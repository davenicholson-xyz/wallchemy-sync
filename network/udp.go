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

type UDPMessage struct {
	Content string
	Sender  *net.UDPAddr
	Time    time.Time
}

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

type UDPConfig struct {
	MulticastAddress string
	Port             int
	FilterSelf       bool
	DatagramSize     int
	ChannelBuffer    int
}

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

func (mc *MulticastClient) listenLoop() {
	defer func() {
		mc.mu.Lock()
		if mc.conn != nil {
			mc.conn.Close()
		}
		close(mc.messageChan)
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

			if mc.filterSelf && mc.localAddr != nil && src.IP.Equal(mc.localAddr.IP) {
				continue
			}

			message := UDPMessage{
				Content: string(buffer[:n]),
				Sender:  src,
				Time:    time.Now(),
			}

			select {
			case mc.messageChan <- message:

			default:

				log.Printf("Warning: message channel full, dropping message from %s", src.String())
			}
		}
	}
}

func (mc *MulticastClient) Messages() <-chan UDPMessage {
	return mc.messageChan
}

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

func (mc *MulticastClient) Stop() {
	mc.mu.RLock()
	if !mc.running {
		mc.mu.RUnlock()
		return
	}
	mc.mu.RUnlock()

	close(mc.stopChan)
}

func (mc *MulticastClient) IsRunning() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.running
}

func (mc *MulticastClient) GetLocalAddr() *net.UDPAddr {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.localAddr
}
