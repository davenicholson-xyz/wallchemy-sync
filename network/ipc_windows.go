//go:build windows

package network

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func createListener(path string) (net.Listener, error) {
	return winio.ListenPipe(path, &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GA;;;AU)",
		MessageMode:        true,
		InputBufferSize:    65536,
		OutputBufferSize:   65536,
	})
}

func SendToIPC(path, message string) (string, error) {
	timeout := 2 * time.Second
	conn, err := winio.DialPipe(path, &timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to IPC: %v", err)
	}
	defer conn.Close()

	return sendAndReceive(conn, message)
}
