package network

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

func NewIPCListener() (net.Listener, string, error) {
	if runtime.GOOS == "windows" {
		pipeName := `\\.\pipe\wallchemy`
		listener, err := net.Listen("pipe", pipeName)
		return listener, pipeName, err
	} else {
		socketPath := "/tmp/wallchemy.sock"
		os.Remove(socketPath)
		listener, err := net.Listen("unix", socketPath)
		return listener, socketPath, err
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	// Trim whitespace and convert to string
	message := strings.TrimSpace(string(buf[:n]))
	fmt.Printf("Received message: %q\n", message)

	// Send acknowledgment
	_, err = conn.Write([]byte("ACK: " + message + "\n"))
	if err != nil {
		log.Println("Write error:", err)
	}
}
