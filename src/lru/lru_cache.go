package lru

import (
	"errors"
	"fmt"
	"time"
)

type LRUCacheConfig struct {
	ItemLimit int64
	TTL       int64
}

type message[T any] struct {
	value  T
	key    string
	action string
}

type LRUCache[T any] interface {
	Has(key string) bool
	Get(key string) (T, error)
	Set(key string, value T) T
}

type StorageItem[T any] struct {
	Value    T
	DeleteAt time.Time
}

func (item *StorageItem[T]) bumpDeleteAt(ms int64) *StorageItem[T] {
	item.DeleteAt = time.Now().Add(time.Duration(ms) * time.Millisecond)
	return item
}

type InMemoryLRUCache[T any] struct {
	Config  LRUCacheConfig
	Storage map[string]*StorageItem[T]

	// channel operations
	messageChannel chan message[T]
}

func (cache InMemoryLRUCache[T]) Has(key string) bool {
	storageItem, exists := cache.Storage[key]
	if !exists {
		return false
	}
	storageItem.bumpDeleteAt(cache.Config.TTL)
	return exists
}

func (cache InMemoryLRUCache[T]) Get(key string) (T, error) {
	storageItem, exists := cache.Storage[key]
	var zero T
	if !exists {
		return zero, errors.New("key not found on LRU cache")
	}
	storageItem.bumpDeleteAt(cache.Config.TTL)
	return storageItem.Value, nil
}

func (cache InMemoryLRUCache[T]) Set(key string, value T) T {
	cache.messageChannel <- message[T]{key: key, value: value, action: "add"}
	<-cache.messageChannel
	return value
}

func (cache InMemoryLRUCache[T]) sweepKeys() {
	now := time.Now()
	for key, value := range cache.Storage {
		diff := now.Sub(value.DeleteAt).Milliseconds()
		if diff >= 0 {
			fmt.Printf("deleted key automatically %s with diff %d \n", key, diff)
			delete(cache.Storage, key)
		}
	}
}

func (cache InMemoryLRUCache[T]) removeOldestKey() {
	var oldestKey string
	var oldestTime time.Time

	// Iterate over all keys to find the oldest one
	for key, value := range cache.Storage {
		// If this is the first key or an older key, update the oldest key tracker
		if oldestKey == "" || value.DeleteAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = value.DeleteAt
		}
	}

	// Remove the oldest key if it exists and is due for deletion
	if oldestKey != "" {
		fmt.Printf("deleted oldest key %s \n", oldestKey)
		delete(cache.Storage, oldestKey)
	}
}

func (cache *InMemoryLRUCache[T]) startMessageListener(interval time.Duration) {
	cache.messageChannel = make(chan message[T])
	go func() {
		for {
			time.Sleep(interval)
			cache.messageChannel <- message[T]{action: "sweep"}
		}
	}()
	go func() {
		for msg := range cache.messageChannel {
			if msg.action == "add" {
				cache.setFromChannel(msg.key, msg.value)
				cache.messageChannel <- message[T]{action: "done"}
			} else if msg.action == "sweep" {
				cache.sweepKeys()
			}
		}
	}()
}

func (cache InMemoryLRUCache[T]) setFromChannel(key string, value T) T {

	if int64(len(cache.Storage)) >= cache.Config.ItemLimit {
		cache.removeOldestKey()
	}

	cache.Storage[key] = &StorageItem[T]{Value: value, DeleteAt: time.Now().Add(time.Duration(cache.Config.TTL) * time.Millisecond)}
	return value
}

type LRUCacheProvider[T any] interface {
	NewLRUCache(config LRUCacheConfig) LRUCache[T]
}

// e.g other
// type RedisCacheProvider[T any] struct{}
type InMemoryLRUCacheProvider[T any] struct{}

func (cacheProvider InMemoryLRUCacheProvider[T]) NewLRUCache(config LRUCacheConfig) LRUCache[T] {
	cache := InMemoryLRUCache[T]{Config: config, Storage: make(map[string]*StorageItem[T])}
	cache.startMessageListener(50 * time.Millisecond)
	return cache
}
