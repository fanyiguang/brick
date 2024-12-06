package base

import (
	"sync"
	"time"
)

type Mutex[c comparable] struct {
	lock sync.Map
}

func (l *Mutex[c]) Lock(key c) {
	for _, ok := l.lock.LoadOrStore(key, ""); ok; {
		time.Sleep(100 * time.Millisecond)
	}
}

func (l *Mutex[c]) LockInterval(key c, interval int) {
	for _, ok := l.lock.LoadOrStore(key, ""); ok; {
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func (l *Mutex[c]) Unlock(key c) {
	l.lock.Delete(key)
}

func (l *Mutex[c]) LocksMap(keys map[c]struct{}) {
	for key, _ := range keys {
		l.Lock(key)
	}
}

func (l *Mutex[c]) UnlocksMap(keys map[c]struct{}) {
	for key, _ := range keys {
		l.Unlock(key)
	}
}

func (l *Mutex[c]) LocksSlice(keys []c) {
	for _, key := range keys {
		l.Lock(key)
	}
}

func (l *Mutex[c]) UnlocksSlice(keys []c) {
	for _, key := range keys {
		l.Unlock(key)
	}
}
