// ABOUTME: ScriptValue to Lua LValue bi-directional converter for the GopherLua engine
// ABOUTME: Provides seamless conversion between ScriptValue system and Lua types with circular reference detection

package gopherlua

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// ScriptValueConverter handles conversion between ScriptValue and lua.LValue
type ScriptValueConverter struct {
	mu        sync.RWMutex
	maxDepth  int
	converter *LuaTypeConverter // Reference to existing converter for custom types
}

// NewScriptValueConverter creates a new ScriptValue converter
func NewScriptValueConverter(converter *LuaTypeConverter) *ScriptValueConverter {
	return &ScriptValueConverter{
		maxDepth:  32,
		converter: converter,
	}
}

// LValueToScriptValue converts a lua.LValue to a ScriptValue
func (c *ScriptValueConverter) LValueToScriptValue(L *lua.LState, lv lua.LValue) (engine.ScriptValue, error) {
	return c.lValueToScriptValueWithDepth(L, lv, 0, make(map[uintptr]bool))
}

// lValueToScriptValueWithDepth performs the actual conversion with depth tracking
func (c *ScriptValueConverter) lValueToScriptValueWithDepth(L *lua.LState, lv lua.LValue, depth int, visited map[uintptr]bool) (engine.ScriptValue, error) {
	// Check depth limit
	if depth > c.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", c.maxDepth)
	}

	// Handle nil first
	if lv == lua.LNil {
		return engine.NewNilValue(), nil
	}

	switch v := lv.(type) {
	case lua.LBool:
		return engine.NewBoolValue(bool(v)), nil

	case lua.LNumber:
		return engine.NewNumberValue(float64(v)), nil

	case lua.LString:
		return engine.NewStringValue(string(v)), nil

	case *lua.LTable:
		return c.tableToScriptValue(L, v, depth, visited)

	case *lua.LFunction:
		// Create a wrapper function that can be called from ScriptValue
		wrapper := &luaFunctionWrapper{
			L:         L,
			function:  v,
			converter: c,
		}
		return engine.NewFunctionValue("lua_function", wrapper), nil

	case *lua.LUserData:
		// Handle user data by extracting the underlying Go value
		if userData := v.Value; userData != nil {
			// Convert the underlying Go value to ScriptValue using our own method
			return c.GoToScriptValue(userData)
		}
		return engine.NewCustomValue("userdata", v), nil

	// Note: gopher-lua doesn't have LChannel, so we handle this differently
	// case *lua.LChannel:
	//	return engine.NewChannelValue("lua_channel", v), nil

	default:
		return engine.NewCustomValue("lua_value", v), nil
	}
}

// tableToScriptValue converts a Lua table to ScriptValue (either ArrayValue or ObjectValue)
func (c *ScriptValueConverter) tableToScriptValue(L *lua.LState, table *lua.LTable, depth int, visited map[uintptr]bool) (engine.ScriptValue, error) {
	// Check for circular references using the table's address
	tablePtr := uintptr(unsafe.Pointer(table))
	if visited[tablePtr] {
		return nil, fmt.Errorf("circular reference detected in table")
	}
	visited[tablePtr] = true
	defer delete(visited, tablePtr)

	// Check if table is array-like (consecutive integer keys starting from 1)
	isArray := true
	arrayLen := table.Len()

	// If table has length > 0, check if it's array-like
	if arrayLen > 0 {
		for i := 1; i <= arrayLen; i++ {
			if table.RawGetInt(i) == lua.LNil {
				isArray = false
				break
			}
		}

		// Also check if there are non-integer keys
		if isArray {
			table.ForEach(func(key, value lua.LValue) {
				if lnum, ok := key.(lua.LNumber); ok {
					if float64(lnum) != float64(int(lnum)) || int(lnum) < 1 || int(lnum) > arrayLen {
						isArray = false
					}
				} else {
					isArray = false
				}
			})
		}
	} else {
		// Empty table or table with no consecutive integer keys
		hasIntegerKeys := false
		table.ForEach(func(key, value lua.LValue) {
			if _, ok := key.(lua.LNumber); ok {
				hasIntegerKeys = true
			}
		})
		isArray = !hasIntegerKeys || arrayLen == 0
	}

	if isArray && arrayLen > 0 {
		// Convert to ArrayValue
		elements := make([]engine.ScriptValue, arrayLen)
		for i := 1; i <= arrayLen; i++ {
			lval := table.RawGetInt(i)
			scriptVal, err := c.lValueToScriptValueWithDepth(L, lval, depth+1, visited)
			if err != nil {
				return nil, fmt.Errorf("error converting array element %d: %w", i-1, err)
			}
			elements[i-1] = scriptVal
		}
		return engine.NewArrayValue(elements), nil
	} else {
		// Convert to ObjectValue
		fields := make(map[string]engine.ScriptValue)
		var conversionError error

		table.ForEach(func(key, value lua.LValue) {
			if conversionError != nil {
				return
			}

			// Convert key to string
			var keyStr string
			switch k := key.(type) {
			case lua.LString:
				keyStr = string(k)
			case lua.LNumber:
				keyStr = fmt.Sprintf("%g", float64(k))
			default:
				keyStr = key.String()
			}

			// Convert value
			scriptVal, err := c.lValueToScriptValueWithDepth(L, value, depth+1, visited)
			if err != nil {
				conversionError = fmt.Errorf("error converting object field %s: %w", keyStr, err)
				return
			}
			fields[keyStr] = scriptVal
		})

		if conversionError != nil {
			return nil, conversionError
		}

		return engine.NewObjectValue(fields), nil
	}
}

