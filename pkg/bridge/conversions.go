// ABOUTME: Type conversion utilities for bridging Go and script types
// ABOUTME: Provides safe conversions between native Go types and script representations

package bridge

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Common type references for conversions
var (
	MapType = reflect.TypeOf(map[string]interface{}{})
)

// Converter defines the interface for type conversion between Go and script languages
type Converter interface {
	// ToScript converts a Go value to a script-compatible representation
	ToScript(value interface{}) (interface{}, error)

	// FromScript converts a script value to a Go type
	FromScript(value interface{}, targetType reflect.Type) (interface{}, error)
}

// BaseConverter provides common type conversion functionality
type BaseConverter struct{}

// ToScript converts Go values to a generic script-compatible format
func (c *BaseConverter) ToScript(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Invalid:
		return nil, nil

	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		// Basic types pass through
		return value, nil

	case reflect.Slice, reflect.Array:
		// Convert slices/arrays to []interface{}
		length := v.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			elem, err := c.ToScript(v.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("error converting array element %d: %w", i, err)
			}
			result[i] = elem
		}
		return result, nil

	case reflect.Map:
		// Convert maps to map[string]interface{}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			val, err := c.ToScript(v.MapIndex(key).Interface())
			if err != nil {
				return nil, fmt.Errorf("error converting map value for key %s: %w", keyStr, err)
			}
			result[keyStr] = val
		}
		return result, nil

	case reflect.Struct:
		// Convert structs to map[string]interface{} using JSON tags
		return c.structToMap(v)

	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return c.ToScript(v.Elem().Interface())

	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		return c.ToScript(v.Elem().Interface())

	default:
		// For unsupported types, try JSON marshaling
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unsupported type %s: %w", v.Kind(), err)
		}
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to convert type %s: %w", v.Kind(), err)
		}
		return result, nil
	}
}

// FromScript converts script values to Go types
func (c *BaseConverter) FromScript(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}

	// Handle basic type conversions
	switch targetType.Kind() {
	case reflect.String:
		return c.toString(value)

	case reflect.Bool:
		return c.toBool(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := c.toNumber(value)
		if err != nil {
			return nil, err
		}
		return c.convertNumber(n, targetType)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := c.toNumber(value)
		if err != nil {
			return nil, err
		}
		return c.convertNumber(n, targetType)

	case reflect.Float32, reflect.Float64:
		n, err := c.toNumber(value)
		if err != nil {
			return nil, err
		}
		return c.convertNumber(n, targetType)

	case reflect.Slice:
		return c.toSlice(value, targetType)

	case reflect.Map:
		return c.toMap(value, targetType)

	case reflect.Struct:
		return c.toStruct(value, targetType)

	case reflect.Ptr:
		// Create a new pointer and convert the value
		ptr := reflect.New(targetType.Elem())
		val, err := c.FromScript(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		ptr.Elem().Set(reflect.ValueOf(val))
		return ptr.Interface(), nil

	case reflect.Interface:
		// For interface{}, return the value as-is
		if targetType.NumMethod() == 0 {
			return value, nil
		}
		// For non-empty interfaces, we can't convert
		return nil, fmt.Errorf("cannot convert to non-empty interface %s", targetType)

	default:
		// Try JSON unmarshaling for complex types
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to type %s: %w", targetType, err)
		}
		result := reflect.New(targetType).Interface()
		if err := json.Unmarshal(data, result); err != nil {
			return nil, fmt.Errorf("failed to convert to type %s: %w", targetType, err)
		}
		return reflect.ValueOf(result).Elem().Interface(), nil
	}
}

// Helper methods

func (c *BaseConverter) structToMap(v reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}

		// Get field name from json tag or use field name
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			if tag == "-" {
				continue // Skip fields with json:"-"
			}
			if idx := findChar(tag, ','); idx != -1 {
				name = tag[:idx]
			} else {
				name = tag
			}
		}

		val, err := c.ToScript(v.Field(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("error converting field %s: %w", name, err)
		}
		result[name] = val
	}

	return result, nil
}

