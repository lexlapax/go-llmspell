// ABOUTME: Optimized type converter with conversion caching, fast paths for common types, and reduced allocations
// ABOUTME: Provides significant performance improvements for high-frequency type conversions between Go and Lua

package gopherlua

import (
	"fmt"
	// "reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	// "github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// OptimizedConverterConfig provides configuration for the optimized converter
type OptimizedConverterConfig struct {
	CacheSize           int
	EnableCaching       bool
	EnableFastPaths     bool
	ReduceAllocations   bool
	OptimizeTableAccess bool
	EnableTypeHints     bool
}

// OptimizedConverter extends LuaTypeConverter with performance optimizations
type OptimizedConverter struct {
	*LuaTypeConverter

	// Caching
	enableCaching  bool
	primitiveCache *sync.Map // For primitive values
	stringIntern   *stringInternPool
	tableCache     *tableConversionCache

	// Fast paths
	enableFastPaths bool

	// Allocation reduction
	reduceAllocations bool
	bufferPool        sync.Pool

	// Table optimization
	optimizeTableAccess bool

	// Type hints
	enableTypeHints bool
	typeHintMap     sync.Map

	// Metrics
	cacheHits   int64
	cacheMisses int64
	fastPathUse int64
}

// stringInternPool provides string interning for reduced memory usage
type stringInternPool struct {
	mu    sync.RWMutex
	pool  map[string]string
	size  int
	limit int
}

// tableConversionCache caches table conversion results
type tableConversionCache struct {
	mu    sync.RWMutex
	cache map[uintptr]interface{}
	size  int
	limit int
}

// conversionBuffer is reused for temporary allocations
type conversionBuffer struct {
	keys   []string
	values []interface{}
}

// NewOptimizedConverter creates a new optimized type converter
func NewOptimizedConverter(config OptimizedConverterConfig) *OptimizedConverter {
	// Set defaults
	if config.CacheSize <= 0 {
		config.CacheSize = 1000
	}

	baseConverter := NewLuaTypeConverter()

	oc := &OptimizedConverter{
		LuaTypeConverter:    baseConverter,
		enableCaching:       config.EnableCaching,
		enableFastPaths:     config.EnableFastPaths,
		reduceAllocations:   config.ReduceAllocations,
		optimizeTableAccess: config.OptimizeTableAccess,
		enableTypeHints:     config.EnableTypeHints,
		primitiveCache:      &sync.Map{},
		stringIntern: &stringInternPool{
			pool:  make(map[string]string),
			limit: config.CacheSize,
		},
		tableCache: &tableConversionCache{
			cache: make(map[uintptr]interface{}),
			limit: config.CacheSize / 2,
		},
	}

	// Initialize buffer pool
	if config.ReduceAllocations {
		oc.bufferPool = sync.Pool{
			New: func() interface{} {
				return &conversionBuffer{
					keys:   make([]string, 0, 16),
					values: make([]interface{}, 0, 16),
				}
			},
		}
	}

	return oc
}

// ToLua converts a Go value to a Lua value with optimizations
func (oc *OptimizedConverter) ToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	// Fast path for nil
	if value == nil {
		atomic.AddInt64(&oc.fastPathUse, 1)
		return lua.LNil, nil
	}

	// Fast paths for common types
	if oc.enableFastPaths {
		if lval := oc.fastPathToLua(L, value); lval != nil {
			atomic.AddInt64(&oc.fastPathUse, 1)
			return lval, nil
		}
	}

	// Check cache for primitive types
	if oc.enableCaching && oc.isCacheablePrimitive(value) {
		cacheKey := oc.generatePrimitiveCacheKey(value)
		if cached, ok := oc.primitiveCache.Load(cacheKey); ok {
			atomic.AddInt64(&oc.cacheHits, 1)
			// For primitives, we can reconstruct the Lua value
			if lval := oc.reconstructLuaValue(L, cached, value); lval != nil {
				return lval, nil
			}
		}
		atomic.AddInt64(&oc.cacheMisses, 1)

		// Convert and cache
		lval, err := oc.LuaTypeConverter.ToLua(L, value)
		if err == nil && lval != nil {
			oc.primitiveCache.Store(cacheKey, oc.getLuaValueType(lval))
		}
		return lval, err
	}

	// Use base converter for complex types
	return oc.LuaTypeConverter.ToLua(L, value)
}

