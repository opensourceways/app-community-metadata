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

func InitServer() {
	server = gin.New()
	//TODO: figure out why
	if app.EnvName == app.EnvDev {
		server.Use(gin.Logger(), gin.Recovery())
	}
	server.Use(middleware.RequestLog())

	AddRoutes(server)

}

func Run() {
	err := server.Run(fmt.Sprintf("0.0.0.0:%d", app.HttpPort))
	if err != nil {
		color.Error.Println(err)
	}
}
