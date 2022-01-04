package cache

import (
	"context"

	"distributed-cache.io/common"
	log "github.com/sirupsen/logrus"
)

type CacheService struct {
	UnimplementedCacheServer
	Cache *common.Cache
}

func (c *CacheService) GetFromCache(ctx context.Context, in *CacheGetRequest) (*CacheGetResponse, error) {
	log.Infof("GET request for %s", in.Key)
	response := &CacheGetResponse{}
	response.Message = c.Cache.Get(in.Key)
	if response.Message == "" {
		response.Code = 404
	} else {
		response.Code = 200
	}
	return response, nil
}

func (c *CacheService) PutIntoCache(ctx context.Context, in *CachePutRequest) (*CachePutResponse, error) {
	log.Infof("PUT request for %s", in.Key)
	response := &CachePutResponse{
		Code: 500,
	}
	c.Cache.Put(in.Key, in.Value)
	response.Code = 200
	return response, nil
}
