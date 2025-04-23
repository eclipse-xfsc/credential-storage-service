package tests

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func GetPrivateKey() ([]byte, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path.Join(dir, "tests", "data", "testPriv.json"))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func CreateTestJWK() (jwk.Key, error) {

	data, _ := GetPrivateKey()
	privkey, err := jwk.ParseKey(data)

	if err != nil {
		return nil, err
	}

	return privkey, err
}

func CreateSelfSignedToken(key jwk.Key, audience string, account string) ([]byte, error) {

	tok, err := jwt.NewBuilder().
		Issuer(`github.com/lestrrat-go/jwx`).
		IssuedAt(time.Now()).
		Audience([]string{audience}).
		Subject(account).
		Expiration(time.Now().Add(time.Hour)).
		Build()
	if err != nil {
		fmt.Printf("failed to build token: %s\n", err)
		return nil, err
	}

	var k interface{}
	key.Raw(&k)
	pub, _ := key.PublicKey()
	headers := jws.NewHeaders()
	headers.Set("jwk", pub)

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256, k, jws.WithProtectedHeaders(headers)))

	if err != nil {
		return nil, err
	}

	return signed, nil
}

func CreateToken(key jwk.Key, audience string, account string, nonce string, addNonce bool) ([]byte, error) {

	tok, err := jwt.NewBuilder().
		Issuer(`github.com/lestrrat-go/jwx`).
		IssuedAt(time.Now()).
		Audience([]string{audience}).
		Subject(account).
		Expiration(time.Now().Add(time.Hour)).
		Claim("nonce", nonce).
		Build()

	if !addNonce {
		tok, err = jwt.NewBuilder().
			Issuer(`github.com/lestrrat-go/jwx`).
			IssuedAt(time.Now()).
			Audience([]string{audience}).
			Subject(account).
			Expiration(time.Now().Add(time.Hour)).
			Build()
	}

	if err != nil {
		fmt.Printf("failed to build token: %s\n", err)
		return nil, err
	}

	var k interface{}
	key.Raw(&k)

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256, k))

	if err != nil {
		return nil, err
	}

	return signed, nil
}
