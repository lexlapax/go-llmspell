// ABOUTME: Tests for Lua state management library - matches current state.lua implementation
// ABOUTME: Tests state.get/set/update/delete/has/keys/clear/snapshot/save/load/export/import methods

package stdlib

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockStateBridge creates a mock state_manager bridge that matches what state.lua expects
func setupMockStateBridge(L *lua.LState) {
	// Create mock bridges table if it doesn't exist
	bridges := L.GetGlobal("bridges")
	if bridges == lua.LNil {
		bridges = L.NewTable()
		L.SetGlobal("bridges", bridges)
	}
	bridgesTable := bridges.(*lua.LTable)

	// Create mock state_manager bridge
	stateManager := L.NewTable()

	// Mock storage for the test
	contextStorage := L.NewTable()

	// Mock createSharedContext
	stateManager.RawSetString("createSharedContext", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		// Return a mock context string
		L.Push(lua.LString("mock_context_" + name))
		return 1
	}))

	// Mock get
	stateManager.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		key := L.CheckString(2)
		
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			L.Push(lua.LNil)
			return 1
		}
		
		value := contextTable.(*lua.LTable).RawGetString(key)
		L.Push(value)
		return 1
	}))

	// Mock set
	stateManager.RawSetString("set", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		key := L.CheckString(2)
		value := L.Get(3)
		
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			contextTable = L.NewTable()
			contextStorage.RawSetString(context, contextTable)
		}
		
		contextTable.(*lua.LTable).RawSetString(key, value)
		return 0
	}))

	// Mock has
	stateManager.RawSetString("has", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		key := L.CheckString(2)
		
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			L.Push(lua.LFalse)
			return 1
		}
		
		value := contextTable.(*lua.LTable).RawGetString(key)
		L.Push(lua.LBool(value != lua.LNil))
		return 1
	}))

	// Mock keys
	stateManager.RawSetString("keys", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			L.Push(L.NewTable())
			return 1
		}
		
		keysTable := L.NewTable()
		i := 1
		contextTable.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			keysTable.RawSetInt(i, k)
			i++
		})
		
		L.Push(keysTable)
		return 1
	}))

	// Mock delete
	stateManager.RawSetString("delete", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		key := L.CheckString(2)
		
		contextTable := contextStorage.RawGetString(context)
		if contextTable != lua.LNil {
			contextTable.(*lua.LTable).RawSetString(key, lua.LNil)
		}
		return 0
	}))

	// Mock clearContext
	stateManager.RawSetString("clearContext", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		contextStorage.RawSetString(context, lua.LNil)
		return 0
	}))

	// Mock createSnapshot
	stateManager.RawSetString("createSnapshot", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			L.Push(L.NewTable())
		} else {
			// Return a copy of the context
			snapshot := L.NewTable()
			contextTable.(*lua.LTable).ForEach(func(k, v lua.LValue) {
				snapshot.RawSet(k, v)
			})
			L.Push(snapshot)
		}
		return 1
	}))

	// Mock saveState
	stateManager.RawSetString("saveState", L.NewFunction(func(L *lua.LState) int {
		// context := L.CheckString(1)
		// name := L.CheckString(2)
		L.Push(lua.LBool(true))
		return 1
	}))

	// Mock loadState
	stateManager.RawSetString("loadState", L.NewFunction(func(L *lua.LState) int {
		// context := L.CheckString(1)
		// name := L.CheckString(2)
		L.Push(L.NewTable())
		return 1
	}))

	// Mock exportState
	stateManager.RawSetString("exportState", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		contextTable := contextStorage.RawGetString(context)
		if contextTable == lua.LNil {
			L.Push(L.NewTable())
		} else {
			L.Push(contextTable)
		}
		return 1
	}))

	// Mock importState
	stateManager.RawSetString("importState", L.NewFunction(func(L *lua.LState) int {
		context := L.CheckString(1)
		data := L.CheckTable(2)
		contextStorage.RawSetString(context, data)
		return 0
	}))

	// Set the state_manager bridge
	bridgesTable.RawSetString("state_manager", stateManager)
}

