package auth_rpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/weedbox/common-modules/http_server"
	"github.com/weedbox/websocket-modules/websocket_server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AuthenticateResponse struct {
	Success bool                   `json:"success"`
	Message bool                   `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type AuthRPC struct {
	params Params
	logger *zap.Logger
	router *gin.RouterGroup
	scope  string
	uri    string

	// Implementations
	authenticator Authenticator
}

type Params struct {
	fx.In

	Lifecycle       fx.Lifecycle
	Logger          *zap.Logger
	HTTPServer      *http_server.HTTPServer
	WebSocketServer *websocket_server.WebSocketServer
}

type Option func(*AuthRPC)

func WithAuthenticator(a Authenticator) Option {
	return func(arpc *AuthRPC) {
		arpc.authenticator = a
	}
}

func WithHTTPAuthenticator() Option {
	return func(arpc *AuthRPC) {
		arpc.authenticator = NewHTTPAuthenticator("")
	}
}

func WithJWTAuthenticator() Option {
	return func(arpc *AuthRPC) {
		arpc.authenticator = NewJWTAuthenticator("")
	}
}

func Module(scope string, uri string, opts ...Option) fx.Option {

	var arpc *AuthRPC

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *AuthRPC {

			arpc := &AuthRPC{
				params:        p,
				logger:        p.Logger.Named(scope),
				scope:         scope,
				uri:           uri,
				authenticator: NewJWTAuthenticator(""),
			}

			arpc.initDefaultConfigs()

			for _, o := range opts {
				o(arpc)
			}

			return arpc
		}),
		fx.Populate(&arpc),
		fx.Invoke(func(p Params) {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: arpc.onStart,
					OnStop:  arpc.onStop,
				},
			)
		}),
	)
}

func (arpc *AuthRPC) getConfigPath(key string) string {
	return fmt.Sprintf("%s.%s", arpc.scope, key)
}

func (arpc *AuthRPC) initDefaultConfigs() {
	viper.SetDefault(arpc.getConfigPath("secret"), "")
	viper.SetDefault(arpc.getConfigPath("auth_url"), "http://0.0.0.0/auth")
}

func (arpc *AuthRPC) onStart(ctx context.Context) error {

	arpc.logger.Info("Starting Auth RPC",
		zap.String("uri", arpc.uri),
	)

	// Set secret if it is JWT authenticator
	if a, ok := arpc.authenticator.(*JWTAuthenticator); ok {
		secret := viper.GetString(arpc.getConfigPath("secret"))
		a.UpdateSecret(secret)
	}

	// Set url if it is HTTP authenticator
	if a, ok := arpc.authenticator.(*HTTPAuthenticator); ok {
		auth_url := viper.GetString(arpc.getConfigPath("auth_url"))
		a.UpdateUrl(auth_url)
	}

	ep := arpc.params.WebSocketServer.GetEndpoint(arpc.uri)
	if ep == nil {
		return errors.New("Not found endpoint")
	}

	ep.GetAdapter().Register("Auth.Authenticate", arpc.authenticate)

	return nil
}

func (arpc *AuthRPC) onStop(ctx context.Context) error {

	arpc.logger.Info("Stopped Auth RPC")

	return nil
}

func (arpc *AuthRPC) authenticate(c *websocket_server.Context) (interface{}, error) {

	parameters := c.GetRequest().Params.([]interface{})

	token := parameters[0].(string)

	if len(token) == 0 {
		return nil, websocket_server.NewError(websocket_server.ErrorCode_InvalidParams, nil)
	}

	res := &AuthenticateResponse{
		Success: true,
	}

	info, err := arpc.authenticator.Authenticate(token)
	if err != nil {
		res.Success = false
		return res, nil
	}

	res.Data = make(map[string]interface{})

	for k, v := range info.Data {
		res.Data[k] = v
		c.GetMeta().Set(k, v)
	}

	return res, nil
}
