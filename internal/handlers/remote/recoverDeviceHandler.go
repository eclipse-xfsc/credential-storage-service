package handlers

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	ljwt "github.com/eclipse-xfsc/ssi-jwt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func RecoverDevice(c *gin.Context, env *common.Environment) {

	if c.GetHeader("Content-Type") != "application/jwt" {
		_ = handlers.ErrorResponse(c, handlers.WrongContentType, errors.New(""))
		return
	}

	ctx := c.Request.Context()

	authModel := ctx.Value(model.AuthModelKey).(model.AuthModel)

	body, err := handlers.ExtractBody(c.Request)

	if err == nil {
		jwt, err := jwt.Parse(body,
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
		if err == nil {
			if jwt.Subject() == authModel.Account {
				msg, err := jws.Parse(body)
				if err == nil {
					newJwk := msg.Signatures()[0].ProtectedHeaders().JWK()
					receipt, err := updateRecord(ctx, authModel, &newJwk, env)
					if err == nil && receipt != nil {
						c.Header("Content-Type", "application/jose")
						c.String(200, receipt.Receipt)
						return
					}
				}
			}
		}
	}

	_ = handlers.ErrorResponse(c, handlers.InvalidRequest, err)
}

func updateRecord(ctx context.Context, authModel model.AuthModel, key *jwk.Key, env *common.Environment) (*model.Receipt, error) {
	logger := env.GetLogger()
	session := env.GetSession()

	sig, nonce, newKey, err := createBasicStructure(ctx, key, env)

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	queryString := fmt.Sprintf(`UPDATE %s.credentials SET device_key=?,
															  signature=?, 
															  recovery_nonce=?,
															  last_update_timestamp=toTimestamp(now()) WHERE 
																  		  accountPartition=? AND 
																					region=? AND 
																					country=? AND
																					account=?;`, authModel.TenantId)

	err2 := session.Query(queryString,
		b64.StdEncoding.EncodeToString(newKey),
		b64.StdEncoding.EncodeToString(sig),
		b64.StdEncoding.EncodeToString(nonce),
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account).WithContext(ctx).Exec()

	if err2 != nil {
		logger.Error(err2, "")
		return nil, err2
	}

	receipt := model.RegistrationModel{
		Recovery_Nonce: b64.StdEncoding.EncodeToString(nonce),
	}

	payload, err := json.Marshal(receipt)

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	return new(model.Receipt).CreateReceipt(ljwt.EncryptJweMessage(payload, jwa.ECDH_ES_A256KW, *key)), nil
}
