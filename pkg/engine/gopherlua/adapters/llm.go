// ABOUTME: LLM bridge adapter that exposes go-llms LLM functionality to Lua scripts
// ABOUTME: Provides agent creation, completion methods, streaming, model selection, provider management, and pool functionality

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// LLMAdapter specializes BridgeAdapter for LLM functionality
type LLMAdapter struct {
	*gopherlua.BridgeAdapter

	// References to related bridges for enhanced functionality
	providersBridge engine.Bridge
	poolBridge      engine.Bridge
}

// NewLLMAdapter creates a new LLM adapter with optional related bridges
func NewLLMAdapter(bridge engine.Bridge, providersBridge engine.Bridge, poolBridge engine.Bridge) *LLMAdapter {
	// Create base adapter
	baseAdapter := gopherlua.NewBridgeAdapter(bridge)

	// Create LLM adapter
	adapter := &LLMAdapter{
		BridgeAdapter:   baseAdapter,
		providersBridge: providersBridge,
		poolBridge:      poolBridge,
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

		// Add provider management methods
		la.addProviderMethods(L, module)

		// Add pool management methods
		la.addPoolMethods(L, module)

		// Add model management methods
		la.addModelMethods(L, module)

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

// addProviderMethods adds provider management methods
func (la *LLMAdapter) addProviderMethods(L *lua.LState, module *lua.LTable) {
	// Create provider namespace
	providers := L.NewTable()

	// createProvider method
	L.SetField(providers, "create", L.NewFunction(func(L *lua.LState) int {
		providerType := L.CheckString(1)
		name := L.CheckString(2)
		config := L.OptTable(3, L.NewTable())

		// Convert config to map
		configMap := la.tableToMap(config)

		// Call provider creation through main bridge
		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(providerType),
			engine.NewStringValue(name),
			engine.NewObjectValue(configMap),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "createProvider", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getProvider method
	L.SetField(providers, "get", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(name)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "getProvider", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// listProviders method
	L.SetField(providers, "list", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := la.GetBridge().ExecuteMethod(ctx, "listProviders", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add provider templates support
	L.SetField(providers, "getTemplate", L.NewFunction(func(L *lua.LState) int {
		templateName := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(templateName)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "getProviderTemplate", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Multi-provider support
	L.SetField(providers, "createMulti", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		providerList := L.CheckTable(2)
		strategy := L.CheckString(3)
		config := L.OptTable(4, L.NewTable())

		// Convert provider list to array
		var providerNames []engine.ScriptValue
		providerList.ForEach(func(k, v lua.LValue) {
			if str, ok := v.(lua.LString); ok {
				providerNames = append(providerNames, engine.NewStringValue(string(str)))
			}
		})

		// Convert config to map
		configMap := la.tableToMap(config)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			engine.NewArrayValue(providerNames),
			engine.NewStringValue(strategy),
			engine.NewObjectValue(configMap),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "createMultiProvider", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add providers namespace to module
	L.SetField(module, "providers", providers)
}

// addPoolMethods adds pool management methods
func (la *LLMAdapter) addPoolMethods(L *lua.LState, module *lua.LTable) {
	// Create pool namespace
	pool := L.NewTable()

	// createPool method
	L.SetField(pool, "create", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		providers := L.CheckTable(2)
		strategy := L.CheckString(3)
		config := L.OptTable(4, L.NewTable())

		// Convert providers to array
		var providerList []engine.ScriptValue
		providers.ForEach(func(k, v lua.LValue) {
			if str, ok := v.(lua.LString); ok {
				providerList = append(providerList, engine.NewStringValue(string(str)))
			}
		})

		// Convert config to map
		configMap := la.tableToMap(config)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			engine.NewArrayValue(providerList),
			engine.NewStringValue(strategy),
			engine.NewObjectValue(configMap),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "createPool", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getPoolHealth method
	L.SetField(pool, "getHealth", L.NewFunction(func(L *lua.LState) int {
		poolName := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(poolName)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "getPoolHealth", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// generateWithPool method
	L.SetField(pool, "generate", L.NewFunction(func(L *lua.LState) int {
		poolName := L.CheckString(1)
		prompt := L.CheckString(2)
		options := L.OptTable(3, L.NewTable())

		// Convert options to map
		optionsMap := la.tableToMap(options)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(poolName),
			engine.NewStringValue(prompt),
			engine.NewObjectValue(optionsMap),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "generateWithPool", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Pool metrics
	L.SetField(pool, "getMetrics", L.NewFunction(func(L *lua.LState) int {
		poolName := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(poolName)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "getPoolMetrics", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add pool namespace to module
	L.SetField(module, "pool", pool)
}

// addModelMethods adds model management methods
func (la *LLMAdapter) addModelMethods(L *lua.LState, module *lua.LTable) {
	// Create models namespace
	models := L.NewTable()

	// listModels method
	L.SetField(models, "list", L.NewFunction(func(L *lua.LState) int {
		provider := L.OptString(1, "")

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(provider)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "listModels", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getModelInfo method
	L.SetField(models, "getInfo", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(modelName)}

		result, err := la.GetBridge().ExecuteMethod(ctx, "getModelInfo", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// checkCapabilities method
	L.SetField(models, "checkCapabilities", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		capability := L.CheckString(2)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(modelName),
			engine.NewStringValue(capability),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "checkModelCapability", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add models namespace to module
	L.SetField(module, "models", models)
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

	// Add token counting utility
	L.SetField(module, "countTokens", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		model := L.OptString(2, "")

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(text),
			engine.NewStringValue(model),
		}

		result, err := la.GetBridge().ExecuteMethod(ctx, "countTokens", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to Lua
		luaResult, err := la.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
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

	// Add pool strategies
	strategies := L.NewTable()
	L.SetField(strategies, "ROUND_ROBIN", lua.LString("round_robin"))
	L.SetField(strategies, "FAILOVER", lua.LString("failover"))
	L.SetField(strategies, "FASTEST", lua.LString("fastest"))
	L.SetField(strategies, "WEIGHTED", lua.LString("weighted"))
	L.SetField(strategies, "LEAST_USED", lua.LString("least_used"))
	L.SetField(module, "STRATEGIES", strategies)
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
	case "generate":
		return la.wrapGenerate(baseWrapped)
	case "generateMessage":
		return la.wrapGenerateMessage(baseWrapped)
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

// wrapGenerate adds generate-specific handling
func (la *LLMAdapter) wrapGenerate(baseFn lua.LGFunction) lua.LGFunction {
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

// wrapGenerateMessage adds message generation handling
func (la *LLMAdapter) wrapGenerateMessage(baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least messages array is provided
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("messages array is required"))
			return 2
		}

		// Validate messages is a table
		if L.Get(1).Type() != lua.LTTable {
			L.Push(lua.LNil)
			L.Push(lua.LString("messages must be an array"))
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

	// Add streaming method to agent
	L.SetField(agent, "stream", L.NewFunction(func(L *lua.LState) int {
		prompt := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		// Get agent ID
		agentId := agent.RawGetString("id")
		if agentId == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("agent has no id"))
			return 2
		}

		// Call agentStream through the module
		module := L.GetGlobal("llm")
		if module.Type() != lua.LTTable {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm module not found"))
			return 2
		}

		agentStreamFn := module.(*lua.LTable).RawGetString("agentStream")
		if agentStreamFn == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("agentStream not found"))
			return 2
		}

		// Call agentStream
		err := L.CallByParam(lua.P{
			Fn:      agentStreamFn,
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
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (la *LLMAdapter) tableToMap(table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			sv, err := la.BridgeAdapter.GetTypeConverter().ToLuaScriptValue(nil, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return result
}

// RegisterAsModule registers the adapter as a module in the module system
func (la *LLMAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge dependencies from metadata
	metadata := la.GetMetadata()
	deps := metadata.Dependencies

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  metadata.Description,
		Dependencies: deps,
		LoadFunc:     la.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (la *LLMAdapter) GetBridge() engine.Bridge {
	return la.BridgeAdapter.GetBridge()
}

// GetMethods returns the available methods
func (la *LLMAdapter) GetMethods() []string {
	// Get base methods
	methods := la.BridgeAdapter.GetMethods()

	// Add LLM-specific methods if not already present
	llmMethods := []string{
		"generate", "generateMessage", "stream",
		"createProvider", "getProvider", "listProviders",
		"createPool", "getPoolHealth", "generateWithPool",
		"listModels", "getModelInfo", "countTokens",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range llmMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
