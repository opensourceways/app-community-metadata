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

package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	"io/ioutil"
	"os"
	"strings"
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
			if strings.HasSuffix(files[0], RepoFile) {
				f, err := os.Open(files[0])
				if err != nil {
					return err
				}
				defer f.Close()
				bytes, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				h.content = string(bytes)
			}
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
