// ABOUTME: PrimitiveConverter provides specialized conversion handlers for primitive types (bool, number, string, nil)
// ABOUTME: Offers granular control over type validation, error reporting, and conversion behavior for basic types

package gopherlua

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// PrimitiveConverter handles conversion of primitive types with specialized validation and error handling
type PrimitiveConverter struct {
	// Configuration options can be added here in the future
}

// NewPrimitiveConverter creates a new primitive type converter
func NewPrimitiveConverter() *PrimitiveConverter {
	return &PrimitiveConverter{}
}

// Bool conversion methods

// ToBool converts any value to a Go boolean with comprehensive type handling
func (pc *PrimitiveConverter) ToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), nil

	case reflect.String:
		str := strings.ToLower(strings.TrimSpace(rv.String()))
		switch str {
		case "true", "yes", "1":
			return true, nil
		case "false", "no", "0", "":
			return false, nil
		default:
			// Non-empty strings that aren't explicit false values are truthy
			return str != "", nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0, nil

	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		// Handle special cases
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return true, nil // NaN and Inf are truthy
		}
		return f != 0.0, nil

	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0, nil

	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return false, nil
		}
		return pc.ToBool(rv.Elem().Interface())

	default:
		// Non-nil values of unsupported types are truthy
		return true, nil
	}
}

// BoolToLua converts a value to Lua LBool
func (pc *PrimitiveConverter) BoolToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	boolVal, err := pc.ToBool(value)
	if err != nil {
		return nil, err
	}
	return lua.LBool(boolVal), nil
}

// Number conversion methods

// ToNumber converts any value to a Go float64 with comprehensive type handling
func (pc *PrimitiveConverter) ToNumber(value interface{}) (float64, error) {
	if value == nil {
		return 0.0, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil

	case reflect.String:
		str := strings.TrimSpace(rv.String())
		if str == "" {
			return 0, fmt.Errorf("cannot convert empty string to number")
		}

		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to number: %w", str, err)
		}
		return num, nil

	case reflect.Bool:
		if rv.Bool() {
			return 1.0, nil
		}
		return 0.0, nil

	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return 0.0, nil
		}
		return pc.ToNumber(rv.Elem().Interface())

	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}

// NumberToLua converts a value to Lua LNumber
func (pc *PrimitiveConverter) NumberToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	numVal, err := pc.ToNumber(value)
	if err != nil {
		return nil, err
	}
	return lua.LNumber(numVal), nil
}

// String conversion methods

// ToString converts any value to a Go string with comprehensive formatting
func (pc *PrimitiveConverter) ToString(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), nil

	case reflect.Bool:
		if rv.Bool() {
			return "true", nil
		}
		return "false", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), nil

	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		// Handle special values
		if math.IsNaN(f) {
			return "NaN", nil
		}
		if math.IsInf(f, 1) {
			return "+Inf", nil
		}
		if math.IsInf(f, -1) {
			return "-Inf", nil
		}
		return strconv.FormatFloat(f, 'g', -1, 64), nil

	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return "", nil
		}
		return pc.ToString(rv.Elem().Interface())

	default:
		// For complex types, use Go's default string representation
		return fmt.Sprintf("%v", value), nil
	}
}

// StringToLua converts a value to Lua LString
func (pc *PrimitiveConverter) StringToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	strVal, err := pc.ToString(value)
	if err != nil {
		return nil, err
	}
	return lua.LString(strVal), nil
}

// Nil handling methods

// IsNil checks if a value is nil
func (pc *PrimitiveConverter) IsNil(value interface{}) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}

// NilToLua converts nil values to Lua LNil with validation
func (pc *PrimitiveConverter) NilToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	if !pc.IsNil(value) {
		return nil, fmt.Errorf("expected nil value, got %T", value)
	}
	return lua.LNil, nil
}

// Type validation methods

// IsBool checks if a value is a boolean type
func (pc *PrimitiveConverter) IsBool(value interface{}) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Bool
}

// IsNumber checks if a value is a numeric type
func (pc *PrimitiveConverter) IsNumber(value interface{}) bool {
	if value == nil {
		return false
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsString checks if a value is a string type
func (pc *PrimitiveConverter) IsString(value interface{}) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.String
}

// Advanced string conversion with strict validation

// ToStringStrict converts a value to string but rejects nil values
func (pc *PrimitiveConverter) ToStringStrict(value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("cannot convert nil to string in strict mode")
	}
	return pc.ToString(value)
}

// Type validation method

// ValidateType validates that a value matches the expected type
func (pc *PrimitiveConverter) ValidateType(expectedType string, value interface{}) error {
	switch expectedType {
	case "bool":
		if !pc.IsBool(value) {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "number":
		if !pc.IsNumber(value) {
			return fmt.Errorf("expected number, got %T", value)
		}
	case "string":
		if !pc.IsString(value) {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "nil":
		if !pc.IsNil(value) {
			return fmt.Errorf("expected nil, got %T", value)
		}
	default:
		return fmt.Errorf("unknown type for validation: %s", expectedType)
	}
	return nil
}
