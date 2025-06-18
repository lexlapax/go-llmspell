// ABOUTME: Centralized conversion utilities for ScriptValue type system
// ABOUTME: Provides standard conversion functions to eliminate duplication across bridges

package engine

import (
	"fmt"
)

// ConvertToScriptValue converts interface{} to appropriate ScriptValue type
// This is the canonical conversion function used across all bridges
func ConvertToScriptValue(v interface{}) ScriptValue {
	switch val := v.(type) {
	case nil:
		return NewNilValue()
	case bool:
		return NewBoolValue(val)
	case int:
		return NewNumberValue(float64(val))
	case int8:
		return NewNumberValue(float64(val))
	case int16:
		return NewNumberValue(float64(val))
	case int32:
		return NewNumberValue(float64(val))
	case int64:
		return NewNumberValue(float64(val))
	case uint:
		return NewNumberValue(float64(val))
	case uint8:
		return NewNumberValue(float64(val))
	case uint16:
		return NewNumberValue(float64(val))
	case uint32:
		return NewNumberValue(float64(val))
	case uint64:
		return NewNumberValue(float64(val))
	case float32:
		return NewNumberValue(float64(val))
	case float64:
		return NewNumberValue(val)
	case string:
		return NewStringValue(val)
	case []interface{}:
		return NewArrayValue(ConvertSliceToScriptValue(val))
	case map[string]interface{}:
		return NewObjectValue(ConvertMapToScriptValue(val))
	case ScriptValue:
		// Already a ScriptValue, return as-is
		return val
	default:
		// Fallback to string representation
		return NewStringValue(fmt.Sprintf("%v", v))
	}
}

// ConvertMapToScriptValue converts a map[string]interface{} to map[string]ScriptValue
func ConvertMapToScriptValue(data map[string]interface{}) map[string]ScriptValue {
	if data == nil {
		return nil
	}
	result := make(map[string]ScriptValue, len(data))
	for k, v := range data {
		result[k] = ConvertToScriptValue(v)
	}
	return result
}

// ConvertSliceToScriptValue converts a []interface{} to []ScriptValue
func ConvertSliceToScriptValue(data []interface{}) []ScriptValue {
	if data == nil {
		return nil
	}
	result := make([]ScriptValue, len(data))
	for i, v := range data {
		result[i] = ConvertToScriptValue(v)
	}
	return result
}

// ConvertFromScriptValue converts ScriptValue back to interface{}
// This is useful for bridges that need to pass data to go-llms functions
func ConvertFromScriptValue(v ScriptValue) interface{} {
	if v == nil || v.IsNil() {
		return nil
	}
	return v.ToGo()
}

// ConvertScriptValueMap converts map[string]ScriptValue to map[string]interface{}
func ConvertScriptValueMap(data map[string]ScriptValue) map[string]interface{} {
	if data == nil {
		return nil
	}
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		result[k] = ConvertFromScriptValue(v)
	}
	return result
}

// ConvertScriptValueSlice converts []ScriptValue to []interface{}
func ConvertScriptValueSlice(data []ScriptValue) []interface{} {
	if data == nil {
		return nil
	}
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = ConvertFromScriptValue(v)
	}
	return result
}

// Validation helpers for bridge method arguments

// ValidateStringArg validates that args[index] is a string and returns its value
func ValidateStringArg(args []ScriptValue, index int, name string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeString {
		return "", fmt.Errorf("%s must be string", name)
	}
	return args[index].(StringValue).Value(), nil
}

// ValidateNumberArg validates that args[index] is a number and returns its value
func ValidateNumberArg(args []ScriptValue, index int, name string) (float64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeNumber {
		return 0, fmt.Errorf("%s must be number", name)
	}
	return args[index].(NumberValue).Value(), nil
}

// ValidateBoolArg validates that args[index] is a boolean and returns its value
func ValidateBoolArg(args []ScriptValue, index int, name string) (bool, error) {
	if len(args) <= index {
		return false, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeBool {
		return false, fmt.Errorf("%s must be boolean", name)
	}
	return args[index].(BoolValue).Value(), nil
}

// ValidateObjectArg validates that args[index] is an object and returns its fields as map[string]interface{}
func ValidateObjectArg(args []ScriptValue, index int, name string) (map[string]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeObject {
		return nil, fmt.Errorf("%s must be object", name)
	}

	result := make(map[string]interface{})
	for k, v := range args[index].(ObjectValue).Fields() {
		result[k] = v.ToGo()
	}
	return result, nil
}

// ValidateArrayArg validates that args[index] is an array and returns its values as []interface{}
func ValidateArrayArg(args []ScriptValue, index int, name string) ([]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeArray {
		return nil, fmt.Errorf("%s must be array", name)
	}

	result := make([]interface{}, 0)
	for _, v := range args[index].(ArrayValue).Elements() {
		result = append(result, v.ToGo())
	}
	return result, nil
}

// ValidateFunctionArg validates that args[index] is a function and returns it
func ValidateFunctionArg(args []ScriptValue, index int, name string) (FunctionValue, error) {
	if len(args) <= index {
		return FunctionValue{}, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != TypeFunction {
		return FunctionValue{}, fmt.Errorf("%s must be function", name)
	}
	return args[index].(FunctionValue), nil
}

// ValidateOptionalStringArg validates an optional string argument with default value
func ValidateOptionalStringArg(args []ScriptValue, index int, defaultValue string) string {
	if len(args) <= index || args[index] == nil || args[index].Type() != TypeString {
		return defaultValue
	}
	return args[index].(StringValue).Value()
}

// ValidateOptionalNumberArg validates an optional number argument with default value
func ValidateOptionalNumberArg(args []ScriptValue, index int, defaultValue float64) float64 {
	if len(args) <= index || args[index] == nil || args[index].Type() != TypeNumber {
		return defaultValue
	}
	return args[index].(NumberValue).Value()
}

// ValidateOptionalBoolArg validates an optional boolean argument with default value
func ValidateOptionalBoolArg(args []ScriptValue, index int, defaultValue bool) bool {
	if len(args) <= index || args[index] == nil || args[index].Type() != TypeBool {
		return defaultValue
	}
	return args[index].(BoolValue).Value()
}
