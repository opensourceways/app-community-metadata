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

package app

type AppInfo struct {
	Tag       string `json:"tag" description:"get tag name"`
	Version   string `json:"version" description:"git repo version."`
	ReleaseAt string `json:"releaseAt" description:"latest commit date"`
}

const (
	EnvProd = "prod"
	EnvTest = "test"
	EnvDev  = "dev"
)

const (
	BaseConfigFile         = "app.toml"
	DefaultHttpPort        = 9500
	DefaultAppName         = "community-metadata"
	DefaultInterval        = 60
	DefaultSyncChannelSize = 100
)

var (
	//App name
	Name string
	//Debug mode
	Debug bool
	//Current host name
	Hostname string
	//App port listen to
	HttpPort = DefaultHttpPort
	//Env name
	EnvName = EnvDev
	//App git info
	GitInfo AppInfo
)
