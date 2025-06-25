// ABOUTME: Hooks bridge adapter that exposes go-llms hook functionality to Lua scripts
// ABOUTME: Provides hook registration, priority ordering, lifecycle execution, and management operations

package impl

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/adapters"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/converters"
)

// HooksAdapter specializes BridgeAdapter for hooks functionality.
// It provides hook registration, priority ordering, lifecycle execution,
// and management operations for extending LLM behavior at various points
// in the processing pipeline.
// bridge_id=bridges.agent_hooks
type HooksAdapter struct {
	*adapters.BridgeAdapter
}

// NewHooksAdapter creates a new hooks adapter with the provided bridge.
// The bridge parameter should be a hooks bridge from go-llms that provides
// hook management functionality. Returns an adapter that can be registered
// as a Lua module.
func NewHooksAdapter(bridge engine.Bridge) *HooksAdapter {
	// Create hooks adapter
	adapter := &HooksAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = adapters.NewBridgeAdapter(bridge)
	}

	// Add hooks-specific methods if not already present
	adapter.ensureHooksMethods()

	return adapter
}

// ensureHooksMethods ensures hooks-specific methods are available
func (ha *HooksAdapter) ensureHooksMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
}

// CreateLuaModule creates a Lua module with hooks-specific enhancements.
// Returns a Lua function that, when called, creates and returns a module table
// containing all hook operations, constants, and convenience methods. The module
// provides a complete hook management API for Lua scripts.
func (ha *HooksAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if ha.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := ha.BridgeAdapter.CreateLuaModule()
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(baseLoader),
				NRet:    1,
				Protect: true,
			})
			if err != nil {
				L.RaiseError("failed to create base module: %v", err)
				return 0
			}

			// Get the base module and copy its methods
			baseModule := L.Get(-1).(*lua.LTable)
			L.Pop(1)

			// Copy base module methods to our module
			baseModule.ForEach(func(k, v lua.LValue) {
				module.RawSet(k, v)
			})
		}

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("hooks"))
		L.SetField(module, "_version", lua.LString("1.0.0"))

		// Add hooks-specific enhancements
		ha.addHooksConstants(L, module)
		ha.addConvenienceMethods(L, module)

		// Push module
		L.Push(module)
		return 1
	}
}

// addHooksConstants adds hook-related constants
func (ha *HooksAdapter) addHooksConstants(L *lua.LState, module *lua.LTable) {
	// Hook types
	types := L.NewTable()
	L.SetField(types, "BEFORE_GENERATE", lua.LString("beforeGenerate"))
	L.SetField(types, "AFTER_GENERATE", lua.LString("afterGenerate"))
	L.SetField(types, "BEFORE_TOOL_CALL", lua.LString("beforeToolCall"))
	L.SetField(types, "AFTER_TOOL_CALL", lua.LString("afterToolCall"))
	L.SetField(module, "TYPES", types)

	// Priority levels
	priority := L.NewTable()
	L.SetField(priority, "HIGHEST", lua.LNumber(1000))
	L.SetField(priority, "HIGH", lua.LNumber(100))
	L.SetField(priority, "NORMAL", lua.LNumber(0))
	L.SetField(priority, "LOW", lua.LNumber(-100))
	L.SetField(priority, "LOWEST", lua.LNumber(-1000))
	L.SetField(module, "PRIORITY", priority)
}

