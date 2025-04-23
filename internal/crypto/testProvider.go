package crypto

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	b64 "encoding/base64"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
)

type TestProvider struct {
}

var aesKeys = make(map[string][]byte, 0)
var rsaKeys = make(map[string]rsa.PrivateKey, 0)
var ecDsaKeys = make(map[string]ecdsa.PrivateKey, 0)

func (l *TestProvider) AddKey(name string, key interface{}) {
	switch v := key.(type) {
	case *ecdsa.PrivateKey:
		ecDsaKeys[name] = *v
	case *rsa.PrivateKey:
		rsaKeys[name] = *v
	}
}

func (l *TestProvider) GetNamespaces(context types.CryptoContext) ([]string, error) {
	return []string{"transit"}, nil
}

func (l *TestProvider) CreateCryptoContext(context types.CryptoContext) error {
	return nil
}

func (l *TestProvider) DestroyCryptoContext(context types.CryptoContext) error {
	return nil
}

func (l *TestProvider) GetKey(parameter types.CryptoIdentifier) (*types.CryptoKey, error) {
	key, ok := rsaKeys[parameter.KeyId]
	if ok {
		pubkey_bytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err == nil {
			return &types.CryptoKey{
				Key:     pubkey_bytes,
				Version: "1",
				CryptoKeyParameter: types.CryptoKeyParameter{
					KeyType:    types.Rsa4096,
					Identifier: parameter,
				},
			}, nil
		}
	}

	key2, ok2 := ecDsaKeys[parameter.KeyId]
	if ok2 {
		pubkey_bytes, err := x509.MarshalPKIXPublicKey(&key2.PublicKey)
		if err == nil {
			return &types.CryptoKey{
				Key:     pubkey_bytes,
				Version: "1",
				CryptoKeyParameter: types.CryptoKeyParameter{
					KeyType:    types.Ecdsap256,
					Identifier: parameter,
				},
			}, nil
		}
	}

	key3, ok3 := aesKeys[parameter.KeyId]

	if ok3 {
		return &types.CryptoKey{
			Key:     key3,
			Version: "1",
			CryptoKeyParameter: types.CryptoKeyParameter{
				KeyType:    types.Aes256GCM,
				Identifier: parameter,
			},
		}, nil
	}

	return nil, nil
}

func (l *TestProvider) IsCryptoContextExisting(context types.CryptoContext) (bool, error) {
	return true, nil
}

func (l *TestProvider) RotateKey(parameter types.CryptoIdentifier) error {
	return nil
}

func (l *TestProvider) GetKeys(parameter types.CryptoFilter) (*types.CryptoKeySet, error) {
	set := new(types.CryptoKeySet)
	set.Keys = make([]types.CryptoKey, 0)

	for i := range aesKeys {
		if parameter.Filter.MatchString(i) {
			identifier := types.CryptoIdentifier{
				CryptoContext: parameter.CryptoContext,
				KeyId:         i,
			}
			key, err := l.GetKey(identifier)

			if err != nil {
				return nil, err
			}

			set.Keys = append(set.Keys, *key)
		}
	}

	for i := range rsaKeys {
		if parameter.Filter.MatchString(i) {
			identifier := types.CryptoIdentifier{
				CryptoContext: parameter.CryptoContext,
				KeyId:         i,
			}
			key, err := l.GetKey(identifier)

			if err != nil {
				return nil, err
			}

			set.Keys = append(set.Keys, *key)
		}
	}

	for i := range ecDsaKeys {
		if parameter.Filter.MatchString(i) {
			identifier := types.CryptoIdentifier{
				CryptoContext: parameter.CryptoContext,
				KeyId:         i,
			}
			key, err := l.GetKey(identifier)

			if err != nil {
				return nil, err
			}

			set.Keys = append(set.Keys, *key)
		}
	}

	return set, nil
}

func (l *TestProvider) DeleteKey(parameter types.CryptoIdentifier) error {
	delete(rsaKeys, parameter.KeyId)
	delete(ecDsaKeys, parameter.KeyId)
	delete(aesKeys, parameter.KeyId)
	return nil
}

func (l *TestProvider) Hash(parameter types.CryptoHashParameter, msg []byte) (b []byte, err error) {
	if parameter.HashAlgorithm == types.Sha2256 {
		msgHash := sha256.New()
		_, err = msgHash.Write(msg)
		if err != nil {
			return nil, err
		}
		msgHashSum := msgHash.Sum(nil)
		return msgHashSum, nil
	} else {
		return nil, errors.ErrUnsupported
	}
}

func (l *TestProvider) GenerateRandom(context types.CryptoContext, number int) ([]byte, error) {
	key := make([]byte, number)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (l *TestProvider) Encrypt(parameter types.CryptoIdentifier, data []byte) ([]byte, error) {

	key, ok := rsaKeys[parameter.KeyId]

	if ok {
		hash := sha256.New()
		ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &key.PublicKey, data, nil)
		if err != nil {
			return nil, err
		}
		return ciphertext, err
	} else {
		key, ok := aesKeys[parameter.KeyId]

		if ok {
			c, err := aes.NewCipher(key)

			if err != nil {
				return nil, err
			}

			gcm, err := cipher.NewGCM(c)

			if err != nil {
				return nil, err
			}

			nonce, err := l.GenerateRandom(parameter.CryptoContext, gcm.NonceSize())

			if err != nil {
				return nil, err
			}

			return gcm.Seal(nonce, nonce, data, nil), nil
		}
	}

	return nil, errors.New("no key found")
}

