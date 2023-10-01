package websocket_server

import "io"

type Context struct {
	client Client
	meta   *Metadata
	req    *RPCRequest
}

func NewContext(client Client, req *RPCRequest) *Context {
	return &Context{
		client: client,
		req:    req,
		meta:   NewMetadata(),
	}
}

func (ctx *Context) GetMeta() *Metadata {
	return ctx.meta
}

func (ctx *Context) GetClient() Client {
	return ctx.client
}

func (ctx *Context) GetReader() io.Reader {
	return ctx.client.GetReader()
}

func (ctx *Context) GetRequest() *RPCRequest {
	return ctx.req
}

func (ctx *Context) Error(code RPCErrorCode, data string) error {
	return NewError(code, data)
}

func (ctx *Context) Send(data []byte) error {
	return ctx.client.Send(data)
}

func (ctx *Context) Respond(res *RPCResponse) error {
	return ctx.client.Respond(res)
}

func (ctx *Context) Notify(eventName string, payload []byte) error {
	return ctx.client.Notify(eventName, payload)
}
