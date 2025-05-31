// ABOUTME: Provides comprehensive type conversion utilities between Go and Lua types
// ABOUTME: Handles complex types including structs, maps, slices, and functions

package lua

import (
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

// LuaConverter handles type conversions between Go and Lua
type LuaConverter struct {
	vm *lua.LState
}

// NewLuaConverter creates a new Lua type converter
func NewLuaConverter(vm *lua.LState) *LuaConverter {
	return &LuaConverter{vm: vm}
}

// ToLua converts a Go value to a Lua value
func (c *LuaConverter) ToLua(value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}

	v := reflect.ValueOf(value)
	return c.goToLua(v)
}

// goToLua performs the actual Go to Lua conversion
func (c *LuaConverter) goToLua(v reflect.Value) lua.LValue {
	if !v.IsValid() {
		return lua.LNil
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return lua.LNil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return lua.LBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(v.Uint())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(v.Float())
	case reflect.String:
		return lua.LString(v.String())
	case reflect.Slice, reflect.Array:
		return c.sliceToLua(v)
	case reflect.Map:
		return c.mapToLua(v)
	case reflect.Struct:
		return c.structToLua(v)
	case reflect.Func:
		return c.funcToLua(v)
	case reflect.Interface:
		if v.IsNil() {
			return lua.LNil
		}
		return c.goToLua(v.Elem())
	default:
		// For unsupported types, return as userdata
		ud := c.vm.NewUserData()
		ud.Value = v.Interface()
		return ud
	}
}

// sliceToLua converts a Go slice/array to a Lua table
func (c *LuaConverter) sliceToLua(v reflect.Value) *lua.LTable {
	table := c.vm.NewTable()
	for i := 0; i < v.Len(); i++ {
		table.RawSetInt(i+1, c.goToLua(v.Index(i))) // Lua arrays are 1-indexed
	}
	return table
}

// mapToLua converts a Go map to a Lua table
func (c *LuaConverter) mapToLua(v reflect.Value) *lua.LTable {
	table := c.vm.NewTable()
	for _, key := range v.MapKeys() {
		lKey := c.goToLua(key)
		lValue := c.goToLua(v.MapIndex(key))
		table.RawSet(lKey, lValue)
	}
	return table
}

// structToLua converts a Go struct to a Lua table
func (c *LuaConverter) structToLua(v reflect.Value) *lua.LTable {
	table := c.vm.NewTable()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		fieldName := field.Name
		// Check for json tag
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if tag == ",omitempty" {
				fieldName = field.Name
			} else {
				fieldName = tag
			}
		}

		table.RawSetString(fieldName, c.goToLua(v.Field(i)))
	}
	return table
}

// funcToLua converts a Go function to a Lua function
func (c *LuaConverter) funcToLua(v reflect.Value) *lua.LFunction {
	return c.vm.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()
		fnType := v.Type()

		// Check argument count
		if fnType.NumIn() != numArgs {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("expected %d arguments, got %d", fnType.NumIn(), numArgs)))
			return 2
		}

		// Convert Lua arguments to Go
		args := make([]reflect.Value, numArgs)
		for i := 0; i < numArgs; i++ {
			lVal := L.Get(i + 1)
			goVal, err := c.FromLua(lVal, fnType.In(i))
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			args[i] = reflect.ValueOf(goVal)
		}

		// Call the Go function
		results := v.Call(args)

		// Convert results back to Lua
		for _, result := range results {
			L.Push(c.goToLua(result))
		}

		return len(results)
	})
}

// FromLua converts a Lua value to a Go value of the specified type
func (c *LuaConverter) FromLua(lval lua.LValue, targetType reflect.Type) (interface{}, error) {
	if targetType.Kind() == reflect.Interface && targetType.NumMethod() == 0 {
		// Special case for interface{}
		return c.luaToInterface(lval), nil
	}

	v, err := c.luaToGo(lval, targetType)
	if err != nil {
		return nil, err
	}
	return v.Interface(), nil
}

// ToInterface converts a Lua value to interface{}
func (c *LuaConverter) ToInterface(lval lua.LValue) interface{} {
	return c.luaToInterface(lval)
}

// luaToInterface converts a Lua value to interface{}
func (c *LuaConverter) luaToInterface(lval lua.LValue) interface{} {
	switch v := lval.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		return c.tableToInterface(v)
	case *lua.LFunction:
		return v
	case *lua.LUserData:
		return v.Value
	default:
		return nil
	}
}

// tableToInterface converts a Lua table to appropriate Go type
func (c *LuaConverter) tableToInterface(table *lua.LTable) interface{} {
	// Check if it's an array
	length := table.Len()
	if length > 0 {
		// Try to convert as array
		arr := make([]interface{}, length)
		for i := 1; i <= length; i++ {
			arr[i-1] = c.luaToInterface(table.RawGetInt(i))
		}
		return arr
	}

	// Convert as map
	m := make(map[string]interface{})
	table.ForEach(func(k, v lua.LValue) {
		if ks, ok := k.(lua.LString); ok {
			m[string(ks)] = c.luaToInterface(v)
		}
	})
	return m
}

