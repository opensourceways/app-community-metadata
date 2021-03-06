package gitsync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gookit/goutil/fsutil"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const SyncTimeout = 300 //5 minutes at most
const DirectoryWalkTimeout = 30
const DefaultSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
const MaxDelay = 5
const MaxCalculateFiles = 100

type GitSyncRunner struct {
	ParentFolder    string
	Meta            *GitMeta
	EventChannel    chan<- *GitEvent
	CloseChannel    chan bool
	SyncInterval    int
	logger          *zap.Logger
	watchFiles      map[string]string
	group           string
	gitSyncPath     string
	WebhookEndpoint string
	closed          bool
}

type HashResult struct {
	path string
	hash string
	err  error
}

func NewGitSyncRunner(group, parentFolder string, repo *GitMeta, eventChannel chan<- *GitEvent, interval int, logger *zap.Logger, gitSyncPath string, webhookEndpoint string) (*GitSyncRunner, error) {
	if !fsutil.DirExist(parentFolder) {
		return nil, errors.New(fmt.Sprintf("parent folder %s doesn't exist", parentFolder))
	}
	//convert relative path into abs
	watchFiles := make(map[string]string)
	for _, r := range repo.WatchFiles {
		//NOTE:
		//git sync will create a nested folder inside and perform file link switch when updated, therefore, the full
		//file path would be like:
		//repo: https://github.com/repo.git
		//watch file: README.md
		//group name: group1
		//local repo path: /developing
		//full file path: /developing/group1/repo/repo/README.md
		path := filepath.Join(parentFolder, GetRepoLocalName(repo.Repo), r)
		watchFiles[path] = DefaultSHA256
	}
	return &GitSyncRunner{
		ParentFolder:    parentFolder,
		Meta:            repo,
		EventChannel:    eventChannel,
		SyncInterval:    interval,
		logger:          logger,
		watchFiles:      watchFiles,
		group:           group,
		gitSyncPath:     gitSyncPath,
		CloseChannel:    make(chan bool, 1),
		WebhookEndpoint: webhookEndpoint,
		closed:          false,
	}, nil
}

func (g *GitSyncRunner) RepoUpdated() {
	g.logger.Info(fmt.Sprintf("repo %s commit id changed.", g.Meta.Repo))
	g.CompareDigestAndNotify()
}

func (g *GitSyncRunner) runCommand(ctx context.Context, cwd, command string, args ...string) (string, error) {
	return g.runCommandWithStdin(ctx, cwd, "", command, args...)
}

func cmdForLog(command string, args ...string) string {
	if strings.ContainsAny(command, " \t\n") {
		command = fmt.Sprintf("%q", command)
	}
	argsCopy := make([]string, len(args))
	copy(argsCopy, args)
	for i := range args {
		if strings.ContainsAny(args[i], " \t\n") {
			argsCopy[i] = fmt.Sprintf("%q", args[i])
		}
	}
	return command + " " + strings.Join(argsCopy, " ")
}

func (g *GitSyncRunner) runCommandWithStdin(ctx context.Context, cwd, stdin, command string, args ...string) (string, error) {
	cmdStr := cmdForLog(command, args...)
	g.logger.Info(fmt.Sprintf("running command cwd %s cmd %s", cwd, cmdStr))

	cmd := exec.CommandContext(ctx, command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf
	cmd.Stdin = bytes.NewBufferString(stdin)

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("run(%s): %w: { stdout: %q, stderr: %q }", cmdStr, ctx.Err(), stdout, stderr)
	}
	if err != nil {
		return "", fmt.Errorf("run(%s): %w: { stdout: %q, stderr: %q }", cmdStr, err, stdout, stderr)
	}
	g.logger.Info(fmt.Sprintf("command result stdout %q, stderr %q", stdout, stderr))

	return stdout, nil
}

func (g *GitSyncRunner) SyncRepo(ctx context.Context, onetime bool) bool {
	args := []string{"--repo", g.Meta.Repo, "--root", g.ParentFolder, "--branch", g.Meta.Branch}
	if len(g.Meta.SubModules) != 0 {
		args = append(args, []string{"--submodules", g.Meta.SubModules}...)
	}
	if !onetime {
		//append webhook related parameters
		args = append(args, []string{"-webhook-url", g.WebhookEndpoint, "-webhook-method", "GET", "--webhook-timeout", "2s"}...)
		args = append(args, []string{"--wait", fmt.Sprintf("%d", g.SyncInterval)}...)
	} else {
		args = append(args, []string{"--one-time"}...)
	}
	_, err := g.runCommand(ctx, "", g.gitSyncPath, args...)
	if err != nil {
		g.logger.Error(fmt.Sprintf("failed to perform git sync operation %s %v", g.Meta.Repo, err))
		return false
	}
	return true
}

