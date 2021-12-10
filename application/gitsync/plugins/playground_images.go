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
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"sync/atomic"
)

const PlaygroundImages = " https://github.com/opensourceways/playground-images.git"

type PlaygoundImagesPlugin struct {
	Images atomic.Value
}

func NewPlaygoundImagesPlugin() gitsync.Plugin {
	return &PlaygoundImagesPlugin{}
}

func (h *PlaygoundImagesPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "playground-images",
		Group:       "infrastructure",
		Description: "get all playground images information",
		Repos: []gitsync.GitMeta{
			{
				Repo:       PlaygroundImages,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"deploy",
				},
			},
		},
	}
}

func (h *PlaygoundImagesPlugin) Load(files map[string][]string) error {
	var images string
	if files, ok := files[PlaygroundImages]; ok {
		if len(files) > 0 {
			//there would be only one file possible
			err := filepath.Walk(files[0], func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return nil
				}

				if strings.HasSuffix(path, "lxd-images.yaml") {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()
					bytes, err := ioutil.ReadAll(f)
					if err != nil {
						return err
					}
					m, err := yaml.YAMLToJSON(bytes)
					if err != nil {
						return err
					}
					images = string(m)
				}
				return nil
			})
			if err != nil {
				return err
			}
			h.Images.Store(images)
		}
	}
	return nil
}

func (h *PlaygoundImagesPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/lxd-images", h.ReadLXDImages)
}

func (h *PlaygoundImagesPlugin) ReadLXDImages(c *gin.Context) {
	images := h.Images.Load()
	if repos == nil {
		c.Data(200, "application/json", []byte(""))
	} else {
		c.Data(200, "application/json", []byte(images.(string)))
	}

}
