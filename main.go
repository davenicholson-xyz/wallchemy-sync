package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/davenicholson-xyz/wallchemy-sync/app"
	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func main() {
	_, err := exec.LookPath("wallchemy")
	if err != nil {
		log.Fatal("wallchemy not found installed")
	}

	port := flag.Int("port", 9999, "port")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	indentifier := fmt.Sprintf("%s-%06d", hostname, rand.Intn(100000))

	app := app.NewApp(*port, indentifier)

	//TODO: Go to a start function and run these.

	udpHandler := func(msg string, src *net.UDPAddr) {
		fmt.Printf("[UDP][%s] From %s: %s\n", indentifier, src.IP, msg)
	}

	udp := network.NewMulticastListener(app.Port, app.Identifier, udpHandler)
	udp.Start()
	defer udp.Stop()

	ipcHandler := func(msg string) string {
		msg = strings.TrimSpace(msg)
		fmt.Printf("%s\n", msg)
		return ""
	}

	ipc := network.NewIPCListener(ipcHandler)
	ipc.Start()
	defer ipc.Stop()

	select {}
}
