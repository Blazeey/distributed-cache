package cache

import (
	"context"

	"distributed-cache.io/common"
)

type CacheService struct {
	UnimplementedCacheServer
	Cache *common.Cache
}

func (c *CacheService) GetFromCache(ctx context.Context, in *CacheGetRequest) (*CacheGetResponse, error) {
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
	response := &CachePutResponse{
		Code: 500,
	}
	c.Cache.Put(in.Key, in.Value)
	response.Code = 200
	return response, nil
}
