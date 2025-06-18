// ABOUTME: Performance benchmarks comparing interface{} vs ScriptValue type system
// ABOUTME: Measures conversion overhead, type checking performance, and memory usage

//go:build bench
// +build bench

package benchmarks

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Import types for direct reference
type (
	ScriptValue = engine.ScriptValue
	StringValue = engine.StringValue
	NumberValue = engine.NumberValue
	BoolValue   = engine.BoolValue
	ArrayValue  = engine.ArrayValue
	ObjectValue = engine.ObjectValue
)

// Benchmark comparing interface{} type assertions vs ScriptValue type checking
func BenchmarkTypeChecking(b *testing.B) {
	b.Run("Interface_TypeAssertion", func(b *testing.B) {
		var values []interface{}
		for i := 0; i < 100; i++ {
			values = append(values, "test string")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, v := range values {
				// Old way: dangerous type assertion
				if str, ok := v.(string); ok {
					_ = str
				}
			}
		}
	})

	b.Run("ScriptValue_TypeCheck", func(b *testing.B) {
		var values []engine.ScriptValue
		for i := 0; i < 100; i++ {
			values = append(values, engine.NewStringValue("test string"))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, v := range values {
				// New way: safe type check
				if v.Type() == engine.TypeString {
					_ = v.(engine.StringValue).Value()
				}
			}
		}
	})
}

// Benchmark method execution with different type systems
func BenchmarkMethodExecution(b *testing.B) {
	// Simulated bridge method with interface{}
	executeOld := func(args []interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, nil
		}
		prompt := args[0].(string)
		maxTokens := args[1].(float64)
		return map[string]interface{}{
			"prompt": prompt,
			"tokens": maxTokens,
		}, nil
	}

	// Simulated bridge method with ScriptValue
	executeNew := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
		if len(args) < 2 {
			return nil, nil
		}
		if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeNumber {
			return nil, nil
		}
		prompt := args[0].(engine.StringValue).Value()
		maxTokens := args[1].(engine.NumberValue).Value()
		return engine.NewObjectValue(map[string]engine.ScriptValue{
			"prompt": engine.NewStringValue(prompt),
			"tokens": engine.NewNumberValue(maxTokens),
		}), nil
	}

	b.Run("Interface_Method", func(b *testing.B) {
		args := []interface{}{"test prompt", float64(100)}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			executeOld(args)
		}
	})

	b.Run("ScriptValue_Method", func(b *testing.B) {
		args := []engine.ScriptValue{engine.NewStringValue("test prompt"), engine.NewNumberValue(100)}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			executeNew(args)
		}
	})
}

// Benchmark complex object conversion
func BenchmarkComplexObjectConversion(b *testing.B) {
	complexData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name":   "Alice",
				"age":    30,
				"active": true,
			},
			map[string]interface{}{
				"name":   "Bob",
				"age":    25,
				"active": false,
			},
		},
		"settings": map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"limits": map[string]interface{}{
				"max_requests": 1000,
				"timeout":      30,
			},
		},
	}

	b.Run("Interface_NoConversion", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Just pass through - no conversion needed
			_ = complexData
		}
	})

	b.Run("ScriptValue_Conversion", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Convert to ScriptValue
			sv := engine.ConvertToScriptValue(complexData)
			// Convert back
			_ = sv.ToGo()
		}
	})
}

// Benchmark array operations
func BenchmarkArrayOperations(b *testing.B) {
	size := 1000

	b.Run("Interface_Array", func(b *testing.B) {
		arr := make([]interface{}, size)
		for i := 0; i < size; i++ {
			arr[i] = i
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sum := 0.0
			for _, v := range arr {
				if num, ok := v.(int); ok {
					sum += float64(num)
				}
			}
		}
	})

	b.Run("ScriptValue_Array", func(b *testing.B) {
		elements := make([]engine.ScriptValue, size)
		for i := 0; i < size; i++ {
			elements[i] = engine.NewNumberValue(float64(i))
		}
		arr := engine.NewArrayValue(elements)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sum := 0.0
			arrVal := arr.(ArrayValue)
			for _, v := range arrVal.Elements() {
				if v.Type() == engine.TypeNumber {
					sum += v.(NumberValue).Value()
				}
			}
		}
	})
}

// Benchmark memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("Interface_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			values := make([]interface{}, 0, 10)
			values = append(values, "string")
			values = append(values, 42)
			values = append(values, true)
			values = append(values, map[string]interface{}{"key": "value"})
			values = append(values, []interface{}{1, 2, 3})
		}
	})

	b.Run("ScriptValue_Allocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			values := make([]engine.ScriptValue, 0, 10)
			values = append(values, engine.NewStringValue("string"))
			values = append(values, engine.NewNumberValue(42))
			values = append(values, engine.NewBoolValue(true))
			values = append(values, engine.NewObjectValue(map[string]engine.ScriptValue{"key": engine.NewStringValue("value")}))
			values = append(values, engine.NewArrayValue([]engine.ScriptValue{engine.NewNumberValue(1), engine.NewNumberValue(2), engine.NewNumberValue(3)}))
		}
	})
}

// Benchmark error handling
func BenchmarkErrorHandling(b *testing.B) {
	b.Run("Interface_PanicRecover", func(b *testing.B) {
		values := []interface{}{"string", 42, true}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, v := range values {
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Recovered from panic
						}
					}()
					// This might panic
					_ = v.(string)
				}()
			}
		}
	})

	b.Run("ScriptValue_SafeCheck", func(b *testing.B) {
		values := []engine.ScriptValue{engine.NewStringValue("string"), engine.NewNumberValue(42), engine.NewBoolValue(true)}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, v := range values {
				// Safe check, no panic possible
				if v.Type() == engine.TypeString {
					_ = v.(StringValue).Value()
				}
			}
		}
	})
}

// Benchmark real-world scenario: processing API response
func BenchmarkAPIResponseProcessing(b *testing.B) {
	// Simulated API response
	apiResponse := map[string]interface{}{
		"id":    "msg-123",
		"model": "gpt-4",
		"choices": []interface{}{
			map[string]interface{}{
				"text":          "Hello, world!",
				"index":         0,
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 5,
			"total_tokens":      15,
		},
	}

	b.Run("Interface_Processing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Extract data with type assertions
			choices := apiResponse["choices"].([]interface{})
			firstChoice := choices[0].(map[string]interface{})
			text := firstChoice["text"].(string)

			usage := apiResponse["usage"].(map[string]interface{})
			totalTokens := usage["total_tokens"].(int)

			_ = text
			_ = totalTokens
		}
	})

	b.Run("ScriptValue_Processing", func(b *testing.B) {
		sv := engine.ConvertToScriptValue(apiResponse)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Extract data with type-safe checks
			obj := sv.(ObjectValue)
			fields := obj.Fields()

			if choices, ok := fields["choices"]; ok && choices.Type() == engine.TypeArray {
				choicesArr := choices.(ArrayValue).Elements()
				if len(choicesArr) > 0 && choicesArr[0].Type() == engine.TypeObject {
					firstChoice := choicesArr[0].(ObjectValue).Fields()
					if text, ok := firstChoice["text"]; ok && text.Type() == engine.TypeString {
						_ = text.(StringValue).Value()
					}
				}
			}

			if usage, ok := fields["usage"]; ok && usage.Type() == engine.TypeObject {
				usageFields := usage.(ObjectValue).Fields()
				if tokens, ok := usageFields["total_tokens"]; ok && tokens.Type() == engine.TypeNumber {
					_ = int(tokens.(NumberValue).Value())
				}
			}
		}
	})
}
