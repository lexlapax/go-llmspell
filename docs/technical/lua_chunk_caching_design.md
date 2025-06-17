# Lua Compiled Chunk Caching Design

## Overview
This document designs a compiled chunk caching system for GopherLua to improve performance by avoiding repeated script compilation and enabling bytecode reuse across LState instances.

## Background

### Compilation Process in GopherLua
1. **Parse**: Source code → Abstract Syntax Tree (AST)
2. **Compile**: AST → FunctionProto (bytecode)
3. **Execute**: FunctionProto → LFunction → execution

### Performance Implications
- Parsing and compilation are expensive operations
- Scripts often executed multiple times
- Multiple LStates may run identical scripts
- Bytecode is read-only and safe to share

## Cache Architecture

### Core Components
```go
// ChunkCache manages compiled Lua chunks
type ChunkCache struct {
    mu         sync.RWMutex
    cache      map[string]*CachedChunk
    config     CacheConfig
    stats      CacheStats
    evictionQ  *list.List // LRU eviction
}

// CachedChunk represents a compiled chunk
type CachedChunk struct {
    Key          string              // Cache key
    Proto        *lua.FunctionProto  // Compiled bytecode
    Source       string              // Original source (optional)
    Hash         string              // Content hash
    Size         int64               // Memory size estimate
    CompileTime  time.Duration       // Compilation duration
    
    // Metadata
    Created      time.Time
    LastAccessed time.Time
    AccessCount  int64
    
    // LRU tracking
    element      *list.Element
}

// CacheConfig defines cache behavior
type CacheConfig struct {
    MaxSize         int64         // Max cache size in bytes
    MaxEntries      int           // Max number of entries
    TTL             time.Duration // Time to live
    EvictionPolicy  EvictionPolicy
    EnableStats     bool
    CompressSource  bool          // Compress source code
    ValidateOnLoad  bool          // Validate chunk on load
}

// CacheStats tracks cache performance
type CacheStats struct {
    Hits        atomic.Int64
    Misses      atomic.Int64
    Evictions   atomic.Int64
    CompileTime atomic.Int64 // Total nanoseconds
    SavedTime   atomic.Int64 // Time saved by cache hits
}
```

### Cache Key Generation
```go
// GenerateCacheKey creates a unique key for a chunk
func GenerateCacheKey(source string, options CompileOptions) string {
    h := sha256.New()
    h.Write([]byte(source))
    
    // Include compilation options in key
    binary.Write(h, binary.LittleEndian, options.OptimizationLevel)
    binary.Write(h, binary.LittleEndian, options.StripDebug)
    
    return hex.EncodeToString(h.Sum(nil))
}

// CompileOptions affects bytecode generation
type CompileOptions struct {
    OptimizationLevel int
    StripDebug        bool
    SourceName        string
}
```

## Compilation and Caching

### Smart Compilation
```go
// CompileOrGetCached compiles a chunk or returns cached version
func (c *ChunkCache) CompileOrGetCached(source string, options CompileOptions) (*CachedChunk, error) {
    key := GenerateCacheKey(source, options)
    
    // Try cache first
    if chunk := c.get(key); chunk != nil {
        c.stats.Hits.Add(1)
        c.stats.SavedTime.Add(int64(chunk.CompileTime))
        return chunk, nil
    }
    
    c.stats.Misses.Add(1)
    
    // Compile new chunk
    start := time.Now()
    proto, err := c.compileChunk(source, options)
    if err != nil {
        return nil, err
    }
    compileTime := time.Since(start)
    
    // Create cached chunk
    chunk := &CachedChunk{
        Key:         key,
        Proto:       proto,
        Source:      source,
        Hash:        key,
        CompileTime: compileTime,
        Created:     time.Now(),
        AccessCount: 1,
    }
    
    // Estimate memory size
    chunk.Size = c.estimateProtoSize(proto)
    
    // Add to cache
    c.put(key, chunk)
    c.stats.CompileTime.Add(int64(compileTime))
    
    return chunk, nil
}

// compileChunk performs actual compilation
func (c *ChunkCache) compileChunk(source string, options CompileOptions) (*lua.FunctionProto, error) {
    reader := strings.NewReader(source)
    
    // Parse to AST
    chunk, err := parse.Parse(reader, options.SourceName)
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }
    
    // Apply optimizations if needed
    if options.OptimizationLevel > 0 {
        chunk = c.optimizeAST(chunk, options.OptimizationLevel)
    }
    
    // Compile to bytecode
    proto, err := lua.Compile(chunk, options.SourceName)
    if err != nil {
        return nil, fmt.Errorf("compile error: %w", err)
    }
    
    // Strip debug info if requested
    if options.StripDebug {
        c.stripDebugInfo(proto)
    }
    
    return proto, nil
}
```

