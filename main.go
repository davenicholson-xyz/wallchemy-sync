package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	indentifier := fmt.Sprintf("%s-%06d", hostname, rand.Intn(100000))

	listener := network.NewMulticastListener(indentifier)
	listener.Start()

}
