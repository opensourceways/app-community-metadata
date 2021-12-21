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

const InfrastructureRepo = "https://gitee.com/openeuler/infrastructure"

type OpenEulerMirrorsPlugin struct {
	Repos atomic.Value
}

func NewOpenEulerMirrorsPlugin() gitsync.Plugin {
	return &OpenEulerMirrorsPlugin{}
}

func (h *OpenEulerMirrorsPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "mirrors",
		Group:       "openeuler",
		Description: "get all openeuler mirror information",
		Repos: []gitsync.GitMeta{
			{
				Repo:       InfrastructureRepo,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"mirrors",
				},
			},
		},
	}
}

func (h *OpenEulerMirrorsPlugin) Load(files map[string][]string) error {
	mirrors := []string{}
	if files, ok := files[InfrastructureRepo]; ok {
		if len(files) > 0 {
			//walk the yaml file to collect all mirror sites
			err := filepath.Walk(files[0], func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.Mode().IsRegular() {
					return nil
				}

				if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
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
					mirrors = append(mirrors, string(m))
				}
				return nil
			})
			if err != nil {
				return err
			}
			h.Repos.Store(mirrors)
		}
	}
	return nil
}

func (h *OpenEulerMirrorsPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/all", h.ReadMirrorYamls)
}

func (h *OpenEulerMirrorsPlugin) SkipRequestLogs() []gitsync.SkipLogMeta {
	return []gitsync.SkipLogMeta{}
}

func (h *OpenEulerMirrorsPlugin) ReadMirrorYamls(c *gin.Context) {
	repos := h.Repos.Load()
	if repos == nil {
		c.Data(200, "application/json", []byte("[]"))
	} else {
		c.Data(200, "application/json", []byte(fmt.Sprintf("[%s]",
			strings.Join(repos.([]string), ","))))
	}

}