// ScriptValueToLValue converts a ScriptValue to a lua.LValue
func (c *ScriptValueConverter) ScriptValueToLValue(L *lua.LState, sv engine.ScriptValue) (lua.LValue, error) {
	return c.scriptValueToLValueWithDepth(L, sv, 0, make(map[uintptr]bool))
}

// scriptValueToLValueWithDepth performs the actual conversion with depth tracking
func (c *ScriptValueConverter) scriptValueToLValueWithDepth(L *lua.LState, sv engine.ScriptValue, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	// Check depth limit
	if depth > c.maxDepth {
		return nil, fmt.Errorf("maximum depth exceeded: %d", c.maxDepth)
	}

	if sv == nil || sv.IsNil() {
		return lua.LNil, nil
	}

	switch sv.Type() {
	case engine.TypeBool:
		if bv, ok := sv.(engine.BoolValue); ok {
			return lua.LBool(bv.Value()), nil
		}
		return lua.LBool(engine.IsTrue(sv)), nil

	case engine.TypeNumber:
		if nv, ok := sv.(engine.NumberValue); ok {
			return lua.LNumber(nv.Value()), nil
		}
		// Fallback conversion
		if num, err := engine.ConvertToNumber(sv); err == nil {
			return lua.LNumber(num), nil
		}
		return nil, fmt.Errorf("cannot convert %s to number", sv.Type())

	case engine.TypeString:
		if strv, ok := sv.(engine.StringValue); ok {
			return lua.LString(strv.Value()), nil
		}
		// Fallback conversion
		if str, err := engine.ConvertToString(sv); err == nil {
			return lua.LString(str), nil
		}
		return lua.LString(sv.String()), nil

	case engine.TypeArray:
		if av, ok := sv.(engine.ArrayValue); ok {
			return c.arrayToLuaTable(L, av, depth, visited)
		}
		return nil, fmt.Errorf("invalid ArrayValue type")

	case engine.TypeObject:
		if ov, ok := sv.(engine.ObjectValue); ok {
			return c.objectToLuaTable(L, ov, depth, visited)
		}
		return nil, fmt.Errorf("invalid ObjectValue type")

	case engine.TypeFunction:
		if fv, ok := sv.(engine.FunctionValue); ok {
			return c.functionToLuaFunction(L, fv)
		}
		return nil, fmt.Errorf("invalid FunctionValue type")

	case engine.TypeError:
		if ev, ok := sv.(engine.ErrorValue); ok {
			// Return the error as a string value in Lua
			if ev.Error() != nil {
				return lua.LString(ev.Error().Error()), nil
			}
			return lua.LString("unknown error"), nil
		}
		return lua.LString("error"), nil

	case engine.TypeChannel:
		if cv, ok := sv.(engine.ChannelValue); ok {
			// For now, we'll represent channels as user data since gopher-lua doesn't have built-in channels
			userData := L.NewUserData()
			userData.Value = cv.Value()
			return userData, nil
		}
		// Create a basic user data to represent the channel
		userData := L.NewUserData()
		userData.Value = make(chan interface{}, 1)
		return userData, nil

	case engine.TypeCustom:
		if cv, ok := sv.(engine.CustomValue); ok {
			// Try to convert the underlying value using the existing converter
			if c.converter != nil {
				return c.converter.ToLua(L, cv.Value())
			}
			// Fallback: create user data
			userData := L.NewUserData()
			userData.Value = cv.Value()
			return userData, nil
		}
		return nil, fmt.Errorf("invalid CustomValue type")

	default:
		return nil, fmt.Errorf("unsupported ScriptValue type: %s", sv.Type())
	}
}

// arrayToLuaTable converts an ArrayValue to a Lua table
func (c *ScriptValueConverter) arrayToLuaTable(L *lua.LState, av engine.ArrayValue, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	elements := av.Elements()
	table := L.NewTable()

	for i, elem := range elements {
		lval, err := c.scriptValueToLValueWithDepth(L, elem, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("error converting array element %d: %w", i, err)
		}
		table.RawSetInt(i+1, lval) // Lua arrays start at 1
	}

	return table, nil
}

