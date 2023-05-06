//go:build darwin

package ipc

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type ipcListener struct {
	closed bool
	net.Listener
	mu sync.Mutex
}

func (i ipcListener) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.closed {
		return nil
	}
	i.closed = true
	return i.Listener.Close()
}

func Listen(address string) (*ipcListener, error) {
	cleanup := func() {
		if _, err := os.Stat(address); err == nil {
			os.RemoveAll(address)
		}
	}

	cleanup()

	listener, err := net.Listen("unix", address)
	if err != nil {
		return nil, err
	}
	return &ipcListener{
		Listener: listener,
	}, nil
}

func NewServer(address string, msgHandler MsgHandler) error {
	listener, err := Listen(address)
	if err != nil {
		return fmt.Errorf("ipc listen %s, %w", address, err)
	}
	go func() {
		for !listener.closed {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				msgHandler(&Conn{
					pipeName: address,
					conn:     conn,
					buffer:   make([]byte, pipeBuffer),
				})
			}(conn)
		}
	}()
	return nil
}

func NewServerWithListen(address string, msgHandler MsgHandler) (net.Listener, error) {
	listener, err := Listen(address)
	if err != nil {
		return nil, fmt.Errorf("ipc listen %s, %w", address, err)
	}
	go func() {
		for !listener.closed {
			conn, err := listener.Accept()
			if err != nil {
				//logrus.Warn(fmt.Sprintf("%v ipc accept connection err： %v", address, err))
				//if err == ErrClosed {
				//	logrus.Warn(fmt.Sprintf("%v close ipc server by received close event err： %v", address, err))
				//	return
				//}
				//continue
				return
			}
			go func(conn net.Conn) {
				msgHandler(&Conn{
					pipeName: address,
					conn:     conn,
					buffer:   make([]byte, pipeBuffer),
				})
			}(conn)
		}
	}()
	return listener, nil
}

func Dial(address string) (net.Conn, error) {
	return net.Dial("unix", address)
}

// DialTimeout acts like Dial, but will time out after the duration of timeout
func DialTimeout(address string, timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{
		Timeout: timeout,
	}
	return dialer.Dial("unix", address)
}
