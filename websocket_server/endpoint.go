package websocket_server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

type Options struct {
	Adapter        Adapter
	MaxClients     int
	OnConnected    func(Client) error
	OnDisconnected func(Client) error
	OnMessage      func(Client) error
}

func NewOptions() *Options {
	return &Options{
		MaxClients: 4096,
		Adapter:    NewAdapter(),
		OnConnected: func(c Client) error {
			return nil
		},
		OnDisconnected: func(c Client) error {
			return nil
		},
		OnMessage: func(c Client) error {
			return nil
		},
	}
}

type Endpoint struct {
	options    *Options
	clientMgr  *ClientManager
	pollerPool *PollerPool
	uri        string
}

func NewEndpoint(uri string, options *Options) *Endpoint {

	ep := &Endpoint{
		options:    options,
		clientMgr:  NewClientManager(),
		pollerPool: NewPollerPool(),
		uri:        uri,
	}

	ep.clientMgr.Run()

	// Initializing epoll pool
	ep.pollerPool.Wait(func(clients []Client) {

		for _, c := range clients {

			err := c.Resume()
			if err != nil {
				c.Close()

				ep.pollerPool.Remove(c)

				// Unregister client
				ep.clientMgr.Unregister(c)

				// Emit event
				ep.options.OnDisconnected(c)
			}
		}
	})

	return ep
}

func (ep *Endpoint) GetUri() string {
	return ep.uri
}

func (ep *Endpoint) GetAdapter() Adapter {
	return ep.options.Adapter
}

func (ep *Endpoint) Establish(c *gin.Context) {

	// Check protocol
	connectionType := c.GetHeader("Connection")
	if connectionType != "Upgrade" && connectionType != "upgrade" {
		c.String(http.StatusOK, fmt.Sprintf("Unsupported header: %s", connectionType))
		return
	}

	// Disallow to establish connection when the number of clients exceeds
	if ep.clientMgr.clientCount >= uint64(ep.options.MaxClients) {
		logger.Warn("Too Many Connections")
		c.String(http.StatusTooManyRequests, "Too Many Connections")
		return
	}

	// Initializing websocket connection
	conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Create client
	client := NewClient(ep.options, conn)

	ep.clientMgr.Register(client)
	ep.pollerPool.Add(client)

	// Emit event
	ep.options.OnConnected(client)
}
