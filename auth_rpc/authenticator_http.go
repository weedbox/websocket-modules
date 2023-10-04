package auth_rpc

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type HTTPAuthenticator struct {
	Url string
}

type HTTPClaims struct {
	jwt.StandardClaims

	UserID string `json:"user_id"`
}

func NewHTTPAuthenticator(url string) *HTTPAuthenticator {
	return &HTTPAuthenticator{
		Url: url,
	}
}

func (ha *HTTPAuthenticator) UpdateUrl(url string) {
	ha.Url = url
}

func (ha *HTTPAuthenticator) Authenticate(token string) (*AuthenticationInfo, error) {

	req, err := http.NewRequest("GET", ha.Url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("authentication service returned error")
	}

	userID := resp.Header.Get("X-Jwt-Id")

	return &AuthenticationInfo{
		UserID: userID,
	}, nil
}

func (ha *HTTPAuthenticator) GenerateToken(info *AuthenticationInfo) (string, error) {
	return "", nil
}
