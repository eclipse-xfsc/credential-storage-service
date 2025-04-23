package model

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type ContextKey string

const (
	AuthModelKey ContextKey = "authModel"
)

type AuthModel struct {
	Account        string
	TenantId       string
	Device_Key     *jwk.Key
	Nonce          string
	Recovery_Nonce string
	Token          *jwt.Token
}
