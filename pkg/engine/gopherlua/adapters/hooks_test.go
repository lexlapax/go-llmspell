// ABOUTME: Tests for Hooks bridge adapter that exposes go-llms hook functionality to Lua scripts
// ABOUTME: Validates hook registration, priority ordering, lifecycle execution, and management operations

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestHooksAdapter_Creation(t *testing.T) {
	t.Run("create_hooks_adapter", func(t *testing.T) {
		// Create hooks bridge mock
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Hooks Bridge",
				Version:     "1.0.0",
				Description: "Bridge for go-llms agent hook system",
			}).
			WithMethod("registerHook", engine.MethodInfo{
				Name:        "registerHook",
				Description: "Register a new hook with lifecycle callbacks",
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock hook registration
				id := args[0].(engine.StringValue).Value()
				return engine.NewStringValue(id), nil
			}).
			WithMethod("listHooks", engine.MethodInfo{
				Name:        "listHooks",
				Description: "List all registered hooks",
				ReturnType:  "array",
			}, nil).
			WithMethod("enableHook", engine.MethodInfo{
				Name:        "enableHook",
				Description: "Enable a disabled hook",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("unregisterHook", engine.MethodInfo{
				Name:        "unregisterHook",
				Description: "Remove a registered hook",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("disableHook", engine.MethodInfo{
				Name:        "disableHook",
				Description: "Disable a hook without removing it",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("getHookInfo", engine.MethodInfo{
				Name:        "getHookInfo",
				Description: "Get information about a specific hook",
				ReturnType:  "object",
			}, nil).
			WithMethod("executeHooks", engine.MethodInfo{
				Name:        "executeHooks",
				Description: "Execute hooks of a specific type",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("clearHooks", engine.MethodInfo{
				Name:        "clearHooks",
				Description: "Remove all registered hooks",
				ReturnType:  "number",
			}, nil)

		// Create adapter
		adapter := NewHooksAdapter(hooksBridge)
		require.NotNil(t, adapter)

		// Should have hooks-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "registerHook")
		assert.Contains(t, methods, "unregisterHook")
		assert.Contains(t, methods, "listHooks")
		assert.Contains(t, methods, "enableHook")
		assert.Contains(t, methods, "disableHook")
		assert.Contains(t, methods, "getHookInfo")
		assert.Contains(t, methods, "executeHooks")
		assert.Contains(t, methods, "clearHooks")
	})

	t.Run("hooks_module_structure", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Hooks Bridge",
			}).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("hook-123"), nil
			}).
			WithMethod("listHooks", engine.MethodInfo{
				Name: "listHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			})

		adapter := NewHooksAdapter(hooksBridge)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Create module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test module structure
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Check basic module properties
			assert(hooks._adapter == "hooks", "should have correct adapter name")
			assert(hooks._version == "1.0.0", "should have correct version")
			
			-- Check hook types constants
			assert(hooks.TYPES.BEFORE_GENERATE == "beforeGenerate", "should have before generate type")
			assert(hooks.TYPES.AFTER_GENERATE == "afterGenerate", "should have after generate type")
			assert(hooks.TYPES.BEFORE_TOOL_CALL == "beforeToolCall", "should have before tool call type")
			assert(hooks.TYPES.AFTER_TOOL_CALL == "afterToolCall", "should have after tool call type")
			
			-- Check priority constants
			assert(hooks.PRIORITY.HIGHEST == 1000, "should have highest priority")
			assert(hooks.PRIORITY.HIGH == 100, "should have high priority")
			assert(hooks.PRIORITY.NORMAL == 0, "should have normal priority")
			assert(hooks.PRIORITY.LOW == -100, "should have low priority")
			assert(hooks.PRIORITY.LOWEST == -1000, "should have lowest priority")
		`)
		assert.NoError(t, err)
	})
}

func TestHooksAdapter_Registration(t *testing.T) {
	t.Run("register_simple_hook", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock hook registration
				id := args[0].(engine.StringValue).Value()
				definition := args[1].(engine.ObjectValue).Fields()

				// Validate the hook definition
				assert.NotNil(t, definition["priority"])
				assert.NotNil(t, definition["beforeGenerate"])

				return engine.NewStringValue(id), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test hook registration
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Register a simple hook
			local hookID = hooks.registerHook("my-hook", {
				priority = hooks.PRIORITY.HIGH,
				beforeGenerate = function(ctx, messages)
					print("Before generate called")
				end
			})
			
			assert(hookID == "my-hook", "should return hook ID")
		`)
		assert.NoError(t, err)
	})

	t.Run("register_full_lifecycle_hook", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				definition := args[1].(engine.ObjectValue).Fields()

				// Validate all lifecycle methods are present
				assert.NotNil(t, definition["beforeGenerate"])
				assert.NotNil(t, definition["afterGenerate"])
				assert.NotNil(t, definition["beforeToolCall"])
				assert.NotNil(t, definition["afterToolCall"])

				return engine.NewStringValue("lifecycle-hook"), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test full lifecycle hook registration
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Register a hook with all lifecycle methods
			local hookID = hooks.registerHook("lifecycle-hook", {
				priority = hooks.PRIORITY.NORMAL,
				beforeGenerate = function(ctx, messages)
					print("Before generate: " .. #messages .. " messages")
				end,
				afterGenerate = function(ctx, response, err)
					if err then
						print("Generate error: " .. err)
					else
						print("Generated response: " .. response.content)
					end
				end,
				beforeToolCall = function(ctx, tool, params)
					print("Calling tool: " .. tool)
				end,
				afterToolCall = function(ctx, tool, result, err)
					if err then
						print("Tool error: " .. err)
					else
						print("Tool result received")
					end
				end
			})
			
			assert(hookID == "lifecycle-hook", "should return hook ID")
		`)
		assert.NoError(t, err)
	})

	t.Run("unregister_hook", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("unregisterHook", engine.MethodInfo{
				Name: "unregisterHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				// Mock successful unregistration
				return engine.NewBoolValue(id == "existing-hook"), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test hook unregistration
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Unregister existing hook
			local removed = hooks.unregisterHook("existing-hook")
			assert(removed == true, "should successfully unregister existing hook")
			
			-- Try to unregister non-existent hook
			local notRemoved = hooks.unregisterHook("non-existent")
			assert(notRemoved == false, "should return false for non-existent hook")
		`)
		assert.NoError(t, err)
	})
}

func TestHooksAdapter_Management(t *testing.T) {
	t.Run("list_hooks", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("listHooks", engine.MethodInfo{
				Name: "listHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock hook list
				hooks := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":       engine.NewStringValue("hook-1"),
						"priority": engine.NewNumberValue(100),
						"enabled":  engine.NewBoolValue(true),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":       engine.NewStringValue("hook-2"),
						"priority": engine.NewNumberValue(50),
						"enabled":  engine.NewBoolValue(false),
					}),
				}
				return engine.NewArrayValue(hooks), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test listing hooks
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- List all hooks - capture multiple return values into a table
			local hookList = {hooks.listHooks()}
			assert(#hookList == 2, "should have 2 hooks")
			
			-- Check first hook
			assert(hookList[1].id == "hook-1", "first hook should have correct ID")
			assert(hookList[1].priority == 100, "first hook should have correct priority")
			assert(hookList[1].enabled == true, "first hook should be enabled")
			
			-- Check second hook
			assert(hookList[2].id == "hook-2", "second hook should have correct ID")
			assert(hookList[2].enabled == false, "second hook should be disabled")
		`)
		assert.NoError(t, err)
	})

	t.Run("enable_disable_hooks", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("enableHook", engine.MethodInfo{
				Name: "enableHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				return engine.NewBoolValue(id == "existing-hook"), nil
			}).
			WithMethod("disableHook", engine.MethodInfo{
				Name: "disableHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				return engine.NewBoolValue(id == "existing-hook"), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test enabling and disabling hooks
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Enable an existing hook
			local enabled = hooks.enableHook("existing-hook")
			assert(enabled == true, "should successfully enable existing hook")
			
			-- Try to enable non-existent hook
			local notEnabled = hooks.enableHook("non-existent")
			assert(notEnabled == false, "should return false for non-existent hook")
			
			-- Disable an existing hook
			local disabled = hooks.disableHook("existing-hook")
			assert(disabled == true, "should successfully disable existing hook")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_hook_info", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("getHookInfo", engine.MethodInfo{
				Name: "getHookInfo",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				if id == "existing-hook" {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":       engine.NewStringValue("existing-hook"),
						"priority": engine.NewNumberValue(75),
						"enabled":  engine.NewBoolValue(true),
					}), nil
				}
				return engine.NewNilValue(), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test getting hook info
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Get info for existing hook
			local info = hooks.getHookInfo("existing-hook")
			assert(info ~= nil, "should return info for existing hook")
			assert(info.id == "existing-hook", "should have correct ID")
			assert(info.priority == 75, "should have correct priority")
			assert(info.enabled == true, "should have correct enabled state")
			
			-- Try to get info for non-existent hook
			local noInfo = hooks.getHookInfo("non-existent")
			assert(noInfo == nil, "should return nil for non-existent hook")
		`)
		assert.NoError(t, err)
	})

	t.Run("clear_all_hooks", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("clearHooks", engine.MethodInfo{
				Name: "clearHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock clearing 3 hooks
				return engine.NewNumberValue(3), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test clearing all hooks
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Clear all hooks
			local cleared = hooks.clearHooks()
			assert(cleared == 3, "should clear 3 hooks")
		`)
		assert.NoError(t, err)
	})
}

func TestHooksAdapter_Execution(t *testing.T) {
	t.Run("execute_hooks_by_type", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("executeHooks", engine.MethodInfo{
				Name: "executeHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				hookType := args[0].(engine.StringValue).Value()
				context := args[1].(engine.ObjectValue).Fields()

				// Validate hook type and context
				assert.Contains(t, []string{"beforeGenerate", "afterGenerate", "beforeToolCall", "afterToolCall"}, hookType)
				assert.NotNil(t, context)

				return engine.NewBoolValue(true), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test executing hooks
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Execute beforeGenerate hooks
			local success1 = hooks.executeHooks(hooks.TYPES.BEFORE_GENERATE, {
				messages = {
					{role = "user", content = "Hello"},
					{role = "assistant", content = "Hi there"}
				}
			})
			assert(success1 == true, "should execute beforeGenerate hooks")
			
			-- Execute afterGenerate hooks
			local success2 = hooks.executeHooks(hooks.TYPES.AFTER_GENERATE, {
				response = {content = "Generated response"},
				error = nil
			})
			assert(success2 == true, "should execute afterGenerate hooks")
			
			-- Execute beforeToolCall hooks
			local success3 = hooks.executeHooks(hooks.TYPES.BEFORE_TOOL_CALL, {
				tool = "calculator",
				params = {operation = "add", a = 5, b = 3}
			})
			assert(success3 == true, "should execute beforeToolCall hooks")
			
			-- Execute afterToolCall hooks
			local success4 = hooks.executeHooks(hooks.TYPES.AFTER_TOOL_CALL, {
				tool = "calculator",
				result = {answer = 8},
				error = nil
			})
			assert(success4 == true, "should execute afterToolCall hooks")
		`)
		assert.NoError(t, err)
	})

	t.Run("priority_ordering", func(t *testing.T) {
		callOrder := []string{}

		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				return engine.NewStringValue(id), nil
			}).
			WithMethod("executeHooks", engine.MethodInfo{
				Name: "executeHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Simulate priority ordering
				// In real implementation, hooks would be sorted by priority
				callOrder = append(callOrder, "high", "normal", "low")
				return engine.NewBoolValue(true), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test priority ordering
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Register hooks with different priorities
			hooks.registerHook("low-priority", {
				priority = hooks.PRIORITY.LOW,
				beforeGenerate = function(ctx, messages)
					-- This should execute last
				end
			})
			
			hooks.registerHook("high-priority", {
				priority = hooks.PRIORITY.HIGH,
				beforeGenerate = function(ctx, messages)
					-- This should execute first
				end
			})
			
			hooks.registerHook("normal-priority", {
				priority = hooks.PRIORITY.NORMAL,
				beforeGenerate = function(ctx, messages)
					-- This should execute in the middle
				end
			})
			
			-- Execute hooks - they should run in priority order
			hooks.executeHooks(hooks.TYPES.BEFORE_GENERATE, {
				messages = {{role = "user", content = "test"}}
			})
		`)
		assert.NoError(t, err)

		// Verify the expected call order
		assert.Equal(t, []string{"high", "normal", "low"}, callOrder)
	})
}

func TestHooksAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(false).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, nil)

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test error handling
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Try to register hook when bridge not initialized
			-- Bridge methods return (result, error) in Lua convention
			local result, err = hooks.registerHook("test-hook", {priority = 0})
			
			assert(result == nil, "result should be nil when bridge not initialized")
			assert(err ~= nil, "should have error when bridge not initialized")
			assert(string.find(err, "not initialized") ~= nil, "should have initialization error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_hook_type", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("executeHooks", engine.MethodInfo{
				Name: "executeHooks",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				hookType := args[0].(engine.StringValue).Value()
				if hookType == "invalidType" {
					return nil, fmt.Errorf("unknown hook type: invalidType")
				}
				return engine.NewBoolValue(true), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test invalid hook type
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Try to execute hooks with invalid type
			-- Bridge methods return (result, error) in Lua convention
			local result, err = hooks.executeHooks("invalidType", {})
			
			assert(result == nil, "result should be nil for invalid hook type")
			assert(err ~= nil, "should have error for invalid hook type")
			assert(string.find(err, "unknown hook type") ~= nil, "should have correct error message: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestHooksAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("hook_builder_pattern", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				definition := args[1].(engine.ObjectValue).Fields()

				// Verify builder pattern created correct hook
				priority := definition["priority"].(engine.NumberValue).Value()
				assert.Equal(t, float64(100), priority) // HIGH priority

				return engine.NewStringValue(id), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test hook builder pattern
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Create hook using builder pattern
			local hookID = hooks.createHook("builder-hook")
				:withPriority(hooks.PRIORITY.HIGH)
				:beforeGenerate(function(ctx, messages)
					print("Before generate")
				end)
				:afterGenerate(function(ctx, response, err)
					print("After generate")
				end)
				:register()
			
			assert(hookID == "builder-hook", "should register hook with builder pattern")
		`)
		assert.NoError(t, err)
	})

	t.Run("batch_operations", func(t *testing.T) {
		hooksBridge := testutils.NewMockBridge("hooks").
			WithInitialized(true).
			WithMethod("registerHook", engine.MethodInfo{
				Name: "registerHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				id := args[0].(engine.StringValue).Value()
				return engine.NewStringValue(id), nil
			}).
			WithMethod("enableHook", engine.MethodInfo{
				Name: "enableHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			}).
			WithMethod("disableHook", engine.MethodInfo{
				Name: "disableHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			})

		adapter := NewHooksAdapter(hooksBridge)
		L := lua.NewState()
		defer L.Close()

		// Create and register module
		loader := adapter.CreateLuaModule()

		// Register in preload for require
		L.PreloadModule("hooks", loader)

		// Also load it directly
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(loader),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)
		module := L.Get(-1).(*lua.LTable)
		L.SetGlobal("hooks", module)

		// Test batch operations
		err = L.DoString(`
			local hooks = require("hooks")
			
			-- Register multiple hooks
			local hookIDs = {"hook1", "hook2", "hook3"}
			for _, id in ipairs(hookIDs) do
				hooks.registerHook(id, {priority = 0})
			end
			
			-- Batch enable hooks
			local enableResults = hooks.batchEnable(hookIDs)
			assert(#enableResults == 3, "should enable 3 hooks")
			for _, result in ipairs(enableResults) do
				assert(result == true, "all hooks should be enabled")
			end
			
			-- Batch disable hooks
			local disableResults = hooks.batchDisable({"hook1", "hook3"})
			assert(#disableResults == 2, "should disable 2 hooks")
		`)
		assert.NoError(t, err)
	})
}
