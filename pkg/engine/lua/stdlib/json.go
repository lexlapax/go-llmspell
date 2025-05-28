// ABOUTME: JSON encoding/decoding module for Lua scripts
// ABOUTME: Provides json.encode() and json.decode() functions

package stdlib

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

// RegisterJSON registers the JSON module with encode/decode functions
func RegisterJSON(L *lua.LState) {
	// Create json module table
	jsonModule := L.NewTable()

	// Register functions
	L.SetField(jsonModule, "encode", L.NewFunction(jsonEncode))
	L.SetField(jsonModule, "decode", L.NewFunction(jsonDecode))

	// Register the module
	L.SetGlobal("json", jsonModule)
}

// jsonEncode encodes a Lua value to JSON string
// Usage: json_str = json.encode(value)
func jsonEncode(L *lua.LState) int {
	value := L.Get(1)

	// Convert Lua value to Go value
	goValue := luaToGo(value)

	// Encode to JSON
	jsonBytes, err := json.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(jsonBytes)))
	return 1
}

// jsonDecode decodes a JSON string to Lua value
// Usage: value, err = json.decode(json_str)
func jsonDecode(L *lua.LState) int {
	jsonStr := L.CheckString(1)

	// Decode JSON
	var goValue interface{}
	err := json.Unmarshal([]byte(jsonStr), &goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert Go value to Lua value
	luaValue := goToLua(L, goValue)
	L.Push(luaValue)
	return 1
}

// luaToGo converts a Lua value to a Go value for JSON encoding
func luaToGo(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Check if it's an array or object
		maxIndex := 0
		length := 0
		isArray := true

		v.ForEach(func(key, value lua.LValue) {
			if _, ok := key.(lua.LNumber); ok {
				length++
				if int(key.(lua.LNumber)) > maxIndex {
					maxIndex = int(key.(lua.LNumber))
				}
			} else {
				isArray = false
			}
		})

		// If it's a pure array (1-indexed, consecutive)
		if isArray && length > 0 && maxIndex == length {
			arr := make([]interface{}, length)
			v.ForEach(func(key, value lua.LValue) {
				if idx, ok := key.(lua.LNumber); ok {
					arr[int(idx)-1] = luaToGo(value)
				}
			})
			return arr
		}

		// Otherwise treat as object
		obj := make(map[string]interface{})
		v.ForEach(func(key, value lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				obj[string(keyStr)] = luaToGo(value)
			} else if keyNum, ok := key.(lua.LNumber); ok {
				obj[string(keyNum.String())] = luaToGo(value)
			}
		})
		return obj

	default:
		return nil
	}
}

// goToLua converts a Go value to a Lua value for JSON decoding
func goToLua(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}

	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		table := L.NewTable()
		for i, elem := range v {
			table.RawSetInt(i+1, goToLua(L, elem))
		}
		return table
	case map[string]interface{}:
		table := L.NewTable()
		for key, val := range v {
			L.SetField(table, key, goToLua(L, val))
		}
		return table
	default:
		return lua.LNil
	}
}