// fastPathToLua provides optimized conversion for common types
func (oc *OptimizedConverter) fastPathToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int32:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case string:
		// Use string interning for common strings
		if oc.enableCaching && len(v) < 64 {
			return lua.LString(oc.internString(v))
		}
		return lua.LString(v)
	case []interface{}:
		if len(v) == 0 {
			return L.NewTable()
		}
		// Fall through to regular conversion for non-empty slices
	case map[string]interface{}:
		if len(v) == 0 {
			return L.NewTable()
		}
		// Fall through to regular conversion for non-empty maps
	}
	return nil
}

// FromLua converts a Lua value to a Go value with optimizations
func (oc *OptimizedConverter) FromLua(value lua.LValue) (interface{}, error) {
	// Fast path for common types
	if oc.enableFastPaths {
		if goVal := oc.fastPathFromLua(value); goVal != nil {
			atomic.AddInt64(&oc.fastPathUse, 1)
			return goVal, nil
		}
	}

	// Check table cache
	if table, ok := value.(*lua.LTable); ok && oc.enableCaching {
		tablePtr := uintptr(unsafe.Pointer(table))
		if cached := oc.getTableCache(tablePtr); cached != nil {
			atomic.AddInt64(&oc.cacheHits, 1)
			return cached, nil
		}
		atomic.AddInt64(&oc.cacheMisses, 1)

		// Convert and cache
		result, err := oc.LuaTypeConverter.FromLua(value)
		if err == nil && result != nil {
			oc.setTableCache(tablePtr, result)
		}
		return result, err
	}

	// Use base converter
	return oc.LuaTypeConverter.FromLua(value)
}

// fastPathFromLua provides optimized conversion from Lua for common types
func (oc *OptimizedConverter) fastPathFromLua(value lua.LValue) interface{} {
	switch v := value.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		// Return as float64 to maintain consistency
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LNilType:
		return nil
	}
	return nil
}

// Helper methods

func (oc *OptimizedConverter) isCacheablePrimitive(value interface{}) bool {
	if value == nil {
		return true
	}

	switch value.(type) {
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, string:
		return true
	}
	return false
}

func (oc *OptimizedConverter) generatePrimitiveCacheKey(value interface{}) string {
	return fmt.Sprintf("%T:%v", value, value)
}

func (oc *OptimizedConverter) getLuaValueType(lval lua.LValue) string {
	switch lval.(type) {
	case lua.LBool:
		return "bool"
	case lua.LNumber:
		return "number"
	case lua.LString:
		return "string"
	case *lua.LNilType:
		return "nil"
	default:
		return "unknown"
	}
}

func (oc *OptimizedConverter) reconstructLuaValue(L *lua.LState, cachedType interface{}, value interface{}) lua.LValue {
	typeStr, ok := cachedType.(string)
	if !ok {
		return nil
	}

	switch typeStr {
	case "bool":
		if b, ok := value.(bool); ok {
			return lua.LBool(b)
		}
	case "number":
		switch v := value.(type) {
		case int:
			return lua.LNumber(v)
		case int32:
			return lua.LNumber(v)
		case int64:
			return lua.LNumber(v)
		case float32:
			return lua.LNumber(v)
		case float64:
			return lua.LNumber(v)
		}
	case "string":
		if s, ok := value.(string); ok {
			return lua.LString(s)
		}
	case "nil":
		return lua.LNil
	}

	return nil
}

// String interning
func (oc *OptimizedConverter) internString(s string) string {
	oc.stringIntern.mu.RLock()
	if interned, ok := oc.stringIntern.pool[s]; ok {
		oc.stringIntern.mu.RUnlock()
		return interned
	}
	oc.stringIntern.mu.RUnlock()

	oc.stringIntern.mu.Lock()
	defer oc.stringIntern.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, ok := oc.stringIntern.pool[s]; ok {
		return interned
	}

	// Add to pool if under limit
	if oc.stringIntern.size < oc.stringIntern.limit {
		oc.stringIntern.pool[s] = s
		oc.stringIntern.size++
		return s
	}

	return s
}

