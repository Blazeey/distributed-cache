package server

import "distributed-cache.io/common"

func Initialize(port int, multicore bool) {
	c := common.InitCache()

	InitServer(9000, true, func(message Message) (response Response) {
		if message.Op == "GET" {
			response.Message = c.Get(message.Key)
		} else if message.Op == "PUT" {
			c.Put(message.Key, message.Value)
			response.Message = "OK"
		}
		response.Code = 200
		return
	})

}