### File-based Caching
```go
// CompileFile compiles a file with caching
func (c *ChunkCache) CompileFile(filepath string, options CompileOptions) (*CachedChunk, error) {
    // Use file path and modification time as cache key
    stat, err := os.Stat(filepath)
    if err != nil {
        return nil, err
    }
    
    key := fmt.Sprintf("file:%s:%d:%d", filepath, stat.ModTime().Unix(), stat.Size())
    
    // Check cache
    if chunk := c.get(key); chunk != nil {
        c.stats.Hits.Add(1)
        return chunk, nil
    }
    
    // Read and compile file
    content, err := os.ReadFile(filepath)
    if err != nil {
        return nil, err
    }
    
    options.SourceName = filepath
    return c.CompileOrGetCached(string(content), options)
}
```

## Memory Management

### Size Estimation
```go
// estimateProtoSize estimates memory usage of a FunctionProto
func (c *ChunkCache) estimateProtoSize(proto *lua.FunctionProto) int64 {
    size := int64(unsafe.Sizeof(*proto))
    
    // Instructions
    size += int64(len(proto.Code)) * 4
    
    // Constants
    for _, k := range proto.Constants {
        switch k.Type() {
        case lua.LTString:
            size += int64(len(k.String()))
        case lua.LTNumber:
            size += 8
        // ... other types
        }
    }
    
    // Function protos
    for _, p := range proto.FunctionPrototypes {
        size += c.estimateProtoSize(p)
    }
    
    // Debug info
    size += int64(len(proto.DbgSourcePositions)) * 4
    size += int64(len(proto.DbgLocals)) * int64(unsafe.Sizeof(lua.DbgLocalInfo{}))
    size += int64(len(proto.DbgUpvalues)) * 16
    
    return size
}
```

### Eviction Policies
```go
// LRU Eviction
func (c *ChunkCache) evictLRU() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Remove least recently used until under size limit
    for c.totalSize > c.config.MaxSize && c.evictionQ.Len() > 0 {
        elem := c.evictionQ.Back()
        chunk := elem.Value.(*CachedChunk)
        
        delete(c.cache, chunk.Key)
        c.evictionQ.Remove(elem)
        c.totalSize -= chunk.Size
        c.stats.Evictions.Add(1)
    }
}

// TTL Eviction
func (c *ChunkCache) evictExpired() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    now := time.Now()
    for key, chunk := range c.cache {
        if now.Sub(chunk.LastAccessed) > c.config.TTL {
            delete(c.cache, key)
            if chunk.element != nil {
                c.evictionQ.Remove(chunk.element)
            }
            c.totalSize -= chunk.Size
            c.stats.Evictions.Add(1)
        }
    }
}
```

## Usage Patterns

### Basic Usage
```go
// Create cache
cache := NewChunkCache(CacheConfig{
    MaxSize:    100 * 1024 * 1024, // 100MB
    MaxEntries: 1000,
    TTL:        1 * time.Hour,
})

// Compile with caching
chunk, err := cache.CompileOrGetCached(`
    function fibonacci(n)
        if n <= 1 then return n end
        return fibonacci(n-1) + fibonacci(n-2)
    end
    return fibonacci(10)
`, CompileOptions{
    OptimizationLevel: 1,
    SourceName:        "fibonacci.lua",
})

// Use with LState
L := lua.NewState()
fn := L.NewFunctionFromProto(chunk.Proto)
L.Push(fn)
L.Call(0, 1)
result := L.Get(-1)
```

### Integration with LuaEngine
```go
type LuaEngine struct {
    pool       *LStatePool
    chunkCache *ChunkCache
    config     EngineConfig
}

func (e *LuaEngine) ExecuteScript(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
    // Get cached chunk
    chunk, err := e.chunkCache.CompileOrGetCached(script, CompileOptions{
        OptimizationLevel: e.config.OptimizationLevel,
        StripDebug:        e.config.Production,
    })
    if err != nil {
        return nil, fmt.Errorf("compilation failed: %w", err)
    }
    
    // Get LState from pool
    L, err := e.pool.Get(ctx)
    if err != nil {
        return nil, err
    }
    defer e.pool.Put(L)
    
    // Create function from cached proto
    fn := L.NewFunctionFromProto(chunk.Proto)
    
    // Set up parameters
    for k, v := range params {
        lv, err := e.converter.ToLValue(L, v)
        if err != nil {
            return nil, err
        }
        L.SetGlobal(k, lv)
    }
    
    // Execute
    L.Push(fn)
    if err := L.PCall(0, lua.MultRet, nil); err != nil {
        return nil, err
    }
    
    // Get results
    return e.extractResults(L)
}
```

