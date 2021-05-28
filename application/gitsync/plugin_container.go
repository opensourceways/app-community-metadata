package gitsync

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
)

type PluginContainer struct {
	Plugin         Plugin
	Ready          bool
	Channel        chan *GitEvent
	FlushChannel   chan int
	Logger         *zap.Logger
	eventContainer map[string][]string
	eventMutex     sync.Mutex
}

func NewPluginContainer(p Plugin) *PluginContainer {
	container := make(map[string][]string)
	for _, repo := range p.GetMeta().Repos {
		container[repo.Repo] = make([]string, 0)
	}
	return &PluginContainer{
		Plugin:         p,
		Ready:          false,
		Channel:        make(chan *GitEvent, 50),
		FlushChannel:   make(chan int, 10),
		eventContainer: container,
	}
}

func (p *PluginContainer) AddEvents(repo, filename string) {
	defer p.eventMutex.Unlock()
	p.eventMutex.Lock()
	p.eventContainer[repo] = append(p.eventContainer[repo], filename)
}

func (p *PluginContainer) FlushEvents() map[string][]string {
	defer p.eventMutex.Unlock()
	p.eventMutex.Lock()
	results := make(map[string][]string)
	err := DeepCopyMap(p.eventContainer, results)
	if err != nil {
		p.Logger.Error(fmt.Sprintf("failed to copy events to plugins %v", err))
	}
	p.eventContainer = make(map[string][]string)
	return results
}

func (p *PluginContainer) StartLoop() {
	for {
		select {
		case event, ok := <-p.Channel:
			if !ok {
				p.Logger.Info(fmt.Sprintf(
					"plugin container[%s] received close channel event, quiting..", p.Plugin.GetMeta().Name))
				return
			}

			if event.GroupName == p.Plugin.GetMeta().Group {
				p.Logger.Info(fmt.Sprintf("event %v received in plugin container %s", event, p.Plugin.GetMeta().Name))
				r := GetRepo(p.Plugin.GetMeta().Repos, event.RepoName)
				if r != nil {
					eventCount := 0
					for _, f := range event.Files {
						if PathIncludes(r.WatchFiles, f) {
							p.AddEvents(r.Repo, f)
							eventCount += 1
						}
					}
					if eventCount != 0 {
						p.Logger.Info(fmt.Sprintf(
							"plugin container[%s] received git event with %d file changes",
							p.Plugin.GetMeta().Name, eventCount))
					}
				}
			}
		case _, ok := <-p.FlushChannel:
			if !ok {
				p.Logger.Info(fmt.Sprintf(
					"plugin container[%s] received close channel event, quiting..", p.Plugin.GetMeta().Name))
				return
			}
			files := p.FlushEvents()
			if len(files) != 0 {
				err := p.Plugin.Load(files)
				if err != nil {
					p.Logger.Error(fmt.Sprintf("plugin container[%s] triggered LOAD function with error %v",
						p.Plugin.GetMeta().Name, err))
				} else {
					p.Logger.Info(fmt.Sprintf("plugin container[%s] triggered LOAD function",
						p.Plugin.GetMeta().Name))
				}
			}
		}
	}
}

func (p *PluginContainer) Close() {
	close(p.Channel)
	close(p.FlushChannel)
}
