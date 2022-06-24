package persistence

import (
	"distributed-cache.io/cache"
	log "github.com/sirupsen/logrus"
)

type PersistenceMessage struct {
	Key   string
	Value string
}

type PersistenceConfig struct {
	LoggerConfig   *LoggerConfig
	loggingChannel chan LogMessage
}

type PersistenceManager struct {
	Config PersistenceConfig
	Cache  *cache.CacheService
}

func (p *PersistenceManager) Initialize() {
	commitLogger := &CommitLogger{
		config: p.Config.LoggerConfig,
	}

	restoreManager := restoreManager{
		loggerConfig: *p.Config.LoggerConfig,
		cache:        p.Cache,
	}

	restoreManager.restore()
	commitLogger.Initialize()
	logChannel := make(chan LogMessage)
	go func() {
		log.Infof("Initializing Persistence Logging listener")
		for message := range logChannel {
			commitLogger.Write(message)
		}
	}()
	p.Config.loggingChannel = logChannel
}

func (p *PersistenceManager) WriteLog(message PersistenceMessage) {
	p.Config.loggingChannel <- LogMessage{
		message: message.Key + "," + message.Value + "\n",
	}
}
