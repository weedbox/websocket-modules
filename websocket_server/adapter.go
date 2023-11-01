package websocket_server

import (
	"errors"
)

var (
	ErrAdapterNotImplemented = errors.New("adapter: not implemented")
)

type Adapter interface {
	HandleMessage(Client) error
	PrepareNotification(eventName string, payload interface{}) ([]byte, error)
	PrepareResponse(*RPCResponse) ([]byte, error)
	Register(method string, fn RPCFunc) error
	Unregister(method string)
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

func (a *adapter) PrepareNotification(eventName string, payload interface{}) ([]byte, error) {
	return []byte(""), ErrAdapterNotImplemented
}

func (a *adapter) Register(method string, fn RPCFunc) error {
	return ErrAdapterNotImplemented
}

func (a *adapter) Unregister(method string) {
}
