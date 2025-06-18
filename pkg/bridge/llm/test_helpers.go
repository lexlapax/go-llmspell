// ABOUTME: Helper functions for test files to convert Go types to ScriptValue types
// ABOUTME: Used by llm package tests to create proper ScriptValue arguments

package llm

import (
	"fmt"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// testConvertMapToScriptValue converts map[string]interface{} to map[string]engine.ScriptValue for tests
func testConvertMapToScriptValue(m map[string]interface{}) map[string]engine.ScriptValue {
	if m == nil {
		return make(map[string]engine.ScriptValue)
	}

	result := make(map[string]engine.ScriptValue)
	for k, v := range m {
		result[k] = testConvertToScriptValue(v)
	}
	return result
}

// testConvertSliceToScriptValue converts []interface{} to []engine.ScriptValue for tests
func testConvertSliceToScriptValue(s []interface{}) []engine.ScriptValue {
	if s == nil {
		return make([]engine.ScriptValue, 0)
	}

	result := make([]engine.ScriptValue, len(s))
	for i, v := range s {
		result[i] = testConvertToScriptValue(v)
	}
	return result
}

// testConvertToScriptValue converts interface{} to ScriptValue for tests
func testConvertToScriptValue(v interface{}) engine.ScriptValue {
	if v == nil {
		return engine.NewNilValue()
	}

	switch val := v.(type) {
	case bool:
		return engine.NewBoolValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int32:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case float32:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case string:
		return engine.NewStringValue(val)
	case []string:
		arr := make([]engine.ScriptValue, len(val))
		for i, s := range val {
			arr[i] = engine.NewStringValue(s)
		}
		return engine.NewArrayValue(arr)
	case []interface{}:
		return engine.NewArrayValue(testConvertSliceToScriptValue(val))
	case map[string]interface{}:
		return engine.NewObjectValue(testConvertMapToScriptValue(val))
	case engine.ScriptValue:
		// Already a ScriptValue
		return val
	default:
		// Convert to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}
