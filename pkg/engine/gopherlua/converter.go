// ABOUTME: LuaTypeConverter implements engine.TypeConverter for Go ↔ Lua type conversions
// ABOUTME: Handles ToLua, FromLua, circular reference detection, conversion caching, and custom type registration

package gopherlua

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// Default configuration constants
const (
	defaultMaxDepth  = 32
	defaultCacheSize = 1000
)

// LuaTypeConverter implements engine.TypeConverter for Lua engine
type LuaTypeConverter struct {
	mu              sync.RWMutex
	customTypes     map[string]*customTypeConverter
	conversionCache *conversionCache
	maxDepth        int
	cacheSize       int
	scriptConverter *ScriptValueConverter
}

// LuaTypeConverterConfig provides configuration options for the converter
type LuaTypeConverterConfig struct {
	MaxDepth  int
	CacheSize int
}

// customTypeConverter stores custom conversion functions
type customTypeConverter struct {
	toLua   func(*lua.LState, interface{}) (lua.LValue, error)
	fromLua func(lua.LValue) (interface{}, error)
}

// conversionCache implements LRU cache for conversion results
type conversionCache struct {
	mu        sync.RWMutex
	data      map[string]interface{}
	order     []string
	maxSize   int
	hits      int64
	misses    int64
	evictions int64
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
}

// NewLuaTypeConverter creates a new type converter with default configuration
func NewLuaTypeConverter() *LuaTypeConverter {
	return NewLuaTypeConverterWithConfig(LuaTypeConverterConfig{
		MaxDepth:  defaultMaxDepth,
		CacheSize: defaultCacheSize,
	})
}

// NewLuaTypeConverterWithConfig creates a new type converter with custom configuration
func NewLuaTypeConverterWithConfig(config LuaTypeConverterConfig) *LuaTypeConverter {
	if config.MaxDepth <= 0 {
		config.MaxDepth = defaultMaxDepth
	}
	if config.CacheSize <= 0 {
		config.CacheSize = defaultCacheSize
	}

	converter := &LuaTypeConverter{
		customTypes:     make(map[string]*customTypeConverter),
		conversionCache: newConversionCache(config.CacheSize),
		maxDepth:        config.MaxDepth,
		cacheSize:       config.CacheSize,
	}

	// Initialize the ScriptValue converter with a reference to this converter
	converter.scriptConverter = NewScriptValueConverter(converter)

	return converter
}

