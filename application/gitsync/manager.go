/*
Copyright 2021 The Opensourceways Group.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitsync

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"github.com/gookit/goutil/fsutil"
	"github.com/opensourceways/app-community-metadata/app"
	"go.uber.org/zap"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	loopbackAddress = "127.0.0.1"
)

var (
	pluginMutex      sync.RWMutex
	pluginsContainer = map[string]*PluginContainer{}
	//repo organized as below:
	//group1:
	//		repo1(localpath)
	//		repo2
	repoContainer = map[string]map[string]*GitMetaContainer{}
	repoMutex     sync.RWMutex
)

type SyncManager struct {
	SyncInterval   int
	baseFolder     string
	eventCh        chan *GitEvent
	notifyInterval int
	Runners        map[string]Runner
	logger         *zap.Logger
	events         map[string]*GitEvent
	routerGroup    *gin.RouterGroup
	gitSyncPath    string
	validateID     int
	enabledplugins map[string]*PluginContainer
}

func NewSyncManager(routerGroup *gin.RouterGroup) (*SyncManager, error) {
	conf := app.Config.StringMap("manager")
	syncValue, _ := strconv.Atoi(conf["syncInterval"])
	syncInterval := math.Min(float64(syncValue), app.DefaultInterval)
	baseFolder := conf["baseFolder"]
	if !fsutil.DirExist(baseFolder) {
		color.Error.Printf("rsync folder %s not existed", baseFolder)
		return nil, errors.New("rsync folder not existed")
	}
	baseFolder, _ = filepath.Abs(baseFolder)
	gitSyncPath, _ := conf["gitSyncPath"]
	if !fsutil.FileExist(gitSyncPath) {
		lookPath, err := exec.LookPath("git-sync")
		if err != nil {
			color.Error.Printf("git sync binary %s not found\n", lookPath)
			return nil, errors.New(fmt.Sprintf("git sync binary %s not found. ", lookPath))
		}
		gitSyncPath = lookPath
	}
	notifyValue, _ := strconv.Atoi(conf["notifyInterval"])
	notifyInterval := math.Min(float64(notifyValue), app.DefaultInterval)
	color.Info.Printf(
		"============ SyncManager(sync: %d notify: %d baseFolder: %s) ============\n",
		int(syncInterval), int(notifyInterval), baseFolder)

	//update plugin container's logger
	for _, v := range pluginsContainer {
		v.Logger = app.Logger
	}

	return &SyncManager{
		SyncInterval:   int(syncInterval),
		notifyInterval: int(notifyInterval),
		baseFolder:     baseFolder,
		eventCh:        make(chan *GitEvent, app.DefaultSyncChannelSize),
		logger:         app.Logger,
		Runners:        make(map[string]Runner),
		events:         make(map[string]*GitEvent),
		routerGroup:    routerGroup,
		gitSyncPath:    gitSyncPath,
		enabledplugins: make(map[string]*PluginContainer, len(pluginsContainer)),
	}, nil
}

func (s *SyncManager) GetEnabledPlugins() map[string]*PluginContainer {
	//return if initialized.
	if len(s.enabledplugins) != 0 {
		return s.enabledplugins
	}
	for name, instance := range pluginsContainer {
		cfg := app.Config.StringMap(fmt.Sprintf("plugins.%s", name))
		if rs, ok := cfg["enabled"]; ok {
			enabled, _ := strconv.ParseBool(rs)
			if !enabled {
				s.logger.Info(fmt.Sprintf("Plugin [%s] disabled by config", name))
				continue
			}
		} else {
			s.logger.Info(fmt.Sprintf("Plugin [%s] disabled by config", name))
			continue
		}
		s.enabledplugins[name] = instance
	}
	return s.enabledplugins
}

func (s *SyncManager) OnePluginInitialized() bool {
	for _, container := range s.GetEnabledPlugins() {
		if container.Ready == true {
			return true
		}
	}
	return false
}

func (s *SyncManager) initializePluginWhenReady(event *GitEvent) {
	defer repoMutex.Unlock()
	repoMutex.Lock()
	//check whether plugin container is ready to register endpoint and handle message
	for _, container := range s.GetEnabledPlugins() {
		if !container.Ready {
			//filter out mismatch router plugins
			if event.GroupName == container.Plugin.GetMeta().Group {
				//whether all repos are ready
				readyRepos := 0
				repos, ok := repoContainer[event.GroupName]
				if !ok {
					continue
				}
				for _, r := range container.Plugin.GetMeta().Repos {
					if m, ok := repos[GetRepoLocalName(r.Repo)]; ok {
						if m.Ready {
							readyRepos += 1
						}
					}
				}
				if readyRepos == len(container.Plugin.GetMeta().Repos) {
					//Register endpoints
					container.Plugin.RegisterEndpoints(s.routerGroup.Group(container.Plugin.GetMeta().Group).Group(container.Plugin.GetMeta().Name))
					go container.StartLoop()
					container.Ready = true
					s.logger.Info(fmt.Sprintf("plugin %s/%s initialized.", container.Plugin.GetMeta().Group, container.Plugin.GetMeta().Name))
				}
			}
		}
	}
}

func (s *SyncManager) dispatchEvents(event *GitEvent) {
	for _, container := range s.GetEnabledPlugins() {
		container.Channel <- event
	}
}
func (s *SyncManager) dispatchFlushEvents(event int) {
	for _, container := range s.GetEnabledPlugins() {
		container.FlushChannel <- event
	}
}

func (s *SyncManager) GetBaseFolder() string {
	return s.baseFolder
}

// Register used to for plugin registration
func Register(pluginName string, plugin Plugin) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()
	//update plugin
	pluginsContainer[pluginName] = NewPluginContainer(plugin)
}

// Update repo container to hold all repo and watch files
func updateRepoContainer(group, localName string, repo GitMeta) {
	r, found := repoContainer[group]
	if found {
		g, rfound := r[localName]
		if rfound {
			//error if repo url not equal
			equal, err := RepoEqualIgnoreSchemaAndLevel(g.Meta.Repo, repo.Repo)
			if err != nil {
				color.Error.Printf(
					"failed to compare url equality between %s and %s, err %v", g.Meta.Repo, repo.Repo, err)
			}
			if !equal {
				color.Error.Printf(
					"repo %s skipped due to the existence of same local repo while remote url differs %s and %s",
					g.Meta.Repo, repo.Repo)
			} else {
				g.Meta.WatchFiles = append(g.Meta.WatchFiles, repo.WatchFiles...)
			}
		} else {
			r[localName] = &GitMetaContainer{
				Meta:  &repo,
				Ready: false,
			}
		}
	} else {
		repoContainer[group] = make(map[string]*GitMetaContainer, 0)
		repoContainer[group][localName] = &GitMetaContainer{
			Meta:  &repo,
			Ready: false,
		}
	}
}

func PluginDetails(c *gin.Context) {
	data := make([]map[string]string, 0)
	defer pluginMutex.RUnlock()
	pluginMutex.RLock()
	for _, p := range pluginsContainer {
		data = append(data, map[string]string{
			"group":       strings.ToLower(p.Plugin.GetMeta().Group),
			"name":        strings.ToLower(p.Plugin.GetMeta().Name),
			"ready":       strings.ToLower(strconv.FormatBool(p.Ready)),
			"description": strings.ToLower(p.Plugin.GetMeta().Description),
			//TODO:add more metadata to plugins
		})
	}
	c.JSON(200, data)
}

func (s *SyncManager) repoUpdateNotify(c *gin.Context) {
	validateID := c.Query("validateID")
	//only allowed from local
	if c.ClientIP() != loopbackAddress || validateID != strconv.Itoa(s.validateID) {
		c.JSON(403, nil)
		return
	}
	group := c.Param("group")
	localName := c.Param("localname")

	if r, ok := s.Runners[fmt.Sprintf("%s/%s", group, localName)]; ok {
		r.RepoUpdated()
		c.JSON(200, nil)
		return
	} else {
		s.logger.Info(fmt.Sprintf("group: %s repo %s not found in repo runner", group, localName))
		c.JSON(404, nil)
		return
	}
}

func (s *SyncManager) getRepoTriggerEndpoint(group, localName string) string {
	return fmt.Sprintf("http://%s:%d%s/repos/%s/%s/trigger?validateID=%d",
		loopbackAddress, app.HttpPort, s.routerGroup.BasePath(), group, localName, s.validateID)
}

func (s *SyncManager) Initialize() error {
	s.validateID = time.Now().Nanosecond()
	s.routerGroup.GET("/plugins", PluginDetails)
	s.routerGroup.GET("/repos/:group/:localname/trigger", s.repoUpdateNotify)
	//update repo container
	for _, plugin := range s.GetEnabledPlugins() {
		for _, repo := range plugin.Plugin.GetMeta().Repos {
			localName := GetRepoLocalName(repo.Repo)
			if localName == "" {
				color.Error.Printf("Failed to get local name of %s", repo.Repo)
			}
			updateRepoContainer(plugin.Plugin.GetMeta().Group, localName, repo)
			s.logger.Info(fmt.Sprintf("Plugin [%s/%s] registered to manager %s", plugin.Plugin.GetMeta().Group, plugin.Plugin.GetMeta().Name,
				localName))
		}
	}
	//initialize repo container with runner
	for group, metas := range repoContainer {
		groupPath := path.Join(s.baseFolder, group)
		for _, meta := range metas {
			localName := GetRepoLocalName(meta.Meta.Repo)
			if localName == "" {
				s.logger.Error(fmt.Sprintf("failed to find local name for repo: %s", meta.Meta.Repo))
				continue
			}
			localPath := filepath.Join(groupPath, localName)
			err := fsutil.Mkdir(localPath, os.FileMode(0755))
			if err != nil {
				s.logger.Error(fmt.Sprintf("failed to create folder for repo: %s", meta.Meta))
				continue
			}
			r, err := NewGitSyncRunner(group, localPath, meta.Meta, s.eventCh, s.SyncInterval, s.logger, s.gitSyncPath,
				s.getRepoTriggerEndpoint(group, localName))
			if err != nil {
				s.logger.Error(fmt.Sprintf("failed to create runner for repo: %s, err: %v", meta.Meta.Repo, err))
				continue
			}

			s.Runners[fmt.Sprintf("%s/%s", group, localName)] = r
		}
	}
	if len(s.Runners) == 0 {
		return errors.New("no plugin configured")
	}
	s.logger.Info("sync manager successfully started")
	return nil
}

func (s *SyncManager) StartLoop() {
	//start sync worker
	for _, r := range s.Runners {
		go r.StartLoop()
	}
	if len(s.Runners) == 0 {
		s.logger.Error(fmt.Sprintf("no sync runner available"))
		return
	}
	// start loop to receive channel event
	go s.handleEvents()
}

func (s *SyncManager) handleEvents() {
	//1. collect event from worker
	//2. push events on notify event
	//3. register endpoint on first ready event
	ticker := time.NewTicker(time.Duration(s.notifyInterval) * time.Second)
	for {
		select {
		case event, ok := <-s.eventCh:
			if !ok {
				return
			}
			eventHandled := false
			if g, ok := repoContainer[event.GroupName]; ok {
				if r, ok := g[GetRepoLocalName(event.RepoName)]; ok {
					if !r.Ready {
						s.logger.Info(fmt.Sprintf("repo %s initialized.", r.Meta.Repo))
						r.Ready = true
					}
					s.initializePluginWhenReady(event)
					s.dispatchEvents(event)
					eventHandled = true
				}
			}
			if !eventHandled {
				s.logger.Info(fmt.Sprintf("event %v discarded due to unable to located repo from container",
					event))
			}
		case <-ticker.C:
			s.dispatchFlushEvents(0)
		}
	}
}

func (s *SyncManager) Close() {
	//close worker
	for key, runner := range s.Runners {
		if err := runner.Close(); err != nil {
			s.logger.Error(fmt.Sprintf("failed to git runner for repo: %s, will be skipped", key))
		}
	}
	//close plugin container
	for _, plugin := range s.GetEnabledPlugins() {
		plugin.Close()
	}
	close(s.eventCh)
}
