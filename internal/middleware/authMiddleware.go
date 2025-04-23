package middleware

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	AccountLockedError     = "Account is locked."
	TokenInvalid           = "Token Invalid."
	BearerMissing          = "Bearer missing."
	TenantIdMissing        = "Tenant Id Missing."
	AccountIdMissing       = "Account Id Missing."
	NonceNotValid          = "Nonce not valid."
	NonceNotPresent        = "Nonce not present."
	RouteDataInvald        = "Route data invalid."
	KeyManipulationError   = "Device Key was manipulated for Account:"
	NoValidSignatureFormat = "Signature not in valid format."
	NoValidKeyFormat       = "Device Key not in valid format."
)

func Auth(env *common.Environment, recovery bool, nonceRequired bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authFunc(env, ctx, recovery, nonceRequired)
	}
}

func authFunc(env *common.Environment, c *gin.Context, recovery bool, nonceRequired bool) {
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
				return dbCheckUp(env, &authModel, context, sink, sig, message)
			}))))

	if err != nil {
		logger.Error(err, "")
		c.JSON(http.StatusForbidden, gin.H{"message": TokenInvalid})
		c.Abort()
		return
	}

	if nonceRequired {
		field, exist := token.Get("nonce")

		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"message": NonceNotPresent})
			c.Abort()
			return
		}

		if recovery {
			if exist && (field.(string) != authModel.Recovery_Nonce || authModel.Recovery_Nonce == "") {
				c.JSON(http.StatusBadRequest, gin.H{"message": NonceNotValid})
				c.Abort()
				return
			}
		} else {
			if exist && (field.(string) != authModel.Nonce || authModel.Nonce == "") {
				c.JSON(http.StatusBadRequest, gin.H{"message": NonceNotValid})
				c.Abort()
				return
			}
		}
	}

	if err != nil && err.Error() == AccountLockedError {
		c.JSON(http.StatusForbidden, gin.H{"message": AccountLockedError})
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

func dbCheckUp(env *common.Environment, authModel *model.AuthModel, context context.Context, sink jws.KeySink, sig *jws.Signature, message *jws.Message) error {

	var device_key = ""
	var locked bool
	var nonce = ""
	var signature = ""
	var recovery_nonce = ""
	session := env.GetSession()
	logger := env.GetLogger()

	queryString := fmt.Sprintf(`SELECT device_key,locked, nonce, signature, recovery_nonce FROM %s.credentials WHERE accountPartition=? AND 
																						     region=? AND 
																						     country=? AND 
																						     account=? LIMIT 1;`, authModel.TenantId)

	query := session.Query(queryString,
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account)

	err := query.Consistency(gocql.LocalQuorum).Scan(&device_key, &locked, &nonce, &signature, &recovery_nonce)

	if err == nil {
		if locked {
			return errors.New(AccountLockedError)
		}

		sig, err := b64.StdEncoding.DecodeString(signature)

		if err != nil {
			return errors.New(NoValidSignatureFormat)
		}

		jwkJson, err := b64.StdEncoding.DecodeString(device_key)

		if err != nil {
			return errors.New(NoValidSignatureFormat)
		}

		b, err := env.GetCryptoProvider().Verify(types.CryptoIdentifier{
			KeyId: env.GetCryptoSignKey(),
			CryptoContext: types.CryptoContext{
				Namespace: env.GetCryptoNamespace(),
				Context:   context,
				Group:     common.StorageCryptoContext,
			},
		}, jwkJson, sig)

		if !b {
			logger.Debug("", "Error: ", err)
			return errors.New(KeyManipulationError + authModel.Account)
		}

		if err == nil {
			key, err := jwk.ParseKey([]byte(jwkJson))

			if err == nil {
				if key.KeyType() == jwa.EC {
					sink.Key(jwa.SignatureAlgorithm(jwa.ES256), key)
				} else {
					if key.KeyType() == jwa.RSA {
						sink.Key(jwa.SignatureAlgorithm(jwa.PS256), key)
					} else {
						return errors.New(handlers.InvalidKeySigningAlgorithm)
					}
				}
			}

			authModel.Device_Key = &key
			authModel.Nonce = nonce
			authModel.Recovery_Nonce = recovery_nonce

			if err == nil {
				return nil
			}
			return err
		}
	}

	return err
}
