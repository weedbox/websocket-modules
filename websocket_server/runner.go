package websocket_server

import "sync"

type RunnerFunc func(*Runner)

type Runner struct {
	closeCh      chan bool
	IsRunning    bool
	waitForReady sync.WaitGroup
	waitForClose sync.WaitGroup
	fn           func(*Runner)
}

func NewRunner(fn RunnerFunc) *Runner {
	runner := &Runner{
		closeCh:   make(chan bool),
		IsRunning: false,
		fn:        fn,
	}
	return runner
}

func (r *Runner) Start() {
	r.waitForReady.Add(1)
	r.IsRunning = true
	go r.fn(r)
	r.waitForReady.Wait()
}

func (r *Runner) Stop() {
	if !r.IsRunning {
		return
	}

	r.IsRunning = false
	r.closeCh <- true
}

func (r *Runner) WaitForClose() {
	if r.IsRunning {
		r.waitForReady.Done()
	}

	for {
		select {
		case <-r.closeCh:
			return
		}
	}
}
