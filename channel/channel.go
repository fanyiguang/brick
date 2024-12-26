package channel

import (
	"errors"
	"time"
)

func Send[a any](ch chan a, val a) {
	ch <- val
}

func NonBlockSend[a any](ch chan a, val a) error {
	select {
	case ch <- val:
		return nil
	default:
		return errors.New("send error")
	}
}

func SendTimeout[a any](ch chan a, val a, timeout int) error {
	select {
	case ch <- val:
		return nil
	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("timeout")
	}
}

func Accept[a any](ch chan a) a {
	return <-ch
}

func NonBlockAccept[a any](ch chan a) (val a, err error) {
	select {
	case val = <-ch:
	default:
		err = errors.New("not accept")
	}
	return
}

func AcceptTimeout[a any](ch chan a, timeout int) (val a, err error) {
	select {
	case val = <-ch:
	case <-time.After(time.Duration(timeout) * time.Second):
		err = errors.New("timeout")
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

func SafeOperation[a any](done chan a, operation func() error) error {
	select {
	case <-done:
		return errors.New("done")
	default:
		return operation()
	}
}
