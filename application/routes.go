package application

import (
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/app-community-metadata/application/controller"
)

// AddRoutes
func AddRoutes(r *gin.Engine) {
	// status
	r.GET("/health", controller.AppHealth)
	r.GET("/ping", controller.PingPong)

	// metadata
	_ = r.Group("/v1")
	{
		//add metadata API here
	}

	// not found routes
	r.NoRoute(func(c *gin.Context) {
		c.Data(404, "text/plain", []byte("not found"))
	})
}
