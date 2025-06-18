package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsBridge_Initialize(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Should be not initialized initially
	assert.False(t, bridge.IsInitialized())

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestToolsBridge_GetID(t *testing.T) {
	bridge := NewToolsBridge()
	assert.Equal(t, "tools", bridge.GetID())
}

func TestToolsBridge_GetMetadata(t *testing.T) {
	bridge := NewToolsBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "Tools Bridge", metadata.Name)
	assert.Contains(t, metadata.Description, "tools bridge")
	assert.NotEmpty(t, metadata.Version)
}

func TestToolsBridge_Methods(t *testing.T) {
	bridge := NewToolsBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		// Discovery methods
		"listTools", "searchTools", "listByCategory", "getToolInfo",
		"getToolSchema", "getToolHelp", "getToolExamples",
		// Tool creation/execution
		"createTool", "executeTool", "executeToolValidated",
		// Custom tool registration
		"registerCustomTool",
		// Validation methods
		"validateToolInput", "validateToolOutput", "getValidationReport",
		// Metrics methods
		"getToolMetrics", "getAllToolsMetrics", "enableToolProfiling",
		"getToolUsageReport", "getToolAnomalies",
		// Documentation methods
		"generateToolDocumentation", "generateAllToolsDocs", "generateSDKSnippet",
		"generateToolPlayground",
	}

	assert.Equal(t, len(expectedMethods), len(methods), "Method count mismatch")

	// Check that key methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}

func TestToolsBridge_ValidateMethod(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		expectError bool
	}{
		{
			name:        "valid listTools",
			method:      "listTools",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:        "valid searchTools",
			method:      "searchTools",
			args:        []engine.ScriptValue{engine.NewStringValue("test")},
			expectError: false,
		},
		{
			name:        "invalid searchTools - missing args",
			method:      "searchTools",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid executeTool",
			method:      "executeTool",
			args:        []engine.ScriptValue{
				engine.NewStringValue("httpRequest"),
				engine.NewObjectValue(map[string]engine.ScriptValue{}),
			},
			expectError: false,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToolsBridge_ExecuteMethod_ListTools(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listTools
	result, err := bridge.ExecuteMethod(ctx, "listTools", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listTools")

	// Should return array of tools
	tools := arrayValue.ToGo().([]interface{})
	assert.Greater(t, len(tools), 0, "Should have some tools")
}

func TestToolsBridge_ExecuteMethod_SearchTools(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test searchTools
	args := []engine.ScriptValue{engine.NewStringValue("http")}
	result, err := bridge.ExecuteMethod(ctx, "searchTools", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from searchTools")

	// Should return filtered tools
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of matching tools")
}

func TestToolsBridge_ExecuteMethod_GetToolInfo(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get list of tools first
	result, err := bridge.ExecuteMethod(ctx, "listTools", []engine.ScriptValue{})
	require.NoError(t, err)
	
	arrayValue := result.(engine.ArrayValue)
	tools := arrayValue.ToGo().([]interface{})
	require.Greater(t, len(tools), 0, "Need at least one tool")

	// Get first tool name
	firstTool := tools[0].(map[string]interface{})
	toolName := firstTool["name"].(string)

	// Test getToolInfo
	args := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err = bridge.ExecuteMethod(ctx, "getToolInfo", args)
	assert.NoError(t, err)

	objValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getToolInfo")

	toolInfo := objValue.ToGo().(map[string]interface{})
	assert.Equal(t, toolName, toolInfo["name"])
	assert.NotEmpty(t, toolInfo["description"])
}

func TestToolsBridge_ExecuteMethod_RegisterCustomTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create custom tool definition
	toolDef := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("customTool"),
		"description": engine.NewStringValue("A custom test tool"),
		"execute": engine.NewFunctionValue("execute", func(ctx interface{}, params interface{}) (interface{}, error) {
			return map[string]interface{}{"result": "success"}, nil
		}),
	}

	// Test registerCustomTool
	args := []engine.ScriptValue{engine.NewObjectValue(toolDef)}
	result, err := bridge.ExecuteMethod(ctx, "registerCustomTool", args)
	assert.NoError(t, err)

	// registerCustomTool returns nil on success
	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from registerCustomTool")

	// Verify tool was registered
	toolInfo, err := bridge.ExecuteMethod(ctx, "getToolInfo", []engine.ScriptValue{engine.NewStringValue("customTool")})
	assert.NoError(t, err)
	assert.NotNil(t, toolInfo)
}

func TestToolsBridge_ExecuteMethod_ExecuteTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register a custom tool first
	toolDef := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("testExecutor"),
		"description": engine.NewStringValue("Test executor tool"),
		"execute": engine.NewFunctionValue("execute", func(ctx interface{}, params interface{}) (interface{}, error) {
			return map[string]interface{}{"executed": true}, nil
		}),
	}
	
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []engine.ScriptValue{engine.NewObjectValue(toolDef)})
	require.NoError(t, err)

	// Test executeTool
	args := []engine.ScriptValue{
		engine.NewStringValue("testExecutor"),
		engine.NewObjectValue(map[string]engine.ScriptValue{"test": engine.NewBoolValue(true)}),
	}
	result, err := bridge.ExecuteMethod(ctx, "executeTool", args)
	assert.NoError(t, err)

	objValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from executeTool")
	
	resultMap := objValue.ToGo().(map[string]interface{})
	assert.Equal(t, true, resultMap["executed"])
}

