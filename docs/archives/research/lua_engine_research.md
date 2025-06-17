# Lua Engine Research - GopherLua Integration

## Overview
GopherLua (github.com/yuin/gopher-lua) is a Lua 5.1 VM and compiler written in Go. It's designed to be easily embedded in Go applications while providing a user-friendly API.

## Key Findings

### 1. Version and Compatibility
- **Current Version**: Implements Lua 5.1 with select Lua 5.2 features (goto statement)
- **Go Compatibility**: Requires Go 1.9+
- **License**: MIT License
- **Repository**: https://github.com/yuin/gopher-lua

### 2. Core Architecture

#### LState Management
- Each Lua VM instance is represented by an `LState` struct
- LState is NOT thread-safe - each goroutine needs its own LState
- States can be pooled for performance optimization

```go
L := lua.NewState()
defer L.Close()
```

#### API Design Philosophy
- Not stack-based like C Lua API
- Prioritizes user-friendliness over raw performance
- Direct method calls instead of stack manipulation

### 3. Type System and Conversion

#### Lua Types (LValue interface)
- `LNil` - nil value
- `LBool` - boolean
- `LNumber` - float64
- `LString` - string
- `LFunction` - Lua function
- `LTable` - Lua table
- `LUserData` - custom Go types
- `LChannel` - Go channels (gopher-lua specific)

#### Type Conversion Methods
```go
// Go to Lua
L.Push(lua.LString("hello"))
L.Push(lua.LNumber(42))
L.Push(lua.LBool(true))

// Lua to Go
lv := L.Get(-1) // Get from stack
if str, ok := lv.(lua.LString); ok {
    goStr := string(str)
}
```

### 4. Module System

#### Standard Libraries
- `base` - Basic functions
- `coroutine` - Coroutine support
- `channel` - Go channel support (gopher-lua specific)
- `table` - Table manipulation
- `io` - I/O facilities
- `os` - Operating system facilities
- `string` - String manipulation
- `math` - Math functions
- `debug` - Debug facilities
- `package` - Module system

#### Custom Module Loading
```go
L.PreloadModule("mymodule", myModuleLoader)
```

### 5. Performance Characteristics
- **Speed**: Comparable to Python3 in micro-benchmarks
- **Memory**: Configurable registry and stack sizes
- **Optimization Options**:
  - State pooling
  - Compiled chunk caching
  - Registry size tuning

### 6. Security Features
- **Sandboxing**: Can restrict library access
- **Resource Limits**:
  - Instruction count limits
  - Memory limits via registry size
  - Call stack depth limits
- **Module Restrictions**: Can control which modules are available

### 7. Goroutine Integration
- **Channels**: Native Go channel support via LChannel type
- **Concurrency**: Each goroutine needs separate LState
- **Communication**: Can share data via channels

### 8. Memory Management
- **GC Integration**: Works with Go's garbage collector
- **Registry**: Configurable size and growth
- **Call Stack**: Configurable depth
- **State Pooling**: Can reuse LState instances

## Implementation Considerations for go-llmspell

### 1. State Management Strategy
- Pool LState instances for performance
- One LState per script execution
- Clear state between executions for isolation

### 2. Type Conversion Architecture
- Implement bidirectional converters for:
  - ScriptValue ↔ LValue
  - Bridge types ↔ Lua tables/userdata
  - Error handling and type validation

### 3. Security Sandbox Design
- Remove dangerous libraries (io, os, debug)
- Implement instruction count limits
- Control memory usage via registry limits
- Whitelist safe functions only

### 4. Module System Integration
- Create Lua wrappers for each bridge
- Use PreloadModule for bridge registration
- Implement lazy loading for efficiency

### 5. Performance Optimizations
- LState pooling with sync.Pool
- Compiled script caching
- Minimal type conversions
- Efficient error propagation

### 6. Error Handling
- Convert Lua errors to Go errors
- Maintain stack traces
- Provide clear error messages
- Handle panics gracefully

## Additional Research Needed
1. Lua 5.4 features and compatibility
2. Advanced coroutine patterns
3. Memory profiling techniques
4. Benchmarking methodology
5. Integration with go-llms type system

## Recommended Next Steps
1. Create prototype implementation
2. Benchmark against requirements
3. Design comprehensive test suite
4. Document Lua API for scripts
5. Implement security sandbox