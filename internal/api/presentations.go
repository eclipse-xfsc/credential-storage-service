package api

import (
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/credentials"

	"github.com/gin-gonic/gin"
)

func AddPresentationRoutes(g *gin.RouterGroup, env *common.Environment) {
	g.PUT("/:id", func(c *gin.Context) {
		handlers.AddPresentation(c, env)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		handlers.Remove(c, env, true)
	})

	if env.GetContentType() == common.EncryptedContentType {
		g.GET("", func(c *gin.Context) {
			handlers.Get(c, env, true)
		})
	}

	if env.GetContentType() == common.NormalContentType {
		handlers.GetPresentations(g, env)
	}
}
