// ABOUTME: Tests for primitive type conversion handlers - specialized bool, number, string converters
// ABOUTME: Validates type validation, error reporting, and edge cases for primitive types

package gopherlua

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestPrimitiveConverter_BoolConversions(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name        string
		input       interface{}
		expected    bool
		expectError bool
	}{
		{
			name:     "bool_true_to_bool",
			input:    true,
			expected: true,
		},
		{
			name:     "bool_false_to_bool",
			input:    false,
			expected: false,
		},
		{
			name:     "string_true_to_bool",
			input:    "true",
			expected: true,
		},
		{
			name:     "string_false_to_bool",
			input:    "false",
			expected: false,
		},
		{
			name:     "string_yes_to_bool",
			input:    "yes",
			expected: true,
		},
		{
			name:     "string_no_to_bool",
			input:    "no",
			expected: false,
		},
		{
			name:     "string_1_to_bool",
			input:    "1",
			expected: true,
		},
		{
			name:     "string_0_to_bool",
			input:    "0",
			expected: false,
		},
		{
			name:     "empty_string_to_bool",
			input:    "",
			expected: false,
		},
		{
			name:     "non_empty_string_to_bool",
			input:    "hello",
			expected: true,
		},
		{
			name:     "int_zero_to_bool",
			input:    0,
			expected: false,
		},
		{
			name:     "int_nonzero_to_bool",
			input:    42,
			expected: true,
		},
		{
			name:     "float_zero_to_bool",
			input:    0.0,
			expected: false,
		},
		{
			name:     "float_nonzero_to_bool",
			input:    3.14,
			expected: true,
		},
		{
			name:     "nil_to_bool",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToBool(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPrimitiveConverter_BoolToLua(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "true_to_LBool",
			input:    true,
			expected: true,
		},
		{
			name:     "false_to_LBool",
			input:    false,
			expected: false,
		},
		{
			name:     "string_true_to_LBool",
			input:    "true",
			expected: true,
		},
		{
			name:     "int_1_to_LBool",
			input:    1,
			expected: true,
		},
		{
			name:     "int_0_to_LBool",
			input:    0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.BoolToLua(L, tt.input)
			require.NoError(t, err)

			lbool, ok := result.(lua.LBool)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bool(lbool))
		})
	}
}

func TestPrimitiveConverter_NumberConversions(t *testing.T) {
	converter := NewPrimitiveConverter()

	tests := []struct {
		name        string
		input       interface{}
		expected    float64
		expectError bool
		errorText   string
	}{
		{
			name:     "int_to_number",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "int8_to_number",
			input:    int8(127),
			expected: 127.0,
		},
		{
			name:     "int16_to_number",
			input:    int16(32767),
			expected: 32767.0,
		},
		{
			name:     "int32_to_number",
			input:    int32(2147483647),
			expected: 2147483647.0,
		},
		{
			name:     "int64_to_number",
			input:    int64(9223372036854775807),
			expected: 9223372036854775807.0,
		},
		{
			name:     "uint_to_number",
			input:    uint(42),
			expected: 42.0,
		},
		{
			name:     "uint8_to_number",
			input:    uint8(255),
			expected: 255.0,
		},
		{
			name:     "uint16_to_number",
			input:    uint16(65535),
			expected: 65535.0,
		},
		{
			name:     "uint32_to_number",
			input:    uint32(4294967295),
			expected: 4294967295.0,
		},
		{
			name:     "uint64_to_number",
			input:    uint64(18446744073709551615),
			expected: 18446744073709551615.0,
		},
		{
			name:     "float32_to_number",
			input:    float32(3.14),
			expected: float64(float32(3.14)), // Account for precision loss
		},
		{
			name:     "float64_to_number",
			input:    3.141592653589793,
			expected: 3.141592653589793,
		},
		{
			name:     "string_number_to_number",
			input:    "42.5",
			expected: 42.5,
		},
		{
			name:     "string_int_to_number",
			input:    "123",
			expected: 123.0,
		},
		{
			name:     "string_negative_to_number",
			input:    "-456.78",
			expected: -456.78,
		},
		{
			name:     "string_scientific_to_number",
			input:    "1.23e-4",
			expected: 1.23e-4,
		},
		{
			name:        "string_invalid_to_number",
			input:       "not a number",
			expectError: true,
			errorText:   "cannot convert",
		},
		{
			name:        "string_empty_to_number",
			input:       "",
			expectError: true,
			errorText:   "cannot convert",
		},
		{
			name:     "bool_true_to_number",
			input:    true,
			expected: 1.0,
		},
		{
			name:     "bool_false_to_number",
			input:    false,
			expected: 0.0,
		},
		{
			name:     "nil_to_number",
			input:    nil,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToNumber(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				require.NoError(t, err)
				if math.IsNaN(tt.expected) {
					assert.True(t, math.IsNaN(result))
				} else if math.IsInf(tt.expected, 0) {
					assert.True(t, math.IsInf(result, int(math.Copysign(1, tt.expected))))
				} else {
					assert.InDelta(t, tt.expected, result, 1e-10)
				}
			}
		})
	}
}

