package safecloser

import "sync"

type Closer struct {
	sync.Once
	doneCh chan struct{}
}

func New() Closer {
	return Closer{
		doneCh: make(chan struct{}),
	}
}

func (c *Closer) DoneCh() chan struct{} {
	return c.doneCh
}

func (c *Closer) Close() {
	c.Do(func() {
		if c.doneCh == nil {
			return
		}
		close(c.doneCh)
	})
}