func (g *GitSyncRunner) CompareDigestAndNotify() {
	var changedFiles []string
	var newDigest string
	var err error
	for k := range g.watchFiles {
		if fsutil.IsDir(k) {
			newDigest = g.CalculateDigestForDirectory(k)
			if newDigest == "" {
				g.logger.Error(fmt.Sprintf("directory %s skipping watch", k))
				continue
			}
		}
		if fsutil.FileExist(k) {
			newDigest, err = g.CalculateDigestForSingleFile(k)
			if err != nil {
				g.logger.Error(fmt.Sprintf("failed to calculate file digest, error %v. skipping watch", err))
				continue
			}
		}
		if newDigest != g.watchFiles[k] {
			g.watchFiles[k] = newDigest
			changedFiles = append(changedFiles, k)
		}
	}
	if len(changedFiles) != 0 {
		event := GitEvent{
			RepoName:  g.Meta.Repo,
			GroupName: g.group,
			Files:     changedFiles,
		}
		g.logger.Info(fmt.Sprintf("new changes detected for repo %s, files %v", g.Meta.Repo, changedFiles))
		g.EventChannel <- &event
	}
}

func (g *GitSyncRunner) CalculateDigestForSingleFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}

func (g *GitSyncRunner) processFileHash(filePath string, done chan struct{}) (<-chan HashResult, <-chan error) {
	resultChannel := make(chan HashResult, 20)
	errorChannel := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		var files = 0
		//NOTE: Improve the performance by calculating top N files only.
		err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			files += 1
			if files > MaxCalculateFiles {
				g.logger.Warn(fmt.Sprintf("only %d files will be calculated for directiry digest,"+
					"rest will be skipped", MaxCalculateFiles))
				return nil
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				f, err := os.Open(path)
				if err != nil {
					resultChannel <- HashResult{path, "", err}
					return
				}
				defer f.Close()
				h := sha256.New()
				if _, err := io.Copy(h, f); err != nil {
					resultChannel <- HashResult{path, "", err}
					return
				}
				resultChannel <- HashResult{path, string(h.Sum(nil)), nil}
			}()
			select {
			case <-done: // HL
				return errors.New(fmt.Sprintf("walk directory %s canceled", filePath))
			default:
				return nil
			}
		})
		go func() {
			wg.Wait()
			close(resultChannel)
			close(errorChannel)
		}()
		if err != nil {
			errorChannel <- err
		}
	}()
	return resultChannel, errorChannel
}

func (g *GitSyncRunner) CalculateDigestForDirectory(filepath string) string {
	var hashes []string
	doneChannel := make(chan struct{})
	resultChannel, errorChannel := g.processFileHash(filepath, doneChannel)
	ticker := time.NewTimer(time.Duration(DirectoryWalkTimeout) * time.Second)
	for {
		select {
		case <-ticker.C:
			g.logger.Error(fmt.Sprintf("calculate directory %s hashes timed out",
				filepath))
			close(doneChannel)

			return ""
		case e, ok := <-errorChannel:
			if ok {
				g.logger.Error(fmt.Sprintf("failed to calculate %s hashes, error %v",
					filepath, e))
				close(doneChannel)
				return ""
			}
		case result, ok := <-resultChannel:
			if ok {
				if result.err != nil {
					g.logger.Warn(fmt.Sprintf("failed to calculate file digest %s due to error %v",
						result.path, result.err))
				} else {
					//we only care about hash currently
					hashes = append(hashes, result.hash)
				}
			} else {
				//calculate result
				sort.Strings(hashes)
				g.logger.Info(fmt.Sprintf("%d files calculated for digesting directory %s", len(hashes),
					filepath))
				h := sha256.New()
				for _, c := range hashes {
					h.Write([]byte(c))
				}
				return string(h.Sum(nil))
			}
		}
	}
}

func (g *GitSyncRunner) WatchSync(ctx context.Context) {
	retry := 1
	for {
		if g.closed {
			g.logger.Info(fmt.Sprintf("received cancel signal, quit git sync..."))
			return
		}
		g.logger.Info(fmt.Sprintf("loop perform git sync (current: %d) for repo %s", retry, g.Meta.Repo))
		//basically it won't quit unless program fails
		_ = g.SyncRepo(ctx, false)
		g.logger.Error(fmt.Sprintf("repo [%s] failed to sync, application will exit, check log for detail",
			g.Meta.Repo))
		retry += 1
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(rand.Intn(MaxDelay)) * time.Second)
	}
}

func (g *GitSyncRunner) StartLoop() {
	//first clone or update
	ctx, _ := context.WithTimeout(context.Background(), time.Second*SyncTimeout)
	success := g.SyncRepo(ctx, true)
	if success {
		g.logger.Info(fmt.Sprintf("repo [%s] successfully cloned", g.Meta.Repo))
		g.CompareDigestAndNotify()
		//start watching with cancel context
		ctx, cancel := context.WithCancel(context.Background())
		go g.WatchSync(ctx)
		for {
			select {
			case _, ok := <-g.CloseChannel:
				if !ok {
					cancel()
					time.Sleep(2 * time.Second)
					g.logger.Info(fmt.Sprintf("git sync runner for repo [%s] received close event, quiting..",
						g.Meta.Repo))
					return
				}
			}
		}
	} else {
		g.logger.Error(fmt.Sprintf("repo [%s] failed to clone", g.Meta.Repo))
	}
}

func (g *GitSyncRunner) Close() error {
	g.closed = true
	close(g.CloseChannel)
	return nil
}

func (g *GitSyncRunner) GetRepo() *GitMeta {
	return g.Meta
}