func TestStateModuleBasicOperations(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Setup mock bridge
	setupMockStateBridge(L)

	// Load state.lua and get the returned module
	err := L.DoString(`state = dofile("state.lua")`)
	require.NoError(t, err, "Failed to load state.lua")

	// Get state module
	stateModule := L.GetGlobal("state")
	require.NotEqual(t, lua.LNil, stateModule, "state module should be loaded")

	t.Run("set_and_get_simple_value", func(t *testing.T) {
		err := L.DoString(`
			state.set("test_key", "test_value")
			result = state.get("test_key")
		`)
		require.NoError(t, err)

		result := L.GetGlobal("result")
		assert.Equal(t, "test_value", string(result.(lua.LString)))
	})

	t.Run("set_and_get_complex_path", func(t *testing.T) {
		err := L.DoString(`
			state.set("user.name", "John")
			state.set("user.age", 30)
			name_result = state.get("user.name")
			age_result = state.get("user.age")
			user_result = state.get("user")
		`)
		require.NoError(t, err)

		nameResult := L.GetGlobal("name_result")
		ageResult := L.GetGlobal("age_result")
		userResult := L.GetGlobal("user_result")

		assert.Equal(t, "John", string(nameResult.(lua.LString)))
		assert.Equal(t, lua.LNumber(30), ageResult.(lua.LNumber))
		assert.Equal(t, lua.LTTable, userResult.Type())
	})

	t.Run("has_method", func(t *testing.T) {
		err := L.DoString(`
			state.set("exists", "value")
			has_exists = state.has("exists")
			has_missing = state.has("missing")
		`)
		require.NoError(t, err)

		hasExists := L.GetGlobal("has_exists")
		hasMissing := L.GetGlobal("has_missing")

		assert.Equal(t, lua.LTrue, hasExists)
		assert.Equal(t, lua.LFalse, hasMissing)
	})

	t.Run("delete_method", func(t *testing.T) {
		err := L.DoString(`
			state.set("to_delete", "value")
			before_delete = state.has("to_delete")
			state.delete("to_delete")
			after_delete = state.has("to_delete")
		`)
		require.NoError(t, err)

		beforeDelete := L.GetGlobal("before_delete")
		afterDelete := L.GetGlobal("after_delete")

		assert.Equal(t, lua.LTrue, beforeDelete)
		assert.Equal(t, lua.LFalse, afterDelete)
	})

	t.Run("keys_method", func(t *testing.T) {
		err := L.DoString(`
			state.clear()  -- Clear previous state
			state.set("key1", "value1")
			state.set("key2", "value2")
			keys_result = state.keys()
			keys_count = #keys_result
		`)
		require.NoError(t, err)

		keysCount := L.GetGlobal("keys_count")
		assert.Equal(t, lua.LNumber(2), keysCount.(lua.LNumber))
	})

	t.Run("update_method", func(t *testing.T) {
		err := L.DoString(`
			state.set("counter", 5)
			new_value = state.update("counter", function(old_value)
				return old_value + 10
			end)
			final_value = state.get("counter")
		`)
		require.NoError(t, err)

		newValue := L.GetGlobal("new_value")
		finalValue := L.GetGlobal("final_value")

		assert.Equal(t, lua.LNumber(15), newValue.(lua.LNumber))
		assert.Equal(t, lua.LNumber(15), finalValue.(lua.LNumber))
	})

	t.Run("clear_method", func(t *testing.T) {
		err := L.DoString(`
			state.set("before_clear", "value")
			state.clear()
			after_clear = state.get("before_clear")
		`)
		require.NoError(t, err)

		afterClear := L.GetGlobal("after_clear")
		assert.Equal(t, lua.LNil, afterClear)
	})

	t.Run("snapshot_method", func(t *testing.T) {
		err := L.DoString(`
			state.set("snap_key", "snap_value")
			snapshot_result = state.snapshot()
			snapshot_type = type(snapshot_result)
		`)
		require.NoError(t, err)

		snapshotType := L.GetGlobal("snapshot_type")
		assert.Equal(t, "table", string(snapshotType.(lua.LString)))
	})

	t.Run("save_and_load_methods", func(t *testing.T) {
		err := L.DoString(`
			save_result = state.save("test_save")
			load_result = state.load("test_save")
			save_success = (save_result == true)
			load_success = (type(load_result) == "table")
		`)
		require.NoError(t, err)

		saveSuccess := L.GetGlobal("save_success")
		loadSuccess := L.GetGlobal("load_success")

		assert.Equal(t, lua.LTrue, saveSuccess)
		assert.Equal(t, lua.LTrue, loadSuccess)
	})

	t.Run("export_and_import_methods", func(t *testing.T) {
		err := L.DoString(`
			state.set("export_key", "export_value")
			exported_data = state.export()
			
			state.clear()
			state.import(exported_data)
			imported_value = state.get("export_key")
		`)
		require.NoError(t, err)

		importedValue := L.GetGlobal("imported_value")
		assert.Equal(t, "export_value", string(importedValue.(lua.LString)))
	})
}

func TestStateModuleErrorHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Setup mock bridge
	setupMockStateBridge(L)

	// Load state.lua and get the returned module
	err := L.DoString(`state = dofile("state.lua")`)
	require.NoError(t, err, "Failed to load state.lua")

	t.Run("set_with_empty_path_errors", func(t *testing.T) {
		err := L.DoString(`
			local success, err = pcall(function()
				state.set("", "value")
			end)
			set_empty_success = success
		`)
		require.NoError(t, err)

		success := L.GetGlobal("set_empty_success")
		assert.Equal(t, lua.LFalse, success)
	})

	t.Run("update_with_non_function_errors", func(t *testing.T) {
		err := L.DoString(`
			local success, err = pcall(function()
				state.update("key", "not_a_function")
			end)
			update_error_success = success
		`)
		require.NoError(t, err)

		success := L.GetGlobal("update_error_success")
		assert.Equal(t, lua.LFalse, success)
	})

	t.Run("delete_with_empty_path_errors", func(t *testing.T) {
		err := L.DoString(`
			local success, err = pcall(function()
				state.delete("")
			end)
			delete_empty_success = success
		`)
		require.NoError(t, err)

		success := L.GetGlobal("delete_empty_success")
		assert.Equal(t, lua.LFalse, success)
	})
}

func TestStateModuleWithoutBridge(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Load state.lua without setting up bridges
	err := L.DoString(`state = dofile("state.lua")`)
	require.NoError(t, err, "Failed to load state.lua")

	t.Run("operations_fail_without_bridge", func(t *testing.T) {
		err := L.DoString(`
			local success, err = pcall(function()
				state.get("test")
			end)
			no_bridge_success = success
		`)
		require.NoError(t, err)

		success := L.GetGlobal("no_bridge_success")
		assert.Equal(t, lua.LFalse, success, "Operations should fail gracefully when bridge is not available")
	})
}

func TestStateModuleNewMethods(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Setup mock bridge
	setupMockStateBridge(L)

	// Load state.lua and get the returned module
	err := L.DoString(`state = dofile("state.lua")`)
	require.NoError(t, err, "Failed to load state.lua")

	t.Run("new_methods_exist", func(t *testing.T) {
		err := L.DoString(`
			-- Test that new methods exist
			create_exists = (type(state.create) == "function")
			values_exists = (type(state.values) == "function")
			list_states_exists = (type(state.list_states) == "function")
			metadata_exists = (type(state.set_metadata) == "function")
			artifacts_exists = (type(state.add_artifact) == "function")
			transforms_exists = (type(state.transforms) == "table")
			context_exists = (type(state.context) == "table")
			persistence_exists = (type(state.persistence) == "table")
		`)
		require.NoError(t, err)

		// Verify new methods exist
		assert.Equal(t, lua.LTrue, L.GetGlobal("create_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("values_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("list_states_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("metadata_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("artifacts_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("transforms_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("context_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("persistence_exists"))
	})

	t.Run("namespace_methods_exist", func(t *testing.T) {
		err := L.DoString(`
			-- Test that namespace methods exist
			transforms_apply_exists = (type(state.transforms.apply) == "function")
			transforms_register_exists = (type(state.transforms.register) == "function")
			context_get_exists = (type(state.context.get) == "function")
			context_set_exists = (type(state.context.set) == "function")
			persistence_save_exists = (type(state.persistence.save) == "function")
			persistence_load_exists = (type(state.persistence.load) == "function")
		`)
		require.NoError(t, err)

		// Verify namespace methods exist
		assert.Equal(t, lua.LTrue, L.GetGlobal("transforms_apply_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("transforms_register_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("context_get_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("context_set_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("persistence_save_exists"))
		assert.Equal(t, lua.LTrue, L.GetGlobal("persistence_load_exists"))
	})
}