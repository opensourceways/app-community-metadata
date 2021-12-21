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
	"errors"
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
	Images    atomic.Value
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
			fileInfo, err := os.Lstat(files[0])
			if err != nil {
				fmt.Println(fmt.Sprintf("failed to get file %s in plugin.", err))
				return err
			}
			if fileInfo.Name() == "lxd-images.yaml" {
				imageFile, err := os.Open(files[0])
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
			} else {
				return errors.New(fmt.Sprintf("unrecognized file %s", fileInfo.Name()))
			}
		}
	}
	if files, ok := files[PlaygroundCourses]; ok {
		if len(files) > 0 {
			for _, f := range files {
				fileInfo, err := os.Lstat(f)
				if err != nil {
					fmt.Println(fmt.Sprintf("failed to get file %s in plugin.", err))
					continue
				}
				if fileInfo.Name() == "environments" {
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
							templates[path] = bytes
						}
						return nil
					})
					if err != nil {
						return err
					}
					h.Templates.Store(templates)
				} else {
					return errors.New(fmt.Sprintf("unrecognized file %s", fileInfo.Name()))
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
	if h.Templates.Load() == nil {
		c.Data(500, "text/html", []byte("server not ready"))
	} else {
		templates := h.Templates.Load().(map[string][]byte)
		fileQuery := c.Query("file")
		if len(fileQuery) == 0 {
			c.Data(404, "text/html", []byte("please specify 'file' parameter"))
		} else {
			var content []byte
			for k, v := range templates {
				if strings.Contains(k, fileQuery) {
					content = v
					break
				}
			}
			if len(content) == 0 {
				c.Data(404, "text/html", []byte(fmt.Sprintf("%s not found", fileQuery)))
			} else {
				c.Data(200, "application/json", content)
			}
		}
	}

}
