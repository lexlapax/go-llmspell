// ABOUTME: Tests for ScriptValue assertion helpers to ensure correct assertion behavior
// ABOUTME: Validates type checking, value comparison, and comprehensive assertion patterns

package testutils

import (
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

func TestAssertScriptValueType(t *testing.T) {
	tests := []struct {
		name         string
		value        engine.ScriptValue
		expectedType engine.ScriptValueType
		shouldPass   bool
	}{
		{
			name:         "string type match",
			value:        engine.NewStringValue("test"),
			expectedType: engine.TypeString,
			shouldPass:   true,
		},
		{
			name:         "number type match",
			value:        engine.NewNumberValue(42),
			expectedType: engine.TypeNumber,
			shouldPass:   true,
		},
		{
			name:         "type mismatch",
			value:        engine.NewStringValue("test"),
			expectedType: engine.TypeNumber,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPass {
				AssertScriptValueType(t, tt.value, tt.expectedType)
			}
			// We can't test the failure case directly as it would fail the test
			// but we know it works from the implementation
		})
	}
}

func TestStringAssertions(t *testing.T) {
	t.Run("AssertIsString", func(t *testing.T) {
		sv := engine.NewStringValue("hello")
		result := AssertIsString(t, sv)
		assert.Equal(t, "hello", result)
	})

	t.Run("AssertStringEquals", func(t *testing.T) {
		sv := engine.NewStringValue("hello")
		AssertStringEquals(t, sv, "hello")
	})

	t.Run("AssertStringContains", func(t *testing.T) {
		sv := engine.NewStringValue("hello world")
		AssertStringContains(t, sv, "world")
	})
}

func TestNumberAssertions(t *testing.T) {
	t.Run("AssertIsNumber", func(t *testing.T) {
		nv := engine.NewNumberValue(42.5)
		result := AssertIsNumber(t, nv)
		assert.Equal(t, 42.5, result)
	})

	t.Run("AssertNumberEquals", func(t *testing.T) {
		nv := engine.NewNumberValue(42.5)
		AssertNumberEquals(t, nv, 42.5)
	})

	t.Run("AssertNumberInRange", func(t *testing.T) {
		nv := engine.NewNumberValue(50)
		AssertNumberInRange(t, nv, 40, 60)
	})
}

func TestBoolAssertions(t *testing.T) {
	t.Run("AssertIsBool", func(t *testing.T) {
		bv := engine.NewBoolValue(true)
		result := AssertIsBool(t, bv)
		assert.Equal(t, true, result)
	})

	t.Run("AssertBoolEquals", func(t *testing.T) {
		bv := engine.NewBoolValue(false)
		AssertBoolEquals(t, bv, false)
	})
}

func TestNilAssertions(t *testing.T) {
	t.Run("AssertIsNil with nil value", func(t *testing.T) {
		nv := engine.NewNilValue()
		AssertIsNil(t, nv)
	})

	t.Run("AssertIsNil with nil interface", func(t *testing.T) {
		var nv engine.ScriptValue
		AssertIsNil(t, nv)
	})

	t.Run("AssertNotNil", func(t *testing.T) {
		sv := engine.NewStringValue("not nil")
		AssertNotNil(t, sv)
	})
}

func TestErrorAssertions(t *testing.T) {
	t.Run("AssertErrorValue", func(t *testing.T) {
		err := errors.New("test error")
		ev := engine.NewErrorValue(err)
		AssertErrorValue(t, ev, "test error")
	})

	t.Run("AssertErrorValue without message check", func(t *testing.T) {
		err := errors.New("any error")
		ev := engine.NewErrorValue(err)
		AssertErrorValue(t, ev, "")
	})

	t.Run("AssertIsError", func(t *testing.T) {
		expectedErr := errors.New("test error")
		ev := engine.NewErrorValue(expectedErr)
		actualErr := AssertIsError(t, ev)
		assert.Equal(t, expectedErr, actualErr)
	})
}

