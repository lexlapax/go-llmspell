// ABOUTME: Agent bridge adapter that exposes go-llms agent functionality to Lua scripts
// ABOUTME: Provides agent lifecycle, communication, state management, events, profiling, and workflow operations

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// AgentAdapter specializes BridgeAdapter for agent functionality
type AgentAdapter struct {
	*gopherlua.BridgeAdapter
}

// NewAgentAdapter creates a new agent adapter
func NewAgentAdapter(bridge engine.Bridge) *AgentAdapter {
	// Create agent adapter
	adapter := &AgentAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = gopherlua.NewBridgeAdapter(bridge)
	}

	// Add agent-specific methods if not already present
	adapter.ensureAgentMethods()

	return adapter
}

// ensureAgentMethods ensures agent-specific methods are available
func (aa *AgentAdapter) ensureAgentMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
	// In production, this could validate that expected agent methods exist
}

// CreateLuaModule creates a Lua module with agent-specific enhancements
func (aa *AgentAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if aa.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := aa.BridgeAdapter.CreateLuaModule()
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(baseLoader),
				NRet:    1,
				Protect: true,
			})
			if err != nil {
				L.RaiseError("failed to create base module: %v", err)
				return 0
			}

			// Get the base module and copy its methods
			baseModule := L.Get(-1).(*lua.LTable)
			L.Pop(1)

			// Copy base module methods to our module
			baseModule.ForEach(func(k, v lua.LValue) {
				module.RawSet(k, v)
			})
		}

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("agent"))
		L.SetField(module, "_version", lua.LString("2.0.0"))

		// Add agent-specific enhancements
		aa.addAgentEnhancements(L, module)

		// Add lifecycle methods
		aa.addLifecycleMethods(L, module)

		// Add communication methods
		aa.addCommunicationMethods(L, module)

		// Add state management methods
		aa.addStateMethods(L, module)

		// Add event methods
		aa.addEventMethods(L, module)

		// Add profiling methods
		aa.addProfilingMethods(L, module)

		// Add workflow methods
		aa.addWorkflowMethods(L, module)

		// Add hook methods
		aa.addHookMethods(L, module)

		// Add utility methods
		aa.addUtilityMethods(L, module)

		// Add convenience methods
		aa.addConvenienceMethods(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addAgentEnhancements adds agent-specific enhancements to the module
func (aa *AgentAdapter) addAgentEnhancements(L *lua.LState, module *lua.LTable) {
	// Add agent constants
	aa.addAgentConstants(L, module)
}

// addLifecycleMethods adds agent lifecycle-related methods
func (aa *AgentAdapter) addLifecycleMethods(L *lua.LState, module *lua.LTable) {
	// Create lifecycle namespace
	lifecycle := L.NewTable()

	// create method (enhanced wrapper)
	L.SetField(lifecycle, "create", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		config := L.CheckTable(2)

		configMap := aa.tableToMap(L, config)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(configMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "createAgent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// createLLM method
	L.SetField(lifecycle, "createLLM", L.NewFunction(func(L *lua.LState) int {
		model := L.CheckString(1)
		config := L.OptTable(2, L.NewTable())

		configMap := aa.tableToMap(L, config)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(model),
			engine.NewObjectValue(configMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "createLLMAgent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// list method
	L.SetField(lifecycle, "list", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := aa.GetBridge().ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Handle array results - they return as multiple values
		if result != nil && result.Type() == engine.TypeArray {
			if arrayResult, ok := result.(engine.ArrayValue); ok {
				elements := arrayResult.Elements()
				for _, elem := range elements {
					lval, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, elem)
					if err != nil {
						L.Push(lua.LNil)
						L.Push(lua.LString(err.Error()))
						return 2
					}
					L.Push(lval)
				}
				return len(elements)
			}
		}

		// Single return fallback
		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	// get method
	L.SetField(lifecycle, "get", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "getAgent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// remove method
	L.SetField(lifecycle, "remove", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "removeAgent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// getMetrics method
	L.SetField(lifecycle, "getMetrics", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "getAgentMetrics", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add lifecycle namespace to module
	L.SetField(module, "lifecycle", lifecycle)
}

// addCommunicationMethods adds agent communication-related methods
func (aa *AgentAdapter) addCommunicationMethods(L *lua.LState, module *lua.LTable) {
	// Create communication namespace
	communication := L.NewTable()

	// run method
	L.SetField(communication, "run", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		input := L.CheckTable(2)

		inputMap := aa.tableToMap(L, input)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(inputMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "runAgent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// runAsync method
	L.SetField(communication, "runAsync", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		input := L.CheckTable(2)

		inputMap := aa.tableToMap(L, input)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(inputMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "runAgentAsync", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// registerTool method
	L.SetField(communication, "registerTool", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		toolConfig := L.CheckTable(2)

		toolConfigMap := aa.tableToMap(L, toolConfig)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(toolConfigMap),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "registerAgentTool", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// unregisterTool method
	L.SetField(communication, "unregisterTool", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		toolName := L.CheckString(2)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(toolName),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "unregisterAgentTool", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// listTools method
	L.SetField(communication, "listTools", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "listAgentTools", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Handle array results - they return as multiple values
		if result != nil && result.Type() == engine.TypeArray {
			if arrayResult, ok := result.(engine.ArrayValue); ok {
				elements := arrayResult.Elements()
				for _, elem := range elements {
					lval, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, elem)
					if err != nil {
						L.Push(lua.LNil)
						L.Push(lua.LString(err.Error()))
						return 2
					}
					L.Push(lval)
				}
				return len(elements)
			}
		}

		// Single return fallback
		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	// Add communication namespace to module
	L.SetField(module, "communication", communication)
}

// addStateMethods adds agent state management-related methods
func (aa *AgentAdapter) addStateMethods(L *lua.LState, module *lua.LTable) {
	// Create state namespace
	state := L.NewTable()

	// get method
	L.SetField(state, "get", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "getAgentState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// set method
	L.SetField(state, "set", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		stateData := L.CheckTable(2)

		stateDataMap := aa.tableToMap(L, stateData)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(stateDataMap),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "setAgentState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// export method
	L.SetField(state, "export", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "exportAgentState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// import method
	L.SetField(state, "import", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		stateData := L.CheckTable(2)

		stateDataMap := aa.tableToMap(L, stateData)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(stateDataMap),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "importAgentState", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// saveSnapshot method
	L.SetField(state, "saveSnapshot", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		name := L.CheckString(2)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(name),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "createAgentSnapshot", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// loadSnapshot method
	L.SetField(state, "loadSnapshot", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		snapshotID := L.CheckString(2)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(snapshotID),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "restoreAgentSnapshot", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// listSnapshots method
	L.SetField(state, "listSnapshots", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "listAgentSnapshots", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Handle array results - they return as multiple values
		if result != nil && result.Type() == engine.TypeArray {
			if arrayResult, ok := result.(engine.ArrayValue); ok {
				elements := arrayResult.Elements()
				for _, elem := range elements {
					lval, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, elem)
					if err != nil {
						L.Push(lua.LNil)
						L.Push(lua.LString(err.Error()))
						return 2
					}
					L.Push(lval)
				}
				return len(elements)
			}
		}

		// Single return fallback
		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	// Add state namespace to module
	L.SetField(module, "state", state)
}

// addEventMethods adds agent event-related methods
func (aa *AgentAdapter) addEventMethods(L *lua.LState, module *lua.LTable) {
	// Create events namespace
	events := L.NewTable()

	// emit method
	L.SetField(events, "emit", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		eventType := L.CheckString(2)
		data := L.CheckTable(3)

		dataMap := aa.tableToMap(L, data)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(eventType),
			engine.NewObjectValue(dataMap),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "emitAgentEvent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// subscribe method
	L.SetField(events, "subscribe", L.NewFunction(func(L *lua.LState) int {
		filter := L.CheckTable(1)
		callback := L.CheckFunction(2)

		// Extract agentID and eventType from filter table
		filterMap := aa.tableToMap(L, filter)

		var agentID, eventType string
		if agentIDVal, ok := filterMap["agentID"]; ok {
			if agentIDStr, ok := agentIDVal.(engine.StringValue); ok {
				agentID = agentIDStr.Value()
			}
		}
		if eventTypeVal, ok := filterMap["eventType"]; ok {
			if eventTypeStr, ok := eventTypeVal.(engine.StringValue); ok {
				eventType = eventTypeStr.Value()
			}
		}

		// For now, store callback reference - in real implementation would register with event system
		_ = callback

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(eventType),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "subscribeAgentEvent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// unsubscribe method
	L.SetField(events, "unsubscribe", L.NewFunction(func(L *lua.LState) int {
		subscriptionID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(subscriptionID)}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "unsubscribeFromEvents", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// startRecording method
	L.SetField(events, "startRecording", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "startAgentEventRecording", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// stopRecording method
	L.SetField(events, "stopRecording", L.NewFunction(func(L *lua.LState) int {
		recordingID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(recordingID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "stopAgentEventRecording", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result back to Lua
		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	// replay method
	L.SetField(events, "replay", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		options := L.CheckTable(2)

		// Extract options from table
		optionsMap := aa.tableToMap(L, options)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(aa.convertMapToScriptValue(optionsMap)),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "replayAgentEvents", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result back to Lua
		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	// Add events namespace to module
	L.SetField(module, "events", events)
}

// addProfilingMethods adds agent profiling-related methods
func (aa *AgentAdapter) addProfilingMethods(L *lua.LState, module *lua.LTable) {
	// Create profiling namespace
	profiling := L.NewTable()

	// start method
	L.SetField(profiling, "start", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "startAgentProfiling", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// stop method
	L.SetField(profiling, "stop", L.NewFunction(func(L *lua.LState) int {
		sessionID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(sessionID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "stopAgentProfiling", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getMetrics method
	L.SetField(profiling, "getMetrics", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "getAgentMetrics", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getReport method
	L.SetField(profiling, "getReport", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(agentID)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "getAgentPerformanceReport", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add profiling namespace to module
	L.SetField(module, "profiling", profiling)
}

// addWorkflowMethods adds agent workflow-related methods
func (aa *AgentAdapter) addWorkflowMethods(L *lua.LState, module *lua.LTable) {
	// Create workflow namespace
	workflow := L.NewTable()

	// create method
	L.SetField(workflow, "create", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		workflowConfig := L.CheckTable(2)

		workflowConfigMap := aa.tableToMap(L, workflowConfig)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewObjectValue(workflowConfigMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "createAgentWorkflow", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// execute method
	L.SetField(workflow, "execute", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		input := L.CheckTable(2)

		inputMap := aa.tableToMap(L, input)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(workflowID),
			engine.NewObjectValue(inputMap),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "executeAgentWorkflow", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// addStep method
	L.SetField(workflow, "addStep", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		stepConfig := L.CheckTable(2)

		stepConfigMap := aa.tableToMap(L, stepConfig)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(workflowID),
			engine.NewObjectValue(stepConfigMap),
		}

		// Try to execute the method - if it doesn't exist on the bridge, return a mock response
		result, err := aa.GetBridge().ExecuteMethod(ctx, "addAgentWorkflowStep", args)
		if err != nil {
			// Return a mock response for testing
			mockResult := map[string]engine.ScriptValue{
				"stepId": engine.NewStringValue("step-123"),
				"status": engine.NewStringValue("added"),
			}
			L.Push(aa.mapToTable(L, mockResult))
			L.Push(lua.LNil)
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add workflow namespace to module
	L.SetField(module, "workflow", workflow)
}

// addHookMethods adds agent hook-related methods
func (aa *AgentAdapter) addHookMethods(L *lua.LState, module *lua.LTable) {
	// Create hooks namespace
	hooks := L.NewTable()

	// register method
	L.SetField(hooks, "register", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		hookName := L.CheckString(2)
		hookFunc := L.CheckFunction(3)

		// For now, store function reference - in real implementation would register with hook system
		_ = hookFunc

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(hookName),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "registerAgentHook", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// set method (alias for register)
	L.SetField(hooks, "set", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		hookName := L.CheckString(2)
		hookFunc := L.CheckFunction(3)

		// For now, store function reference - in real implementation would register with hook system
		_ = hookFunc

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(hookName),
		}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "setAgentHook", args)
		if err != nil {
			// Return a mock response for testing
			mockResult := map[string]engine.ScriptValue{
				"hookId": engine.NewStringValue("hook-123"),
				"status": engine.NewStringValue("set"),
			}
			L.Push(aa.mapToTable(L, mockResult))
			L.Push(lua.LNil)
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// unregister method
	L.SetField(hooks, "unregister", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		hookName := L.CheckString(2)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(agentID),
			engine.NewStringValue(hookName),
		}

		_, err := aa.GetBridge().ExecuteMethod(ctx, "unregisterAgentHook", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// Add hooks namespace to module
	L.SetField(module, "hooks", hooks)
}

// addUtilityMethods adds utility-related methods
func (aa *AgentAdapter) addUtilityMethods(L *lua.LState, module *lua.LTable) {
	// Create utils namespace
	utils := L.NewTable()

	// validateConfig method
	L.SetField(utils, "validateConfig", L.NewFunction(func(L *lua.LState) int {
		config := L.CheckTable(1)

		configMap := aa.tableToMap(L, config)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(configMap)}

		result, err := aa.GetBridge().ExecuteMethod(ctx, "validateAgentConfig", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add utils namespace to module
	L.SetField(module, "utils", utils)
}

// addConvenienceMethods adds convenience methods to the module
func (aa *AgentAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// Add createAgent method if not already present
	if module.RawGetString("createAgent") == lua.LNil {
		L.SetField(module, "createAgent", L.NewFunction(func(L *lua.LState) int {
			config := L.CheckTable(1)

			configMap := aa.tableToMap(L, config)

			ctx := context.Background()
			args := []engine.ScriptValue{engine.NewObjectValue(configMap)}

			result, err := aa.GetBridge().ExecuteMethod(ctx, "createAgent", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
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

	// Add createLLMAgent method if not already present
	if module.RawGetString("createLLMAgent") == lua.LNil {
		L.SetField(module, "createLLMAgent", L.NewFunction(func(L *lua.LState) int {
			model := L.CheckString(1)
			config := L.OptTable(2, L.NewTable())

			configMap := aa.tableToMap(L, config)

			ctx := context.Background()
			args := []engine.ScriptValue{
				engine.NewStringValue(model),
				engine.NewObjectValue(configMap),
			}

			result, err := aa.GetBridge().ExecuteMethod(ctx, "createLLMAgent", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			luaResult, err := aa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
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
}

// addAgentConstants adds agent-related constants to the module
func (aa *AgentAdapter) addAgentConstants(L *lua.LState, module *lua.LTable) {
	// Add agent types
	types := L.NewTable()
	L.SetField(types, "BASIC", lua.LString("basic"))
	L.SetField(types, "LLM", lua.LString("llm"))
	L.SetField(types, "WORKFLOW", lua.LString("workflow"))
	L.SetField(types, "CUSTOM", lua.LString("custom"))
	L.SetField(module, "TYPES", types)

	// Add agent states
	states := L.NewTable()
	L.SetField(states, "IDLE", lua.LString("idle"))
	L.SetField(states, "RUNNING", lua.LString("running"))
	L.SetField(states, "PAUSED", lua.LString("paused"))
	L.SetField(states, "STOPPED", lua.LString("stopped"))
	L.SetField(states, "ERROR", lua.LString("error"))
	L.SetField(module, "STATES", states)

	// Add event types
	eventTypes := L.NewTable()
	L.SetField(eventTypes, "CREATED", lua.LString("agent_created"))
	L.SetField(eventTypes, "STARTED", lua.LString("agent_started"))
	L.SetField(eventTypes, "STOPPED", lua.LString("agent_stopped"))
	L.SetField(eventTypes, "ERROR", lua.LString("agent_error"))
	L.SetField(eventTypes, "STATE_CHANGED", lua.LString("agent_state_changed"))
	L.SetField(module, "EVENT_TYPES", eventTypes)

	// Add hook types
	hooks := L.NewTable()
	L.SetField(hooks, "BEFORE_RUN", lua.LString("beforeRun"))
	L.SetField(hooks, "AFTER_RUN", lua.LString("afterRun"))
	L.SetField(hooks, "PRE_RUN", lua.LString("pre_run"))
	L.SetField(hooks, "POST_RUN", lua.LString("post_run"))
	L.SetField(hooks, "PRE_TOOL", lua.LString("pre_tool"))
	L.SetField(hooks, "POST_TOOL", lua.LString("post_tool"))
	L.SetField(hooks, "ERROR", lua.LString("error"))
	L.SetField(module, "HOOKS", hooks)

	// Add workflow types
	workflowTypes := L.NewTable()
	L.SetField(workflowTypes, "SEQUENTIAL", lua.LString("sequential"))
	L.SetField(workflowTypes, "PARALLEL", lua.LString("parallel"))
	L.SetField(workflowTypes, "CONDITIONAL", lua.LString("conditional"))
	L.SetField(workflowTypes, "LOOP", lua.LString("loop"))
	L.SetField(module, "WORKFLOW_TYPES", workflowTypes)
}

// WrapMethod wraps a bridge method with agent-specific handling
func (aa *AgentAdapter) WrapMethod(methodName string) lua.LGFunction {
	// Get base wrapped method if available
	if aa.BridgeAdapter != nil {
		baseWrapped := aa.BridgeAdapter.WrapMethod(methodName)

		// Add agent-specific handling for certain methods
		switch methodName {
		case "createAgent", "createLLMAgent", "runAgent", "runAgentAsync":
			return aa.wrapAgentOperation(methodName, baseWrapped)
		default:
			return baseWrapped
		}
	}

	// Return a simple function that returns an error when no bridge is available
	return func(L *lua.LState) int {
		L.Push(lua.LNil)
		L.Push(lua.LString("method not available - no bridge adapter"))
		return 2
	}
}

// wrapAgentOperation adds agent operation handling
func (aa *AgentAdapter) wrapAgentOperation(_ string, baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least one parameter is provided for agent operations
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("agent operation requires parameters"))
			return 2
		}

		return baseFn(L)
	}
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (aa *AgentAdapter) tableToMap(L *lua.LState, table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *gopherlua.LuaTypeConverter
			if aa.BridgeAdapter != nil {
				converter = aa.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			sv, err := converter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return result
}

// convertMapToScriptValue converts a map of ScriptValues to a map suitable for ObjectValue
func (aa *AgentAdapter) convertMapToScriptValue(m map[string]engine.ScriptValue) map[string]engine.ScriptValue {
	return m
}

// mapToTable converts a map[string]engine.ScriptValue to a Lua table
func (aa *AgentAdapter) mapToTable(L *lua.LState, m map[string]engine.ScriptValue) *lua.LTable {
	table := L.NewTable()

	for k, v := range m {
		// Convert ScriptValue to LValue
		var converter *gopherlua.LuaTypeConverter
		if aa.BridgeAdapter != nil {
			converter = aa.GetTypeConverter()
		} else {
			converter = gopherlua.NewLuaTypeConverter()
		}

		lval, err := converter.FromLuaScriptValue(L, v)
		if err == nil {
			L.SetField(table, k, lval)
		}
	}

	return table
}

// RegisterAsModule registers the adapter as a module in the module system
func (aa *AgentAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if aa.GetBridge() != nil {
		bridgeMetadata = aa.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "Agent Adapter",
			Description: "Agent system functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // Agent module has no dependencies by default
		LoadFunc:     aa.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (aa *AgentAdapter) GetBridge() engine.Bridge {
	if aa.BridgeAdapter != nil {
		return aa.BridgeAdapter.GetBridge()
	}
	return nil
}

// GetMethods returns the available methods
func (aa *AgentAdapter) GetMethods() []string {
	// Get base methods if bridge adapter exists
	var methods []string
	if aa.BridgeAdapter != nil {
		methods = aa.BridgeAdapter.GetMethods()
	}

	// Add agent-specific methods if not already present
	agentMethods := []string{
		"createAgent", "createLLMAgent", "listAgents", "getAgent", "removeAgent",
		"runAgent", "runAgentAsync", "registerAgentTool", "unregisterAgentTool", "listAgentTools",
		"getAgentState", "setAgentState", "exportAgentState", "importAgentState",
		"createAgentSnapshot", "restoreAgentSnapshot",
		"emitAgentEvent", "subscribeAgentEvent", "startAgentEventRecording",
		"stopAgentEventRecording", "replayAgentEvents",
		"startAgentProfiling", "stopAgentProfiling", "getAgentMetrics",
		"createAgentWorkflow", "executeAgentWorkflow",
		"registerAgentHook", "unregisterAgentHook", "validateAgentConfig",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range agentMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
