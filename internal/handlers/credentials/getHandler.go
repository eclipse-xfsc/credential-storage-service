package handlers

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/connection"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	oid4vip "github.com/eclipse-xfsc/oid4-vci-vp-library/model/presentation"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

const getError = "Error during record get."
const badContentTypeError = "Bad content type"
const badBodyError = "Body error"

func Get(c *gin.Context, env *common.Environment, presentation bool) any {
	ctx := c.Request.Context()
	authModel := c.Request.Context().Value(model.AuthModelKey).(model.AuthModel)

	if c.ContentType() != common.EncryptedContentType {
		handlers.ErrorResponse(c, badContentTypeError, errors.New(badContentTypeError))
		return nil
	}

	model, err := getCredentials(ctx, authModel, env, nil, presentation)

	if err != nil {
		handlers.ErrorResponse(c, getError, err)
		return nil
	}

	c.JSON(200, model)

	return nil
}

func post(c *gin.Context, env *common.Environment, presentation bool) any {
	ctx := c.Request.Context()
	authModel := c.Request.Context().Value(model.AuthModelKey).(model.AuthModel)

	if c.ContentType() != common.NormalContentType {
		handlers.ErrorResponse(c, badContentTypeError, errors.New(badContentTypeError))
		return nil
	}

	body, err := handlers.ExtractBody(c.Request)

	if err != nil {
		handlers.ErrorResponse(c, badBodyError, err)
		return nil
	}

	var payload oid4vip.PresentationDefinition

	if len(body) > 0 {
		err = json.Unmarshal(body, &payload)

		if err != nil {
			handlers.ErrorResponse(c, badBodyError, err)
			return nil
		}
	}

	model, err := getCredentials(ctx, authModel, env, &payload, presentation)

	if err != nil {
		handlers.ErrorResponse(c, getError, err)
		return nil
	}

	c.JSON(200, model)

	return nil
}

func getCredentials(ctx context.Context, authModel model.AuthModel, env *common.Environment, filter *oid4vip.PresentationDefinition, presentation bool) (*model.GetCredentialModel, error) {
	logger := env.GetLogger()
	session := env.GetSession()

	if env.GetContentType() == common.EncryptedContentType {
		receipt := handlers.CreateTransactionReciept(ctx, authModel, env)

		if receipt != nil {
			credentials, err := loadCredentials(authModel, *env, session, presentation)

			if err != nil {
				return nil, err
			}

			model := model.GetCredentialModel{
				Credentials: make(map[string]interface{}),
				Receipt:     receipt.Receipt,
			}

			if credentials == nil {
				return &model, nil
			}

			for k, v := range credentials {
				cipher, err := b64.RawStdEncoding.DecodeString(v)
				if err == nil {
					msg, err := crypto.DecryptMessage(authModel.Account, cipher, env.GetCryptoNamespace(), common.StorageCryptoContext, ctx, env.GetCryptoProvider())

					if err != nil {
						logger.Error(err, "")
						continue
					}

					if err == nil {
						model.Credentials[k] = string(msg)
					} else {
						logger.Error(err, "")
					}
				}

			}
			return &model, nil
		}

		return nil, errors.New("Receipt Failure")
	}

	if env.GetContentType() == common.NormalContentType {
		credentials, err := loadCredentials(authModel, *env, session, presentation)

		if err != nil {
			return nil, err
		}
		model := model.GetCredentialModel{
			Credentials: make(map[string]interface{}),
		}

		if credentials == nil {
			return &model, nil
		}

		foundCredentials := make(map[string]interface{}, 0)
		for k, v := range credentials {
			cipher, err := b64.RawStdEncoding.DecodeString(v)
			if err == nil {
				msg, err := crypto.DecryptMessage(authModel.Account, cipher, env.GetCryptoNamespace(), common.StorageCryptoContext, ctx, env.GetCryptoProvider())

				if err != nil {
					logger.Error(err, "")
					continue
				}

				foundCredentials[k] = string(msg)
			}
		}

		logger.Info("Found credentials before filter", "amount", len(foundCredentials))

		res, err := filter.Filter(foundCredentials)

		if err != nil {
			return nil, err
		}

		model.Groups = res

		return &model, nil
	}

	return nil, errors.New("Error Getting Credentials.")
}

func loadCredentials(authModel model.AuthModel, env common.Environment, session connection.SessionInterface, presentation bool) (map[string]string, error) {

	object := "credentials"

	if presentation {
		object = "presentations"
	}

	var objects map[string]string
	queryString := fmt.Sprintf(`SELECT %s FROM %s.credentials WHERE accountPartition=? AND 
																					region=? AND 
																					country=? AND 
																					account=? AND 
																					locked=False;`, object, authModel.TenantId)
	err := session.Query(queryString,
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account).Consistency(gocql.LocalQuorum).Scan(&objects)

	if err != nil && errors.Is(gocql.ErrNotFound, err) {
		return make(map[string]string), nil
	} else if err != nil {
		return nil, errors.Join(errors.New("db query error"), err)
	} else {
		return objects, err
	}
}

// GetPresentations godoc
// @Summary Add a presentation to the storage
// @Description Add a presentation to the storage
// @Tags presentations
// @Produce json
// @Accept  json
// @Param request body oid4vip.PresentationDefinition false "Presentation definition details"
// @Param account path string true "Account ID"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} model.GetCredentialModel "presentations"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /presentations [post]
func GetPresentations(g *gin.RouterGroup, env *common.Environment) gin.IRoutes {
	return g.POST("", func(c *gin.Context) {
		post(c, env, true)
	})
}

// GetCredentials godoc
// @Summary Get credentials from the storage
// @Description Get credentials from the storage
// @Tags credentials
// @Produce json
// @Accept  json
// @Param request body  oid4vip.PresentationDefinition false "Presentation definition details"
// @Param account path string true "Account ID"
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} model.GetCredentialModel "credentials"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /credentials [post]
func GetCredentials(g *gin.RouterGroup, env *common.Environment) gin.IRoutes {
	return g.POST("", func(c *gin.Context) {
		post(c, env, false)
	})
}
