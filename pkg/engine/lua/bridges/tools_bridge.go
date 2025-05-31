// ABOUTME: Lua bridge implementation for the tool system
// ABOUTME: Exposes tool registration and execution capabilities to Lua scripts

package bridges

import (
	"context"

	engLua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	lua "github.com/yuin/gopher-lua"
)

// RegisterToolsModule registers the tools module in Lua
func RegisterToolsModule(L *lua.LState, toolBridge ToolBridgeInterface) error {
	// Create tools module
	toolsMod := L.NewTable()

	// Create converter
	converter := engLua.NewLuaConverter(L)

	// Register functions
	L.SetField(toolsMod, "register", L.NewFunction(toolsRegister(toolBridge, converter)))
	L.SetField(toolsMod, "execute", L.NewFunction(toolsExecute(toolBridge, converter)))
	L.SetField(toolsMod, "get", L.NewFunction(toolsGet(toolBridge, converter)))
	L.SetField(toolsMod, "list", L.NewFunction(toolsList(toolBridge, converter)))
	L.SetField(toolsMod, "remove", L.NewFunction(toolsRemove(toolBridge)))
	L.SetField(toolsMod, "validate", L.NewFunction(toolsValidate(toolBridge, converter)))

	// Register the module
	L.SetGlobal("tools", toolsMod)
	return nil
}

// toolsRegister creates a Lua function for registering tools
func toolsRegister(tb ToolBridgeInterface, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get arguments
		name := L.CheckString(1)
		description := L.CheckString(2)

		// Get parameters table
		if L.Get(3).Type() != lua.LTTable {
			L.ArgError(3, "parameters must be a table")
			return 0
		}
		paramsInterface := converter.ToInterface(L.Get(3))
		params, ok := paramsInterface.(map[string]interface{})
		if !ok {
			L.ArgError(3, "parameters must be a table/object")
			return 0
		}

		// Get function
		if L.Get(4).Type() != lua.LTFunction {
			L.ArgError(4, "handler must be a function")
			return 0
		}
		fn := L.Get(4).(*lua.LFunction)

		// Create a Lua tool wrapper
		luaTool := NewLuaTool(name, description, params, fn, L, converter)

		// Create a Go function that delegates to the Lua tool
		goFunc := func(p map[string]interface{}) (interface{}, error) {
			return luaTool.Execute(context.Background(), p)
		}

		// Register the tool
		err := tb.RegisterTool(name, description, params, goFunc)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// toolsExecute creates a Lua function for executing tools
func toolsExecute(tb ToolBridgeInterface, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get arguments
		name := L.CheckString(1)

		// Get parameters
		var params map[string]interface{}
		if L.GetTop() >= 2 && L.Get(2).Type() == lua.LTTable {
			paramsInterface := converter.ToInterface(L.Get(2))
			params, _ = paramsInterface.(map[string]interface{})
			if params == nil {
				params = make(map[string]interface{})
			}
		} else {
			params = make(map[string]interface{})
		}

		// Execute the tool
		ctx := context.Background()
		result, err := tb.ExecuteTool(ctx, name, params)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua value
		L.Push(converter.ToLua(result))
		return 1
	}
}

// toolsGet creates a Lua function for getting tool information
func toolsGet(tb ToolBridgeInterface, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		info, err := tb.GetTool(name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert to Lua table
		L.Push(converter.ToLua(info))
		return 1
	}
}

// toolsList creates a Lua function for listing all tools
func toolsList(tb ToolBridgeInterface, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		tools := tb.ListTools()

		// Convert to Lua value
		L.Push(converter.ToLua(tools))
		return 1
	}
}

// toolsRemove creates a Lua function for removing tools
func toolsRemove(tb ToolBridgeInterface) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		err := tb.RemoveTool(name)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// toolsValidate creates a Lua function for validating parameters
func toolsValidate(tb ToolBridgeInterface, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get arguments
		name := L.CheckString(1)

		// Get parameters
		var params map[string]interface{}
		if L.GetTop() >= 2 && L.Get(2).Type() == lua.LTTable {
			paramsInterface := converter.ToInterface(L.Get(2))
			params, _ = paramsInterface.(map[string]interface{})
			if params == nil {
				params = make(map[string]interface{})
			}
		} else {
			params = make(map[string]interface{})
		}

		// Validate parameters
		err := tb.ValidateParameters(name, params)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}
