// ABOUTME: This file defines common type representations and validation for the multi-engine system.
// ABOUTME: It provides type conversion utilities and validation frameworks used across all script engines.

package engine

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// BaseTypeConverter provides a default implementation of TypeConverter interface.
// Engine-specific converters can embed this and override specific methods.
type BaseTypeConverter struct {
	engineName string
	adapters   map[string]TypeAdapter
}

// TypeAdapter handles conversion for specific types.
type TypeAdapter interface {
	ToNative(v interface{}) (interface{}, error)
	FromNative(v interface{}) (interface{}, error)
	SupportsType(typeName string) bool
}

// NewBaseTypeConverter creates a new base type converter.
func NewBaseTypeConverter(engineName string) *BaseTypeConverter {
	return &BaseTypeConverter{
		engineName: engineName,
		adapters:   make(map[string]TypeAdapter),
	}
}

// RegisterAdapter registers a type adapter for specific types.
func (c *BaseTypeConverter) RegisterAdapter(typeName string, adapter TypeAdapter) {
	c.adapters[typeName] = adapter
}

// ToBoolean converts a value to boolean.
func (c *BaseTypeConverter) ToBoolean(v interface{}) (bool, error) {
	if v == nil {
		return false, nil
	}

	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		switch strings.ToLower(val) {
		case "true", "yes", "1", "on":
			return true, nil
		case "false", "no", "0", "off", "":
			return false, nil
		default:
			return false, fmt.Errorf("cannot convert string %q to boolean", val)
		}
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(val).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(val).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(val).Float() != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to boolean", v)
	}
}

// ToNumber converts a value to float64.
func (c *BaseTypeConverter) ToNumber(v interface{}) (float64, error) {
	if v == nil {
		return 0, nil
	}

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
		if val == "" {
			return 0, nil
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to number: %w", val, err)
		}
		return f, nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// ToString converts a value to string.
func (c *BaseTypeConverter) ToString(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}

	switch val := v.(type) {
	case string:
		return val, nil
	case bool:
		return strconv.FormatBool(val), nil
	case int:
		return strconv.Itoa(val), nil
	case int8:
		return strconv.FormatInt(int64(val), 10), nil
	case int16:
		return strconv.FormatInt(int64(val), 10), nil
	case int32:
		return strconv.FormatInt(int64(val), 10), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case uint:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint64:
		return strconv.FormatUint(val, 10), nil
	case float32:
		return strconv.FormatFloat(float64(val), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64), nil
	case time.Time:
		return val.Format(time.RFC3339), nil
	case fmt.Stringer:
		return val.String(), nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// ToArray converts a value to []interface{}.
func (c *BaseTypeConverter) ToArray(v interface{}) ([]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case []interface{}:
		return val, nil
	case []string:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	case []int:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	case []float64:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	case []bool:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result, nil
	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil, fmt.Errorf("cannot convert %T to array", v)
		}
		
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result, nil
	}
}

// ToMap converts a value to map[string]interface{}.
func (c *BaseTypeConverter) ToMap(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return val, nil
	case map[string]string:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = v
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			key, err := c.ToString(k)
			if err != nil {
				return nil, fmt.Errorf("cannot convert map key to string: %w", err)
			}
			result[key] = v
		}
		return result, nil
	default:
		// Use reflection for struct types
		rv := reflect.ValueOf(v)
		rt := reflect.TypeOf(v)
		
		// Dereference pointer if needed
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
			rt = rt.Elem()
		}
		
		if rv.Kind() != reflect.Struct {
			return nil, fmt.Errorf("cannot convert %T to map", v)
		}
		
		result := make(map[string]interface{})
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			value := rv.Field(i)
			
			// Skip unexported fields
			if !field.IsExported() {
				continue
			}
			
			// Use json tag if available, otherwise use field name
			key := field.Name
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				if comma := strings.Index(tag, ","); comma > 0 {
					key = tag[:comma]
				} else {
					key = tag
				}
			}
			
			result[key] = value.Interface()
		}
		
		return result, nil
	}
}