// luaToGo converts a Lua value to a specific Go type
func (c *LuaConverter) luaToGo(lval lua.LValue, targetType reflect.Type) (reflect.Value, error) {
	// Handle pointer types
	if targetType.Kind() == reflect.Ptr {
		if lval.Type() == lua.LTNil {
			return reflect.Zero(targetType), nil
		}
		elem, err := c.luaToGo(lval, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(elem)
		return ptr, nil
	}

	switch targetType.Kind() {
	case reflect.Bool:
		if v, ok := lval.(lua.LBool); ok {
			return reflect.ValueOf(bool(v)), nil
		}
		return reflect.Value{}, fmt.Errorf("expected bool, got %s", lval.Type())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, ok := lval.(lua.LNumber); ok {
			return reflect.ValueOf(int64(v)).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", lval.Type())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, ok := lval.(lua.LNumber); ok {
			return reflect.ValueOf(uint64(v)).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", lval.Type())

	case reflect.Float32, reflect.Float64:
		if v, ok := lval.(lua.LNumber); ok {
			return reflect.ValueOf(float64(v)).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", lval.Type())

	case reflect.String:
		if v, ok := lval.(lua.LString); ok {
			return reflect.ValueOf(string(v)), nil
		}
		return reflect.Value{}, fmt.Errorf("expected string, got %s", lval.Type())

	case reflect.Slice:
		if table, ok := lval.(*lua.LTable); ok {
			return c.tableToSlice(table, targetType)
		}
		return reflect.Value{}, fmt.Errorf("expected table, got %s", lval.Type())

	case reflect.Map:
		if table, ok := lval.(*lua.LTable); ok {
			return c.tableToMap(table, targetType)
		}
		return reflect.Value{}, fmt.Errorf("expected table, got %s", lval.Type())

	case reflect.Struct:
		if table, ok := lval.(*lua.LTable); ok {
			return c.tableToStruct(table, targetType)
		}
		return reflect.Value{}, fmt.Errorf("expected table, got %s", lval.Type())

	case reflect.Interface:
		// Handle interface types
		if targetType.NumMethod() == 0 {
			// interface{}
			return reflect.ValueOf(c.luaToInterface(lval)), nil
		}
		// For other interfaces, try to get the underlying value
		if ud, ok := lval.(*lua.LUserData); ok {
			val := reflect.ValueOf(ud.Value)
			if val.Type().Implements(targetType) {
				return val, nil
			}
		}
		return reflect.Value{}, fmt.Errorf("value does not implement interface %s", targetType)

	default:
		return reflect.Value{}, fmt.Errorf("unsupported type: %s", targetType)
	}
}

// tableToSlice converts a Lua table to a Go slice
func (c *LuaConverter) tableToSlice(table *lua.LTable, targetType reflect.Type) (reflect.Value, error) {
	length := table.Len()
	slice := reflect.MakeSlice(targetType, length, length)
	elemType := targetType.Elem()

	for i := 1; i <= length; i++ {
		lVal := table.RawGetInt(i)
		elem, err := c.luaToGo(lVal, elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error converting element %d: %w", i, err)
		}
		slice.Index(i - 1).Set(elem)
	}

	return slice, nil
}

// tableToMap converts a Lua table to a Go map
func (c *LuaConverter) tableToMap(table *lua.LTable, targetType reflect.Type) (reflect.Value, error) {
	mapVal := reflect.MakeMap(targetType)
	keyType := targetType.Key()
	elemType := targetType.Elem()

	var convErr error
	table.ForEach(func(k, v lua.LValue) {
		if convErr != nil {
			return
		}

		key, err := c.luaToGo(k, keyType)
		if err != nil {
			convErr = fmt.Errorf("error converting key: %w", err)
			return
		}

		value, err := c.luaToGo(v, elemType)
		if err != nil {
			convErr = fmt.Errorf("error converting value: %w", err)
			return
		}

		mapVal.SetMapIndex(key, value)
	})

	if convErr != nil {
		return reflect.Value{}, convErr
	}

	return mapVal, nil
}

// tableToStruct converts a Lua table to a Go struct
func (c *LuaConverter) tableToStruct(table *lua.LTable, targetType reflect.Type) (reflect.Value, error) {
	structVal := reflect.New(targetType).Elem()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		fieldName := field.Name
		// Check for json tag
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			if tag != ",omitempty" {
				fieldName = tag
			}
		}

		lVal := table.RawGetString(fieldName)
		if lVal.Type() != lua.LTNil {
			fieldValue, err := c.luaToGo(lVal, field.Type)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting field %s: %w", fieldName, err)
			}
			structVal.Field(i).Set(fieldValue)
		}
	}

	return structVal, nil
}
