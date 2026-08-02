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

func createAuthModel(account, tenantId, region, country string, deviceKey *jwk.Key) model.AuthModel {
	authModel := model.AuthModel{
		TenantId:   tenantId,
		Account:    account,
		Country:    country,
		Region:     region,
		Device_Key: deviceKey,
	}
	return authModel
}

func authModelFunc(c *gin.Context, deviceKey *jwk.Key) {
	account := c.Param("account")
	tenantId := c.Param("tenantId")
	region := c.Param("region")
	country := c.Param("country")

	authModel := createAuthModel(account, tenantId, region, country, deviceKey)

	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": AccountIdMissing})
		c.Abort()
		return
	}

	if tenantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": TenantIdMissing})
		c.Abort()
		return
	}

	if region == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": RegionMissing})
		c.Abort()
		return
	}

	if country == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": CountryMissing})
		c.Abort()
		return
	}

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), model.AuthModelKey, authModel))
	c.Next()
}
