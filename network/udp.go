package network

import (
	"fmt"
	"log"
	"net"
)

const (
	multicastAddress = "239.192.0.1:9999"
	maxDatagramSize  = 8192
)

type MulticastListener struct {
	identifier string
	conn       *net.UDPConn
}

func NewMulticastListener(id string) *MulticastListener {
	return &MulticastListener{identifier: id}
}

func (ml *MulticastListener) Start() {
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		log.Fatal("ResolveUDPAddr fail:", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("ListenMulticastUDP fail:", err)
	}
	ml.conn = conn

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("SetReadBuffer failed: %v", err)
	}

	fmt.Printf("Listening for multicast address on %s\n", multicastAddress)

	ml.listenLoop()
}

func (ml *MulticastListener) listenLoop() {
	for {
		buffer := make([]byte, maxDatagramSize)

		n, src, err := ml.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}

		message := string(buffer[:n])
		ml.handleMessage(message, src)
	}
}

func (ml *MulticastListener) handleMessage(msg string, src *net.UDPAddr) {
	switch msg {
	case "HELLO":
		fmt.Printf("[%s] New connection from %s\n", ml.identifier, src.IP)
	case "GOODBYE":
		fmt.Printf("[%s] Connection closed from %s\n", ml.identifier, src.IP)
	default:
		fmt.Printf("[%s] Received from %s: %s\n", ml.identifier, src.IP, msg)
	}
}
