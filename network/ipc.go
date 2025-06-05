package network

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	DefaultBufferSize    = 1024
	DefaultChannelBuffer = 100
)

// IPCMessage represents a message received via IPC
type IPCMessage struct {
	Content    string
	Connection net.Conn
	Time       time.Time
	ResponseCh chan string // Channel to send response back
}

// IPCClient provides cross-platform IPC communication
type IPCClient struct {
	listener    net.Listener
	path        string
	stopChan    chan struct{}
	messageChan chan IPCMessage
	bufferSize  int
	mu          sync.RWMutex
	running     bool
}

// IPCConfig holds configuration options for the IPCClient
type IPCConfig struct {
	AppName       string // Application name for auto-generating paths
	Path          string // Custom path (optional, overrides AppName)
	BufferSize    int    // Buffer size for reading messages
	ChannelBuffer int    // Buffer size for message channel
}

// NewIPCClient creates a new IPC client with the given configuration
func NewIPCClient(config IPCConfig) (*IPCClient, error) {
	if config.BufferSize == 0 {
		config.BufferSize = DefaultBufferSize
	}
	if config.ChannelBuffer == 0 {
		config.ChannelBuffer = DefaultChannelBuffer
	}

	var listener net.Listener
	var path string
	var err error

	if config.Path != "" {
		// Use custom path
		path = config.Path
		if runtime.GOOS == "windows" {
			listener, err = net.Listen("pipe", path)
		} else {
			os.Remove(path) // Remove existing socket
			listener, err = net.Listen("unix", path)
		}
	} else if config.AppName != "" {
		// Generate path from app name
		if runtime.GOOS == "windows" {
			path = fmt.Sprintf(`\\.\pipe\%s`, config.AppName)
			listener, err = net.Listen("pipe", path)
		} else {
			path = fmt.Sprintf("/tmp/%s.sock", config.AppName)
			os.Remove(path)
			listener, err = net.Listen("unix", path)
		}
	} else {
		// Use default paths
		if runtime.GOOS == "windows" {
			path = `\\.\pipe\app_ipc`
			listener, err = net.Listen("pipe", path)
		} else {
			path = "/tmp/app_ipc.sock"
			os.Remove(path)
			listener, err = net.Listen("unix", path)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create IPC listener: %v", err)
	}

	return &IPCClient{
		listener:    listener,
		path:        path,
		stopChan:    make(chan struct{}),
		messageChan: make(chan IPCMessage, config.ChannelBuffer),
		bufferSize:  config.BufferSize,
	}, nil
}

// Start begins listening for IPC connections
func (ipc *IPCClient) Start() error {
	ipc.mu.Lock()
	defer ipc.mu.Unlock()

	if ipc.running {
		return fmt.Errorf("IPC client already running")
	}

	ipc.running = true
	fmt.Printf("IPC listening on %s (%s)\n", ipc.path, runtime.GOOS)
	go ipc.listenLoop()

	return nil
}

// listenLoop runs the main listening loop in a goroutine
func (ipc *IPCClient) listenLoop() {
	defer func() {
		ipc.mu.Lock()
		if ipc.listener != nil {
			ipc.listener.Close()
		}
		close(ipc.messageChan)
		ipc.running = false
		ipc.mu.Unlock()
	}()

	for {
		select {
		case <-ipc.stopChan:
			fmt.Println("IPC client shutting down")
			return
		default:
			conn, err := ipc.listener.Accept()
			if err != nil {
				select {
				case <-ipc.stopChan:
					return
				default:
					log.Printf("Accept error: %v", err)
					continue
				}
			}
			go ipc.handleConnection(conn)
		}
	}
}

// handleConnection processes an individual connection
func (ipc *IPCClient) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, ipc.bufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	// Create response channel
	responseCh := make(chan string, 1)

	message := IPCMessage{
		Content:    string(buf[:n]),
		Connection: conn,
		Time:       time.Now(),
		ResponseCh: responseCh,
	}

	// Send message to channel (non-blocking)
	select {
	case ipc.messageChan <- message:
		// Wait for response
		select {
		case response := <-responseCh:
			_, err := conn.Write([]byte(response))
			if err != nil {
				log.Printf("Write error: %v", err)
			}
		}
	default:
		log.Printf("Message channel full, dropping connection")
		conn.Write([]byte("ERROR: Server busy"))
	}
}

// Messages returns the read-only message channel
func (ipc *IPCClient) Messages() <-chan IPCMessage {
	return ipc.messageChan
}

// Stop gracefully shuts down the IPC client
func (ipc *IPCClient) Stop() {
	ipc.mu.RLock()
	if !ipc.running {
		ipc.mu.RUnlock()
		return
	}
	ipc.mu.RUnlock()

	close(ipc.stopChan)

	// Clean up socket file on Unix systems
	if runtime.GOOS != "windows" {
		os.Remove(ipc.path)
	}
}

// IsRunning returns whether the client is currently running
func (ipc *IPCClient) IsRunning() bool {
	ipc.mu.RLock()
	defer ipc.mu.RUnlock()
	return ipc.running
}

// GetPath returns the IPC path
func (ipc *IPCClient) GetPath() string {
	ipc.mu.RLock()
	defer ipc.mu.RUnlock()
	return ipc.path
}

// SendToIPC is a utility function to send a message to an IPC server
func SendToIPC(path, message string) (string, error) {
	var conn net.Conn
	var err error

	if runtime.GOOS == "windows" {
		conn, err = net.Dial("pipe", path)
	} else {
		conn, err = net.Dial("unix", path)
	}

	if err != nil {
		return "", fmt.Errorf("failed to connect to IPC: %v", err)
	}
	defer conn.Close()

	// Send message
	_, err = conn.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	// Read response
	buf := make([]byte, DefaultBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(buf[:n]), nil
}
