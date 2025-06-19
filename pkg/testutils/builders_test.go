// ABOUTME: Tests for ScriptValue builders to ensure correct value creation
// ABOUTME: Validates fluent builders, quick creators, and factory methods

package testutils

import (
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptValueBuilder(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		value := NewScriptValueBuilder().
			String("hello").
			BuildSingle()

		assert.Equal(t, engine.TypeString, value.Type())
		assert.Equal(t, "hello", value.(engine.StringValue).Value())
	})

	t.Run("multiple values", func(t *testing.T) {
		values := NewScriptValueBuilder().
			String("hello").
			Number(42.5).
			Bool(true).
			Nil().
			Build()

		require.Len(t, values, 4)
		assert.Equal(t, engine.TypeString, values[0].Type())
		assert.Equal(t, engine.TypeNumber, values[1].Type())
		assert.Equal(t, engine.TypeBool, values[2].Type())
		assert.Equal(t, engine.TypeNil, values[3].Type())
	})

	t.Run("object building", func(t *testing.T) {
		value := NewScriptValueBuilder().
			Object(map[string]interface{}{
				"name":  "test",
				"value": 123,
				"flag":  true,
			}).
			BuildSingle()

		assert.Equal(t, engine.TypeObject, value.Type())
		fields := value.(engine.ObjectValue).Fields()
		assert.Len(t, fields, 3)
		assert.Equal(t, "test", fields["name"].(engine.StringValue).Value())
		assert.Equal(t, 123.0, fields["value"].(engine.NumberValue).Value())
		assert.Equal(t, true, fields["flag"].(engine.BoolValue).Value())
	})

	t.Run("array building", func(t *testing.T) {
		value := NewScriptValueBuilder().
			Array("one", 2, true, nil).
			BuildSingle()

		assert.Equal(t, engine.TypeArray, value.Type())
		elements := value.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 4)
		assert.Equal(t, "one", elements[0].(engine.StringValue).Value())
		assert.Equal(t, 2.0, elements[1].(engine.NumberValue).Value())
		assert.Equal(t, true, elements[2].(engine.BoolValue).Value())
		assert.True(t, elements[3].IsNil())
	})

	t.Run("error values", func(t *testing.T) {
		err := errors.New("test error")
		value1 := NewScriptValueBuilder().Error(err).BuildSingle()
		value2 := NewScriptValueBuilder().ErrorString("test error string").BuildSingle()

		assert.Equal(t, engine.TypeError, value1.Type())
		assert.Equal(t, engine.TypeError, value2.Type())
		assert.Equal(t, "test error", value1.(engine.ErrorValue).Error().Error())
		assert.Equal(t, "test error string", value2.(engine.ErrorValue).Error().Error())
	})

	t.Run("custom values", func(t *testing.T) {
		customData := struct{ Name string }{Name: "custom"}
		value := NewScriptValueBuilder().
			Custom("MyType", customData).
			BuildSingle()

		assert.Equal(t, engine.TypeCustom, value.Type())
		cv := value.(engine.CustomValue)
		assert.Equal(t, "MyType", cv.TypeName())
		assert.Equal(t, customData, cv.Value())
	})

	t.Run("empty builder returns nil", func(t *testing.T) {
		value := NewScriptValueBuilder().BuildSingle()
		assert.True(t, value.IsNil())
	})
}

func TestQuickCreators(t *testing.T) {
	t.Run("StringValue", func(t *testing.T) {
		v := StringValue("test")
		assert.Equal(t, engine.TypeString, v.Type())
		assert.Equal(t, "test", v.(engine.StringValue).Value())
	})

	t.Run("NumberValue", func(t *testing.T) {
		v := NumberValue(42.5)
		assert.Equal(t, engine.TypeNumber, v.Type())
		assert.Equal(t, 42.5, v.(engine.NumberValue).Value())
	})

	t.Run("IntValue", func(t *testing.T) {
		v := IntValue(42)
		assert.Equal(t, engine.TypeNumber, v.Type())
		assert.Equal(t, 42.0, v.(engine.NumberValue).Value())
	})

	t.Run("BoolValue", func(t *testing.T) {
		v := BoolValue(true)
		assert.Equal(t, engine.TypeBool, v.Type())
		assert.Equal(t, true, v.(engine.BoolValue).Value())
	})

	t.Run("NilValue", func(t *testing.T) {
		v := NilValue()
		assert.Equal(t, engine.TypeNil, v.Type())
		assert.True(t, v.IsNil())
	})

	t.Run("ErrorValue", func(t *testing.T) {
		err := errors.New("test error")
		v := ErrorValue(err)
		assert.Equal(t, engine.TypeError, v.Type())
		assert.Equal(t, err, v.(engine.ErrorValue).Error())
	})

	t.Run("ErrorStringValue", func(t *testing.T) {
		v := ErrorStringValue("error message")
		assert.Equal(t, engine.TypeError, v.Type())
		assert.Equal(t, "error message", v.(engine.ErrorValue).Error().Error())
	})
}

