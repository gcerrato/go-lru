package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type UserData struct {
	ID   int
	Name string
	Age  int
}

func TestLRUCache(t *testing.T) {

	t.Run("LRU cache: Has", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("returns false for a missing key and true for an existing one", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 1000})
			assert.Equal(t, false, lruCache.Has("user123"), "Should find key 'user123' missing")
			lruCache.Set("user123", UserData{ID: 1, Name: "Alice", Age: 30})
			assert.Equal(t, true, lruCache.Has("user123"), "Should find key 'user123' present")
		})

		t.Run("considers evicted keys as absent", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 2000})
			lruCache.Set("user123", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user456", UserData{ID: 2, Name: "Bob", Age: 25})
			assert.Equal(t, false, lruCache.Has("user123"), "Expecting key 'user123' to not be in the cache")
			assert.Equal(t, true, lruCache.Has("user456"), "Expecting key 'user456' to be in the cache")
		})

		t.Run("treats expired keys as nonexistent", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 250})
			lruCache.Set("user789", UserData{ID: 3, Name: "Charlie", Age: 40})
			assert.Equal(t, true, lruCache.Has("user789"), "Key 'user789' should exist before expiry")

			time.Sleep(300 * time.Millisecond)
			assert.Equal(t, false, lruCache.Has("user789"), "Key 'user789' should be gone after expiry")
		})

		t.Run("acknowledges multiple existing keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 5, TTL: 2000})
			lruCache.Set("user111", UserData{ID: 4, Name: "Dan", Age: 35})
			lruCache.Set("user222", UserData{ID: 5, Name: "Eve", Age: 28})
			assert.Equal(t, true, lruCache.Has("user111"), "Key 'user111' should be found")
			assert.Equal(t, true, lruCache.Has("user222"), "Key 'user222' should be found")
		})

		t.Run("considers key valid after TTL extension", func(t *testing.T) {
			cacheProvider := InMemoryLRUCacheProvider[UserData]{}

			t.Run("key remains valid after extending TTL", func(t *testing.T) {
				lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 250})
				lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})

				time.Sleep(200 * time.Millisecond)
				assert.Equal(t, true, lruCache.Has("user1"), "Key 'user1' should still be valid before TTL expiry")

				time.Sleep(200 * time.Millisecond)
				assert.Equal(t, true, lruCache.Has("user1"), "Key 'user1' should be renewed after further access")
			})
		})
	})

	t.Run("LRU cache: Get", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("fetches no value and returns error for missing key, fetches value for present key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 1500})
			value, err := lruCache.Get("user001")
			assert.Error(t, err, "An error should occur for missing key 'user001'")
			assert.Empty(t, value, "Returned value should be null for missing key 'user001'")

			lruCache.Set("user001", UserData{ID: 6, Name: "Frank", Age: 22})
			value, err = lruCache.Get("user001")
			assert.NoError(t, err, "No error expected for existing key 'user001'")
			assert.Equal(t, UserData{ID: 6, Name: "Frank", Age: 22}, value, "Value for 'user001' should be 'Frank'")
		})

		t.Run("provides no value and error for evicted keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 2000})
			lruCache.Set("user002", UserData{ID: 7, Name: "Grace", Age: 29})
			lruCache.Set("user003", UserData{ID: 8, Name: "Hank", Age: 32})
			value, err := lruCache.Get("user002")
			assert.Error(t, err, "An error should occur for evicted key 'user002'")
			assert.Empty(t, value, "Returned value should be null for evicted key 'user002'")

			value, err = lruCache.Get("user003")
			assert.NoError(t, err, "No error expected for existing key 'user003'")
			assert.Equal(t, UserData{ID: 8, Name: "Hank", Age: 32}, value, "Expected value should be 'Hank' for 'user003'")
		})

		t.Run("fetches no value and returns error for expired keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 800})
			lruCache.Set("user004", UserData{ID: 9, Name: "Ivy", Age: 21})
			value, err := lruCache.Get("user004")
			assert.NoError(t, err, "No error expected for key 'user004' before expiry")
			assert.Equal(t, UserData{ID: 9, Name: "Ivy", Age: 21}, value, "Expected value should be 'Ivy' for 'user004' before expiration")

			time.Sleep(900 * time.Millisecond)
			value, err = lruCache.Get("user004")
			assert.Error(t, err, "An error should occur for expired key 'user004'")
			assert.Empty(t, value, "Returned value should be null for expired key 'user004'")
		})

		t.Run("fetches correct values for existing multiple keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 5, TTL: 3000})
			lruCache.Set("user005", UserData{ID: 10, Name: "Jake", Age: 33})
			lruCache.Set("user006", UserData{ID: 11, Name: "Kara", Age: 27})

			value, err := lruCache.Get("user005")
			assert.NoError(t, err, "No error expected for existing key 'user005'")
			assert.Equal(t, UserData{ID: 10, Name: "Jake", Age: 33}, value, "Expected value should be 'Jake' for 'user005'")

			value, err = lruCache.Get("user006")
			assert.NoError(t, err, "No error expected for existing key 'user006'")
			assert.Equal(t, UserData{ID: 11, Name: "Kara", Age: 27}, value, "Expected value should be 'Kara' for 'user006'")
		})

		t.Run("fetches correct value for a key with extended TTL", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 4, TTL: 250})
			lruCache.Set("user007", UserData{ID: 12, Name: "Liam", Age: 23})

			time.Sleep(200 * time.Millisecond)
			value, err := lruCache.Get("user007")
			assert.NoError(t, err, "No error expected for key 'user007' before expiration")
			assert.Equal(t, UserData{ID: 12, Name: "Liam", Age: 23}, value, "Expected value should be 'Liam' for 'user007' before expiration")

			time.Sleep(100 * time.Millisecond)
			lruCache.Set("user007", UserData{ID: 12, Name: "Liam", Age: 23})

			time.Sleep(150 * time.Millisecond)
			value, err = lruCache.Get("user007")
			assert.NoError(t, err, "No error expected for key 'user007' after TTL extension")
			assert.Equal(t, UserData{ID: 12, Name: "Liam", Age: 23}, value, "Expected value should be 'Liam' for 'user007' after TTL extension")
		})
	})

	t.Run("LRU cache: set with UserData", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("inserts a record into cache", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value)

			value, err = lruCache.Get("user2")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value)
		})

		t.Run("overwrites an existing record for the same key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice Updated", Age: 31})

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice Updated", Age: 31}, value)
		})

		t.Run("updates existing record and extends TTL", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 250})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})

			time.Sleep(200 * time.Millisecond)
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice Updated", Age: 31})
			time.Sleep(200 * time.Millisecond)

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice Updated", Age: 31}, value)
		})

		t.Run("inserts record and removes least recently used entry", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 2, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})

			value, err := lruCache.Get("user1")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user2")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value)

			value, err = lruCache.Get("user3")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 3, Name: "Charlie", Age: 35}, value)
		})

		t.Run("accounts for 'get' operation in least recently used strategy", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			time.Sleep(100 * time.Millisecond)
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			time.Sleep(100 * time.Millisecond)
			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value)

			lruCache.Set("user4", UserData{ID: 4, Name: "Dave", Age: 40})
			time.Sleep(100 * time.Millisecond)

			value, err = lruCache.Get("user2")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value)

			value, err = lruCache.Get("user3")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 3, Name: "Charlie", Age: 35}, value)

			value, err = lruCache.Get("user4")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 4, Name: "Dave", Age: 40}, value)
		})

		t.Run("includes 'has' operation in least recently used strategy", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})

			assert.True(t, lruCache.Has("user1"))

			lruCache.Set("user4", UserData{ID: 4, Name: "Dave", Age: 40})

			value, err := lruCache.Get("user2")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value)

			value, err = lruCache.Get("user3")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 3, Name: "Charlie", Age: 35}, value)

			value, err = lruCache.Get("user4")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 4, Name: "Dave", Age: 40}, value)
		})

		t.Run("considers 'set' operation in least recently used strategy", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice Updated", Age: 31})
			lruCache.Set("user4", UserData{ID: 4, Name: "Dave", Age: 40})

			value, err := lruCache.Get("user2")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user3")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 3, Name: "Charlie", Age: 35}, value)

			value, err = lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice Updated", Age: 31}, value)

			value, err = lruCache.Get("user4")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 4, Name: "Dave", Age: 40}, value)
		})

		t.Run("allows new record insertion without evicting when old records expire due to TTL", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 2, TTL: 250})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})

			time.Sleep(300 * time.Millisecond)
			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})

			value, err := lruCache.Get("user1")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user2")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("user3")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 3, Name: "Charlie", Age: 35}, value)
		})
	})

	t.Run("Arbitrary operations with UserData", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("executes correctly for mixed cache operations", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 250})

			value, err := lruCache.Get("user1")
			assert.Error(t, err, "Expecting error for non-existing key 'user1'")
			assert.Empty(t, value)

			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			assert.False(t, lruCache.Has("user2"), "Should not contain key 'user2'")

			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			value, err = lruCache.Get("user1")
			assert.NoError(t, err, "No error should occur for existing key 'user1'")
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value, "Value for 'user1' should be 'Alice'")

			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "No error should occur for existing key 'user2'")
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value, "Value for 'user2' should be 'Bob'")

			time.Sleep(200 * time.Millisecond)
			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "No error should occur for key 'user2' after 200ms")
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value)

			time.Sleep(200 * time.Millisecond)
			assert.False(t, lruCache.Has("user1"), "Key 'user1' should be expired")
			assert.True(t, lruCache.Has("user2"), "Key 'user2' should not be expired")

			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})
			lruCache.Set("user1", UserData{ID: 4, Name: "Alice Updated", Age: 31})
			lruCache.Set("user2", UserData{ID: 5, Name: "Bob Updated", Age: 26})
			lruCache.Set("user4", UserData{ID: 6, Name: "Dave", Age: 40})

			value, err = lruCache.Get("user3")
			assert.Error(t, err, "Expecting error for evicted key 'user3'")
			assert.Empty(t, value)

			value, err = lruCache.Get("user1")
			assert.NoError(t, err, "No error should occur for key 'user1'")
			assert.Equal(t, UserData{ID: 4, Name: "Alice Updated", Age: 31}, value, "Value for 'user1' should be 'Alice Updated'")

			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "No error should occur for key 'user2'")
			assert.Equal(t, UserData{ID: 5, Name: "Bob Updated", Age: 26}, value, "Value for 'user2' should be 'Bob Updated'")

			value, err = lruCache.Get("user4")
			assert.NoError(t, err, "No error should occur for key 'user4'")
			assert.Equal(t, UserData{ID: 6, Name: "Dave", Age: 40}, value, "Value for 'user4' should be 'Dave'")
		})
	})
}
