package lru

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type LRUCacheConfig struct {
	ItemLimit int64
	TTL       int64
}

type LRUCacher[T any] interface {
	Has(key string) bool
	Get(key string) (T, error)
	Set(key string, value T) T
}

type StorageItem[T any] struct {
	Value    T
	DeleteAt time.Time
}

type SafeMap[T any] struct {
	SafeMap map[string]*StorageItem[T]
	mu      sync.RWMutex
}

func NewSafeMap[T any]() *SafeMap[T] {
	return &SafeMap[T]{SafeMap: make(map[string]*StorageItem[T])}
}

func (item *StorageItem[T]) bumpDeleteAt(ms int64) *StorageItem[T] {
	item.DeleteAt = time.Now().Add(time.Duration(ms) * time.Millisecond)
	return item
}

type InMemoryLRUCache[T any] struct {
	Config  LRUCacheConfig
	Storage *SafeMap[T]
}

func (cache *InMemoryLRUCache[T]) Has(key string) bool {
	cache.Storage.mu.Lock()
	defer cache.Storage.mu.Unlock()
	storageItem, exists := cache.Storage.SafeMap[key]
	if !exists {
		return false
	}
	storageItem.bumpDeleteAt(cache.Config.TTL)
	return exists
}

func (cache *InMemoryLRUCache[T]) Get(key string) (T, error) {
	cache.Storage.mu.Lock()
	defer cache.Storage.mu.Unlock()
	storageItem, exists := cache.Storage.SafeMap[key]
	var zero T
	if !exists {
		return zero, errors.New("key not found on LRU cache")
	}
	storageItem.bumpDeleteAt(cache.Config.TTL)
	return storageItem.Value, nil
}

func (cache *InMemoryLRUCache[T]) Set(key string, value T) T {
	cache.Storage.mu.Lock()
	defer cache.Storage.mu.Unlock()
	if int64(len(cache.Storage.SafeMap)) >= cache.Config.ItemLimit {
		cache.removeOldestKey()
	}

	cache.Storage.SafeMap[key] = &StorageItem[T]{Value: value, DeleteAt: time.Now().Add(time.Duration(cache.Config.TTL) * time.Millisecond)}
	return value
}

func (cache *InMemoryLRUCache[T]) sweepKeys() {
	now := time.Now()
	for key, value := range cache.Storage.SafeMap {
		diff := now.Sub(value.DeleteAt).Milliseconds()
		if diff >= 0 {
			fmt.Printf("deleted key automatically %s with diff %d \n", key, diff)
			delete(cache.Storage.SafeMap, key)
		}
	}
}

func (cache *InMemoryLRUCache[T]) removeOldestKey() {
	var oldestKey string
	var oldestTime time.Time

	for key, value := range cache.Storage.SafeMap {
		if oldestKey == "" || value.DeleteAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = value.DeleteAt
		}
	}

	if oldestKey != "" {
		fmt.Printf("deleted oldest key %s \n", oldestKey)
		delete(cache.Storage.SafeMap, oldestKey)
	}
}

func (cache *InMemoryLRUCache[T]) startMessageListener(interval time.Duration) {
	for {
		time.Sleep(interval)
		cache.sweepKeys()
	}
}

type LRUCacheProvider[T any] interface {
	NewLRUCache(config LRUCacheConfig) LRUCacher[T]
}

// e.g other
// type RedisCacheProvider[T any] struct{}
type InMemoryLRUCacheProvider[T any] struct{}

func (cacheProvider InMemoryLRUCacheProvider[T]) NewLRUCache(config LRUCacheConfig) LRUCacher[T] {
	safeMap := NewSafeMap[T]()
	cache := InMemoryLRUCache[T]{Config: config, Storage: safeMap}
	go cache.startMessageListener(50 * time.Millisecond)
	return &cache
}
