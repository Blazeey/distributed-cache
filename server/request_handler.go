package server

import (
	"context"
	"fmt"
	"time"

	"distributed-cache.io/cache"
	"distributed-cache.io/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestHandler struct {
	connections map[uint32]cache.CacheClient
	cache       *cache.CacheService
}

func (handler RequestHandler) processMessage(message Message) (response Response) {
	log.Printf("PROCESSING MESSAGE : %s", message)
	if message.Op == "GET" {
		return handler.processGetRequest(message)
	} else if message.Op == "PUT" {
		return handler.processPutRequest(message)
	}
	return Response{
		Message: "Invalid operation",
		Code:    404,
	}
}

func (handler RequestHandler) processGetRequest(message Message) (response Response) {
	isLocal, cacheClient := handler.getRequestConnectionDetails(message)
	request := &cache.CacheGetRequest{Key: message.Key}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cacheResponse *cache.CacheGetResponse
	var err error
	if isLocal {
		cacheResponse, err = handler.cache.GetFromCache(ctx, request)
	} else {
		cacheResponse, err = cacheClient.GetFromCache(ctx, request)
	}
	if err != nil {
		s := fmt.Sprintf("Unable to retrieve from cache for %s", message.Key)
		log.Fatalf(s, err)
		return Response{
			Message: s,
			Code:    500,
		}
	}
	return Response{
		Message: cacheResponse.Message,
		Code:    cacheResponse.Code,
	}
}

func (handler RequestHandler) processPutRequest(message Message) (response Response) {
	isLocal, cacheClient := handler.getRequestConnectionDetails(message)
	request := &cache.CachePutRequest{Key: message.Key, Value: message.Value}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cacheResponse *cache.CachePutResponse
	var err error
	if isLocal {
		cacheResponse, err = handler.cache.PutIntoCache(ctx, request)
	} else {
		cacheResponse, err = cacheClient.PutIntoCache(ctx, request)
	}
	if err != nil {
		s := fmt.Sprintf("Unable to put into cache for %s", message.Key)
		log.Fatalf(s)
		return Response{
			Message: s,
			Code:    500,
		}
	}
	return Response{
		Message: cacheResponse.Message,
		Code:    cacheResponse.Code,
	}
}

func (handler RequestHandler) getRequestConnectionDetails(message Message) (isLocal bool, cacheClient cache.CacheClient) {
	key := message.Key
	hash := common.Murmur3(key)
	log.Infoln("Hash for %s is %d", key, hash)
	nodeToRequest := tokenRing.getAssignedNode(hash)
	isLocal = nodeToRequest.IsCurrentNode
	if isLocal {
		return
	}
	if client, ok := handler.connections[hash]; ok {
		cacheClient = client
	} else {
		// Finding gRPC port using hacky method
		port := (nodeToRequest.Host.Port() % 10) + 9050
		ip := nodeToRequest.Host.IP().String()
		address := common.GetAddress(ip, port)
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Cannot connect to %s for %s", nodeToRequest.Host.Address(), key)
		}
		cacheClient = cache.NewCacheClient(conn)
		handler.connections[hash] = cacheClient
	}
	return
}
