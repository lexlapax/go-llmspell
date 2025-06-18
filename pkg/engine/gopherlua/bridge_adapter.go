// ABOUTME: Bridge adapter system that wraps go-llms bridges for Lua script access
// ABOUTME: Provides automatic method discovery, type conversion, and Lua module generation

package gopherlua

import (
	"context"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// BridgeAdapter wraps a go-llms bridge for Lua script access
type BridgeAdapter struct {
	bridge      engine.Bridge
	converter   *LuaTypeConverter
	methodCache map[string]lua.LGFunction
	methodInfo  map[string]engine.MethodInfo
	validation  bool
	mu          sync.RWMutex
}

// NewBridgeAdapter creates a new bridge adapter
func NewBridgeAdapter(b engine.Bridge) *BridgeAdapter {
	adapter := &BridgeAdapter{
		bridge:      b,
		converter:   NewLuaTypeConverter(),
		methodCache: make(map[string]lua.LGFunction),
		methodInfo:  make(map[string]engine.MethodInfo),
		validation:  false,
	}

	// Cache method info
	for _, method := range b.Methods() {
		adapter.methodInfo[method.Name] = method
	}

	return adapter
}

// GetBridge returns the wrapped bridge
func (ba *BridgeAdapter) GetBridge() engine.Bridge {
	return ba.bridge
}

// GetID returns the bridge ID
func (ba *BridgeAdapter) GetID() string {
	return ba.bridge.GetID()
}

// GetMetadata returns the bridge metadata
func (ba *BridgeAdapter) GetMetadata() engine.BridgeMetadata {
	return ba.bridge.GetMetadata()
}

// GetMethods returns the available method names
func (ba *BridgeAdapter) GetMethods() []string {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	methods := make([]string, 0, len(ba.methodInfo))
	for name := range ba.methodInfo {
		methods = append(methods, name)
	}
	return methods
}

// GetMethodInfo returns information about a specific method
func (ba *BridgeAdapter) GetMethodInfo(name string) (engine.MethodInfo, error) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	info, exists := ba.methodInfo[name]
	if !exists {
		return engine.MethodInfo{}, fmt.Errorf("method not found: %s", name)
	}
	return info, nil
}

// SetTypeConverter sets a custom type converter
func (ba *BridgeAdapter) SetTypeConverter(converter *LuaTypeConverter) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.converter = converter
}

// GetTypeConverter returns the type converter
func (ba *BridgeAdapter) GetTypeConverter() *LuaTypeConverter {
	ba.mu.RLock()
	defer ba.mu.RUnlock()
	return ba.converter
}

// EnableValidation enables or disables method argument validation
func (ba *BridgeAdapter) EnableValidation(enable bool) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.validation = enable
}

// CreateLuaModule returns a Lua module loader function
func (ba *BridgeAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add metadata
		metadata := ba.GetMetadata()
		L.SetField(module, "_bridge", lua.LString(ba.GetID()))
		L.SetField(module, "_version", lua.LString(metadata.Version))
		L.SetField(module, "_description", lua.LString(metadata.Description))

		// Add methods
		for name := range ba.methodInfo {
			L.SetField(module, name, L.NewFunction(ba.WrapMethod(name)))
		}

		// Push module
		L.Push(module)
		return 1
	}
}

// WrapMethod wraps a bridge method for Lua
func (ba *BridgeAdapter) WrapMethod(methodName string) lua.LGFunction {
	ba.mu.RLock()
	if fn, exists := ba.methodCache[methodName]; exists {
		ba.mu.RUnlock()
		return fn
	}
	ba.mu.RUnlock()

	// Create wrapped function
	fn := func(L *lua.LState) (returnCount int) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("panic: %v", r)))
				returnCount = 2
			}
		}()

		// Get arguments and convert to ScriptValue
		nArgs := L.GetTop()
		args := make([]engine.ScriptValue, nArgs)
		for i := 0; i < nArgs; i++ {
			luaVal := L.Get(i + 1)
			scriptVal, err := ba.converter.ToLuaScriptValue(L, luaVal)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("argument conversion error: %v", err)))
				return 2
			}
			args[i] = scriptVal
		}

		// Validate if enabled
		if ba.validation {
			if err := ba.bridge.ValidateMethod(methodName, args); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("validation error: %v", err)))
				return 2
			}
		}

		// Execute bridge method
		ctx := context.Background() // TODO: Pass context from caller
		result, err := ba.bridge.ExecuteMethod(ctx, methodName, args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert ScriptValue result back to Lua
		if result != nil {
			// Check if result is an array (for multiple returns)
			if result.Type() == engine.TypeArray {
				if arrayResult, ok := result.(engine.ArrayValue); ok {
					elements := arrayResult.Elements()
					for _, elem := range elements {
						lval, err := ba.converter.FromLuaScriptValue(L, elem)
						if err != nil {
							L.Push(lua.LNil)
							L.Push(lua.LString(fmt.Sprintf("result conversion error: %v", err)))
							return 2
						}
						L.Push(lval)
					}
					return len(elements)
				}
			}

			// Single return
			lval, err := ba.converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("result conversion error: %v", err)))
				return 2
			}
			L.Push(lval)
			return 1
		} else {
			// Nil result
			L.Push(lua.LNil)
			return 1
		}
	}

	// Cache the function
	ba.mu.Lock()
	ba.methodCache[methodName] = fn
	ba.mu.Unlock()

	return fn
}

// RegisterAsModule registers the adapter as a module in the module system
func (ba *BridgeAdapter) RegisterAsModule(ms *ModuleSystem, name string) error {
	// Get bridge dependencies from metadata
	metadata := ba.GetMetadata()
	deps := metadata.Dependencies

	// Create module definition
	module := ModuleDefinition{
		Name:         name,
		Description:  metadata.Description,
		Dependencies: deps,
		LoadFunc:     ba.CreateLuaModule(),
	}

	// Register with module system
	return ms.Register(module)
}