func TestPrimitiveConverter_NumberToLua(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{
			name:     "int_to_LNumber",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "float_to_LNumber",
			input:    3.14159,
			expected: 3.14159,
		},
		{
			name:     "string_number_to_LNumber",
			input:    "123.45",
			expected: 123.45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.NumberToLua(L, tt.input)
			require.NoError(t, err)

			lnum, ok := result.(lua.LNumber)
			require.True(t, ok)
			assert.InDelta(t, tt.expected, float64(lnum), 1e-10)
		})
	}
}

func TestPrimitiveConverter_StringConversions(t *testing.T) {
	converter := NewPrimitiveConverter()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string_to_string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty_string_to_string",
			input:    "",
			expected: "",
		},
		{
			name:     "int_to_string",
			input:    42,
			expected: "42",
		},
		{
			name:     "float_to_string",
			input:    3.14159,
			expected: "3.14159",
		},
		{
			name:     "bool_true_to_string",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool_false_to_string",
			input:    false,
			expected: "false",
		},
		{
			name:     "nil_to_string",
			input:    nil,
			expected: "",
		},
		{
			name:     "complex_type_to_string",
			input:    []int{1, 2, 3},
			expected: "[1 2 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToString(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrimitiveConverter_StringToLua(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string_to_LString",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int_to_LString",
			input:    123,
			expected: "123",
		},
		{
			name:     "bool_to_LString",
			input:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.StringToLua(L, tt.input)
			require.NoError(t, err)

			lstr, ok := result.(lua.LString)
			require.True(t, ok)
			assert.Equal(t, tt.expected, string(lstr))
		})
	}
}

func TestPrimitiveConverter_NilHandling(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("nil_to_LNil", func(t *testing.T) {
		result, err := converter.NilToLua(L, nil)
		require.NoError(t, err)
		assert.Equal(t, lua.LNil, result)
	})

	t.Run("non_nil_to_LNil_error", func(t *testing.T) {
		_, err := converter.NilToLua(L, "not nil")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected nil")
	})

	t.Run("is_nil_check_true", func(t *testing.T) {
		assert.True(t, converter.IsNil(nil))
	})

	t.Run("is_nil_check_false", func(t *testing.T) {
		assert.False(t, converter.IsNil("not nil"))
		assert.False(t, converter.IsNil(0))
		assert.False(t, converter.IsNil(false))
		assert.False(t, converter.IsNil(""))
	})
}

