// ABOUTME: Tests for function conversion handlers - Go function to LFunction wrapping
// ABOUTME: Validates argument conversion, return value handling, panic recovery, and variadic functions

package gopherlua

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestFunctionConverter_BasicFunctionWrapping(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("simple_string_function", func(t *testing.T) {
		goFunc := func(input string) string {
			return "Hello, " + input
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)
		assert.Equal(t, lua.LTFunction, luaFunc.Type())

		// Test calling the wrapped function
		L.Push(luaFunc)
		L.Push(lua.LString("World"))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, "Hello, World", result.String())
		L.Pop(1)
	})

	t.Run("math_function", func(t *testing.T) {
		goFunc := func(a, b float64) float64 {
			return a + b
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test calling the wrapped function
		L.Push(luaFunc)
		L.Push(lua.LNumber(10))
		L.Push(lua.LNumber(20))
		err = L.PCall(2, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, float64(30), float64(result.(lua.LNumber)))
		L.Pop(1)
	})

	t.Run("boolean_function", func(t *testing.T) {
		goFunc := func(value int) bool {
			return value > 0
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test with positive value
		L.Push(luaFunc)
		L.Push(lua.LNumber(5))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, true, bool(result.(lua.LBool)))
		L.Pop(1)

		// Test with negative value
		L.Push(luaFunc)
		L.Push(lua.LNumber(-5))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result = L.Get(-1)
		assert.Equal(t, false, bool(result.(lua.LBool)))
		L.Pop(1)
	})
}