// newConversionCache creates a new LRU cache
func newConversionCache(maxSize int) *conversionCache {
	return &conversionCache{
		data:    make(map[string]interface{}),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// ToLua converts a Go value to a Lua value
func (ltc *LuaTypeConverter) ToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	// Check cache first for simple types
	if ltc.isCacheable(value) {
		cacheKey := ltc.generateCacheKey(value)
		if cached := ltc.conversionCache.get(cacheKey); cached != nil {
			ltc.conversionCache.recordHit()
			// Note: We can't cache Lua values directly as they're tied to specific LState
			// This is a simplified cache implementation for demonstration
		} else {
			ltc.conversionCache.recordMiss()
		}
	}

	return ltc.toLuaWithDepth(L, value, 0, make(map[uintptr]bool))
}

// toLuaWithDepth performs the actual conversion with depth and circular reference tracking
func (ltc *LuaTypeConverter) toLuaWithDepth(L *lua.LState, value interface{}, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	// Check depth limit
	if depth > ltc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", ltc.maxDepth)
	}

	// Handle nil
	if value == nil {
		return lua.LNil, nil
	}

	// Get reflection info
	rv := reflect.ValueOf(value)
	rt := rv.Type()

	// Check for circular references in maps and slices
	if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice {
		if rv.IsValid() && rv.Pointer() != 0 {
			if visited[rv.Pointer()] {
				return nil, fmt.Errorf("circular reference detected")
			}
			visited[rv.Pointer()] = true
			defer delete(visited, rv.Pointer())
		}
	}

	// Check for custom type converter
	ltc.mu.RLock()
	typeName := rt.String()
	// Also check simple type name (without package)
	if rt.Name() != "" {
		if converter, exists := ltc.customTypes[rt.Name()]; exists {
			ltc.mu.RUnlock()
			return converter.toLua(L, value)
		}
	}
	if converter, exists := ltc.customTypes[typeName]; exists {
		ltc.mu.RUnlock()
		return converter.toLua(L, value)
	}
	ltc.mu.RUnlock()

	// Handle standard types
	switch rv.Kind() {
	case reflect.Bool:
		return lua.LBool(rv.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(rv.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return lua.LNumber(rv.Float()), nil

	case reflect.String:
		return lua.LString(rv.String()), nil

	case reflect.Slice, reflect.Array:
		return ltc.sliceToLuaTable(L, rv, depth, visited)

	case reflect.Map:
		return ltc.mapToLuaTable(L, rv, depth, visited)

	case reflect.Struct:
		return ltc.structToLuaTable(L, rv, depth, visited)

	case reflect.Ptr:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return ltc.toLuaWithDepth(L, rv.Elem().Interface(), depth+1, visited)

	case reflect.Interface:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return ltc.toLuaWithDepth(L, rv.Elem().Interface(), depth+1, visited)

	default:
		return nil, fmt.Errorf("unsupported type for conversion to Lua: %s", rt.String())
	}
}

// sliceToLuaTable converts a Go slice/array to a Lua table
func (ltc *LuaTypeConverter) sliceToLuaTable(L *lua.LState, rv reflect.Value, depth int, visited map[uintptr]bool) (*lua.LTable, error) {
	table := L.NewTable()

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		luaValue, err := ltc.toLuaWithDepth(L, elem, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert slice element %d: %w", i, err)
		}
		table.RawSetInt(i+1, luaValue) // Lua arrays are 1-indexed
	}

	return table, nil
}

// mapToLuaTable converts a Go map to a Lua table
func (ltc *LuaTypeConverter) mapToLuaTable(L *lua.LState, rv reflect.Value, depth int, visited map[uintptr]bool) (*lua.LTable, error) {
	table := L.NewTable()

	for _, key := range rv.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		value := rv.MapIndex(key).Interface()

		luaValue, err := ltc.toLuaWithDepth(L, value, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert map value for key %s: %w", keyStr, err)
		}

		table.RawSetString(keyStr, luaValue)
	}

	return table, nil
}

// structToLuaTable converts a Go struct to a Lua table
func (ltc *LuaTypeConverter) structToLuaTable(L *lua.LState, rv reflect.Value, depth int, visited map[uintptr]bool) (*lua.LTable, error) {
	table := L.NewTable()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldValue := rv.Field(i).Interface()
		luaValue, err := ltc.toLuaWithDepth(L, fieldValue, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert struct field %s: %w", field.Name, err)
		}

		table.RawSetString(field.Name, luaValue)
	}

	return table, nil
}

// FromLua converts a Lua value to a Go value
func (ltc *LuaTypeConverter) FromLua(value lua.LValue) (interface{}, error) {
	return ltc.fromLuaWithDepth(value, 0, make(map[*lua.LTable]bool))
}

// fromLuaWithDepth performs the actual conversion with depth and circular reference tracking
func (ltc *LuaTypeConverter) fromLuaWithDepth(value lua.LValue, depth int, visited map[*lua.LTable]bool) (interface{}, error) {
	// Check depth limit
	if depth > ltc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", ltc.maxDepth)
	}

	switch lv := value.(type) {
	case lua.LBool:
		return bool(lv), nil

	case lua.LNumber:
		return float64(lv), nil

	case lua.LString:
		return string(lv), nil

	case *lua.LNilType:
		return nil, nil

	case *lua.LTable:
		// Check for circular references
		if visited[lv] {
			return nil, fmt.Errorf("circular reference detected in Lua table")
		}
		visited[lv] = true
		defer delete(visited, lv)

		return ltc.luaTableToGo(lv, depth, visited)

	case *lua.LUserData:
		// Handle user data - return the Go value if available
		return lv.Value, nil

	case *lua.LFunction:
		// Functions are not directly convertible - return a placeholder
		return fmt.Sprintf("function<%p>", lv), nil

	default:
		return nil, fmt.Errorf("unsupported Lua type for conversion to Go: %T", value)
	}
}