func TestArrayCreators(t *testing.T) {
	t.Run("ArrayFromStrings", func(t *testing.T) {
		arr := ArrayFromStrings("one", "two", "three")
		assert.Equal(t, engine.TypeArray, arr.Type())

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 3)
		assert.Equal(t, "one", elements[0].(engine.StringValue).Value())
		assert.Equal(t, "two", elements[1].(engine.StringValue).Value())
		assert.Equal(t, "three", elements[2].(engine.StringValue).Value())
	})

	t.Run("ArrayFromNumbers", func(t *testing.T) {
		arr := ArrayFromNumbers(1.5, 2.5, 3.5)
		assert.Equal(t, engine.TypeArray, arr.Type())

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 3)
		assert.Equal(t, 1.5, elements[0].(engine.NumberValue).Value())
		assert.Equal(t, 2.5, elements[1].(engine.NumberValue).Value())
		assert.Equal(t, 3.5, elements[2].(engine.NumberValue).Value())
	})

	t.Run("ArrayFromInts", func(t *testing.T) {
		arr := ArrayFromInts(1, 2, 3)
		assert.Equal(t, engine.TypeArray, arr.Type())

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 3)
		assert.Equal(t, 1.0, elements[0].(engine.NumberValue).Value())
		assert.Equal(t, 2.0, elements[1].(engine.NumberValue).Value())
		assert.Equal(t, 3.0, elements[2].(engine.NumberValue).Value())
	})
}

func TestInterfaceToScriptValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected func(t *testing.T, v engine.ScriptValue)
	}{
		{
			name:  "nil",
			input: nil,
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.True(t, v.IsNil())
			},
		},
		{
			name:  "bool",
			input: true,
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeBool, v.Type())
				assert.Equal(t, true, v.(engine.BoolValue).Value())
			},
		},
		{
			name:  "string",
			input: "hello",
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeString, v.Type())
				assert.Equal(t, "hello", v.(engine.StringValue).Value())
			},
		},
		{
			name:  "int",
			input: 42,
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeNumber, v.Type())
				assert.Equal(t, 42.0, v.(engine.NumberValue).Value())
			},
		},
		{
			name:  "float64",
			input: 42.5,
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeNumber, v.Type())
				assert.Equal(t, 42.5, v.(engine.NumberValue).Value())
			},
		},
		{
			name:  "slice",
			input: []interface{}{"a", 1, true},
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeArray, v.Type())
				elements := v.(engine.ArrayValue).Elements()
				assert.Len(t, elements, 3)
			},
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value", "num": 123},
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeObject, v.Type())
				fields := v.(engine.ObjectValue).Fields()
				assert.Len(t, fields, 2)
			},
		},
		{
			name:  "error",
			input: errors.New("test error"),
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeError, v.Type())
				assert.Equal(t, "test error", v.(engine.ErrorValue).Error().Error())
			},
		},
		{
			name:  "custom type",
			input: struct{ Name string }{Name: "test"},
			expected: func(t *testing.T, v engine.ScriptValue) {
				assert.Equal(t, engine.TypeCustom, v.Type())
				cv := v.(engine.CustomValue)
				assert.Equal(t, "struct { Name string }", cv.TypeName())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterfaceToScriptValue(tt.input)
			tt.expected(t, result)
		})
	}
}

func TestFactoryMethods(t *testing.T) {
	t.Run("CreateTestObject", func(t *testing.T) {
		obj := CreateTestObject("123", "test object", 99.9)
		assert.Equal(t, engine.TypeObject, obj.Type())

		fields := obj.(engine.ObjectValue).Fields()
		assert.Equal(t, "123", fields["id"].(engine.StringValue).Value())
		assert.Equal(t, "test object", fields["name"].(engine.StringValue).Value())
		assert.Equal(t, 99.9, fields["value"].(engine.NumberValue).Value())
		assert.Equal(t, "test", fields["type"].(engine.StringValue).Value())
	})

	t.Run("CreateTestArray", func(t *testing.T) {
		arr := CreateTestArray(6)
		assert.Equal(t, engine.TypeArray, arr.Type())

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 6)

		// Check pattern: string, number, bool, string, number, bool
		assert.Equal(t, engine.TypeString, elements[0].Type())
		assert.Equal(t, engine.TypeNumber, elements[1].Type())
		assert.Equal(t, engine.TypeBool, elements[2].Type())
		assert.Equal(t, engine.TypeString, elements[3].Type())
		assert.Equal(t, engine.TypeNumber, elements[4].Type())
		assert.Equal(t, engine.TypeBool, elements[5].Type())
	})

	t.Run("CreateNestedObject", func(t *testing.T) {
		obj := CreateNestedObject()
		assert.Equal(t, engine.TypeObject, obj.Type())

		fields := obj.(engine.ObjectValue).Fields()
		assert.Contains(t, fields, "level1")
		assert.Contains(t, fields, "metadata")

		// Check deep nesting
		level1 := fields["level1"].(engine.ObjectValue).Fields()
		level2 := level1["level2"].(engine.ObjectValue).Fields()
		level3 := level2["level3"].(engine.ObjectValue).Fields()

		assert.Equal(t, "deep", level3["value"].(engine.StringValue).Value())
		assert.Equal(t, engine.TypeArray, level3["array"].Type())
	})

	t.Run("CreateComplexArray", func(t *testing.T) {
		arr := CreateComplexArray()
		assert.Equal(t, engine.TypeArray, arr.Type())

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 7)

		// Verify types
		assert.Equal(t, engine.TypeString, elements[0].Type())
		assert.Equal(t, engine.TypeNumber, elements[1].Type())
		assert.Equal(t, engine.TypeBool, elements[2].Type())
		assert.True(t, elements[3].IsNil())
		assert.Equal(t, engine.TypeObject, elements[4].Type())
		assert.Equal(t, engine.TypeArray, elements[5].Type())
		assert.Equal(t, engine.TypeObject, elements[6].Type())
	})
}