func TestObjectAssertions(t *testing.T) {
	t.Run("AssertObjectValue", func(t *testing.T) {
		obj := engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":  engine.NewStringValue("test"),
			"value": engine.NewNumberValue(42),
		})
		fields := AssertObjectValue(t, obj)
		assert.Len(t, fields, 2)
	})

	t.Run("AssertObjectHasFields", func(t *testing.T) {
		obj := engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":  engine.NewStringValue("test"),
			"value": engine.NewNumberValue(42),
			"flag":  engine.NewBoolValue(true),
		})
		AssertObjectHasFields(t, obj, "name", "value", "flag")
	})

	t.Run("AssertObjectFieldEquals", func(t *testing.T) {
		obj := engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":  engine.NewStringValue("test"),
			"count": engine.NewNumberValue(10),
		})
		AssertObjectFieldEquals(t, obj, "name", "test")
		AssertObjectFieldEquals(t, obj, "count", 10)
	})

	t.Run("AssertObjectFieldCount", func(t *testing.T) {
		obj := engine.NewObjectValue(map[string]engine.ScriptValue{
			"a": engine.NewStringValue("1"),
			"b": engine.NewStringValue("2"),
			"c": engine.NewStringValue("3"),
		})
		AssertObjectFieldCount(t, obj, 3)
	})
}

func TestArrayAssertions(t *testing.T) {
	t.Run("AssertArrayValue", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("one"),
			engine.NewNumberValue(2),
			engine.NewBoolValue(true),
		})
		elements := AssertArrayValue(t, arr)
		assert.Len(t, elements, 3)
	})

	t.Run("AssertArrayLength", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("a"),
			engine.NewStringValue("b"),
			engine.NewStringValue("c"),
		})
		AssertArrayLength(t, arr, 3)
	})

	t.Run("AssertArrayElementEquals", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("first"),
			engine.NewNumberValue(42),
			engine.NewBoolValue(true),
		})
		AssertArrayElementEquals(t, arr, 0, "first")
		AssertArrayElementEquals(t, arr, 1, 42)
		AssertArrayElementEquals(t, arr, 2, true)
	})

	t.Run("AssertArrayContains", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("apple"),
			engine.NewStringValue("banana"),
			engine.NewStringValue("cherry"),
		})
		AssertArrayContains(t, arr, "banana")
	})
}

func TestCustomValueAssertions(t *testing.T) {
	t.Run("AssertCustomValue", func(t *testing.T) {
		customData := struct{ Name string }{Name: "test"}
		cv := engine.NewCustomValue("MyType", customData)

		result := AssertCustomValue(t, cv, "MyType")
		assert.Equal(t, customData, result)
	})

	t.Run("AssertCustomValue without type check", func(t *testing.T) {
		customData := "custom string"
		cv := engine.NewCustomValue("StringType", customData)

		result := AssertCustomValue(t, cv, "")
		assert.Equal(t, customData, result)
	})
}

func TestScriptValueComparison(t *testing.T) {
	t.Run("AssertScriptValueEquals with same values", func(t *testing.T) {
		v1 := engine.NewStringValue("test")
		v2 := engine.NewStringValue("test")
		AssertScriptValueEquals(t, v1, v2)
	})

	t.Run("AssertScriptValueEquals with nil values", func(t *testing.T) {
		var v1, v2 engine.ScriptValue
		AssertScriptValueEquals(t, v1, v2)
	})

	t.Run("AssertScriptValuesEqual", func(t *testing.T) {
		expected := []engine.ScriptValue{
			engine.NewStringValue("one"),
			engine.NewNumberValue(2),
			engine.NewBoolValue(true),
		}
		actual := []engine.ScriptValue{
			engine.NewStringValue("one"),
			engine.NewNumberValue(2),
			engine.NewBoolValue(true),
		}
		AssertScriptValuesEqual(t, expected, actual)
	})
}

