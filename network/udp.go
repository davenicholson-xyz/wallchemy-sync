package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	DefaultMulticastAddress = "239.192.0.1"
	DefaultDatagramSize     = 8192
)

// UDPMessage represents a received multicast message
type UDPMessage struct {
	Content string
	Sender  *net.UDPAddr
	Time    time.Time
}

// MulticastClient provides multicast UDP communication capabilities
type MulticastClient struct {
	multicastAddr string
	port          int
	conn          *net.UDPConn
	stopChan      chan struct{}
	messageChan   chan UDPMessage
	localAddr     *net.UDPAddr
	filterSelf    bool
	datagramSize  int
	mu            sync.RWMutex
	running       bool
}

// UDPConfig holds configuration options for the MulticastClient
type UDPConfig struct {
	MulticastAddress string // Default: DefaultMulticastAddress
	Port             int    // Required
	FilterSelf       bool   // Whether to ignore messages from self
	DatagramSize     int    // Default: DefaultDatagramSize
	ChannelBuffer    int    // Default: DefaultChannelBuffer
}

// NewMulticastClient creates a new multicast client with the given configuration
func NewMulticastClient(config UDPConfig) *MulticastClient {
	if config.MulticastAddress == "" {
		config.MulticastAddress = DefaultMulticastAddress
	}
	if config.DatagramSize == 0 {
		config.DatagramSize = DefaultDatagramSize
	}
	if config.ChannelBuffer == 0 {
		config.ChannelBuffer = DefaultChannelBuffer
	}

	return &MulticastClient{
		multicastAddr: config.MulticastAddress,
		port:          config.Port,
		filterSelf:    config.FilterSelf,
		datagramSize:  config.DatagramSize,
		stopChan:      make(chan struct{}),
		messageChan:   make(chan UDPMessage, config.ChannelBuffer),
	}
}

// Start begins listening for multicast messages
func (mc *MulticastClient) Start() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return fmt.Errorf("multicast client already running")
	}

	addrStr := fmt.Sprintf("%s:%d", mc.multicastAddr, mc.port)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to listen on multicast UDP: %v", err)
	}

	mc.conn = conn
	mc.localAddr = conn.LocalAddr().(*net.UDPAddr)
	mc.running = true

	if err := conn.SetReadBuffer(mc.datagramSize); err != nil {
		log.Printf("Warning: SetReadBuffer failed: %v", err)
	}

	fmt.Printf("Multicast client listening on %s\n", addrStr)
	go mc.listenLoop()

	return nil
}

// listenLoop runs the main listening loop in a goroutine
func (mc *MulticastClient) listenLoop() {
	defer func() {
		mc.mu.Lock()
		if mc.conn != nil {
			mc.conn.Close()
		}
		close(mc.messageChan) // Close channel when done
		mc.running = false
		mc.mu.Unlock()
	}()

	buffer := make([]byte, mc.datagramSize)

	for {
		select {
		case <-mc.stopChan:
			fmt.Println("Multicast client shutting down")
			return
		default:
			n, src, err := mc.conn.ReadFromUDP(buffer)
			if err != nil {
				select {
				case <-mc.stopChan:
					return
				default:
					log.Printf("ReadFromUDP failed: %v", err)
					continue
				}
			}

			// Filter self messages if enabled
			if mc.filterSelf && mc.localAddr != nil && src.IP.Equal(mc.localAddr.IP) {
				continue
			}

			message := UDPMessage{
				Content: string(buffer[:n]),
				Sender:  src,
				Time:    time.Now(),
			}

			// Send to channel with non-blocking send
			select {
			case mc.messageChan <- message:
				// Message sent successfully
			default:
				// Channel full, drop message
				log.Printf("Warning: message channel full, dropping message from %s", src.String())
			}
		}
	}
}

// Messages returns the read-only message channel
func (mc *MulticastClient) Messages() <-chan UDPMessage {
	return mc.messageChan
}

// Broadcast sends a message to the multicast group
func (mc *MulticastClient) Broadcast(message string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if !mc.running || mc.conn == nil {
		return fmt.Errorf("multicast client not running")
	}

	addr := &net.UDPAddr{
		IP:   net.ParseIP(mc.multicastAddr),
		Port: mc.port,
	}

	_, err := mc.conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		return fmt.Errorf("failed to broadcast message: %v", err)
	}

	return nil
}

// Stop gracefully shuts down the multicast client
func (mc *MulticastClient) Stop() {
	mc.mu.RLock()
	if !mc.running {
		mc.mu.RUnlock()
		return
	}
	mc.mu.RUnlock()

	close(mc.stopChan)
}

// IsRunning returns whether the client is currently running
func (mc *MulticastClient) IsRunning() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.running
}

// GetLocalAddr returns the local address of the connection
func (mc *MulticastClient) GetLocalAddr() *net.UDPAddr {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.localAddr
}
