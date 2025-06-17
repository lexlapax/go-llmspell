# Lua Type Conversion Design - ScriptValue ↔ LValue

## Overview
This document designs the bidirectional type conversion system between go-llmspell's ScriptValue types and GopherLua's LValue types.

## Type Mapping

### Basic Type Mappings
| Go Type | ScriptValue | LValue Type | Notes |
|---------|-------------|-------------|-------|
| `nil` | `nil` | `LNil` | Direct mapping |
| `bool` | `bool` | `LBool` | Direct mapping |
| `float64` | `number` | `LNumber` | All numbers are float64 |
| `string` | `string` | `LString` | Direct mapping |
| `[]interface{}` | `array` | `*LTable` | Sequential table |
| `map[string]interface{}` | `object` | `*LTable` | Hash table |
| `Function` | `function` | `*LFunction` | Wrapped Go function |
| `error` | `error` | `*LUserData` | Custom error type |
| `chan interface{}` | `channel` | `*LChannel` | Go channels |

### Complex Type Mappings
| Go Type | LValue Strategy | Notes |
|---------|-----------------|-------|
| Bridge objects | `*LUserData` | Wrapped with metadata |
| Custom structs | `*LTable` or `*LUserData` | Configurable |
| Time/Duration | `*LUserData` | With string conversion |
| Enums | `LString` or `LNumber` | Based on underlying type |

## Conversion Architecture

```go
// LuaTypeConverter implements TypeConverter for Lua engine
type LuaTypeConverter struct {
    *BaseTypeConverter
    
    // Conversion options
    config ConversionConfig
    
    // Type registries
    userDataTypes map[reflect.Type]UserDataHandler
    tableHandlers map[reflect.Type]TableHandler
    
    // Performance optimizations
    stringCache  map[string]lua.LString
    numberCache  map[float64]lua.LNumber
    
    // Circular reference detection
    visited map[uintptr]lua.LValue
}

// ConversionConfig controls conversion behavior
type ConversionConfig struct {
    // Table conversion
    ArrayMetatable  string // Metatable name for arrays
    ObjectMetatable string // Metatable name for objects
    MaxTableDepth   int    // Maximum nesting depth
    
    // UserData handling
    PreferUserData  bool   // Prefer UserData over Table for structs
    ExposePrivate   bool   // Expose private fields
    
    // Performance
    EnableCaching   bool   // Cache common values
    CacheSize       int    // Max cache entries
    
    // Safety
    DetectCycles    bool   // Detect circular references
    MaxStringLength int    // Limit string conversions
}
```

## Core Conversion Methods

### Go to Lua Conversion

