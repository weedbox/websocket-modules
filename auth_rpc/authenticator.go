package auth_rpc

import (
	"errors"
)

var (
	ErrInvalidToken = errors.New("authenticator: invalid token")
)

type AuthenticationInfo struct {
	Data map[string]interface{}
}

type Authenticator interface {
	Authenticate(token string) (*AuthenticationInfo, error)
	GenerateToken(info *AuthenticationInfo) (string, error)
}
