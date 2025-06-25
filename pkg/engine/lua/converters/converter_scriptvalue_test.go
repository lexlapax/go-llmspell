// ABOUTME: Comprehensive tests for ScriptValue to Lua LValue bi-directional conversion
// ABOUTME: Tests circular reference detection, type conversion accuracy, and edge cases

package gopherlua

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

func TestLValueToScriptValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := NewScriptValueConverter(NewLuaTypeConverter())

	tests := []struct {
		name        string
		setup       func() lua.LValue
		expected    engine.ScriptValue
		expectError bool
	}{
		{
			name: "nil value",
			setup: func() lua.LValue {
				return lua.LNil
			},
			expected: engine.NewNilValue(),
		},
		{
			name: "boolean true",
			setup: func() lua.LValue {
				return lua.LBool(true)
			},
			expected: engine.NewBoolValue(true),
		},
		{
			name: "boolean false",
			setup: func() lua.LValue {
				return lua.LBool(false)
			},
			expected: engine.NewBoolValue(false),
		},
		{
			name: "number",
			setup: func() lua.LValue {
				return lua.LNumber(42.5)
			},
			expected: engine.NewNumberValue(42.5),
		},
		{
			name: "string",
			setup: func() lua.LValue {
				return lua.LString("hello world")
			},
			expected: engine.NewStringValue("hello world"),
		},
		{
			name: "array table",
			setup: func() lua.LValue {
				table := L.NewTable()
				table.RawSetInt(1, lua.LString("first"))
				table.RawSetInt(2, lua.LString("second"))
				table.RawSetInt(3, lua.LNumber(42))
				return table
			},
			expected: engine.NewArrayValue([]engine.ScriptValue{
				engine.NewStringValue("first"),
				engine.NewStringValue("second"),
				engine.NewNumberValue(42),
			}),
		},
		{
			name: "object table",
			setup: func() lua.LValue {
				table := L.NewTable()
				table.RawSetString("name", lua.LString("test"))
				table.RawSetString("value", lua.LNumber(123))
				table.RawSetString("active", lua.LBool(true))
				return table
			},
			expected: engine.NewObjectValue(map[string]engine.ScriptValue{
				"name":   engine.NewStringValue("test"),
				"value":  engine.NewNumberValue(123),
				"active": engine.NewBoolValue(true),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lval := tt.setup()
			result, err := converter.LValueToScriptValue(L, lval)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("expected %v (%s), got %v (%s)",
					tt.expected, tt.expected.Type(), result, result.Type())
			}
		})
	}
}

func TestScriptValueToLValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := NewScriptValueConverter(NewLuaTypeConverter())

	tests := []struct {
		name        string
		input       engine.ScriptValue
		verify      func(lua.LValue) bool
		expectError bool
	}{
		{
			name:  "nil value",
			input: engine.NewNilValue(),
			verify: func(lv lua.LValue) bool {
				return lv == lua.LNil
			},
		},
		{
			name:  "boolean true",
			input: engine.NewBoolValue(true),
			verify: func(lv lua.LValue) bool {
				return lv == lua.LBool(true)
			},
		},
		{
			name:  "boolean false",
			input: engine.NewBoolValue(false),
			verify: func(lv lua.LValue) bool {
				return lv == lua.LBool(false)
			},
		},
		{
			name:  "number",
			input: engine.NewNumberValue(42.5),
			verify: func(lv lua.LValue) bool {
				if lnum, ok := lv.(lua.LNumber); ok {
					return float64(lnum) == 42.5
				}
				return false
			},
		},
		{
			name:  "string",
			input: engine.NewStringValue("hello world"),
			verify: func(lv lua.LValue) bool {
				if lstr, ok := lv.(lua.LString); ok {
					return string(lstr) == "hello world"
				}
				return false
			},
		},
		{
			name: "array",
			input: engine.NewArrayValue([]engine.ScriptValue{
				engine.NewStringValue("first"),
				engine.NewStringValue("second"),
				engine.NewNumberValue(42),
			}),
			verify: func(lv lua.LValue) bool {
				if table, ok := lv.(*lua.LTable); ok {
					return table.Len() == 3 &&
						table.RawGetInt(1).String() == "first" &&
						table.RawGetInt(2).String() == "second" &&
						table.RawGetInt(3).String() == "42"
				}
				return false
			},
		},
		{
			name: "object",
			input: engine.NewObjectValue(map[string]engine.ScriptValue{
				"name":   engine.NewStringValue("test"),
				"value":  engine.NewNumberValue(123),
				"active": engine.NewBoolValue(true),
			}),
			verify: func(lv lua.LValue) bool {
				if table, ok := lv.(*lua.LTable); ok {
					name := table.RawGetString("name")
					value := table.RawGetString("value")
					active := table.RawGetString("active")

					return name.String() == "test" &&
						value.String() == "123" &&
						active.String() == "true"
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ScriptValueToLValue(L, tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !tt.verify(result) {
				t.Errorf("verification failed for result: %v (type: %T)", result, result)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := NewScriptValueConverter(NewLuaTypeConverter())

	tests := []struct {
		name  string
		value engine.ScriptValue
	}{
		{
			name:  "nil",
			value: engine.NewNilValue(),
		},
		{
			name:  "boolean",
			value: engine.NewBoolValue(true),
		},
		{
			name:  "number",
			value: engine.NewNumberValue(42.5),
		},
		{
			name:  "string",
			value: engine.NewStringValue("test"),
		},
		{
			name: "simple array",
			value: engine.NewArrayValue([]engine.ScriptValue{
				engine.NewStringValue("a"),
				engine.NewNumberValue(1),
				engine.NewBoolValue(true),
			}),
		},
		{
			name: "simple object",
			value: engine.NewObjectValue(map[string]engine.ScriptValue{
				"str":  engine.NewStringValue("value"),
				"num":  engine.NewNumberValue(42),
				"bool": engine.NewBoolValue(false),
			}),
		},
		{
			name: "nested structure",
			value: engine.NewObjectValue(map[string]engine.ScriptValue{
				"nested": engine.NewArrayValue([]engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"deep": engine.NewStringValue("value"),
					}),
				}),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert ScriptValue to LValue
			lval, err := converter.ScriptValueToLValue(L, tt.value)
			if err != nil {
				t.Fatalf("ScriptValue to LValue conversion failed: %v", err)
			}

			// Convert LValue back to ScriptValue
			result, err := converter.LValueToScriptValue(L, lval)
			if err != nil {
				t.Fatalf("LValue to ScriptValue conversion failed: %v", err)
			}

			// Check if they're equal
			if !result.Equals(tt.value) {
				t.Errorf("round trip conversion failed:\noriginal: %v\nresult:   %v",
					tt.value, result)
			}
		})
	}
}

func TestCircularReferenceDetection(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := NewScriptValueConverter(NewLuaTypeConverter())

	// Create a table with circular reference
	table := L.NewTable()
	table.RawSetString("self", table) // Circular reference

	_, err := converter.LValueToScriptValue(L, table)
	if err == nil {
		t.Errorf("expected error for circular reference but got none")
	}

	if err != nil && !strings.Contains(err.Error(), "circular reference detected") {
		t.Errorf("expected circular reference error, got: %v", err)
	}
}

func TestGoValueConversion(t *testing.T) {
	converter := NewScriptValueConverter(NewLuaTypeConverter())

	tests := []struct {
		name     string
		input    interface{}
		expected engine.ScriptValue
	}{
		{
			name:     "nil",
			input:    nil,
			expected: engine.NewNilValue(),
		},
		{
			name:     "bool",
			input:    true,
			expected: engine.NewBoolValue(true),
		},
		{
			name:     "int",
			input:    42,
			expected: engine.NewNumberValue(42),
		},
		{
			name:     "float64",
			input:    3.14,
			expected: engine.NewNumberValue(3.14),
		},
		{
			name:     "string",
			input:    "hello",
			expected: engine.NewStringValue("hello"),
		},
		{
			name:  "slice",
			input: []interface{}{"a", 1, true},
			expected: engine.NewArrayValue([]engine.ScriptValue{
				engine.NewStringValue("a"),
				engine.NewNumberValue(1),
				engine.NewBoolValue(true),
			}),
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: engine.NewObjectValue(map[string]engine.ScriptValue{
				"key1": engine.NewStringValue("value1"),
				"key2": engine.NewNumberValue(42),
			}),
		},
		{
			name:     "error",
			input:    fmt.Errorf("test error"),
			expected: engine.NewErrorValue(fmt.Errorf("test error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.GoToScriptValue(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestScriptValueToGo(t *testing.T) {
	converter := NewScriptValueConverter(NewLuaTypeConverter())

	tests := []struct {
		name     string
		input    engine.ScriptValue
		expected interface{}
	}{
		{
			name:     "nil",
			input:    engine.NewNilValue(),
			expected: nil,
		},
		{
			name:     "bool",
			input:    engine.NewBoolValue(true),
			expected: true,
		},
		{
			name:     "number",
			input:    engine.NewNumberValue(42.5),
			expected: 42.5,
		},
		{
			name:     "string",
			input:    engine.NewStringValue("hello"),
			expected: "hello",
		},
		{
			name: "array",
			input: engine.NewArrayValue([]engine.ScriptValue{
				engine.NewStringValue("a"),
				engine.NewNumberValue(1),
			}),
			expected: []interface{}{"a", float64(1)},
		},
		{
			name: "object",
			input: engine.NewObjectValue(map[string]engine.ScriptValue{
				"key": engine.NewStringValue("value"),
			}),
			expected: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ScriptValueToGo(tt.input)

			// For complex types, we need deep comparison
			switch expected := tt.expected.(type) {
			case []interface{}:
				if resultSlice, ok := result.([]interface{}); ok {
					if len(resultSlice) != len(expected) {
						t.Errorf("slice length mismatch: expected %d, got %d",
							len(expected), len(resultSlice))
						return
					}
					for i, v := range expected {
						if resultSlice[i] != v {
							t.Errorf("slice element %d mismatch: expected %v, got %v",
								i, v, resultSlice[i])
						}
					}
				} else {
					t.Errorf("expected slice, got %T", result)
				}
			case map[string]interface{}:
				if resultMap, ok := result.(map[string]interface{}); ok {
					if len(resultMap) != len(expected) {
						t.Errorf("map length mismatch: expected %d, got %d",
							len(expected), len(resultMap))
						return
					}
					for k, v := range expected {
						if resultMap[k] != v {
							t.Errorf("map value for key %s mismatch: expected %v, got %v",
								k, v, resultMap[k])
						}
					}
				} else {
					t.Errorf("expected map, got %T", result)
				}
			default:
				if result != expected {
					t.Errorf("expected %v, got %v", expected, result)
				}
			}
		})
	}
}

func TestMaxDepthLimiting(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	converter := NewScriptValueConverter(NewLuaTypeConverter())
	converter.maxDepth = 2 // Set low depth limit for testing

	// Create deeply nested structure
	deep := engine.NewObjectValue(map[string]engine.ScriptValue{
		"level1": engine.NewObjectValue(map[string]engine.ScriptValue{
			"level2": engine.NewObjectValue(map[string]engine.ScriptValue{
				"level3": engine.NewStringValue("too deep"),
			}),
		}),
	})

	_, err := converter.ScriptValueToLValue(L, deep)
	if err == nil {
		t.Errorf("expected depth limit error but got none")
	}
}
