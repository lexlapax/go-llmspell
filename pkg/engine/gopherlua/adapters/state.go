// ABOUTME: State bridge adapter that exposes go-llms state management functionality to Lua scripts
// ABOUTME: Provides state creation, context management, transforms, validation, persistence, and merging operations

package adapters

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// StateAdapter specializes BridgeAdapter for state management functionality
type StateAdapter struct {
	*gopherlua.BridgeAdapter

	// Optional related bridges for enhanced functionality
	contextBridge engine.Bridge // StateContextBridge for shared contexts
}

// NewStateAdapter creates a new state adapter
func NewStateAdapter(bridge engine.Bridge) *StateAdapter {
	// Create state adapter
	adapter := &StateAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = gopherlua.NewBridgeAdapter(bridge)
	}

	// Add state-specific methods if not already present
	adapter.ensureStateMethods()

	return adapter
}

// NewStateAdapterWithContext creates a new state adapter with context bridge
func NewStateAdapterWithContext(bridge engine.Bridge, contextBridge engine.Bridge) *StateAdapter {
	adapter := NewStateAdapter(bridge)
	adapter.contextBridge = contextBridge
	return adapter
}

// ensureStateMethods ensures state-specific methods are available
func (sa *StateAdapter) ensureStateMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
	// In production, this could validate that expected state methods exist
}

