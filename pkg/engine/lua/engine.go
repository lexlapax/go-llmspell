// ABOUTME: Lua script engine implementation using gopher-lua
// ABOUTME: Provides sandboxed Lua execution with Go function bindings

package lua

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine implements the Engine interface for Lua scripts
type LuaEngine struct {
	vm       *lua.LState
	config   *engine.Config
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	loaded   bool
	bindings map[string]interface{}
}

// NewLuaEngine creates a new Lua engine instance
func NewLuaEngine(config *engine.Config) (*LuaEngine, error) {
	if config == nil {
		config = &engine.Config{
			MaxExecutionTime: 30,               // 30 seconds default
			MaxMemory:        64 * 1024 * 1024, // 64MB default
		}
	}

	engine := &LuaEngine{
		config:   config,
		bindings: make(map[string]interface{}),
	}

	// Initialize the Lua VM
	if err := engine.initVM(); err != nil {
		return nil, err
	}

	return engine, nil
}

// initVM initializes a new Lua VM with security settings
func (e *LuaEngine) initVM() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Close existing VM if any
	if e.vm != nil {
		e.vm.Close()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.MaxExecutionTime)*time.Second)
	e.ctx = ctx
	e.cancel = cancel

	// Create new Lua state
	e.vm = lua.NewState()

	// Configure VM options
	e.vm.SetContext(ctx)

	// Disable dangerous functions for security
	e.vm.SetGlobal("dofile", lua.LNil)
	e.vm.SetGlobal("loadfile", lua.LNil)
	e.vm.SetGlobal("load", lua.LNil)
	e.vm.SetGlobal("loadstring", lua.LNil)
	e.vm.SetGlobal("require", lua.LNil)

	// Disable io and os libraries for security
	e.vm.SetGlobal("io", lua.LNil)
	e.vm.SetGlobal("os", lua.LNil)

	// Disable debug library
	e.vm.SetGlobal("debug", lua.LNil)

	// Register all previously registered bindings
	for name, fn := range e.bindings {
		if err := e.registerFunctionInternal(name, fn); err != nil {
			return fmt.Errorf("failed to re-register function %s: %w", name, err)
		}
	}

	e.loaded = false
	return nil
}

// Name returns the name of the engine
func (e *LuaEngine) Name() string {
	return "lua"
}

// LoadScript loads a script from a reader
func (e *LuaEngine) LoadScript(reader io.Reader) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.vm == nil {
		return fmt.Errorf("Lua VM not initialized")
	}

	// Read the script
	script, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	// Compile the script
	fn, err := e.vm.LoadString(string(script))
	if err != nil {
		return fmt.Errorf("failed to compile script: %w", err)
	}

	// Push the compiled function onto the stack
	e.vm.Push(fn)

	e.loaded = true
	return nil
}

// LoadScriptFile loads a script from a file path
func (e *LuaEngine) LoadScriptFile(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.vm == nil {
		return fmt.Errorf("Lua VM not initialized")
	}

	// Load and compile the script file
	fn, err := e.vm.LoadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load script file: %w", err)
	}

	// Push the compiled function onto the stack
	e.vm.Push(fn)

	e.loaded = true
	return nil
}

// Execute runs the loaded script
func (e *LuaEngine) Execute(ctx context.Context) error {
	e.mu.Lock()

	if !e.loaded {
		e.mu.Unlock()
		return fmt.Errorf("no script loaded")
	}

	// Update VM context
	e.vm.SetContext(ctx)

	// Run the script (synchronously to avoid race conditions)
	err := e.vm.PCall(0, lua.MultRet, nil)
	e.mu.Unlock()

	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}

// RegisterFunction registers a Go function to be callable from Lua
func (e *LuaEngine) RegisterFunction(name string, fn interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Store the binding for re-registration after VM reset
	e.bindings[name] = fn

	return e.registerFunctionInternal(name, fn)
}