// ToStruct converts a value to the target struct.
func (c *BaseTypeConverter) ToStruct(v interface{}, target interface{}) error {
	if v == nil {
		return nil
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	if !targetValue.CanSet() {
		return fmt.Errorf("target cannot be set")
	}

	sourceMap, err := c.ToMap(v)
	if err != nil {
		return fmt.Errorf("cannot convert source to map: %w", err)
	}

	return c.mapToStruct(sourceMap, targetValue)
}

// FromStruct converts a struct to map[string]interface{}.
func (c *BaseTypeConverter) FromStruct(v interface{}) (map[string]interface{}, error) {
	return c.ToMap(v)
}

// ToFunction converts a value to Function interface.
func (c *BaseTypeConverter) ToFunction(v interface{}) (Function, error) {
	if fn, ok := v.(Function); ok {
		return fn, nil
	}
	
	// Check if there's a registered adapter for this type
	typeName := reflect.TypeOf(v).String()
	if adapter, exists := c.adapters[typeName]; exists {
		native, err := adapter.ToNative(v)
		if err != nil {
			return nil, err
		}
		if fn, ok := native.(Function); ok {
			return fn, nil
		}
	}
	
	return nil, fmt.Errorf("cannot convert %T to Function", v)
}

// FromFunction converts a Function to engine-specific representation.
func (c *BaseTypeConverter) FromFunction(fn Function) (interface{}, error) {
	// Check if there's a registered adapter for functions
	if adapter, exists := c.adapters["function"]; exists {
		return adapter.FromNative(fn)
	}
	
	// Default: return as-is
	return fn, nil
}

// SupportsType checks if a type is supported by this converter.
func (c *BaseTypeConverter) SupportsType(typeName string) bool {
	// Check built-in types
	builtinTypes := []string{
		"bool", "string", "int", "float64", "[]interface{}", "map[string]interface{}",
		"int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "time.Time",
	}
	
	for _, builtin := range builtinTypes {
		if builtin == typeName {
			return true
		}
	}
	
	// Check registered adapters
	_, exists := c.adapters[typeName]
	return exists
}

// GetTypeInfo returns information about a supported type.
func (c *BaseTypeConverter) GetTypeInfo(typeName string) TypeInfo {
	// Return basic type info for built-in types
	switch typeName {
	case "bool":
		return TypeInfo{
			Name:        "bool",
			Category:    TypeCategoryPrimitive,
			Description: "Boolean value (true/false)",
		}
	case "string":
		return TypeInfo{
			Name:        "string",
			Category:    TypeCategoryPrimitive,
			Description: "Text string value",
		}
	case "number", "float64":
		return TypeInfo{
			Name:        "number",
			Category:    TypeCategoryPrimitive,
			Description: "Numeric value (integer or floating point)",
		}
	case "array", "[]interface{}":
		return TypeInfo{
			Name:        "array",
			Category:    TypeCategoryArray,
			Description: "Ordered collection of values",
		}
	case "object", "map[string]interface{}":
		return TypeInfo{
			Name:        "object",
			Category:    TypeCategoryObject,
			Description: "Key-value collection of properties",
		}
	case "function":
		return TypeInfo{
			Name:        "function",
			Category:    TypeCategoryFunction,
			Description: "Callable function with parameters and return value",
		}
	default:
		return TypeInfo{
			Name:        typeName,
			Category:    TypeCategoryCustom,
			Description: "Custom type specific to " + c.engineName,
		}
	}
}

// mapToStruct uses reflection to populate a struct from a map.
func (c *BaseTypeConverter) mapToStruct(source map[string]interface{}, target reflect.Value) error {
	targetType := target.Type()
	
	for i := 0; i < target.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := target.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get the key to look for in the source map
		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if comma := strings.Index(tag, ","); comma > 0 {
				key = tag[:comma]
			} else {
				key = tag
			}
		}
		
		// Check if the key exists in the source map
		sourceValue, exists := source[key]
		if !exists {
			continue
		}
		
		// Convert and set the value
		if err := c.setFieldValue(sourceValue, fieldValue); err != nil {
			return fmt.Errorf("error setting field %s: %w", field.Name, err)
		}
	}
	
	return nil
}

