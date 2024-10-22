package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {

	t.Run("LRU cache: has", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[string]{}

		t.Run("should return false for non-existent key and true for existing", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			assert.Equal(t, false, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be false")
			lruCache.Set("foo", "bar")
			assert.Equal(t, true, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be true")
		})

		t.Run("should return false for evicted key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar") // This should evict "foo"
			assert.Equal(t, false, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be false")
			assert.Equal(t, true, lruCache.Has("baz"), "Expected lruCache.Has(\"baz\") to be true")
		})

		t.Run("should return false for expired key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 500})
			lruCache.Set("foo", "bar")
			assert.Equal(t, true, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be true before expiration")

			time.Sleep(600 * time.Millisecond) // Wait for the key to expire
			assert.Equal(t, false, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be false after expiration")
		})

		t.Run("should return true for multiple existing keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar")
			assert.Equal(t, true, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be true")
			assert.Equal(t, true, lruCache.Has("baz"), "Expected lruCache.Has(\"baz\") to be true")
		})

		t.Run("should return true for key with TTL extended", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 500})
			lruCache.Set("foo", "bar")

			time.Sleep(400 * time.Millisecond) // Before TTL expires
			assert.Equal(t, true, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be true before TTL expires")

			time.Sleep(400 * time.Millisecond) // Further access to extend TTL
			assert.Equal(t, true, lruCache.Has("foo"), "Expected lruCache.Has(\"foo\") to be true after further access")
		})
	})

	t.Run("LRU cache: get", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[string]{}

		t.Run("should return empty and error for non-existent key, and value for existing key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			value, err := lruCache.Get("foo")
			assert.Error(t, err, "Expected an error for non-existent key 'foo'")
			assert.Empty(t, value, "Expected value to be nil for non-existent key 'foo'")

			lruCache.Set("foo", "bar")
			value, err = lruCache.Get("foo")
			assert.NoError(t, err, "Did not expect an error for existing key 'foo'")
			assert.Equal(t, "bar", value, "Expected lruCache.Get(\"foo\") to return 'bar'")
		})

		t.Run("should return empty and error for evicted key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar") // This should evict "foo"
			value, err := lruCache.Get("foo")
			assert.Error(t, err, "Expected an error for evicted key 'foo'")
			assert.Empty(t, value, "Expected value to be nil for evicted key 'foo'")

			value, err = lruCache.Get("baz")
			assert.NoError(t, err, "Did not expect an error for existing key 'baz'")
			assert.Equal(t, "bar", value, "Expected lruCache.Get(\"baz\") to return 'bar'")
		})

		t.Run("should return nil and error for expired key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 1, TTL: 500})
			lruCache.Set("foo", "bar")
			value, err := lruCache.Get("foo")
			assert.NoError(t, err, "Did not expect an error for key 'foo' before expiration")
			assert.Equal(t, "bar", value, "Expected lruCache.Get(\"foo\") to return 'bar' before expiration")

			time.Sleep(600 * time.Millisecond) // Wait for the key to expire
			value, err = lruCache.Get("foo")
			assert.Error(t, err, "Expected an error for expired key 'foo'")
			assert.Empty(t, value, "Expected value to be nil for expired key 'foo'")
		})

		t.Run("should return value for multiple existing keys", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("foo", "foo")
			lruCache.Set("baz", "baz")

			value, err := lruCache.Get("foo")
			assert.NoError(t, err, "Did not expect an error for existing key 'foo'")
			assert.Equal(t, "foo", value, "Expected lruCache.Get(\"foo\") to return 'foo'")

			value, err = lruCache.Get("baz")
			assert.NoError(t, err, "Did not expect an error for existing key 'baz'")
			assert.Equal(t, "baz", value, "Expected lruCache.Get(\"baz\") to return 'baz'")
		})

		t.Run("should return value for key with TTL extended", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 500})
			lruCache.Set("foo", "bar")

			time.Sleep(400 * time.Millisecond) // Before TTL expires
			value, err := lruCache.Get("foo")
			assert.NoError(t, err, "Did not expect an error for existing key 'foo' before TTL expires")
			assert.Equal(t, "bar", value, "Expected lruCache.Get(\"foo\") to return 'bar'")

			time.Sleep(400 * time.Millisecond) // Access it again, extending the TTL
			value, err = lruCache.Get("foo")
			assert.NoError(t, err, "Did not expect an error for existing key 'foo' after TTL extension")
			assert.Equal(t, "bar", value, "Expected lruCache.Get(\"foo\") to return 'bar' after TTL extension")
		})
	})

	t.Run("LRU cache: set", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[string]{}

		t.Run("should set record in cache", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("bar", "foo")

			value, err := lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "foo", value)
		})

		t.Run("should overwrite previous record in cache with the same key", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("foo", "baz")

			value, err := lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)
		})

		t.Run("should overwrite previous record in cache with the same key and extend the TTL", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 10, TTL: 500})
			lruCache.Set("foo", "bar")

			time.Sleep(400 * time.Millisecond)
			lruCache.Set("foo", "baz")
			time.Sleep(400 * time.Millisecond)

			value, err := lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)
		})

		t.Run("should set record while evicting least recently used", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 2, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar")
			lruCache.Set("bar", "baz")

			value, err := lruCache.Get("foo")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("baz")
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)
		})

		t.Run("should take `get` operation into account while evicting least recently used", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("foo", "bar")
			time.Sleep(100 * time.Millisecond)
			lruCache.Set("baz", "bar")
			time.Sleep(100 * time.Millisecond)
			lruCache.Set("bar", "baz")

			value, err := lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)

			lruCache.Set("foobar", "barbaz")
			time.Sleep(100 * time.Millisecond)

			value, err = lruCache.Get("baz")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)

			value, err = lruCache.Get("foobar")
			assert.NoError(t, err)
			assert.Equal(t, "barbaz", value)
		})

		t.Run("should take `has` operation into account while evicting least recently used", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar")
			lruCache.Set("bar", "baz")

			assert.True(t, lruCache.Has("foo"))

			lruCache.Set("foobar", "barbaz")

			value, err := lruCache.Get("baz")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "bar", value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)

			value, err = lruCache.Get("foobar")
			assert.NoError(t, err)
			assert.Equal(t, "barbaz", value)
		})

		t.Run("should take `set` operation into account while evicting least recently used", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 50000})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar")
			lruCache.Set("bar", "baz")
			lruCache.Set("foo", "barbaz")
			lruCache.Set("foobar", "barbaz")

			value, err := lruCache.Get("baz")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)

			value, err = lruCache.Get("foo")
			assert.NoError(t, err)
			assert.Equal(t, "barbaz", value)

			value, err = lruCache.Get("foobar")
			assert.NoError(t, err)
			assert.Equal(t, "barbaz", value)
		})

		t.Run("should set record without evicting least recently used while TTL is reached", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 2, TTL: 500})
			lruCache.Set("foo", "bar")
			lruCache.Set("baz", "bar")

			time.Sleep(600 * time.Millisecond)
			lruCache.Set("bar", "baz")

			value, err := lruCache.Get("foo")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("baz")
			assert.Error(t, err)
			assert.Empty(t, value)

			value, err = lruCache.Get("bar")
			assert.NoError(t, err)
			assert.Equal(t, "baz", value)
		})
	})

	t.Run("Arbitrary operations", func(t *testing.T) {
		cacheProvider := InMemoryLRUCacheProvider[string]{}

		t.Run("should behave correctly", func(t *testing.T) {
			lruCache := cacheProvider.NewLRUCache(LRUCacheConfig{ItemLimit: 3, TTL: 500})

			// Check that a non-existent key returns an error or empty value
			value, err := lruCache.Get("key1")
			assert.Error(t, err, "Expected an error for non-existent key 'key1'")
			assert.Empty(t, value)

			// Set and verify key1
			lruCache.Set("key1", "value1")
			assert.False(t, lruCache.Has("key2"), "Expected lruCache.Has(\"key2\") to be false")

			// Set key2 and check both key1 and key2
			lruCache.Set("key2", "value2")
			value, err = lruCache.Get("key1")
			assert.NoError(t, err, "Did not expect an error for existing key 'key1'")
			assert.Equal(t, "value1", value, "Expected lruCache.Get(\"key1\") to return 'value1'")

			value, err = lruCache.Get("key2")
			assert.NoError(t, err, "Did not expect an error for existing key 'key2'")
			assert.Equal(t, "value2", value, "Expected lruCache.Get(\"key2\") to return 'value2'")

			// Wait for 400ms and check if key2 still exists
			time.Sleep(400 * time.Millisecond)
			value, err = lruCache.Get("key2")
			assert.NoError(t, err, "Did not expect an error for existing key 'key2' after 400ms")
			assert.Equal(t, "value2", value)

			// Wait for additional 400ms to trigger TTL expiration for key1
			time.Sleep(400 * time.Millisecond)
			assert.False(t, lruCache.Has("key1"), "Expected lruCache.Has(\"key1\") to be false after TTL expired")
			assert.True(t, lruCache.Has("key2"), "Expected lruCache.Has(\"key2\") to be true")

			// Set key3, key1 again, key2 with new value, and key4
			lruCache.Set("key3", "value3")
			lruCache.Set("key1", "value1")
			lruCache.Set("key2", "-value2-")
			lruCache.Set("key4", "value4")

			// Check eviction of least recently used key (key3 should be evicted)
			value, err = lruCache.Get("key3")
			assert.Error(t, err, "Expected an error for evicted key 'key3'")
			assert.Empty(t, value)

			// Check the existence and values of the remaining keys
			value, err = lruCache.Get("key1")
			assert.NoError(t, err, "Did not expect an error for key 'key1'")
			assert.Equal(t, "value1", value)

			value, err = lruCache.Get("key2")
			assert.NoError(t, err, "Did not expect an error for key 'key2'")
			assert.Equal(t, "-value2-", value)

			value, err = lruCache.Get("key4")
			assert.NoError(t, err, "Did not expect an error for key 'key4'")
			assert.Equal(t, "value4", value)
		})

	})

}
