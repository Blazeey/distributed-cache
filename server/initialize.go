package server

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"distributed-cache.io/common"
)

func Initialize(port int, multicore bool) {
	c := common.InitCache()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go InitServer(9000, true, func(message Message) (response Response) {
		log.Printf("PROCESSING MESSAGE : %s", message)
		if message.Op == "GET" {
			response.Message = c.Get(message.Key)
		} else if message.Op == "PUT" {
			c.Put(message.Key, message.Value)
			response.Message = "OK"
		}
		response.Code = 200
		return
	})
	wg.Wait()
}
