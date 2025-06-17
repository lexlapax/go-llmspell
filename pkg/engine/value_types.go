// ABOUTME: Defines the unified ScriptValue type system for cross-engine value representation
// ABOUTME: Provides type-safe value conversion and manipulation across script engines

package engine

import (
	"fmt"
	"reflect"
	"strings"
)

// ScriptValueType represents the type of a ScriptValue
type ScriptValueType int

const (
	TypeNil ScriptValueType = iota
	TypeBool
	TypeNumber
	TypeString
	TypeArray
	TypeObject
	TypeFunction
	TypeError
	TypeChannel
	TypeCustom
)

// String returns the string representation of the type
func (t ScriptValueType) String() string {
	switch t {
	case TypeNil:
		return "nil"
	case TypeBool:
		return "bool"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeFunction:
		return "function"
	case TypeError:
		return "error"
	case TypeChannel:
		return "channel"
	case TypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// ScriptValue represents a value that can be passed between script engines and Go
type ScriptValue interface {
	// Type returns the type of this value
	Type() ScriptValueType

	// IsNil returns true if this value represents nil/null/undefined
	IsNil() bool

	// String returns a string representation of the value
	String() string

	// ToGo converts the value to a native Go type
	ToGo() interface{}

	// Equals checks if this value equals another value
	Equals(other ScriptValue) bool
}

// NilValue represents a nil/null/undefined value
type NilValue struct{}

func (n NilValue) Type() ScriptValueType { return TypeNil }
func (n NilValue) IsNil() bool           { return true }
func (n NilValue) String() string        { return "nil" }
func (n NilValue) ToGo() interface{}     { return nil }
func (n NilValue) Equals(other ScriptValue) bool {
	return other != nil && other.Type() == TypeNil
}

// BoolValue represents a boolean value
type BoolValue struct {
	value bool
}

func (b BoolValue) Type() ScriptValueType { return TypeBool }
func (b BoolValue) IsNil() bool           { return false }
func (b BoolValue) String() string        { return fmt.Sprintf("%t", b.value) }
func (b BoolValue) ToGo() interface{}     { return b.value }
func (b BoolValue) Value() bool           { return b.value }
func (b BoolValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeBool {
		return false
	}
	if otherBool, ok := other.(BoolValue); ok {
		return b.value == otherBool.value
	}
	return false
}

// NumberValue represents a numeric value
type NumberValue struct {
	value float64
}

func (n NumberValue) Type() ScriptValueType { return TypeNumber }
func (n NumberValue) IsNil() bool           { return false }
func (n NumberValue) String() string        { return fmt.Sprintf("%g", n.value) }
func (n NumberValue) ToGo() interface{}     { return n.value }
func (n NumberValue) Value() float64        { return n.value }
func (n NumberValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeNumber {
		return false
	}
	if otherNum, ok := other.(NumberValue); ok {
		return n.value == otherNum.value
	}
	return false
}

// StringValue represents a string value
type StringValue struct {
	value string
}

func (s StringValue) Type() ScriptValueType { return TypeString }
func (s StringValue) IsNil() bool           { return false }
func (s StringValue) String() string        { return s.value }
func (s StringValue) ToGo() interface{}     { return s.value }
func (s StringValue) Value() string         { return s.value }
func (s StringValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeString {
		return false
	}
	if otherStr, ok := other.(StringValue); ok {
		return s.value == otherStr.value
	}
	return false
}

// ArrayValue represents an array/list value
type ArrayValue struct {
	elements []ScriptValue
}

func (a ArrayValue) Type() ScriptValueType { return TypeArray }
func (a ArrayValue) IsNil() bool           { return false }
func (a ArrayValue) String() string {
	var parts []string
	for _, elem := range a.elements {
		parts = append(parts, elem.String())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func (a ArrayValue) ToGo() interface{} {
	result := make([]interface{}, len(a.elements))
	for i, elem := range a.elements {
		result[i] = elem.ToGo()
	}
	return result
}
func (a ArrayValue) Elements() []ScriptValue { return a.elements }
func (a ArrayValue) Len() int                { return len(a.elements) }
func (a ArrayValue) Get(index int) (ScriptValue, bool) {
	if index < 0 || index >= len(a.elements) {
		return nil, false
	}
	return a.elements[index], true
}
func (a ArrayValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeArray {
		return false
	}
	if otherArr, ok := other.(ArrayValue); ok {
		if len(a.elements) != len(otherArr.elements) {
			return false
		}
		for i, elem := range a.elements {
			if !elem.Equals(otherArr.elements[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// ObjectValue represents an object/map/table value
type ObjectValue struct {
	fields map[string]ScriptValue
}

func (o ObjectValue) Type() ScriptValueType { return TypeObject }
func (o ObjectValue) IsNil() bool           { return false }
func (o ObjectValue) String() string {
	var parts []string
	for k, v := range o.fields {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v.String()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func (o ObjectValue) ToGo() interface{} {
	result := make(map[string]interface{})
	for k, v := range o.fields {
		result[k] = v.ToGo()
	}
	return result
}
func (o ObjectValue) Fields() map[string]ScriptValue { return o.fields }
func (o ObjectValue) Get(key string) (ScriptValue, bool) {
	val, ok := o.fields[key]
	return val, ok
}
func (o ObjectValue) Set(key string, value ScriptValue) {
	o.fields[key] = value
}
func (o ObjectValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeObject {
		return false
	}
	if otherObj, ok := other.(ObjectValue); ok {
		if len(o.fields) != len(otherObj.fields) {
			return false
		}
		for k, v := range o.fields {
			otherV, ok := otherObj.fields[k]
			if !ok || !v.Equals(otherV) {
				return false
			}
		}
		return true
	}
	return false
}

// FunctionValue represents a function value
type FunctionValue struct {
	name     string
	function interface{} // Can be various function types depending on engine
}

func (f FunctionValue) Type() ScriptValueType { return TypeFunction }
func (f FunctionValue) IsNil() bool           { return f.function == nil }
func (f FunctionValue) String() string {
	if f.name != "" {
		return fmt.Sprintf("function:%s", f.name)
	}
	return "function"
}
func (f FunctionValue) ToGo() interface{}     { return f.function }
func (f FunctionValue) Name() string          { return f.name }
func (f FunctionValue) Function() interface{} { return f.function }
func (f FunctionValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeFunction {
		return false
	}
	if otherFunc, ok := other.(FunctionValue); ok {
		// Function equality is based on identity
		return reflect.ValueOf(f.function).Pointer() == reflect.ValueOf(otherFunc.function).Pointer()
	}
	return false
}

// ErrorValue represents an error value
type ErrorValue struct {
	error error
}

func (e ErrorValue) Type() ScriptValueType { return TypeError }
func (e ErrorValue) IsNil() bool           { return e.error == nil }
func (e ErrorValue) String() string {
	if e.error != nil {
		return e.error.Error()
	}
	return ""
}
func (e ErrorValue) ToGo() interface{} { return e.error }
func (e ErrorValue) Error() error      { return e.error }
func (e ErrorValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeError {
		return false
	}
	if otherErr, ok := other.(ErrorValue); ok {
		// Error equality is based on error message
		if e.error == nil && otherErr.error == nil {
			return true
		}
		if e.error != nil && otherErr.error != nil {
			return e.error.Error() == otherErr.error.Error()
		}
	}
	return false
}

// ChannelValue represents a channel for inter-script communication
type ChannelValue struct {
	channel interface{} // Engine-specific channel type
	id      string
}

func (c ChannelValue) Type() ScriptValueType { return TypeChannel }
func (c ChannelValue) IsNil() bool           { return c.channel == nil }
func (c ChannelValue) String() string        { return fmt.Sprintf("channel:%s", c.id) }
func (c ChannelValue) ToGo() interface{}     { return c.channel }
func (c ChannelValue) Value() interface{}    { return c.channel }
func (c ChannelValue) ID() string            { return c.id }
func (c ChannelValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeChannel {
		return false
	}
	if otherChan, ok := other.(ChannelValue); ok {
		return c.id == otherChan.id
	}
	return false
}

// CustomValue represents a custom/user-defined value
type CustomValue struct {
	typeName string
	value    interface{}
}

func (c CustomValue) Type() ScriptValueType { return TypeCustom }
func (c CustomValue) IsNil() bool           { return c.value == nil }
func (c CustomValue) String() string        { return fmt.Sprintf("%s:%v", c.typeName, c.value) }
func (c CustomValue) ToGo() interface{}     { return c.value }
func (c CustomValue) TypeName() string      { return c.typeName }
func (c CustomValue) Value() interface{}    { return c.value }
func (c CustomValue) Equals(other ScriptValue) bool {
	if other == nil || other.Type() != TypeCustom {
		return false
	}
	if otherCustom, ok := other.(CustomValue); ok {
		if c.typeName != otherCustom.typeName {
			return false
		}
		// Custom equality is based on reflect.DeepEqual
		return reflect.DeepEqual(c.value, otherCustom.value)
	}
	return false
}

// Constructor functions

// NewNilValue creates a new nil value
func NewNilValue() ScriptValue {
	return NilValue{}
}

// NewBoolValue creates a new boolean value
func NewBoolValue(v bool) ScriptValue {
	return BoolValue{value: v}
}

// NewNumberValue creates a new number value
func NewNumberValue(v float64) ScriptValue {
	return NumberValue{value: v}
}

// NewStringValue creates a new string value
func NewStringValue(v string) ScriptValue {
	return StringValue{value: v}
}

// NewArrayValue creates a new array value
func NewArrayValue(elements []ScriptValue) ScriptValue {
	return ArrayValue{elements: elements}
}

// NewObjectValue creates a new object value
func NewObjectValue(fields map[string]ScriptValue) ScriptValue {
	if fields == nil {
		fields = make(map[string]ScriptValue)
	}
	return ObjectValue{fields: fields}
}

// NewFunctionValue creates a new function value
func NewFunctionValue(name string, fn interface{}) ScriptValue {
	return FunctionValue{name: name, function: fn}
}

// NewErrorValue creates a new error value
func NewErrorValue(err error) ScriptValue {
	return ErrorValue{error: err}
}

// NewChannelValue creates a new channel value
func NewChannelValue(id string, ch interface{}) ScriptValue {
	return ChannelValue{id: id, channel: ch}
}

// NewCustomValue creates a new custom value
func NewCustomValue(typeName string, value interface{}) ScriptValue {
	return CustomValue{typeName: typeName, value: value}
}

// Helper functions

// IsTrue returns whether a ScriptValue is truthy
func IsTrue(v ScriptValue) bool {
	if v == nil || v.IsNil() {
		return false
	}
	switch v.Type() {
	case TypeBool:
		if bv, ok := v.(BoolValue); ok {
			return bv.Value()
		}
	case TypeNumber:
		if nv, ok := v.(NumberValue); ok {
			return nv.Value() != 0
		}
	case TypeString:
		if sv, ok := v.(StringValue); ok {
			return sv.Value() != ""
		}
	case TypeArray:
		if av, ok := v.(ArrayValue); ok {
			return av.Len() > 0
		}
	case TypeObject:
		if ov, ok := v.(ObjectValue); ok {
			return len(ov.Fields()) > 0
		}
	case TypeFunction, TypeChannel, TypeCustom:
		return true
	case TypeError:
		return false
	}
	return false
}

// ConvertToString attempts to convert a ScriptValue to a string
func ConvertToString(v ScriptValue) (string, error) {
	if v == nil || v.IsNil() {
		return "", nil
	}
	return v.String(), nil
}

// ConvertToNumber attempts to convert a ScriptValue to a number
func ConvertToNumber(v ScriptValue) (float64, error) {
	if v == nil || v.IsNil() {
		return 0, nil
	}
	switch v.Type() {
	case TypeNumber:
		if nv, ok := v.(NumberValue); ok {
			return nv.Value(), nil
		}
	case TypeBool:
		if bv, ok := v.(BoolValue); ok {
			if bv.Value() {
				return 1, nil
			}
			return 0, nil
		}
	case TypeString:
		if sv, ok := v.(StringValue); ok {
			var num float64
			_, err := fmt.Sscanf(sv.Value(), "%f", &num)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to number", sv.Value())
			}
			return num, nil
		}
	}
	return 0, fmt.Errorf("cannot convert %s to number", v.Type())
}

// ConvertToBool attempts to convert a ScriptValue to a boolean
func ConvertToBool(v ScriptValue) (bool, error) {
	return IsTrue(v), nil
}