// Table caching
func (oc *OptimizedConverter) getTableCache(ptr uintptr) interface{} {
	oc.tableCache.mu.RLock()
	defer oc.tableCache.mu.RUnlock()
	return oc.tableCache.cache[ptr]
}

func (oc *OptimizedConverter) setTableCache(ptr uintptr, value interface{}) {
	oc.tableCache.mu.Lock()
	defer oc.tableCache.mu.Unlock()

	if oc.tableCache.size < oc.tableCache.limit {
		oc.tableCache.cache[ptr] = value
		oc.tableCache.size++
	}
}

// GetCacheStats returns detailed cache statistics
func (oc *OptimizedConverter) GetCacheStats() CacheStats {
	primitiveSize := 0
	oc.primitiveCache.Range(func(_, _ interface{}) bool {
		primitiveSize++
		return true
	})

	// Lock for reading string intern size
	oc.stringIntern.mu.RLock()
	stringInternSize := oc.stringIntern.size
	oc.stringIntern.mu.RUnlock()

	// Lock for reading table cache size
	oc.tableCache.mu.RLock()
	tableCacheSize := oc.tableCache.size
	oc.tableCache.mu.RUnlock()

	return CacheStats{
		Hits:      atomic.LoadInt64(&oc.cacheHits),
		Misses:    atomic.LoadInt64(&oc.cacheMisses),
		Size:      primitiveSize + stringInternSize + tableCacheSize,
		Evictions: 0, // Not tracked in this implementation
	}
}

// ResetCacheStats resets cache statistics
func (oc *OptimizedConverter) ResetCacheStats() {
	atomic.StoreInt64(&oc.cacheHits, 0)
	atomic.StoreInt64(&oc.cacheMisses, 0)
	atomic.StoreInt64(&oc.fastPathUse, 0)
}

// Type hint support for Lua scripts
func (oc *OptimizedConverter) InstallTypeHints(L *lua.LState) {
	if !oc.enableTypeHints {
		return
	}

	// Numeric array hint
	L.SetGlobal("__hint_numeric_array", L.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		// Mark table as numeric array for optimized access
		oc.typeHintMap.Store(uintptr(unsafe.Pointer(table)), "numeric_array")
		L.Push(table)
		return 1
	}))

	// String array hint
	L.SetGlobal("__hint_string_array", L.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		oc.typeHintMap.Store(uintptr(unsafe.Pointer(table)), "string_array")
		L.Push(table)
		return 1
	}))

	// String concatenation optimization
	L.SetGlobal("__hint_string_concat", L.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		length := table.Len()

		// Pre-calculate total length
		totalLen := 0
		parts := make([]string, length)
		for i := 1; i <= length; i++ {
			if str, ok := table.RawGetInt(i).(lua.LString); ok {
				parts[i-1] = string(str)
				totalLen += len(str)
			}
		}

		// Single allocation for result
		result := make([]byte, 0, totalLen)
		for _, part := range parts {
			result = append(result, part...)
		}

		L.Push(lua.LString(result))
		return 1
	}))

	// Bulk conversion hint
	L.SetGlobal("__hint_bulk_convert", L.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		length := table.Len()
		result := L.NewTable()

		for i := 1; i <= length; i++ {
			item := table.RawGetInt(i)
			if itemTable, ok := item.(*lua.LTable); ok {
				typeField := itemTable.RawGetString("type")
				valueField := itemTable.RawGetString("value")

				var converted lua.LValue
				switch typeStr, _ := typeField.(lua.LString); string(typeStr) {
				case "number":
					if str, ok := valueField.(lua.LString); ok {
						var num float64
						if _, err := fmt.Sscanf(string(str), "%f", &num); err == nil {
							converted = lua.LNumber(num)
						} else {
							converted = valueField
						}
					}
				case "boolean":
					if str, ok := valueField.(lua.LString); ok {
						converted = lua.LBool(string(str) == "true")
					}
				case "string":
					converted = lua.LString(fmt.Sprint(valueField))
				default:
					converted = valueField
				}

				result.RawSetInt(i, converted)
			}
		}

		L.Push(result)
		return 1
	}))
}
