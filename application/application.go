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

package application

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"github.com/opensourceways/app-community-metadata/app"
	"github.com/opensourceways/app-community-metadata/application/middleware"
)

var server *gin.Engine

func Server() *gin.Engine {
	return server
}

func InitServer(s middleware.SkipRequestLog) {
	server = gin.New()
	//TODO: figure out why
	if app.EnvName == app.EnvDev {
		server.Use(gin.Logger(), gin.Recovery())
	}
	AddRoutes(server)

}

func Run() {
	//NOTE: application will use loopback address 127.0.0.1 for internal usage, please don't remove 127.0.0.1 address
	err := server.Run(fmt.Sprintf("0.0.0.0:%d", app.HttpPort))
	if err != nil {
		color.Error.Println(err)
	}
}
