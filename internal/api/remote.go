package api

import (
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/remote"

	"github.com/gin-gonic/gin"
)

func AddRegistrationRoutes(g *gin.RouterGroup, env *common.Environment) {
	g.GET("/register", func(c *gin.Context) {
		handlers.AddDevice(c, env)
	})

}

func AddRecoverRoutes(g *gin.RouterGroup, env *common.Environment) {
	g.PATCH("/recover", func(c *gin.Context) {
		handlers.RecoverDevice(c, env)
	})
}

func AddRemoteRoutes(g *gin.RouterGroup, env *common.Environment) {

	g.GET("/session", func(c *gin.Context) {
		handlers.CreateSession(c, env)
	})

	g.DELETE("/delete", func(c *gin.Context) {
		handlers.DeleteDevice(c, env)
	})
}