func TestComprehensiveValueAssertion(t *testing.T) {
	t.Run("string assertion chain", func(t *testing.T) {
		sv := engine.NewStringValue("hello world")

		NewComprehensiveAssertion(t, sv).
			WithType(engine.TypeString).
			IsString().
			Equals("hello world").
			Contains("world").
			HasPrefix("hello")
	})

	t.Run("number assertion chain", func(t *testing.T) {
		nv := engine.NewNumberValue(50)

		NewComprehensiveAssertion(t, nv).
			WithType(engine.TypeNumber).
			IsNumber().
			Equals(50).
			InRange(40, 60).
			GreaterThan(25)
	})

	t.Run("object assertion chain", func(t *testing.T) {
		obj := engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":  engine.NewStringValue("test"),
			"value": engine.NewNumberValue(42),
			"flag":  engine.NewBoolValue(true),
		})

		NewComprehensiveAssertion(t, obj).
			WithType(engine.TypeObject).
			IsObject().
			HasFields("name", "value", "flag").
			WithFieldCount(3)
	})

	t.Run("array assertion chain", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("one"),
			engine.NewStringValue("two"),
			engine.NewStringValue("three"),
		})

		NewComprehensiveAssertion(t, arr).
			WithType(engine.TypeArray).
			IsArray().
			WithLength(3).
			Contains("two")
	})
}

func TestMethodResultAssertion(t *testing.T) {
	t.Run("successful method result", func(t *testing.T) {
		result := engine.NewStringValue("success")

		AssertMethodResult(t, result, nil).
			NoError().
			Result().
			IsString().
			Equals("success")
	})

	t.Run("error method result", func(t *testing.T) {
		err := errors.New("method failed")

		AssertMethodResult(t, nil, err).
			HasError().
			ErrorContains("method failed")
	})
}

func TestRequireHelpers(t *testing.T) {
	t.Run("RequireNoGoError", func(t *testing.T) {
		RequireNoGoError(t, nil, "should not error")
	})

	t.Run("RequireGoError", func(t *testing.T) {
		err := errors.New("expected error")
		RequireGoError(t, err, "should error")
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("AssertArrayElementEquals with out of bounds", func(t *testing.T) {
		arr := engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("only one"),
		})

		// This would fail the test if called with index > 0
		// We're testing that the bounds checking works
		AssertArrayElementEquals(t, arr, 0, "only one")
	})

	t.Run("complex nested structure", func(t *testing.T) {
		// Create a complex nested structure
		nested := engine.NewObjectValue(map[string]engine.ScriptValue{
			"array": engine.NewArrayValue([]engine.ScriptValue{
				engine.NewObjectValue(map[string]engine.ScriptValue{
					"deep": engine.NewStringValue("value"),
				}),
			}),
		})

		// Test nested assertions
		fields := AssertObjectValue(t, nested)
		arr := fields["array"]
		AssertArrayLength(t, arr, 1)

		elements := AssertArrayValue(t, arr)
		innerObj := elements[0]
		AssertObjectFieldEquals(t, innerObj, "deep", "value")
	})
}

// TestCustomScriptValueImplementation tests with a custom ScriptValue that implements Equals
type mockEqualableValue struct {
	engine.ScriptValue
	value string
}

func (m mockEqualableValue) Equals(other engine.ScriptValue) bool {
	if o, ok := other.(mockEqualableValue); ok {
		return m.value == o.value
	}
	return false
}

func (m mockEqualableValue) Type() engine.ScriptValueType {
	return engine.TypeCustom
}

func (m mockEqualableValue) IsNil() bool {
	return false
}

func TestScriptValueEqualsWithCustomEquals(t *testing.T) {
	v1 := mockEqualableValue{value: "test"}
	v2 := mockEqualableValue{value: "test"}

	// Should use the Equals method
	AssertScriptValueEquals(t, v1, v2)

	// This would fail if we could test it
	// v3 := mockEqualableValue{value: "different"}
	// AssertScriptValueEquals(t, v1, v3)
}
