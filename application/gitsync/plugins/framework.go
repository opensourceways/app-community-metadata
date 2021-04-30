package plugins

import "github.com/opensourceways/app-community-metadata/application/gitsync"

func init() {
	gitsync.Register("helloworld", NewHelloWorldPlugin())
}
