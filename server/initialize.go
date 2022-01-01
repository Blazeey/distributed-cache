package server

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"distributed-cache.io/common"
)

func Initialize(port int, multicore bool, healthyNode string, membershipPort int) {
	c := common.InitCache()

	log.SetFormatter(&log.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	log.SetOutput(os.Stdout)

	processMessage := func(message Message) (response Response) {
		log.Printf("PROCESSING MESSAGE : %s", message)
		if message.Op == "GET" {
			response.Message = c.Get(message.Key)
		} else if message.Op == "PUT" {
			c.Put(message.Key, message.Value)
			response.Message = "OK"
		}
		response.Code = 200
		return
	}

	serverConfig := ServerConfig{
		serverPort:     port,
		membershipPort: membershipPort,
		multicore:      multicore,
		healthyNode:    healthyNode,
		callback:       processMessage,
	}

	InitServer(serverConfig)
}
