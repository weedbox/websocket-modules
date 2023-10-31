package auth_rpc

import (
	"github.com/golang-jwt/jwt"
)

type JWTAuthenticator struct {
	Secret []byte
}

type JWTClaims struct {
	jwt.StandardClaims

	UserID string `json:"user_id"`
}

func NewJWTAuthenticator(secret string) *JWTAuthenticator {
	return &JWTAuthenticator{
		Secret: []byte(secret),
	}
}

func (ja *JWTAuthenticator) UpdateSecret(secret string) {
	ja.Secret = []byte(secret)
}

func (ja *JWTAuthenticator) Authenticate(token string) (*AuthenticationInfo, error) {

	t, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return ja.Secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := t.Claims.(*JWTClaims)
	if !ok || !t.Valid {
		return nil, ErrInvalidToken
	}

	return &AuthenticationInfo{
		Data: map[string]interface{}{
			"id": claims.UserID,
		},
	}, nil
}

func (ja *JWTAuthenticator) GenerateToken(info *AuthenticationInfo) (string, error) {

	claims := JWTClaims{
		UserID: info.Data["id"].(string),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signed, err := token.SignedString(ja.Secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}