// CreateLuaModule creates a Lua module with state-specific enhancements
func (sa *StateAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if sa.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := sa.BridgeAdapter.CreateLuaModule()
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

			// Copy base module methods to our module, but wrap certain methods
			baseModule.ForEach(func(k, v lua.LValue) {
				if keyStr, ok := k.(lua.LString); ok {
					methodName := string(keyStr)
					switch methodName {
					case "createState":
						// Use our wrapped version for createState
						if baseFn, ok := v.(*lua.LFunction); ok {
							module.RawSet(k, L.NewFunction(sa.wrapCreateState(baseFn.GFunction)))
						} else {
							module.RawSet(k, v)
						}
					case "get", "set", "has", "keys", "values":
						// Use wrapped versions for state operations
						if baseFn, ok := v.(*lua.LFunction); ok {
							module.RawSet(k, L.NewFunction(sa.wrapStateOperation(methodName, baseFn.GFunction)))
						} else {
							module.RawSet(k, v)
						}
					default:
						// Copy other methods as-is
						module.RawSet(k, v)
					}
				} else {
					module.RawSet(k, v)
				}
			})
		}

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("state"))
		L.SetField(module, "_version", lua.LString("1.0.0"))

		// Add state-specific enhancements
		sa.addStateEnhancements(L, module)

		// Add transform methods
		sa.addTransformMethods(L, module)

		// Add context methods
		sa.addContextMethods(L, module)

		// Add persistence methods
		sa.addPersistenceMethods(L, module)

		// Add convenience methods
		sa.addConvenienceMethods(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addStateEnhancements adds state-specific enhancements to the module
func (sa *StateAdapter) addStateEnhancements(L *lua.LState, module *lua.LTable) {
	// Add constructor alias
	if stateCreate := module.RawGetString("createState"); stateCreate != lua.LNil {
		L.SetField(module, "State", stateCreate)
	}

	// Add state constants
	sa.addStateConstants(L, module)
}

// addTransformMethods adds transform-related methods (flattened to module level)
func (sa *StateAdapter) addTransformMethods(L *lua.LState, module *lua.LTable) {
	// transformsApply method (flattened from transforms.apply)
	L.SetField(module, "transformsApply", L.NewFunction(func(L *lua.LState) int {
		transformName := L.CheckString(1)
		stateObj := L.CheckTable(2)
		options := L.OptTable(3, L.NewTable())

		// Convert Lua table to map
		stateMap := sa.tableToMap(L, stateObj)
		optionsMap := sa.tableToMap(L, options)

		// Call applyTransform through bridge
		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(transformName),
			engine.NewObjectValue(stateMap),
			engine.NewObjectValue(optionsMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "applyTransform", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// transformsRegister method (flattened from transforms.register)
	L.SetField(module, "transformsRegister", L.NewFunction(func(L *lua.LState) int {
		transformName := L.CheckString(1)
		_ = L.CheckFunction(2) // transformFunc - acknowledged but not used in this simplified implementation

		// Convert Lua function to Go interface
		transformGo := func(ctx context.Context, state interface{}) (interface{}, error) {
			// This is a simplified implementation
			// In practice, we'd need to properly convert between Go and Lua
			return state, nil
		}

		// Call registerTransform through bridge
		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(transformName),
			engine.NewCustomValue("function", transformGo),
		}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "registerTransform", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// transformsChain method (flattened from transforms.chain)
	L.SetField(module, "transformsChain", L.NewFunction(func(L *lua.LState) int {
		transformNames := L.CheckTable(1)
		stateObj := L.CheckTable(2)

		// Convert transform names to array
		var transforms []engine.ScriptValue
		transformNames.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTString {
				transforms = append(transforms, engine.NewStringValue(string(v.(lua.LString))))
			}
		})

		stateMap := sa.tableToMap(L, stateObj)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewArrayValue(transforms),
			engine.NewObjectValue(stateMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "chainTransforms", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// transformsValidate method (flattened from transforms.validate)
	L.SetField(module, "transformsValidate", L.NewFunction(func(L *lua.LState) int {
		transformName := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(transformName)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "validateTransform", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// transformsGetAvailable method (flattened from transforms.getAvailable)
	L.SetField(module, "transformsGetAvailable", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := sa.GetBridge().ExecuteMethod(ctx, "getAvailableTransforms", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add built-in transform constants directly to module
	L.SetField(module, "TRANSFORM_FILTER", lua.LString("filter"))
	L.SetField(module, "TRANSFORM_FLATTEN", lua.LString("flatten"))
	L.SetField(module, "TRANSFORM_SANITIZE", lua.LString("sanitize"))
}

// addContextMethods adds context-related methods (flattened to module level)
func (sa *StateAdapter) addContextMethods(L *lua.LState, module *lua.LTable) {
	// contextGet method (flattened from context.get)
	L.SetField(module, "contextGet", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(key)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "getContext", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// contextSet method (flattened from context.set)
	L.SetField(module, "contextSet", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.Get(2)

		// Convert value to ScriptValue
		var converter *gopherlua.LuaTypeConverter
		if sa.BridgeAdapter != nil {
			converter = sa.GetTypeConverter()
		} else {
			converter = gopherlua.NewLuaTypeConverter()
		}

		valueScriptValue, err := converter.ToLuaScriptValue(L, value)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(key),
			valueScriptValue,
		}

		_, err = sa.GetBridge().ExecuteMethod(ctx, "setContext", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// contextMerge method (flattened from context.merge)
	L.SetField(module, "contextMerge", L.NewFunction(func(L *lua.LState) int {
		contextData := L.CheckTable(1)

		contextMap := sa.tableToMap(L, contextData)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(contextMap)}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "mergeContext", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// contextClear method (flattened from context.clear)
	L.SetField(module, "contextClear", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := sa.GetBridge().ExecuteMethod(ctx, "clearContext", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// contextCreateShared method (was createShared in namespace, now flattened)
	L.SetField(module, "contextCreateShared", L.NewFunction(func(L *lua.LState) int {
		parentContext := L.OptTable(1, nil)

		var args []engine.ScriptValue
		if parentContext != nil {
			parentMap := sa.tableToMap(L, parentContext)
			args = append(args, engine.NewObjectValue(parentMap))
		}

		if sa.contextBridge != nil {
			ctx := context.Background()
			result, err := sa.contextBridge.ExecuteMethod(ctx, "createSharedContext", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Check if we have a type converter available
			var converter *gopherlua.LuaTypeConverter
			if sa.BridgeAdapter != nil {
				converter = sa.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("context bridge not available"))
		return 2
	}))

	// contextWithInheritance method (was withInheritance in namespace, now flattened)
	L.SetField(module, "contextWithInheritance", L.NewFunction(func(L *lua.LState) int {
		sharedContext := L.CheckTable(1)
		inheritMessages := L.CheckBool(2)
		inheritArtifacts := L.CheckBool(3)
		inheritMetadata := L.CheckBool(4)

		if sa.contextBridge != nil {
			ctx := context.Background()
			contextMap := sa.tableToMap(L, sharedContext)
			args := []engine.ScriptValue{
				engine.NewObjectValue(contextMap),
				engine.NewBoolValue(inheritMessages),
				engine.NewBoolValue(inheritArtifacts),
				engine.NewBoolValue(inheritMetadata),
			}

			result, err := sa.contextBridge.ExecuteMethod(ctx, "withInheritanceConfig", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Check if we have a type converter available
			var converter *gopherlua.LuaTypeConverter
			if sa.BridgeAdapter != nil {
				converter = sa.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("context bridge not available"))
		return 2
	}))
}

// addPersistenceMethods adds persistence-related methods (flattened to module level)
func (sa *StateAdapter) addPersistenceMethods(L *lua.LState, module *lua.LTable) {
	// persistenceSave method (flattened from persistence.save)
	L.SetField(module, "persistenceSave", L.NewFunction(func(L *lua.LState) int {
		stateObj := L.CheckTable(1)

		stateMap := sa.tableToMap(L, stateObj)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(stateMap)}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "saveState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// persistenceLoad method (flattened from persistence.load)
	L.SetField(module, "persistenceLoad", L.NewFunction(func(L *lua.LState) int {
		stateID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(stateID)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "loadState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// persistenceExists method (flattened from persistence.exists)
	L.SetField(module, "persistenceExists", L.NewFunction(func(L *lua.LState) int {
		stateID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(stateID)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "stateExists", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// persistenceListVersions method (flattened from persistence.listVersions)
	L.SetField(module, "persistenceListVersions", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := sa.GetBridge().ExecuteMethod(ctx, "listStates", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// persistenceDelete method (flattened from persistence.delete)
	L.SetField(module, "persistenceDelete", L.NewFunction(func(L *lua.LState) int {
		stateID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(stateID)}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "deleteState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))
}

// addConvenienceMethods adds convenience methods to the module
func (sa *StateAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// Enhance mergeStates to accept arrays properly
	L.SetField(module, "mergeStates", L.NewFunction(func(L *lua.LState) int {
		statesTable := L.CheckTable(1)
		strategy := L.CheckString(2)

		// Convert Lua array to Go array
		var states []engine.ScriptValue
		statesTable.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTTable {
				stateMap := sa.tableToMap(L, v.(*lua.LTable))
				states = append(states, engine.NewObjectValue(stateMap))
			}
		})

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewArrayValue(states),
			engine.NewStringValue(strategy),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "mergeStates", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))
}

// addStateConstants adds state-related constants to the module
func (sa *StateAdapter) addStateConstants(L *lua.LState, module *lua.LTable) {
	// Add merge strategies
	strategies := L.NewTable()
	L.SetField(strategies, "LAST", lua.LString("last"))
	L.SetField(strategies, "MERGE_ALL", lua.LString("merge_all"))
	L.SetField(strategies, "UNION", lua.LString("union"))
	L.SetField(module, "MERGE_STRATEGIES", strategies)

	// Add transform types
	transformTypes := L.NewTable()
	L.SetField(transformTypes, "FILTER", lua.LString("filter"))
	L.SetField(transformTypes, "FLATTEN", lua.LString("flatten"))
	L.SetField(transformTypes, "SANITIZE", lua.LString("sanitize"))
	L.SetField(module, "TRANSFORM_TYPES", transformTypes)
}

// WrapMethod wraps a bridge method with state-specific handling
func (sa *StateAdapter) WrapMethod(methodName string) lua.LGFunction {
	// Get base wrapped method if available
	if sa.BridgeAdapter != nil {
		baseWrapped := sa.BridgeAdapter.WrapMethod(methodName)

		// Add state-specific handling for certain methods
		switch methodName {
		case "createState":
			return sa.wrapCreateState(baseWrapped)
		case "get", "set", "has", "keys", "values":
			return sa.wrapStateOperation(methodName, baseWrapped)
		default:
			return baseWrapped
		}
	}

	// Return a simple function that returns an error when no bridge is available
	return func(L *lua.LState) int {
		L.Push(lua.LNil)
		L.Push(lua.LString("method not available - no bridge adapter"))
		return 2
	}
}

// wrapCreateState adds state-specific handling for createState
func (sa *StateAdapter) wrapCreateState(baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Call base function
		returnCount := baseFn(L)

		// If successful, enhance the state object
		if returnCount > 0 && L.Get(-returnCount).Type() == lua.LTTable {
			state := L.Get(-returnCount).(*lua.LTable)

			// Save current stack size
			stackSize := L.GetTop()

			// Get the module from the package.loaded table
			L.GetField(L.GetField(L.Get(lua.RegistryIndex), "_LOADED"), "state")
			if L.Get(-1).Type() == lua.LTTable {
				stateModule := L.Get(-1).(*lua.LTable)
				sa.enhanceStateObjectWithModule(L, state, stateModule)
			}

			// Restore stack to original size
			L.SetTop(stackSize)
		}

		return returnCount
	}
}

// wrapStateOperation adds state operation handling
func (sa *StateAdapter) wrapStateOperation(_ string, baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least state parameter is provided
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("state parameter is required"))
			return 2
		}

		return baseFn(L)
	}
}

// enhanceStateObjectWithModule adds convenience methods to state objects
func (sa *StateAdapter) enhanceStateObjectWithModule(L *lua.LState, state *lua.LTable, stateModule *lua.LTable) {

	// Add get method to state
	L.SetField(state, "get", L.NewFunction(func(L *lua.LState) int {
		// When called with colon syntax, first arg is self (state), second is key
		self := L.CheckTable(1)
		key := L.CheckString(2)

		// Call the bridge directly to avoid recursion through wrapped methods
		if sa.BridgeAdapter != nil {
			// Convert arguments to ScriptValues
			converter := sa.GetTypeConverter()
			stateScriptValue, err := converter.ToLuaScriptValue(L, self)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("state conversion error: %v", err)))
				return 2
			}

			keyScriptValue := engine.NewStringValue(key)

			// Call bridge method directly
			ctx := context.Background()
			result, err := sa.BridgeAdapter.GetBridge().ExecuteMethod(ctx, "get", []engine.ScriptValue{stateScriptValue, keyScriptValue})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Convert result back to Lua
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("result conversion error: %v", err)))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil) // No error
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("bridge adapter not available"))
		return 2
	}))

	// Add set method to state
	L.SetField(state, "set", L.NewFunction(func(L *lua.LState) int {
		// When called with colon syntax, first arg is self (state), second is key, third is value
		self := L.CheckTable(1)
		key := L.CheckString(2)
		value := L.Get(3)

		// Call the bridge directly to avoid recursion
		if sa.BridgeAdapter != nil {
			converter := sa.GetTypeConverter()
			stateScriptValue, err := converter.ToLuaScriptValue(L, self)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("state conversion error: %v", err)))
				return 2
			}

			keyScriptValue := engine.NewStringValue(key)
			valueScriptValue, err := converter.ToLuaScriptValue(L, value)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("value conversion error: %v", err)))
				return 2
			}

			// Call bridge method directly
			ctx := context.Background()
			result, err := sa.BridgeAdapter.GetBridge().ExecuteMethod(ctx, "set", []engine.ScriptValue{stateScriptValue, keyScriptValue, valueScriptValue})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Convert result back to Lua (should be nil for set)
			if result != nil {
				luaResult, err := converter.FromLuaScriptValue(L, result)
				if err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(fmt.Sprintf("result conversion error: %v", err)))
					return 2
				}
				L.Push(luaResult)
			} else {
				L.Push(lua.LNil)
			}
			L.Push(lua.LNil) // No error
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("bridge adapter not available"))
		return 2
	}))

	// Add has method to state
	L.SetField(state, "has", L.NewFunction(func(L *lua.LState) int {
		// When called with colon syntax, first arg is self (state), second is key
		self := L.CheckTable(1)
		key := L.CheckString(2)

		// Call the bridge directly to avoid recursion
		if sa.BridgeAdapter != nil {
			converter := sa.GetTypeConverter()
			stateScriptValue, err := converter.ToLuaScriptValue(L, self)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("state conversion error: %v", err)))
				return 2
			}

			keyScriptValue := engine.NewStringValue(key)

			// Call bridge method directly
			ctx := context.Background()
			result, err := sa.BridgeAdapter.GetBridge().ExecuteMethod(ctx, "has", []engine.ScriptValue{stateScriptValue, keyScriptValue})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Convert result back to Lua
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("result conversion error: %v", err)))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil) // No error
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("bridge adapter not available"))
		return 2
	}))
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (sa *StateAdapter) tableToMap(L *lua.LState, table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *gopherlua.LuaTypeConverter
			if sa.BridgeAdapter != nil {
				converter = sa.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			sv, err := converter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return result
}

// RegisterAsModule registers the adapter as a module in the module system
func (sa *StateAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if sa.GetBridge() != nil {
		bridgeMetadata = sa.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "State Adapter",
			Description: "State management functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // State module has no dependencies by default
		LoadFunc:     sa.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (sa *StateAdapter) GetBridge() engine.Bridge {
	if sa.BridgeAdapter != nil {
		return sa.BridgeAdapter.GetBridge()
	}
	return nil
}

// GetMethods returns the available methods
func (sa *StateAdapter) GetMethods() []string {
	// Get base methods if bridge adapter exists
	var methods []string
	if sa.BridgeAdapter != nil {
		methods = sa.BridgeAdapter.GetMethods()
	}

	// Add state-specific methods if not already present
	stateMethods := []string{
		// Base state methods
		"createState", "saveState", "loadState", "deleteState", "listStates",
		"get", "set", "delete", "has", "keys", "values",
		"setMetadata", "getMetadata", "getAllMetadata",
		"addArtifact", "getArtifact", "artifacts",
		"addMessage", "messages",
		"applyTransform", "registerTransform",
		"mergeStates", "validateState",
		// Flattened transform methods
		"transformsApply", "transformsRegister", "transformsChain",
		"transformsValidate", "transformsGetAvailable",
		// Flattened context methods
		"contextGet", "contextSet", "contextMerge", "contextClear",
		"contextCreateShared", "contextWithInheritance",
		// Flattened persistence methods
		"persistenceSave", "persistenceLoad", "persistenceExists",
		"persistenceDelete", "persistenceListVersions",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range stateMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
