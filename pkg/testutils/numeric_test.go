// ABOUTME: Tests for numeric conversion helpers ensuring safe and panic-based conversions work correctly
// ABOUTME: Validates ToFloat64, MustFloat64, ToInt64, MustInt64 and related numeric utilities

package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		wantErr  bool
	}{
		{"float64", float64(42.5), 42.5, false},
		{"float32", float32(42.5), 42.5, false},
		{"int", int(42), 42.0, false},
		{"int8", int8(42), 42.0, false},
		{"int16", int16(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(42), 42.0, false},
		{"uint16", uint16(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"string_valid", "42.5", 42.5, false},
		{"string_invalid", "not_a_number", 0, true},
		{"unsupported_type", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToFloat64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMustFloat64(t *testing.T) {
	// Test successful conversion
	result := MustFloat64(42)
	assert.Equal(t, float64(42), result)

	result = MustFloat64("42.5")
	assert.Equal(t, 42.5, result)

	// Test panic on invalid input
	assert.Panics(t, func() {
		MustFloat64("invalid")
	})

	assert.Panics(t, func() {
		MustFloat64([]int{1, 2, 3})
	})
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		wantErr  bool
	}{
		{"int64", int64(42), 42, false},
		{"int", int(42), 42, false},
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64_valid", uint64(42), 42, false},
		{"uint64_overflow", uint64(9223372036854775808), 0, true}, // max int64 + 1
		{"float64", float64(42.7), 42, false},
		{"float32", float32(42.7), 42, false},
		{"string_valid", "42", 42, false},
		{"string_invalid", "not_a_number", 0, true},
		{"unsupported_type", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInt64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMustInt64(t *testing.T) {
	// Test successful conversion
	result := MustInt64(42)
	assert.Equal(t, int64(42), result)

	result = MustInt64("42")
	assert.Equal(t, int64(42), result)

	// Test panic on invalid input
	assert.Panics(t, func() {
		MustInt64("invalid")
	})

	assert.Panics(t, func() {
		MustInt64([]int{1, 2, 3})
	})
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		wantErr  bool
	}{
		{"int", int(42), 42, false},
		{"int64_valid", int64(42), 42, false},
		{"int64_overflow_positive", int64(2147483648), 0, true},  // max int32 + 1
		{"int64_overflow_negative", int64(-2147483649), 0, true}, // min int32 - 1
		{"string_valid", "42", 42, false},
		{"string_invalid", "not_a_number", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMustInt(t *testing.T) {
	// Test successful conversion
	result := MustInt(42)
	assert.Equal(t, 42, result)

	result = MustInt("42")
	assert.Equal(t, 42, result)

	// Test panic on invalid input
	assert.Panics(t, func() {
		MustInt("invalid")
	})

	assert.Panics(t, func() {
		MustInt(int64(2147483648)) // overflow
	})
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"int", int(42), true},
		{"int8", int8(42), true},
		{"int16", int16(42), true},
		{"int32", int32(42), true},
		{"int64", int64(42), true},
		{"uint", uint(42), true},
		{"uint8", uint8(42), true},
		{"uint16", uint16(42), true},
		{"uint32", uint32(42), true},
		{"uint64", uint64(42), true},
		{"float32", float32(42.5), true},
		{"float64", float64(42.5), true},
		{"string", "42", false},
		{"bool", true, false},
		{"slice", []int{1, 2, 3}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumericEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"same_int", 42, 42, true},
		{"int_float64", 42, 42.0, true},
		{"int8_int64", int8(42), int64(42), true},
		{"float32_float64", float32(42.5), float64(42.5), true},
		{"different_values", 42, 43, false},
		{"string_int", "42", 42, false}, // strings not converted
		{"non_numeric", "hello", "world", false},
		{"one_non_numeric", 42, "42", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NumericEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinMaxFloat64(t *testing.T) {
	assert.Equal(t, float64(1.5), MinFloat64(1.5, 2.5))
	assert.Equal(t, float64(2.5), MaxFloat64(1.5, 2.5))
	assert.Equal(t, float64(-2.5), MinFloat64(-1.5, -2.5))
	assert.Equal(t, float64(-1.5), MaxFloat64(-1.5, -2.5))
	assert.Equal(t, float64(42.0), MinFloat64(42.0, 42.0))
	assert.Equal(t, float64(42.0), MaxFloat64(42.0, 42.0))
}

func TestMinMaxInt64(t *testing.T) {
	assert.Equal(t, int64(1), MinInt64(1, 2))
	assert.Equal(t, int64(2), MaxInt64(1, 2))
	assert.Equal(t, int64(-2), MinInt64(-1, -2))
	assert.Equal(t, int64(-1), MaxInt64(-1, -2))
	assert.Equal(t, int64(42), MinInt64(42, 42))
	assert.Equal(t, int64(42), MaxInt64(42, 42))
}

func TestNumericConversionEdgeCases(t *testing.T) {
	// Test zero values
	result, err := ToFloat64(0)
	require.NoError(t, err)
	assert.Equal(t, float64(0), result)

	// Test negative values
	result, err = ToFloat64(-42)
	require.NoError(t, err)
	assert.Equal(t, float64(-42), result)

	// Test large values
	result, err = ToFloat64(uint64(18446744073709551615)) // max uint64
	require.NoError(t, err)
	assert.Equal(t, float64(18446744073709551615), result)
}

func TestStringNumericConversions(t *testing.T) {
	// Test various string formats
	tests := []struct {
		input    string
		expected float64
		wantErr  bool
	}{
		{"0", 0, false},
		{"-42", -42, false},
		{"42.5", 42.5, false},
		{"-42.5", -42.5, false},
		{"1.23e4", 12300, false},
		{"-1.23e-2", -0.0123, false},
		{"", 0, true},
		{" ", 0, true},
		{"abc", 0, true},
		{"42.5.5", 0, true},
	}

	for _, tt := range tests {
		t.Run("string_"+tt.input, func(t *testing.T) {
			result, err := ToFloat64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
