//go:build windows

// Package ipc windows直接借鉴 https://github.com/natefinch/npipe， 使用named pipe方便二次修改
// mac, linux 使用unix socket， 这是由内核决定，所以与内核交互也被限制了
// 使用标准net包的接口 - Dial, Listen, Accept, Conn, Listener.
package ipc

import (
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

// ErrPipeNotConnected 当进程结束，但piped name未正常关闭时会报这个错误 - 对端pipe没有进程
// windows报No process is on the other end of the pipe，对应windows/zerrors_windows错误码为233
// 这里统一这个错误，方便处理
var ErrPipeNotConnected = errors.New("no process is on the other end of the pipe")

func NewServerWithListen(address string, msgHandler MsgHandler) (net.Listener, error) {
	listener, err := Listen(address)
	if err != nil {
		return nil, fmt.Errorf("ipc listen %s, %w", address, err)
	}
	go func() {
		for !listener.closed {
			conn, err := listener.Accept()
			if err != nil {
				logrus.Warn(fmt.Sprintf("%v ipc accept connection err： %v", address, err))
				if err == ErrClosed {
					logrus.Warn(fmt.Sprintf("%v close ipc server by received close event err： %v", address, err))
					return
				}
				continue
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
				callerInfo, err := getCallerInfoFromNamedPipeConn(conn)
				if err != nil {
					return
				}
				msgHandler(&Conn{
					pipeName:   address,
					conn:       conn,
					buffer:     make([]byte, pipeBuffer),
					callerInfo: callerInfo,
				})
			}(conn)
		}
	}()
	return nil
}

func NewClientHandler(address string, msgHandler MsgHandler) error {
	conn, err := Dial(address)
	if err != nil {
		return fmt.Errorf("ipc connect error %s, %w", address, err)
	}
	go func(conn net.Conn) {
		callerInfo, err := getCallerInfoFromNamedPipeConn(conn)
		if err != nil {
			return
		}
		msgHandler(&Conn{
			pipeName:   address,
			conn:       conn,
			buffer:     make([]byte, pipeBuffer),
			callerInfo: callerInfo,
		})
	}(conn)
	return nil
}
