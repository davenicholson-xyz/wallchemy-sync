package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/davenicholson-xyz/wallchemy-sync/app"
	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func main() {
	_, err := exec.LookPath("wallchemy")
	if err != nil {
		log.Fatal("wallchemy not found in path")
	}

	fg := flag.Bool("fg", false, "run fg")
	port := flag.Int("port", 9999, "port")
	flag.Parse()

	if !*fg {
		daemonize()
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	indentifier := fmt.Sprintf("%s-%06d", hostname, rand.Intn(100000))

	udp := network.NewMulticastListener(*port, indentifier, app.HandleUDP, false)
	udp.Start()
	defer udp.Stop()

	ipc := network.NewIPCListener(app.HandleIPC(udp))
	ipc.Start()
	defer ipc.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("wallchemy-sync started on port %d", *port)
	<-sigChan
	log.Println("Shutting down...")
}

func daemonize() {
	if os.Getppid() == 1 {
		return
	}

	if os.Getenv("DAEMON_CHILD") != "1" {
		os.Setenv("DAEMON_CHILD", "1")
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Start()
		os.Exit(0)
	}

	syscall.Setsid()

	os.Chdir("/")

	if f, err := os.OpenFile("/dev/null", os.O_RDWR, 0); err == nil {
		os.Stdin = f
		os.Stdout = f
		os.Stderr = f
	}
}
