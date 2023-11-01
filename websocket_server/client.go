package websocket_server

import (
	"errors"
	"io"
	"net"

	"github.com/google/uuid"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	ErrConnectionClosed = errors.New("client: conn closed")
)

type Packet struct {
	Length  int
	Payload []byte
}

type Client interface {
	GetOptions() *Options
	GetConnection() net.Conn
	GetClientID() uuid.UUID
	GetMeta() *Metadata
	GetReader() io.Reader
	Send(data []byte) error
	Respond(res *RPCResponse) error
	Notify(eventName string, payload interface{}) error
	Resume() error
	CreateRunner(func(*Runner)) *Runner
	Release()
	Close()
}

type client struct {
	conn    net.Conn
	reader  *wsutil.Reader
	options *Options
	id      uuid.UUID
	meta    *Metadata
	runners []*Runner
}

func NewClient(options *Options, conn net.Conn) Client {

	c := &client{
		options: options,
		id:      uuid.New(),
		conn:    conn,
		meta:    NewMetadata(),
		runners: make([]*Runner, 0),
	}

	c.reader = wsutil.NewReader(c.conn, ws.StateServerSide)

	return c
}

func (c *client) GetOptions() *Options {
	return c.options
}

func (c *client) GetConnection() net.Conn {
	return c.conn
}

func (c *client) GetReader() io.Reader {
	return c.reader
}

func (c *client) GetMeta() *Metadata {
	return c.meta
}

func (c *client) Release() {

	// Stop all runners
	for _, r := range c.runners {
		r.Stop()
	}

	c.runners = make([]*Runner, 0)
}

func (c *client) Close() {
	c.Release()
	c.conn.Close()
}

func (c *client) Resume() error {

	header, err := c.reader.NextFrame()
	if err != nil {
		return err
	}

	// EOF
	if header.OpCode == ws.OpClose {
		return ErrConnectionClosed
	}

	if err := c.options.Adapter.HandleMessage(c); err != nil {
		return err
	}

	return nil
}

func (c *client) CreateRunner(fn func(*Runner)) *Runner {
	r := NewRunner(fn)
	c.runners = append(c.runners, r)
	return r
}

func (c *client) GetClientID() uuid.UUID {
	return c.id
}

func (c *client) Send(data []byte) error {

	w := wsutil.NewWriter(c.conn, ws.StateServerSide, ws.OpText)

	w.Write(data)

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func (c *client) Respond(res *RPCResponse) error {

	data, err := c.options.Adapter.PrepareResponse(res)
	if err != nil {
		return err
	}

	return c.Send(data)
}

func (c *client) Notify(eventName string, payload interface{}) error {

	data, err := c.options.Adapter.PrepareNotification(eventName, payload)
	if err != nil {
		return err
	}

	return c.Send(data)
}