## Optimization Strategies

### AST Optimizations
```go
func (c *ChunkCache) optimizeAST(chunk ast.Chunk, level int) ast.Chunk {
    if level >= 1 {
        // Constant folding
        chunk = c.constantFolding(chunk)
        
        // Dead code elimination
        chunk = c.deadCodeElimination(chunk)
    }
    
    if level >= 2 {
        // Function inlining for small functions
        chunk = c.inlineSmallFunctions(chunk)
        
        // Loop optimizations
        chunk = c.optimizeLoops(chunk)
    }
    
    return chunk
}
```

### Debug Info Stripping
```go
func (c *ChunkCache) stripDebugInfo(proto *lua.FunctionProto) {
    // Clear debug information to reduce memory
    proto.DbgSourcePositions = nil
    proto.DbgLocals = nil
    proto.DbgCalls = nil
    proto.DbgUpvalues = nil
    
    // Recursively strip from inner functions
    for _, fp := range proto.FunctionPrototypes {
        c.stripDebugInfo(fp)
    }
}
```

## Persistence

### Disk Cache
```go
// SaveToDisk persists cache to disk
func (c *ChunkCache) SaveToDisk(path string) error {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    data := CachePersistData{
        Version: 1,
        Entries: make([]PersistedEntry, 0, len(c.cache)),
    }
    
    for key, chunk := range c.cache {
        // Serialize proto to custom format
        protoData, err := c.serializeProto(chunk.Proto)
        if err != nil {
            continue
        }
        
        data.Entries = append(data.Entries, PersistedEntry{
            Key:         key,
            ProtoData:   protoData,
            CompileTime: chunk.CompileTime,
            Hash:        chunk.Hash,
        })
    }
    
    // Save with compression
    return c.saveCompressed(path, data)
}

// LoadFromDisk restores cache from disk
func (c *ChunkCache) LoadFromDisk(path string) error {
    data, err := c.loadCompressed(path)
    if err != nil {
        return err
    }
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    for _, entry := range data.Entries {
        proto, err := c.deserializeProto(entry.ProtoData)
        if err != nil {
            continue
        }
        
        chunk := &CachedChunk{
            Key:         entry.Key,
            Proto:       proto,
            Hash:        entry.Hash,
            CompileTime: entry.CompileTime,
            Created:     time.Now(),
        }
        
        c.cache[entry.Key] = chunk
    }
    
    return nil
}
```

## Benchmarking

### Performance Metrics
```go
func (c *ChunkCache) GetMetrics() CacheMetrics {
    return CacheMetrics{
        HitRate:       float64(c.stats.Hits.Load()) / float64(c.stats.Hits.Load() + c.stats.Misses.Load()),
        TotalHits:     c.stats.Hits.Load(),
        TotalMisses:   c.stats.Misses.Load(),
        TotalEvictions: c.stats.Evictions.Load(),
        AvgCompileTime: time.Duration(c.stats.CompileTime.Load() / c.stats.Misses.Load()),
        TimeSaved:     time.Duration(c.stats.SavedTime.Load()),
        CacheSize:     c.totalSize,
        EntryCount:    len(c.cache),
    }
}
```

### Benchmark Example
```go
func BenchmarkChunkCache(b *testing.B) {
    cache := NewChunkCache(DefaultCacheConfig())
    
    script := `
        local function factorial(n)
            if n <= 1 then return 1 end
            return n * factorial(n - 1)
        end
        return factorial(20)
    `
    
    b.ResetTimer()
    
    b.Run("WithCache", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            chunk, _ := cache.CompileOrGetCached(script, CompileOptions{})
            L := lua.NewState()
            L.Push(L.NewFunctionFromProto(chunk.Proto))
            L.Call(0, 1)
            L.Close()
        }
    })
    
    b.Run("WithoutCache", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            L := lua.NewState()
            L.DoString(script)
            L.Close()
        }
    })
}
```

## Best Practices

1. **Cache Warming**: Pre-compile frequently used scripts at startup
2. **Key Strategy**: Use content hash for dynamic scripts, file path + mtime for files
3. **Memory Limits**: Set appropriate cache size based on available memory
4. **TTL Settings**: Balance between memory usage and compilation overhead
5. **Production Mode**: Strip debug info in production for smaller cache
6. **Monitoring**: Track cache metrics to tune configuration
7. **Thread Safety**: Cache is safe for concurrent access
8. **Error Handling**: Failed compilations should not be cached