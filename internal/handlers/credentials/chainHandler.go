package handlers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	"github.com/gin-gonic/gin"
)

const NoValidJsonBody = "No valid Json Body."
const NoValidChain = "No chain building possible"
const CredentialNotFound = "Credential not found"

func Chain(c *gin.Context, env *common.Environment) {
	contentType := c.GetHeader("Content-Type")
	if contentType != env.GetContentType() {
		_ = handlers.ErrorResponse(c, handlers.WrongContentType, errors.New(""))
		return
	}
	ctx := c.Request.Context()

	authModel := ctx.Value(model.AuthModelKey).(model.AuthModel)

	body, err := handlers.ExtractBody(c.Request)
	if err == nil {
		if contentType == common.NormalContentType {
			var chainStatement model.ChainStatement
			err = json.Unmarshal(body, &chainStatement)
			if err == nil {
				g, err := ChainCredentials(ctx, chainStatement, authModel, env)

				if err == nil {
					c.JSON(200, g)
					return
				} else {
					handlers.ErrorResponse(c, NoValidChain, err)
					return
				}
			} else {
				handlers.ErrorResponse(c, NoValidJsonBody, err)
				return
			}

		} else {
			handlers.ErrorResponse(c, handlers.WrongContentType, err)
			return
		}
	}

	handlers.ErrorResponse(c, handlers.NoBodyError, err)
}

func ChainCredentials(ctx context.Context, chainStatement model.ChainStatement, authModel model.AuthModel, env *common.Environment) (map[string]interface{}, error) {

	if chainStatement.Root == "" || chainStatement.ChainName == "" {
		return nil, errors.New(handlers.InvalidRequest)
	}

	credentials, err := getCredentials(ctx, authModel, env, nil, false)

	if err != nil {
		return nil, err
	}

	c, ok := credentials.Credentials[chainStatement.Root]

	if !ok {
		return nil, errors.New(CredentialNotFound)
	}

	return ChainRecurse(credentials, chainStatement.Chain, chainStatement.ChainName, c.(map[string]interface{}))
}

func ChainRecurse(credentials *model.GetCredentialModel, chain []model.ChainItem, chainName string, root map[string]interface{}) (map[string]interface{}, error) {

	if chain != nil {
		tempChain := make([]map[string]interface{}, 0)

		for _, c := range chain {

			if c.Item == "" {
				return nil, errors.New(NoValidChain)
			}

			cred, ok := credentials.Credentials[c.Item]

			if !ok {
				return nil, errors.New(CredentialNotFound)
			}

			currentMap := cred.(map[string]interface{})
			tempChain = append(tempChain, currentMap)

			if c.Chain != nil {
				ChainRecurse(credentials, c.Chain, chainName, currentMap)
			}
		}

		root[chainName] = tempChain
	}

	return root, nil
}
