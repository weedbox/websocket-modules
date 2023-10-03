package websocket_endpoint

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/weedbox/common-modules/http_server"
	"github.com/weedbox/websocket-modules/jsonrpc"
	"github.com/weedbox/websocket-modules/websocket_server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Endpoint struct {
	params   Params
	logger   *zap.Logger
	router   *gin.RouterGroup
	scope    string
	uri      string
	endpoint *websocket_server.Endpoint
}

type Params struct {
	fx.In

	Lifecycle       fx.Lifecycle
	Logger          *zap.Logger
	HTTPServer      *http_server.HTTPServer
	WebSocketServer *websocket_server.WebSocketServer
}

func Module(scope string, uri string) fx.Option {

	var ep *Endpoint

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *Endpoint {

			ep := &Endpoint{
				params: p,
				logger: p.Logger.Named(scope),
				scope:  scope,
				uri:    uri,
			}

			return ep
		}),
		fx.Populate(&ep),
		fx.Invoke(func(p Params) {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: ep.onStart,
					OnStop:  ep.onStop,
				},
			)
		}),
	)
}

func (ep *Endpoint) onStart(ctx context.Context) error {

	ep.logger.Info("Starting Websocket Endpoint", zap.String("uri", ep.uri))

	opts := websocket_server.NewOptions()
	opts.Adapter = websocket_server.NewRPCAdapter(
		websocket_server.WithRPCBackend(&jsonrpc.JSONRPC{}),
	)

	// Create endpoint
	e, err := ep.params.WebSocketServer.CreateEndpoint(ep.uri, opts)
	if err != nil {
		return err
	}

	ep.endpoint = e

	return nil
}

func (ep *Endpoint) onStop(ctx context.Context) error {
	ep.logger.Info("Stopped Websocket Endpoint", zap.String("uri", ep.uri))
	return nil
}
