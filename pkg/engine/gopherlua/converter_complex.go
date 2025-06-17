// ABOUTME: ComplexConverter handles conversion of complex Go types (maps, slices, structs, interfaces)
// ABOUTME: Provides struct tag support, field mapping, circular reference detection, and nested type handling

package gopherlua

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// ComplexConverter handles conversion of complex types with advanced features
type ComplexConverter struct {
	primitiveConverter *PrimitiveConverter
	maxDepth           int
}

// StructTagInfo contains parsed struct tag information
type StructTagInfo struct {
	Name      string
	Omitempty bool
	Skip      bool
	Required  bool
}

// NewComplexConverter creates a new complex type converter
func NewComplexConverter() *ComplexConverter {
	return &ComplexConverter{
		primitiveConverter: NewPrimitiveConverter(),
		maxDepth:           32, // Default max depth for nested structures
	}
}

// Map conversion methods

// MapToLua converts a Go map to a Lua table
func (cc *ComplexConverter) MapToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	return cc.mapToLuaWithDepth(L, value, 0, make(map[uintptr]bool))
}

func (cc *ComplexConverter) mapToLuaWithDepth(L *lua.LState, value interface{}, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	if depth > cc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", cc.maxDepth)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected map, got %T", value)
	}

	// Check for circular references
	if rv.Pointer() != 0 {
		if visited[rv.Pointer()] {
			return nil, fmt.Errorf("circular reference detected in map")
		}
		visited[rv.Pointer()] = true
		defer delete(visited, rv.Pointer())
	}

	table := L.NewTable()

	for _, key := range rv.MapKeys() {
		// Convert key to string
		keyStr, err := cc.convertMapKeyToString(key.Interface())
		if err != nil {
			return nil, fmt.Errorf("unsupported map key type %T: %w", key.Interface(), err)
		}

		// Convert value
		mapValue := rv.MapIndex(key).Interface()
		luaValue, err := cc.convertToLuaWithDepth(L, mapValue, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert map value for key %s: %w", keyStr, err)
		}

		table.RawSetString(keyStr, luaValue)
	}

	return table, nil
}

func (cc *ComplexConverter) convertMapKeyToString(key interface{}) (string, error) {
	switch k := key.(type) {
	case string:
		return k, nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", k), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", k), nil
	case float32, float64:
		return fmt.Sprintf("%g", k), nil
	case bool:
		return fmt.Sprintf("%t", k), nil
	default:
		return "", fmt.Errorf("unsupported map key type: %T", key)
	}
}

// Slice conversion methods

// SliceToLua converts a Go slice or array to a Lua table
func (cc *ComplexConverter) SliceToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	return cc.sliceToLuaWithDepth(L, value, 0, make(map[uintptr]bool))
}

func (cc *ComplexConverter) sliceToLuaWithDepth(L *lua.LState, value interface{}, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	if depth > cc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", cc.maxDepth)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected slice or array, got %T", value)
	}

	// Check for circular references (only for slices, not arrays)
	if rv.Kind() == reflect.Slice && rv.Pointer() != 0 {
		if visited[rv.Pointer()] {
			return nil, fmt.Errorf("circular reference detected in slice")
		}
		visited[rv.Pointer()] = true
		defer delete(visited, rv.Pointer())
	}

	table := L.NewTable()

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		luaValue, err := cc.convertToLuaWithDepth(L, elem, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert slice element %d: %w", i, err)
		}
		table.RawSetInt(i+1, luaValue) // Lua arrays are 1-indexed
	}

	return table, nil
}

// Struct conversion methods

// StructToLua converts a Go struct to a Lua table with struct tag support
func (cc *ComplexConverter) StructToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	return cc.structToLuaWithDepth(L, value, 0, make(map[uintptr]bool))
}

func (cc *ComplexConverter) structToLuaWithDepth(L *lua.LState, value interface{}, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	if depth > cc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", cc.maxDepth)
	}

	rv := reflect.ValueOf(value)
	rt := rv.Type()

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", value)
	}

	table := L.NewTable()

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse struct tags
		tagInfo := cc.ParseStructTag(field)

		// Skip fields marked with "-"
		if tagInfo.Skip {
			continue
		}

		// Handle omitempty
		if tagInfo.Omitempty && cc.isEmptyValue(fieldValue) {
			continue
		}

		// Convert field value
		luaValue, err := cc.convertToLuaWithDepth(L, fieldValue.Interface(), depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert struct field %s: %w", field.Name, err)
		}

		// Use tag name or field name
		fieldName := tagInfo.Name
		if fieldName == "" {
			fieldName = field.Name
		}

		table.RawSetString(fieldName, luaValue)
	}

	return table, nil
}

// ParseStructTag parses struct field tags for lua conversion
func (cc *ComplexConverter) ParseStructTag(field reflect.StructField) StructTagInfo {
	tag := field.Tag.Get("lua")
	if tag == "" {
		return StructTagInfo{
			Name:      field.Name,
			Omitempty: false,
			Skip:      false,
			Required:  false,
		}
	}

	if tag == "-" {
		return StructTagInfo{
			Name:      "",
			Omitempty: false,
			Skip:      true,
			Required:  false,
		}
	}

	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		name = field.Name
	}

	var omitempty, required bool
	for i := 1; i < len(parts); i++ {
		switch strings.TrimSpace(parts[i]) {
		case "omitempty":
			omitempty = true
		case "required":
			required = true
		}
	}

	return StructTagInfo{
		Name:      name,
		Omitempty: omitempty,
		Skip:      false,
		Required:  required,
	}
}

