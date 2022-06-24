package persistence

import (
	"bufio"
	"context"
	"os"
	"strings"

	"distributed-cache.io/cache"
	log "github.com/sirupsen/logrus"
)

type restoreManager struct {
	loggerConfig LoggerConfig
	cache        *cache.CacheService
}

func (r restoreManager) restore() {
	log.Infof("Initializing DB restore from %s", r.loggerConfig.FilePath)
	f, err := os.Open(r.loggerConfig.FilePath)
	if err != nil {
		log.Errorf("Unable to restore - File not found %s", r.loggerConfig.FilePath)
	} else {
		log.Infof("Started reading DB data")
		scanner := bufio.NewScanner(f)
		ctx, cancel := context.WithCancel(context.Background())
		for scanner.Scan() {
			m := strings.Split(scanner.Text(), ",")
			value := &cache.CachePutRequest{Key: m[0], Value: m[1]}
			r.cache.PutIntoCache(ctx, value)
		}
		defer cancel()
		log.Infof("Completed DB restoration")
	}
	defer f.Close()
}
