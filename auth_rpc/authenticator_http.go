package auth_rpc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type HTTPAuthenticator struct {
	Url string
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

	info := &AuthenticationInfo{
		Data: make(map[string]interface{}),
	}

	for k, v := range resp.Header {

		var key string
		fmt.Sscanf(k, "X-Jwt-%s", &key)

		if len(key) == 0 {
			continue
		}

		key = strings.ToLower(key)

		if len(v) == 1 {
			info.Data[key] = v[0]
		} else {
			info.Data[key] = v
		}
	}

	return info, nil
}

func (ha *HTTPAuthenticator) GenerateToken(info *AuthenticationInfo) (string, error) {
	return "", nil
}
