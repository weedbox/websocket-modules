package websocket_server

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/weedbox/common-modules/http_server"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var logger *zap.Logger

type WebSocketServer struct {
	params    Params
	logger    *zap.Logger
	router    *gin.RouterGroup
	scope     string
	endpoints map[*Endpoint]*Endpoint
}

type Params struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Logger     *zap.Logger
	HTTPServer *http_server.HTTPServer
}

func Module(scope string) fx.Option {

	var wss *WebSocketServer

	return fx.Module(
		scope,
		fx.Provide(func(p Params) *WebSocketServer {

			wss := &WebSocketServer{
				params:    p,
				logger:    p.Logger.Named(scope),
				scope:     scope,
				endpoints: make(map[*Endpoint]*Endpoint),
			}

			logger = wss.logger

			return wss
		}),
		fx.Populate(&wss),
		fx.Invoke(func(p Params) {

			p.Lifecycle.Append(
				fx.Hook{
					OnStart: wss.onStart,
					OnStop:  wss.onStop,
				},
			)
		}),
	)

}

func (wss *WebSocketServer) onStart(ctx context.Context) error {
	wss.logger.Info("Starting WebSocketServer")
	return nil
}

func (wss *WebSocketServer) onStop(ctx context.Context) error {
	wss.logger.Info("Stopped WebSocketServer")
	return nil
}

func (wss *WebSocketServer) CreateEndpoint(uri string, opts *Options) (*Endpoint, error) {

	ep := NewEndpoint(uri, opts)

	wss.params.HTTPServer.GetRouter().GET(uri, func(c *gin.Context) {
		ep.Establish(c)
	})

	wss.endpoints[ep] = ep

	return ep, nil
}

func (wss *WebSocketServer) RemoveEndpoint(ep *Endpoint) error {
	delete(wss.endpoints, ep)
	return nil
}
