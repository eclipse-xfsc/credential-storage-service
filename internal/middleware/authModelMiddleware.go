package middleware

import (
	"context"
	"net/http"

	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func AuthModel() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authModelFunc(ctx, nil)
	}
}

func AuthTestModel(deviceKey *jwk.Key) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authModelFunc(ctx, deviceKey)
	}
}

func createAuthModel(account string, tenantId string, deviceKey *jwk.Key) model.AuthModel {
	authModel := model.AuthModel{
		TenantId:   tenantId,
		Account:    account,
		Device_Key: deviceKey,
	}
	return authModel
}

func authModelFunc(c *gin.Context, deviceKey *jwk.Key) {
	account := c.Param("account")
	tenantId := c.Param("tenantId")
	authModel := createAuthModel(account, tenantId, deviceKey)

	if tenantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": TenantIdMissing})
		c.Abort()
		return
	}

	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": AccountIdMissing})
		c.Abort()
		return
	}

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), model.AuthModelKey, authModel))
	c.Next()
}
