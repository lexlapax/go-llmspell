// ABOUTME: Helper functions for ScriptValue conversions in bridge implementations
// ABOUTME: Reduces boilerplate code when converting between Go types and ScriptValues

package structured

import (
	"fmt"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ConvertMapToScriptValue converts a map[string]interface{} to map[string]engine.ScriptValue
func ConvertMapToScriptValue(data map[string]interface{}) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)
	for k, v := range data {
		result[k] = ConvertInterfaceToScriptValue(v)
	}
	return result
}

// ConvertSliceToScriptValue converts a []interface{} to []engine.ScriptValue
func ConvertSliceToScriptValue(data []interface{}) []engine.ScriptValue {
	result := make([]engine.ScriptValue, len(data))
	for i, v := range data {
		result[i] = ConvertInterfaceToScriptValue(v)
	}
	return result
}

// ConvertInterfaceToScriptValue converts interface{} to appropriate ScriptValue type
func ConvertInterfaceToScriptValue(v interface{}) engine.ScriptValue {
	switch val := v.(type) {
	case string:
		return engine.NewStringValue(val)
	case float64:
		return engine.NewNumberValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case bool:
		return engine.NewBoolValue(val)
	case nil:
		return engine.NewNilValue()
	case []interface{}:
		return engine.NewArrayValue(ConvertSliceToScriptValue(val))
	case map[string]interface{}:
		return engine.NewObjectValue(ConvertMapToScriptValue(val))
	default:
		// Fallback to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}

// ValidateStringArg validates that args[index] is a string and returns its value
func ValidateStringArg(args []engine.ScriptValue, index int, name string) (string, error) {
	if len(args) <= index {
		return "", fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeString {
		return "", fmt.Errorf("%s must be string", name)
	}
	return args[index].(engine.StringValue).Value(), nil
}

// ValidateNumberArg validates that args[index] is a number and returns its value
func ValidateNumberArg(args []engine.ScriptValue, index int, name string) (float64, error) {
	if len(args) <= index {
		return 0, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeNumber {
		return 0, fmt.Errorf("%s must be number", name)
	}
	return args[index].(engine.NumberValue).Value(), nil
}

// ValidateObjectArg validates that args[index] is an object and returns its fields as map[string]interface{}
func ValidateObjectArg(args []engine.ScriptValue, index int, name string) (map[string]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeObject {
		return nil, fmt.Errorf("%s must be object", name)
	}

	result := make(map[string]interface{})
	for k, v := range args[index].(engine.ObjectValue).Fields() {
		result[k] = v.ToGo()
	}
	return result, nil
}

// ValidateArrayArg validates that args[index] is an array and returns its values as []interface{}
func ValidateArrayArg(args []engine.ScriptValue, index int, name string) ([]interface{}, error) {
	if len(args) <= index {
		return nil, fmt.Errorf("%s argument required", name)
	}
	if args[index] == nil || args[index].Type() != engine.TypeArray {
		return nil, fmt.Errorf("%s must be array", name)
	}

	result := make([]interface{}, 0)
	for _, v := range args[index].(engine.ArrayValue).Elements() {
		result = append(result, v.ToGo())
	}
	return result, nil
}
