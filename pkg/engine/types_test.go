// ABOUTME: Test suite for common type representations and validation for the multi-engine system.
// ABOUTME: Validates type conversion utilities and validation frameworks used across all script engines.

package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock TypeAdapter for testing
type mockTypeAdapter struct {
	supportedTypes []string
}

func (m *mockTypeAdapter) ToNative(v interface{}) (interface{}, error) {
	// Simple conversion for testing
	if str, ok := v.(string); ok && str == "test-value" {
		return "native-test-value", nil
	}
	return nil, errors.New("unsupported conversion")
}

func (m *mockTypeAdapter) FromNative(v interface{}) (interface{}, error) {
	// Simple conversion for testing
	if str, ok := v.(string); ok && str == "native-test-value" {
		return "test-value", nil
	}
	return nil, errors.New("unsupported conversion")
}

func (m *mockTypeAdapter) SupportsType(typeName string) bool {
	for _, t := range m.supportedTypes {
		if t == typeName {
			return true
		}
	}
	return false
}

// Test struct for conversion testing
type testStruct struct {
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Active    bool      `json:"active"`
	Score     float64   `json:"score"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	private   string    // Should be ignored
}

// Tests for BaseTypeConverter
func TestBaseTypeConverter(t *testing.T) {
	t.Run("NewBaseTypeConverter", func(t *testing.T) {
		converter := NewBaseTypeConverter("test-engine")
		assert.NotNil(t, converter)
		assert.Equal(t, "test-engine", converter.engineName)
		assert.NotNil(t, converter.adapters)
	})

	t.Run("RegisterAdapter", func(t *testing.T) {
		converter := NewBaseTypeConverter("test")
		adapter := &mockTypeAdapter{
			supportedTypes: []string{"custom"},
		}

		converter.RegisterAdapter("custom", adapter)
		
		// Test that the adapter was registered
		assert.True(t, converter.SupportsType("custom"))
	})
}

// Tests for ToBoolean conversion
func TestToBoolean(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name     string
		input    interface{}
		expected bool
		wantErr  bool
	}{
		// Boolean inputs
		{"bool true", true, true, false},
		{"bool false", false, false, false},

		// String inputs
		{"string true", "true", true, false},
		{"string True", "True", true, false},
		{"string yes", "yes", true, false},
		{"string 1", "1", true, false},
		{"string on", "on", true, false},
		{"string false", "false", false, false},
		{"string False", "False", false, false},
		{"string no", "no", false, false},
		{"string 0", "0", false, false},
		{"string off", "off", false, false},
		{"string empty", "", false, false},
		{"string invalid", "invalid", false, true},

		// Numeric inputs
		{"int 0", 0, false, false},
		{"int 1", 1, true, false},
		{"int -1", -1, true, false},
		{"int64 0", int64(0), false, false},
		{"int64 42", int64(42), true, false},
		{"uint 0", uint(0), false, false},
		{"uint 1", uint(1), true, false},
		{"float32 0", float32(0), false, false},
		{"float32 1.5", float32(1.5), true, false},
		{"float64 0", float64(0), false, false},
		{"float64 -0.5", float64(-0.5), true, false},

		// Special cases
		{"nil", nil, false, false},
		{"unsupported type", struct{}{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToBoolean(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Tests for ToNumber conversion
func TestToNumber(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name     string
		input    interface{}
		expected float64
		wantErr  bool
	}{
		// Float inputs
		{"float64", float64(3.14), 3.14, false},
		{"float32", float32(2.5), 2.5, false},

		// Integer inputs
		{"int", 42, 42.0, false},
		{"int8", int8(127), 127.0, false},
		{"int16", int16(1000), 1000.0, false},
		{"int32", int32(100000), 100000.0, false},
		{"int64", int64(1000000), 1000000.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(255), 255.0, false},
		{"uint16", uint16(65535), 65535.0, false},
		{"uint32", uint32(100000), 100000.0, false},
		{"uint64", uint64(1000000), 1000000.0, false},

		// String inputs
		{"string number", "123.45", 123.45, false},
		{"string integer", "42", 42.0, false},
		{"string negative", "-10.5", -10.5, false},
		{"string empty", "", 0.0, false},
		{"string invalid", "not a number", 0, true},

		// Boolean inputs
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},

		// Special cases
		{"nil", nil, 0.0, false},
		{"unsupported type", struct{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToNumber(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Tests for ToString conversion
func TestToString(t *testing.T) {
	converter := NewBaseTypeConverter("test")
	now := time.Now()

	tests := []struct {
		name     string
		input    interface{}
		expected string
		checkFunc func(string) bool
	}{
		// String input
		{"string", "hello", "hello", nil},

		// Boolean inputs
		{"bool true", true, "true", nil},
		{"bool false", false, "false", nil},

		// Numeric inputs
		{"int", 42, "42", nil},
		{"int8", int8(-128), "-128", nil},
		{"int16", int16(1000), "1000", nil},
		{"int32", int32(100000), "100000", nil},
		{"int64", int64(1000000), "1000000", nil},
		{"uint", uint(42), "42", nil},
		{"uint8", uint8(255), "255", nil},
		{"uint16", uint16(65535), "65535", nil},
		{"uint32", uint32(100000), "100000", nil},
		{"uint64", uint64(1000000), "1000000", nil},
		{"float32", float32(3.14), "3.14", nil},
		{"float64", float64(2.71828), "2.71828", nil},

		// Time input
		{"time", now, now.Format(time.RFC3339), nil},

		// Special cases
		{"nil", nil, "", nil},
		{"struct", struct{ Name string }{Name: "test"}, "", func(s string) bool {
			return s == "{test}" || s == "{Name:test}"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToString(tt.input)
			assert.NoError(t, err)
			
			if tt.checkFunc != nil {
				assert.True(t, tt.checkFunc(result), "Result: %s", result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Tests for ToArray conversion
func TestToArray(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		// Already []interface{}
		{"interface slice", []interface{}{"a", 1, true}, []interface{}{"a", 1, true}, false},

		// Type-specific slices
		{"string slice", []string{"a", "b", "c"}, []interface{}{"a", "b", "c"}, false},
		{"int slice", []int{1, 2, 3}, []interface{}{1, 2, 3}, false},
		{"float64 slice", []float64{1.1, 2.2, 3.3}, []interface{}{1.1, 2.2, 3.3}, false},
		{"bool slice", []bool{true, false, true}, []interface{}{true, false, true}, false},

		// Array (via reflection)
		{"array", [3]int{1, 2, 3}, []interface{}{1, 2, 3}, false},

		// Special cases
		{"nil", nil, nil, false},
		{"not slice", "not a slice", nil, true},
		{"struct", struct{}{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToArray(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Tests for ToMap conversion
func TestToMap(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		// Already map[string]interface{}
		{
			"map string interface",
			map[string]interface{}{"key": "value", "num": 42},
			map[string]interface{}{"key": "value", "num": 42},
			false,
		},

		// map[string]string
		{
			"map string string",
			map[string]string{"key": "value", "name": "test"},
			map[string]interface{}{"key": "value", "name": "test"},
			false,
		},

		// map[interface{}]interface{}
		{
			"map interface interface",
			map[interface{}]interface{}{"key": "value", 123: "number"},
			map[string]interface{}{"key": "value", "123": "number"},
			false,
		},

		// Struct conversion
		{
			"struct with json tags",
			testStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Score:  95.5,
				Tags:   []string{"go", "test"},
			},
			map[string]interface{}{
				"name":       "John",
				"age":        30,
				"active":     true,
				"score":      95.5,
				"tags":       []string{"go", "test"},
				"created_at": time.Time{},
				"metadata":   map[string]interface{}(nil),
			},
			false,
		},

		// Pointer to struct
		{
			"pointer to struct",
			&testStruct{Name: "Jane", Age: 25},
			map[string]interface{}{
				"name":       "Jane",
				"age":        25,
				"active":     false,
				"score":      float64(0),
				"tags":       []string(nil),
				"created_at": time.Time{},
				"metadata":   map[string]interface{}(nil),
			},
			false,
		},

		// Special cases
		{"nil", nil, nil, false},
		{"not a map", []int{1, 2, 3}, nil, true},
		{"int", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToMap(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Tests for ToStruct conversion
func TestToStruct(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	t.Run("Valid conversion", func(t *testing.T) {
		source := map[string]interface{}{
			"name":   "Alice",
			"age":    28,
			"active": true,
			"score":  88.5,
			"tags":   []interface{}{"developer", "golang"},
		}

		var target testStruct
		err := converter.ToStruct(source, &target)
		
		assert.NoError(t, err)
		assert.Equal(t, "Alice", target.Name)
		assert.Equal(t, 28, target.Age)
		assert.True(t, target.Active)
		assert.Equal(t, 88.5, target.Score)
		assert.Equal(t, []string{"developer", "golang"}, target.Tags)
	})

	t.Run("Nil source", func(t *testing.T) {
		var target testStruct
		err := converter.ToStruct(nil, &target)
		assert.NoError(t, err)
	})

	t.Run("Non-pointer target", func(t *testing.T) {
		source := map[string]interface{}{"name": "Bob"}
		var target testStruct
		err := converter.ToStruct(source, target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target must be a pointer")
	})

	t.Run("Invalid source type", func(t *testing.T) {
		var target testStruct
		err := converter.ToStruct("not a map", &target)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert source to map")
	})

	t.Run("Type conversion in fields", func(t *testing.T) {
		source := map[string]interface{}{
			"name":   "Charlie",
			"age":    "35", // String that should convert to int
			"active": "true", // String that should convert to bool
			"score":  "92.3", // String that should convert to float
		}

		var target testStruct
		err := converter.ToStruct(source, &target)
		
		assert.NoError(t, err)
		assert.Equal(t, "Charlie", target.Name)
		assert.Equal(t, 35, target.Age)
		assert.True(t, target.Active)
		assert.Equal(t, 92.3, target.Score)
	})
}

// Tests for Function conversions
func TestFunctionConversions(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	t.Run("ToFunction without adapter", func(t *testing.T) {
		// Test with non-function type
		_, err := converter.ToFunction("not a function")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("ToFunction with adapter", func(t *testing.T) {
		// Register a mock adapter
		adapter := &mockTypeAdapter{
			supportedTypes: []string{"string"},
		}
		converter.RegisterAdapter("string", adapter)

		// This should still fail as our mock doesn't return a Function
		_, err := converter.ToFunction("test-value")
		assert.Error(t, err)
	})

	t.Run("FromFunction without adapter", func(t *testing.T) {
		// Create a mock function
		mockFn := &mockFunction{}
		
		result, err := converter.FromFunction(mockFn)
		assert.NoError(t, err)
		assert.Equal(t, mockFn, result) // Should return as-is
	})

	t.Run("FromFunction with adapter", func(t *testing.T) {
		// Register adapter for functions
		adapter := &mockTypeAdapter{
			supportedTypes: []string{"function"},
		}
		converter.RegisterAdapter("function", adapter)

		mockFn := &mockFunction{}
		_, err := converter.FromFunction(mockFn)
		assert.Error(t, err) // Our mock adapter doesn't handle functions properly
	})
}

// Mock Function implementation
type mockFunction struct{}

func (m *mockFunction) Call(args ...interface{}) (interface{}, error) {
	return "result", nil
}

func (m *mockFunction) Bind(thisArg interface{}) Function {
	return m
}

func (m *mockFunction) GetSignature() FunctionSignature {
	return FunctionSignature{
		Name:       "mockFunction",
		ReturnType: "string",
	}
}

// Tests for type support checking
func TestSupportsType(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	// Test built-in types
	builtinTypes := []string{
		"bool", "string", "int", "float64", "[]interface{}", "map[string]interface{}",
		"int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "time.Time",
	}

	for _, typeName := range builtinTypes {
		t.Run("builtin "+typeName, func(t *testing.T) {
			assert.True(t, converter.SupportsType(typeName))
		})
	}

	// Test unsupported type
	t.Run("unsupported type", func(t *testing.T) {
		assert.False(t, converter.SupportsType("custom"))
	})

	// Test with registered adapter
	t.Run("type with adapter", func(t *testing.T) {
		adapter := &mockTypeAdapter{
			supportedTypes: []string{"custom"},
		}
		converter.RegisterAdapter("custom", adapter)
		
		assert.True(t, converter.SupportsType("custom"))
	})
}

// Tests for GetTypeInfo
func TestGetTypeInfo(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name         string
		typeName     string
		expectedCat  TypeCategory
		expectedDesc string
	}{
		{"bool", "bool", TypeCategoryPrimitive, "Boolean value (true/false)"},
		{"string", "string", TypeCategoryPrimitive, "Text string value"},
		{"number", "number", TypeCategoryPrimitive, "Numeric value (integer or floating point)"},
		{"float64", "float64", TypeCategoryPrimitive, "Numeric value (integer or floating point)"},
		{"array", "array", TypeCategoryArray, "Ordered collection of values"},
		{"[]interface{}", "[]interface{}", TypeCategoryArray, "Ordered collection of values"},
		{"object", "object", TypeCategoryObject, "Key-value collection of properties"},
		{"map[string]interface{}", "map[string]interface{}", TypeCategoryObject, "Key-value collection of properties"},
		{"function", "function", TypeCategoryFunction, "Callable function with parameters and return value"},
		{"custom", "customType", TypeCategoryCustom, "Custom type specific to test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := converter.GetTypeInfo(tt.typeName)
			// GetTypeInfo normalizes type names
			switch tt.typeName {
			case "float64":
				assert.Equal(t, "number", info.Name)
			case "[]interface{}":
				assert.Equal(t, "array", info.Name)
			case "map[string]interface{}":
				assert.Equal(t, "object", info.Name)
			case "customType":
				assert.Equal(t, tt.typeName, info.Name)
			default:
				assert.Equal(t, tt.typeName, info.Name)
			}
			assert.Equal(t, tt.expectedCat, info.Category)
			assert.Equal(t, tt.expectedDesc, info.Description)
		})
	}
}

// Tests for ValidateType
func TestValidateType(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name         string
		value        interface{}
		expectedType string
		wantErr      bool
	}{
		// Boolean validation
		{"bool value for bool", true, "bool", false},
		{"bool value for boolean", false, "boolean", false},
		{"string for bool", "true", "bool", false},
		{"invalid for bool", struct{}{}, "bool", true},

		// String validation
		{"string for string", "hello", "string", false},
		{"number for string", 42, "string", false},
		{"struct for string", struct{}{}, "string", false},

		// Number validation
		{"int for number", 42, "number", false},
		{"float for float64", 3.14, "float64", false},
		{"int for int", 100, "int", false},
		{"string number for number", "123", "number", false},
		{"invalid string for number", "abc", "number", true},

		// Array validation
		{"slice for array", []interface{}{1, 2, 3}, "array", false},
		{"slice for []interface{}", []string{"a", "b"}, "[]interface{}", false},
		{"non-slice for array", "not array", "array", true},

		// Object validation
		{"map for object", map[string]interface{}{"key": "value"}, "object", false},
		{"map for map", map[string]string{"key": "value"}, "map", false},
		{"struct for object", testStruct{Name: "test"}, "map[string]interface{}", false},
		{"slice for object", []int{1, 2}, "object", true},

		// Function validation
		{"function for function", &mockFunction{}, "function", false},
		{"non-function for function", "not func", "function", true},

		// Nil validation
		{"nil for bool", nil, "bool", false},
		{"nil for string", nil, "string", false},
		{"nil for number", nil, "number", false},
		{"nil for array", nil, "array", false},
		{"nil for object", nil, "object", false},

		// Unsupported type
		{"value for unsupported", "value", "unsupported", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ValidateType(tt.value, tt.expectedType)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Tests for GetConversionPath
func TestGetConversionPath(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	tests := []struct {
		name     string
		fromType string
		toType   string
		expected []string
		wantErr  bool
	}{
		// Direct conversions
		{"string to bool", "string", "bool", []string{"string", "bool"}, false},
		{"string to number", "string", "number", []string{"string", "number"}, false},
		{"string to array", "string", "array", []string{"string", "array"}, false},
		{"number to bool", "number", "bool", []string{"number", "bool"}, false},
		{"number to string", "number", "string", []string{"number", "string"}, false},
		{"bool to string", "bool", "string", []string{"bool", "string"}, false},
		{"bool to number", "bool", "number", []string{"bool", "number"}, false},
		{"array to string", "array", "string", []string{"array", "string"}, false},
		{"object to string", "object", "string", []string{"object", "string"}, false},

		// No conversion path
		{"array to bool", "array", "bool", nil, true},
		{"object to number", "object", "number", nil, true},
		{"bool to array", "bool", "array", nil, true},
		{"custom to custom2", "custom", "custom2", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := converter.GetConversionPath(tt.fromType, tt.toType)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no conversion path")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, path)
			}
		})
	}
}

// Tests for complex nested conversions
func TestComplexConversions(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	t.Run("Nested map to struct", func(t *testing.T) {
		type NestedStruct struct {
			Inner struct {
				Value string `json:"value"`
				Count int    `json:"count"`
			} `json:"inner"`
			Items []string `json:"items"`
		}

		source := map[string]interface{}{
			"inner": map[string]interface{}{
				"value": "nested",
				"count": 42,
			},
			"items": []interface{}{"a", "b", "c"},
		}

		var target NestedStruct
		err := converter.ToStruct(source, &target)
		
		require.NoError(t, err)
		assert.Equal(t, "nested", target.Inner.Value)
		assert.Equal(t, 42, target.Inner.Count)
		assert.Equal(t, []string{"a", "b", "c"}, target.Items)
	})

	t.Run("Struct with map field", func(t *testing.T) {
		source := map[string]interface{}{
			"name": "Test",
			"metadata": map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		}

		var target testStruct
		err := converter.ToStruct(source, &target)
		
		require.NoError(t, err)
		assert.Equal(t, "Test", target.Name)
		assert.Equal(t, "value1", target.Metadata["key1"])
		assert.Equal(t, 123, target.Metadata["key2"])
	})

	t.Run("Array of mixed types", func(t *testing.T) {
		input := []interface{}{
			"string",
			42,
			true,
			3.14,
			[]interface{}{1, 2, 3},
			map[string]interface{}{"key": "value"},
		}

		result, err := converter.ToArray(input)
		
		require.NoError(t, err)
		assert.Len(t, result, 6)
		assert.Equal(t, "string", result[0])
		assert.Equal(t, 42, result[1])
		assert.Equal(t, true, result[2])
		assert.Equal(t, 3.14, result[3])
		assert.Equal(t, []interface{}{1, 2, 3}, result[4])
		assert.Equal(t, map[string]interface{}{"key": "value"}, result[5])
	})
}

// Tests for error handling
func TestTypeConverterErrorHandling(t *testing.T) {
	converter := NewBaseTypeConverter("test")

	t.Run("Invalid string to number", func(t *testing.T) {
		_, err := converter.ToNumber("not a number")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert string")
	})

	t.Run("Invalid type to boolean", func(t *testing.T) {
		_, err := converter.ToBoolean(make(chan int))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("Invalid type to array", func(t *testing.T) {
		_, err := converter.ToArray(42)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("Invalid type to map", func(t *testing.T) {
		_, err := converter.ToMap("not a map")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("Invalid map key type", func(t *testing.T) {
		// Skip this test as ToString now handles all types via fmt.Sprintf
		t.Skip("ToString handles all types via fmt.Sprintf, so this test is no longer valid")
	})
}

// Benchmark tests
func BenchmarkToBoolean(b *testing.B) {
	converter := NewBaseTypeConverter("bench")
	
	b.Run("bool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = converter.ToBoolean(true)
		}
	})
	
	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = converter.ToBoolean("true")
		}
	})
	
	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = converter.ToBoolean(1)
		}
	})
}

func BenchmarkToMap(b *testing.B) {
	converter := NewBaseTypeConverter("bench")
	testData := testStruct{
		Name:   "Benchmark",
		Age:    30,
		Active: true,
		Score:  95.5,
		Tags:   []string{"test", "bench"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.ToMap(testData)
	}
}

func BenchmarkToStruct(b *testing.B) {
	converter := NewBaseTypeConverter("bench")
	source := map[string]interface{}{
		"name":   "Benchmark",
		"age":    30,
		"active": true,
		"score":  95.5,
		"tags":   []interface{}{"test", "bench"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var target testStruct
		_ = converter.ToStruct(source, &target)
	}
}