// ABOUTME: Utility bridge adapter that exposes go-llms utility functionality to Lua scripts
// ABOUTME: Provides auth, debug, errors, json, llm utils, logging, slog, and general utility functionality

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// UtilsAdapter combines multiple utility bridges into a unified adapter
type UtilsAdapter struct {
	authBridge    engine.Bridge
	debugBridge   engine.Bridge
	errorsBridge  engine.Bridge
	jsonBridge    engine.Bridge
	llmBridge     engine.Bridge
	loggerBridge  engine.Bridge
	slogBridge    engine.Bridge
	utilBridge    engine.Bridge
	typeConverter *gopherlua.LuaTypeConverter
}

// NewUtilsAdapter creates a new utility adapter
func NewUtilsAdapter(authBridge, debugBridge, errorsBridge, jsonBridge, llmBridge, loggerBridge, slogBridge, utilBridge engine.Bridge) *UtilsAdapter {
	return &UtilsAdapter{
		authBridge:    authBridge,
		debugBridge:   debugBridge,
		errorsBridge:  errorsBridge,
		jsonBridge:    jsonBridge,
		llmBridge:     llmBridge,
		loggerBridge:  loggerBridge,
		slogBridge:    slogBridge,
		utilBridge:    utilBridge,
		typeConverter: gopherlua.NewLuaTypeConverter(),
	}
}

// GetAdapterName returns the adapter name
func (ua *UtilsAdapter) GetAdapterName() string {
	return "utils"
}

// CreateLuaModule creates a Lua module with utility enhancements
func (ua *UtilsAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("utils"))
		L.SetField(module, "_version", lua.LString("1.0.0"))

		// Add auth namespace
		ua.addAuthMethods(L, module)

		// Add debug namespace
		ua.addDebugMethods(L, module)

		// Add errors namespace
		ua.addErrorMethods(L, module)

		// Add json namespace
		ua.addJSONMethods(L, module)

		// Add llm namespace
		ua.addLLMMethods(L, module)

		// Add logger namespace
		ua.addLoggerMethods(L, module)

		// Add slog namespace
		ua.addSlogMethods(L, module)

		// Add general utilities namespace
		ua.addGeneralMethods(L, module)

		// Add utility constants
		ua.addUtilityConstants(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addAuthMethods adds authentication methods
func (ua *UtilsAdapter) addAuthMethods(L *lua.LState, module *lua.LTable) {
	// Create auth namespace
	auth := L.NewTable()

	// authenticate method
	L.SetField(auth, "authenticate", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		credentials := L.CheckTable(1)
		scheme := L.CheckString(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, credentials),
			engine.NewStringValue(scheme),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "authenticate", args)
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

	// validateToken method
	L.SetField(auth, "validateToken", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		token := L.CheckString(1)
		options := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(token),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "validateToken", args)
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

	// refreshToken method
	L.SetField(auth, "refreshToken", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		refreshToken := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(refreshToken),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "refreshToken", args)
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

	// Add auth namespace to module
	L.SetField(module, "auth", auth)
}

// addDebugMethods adds debug methods
func (ua *UtilsAdapter) addDebugMethods(L *lua.LState, module *lua.LTable) {
	// Create debug namespace
	debug := L.NewTable()

	// setLevel method
	L.SetField(debug, "setLevel", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		component := L.CheckString(1)
		level := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(component),
			engine.NewStringValue(level),
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "setDebugLevel", args)
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

	// log method
	L.SetField(debug, "log", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		component := L.CheckString(1)
		message := L.CheckString(2)
		data := L.CheckTable(3)

		args := []engine.ScriptValue{
			engine.NewStringValue(component),
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, data),
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "debugLog", args)
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

	// getConfig method
	L.SetField(debug, "getConfig", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "getDebugConfig", []engine.ScriptValue{})
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

	// Add debug namespace to module
	L.SetField(module, "debug", debug)
}

// addErrorMethods adds error handling methods
func (ua *UtilsAdapter) addErrorMethods(L *lua.LState, module *lua.LTable) {
	// Create errors namespace
	errors := L.NewTable()

	// createError method
	L.SetField(errors, "createError", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		code := L.CheckString(2)
		category := L.CheckString(3)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			engine.NewStringValue(code),
			engine.NewStringValue(category),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "createError", args)
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

	// wrapError method
	L.SetField(errors, "wrapError", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		originalError := L.CheckTable(1)
		contextData := L.CheckTable(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, originalError),
			ua.tableToScriptValue(L, contextData),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "wrapError", args)
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

	// aggregateErrors method
	L.SetField(errors, "aggregateErrors", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		errorsTable := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, errorsTable),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "aggregateErrors", args)
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

	// categorizeError method
	L.SetField(errors, "categorizeError", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		errorData := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, errorData),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "categorizeError", args)
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

	// Add errors namespace to module
	L.SetField(module, "errors", errors)
}

