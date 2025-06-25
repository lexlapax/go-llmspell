// ABOUTME: Integration tests for stdlib modules verifying full Lua → Adapter → Bridge flow
// ABOUTME: Tests that snake_case Lua APIs correctly call through to camelCase adapter methods

package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// TestStdlibIntegrationFlow tests the complete flow from Lua snake_case API
// through to adapter camelCase methods for key stdlib modules
func TestStdlibIntegrationFlow(t *testing.T) {
	t.Run("state module integration", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Track adapter method calls
		var adapterCalls []string
		
		// Setup mock state bridge
		bridges := L.NewTable()
		L.SetGlobal("bridges", bridges)
		
		// Create mock context for state operations
		mockContext := L.NewTable()
		
		stateBridge := L.NewTable()
		stateBridge.RawSetString("createSharedContext", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "createSharedContext")
			L.Push(mockContext)
			return 1
		}))
		stateBridge.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "get")
			_ = L.CheckTable(1) // context
			key := L.CheckString(2)
			if key == "test_key" {
				L.Push(lua.LString("test_value"))
			} else {
				L.Push(lua.LNil)
			}
			return 1
		}))
		stateBridge.RawSetString("set", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "set")
			_ = L.CheckTable(1) // context
			_ = L.CheckString(2) // key
			_ = L.Get(3) // value
			L.Push(lua.LBool(true))
			return 1
		}))
		stateBridge.RawSetString("update", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "update")
			_ = L.CheckTable(1) // context
			_ = L.CheckString(2) // key
			_ = L.CheckFunction(3) // updater function
			L.Push(lua.LBool(true))
			return 1
		}))
		stateBridge.RawSetString("keys", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "keys")
			L.Push(L.NewTable()) // empty keys
			return 1
		}))
		
		bridges.RawSetString("state_manager", stateBridge)
		LoadModule(t, L, "state")

		// Test snake_case Lua API calls adapter methods
		err := L.DoString(`
			local state = require("state")
			
			-- Test state.get (snake_case) calls adapter get (camelCase)
			local val = state.get("test_key")
			assert(val == "test_value", "should get value")
			
			-- Test state.set (snake_case) calls adapter set (camelCase)
			state.set("key", "value")
			
			-- Test state.update (snake_case) calls adapter update (camelCase)
			state.update("key", function(old) return "new" end)
		`)
		assert.NoError(t, err)
		
		// Verify adapter methods were called
		// Note: state.update() internally calls get() and set()
		expected := []string{
			"createSharedContext", // from ensure_initialized
			"get",                 // from state.get("test_key")
			"set",                 // from state.set("key", "value")
			"get",                 // from state.update() -> state.get()
			"set",                 // from state.update() -> state.set()
		}
		assert.Equal(t, expected, adapterCalls)
	})

	t.Run("workflow module integration", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Track adapter method calls
		var adapterCalls []string
		
		// Setup mock workflow bridge
		bridges := L.NewTable()
		L.SetGlobal("bridges", bridges)
		
		workflowBridge := L.NewTable()
		workflowBridge.RawSetString("createWorkflow", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "createWorkflow")
			workflow := L.NewTable()
			workflow.RawSetString("id", lua.LString(L.CheckString(1)))
			workflow.RawSetString("status", lua.LString("created"))
			L.Push(workflow)
			return 1
		}))
		workflowBridge.RawSetString("executeWorkflow", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "executeWorkflow")
			result := L.NewTable()
			result.RawSetString("status", lua.LString("completed"))
			L.Push(result)
			return 1
		}))
		workflowBridge.RawSetString("createBuilder", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "createBuilder")
			builder := L.NewTable()
			
			// Builder methods
			builder.RawSetString("withType", L.NewFunction(func(L *lua.LState) int {
				adapterCalls = append(adapterCalls, "builder.withType")
				L.Push(builder) // Return self for chaining
				return 1
			}))
			builder.RawSetString("build", L.NewFunction(func(L *lua.LState) int {
				adapterCalls = append(adapterCalls, "builder.build")
				workflow := L.NewTable()
				workflow.RawSetString("id", lua.LString("built_workflow"))
				L.Push(workflow)
				return 1
			}))
			
			L.Push(builder)
			return 1
		}))
		
		bridges.RawSetString("agent_workflow", workflowBridge)
		LoadModule(t, L, "workflow")

		// Test snake_case Lua API calls adapter methods
		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Test workflow.create_workflow (snake_case) calls adapter createWorkflow (camelCase)
			local wf = workflow.create_workflow("test_id", {name = "test"})
			assert(wf.id == "test_id", "should create workflow")
			
			-- Test workflow.execute_workflow (snake_case) calls adapter executeWorkflow (camelCase)
			local result = workflow.execute_workflow("test_id")
			assert(result.status == "completed", "should execute workflow")
			
			-- Test builder pattern with snake_case (Lua-side builder)
			local builder = workflow.builder:new("builder_test")
			local built = builder:with_type("sequential"):build()
			assert(built.id == "builder_test", "should build workflow")
		`)
		assert.NoError(t, err)
		
		// Verify adapter methods were called
		// Note: Lua-side builder doesn't call adapter methods until build()
		expected := []string{
			"createWorkflow",    // from create_workflow
			"executeWorkflow",   // from execute_workflow
			"createWorkflow",    // from builder:build()
		}
		assert.Equal(t, expected, adapterCalls)
	})

	t.Run("hooks module integration", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Track adapter method calls
		var adapterCalls []string
		
		// Setup mock hooks bridge
		bridges := L.NewTable()
		L.SetGlobal("bridges", bridges)
		
		hooksBridge := L.NewTable()
		hooksBridge.RawSetString("registerHook", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "registerHook")
			hook := L.NewTable()
			hook.RawSetString("id", lua.LString(L.CheckString(1)))
			L.Push(hook)
			return 1
		}))
		hooksBridge.RawSetString("enableHook", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "enableHook")
			L.Push(lua.LBool(true))
			return 1
		}))
		hooksBridge.RawSetString("createHook", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "createHook")
			builder := L.NewTable()
			
			// Builder methods use camelCase in adapter
			builder.RawSetString("withPriority", L.NewFunction(func(L *lua.LState) int {
				adapterCalls = append(adapterCalls, "builder.withPriority")
				L.Push(builder)
				return 1
			}))
			builder.RawSetString("register", L.NewFunction(func(L *lua.LState) int {
				adapterCalls = append(adapterCalls, "builder.register")
				hook := L.NewTable()
				hook.RawSetString("id", lua.LString("built_hook"))
				L.Push(hook)
				return 1
			}))
			
			L.Push(builder)
			return 1
		}))
		
		bridges.RawSetString("agent_hooks", hooksBridge)
		LoadModule(t, L, "hooks")

		// Test snake_case Lua API calls adapter methods
		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Test hooks.register_hook (snake_case) calls adapter registerHook (camelCase)
			local hook = hooks.register_hook("test_hook", {})
			assert(hook.id == "test_hook", "should register hook")
			
			-- Test hooks.enable_hook (snake_case) calls adapter enableHook (camelCase)
			local enabled = hooks.enable_hook("test_hook")
			assert(enabled == true, "should enable hook")
			
			-- Test builder pattern - Lua uses snake_case, adapter uses camelCase
			local built = hooks.create_hook("builder_hook")
				:withPriority(100)
				:register()
			assert(built.id == "built_hook", "should build hook")
		`)
		assert.NoError(t, err)
		
		// Verify adapter methods were called
		expected := []string{
			"registerHook",
			"enableHook",
			"createHook",
			"builder.withPriority",
			"builder.register",
		}
		assert.Equal(t, expected, adapterCalls)
	})

	t.Run("modelinfo module integration", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Track adapter method calls
		var adapterCalls []string
		
		// Setup mock modelinfo bridge
		bridges := L.NewTable()
		L.SetGlobal("bridges", bridges)
		
		modelInfoBridge := L.NewTable()
		modelInfoBridge.RawSetString("discoveryScan", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "discoveryScan")
			result := L.NewTable()
			result.RawSetString("models", lua.LNumber(3))
			L.Push(result)
			return 1
		}))
		modelInfoBridge.RawSetString("capabilitiesCheck", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "capabilitiesCheck")
			caps := L.NewTable()
			caps.RawSetString("functionCalling", lua.LBool(true))
			L.Push(caps)
			return 1
		}))
		modelInfoBridge.RawSetString("selectionFind", L.NewFunction(func(L *lua.LState) int {
			adapterCalls = append(adapterCalls, "selectionFind")
			model := L.NewTable()
			model.RawSetString("name", lua.LString("gpt-4"))
			L.Push(model)
			return 1
		}))
		
		bridges.RawSetString("llm_modelinfo", modelInfoBridge)
		LoadModule(t, L, "modelinfo")

		// Test snake_case Lua API calls adapter methods
		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test modelinfo.discovery_scan (snake_case) calls adapter discoveryScan (camelCase)
			local scan = modelinfo.discovery_scan()
			assert(scan.models == 3, "should scan models")
			
			-- Test modelinfo.capabilities_check (snake_case) calls adapter capabilitiesCheck (camelCase)
			local caps = modelinfo.capabilities_check("gpt-4")
			assert(caps.functionCalling == true, "should check capabilities")
			
			-- Test modelinfo.selection_find (snake_case) calls adapter selectionFind (camelCase)
			local model = modelinfo.selection_find({priority = "cost"})
			assert(model.name == "gpt-4", "should find model")
		`)
		assert.NoError(t, err)
		
		// Verify adapter methods were called
		expected := []string{
			"discoveryScan",
			"capabilitiesCheck",
			"selectionFind",
		}
		assert.Equal(t, expected, adapterCalls)
	})
}

// TestBuilderPatternNaming verifies that builder patterns use consistent snake_case naming
func TestBuilderPatternNaming(t *testing.T) {
	t.Run("workflow builder uses snake_case", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Builder methods should use snake_case
			local wf = workflow.builder:new("test_workflow")
				:with_type("sequential")       -- snake_case
				:with_name("Test Workflow")    -- snake_case
				:with_description("A test")    -- snake_case
				:build()
			
			assert(wf.id == "test_workflow", "should set id")
			assert(wf.status == "created", "should have created status")
		`)
		assert.NoError(t, err)
	})

	t.Run("hooks builder uses snake_case in Lua", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupHooksBridge(t, L)
		LoadModule(t, L, "hooks")

		err := L.DoString(`
			local hooks = require("hooks")
			
			-- Lua builder helper uses snake_case
			local hook = hooks.new_builder("test_hook")
				:with_priority(hooks.PRIORITY.HIGH)     -- snake_case
				:before_generate(function() end)        -- snake_case
				:after_generate(function() end)         -- snake_case
				:before_tool_call(function() end)       -- snake_case
				:after_tool_call(function() end)        -- snake_case
				:register()
			
			assert(hook.id == "test_hook", "should create hook")
			assert(hook.priority == 100, "should set priority")
		`)
		assert.NoError(t, err)
	})
}

// TestNamespaceAPIs verifies that namespace APIs use snake_case
func TestNamespaceAPIs(t *testing.T) {
	t.Run("modelinfo namespaces use snake_case", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Namespace methods should use snake_case
			assert(type(modelinfo.discovery.scan) == "function", "discovery.scan")
			assert(type(modelinfo.discovery.get_providers) == "function", "discovery.get_providers")
			
			assert(type(modelinfo.capabilities.check) == "function", "capabilities.check")
			assert(type(modelinfo.capabilities.get_details) == "function", "capabilities.get_details")
			
			assert(type(modelinfo.selection.find) == "function", "selection.find")
			assert(type(modelinfo.selection.rank) == "function", "selection.rank")
			
			-- Test they work
			local scan = modelinfo.discovery.scan()
			assert(scan.models == 3, "namespace method should work")
		`)
		assert.NoError(t, err)
	})
}