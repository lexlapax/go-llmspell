// ABOUTME: Tests for State bridge adapter that exposes go-llms state management functionality to Lua scripts
// ABOUTME: Validates state and context management, transforms, validation, persistence, and merging operations

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestStateAdapter_Creation(t *testing.T) {
	t.Run("create_state_adapter", func(t *testing.T) {
		// Create state bridge mock
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "State Bridge",
				Version:     "1.0.0",
				Description: "Provides state management functionality",
			}).
			WithMethod("createState", engine.MethodInfo{
				Name:        "createState",
				Description: "Create a new state object",
				ReturnType:  "State",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock state creation
				stateFields := map[string]engine.ScriptValue{
					"id":       engine.NewStringValue("state-123"),
					"created":  engine.NewStringValue("2024-01-01T00:00:00Z"),
					"modified": engine.NewStringValue("2024-01-01T00:00:00Z"),
					"version":  engine.NewNumberValue(1),
					"data":     engine.NewObjectValue(map[string]engine.ScriptValue{}),
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{}),
				}
				return engine.NewObjectValue(stateFields), nil
			}).
			WithMethod("get", engine.MethodInfo{
				Name:        "get",
				Description: "Get a value from state",
				ReturnType:  "any",
			}, nil).
			WithMethod("set", engine.MethodInfo{
				Name:        "set",
				Description: "Set a value in state",
				ReturnType:  "void",
			}, nil).
			WithMethod("has", engine.MethodInfo{
				Name:        "has",
				Description: "Check if state has a key",
				ReturnType:  "boolean",
			}, nil)

		// Create adapter
		adapter := NewStateAdapter(stateBridge)
		require.NotNil(t, adapter)

		// Should have state-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "createState")
		assert.Contains(t, methods, "get")
		assert.Contains(t, methods, "set")
		assert.Contains(t, methods, "has")
		assert.Contains(t, methods, "saveState")
		assert.Contains(t, methods, "loadState")
		assert.Contains(t, methods, "applyTransform")
		assert.Contains(t, methods, "mergeStates")
	})

	t.Run("state_module_structure", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "State Bridge",
			}).
			WithMethod("createState", engine.MethodInfo{
				Name: "createState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-state"), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check standard methods exist
		assert.NotEqual(t, lua.LNil, module.RawGetString("createState"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("State")) // Constructor alias

		// Check flattened namespace methods exist
		// Transform methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("transformsApply"), "transformsApply should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("transformsRegister"), "transformsRegister should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("transformsChain"), "transformsChain should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("transformsValidate"), "transformsValidate should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("transformsGetAvailable"), "transformsGetAvailable should exist")

		// Context methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextGet"), "contextGet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextSet"), "contextSet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextMerge"), "contextMerge should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextClear"), "contextClear should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextCreateShared"), "contextCreateShared should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("contextWithInheritance"), "contextWithInheritance should exist")

		// Persistence methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("persistenceSave"), "persistenceSave should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("persistenceLoad"), "persistenceLoad should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("persistenceExists"), "persistenceExists should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("persistenceDelete"), "persistenceDelete should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("persistenceListVersions"), "persistenceListVersions should exist")

		// Check namespaces don't exist (flattened)
		assert.Equal(t, lua.LNil, module.RawGetString("transforms"), "transforms namespace should not exist")
		assert.Equal(t, lua.LNil, module.RawGetString("context"), "context namespace should not exist")
		assert.Equal(t, lua.LNil, module.RawGetString("persistence"), "persistence namespace should not exist")
	})
}