```go
// ToLValue converts any Go value to Lua value
func (c *LuaTypeConverter) ToLValue(L *lua.LState, v interface{}) (lua.LValue, error) {
    // Handle nil
    if v == nil {
        return lua.LNil, nil
    }
    
    // Check cache for common values
    if c.config.EnableCaching {
        if cached := c.checkCache(v); cached != nil {
            return cached, nil
        }
    }
    
    // Handle basic types
    switch val := v.(type) {
    case lua.LValue:
        return val, nil // Already LValue
        
    case bool:
        if val {
            return lua.LTrue, nil
        }
        return lua.LFalse, nil
        
    case string:
        return c.toLString(val), nil
        
    case int, int8, int16, int32, int64:
        return lua.LNumber(reflect.ValueOf(val).Int()), nil
        
    case uint, uint8, uint16, uint32, uint64:
        return lua.LNumber(reflect.ValueOf(val).Uint()), nil
        
    case float32, float64:
        return lua.LNumber(reflect.ValueOf(val).Float()), nil
        
    case []interface{}:
        return c.arrayToTable(L, val)
        
    case map[string]interface{}:
        return c.mapToTable(L, val)
        
    case error:
        return c.errorToUserData(L, val)
        
    case Function:
        return c.functionToLFunction(L, val)
        
    default:
        return c.complexToLValue(L, v)
    }
}

// arrayToTable converts Go slice to Lua table
func (c *LuaTypeConverter) arrayToTable(L *lua.LState, arr []interface{}) (*lua.LTable, error) {
    if c.config.DetectCycles {
        ptr := reflect.ValueOf(arr).Pointer()
        if existing, found := c.visited[ptr]; found {
            return existing.(*lua.LTable), nil
        }
    }
    
    table := L.NewTable()
    
    if c.config.DetectCycles {
        c.visited[reflect.ValueOf(arr).Pointer()] = table
    }
    
    // Set array metatable if configured
    if c.config.ArrayMetatable != "" {
        L.SetMetatable(table, L.GetGlobal(c.config.ArrayMetatable))
    }
    
    // Convert elements
    for i, elem := range arr {
        lv, err := c.ToLValue(L, elem)
        if err != nil {
            return nil, fmt.Errorf("array[%d]: %w", i, err)
        }
        table.RawSetInt(i+1, lv) // Lua arrays are 1-indexed
    }
    
    return table, nil
}

// mapToTable converts Go map to Lua table
func (c *LuaTypeConverter) mapToTable(L *lua.LState, m map[string]interface{}) (*lua.LTable, error) {
    if c.config.DetectCycles {
        ptr := reflect.ValueOf(m).Pointer()
        if existing, found := c.visited[ptr]; found {
            return existing.(*lua.LTable), nil
        }
    }
    
    table := L.NewTable()
    
    if c.config.DetectCycles {
        c.visited[reflect.ValueOf(m).Pointer()] = table
    }
    
    // Set object metatable if configured
    if c.config.ObjectMetatable != "" {
        L.SetMetatable(table, L.GetGlobal(c.config.ObjectMetatable))
    }
    
    // Convert entries
    for k, v := range m {
        lv, err := c.ToLValue(L, v)
        if err != nil {
            return nil, fmt.Errorf("map[%s]: %w", k, err)
        }
        table.RawSetString(k, lv)
    }
    
    return table, nil
}
```

### Lua to Go Conversion

```go
// FromLValue converts Lua value to Go value
func (c *LuaTypeConverter) FromLValue(lv lua.LValue) (interface{}, error) {
    switch lv.Type() {
    case lua.LTNil:
        return nil, nil
        
    case lua.LTBool:
        return bool(lv.(lua.LBool)), nil
        
    case lua.LTNumber:
        return float64(lv.(lua.LNumber)), nil
        
    case lua.LTString:
        return string(lv.(lua.LString)), nil
        
    case lua.LTTable:
        return c.tableToGo(lv.(*lua.LTable))
        
    case lua.LTFunction:
        return c.lfunctionToGo(lv.(*lua.LFunction))
        
    case lua.LTUserData:
        return c.userDataToGo(lv.(*lua.LUserData))
        
    case lua.LTChannel:
        return lv.(*lua.LChannel).Value, nil
        
    default:
        return nil, fmt.Errorf("unsupported LValue type: %s", lv.Type())
    }
}

// tableToGo converts Lua table to Go value (array or map)
func (c *LuaTypeConverter) tableToGo(table *lua.LTable) (interface{}, error) {
    // Check for circular reference
    if c.config.DetectCycles {
        ptr := reflect.ValueOf(table).Pointer()
        if existing, found := c.visited[ptr]; found {
            return existing, nil
        }
    }
    
    // Detect if table is array or map
    if c.isArray(table) {
        return c.tableToArray(table)
    }
    return c.tableToMap(table)
}

// isArray checks if table is array-like (sequential integer keys from 1)
func (c *LuaTypeConverter) isArray(table *lua.LTable) bool {
    length := table.Len()
    if length == 0 {
        // Empty table - check if any non-integer keys exist
        hasNonIntKey := false
        table.ForEach(func(k, v lua.LValue) {
            if k.Type() != lua.LTNumber {
                hasNonIntKey = true
            }
        })
        return !hasNonIntKey
    }
    
    // Check if all keys are sequential integers starting from 1
    for i := 1; i <= length; i++ {
        if table.RawGetInt(i) == lua.LNil {
            return false
        }
    }
    
    // Verify no additional keys beyond length
    count := 0
    table.ForEach(func(k, v lua.LValue) {
        count++
    })
    
    return count == length
}
```

