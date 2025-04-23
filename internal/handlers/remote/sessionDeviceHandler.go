package handlers

import (
	"errors"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/gin-gonic/gin"
)

func CreateSession(c *gin.Context, env *common.Environment) {
	ctx := c.Request.Context()

	authModel := ctx.Value(model.AuthModelKey).(model.AuthModel)

	receipt := handlers.CreateTransactionReciept(ctx, authModel, env)

	if receipt != nil {
		c.Header("Content-Type", "application/jose")
		c.String(200, receipt.Receipt)
		return
	} else {
		_ = handlers.ErrorResponse(c, handlers.InvalidRequest, errors.New(""))
		return
	}
}
