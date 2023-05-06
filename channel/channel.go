package channel

import (
	"time"
)

func Send[a any](ch chan a, val a) {
	ch <- val
}

func NonBlockSend[a any](ch chan a, val a) {
	select {
	case ch <- val:
	default:
	}
}

func SendTimeout[a any](ch chan a, val a, timeout int) {
	select {
	case ch <- val:
	case <-time.After(time.Duration(timeout) * time.Second):
	}
}

func Accept[a any](ch chan a) a {
	return <-ch
}

func NonBlockAccept[a any](ch chan a) (val a, ok bool) {
	select {
	case val = <-ch:
		ok = true
	default:
		ok = false
	}
	return
}

func AcceptTimeout[a any](ch chan a, timeout int) (val a, ok bool) {
	select {
	case val = <-ch:
		ok = true
	case <-time.After(time.Duration(timeout) * time.Second):
		ok = false
	}
	return
}

// Close 并非绝对安全，存在并发问题
func Close[a any](ch chan a) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}

func SafeOperation[a any](done chan a, operation func()) {
	select {
	case <-done:
	default:
		operation()
	}
}
