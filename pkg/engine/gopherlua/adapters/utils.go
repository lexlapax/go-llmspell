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

		// Add all flattened methods
		ua.addAuthMethods(L, module)
		ua.addDebugMethods(L, module)
		ua.addErrorMethods(L, module)
		ua.addJSONMethods(L, module)
		ua.addLLMMethods(L, module)
		ua.addLoggerMethods(L, module)
		ua.addSlogMethods(L, module)
		ua.addGeneralMethods(L, module)

		// Add utility constants
		ua.addUtilityConstants(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addAuthMethods adds authentication methods (flattened to module level)
func (ua *UtilsAdapter) addAuthMethods(L *lua.LState, module *lua.LTable) {
	// authAuthenticate method (flattened from auth.authenticate)
	L.SetField(module, "authAuthenticate", L.NewFunction(func(L *lua.LState) int {
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

	// authValidateToken method (flattened from auth.validateToken)
	L.SetField(module, "authValidateToken", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		token := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

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

	// authRefreshToken method (flattened from auth.refreshToken)
	L.SetField(module, "authRefreshToken", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// authGenerateToken method (flattened from auth.generateToken)
	L.SetField(module, "authGenerateToken", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		userData := L.CheckTable(1)
		options := L.OptTable(2, L.NewTable())

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, userData),
			ua.tableToScriptValue(L, options),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "generateToken", args)
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

	// authHashPassword method (flattened from auth.hashPassword)
	L.SetField(module, "authHashPassword", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		password := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(password),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "hashPassword", args)
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

	// authVerifyPassword method (flattened from auth.verifyPassword)
	L.SetField(module, "authVerifyPassword", L.NewFunction(func(L *lua.LState) int {
		if ua.authBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("auth bridge not initialized"))
			return 2
		}

		password := L.CheckString(1)
		hash := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(password),
			engine.NewStringValue(hash),
		}

		result, err := ua.authBridge.ExecuteMethod(context.Background(), "verifyPassword", args)
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

// addDebugMethods adds debug methods (flattened to module level)
func (ua *UtilsAdapter) addDebugMethods(L *lua.LState, module *lua.LTable) {
	// debugSetLevel method (flattened from debug.setLevel)
	L.SetField(module, "debugSetLevel", L.NewFunction(func(L *lua.LState) int {
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

	// debugLog method (flattened from debug.log)
	L.SetField(module, "debugLog", L.NewFunction(func(L *lua.LState) int {
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

	// debugGetConfig method (flattened from debug.getConfig)
	L.SetField(module, "debugGetConfig", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// debugTrace method (flattened from debug.trace)
	L.SetField(module, "debugTrace", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(message),
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "trace", args)
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

	// debugProfile method (flattened from debug.profile)
	L.SetField(module, "debugProfile", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		name := L.CheckString(1)
		action := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			engine.NewStringValue(action),
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "profile", args)
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

	// debugDump method (flattened from debug.dump)
	L.SetField(module, "debugDump", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		value := L.Get(1)

		// Convert to script value
		sv, err := ua.typeConverter.ToLuaScriptValue(L, value)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		args := []engine.ScriptValue{sv}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "dump", args)
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

	// debugAssert method (flattened from debug.assert)
	L.SetField(module, "debugAssert", L.NewFunction(func(L *lua.LState) int {
		if ua.debugBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("debug bridge not initialized"))
			return 2
		}

		condition := L.CheckBool(1)
		message := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewBoolValue(condition),
			engine.NewStringValue(message),
		}

		result, err := ua.debugBridge.ExecuteMethod(context.Background(), "assert", args)
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

// addErrorMethods adds error handling methods (flattened to module level)
func (ua *UtilsAdapter) addErrorMethods(L *lua.LState, module *lua.LTable) {
	// errorsCreateError method (flattened from errors.createError)
	L.SetField(module, "errorsCreateError", L.NewFunction(func(L *lua.LState) int {
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

	// errorsWrapError method (renamed from wrapError)
	L.SetField(module, "errorsWrapError", L.NewFunction(func(L *lua.LState) int {
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

	// errorsAggregateErrors method (flattened from errors.aggregateErrors)
	L.SetField(module, "errorsAggregateErrors", L.NewFunction(func(L *lua.LState) int {
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

	// errorsCategorizeError method (flattened from errors.categorizeError)
	L.SetField(module, "errorsCategorizeError", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// errorsWrap method (flattened from errors.wrap - alias for consistency with TODO)
	L.SetField(module, "errorsWrap", L.NewFunction(func(L *lua.LState) int {
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

	// errorsUnwrap method (flattened from errors.unwrap)
	L.SetField(module, "errorsUnwrap", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		errorData := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, errorData),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "unwrapError", args)
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

	// errorsIsType method (flattened from errors.isType)
	L.SetField(module, "errorsIsType", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		errorData := L.CheckTable(1)
		errorType := L.CheckString(2)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, errorData),
			engine.NewStringValue(errorType),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "isErrorType", args)
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

	// errorsGetStack method (flattened from errors.getStack)
	L.SetField(module, "errorsGetStack", L.NewFunction(func(L *lua.LState) int {
		if ua.errorsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("errors bridge not initialized"))
			return 2
		}

		errorData := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, errorData),
		}

		result, err := ua.errorsBridge.ExecuteMethod(context.Background(), "getErrorStack", args)
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

// addJSONMethods adds JSON processing methods (flattened to module level)
func (ua *UtilsAdapter) addJSONMethods(L *lua.LState, module *lua.LTable) {
	// jsonParse method (flattened from json.parse)
	L.SetField(module, "jsonParse", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

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

	// jsonToJSON method (flattened from json.toJSON)
	L.SetField(module, "jsonToJSON", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		data := L.CheckTable(1)
		options := L.OptTable(2, L.NewTable())

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

	// jsonValidateJSONSchema method (flattened from json.validateJSONSchema)
	L.SetField(module, "jsonValidateJSONSchema", L.NewFunction(func(L *lua.LState) int {
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

	// jsonExtractStructuredData method (flattened from json.extractStructuredData)
	L.SetField(module, "jsonExtractStructuredData", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// jsonEncode method (flattened from json.encode)
	L.SetField(module, "jsonEncode", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		data := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, data),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "encode", args)
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

	// jsonDecode method (flattened from json.decode)
	L.SetField(module, "jsonDecode", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "decode", args)
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

	// jsonValidate method (flattened from json.validate)
	L.SetField(module, "jsonValidate", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "validate", args)
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

	// jsonPrettify method (flattened from json.prettify)
	L.SetField(module, "jsonPrettify", L.NewFunction(func(L *lua.LState) int {
		if ua.jsonBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("json bridge not initialized"))
			return 2
		}

		text := L.CheckString(1)

		args := []engine.ScriptValue{
			engine.NewStringValue(text),
		}

		result, err := ua.jsonBridge.ExecuteMethod(context.Background(), "prettify", args)
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

// addLoggerMethods adds logger methods (flattened to module level)
func (ua *UtilsAdapter) addLoggerMethods(L *lua.LState, module *lua.LTable) {
	// loggerCreateLogger method (flattened from logger.createLogger)
	L.SetField(module, "loggerCreateLogger", L.NewFunction(func(L *lua.LState) int {
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

	// loggerLog method (flattened from logger.log)
	L.SetField(module, "loggerLog", L.NewFunction(func(L *lua.LState) int {
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

	// loggerSetLogLevel method (flattened from logger.setLogLevel)
	L.SetField(module, "loggerSetLogLevel", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// loggerError method (flattened from logger.error)
	L.SetField(module, "loggerError", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		contextTable := L.OptTable(2, L.NewTable())

		args := []engine.ScriptValue{
			engine.NewStringValue("ERROR"),
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

	// loggerWarn method (flattened from logger.warn)
	L.SetField(module, "loggerWarn", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		contextTable := L.OptTable(2, L.NewTable())

		args := []engine.ScriptValue{
			engine.NewStringValue("WARN"),
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

	// loggerInfo method (flattened from logger.info)
	L.SetField(module, "loggerInfo", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		contextTable := L.OptTable(2, L.NewTable())

		args := []engine.ScriptValue{
			engine.NewStringValue("INFO"),
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

	// loggerDebug method (flattened from logger.debug)
	L.SetField(module, "loggerDebug", L.NewFunction(func(L *lua.LState) int {
		if ua.loggerBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("logger bridge not initialized"))
			return 2
		}

		message := L.CheckString(1)
		contextTable := L.OptTable(2, L.NewTable())

		args := []engine.ScriptValue{
			engine.NewStringValue("DEBUG"),
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
}

// addSlogMethods adds structured logging methods (flattened to module level)
func (ua *UtilsAdapter) addSlogMethods(L *lua.LState, module *lua.LTable) {
	// slogInfo method (flattened from slog.info)
	L.SetField(module, "slogInfo", L.NewFunction(func(L *lua.LState) int {
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

	// slogWarn method (flattened from slog.warn)
	L.SetField(module, "slogWarn", L.NewFunction(func(L *lua.LState) int {
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

	// slogError method (flattened from slog.error)
	L.SetField(module, "slogError", L.NewFunction(func(L *lua.LState) int {
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

	// slogDebug method (flattened from slog.debug)
	L.SetField(module, "slogDebug", L.NewFunction(func(L *lua.LState) int {
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

	// Add method from TODO that may be missing
	// slogWithFields method (flattened from slog.withFields)
	L.SetField(module, "slogWithFields", L.NewFunction(func(L *lua.LState) int {
		if ua.slogBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("slog bridge not initialized"))
			return 2
		}

		fields := L.CheckTable(1)

		args := []engine.ScriptValue{
			ua.tableToScriptValue(L, fields),
		}

		result, err := ua.slogBridge.ExecuteMethod(context.Background(), "withFields", args)
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

// addGeneralMethods adds general utility methods (flattened to module level)
func (ua *UtilsAdapter) addGeneralMethods(L *lua.LState, module *lua.LTable) {
	// generalGenerateUUID method (flattened from general.generateUUID)
	L.SetField(module, "generalGenerateUUID", L.NewFunction(func(L *lua.LState) int {
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

	// generalHash method (flattened from general.hash)
	L.SetField(module, "generalHash", L.NewFunction(func(L *lua.LState) int {
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

	// generalRetry method (flattened from general.retry)
	L.SetField(module, "generalRetry", L.NewFunction(func(L *lua.LState) int {
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

	// generalSleep method (flattened from general.sleep)
	L.SetField(module, "generalSleep", L.NewFunction(func(L *lua.LState) int {
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

	// Add methods from TODO that may be missing
	// generalUuid method (flattened from general.uuid - alias)
	L.SetField(module, "generalUuid", L.NewFunction(func(L *lua.LState) int {
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

	// generalEncode method (flattened from general.encode)
	L.SetField(module, "generalEncode", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		data := L.CheckString(1)
		encoding := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(data),
			engine.NewStringValue(encoding),
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "encode", args)
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

	// generalDecode method (flattened from general.decode)
	L.SetField(module, "generalDecode", L.NewFunction(func(L *lua.LState) int {
		if ua.utilBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("util bridge not initialized"))
			return 2
		}

		data := L.CheckString(1)
		encoding := L.CheckString(2)

		args := []engine.ScriptValue{
			engine.NewStringValue(data),
			engine.NewStringValue(encoding),
		}

		result, err := ua.utilBridge.ExecuteMethod(context.Background(), "decode", args)
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
		// Auth methods (flattened)
		"authAuthenticate", "authValidateToken", "authRefreshToken", "authGenerateToken",
		"authHashPassword", "authVerifyPassword",
		// Debug methods (flattened)
		"debugSetLevel", "debugLog", "debugGetConfig", "debugTrace",
		"debugProfile", "debugDump", "debugAssert",
		// Error methods (flattened)
		"errorsCreateError", "errorsWrapError", "errorsAggregateErrors", "errorsCategorizeError",
		"errorsWrap", "errorsUnwrap", "errorsIsType", "errorsGetStack",
		// JSON methods (flattened)
		"jsonParse", "jsonToJSON", "jsonValidateJSONSchema", "jsonExtractStructuredData",
		"jsonEncode", "jsonDecode", "jsonValidate", "jsonPrettify",
		// LLM methods (flattened)
		"llmCreateProvider", "llmGenerateTyped", "llmTrackCost",
		"llmParseResponse", "llmFormatPrompt", "llmCountTokens", "llmSplitMessage",
		// Logger methods (flattened)
		"loggerCreateLogger", "loggerLog", "loggerSetLogLevel",
		"loggerError", "loggerWarn", "loggerInfo", "loggerDebug",
		// Slog methods (flattened)
		"slogInfo", "slogWarn", "slogError", "slogDebug", "slogWithFields",
		// General methods (flattened)
		"generalGenerateUUID", "generalHash", "generalRetry", "generalSleep",
		"generalUuid", "generalEncode", "generalDecode",
	}

	return methods
}