// luaTableToGo converts a Lua table to appropriate Go type (slice or map)
func (ltc *LuaTypeConverter) luaTableToGo(table *lua.LTable, depth int, visited map[*lua.LTable]bool) (interface{}, error) {
	// Check if this is an array-like table (consecutive integer keys starting from 1)
	if ltc.isArrayLikeTable(table) {
		return ltc.luaTableToSlice(table, depth, visited)
	}

	// Otherwise, treat as a map
	return ltc.luaTableToMap(table, depth, visited)
}

// isArrayLikeTable checks if a Lua table has consecutive integer keys starting from 1
func (ltc *LuaTypeConverter) isArrayLikeTable(table *lua.LTable) bool {
	length := table.Len()
	if length == 0 {
		// Empty table - could be either, but we'll treat as slice for consistency
		hasStringKeys := false
		table.ForEach(func(key, value lua.LValue) {
			if _, ok := key.(lua.LString); ok {
				hasStringKeys = true
			}
		})
		return !hasStringKeys
	}

	// Check if all keys from 1 to length exist
	for i := 1; i <= length; i++ {
		if table.RawGetInt(i) == lua.LNil {
			return false
		}
	}

	// Check if there are any non-integer keys
	hasNonIntegerKeys := false
	table.ForEach(func(key, value lua.LValue) {
		if num, ok := key.(lua.LNumber); ok {
			if float64(num) != float64(int(num)) || int(num) < 1 || int(num) > length {
				hasNonIntegerKeys = true
			}
		} else {
			hasNonIntegerKeys = true
		}
	})

	return !hasNonIntegerKeys
}

// luaTableToSlice converts a Lua table to a Go slice
func (ltc *LuaTypeConverter) luaTableToSlice(table *lua.LTable, depth int, visited map[*lua.LTable]bool) ([]interface{}, error) {
	length := table.Len()
	result := make([]interface{}, length)

	for i := 1; i <= length; i++ {
		luaValue := table.RawGetInt(i)
		goValue, err := ltc.fromLuaWithDepth(luaValue, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array element %d: %w", i, err)
		}
		result[i-1] = goValue // Convert from 1-indexed to 0-indexed
	}

	return result, nil
}

// luaTableToMap converts a Lua table to a Go map
func (ltc *LuaTypeConverter) luaTableToMap(table *lua.LTable, depth int, visited map[*lua.LTable]bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	table.ForEach(func(key, value lua.LValue) {
		keyStr := fmt.Sprintf("%v", key)
		goValue, err := ltc.fromLuaWithDepth(value, depth+1, visited)
		if err != nil {
			// Skip values that can't be converted rather than failing entirely
			return
		}
		result[keyStr] = goValue
	})

	return result, nil
}

// RegisterCustomType registers a custom type converter
func (ltc *LuaTypeConverter) RegisterCustomType(
	typeName string,
	toLua func(*lua.LState, interface{}) (lua.LValue, error),
	fromLua func(lua.LValue) (interface{}, error),
) error {
	ltc.mu.Lock()
	defer ltc.mu.Unlock()

	if _, exists := ltc.customTypes[typeName]; exists {
		return fmt.Errorf("type converter for %s is already registered", typeName)
	}

	ltc.customTypes[typeName] = &customTypeConverter{
		toLua:   toLua,
		fromLua: fromLua,
	}

	return nil
}

// GetCacheStats returns cache performance statistics
func (ltc *LuaTypeConverter) GetCacheStats() CacheStats {
	ltc.conversionCache.mu.RLock()
	defer ltc.conversionCache.mu.RUnlock()

	return CacheStats{
		Hits:      ltc.conversionCache.hits,
		Misses:    ltc.conversionCache.misses,
		Evictions: ltc.conversionCache.evictions,
		Size:      len(ltc.conversionCache.data),
	}
}

// Cache helper methods
func (ltc *LuaTypeConverter) isCacheable(value interface{}) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	default:
		return false // Don't cache complex types
	}
}

