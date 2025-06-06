//go:build !windows

package network

import (
	"fmt"
	"net"
	"os"
)

func createListener(path string) (net.Listener, error) {
	// Remove existing socket file
	os.Remove(path)
	return net.Listen("unix", path)
}

func SendToIPC(path, message string) (string, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return "", fmt.Errorf("failed to connect to IPC: %v", err)
	}
	defer conn.Close()

	return sendAndReceive(conn, message)
}
