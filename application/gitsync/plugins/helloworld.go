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
	}
}

func (h *HelloWorldPlugin) Load(filesmap map[string][]string) error {
	return nil
}

func (h *HelloWorldPlugin) RegisterEndpoints(group gin.RouterGroup) {
	return
}
