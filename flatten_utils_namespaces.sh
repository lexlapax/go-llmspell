#!/bin/bash

# Script to help flatten the remaining namespaces in utils.go

cat << 'EOF' > /tmp/llm_methods.go
// addLLMMethods adds LLM utility methods (flattened to module level)
func (ua *UtilsAdapter) addLLMMethods(L *lua.LState, module *lua.LTable) {
	// llmCreateProvider method (flattened from llm.createProvider)
	L.SetField(module, "llmCreateProvider", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		providerType := L.CheckString(1)
		config := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(providerType),
			ua.tableToScriptValue(L, config),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "createProvider", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// llmGenerateTyped method (flattened from llm.generateTyped)
	L.SetField(module, "llmGenerateTyped", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		prompt := L.CheckString(1)
		schema := L.CheckTable(2)
		options := L.CheckTable(3)

		args := []engine.ScriptValue{
			engine.NewStringValue(prompt),
			ua.tableToScriptValue(L, schema),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "generateTyped", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// llmTrackCost method (flattened from llm.trackCost)
	L.SetField(module, "llmTrackCost", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		operation := L.CheckString(1)
		tokens := L.CheckNumber(2)
		model := L.CheckString(3)

		args := []engine.ScriptValue{
			engine.NewStringValue(operation),
			engine.NewNumberValue(float64(tokens)),
			engine.NewStringValue(model),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "trackCost", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add methods from TODO that may be missing
	// llmParseResponse method (flattened from llm.parseResponse)
	L.SetField(module, "llmParseResponse", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		response := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(response),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "parseResponse", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// llmFormatPrompt method (flattened from llm.formatPrompt)
	L.SetField(module, "llmFormatPrompt", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		template := L.CheckString(1)
		data := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(template),
			ua.tableToScriptValue(L, data),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "formatPrompt", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// llmCountTokens method (flattened from llm.countTokens)
	L.SetField(module, "llmCountTokens", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)
		model := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
			engine.NewStringValue(model),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "countTokens", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// llmSplitMessage method (flattened from llm.splitMessage)
	L.SetField(module, "llmSplitMessage", L.NewFunction(func(L *lua.LState) int {
		if ua.llmBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("llm bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		maxTokens := L.CheckNumber(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			engine.NewNumberValue(float64(maxTokens)),
		}

		result, err := ua.llmBridge.ExecuteMethod(context.Background(), "splitMessage", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ua.typeConverter.FromLuaScriptValue(L, result)
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
EOF

echo "LLM methods flattened template created at /tmp/llm_methods.go"