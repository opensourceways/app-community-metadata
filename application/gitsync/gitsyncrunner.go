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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const SyncTimeout = 60
const SyncRetry = 3
const DefaultSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"

type GitSyncRunner struct {
	ParentFolder string
	Meta         *GitMeta
	EventChannel chan<- *GitEvent
	CloseChannel chan bool
	SyncInterval int
	logger       *zap.Logger
	watchFiles   map[string]string
	group        string
	gitSyncPath  string
}

func NewGitSyncRunner(group, parentFolder string, repo *GitMeta, eventChannel chan<- *GitEvent, interval int, logger *zap.Logger, gitSyncPath string) (*GitSyncRunner, error) {
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
		ParentFolder: parentFolder,
		Meta:         repo,
		EventChannel: eventChannel,
		SyncInterval: interval,
		logger:       logger,
		watchFiles:   watchFiles,
		group:        group,
		gitSyncPath:  gitSyncPath,
		CloseChannel: make(chan bool, 1),
	}, nil
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

func (g *GitSyncRunner) SyncRepo() bool {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(SyncTimeout))
	args := []string{"--repo", g.Meta.Repo, "--root", g.ParentFolder, "--branch", g.Meta.Branch, "--one-time"}
	if len(g.Meta.SubModules) != 0 {
		args = append(args, []string{"--submodules", g.Meta.SubModules}...)
	}
	result, err := g.runCommand(ctx, "", g.gitSyncPath, args...)
	if err != nil {
		g.logger.Error(fmt.Sprintf("failed to perform git sync operation %s %v", g.Meta.Repo, err))
		return false
	}
	g.logger.Info(result)
	return true

}

func (g *GitSyncRunner) CompareDigestAndNotify() {
	var changedFiles []string
	for k, _ := range g.watchFiles {
		if fsutil.FileExist(k) {
			newDigest, err := g.CalculateDigestForSingleFile(k)
			if err != nil {
				g.logger.Error(fmt.Sprintf("failed to calculate file digest, error %v. skipping watch", err))
				continue
			}
			if newDigest != g.watchFiles[k] {
				g.watchFiles[k] = newDigest
				//convert abs path to relative path
				rootFolder := filepath.Join(g.ParentFolder, GetRepoLocalName(g.Meta.Repo))
				rel, err := filepath.Rel(rootFolder, k)
				if err != nil {
					g.logger.Error(fmt.Sprintf("failed to calculate relative path of file %s base folder %s,"+
						"error %v, skip watching", k, rootFolder, err))
					continue
				}
				changedFiles = append(changedFiles, rel)
			}
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
//TODO: support calculate folder digest: https://blog.golang.org/pipelines/parallel.go
func (g GitSyncRunner) CalculateDigestForSingleFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}

func (g *GitSyncRunner) StartLoop() {
	success := g.SyncRepo()
	if !success {
		return
	}
	g.logger.Info("repo successfully cloned")
	g.CompareDigestAndNotify()
	g.Watching()
}

func (g *GitSyncRunner) Watching() {
	ticker := time.NewTicker(time.Duration(g.SyncInterval) * time.Second)
	retryCount := 0
	for {
		select {
		case <-ticker.C:
			success := g.SyncRepo()
			if !success {
				retryCount += 1
				g.logger.Error(fmt.Sprintf("failed to perform repo sync operation, current retry [%d]", retryCount))
			}
			if retryCount >= SyncRetry {
				return
			}
			if success {
				g.CompareDigestAndNotify()
			}
		case _, ok := <-g.CloseChannel:
			if !ok {
				g.logger.Info("git sync runner received close event, quiting..")
				return
			}
		}
	}
}

func (g *GitSyncRunner) Close() error {
	close(g.CloseChannel)
	return nil
}

func (g *GitSyncRunner) GetRepo() *GitMeta {
	return g.Meta
}
