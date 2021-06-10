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

package gitsync

import "github.com/gin-gonic/gin"

type RepoSchema string

const (
	Https RepoSchema = "https"
	Ssh   RepoSchema = "ssh"
)

type GitEvent struct {
	GroupName string
	RepoName  string
	Files     []string
}

type GitMeta struct {
	//Git repo to watch
	Repo string
	//Git branch
	Branch string
	//Whether to checkout submodules
	SubModules string
	//Git repo schema, https or ssh
	Schema RepoSchema
	//Files to watch, relatively
	WatchFiles []string
}

type GitMetaContainer struct {
	Meta  *GitMeta
	Ready bool
}

type PluginMeta struct {
	//Plugin name used for identity, case insensitive
	Name string
	//Description for this plugin
	Description string
	//API groups, the exposed api endpoint would be: /version/data/routerGroup-name/register-endpoint
	Group string
	//Git repositories to watch
	Repos []GitMeta
}

type Plugin interface {
	GetMeta() *PluginMeta
	Load(files map[string][]string) error
	RegisterEndpoints(group *gin.RouterGroup)
}

type EventFilter interface {
	StartLoop()
}

type Runner interface {
	GetRepo() *GitMeta
	StartLoop()
	Close() error
	RepoUpdated()
}