func TestObjectBuilder(t *testing.T) {
	t.Run("basic object", func(t *testing.T) {
		obj := NewObjectBuilder().
			SetString("name", "test").
			SetNumber("value", 42.5).
			SetInt("count", 10).
			SetBool("enabled", true).
			SetNil("empty").
			Build()

		assert.Equal(t, engine.TypeObject, obj.Type())
		fields := obj.(engine.ObjectValue).Fields()

		assert.Equal(t, "test", fields["name"].(engine.StringValue).Value())
		assert.Equal(t, 42.5, fields["value"].(engine.NumberValue).Value())
		assert.Equal(t, 10.0, fields["count"].(engine.NumberValue).Value())
		assert.Equal(t, true, fields["enabled"].(engine.BoolValue).Value())
		assert.True(t, fields["empty"].IsNil())
	})

	t.Run("nested object", func(t *testing.T) {
		nested := NewObjectBuilder().
			SetString("inner", "value").
			SetInt("level", 2)

		obj := NewObjectBuilder().
			SetString("outer", "test").
			SetObject("nested", nested).
			Build()

		fields := obj.(engine.ObjectValue).Fields()
		nestedObj := fields["nested"].(engine.ObjectValue).Fields()

		assert.Equal(t, "value", nestedObj["inner"].(engine.StringValue).Value())
		assert.Equal(t, 2.0, nestedObj["level"].(engine.NumberValue).Value())
	})

	t.Run("object with array", func(t *testing.T) {
		obj := NewObjectBuilder().
			SetArray("items", "one", "two", "three").
			SetArray("numbers", 1, 2, 3).
			Build()

		fields := obj.(engine.ObjectValue).Fields()

		items := fields["items"].(engine.ArrayValue).Elements()
		assert.Len(t, items, 3)
		assert.Equal(t, "one", items[0].(engine.StringValue).Value())

		numbers := fields["numbers"].(engine.ArrayValue).Elements()
		assert.Len(t, numbers, 3)
		assert.Equal(t, 1.0, numbers[0].(engine.NumberValue).Value())
	})
}

func TestArrayBuilder(t *testing.T) {
	t.Run("basic array", func(t *testing.T) {
		arr := NewArrayBuilder().
			AddString("hello").
			AddNumber(42.5).
			AddInt(10).
			AddBool(true).
			AddNil().
			Build()

		assert.Equal(t, engine.TypeArray, arr.Type())
		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 5)

		assert.Equal(t, "hello", elements[0].(engine.StringValue).Value())
		assert.Equal(t, 42.5, elements[1].(engine.NumberValue).Value())
		assert.Equal(t, 10.0, elements[2].(engine.NumberValue).Value())
		assert.Equal(t, true, elements[3].(engine.BoolValue).Value())
		assert.True(t, elements[4].IsNil())
	})

	t.Run("array with nested structures", func(t *testing.T) {
		nestedObj := NewObjectBuilder().
			SetString("type", "nested")

		nestedArr := NewArrayBuilder().
			AddInt(1).
			AddInt(2).
			AddInt(3)

		arr := NewArrayBuilder().
			AddString("start").
			AddObject(nestedObj).
			AddArray(nestedArr).
			AddString("end").
			Build()

		elements := arr.(engine.ArrayValue).Elements()
		assert.Len(t, elements, 4)

		assert.Equal(t, engine.TypeString, elements[0].Type())
		assert.Equal(t, engine.TypeObject, elements[1].Type())
		assert.Equal(t, engine.TypeArray, elements[2].Type())
		assert.Equal(t, engine.TypeString, elements[3].Type())

		// Check nested array
		nestedElements := elements[2].(engine.ArrayValue).Elements()
		assert.Len(t, nestedElements, 3)
		assert.Equal(t, 1.0, nestedElements[0].(engine.NumberValue).Value())
	})
}
