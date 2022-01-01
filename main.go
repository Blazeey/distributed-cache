package main

import (
	"flag"

	"distributed-cache.io/server"
)

func main() {
	serverPort := flag.Int("port", 9000, "Server Port")
	healthyNode := flag.String("node", "", "Healthy Node")
	membershipPort := flag.Int("membershipPort", 11000, "Membership Port")
	flag.Parse()
	server.Initialize(*serverPort, true, *healthyNode, *membershipPort)
}
