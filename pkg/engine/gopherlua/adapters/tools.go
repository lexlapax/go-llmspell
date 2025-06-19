// ABOUTME: Tools bridge adapter that exposes go-llms tool functionality to Lua scripts
// ABOUTME: Provides tool discovery, execution, registration, validation, and metrics capabilities

package adapters

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// ToolsAdapter bridges go-llms tools functionality to Lua
type ToolsAdapter struct {
	bridge         engine.Bridge
	registryBridge engine.Bridge // Tool registry bridge for enhanced functionality
}

// NewToolsAdapter creates a new tools adapter
func NewToolsAdapter(bridge engine.Bridge) *ToolsAdapter {
	return &ToolsAdapter{
		bridge: bridge,
	}
}

// NewToolsAdapterWithRegistry creates a new tools adapter with registry bridge
func NewToolsAdapterWithRegistry(bridge engine.Bridge, registryBridge engine.Bridge) *ToolsAdapter {
	return &ToolsAdapter{
		bridge:         bridge,
		registryBridge: registryBridge,
	}
}

// GetAdapterName returns the adapter name
func (ta *ToolsAdapter) GetAdapterName() string {
	return "tools"
}

// GetBridge returns the underlying bridge
func (ta *ToolsAdapter) GetBridge() engine.Bridge {
	return ta.bridge
}

// CreateLuaModule creates a Lua module for tools
func (ta *ToolsAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Set adapter info
		L.SetField(module, "_adapter", lua.LString("tools"))
		L.SetField(module, "_version", lua.LString("2.0.0"))

		// Add tool categories
		categories := L.NewTable()
		L.SetField(categories, "MATH", lua.LString("math"))
		L.SetField(categories, "API", lua.LString("api"))
		L.SetField(categories, "TEXT", lua.LString("text"))
		L.SetField(categories, "FILE", lua.LString("file"))
		L.SetField(categories, "SYSTEM", lua.LString("system"))
		L.SetField(module, "CATEGORIES", categories)

		// Add permission types
		permissions := L.NewTable()
		L.SetField(permissions, "NETWORK", lua.LString("network"))
		L.SetField(permissions, "FILE_READ", lua.LString("file_read"))
		L.SetField(permissions, "FILE_WRITE", lua.LString("file_write"))
		L.SetField(permissions, "SYSTEM", lua.LString("system"))
		L.SetField(module, "PERMISSIONS", permissions)

		// Add resource usage levels
		resourceUsage := L.NewTable()
		L.SetField(resourceUsage, "LOW", lua.LString("low"))
		L.SetField(resourceUsage, "MEDIUM", lua.LString("medium"))
		L.SetField(resourceUsage, "HIGH", lua.LString("high"))
		L.SetField(module, "RESOURCE_USAGE", resourceUsage)

		// Core tool discovery methods
		L.SetField(module, "listTools", L.NewFunction(ta.listTools))
		L.SetField(module, "searchTools", L.NewFunction(ta.searchTools))
		L.SetField(module, "getToolInfo", L.NewFunction(ta.getToolInfo))
		L.SetField(module, "getToolSchema", L.NewFunction(ta.getToolSchema))
		L.SetField(module, "getCategories", L.NewFunction(ta.getToolCategories))
		L.SetField(module, "listByCategory", L.NewFunction(ta.listToolsByCategory))
		L.SetField(module, "listByTags", L.NewFunction(ta.listToolsByTags))

		// Tool execution methods
		L.SetField(module, "executeTool", L.NewFunction(ta.executeTool))
		L.SetField(module, "executeAsync", L.NewFunction(ta.executeToolAsync))

		// Custom tool registration
		L.SetField(module, "registerCustomTool", L.NewFunction(ta.registerCustomTool))

		// Validation methods
		L.SetField(module, "validateToolInput", L.NewFunction(ta.validateToolInput))

		// Metrics methods
		L.SetField(module, "getToolMetrics", L.NewFunction(ta.getToolMetrics))

		// Builder pattern support
		L.SetField(module, "createBuilder", L.NewFunction(ta.createToolBuilder))

		// Registry bridge methods (if available)
		if ta.registryBridge != nil {
			// Tool discovery methods from registry
			L.SetField(module, "getTool", L.NewFunction(ta.getTool))
			L.SetField(module, "listToolsByPermission", L.NewFunction(ta.listToolsByPermission))
			L.SetField(module, "listToolsByResourceUsage", L.NewFunction(ta.listToolsByResourceUsage))

			// Tool documentation from registry
			L.SetField(module, "getToolDocumentation", L.NewFunction(ta.getToolDocumentation))

			// MCP export functionality
			L.SetField(module, "exportToolToMCP", L.NewFunction(ta.exportToolToMCP))
			L.SetField(module, "exportAllToolsToMCP", L.NewFunction(ta.exportAllToolsToMCP))

			// Registry management
			L.SetField(module, "clearRegistry", L.NewFunction(ta.clearRegistry))
			L.SetField(module, "getRegistryStats", L.NewFunction(ta.getRegistryStats))
		}

		L.Push(module)
		return 1
	}
}