// objectToLuaTable converts an ObjectValue to a Lua table
func (c *ScriptValueConverter) objectToLuaTable(L *lua.LState, ov engine.ObjectValue, depth int, visited map[uintptr]bool) (lua.LValue, error) {
	fields := ov.Fields()
	table := L.NewTable()

	for key, value := range fields {
		lval, err := c.scriptValueToLValueWithDepth(L, value, depth+1, visited)
		if err != nil {
			return nil, fmt.Errorf("error converting object field %s: %w", key, err)
		}
		table.RawSetString(key, lval)
	}

	return table, nil
}

// functionToLuaFunction converts a FunctionValue to a Lua function
func (c *ScriptValueConverter) functionToLuaFunction(L *lua.LState, fv engine.FunctionValue) (lua.LValue, error) {
	// Check if it's already a Lua function wrapper
	if wrapper, ok := fv.Function().(*luaFunctionWrapper); ok {
		return wrapper.function, nil
	}

	// Create a new Lua function that wraps the ScriptValue function
	fn := L.NewFunction(func(L *lua.LState) int {
		// Get arguments
		argCount := L.GetTop()
		args := make([]engine.ScriptValue, argCount)

		for i := 1; i <= argCount; i++ {
			arg := L.Get(i)
			scriptVal, err := c.LValueToScriptValue(L, arg)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("error converting argument %d: %v", i, err)))
				return 1
			}
			args[i-1] = scriptVal
		}

		// Call the function
		result, err := fv.Call(args)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

		// Convert result back to Lua
		if result != nil {
			lval, err := c.ScriptValueToLValue(L, result)
			if err != nil {
				L.Push(lua.LString(fmt.Sprintf("error converting result: %v", err)))
				return 1
			}
			L.Push(lval)
			return 1
		}

		return 0
	})

	return fn, nil
}

// luaFunctionWrapper wraps a Lua function for ScriptValue calls
type luaFunctionWrapper struct {
	L         *lua.LState
	function  *lua.LFunction
	converter *ScriptValueConverter
}

// Call implements the function call for ScriptValue
func (w *luaFunctionWrapper) Call(args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Convert ScriptValue arguments to Lua values
	luaArgs := make([]lua.LValue, len(args))
	for i, arg := range args {
		lval, err := w.converter.ScriptValueToLValue(w.L, arg)
		if err != nil {
			return nil, fmt.Errorf("error converting argument %d: %w", i, err)
		}
		luaArgs[i] = lval
	}

	// Call the Lua function
	err := w.L.CallByParam(lua.P{
		Fn:      w.function,
		NRet:    1,
		Protect: true,
	}, luaArgs...)

	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Convert result back to ScriptValue
	result := w.L.Get(-1)
	w.L.Pop(1)

	return w.converter.LValueToScriptValue(w.L, result)
}

// GoToScriptValue converts a Go interface{} value to ScriptValue using the existing converter
func (c *ScriptValueConverter) GoToScriptValue(value interface{}) (engine.ScriptValue, error) {
	if value == nil {
		return engine.NewNilValue(), nil
	}

	switch v := value.(type) {
	case bool:
		return engine.NewBoolValue(v), nil
	case int:
		return engine.NewNumberValue(float64(v)), nil
	case int8:
		return engine.NewNumberValue(float64(v)), nil
	case int16:
		return engine.NewNumberValue(float64(v)), nil
	case int32:
		return engine.NewNumberValue(float64(v)), nil
	case int64:
		return engine.NewNumberValue(float64(v)), nil
	case uint:
		return engine.NewNumberValue(float64(v)), nil
	case uint8:
		return engine.NewNumberValue(float64(v)), nil
	case uint16:
		return engine.NewNumberValue(float64(v)), nil
	case uint32:
		return engine.NewNumberValue(float64(v)), nil
	case uint64:
		return engine.NewNumberValue(float64(v)), nil
	case float32:
		return engine.NewNumberValue(float64(v)), nil
	case float64:
		return engine.NewNumberValue(v), nil
	case string:
		return engine.NewStringValue(v), nil
	case []interface{}:
		elements := make([]engine.ScriptValue, len(v))
		for i, elem := range v {
			scriptVal, err := c.GoToScriptValue(elem)
			if err != nil {
				return nil, fmt.Errorf("error converting array element %d: %w", i, err)
			}
			elements[i] = scriptVal
		}
		return engine.NewArrayValue(elements), nil
	case map[string]interface{}:
		fields := make(map[string]engine.ScriptValue)
		for key, val := range v {
			scriptVal, err := c.GoToScriptValue(val)
			if err != nil {
				return nil, fmt.Errorf("error converting object field %s: %w", key, err)
			}
			fields[key] = scriptVal
		}
		return engine.NewObjectValue(fields), nil
	case error:
		return engine.NewErrorValue(v), nil
	default:
		return engine.NewCustomValue(fmt.Sprintf("%T", v), v), nil
	}
}

// ScriptValueToGo converts a ScriptValue to a Go interface{} value
func (c *ScriptValueConverter) ScriptValueToGo(sv engine.ScriptValue) interface{} {
	if sv == nil || sv.IsNil() {
		return nil
	}
	return sv.ToGo()
}
