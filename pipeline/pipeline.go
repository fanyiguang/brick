package pipeline

import (
	"brick/channel"
	"runtime"
	"sync"
	"time"
)

type PipeLine[V any] struct {
	mt          sync.Mutex
	handler     func([]V)
	VCh         chan V
	cacheBuffer []V
	doneCh      chan struct{}
	limit       int
	mode        int
	doInterval  int
	started     bool
}

func New[V any](handler func([]V), options ...Option[V]) *PipeLine[V] {
	pipeline := &PipeLine[V]{
		handler: handler,
		VCh:     make(chan V),
		limit:   50,
		doneCh:  make(chan struct{}),
	}

	for _, option := range options {
		option(pipeline)
	}

	pipeline.cacheBuffer = make([]V, 0, pipeline.limit)
	return pipeline
}

func (p *PipeLine[V]) Start() {
	p.mt.Lock()
	defer p.mt.Unlock()

	if !p.started {
		switch p.mode {
		case modeIntervalReport:
			go p.intervalAccept()
		case modeLimitReport:
			fallthrough
		default:
			go p.limitAccept()
		}
		p.started = true
	}
}

func (p *PipeLine[V]) limitAccept() {
	for {
		select {
		case _item := <-p.VCh:
			p.cacheBuffer = append(p.cacheBuffer, _item)
			if len(p.cacheBuffer) >= p.limit {
				go p.do(p.cacheBuffer)
				p.cacheBuffer = make([]V, 0, p.limit)
			}
		case <-p.doneCh:
			return
		}
	}
}

func (p *PipeLine[V]) intervalAccept() {
	if p.doInterval <= 0 {
		p.doInterval = 30
	}
	ticker := time.NewTicker(time.Duration(p.doInterval) * time.Second)
	for {
		select {
		case _item := <-p.VCh:
			p.cacheBuffer = append(p.cacheBuffer, _item)
		case <-ticker.C:
			if len(p.cacheBuffer) > 0 {
				go p.do(p.cacheBuffer)
				p.cacheBuffer = make([]V, 0, p.limit)
			}
		case <-p.doneCh:
			return
		}
	}
}

func (p *PipeLine[V]) Push(val V) {
	p.mt.Lock()
	defer p.mt.Unlock()

	channel.SafeOperation(p.doneCh, func() {
		channel.Send(p.VCh, val)
	})
}

func (p *PipeLine[V]) PushTimeout(val V, timeout int) {
	p.mt.Lock()
	defer p.mt.Unlock()

	channel.SafeOperation(p.doneCh, func() {
		channel.SendTimeout(p.VCh, val, timeout)
	})
}

func (p *PipeLine[V]) do(data []V) {
	if p.handler != nil {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64*1024)
				buf = buf[:runtime.Stack(buf, false)]
				//log.ErrorF("panic in fanout proc, err: %s, stack: %s", r, buf)
			}
		}()
		p.handler(data)
	}
}

func (p *PipeLine[V]) Close() {
	p.mt.Lock()
	defer p.mt.Unlock()

	channel.Close(p.doneCh)
	if len(p.cacheBuffer) > 0 {
		go p.do(p.cacheBuffer)
		p.cacheBuffer = make([]V, 0)
	}
}
