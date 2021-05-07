package runner

import (
	"errors"
	"github.com/gookit/goutil/fsutil"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
)

type GitSyncRunner struct {
	ParentFolder string
	Repo *gitsync.GitMeta
	EventChannel chan <- *gitsync.GitEvent
	SyncInterval int
}

func NewGitSyncRunner(parentFolder string, repo *gitsync.GitMeta, eventChannel chan <- *gitsync.GitEvent, interval int) (*GitSyncRunner, error) {
	if !fsutil.DirExist(parentFolder) {
		return nil, errors.New("parent folder doesn't exist")
	}
	return &GitSyncRunner{
		ParentFolder: parentFolder,
		Repo: repo,
		EventChannel: eventChannel,
		SyncInterval: interval,
	}, nil
}

func (g *GitSyncRunner) StartLoop() {
}

func (g *GitSyncRunner) Close() error {
	return nil
}

func (g *GitSyncRunner) GetRepo() *gitsync.GitMeta {
	return g.Repo
}




