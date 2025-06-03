package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"

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

	listener := network.NewMulticastListener(app.Port, app.Identifier)
	listener.Start()
	defer listener.Stop()

	ipcListener, addr, err := network.NewIPCListener()
	if err != nil {
		log.Fatal("Failed to ceate listener:", err)
	}
	defer ipcListener.Close()

	if runtime.GOOS != "windows" {
		defer os.Remove(addr)
	}

	fmt.Printf("IPC listening on %s (%s)", addr, runtime.GOOS)

	for {
		conn, err := ipcListener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}

		go network.HandleConnection(conn)
	}

}
