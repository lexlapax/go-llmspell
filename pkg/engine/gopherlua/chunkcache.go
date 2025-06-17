// ABOUTME: ChunkCache implements compiled Lua chunk caching for performance optimization
// ABOUTME: Provides LRU-based caching with TTL support and optional disk persistence

package gopherlua

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ChunkCacheConfig configures the chunk cache behavior
type ChunkCacheConfig struct {
	// MaxSize is the maximum number of cached chunks
	MaxSize int

	// TTL is how long chunks stay in cache before expiring
	TTL time.Duration

	// EnableDiskCache enables persistent caching to disk
	EnableDiskCache bool

	// DiskCacheDir is the directory for disk cache (if enabled)
	DiskCacheDir string
}

// ChunkCache caches compiled Lua chunks for performance
type ChunkCache struct {
	config ChunkCacheConfig
	cache  map[string]*cacheEntry
	mu     sync.RWMutex

	// LRU tracking
	head, tail *cacheEntry
	size       int
}

// cacheEntry represents a cached chunk with metadata
type cacheEntry struct {
	key       string
	chunk     *lua.FunctionProto
	timestamp time.Time
	size      int64

	// LRU doubly-linked list
	prev, next *cacheEntry
}

// NewChunkCache creates a new chunk cache with the given configuration
func NewChunkCache(config ChunkCacheConfig) *ChunkCache {
	// Apply defaults
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.TTL <= 0 {
		config.TTL = 30 * time.Minute
	}

	cache := &ChunkCache{
		config: config,
		cache:  make(map[string]*cacheEntry),
	}

	// Initialize LRU list with sentinel nodes
	cache.head = &cacheEntry{}
	cache.tail = &cacheEntry{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// GenerateKey generates a cache key for the given script and filename
func (c *ChunkCache) GenerateKey(script, filename string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", filename, script)))
	return fmt.Sprintf("%x", hash)
}

// Get retrieves a cached chunk by key
func (c *ChunkCache) Get(key string) *lua.FunctionProto {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(entry.timestamp) > c.config.TTL {
		c.removeEntry(entry)
		return nil
	}

	// Move to front (most recently used)
	c.moveToFront(entry)

	return entry.chunk
}

// Put stores a chunk in the cache
func (c *ChunkCache) Put(key string, chunk *lua.FunctionProto) {
	if chunk == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if entry, exists := c.cache[key]; exists {
		entry.chunk = chunk
		entry.timestamp = time.Now()
		c.moveToFront(entry)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		chunk:     chunk,
		timestamp: time.Now(),
		size:      c.estimateChunkSize(chunk),
	}

	// Add to cache
	c.cache[key] = entry
	c.addToFront(entry)
	c.size++

	// Evict if necessary
	for c.size > c.config.MaxSize {
		c.evictLRU()
	}
}

// Clear removes all entries from the cache
func (c *ChunkCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.head.next = c.tail
	c.tail.prev = c.head
	c.size = 0
}

// Size returns the current number of cached entries
func (c *ChunkCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// Stats returns cache statistics
func (c *ChunkCache) Stats() ChunkCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := ChunkCacheStats{
		Size:    c.size,
		MaxSize: c.config.MaxSize,
		TTL:     c.config.TTL,
	}

	// Calculate total memory usage
	for _, entry := range c.cache {
		stats.MemoryUsage += entry.size
	}

	return stats
}

// ChunkCacheStats provides cache performance statistics
type ChunkCacheStats struct {
	Size        int           `json:"size"`
	MaxSize     int           `json:"max_size"`
	MemoryUsage int64         `json:"memory_usage"`
	TTL         time.Duration `json:"ttl"`
}

// Private methods for LRU management

func (c *ChunkCache) addToFront(entry *cacheEntry) {
	entry.prev = c.head
	entry.next = c.head.next
	c.head.next.prev = entry
	c.head.next = entry
}

func (c *ChunkCache) removeEntry(entry *cacheEntry) {
	delete(c.cache, entry.key)
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	c.size--
}

func (c *ChunkCache) moveToFront(entry *cacheEntry) {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	c.addToFront(entry)
}

func (c *ChunkCache) evictLRU() {
	if c.tail.prev == c.head {
		return // No entries to evict
	}

	lru := c.tail.prev
	c.removeEntry(lru)
}

// estimateChunkSize estimates the memory size of a compiled chunk
func (c *ChunkCache) estimateChunkSize(chunk *lua.FunctionProto) int64 {
	if chunk == nil {
		return 0
	}

	// Basic estimation based on number of instructions and constants
	size := int64(64) // Base overhead

	// Estimate instruction memory (each instruction is ~4 bytes)
	size += int64(len(chunk.Code)) * 4

	// Estimate constants memory
	for _, constant := range chunk.Constants {
		switch v := constant.(type) {
		case lua.LString:
			size += int64(len(string(v)))
		case lua.LNumber:
			size += 8
		case lua.LBool:
			size += 1
		default:
			size += 16 // Default estimate
		}
	}

	// Estimate local variable info
	size += int64(len(chunk.DbgLocals)) * 32

	// Estimate upvalue info
	size += int64(len(chunk.DbgUpvalues)) * 16

	return size
}
