package handlers

import (
	"errors"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, err string, exception error) error {
	env := common.GetEnvironment()
	log := env.GetLogger()
	log.Error(nil, err)
	if exception != nil {
		log.Error(exception, "Detailed Error: ")
	}
	c.JSON(400, gin.H{
		"message": err,
	})
	return errors.New(err)
}

func InternalErrorResponse(c *gin.Context, err string, exception error) error {
	env := common.GetEnvironment()
	log := env.GetLogger()
	log.Error(nil, err)
	if exception != nil {
		log.Error(exception, "Detailed Error: ")
	}

	c.JSON(500, gin.H{
		"message": err,
	})
	return errors.New(err)
}
