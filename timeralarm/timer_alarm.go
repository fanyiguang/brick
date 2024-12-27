package timeralarm

import (
	"github.com/fanyiguang/brick/Go"
	"github.com/fanyiguang/brick/channel"
	"github.com/fanyiguang/brick/safecloser"
	"sync"
	"time"
)

type TimerAlarm struct {
	safecloser.Closer
	startOnce sync.Once
	timerCh   chan struct{}
	finishFn  func()
	repeated  bool
}

func New(finishFn func(), repeated bool) *TimerAlarm {
	return &TimerAlarm{
		timerCh:  make(chan struct{}, 1),
		finishFn: finishFn,
		Closer:   safecloser.New(),
		repeated: repeated,
	}
}

func (t *TimerAlarm) Start(timeout int) {
	t.startOnce.Do(func() {
		Go.Go(func() {
			t.start(timeout)
		})
	})
}

func (t *TimerAlarm) start(timeout int) {
	ticker := time.NewTicker(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-ticker.C:
			_, err := channel.NonBlockAccept(t.timerCh)
			if err != nil {
				t.finishFn()
				if !t.repeated {
					ticker.Stop()
					return
				}
			}
		case <-t.DoneCh():
			return
		}
	}
}

func (t *TimerAlarm) ResetTimer() {
	_ = channel.NonBlockSend(t.timerCh, struct{}{})
}

func (t *TimerAlarm) Close() {
	t.Closer.Close()
}
