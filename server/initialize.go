package server

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func Initialize(port int, multicore bool, healthyNode string, membershipPort int, grpcPort int, numTokens int, dbLog string) {
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	log.SetOutput(os.Stdout)

	serverConfig := ServerConfig{
		serverPort:     port,
		membershipPort: membershipPort,
		grpcPort:       grpcPort,
		multicore:      multicore,
		healthyNode:    healthyNode,
		numTokens:      numTokens,
		dbLog:          dbLog,
	}

	InitServer(serverConfig)
}
