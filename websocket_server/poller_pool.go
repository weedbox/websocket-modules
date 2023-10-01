package websocket_server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/smallnest/epoller"
)

type PollerPool struct {
	connCount int64
	poller    epoller.Poller
	clients   map[net.Conn]Client
	mutex     sync.RWMutex
	fn        func([]Client)
}

func NewPollerPool() *PollerPool {

	pp := &PollerPool{
		connCount: 0,
		clients:   make(map[net.Conn]Client),
		fn:        func([]Client) {},
	}

	poller, err := epoller.NewPollerWithBuffer(10240)
	if err != nil {
		fmt.Printf("poller: %v\n", err)
		return nil
	}

	pp.poller = poller

	go pp.wait()

	return pp
}

func (pp *PollerPool) wait() {

	for {
		// Getting connections which is triggered
		conns, err := pp.poller.WaitWithBuffer()
		if err != nil {
			if err.Error() != "bad file descriptor" {
				fmt.Printf("failed to poll: %v\n", err)
			}
			continue
		}

		if len(conns) == 0 {
			continue
		}

		pp.mutex.Lock()

		clients := make([]Client, 0, len(conns))
		for _, conn := range conns {
			if client, ok := pp.clients[conn]; ok {
				clients = append(clients, client)
			}
		}

		pp.mutex.Unlock()

		if len(clients) == 0 {
			continue
		}

		pp.fn(clients)
	}
}

func (pp *PollerPool) Add(c Client) error {

	pp.mutex.Lock()

	conn := c.GetConnection()
	pp.clients[conn] = c

	err := pp.poller.Add(conn)
	if err != nil {
		delete(pp.clients, conn)
		return err
	}

	pp.mutex.Unlock()

	atomic.AddInt64(&pp.connCount, 1)

	return nil
}

func (pp *PollerPool) Remove(c Client) error {

	conn := c.GetConnection()

	pp.mutex.Lock()
	pp.poller.Remove(conn)
	delete(pp.clients, conn)
	pp.mutex.Unlock()

	atomic.AddInt64(&pp.connCount, -1)

	return nil
}

func (pp *PollerPool) Wait(fn func([]Client)) {
	pp.fn = fn
}
