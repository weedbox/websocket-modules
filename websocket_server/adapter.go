package websocket_server

import (
	"errors"
)

var (
	ErrAdapterNotImplemented = errors.New("adapter: not implemented")
)

type Adapter interface {
	HandleMessage(Client) error
	PrepareNotification(eventName string, payload []byte) ([]byte, error)
	PrepareResponse(*RPCResponse) ([]byte, error)
}

type adapter struct {
}

func NewAdapter() Adapter {
	return &adapter{}
}

func (a *adapter) HandleMessage(c Client) error {
	return c.GetOptions().OnMessage(c)
}

func (a *adapter) PrepareResponse(res *RPCResponse) ([]byte, error) {
	return []byte(""), ErrAdapterNotImplemented
}

func (a *adapter) PrepareNotification(eventName string, payload []byte) ([]byte, error) {
	return []byte(""), ErrAdapterNotImplemented
}
