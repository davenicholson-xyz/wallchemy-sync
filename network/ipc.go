package network

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

type IPCListener struct {
	listener net.Listener
	path     string
	stopChan chan struct{}
}

func NewIPCListener() *IPCListener {
	var listener net.Listener
	var path string
	var err error

	if runtime.GOOS == "windows" {
		path = `\\.\pipe\wallchemy`
		listener, err = net.Listen("pipe", path)
	} else {
		path = "/tmp/wallchemy.sock"
		os.Remove(path)
		listener, err = net.Listen("unix", path)
	}

	if err != nil {
		log.Fatal("Could not stort IPC:", err)
	}

	return &IPCListener{
		listener: listener,
		path:     path,
		stopChan: make(chan struct{}),
	}
}

func (ipc *IPCListener) Start() {
	go ipc.listenLoop()
	fmt.Printf("IPC listening on %s (%s)", ipc.path, runtime.GOOS)
}

func (ipc *IPCListener) Stop() {
	close(ipc.stopChan)
	ipc.listener.Close()
	if runtime.GOOS != "windows" {
		os.Remove(ipc.path)
	}
}

func (ipc *IPCListener) listenLoop() {
	defer ipc.Stop()

	for {
		select {
		case <-ipc.stopChan:
			return
		default:
			conn, err := ipc.listener.Accept()
			if err != nil {
				select {
				case <-ipc.stopChan:
					return
				default:
					log.Println("Accept error:", err)
				}
				continue
			}

			go HandleConnection(conn)
		}
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

	message := strings.TrimSpace(string(buf[:n]))
	fmt.Printf("Received message: %q\n", message)

	_, err = conn.Write([]byte("ACK: " + message + "\n"))
	if err != nil {
		log.Println("Write error:", err)
	}
}
