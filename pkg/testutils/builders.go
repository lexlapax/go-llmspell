// ABOUTME: ScriptValue builders provide fluent API for creating test data and complex value structures
// ABOUTME: Simplifies test data creation with method chaining and quick creator functions

package testutils

import (
	"fmt"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ScriptValueBuilder provides fluent API for building test ScriptValues
type ScriptValueBuilder struct {
	values []engine.ScriptValue
}

// NewScriptValueBuilder creates a new builder instance
func NewScriptValueBuilder() *ScriptValueBuilder {
	return &ScriptValueBuilder{
		values: make([]engine.ScriptValue, 0),
	}
}

// String adds a string value to the builder
func (b *ScriptValueBuilder) String(s string) *ScriptValueBuilder {
	b.values = append(b.values, engine.NewStringValue(s))
	return b
}

// Number adds a number value to the builder
func (b *ScriptValueBuilder) Number(n float64) *ScriptValueBuilder {
	b.values = append(b.values, engine.NewNumberValue(n))
	return b
}

// Int adds an integer as a number value to the builder
func (b *ScriptValueBuilder) Int(n int) *ScriptValueBuilder {
	return b.Number(float64(n))
}

// Bool adds a boolean value to the builder
func (b *ScriptValueBuilder) Bool(v bool) *ScriptValueBuilder {
	b.values = append(b.values, engine.NewBoolValue(v))
	return b
}

// Nil adds a nil value to the builder
func (b *ScriptValueBuilder) Nil() *ScriptValueBuilder {
	b.values = append(b.values, engine.NewNilValue())
	return b
}

// Object adds an object value to the builder
func (b *ScriptValueBuilder) Object(fields map[string]interface{}) *ScriptValueBuilder {
	obj := ObjectFromMap(fields)
	b.values = append(b.values, obj)
	return b
}

// Array adds an array value to the builder
func (b *ScriptValueBuilder) Array(elements ...interface{}) *ScriptValueBuilder {
	arr := ArrayFromSlice(elements)
	b.values = append(b.values, arr)
	return b
}

// Custom adds a custom value to the builder
func (b *ScriptValueBuilder) Custom(typeName string, value interface{}) *ScriptValueBuilder {
	b.values = append(b.values, engine.NewCustomValue(typeName, value))
	return b
}

// Error adds an error value to the builder
func (b *ScriptValueBuilder) Error(err error) *ScriptValueBuilder {
	b.values = append(b.values, engine.NewErrorValue(err))
	return b
}

// ErrorString adds an error value with string message to the builder
func (b *ScriptValueBuilder) ErrorString(msg string) *ScriptValueBuilder {
	return b.Error(fmt.Errorf("%s", msg))
}

// Build returns the built values as a slice
func (b *ScriptValueBuilder) Build() []engine.ScriptValue {
	return b.values
}

// BuildSingle returns a single value (the first one added)
func (b *ScriptValueBuilder) BuildSingle() engine.ScriptValue {
	if len(b.values) == 0 {
		return engine.NewNilValue()
	}
	return b.values[0]
}

// Quick creator functions for common patterns

// StringValue creates a string ScriptValue
func StringValue(s string) engine.ScriptValue {
	return engine.NewStringValue(s)
}

// NumberValue creates a number ScriptValue
func NumberValue(n float64) engine.ScriptValue {
	return engine.NewNumberValue(n)
}

// IntValue creates a number ScriptValue from an integer
func IntValue(n int) engine.ScriptValue {
	return engine.NewNumberValue(float64(n))
}

// BoolValue creates a boolean ScriptValue
func BoolValue(b bool) engine.ScriptValue {
	return engine.NewBoolValue(b)
}

// NilValue creates a nil ScriptValue
func NilValue() engine.ScriptValue {
	return engine.NewNilValue()
}

// ErrorValue creates an error ScriptValue
func ErrorValue(err error) engine.ScriptValue {
	return engine.NewErrorValue(err)
}

// ErrorStringValue creates an error ScriptValue from string
func ErrorStringValue(msg string) engine.ScriptValue {
	return engine.NewErrorValue(fmt.Errorf("%s", msg))
}

// ObjectFromMap converts a map[string]interface{} to ObjectValue
func ObjectFromMap(m map[string]interface{}) engine.ScriptValue {
	fields := make(map[string]engine.ScriptValue)
	for k, v := range m {
		fields[k] = InterfaceToScriptValue(v)
	}
	return engine.NewObjectValue(fields)
}

// ArrayFromSlice converts a []interface{} to ArrayValue
func ArrayFromSlice(s []interface{}) engine.ScriptValue {
	elements := make([]engine.ScriptValue, len(s))
	for i, v := range s {
		elements[i] = InterfaceToScriptValue(v)
	}
	return engine.NewArrayValue(elements)
}

// ArrayFromStrings creates an array of string values
func ArrayFromStrings(strings ...string) engine.ScriptValue {
	elements := make([]engine.ScriptValue, len(strings))
	for i, s := range strings {
		elements[i] = engine.NewStringValue(s)
	}
	return engine.NewArrayValue(elements)
}

// ArrayFromNumbers creates an array of number values
func ArrayFromNumbers(numbers ...float64) engine.ScriptValue {
	elements := make([]engine.ScriptValue, len(numbers))
	for i, n := range numbers {
		elements[i] = engine.NewNumberValue(n)
	}
	return engine.NewArrayValue(elements)
}

// ArrayFromInts creates an array of number values from integers
func ArrayFromInts(ints ...int) engine.ScriptValue {
	elements := make([]engine.ScriptValue, len(ints))
	for i, n := range ints {
		elements[i] = engine.NewNumberValue(float64(n))
	}
	return engine.NewArrayValue(elements)
}

// InterfaceToScriptValue converts an interface{} to appropriate ScriptValue
func InterfaceToScriptValue(v interface{}) engine.ScriptValue {
	switch val := v.(type) {
	case nil:
		return engine.NewNilValue()
	case bool:
		return engine.NewBoolValue(val)
	case string:
		return engine.NewStringValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int8:
		return engine.NewNumberValue(float64(val))
	case int16:
		return engine.NewNumberValue(float64(val))
	case int32:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case uint:
		return engine.NewNumberValue(float64(val))
	case uint8:
		return engine.NewNumberValue(float64(val))
	case uint16:
		return engine.NewNumberValue(float64(val))
	case uint32:
		return engine.NewNumberValue(float64(val))
	case uint64:
		return engine.NewNumberValue(float64(val))
	case float32:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case []interface{}:
		return ArrayFromSlice(val)
	case map[string]interface{}:
		return ObjectFromMap(val)
	case error:
		return engine.NewErrorValue(val)
	case engine.ScriptValue:
		return val // Already a ScriptValue
	default:
		// For unknown types, create a custom value
		return engine.NewCustomValue(fmt.Sprintf("%T", v), v)
	}
}

// Test data factory methods

// CreateTestObject creates a test object with common fields
func CreateTestObject(id string, name string, value float64) engine.ScriptValue {
	return ObjectFromMap(map[string]interface{}{
		"id":    id,
		"name":  name,
		"value": value,
		"type":  "test",
	})
}

// CreateTestArray creates a test array with mixed types
func CreateTestArray(size int) engine.ScriptValue {
	elements := make([]interface{}, size)
	for i := 0; i < size; i++ {
		switch i % 3 {
		case 0:
			elements[i] = fmt.Sprintf("item-%d", i)
		case 1:
			elements[i] = float64(i)
		case 2:
			elements[i] = i%2 == 0
		}
	}
	return ArrayFromSlice(elements)
}

// CreateNestedObject creates a nested object structure for testing
func CreateNestedObject() engine.ScriptValue {
	return ObjectFromMap(map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"value": "deep",
					"array": []interface{}{1, 2, 3},
				},
				"name": "nested",
			},
			"items": []interface{}{
				map[string]interface{}{"id": 1, "name": "first"},
				map[string]interface{}{"id": 2, "name": "second"},
			},
		},
		"metadata": map[string]interface{}{
			"created": time.Now().Format(time.RFC3339),
			"version": "1.0.0",
			"enabled": true,
		},
	})
}

