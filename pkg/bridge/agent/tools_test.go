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
	assert.Equal(t, "2.1.0", metadata.Version)
	assert.Contains(t, metadata.Description, "tools bridge")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestToolsBridge_Methods(t *testing.T) {
	bridge := NewToolsBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"discoverTools", "executeTool", "validateTool", "getToolSchema",
		"listAvailableTools", "getToolDefinition", "createCustomTool",
		"registerTool", "unregisterTool", "enableTool", "disableTool",
		"isToolEnabled", "getToolMetrics", "resetToolMetrics", "benchmarkTool",
		"getToolBenchmarks", "setToolTimeout", "getToolTimeout",
		"addToolValidator", "removeToolValidator", "validateToolInput",
		"validateToolOutput", "generateToolDocumentation", "exportToolDocumentation",
		"createToolChain", "executeToolChain", "getToolChain", "removeToolChain",
		"listToolChains", "addToolToChain", "removeToolFromChain",
		"setToolChainTimeout", "executeToolsInParallel", "getToolExecutionHistory",
		"clearToolExecutionHistory", "exportToolExecutionReport",
	}

	assert.GreaterOrEqual(t, len(methods), len(expectedMethods))

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
			name:        "valid discoverTools",
			method:      "discoverTools",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:        "valid executeTool",
			method:      "executeTool",
			args:        []engine.ScriptValue{engine.NewStringValue("calculator"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
			expectError: false,
		},
		{
			name:        "invalid executeTool - missing args",
			method:      "executeTool",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid getToolDefinition",
			method:      "getToolDefinition",
			args:        []engine.ScriptValue{engine.NewStringValue("calculator")},
			expectError: false,
		},
		{
			name:        "valid createCustomTool",
			method:      "createCustomTool",
			args:        []engine.ScriptValue{engine.NewStringValue("custom"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
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

func TestToolsBridge_ExecuteMethod_DiscoverTools(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test discoverTools
	result, err := bridge.ExecuteMethod(ctx, "discoverTools", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from discoverTools")

	// Should return array of discovered tools
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools")
}

func TestToolsBridge_ExecuteMethod_ListAvailableTools(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listAvailableTools
	result, err := bridge.ExecuteMethod(ctx, "listAvailableTools", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listAvailableTools")

	// Should return array of available tools
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of available tools")
}

func TestToolsBridge_ExecuteMethod_GetToolDefinition(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolDefinition with non-existent tool
	args := []engine.ScriptValue{engine.NewStringValue("non-existent-tool")}
	result, err := bridge.ExecuteMethod(ctx, "getToolDefinition", args)
	assert.NoError(t, err) // Should return error value, not Go error

	// Could be either ObjectValue (success) or ErrorValue (not found)
	switch result.(type) {
	case engine.ObjectValue:
		// Tool exists and definition returned
	case engine.ErrorValue:
		// Tool not found - this is expected for non-existent tools
		errorValue := result.(engine.ErrorValue)
		assert.Contains(t, errorValue.Error().Error(), "not found")
	default:
		t.Fatalf("Expected ObjectValue or ErrorValue, got %T", result)
	}
}

func TestToolsBridge_ExecuteMethod_CreateCustomTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test createCustomTool
	toolDefinition := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("custom-tool"),
		"description": engine.NewStringValue("A custom tool for testing"),
		"category":    engine.NewStringValue("test"),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("custom-tool"),
		engine.NewObjectValue(toolDefinition),
	}

	result, err := bridge.ExecuteMethod(ctx, "createCustomTool", args)
	assert.NoError(t, err)

	// Should return ObjectValue (tool created) or ErrorValue (creation failed)
	switch result.(type) {
	case engine.ObjectValue:
		// Tool created successfully
	case engine.ErrorValue:
		// Creation failed - this might be expected if tools are read-only
	default:
		t.Fatalf("Expected ObjectValue or ErrorValue, got %T", result)
	}
}

func TestToolsBridge_ExecuteMethod_ExecuteTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test executeTool with simple parameters
	toolParams := map[string]engine.ScriptValue{
		"input": engine.NewStringValue("test input"),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("echo"), // Assuming echo tool exists
		engine.NewObjectValue(toolParams),
	}

	result, err := bridge.ExecuteMethod(ctx, "executeTool", args)
	assert.NoError(t, err)

	// Could be any type depending on tool output
	assert.NotNil(t, result, "Should return some result from tool execution")
}

func TestToolsBridge_ExecuteMethod_ValidateTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test validateTool
	toolDefinition := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("test-tool"),
		"description": engine.NewStringValue("A test tool"),
	}

	args := []engine.ScriptValue{engine.NewObjectValue(toolDefinition)}
	result, err := bridge.ExecuteMethod(ctx, "validateTool", args)
	assert.NoError(t, err)

	// Should return ValidationResult
	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from validateTool")

	validation := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, validation, "valid")
}

func TestToolsBridge_ExecuteMethod_GetToolSchema(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolSchema
	args := []engine.ScriptValue{engine.NewStringValue("calculator")}
	result, err := bridge.ExecuteMethod(ctx, "getToolSchema", args)
	assert.NoError(t, err)

	// Could be ObjectValue (schema found) or ErrorValue (not found)
	switch result.(type) {
	case engine.ObjectValue:
		// Schema found
	case engine.ErrorValue:
		// Schema not found
	default:
		t.Fatalf("Expected ObjectValue or ErrorValue, got %T", result)
	}
}

