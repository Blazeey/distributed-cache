package swim

import (
	"distributed-cache.io/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: Build a connection pool. Check if active TCP connection using WireShark?

type SwimClientConnectionPool struct {
	connections map[string]SwimClient
}

func (c SwimClientConnectionPool) getRemoteConnection(ip string, port uint16) (client SwimClient) {
	address := common.GetAddress(ip, port)
	if cachedClient, ok := c.connections[address]; ok {
		client = cachedClient
	} else {
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Cannot connect to %s", address)
		}
		client = NewSwimClient(conn)
		c.connections[address] = client
		return
	}
	return
}
