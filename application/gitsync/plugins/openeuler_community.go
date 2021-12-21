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

const CommunityRepo = "https://gitee.com/openeuler/community"

type OpenEulerCommunityPlugin struct {
	Sigs atomic.Value
}

func NewOpenEulerCommunityPlugin() gitsync.Plugin {
	return &OpenEulerCommunityPlugin{}
}

func (h *OpenEulerCommunityPlugin) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "community",
		Group:       "openeuler",
		Description: "get openEuler community information",
		Repos: []gitsync.GitMeta{
			{
				Repo:       CommunityRepo,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"sig/sigs.yaml",
				},
			},
		},
	}
}

func (h *OpenEulerCommunityPlugin) Load(files map[string][]string) error {
	if files, ok := files[CommunityRepo]; ok {
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
			sigs, err := yaml.YAMLToJSON(bytes)
			if err != nil {
				return err
			}
			h.Sigs.Store(sigs)
		}
	}
	return nil
}

func (h *OpenEulerCommunityPlugin) RegisterEndpoints(group *gin.RouterGroup) {
	group.GET("/sigs", h.ReadSigsYaml)
}

func (h *OpenEulerCommunityPlugin) SkipRequestLogs() []gitsync.SkipLogMeta {
	return []gitsync.SkipLogMeta{}
}

func (h *OpenEulerCommunityPlugin) ReadSigsYaml(c *gin.Context) {
	sigs := h.Sigs.Load()
	if sigs == nil {
		c.Data(200, "application/json", []byte("[]"))
	} else {
		c.Data(200, "application/json", sigs.([]byte))
	}

}
