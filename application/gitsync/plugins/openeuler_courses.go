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
	"strings"
	"sync"
	"sync/atomic"
)

const OpenEulerMoocStudioCourses = "https://github.com/opensourceways/playground-courses"

type OpenEulerMoocStudioMetaPlugins struct {
	Images     atomic.Value
	Templates  atomic.Value
	Group      *gin.RouterGroup
	StaticOnce sync.Once
}

func NewOpenEulerMoocStudioMetaPlugins() gitsync.Plugin {
	return &OpenEulerMoocStudioMetaPlugins{}
}

func (h *OpenEulerMoocStudioMetaPlugins) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "openeulermoocstudio",
		Group:       "infrastructure",
		Description: "get all mooc studio courses information for openEuler",
		Repos: []gitsync.GitMeta{
			{
				Repo:       OpenEulerMoocStudioCourses,
				Branch:     "main",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"environments",
					"courses",
				},
			},
		},
	}
}

func (h *OpenEulerMoocStudioMetaPlugins) Load(files map[string][]string) error {
	if files, ok := files[OpenEulerMoocStudioCourses]; ok {
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
				} else if fileInfo.Name() == "courses" {
					//path will not change
					h.StaticOnce.Do(func() {
						h.Group.Static("courses", f)
					})
				} else {
					return errors.New(fmt.Sprintf("unrecognized file %s", fileInfo.Name()))
				}
			}
		}
	}
	return nil
}

func (h *OpenEulerMoocStudioMetaPlugins) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/templates", h.ReadTemplates)
	h.Group = group
}

func (h *OpenEulerMoocStudioMetaPlugins) ReadTemplates(c *gin.Context) {
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
