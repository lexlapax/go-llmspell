// ABOUTME: FunctionConverter handles wrapping Go functions for Lua execution
// ABOUTME: Provides argument conversion, return value handling, panic recovery, and variadic function support

package gopherlua

import (
	"fmt"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

// FunctionConverter handles conversion of Go functions to Lua functions
type FunctionConverter struct {
	primitiveConverter *PrimitiveConverter
	complexConverter   *ComplexConverter
}

// NewFunctionConverter creates a new function converter
func NewFunctionConverter() *FunctionConverter {
	return &FunctionConverter{
		primitiveConverter: NewPrimitiveConverter(),
		complexConverter:   NewComplexConverter(),
	}
}

// WrapGoFunction wraps a Go function for use in Lua
func (fc *FunctionConverter) WrapGoFunction(L *lua.LState, fn interface{}) (lua.LValue, error) {
	if fn == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected function, got %T", fn)
	}

	// Validate function signature
	if err := fc.validateFunctionSignature(fnType); err != nil {
		return nil, err
	}

	// Create Lua function wrapper
	return L.NewFunction(func(L *lua.LState) int {
		return fc.callGoFunction(L, fnValue, fnType)
	}), nil
}

// validateFunctionSignature checks if the function signature is supported
func (fc *FunctionConverter) validateFunctionSignature(fnType reflect.Type) error {
	// Check parameter types
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		if !fc.isSupportedType(paramType) {
			return fmt.Errorf("unsupported parameter type: %s", paramType.String())
		}
	}

	// Check return types
	for i := 0; i < fnType.NumOut(); i++ {
		returnType := fnType.Out(i)
		// Allow error as the last return type
		if i == fnType.NumOut()-1 && returnType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}
		if !fc.isSupportedType(returnType) {
			return fmt.Errorf("unsupported return type: %s", returnType.String())
		}
	}

	return nil
}

// isSupportedType checks if a type is supported for conversion
func (fc *FunctionConverter) isSupportedType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.String:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice, reflect.Array:
		return fc.isSupportedType(t.Elem())
	case reflect.Map:
		return fc.isSupportedType(t.Key()) && fc.isSupportedType(t.Elem())
	case reflect.Interface:
		// Allow empty interface and error interface
		return t == reflect.TypeOf((*interface{})(nil)).Elem() ||
			t.Implements(reflect.TypeOf((*error)(nil)).Elem())
	case reflect.Ptr:
		return fc.isSupportedType(t.Elem())
	default:
		return false
	}
}

// callGoFunction executes the Go function with converted arguments
func (fc *FunctionConverter) callGoFunction(L *lua.LState, fnValue reflect.Value, fnType reflect.Type) int {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			L.RaiseError("function panicked: %v", r)
		}
	}()

	// Convert Lua arguments to Go values
	args, err := fc.convertLuaArgs(L, fnType)
	if err != nil {
		L.RaiseError("argument conversion failed: %s", err.Error())
		return 0
	}

	// Call the Go function
	var results []reflect.Value
	if fnType.IsVariadic() {
		results = fnValue.CallSlice(args)
	} else {
		results = fnValue.Call(args)
	}

	// Handle results
	return fc.handleResults(L, results, fnType)
}

