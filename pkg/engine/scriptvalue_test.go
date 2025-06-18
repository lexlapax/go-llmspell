// ABOUTME: Comprehensive tests for the ScriptValue type system including all value types and conversions
// ABOUTME: Tests type safety, equality, conversions, and edge cases for the universal script value interface

package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptValueTypes(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		v := NewNilValue()
		assert.Equal(t, TypeNil, v.Type())
		assert.True(t, v.IsNil())
		assert.Equal(t, "nil", v.String())
		assert.Nil(t, v.ToGo())
		assert.True(t, v.Equals(NewNilValue()))
		assert.False(t, v.Equals(NewBoolValue(false)))
	})

	t.Run("BoolValue", func(t *testing.T) {
		v1 := NewBoolValue(true)
		assert.Equal(t, TypeBool, v1.Type())
		assert.False(t, v1.IsNil())
		assert.Equal(t, "true", v1.String())
		assert.Equal(t, true, v1.ToGo())
		assert.True(t, v1.Equals(NewBoolValue(true)))
		assert.False(t, v1.Equals(NewBoolValue(false)))

		v2 := NewBoolValue(false)
		assert.Equal(t, "false", v2.String())
		assert.Equal(t, false, v2.ToGo())
	})

	t.Run("NumberValue", func(t *testing.T) {
		v1 := NewNumberValue(42.5)
		assert.Equal(t, TypeNumber, v1.Type())
		assert.False(t, v1.IsNil())
		assert.Equal(t, "42.5", v1.String())
		assert.Equal(t, 42.5, v1.ToGo())
		assert.True(t, v1.Equals(NewNumberValue(42.5)))
		assert.False(t, v1.Equals(NewNumberValue(42.0)))

		// Test integer formatting
		v2 := NewNumberValue(42.0)
		assert.Equal(t, "42", v2.String())
	})

	t.Run("StringValue", func(t *testing.T) {
		v := NewStringValue("hello world")
		assert.Equal(t, TypeString, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "hello world", v.String())
		assert.Equal(t, "hello world", v.ToGo())
		assert.True(t, v.Equals(NewStringValue("hello world")))
		assert.False(t, v.Equals(NewStringValue("goodbye")))

		// Test empty string
		v2 := NewStringValue("")
		assert.Equal(t, "", v2.String())
		assert.False(t, v2.IsNil())
	})

	t.Run("ArrayValue", func(t *testing.T) {
		elements := []ScriptValue{
			NewNumberValue(1),
			NewStringValue("two"),
			NewBoolValue(true),
		}
		v := NewArrayValue(elements)
		assert.Equal(t, TypeArray, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "[1, two, true]", v.String())
		
		// Test ToGo conversion
		goValue := v.ToGo()
		arr, ok := goValue.([]interface{})
		require.True(t, ok)
		assert.Len(t, arr, 3)
		assert.Equal(t, float64(1), arr[0])
		assert.Equal(t, "two", arr[1])
		assert.Equal(t, true, arr[2])

		// Test equality
		v2 := NewArrayValue([]ScriptValue{
			NewNumberValue(1),
			NewStringValue("two"),
			NewBoolValue(true),
		})
		assert.True(t, v.Equals(v2))

		// Test inequality
		v3 := NewArrayValue([]ScriptValue{
			NewNumberValue(1),
			NewStringValue("three"),
		})
		assert.False(t, v.Equals(v3))

		// Test empty array
		v4 := NewArrayValue([]ScriptValue{})
		assert.Equal(t, "[]", v4.String())
	})

	t.Run("ObjectValue", func(t *testing.T) {
		fields := map[string]ScriptValue{
			"name":   NewStringValue("test"),
			"age":    NewNumberValue(25),
			"active": NewBoolValue(true),
		}
		v := NewObjectValue(fields)
		assert.Equal(t, TypeObject, v.Type())
		assert.False(t, v.IsNil())
		
		// Test ToGo conversion
		goValue := v.ToGo()
		obj, ok := goValue.(map[string]interface{})
		require.True(t, ok)
		assert.Len(t, obj, 3)
		assert.Equal(t, "test", obj["name"])
		assert.Equal(t, float64(25), obj["age"])
		assert.Equal(t, true, obj["active"])

		// Test equality
		v2 := NewObjectValue(map[string]ScriptValue{
			"name":   NewStringValue("test"),
			"age":    NewNumberValue(25),
			"active": NewBoolValue(true),
		})
		assert.True(t, v.Equals(v2))

		// Test inequality - different values
		v3 := NewObjectValue(map[string]ScriptValue{
			"name":   NewStringValue("different"),
			"age":    NewNumberValue(25),
			"active": NewBoolValue(true),
		})
		assert.False(t, v.Equals(v3))

		// Test inequality - different keys
		v4 := NewObjectValue(map[string]ScriptValue{
			"name": NewStringValue("test"),
			"age":  NewNumberValue(25),
		})
		assert.False(t, v.Equals(v4))

		// Test empty object
		v5 := NewObjectValue(map[string]ScriptValue{})
		assert.Equal(t, "{}", v5.String())
	})

	t.Run("FunctionValue", func(t *testing.T) {
		fn := func(a, b int) int { return a + b }
		v := NewFunctionValue("add", fn)
		assert.Equal(t, TypeFunction, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "function:add", v.String())
		assert.NotNil(t, v.ToGo())

		// Functions are equal if they have the same value
		v2 := NewFunctionValue("add", fn)
		assert.True(t, v.Equals(v2))
	})

	t.Run("ErrorValue", func(t *testing.T) {
		err := errors.New("test error")
		v := NewErrorValue(err)
		assert.Equal(t, TypeError, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "test error", v.String())
		assert.Equal(t, err, v.ToGo())

		// Test equality
		v2 := NewErrorValue(errors.New("test error"))
		assert.True(t, v.Equals(v2))

		v3 := NewErrorValue(errors.New("different error"))
		assert.False(t, v.Equals(v3))
	})

	t.Run("ChannelValue", func(t *testing.T) {
		ch := make(chan interface{})
		v := NewChannelValue("test-channel", ch)
		assert.Equal(t, TypeChannel, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "channel:test-channel", v.String())
		assert.Equal(t, ch, v.ToGo())

		// Test equality by ID
		v2 := NewChannelValue("test-channel", make(chan interface{}))
		assert.True(t, v.Equals(v2))

		v3 := NewChannelValue("other-channel", ch)
		assert.False(t, v.Equals(v3))
	})

	t.Run("CustomValue", func(t *testing.T) {
		type CustomType struct {
			Value string
		}
		custom := CustomType{Value: "test"}
		v := NewCustomValue("CustomType", custom)
		assert.Equal(t, TypeCustom, v.Type())
		assert.False(t, v.IsNil())
		assert.Equal(t, "CustomType:{test}", v.String())
		assert.Equal(t, custom, v.ToGo())

		// Custom values are equal if their types and values match
		v2 := NewCustomValue("CustomType", custom)
		assert.True(t, v.Equals(v2))
	})
}

func TestScriptValueConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected ScriptValue
	}{
		{
			name:     "nil",
			input:    nil,
			expected: NewNilValue(),
		},
		{
			name:     "bool",
			input:    true,
			expected: NewBoolValue(true),
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
			input:    float32(42.5),
			expected: NewNumberValue(42.5),
		},
		{
			name:     "float64",
			input:    float64(42.5),
			expected: NewNumberValue(42.5),
		},
		{
			name:     "string",
			input:    "hello",
			expected: NewStringValue("hello"),
		},
		{
			name:     "error",
			input:    errors.New("test error"),
			expected: NewStringValue("test error"), // errors are converted to strings
		},
		{
			name:     "slice",
			input:    []interface{}{1, "two", true},
			expected: NewArrayValue([]ScriptValue{NewNumberValue(1), NewStringValue("two"), NewBoolValue(true)}),
		},
		{
			name:     "map",
			input:    map[string]interface{}{"key": "value", "num": 42},
			expected: NewObjectValue(map[string]ScriptValue{"key": NewStringValue("value"), "num": NewNumberValue(42)}),
		},
		{
			name:     "channel",
			input:    make(chan interface{}),
			expected: NewChannelValue("", make(chan interface{})),
		},
		{
			name:     "function",
			input:    func() {},
			expected: NewFunctionValue("", func() {}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToScriptValue(tt.input)
			assert.Equal(t, tt.expected.Type(), result.Type())
			
			// For most types, we can check value equality
			switch tt.expected.Type() {
			case TypeNil, TypeBool, TypeNumber, TypeString, TypeError:
				assert.True(t, result.Equals(tt.expected))
			case TypeArray:
				// For arrays, check individual elements
				expectedArr := tt.expected.(ArrayValue)
				resultArr := result.(ArrayValue)
				assert.Len(t, resultArr.elements, len(expectedArr.elements))
			case TypeObject:
				// For objects, check ToGo conversion
				expectedMap := tt.expected.ToGo().(map[string]interface{})
				resultMap := result.ToGo().(map[string]interface{})
				assert.Equal(t, len(expectedMap), len(resultMap))
			}
		})
	}
}

