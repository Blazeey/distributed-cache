package main

import (
	"distributed-cache.io/server"
)

func main() {
	server.Initialize(9000, true)
}
