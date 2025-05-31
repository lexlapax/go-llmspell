// ABOUTME: Tests for the Lua tools bridge implementation
// ABOUTME: Verifies tool registration, execution, and management through Lua

package bridges

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// mockToolBridge implements a test double for bridge.ToolBridge
type mockToolBridge struct {
	tools              map[string]*mockToolInfo
	registerCalled     bool
	registerErr        error
	executeCalled      bool
	executeResult      interface{}
	executeErr         error
	validateCalled     bool
	validateErr        error
	lastExecutedTool   string
	lastExecutedParams map[string]interface{}
}

type mockToolInfo struct {
	name        string
	description string
	parameters  map[string]interface{}
	handler     func(map[string]interface{}) (interface{}, error)
}

func newMockToolBridge() *mockToolBridge {
	return &mockToolBridge{
		tools: make(map[string]*mockToolInfo),
	}
}

func (m *mockToolBridge) RegisterTool(name, description string, parameters map[string]interface{}, handler func(map[string]interface{}) (interface{}, error)) error {
	m.registerCalled = true
	if m.registerErr != nil {
		return m.registerErr
	}

	m.tools[name] = &mockToolInfo{
		name:        name,
		description: description,
		parameters:  parameters,
		handler:     handler,
	}
	return nil
}

func (m *mockToolBridge) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	m.executeCalled = true
	m.lastExecutedTool = name
	m.lastExecutedParams = params

	if m.executeErr != nil {
		return nil, m.executeErr
	}

	if m.executeResult != nil {
		return m.executeResult, nil
	}

	tool, exists := m.tools[name]
	if !exists {
		return nil, errors.New("tool not found")
	}

	if tool.handler != nil {
		return tool.handler(params)
	}

	// Default behavior
	return map[string]interface{}{
		"tool":   name,
		"params": params,
		"result": "success",
	}, nil
}

func (m *mockToolBridge) GetTool(name string) (map[string]interface{}, error) {
	tool, exists := m.tools[name]
	if !exists {
		return nil, errors.New("tool not found")
	}

	return map[string]interface{}{
		"name":        tool.name,
		"description": tool.description,
		"parameters":  tool.parameters,
	}, nil
}

func (m *mockToolBridge) ListTools() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(m.tools))
	for _, tool := range m.tools {
		result = append(result, map[string]interface{}{
			"name":        tool.name,
			"description": tool.description,
			"parameters":  tool.parameters,
		})
	}
	return result
}

func (m *mockToolBridge) RemoveTool(name string) error {
	if _, exists := m.tools[name]; !exists {
		return errors.New("tool not found")
	}
	delete(m.tools, name)
	return nil
}

func (m *mockToolBridge) ValidateParameters(name string, params map[string]interface{}) error {
	m.validateCalled = true
	if m.validateErr != nil {
		return m.validateErr
	}

	_, exists := m.tools[name]
	if !exists {
		return errors.New("tool not found")
	}

	// Simple validation - just check if required params are present
	// In real implementation, this would use JSON schema validation
	return nil
}

func TestRegisterToolsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	err := RegisterToolsModule(L, mockBridge)
	require.NoError(t, err)

	// Check if tools module is registered
	tools := L.GetGlobal("tools")
	require.NotEqual(t, lua.LNil, tools)
	require.Equal(t, lua.LTTable, tools.Type())

	// Check if all functions are registered
	toolsTable := tools.(*lua.LTable)
	functions := []string{
		"register", "execute", "get", "list", "remove", "validate",
	}

	for _, fn := range functions {
		f := toolsTable.RawGetString(fn)
		assert.NotEqual(t, lua.LNil, f, "Function %s should be registered", fn)
		assert.Equal(t, lua.LTFunction, f.Type(), "tools.%s should be a function", fn)
	}
}