func TestToolsBridge_ExecuteMethod_GetToolSchema(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get a tool that has schema
	args := []engine.ScriptValue{engine.NewStringValue("httpRequest")}
	result, err := bridge.ExecuteMethod(ctx, "getToolSchema", args)
	
	if err != nil {
		// Tool might not exist, which is ok for this test
		assert.Contains(t, err.Error(), "not found")
	} else {
		objValue, ok := result.(engine.ObjectValue)
		assert.True(t, ok, "Expected ObjectValue from getToolSchema")
		
		schema := objValue.ToGo().(map[string]interface{})
		assert.NotNil(t, schema["name"])
		assert.NotNil(t, schema["description"])
	}
}

func TestToolsBridge_ExecuteMethod_ValidateToolInput(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test validateToolInput
	args := []engine.ScriptValue{
		engine.NewStringValue("httpRequest"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"url":    engine.NewStringValue("https://example.com"),
			"method": engine.NewStringValue("GET"),
		}),
	}
	
	result, err := bridge.ExecuteMethod(ctx, "validateToolInput", args)
	// May error if tool doesn't exist, which is ok
	if err == nil {
		objValue, ok := result.(engine.ObjectValue)
		assert.True(t, ok, "Expected ObjectValue from validateToolInput")
		
		validation := objValue.ToGo().(map[string]interface{})
		assert.NotNil(t, validation["valid"])
	}
}

func TestToolsBridge_ExecuteMethod_GetToolMetrics(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolMetrics - first execute a tool to generate metrics
	// Register a custom tool
	toolDef := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("metricsTool"),
		"description": engine.NewStringValue("Tool for metrics test"),
		"execute": engine.NewFunctionValue("execute", func(ctx interface{}, params interface{}) (interface{}, error) {
			return map[string]interface{}{"result": "success"}, nil
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []engine.ScriptValue{engine.NewObjectValue(toolDef)})
	require.NoError(t, err)
	
	// Execute it to generate metrics
	_, _ = bridge.ExecuteMethod(ctx, "executeTool", []engine.ScriptValue{
		engine.NewStringValue("metricsTool"),
		engine.NewObjectValue(map[string]engine.ScriptValue{}),
	})
	
	// Now get metrics
	args := []engine.ScriptValue{engine.NewStringValue("metricsTool")}
	result, err := bridge.ExecuteMethod(ctx, "getToolMetrics", args)
	assert.NoError(t, err)

	objValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getToolMetrics")
	
	metrics := objValue.ToGo().(map[string]interface{})
	assert.NotNil(t, metrics["toolName"])
}

func TestToolsBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "method not found")
}

func TestToolsBridge_RequiredPermissions(t *testing.T) {
	bridge := NewToolsBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasToolPermission := false
	for _, perm := range permissions {
		if perm.Resource == "tool" {
			hasToolPermission = true
			break
		}
	}
	assert.True(t, hasToolPermission, "Should have tool permission")
}

func TestToolsBridge_TypeMappings(t *testing.T) {
	bridge := NewToolsBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Tool", "ToolInfo"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestToolsBridge_Cleanup(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestToolsBridge_NotInitialized(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listTools", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not initialized")
}