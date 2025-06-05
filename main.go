package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/davenicholson-xyz/wallchemy-sync/daemonize"
	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func main() {
	_, err := exec.LookPath("wallchemy")
	if err != nil {
		log.Fatal("wallchemy not found in path")
	}

	fg := flag.Bool("fg", false, "start in foreground")
	port := flag.Int("port", 9999, "port")
	flag.Parse()

	if !*fg {
		if !daemonize.Start() {
			return
		}
		log.Println("Running as daemon")
	}

	udp := network.NewMulticastClient(network.UDPConfig{Port: *port})
	udp.Start()
	defer udp.Stop()

	ipc, _ := network.NewIPCClient(network.IPCConfig{AppName: "wallchemy"})
	ipc.Start()
	defer ipc.Stop()

	var wg sync.WaitGroup
	wg.Add(2)

	stopChan := make(chan struct{})

	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			case msg, ok := <-ipc.Messages():
				if !ok {
					return
				}
				fmt.Printf("[IPC] Received %s\n", msg.Content)

				if err := udp.Broadcast(msg.Content); err != nil {
					log.Printf("Failed to broadcast IPC message: %v", err)
					msg.ResponseCh <- "ERROR: Broadcast failed"
				} else {
					msg.ResponseCh <- "OK: Message broadcasted"
				}

			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			case msg, ok := <-udp.Messages():
				if !ok {
					return
				}
				fmt.Printf("[UDP] Received from %s: %s\n", msg.Sender.String(), msg.Content)

				cmd := exec.Command("wallchemy", "-fromsync", "-id", msg.Content)
				_, err := cmd.Output()
				if err != nil {
					log.Fatal(err)
				}

			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")

	close(stopChan)

	wg.Wait()

	log.Println("Shutdown complete")

}
