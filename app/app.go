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
	BaseConfigFile  = "app.toml"
	DefaultHttpPort = 9500
	DefaultAppName  = "community-metadata"
	DefaultInterval = 60
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