// CreateComplexArray creates an array with nested structures
func CreateComplexArray() engine.ScriptValue {
	return ArrayFromSlice([]interface{}{
		"string value",
		42.5,
		true,
		nil,
		map[string]interface{}{
			"nested": "object",
			"count":  10,
		},
		[]interface{}{1, 2, 3},
		map[string]interface{}{
			"array": []interface{}{
				map[string]interface{}{"deep": true},
			},
		},
	})
}

// ObjectBuilder provides a fluent interface for building objects
type ObjectBuilder struct {
	fields map[string]interface{}
}

// NewObjectBuilder creates a new object builder
func NewObjectBuilder() *ObjectBuilder {
	return &ObjectBuilder{
		fields: make(map[string]interface{}),
	}
}

// Set adds a field to the object
func (b *ObjectBuilder) Set(key string, value interface{}) *ObjectBuilder {
	b.fields[key] = value
	return b
}

// SetString adds a string field
func (b *ObjectBuilder) SetString(key string, value string) *ObjectBuilder {
	return b.Set(key, value)
}

// SetNumber adds a number field
func (b *ObjectBuilder) SetNumber(key string, value float64) *ObjectBuilder {
	return b.Set(key, value)
}

// SetInt adds an integer field
func (b *ObjectBuilder) SetInt(key string, value int) *ObjectBuilder {
	return b.Set(key, value)
}