func (l *TestProvider) Decrypt(parameter types.CryptoIdentifier, data []byte) ([]byte, error) {
	key, ok := rsaKeys[parameter.KeyId]

	if ok {
		hash := sha256.New()
		ciphertext, err := rsa.DecryptOAEP(hash, rand.Reader, &key, data, nil)
		if err != nil {
			return nil, err
		}
		return ciphertext, err
	} else {
		key, ok := aesKeys[parameter.KeyId]

		if ok {
			c, err := aes.NewCipher(key)

			if err != nil {
				return nil, err
			}

			gcm, err := cipher.NewGCM(c)

			if err != nil {
				return nil, err
			}

			nonceSize := gcm.NonceSize()

			if len(data) < nonceSize {
				return nil, errors.New("nonce size not valid")
			}

			nonce, ciphertext := data[:nonceSize], data[nonceSize:]
			plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

			if err != nil {
				return nil, err
			}

			return plaintext, nil
		}
	}
	return nil, errors.New("no key found")
}

func (l *TestProvider) Sign(parameter types.CryptoIdentifier, data []byte) (b []byte, err error) {

	key, ok := rsaKeys[parameter.KeyId]

	if ok {
		hashed := sha256.Sum256(data)
		signature, err := rsa.SignPSS(rand.Reader, &key, crypto.SHA256, hashed[:], nil)
		if err != nil {
			return nil, err
		}

		return signature, nil
	}

	key2, ok := ecDsaKeys[parameter.KeyId]

	if ok {
		hashed := sha256.Sum256(data)
		signature, err := ecdsa.SignASN1(rand.Reader, &key2, hashed[:])
		if err != nil {
			return nil, err
		}

		return signature, nil
	}

	return nil, errors.ErrUnsupported
}

func (l *TestProvider) Verify(parameter types.CryptoIdentifier, data []byte, signature []byte) (b bool, err error) {

	key, ok := rsaKeys[parameter.KeyId]
	if ok {
		hashed := sha256.Sum256(data)
		err = rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, hashed[:], signature, nil)
		if err == nil {
			return true, nil
		}
	}

	key2, ok := ecDsaKeys[parameter.KeyId]
	if ok {
		hashed := sha256.Sum256(data)
		result := ecdsa.VerifyASN1(&key2.PublicKey, hashed[:], signature)
		return result, nil
	}

	fmt.Println("could not verify signature: ", err)
	return false, errors.ErrUnsupported
}

func (l *TestProvider) IsKeyExisting(identifer types.CryptoIdentifier) (bool, error) {
	_, ok := rsaKeys[identifer.KeyId]

	if ok {
		return true, nil
	}

	_, ok = ecDsaKeys[identifer.KeyId]

	if ok {
		return true, nil
	}

	_, ok = aesKeys[identifer.KeyId]

	return ok, nil
}

func (l *TestProvider) GetPublicKeyPem(context string, keyId string) (p *pem.Block, err error) {

	key, ok := rsaKeys[keyId]

	if ok {
		pub := key.Public()
		pubPEM := &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		}
		return pubPEM, err
	} else {
		key, ok := ecDsaKeys[keyId]

		if ok {
			pub := key.Public()
			pubPEM := &pem.Block{
				Type:  "ECDSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
			}
			return pubPEM, err
		}
	}
	return nil, errors.New("no Key found for keyID+context")
}

func (l *TestProvider) GenerateKey(parameter types.CryptoKeyParameter) error {

	if parameter.KeyType == types.Rsa4096 {

		_, ok := rsaKeys[parameter.Identifier.KeyId]

		if !ok {
			keyNew, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				return err
			}
			rsaKeys[parameter.Identifier.KeyId] = *keyNew
		}
		return nil
	}

	if parameter.KeyType == types.Ecdsap256 {

		_, ok := ecDsaKeys[parameter.Identifier.KeyId]

		if !ok {
			keyNew, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				return err
			}
			ecDsaKeys[parameter.Identifier.KeyId] = *keyNew
		}

		return nil
	}

	if parameter.KeyType == types.Aes256GCM {
		_, ok := aesKeys[parameter.Identifier.KeyId]

		if !ok {
			_, err := l.GenerateRandom(parameter.Identifier.CryptoContext, 32)
			if err != nil {
				return err
			}
			aesKeys[parameter.Identifier.KeyId] = []byte("12345678901234567890123456789012")
			return nil
		}
	}

	return errors.ErrUnsupported
}

func (l *TestProvider) GetSeed(context context.Context) string {
	ctx := types.CryptoContext{
		Namespace: "random",
		Context:   context,
	}
	b, _ := l.GenerateRandom(ctx, 32)
	return b64.StdEncoding.EncodeToString(b)
}

func (l *TestProvider) GetSupportedHashAlgs() []types.HashAlgorithm {
	return make([]types.HashAlgorithm, 0)
}

func (l *TestProvider) GetSupportedKeysAlgs() []types.KeyType {
	return make([]types.KeyType, 0)
}
