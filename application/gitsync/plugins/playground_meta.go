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

const PlaygroundImages = "https://github.com/opensourceways/playground-images"
const PlaygroundCourses = "https://github.com/opensourceways/playground-courses"

type PlaygoundMetaPlugins struct {
	Images atomic.Value
	Templates atomic.Value
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
			for _, f := range files {
				fileInfo, err := os.Lstat(f)
				if err != nil {
					fmt.Println(fmt.Sprintf("failed to get file %s in plugin.",  err))
					continue
				}
				if fileInfo.Name() == "lxd-images.yaml" {
					imageFile, err := os.Open(f)
					if err != nil {
						return err
					}
					defer imageFile.Close()
					bytes, err := ioutil.ReadAll(imageFile)
					if err != nil {
						return err
					}
					images, err := yaml.YAMLToJSON(bytes)
					if err != nil {
						return err
					}
					h.Images.Store(images)
				} else if fileInfo.Name() == "environments" {
					templates := make(map[string][]byte)
					err := filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if !info.Mode().IsRegular() {
							return nil
						}
						//read all tmpl file
						if strings.HasSuffix(path, ".tmpl") {
							templateFile, err := os.Open(path)
							if err != nil {
								return err
							}
							defer templateFile.Close()
							bytes, err := ioutil.ReadAll(templateFile)
							if err != nil {
								return err
							}
							m, err := yaml.YAMLToJSON(bytes)
							if err != nil {
								return err
							}
							templates[path] = m
						}
						return nil
					})
					if err != nil {
						return err
					}
					h.Templates.Store(templates)
				}
			}
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
	images := h.Images.Load().(map[string][]byte)
	if images == nil {
		c.Data(200, "application/json", []byte(""))
	} else {
		fmt.Println(images)
		c.Data(200, "application/json", []byte(""))
	}

}

