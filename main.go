package main

import (
	"flag"

	"distributed-cache.io/server"
)

func main() {
	serverPort := flag.Int("port", 9000, "Server Port")
	healthyNode := flag.String("node", "", "Healthy Node")
	membershipPort := flag.Int("membershipPort", 11000, "Membership Port")
	grpcPort := flag.Int("grpcPort", 9050, "gRPC Port")
	numTokens := flag.Int("numTokens", 256, "Total token ranges")
	flag.Parse()
	server.Initialize(*serverPort, true, *healthyNode, *membershipPort, *grpcPort, *numTokens)
}