// registerFunctionInternal registers a function without locking
func (e *LuaEngine) registerFunctionInternal(name string, fn interface{}) error {
	if e.vm == nil {
		return fmt.Errorf("Lua VM not initialized")
	}

	// Wrap the Go function for Lua
	luaFn := wrapGoFunction(fn)
	e.vm.SetGlobal(name, e.vm.NewFunction(luaFn))

	return nil
}

// SetVariable sets a variable in the Lua context
func (e *LuaEngine) SetVariable(name string, value interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.vm == nil {
		return fmt.Errorf("Lua VM not initialized")
	}

	// Convert Go value to Lua value
	luaValue, err := goToLua(e.vm, value)
	if err != nil {
		return fmt.Errorf("failed to convert value: %w", err)
	}

	e.vm.SetGlobal(name, luaValue)
	return nil
}

// GetVariable gets a variable from the Lua context
func (e *LuaEngine) GetVariable(name string) (interface{}, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.vm == nil {
		return nil, fmt.Errorf("Lua VM not initialized")
	}

	luaValue := e.vm.GetGlobal(name)
	return luaToGo(luaValue), nil
}

// Close cleans up the Lua engine
func (e *LuaEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		e.cancel()
	}

	if e.vm != nil {
		e.vm.Close()
		e.vm = nil
	}

	return nil
}

// GetLuaState returns the underlying Lua state for advanced usage
// This is needed for registering complex bridges
func (e *LuaEngine) GetLuaState() *lua.LState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vm
}

// Reset resets the Lua engine to a clean state
func (e *LuaEngine) Reset() error {
	return e.initVM()
}

// wrapGoFunction wraps a Go function to be callable from Lua
func wrapGoFunction(fn interface{}) lua.LGFunction {
	return func(L *lua.LState) int {
		// This is a simplified wrapper - in production, we'd use reflection
		// to properly handle different function signatures

		// For now, we'll handle common cases manually
		// This will be expanded in the conversions.go file

		// Example: function that takes a string and returns a string
		if f, ok := fn.(func(string) string); ok {
			arg := L.CheckString(1)
			result := f(arg)
			L.Push(lua.LString(result))
			return 1
		}

		// Function that takes no args and returns a string
		if f, ok := fn.(func() string); ok {
			result := f()
			L.Push(lua.LString(result))
			return 1
		}

		// Function that takes string and int, returns string
		if f, ok := fn.(func(string, int) string); ok {
			arg1 := L.CheckString(1)
			arg2 := L.CheckInt(2)
			result := f(arg1, arg2)
			L.Push(lua.LString(result))
			return 1
		}

		// Function that returns []string
		if f, ok := fn.(func() []string); ok {
			result := f()
			table := L.NewTable()
			for i, s := range result {
				table.RawSetInt(i+1, lua.LString(s))
			}
			L.Push(table)
			return 1
		}

		// Example: function that takes no args and returns nothing
		if f, ok := fn.(func()); ok {
			f()
			return 0
		}

		L.RaiseError("unsupported function signature")
		return 0
	}
}

// Basic type conversions (will be expanded in conversions.go)

// goToLua converts a Go value to a Lua value
func goToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	switch v := value.(type) {
	case bool:
		return lua.LBool(v), nil
	case int:
		return lua.LNumber(float64(v)), nil
	case int64:
		return lua.LNumber(float64(v)), nil
	case float64:
		return lua.LNumber(v), nil
	case string:
		return lua.LString(v), nil
	case []byte:
		return lua.LString(string(v)), nil
	default:
		// For complex types, we'll implement proper conversion in conversions.go
		return lua.LNil, fmt.Errorf("unsupported type: %T", value)
	}
}

// luaToGo converts a Lua value to a Go value
func luaToGo(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LNilType:
		return nil
	case *lua.LTable:
		// Basic table conversion - will be enhanced in conversions.go
		return luaTableToMap(v)
	default:
		return nil
	}
}

// luaTableToMap converts a Lua table to a Go map (basic implementation)
func luaTableToMap(table *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})

	table.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			result[string(keyStr)] = luaToGo(value)
		}
	})

	return result
}
