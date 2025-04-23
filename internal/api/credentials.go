package api

import (
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/credentials"
	"github.com/gin-gonic/gin"
)

func AddCredentialRoutes(g *gin.RouterGroup, env *common.Environment) {
	g.PUT("/:id", func(c *gin.Context) {
		handlers.AddCredential(c, env)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		handlers.Remove(c, env, false)
	})

	if env.GetContentType() == common.EncryptedContentType {
		g.GET("", func(c *gin.Context) {
			handlers.Get(c, env, false)
		})
	}

	if env.GetContentType() == common.NormalContentType {
		handlers.GetCredentials(g, env)

		g.POST("/chain", func(c *gin.Context) {
			handlers.Chain(c, env)
		})
	}
}
