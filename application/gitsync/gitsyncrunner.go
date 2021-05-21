package gitsync

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gookit/goutil/fsutil"
	"go.uber.org/zap"
	"path/filepath"
)

type GitSyncRunner struct {
	ParentFolder string
	Meta         *GitMeta
	EventChannel chan<- *GitEvent
	SyncInterval int
	watcher      *fsnotify.Watcher
	logger       *zap.Logger
	watchFiles   map[string]string
	group        string
}

func NewGitSyncRunner(group, parentFolder string, repo *GitMeta, eventChannel chan<- *GitEvent, interval int, logger *zap.Logger) (*GitSyncRunner, error) {
	if !fsutil.DirExist(parentFolder) {
		return nil, errors.New("parent folder doesn't exist")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	//convert relative path into abs
	watchFiles := make(map[string]string)
	for _, r := range repo.WatchFiles {
		path, _ := filepath.Abs(r)
		watchFiles[path] = r
	}
	return &GitSyncRunner{
		ParentFolder: parentFolder,
		Meta:         repo,
		EventChannel: eventChannel,
		SyncInterval: interval,
		watcher:      watcher,
		logger:       logger,
		watchFiles:   watchFiles,
		group: 		  group,
	}, nil
}

func (g *GitSyncRunner) StartLoop() {
	//clone and watch repo
	//watch file changes
	//1. watch parent folder instead of file itself
	//2. send out event only after the file is confirmed existed
	g.Watching()
}

func (g *GitSyncRunner) Watching() {
	for {
		select {
		case event, ok := <-g.watcher.Events:
			if !ok {
				return
			}
			g.logger.Info(fmt.Sprintf("%s received event %s", g.Meta.Repo, event))
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				for k, path := range g.watchFiles {
					if path == event.Name {
						//ensure file exists
						if fsutil.FileExist(event.Name) {
							event := &GitEvent{
								GroupName: g.group,
								RepoName:  g.Meta.Repo,
								Files: []string{
									k,
								},
							}
							g.EventChannel <- event
						}
					}
				}
			}
		case err, ok := <-g.watcher.Errors:
			if !ok {
				return
			}
			g.logger.Info(fmt.Sprintf("%s received error %v", g.Meta.Repo, err))
		}
	}
}

func (g *GitSyncRunner) Close() error {
	g.watcher.Close()
	return nil
}

func (g *GitSyncRunner) GetRepo() *GitMeta {
	return g.Meta
}