func TestStateAdapter_StateCreation(t *testing.T) {
	t.Run("create_simple_state", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("createState", engine.MethodInfo{
				Name: "createState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock state object
				stateFields := map[string]engine.ScriptValue{
					"id":       engine.NewStringValue("state-456"),
					"created":  engine.NewStringValue("2024-01-01T00:00:00Z"),
					"modified": engine.NewStringValue("2024-01-01T00:00:00Z"),
					"version":  engine.NewNumberValue(1),
					"data":     engine.NewObjectValue(map[string]engine.ScriptValue{}),
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{}),
				}
				return engine.NewObjectValue(stateFields), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Create state from Lua
		err = L.DoString(`
			local state = require("state")
			local newState = state.createState()
			assert(newState ~= nil)
			assert(newState.id == "state-456")
			assert(newState.version == 1)
		`)
		assert.NoError(t, err)
	})

	t.Run("create_state_with_initial_data", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("createState", engine.MethodInfo{
				Name: "createState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Check if initial data was provided
				var initialData map[string]engine.ScriptValue
				if len(args) > 0 && args[0].Type() == engine.TypeObject {
					initialData = args[0].(engine.ObjectValue).Fields()
				} else {
					initialData = map[string]engine.ScriptValue{}
				}

				stateFields := map[string]engine.ScriptValue{
					"id":       engine.NewStringValue("state-789"),
					"data":     engine.NewObjectValue(initialData),
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{}),
				}
				return engine.NewObjectValue(stateFields), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Create state with initial data
		err = L.DoString(`
			local state = require("state")
			local newState = state.createState({
				name = "test",
				value = 42
			})
			assert(newState.data.name == "test")
			assert(newState.data.value == 42)
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_StateOperations(t *testing.T) {
	t.Run("get_set_operations", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("get", engine.MethodInfo{
				Name: "get",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[1].Type() == engine.TypeString {
					key := args[1].(engine.StringValue).Value()
					if key == "test_key" {
						return engine.NewStringValue("test_value"), nil
					}
				}
				return engine.NewNilValue(), nil
			}).
			WithMethod("set", engine.MethodInfo{
				Name: "set",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful set operation
				return engine.NewNilValue(), nil
			}).
			WithMethod("has", engine.MethodInfo{
				Name: "has",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[1].Type() == engine.TypeString {
					key := args[1].(engine.StringValue).Value()
					return engine.NewBoolValue(key == "test_key"), nil
				}
				return engine.NewBoolValue(false), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test get/set operations
		err = L.DoString(`
			local state = require("state")
			local mockState = { id = "state-123" }
			
			-- Test set operation
			local setResult, setErr = state.set(mockState, "test_key", "test_value")
			assert(setErr == nil, "set should not error: " .. tostring(setErr))
			
			-- Test has operation
			local hasResult, hasErr = state.has(mockState, "test_key")
			assert(hasErr == nil, "has should not error: " .. tostring(hasErr))
			assert(hasResult == true, "should have test_key")
			
			-- Test get operation
			local getValue, getErr = state.get(mockState, "test_key")
			assert(getErr == nil, "get should not error: " .. tostring(getErr))
			assert(getValue == "test_value", "should get correct value")
		`)
		assert.NoError(t, err)
	})

	t.Run("keys_values_operations", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("keys", engine.MethodInfo{
				Name: "keys",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				keys := []engine.ScriptValue{
					engine.NewStringValue("key1"),
					engine.NewStringValue("key2"),
				}
				return engine.NewArrayValue(keys), nil
			}).
			WithMethod("values", engine.MethodInfo{
				Name: "values",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				values := []engine.ScriptValue{
					engine.NewStringValue("value1"),
					engine.NewStringValue("value2"),
				}
				return engine.NewArrayValue(values), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test keys/values operations
		err = L.DoString(`
			local state = require("state")
			local mockState = { id = "state-123" }
			
			-- Test keys operation - get individual keys as multiple returns
			local key1, key2 = state.keys(mockState)
			assert(key1 == "key1", "first key should be key1")
			assert(key2 == "key2", "second key should be key2")
			
			-- Test values operation - get individual values as multiple returns
			local val1, val2 = state.values(mockState)
			assert(val1 == "value1", "first value should be value1")
			assert(val2 == "value2", "second value should be value2")
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_StateTransforms(t *testing.T) {
	t.Run("apply_transform", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("applyTransform", engine.MethodInfo{
				Name: "applyTransform",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeString {
					transformName := args[0].(engine.StringValue).Value()
					// Return transformed state
					stateFields := map[string]engine.ScriptValue{
						"id":          engine.NewStringValue("state-transformed"),
						"transformed": engine.NewBoolValue(true),
						"transform":   engine.NewStringValue(transformName),
					}
					return engine.NewObjectValue(stateFields), nil
				}
				return engine.NewNilValue(), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test transform application
		err = L.DoString(`
			local state = require("state")
			local mockState = { id = "state-123" }
			
			-- Apply filter transform
			local transformed, err = state.transformsApply("filter", mockState)
			assert(err == nil, "transform should not error: " .. tostring(err))
			assert(transformed.transformed == true, "state should be transformed")
			assert(transformed.transform == "filter", "should record transform name")
		`)
		assert.NoError(t, err)
	})

	t.Run("register_custom_transform", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("registerTransform", engine.MethodInfo{
				Name: "registerTransform",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful registration
				return engine.NewNilValue(), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test custom transform registration
		err = L.DoString(`
			local state = require("state")
			
			-- Define a custom transform function
			local function customTransform(stateObj)
				-- Transform logic would go here
				return stateObj
			end
			
			-- Register the transform
			local result, err = state.transformsRegister("custom", customTransform)
			assert(err == nil, "registration should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_StatePersistence(t *testing.T) {
	t.Run("save_load_state", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("saveState", engine.MethodInfo{
				Name: "saveState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful save
				return engine.NewNilValue(), nil
			}).
			WithMethod("loadState", engine.MethodInfo{
				Name: "loadState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 1 && args[0].Type() == engine.TypeString {
					id := args[0].(engine.StringValue).Value()
					// Return loaded state
					stateFields := map[string]engine.ScriptValue{
						"id":     engine.NewStringValue(id),
						"loaded": engine.NewBoolValue(true),
					}
					return engine.NewObjectValue(stateFields), nil
				}
				return engine.NewNilValue(), nil
			}).
			WithMethod("deleteState", engine.MethodInfo{
				Name: "deleteState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful delete
				return engine.NewNilValue(), nil
			}).
			WithMethod("listStates", engine.MethodInfo{
				Name: "listStates",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				states := []engine.ScriptValue{
					engine.NewStringValue("state-1"),
					engine.NewStringValue("state-2"),
				}
				return engine.NewArrayValue(states), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test persistence operations
		err = L.DoString(`
			local state = require("state")
			local mockState = { id = "state-123", data = { key = "value" } }
			
			-- Test save
			local saveResult, saveErr = state.persistenceSave(mockState)
			assert(saveErr == nil, "save should not error: " .. tostring(saveErr))
			
			-- Test load
			local loadedState, loadErr = state.persistenceLoad("state-123")
			assert(loadErr == nil, "load should not error: " .. tostring(loadErr))
			assert(loadedState.loaded == true, "state should be loaded")
			
			-- Test list
			local statesList, listErr = state.persistenceListVersions()
			assert(listErr == nil, "list should not error: " .. tostring(listErr))
			assert(#statesList == 2, "should have 2 states")
			
			-- Test delete
			local deleteResult, deleteErr = state.persistenceDelete("state-123")
			assert(deleteErr == nil, "delete should not error: " .. tostring(deleteErr))
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_StateContext(t *testing.T) {
	t.Run("create_shared_context", func(t *testing.T) {
		// Mock the context bridge - this would normally be provided separately
		contextBridge := testutils.NewMockBridge("state_context").
			WithInitialized(true).
			WithMethod("createSharedContext", engine.MethodInfo{
				Name: "createSharedContext",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock shared context
				contextFields := map[string]engine.ScriptValue{
					"_id":              engine.NewStringValue("context-123"),
					"_type":            engine.NewStringValue("SharedStateContext"),
					"inheritMessages":  engine.NewBoolValue(true),
					"inheritArtifacts": engine.NewBoolValue(true),
					"inheritMetadata":  engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(contextFields), nil
			})

		adapter := NewStateAdapter(nil) // No state bridge needed for this test
		adapter.contextBridge = contextBridge

		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test shared context creation
		err = L.DoString(`
			local state = require("state")
			
			-- Create shared context
			local sharedContext, err = state.contextCreateShared()
			assert(err == nil, "createShared should not error: " .. tostring(err))
			assert(sharedContext._type == "SharedStateContext", "should be shared context")
			assert(sharedContext.inheritMessages == true, "should inherit messages")
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_StateMerging(t *testing.T) {
	t.Run("merge_states", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("mergeStates", engine.MethodInfo{
				Name: "mergeStates",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 {
					// First arg is array of states, second is strategy
					var stateCount int
					if args[0].Type() == engine.TypeArray {
						if av, ok := args[0].(engine.ArrayValue); ok {
							stateCount = len(av.Elements())
						}
					}

					// Return merged state
					mergedFields := map[string]engine.ScriptValue{
						"id":     engine.NewStringValue("merged-state"),
						"merged": engine.NewBoolValue(true),
						"count":  engine.NewNumberValue(float64(stateCount)),
					}
					return engine.NewObjectValue(mergedFields), nil
				}
				return engine.NewNilValue(), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test state merging
		err = L.DoString(`
			local state = require("state")
			local state1 = { id = "state-1", data = { key1 = "value1" } }
			local state2 = { id = "state-2", data = { key2 = "value2" } }
			
			-- Merge states with merge_all strategy
			local merged, err = state.mergeStates({state1, state2}, "merge_all")
			assert(err == nil, "merge should not error: " .. tostring(err))
			assert(merged.merged == true, "state should be merged")
			assert(merged.count == 2, "should have merged 2 states")
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("get", engine.MethodInfo{
				Name: "get",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("state not found")
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test error handling
		err = L.DoString(`
			local state = require("state")
			local mockState = { id = "state-123" }
			
			local result, err = state.get(mockState, "nonexistent")
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "state not found"), "should contain error message")
		`)
		assert.NoError(t, err)
	})
}

func TestStateAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("enhanced_state_object", func(t *testing.T) {
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("createState", engine.MethodInfo{
				Name: "createState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				stateFields := map[string]engine.ScriptValue{
					"id":   engine.NewStringValue("state-enhanced"),
					"data": engine.NewObjectValue(map[string]engine.ScriptValue{}),
				}
				return engine.NewObjectValue(stateFields), nil
			}).
			WithMethod("get", engine.MethodInfo{
				Name: "get",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("enhanced_value"), nil
			})

		adapter := NewStateAdapter(stateBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "state")
		require.NoError(t, err)

		err = ms.LoadModule(L, "state")
		require.NoError(t, err)

		// Test enhanced state object methods
		err = L.DoString(`
			local state = require("state")
			local newState = state.createState()
			
			-- Enhanced state object should have convenience methods
			assert(type(newState) == "table", "state should be a table, got: " .. type(newState))
			assert(type(newState.get) == "function", "state should have get method")
			assert(type(newState.set) == "function", "state should have set method")
			assert(type(newState.has) == "function", "state should have has method")
			
			-- Test convenience method
			local value, err = newState:get("test_key")
			assert(err == nil, "get method should work: " .. tostring(err))
			assert(value == "enhanced_value", "should get enhanced value")
		`)
		assert.NoError(t, err)
	})
}