func TestFunctionConverter_MultipleReturnValues(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("two_return_values", func(t *testing.T) {
		goFunc := func(input string) (string, int) {
			return input + "_processed", len(input)
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test calling the wrapped function
		L.Push(luaFunc)
		L.Push(lua.LString("test"))
		err = L.PCall(1, 2, nil)
		require.NoError(t, err)

		// Check first return value
		result1 := L.Get(-2)
		assert.Equal(t, "test_processed", result1.String())

		// Check second return value
		result2 := L.Get(-1)
		assert.Equal(t, float64(4), float64(result2.(lua.LNumber)))

		L.Pop(2)
	})

	t.Run("three_return_values", func(t *testing.T) {
		goFunc := func(a, b int) (int, int, int) {
			return a + b, a - b, a * b
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test calling the wrapped function
		L.Push(luaFunc)
		L.Push(lua.LNumber(10))
		L.Push(lua.LNumber(3))
		err = L.PCall(2, 3, nil)
		require.NoError(t, err)

		// Check return values
		sum := L.Get(-3)
		diff := L.Get(-2)
		prod := L.Get(-1)

		assert.Equal(t, float64(13), float64(sum.(lua.LNumber)))
		assert.Equal(t, float64(7), float64(diff.(lua.LNumber)))
		assert.Equal(t, float64(30), float64(prod.(lua.LNumber)))

		L.Pop(3)
	})
}

func TestFunctionConverter_ErrorHandling(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("function_with_error_return", func(t *testing.T) {
		goFunc := func(input string) (string, error) {
			if input == "error" {
				return "", fmt.Errorf("test error")
			}
			return "success: " + input, nil
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test successful call
		L.Push(luaFunc)
		L.Push(lua.LString("test"))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, "success: test", result.String())
		L.Pop(1)

		// Test error case
		L.Push(luaFunc)
		L.Push(lua.LString("error"))
		err = L.PCall(1, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
	})

	t.Run("panic_recovery", func(t *testing.T) {
		goFunc := func(input string) string {
			if input == "panic" {
				panic("test panic")
			}
			return "no panic"
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test normal operation
		L.Push(luaFunc)
		L.Push(lua.LString("normal"))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, "no panic", result.String())
		L.Pop(1)

		// Test panic recovery
		L.Push(luaFunc)
		L.Push(lua.LString("panic"))
		err = L.PCall(1, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic")
	})
}

func TestFunctionConverter_VariadicFunctions(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("variadic_sum", func(t *testing.T) {
		goFunc := func(numbers ...float64) float64 {
			sum := 0.0
			for _, n := range numbers {
				sum += n
			}
			return sum
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test with multiple arguments
		L.Push(luaFunc)
		L.Push(lua.LNumber(1))
		L.Push(lua.LNumber(2))
		L.Push(lua.LNumber(3))
		L.Push(lua.LNumber(4))
		err = L.PCall(4, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, float64(10), float64(result.(lua.LNumber)))
		L.Pop(1)
	})

	t.Run("variadic_string_join", func(t *testing.T) {
		goFunc := func(separator string, parts ...string) string {
			if len(parts) == 0 {
				return ""
			}
			result := parts[0]
			for i := 1; i < len(parts); i++ {
				result += separator + parts[i]
			}
			return result
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test with multiple string arguments
		L.Push(luaFunc)
		L.Push(lua.LString(", "))
		L.Push(lua.LString("one"))
		L.Push(lua.LString("two"))
		L.Push(lua.LString("three"))
		err = L.PCall(4, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, "one, two, three", result.String())
		L.Pop(1)
	})
}

func TestFunctionConverter_ArgumentValidation(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("wrong_argument_count", func(t *testing.T) {
		goFunc := func(a, b int) int {
			return a + b
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test with wrong number of arguments
		L.Push(luaFunc)
		L.Push(lua.LNumber(10))
		// Missing second argument
		err = L.PCall(1, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "argument")
	})

	t.Run("wrong_argument_type", func(t *testing.T) {
		goFunc := func(input int) int {
			return input * 2
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Test with wrong argument type
		L.Push(luaFunc)
		L.Push(lua.LString("not a number"))
		err = L.PCall(1, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected number")
	})
}

func TestFunctionConverter_ComplexTypes(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("slice_parameter", func(t *testing.T) {
		goFunc := func(items []string) int {
			return len(items)
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Create a Lua table (will be converted to slice)
		table := L.NewTable()
		table.RawSetInt(1, lua.LString("one"))
		table.RawSetInt(2, lua.LString("two"))
		table.RawSetInt(3, lua.LString("three"))

		L.Push(luaFunc)
		L.Push(table)
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, float64(3), float64(result.(lua.LNumber)))
		L.Pop(1)
	})

	t.Run("map_parameter", func(t *testing.T) {
		goFunc := func(data map[string]interface{}) int {
			return len(data)
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		// Create a Lua table (will be converted to map)
		table := L.NewTable()
		table.RawSetString("key1", lua.LString("value1"))
		table.RawSetString("key2", lua.LNumber(42))
		table.RawSetString("key3", lua.LBool(true))

		L.Push(luaFunc)
		L.Push(table)
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, float64(3), float64(result.(lua.LNumber)))
		L.Pop(1)
	})
}

func TestFunctionConverter_InvalidFunctions(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name      string
		function  interface{}
		errorText string
	}{
		{
			name:      "not_a_function",
			function:  "not a function",
			errorText: "expected function",
		},
		{
			name:      "nil_function",
			function:  nil,
			errorText: "function cannot be nil",
		},
		{
			name: "unsupported_parameter_type",
			function: func(ch chan int) {
				// Channels are not supported
			},
			errorText: "unsupported parameter type",
		},
		{
			name: "unsupported_return_type",
			function: func() chan int {
				return make(chan int)
			},
			errorText: "unsupported return type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := converter.WrapGoFunction(L, tt.function)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorText)
		})
	}
}

func TestFunctionConverter_PerformanceAndEdgeCases(t *testing.T) {
	converter := NewFunctionConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("function_with_no_parameters", func(t *testing.T) {
		goFunc := func() string {
			return "hello"
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		L.Push(luaFunc)
		err = L.PCall(0, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Equal(t, "hello", result.String())
		L.Pop(1)
	})

	t.Run("function_with_no_return_values", func(t *testing.T) {
		called := false
		goFunc := func(input string) {
			called = true
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		L.Push(luaFunc)
		L.Push(lua.LString("test"))
		err = L.PCall(1, 0, nil)
		require.NoError(t, err)

		assert.True(t, called)
	})

	t.Run("function_with_interface_parameter", func(t *testing.T) {
		goFunc := func(value interface{}) string {
			return fmt.Sprintf("%T", value)
		}

		luaFunc, err := converter.WrapGoFunction(L, goFunc)
		require.NoError(t, err)

		L.Push(luaFunc)
		L.Push(lua.LString("test"))
		err = L.PCall(1, 1, nil)
		require.NoError(t, err)

		result := L.Get(-1)
		assert.Contains(t, result.String(), "string")
		L.Pop(1)
	})
}
