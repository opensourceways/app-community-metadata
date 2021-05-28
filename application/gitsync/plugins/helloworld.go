package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	"io/ioutil"
	"os"
)

const RepoName = "https://github.com/TommyLike/SampleApp"
const RepoFile = "README.md"
const Dockerfiles = "Dockerfiles"

type HelloWorldPlugin struct {
	content string
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
				Repo:       RepoName,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					RepoFile,
					Dockerfiles,
				},
			},
		},
	}
}

func (h *HelloWorldPlugin) Load(files map[string][]string) error {
	if files, ok := files[RepoName]; ok {
		if len(files) > 0 {
			f, err := os.Open(files[0])
			defer f.Close()
			if err != nil {
				return err
			}
			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			h.content = string(bytes)
		}
	}
	return nil
}

func (h *HelloWorldPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/readme", h.ReadmeContent)
}

func (h *HelloWorldPlugin) ReadmeContent(c *gin.Context) {
	c.JSON(200, h.content)
}
