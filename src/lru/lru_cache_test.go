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

		t.Run("should return false for non-existent key and true for existing", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 1000})
			assert.Equal(t, false, lruCache.Has("user123"), "Expected lruCache.Has(\"user123\") to be false")
			lruCache.Set("user123", UserData{ID: 1, Name: "Alice", Age: 30})
			assert.Equal(t, true, lruCache.Has("user123"), "Expected lruCache.Has(\"user123\") to be true")
		})

		t.Run("should return false for evicted key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 2000})
			lruCache.Set("user123", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user456", UserData{ID: 2, Name: "Bob", Age: 25})
			assert.Equal(t, false, lruCache.Has("user123"), "Expected lruCache.Has(\"user123\") to be false")
			assert.Equal(t, true, lruCache.Has("user456"), "Expected lruCache.Has(\"user456\") to be true")
		})

		t.Run("should return false for expired key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 500})
			lruCache.Set("user789", UserData{ID: 3, Name: "Charlie", Age: 40})
			assert.Equal(t, true, lruCache.Has("user789"), "Expected lruCache.Has(\"user789\") to be true before expiration")

			time.Sleep(600 * time.Millisecond)
			assert.Equal(t, false, lruCache.Has("user789"), "Expected lruCache.Has(\"user789\") to be false after expiration")
		})

		t.Run("should return true for multiple existing keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 5, TTL: 2000})
			lruCache.Set("user111", UserData{ID: 4, Name: "Dan", Age: 35})
			lruCache.Set("user222", UserData{ID: 5, Name: "Eve", Age: 28})
			assert.Equal(t, true, lruCache.Has("user111"), "Expected lruCache.Has(\"user111\") to be true")
			assert.Equal(t, true, lruCache.Has("user222"), "Expected lruCache.Has(\"user222\") to be true")
		})

		t.Run("should return true for key with TTL extended (UserData)", func(t *testing.T) {
			cacheProvider := InMemoryLRUCacheProvider[UserData]{}

			t.Run("should return true for key with TTL extended", func(t *testing.T) {
				lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 500})
				lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})

				time.Sleep(400 * time.Millisecond)
				assert.Equal(t, true, lruCache.Has("user1"), "Expected lruCache.Has(\"user1\") to be true before TTL expires")

				time.Sleep(400 * time.Millisecond)
				assert.Equal(t, true, lruCache.Has("user1"), "Expected lruCache.Has(\"user1\") to be true after further access")
			})
		})
	})

	t.Run("LRU cache: Get", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("should return empty and error for non-existent key, and value for existing key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 1500})
			value, err := lruCache.Get("user001")
			assert.Error(t, err, "Expected an error for non-existent key 'user001'")
			assert.Empty(t, value, "Expected value to be nil for non-existent key 'user001'")

			lruCache.Set("user001", UserData{ID: 6, Name: "Frank", Age: 22})
			value, err = lruCache.Get("user001")
			assert.NoError(t, err, "Did not expect an error for existing key 'user001'")
			assert.Equal(t, UserData{ID: 6, Name: "Frank", Age: 22}, value, "Expected lruCache.Get(\"user001\") to return 'Frank'")
		})

		t.Run("should return empty and error for evicted key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 2000})
			lruCache.Set("user002", UserData{ID: 7, Name: "Grace", Age: 29})
			lruCache.Set("user003", UserData{ID: 8, Name: "Hank", Age: 32})
			value, err := lruCache.Get("user002")
			assert.Error(t, err, "Expected an error for evicted key 'user002'")
			assert.Empty(t, value, "Expected value to be nil for evicted key 'user002'")

			value, err = lruCache.Get("user003")
			assert.NoError(t, err, "Did not expect an error for existing key 'user003'")
			assert.Equal(t, UserData{ID: 8, Name: "Hank", Age: 32}, value, "Expected lruCache.Get(\"user003\") to return 'Hank'")
		})

		t.Run("should return nil and error for expired key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 800})
			lruCache.Set("user004", UserData{ID: 9, Name: "Ivy", Age: 21})
			value, err := lruCache.Get("user004")
			assert.NoError(t, err, "Did not expect an error for key 'user004' before expiration")
			assert.Equal(t, UserData{ID: 9, Name: "Ivy", Age: 21}, value, "Expected lruCache.Get(\"user004\") to return 'Ivy' before expiration")

			time.Sleep(900 * time.Millisecond)
			value, err = lruCache.Get("user004")
			assert.Error(t, err, "Expected an error for expired key 'user004'")
			assert.Empty(t, value, "Expected value to be nil for expired key 'user004'")
		})

		t.Run("should return value for multiple existing keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 5, TTL: 3000})
			lruCache.Set("user005", UserData{ID: 10, Name: "Jake", Age: 33})
			lruCache.Set("user006", UserData{ID: 11, Name: "Kara", Age: 27})

			value, err := lruCache.Get("user005")
			assert.NoError(t, err, "Did not expect an error for existing key 'user005'")
			assert.Equal(t, UserData{ID: 10, Name: "Jake", Age: 33}, value, "Expected lruCache.Get(\"user005\") to return 'Jake'")

			value, err = lruCache.Get("user006")
			assert.NoError(t, err, "Did not expect an error for existing key 'user006'")
			assert.Equal(t, UserData{ID: 11, Name: "Kara", Age: 27}, value, "Expected lruCache.Get(\"user006\") to return 'Kara'")
		})

		t.Run("should return value for key with TTL extended", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 4, TTL: 500})
			lruCache.Set("user007", UserData{ID: 12, Name: "Liam", Age: 23})

			time.Sleep(400 * time.Millisecond)
			value, err := lruCache.Get("user007")
			assert.NoError(t, err, "Did not expect an error for key 'user007' before expiration")
			assert.Equal(t, UserData{ID: 12, Name: "Liam", Age: 23}, value, "Expected lruCache.Get(\"user007\") to return 'Liam' before expiration")

			time.Sleep(200 * time.Millisecond)
			lruCache.Set("user007", UserData{ID: 12, Name: "Liam", Age: 23})

			time.Sleep(300 * time.Millisecond)
			value, err = lruCache.Get("user007")
			assert.NoError(t, err, "Did not expect an error for key 'user007' with extended TTL")
			assert.Equal(t, UserData{ID: 12, Name: "Liam", Age: 23}, value, "Expected lruCache.Get(\"user007\") to return 'Liam' after extended TTL")
		})
	})

	t.Run("LRU cache: set with UserData", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[UserData]{}

		t.Run("should set record in cache", func(t *testing.T) {
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

		t.Run("should overwrite previous record in cache with the same key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice Updated", Age: 31})

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice Updated", Age: 31}, value)
		})

		t.Run("should overwrite previous record in cache with the same key and extend the TTL", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 500})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})

			time.Sleep(400 * time.Millisecond)
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice Updated", Age: 31})
			time.Sleep(400 * time.Millisecond)

			value, err := lruCache.Get("user1")
			assert.NoError(t, err)
			assert.Equal(t, UserData{ID: 1, Name: "Alice Updated", Age: 31}, value)
		})

		t.Run("should set record while evicting least recently used", func(t *testing.T) {
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

		t.Run("should take `get` operation into account while evicting least recently used", func(t *testing.T) {
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

		t.Run("should take `has` operation into account while evicting least recently used", func(t *testing.T) {
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

		t.Run("should take `set` operation into account while evicting least recently used", func(t *testing.T) {
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

		t.Run("should set record without evicting least recently used while TTL is reached", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 2, TTL: 500})
			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})

			time.Sleep(600 * time.Millisecond)
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

		t.Run("should behave correctly", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 500})

			value, err := lruCache.Get("user1")
			assert.Error(t, err, "Expected an error for non-existent key 'user1'")
			assert.Empty(t, value)

			lruCache.Set("user1", UserData{ID: 1, Name: "Alice", Age: 30})
			assert.False(t, lruCache.Has("user2"), "Expected lruCache.Has(\"user2\") to be false")

			lruCache.Set("user2", UserData{ID: 2, Name: "Bob", Age: 25})
			value, err = lruCache.Get("user1")
			assert.NoError(t, err, "Did not expect an error for existing key 'user1'")
			assert.Equal(t, UserData{ID: 1, Name: "Alice", Age: 30}, value, "Expected lruCache.Get(\"user1\") to return 'Alice'")

			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "Did not expect an error for existing key 'user2'")
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value, "Expected lruCache.Get(\"user2\") to return 'Bob'")

			time.Sleep(400 * time.Millisecond)
			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "Did not expect an error for existing key 'user2' after 400ms")
			assert.Equal(t, UserData{ID: 2, Name: "Bob", Age: 25}, value)

			time.Sleep(400 * time.Millisecond)
			assert.False(t, lruCache.Has("user1"), "Expected lruCache.Has(\"user1\") to be false after TTL expired")
			assert.True(t, lruCache.Has("user2"), "Expected lruCache.Has(\"user2\") to be true")

			lruCache.Set("user3", UserData{ID: 3, Name: "Charlie", Age: 35})
			lruCache.Set("user1", UserData{ID: 4, Name: "Alice Updated", Age: 31})
			lruCache.Set("user2", UserData{ID: 5, Name: "Bob Updated", Age: 26})
			lruCache.Set("user4", UserData{ID: 6, Name: "Dave", Age: 40})

			value, err = lruCache.Get("user3")
			assert.Error(t, err, "Expected an error for evicted key 'user3'")
			assert.Empty(t, value)

			value, err = lruCache.Get("user1")
			assert.NoError(t, err, "Did not expect an error for key 'user1'")
			assert.Equal(t, UserData{ID: 4, Name: "Alice Updated", Age: 31}, value, "Expected lruCache.Get(\"user1\") to return 'Alice Updated'")

			value, err = lruCache.Get("user2")
			assert.NoError(t, err, "Did not expect an error for key 'user2'")
			assert.Equal(t, UserData{ID: 5, Name: "Bob Updated", Age: 26}, value, "Expected lruCache.Get(\"user2\") to return 'Bob Updated'")

			value, err = lruCache.Get("user4")
			assert.NoError(t, err, "Did not expect an error for key 'user4'")
			assert.Equal(t, UserData{ID: 6, Name: "Dave", Age: 40}, value, "Expected lruCache.Get(\"user4\") to return 'Dave'")
		})
	})
}