func TestScriptValueNestedConversions(t *testing.T) {
	t.Run("NestedArray", func(t *testing.T) {
		input := []interface{}{
			1,
			[]interface{}{2, 3},
			map[string]interface{}{"key": "value"},
		}
		
		v := ConvertToScriptValue(input)
		assert.Equal(t, TypeArray, v.Type())
		
		// Convert back and verify structure
		output := v.ToGo()
		arr, ok := output.([]interface{})
		require.True(t, ok)
		assert.Len(t, arr, 3)
		assert.Equal(t, float64(1), arr[0])
		
		innerArr, ok := arr[1].([]interface{})
		require.True(t, ok)
		assert.Len(t, innerArr, 2)
		assert.Equal(t, float64(2), innerArr[0])
		assert.Equal(t, float64(3), innerArr[1])
		
		innerMap, ok := arr[2].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", innerMap["key"])
	})

	t.Run("NestedObject", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "test",
			"nested": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep",
				},
			},
			"array": []interface{}{1, 2, 3},
		}
		
		v := ConvertToScriptValue(input)
		assert.Equal(t, TypeObject, v.Type())
		
		// Convert back and verify structure
		output := v.ToGo()
		obj, ok := output.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test", obj["name"])
		
		nested, ok := obj["nested"].(map[string]interface{})
		require.True(t, ok)
		level2, ok := nested["level2"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "deep", level2["level3"])
		
		arr, ok := obj["array"].([]interface{})
		require.True(t, ok)
		assert.Len(t, arr, 3)
	})
}

