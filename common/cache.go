package common

import "sync"

type CacheValue struct {
	value string
}

type Cache struct {
	mutex sync.RWMutex
	m     map[string]*Node
	dll   DoublyLinkedList
}

func InitCache() *Cache {
	var cache Cache
	cache.m = make(map[string]*Node)
	return &cache
}

func (cache *Cache) Put(key string, value string) bool {
	cache.mutex.Lock()
	cache.m[key] = cache.dll.AddToFront(&CacheValue{value})
	cache.mutex.Unlock()
	return true
}

func (cache *Cache) Get(key string) string {
	cache.mutex.RLock()
	value := cache.m[key]
	if value == nil {
		return ""
	}
	cache.mutex.RUnlock()
	return value.value.value
}
