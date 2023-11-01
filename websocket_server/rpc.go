package websocket_server

import (
	"errors"

	"go.uber.org/zap"
)

var (
	ErrMethodNotFound = errors.New("rpc: method not found")
)

type RPCAdapterOpt func(*RPCAdapter)
type RPCFunc func(*Context) (interface{}, error)

type RPCRequest struct {
	ID     int64
	Method string
	Params interface{}
}

type RPCResponse struct {
	ID     int64
	Error  error
	Result interface{}
}

type RPCAdapter struct {
	backend      Backend
	requestQueue *RequestQueue
	methods      map[string]RPCFunc
}

func WithRPCBackend(b Backend) RPCAdapterOpt {
	return func(a *RPCAdapter) {
		a.backend = b
	}
}

func NewRPCAdapter(opts ...RPCAdapterOpt) *RPCAdapter {

	ra := &RPCAdapter{
		requestQueue: NewRequestQueue(),
		methods:      make(map[string]RPCFunc),
	}

	for _, o := range opts {
		o(ra)
	}

	if ra.backend == nil {
		ra.backend = NewBackend()
	}

	ra.requestQueue.Consume(ra.consume)

	return ra
}

func (ra *RPCAdapter) consume(c *Context) error {

	err := ra.handleRequest(c)
	if err != nil {

		if errors.Is(err, ErrMethodNotFound) {
			err = NewError(ErrorCode_NotFound, nil)
		}

		// Error
		res := &RPCResponse{
			ID:     c.GetRequest().ID,
			Error:  err,
			Result: "",
		}

		return ra.respond(c, res)
	}

	return nil
}

func (ra *RPCAdapter) handleRequest(c *Context) error {

	method := c.GetRequest().Method

	fn, ok := ra.methods[method]
	if !ok {
		return ErrMethodNotFound
	}

	// Invoke
	returnedValue, err := fn(c)

	// Response with returned value
	res := &RPCResponse{
		ID:     c.GetRequest().ID,
		Error:  err,
		Result: returnedValue,
	}

	return ra.respond(c, res)
}

func (ra *RPCAdapter) respond(c *Context, res *RPCResponse) error {

	data, err := ra.PrepareResponse(res)
	if err != nil {
		return err
	}

	return c.Send(data)
}

func (ra *RPCAdapter) HandleMessage(c Client) error {

	r := c.GetReader()

	// Parse message
	req, err := ra.backend.ParseRequest(r)
	if err != nil {
		// Ignora unrecognized message
		return err
	}

	// Preparing context
	ctx := NewContext(c, req)

	// Push to queue for processing
	ra.requestQueue.Push(ctx)

	return nil
}

func (ra *RPCAdapter) Register(method string, fn RPCFunc) error {
	logger.Info("Registering", zap.String("method", method))
	ra.methods[method] = fn
	return nil
}

func (ra *RPCAdapter) Unregister(method string) {
	delete(ra.methods, method)
}

func (ra *RPCAdapter) PrepareResponse(res *RPCResponse) ([]byte, error) {
	return ra.backend.PrepareResponse(res)
}

func (ra *RPCAdapter) PrepareNotification(eventName string, payload interface{}) ([]byte, error) {
	return ra.backend.PrepareNotification(eventName, payload)
}