func (c *BaseConverter) toString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

func (c *BaseConverter) toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	case string:
		return v != "" && v != "false" && v != "0", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

func (c *BaseConverter) toNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		var n float64
		if _, err := fmt.Sscanf(v, "%f", &n); err != nil {
			return 0, fmt.Errorf("cannot convert string %q to number", v)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}

func (c *BaseConverter) convertNumber(n float64, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return int(n), nil
	case reflect.Int8:
		return int8(n), nil
	case reflect.Int16:
		return int16(n), nil
	case reflect.Int32:
		return int32(n), nil
	case reflect.Int64:
		return int64(n), nil
	case reflect.Uint:
		return uint(n), nil
	case reflect.Uint8:
		return uint8(n), nil
	case reflect.Uint16:
		return uint16(n), nil
	case reflect.Uint32:
		return uint32(n), nil
	case reflect.Uint64:
		return uint64(n), nil
	case reflect.Float32:
		return float32(n), nil
	case reflect.Float64:
		return n, nil
	default:
		return nil, fmt.Errorf("cannot convert number to %s", targetType)
	}
}

func (c *BaseConverter) toSlice(value interface{}, targetType reflect.Type) (interface{}, error) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot convert %T to slice", value)
	}

	elemType := targetType.Elem()
	result := reflect.MakeSlice(targetType, v.Len(), v.Len())

	for i := 0; i < v.Len(); i++ {
		elem, err := c.FromScript(v.Index(i).Interface(), elemType)
		if err != nil {
			return nil, fmt.Errorf("error converting slice element %d: %w", i, err)
		}
		result.Index(i).Set(reflect.ValueOf(elem))
	}

	return result.Interface(), nil
}

func (c *BaseConverter) toMap(value interface{}, targetType reflect.Type) (interface{}, error) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot convert %T to map", value)
	}

	keyType := targetType.Key()
	elemType := targetType.Elem()
	result := reflect.MakeMap(targetType)

	for _, key := range v.MapKeys() {
		// Convert key
		k, err := c.FromScript(key.Interface(), keyType)
		if err != nil {
			return nil, fmt.Errorf("error converting map key: %w", err)
		}

		// Convert value
		val, err := c.FromScript(v.MapIndex(key).Interface(), elemType)
		if err != nil {
			return nil, fmt.Errorf("error converting map value: %w", err)
		}

		result.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(val))
	}

	return result.Interface(), nil
}

func (c *BaseConverter) toStruct(value interface{}, targetType reflect.Type) (interface{}, error) {
	// Expect value to be a map
	m, ok := value.(map[string]interface{})
	if !ok {
		// Try to convert to map first
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Map {
			m = make(map[string]interface{})
			for _, key := range v.MapKeys() {
				keyStr := fmt.Sprintf("%v", key.Interface())
				m[keyStr] = v.MapIndex(key).Interface()
			}
		} else {
			return nil, fmt.Errorf("cannot convert %T to struct", value)
		}
	}

	result := reflect.New(targetType).Elem()

	// Create a map of field names to field indices
	fieldMap := make(map[string]int)
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}

		// Use json tag if available
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if idx := findChar(tag, ','); idx != -1 {
				name = tag[:idx]
			} else {
				name = tag
			}
		}
		fieldMap[name] = i
	}

	// Set field values
	for name, value := range m {
		if idx, ok := fieldMap[name]; ok {
			field := targetType.Field(idx)
			val, err := c.FromScript(value, field.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting field %s: %w", name, err)
			}
			result.Field(idx).Set(reflect.ValueOf(val))
		}
	}

	return result.Interface(), nil
}

// findChar finds the first occurrence of a character in a string
func findChar(s string, c rune) int {
	for i, r := range s {
		if r == c {
			return i
		}
	}
	return -1
}