func TestToolsRegister(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Test successful registration
	err := L.DoString(`
		local success, err = tools.register(
			"test_tool",
			"A test tool",
			{
				input = {
					type = "string",
					description = "Input parameter",
					required = true
				}
			},
			function(params)
				return "Processed: " .. params.input
			end
		)
		
		assert(success == true, "Registration should succeed")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.registerCalled)
	assert.Contains(t, mockBridge.tools, "test_tool")

	// Test registration with error
	mockBridge.registerErr = errors.New("registration failed")
	err = L.DoString(`
		local success, err = tools.register(
			"fail_tool",
			"A failing tool",
			{},
			function(params)
				return "Should not be called"
			end
		)
		
		assert(success == nil, "Success should be nil on error")
		assert(err == "registration failed", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestToolsExecute(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Register a tool first
	mockBridge.tools["echo_tool"] = &mockToolInfo{
		name:        "echo_tool",
		description: "Echoes input",
		parameters:  map[string]interface{}{},
		handler: func(params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"echo": params["message"],
			}, nil
		},
	}

	// Test successful execution
	err := L.DoString(`
		local result, err = tools.execute("echo_tool", {message = "Hello"})
		
		assert(result ~= nil, "Result should not be nil")
		assert(err == nil, "Error should be nil")
		assert(result.echo == "Hello", "Echo should match input")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.executeCalled)
	assert.Equal(t, "echo_tool", mockBridge.lastExecutedTool)
	assert.Equal(t, "Hello", mockBridge.lastExecutedParams["message"])

	// Test execution without parameters
	err = L.DoString(`
		local result, err = tools.execute("echo_tool")
		
		assert(result ~= nil, "Result should not be nil")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)

	// Test execution of non-existent tool
	err = L.DoString(`
		local result, err = tools.execute("non_existent")
		
		assert(result == nil, "Result should be nil for non-existent tool")
		assert(err == "tool not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestToolsGet(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Add a tool
	mockBridge.tools["info_tool"] = &mockToolInfo{
		name:        "info_tool",
		description: "Tool with info",
		parameters: map[string]interface{}{
			"param1": map[string]interface{}{
				"type":        "string",
				"description": "First parameter",
			},
		},
	}

	// Test getting tool info
	err := L.DoString(`
		local info, err = tools.get("info_tool")
		
		assert(info ~= nil, "Info should not be nil")
		assert(err == nil, "Error should be nil")
		assert(info.name == "info_tool", "Name should match")
		assert(info.description == "Tool with info", "Description should match")
		assert(info.parameters ~= nil, "Parameters should not be nil")
		assert(info.parameters.param1 ~= nil, "Parameter should exist")
		assert(info.parameters.param1.type == "string", "Parameter type should match")
	`)
	require.NoError(t, err)

	// Test getting non-existent tool
	err = L.DoString(`
		local info, err = tools.get("non_existent")
		
		assert(info == nil, "Info should be nil for non-existent tool")
		assert(err == "tool not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestToolsList(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Add some tools
	mockBridge.tools["tool1"] = &mockToolInfo{
		name:        "tool1",
		description: "First tool",
		parameters:  map[string]interface{}{},
	}
	mockBridge.tools["tool2"] = &mockToolInfo{
		name:        "tool2",
		description: "Second tool",
		parameters:  map[string]interface{}{},
	}

	// Test listing tools
	err := L.DoString(`
		local tools_list = tools.list()
		
		assert(#tools_list >= 2, "Should have at least 2 tools")
		
		-- Check if our tools are in the list
		local found_tool1 = false
		local found_tool2 = false
		
		for _, tool in ipairs(tools_list) do
			if tool.name == "tool1" then
				found_tool1 = true
				assert(tool.description == "First tool", "Description should match")
			elseif tool.name == "tool2" then
				found_tool2 = true
				assert(tool.description == "Second tool", "Description should match")
			end
		end
		
		assert(found_tool1, "Tool1 should be in the list")
		assert(found_tool2, "Tool2 should be in the list")
	`)
	require.NoError(t, err)
}

func TestToolsRemove(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Add a tool
	mockBridge.tools["removable_tool"] = &mockToolInfo{
		name: "removable_tool",
	}

	// Test removing tool
	err := L.DoString(`
		local success, err = tools.remove("removable_tool")
		
		assert(success == true, "Remove should succeed")
		assert(err == nil, "Error should be nil")
		
		-- Try to get the removed tool
		local info, err2 = tools.get("removable_tool")
		assert(info == nil, "Removed tool should not exist")
		assert(err2 == "tool not found", "Should get not found error")
	`)
	require.NoError(t, err)

	// Verify tool was removed
	_, exists := mockBridge.tools["removable_tool"]
	assert.False(t, exists)

	// Test removing non-existent tool
	err = L.DoString(`
		local success, err = tools.remove("non_existent")
		
		assert(success == false, "Remove should fail for non-existent tool")
		assert(err == "tool not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestToolsValidate(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Add a tool
	mockBridge.tools["validated_tool"] = &mockToolInfo{
		name: "validated_tool",
		parameters: map[string]interface{}{
			"required_param": map[string]interface{}{
				"type":     "string",
				"required": true,
			},
		},
	}

	// Test successful validation
	err := L.DoString(`
		local success, err = tools.validate("validated_tool", {
			required_param = "value"
		})
		
		assert(success == true, "Validation should succeed")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.validateCalled)

	// Test validation with error
	mockBridge.validateErr = errors.New("validation failed")
	err = L.DoString(`
		local success, err = tools.validate("validated_tool", {})
		
		assert(success == false, "Validation should fail")
		assert(err == "validation failed", "Error message should match")
	`)
	require.NoError(t, err)

	// Test validation of non-existent tool
	mockBridge.validateErr = nil
	err = L.DoString(`
		local success, err = tools.validate("non_existent", {})
		
		assert(success == false, "Validation should fail for non-existent tool")
		assert(err == "tool not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestToolsIntegration(t *testing.T) {
	// Skip if not integration test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	L := lua.NewState()
	defer L.Close()

	// Create a real tool bridge
	toolRegistry := tools.DefaultRegistry
	realBridge := bridge.NewToolBridge(toolRegistry)
	require.NoError(t, RegisterToolsModule(L, realBridge))

	// Test registering and executing a tool through Lua
	err := L.DoString(`
		-- Register a simple calculator tool
		local success, err = tools.register(
			"add",
			"Adds two numbers",
			{
				a = {type = "number", description = "First number", required = true},
				b = {type = "number", description = "Second number", required = true}
			},
			function(params)
				return {result = params.a + params.b}
			end
		)
		
		assert(success == true, "Registration should succeed")
		
		-- Execute the tool
		local result, err = tools.execute("add", {a = 5, b = 3})
		assert(result ~= nil, "Result should not be nil")
		assert(result.result == 8, "5 + 3 should equal 8")
		
		-- List tools to verify registration
		local tools_list = tools.list()
		local found = false
		for _, tool in ipairs(tools_list) do
			if tool.name == "add" then
				found = true
				break
			end
		end
		assert(found, "Add tool should be in the list")
		
		-- Remove the tool
		local removed, err = tools.remove("add")
		assert(removed == true, "Tool should be removed successfully")
	`)
	require.NoError(t, err)
}

func TestLuaToolExecution(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockToolBridge()
	require.NoError(t, RegisterToolsModule(L, mockBridge))

	// Test that Lua functions are properly wrapped and executed
	err := L.DoString(`
		-- Register a tool that uses Lua-specific features
		local counter = 0
		local success, err = tools.register(
			"stateful_tool",
			"Tool with state",
			{
				increment = {type = "number", description = "Amount to increment", required = false}
			},
			function(params)
				local inc = params.increment or 1
				counter = counter + inc
				return {
					counter = counter,
					message = "Counter is now " .. counter
				}
			end
		)
		
		assert(success == true, "Registration should succeed")
	`)
	require.NoError(t, err)

	// Now execute the registered tool through the Go bridge
	tool := mockBridge.tools["stateful_tool"]
	require.NotNil(t, tool)
	require.NotNil(t, tool.handler)

	// Execute the tool multiple times
	result1, err := tool.handler(map[string]interface{}{"increment": float64(5)})
	require.NoError(t, err)
	resultMap1 := result1.(map[string]interface{})
	assert.Equal(t, float64(5), resultMap1["counter"])

	result2, err := tool.handler(map[string]interface{}{"increment": float64(3)})
	require.NoError(t, err)
	resultMap2 := result2.(map[string]interface{})
	assert.Equal(t, float64(8), resultMap2["counter"])
}
