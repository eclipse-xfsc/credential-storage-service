package middleware

import (
	"context"
	"net/http"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func SelfSignedAuth(env *common.Environment) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		selfSignedAuthFunc(env, ctx)
	}
}

func selfSignedAuthFunc(env *common.Environment, c *gin.Context) {
	logger := env.GetLogger()
	authObject := c.Request.Context().Value(model.AuthModelKey)

	if authObject == nil {
		c.JSON(http.StatusForbidden, gin.H{"message": RouteDataInvald})
		c.Abort()
		return
	}
	authModel := authObject.(model.AuthModel)
	token, err := jwt.ParseRequest(c.Request,
		jwt.WithSubject(authModel.Account),
		jwt.WithKeyProvider(jws.KeyProviderFunc(jws.KeyProviderFunc(
			func(context context.Context, sink jws.KeySink, sig *jws.Signature, message *jws.Message) error {
				alg := sig.ProtectedHeaders().Algorithm()

				key := sig.ProtectedHeaders().JWK()

				if key == nil {
					return jwt.ErrInvalidJWT()
				}

				sink.Key(alg, key)
				authModel.Device_Key = &key
				return nil
			}))))

	if err != nil {
		logger.Debug("", "Error: ", err)
		c.JSON(http.StatusForbidden, gin.H{"message": TokenInvalid})
		c.Abort()
		return
	}

	if token == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": BearerMissing})
		c.Abort()
		return
	}
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), model.AuthModelKey, authModel))
	c.Next()
}
