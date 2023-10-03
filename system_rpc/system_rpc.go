package system_rpc

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/weedbox/common-modules/http_server"
	"github.com/weedbox/websocket-modules/websocket_server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SystemRPC struct {
	params Params
	logger *zap.Logger
	router *gin.RouterGroup
	scope  string
	uri    string
}

type Params struct {
	fx.In

	Lifecycle       fx.Lifecycle
	Logger          *zap.Logger
	HTTPServer      *http_server.HTTPServer
	WebSocketServer *websocket_server.WebSocketServer
}

func Module(scope string, uri string) fx.Option {

	var srpc *SystemRPC

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *SystemRPC {

			srpc := &SystemRPC{
				params: p,
				logger: p.Logger.Named(scope),
				scope:  scope,
				uri:    uri,
			}

			return srpc
		}),
		fx.Populate(&srpc),
		fx.Invoke(func(p Params) {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: srpc.onStart,
					OnStop:  srpc.onStop,
				},
			)
		}),
	)
}

func (srpc *SystemRPC) onStart(ctx context.Context) error {

	srpc.logger.Info("Starting System RPC", zap.String("uri", srpc.uri))

	ep := srpc.params.WebSocketServer.GetEndpoint(srpc.uri)
	if ep == nil {
		return errors.New("Not found endpoint")
	}

	ep.GetAdapter().Register("System.Ping", srpc.ping)

	return nil
}

func (srpc *SystemRPC) onStop(ctx context.Context) error {
	srpc.logger.Info("Stopped System RPC", zap.String("uri", srpc.uri))
	return nil
}

type Hello struct {
	Name string                 `json:"name"`
	Map  map[string]interface{} `json:"map"`
}

func (srpc *SystemRPC) ping(c *websocket_server.Context) (interface{}, error) {
	return nil, nil
}
