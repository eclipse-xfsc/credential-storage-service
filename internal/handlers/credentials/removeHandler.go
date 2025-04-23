package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

const deleteCredentialError = "Credential couldnt be deleted."

func Remove(c *gin.Context, env *common.Environment, presentation bool) any {
	id := c.Param("id")
	ctx := c.Request.Context()
	authModel := c.Request.Context().Value(model.AuthModelKey).(model.AuthModel)

	if env.GetContentType() == common.EncryptedContentType {
		receipt := handlers.CreateTransactionReciept(ctx, authModel, env)
		if receipt != nil {
			err := removeCredential(ctx, id, authModel, env, presentation)

			if err == nil {
				c.Header("Content-Type", common.EncryptedContentType)
				c.String(200, receipt.Receipt)
				return nil
			} else {
				handlers.ErrorResponse(c, deleteCredentialError, err)
			}
		}
		return handlers.ErrorResponse(c, "Receipt Error", errors.New("receipt error"))
	}

	if env.GetContentType() == common.NormalContentType {
		err := removeCredential(ctx, id, authModel, env, presentation)

		if err == nil {
			return nil
		} else {
			handlers.ErrorResponse(c, deleteCredentialError, err)
			return err
		}

	}

	handlers.ErrorResponse(c, deleteCredentialError, errors.ErrUnsupported)
	return errors.ErrUnsupported
}

func removeCredential(ctx context.Context, id string, authModel model.AuthModel, env *common.Environment, presentation bool) error {

	object := "credentials"

	if presentation {
		object = "presentations"
	}

	session := env.GetSession()

	queryString := fmt.Sprintf(`DELETE %s[?] FROM %s.credentials  WHERE accountPartition=? AND
																						   region=? AND 
																						   country=? AND 
																						   account=?;`, object, authModel.TenantId)
	err := session.Query(queryString,
		id,
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account).Consistency(gocql.LocalQuorum).WithContext(ctx).Exec()

	if err != nil {
		return errors.New(deleteCredentialError)
	}

	return nil
}
