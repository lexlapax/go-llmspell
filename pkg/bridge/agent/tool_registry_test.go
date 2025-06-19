package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsRegistryBridge_Initialize(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestToolsRegistryBridge_GetID(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	assert.Equal(t, "tools_registry", bridge.GetID())
}

func TestToolsRegistryBridge_GetMetadata(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "tools_registry", metadata.Name)
	assert.Equal(t, "v1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "tools registry")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestToolsRegistryBridge_Methods(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"listTools", "getTool", "searchTools", "listToolsByCategory",
		"listToolsByTags", "getToolCategories", "listToolsByPermission",
		"listToolsByResourceUsage", "getToolDocumentation", "registerTool",
		"exportToolToMCP", "exportAllToolsToMCP", "clearRegistry", "getRegistryStats",
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

func TestToolsRegistryBridge_ValidateMethod(t *testing.T) {
	bridge := NewToolsRegistryBridge()
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
			name:        "valid getTool",
			method:      "getTool",
			args:        []engine.ScriptValue{sv("calculator")},
			expectError: false,
		},
		{
			name:        "invalid getTool - missing args",
			method:      "getTool",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid searchTools",
			method:      "searchTools",
			args:        []engine.ScriptValue{sv("math")},
			expectError: false,
		},
		{
			name:        "valid listToolsByTags",
			method:      "listToolsByTags",
			args:        []engine.ScriptValue{svArray("math")},
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

func TestToolsRegistryBridge_ExecuteMethod_ListTools(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listTools
	result, err := bridge.ExecuteMethod(ctx, "listTools", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listTools")

	// Should return array (may be empty if no tools registered)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools")
}

func TestToolsRegistryBridge_ExecuteMethod_GetTool(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getTool with non-existent tool
	args := []engine.ScriptValue{sv("non-existent-tool")}
	result, err := bridge.ExecuteMethod(ctx, "getTool", args)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for non-existent tool")
	assert.Contains(t, errorValue.Error().Error(), "not found")
}

func TestToolsRegistryBridge_ExecuteMethod_SearchTools(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test searchTools
	args := []engine.ScriptValue{sv("test")}
	result, err := bridge.ExecuteMethod(ctx, "searchTools", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from searchTools")

	// Should return array (may be empty)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of matching tools")
}

func TestToolsRegistryBridge_ExecuteMethod_ListToolsByCategory(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listToolsByCategory
	args := []engine.ScriptValue{sv("math")}
	result, err := bridge.ExecuteMethod(ctx, "listToolsByCategory", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listToolsByCategory")

	// Should return array (may be empty)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools in category")
}

func TestToolsRegistryBridge_ExecuteMethod_ListToolsByTags(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listToolsByTags
	tags := svArray("math", "utility")
	args := []engine.ScriptValue{tags}
	result, err := bridge.ExecuteMethod(ctx, "listToolsByTags", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listToolsByTags")

	// Should return array (may be empty)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools with tags")
}

func TestToolsRegistryBridge_ExecuteMethod_GetToolCategories(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolCategories
	result, err := bridge.ExecuteMethod(ctx, "getToolCategories", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from getToolCategories")

	// Should return array of categories
	categories := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(categories), 0, "Should return array of categories")
}

func TestToolsRegistryBridge_ExecuteMethod_ListToolsByPermission(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listToolsByPermission
	args := []engine.ScriptValue{sv("file:read")}
	result, err := bridge.ExecuteMethod(ctx, "listToolsByPermission", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listToolsByPermission")

	// Should return array (may be empty)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools with permission")
}

func TestToolsRegistryBridge_ExecuteMethod_ListToolsByResourceUsage(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listToolsByResourceUsage
	criteria := map[string]interface{}{
		"maxMemory":          "low",
		"requiresNetwork":    false,
		"requiresFileSystem": true,
	}
	args := []engine.ScriptValue{svMap(criteria)}
	result, err := bridge.ExecuteMethod(ctx, "listToolsByResourceUsage", args)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listToolsByResourceUsage")

	// Should return array (may be empty)
	tools := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(tools), 0, "Should return array of tools matching criteria")
}

func TestToolsRegistryBridge_ExecuteMethod_GetToolDocumentation(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getToolDocumentation with non-existent tool
	args := []engine.ScriptValue{sv("non-existent-tool")}
	result, err := bridge.ExecuteMethod(ctx, "getToolDocumentation", args)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for non-existent tool documentation")
	assert.Contains(t, errorValue.Error().Error(), "failed to get tool documentation")
}

func TestToolsRegistryBridge_ExecuteMethod_RegisterTool(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test registerTool (should return error as not implemented)
	args := []engine.ScriptValue{
		sv("test-tool"),
		svMap(map[string]interface{}{}),
		svMap(map[string]interface{}{}),
	}
	result, err := bridge.ExecuteMethod(ctx, "registerTool", args)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for registerTool (not implemented)")
	assert.Contains(t, errorValue.Error().Error(), "not yet implemented")
}

func TestToolsRegistryBridge_ExecuteMethod_ExportToolToMCP(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test exportToolToMCP with non-existent tool
	args := []engine.ScriptValue{sv("non-existent-tool")}
	result, err := bridge.ExecuteMethod(ctx, "exportToolToMCP", args)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for non-existent tool export")
	assert.Contains(t, errorValue.Error().Error(), "failed to export tool to MCP")
}

func TestToolsRegistryBridge_ExecuteMethod_ExportAllToolsToMCP(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test exportAllToolsToMCP
	result, err := bridge.ExecuteMethod(ctx, "exportAllToolsToMCP", []engine.ScriptValue{})
	assert.NoError(t, err)

	// Could be either ObjectValue (success) or ErrorValue (if registry fails)
	switch v := result.(type) {
	case engine.ObjectValue:
		catalog := v.ToGo().(map[string]interface{})
		assert.Contains(t, catalog, "tools")
		assert.Contains(t, catalog, "version")
	case engine.ErrorValue:
		assert.Contains(t, v.Error(), "failed to export tools to MCP catalog")
	default:
		t.Fatalf("Expected ObjectValue or ErrorValue, got %T", result)
	}
}

func TestToolsRegistryBridge_ExecuteMethod_ClearRegistry(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test clearRegistry
	result, err := bridge.ExecuteMethod(ctx, "clearRegistry", []engine.ScriptValue{})
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from clearRegistry")
}

func TestToolsRegistryBridge_ExecuteMethod_GetRegistryStats(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getRegistryStats
	result, err := bridge.ExecuteMethod(ctx, "getRegistryStats", []engine.ScriptValue{})
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getRegistryStats")

	stats := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, stats, "total_tools")
	assert.Contains(t, stats, "total_categories")
	assert.Contains(t, stats, "categories")
	assert.Contains(t, stats, "tools_by_category")
	assert.Contains(t, stats, "deprecated_tools")
	assert.Contains(t, stats, "experimental_tools")
}

func TestToolsRegistryBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestToolsRegistryBridge_RequiredPermissions(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasStoragePermission := false
	hasMemoryPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionStorage {
			hasStoragePermission = true
		}
		if perm.Type == engine.PermissionMemory {
			hasMemoryPermission = true
		}
	}
	assert.True(t, hasStoragePermission, "Should have storage permission")
	assert.True(t, hasMemoryPermission, "Should have memory permission")
}

func TestToolsRegistryBridge_TypeMappings(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"tool_registry_entry", "tool_metadata", "tool_documentation"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestToolsRegistryBridge_Cleanup(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestToolsRegistryBridge_NotInitialized(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listTools", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
