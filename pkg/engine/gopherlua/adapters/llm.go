// ABOUTME: LLM bridge adapter that exposes go-llms LLM functionality to Lua scripts
// ABOUTME: Provides agent creation, completion methods, streaming, model selection, and token counting

package adapters

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// LLMAdapter specializes BridgeAdapter for LLM functionality
type LLMAdapter struct {
	*gopherlua.BridgeAdapter
}

// NewLLMAdapter creates a new LLM adapter
func NewLLMAdapter(bridge engine.Bridge) *LLMAdapter {
	// Create base adapter
	baseAdapter := gopherlua.NewBridgeAdapter(bridge)

	// Create LLM adapter
	adapter := &LLMAdapter{
		BridgeAdapter: baseAdapter,
	}

	// Add LLM-specific methods if not already present
	adapter.ensureLLMMethods()

	return adapter
}

// ensureLLMMethods ensures LLM-specific methods are available
func (la *LLMAdapter) ensureLLMMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
	// In production, this could validate that expected LLM methods exist
}

// CreateLuaModule creates a Lua module with LLM-specific enhancements
func (la *LLMAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Get base module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(la.BridgeAdapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		if err != nil {
			L.RaiseError("failed to create base module: %v", err)
			return 0
		}

		// Get the module
		module := L.Get(-1).(*lua.LTable)

		// Add LLM-specific enhancements
		la.addLLMEnhancements(L, module)

		// Module is already on stack
		return 1
	}
}

// addLLMEnhancements adds LLM-specific enhancements to the module
func (la *LLMAdapter) addLLMEnhancements(L *lua.LState, module *lua.LTable) {
	// Add constructor alias
	if agentCreate := module.RawGetString("createAgent"); agentCreate != lua.LNil {
		L.SetField(module, "Agent", agentCreate)
	}

	// Add convenience methods
	la.addConvenienceMethods(L, module)

	// Add constants
	la.addConstants(L, module)
}

// addConvenienceMethods adds convenience methods to the module
func (la *LLMAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// Add quick completion method that uses default model
	L.SetField(module, "quick", L.NewFunction(func(L *lua.LState) int {
		prompt := L.CheckString(1)

		// Call complete with just the prompt
		completeFn := module.RawGetString("complete")
		if completeFn == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("complete method not found"))
			return 2
		}

		// Call the complete function
		err := L.CallByParam(lua.P{
			Fn:      completeFn,
			NRet:    2,
			Protect: true,
		}, lua.LString(prompt))

		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Results are already on stack
		return 2
	}))

	// Add batch completion method
	L.SetField(module, "batchComplete", L.NewFunction(func(L *lua.LState) int {
		prompts := L.CheckTable(1)
		options := L.OptTable(2, L.NewTable())

		results := L.NewTable()
		var lastError string

		// Process each prompt
		prompts.ForEach(func(k, v lua.LValue) {
			if str, ok := v.(lua.LString); ok {
				// Call complete for this prompt
				completeFn := module.RawGetString("complete")
				if completeFn != lua.LNil {
					err := L.CallByParam(lua.P{
						Fn:      completeFn,
						NRet:    2,
						Protect: true,
					}, str, options)

					if err == nil {
						result := L.Get(-2)
						resultErr := L.Get(-1)
						L.Pop(2)

						if resultErr == lua.LNil {
							results.Append(result)
						} else {
							lastError = resultErr.String()
							results.Append(lua.LNil)
						}
					} else {
						lastError = err.Error()
						results.Append(lua.LNil)
					}
				}
			}
		})

		L.Push(results)
		if lastError != "" {
			L.Push(lua.LString(lastError))
		} else {
			L.Push(lua.LNil)
		}
		return 2
	}))
}

