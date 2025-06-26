// ABOUTME: Direct test of UtilsAdapter JSON functionality
// ABOUTME: Isolate adapter-level JSON issues without Lua

package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestUtilsAdapterJSONDirect tests JSON functionality at adapter level
func TestUtilsAdapterJSONDirect(t *testing.T) {
	// Create a real JSON bridge with proper marshal implementation
	jsonBridge := testutils.NewMockBridge("util_json").
		WithInitialized(true).
		WithMethod("marshal", engine.MethodInfo{
			Name: "marshal",
		}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
			// This simulates what the real bridge does - returns a StringValue
			return engine.NewStringValue(`{"test":"value","number":42}`), nil
		})

	// Create UtilsAdapter with only JSON bridge
	adapter := NewUtilsAdapter(nil, nil, nil, jsonBridge, nil, nil, nil, nil)
	require.NotNil(t, adapter)

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Get the module (this is what bridges.util_core becomes)
	moduleFn := adapter.CreateLuaModule()
	require.NotNil(t, moduleFn)

	// Call the module function to get the module table
	L.Push(L.NewFunction(moduleFn))
	err := L.PCall(0, 1, nil)
	require.NoError(t, err)

	// Get the module table from stack
	module := L.CheckTable(-1)
	L.Pop(1)

	// Check if jsonEncode exists
	jsonEncodeFn := module.RawGetString("jsonEncode")
	assert.NotEqual(t, lua.LNil, jsonEncodeFn, "jsonEncode should exist")
	assert.Equal(t, lua.LTFunction, jsonEncodeFn.Type(), "jsonEncode should be a function")

	// Test calling jsonEncode
	t.Run("call jsonEncode directly", func(t *testing.T) {
		// Push function
		L.Push(jsonEncodeFn)
		
		// Push argument (a table)
		testTable := L.NewTable()
		L.SetField(testTable, "name", lua.LString("test"))
		L.SetField(testTable, "value", lua.LNumber(42))
		L.Push(testTable)

		// Call function
		err := L.PCall(1, 2, nil)
		require.NoError(t, err)

		// Check results
		errorVal := L.Get(-1)
		resultVal := L.Get(-2)
		L.Pop(2)

		// Should have no error
		assert.Equal(t, lua.LNil, errorVal, "Should have no error")
		
		// Should have JSON string result
		assert.Equal(t, lua.LTString, resultVal.Type(), "Result should be a string")
		jsonStr := resultVal.String()
		assert.Equal(t, `{"test":"value","number":42}`, jsonStr, "Should return JSON string")
	})

	// Test with nil JSON bridge
	t.Run("nil json bridge", func(t *testing.T) {
		// Create adapter with no JSON bridge
		adapterNoJSON := NewUtilsAdapter(nil, nil, nil, nil, nil, nil, nil, nil)
		
		L2 := lua.NewState()
		defer L2.Close()

		// Get module
		moduleFn2 := adapterNoJSON.CreateLuaModule()
		L2.Push(L2.NewFunction(moduleFn2))
		err := L2.PCall(0, 1, nil)
		require.NoError(t, err)

		module2 := L2.CheckTable(-1)
		L2.Pop(1)

		// jsonEncode should still exist
		jsonEncodeFn2 := module2.RawGetString("jsonEncode")
		assert.NotEqual(t, lua.LNil, jsonEncodeFn2)

		// But calling it should return error
		L2.Push(jsonEncodeFn2)
		testTable2 := L2.NewTable()
		L2.SetField(testTable2, "test", lua.LString("value"))
		L2.Push(testTable2)

		err = L2.PCall(1, 2, nil)
		require.NoError(t, err)

		errorVal2 := L2.Get(-1)
		resultVal2 := L2.Get(-2)
		L2.Pop(2)

		// Should have error
		assert.Equal(t, lua.LNil, resultVal2, "Result should be nil")
		assert.Equal(t, lua.LTString, errorVal2.Type(), "Error should be a string")
		assert.Equal(t, "json bridge not initialized", errorVal2.String())
	})
}