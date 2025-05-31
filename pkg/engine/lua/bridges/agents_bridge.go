// ABOUTME: Lua bridge implementation for the agent system
// ABOUTME: Exposes agent creation, execution, and management capabilities to Lua scripts

package bridges

import (
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/agents"
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	engLua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	lua "github.com/yuin/gopher-lua"
)

// luaAgentRegistry stores Lua-based agents
// This is a workaround since the main agent registry is designed for provider-based factories
var luaAgentRegistry sync.Map

// RegisterAgentsModule registers the agents module in Lua
func RegisterAgentsModule(L *lua.LState, agentBridge bridge.AgentBridge) error {
	// Create agents module
	agentsMod := L.NewTable()

	// Create converter
	converter := engLua.NewLuaConverter(L)

	// Register functions
	L.SetField(agentsMod, "create", L.NewFunction(agentsCreate(agentBridge, converter)))
	L.SetField(agentsMod, "execute", L.NewFunction(agentsExecute(agentBridge, converter)))
	L.SetField(agentsMod, "stream", L.NewFunction(agentsStream(agentBridge, converter)))
	L.SetField(agentsMod, "list", L.NewFunction(agentsList(agentBridge, converter)))
	L.SetField(agentsMod, "get", L.NewFunction(agentsGetInfo(agentBridge, converter)))
	L.SetField(agentsMod, "remove", L.NewFunction(agentsRemove(agentBridge)))
	L.SetField(agentsMod, "update_system_prompt", L.NewFunction(agentsUpdateSystemPrompt(agentBridge)))
	L.SetField(agentsMod, "add_tool", L.NewFunction(agentsAddTool(agentBridge)))
	L.SetField(agentsMod, "register", L.NewFunction(agentsRegister(L)))

	// Register the module
	L.SetGlobal("agents", agentsMod)
	return nil
}

// agentsCreate creates a Lua function for creating agents
func agentsCreate(ab bridge.AgentBridge, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get configuration table
		if L.Get(1).Type() != lua.LTTable {
			L.ArgError(1, "configuration must be a table")
			return 0
		}
		configInterface := converter.ToInterface(L.Get(1))
		config, ok := configInterface.(map[string]interface{})
		if !ok {
			L.ArgError(1, "configuration must be a table/object")
			return 0
		}

		// Create the agent
		name, err := ab.Create(config)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(name))
		return 1
	}
}

// agentsExecute creates a Lua function for executing agents
func agentsExecute(ab bridge.AgentBridge, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get arguments
		agentName := L.CheckString(1)
		input := L.CheckString(2)

		// Get options (optional)
		var options map[string]interface{}
		if L.GetTop() >= 3 && L.Get(3).Type() == lua.LTTable {
			optionsInterface := converter.ToInterface(L.Get(3))
			options, _ = optionsInterface.(map[string]interface{})
		}

		// Execute the agent
		result, err := ab.Execute(agentName, input, options)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(result))
		return 1
	}
}

// agentsStream creates a Lua function for streaming agent execution
func agentsStream(ab bridge.AgentBridge, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get arguments
		agentName := L.CheckString(1)
		input := L.CheckString(2)

		// Get callback function
		if L.Get(3).Type() != lua.LTFunction {
			L.ArgError(3, "callback must be a function")
			return 0
		}
		callbackFn := L.Get(3).(*lua.LFunction)

		// Get options (optional)
		var options map[string]interface{}
		if L.GetTop() >= 4 && L.Get(4).Type() == lua.LTTable {
			optionsInterface := converter.ToInterface(L.Get(4))
			options, _ = optionsInterface.(map[string]interface{})
		}

		// Create Go callback that calls Lua function
		callback := func(chunk string) error {
			// Call the Lua callback directly
			err := L.CallByParam(lua.P{
				Fn:      callbackFn,
				NRet:    1,
				Protect: true,
			}, lua.LString(chunk))

			if err != nil {
				return fmt.Errorf("callback error: %v", err)
			}

			// Check if callback returned false to stop streaming
			ret := L.Get(-1)
			L.Pop(1)

			if ret == lua.LFalse {
				return fmt.Errorf("streaming stopped by callback")
			}

			return nil
		}

		// Stream the agent execution
		err := ab.Stream(agentName, input, options, callback)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// agentsList creates a Lua function for listing all agents
func agentsList(ab bridge.AgentBridge, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		agents := ab.List()

		// Convert to Lua value
		L.Push(converter.ToLua(agents))
		return 1
	}
}

// agentsGetInfo creates a Lua function for getting agent information
func agentsGetInfo(ab bridge.AgentBridge, converter *engLua.LuaConverter) lua.LGFunction {
	return func(L *lua.LState) int {
		agentName := L.CheckString(1)

		info, err := ab.GetInfo(agentName)
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

// agentsRemove creates a Lua function for removing agents
func agentsRemove(ab bridge.AgentBridge) lua.LGFunction {
	return func(L *lua.LState) int {
		agentName := L.CheckString(1)

		err := ab.Remove(agentName)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// agentsUpdateSystemPrompt creates a Lua function for updating system prompts
func agentsUpdateSystemPrompt(ab bridge.AgentBridge) lua.LGFunction {
	return func(L *lua.LState) int {
		agentName := L.CheckString(1)
		prompt := L.CheckString(2)

		err := ab.UpdateSystemPrompt(agentName, prompt)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// agentsAddTool creates a Lua function for adding tools to agents
func agentsAddTool(ab bridge.AgentBridge) lua.LGFunction {
	return func(L *lua.LState) int {
		agentName := L.CheckString(1)
		toolName := L.CheckString(2)

		err := ab.AddTool(agentName, toolName)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// agentsRegister creates a Lua function for registering Lua-based agents
func agentsRegister(L *lua.LState) lua.LGFunction {
	return func(L *lua.LState) int {
		// Get agent name
		name := L.CheckString(1)

		// Get agent implementation (function or table)
		agentImpl := L.Get(2)
		if agentImpl.Type() != lua.LTFunction && agentImpl.Type() != lua.LTTable {
			L.ArgError(2, "agent implementation must be a function or table")
			return 0
		}

		// Store the Lua agent in our registry
		luaAgentRegistry.Store(name, agentImpl)

		// Register a factory for this specific agent
		registry := agents.DefaultRegistry()

		// Create a unique provider name for this Lua agent
		providerName := "lua-" + name

		// Register the factory
		err := registry.Register(providerName, func(config agents.Config) (agents.Agent, error) {
			// Retrieve the stored Lua implementation
			if impl, ok := luaAgentRegistry.Load(config.Name); ok {
				return NewLuaAgent(config.Name, impl.(lua.LValue), L), nil
			}
			return nil, fmt.Errorf("lua agent implementation not found for %s", config.Name)
		})

		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Create the agent instance through the registry
		config := agents.Config{
			Name:     name,
			Provider: providerName,
			Model:    "lua-script",
		}

		_, err = registry.Create(config)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}
