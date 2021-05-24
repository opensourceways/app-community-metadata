package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
)

type HelloWorldPlugin struct {
}

func NewHelloWorldPlugin() gitsync.Plugin {
	return &HelloWorldPlugin{}
}

func (h *HelloWorldPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "helloworld",
		Group:       "sample",
		Description: "used for demonstration",
		Repos: []gitsync.GitMeta{
			{
				Repo:       "https://github.com/TommyLike/SampleApp",
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"README.md",
				},
			},
		},
	}
}

func (h *HelloWorldPlugin) Load(files map[string][]string) error {
	return nil
}

func (h *HelloWorldPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	return
}