## Bridge Object Handling

```go
// BridgeUserData wraps bridge objects for Lua
type BridgeUserData struct {
    Bridge   interface{}         // The actual bridge object
    TypeName string              // Type identifier
    Methods  map[string]lua.LGFunction // Cached methods
}

// bridgeToUserData converts bridge object to UserData
func (c *LuaTypeConverter) bridgeToUserData(L *lua.LState, bridge interface{}) *lua.LUserData {
    ud := L.NewUserData()
    ud.Value = &BridgeUserData{
        Bridge:   bridge,
        TypeName: reflect.TypeOf(bridge).String(),
        Methods:  make(map[string]lua.LGFunction),
    }
    
    // Set bridge metatable for method access
    L.SetMetatable(ud, c.getBridgeMetatable(L, bridge))
    
    return ud
}

// getBridgeMetatable creates or retrieves metatable for bridge type
func (c *LuaTypeConverter) getBridgeMetatable(L *lua.LState, bridge interface{}) *lua.LTable {
    typeName := reflect.TypeOf(bridge).String()
    
    // Check if metatable already exists
    L.GetGlobal("__bridge_metatables")
    if !lua.LVIsFalse(L.Get(-1)) {
        mt := L.GetField(L.Get(-1), typeName)
        if !lua.LVIsFalse(mt) {
            L.Pop(1)
            return mt.(*lua.LTable)
        }
    }
    L.Pop(1)
    
    // Create new metatable
    mt := L.NewTable()
    
    // __index for method access
    L.SetField(mt, "__index", L.NewFunction(c.bridgeIndex))
    
    // __tostring for debugging
    L.SetField(mt, "__tostring", L.NewFunction(c.bridgeToString))
    
    // Store metatable
    L.GetGlobal("__bridge_metatables")
    if lua.LVIsFalse(L.Get(-1)) {
        L.Pop(1)
        L.SetGlobal("__bridge_metatables", L.NewTable())
        L.GetGlobal("__bridge_metatables")
    }
    L.SetField(L.Get(-1), typeName, mt)
    L.Pop(1)
    
    return mt
}
```

## Performance Optimizations

```go
// String caching for common values
func (c *LuaTypeConverter) toLString(s string) lua.LString {
    if !c.config.EnableCaching {
        return lua.LString(s)
    }
    
    if cached, found := c.stringCache[s]; found {
        return cached
    }
    
    // Limit cache size
    if len(c.stringCache) >= c.config.CacheSize {
        // Simple eviction: clear half the cache
        count := 0
        for k := range c.stringCache {
            delete(c.stringCache, k)
            count++
            if count >= c.config.CacheSize/2 {
                break
            }
        }
    }
    
    ls := lua.LString(s)
    c.stringCache[s] = ls
    return ls
}

// Number caching for common values (-100 to 100, common floats)
func (c *LuaTypeConverter) toLNumber(n float64) lua.LNumber {
    if !c.config.EnableCaching {
        return lua.LNumber(n)
    }
    
    // Only cache "nice" numbers
    if n == float64(int64(n)) && n >= -100 && n <= 100 {
        if cached, found := c.numberCache[n]; found {
            return cached
        }
        c.numberCache[n] = lua.LNumber(n)
        return c.numberCache[n]
    }
    
    return lua.LNumber(n)
}
```

## Error Handling

```go
// ConversionError provides detailed error context
type ConversionError struct {
    Path      []string      // Conversion path (e.g., ["root", "field1", "[2]"])
    FromType  reflect.Type  // Source Go type
    ToType    string        // Target Lua type
    Reason    string        // Error reason
    Wrapped   error         // Underlying error
}

func (e *ConversionError) Error() string {
    path := strings.Join(e.Path, ".")
    return fmt.Sprintf("conversion error at %s: cannot convert %v to %s: %s",
        path, e.FromType, e.ToType, e.Reason)
}

// errorToUserData converts Go error to Lua UserData
func (c *LuaTypeConverter) errorToUserData(L *lua.LState, err error) *lua.LUserData {
    ud := L.NewUserData()
    ud.Value = err
    
    // Set error metatable
    mt := L.NewTable()
    L.SetField(mt, "__tostring", L.NewFunction(func(L *lua.LState) int {
        ud := L.CheckUserData(1)
        if err, ok := ud.Value.(error); ok {
            L.Push(lua.LString(err.Error()))
        } else {
            L.Push(lua.LString("error"))
        }
        return 1
    }))
    
    L.SetMetatable(ud, mt)
    return ud
}
```