func TestToolsBridge_ExecuteMethod_RegisterUnregisterTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	toolName := "test-tool"
	toolDefinition := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue(toolName),
		"description": engine.NewStringValue("A test tool"),
	}

	// Test registerTool
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(toolName),
		engine.NewObjectValue(toolDefinition),
	}

	result, err := bridge.ExecuteMethod(ctx, "registerTool", registerArgs)
	assert.NoError(t, err)

	// Should return success indicator or error
	switch result.(type) {
	case engine.BoolValue:
		// Registration result
	case engine.ErrorValue:
		// Registration failed
	case engine.NilValue:
		// Registration completed without explicit result
	default:
		t.Fatalf("Expected BoolValue, ErrorValue, or NilValue, got %T", result)
	}

	// Test unregisterTool
	unregisterArgs := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err = bridge.ExecuteMethod(ctx, "unregisterTool", unregisterArgs)
	assert.NoError(t, err)

	// Should return success indicator
	switch result.(type) {
	case engine.BoolValue:
		// Unregistration result
	case engine.ErrorValue:
		// Unregistration failed
	case engine.NilValue:
		// Unregistration completed
	default:
		t.Fatalf("Expected BoolValue, ErrorValue, or NilValue, got %T", result)
	}
}

func TestToolsBridge_ExecuteMethod_EnableDisableTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	toolName := "test-tool"

	// Test enableTool
	enableArgs := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err := bridge.ExecuteMethod(ctx, "enableTool", enableArgs)
	assert.NoError(t, err)

	_, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from enableTool")

	// Test isToolEnabled
	isEnabledArgs := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err = bridge.ExecuteMethod(ctx, "isToolEnabled", isEnabledArgs)
	assert.NoError(t, err)

	_, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isToolEnabled")

	// Test disableTool
	disableArgs := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err = bridge.ExecuteMethod(ctx, "disableTool", disableArgs)
	assert.NoError(t, err)

	_, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from disableTool")
}

func TestToolsBridge_ExecuteMethod_GetToolMetrics(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolMetrics
	args := []engine.ScriptValue{engine.NewStringValue("test-tool")}
	result, err := bridge.ExecuteMethod(ctx, "getToolMetrics", args)
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getToolMetrics")

	metrics := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, metrics, "execution_count")
	assert.Contains(t, metrics, "success_count")
}

func TestToolsBridge_ExecuteMethod_BenchmarkTool(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test benchmarkTool
	benchmarkConfig := map[string]engine.ScriptValue{
		"iterations": engine.NewNumberValue(10),
		"timeout":    engine.NewNumberValue(5000),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("echo"),
		engine.NewObjectValue(benchmarkConfig),
	}

	result, err := bridge.ExecuteMethod(ctx, "benchmarkTool", args)
	assert.NoError(t, err)

	// Should return benchmark results
	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from benchmarkTool")

	benchmark := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, benchmark, "average_duration")
	assert.Contains(t, benchmark, "total_iterations")
}

func TestToolsBridge_ExecuteMethod_SetGetToolTimeout(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	toolName := "test-tool"
	timeout := float64(5000) // 5 seconds

	// Test setToolTimeout
	setArgs := []engine.ScriptValue{
		engine.NewStringValue(toolName),
		engine.NewNumberValue(timeout),
	}
	result, err := bridge.ExecuteMethod(ctx, "setToolTimeout", setArgs)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from setToolTimeout")

	// Test getToolTimeout
	getArgs := []engine.ScriptValue{engine.NewStringValue(toolName)}
	result, err = bridge.ExecuteMethod(ctx, "getToolTimeout", getArgs)
	assert.NoError(t, err)

	numberValue, ok := result.(engine.NumberValue)
	assert.True(t, ok, "Expected NumberValue from getToolTimeout")
	assert.Equal(t, timeout, numberValue.Value(), "Timeout should match what was set")
}

func TestToolsBridge_ExecuteMethod_CreateToolChain(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test createToolChain
	chainConfig := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("test-chain"),
		"description": engine.NewStringValue("A test tool chain"),
		"tools": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("tool1"),
			engine.NewStringValue("tool2"),
		}),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("test-chain"),
		engine.NewObjectValue(chainConfig),
	}

	result, err := bridge.ExecuteMethod(ctx, "createToolChain", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (chain ID) from createToolChain")
	assert.Equal(t, "test-chain", stringValue.Value(), "Chain ID should match input")
}

func TestToolsBridge_ExecuteMethod_ExecuteToolsInParallel(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test executeToolsInParallel
	toolSpecs := []engine.ScriptValue{
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":   engine.NewStringValue("tool1"),
			"params": engine.NewObjectValue(map[string]engine.ScriptValue{}),
		}),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":   engine.NewStringValue("tool2"),
			"params": engine.NewObjectValue(map[string]engine.ScriptValue{}),
		}),
	}

	args := []engine.ScriptValue{engine.NewArrayValue(toolSpecs)}
	result, err := bridge.ExecuteMethod(ctx, "executeToolsInParallel", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from executeToolsInParallel")

	results := arrayValue.ToGo().([]interface{})
	assert.Equal(t, 2, len(results), "Should return results for both tools")
}

func TestToolsBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewToolsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestToolsBridge_RequiredPermissions(t *testing.T) {
	bridge := NewToolsBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasToolsPermission := false
	for _, perm := range permissions {
		if perm.Resource == "tools" {
			hasToolsPermission = true
			break
		}
	}
	assert.True(t, hasToolsPermission, "Should have tools permission")
}

func TestToolsBridge_TypeMappings(t *testing.T) {
	bridge := NewToolsBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Tool", "ToolDefinition", "ToolChain"}
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
	result, err := bridge.ExecuteMethod(ctx, "discoverTools", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
