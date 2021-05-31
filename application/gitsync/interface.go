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
