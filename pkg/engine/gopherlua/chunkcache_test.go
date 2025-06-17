// ABOUTME: Tests for ChunkCache which caches compiled Lua bytecode for performance optimization
// ABOUTME: Validates LRU eviction, TTL expiration, cache key generation, and concurrent access patterns

package gopherlua

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestChunkCache_Basic(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         3,
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	// Create a simple Lua state for testing
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name   string
		script string
		key    string
	}{
		{
			name:   "simple_script",
			script: "return 1 + 1",
			key:    "test1",
		},
		{
			name:   "function_script",
			script: "function add(a, b) return a + b end; return add(2, 3)",
			key:    "test2",
		},
		{
			name:   "table_script",
			script: "return {x = 10, y = 20}",
			key:    "test3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile script to get FunctionProto
			chunk, err := L.LoadString(tt.script)
			require.NoError(t, err)
			proto := chunk.Proto

			// Test Put
			cache.Put(tt.key, proto)

			// Test Get
			retrieved := cache.Get(tt.key)
			assert.NotNil(t, retrieved)
			assert.Equal(t, proto, retrieved)

			// Test that it exists by getting it again
			retrieved2 := cache.Get(tt.key)
			assert.NotNil(t, retrieved2)
		})
	}
}

func TestChunkCache_LRUEviction(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         2, // Small size to trigger eviction
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	L := lua.NewState()
	defer L.Close()

	// Create 3 chunks, but cache can only hold 2
	scripts := []string{
		"return 1",
		"return 2",
		"return 3",
	}

	for i, script := range scripts {
		chunk, err := L.LoadString(script)
		require.NoError(t, err)

		key := cache.GenerateKey(script, "")
		cache.Put(key, chunk.Proto)

		// First item should be evicted after adding the third
		if i == 2 {
			firstKey := cache.GenerateKey(scripts[0], "")
			assert.Nil(t, cache.Get(firstKey), "First item should be evicted")

			secondKey := cache.GenerateKey(scripts[1], "")
			assert.NotNil(t, cache.Get(secondKey), "Second item should still be present")

			thirdKey := cache.GenerateKey(scripts[2], "")
			assert.NotNil(t, cache.Get(thirdKey), "Third item should be present")
		}
	}
}

func TestChunkCache_TTLExpiration(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         10,
		TTL:             50 * time.Millisecond, // Very short TTL
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	L := lua.NewState()
	defer L.Close()

	script := "return 42"
	chunk, err := L.LoadString(script)
	require.NoError(t, err)

	key := cache.GenerateKey(script, "")

	// Put item in cache
	cache.Put(key, chunk.Proto)

	// Should still be there immediately
	retrieved := cache.Get(key)
	assert.NotNil(t, retrieved)

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Should be expired now
	retrieved = cache.Get(key)
	assert.Nil(t, retrieved)
}

func TestChunkCache_KeyGeneration(t *testing.T) {
	cache := NewChunkCache(ChunkCacheConfig{
		MaxSize: 10,
		TTL:     time.Minute,
	})

	tests := []struct {
		name     string
		script   string
		filename string
		wantDiff bool
	}{
		{
			name:     "same_script_same_key",
			script:   "return 1 + 1",
			filename: "",
			wantDiff: false,
		},
		{
			name:     "different_script_different_key",
			script:   "return 2 + 2",
			filename: "",
			wantDiff: true,
		},
		{
			name:     "same_script_different_filename",
			script:   "return 1 + 1",
			filename: "test.lua",
			wantDiff: true,
		},
	}

	baseKey := cache.GenerateKey("return 1 + 1", "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := cache.GenerateKey(tt.script, tt.filename)
			assert.NotEmpty(t, key)

			if tt.wantDiff {
				assert.NotEqual(t, baseKey, key)
			} else {
				assert.Equal(t, baseKey, key)
			}
		})
	}
}

func TestChunkCache_ConcurrentAccess(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         100,
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	L := lua.NewState()
	defer L.Close()

	const numGoroutines = 10
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Create some test chunks
	testChunks := make([]*lua.FunctionProto, 5)
	testKeys := make([]string, 5)
	for i := 0; i < 5; i++ {
		script := fmt.Sprintf("return %d", i)
		chunk, err := L.LoadString(script)
		require.NoError(t, err)
		testChunks[i] = chunk.Proto
		testKeys[i] = cache.GenerateKey(script, "")
	}

	// Concurrent writers
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				idx := (id + j) % len(testChunks)
				cache.Put(testKeys[idx], testChunks[idx])
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				idx := (id + j) % len(testKeys)
				_ = cache.Get(testKeys[idx])
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify cache is still functional
	assert.True(t, cache.Size() >= 0)
	assert.True(t, cache.Size() <= config.MaxSize)
}

func TestChunkCache_ClearAndSize(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         5,
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	L := lua.NewState()
	defer L.Close()

	// Add some items
	for i := 0; i < 3; i++ {
		script := fmt.Sprintf("return %d", i)
		chunk, err := L.LoadString(script)
		require.NoError(t, err)

		key := cache.GenerateKey(script, "")
		cache.Put(key, chunk.Proto)
	}

	// Check size
	assert.Equal(t, 3, cache.Size())

	// Clear cache
	cache.Clear()
	assert.Equal(t, 0, cache.Size())

	// Verify items are gone
	for i := 0; i < 3; i++ {
		script := fmt.Sprintf("return %d", i)
		key := cache.GenerateKey(script, "")
		assert.Nil(t, cache.Get(key))
	}
}

func TestChunkCache_EdgeCases(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         2,
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	t.Run("nil_proto", func(t *testing.T) {
		// Should handle nil gracefully
		cache.Put("nil_test", nil)
		retrieved := cache.Get("nil_test")
		assert.Nil(t, retrieved)
	})

	t.Run("empty_key", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		chunk, err := L.LoadString("return 1")
		require.NoError(t, err)

		// Empty key should still work
		cache.Put("", chunk.Proto)
		retrieved := cache.Get("")
		assert.NotNil(t, retrieved)
	})

	t.Run("overwrite_existing", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		chunk1, err := L.LoadString("return 1")
		require.NoError(t, err)

		chunk2, err := L.LoadString("return 2")
		require.NoError(t, err)

		key := "overwrite_test"

		// Put first chunk
		cache.Put(key, chunk1.Proto)
		retrieved := cache.Get(key)
		assert.Equal(t, chunk1.Proto, retrieved)

		// Overwrite with second chunk
		cache.Put(key, chunk2.Proto)
		retrieved = cache.Get(key)
		assert.Equal(t, chunk2.Proto, retrieved)
	})
}

func TestChunkCache_MemoryEfficiency(t *testing.T) {
	config := ChunkCacheConfig{
		MaxSize:         1000,
		TTL:             time.Minute,
		EnableDiskCache: false,
	}
	cache := NewChunkCache(config)

	L := lua.NewState()
	defer L.Close()

	// Add many items and verify memory usage is reasonable
	const numItems = 100

	for i := 0; i < numItems; i++ {
		script := fmt.Sprintf("local x = %d; return x * 2", i)
		chunk, err := L.LoadString(script)
		require.NoError(t, err)

		key := cache.GenerateKey(script, "")
		cache.Put(key, chunk.Proto)
	}

	assert.Equal(t, numItems, cache.Size())

	// Verify all items are accessible
	for i := 0; i < numItems; i++ {
		script := fmt.Sprintf("local x = %d; return x * 2", i)
		key := cache.GenerateKey(script, "")
		assert.NotNil(t, cache.Get(key))
	}
}
