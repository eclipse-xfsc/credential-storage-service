package health

import (
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"

	"github.com/gin-gonic/gin"
)

func AddHealth(c *gin.Context, env *common.Environment) {
	if env.GetSession() == nil || env.GetSession().Closed() {
		c.JSON(400, gin.H{
			"status": "unhealthy",
		})
	} else {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	}
}
