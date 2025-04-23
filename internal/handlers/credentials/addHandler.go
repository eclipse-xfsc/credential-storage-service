package handlers

import (
	"errors"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/eclipse-xfsc/credential-storage-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
)

func add(c *gin.Context, env *common.Environment, presentation bool) {
	contentType := c.GetHeader("Content-Type")
	if contentType != env.GetContentType() {
		_ = handlers.ErrorResponse(c, handlers.WrongContentType, errors.New(""))
		return
	}
	ctx := c.Request.Context()

	id := c.Param("id")

	authModel := ctx.Value(model.AuthModelKey).(model.AuthModel)

	body, err := handlers.ExtractBody(c.Request)
	if err == nil {
		if contentType == common.EncryptedContentType {
			msg, err := jwe.Parse(body)
			if err == nil {
				if msg.ProtectedHeaders().Algorithm() == jwa.ECDH_ES_A256KW {
					if msg.ProtectedHeaders().ContentEncryption() == jwa.ContentEncryptionAlgorithm(jwa.A256GCM) {
						recipients := msg.Recipients()
						if len(recipients) == 1 {
							receipt, err := services.StoreMessage(ctx, id, body, authModel, env, presentation)
							if err == nil && receipt != nil {
								c.Header("Content-Type", common.EncryptedContentType)
								c.String(200, receipt.Receipt)
								return
							} else {
								_ = handlers.ErrorResponse(c, handlers.StoreMessageFailed, err)
								return
							}
						} else {
							_ = handlers.ErrorResponse(c, handlers.InvalidAmountOfRecipients, err)
							return
						}
					} else {
						_ = handlers.ErrorResponse(c, handlers.InvalidContentEncryptionAlgorithm, err)
						return
					}
				} else {
					_ = handlers.ErrorResponse(c, handlers.InvalidKeyEncryptionAlgorithm, err)
					return
				}
			} else {
				_ = handlers.ErrorResponse(c, handlers.BodyParseError, err)
				return
			}
		}

		if contentType == common.NormalContentType {
			if err == nil {
				_, err := services.StoreMessage(ctx, id, body, authModel, env, presentation)
				if err != nil {
					_ = handlers.ErrorResponse(c, handlers.StoreMessageFailed, err)
					return
				}
				return
			} else {
				_ = handlers.ErrorResponse(c, handlers.BodyParseError, err)
				return
			}
		}
	}

	_ = handlers.ErrorResponse(c, handlers.NoBodyError, err)
}

// AddPresentation  godoc
// @Summary Add a presentation to the storage
// @Description Add a presentation to the storage
// @Tags presentations
// @Accept  application/json
// @Param Content-Type header string true "application/json"
// @Param data body string true "The VerifiablePresentation raw data to upload"
// @Param account path string true "Account ID"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "ID of the presentation"
// @Success 200 {string} string "Receipt"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /presentations/{id} [put]
func AddPresentation(c *gin.Context, env *common.Environment) {
	add(c, env, true)
}

// AddCredential  godoc
// @Summary Add a credential to the storage
// @Description Add a credential to the storage
// @Tags credentials
// @Accept  application/json
// @Param Content-Type header string true "application/json"
// @Param data body string true "The VerifiableCredential raw data to upload"
// @Param account path string true "Account ID"
// @Param tenantId path string true "Tenant ID"
// @Param id path string true "ID of the credential"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /credentials/{id} [put]
func AddCredential(c *gin.Context, env *common.Environment) {
	add(c, env, false)
}