// setFieldValue sets a struct field value with type conversion.
func (c *BaseTypeConverter) setFieldValue(source interface{}, target reflect.Value) error {
	if source == nil {
		return nil
	}
	
	sourceValue := reflect.ValueOf(source)
	targetType := target.Type()
	
	// Direct assignment if types match
	if sourceValue.Type().AssignableTo(targetType) {
		target.Set(sourceValue)
		return nil
	}
	
	// Type conversion based on target type
	switch targetType.Kind() {
	case reflect.Bool:
		val, err := c.ToBoolean(source)
		if err != nil {
			return err
		}
		target.SetBool(val)
		
	case reflect.String:
		val, err := c.ToString(source)
		if err != nil {
			return err
		}
		target.SetString(val)
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := c.ToNumber(source)
		if err != nil {
			return err
		}
		target.SetInt(int64(val))
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := c.ToNumber(source)
		if err != nil {
			return err
		}
		target.SetUint(uint64(val))
		
	case reflect.Float32, reflect.Float64:
		val, err := c.ToNumber(source)
		if err != nil {
			return err
		}
		target.SetFloat(val)
		
	case reflect.Slice:
		val, err := c.ToArray(source)
		if err != nil {
			return err
		}
		
		slice := reflect.MakeSlice(targetType, len(val), len(val))
		for i, item := range val {
			if err := c.setFieldValue(item, slice.Index(i)); err != nil {
				return fmt.Errorf("error setting slice element %d: %w", i, err)
			}
		}
		target.Set(slice)
		
	case reflect.Map:
		val, err := c.ToMap(source)
		if err != nil {
			return err
		}
		
		mapValue := reflect.MakeMap(targetType)
		for k, v := range val {
			keyValue := reflect.ValueOf(k)
			valueValue := reflect.New(targetType.Elem()).Elem()
			
			if err := c.setFieldValue(v, valueValue); err != nil {
				return fmt.Errorf("error setting map value for key %s: %w", k, err)
			}
			
			mapValue.SetMapIndex(keyValue, valueValue)
		}
		target.Set(mapValue)
		
	case reflect.Struct:
		val, err := c.ToMap(source)
		if err != nil {
			return err
		}
		return c.mapToStruct(val, target)
		
	case reflect.Ptr:
		if target.IsNil() {
			target.Set(reflect.New(targetType.Elem()))
		}
		return c.setFieldValue(source, target.Elem())
		
	default:
		return fmt.Errorf("unsupported target type: %s", targetType.Kind())
	}
	
	return nil
}

// ValidateType validates that a value conforms to the expected type.
func (c *BaseTypeConverter) ValidateType(value interface{}, expectedType string) error {
	if value == nil {
		return nil // nil is valid for any type
	}
	
	switch expectedType {
	case "bool", "boolean":
		_, err := c.ToBoolean(value)
		return err
	case "string":
		_, err := c.ToString(value)
		return err
	case "number", "float64", "int":
		_, err := c.ToNumber(value)
		return err
	case "array", "[]interface{}":
		_, err := c.ToArray(value)
		return err
	case "object", "map", "map[string]interface{}":
		_, err := c.ToMap(value)
		return err
	case "function":
		_, err := c.ToFunction(value)
		return err
	default:
		// Check if type is supported by registered adapters
		if !c.SupportsType(expectedType) {
			return fmt.Errorf("unsupported type: %s", expectedType)
		}
		return nil
	}
}

// GetConversionPath returns the conversion path between two types.
func (c *BaseTypeConverter) GetConversionPath(fromType, toType string) ([]string, error) {
	// Simple direct conversions
	directConversions := map[string][]string{
		"string": {"bool", "number", "array"},
		"number": {"bool", "string"},
		"bool":   {"string", "number"},
		"array":  {"string"},
		"object": {"string"},
	}
	
	if targets, exists := directConversions[fromType]; exists {
		for _, target := range targets {
			if target == toType {
				return []string{fromType, toType}, nil
			}
		}
	}
	
	// No conversion path found
	return nil, fmt.Errorf("no conversion path from %s to %s", fromType, toType)
}