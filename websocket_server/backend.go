package websocket_server

import (
	"errors"
	"io"
)

var (
	ErrBackendNotImplemented = errors.New("backend: not implemented")
)

type Backend interface {
	ParseRequest(r io.Reader) (*RPCRequest, error)
	PrepareNotification(eventName string, payload interface{}) ([]byte, error)
	PrepareResponse(*RPCResponse) ([]byte, error)
}

type backend struct {
}

func NewBackend() Backend {
	return &backend{}
}

func (b *backend) ParseRequest(r io.Reader) (*RPCRequest, error) {
	return nil, ErrBackendNotImplemented
}

func (b *backend) PrepareNotification(eventName string, payload interface{}) ([]byte, error) {
	return []byte(""), ErrBackendNotImplemented
}

func (b *backend) PrepareResponse(*RPCResponse) ([]byte, error) {
	return []byte(""), ErrBackendNotImplemented
}
