// ABOUTME: Tests for the Lua tool implementation
// ABOUTME: Verifies Lua function wrapping and execution as tools

package bridges

import (
	"context"
	"strings"
	"testing"

	engLua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestNewLuaTool(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a simple Lua function
	err := L.DoString(`
		function testFunc(params)
			return "Hello, " .. params.name
		end
	`)
	require.NoError(t, err)

	fn := L.GetGlobal("testFunc").(*lua.LFunction)
	converter := engLua.NewLuaConverter(L)

	tool := NewLuaTool(
		"test_tool",
		"A test tool",
		map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name to greet",
			},
		},
		fn,
		L,
		converter,
	)

	assert.NotNil(t, tool)
	assert.Equal(t, "test_tool", tool.name)
	assert.Equal(t, "A test tool", tool.description)
	assert.NotNil(t, tool.parameters)
	assert.NotNil(t, tool.fn)
	assert.NotNil(t, tool.L)
	assert.NotNil(t, tool.converter)
}

func TestLuaToolExecute(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := engLua.NewLuaConverter(L)

	tests := []struct {
		name           string
		luaCode        string
		params         map[string]interface{}
		expectedResult interface{}
		expectError    bool
		errorContains  string
	}{
		{
			name: "simple string concatenation",
			luaCode: `
				function concat(params)
					return params.a .. params.b
				end
			`,
			params: map[string]interface{}{
				"a": "Hello, ",
				"b": "World!",
			},
			expectedResult: "Hello, World!",
		},
		{
			name: "numeric calculation",
			luaCode: `
				function calculate(params)
					return params.x + params.y * 2
				end
			`,
			params: map[string]interface{}{
				"x": float64(10),
				"y": float64(5),
			},
			expectedResult: float64(20),
		},
		{
			name: "return table",
			luaCode: `
				function process(params)
					return {
						input = params.text,
						length = string.len(params.text),
						upper = string.upper(params.text)
					}
				end
			`,
			params: map[string]interface{}{
				"text": "hello",
			},
			expectedResult: map[string]interface{}{
				"input":  "hello",
				"length": float64(5),
				"upper":  "HELLO",
			},
		},
		{
			name: "return nil",
			luaCode: `
				function nothing(params)
					return nil
				end
			`,
			params:         map[string]interface{}{},
			expectedResult: nil,
		},
		{
			name: "no return value",
			luaCode: `
				function noreturn(params)
					-- Do nothing
				end
			`,
			params:         map[string]interface{}{},
			expectedResult: nil,
		},
		{
			name: "explicit error return",
			luaCode: `
				function failing(params)
					return nil, "Something went wrong"
				end
			`,
			params:        map[string]interface{}{},
			expectError:   true,
			errorContains: "Something went wrong",
		},
		{
			name: "runtime error",
			luaCode: `
				function buggy(params)
					error("Runtime error!")
				end
			`,
			params:        map[string]interface{}{},
			expectError:   true,
			errorContains: "Runtime error!",
		},
		{
			name: "accessing nested params",
			luaCode: `
				function nested(params)
					return params.user.name .. " is " .. params.user.age .. " years old"
				end
			`,
			params: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"age":  float64(30),
				},
			},
			expectedResult: "Alice is 30 years old",
		},
		{
			name: "array handling",
			luaCode: `
				function sumArray(params)
					local sum = 0
					for i, v in ipairs(params.numbers) do
						sum = sum + v
					end
					return sum
				end
			`,
			params: map[string]interface{}{
				"numbers": []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
			},
			expectedResult: float64(15),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the Lua function
			err := L.DoString(tt.luaCode)
			require.NoError(t, err)

			// Get the function by its known name
			// Extract function name from the Lua code
			var fnName string
			if strings.Contains(tt.luaCode, "function concat") {
				fnName = "concat"
			} else if strings.Contains(tt.luaCode, "function calculate") {
				fnName = "calculate"
			} else if strings.Contains(tt.luaCode, "function process") {
				fnName = "process"
			} else if strings.Contains(tt.luaCode, "function nothing") {
				fnName = "nothing"
			} else if strings.Contains(tt.luaCode, "function noreturn") {
				fnName = "noreturn"
			} else if strings.Contains(tt.luaCode, "function failing") {
				fnName = "failing"
			} else if strings.Contains(tt.luaCode, "function buggy") {
				fnName = "buggy"
			} else if strings.Contains(tt.luaCode, "function nested") {
				fnName = "nested"
			} else if strings.Contains(tt.luaCode, "function sumArray") {
				fnName = "sumArray"
			}
			require.NotEmpty(t, fnName, "Could not determine function name from Lua code")

			fn := L.GetGlobal(fnName).(*lua.LFunction)
			require.NotNil(t, fn, "Function %s not found", fnName)

			// Create the tool
			tool := NewLuaTool(fnName, "Test function", map[string]interface{}{}, fn, L, converter)

			// Execute the tool
			result, err := tool.Execute(context.Background(), tt.params)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestLuaToolConcurrency(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a stateful Lua function
	err := L.DoString(`
		local counter = 0
		function increment(params)
			counter = counter + (params.amount or 1)
			return counter
		end
	`)
	require.NoError(t, err)

	fn := L.GetGlobal("increment").(*lua.LFunction)
	converter := engLua.NewLuaConverter(L)

	tool := NewLuaTool("increment", "Increments counter", map[string]interface{}{}, fn, L, converter)

	// Execute sequentially (Lua is not thread-safe, so the mutex should serialize access)
	results := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"amount": float64(1),
		})
		require.NoError(t, err)
		results[i] = result
	}

	// Check that we got sequential increments
	for i, result := range results {
		assert.Equal(t, float64(i+1), result)
	}
}

func TestLuaToolStackManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := engLua.NewLuaConverter(L)

	// Create a function that pushes multiple values
	err := L.DoString(`
		function multiReturn(params)
			return {first = "first", second = "second", third = "third"}
		end
	`)
	require.NoError(t, err)

	fn := L.GetGlobal("multiReturn").(*lua.LFunction)
	tool := NewLuaTool("multi", "Multi-return function", map[string]interface{}{}, fn, L, converter)

	// Get initial stack size
	initialTop := L.GetTop()

	// Execute multiple times
	for i := 0; i < 5; i++ {
		result, err := tool.Execute(context.Background(), map[string]interface{}{})
		require.NoError(t, err)
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "first", resultMap["first"])

		// Check stack is restored
		assert.Equal(t, initialTop, L.GetTop(), "Stack should be restored after execution")
	}
}

func TestLuaToolErrorHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := engLua.NewLuaConverter(L)

	tests := []struct {
		name          string
		luaCode       string
		errorContains string
	}{
		{
			name: "nil function access",
			luaCode: `
				function badAccess(params)
					local x = nil
					return x.field -- This will error
				end
			`,
			errorContains: "attempt to index",
		},
		{
			name: "type error",
			luaCode: `
				function typeError(params)
					return "string" + 5 -- Type error
				end
			`,
			errorContains: "cannot perform add operation",
		},
		{
			name: "missing parameter",
			luaCode: `
				function needsParam(params)
					return params.required.value -- Will fail if required is nil
				end
			`,
			errorContains: "attempt to index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := L.DoString(tt.luaCode)
			require.NoError(t, err)

			// Get the function by its known name
			var fnName string
			if strings.Contains(tt.luaCode, "function badAccess") {
				fnName = "badAccess"
			} else if strings.Contains(tt.luaCode, "function typeError") {
				fnName = "typeError"
			} else if strings.Contains(tt.luaCode, "function needsParam") {
				fnName = "needsParam"
			}

			fn := L.GetGlobal(fnName).(*lua.LFunction)

			tool := NewLuaTool("error_test", "Error test", map[string]interface{}{}, fn, L, converter)

			// Execute and expect error
			_, err = tool.Execute(context.Background(), map[string]interface{}{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestLuaToolComplexDataTypes(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := engLua.NewLuaConverter(L)

	// Test handling of complex nested structures
	err := L.DoString(`
		function processComplex(params)
			local result = {
				users = {},
				metadata = {
					processed = true,
					timestamp = params.timestamp
				}
			}
			
			for i, user in ipairs(params.users) do
				table.insert(result.users, {
					id = user.id,
					name = string.upper(user.name),
					tags = user.tags
				})
			end
			
			return result
		end
	`)
	require.NoError(t, err)

	fn := L.GetGlobal("processComplex").(*lua.LFunction)
	tool := NewLuaTool("complex", "Complex data processor", map[string]interface{}{}, fn, L, converter)

	input := map[string]interface{}{
		"timestamp": "2024-01-01",
		"users": []interface{}{
			map[string]interface{}{
				"id":   float64(1),
				"name": "alice",
				"tags": []interface{}{"admin", "user"},
			},
			map[string]interface{}{
				"id":   float64(2),
				"name": "bob",
				"tags": []interface{}{"user"},
			},
		},
	}

	result, err := tool.Execute(context.Background(), input)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Check metadata
	metadata, ok := resultMap["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, metadata["processed"])
	assert.Equal(t, "2024-01-01", metadata["timestamp"])

	// Check users
	users, ok := resultMap["users"].([]interface{})
	require.True(t, ok)
	assert.Len(t, users, 2)

	user1, ok := users[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), user1["id"])
	assert.Equal(t, "ALICE", user1["name"])
}
