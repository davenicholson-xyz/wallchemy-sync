package network

import (
	"fmt"
	"log"
	"net"
)

const (
	multicastAddress = "239.192.0.1"
	maxDatagramSize  = 8192
)

type MulticastListener struct {
	identifier string
	port       int
	conn       *net.UDPConn
	stopChan   chan struct{}
	handler    func(string, *net.UDPAddr)
	localAddr  *net.UDPAddr
}

func NewMulticastListener(port int, id string, handlerFn func(string, *net.UDPAddr)) *MulticastListener {
	return &MulticastListener{
		identifier: id,
		port:       port,
		stopChan:   make(chan struct{}),
		handler:    handlerFn,
	}
}

func (ml *MulticastListener) Start() {
	addrStr := fmt.Sprintf("%s:%d", multicastAddress, ml.port)

	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		log.Fatal("ResolveUDPAddr fail:", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("ListenMulticastUDP fail:", err)
	}
	ml.conn = conn
	ml.localAddr = conn.LocalAddr().(*net.UDPAddr)

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("SetReadBuffer failed: %v", err)
	}

	fmt.Printf("Listening for multicast address on %s\n", addrStr)

	go ml.listenLoop()
}

func (ml *MulticastListener) listenLoop() {
	defer ml.conn.Close()

	for {
		select {
		case <-ml.stopChan:
			fmt.Printf("[%s] Multicast listener shutting down\n", ml.identifier)
			return
		default:
			buffer := make([]byte, maxDatagramSize)
			n, src, err := ml.conn.ReadFromUDP(buffer)
			if err != nil {
				select {
				case <-ml.stopChan:
					return
				default:
					log.Printf("[%s] ReadFromUDP failed: %v\n", ml.identifier, err)
				}
				continue
			}

			message := string(buffer[:n])
			ml.handler(message, src)
		}
	}
}

func (ml *MulticastListener) Stop() {
	close(ml.stopChan)
}