func (ltc *LuaTypeConverter) generateCacheKey(value interface{}) string {
	return fmt.Sprintf("%T:%v", value, value)
}

func (cc *conversionCache) get(key string) interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data[key]
}

func (cc *conversionCache) recordHit() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.hits++
}

func (cc *conversionCache) recordMiss() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.misses++
}

// Implementation of engine.TypeConverter interface

// ToFunction converts a ScriptValue to engine.Function
func (ltc *LuaTypeConverter) ToFunction(v engine.ScriptValue) (engine.Function, error) {
	if v == nil || v.Type() != engine.TypeFunction {
		return nil, fmt.Errorf("cannot convert %s to Function", v.Type())
	}

	if fv, ok := v.(engine.FunctionValue); ok {
		if fn, ok := fv.Function().(engine.Function); ok {
			return fn, nil
		}
	}

	return nil, fmt.Errorf("ScriptValue does not contain a valid Function")
}

// FromFunction converts engine.Function to ScriptValue
func (ltc *LuaTypeConverter) FromFunction(fn engine.Function) (engine.ScriptValue, error) {
	return engine.NewFunctionValue("function", fn), nil
}

// SupportsType checks if the converter supports a given type
func (ltc *LuaTypeConverter) SupportsType(typeName string) bool {
	// Check custom types
	ltc.mu.RLock()
	_, exists := ltc.customTypes[typeName]
	ltc.mu.RUnlock()

	if exists {
		return true
	}

	// Check built-in types
	supportedTypes := map[string]bool{
		"bool":                   true,
		"int":                    true,
		"int8":                   true,
		"int16":                  true,
		"int32":                  true,
		"int64":                  true,
		"uint":                   true,
		"uint8":                  true,
		"uint16":                 true,
		"uint32":                 true,
		"uint64":                 true,
		"float32":                true,
		"float64":                true,
		"string":                 true,
		"[]interface{}":          true,
		"map[string]interface{}": true,
	}

	return supportedTypes[typeName]
}

// GetTypeInfo returns information about a supported type
func (ltc *LuaTypeConverter) GetTypeInfo(typeName string) engine.TypeInfo {
	// This is a placeholder implementation
	return engine.TypeInfo{
		Name:        typeName,
		Category:    engine.TypeCategoryPrimitive,
		Description: fmt.Sprintf("Type information for %s", typeName),
		Methods:     []string{},
		Properties:  []string{},
		Metadata:    make(map[string]interface{}),
	}
}

// ScriptValue integration methods

// ToScriptValue converts a Go value to a ScriptValue
func (ltc *LuaTypeConverter) ToScriptValue(value interface{}) (engine.ScriptValue, error) {
	return ltc.scriptConverter.GoToScriptValue(value)
}

// FromScriptValue converts a ScriptValue to a Go value
func (ltc *LuaTypeConverter) FromScriptValue(sv engine.ScriptValue) interface{} {
	return ltc.scriptConverter.ScriptValueToGo(sv)
}

// ToLuaScriptValue converts a lua.LValue to a ScriptValue
func (ltc *LuaTypeConverter) ToLuaScriptValue(L *lua.LState, lv lua.LValue) (engine.ScriptValue, error) {
	return ltc.scriptConverter.LValueToScriptValue(L, lv)
}

// FromLuaScriptValue converts a ScriptValue to a lua.LValue
func (ltc *LuaTypeConverter) FromLuaScriptValue(L *lua.LState, sv engine.ScriptValue) (lua.LValue, error) {
	return ltc.scriptConverter.ScriptValueToLValue(L, sv)
}

// ToLuaWithScriptValue converts a Go value to Lua via ScriptValue (for consistency)
func (ltc *LuaTypeConverter) ToLuaWithScriptValue(L *lua.LState, value interface{}) (lua.LValue, error) {
	// First convert to ScriptValue
	sv, err := ltc.ToScriptValue(value)
	if err != nil {
		return nil, fmt.Errorf("error converting to ScriptValue: %w", err)
	}

	// Then convert ScriptValue to Lua
	return ltc.FromLuaScriptValue(L, sv)
}