## Special Type Handlers

```go
// UserDataHandler customizes UserData conversion
type UserDataHandler interface {
    ToUserData(L *lua.LState, v interface{}) (*lua.LUserData, error)
    FromUserData(ud *lua.LUserData) (interface{}, error)
}

// TableHandler customizes Table conversion
type TableHandler interface {
    ToTable(L *lua.LState, v interface{}) (*lua.LTable, error)
    FromTable(table *lua.LTable) (interface{}, error)
}

// RegisterUserDataType registers custom UserData handler
func (c *LuaTypeConverter) RegisterUserDataType(t reflect.Type, handler UserDataHandler) {
    c.userDataTypes[t] = handler
}

// RegisterTableType registers custom Table handler
func (c *LuaTypeConverter) RegisterTableType(t reflect.Type, handler TableHandler) {
    c.tableHandlers[t] = handler
}
```

## Function Conversion

```go
// functionToLFunction wraps Go function as Lua function
func (c *LuaTypeConverter) functionToLFunction(L *lua.LState, fn Function) *lua.LFunction {
    return L.NewFunction(func(L *lua.LState) int {
        // Extract arguments
        nargs := L.GetTop()
        args := make([]interface{}, nargs)
        for i := 0; i < nargs; i++ {
            arg, err := c.FromLValue(L.Get(i + 1))
            if err != nil {
                L.RaiseError("argument %d: %v", i+1, err)
                return 0
            }
            args[i] = arg
        }
        
        // Call Go function
        ctx := context.Background() // TODO: Get from Lua registry
        results, err := fn.Call(ctx, args)
        if err != nil {
            L.RaiseError("function error: %v", err)
            return 0
        }
        
        // Convert results
        for _, result := range results {
            lv, err := c.ToLValue(L, result)
            if err != nil {
                L.RaiseError("result conversion: %v", err)
                return 0
            }
            L.Push(lv)
        }
        
        return len(results)
    })
}

// lfunctionToGo wraps Lua function as Go function
func (c *LuaTypeConverter) lfunctionToGo(fn *lua.LFunction) Function {
    return &LuaFunction{
        fn:        fn,
        converter: c,
    }
}

// LuaFunction wraps Lua function as Go Function
type LuaFunction struct {
    fn        *lua.LFunction
    converter *LuaTypeConverter
    L         *lua.LState // Need access to LState for calls
}

func (f *LuaFunction) Call(ctx context.Context, args []interface{}) ([]interface{}, error) {
    // This is tricky - need LState access
    // Will be resolved in engine implementation
    return nil, fmt.Errorf("Lua function calls require LState context")
}
```

## Testing Strategy

1. **Roundtrip Tests**: Verify Go→Lua→Go conversions
2. **Edge Cases**: nil, empty collections, deep nesting
3. **Performance Tests**: Benchmark conversions
4. **Memory Tests**: Verify no leaks with cycles
5. **Error Tests**: Invalid conversions, limits

## Integration Points

1. **With LuaEngine**: Converter instance per engine
2. **With Bridges**: Auto-convert bridge returns
3. **With Sandbox**: Respect type restrictions
4. **With Pool**: Reset converter state

## Configuration Defaults

```go
func DefaultConversionConfig() ConversionConfig {
    return ConversionConfig{
        ArrayMetatable:  "__array",
        ObjectMetatable: "__object",
        MaxTableDepth:   100,
        PreferUserData:  true,
        ExposePrivate:   false,
        EnableCaching:   true,
        CacheSize:       1000,
        DetectCycles:    true,
        MaxStringLength: 1024 * 1024, // 1MB
    }
}
```