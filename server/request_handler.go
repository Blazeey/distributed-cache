package server

import (
	"google.golang.org/grpc"
)

type RequestHandler struct {
	connections *map[string]*grpc.ClientConn
}

func ProcessMessage(message Message) (response Response) {

	key := message.Key
	hash := murmur3(key)
	tokenRing.getAssignedNode(hash)
	// grpc.Dial()
	// log.Printf("PROCESSING MESSAGE : %s", message)
	// if message.Op == "GET" {
	// 	response.Message = c.Get(message.Key)
	// } else if message.Op == "PUT" {
	// 	c.Put(message.Key, message.Value)
	// 	response.Message = "OK"
	// }
	response.Code = 200
	return
}