// addJSONMethods adds JSON processing methods
func (ua *UtilsAdapter) addJSONMethods(L *lua.LState, module *lua.LTable) {
	// Create json namespace
	json := L.NewTable()

	// parse method
	L.SetField(json, "parse", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)
		options := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "parseJSON", args)
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

	// toJSON method
	L.SetField(json, "toJSON", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		data := L.CheckTable(1)
		options := L.CheckTable(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, data),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "toJSON", args)
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

	// validateJSONSchema method
	L.SetField(json, "validateJSONSchema", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		data := L.CheckTable(1)
		schema := L.CheckTable(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, data),
			ua.tableToScriptValue(L, schema),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "validateJSONSchema", args)
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

	// extractStructuredData method
	L.SetField(json, "extractStructuredData", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)
		schema := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
			ua.tableToScriptValue(L, schema),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "extractStructuredData", args)
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

	// Add json namespace to module
	L.SetField(module, "json", json)
}

// addLLMMethods adds LLM utility methods
func (ua *UtilsAdapter) addLLMMethods(L *lua.LState, module *lua.LTable) {
	// Create llm namespace
	llm := L.NewTable()

	// createProvider method
	L.SetField(llm, "createProvider", L.NewFunction(func(L *lua.LState) int {
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

	// generateTyped method
	L.SetField(llm, "generateTyped", L.NewFunction(func(L *lua.LState) int {
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

	// trackCost method
	L.SetField(llm, "trackCost", L.NewFunction(func(L *lua.LState) int {
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

	// Add llm namespace to module
	L.SetField(module, "llm", llm)
}

// addLoggerMethods adds logger methods
func (ua *UtilsAdapter) addLoggerMethods(L *lua.LState, module *lua.LTable) {
	// Create logger namespace
	logger := L.NewTable()

	// createLogger method
	L.SetField(logger, "createLogger", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		component := L.CheckString(1)
		config := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(component),
			ua.tableToScriptValue(L, config),
		}

		result, err := ua.loggerBridge.ExecuteMethod(context.Background(), "createLogger", args)
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

	// log method
	L.SetField(logger, "log", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		level := L.CheckString(1)
		message := L.CheckString(2)
		contextTable := L.CheckTable(3)

		args := []engine.ScriptValue{
			engine.NewStringValue(level),
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, contextTable),
		}

		result, err := ua.loggerBridge.ExecuteMethod(context.Background(), "log", args)
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

	// setLogLevel method
	L.SetField(logger, "setLogLevel", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		component := L.CheckString(1)
		level := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(component),
			engine.NewStringValue(level),
		}

		result, err := ua.loggerBridge.ExecuteMethod(context.Background(), "setLogLevel", args)
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

	// Add logger namespace to module
	L.SetField(module, "logger", logger)
}

// addSlogMethods adds structured logging methods
func (ua *UtilsAdapter) addSlogMethods(L *lua.LState, module *lua.LTable) {
	// Create slog namespace
	slog := L.NewTable()

	// info method
	L.SetField(slog, "info", L.NewFunction(func(L *lua.LState) int {
		if ua.slogBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("slog bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		fields := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, fields),
		}

		result, err := ua.slogBridge.ExecuteMethod(context.Background(), "info", args)
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

	// warn method
	L.SetField(slog, "warn", L.NewFunction(func(L *lua.LState) int {
		if ua.slogBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("slog bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		fields := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, fields),
		}

		result, err := ua.slogBridge.ExecuteMethod(context.Background(), "warn", args)
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

	// error method
	L.SetField(slog, "error", L.NewFunction(func(L *lua.LState) int {
		if ua.slogBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("slog bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		fields := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, fields),
		}

		result, err := ua.slogBridge.ExecuteMethod(context.Background(), "error", args)
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

	// debug method
	L.SetField(slog, "debug", L.NewFunction(func(L *lua.LState) int {
		if ua.slogBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("slog bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		fields := L.CheckTable(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
			ua.tableToScriptValue(L, fields),
		}

		result, err := ua.slogBridge.ExecuteMethod(context.Background(), "debug", args)
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

	// Add slog namespace to module
	L.SetField(module, "slog", slog)
}

// addGeneralMethods adds general utility methods
func (ua *UtilsAdapter) addGeneralMethods(L *lua.LState, module *lua.LTable) {
	// Create general namespace
	general := L.NewTable()

	// generateUUID method
	L.SetField(general, "generateUUID", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "generateUUID", []engine.ScriptValue{})
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

	// hash method
	L.SetField(general, "hash", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		data := L.CheckString(1)
		algorithm := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(data),
			engine.NewStringValue(algorithm),
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "hash", args)
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

	// retry method
	L.SetField(general, "retry", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		operation := L.CheckTable(1)
		options := L.CheckTable(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, operation),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "retry", args)
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

	// sleep method
	L.SetField(general, "sleep", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		duration := L.CheckNumber(1)

		args := []engine.ScriptValue{
			engine.NewNumberValue(float64(duration)),
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "sleep", args)
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

	// Add general namespace to module
	L.SetField(module, "general", general)
}

// addUtilityConstants adds utility-related constants
func (ua *UtilsAdapter) addUtilityConstants(L *lua.LState, module *lua.LTable) {
	// Add log levels
	logLevels := L.NewTable()
	L.SetField(logLevels, "DEBUG", lua.LString("DEBUG"))
	L.SetField(logLevels, "INFO", lua.LString("INFO"))
	L.SetField(logLevels, "WARN", lua.LString("WARN"))
	L.SetField(logLevels, "ERROR", lua.LString("ERROR"))
	L.SetField(module, "LOG_LEVELS", logLevels)

	// Add auth schemes
	authSchemes := L.NewTable()
	L.SetField(authSchemes, "OAUTH2", lua.LString("oauth2"))
	L.SetField(authSchemes, "BASIC", lua.LString("basic"))
	L.SetField(authSchemes, "BEARER", lua.LString("bearer"))
	L.SetField(authSchemes, "API_KEY", lua.LString("api_key"))
	L.SetField(module, "AUTH_SCHEMES", authSchemes)

	// Add hash algorithms
	hashAlgorithms := L.NewTable()
	L.SetField(hashAlgorithms, "SHA256", lua.LString("sha256"))
	L.SetField(hashAlgorithms, "SHA1", lua.LString("sha1"))
	L.SetField(hashAlgorithms, "MD5", lua.LString("md5"))
	L.SetField(module, "HASH_ALGORITHMS", hashAlgorithms)

	// Add error categories
	errorCategories := L.NewTable()
	L.SetField(errorCategories, "VALIDATION", lua.LString("validation"))
	L.SetField(errorCategories, "NETWORK", lua.LString("network"))
	L.SetField(errorCategories, "AUTH", lua.LString("auth"))
	L.SetField(errorCategories, "SYSTEM", lua.LString("system"))
	L.SetField(module, "ERROR_CATEGORIES", errorCategories)
}

// tableToScriptValue converts a Lua table to a ScriptValue
func (ua *UtilsAdapter) tableToScriptValue(L *lua.LState, table *lua.LTable) engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			sv, err := ua.typeConverter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return engine.NewObjectValue(result)
}

// RegisterAsModule registers the adapter as a module in the module system
func (ua *UtilsAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Create module definition using our CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  "Comprehensive utility functions for auth, debug, errors, json, logging and general purposes",
		Dependencies: []string{}, // Utils module has no dependencies by default
		LoadFunc:     ua.CreateLuaModule(),
	}

	// Register the module
	return ms.Register(module)
}

// GetMethods returns the available methods
func (ua *UtilsAdapter) GetMethods() []string {
	methods := []string{
		// Auth methods
		"authenticate", "validateToken", "refreshToken",
		// Debug methods
		"setDebugLevel", "debugLog", "getDebugConfig",
		// Error methods
		"createError", "wrapError", "aggregateErrors", "categorizeError",
		// JSON methods
		"parseJSON", "toJSON", "validateJSONSchema", "extractStructuredData",
		// LLM methods
		"createProvider", "generateTyped", "trackCost",
		// Logger methods
		"createLogger", "log", "setLogLevel",
		// Slog methods
		"info", "warn", "error", "debug",
		// General methods
		"generateUUID", "hash", "retry", "sleep",
	}

	return methods
}
