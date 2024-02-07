package network

import (
	"fmt"
	"net"
)

func GetTCPIdlePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func GetUDPIdlePort() (int, error) {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.LocalAddr().(*net.UDPAddr).Port, nil
}

func GetIdlePort() (int, error) {
	for i := 0; i < 10; i++ {
		port, err := GetTCPIdlePort()
		if err != nil {
			return 0, err
		}
		address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", "0.0.0.0", port))
		if err != nil {
			continue
		}
		listener, err := net.ListenUDP("udp", address)
		if err != nil {
			continue
		}
		_ = listener.Close()
		return port, nil
	}
	return 0, nil
}