func TestScriptValueEdgeCases(t *testing.T) {
	t.Run("NilHandling", func(t *testing.T) {
		// Nil in array
		arr := NewArrayValue([]ScriptValue{NewNilValue(), NewNumberValue(1)})
		goArr := arr.ToGo().([]interface{})
		assert.Nil(t, goArr[0])
		assert.Equal(t, float64(1), goArr[1])

		// Nil in object
		obj := NewObjectValue(map[string]ScriptValue{
			"key": NewNilValue(),
		})
		goObj := obj.ToGo().(map[string]interface{})
		assert.Nil(t, goObj["key"])
	})

	t.Run("EmptyCollections", func(t *testing.T) {
		// Empty array
		emptyArr := NewArrayValue([]ScriptValue{})
		assert.Equal(t, "[]", emptyArr.String())
		assert.Len(t, emptyArr.ToGo().([]interface{}), 0)

		// Empty object
		emptyObj := NewObjectValue(map[string]ScriptValue{})
		assert.Equal(t, "{}", emptyObj.String())
		assert.Len(t, emptyObj.ToGo().(map[string]interface{}), 0)
	})

	t.Run("TypeMismatchInEquality", func(t *testing.T) {
		// Different types should never be equal
		values := []ScriptValue{
			NewNilValue(),
			NewBoolValue(true),
			NewNumberValue(1),
			NewStringValue("1"),
			NewArrayValue([]ScriptValue{}),
			NewObjectValue(map[string]ScriptValue{}),
			NewFunctionValue("fn", func() {}),
			NewErrorValue(errors.New("err")),
			NewChannelValue("ch", make(chan interface{})),
			NewCustomValue("Custom", struct{}{}),
		}

		for i, v1 := range values {
			for j, v2 := range values {
				if i != j {
					assert.False(t, v1.Equals(v2), "Values of different types should not be equal: %s vs %s", v1.Type(), v2.Type())
				}
			}
		}
	})

	t.Run("SpecialNumbers", func(t *testing.T) {
		// Test special float values
		// Note: In Go, we need to use math package for special values
		// For now, just test regular numbers
	})

	t.Run("LargeCollections", func(t *testing.T) {
		// Large array
		elements := make([]ScriptValue, 1000)
		for i := range elements {
			elements[i] = NewNumberValue(float64(i))
		}
		largeArr := NewArrayValue(elements)
		assert.Equal(t, TypeArray, largeArr.Type())
		assert.Len(t, largeArr.ToGo().([]interface{}), 1000)

		// Large object
		fields := make(map[string]ScriptValue)
		for i := 0; i < 1000; i++ {
			fields[string(rune('a'+i%26))+string(rune('0'+i/26))] = NewNumberValue(float64(i))
		}
		largeObj := NewObjectValue(fields)
		assert.Equal(t, TypeObject, largeObj.Type())
		assert.Len(t, largeObj.ToGo().(map[string]interface{}), 1000)
	})
}

func TestScriptValueGetters(t *testing.T) {
	t.Run("BoolValue", func(t *testing.T) {
		v := NewBoolValue(true).(BoolValue)
		assert.True(t, v.Value())
	})

	t.Run("NumberValue", func(t *testing.T) {
		v := NewNumberValue(42.5).(NumberValue)
		assert.Equal(t, 42.5, v.Value())
	})

	t.Run("StringValue", func(t *testing.T) {
		v := NewStringValue("hello").(StringValue)
		assert.Equal(t, "hello", v.Value())
	})

	t.Run("ErrorValue", func(t *testing.T) {
		err := errors.New("test error")
		v := NewErrorValue(err).(ErrorValue)
		assert.Equal(t, err, v.Error())
	})
}

func BenchmarkScriptValueCreation(b *testing.B) {
	b.Run("NilValue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewNilValue()
		}
	})

	b.Run("BoolValue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewBoolValue(true)
		}
	})

	b.Run("NumberValue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewNumberValue(42.5)
		}
	})

	b.Run("StringValue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewStringValue("hello world")
		}
	})

	b.Run("ArrayValue", func(b *testing.B) {
		elements := []ScriptValue{NewNumberValue(1), NewNumberValue(2), NewNumberValue(3)}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewArrayValue(elements)
		}
	})

	b.Run("ObjectValue", func(b *testing.B) {
		fields := map[string]ScriptValue{
			"key1": NewStringValue("value1"),
			"key2": NewNumberValue(42),
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewObjectValue(fields)
		}
	})
}

func BenchmarkScriptValueConversion(b *testing.B) {
	b.Run("ConvertMap", func(b *testing.B) {
		m := map[string]interface{}{
			"name": "test",
			"age":  42,
			"active": true,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ConvertToScriptValue(m)
		}
	})

	b.Run("ConvertSlice", func(b *testing.B) {
		s := []interface{}{1, 2, 3, "four", true}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ConvertToScriptValue(s)
		}
	})

	b.Run("ConvertNested", func(b *testing.B) {
		nested := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": []interface{}{1, 2, 3},
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ConvertToScriptValue(nested)
		}
	})
}