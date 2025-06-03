package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/davenicholson-xyz/wallchemy-sync/app"
	"github.com/davenicholson-xyz/wallchemy-sync/network"
)

func main() {

	port := flag.Int("port", 9999, "port")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	indentifier := fmt.Sprintf("%s-%06d", hostname, rand.Intn(100000))

	app := app.NewApp(*port, indentifier)
	fmt.Println(app)

	listener := network.NewMulticastListener(app.Port, app.Identifier)
	listener.Start()
}
