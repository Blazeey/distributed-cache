package persistence

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

type LoggerConfig struct {
	FilePath string
}

type LogMessage struct {
	message string
}

type Writer interface {
	Write(message LogMessage)
	Initialize()
}

type CommitLogger struct {
	mutex       sync.RWMutex
	config      *LoggerConfig
	currentFile *os.File
}

func (logger *CommitLogger) Initialize() {
	log.Infof("Initializing Commit Logger")
	file, err := os.Create(logger.config.FilePath)
	if err != nil {
		log.Errorf("Failed to create file %s", logger.config.FilePath, err)
		os.Exit(1)
	}
	logger.currentFile = file
	log.Infof("Commit Logger initialization complete - Writing to %s", logger.config.FilePath)
}

func (logger *CommitLogger) Write(logMessage LogMessage) {
	logger.mutex.Lock()
	n, err := logger.currentFile.WriteString(logMessage.message)
	if err != nil {
		log.Errorf("Failed to write to %s", logger.currentFile.Name(), err)
	}
	if n != len(logMessage.message) {
		log.Errorf("Failed to write data to %s", logger.currentFile.Name())
	}
	logger.mutex.Unlock()
}
