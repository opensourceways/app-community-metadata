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
	"os"
	"sync/atomic"
)

const OpenDesignResources = "https://gitee.com/openeuler/opendesign-resources"

type OpenDesignResourcesPlugins struct {
	Images     atomic.Value
	Templates  atomic.Value
	Group      *gin.RouterGroup
}

func NewOpenDesignResourcesPlugins() gitsync.Plugin {
	return &OpenDesignResourcesPlugins{}
}

func (h *OpenDesignResourcesPlugins) GetMeta() *gitsync.PluginMeta {
	return &gitsync.PluginMeta{
		Name:        "opendesign",
		Group:       "openeuler",
		Description: "get all resource for open design sig",
		Repos: []gitsync.GitMeta{
			{
				Repo:       OpenDesignResources,
				Branch:     "master",
				SubModules: "recursive",
				Schema:     gitsync.Https,
				WatchFiles: []string{
					"packages",
				},
			},
		},
	}
}

func (h *OpenDesignResourcesPlugins) Load(files map[string][]string) error {
	if files, ok := files[OpenDesignResources]; ok {
		if len(files) > 0 {
			for _, f := range files {
				fileInfo, err := os.Lstat(f)
				if err != nil {
					fmt.Println(fmt.Sprintf("failed to get file %s in plugin.", err))
					continue
				}
				if fileInfo.Name() == "packages" {
					h.Group.Static("packages", f)
				} else {
					return errors.New(fmt.Sprintf("unrecognized file %s", fileInfo.Name()))
				}
			}
		}
	}
	return nil
}

func (h *OpenDesignResourcesPlugins) RegisterEndpoints(group *gin.RouterGroup) {
	h.Group = group
}