// addConstants adds LLM-related constants to the module
func (la *LLMAdapter) addConstants(L *lua.LState, module *lua.LTable) {
	// Add model constants
	models := L.NewTable()
	L.SetField(models, "GPT4", lua.LString("gpt-4"))
	L.SetField(models, "GPT35_TURBO", lua.LString("gpt-3.5-turbo"))
	L.SetField(models, "CLAUDE3", lua.LString("claude-3"))
	L.SetField(models, "CLAUDE2", lua.LString("claude-2"))
	L.SetField(module, "MODELS", models)

	// Add default options
	defaults := L.NewTable()
	L.SetField(defaults, "temperature", lua.LNumber(0.7))
	L.SetField(defaults, "maxTokens", lua.LNumber(1000))
	L.SetField(defaults, "topP", lua.LNumber(1.0))
	L.SetField(module, "DEFAULTS", defaults)

	// Add error codes
	errors := L.NewTable()
	L.SetField(errors, "RATE_LIMIT", lua.LString("rate_limit_exceeded"))
	L.SetField(errors, "INVALID_MODEL", lua.LString("invalid_model"))
	L.SetField(errors, "CONTEXT_LENGTH", lua.LString("context_length_exceeded"))
	L.SetField(module, "ERRORS", errors)
}

// WrapMethod wraps a bridge method with LLM-specific handling
func (la *LLMAdapter) WrapMethod(methodName string) lua.LGFunction {
	// Get base wrapped method
	baseWrapped := la.BridgeAdapter.WrapMethod(methodName)

	// Add LLM-specific handling for certain methods
	switch methodName {
	case "createAgent":
		return la.wrapCreateAgent(baseWrapped)
	case "complete":
		return la.wrapComplete(baseWrapped)
	case "stream":
		return la.wrapStream(baseWrapped)
	default:
		return baseWrapped
	}
}

// wrapCreateAgent adds agent-specific handling
func (la *LLMAdapter) wrapCreateAgent(baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure config is provided
		if L.GetTop() == 0 {
			L.Push(L.NewTable()) // Empty config
		}

		// Call base function
		returnCount := baseFn(L)

		// If successful, enhance the agent object
		if returnCount > 0 && L.Get(-returnCount).Type() == lua.LTTable {
			agent := L.Get(-returnCount).(*lua.LTable)
			la.enhanceAgentObject(L, agent)
		}

		return returnCount
	}
}

// wrapComplete adds completion-specific handling
func (la *LLMAdapter) wrapComplete(baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least prompt is provided
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("prompt is required"))
			return 2
		}

		// If no options provided, add empty table
		if L.GetTop() == 1 {
			L.Push(L.NewTable())
		}

		return baseFn(L)
	}
}

// wrapStream adds streaming-specific handling
func (la *LLMAdapter) wrapStream(baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure callback is provided in options
		if L.GetTop() >= 2 {
			options := L.Get(2)
			if options.Type() == lua.LTTable {
				optTable := options.(*lua.LTable)
				callback := optTable.RawGetString("onChunk")
				if callback == lua.LNil {
					// Add default callback that collects chunks
					chunks := L.NewTable()
					L.SetField(optTable, "_chunks", chunks)
					L.SetField(optTable, "onChunk", L.NewFunction(func(L *lua.LState) int {
						chunk := L.Get(1)
						chunks.Append(chunk)
						return 0
					}))
				}
			}
		}

		return baseFn(L)
	}
}

// enhanceAgentObject adds methods to the agent object
func (la *LLMAdapter) enhanceAgentObject(L *lua.LState, agent *lua.LTable) {
	// Add complete method to agent
	L.SetField(agent, "complete", L.NewFunction(func(L *lua.LState) int {
		prompt := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		// Get agent ID
		agentId := agent.RawGetString("id")
		if agentId == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("agent has no id"))
			return 2
		}

		// Call agentComplete through the module
		module := L.GetGlobal("llm")
		if module.Type() != lua.LTTable {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm module not found"))
			return 2
		}

		agentCompleteFn := module.(*lua.LTable).RawGetString("agentComplete")
		if agentCompleteFn == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("agentComplete not found"))
			return 2
		}

		// Call agentComplete
		err := L.CallByParam(lua.P{
			Fn:      agentCompleteFn,
			NRet:    2,
			Protect: true,
		}, agentId, lua.LString(prompt), options)

		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		return 2
	}))

	// Add info method
	L.SetField(agent, "info", L.NewFunction(func(L *lua.LState) int {
		info := L.NewTable()
		L.SetField(info, "id", agent.RawGetString("id"))
		L.SetField(info, "model", agent.RawGetString("model"))
		L.SetField(info, "type", agent.RawGetString("type"))
		L.Push(info)
		return 1
	}))
}
