package pipeline

import (
	"github.com/fanyiguang/brick/channel"
	"time"
)

type PipeLine[V any] struct {
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
	channel.Close(pipeline.doneCh)
	for _, option := range options {
		option(pipeline)
	}
	pipeline.cacheBuffer = make([]V, 0, pipeline.limit)
	return pipeline
}

func (p *PipeLine[V]) Start() {
	if !p.started {
		p.doneCh = make(chan struct{})
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

func (p *PipeLine[V]) Push(val V) error {
	return channel.SafeOperation(p.doneCh, func() error {
		channel.Send(p.VCh, val)
		return nil
	})
}

func (p *PipeLine[V]) PushTimeout(val V, timeout int) error {
	return channel.SafeOperation(p.doneCh, func() error {
		return channel.SendTimeout(p.VCh, val, timeout)
	})
}

func (p *PipeLine[V]) do(data []V) {
	if p.handler != nil {
		p.handler(data)
	}
}

func (p *PipeLine[V]) Close(wait int) {
	if p.started {
		channel.Close(p.doneCh)
		time.Sleep(time.Duration(p.waitTime(wait)) * time.Second)
		if len(p.cacheBuffer) > 0 {
			go p.do(p.cacheBuffer)
			p.cacheBuffer = make([]V, p.limit)
		}
		p.started = false
	}
}

func (p *PipeLine[V]) waitTime(wait int) int {
	if wait < 0 {
		wait = 2
	}
	return wait
}