func (cc *ComplexConverter) isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// Interface conversion methods

// InterfaceToLua converts an interface{} to a Lua value by examining its concrete type
func (cc *ComplexConverter) InterfaceToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return lua.LNil, nil
		}
		// Get the concrete value from the interface
		value = rv.Elem().Interface()
	}

	return cc.convertToLuaWithDepth(L, value, 0, make(map[uintptr]bool))
}

// FromLua converts a Lua value to a Go value (reverse conversion)
func (cc *ComplexConverter) FromLua(value lua.LValue) (interface{}, error) {
	return cc.fromLuaWithDepth(value, 0, make(map[*lua.LTable]bool))
}

func (cc *ComplexConverter) fromLuaWithDepth(value lua.LValue, depth int, visited map[*lua.LTable]bool) (interface{}, error) {
	if depth > cc.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", cc.maxDepth)
	}

	switch lv := value.(type) {
	case lua.LBool:
		return bool(lv), nil
	case lua.LNumber:
		return float64(lv), nil
	case lua.LString:
		return string(lv), nil
	case *lua.LNilType:
		return nil, nil
	case *lua.LTable:
		// Check for circular references
		if visited[lv] {
			return nil, fmt.Errorf("circular reference detected in Lua table")
		}
		visited[lv] = true
		defer delete(visited, lv)

		return cc.luaTableToGo(lv, depth, visited)
	case *lua.LUserData:
		return lv.Value, nil
	case *lua.LFunction:
		return fmt.Sprintf("function<%p>", lv), nil
	default:
		return nil, fmt.Errorf("unsupported Lua type for conversion to Go: %T", value)
	}
}

func (cc *ComplexConverter) luaTableToGo(table *lua.LTable, depth int, visited map[*lua.LTable]bool) (interface{}, error) {
	// Check if this is an array-like table (consecutive integer keys starting from 1)
	if cc.isArrayLikeTable(table) {
		return cc.luaTableToSlice(table, depth, visited)
	}

	// Otherwise, treat as a map
	return cc.luaTableToMap(table, depth, visited)
}

func (cc *ComplexConverter) isArrayLikeTable(table *lua.LTable) bool {
	length := table.Len()
	if length == 0 {
		// Empty table - check if it has any non-integer keys
		hasStringKeys := false
		table.ForEach(func(key, value lua.LValue) {
			if _, ok := key.(lua.LString); ok {
				hasStringKeys = true
			}
		})
		return !hasStringKeys
	}

	// Check if all keys from 1 to length exist
	for i := 1; i <= length; i++ {
		if table.RawGetInt(i) == lua.LNil {
			return false
		}
	}

	// Check if there are any non-integer keys
	hasNonIntegerKeys := false
	table.ForEach(func(key, value lua.LValue) {
		if num, ok := key.(lua.LNumber); ok {
			if float64(num) != float64(int(num)) || int(num) < 1 || int(num) > length {
				hasNonIntegerKeys = true
			}
		} else {
			hasNonIntegerKeys = true
		}
	})

	return !hasNonIntegerKeys
}

func (cc *ComplexConverter) luaTableToSlice(table *lua.LTable, depth int, visited map[*lua.LTable]bool) ([]interface{}, error) {
	length := table.Len()
	result := make([]interface{}, length)

	for i := 1; i <= length; i++ {
		luaValue := table.RawGetInt(i)
		goValue, err := cc.fromLuaWithDepth(luaValue, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array element %d: %w", i, err)
		}
		result[i-1] = goValue // Convert from 1-indexed to 0-indexed
	}

	return result, nil
}

func (cc *ComplexConverter) luaTableToMap(table *lua.LTable, depth int, visited map[*lua.LTable]bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	table.ForEach(func(key, value lua.LValue) {
		keyStr := fmt.Sprintf("%v", key)
		goValue, err := cc.fromLuaWithDepth(value, depth+1, visited)
		if err != nil {
			// Skip values that can't be converted rather than failing entirely
			return
		}
		result[keyStr] = goValue
	})

	return result, nil
}

// Helper method to route to appropriate conversion based on type
func (cc *ComplexConverter) convertToLuaWithDepth(L *lua.LState, value interface{}, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return cc.primitiveConverter.BoolToLua(L, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return cc.primitiveConverter.NumberToLua(L, value)
	case reflect.String:
		return cc.primitiveConverter.StringToLua(L, value)
	case reflect.Map:
		return cc.mapToLuaWithDepth(L, value, depth, visited)
	case reflect.Slice, reflect.Array:
		return cc.sliceToLuaWithDepth(L, value, depth, visited)
	case reflect.Struct:
		return cc.structToLuaWithDepth(L, value, depth, visited)
	case reflect.Ptr:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return cc.convertToLuaWithDepth(L, rv.Elem().Interface(), depth+1, visited)
	case reflect.Interface:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return cc.convertToLuaWithDepth(L, rv.Elem().Interface(), depth+1, visited)
	default:
		return nil, fmt.Errorf("unsupported type for conversion to Lua: %T", value)
	}
}
