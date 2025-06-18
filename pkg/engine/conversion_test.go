// ABOUTME: Comprehensive tests for centralized ScriptValue conversion utilities
// ABOUTME: Ensures conversion functions work correctly and handle all edge cases

package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToScriptValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected ScriptValue
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: NewNilValue(),
		},
		{
			name:     "boolean true",
			input:    true,
			expected: NewBoolValue(true),
		},
		{
			name:     "boolean false",
			input:    false,
			expected: NewBoolValue(false),
		},
		{
			name:     "int",
			input:    42,
			expected: NewNumberValue(42),
		},
		{
			name:     "int8",
			input:    int8(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "int16",
			input:    int16(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "int32",
			input:    int32(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "int64",
			input:    int64(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "uint",
			input:    uint(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "uint8",
			input:    uint8(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "uint16",
			input:    uint16(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "uint32",
			input:    uint32(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "uint64",
			input:    uint64(42),
			expected: NewNumberValue(42),
		},
		{
			name:     "float32",
			input:    float32(3.14),
			expected: NewNumberValue(float64(float32(3.14))),
		},
		{
			name:     "float64",
			input:    3.14,
			expected: NewNumberValue(3.14),
		},
		{
			name:     "string",
			input:    "hello",
			expected: NewStringValue("hello"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: NewStringValue(""),
		},
		{
			name:  "simple slice",
			input: []interface{}{"hello", 42, true},
			expected: NewArrayValue([]ScriptValue{
				NewStringValue("hello"),
				NewNumberValue(42),
				NewBoolValue(true),
			}),
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: NewArrayValue([]ScriptValue{}),
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"name":   "test",
				"age":    25,
				"active": true,
			},
			expected: NewObjectValue(map[string]ScriptValue{
				"name":   NewStringValue("test"),
				"age":    NewNumberValue(25),
				"active": NewBoolValue(true),
			}),
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: NewObjectValue(map[string]ScriptValue{}),
		},
		{
			name:     "already ScriptValue",
			input:    NewStringValue("already converted"),
			expected: NewStringValue("already converted"),
		},
		{
			name:     "unknown type fallback",
			input:    struct{ Name string }{Name: "test"},
			expected: NewStringValue("{test}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToScriptValue(tt.input)

			// Compare types
			assert.Equal(t, tt.expected.Type(), result.Type())

			// Compare values by converting both to Go and comparing
			assert.Equal(t, tt.expected.ToGo(), result.ToGo())
		})
	}
}

func TestConvertNestedStructures(t *testing.T) {
	input := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "Alice",
				"age":  30,
				"preferences": map[string]interface{}{
					"theme":         "dark",
					"notifications": true,
				},
			},
			map[string]interface{}{
				"name": "Bob",
				"age":  25,
				"preferences": map[string]interface{}{
					"theme":         "light",
					"notifications": false,
				},
			},
		},
		"count": 2,
	}

	result := ConvertToScriptValue(input)
	require.Equal(t, TypeObject, result.Type())

	obj := result.(ObjectValue)

	// Check users array
	users, exists := obj.Get("users")
	require.True(t, exists)
	require.Equal(t, TypeArray, users.Type())

	usersArray := users.(ArrayValue)
	assert.Equal(t, 2, usersArray.Len())

	// Check first user
	firstUser, exists := usersArray.Get(0)
	require.True(t, exists)
	require.Equal(t, TypeObject, firstUser.Type())

	firstUserObj := firstUser.(ObjectValue)
	name, exists := firstUserObj.Get("name")
	require.True(t, exists)
	assert.Equal(t, "Alice", name.(StringValue).Value())
}

func TestConvertMapToScriptValue(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		result := ConvertMapToScriptValue(nil)
		assert.Nil(t, result)
	})

	t.Run("empty map", func(t *testing.T) {
		result := ConvertMapToScriptValue(map[string]interface{}{})
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("simple map", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "test",
			"age":  25,
		}
		result := ConvertMapToScriptValue(input)

		assert.Len(t, result, 2)
		assert.Equal(t, TypeString, result["name"].Type())
		assert.Equal(t, "test", result["name"].(StringValue).Value())
		assert.Equal(t, TypeNumber, result["age"].Type())
		assert.Equal(t, float64(25), result["age"].(NumberValue).Value())
	})
}

func TestConvertSliceToScriptValue(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := ConvertSliceToScriptValue(nil)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := ConvertSliceToScriptValue([]interface{}{})
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("simple slice", func(t *testing.T) {
		input := []interface{}{"hello", 42, true}
		result := ConvertSliceToScriptValue(input)

		require.Len(t, result, 3)
		assert.Equal(t, TypeString, result[0].Type())
		assert.Equal(t, "hello", result[0].(StringValue).Value())
		assert.Equal(t, TypeNumber, result[1].Type())
		assert.Equal(t, float64(42), result[1].(NumberValue).Value())
		assert.Equal(t, TypeBool, result[2].Type())
		assert.Equal(t, true, result[2].(BoolValue).Value())
	})
}

func TestConvertFromScriptValue(t *testing.T) {
	tests := []struct {
		name     string
		input    ScriptValue
		expected interface{}
	}{
		{
			name:     "nil value",
			input:    NewNilValue(),
			expected: nil,
		},
		{
			name:     "string value",
			input:    NewStringValue("hello"),
			expected: "hello",
		},
		{
			name:     "number value",
			input:    NewNumberValue(42),
			expected: float64(42),
		},
		{
			name:     "boolean value",
			input:    NewBoolValue(true),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFromScriptValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("nil ScriptValue", func(t *testing.T) {
		result := ConvertFromScriptValue(nil)
		assert.Nil(t, result)
	})
}

func TestValidateStringArg(t *testing.T) {
	args := []ScriptValue{
		NewStringValue("hello"),
		NewNumberValue(42),
		NewNilValue(),
	}

	t.Run("valid string arg", func(t *testing.T) {
		result, err := ValidateStringArg(args, 0, "first")
		assert.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("non-string arg", func(t *testing.T) {
		_, err := ValidateStringArg(args, 1, "second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "second must be string")
	})

	t.Run("missing arg", func(t *testing.T) {
		_, err := ValidateStringArg(args, 5, "missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument required")
	})

	t.Run("nil arg", func(t *testing.T) {
		_, err := ValidateStringArg(args, 2, "nil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil must be string")
	})
}

func TestValidateNumberArg(t *testing.T) {
	args := []ScriptValue{
		NewNumberValue(42),
		NewStringValue("hello"),
		NewNilValue(),
	}

	t.Run("valid number arg", func(t *testing.T) {
		result, err := ValidateNumberArg(args, 0, "first")
		assert.NoError(t, err)
		assert.Equal(t, float64(42), result)
	})

	t.Run("non-number arg", func(t *testing.T) {
		_, err := ValidateNumberArg(args, 1, "second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "second must be number")
	})

	t.Run("missing arg", func(t *testing.T) {
		_, err := ValidateNumberArg(args, 5, "missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument required")
	})
}

func TestValidateBoolArg(t *testing.T) {
	args := []ScriptValue{
		NewBoolValue(true),
		NewStringValue("hello"),
		NewNilValue(),
	}

	t.Run("valid bool arg", func(t *testing.T) {
		result, err := ValidateBoolArg(args, 0, "first")
		assert.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("non-bool arg", func(t *testing.T) {
		_, err := ValidateBoolArg(args, 1, "second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "second must be boolean")
	})

	t.Run("missing arg", func(t *testing.T) {
		_, err := ValidateBoolArg(args, 5, "missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument required")
	})
}

func TestValidateObjectArg(t *testing.T) {
	objectValue := NewObjectValue(map[string]ScriptValue{
		"name": NewStringValue("test"),
		"age":  NewNumberValue(25),
	})

	args := []ScriptValue{
		objectValue,
		NewStringValue("hello"),
		NewNilValue(),
	}

	t.Run("valid object arg", func(t *testing.T) {
		result, err := ValidateObjectArg(args, 0, "first")
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"name": "test",
			"age":  float64(25),
		}, result)
	})

	t.Run("non-object arg", func(t *testing.T) {
		_, err := ValidateObjectArg(args, 1, "second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "second must be object")
	})

	t.Run("missing arg", func(t *testing.T) {
		_, err := ValidateObjectArg(args, 5, "missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument required")
	})
}

func TestValidateArrayArg(t *testing.T) {
	arrayValue := NewArrayValue([]ScriptValue{
		NewStringValue("hello"),
		NewNumberValue(42),
	})

	args := []ScriptValue{
		arrayValue,
		NewStringValue("hello"),
		NewNilValue(),
	}

	t.Run("valid array arg", func(t *testing.T) {
		result, err := ValidateArrayArg(args, 0, "first")
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{"hello", float64(42)}, result)
	})

	t.Run("non-array arg", func(t *testing.T) {
		_, err := ValidateArrayArg(args, 1, "second")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "second must be array")
	})

	t.Run("missing arg", func(t *testing.T) {
		_, err := ValidateArrayArg(args, 5, "missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument required")
	})
}

func TestValidateOptionalArgs(t *testing.T) {
	args := []ScriptValue{
		NewStringValue("hello"),
		NewNumberValue(42),
		NewBoolValue(true),
	}

	t.Run("optional string - present", func(t *testing.T) {
		result := ValidateOptionalStringArg(args, 0, "default")
		assert.Equal(t, "hello", result)
	})

	t.Run("optional string - missing", func(t *testing.T) {
		result := ValidateOptionalStringArg(args, 5, "default")
		assert.Equal(t, "default", result)
	})

	t.Run("optional string - wrong type", func(t *testing.T) {
		result := ValidateOptionalStringArg(args, 1, "default")
		assert.Equal(t, "default", result)
	})

	t.Run("optional number - present", func(t *testing.T) {
		result := ValidateOptionalNumberArg(args, 1, 99)
		assert.Equal(t, float64(42), result)
	})

	t.Run("optional number - missing", func(t *testing.T) {
		result := ValidateOptionalNumberArg(args, 5, 99)
		assert.Equal(t, float64(99), result)
	})

	t.Run("optional bool - present", func(t *testing.T) {
		result := ValidateOptionalBoolArg(args, 2, false)
		assert.Equal(t, true, result)
	})

	t.Run("optional bool - missing", func(t *testing.T) {
		result := ValidateOptionalBoolArg(args, 5, false)
		assert.Equal(t, false, result)
	})
}

func TestConvertScriptValueMap(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		result := ConvertScriptValueMap(nil)
		assert.Nil(t, result)
	})

	t.Run("simple map", func(t *testing.T) {
		input := map[string]ScriptValue{
			"name": NewStringValue("test"),
			"age":  NewNumberValue(25),
		}
		result := ConvertScriptValueMap(input)

		expected := map[string]interface{}{
			"name": "test",
			"age":  float64(25),
		}
		assert.Equal(t, expected, result)
	})
}

func TestConvertScriptValueSlice(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		result := ConvertScriptValueSlice(nil)
		assert.Nil(t, result)
	})

	t.Run("simple slice", func(t *testing.T) {
		input := []ScriptValue{
			NewStringValue("hello"),
			NewNumberValue(42),
			NewBoolValue(true),
		}
		result := ConvertScriptValueSlice(input)

		expected := []interface{}{"hello", float64(42), true}
		assert.Equal(t, expected, result)
	})
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkConvertToScriptValue(b *testing.B) {
	testData := map[string]interface{}{
		"name":   "test",
		"age":    25,
		"active": true,
		"tags":   []interface{}{"tag1", "tag2", "tag3"},
		"metadata": map[string]interface{}{
			"version": "1.0",
			"count":   100,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertToScriptValue(testData)
	}
}

func BenchmarkValidateStringArg(b *testing.B) {
	args := []ScriptValue{
		NewStringValue("test"),
		NewNumberValue(42),
		NewBoolValue(true),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateStringArg(args, 0, "test")
	}
}
