// ABOUTME: Tests for the hooks Lua module with mock bridge
// ABOUTME: Validates LLM pipeline hooks functionality

package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// setupHooksBridge sets up mock hooks bridge
func setupHooksBridge(t *testing.T, L *lua.LState) {
	t.Helper()

	// Create bridges table
	bridges := L.NewTable()
	L.SetGlobal("bridges", bridges)

	// Create hooks bridge
	hooksBridge := L.NewTable()

	// Mock hook storage
	hooks := make(map[string]*lua.LTable)

	// Hook management methods
	hooksBridge.RawSetString("registerHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		definition := L.CheckTable(2)

		// Create hook
		hook := L.NewTable()
		hook.RawSetString("id", lua.LString(hookID))
		hook.RawSetString("enabled", lua.LTrue)
		
		// Copy definition fields
		definition.ForEach(func(k, v lua.LValue) {
			hook.RawSet(k, v)
		})

		hooks[hookID] = hook
		L.Push(hook)
		return 1
	}))

	hooksBridge.RawSetString("unregisterHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if _, ok := hooks[hookID]; ok {
			delete(hooks, hookID)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	hooksBridge.RawSetString("enableHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if hook, ok := hooks[hookID]; ok {
			hook.RawSetString("enabled", lua.LTrue)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	hooksBridge.RawSetString("disableHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if hook, ok := hooks[hookID]; ok {
			hook.RawSetString("enabled", lua.LFalse)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	hooksBridge.RawSetString("getHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if hook, ok := hooks[hookID]; ok {
			L.Push(hook)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	hooksBridge.RawSetString("listHooks", L.NewFunction(func(L *lua.LState) int {
		filter := L.Get(1) // optional
		
		list := L.NewTable()
		i := 1
		for _, hook := range hooks {
			// Apply filter if provided
			if filter != lua.LNil && filter.Type() == lua.LTTable {
				filterTable := filter.(*lua.LTable)
				hookType := filterTable.RawGetString("type")
				if hookType != lua.LNil {
					// Check if hook has matching type
					hasType := false
					if hookType.String() == "beforeGenerate" && hook.RawGetString("beforeGenerate") != lua.LNil {
						hasType = true
					} else if hookType.String() == "afterGenerate" && hook.RawGetString("afterGenerate") != lua.LNil {
						hasType = true
					} else if hookType.String() == "beforeToolCall" && hook.RawGetString("beforeToolCall") != lua.LNil {
						hasType = true
					} else if hookType.String() == "afterToolCall" && hook.RawGetString("afterToolCall") != lua.LNil {
						hasType = true
					}
					if !hasType {
						continue
					}
				}
			}
			
			list.RawSetInt(i, hook)
			i++
		}
		L.Push(list)
		return 1
	}))

	hooksBridge.RawSetString("clearHooks", L.NewFunction(func(L *lua.LState) int {
		hookType := L.Get(1) // optional
		
		if hookType == lua.LNil {
			// Clear all hooks
			hooks = make(map[string]*lua.LTable)
		} else {
			// Clear hooks of specific type
			typeStr := hookType.String()
			for id, hook := range hooks {
				if typeStr == "beforeGenerate" && hook.RawGetString("beforeGenerate") != lua.LNil {
					delete(hooks, id)
				} else if typeStr == "afterGenerate" && hook.RawGetString("afterGenerate") != lua.LNil {
					delete(hooks, id)
				} else if typeStr == "beforeToolCall" && hook.RawGetString("beforeToolCall") != lua.LNil {
					delete(hooks, id)
				} else if typeStr == "afterToolCall" && hook.RawGetString("afterToolCall") != lua.LNil {
					delete(hooks, id)
				}
			}
		}
		
		L.Push(lua.LTrue)
		return 1
	}))

	// Hook execution methods
	hooksBridge.RawSetString("executeHooks", L.NewFunction(func(L *lua.LState) int {
		hookType := L.CheckString(1)
		_ = L.CheckTable(2) // context
		
		results := L.NewTable()
		for _, hook := range hooks {
			if hook.RawGetString("enabled") == lua.LTrue {
				var fn lua.LValue
				if hookType == "beforeGenerate" {
					fn = hook.RawGetString("beforeGenerate")
				} else if hookType == "afterGenerate" {
					fn = hook.RawGetString("afterGenerate")
				} else if hookType == "beforeToolCall" {
					fn = hook.RawGetString("beforeToolCall")
				} else if hookType == "afterToolCall" {
					fn = hook.RawGetString("afterToolCall")
				}
				
				if fn != lua.LNil && fn.Type() == lua.LTFunction {
					// Mock execution - just add hook ID to results
					results.Append(hook.RawGetString("id"))
				}
			}
		}
		L.Push(results)
		return 1
	}))

	hooksBridge.RawSetString("executeHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		_ = L.CheckTable(2) // context
		
		if hook, ok := hooks[hookID]; ok && hook.RawGetString("enabled") == lua.LTrue {
			result := L.NewTable()
			result.RawSetString("hook_id", lua.LString(hookID))
			result.RawSetString("executed", lua.LTrue)
			L.Push(result)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	// Priority management
	hooksBridge.RawSetString("setHookPriority", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		priority := L.CheckNumber(2)
		
		if hook, ok := hooks[hookID]; ok {
			hook.RawSetString("priority", priority)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	hooksBridge.RawSetString("getHookPriority", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if hook, ok := hooks[hookID]; ok {
			priority := hook.RawGetString("priority")
			if priority != lua.LNil {
				L.Push(priority)
			} else {
				L.Push(lua.LNumber(0)) // Default priority
			}
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	// Hook state queries
	hooksBridge.RawSetString("isHookEnabled", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		if hook, ok := hooks[hookID]; ok {
			L.Push(hook.RawGetString("enabled"))
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	hooksBridge.RawSetString("getEnabledHooks", L.NewFunction(func(L *lua.LState) int {
		hookType := L.Get(1) // optional
		
		list := L.NewTable()
		i := 1
		for _, hook := range hooks {
			if hook.RawGetString("enabled") == lua.LTrue {
				// Apply type filter if provided
				if hookType != lua.LNil {
					typeStr := hookType.String()
					hasType := false
					if typeStr == "beforeGenerate" && hook.RawGetString("beforeGenerate") != lua.LNil {
						hasType = true
					} else if typeStr == "afterGenerate" && hook.RawGetString("afterGenerate") != lua.LNil {
						hasType = true
					} else if typeStr == "beforeToolCall" && hook.RawGetString("beforeToolCall") != lua.LNil {
						hasType = true
					} else if typeStr == "afterToolCall" && hook.RawGetString("afterToolCall") != lua.LNil {
						hasType = true
					}
					if !hasType {
						continue
					}
				}
				
				list.RawSetInt(i, hook)
				i++
			}
		}
		L.Push(list)
		return 1
	}))

	hooksBridge.RawSetString("getDisabledHooks", L.NewFunction(func(L *lua.LState) int {
		hookType := L.Get(1) // optional
		
		list := L.NewTable()
		i := 1
		for _, hook := range hooks {
			if hook.RawGetString("enabled") != lua.LTrue {
				// Apply type filter if provided
				if hookType != lua.LNil {
					typeStr := hookType.String()
					hasType := false
					if typeStr == "beforeGenerate" && hook.RawGetString("beforeGenerate") != lua.LNil {
						hasType = true
					} else if typeStr == "afterGenerate" && hook.RawGetString("afterGenerate") != lua.LNil {
						hasType = true
					} else if typeStr == "beforeToolCall" && hook.RawGetString("beforeToolCall") != lua.LNil {
						hasType = true
					} else if typeStr == "afterToolCall" && hook.RawGetString("afterToolCall") != lua.LNil {
						hasType = true
					}
					if !hasType {
						continue
					}
				}
				
				list.RawSetInt(i, hook)
				i++
			}
		}
		L.Push(list)
		return 1
	}))

	// Batch operations (from adapter)
	hooksBridge.RawSetString("batchEnable", L.NewFunction(func(L *lua.LState) int {
		hookIDs := L.CheckTable(1)
		results := L.NewTable()
		
		hookIDs.ForEach(func(_, v lua.LValue) {
			if id, ok := v.(lua.LString); ok {
				if hook, exists := hooks[string(id)]; exists {
					hook.RawSetString("enabled", lua.LTrue)
					results.Append(lua.LTrue)
				} else {
					results.Append(lua.LFalse)
				}
			}
		})
		
		L.Push(results)
		return 1
	}))

	hooksBridge.RawSetString("batchDisable", L.NewFunction(func(L *lua.LState) int {
		hookIDs := L.CheckTable(1)
		results := L.NewTable()
		
		hookIDs.ForEach(func(_, v lua.LValue) {
			if id, ok := v.(lua.LString); ok {
				if hook, exists := hooks[string(id)]; exists {
					hook.RawSetString("enabled", lua.LFalse)
					results.Append(lua.LTrue)
				} else {
					results.Append(lua.LFalse)
				}
			}
		})
		
		L.Push(results)
		return 1
	}))

	// createHook builder from adapter
	hooksBridge.RawSetString("createHook", L.NewFunction(func(L *lua.LState) int {
		hookID := L.CheckString(1)
		
		// Create builder table
		builder := L.NewTable()
		
		// Internal definition
		definition := L.NewTable()
		definition.RawSetString("priority", lua.LNumber(0))
		
		// Builder methods
		builder.RawSetString("withPriority", L.NewFunction(func(L *lua.LState) int {
			_ = L.CheckTable(1) // self
			priority := L.CheckNumber(2)
			definition.RawSetString("priority", priority)
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("beforeGenerate", L.NewFunction(func(L *lua.LState) int {
			_ = L.CheckTable(1) // self
			fn := L.CheckFunction(2)
			definition.RawSetString("beforeGenerate", fn)
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("afterGenerate", L.NewFunction(func(L *lua.LState) int {
			_ = L.CheckTable(1) // self
			fn := L.CheckFunction(2)
			definition.RawSetString("afterGenerate", fn)
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("beforeToolCall", L.NewFunction(func(L *lua.LState) int {
			_ = L.CheckTable(1) // self
			fn := L.CheckFunction(2)
			definition.RawSetString("beforeToolCall", fn)
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("afterToolCall", L.NewFunction(func(L *lua.LState) int {
			_ = L.CheckTable(1) // self
			fn := L.CheckFunction(2)
			definition.RawSetString("afterToolCall", fn)
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("register", L.NewFunction(func(L *lua.LState) int {
			// Register the hook
			hook := L.NewTable()
			hook.RawSetString("id", lua.LString(hookID))
			hook.RawSetString("enabled", lua.LTrue)
			
			definition.ForEach(func(k, v lua.LValue) {
				hook.RawSet(k, v)
			})
			
			hooks[hookID] = hook
			L.Push(hook)
			return 1
		}))
		
		L.Push(builder)
		return 1
	}))

	bridges.RawSetString("agent_hooks", hooksBridge)
}

func TestHooksModule(t *testing.T) {
	t.Run("module loads with bridge", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			assert(type(hooks) == "table", "hooks should be a table")
			
			-- Check constants exist
			assert(type(hooks.TYPES) == "table", "TYPES should be a table")
			assert(hooks.TYPES.BEFORE_GENERATE == "beforeGenerate", "TYPES.BEFORE_GENERATE should be 'beforeGenerate'")
			assert(hooks.TYPES.AFTER_GENERATE == "afterGenerate", "TYPES.AFTER_GENERATE should be 'afterGenerate'")
			assert(hooks.TYPES.BEFORE_TOOL_CALL == "beforeToolCall", "TYPES.BEFORE_TOOL_CALL should be 'beforeToolCall'")
			assert(hooks.TYPES.AFTER_TOOL_CALL == "afterToolCall", "TYPES.AFTER_TOOL_CALL should be 'afterToolCall'")
			
			assert(type(hooks.PRIORITY) == "table", "PRIORITY should be a table")
			assert(hooks.PRIORITY.HIGHEST == 1000, "PRIORITY.HIGHEST should be 1000")
			assert(hooks.PRIORITY.HIGH == 100, "PRIORITY.HIGH should be 100")
			assert(hooks.PRIORITY.NORMAL == 0, "PRIORITY.NORMAL should be 0")
			assert(hooks.PRIORITY.LOW == -100, "PRIORITY.LOW should be -100")
			assert(hooks.PRIORITY.LOWEST == -1000, "PRIORITY.LOWEST should be -1000")
			
			-- Check methods exist
			assert(type(hooks.register_hook) == "function", "register_hook should be a function")
			assert(type(hooks.enable_hook) == "function", "enable_hook should be a function")
			assert(type(hooks.create_hook) == "function", "create_hook should be a function")
			assert(type(hooks.batch_enable) == "function", "batch_enable should be a function")
		`)
		assert.NoError(t, err)
	})

	t.Run("hook registration and management", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register a hook
			local hook = hooks.register_hook("test_hook", {
				priority = hooks.PRIORITY.HIGH,
				beforeGenerate = function(context)
					print("Before generate hook")
				end
			})
			assert(hook.id == "test_hook", "hook id should match")
			assert(hook.enabled == true, "hook should be enabled by default")
			assert(hook.priority == 100, "hook priority should be HIGH")
			
			-- Get hook
			local retrieved = hooks.get_hook("test_hook")
			assert(retrieved.id == "test_hook", "retrieved hook id should match")
			
			-- Disable hook
			local disabled = hooks.disable_hook("test_hook")
			assert(disabled == true, "should disable hook")
			
			-- Check if disabled
			local enabled = hooks.is_hook_enabled("test_hook")
			assert(enabled == false, "hook should be disabled")
			
			-- Enable hook
			local enabled_result = hooks.enable_hook("test_hook")
			assert(enabled_result == true, "should enable hook")
			
			-- Unregister hook
			local unregistered = hooks.unregister_hook("test_hook")
			assert(unregistered == true, "should unregister hook")
			
			-- Try to get unregistered hook
			local gone = hooks.get_hook("test_hook")
			assert(gone == nil, "unregistered hook should be nil")
		`)
		assert.NoError(t, err)
	})

	t.Run("hook lifecycle methods", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register multiple hooks with different types
			hooks.register_hook("before_gen_1", {
				priority = hooks.PRIORITY.HIGH,
				beforeGenerate = function(ctx) return ctx end
			})
			
			hooks.register_hook("after_gen_1", {
				priority = hooks.PRIORITY.NORMAL,
				afterGenerate = function(ctx) return ctx end
			})
			
			hooks.register_hook("before_tool_1", {
				priority = hooks.PRIORITY.LOW,
				beforeToolCall = function(ctx) return ctx end
			})
			
			hooks.register_hook("after_tool_1", {
				priority = hooks.PRIORITY.LOWEST,
				afterToolCall = function(ctx) return ctx end
			})
			
			-- List all hooks
			local all_hooks = hooks.list_hooks()
			assert(#all_hooks == 4, "should have 4 hooks")
			
			-- List hooks by type
			local before_gen_hooks = hooks.list_hooks({type = "beforeGenerate"})
			assert(#before_gen_hooks == 1, "should have 1 beforeGenerate hook")
			
			-- Execute hooks
			local context = {data = "test"}
			local results = hooks.execute_hooks("beforeGenerate", context)
			assert(#results >= 1, "should execute at least 1 hook")
			
			-- Clear hooks of specific type
			hooks.clear_hooks("beforeGenerate")
			before_gen_hooks = hooks.list_hooks({type = "beforeGenerate"})
			assert(#before_gen_hooks == 0, "should have no beforeGenerate hooks after clear")
			
			-- Clear all hooks
			hooks.clear_hooks()
			all_hooks = hooks.list_hooks()
			assert(#all_hooks == 0, "should have no hooks after clear all")
		`)
		assert.NoError(t, err)
	})

	t.Run("priority management", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register hook with default priority
			hooks.register_hook("priority_test", {
				beforeGenerate = function(ctx) return ctx end
			})
			
			-- Get priority (should be default)
			local priority = hooks.get_hook_priority("priority_test")
			assert(priority == 0, "default priority should be NORMAL (0)")
			
			-- Set new priority
			local set_result = hooks.set_hook_priority("priority_test", hooks.PRIORITY.HIGHEST)
			assert(set_result == true, "should set priority")
			
			-- Verify new priority
			priority = hooks.get_hook_priority("priority_test")
			assert(priority == 1000, "priority should be HIGHEST (1000)")
			
			-- Test with non-existent hook
			priority = hooks.get_hook_priority("non_existent")
			assert(priority == nil, "non-existent hook should return nil priority")
		`)
		assert.NoError(t, err)
	})

	t.Run("batch operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register multiple hooks
			hooks.register_hook("batch_1", {beforeGenerate = function() end})
			hooks.register_hook("batch_2", {afterGenerate = function() end})
			hooks.register_hook("batch_3", {beforeToolCall = function() end})
			
			-- Batch disable
			local results = hooks.batch_disable({"batch_1", "batch_2", "batch_3"})
			assert(#results == 3, "should have 3 results")
			assert(results[1] == true and results[2] == true and results[3] == true, "all should succeed")
			
			-- Verify disabled
			assert(hooks.is_hook_enabled("batch_1") == false, "batch_1 should be disabled")
			assert(hooks.is_hook_enabled("batch_2") == false, "batch_2 should be disabled")
			assert(hooks.is_hook_enabled("batch_3") == false, "batch_3 should be disabled")
			
			-- Batch enable
			results = hooks.batch_enable({"batch_1", "batch_3"})
			assert(#results == 2, "should have 2 results")
			assert(results[1] == true and results[2] == true, "both should succeed")
			
			-- Verify enabled/disabled
			assert(hooks.is_hook_enabled("batch_1") == true, "batch_1 should be enabled")
			assert(hooks.is_hook_enabled("batch_2") == false, "batch_2 should still be disabled")
			assert(hooks.is_hook_enabled("batch_3") == true, "batch_3 should be enabled")
			
			-- Test with non-existent hooks
			results = hooks.batch_enable({"batch_1", "non_existent", "batch_2"})
			assert(#results == 3, "should have 3 results")
			assert(results[1] == true and results[2] == false and results[3] == true, "only existing should succeed")
		`)
		assert.NoError(t, err)
	})

	t.Run("hook builder pattern", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Use bridge builder
			local hook = hooks.create_hook("builder_test")
				:withPriority(hooks.PRIORITY.HIGH)
				:beforeGenerate(function(ctx)
					ctx.modified = true
					return ctx
				end)
				:afterGenerate(function(ctx)
					ctx.completed = true
					return ctx
				end)
				:register()
			
			assert(hook.id == "builder_test", "hook id should match")
			assert(hook.priority == 100, "priority should be HIGH")
			assert(type(hook.beforeGenerate) == "function", "should have beforeGenerate function")
			assert(type(hook.afterGenerate) == "function", "should have afterGenerate function")
			
			-- Use Lua builder helper
			local hook2 = hooks.new_builder("lua_builder_test")
				:with_priority(hooks.PRIORITY.LOWEST)
				:before_tool_call(function(tool, args)
					print("Before tool call:", tool)
					return tool, args
				end)
				:after_tool_call(function(tool, result)
					print("After tool call:", tool)
					return result
				end)
				:register()
			
			assert(hook2.id == "lua_builder_test", "hook id should match")
			assert(hook2.priority == -1000, "priority should be LOWEST")
			assert(type(hook2.beforeToolCall) == "function", "should have beforeToolCall function")
			assert(type(hook2.afterToolCall) == "function", "should have afterToolCall function")
		`)
		assert.NoError(t, err)
	})

	t.Run("simple hook creation", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Create simple hook
			local hook = hooks.create_simple_hook(
				"simple_test",
				hooks.TYPES.BEFORE_GENERATE,
				function(ctx)
					ctx.simple = true
					return ctx
				end,
				hooks.PRIORITY.HIGH
			)
			
			assert(hook.id == "simple_test", "hook id should match")
			assert(hook.priority == 100, "priority should be HIGH")
			assert(type(hook.beforeGenerate) == "function", "should have beforeGenerate function")
			
			-- Create with default priority
			local hook2 = hooks.create_simple_hook(
				"simple_test2",
				hooks.TYPES.AFTER_TOOL_CALL,
				function(tool, result)
					return result
				end
			)
			
			assert(hook2.id == "simple_test2", "hook id should match")
			assert(hook2.priority == 0, "priority should be NORMAL (default)")
			assert(type(hook2.afterToolCall) == "function", "should have afterToolCall function")
			
			-- Test invalid hook type
			local success, err = pcall(function()
				hooks.create_simple_hook("bad_hook", "invalid_type", function() end)
			end)
			assert(not success, "should fail with invalid hook type")
			assert(err:find("Invalid hook type"), "should have correct error message")
		`)
		assert.NoError(t, err)
	})

	t.Run("hook state queries", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register hooks in different states
			hooks.register_hook("enabled_1", {
				beforeGenerate = function() end,
				priority = hooks.PRIORITY.HIGH
			})
			
			hooks.register_hook("enabled_2", {
				afterGenerate = function() end,
				priority = hooks.PRIORITY.NORMAL
			})
			
			hooks.register_hook("disabled_1", {
				beforeToolCall = function() end,
				priority = hooks.PRIORITY.LOW
			})
			
			hooks.register_hook("disabled_2", {
				afterToolCall = function() end,
				priority = hooks.PRIORITY.LOWEST
			})
			
			-- Disable some hooks
			hooks.disable_hook("disabled_1")
			hooks.disable_hook("disabled_2")
			
			-- Get enabled hooks
			local enabled = hooks.get_enabled_hooks()
			assert(#enabled == 2, "should have 2 enabled hooks")
			
			-- Get enabled hooks by type
			local enabled_before_gen = hooks.get_enabled_hooks("beforeGenerate")
			assert(#enabled_before_gen == 1, "should have 1 enabled beforeGenerate hook")
			
			-- Get disabled hooks
			local disabled = hooks.get_disabled_hooks()
			assert(#disabled == 2, "should have 2 disabled hooks")
			
			-- Get disabled hooks by type
			local disabled_before_tool = hooks.get_disabled_hooks("beforeToolCall")
			assert(#disabled_before_tool == 1, "should have 1 disabled beforeToolCall hook")
		`)
		assert.NoError(t, err)
	})

	t.Run("hook execution", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Register hooks
			hooks.register_hook("exec_1", {
				beforeGenerate = function(ctx)
					ctx.exec_1 = true
					return ctx
				end,
				priority = hooks.PRIORITY.HIGH
			})
			
			hooks.register_hook("exec_2", {
				beforeGenerate = function(ctx)
					ctx.exec_2 = true
					return ctx
				end,
				priority = hooks.PRIORITY.LOW
			})
			
			hooks.register_hook("exec_3", {
				afterGenerate = function(ctx)
					ctx.exec_3 = true
					return ctx
				end
			})
			
			-- Execute specific hook
			local context = {original = true}
			local result = hooks.execute_hook("exec_1", context)
			assert(result.hook_id == "exec_1", "should execute specific hook")
			assert(result.executed == true, "hook should be executed")
			
			-- Execute hooks by type
			local results = hooks.execute_hooks("beforeGenerate", context)
			assert(#results >= 2, "should execute at least 2 beforeGenerate hooks")
			
			-- Disable a hook and re-execute
			hooks.disable_hook("exec_2")
			results = hooks.execute_hooks("beforeGenerate", context)
			assert(#results >= 1, "should execute at least 1 hook (exec_2 disabled)")
			
			-- Execute non-existent hook
			result = hooks.execute_hook("non_existent", context)
			assert(result == nil, "non-existent hook should return nil")
		`)
		assert.NoError(t, err)
	})

	t.Run("missing bridge graceful failure", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Don't setup bridge
		L.SetGlobal("bridges", L.NewTable())
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			local success, err = pcall(function()
				hooks.register_hook("test", {})
			end)
			assert(not success, "should fail without bridge")
			assert(err:find("Hooks bridge not available"), "should have correct error message")
		`)
		assert.NoError(t, err)
	})
}

func TestHooksIntegration(t *testing.T) {
	t.Run("complete hook scenario", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- 1. Create a comprehensive hook setup
			local request_counter = 0
			local tokens_used = 0
			
			-- Before generate hook for rate limiting
			hooks.new_builder("rate_limiter")
				:with_priority(hooks.PRIORITY.HIGHEST)
				:before_generate(function(context)
					request_counter = request_counter + 1
					if request_counter > 10 then
						error("Rate limit exceeded")
					end
					context.request_id = request_counter
					return context
				end)
				:register()
			
			-- Before generate hook for prompt modification
			hooks.create_simple_hook(
				"prompt_modifier",
				hooks.TYPES.BEFORE_GENERATE,
				function(context)
					if context.prompt then
						context.prompt = "[SYSTEM] " .. context.prompt
					end
					return context
				end,
				hooks.PRIORITY.HIGH
			)
			
			-- After generate hook for token counting
			local token_counter = hooks.create_hook("token_counter")
				:withPriority(hooks.PRIORITY.NORMAL)
				:afterGenerate(function(context)
					if context.tokens then
						tokens_used = tokens_used + context.tokens
					end
					context.total_tokens_used = tokens_used
					return context
				end)
				:register()
			
			-- Before tool call hook for validation
			hooks.register_hook("tool_validator", {
				priority = hooks.PRIORITY.HIGH,
				beforeToolCall = function(tool_name, args)
					if tool_name == "dangerous_tool" then
						error("Dangerous tool blocked")
					end
					return tool_name, args
				end
			})
			
			-- After tool call hook for logging
			hooks.register_hook("tool_logger", {
				priority = hooks.PRIORITY.LOW,
				afterToolCall = function(tool_name, result)
					print(string.format("Tool '%s' executed", tool_name))
					return result
				end
			})
			
			-- 2. Test the hooks
			local context = {prompt = "Hello AI", tokens = 50}
			
			-- Execute beforeGenerate hooks
			local before_results = hooks.execute_hooks("beforeGenerate", context)
			assert(#before_results >= 2, "should execute rate_limiter and prompt_modifier")
			
			-- Execute afterGenerate hooks
			local after_results = hooks.execute_hooks("afterGenerate", context)
			assert(#after_results >= 1, "should execute token_counter")
			
			-- 3. Manage hook states
			local all_hooks = hooks.list_hooks()
			assert(#all_hooks == 5, "should have 5 hooks total")
			
			-- Disable tool validation temporarily
			hooks.disable_hook("tool_validator")
			
			-- Get hook states
			local enabled = hooks.get_enabled_hooks()
			local disabled = hooks.get_disabled_hooks()
			assert(#enabled == 4, "should have 4 enabled hooks")
			assert(#disabled == 1, "should have 1 disabled hook")
			
			-- 4. Batch operations
			local tool_hooks = {"tool_validator", "tool_logger"}
			hooks.batch_enable(tool_hooks)
			
			-- 5. Priority-based execution order
			local priorities = {}
			priorities[1] = hooks.get_hook_priority("rate_limiter")  -- HIGHEST (1000)
			priorities[2] = hooks.get_hook_priority("prompt_modifier") -- HIGH (100)
			priorities[3] = hooks.get_hook_priority("token_counter")   -- NORMAL (0)
			priorities[4] = hooks.get_hook_priority("tool_logger")     -- LOW (-100)
			
			assert(priorities[1] > priorities[2], "rate_limiter should have higher priority")
			assert(priorities[2] > priorities[3], "prompt_modifier should have higher priority than token_counter")
			assert(priorities[3] > priorities[4], "token_counter should have higher priority than tool_logger")
			
			-- 6. Clean up specific hook types
			hooks.clear_hooks("beforeToolCall")
			local before_tool_hooks = hooks.list_hooks({type = "beforeToolCall"})
			assert(#before_tool_hooks == 0, "should have no beforeToolCall hooks")
			
			-- Verify other hooks still exist
			local after_tool_hooks = hooks.list_hooks({type = "afterToolCall"})
			assert(#after_tool_hooks == 1, "afterToolCall hooks should still exist")
		`)
		assert.NoError(t, err)
	})
}