// FromLuaWithScriptValue converts a Lua value to Go via ScriptValue (for consistency)
func (ltc *LuaTypeConverter) FromLuaWithScriptValue(L *lua.LState, lv lua.LValue) (interface{}, error) {
	// First convert to ScriptValue
	sv, err := ltc.ToLuaScriptValue(L, lv)
	if err != nil {
		return nil, fmt.Errorf("error converting to ScriptValue: %w", err)
	}

	// Then convert ScriptValue to Go
	return ltc.FromScriptValue(sv), nil
}

// Interface implementation methods for TypeConverter

// FromInterface converts a Go interface{} to ScriptValue
func (ltc *LuaTypeConverter) FromInterface(v interface{}) (engine.ScriptValue, error) {
	return ltc.ToScriptValue(v)
}

// ToInterface converts a ScriptValue to Go interface{}
func (ltc *LuaTypeConverter) ToInterface(v engine.ScriptValue) (interface{}, error) {
	if v == nil || v.IsNil() {
		return nil, nil
	}
	return ltc.FromScriptValue(v), nil
}

// ToBoolean converts a ScriptValue to boolean
func (ltc *LuaTypeConverter) ToBoolean(v engine.ScriptValue) (bool, error) {
	if v == nil || v.IsNil() {
		return false, nil
	}
	return engine.IsTrue(v), nil
}

// ToNumber converts a ScriptValue to float64
func (ltc *LuaTypeConverter) ToNumber(v engine.ScriptValue) (float64, error) {
	return engine.ConvertToNumber(v)
}

// ToString converts a ScriptValue to string
func (ltc *LuaTypeConverter) ToString(v engine.ScriptValue) (string, error) {
	return engine.ConvertToString(v)
}

// ToArray converts a ScriptValue to []ScriptValue
func (ltc *LuaTypeConverter) ToArray(v engine.ScriptValue) ([]engine.ScriptValue, error) {
	if v == nil || v.IsNil() {
		return nil, nil
	}

	if v.Type() == engine.TypeArray {
		if av, ok := v.(engine.ArrayValue); ok {
			return av.Elements(), nil
		}
	}

	return nil, fmt.Errorf("cannot convert %s to array", v.Type())
}

// ToMap converts a ScriptValue to map[string]ScriptValue
func (ltc *LuaTypeConverter) ToMap(v engine.ScriptValue) (map[string]engine.ScriptValue, error) {
	if v == nil || v.IsNil() {
		return nil, nil
	}

	if v.Type() == engine.TypeObject {
		if ov, ok := v.(engine.ObjectValue); ok {
			return ov.Fields(), nil
		}
	}

	return nil, fmt.Errorf("cannot convert %s to map", v.Type())
}

// ToStruct converts a ScriptValue to the target struct
func (ltc *LuaTypeConverter) ToStruct(v engine.ScriptValue, target interface{}) error {
	if v == nil || v.IsNil() {
		return nil
	}

	// Use reflection to convert object fields to struct
	if v.Type() != engine.TypeObject {
		return fmt.Errorf("cannot convert %s to struct, expected object", v.Type())
	}

	ov, ok := v.(engine.ObjectValue)
	if !ok {
		return fmt.Errorf("value is not an ObjectValue")
	}

	// Get target type
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetStruct := targetValue.Elem()
	targetType := targetStruct.Type()

	// Map object fields to struct fields
	fields := ov.Fields()
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		if !field.IsExported() {
			continue
		}

		if fieldValue, exists := fields[field.Name]; exists {
			goValue := ltc.FromScriptValue(fieldValue)
			fieldValueReflect := reflect.ValueOf(goValue)

			if fieldValueReflect.Type().ConvertibleTo(field.Type) {
				targetStruct.Field(i).Set(fieldValueReflect.Convert(field.Type))
			}
		}
	}

	return nil
}

// FromStruct converts a Go struct to ScriptValue
func (ltc *LuaTypeConverter) FromStruct(v interface{}) (engine.ScriptValue, error) {
	return ltc.ToScriptValue(v)
}
