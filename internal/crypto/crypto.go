package crypto

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
	"github.com/eclipse-xfsc/ssi-jwt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var cProvider types.CryptoProvider

func GetCryptoProvider() types.CryptoProvider {
	return cProvider
}

/*
	Usage: Creates an Crypto Provider
*/

func CreateCryptoProvider(unitTestMode bool, stdCryptoProvider types.CryptoProvider) {
	if unitTestMode {
		logrus.Info("unitTestMode crypto provider is used")
		cProvider = new(TestProvider)
	} else {
		cProvider = stdCryptoProvider
	}
}

/*
	Usage: Encrypts JWE messages before they are going out to storage.

	Notes: Keys will be generated during account generation in Crypto Provider in cause of performance considerations.
*/

func EncryptMessage(id string, namespace string, group string, msg []byte, ctx context.Context, provider types.CryptoProvider) ([]byte, error) {
	identifier := types.CryptoIdentifier{
		KeyId: id,
		CryptoContext: types.CryptoContext{
			Namespace: namespace,
			Context:   ctx,
			Group:     group,
		},
	}
	var err error
	var exists bool
	if exists, err = provider.IsKeyExisting(identifier); err == nil && !exists {
		err = provider.GenerateKey(types.CryptoKeyParameter{
			Identifier: identifier,
			KeyType:    types.Aes256GCM,
		})
		if err != nil {
			err = errors.Join(errors.New("failed to generate key"), err)
		}
	}
	if err != nil {
		logrus.Error(errors.Join(errors.New("failed to check existence"), err))
		return nil, err
	}

	cipher, err := provider.Encrypt(identifier, msg)
	if err == nil && cipher != nil {
		return cipher, nil
	}
	if err != nil {
		err = errors.Join(errors.New("failed to encrypt data"), err)
		return nil, err
	}

	return nil, err
}

/*
	Usage: Decrypts JWE messages before they are going out to storage.

	Notes: Keys will be generated during account generation in Crypto Provider in cause of performance considerations.
*/

func DecryptMessage(id string, cipher []byte, namespace string, group string, ctx context.Context, provider types.CryptoProvider) ([]byte, error) {

	identifier := types.CryptoIdentifier{
		KeyId: id,
		CryptoContext: types.CryptoContext{
			Namespace: namespace,
			Context:   ctx,
			Group:     group,
		},
	}

	data, err := provider.Decrypt(identifier, cipher)

	if err == nil && data != nil {
		return data, nil
	} else {
		return nil, err
	}
}

func GenerateNonce(namespace string, group string, ctx context.Context) ([]byte, error) {
	return cProvider.GenerateRandom(
		types.CryptoContext{Namespace: namespace, Context: ctx, Group: group}, 32)
}

func CreateJweMessage(payload any, key jwk.Key) (*jwe.Message, error) {
	p, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	switch key.KeyType() {
	case jwa.EC:
		return jwt.EncryptJweMessage(p, jwa.ECDH_ES_A256KW, key), nil
	case jwa.RSA:
		return jwt.EncryptJweMessage(p, jwa.RSA_OAEP_256, key), nil
	}

	return nil, err
}