// convertLuaArgs converts Lua arguments to Go values
func (fc *FunctionConverter) convertLuaArgs(L *lua.LState, fnType reflect.Type) ([]reflect.Value, error) {
	numArgs := L.GetTop()
	numParams := fnType.NumIn()
	isVariadic := fnType.IsVariadic()

	// Check argument count
	if isVariadic {
		if numArgs < numParams-1 {
			return nil, fmt.Errorf("expected at least %d arguments, got %d", numParams-1, numArgs)
		}
	} else {
		if numArgs != numParams {
			return nil, fmt.Errorf("expected %d arguments, got %d", numParams, numArgs)
		}
	}

	args := make([]reflect.Value, 0, numParams)

	// Convert regular parameters
	regularParams := numParams
	if isVariadic {
		regularParams = numParams - 1
	}

	for i := 0; i < regularParams; i++ {
		luaArg := L.Get(i + 1)
		paramType := fnType.In(i)

		goValue, err := fc.convertLuaValueToGo(luaArg, paramType)
		if err != nil {
			return nil, fmt.Errorf("parameter %d: %w", i+1, err)
		}

		args = append(args, goValue)
	}

	// Handle variadic parameters
	if isVariadic {
		variadicType := fnType.In(numParams - 1).Elem()
		variadicArgs := make([]reflect.Value, 0)

		for i := regularParams; i < numArgs; i++ {
			luaArg := L.Get(i + 1)
			goValue, err := fc.convertLuaValueToGo(luaArg, variadicType)
			if err != nil {
				return nil, fmt.Errorf("variadic parameter %d: %w", i+1, err)
			}
			variadicArgs = append(variadicArgs, goValue)
		}

		// Create variadic slice
		variadicSlice := reflect.MakeSlice(fnType.In(numParams-1), len(variadicArgs), len(variadicArgs))
		for i, arg := range variadicArgs {
			variadicSlice.Index(i).Set(arg)
		}

		args = append(args, variadicSlice)
	}

	return args, nil
}

