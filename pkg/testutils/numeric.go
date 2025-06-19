// ABOUTME: Numeric conversion helpers for consistent type handling in tests
// ABOUTME: Provides safe and panic-based numeric converters for test value creation

package testutils

import (
	"fmt"
	"strconv"
)

// ToFloat64 safely converts various numeric types to float64
func ToFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// MustFloat64 converts to float64 or panics - useful for test data setup
func MustFloat64(v interface{}) float64 {
	result, err := ToFloat64(v)
	if err != nil {
		panic(fmt.Sprintf("MustFloat64: %v", err))
	}
	return result
}

// ToInt64 safely converts various numeric types to int64
func ToInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		if val > 9223372036854775807 { // max int64
			return 0, fmt.Errorf("uint64 value %d too large for int64", val)
		}
		return int64(val), nil
	case float64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// MustInt64 converts to int64 or panics - useful for test data setup
func MustInt64(v interface{}) int64 {
	result, err := ToInt64(v)
	if err != nil {
		panic(fmt.Sprintf("MustInt64: %v", err))
	}
	return result
}

// ToInt safely converts various numeric types to int
func ToInt(v interface{}) (int, error) {
	i64, err := ToInt64(v)
	if err != nil {
		return 0, err
	}

	// Check for overflow on 32-bit systems
	if i64 > 2147483647 || i64 < -2147483648 {
		return 0, fmt.Errorf("int64 value %d out of range for int", i64)
	}

	return int(i64), nil
}

// MustInt converts to int or panics - useful for test data setup
func MustInt(v interface{}) int {
	result, err := ToInt(v)
	if err != nil {
		panic(fmt.Sprintf("MustInt: %v", err))
	}
	return result
}

// IsNumeric checks if a value is of any numeric type
func IsNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

// NumericEqual compares two values for numeric equality, handling type conversions
func NumericEqual(a, b interface{}) bool {
	// If both are the same type, compare directly
	if a == b {
		return true
	}

	// Both values must be numeric for numeric comparison
	if !IsNumeric(a) || !IsNumeric(b) {
		return false
	}

	// Convert both to float64 for comparison
	aFloat, err1 := ToFloat64(a)
	bFloat, err2 := ToFloat64(b)

	if err1 != nil || err2 != nil {
		return false
	}

	return aFloat == bFloat
}

// MinFloat64 returns the minimum of two float64 values
func MinFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat64 returns the maximum of two float64 values
func MaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 returns the minimum of two int64 values
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two int64 values
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
