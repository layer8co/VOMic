package vomic

import (
	"errors"
	"fmt"
	"net"
)

func VerifyPort(port int) error {
	if port == 0 {
		return errors.New("port cannot be zero")
	}
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}

func IsUDPPortAvailable(interfaceAddr string, port int) error {
	if interfaceAddr == "" {
		interfaceAddr = "*"
	}

	conn, err := net.Listen("udp", fmt.Sprintf("%s:%s", interfaceAddr, port))

	if err != nil {
		return err
	}
	conn.Close()

	return nil
}
