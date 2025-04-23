package handlers

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
)

func CreateTransactionReciept(ctx context.Context, authModel model.AuthModel, env *common.Environment) *model.Receipt {
	log := env.GetLogger()
	nonce, err := crypto.GenerateNonce(env.GetCryptoNamespace(), common.StorageCryptoContext, ctx)

	if err == nil {
		expire := time.Now().Add(time.Minute * 5)
		transaction := model.TransactionModel{
			Nonce:  b64.StdEncoding.EncodeToString([]byte(nonce)),
			Expire: expire.Unix(),
		}

		session := env.GetSession()

		if session == nil {
			return nil
		}

		queryString := fmt.Sprintf(`UPDATE %s.credentials USING TTL %s SET nonce=? WHERE accountPartition=? AND 
																					region=? AND 
																					country=? AND
																					account=?;`,
			authModel.TenantId,
			strconv.Itoa(5*60))

		err := session.Query(queryString,
			b64.StdEncoding.EncodeToString([]byte(nonce)),
			env.GetAccountPartition(authModel.Account),
			env.GetRegion(),
			env.GetCountry(),
			authModel.Account).WithContext(ctx).Exec()

		if err == nil {

			if authModel.Device_Key == nil {
				log.Error(err, "Device Key not present.")
				return nil
			}

			msg, err := crypto.CreateJweMessage(transaction, *authModel.Device_Key)

			if err == nil {
				return new(model.Receipt).CreateReceipt(msg)
			}
		}
	}

	if err != nil {
		log.Error(err, "")
	}

	return nil
}
