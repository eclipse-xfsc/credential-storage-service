package handlers

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	crypt "github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/eclipse-xfsc/crypto-provider-core/types"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func AddDevice(c *gin.Context, env *common.Environment) {
	logger := env.GetLogger()

	if c.GetHeader("Content-Type") != "application/json" {
		handlers.ErrorResponse(c, handlers.WrongContentType, errors.New(""))
		return
	}

	authObject := c.Request.Context().Value(model.AuthModelKey)

	if authObject == nil {
		return
	}

	authModel := authObject.(model.AuthModel)

	key := *authModel.Device_Key

	if key.KeyType() != jwa.EC && key.KeyType() != jwa.RSA {
		handlers.ErrorResponse(c, handlers.InvalidKeySigningAlgorithm, errors.New(""))
		return
	}

	exist, err := checkExist(authModel, env)
	if !exist {
		if err == nil {
			receipt, err := storeRecord(c.Request.Context(), authModel, env)
			if err == nil && receipt != nil {
				c.Header("Content-Type", "application/jose")
				c.String(200, receipt.Receipt)
				return
			} else {
				logger.Debug("", "Error", err)
				handlers.ErrorResponse(c, handlers.DeviceRegistrationFailed, err)
				return
			}
		} else {
			logger.Debug("", "Error", err)
			handlers.ErrorResponse(c, handlers.DeviceRegistrationFailed, err)
		}
	} else {
		logger.Debug("", "Error", err)
		handlers.ErrorResponse(c, handlers.DeviceAlreadyExist, err)
	}

}

func checkExist(authModel model.AuthModel, env *common.Environment) (bool, error) {
	var account = ""
	session := env.GetSession()

	queryString := fmt.Sprintf(`SELECT account FROM %s.credentials WHERE accountPartition=? AND 
																					region=? AND 
																					country=? AND 
																					account=? LIMIT 1;`, authModel.TenantId)

	query := session.Query(queryString,
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account)

	err := query.Consistency(gocql.LocalQuorum).Scan(&account)

	if err == gocql.ErrNotFound {
		err = nil
		account = ""
	}

	return account != "", err
}

func createBasicStructure(ctx context.Context, key *jwk.Key, env *common.Environment) ([]byte, []byte, []byte, error) {
	logger := env.GetLogger()

	nonce, err := env.GetCryptoProvider().GenerateRandom(types.CryptoContext{Namespace: env.GetCryptoNamespace(), Context: ctx, Group: common.StorageCryptoContext}, 64)
	if err != nil {
		return nil, nil, nil, err
	}

	var dk bytes.Buffer
	json.NewEncoder(&dk).Encode(key)

	sign, err := env.GetCryptoProvider().Sign(types.CryptoIdentifier{
		KeyId: env.GetCryptoSignKey(),
		CryptoContext: types.CryptoContext{
			Namespace: env.GetCryptoNamespace(),
			Context:   ctx,
			Group:     common.StorageCryptoContext,
		},
	}, dk.Bytes())

	if err != nil {
		logger.Error(err, "")
		return nil, nil, nil, err
	}

	return sign, nonce, dk.Bytes(), nil
}

func storeRecord(ctx context.Context, authModel model.AuthModel, env *common.Environment) (*model.Receipt, error) {
	session := env.GetSession()
	logger := env.GetLogger()

	sig, nonce, key, err := createBasicStructure(ctx, authModel.Device_Key, env)

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	queryString := fmt.Sprintf(`INSERT INTO %s.credentials 
									( accountPartition, 
									  region, 
									  country,
									  account,
									  last_update_timestamp,
									  recovery_nonce,
									  device_key,
									  signature,
									  locked) VALUES (?, ?, ?, ?, toTimestamp(now()), ?, ? , ? ,False);`, authModel.TenantId)

	err = session.Query(queryString,
		env.GetAccountPartition(authModel.Account),
		env.GetRegion(),
		env.GetCountry(),
		authModel.Account,
		b64.StdEncoding.EncodeToString(nonce),
		b64.StdEncoding.EncodeToString(key),
		b64.StdEncoding.EncodeToString(sig),
	).WithContext(ctx).Exec()

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	receipt := model.RegistrationModel{
		Recovery_Nonce: b64.StdEncoding.EncodeToString(nonce),
	}

	msg, err := crypt.CreateJweMessage(receipt, *authModel.Device_Key)

	if err != nil {
		logger.Error(err, "")
		return nil, err
	}

	return new(model.Receipt).CreateReceipt(msg), nil
}
