package gitsync

import (
	"errors"
	"github.com/gookit/color"
	"github.com/gookit/goutil/fsutil"
	"github.com/opensourceways/app-community-metadata/app"
	"math"
	"strconv"
	"sync"
)

var (
	pluginMutex sync.RWMutex
	pluginsContainer = map[string]Plugin{}
	//repo organized as below:
	//group1:
	//		repo1
	//		repo2
	repoContainer = map[string]map[string]*GitMeta{}
)

type SyncManager struct {
	SyncInterval   int
	baseFolder     string
	eventCh        chan string
	notifyInterval int
}

func NewSyncManager() (*SyncManager, error) {
	conf := app.Config.StringMap("manager")
	syncValue, _ := strconv.Atoi(conf["syncInterval"])
	syncInterval := math.Min(float64(syncValue), app.DefaultInterval)
	baseFolder := conf["baseFolder"]
	if !fsutil.DirExist(baseFolder) {
		color.Error.Printf("rsync folder %s not existed", baseFolder)
		return nil, errors.New("rsync folder not existed")
	}
	notifyValue, _ := strconv.Atoi(conf["notifyInterval"])
	notifyInterval := math.Min(float64(notifyValue), app.DefaultInterval)
	color.Info.Printf(
		"============ SyncManager(sync: %d notify: %d baseFolder: %s) ============\n",
		int(syncInterval), int(notifyInterval), baseFolder)

	return &SyncManager{
		SyncInterval: int(syncInterval),
		notifyInterval: int(notifyInterval),
		baseFolder: baseFolder,
		eventCh: make(chan string, app.DefaultSyncChannelSize),
	}, nil
}

func (s *SyncManager) GetBaseFolder() string {
	return s.baseFolder
}

// Register used to for plugin registration
func Register(pluginName string, plugin Plugin) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()
	//update plugin
	pluginsContainer[pluginName] = plugin
	//update repo
	for _, repo := range plugin.GetMeta().Repos {
		localName, err := GetRepoLocalName(&repo)
		if err != nil {
			color.Error.Printf("failed to get local name of %s, error: %v", repo.Repo, err)
		}
		updateRepoContainer(plugin.GetMeta().Group, localName, &repo)
		color.Info.Printf("plugin %s registered to manager", plugin.GetMeta().Name)
	}
}

// Update repo container to hold all repo and watch files
func updateRepoContainer(group, localName string, repo *GitMeta) {
	r, found := repoContainer[group]
	if found {
		g, rfound := r[localName]
		if rfound {
			//error if repo url not equal
			equal, err := RepoEqual(g.Repo, repo.Repo)
			if err != nil {
				color.Error.Printf(
					"failed to compare url equality between %s and %s, err %v", g.Repo, repo.Repo, err)
			}
			if ! equal {
				color.Error.Printf(
					"repo %s skipped due to the existence of same local repo while remote url differs %s and %s",
					g.Repo, repo.Repo)
			} else {
				g.WatchFiles = append(g.WatchFiles, repo.WatchFiles...)
			}
		}else {
			r[localName] = repo
		}
	}else {
		repoContainer[group] = make(map[string]*GitMeta, 0)
		repoContainer[group][localName] = repo
	}
}

func (s *SyncManager) StartLoop() error {
	//start sync worker
	return nil
}