// addConvenienceMethods adds convenience methods for hooks
func (ha *HooksAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// createHook - Builder pattern for creating hooks
	L.SetField(module, "createHook", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)

		// Create builder table
		builder := L.NewTable()
		builderMeta := L.NewTable()

		// Hook definition being built
		definition := L.NewTable()

		// Set default priority
		L.SetField(definition, "priority", lua.LNumber(0))

		// Builder methods
		L.SetField(builderMeta, "__index", L.NewFunction(func(L *lua.LState) int {
			method := L.CheckString(2)

			switch method {
			case "withPriority":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					priority := L.CheckNumber(2)
					L.SetField(definition, "priority", priority)
					L.Push(builder) // Return builder for chaining
					return 1
				}))
				return 1

			case "beforeGenerate":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					fn := L.CheckFunction(2)
					L.SetField(definition, "beforeGenerate", fn)
					L.Push(builder) // Return builder for chaining
					return 1
				}))
				return 1

			case "afterGenerate":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					fn := L.CheckFunction(2)
					L.SetField(definition, "afterGenerate", fn)
					L.Push(builder) // Return builder for chaining
					return 1
				}))
				return 1

			case "beforeToolCall":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					fn := L.CheckFunction(2)
					L.SetField(definition, "beforeToolCall", fn)
					L.Push(builder) // Return builder for chaining
					return 1
				}))
				return 1

			case "afterToolCall":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					fn := L.CheckFunction(2)
					L.SetField(definition, "afterToolCall", fn)
					L.Push(builder) // Return builder for chaining
					return 1
				}))
				return 1

			case "register":
				L.Push(L.NewFunction(func(L *lua.LState) int {
					// Call registerHook
					if ha.BridgeAdapter != nil {
						ctx := context.Background()
						args := []engine.ScriptValue{
							engine.NewStringValue(id),
							ha.tableToScriptValue(L, definition),
						}

						result, err := ha.GetBridge().ExecuteMethod(ctx, "registerHook", args)
						if err != nil {
							L.Push(lua.LNil)
							L.Push(lua.LString(err.Error()))
							return 2
						}

						// Convert result back to Lua
						luaResult, err := ha.GetTypeConverter().FromLuaScriptValue(L, result)
						if err != nil {
							L.Push(lua.LNil)
							L.Push(lua.LString(err.Error()))
							return 2
						}

						L.Push(luaResult)
						return 1
					}
					L.Push(lua.LNil)
					return 1
				}))
				return 1

			default:
				L.Push(lua.LNil)
				return 1
			}
		}))

		L.SetMetatable(builder, builderMeta)
		L.Push(builder)
		return 1
	}))

	// batchEnable - Enable multiple hooks at once
	L.SetField(module, "batchEnable", L.NewFunction(func(L *lua.LState) int {
		hookIDs := L.CheckTable(1)
		results := L.NewTable()

		hookIDs.ForEach(func(_, v lua.LValue) {
			if id, ok := v.(lua.LString); ok {
				ctx := context.Background()
				args := []engine.ScriptValue{engine.NewStringValue(string(id))}

				result, err := ha.GetBridge().ExecuteMethod(ctx, "enableHook", args)
				if err != nil {
					results.Append(lua.LFalse)
				} else {
					luaResult, _ := ha.GetTypeConverter().FromLuaScriptValue(L, result)
					results.Append(luaResult)
				}
			}
		})

		L.Push(results)
		return 1
	}))

	// batchDisable - Disable multiple hooks at once
	L.SetField(module, "batchDisable", L.NewFunction(func(L *lua.LState) int {
		hookIDs := L.CheckTable(1)
		results := L.NewTable()

		hookIDs.ForEach(func(_, v lua.LValue) {
			if id, ok := v.(lua.LString); ok {
				ctx := context.Background()
				args := []engine.ScriptValue{engine.NewStringValue(string(id))}

				result, err := ha.GetBridge().ExecuteMethod(ctx, "disableHook", args)
				if err != nil {
					results.Append(lua.LFalse)
				} else {
					luaResult, _ := ha.GetTypeConverter().FromLuaScriptValue(L, result)
					results.Append(luaResult)
				}
			}
		})

		L.Push(results)
		return 1
	}))
}

// tableToScriptValue converts a Lua table to a ScriptValue
func (ha *HooksAdapter) tableToScriptValue(L *lua.LState, table *lua.LTable) engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *converters.LuaTypeConverter
			if ha.BridgeAdapter != nil {
				converter = ha.GetTypeConverter()
			} else {
				converter = converters.NewLuaTypeConverter()
			}

			sv, err := converter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return engine.NewObjectValue(result)
}


// GetMethods returns the available methods as a map of method names to availability.
// The returned map includes all hook-related operations such as registration,
// enabling/disabling, and batch operations. Returns true for each available method.
func (ha *HooksAdapter) GetMethods() map[string]bool {
	methods := make(map[string]bool)

	// Base methods from bridge
	if ha.BridgeAdapter != nil && ha.GetBridge() != nil {
		bridgeMethods := ha.GetBridge().Methods()
		for _, method := range bridgeMethods {
			methods[method.Name] = true
		}
	}

	// Additional convenience methods
	methods["createHook"] = true
	methods["batchEnable"] = true
	methods["batchDisable"] = true

	return methods
}
