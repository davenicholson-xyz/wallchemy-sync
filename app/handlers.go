package app

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func HandleUDP(msg string, src *net.UDPAddr) {
	cmd := exec.Command("wallchemy", "-fromsync", "-id", msg)
	_, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
}

func HandleIPC(udp *network.MulticastListener) func(string) string {
	return func(msg string) string {
		msg = strings.TrimSpace(msg)
		udp.Broadcast(fmt.Sprintf("%s", msg), true)
		fmt.Printf("%s\n", msg)
		return ""
	}
}