// GetMethods returns available adapter methods
func (ta *ToolsAdapter) GetMethods() []string {
	methods := []string{
		"listTools", "searchTools", "getToolInfo", "getToolSchema",
		"getCategories", "listByCategory", "listByTags",
		"executeTool", "executeAsync",
		"registerCustomTool",
		"validateToolInput",
		"getToolMetrics",
		"createBuilder",
	}

	// Add registry methods if available
	if ta.registryBridge != nil {
		methods = append(methods,
			"getTool", "listToolsByPermission", "listToolsByResourceUsage",
			"getToolDocumentation", "exportToolToMCP", "exportAllToolsToMCP",
			"clearRegistry", "getRegistryStats",
		)
	}

	return methods
}

// RegisterAsModule registers the adapter as a module in the module system
func (ta *ToolsAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if ta.GetBridge() != nil {
		bridgeMetadata = ta.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "Tools Adapter",
			Description: "Tool discovery and execution functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // Tools module has no dependencies by default
		LoadFunc:     ta.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// Core tool discovery methods

func (ta *ToolsAdapter) listTools(L *lua.LState) int {
	result, err := ta.bridge.ExecuteMethod(context.Background(), "listTools", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

func (ta *ToolsAdapter) searchTools(L *lua.LState) int {
	query := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(query),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "searchTools", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

func (ta *ToolsAdapter) getToolInfo(L *lua.LState) int {
	toolName := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "getToolInfo", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (ta *ToolsAdapter) getToolSchema(L *lua.LState) int {
	toolName := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "getToolSchema", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (ta *ToolsAdapter) getToolCategories(L *lua.LState) int {
	result, err := ta.bridge.ExecuteMethod(context.Background(), "getToolCategories", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

func (ta *ToolsAdapter) listToolsByCategory(L *lua.LState) int {
	category := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(category),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "listToolsByCategory", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

func (ta *ToolsAdapter) listToolsByTags(L *lua.LState) int {
	tags := L.CheckTable(1)

	// Convert Lua table to array of strings
	tagArray := []engine.ScriptValue{}
	tags.ForEach(func(k, v lua.LValue) {
		if str, ok := v.(lua.LString); ok {
			tagArray = append(tagArray, engine.NewStringValue(string(str)))
		}
	})

	args := []engine.ScriptValue{
		engine.NewArrayValue(tagArray),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "listToolsByTags", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

// Tool execution methods

func (ta *ToolsAdapter) executeTool(L *lua.LState) int {
	toolName := L.CheckString(1)
	params := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
		luaToScriptValue(params),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "executeTool", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (ta *ToolsAdapter) executeToolAsync(L *lua.LState) int {
	toolName := L.CheckString(1)
	params := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
		luaToScriptValue(params),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "executeToolAsync", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// Custom tool registration

func (ta *ToolsAdapter) registerCustomTool(L *lua.LState) int {
	toolDef := L.CheckTable(1)

	args := []engine.ScriptValue{
		luaToScriptValue(toolDef),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "registerCustomTool", args)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result to boolean
	if boolVal, ok := result.(engine.BoolValue); ok {
		L.Push(lua.LBool(boolVal.Value()))
	} else {
		L.Push(lua.LTrue) // Default to true if method succeeded
	}
	L.Push(lua.LNil)
	return 2
}

// Validation methods

func (ta *ToolsAdapter) validateToolInput(L *lua.LState) int {
	toolName := L.CheckString(1)
	params := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
		luaToScriptValue(params),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "validateToolInput", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// Metrics methods

func (ta *ToolsAdapter) getToolMetrics(L *lua.LState) int {
	toolName := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.bridge.ExecuteMethod(context.Background(), "getToolMetrics", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// Builder pattern support

func (ta *ToolsAdapter) createToolBuilder(L *lua.LState) int {
	toolName := L.CheckString(1)

	// Create builder table with methods
	builder := L.NewTable()

	// Store tool definition
	toolDef := L.NewTable()
	L.SetField(toolDef, "name", lua.LString(toolName))
	L.SetField(builder, "_toolDef", toolDef)
	L.SetField(builder, "_adapter", L.NewUserData()) // Store adapter reference

	// Builder methods
	L.SetField(builder, "withDescription", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		description := L.CheckString(2)

		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)
		L.SetField(toolDef, "description", lua.LString(description))

		L.Push(self) // Return self for chaining
		return 1
	}))

	L.SetField(builder, "withCategory", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		category := L.CheckString(2)

		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)
		L.SetField(toolDef, "category", lua.LString(category))

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "withTags", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		tags := L.CheckTable(2)

		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)
		L.SetField(toolDef, "tags", tags)

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "withParameter", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		name := L.CheckString(2)
		paramType := L.CheckString(3)
		description := L.CheckString(4)
		required := L.CheckBool(5)

		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)

		// Get or create parameters schema
		var paramsSchema *lua.LTable
		if ps := L.GetField(toolDef, "parameterSchema"); ps == lua.LNil {
			paramsSchema = L.NewTable()
			props := L.NewTable()
			reqArray := L.NewTable()
			L.SetField(paramsSchema, "type", lua.LString("object"))
			L.SetField(paramsSchema, "properties", props)
			L.SetField(paramsSchema, "required", reqArray)
			L.SetField(toolDef, "parameterSchema", paramsSchema)
		} else {
			paramsSchema = ps.(*lua.LTable)
		}

		// Add parameter
		props := L.GetField(paramsSchema, "properties").(*lua.LTable)
		paramDef := L.NewTable()
		L.SetField(paramDef, "type", lua.LString(paramType))
		L.SetField(paramDef, "description", lua.LString(description))
		L.SetField(props, name, paramDef)

		// Add to required if needed
		if required {
			reqArray := L.GetField(paramsSchema, "required").(*lua.LTable)
			reqArray.Append(lua.LString(name))
		}

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "withExecute", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		executeFn := L.CheckFunction(2)

		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)
		L.SetField(toolDef, "execute", executeFn)

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "build", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		toolDef := L.GetField(self, "_toolDef").(*lua.LTable)

		// Register the tool using the adapter
		args := []engine.ScriptValue{
			luaToScriptValue(toolDef),
		}

		result, err := ta.bridge.ExecuteMethod(context.Background(), "registerCustomTool", args)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result to boolean
		if boolVal, ok := result.(engine.BoolValue); ok {
			L.Push(lua.LBool(boolVal.Value()))
		} else {
			L.Push(lua.LTrue)
		}
		L.Push(lua.LNil)
		return 2
	}))

	L.Push(builder)
	return 1
}

// Helper function to push array as multiple return values
func pushArrayAsMultipleValues(L *lua.LState, arrayVal engine.ArrayValue) int {
	elements := arrayVal.Elements()
	for _, val := range elements {
		L.Push(scriptValueToLua(L, val))
	}
	return len(elements)
}

// Helper function to convert ScriptValue to Lua value
func scriptValueToLua(L *lua.LState, sv engine.ScriptValue) lua.LValue {
	switch v := sv.(type) {
	case engine.StringValue:
		return lua.LString(v.Value())
	case engine.NumberValue:
		return lua.LNumber(v.Value())
	case engine.BoolValue:
		return lua.LBool(v.Value())
	case engine.ArrayValue:
		table := L.NewTable()
		for _, elem := range v.Elements() {
			table.Append(scriptValueToLua(L, elem))
		}
		return table
	case engine.ObjectValue:
		table := L.NewTable()
		for k, val := range v.Fields() {
			L.SetField(table, k, scriptValueToLua(L, val))
		}
		return table
	case engine.NilValue:
		return lua.LNil
	case engine.ErrorValue:
		return lua.LString(fmt.Sprintf("error: %v", v.Error()))
	default:
		return lua.LNil
	}
}

// Helper function to convert Lua value to ScriptValue
func luaToScriptValue(lv lua.LValue) engine.ScriptValue {
	switch v := lv.(type) {
	case lua.LString:
		return engine.NewStringValue(string(v))
	case lua.LNumber:
		return engine.NewNumberValue(float64(v))
	case lua.LBool:
		return engine.NewBoolValue(bool(v))
	case *lua.LTable:
		// Check if it's an array by looking for numeric keys starting from 1
		length := v.Len()
		isArray := false

		if length > 0 {
			// Check if all keys 1..length exist
			isArray = true
			for i := 1; i <= length; i++ {
				if v.RawGetInt(i) == lua.LNil {
					isArray = false
					break
				}
			}

			// Also check if there are any string keys (which would make it an object)
			if isArray {
				hasStringKeys := false
				v.ForEach(func(k, val lua.LValue) {
					if _, ok := k.(lua.LString); ok {
						hasStringKeys = true
					}
				})
				if hasStringKeys {
					isArray = false
				}
			}
		}

		if isArray && length > 0 {
			values := make([]engine.ScriptValue, length)
			for i := 1; i <= length; i++ {
				values[i-1] = luaToScriptValue(v.RawGetInt(i))
			}
			return engine.NewArrayValue(values)
		}

		// Object (or empty table treated as object)
		obj := make(map[string]engine.ScriptValue)
		v.ForEach(func(k, val lua.LValue) {
			if key, ok := k.(lua.LString); ok {
				obj[string(key)] = luaToScriptValue(val)
			}
		})
		return engine.NewObjectValue(obj)
	case *lua.LFunction:
		// Functions can't be directly converted, return a placeholder
		return engine.NewFunctionValue("lua_function", func(ctx context.Context, args []interface{}) (interface{}, error) {
			return nil, fmt.Errorf("lua function execution not supported")
		})
	default:
		return engine.NewNilValue()
	}
}

// Registry bridge methods

// getTool gets complete tool information from registry
func (ta *ToolsAdapter) getTool(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	toolName := L.CheckString(1)
	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "getTool", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// listToolsByPermission lists tools requiring specific permission
func (ta *ToolsAdapter) listToolsByPermission(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	permission := L.CheckString(1)
	args := []engine.ScriptValue{
		engine.NewStringValue(permission),
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "listToolsByPermission", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

// listToolsByResourceUsage lists tools matching resource criteria
func (ta *ToolsAdapter) listToolsByResourceUsage(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	criteria := L.CheckTable(1)
	args := []engine.ScriptValue{
		luaToScriptValue(criteria),
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "listToolsByResourceUsage", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayVal, ok := result.(engine.ArrayValue); ok {
		return pushArrayAsMultipleValues(L, arrayVal)
	}

	L.Push(lua.LNil)
	L.Push(lua.LString("unexpected result type"))
	return 2
}

// getToolDocumentation gets comprehensive documentation for a tool
func (ta *ToolsAdapter) getToolDocumentation(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	toolName := L.CheckString(1)
	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "getToolDocumentation", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// exportToolToMCP exports single tool to MCP format
func (ta *ToolsAdapter) exportToolToMCP(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	toolName := L.CheckString(1)
	args := []engine.ScriptValue{
		engine.NewStringValue(toolName),
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "exportToolToMCP", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// exportAllToolsToMCP exports all tools to MCP catalog
func (ta *ToolsAdapter) exportAllToolsToMCP(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "exportAllToolsToMCP", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// clearRegistry clears all tools from registry (testing only)
func (ta *ToolsAdapter) clearRegistry(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "clearRegistry", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// getRegistryStats gets registry statistics and metrics
func (ta *ToolsAdapter) getRegistryStats(L *lua.LState) int {
	if ta.registryBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("registry bridge not available"))
		return 2
	}

	result, err := ta.registryBridge.ExecuteMethod(context.Background(), "getRegistryStats", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}
