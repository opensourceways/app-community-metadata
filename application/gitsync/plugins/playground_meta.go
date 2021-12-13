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
	"sigs.k8s.io/yaml"
	"sync/atomic"
)

const PlaygroundImages = "https://github.com/opensourceways/playground-images"
const PlaygroundCourses = "https://github.com/opensourceways/playground-courses"

type PlaygoundMetaPlugins struct {
	Images atomic.Value
}

func NewPlaygoundMetaPlugin() gitsync.Plugin {
	return &PlaygoundMetaPlugins{}
}

func (h *PlaygoundMetaPlugins) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "playground-meta",
		Group:       "infrastructure",
		Description: "get all playground meta information",
		Repos: []gitsync.GitMeta{
			{
				Repo:       PlaygroundImages,
				Branch:     "main",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"deploy/lxd-images.yaml",
				},
			},
			{
				Repo:       PlaygroundCourses,
				Branch:     "main",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"environments",
				},
			},
		},
	}
}

func (h *PlaygoundMetaPlugins) Load(files map[string][]string) error {
	if files, ok := files[PlaygroundImages]; ok {
		if len(files) > 0 {
			f, err := os.Open(files[0])
			if err != nil {
				return err
			}
			defer f.Close()
			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			images, err := yaml.YAMLToJSON(bytes)
			if err != nil {
				return err
			}
			h.Images.Store(images)
		}
	}
	return nil
}

func (h *PlaygoundMetaPlugins) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/images", h.ReadImages)
	group.GET("/templates", h.ReadTemplates)
}

func (h *PlaygoundMetaPlugins) ReadImages(c *gin.Context) {
	images := h.Images.Load()
	if images == nil {
		c.Data(200, "application/json", []byte(""))
	} else {
		c.Data(200, "application/json", images.([]byte))
	}

}

func (h *PlaygoundMetaPlugins) ReadTemplates(c *gin.Context) {
	images := h.Images.Load()
	if images == nil {
		c.Data(200, "application/json", []byte(""))
	} else {
		c.Data(200, "application/json", images.([]byte))
	}

}