// SetBool adds a boolean field
func (b *ObjectBuilder) SetBool(key string, value bool) *ObjectBuilder {
	return b.Set(key, value)
}

// SetNil adds a nil field
func (b *ObjectBuilder) SetNil(key string) *ObjectBuilder {
	return b.Set(key, nil)
}

// SetObject adds a nested object field
func (b *ObjectBuilder) SetObject(key string, obj *ObjectBuilder) *ObjectBuilder {
	return b.Set(key, obj.fields)
}

// SetArray adds an array field
func (b *ObjectBuilder) SetArray(key string, elements ...interface{}) *ObjectBuilder {
	return b.Set(key, elements)
}

// Build creates the ScriptValue object
func (b *ObjectBuilder) Build() engine.ScriptValue {
	return ObjectFromMap(b.fields)
}

// ArrayBuilder provides a fluent interface for building arrays
type ArrayBuilder struct {
	elements []interface{}
}

// NewArrayBuilder creates a new array builder
func NewArrayBuilder() *ArrayBuilder {
	return &ArrayBuilder{
		elements: make([]interface{}, 0),
	}
}

// Add adds an element to the array
func (b *ArrayBuilder) Add(value interface{}) *ArrayBuilder {
	b.elements = append(b.elements, value)
	return b
}

// AddString adds a string element
func (b *ArrayBuilder) AddString(value string) *ArrayBuilder {
	return b.Add(value)
}

// AddNumber adds a number element
func (b *ArrayBuilder) AddNumber(value float64) *ArrayBuilder {
	return b.Add(value)
}

// AddInt adds an integer element
func (b *ArrayBuilder) AddInt(value int) *ArrayBuilder {
	return b.Add(value)
}

// AddBool adds a boolean element
func (b *ArrayBuilder) AddBool(value bool) *ArrayBuilder {
	return b.Add(value)
}

// AddNil adds a nil element
func (b *ArrayBuilder) AddNil() *ArrayBuilder {
	return b.Add(nil)
}

// AddObject adds an object element
func (b *ArrayBuilder) AddObject(obj *ObjectBuilder) *ArrayBuilder {
	return b.Add(obj.fields)
}

// AddArray adds a nested array element
func (b *ArrayBuilder) AddArray(arr *ArrayBuilder) *ArrayBuilder {
	return b.Add(arr.elements)
}

// Build creates the ScriptValue array
func (b *ArrayBuilder) Build() engine.ScriptValue {
	return ArrayFromSlice(b.elements)
}