// convertLuaValueToGo converts a Lua value to a Go value of the specified type
func (fc *FunctionConverter) convertLuaValueToGo(luaValue lua.LValue, targetType reflect.Type) (reflect.Value, error) {
	// Handle nil
	if luaValue == lua.LNil {
		if targetType.Kind() == reflect.Ptr || targetType.Kind() == reflect.Interface {
			return reflect.Zero(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot convert nil to %s", targetType.String())
	}

	// Handle interface{} - accept any converted value
	if targetType == reflect.TypeOf((*interface{})(nil)).Elem() {
		goValue, err := fc.complexConverter.FromLua(luaValue)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(goValue), nil
	}

	// Convert based on target type
	switch targetType.Kind() {
	case reflect.Bool:
		if lbool, ok := luaValue.(lua.LBool); ok {
			return reflect.ValueOf(bool(lbool)), nil
		}
		return reflect.Value{}, fmt.Errorf("expected boolean, got %s", luaValue.Type().String())

	case reflect.String:
		if lstr, ok := luaValue.(lua.LString); ok {
			return reflect.ValueOf(string(lstr)), nil
		}
		return reflect.Value{}, fmt.Errorf("expected string, got %s", luaValue.Type().String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if lnum, ok := luaValue.(lua.LNumber); ok {
			val := int64(lnum)
			return reflect.ValueOf(val).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", luaValue.Type().String())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if lnum, ok := luaValue.(lua.LNumber); ok {
			val := uint64(lnum)
			return reflect.ValueOf(val).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", luaValue.Type().String())

	case reflect.Float32, reflect.Float64:
		if lnum, ok := luaValue.(lua.LNumber); ok {
			val := float64(lnum)
			return reflect.ValueOf(val).Convert(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("expected number, got %s", luaValue.Type().String())

	case reflect.Slice:
		if ltable, ok := luaValue.(*lua.LTable); ok {
			goSlice, err := fc.complexConverter.luaTableToSlice(ltable, 0, make(map[*lua.LTable]bool))
			if err != nil {
				return reflect.Value{}, err
			}
			// Convert []interface{} to target slice type
			return fc.convertInterfaceSliceToTargetType(goSlice, targetType)
		}
		return reflect.Value{}, fmt.Errorf("expected table, got %s", luaValue.Type().String())

	case reflect.Map:
		if ltable, ok := luaValue.(*lua.LTable); ok {
			goMap, err := fc.complexConverter.luaTableToMap(ltable, 0, make(map[*lua.LTable]bool))
			if err != nil {
				return reflect.Value{}, err
			}
			// Convert map[string]interface{} to target map type
			return fc.convertInterfaceMapToTargetType(goMap, targetType)
		}
		return reflect.Value{}, fmt.Errorf("expected table, got %s", luaValue.Type().String())

	default:
		return reflect.Value{}, fmt.Errorf("unsupported target type: %s", targetType.String())
	}
}

// convertInterfaceSliceToTargetType converts []interface{} to the target slice type
func (fc *FunctionConverter) convertInterfaceSliceToTargetType(source []interface{}, targetType reflect.Type) (reflect.Value, error) {
	elemType := targetType.Elem()
	result := reflect.MakeSlice(targetType, len(source), len(source))

	for i, item := range source {
		itemValue := reflect.ValueOf(item)
		if itemValue.Type().ConvertibleTo(elemType) {
			result.Index(i).Set(itemValue.Convert(elemType))
		} else {
			return reflect.Value{}, fmt.Errorf("cannot convert slice element %d from %T to %s", i, item, elemType.String())
		}
	}

	return result, nil
}

// convertInterfaceMapToTargetType converts map[string]interface{} to the target map type
func (fc *FunctionConverter) convertInterfaceMapToTargetType(source map[string]interface{}, targetType reflect.Type) (reflect.Value, error) {
	keyType := targetType.Key()
	valueType := targetType.Elem()
	result := reflect.MakeMap(targetType)

	for k, v := range source {
		// Convert key
		var keyValue reflect.Value
		if keyType.Kind() == reflect.String {
			keyValue = reflect.ValueOf(k)
		} else {
			return reflect.Value{}, fmt.Errorf("unsupported map key type: %s", keyType.String())
		}

		// Convert value
		valueValue := reflect.ValueOf(v)
		if valueValue.Type().ConvertibleTo(valueType) {
			result.SetMapIndex(keyValue, valueValue.Convert(valueType))
		} else {
			return reflect.Value{}, fmt.Errorf("cannot convert map value for key %s from %T to %s", k, v, valueType.String())
		}
	}

	return result, nil
}

// handleResults converts Go function results to Lua values
func (fc *FunctionConverter) handleResults(L *lua.LState, results []reflect.Value, fnType reflect.Type) int {
	numResults := len(results)
	if numResults == 0 {
		return 0
	}

	// Check if the last result is an error
	hasError := false
	if numResults > 0 {
		lastResult := results[numResults-1]
		if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			hasError = true
			// Check if error is not nil
			if !lastResult.IsNil() {
				err := lastResult.Interface().(error)
				L.RaiseError("function returned error: %s", err.Error())
				return 0
			}
		}
	}

	// Convert non-error results to Lua
	resultCount := numResults
	if hasError {
		resultCount = numResults - 1
	}

	for i := 0; i < resultCount; i++ {
		result := results[i]
		luaValue, err := fc.convertGoValueToLua(L, result.Interface())
		if err != nil {
			L.RaiseError("return value conversion failed: %s", err.Error())
			return 0
		}
		L.Push(luaValue)
	}

	return resultCount
}

// convertGoValueToLua converts a Go value to a Lua value
func (fc *FunctionConverter) convertGoValueToLua(L *lua.LState, value interface{}) (lua.LValue, error) {
	if value == nil {
		return lua.LNil, nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return fc.primitiveConverter.BoolToLua(L, value)
	case reflect.String:
		return fc.primitiveConverter.StringToLua(L, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fc.primitiveConverter.NumberToLua(L, value)
	case reflect.Slice, reflect.Array:
		return fc.complexConverter.SliceToLua(L, value)
	case reflect.Map:
		return fc.complexConverter.MapToLua(L, value)
	case reflect.Struct:
		return fc.complexConverter.StructToLua(L, value)
	case reflect.Ptr:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return fc.convertGoValueToLua(L, rv.Elem().Interface())
	case reflect.Interface:
		if rv.IsNil() {
			return lua.LNil, nil
		}
		return fc.convertGoValueToLua(L, rv.Elem().Interface())
	default:
		return nil, fmt.Errorf("unsupported return type: %T", value)
	}
}
