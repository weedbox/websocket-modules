package websocket_server

import "sync/atomic"

type ClientManager struct {
	clientCount uint64

	// Registered clients.
	clients map[Client]struct{}

	// Register requests from the clients.
	register chan Client

	// Unregister requests from clients.
	unregister chan Client

	// Close all clients
	closeAll chan interface{}
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		register:   make(chan Client, 1024),
		unregister: make(chan Client, 1024),
		closeAll:   make(chan interface{}),
		clients:    make(map[Client]struct{}),
	}
}

func (clientMgr *ClientManager) Register(c Client) {
	clientMgr.register <- c
}

func (clientMgr *ClientManager) Unregister(c Client) {
	clientMgr.unregister <- c
}

func (clientMgr *ClientManager) Close() {
	clientMgr.closeAll <- true
}

func (clientMgr *ClientManager) Run() {

	go func() {
		for {
			select {
			case client := <-clientMgr.register:
				atomic.AddUint64((*uint64)(&clientMgr.clientCount), 1)
				clientMgr.clients[client] = struct{}{}

			case client := <-clientMgr.unregister:
				if _, ok := clientMgr.clients[client]; ok {
					delete(clientMgr.clients, client)
					atomic.AddUint64((*uint64)(&clientMgr.clientCount), ^uint64(0))
				}
			case <-clientMgr.closeAll:
				for client := range clientMgr.clients {
					client.Close()
				}
				return
			}
		}
	}()
}
