package services

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/connection"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	"github.com/sirupsen/logrus"
)

func StoreMessage(ctx context.Context,
	id string,
	msg []byte,
	authModel model.AuthModel, env *common.Environment, presentation bool) (*model.Receipt, error) {

	session := env.GetSession()

	if env.GetContentType() == common.EncryptedContentType {
		cipher, err := crypto.EncryptMessage(authModel.Account, env.GetCryptoNamespace(), common.StorageCryptoContext, msg, ctx, env.GetCryptoProvider())
		if err == nil && cipher != nil {

			receipt := handlers.CreateTransactionReciept(ctx, authModel, env)

			if receipt != nil {
				err = executeStoring(id, ctx, cipher, session, authModel, env, presentation)
				return receipt, err
			}
		}

		if err != nil {
			logrus.Error(err.Error())
		}
		return nil, err
	}

	if env.GetContentType() == common.NormalContentType {
		cipher, err := crypto.EncryptMessage(authModel.Account, env.GetCryptoNamespace(), common.StorageCryptoContext, msg, ctx, env.GetCryptoProvider())
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		err = executeStoring(id, ctx, cipher, session, authModel, env, presentation)

		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}

		return nil, err
	}

	return nil, errors.New("content types doesnt fit.")

}

func executeStoring(id string, ctx context.Context, msg []byte, session connection.SessionInterface, authModel model.AuthModel, env *common.Environment, presentation bool) error {

	object := "credentials"

	if presentation {
		object = "presentations"
	}

	queryString := fmt.Sprintf(`UPDATE %s.credentials SET %s[?] = ?, locked=False, last_update_timestamp=toTimestamp(now()) WHERE 
																  		  accountPartition=? AND 
																					region=? AND 
																					country=? AND
																					account=?;`, authModel.TenantId, object)

	return session.Query(queryString,
		id,
		b64.RawStdEncoding.EncodeToString(msg),
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account).WithContext(ctx).Exec()
}
