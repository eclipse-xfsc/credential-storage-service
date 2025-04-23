package api

import (
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/health"

	"github.com/gin-gonic/gin"
)

func AddHealth(g *gin.RouterGroup, env *common.Environment) {
	g.GET("/health", func(c *gin.Context) {
		handlers.AddHealth(c, env)
	})
}
