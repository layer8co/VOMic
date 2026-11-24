package vomic

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
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
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", interfaceAddr, port))

	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

func IsPactlExists() bool {
	_, err := exec.LookPath("pactl")

	if err != nil {
		return false
	}

	return true
}