func TestPrimitiveConverter_TypeValidation(t *testing.T) {
	converter := NewPrimitiveConverter()

	tests := []struct {
		name     string
		input    interface{}
		isBool   bool
		isNumber bool
		isString bool
		isNil    bool
	}{
		{
			name:     "bool_value",
			input:    true,
			isBool:   true,
			isNumber: false,
			isString: false,
			isNil:    false,
		},
		{
			name:     "int_value",
			input:    42,
			isBool:   false,
			isNumber: true,
			isString: false,
			isNil:    false,
		},
		{
			name:     "float_value",
			input:    3.14,
			isBool:   false,
			isNumber: true,
			isString: false,
			isNil:    false,
		},
		{
			name:     "string_value",
			input:    "hello",
			isBool:   false,
			isNumber: false,
			isString: true,
			isNil:    false,
		},
		{
			name:     "nil_value",
			input:    nil,
			isBool:   false,
			isNumber: false,
			isString: false,
			isNil:    true,
		},
		{
			name:     "complex_value",
			input:    []int{1, 2, 3},
			isBool:   false,
			isNumber: false,
			isString: false,
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isBool, converter.IsBool(tt.input), "IsBool check")
			assert.Equal(t, tt.isNumber, converter.IsNumber(tt.input), "IsNumber check")
			assert.Equal(t, tt.isString, converter.IsString(tt.input), "IsString check")
			assert.Equal(t, tt.isNil, converter.IsNil(tt.input), "IsNil check")
		})
	}
}

func TestPrimitiveConverter_ErrorReporting(t *testing.T) {
	converter := NewPrimitiveConverter()

	tests := []struct {
		name        string
		operation   func() error
		expectError bool
		errorText   string
	}{
		{
			name: "invalid_number_conversion",
			operation: func() error {
				_, err := converter.ToNumber("not a number")
				return err
			},
			expectError: true,
			errorText:   "cannot convert",
		},
		{
			name: "nil_to_string_with_validation",
			operation: func() error {
				_, err := converter.ToStringStrict(nil)
				return err
			},
			expectError: true,
			errorText:   "cannot convert nil",
		},
		{
			name: "type_mismatch_validation",
			operation: func() error {
				return converter.ValidateType("string", 42)
			},
			expectError: true,
			errorText:   "expected string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrimitiveConverter_EdgeCases(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("very_large_numbers", func(t *testing.T) {
		largeInt := int64(9223372036854775807) // Max int64
		result, err := converter.ToNumber(largeInt)
		require.NoError(t, err)
		assert.Equal(t, float64(largeInt), result)
	})

	t.Run("very_small_numbers", func(t *testing.T) {
		smallFloat := 1e-15
		result, err := converter.ToNumber(smallFloat)
		require.NoError(t, err)
		assert.Equal(t, smallFloat, result)
	})

	t.Run("unicode_strings", func(t *testing.T) {
		unicode := "Hello 世界 🌍"
		result, err := converter.ToString(unicode)
		require.NoError(t, err)
		assert.Equal(t, unicode, result)

		luaResult, err := converter.StringToLua(L, unicode)
		require.NoError(t, err)
		assert.Equal(t, unicode, string(luaResult.(lua.LString)))
	})

	t.Run("special_float_values", func(t *testing.T) {
		// Test NaN
		nan := math.NaN()
		result, err := converter.ToNumber(nan)
		require.NoError(t, err)
		assert.True(t, math.IsNaN(result))

		// Test Infinity
		inf := math.Inf(1)
		result, err = converter.ToNumber(inf)
		require.NoError(t, err)
		assert.True(t, math.IsInf(result, 1))

		// Test Negative Infinity
		negInf := math.Inf(-1)
		result, err = converter.ToNumber(negInf)
		require.NoError(t, err)
		assert.True(t, math.IsInf(result, -1))
	})
}

func TestPrimitiveConverter_Performance(t *testing.T) {
	converter := NewPrimitiveConverter()
	L := lua.NewState()
	defer L.Close()

	t.Run("bulk_conversions", func(t *testing.T) {
		// Test performance with many conversions
		const numConversions = 1000

		for i := 0; i < numConversions; i++ {
			// Bool conversions
			_, err := converter.ToBool(i%2 == 0)
			require.NoError(t, err)

			// Number conversions
			_, err = converter.ToNumber(float64(i) * 3.14159)
			require.NoError(t, err)

			// String conversions
			_, err = converter.ToString(i)
			require.NoError(t, err)
		}
	})
}
