package websocket_server

const RequestQueueSize = 32

type RequestHandler func(*Context) error

type RequestTask struct {
	Ctx *Context
}

type RequestQueue struct {
	incoming chan *Context
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{
		incoming: make(chan *Context, 1024),
	}
}

func (rq *RequestQueue) Consume(fn RequestHandler) error {

	for i := 0; i < RequestQueueSize; i++ {
		go rq.consume(i, fn)
	}

	return nil
}

func (rq *RequestQueue) consume(id int, fn RequestHandler) {

	for ctx := range rq.incoming {

		err := fn(ctx)
		if err != nil {
			continue
		}
	}
}

func (rq *RequestQueue) Push(c *Context) {
	rq.incoming <- c
}
