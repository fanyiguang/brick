package ipc

import (
	"io"
	"net"
	"time"
)

const (
	pipeBuffer = 4096
)

type CallerInfo struct {
	Addr net.Addr
	PID  int32
	UID  uint32
	GID  uint32
}

// Conn 包装一个真实连接，主要是为了内置buffer，做复用，不用每次都建一个buffer
type Conn struct {
	pipeName   string
	conn       net.Conn
	buffer     []byte
	backup     []byte // 处理粘包，动态分配
	closed     bool
	callerInfo CallerInfo
}

func (c *Conn) PipeName() string {
	return c.pipeName
}

func (c *Conn) IsClose() bool {
	return c.closed
}

func (c *Conn) CallerInfo() CallerInfo {
	return c.callerInfo
}

// Read 读取完整数据，以0字节结尾
func (c *Conn) Read() ([]byte, error) {
	var data []byte
	for !c.closed {
		// 先判断粘包未处理的数据
		if len(c.backup) > 0 {
			n, ok := checkPacketEnding(c.backup)
			if ok {
				data = append(data, c.backup[:n]...)
				if n < len(c.backup) {
					c.backup = append(c.backup[:0], c.backup[n+1:]...)
				} else {
					c.backup = []byte{}
				}
				if len(data) == 0 {
					continue
				}
				return data, nil
			}
			data = append(data, c.backup...)
			c.backup = []byte{}
		}
		// log.Default().Debug(fmt.Sprintf("1 backup:%d buffer:%d data:%d", len(c.backup), len(c.buffer), len(data)))
		n, err := c.conn.Read(c.buffer)
		if err != nil {
			if err == io.EOF {
				c.closed = true
			}
			return data, err
		}
		// log.Default().Debug(fmt.Sprintf("2 backup:%d buffer:%d n:%d data:%d", len(c.backup), len(c.buffer), n, len(data)))
		if n == 0 {
			continue
		}
		_n, ok := checkPacketEnding(c.buffer[:n])
		// log.Default().Debug(fmt.Sprintf("3 backup:%d buffer:%d n:%d ok:%t data:%d", len(c.backup), len(c.buffer), n, ok, len(data)))
		if ok {
			data = append(data, c.buffer[:_n]...)
			// n包含0的长度， _n是0的位置
			if _n < n-1 {
				c.backup = append(c.backup[:0], c.buffer[_n+1:n]...)
			}
			// log.Default().Debug(fmt.Sprintf("4 backup:%d buffer:%d n:%d _n:%d ok:%t data:%d", len(c.backup), len(c.buffer), n, _n, ok, len(data)))
			return data, nil
		}
		data = append(data, c.buffer[:n]...)

		//data = append(data, c.buffer[:n]...)
		//if n == 0 || c.buffer[n-1] == 0 {
		//	if len(data) > 0 {
		//		return data[:len(data)-1], nil
		//	}
		//}
		//TODO 先临时限制数据大小，大小为100M，如果没有以0结尾的数据会无限增大数据，导致OOM
		if len(data) > 104857600 {
			panic("data too large")
		}

	}
	return data, nil
}

// checkPacketEnding 判断是否包含0字节，处理粘包情况
func checkPacketEnding(data []byte) (int, bool) {
	if len(data) == 0 {
		return 0, false
	}

	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return i, true
		}
	}
	return len(data), false
}

// Write 使用0字节结尾
func (c *Conn) Write(data []byte) error {
	data = append(data, '\000')
	_, err := c.conn.Write(data)
	return err
}

func (c *Conn) Close() error {
	c.closed = true
	return c.conn.Close()
}

// MsgHandler ipc信息处理
type MsgHandler func(conn *Conn)

func NewClient(address string) (*Conn, error) {
	conn, err := Dial(address)
	if err != nil {
		return nil, err
	}
	return &Conn{
		pipeName: address,
		conn:     conn,
		buffer:   make([]byte, pipeBuffer),
		backup:   []byte{},
	}, nil
}

func NewClientWithTimeout(address string, timeout time.Duration) (*Conn, error) {
	conn, err := DialTimeout(address, timeout)
	if err != nil {
		return nil, err
	}
	return &Conn{
		pipeName: address,
		conn:     conn,
		buffer:   make([]byte, pipeBuffer),
		backup:   []byte{},
	}, nil
}
