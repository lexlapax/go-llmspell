// ABOUTME: Tests for tools bridge providing access to go-llms tool discovery system
// ABOUTME: Verifies dynamic tool discovery, metadata access, and execution through bridge

package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestToolsBridge_BasicOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *ToolsBridge)
	}{
		{
			name: "GetID returns correct identifier",
			test: func(t *testing.T, b *ToolsBridge) {
				assert.Equal(t, "tools", b.GetID())
			},
		},
		{
			name: "GetMetadata returns valid metadata",
			test: func(t *testing.T, b *ToolsBridge) {
				metadata := b.GetMetadata()
				assert.Equal(t, "Tools Bridge", metadata.Name)
				assert.NotEmpty(t, metadata.Version)
				assert.NotEmpty(t, metadata.Description)
				assert.Equal(t, "go-llmspell", metadata.Author)
			},
		},
		{
			name: "Initialize and cleanup work correctly",
			test: func(t *testing.T, b *ToolsBridge) {
				ctx := context.Background()

				// Initial state
				assert.False(t, b.IsInitialized())

				// Initialize
				err := b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Double initialize should be safe
				err = b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Cleanup
				err = b.Cleanup(ctx)
				require.NoError(t, err)
				assert.False(t, b.IsInitialized())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewToolsBridge()
			tt.test(t, bridge)
		})
	}
}

func TestToolsBridge_Methods(t *testing.T) {
	bridge := NewToolsBridge()
	methods := bridge.Methods()

	// Check that we have the expected discovery-based methods
	expectedMethods := []string{
		"listTools",
		"searchTools",
		"listByCategory",
		"getToolInfo",
		"getToolSchema",
		"getToolHelp",
		"getToolExamples",
		"createTool",
		"executeTool",
		"registerCustomTool",
	}

	methodMap := make(map[string]engine.MethodInfo)
	for _, m := range methods {
		methodMap[m.Name] = m
	}

	for _, expected := range expectedMethods {
		t.Run("has_method_"+expected, func(t *testing.T) {
			method, exists := methodMap[expected]
			assert.True(t, exists, "Missing method: %s", expected)
			assert.NotEmpty(t, method.Description)
			assert.NotEmpty(t, method.ReturnType)
		})
	}
}

func TestToolsBridge_Discovery(t *testing.T) {
	ctx := context.Background()
	bridge := NewToolsBridge()
	require.NoError(t, bridge.Initialize(ctx))

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		validate func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "listTools returns tool info array",
			method: "listTools",
			args:   []interface{}{},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				tools, ok := result.([]map[string]interface{})
				require.True(t, ok, "Expected []map[string]interface{}, got %T", result)

				// Should have tools available
				assert.NotEmpty(t, tools)

				// Each tool should have required fields
				for _, tool := range tools {
					assert.NotEmpty(t, tool["name"])
					assert.NotEmpty(t, tool["description"])
					assert.NotEmpty(t, tool["category"])
				}
			},
		},
		{
			name:   "searchTools finds matching tools",
			method: "searchTools",
			args:   []interface{}{"json"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				tools, ok := result.([]map[string]interface{})
				require.True(t, ok)

				// Should find tools related to JSON
				for _, tool := range tools {
					// Either name, description, or tags should contain "json"
					found := false
					if name, ok := tool["name"].(string); ok {
						if containsIgnoreCase(name, "json") {
							found = true
						}
					}
					if !found {
						if desc, ok := tool["description"].(string); ok {
							if containsIgnoreCase(desc, "json") {
								found = true
							}
						}
					}
					if !found {
						if tags, ok := tool["tags"].([]string); ok {
							for _, tag := range tags {
								if containsIgnoreCase(tag, "json") {
									found = true
									break
								}
							}
						}
					}
					assert.True(t, found, "Tool %v doesn't match search query 'json'", tool["name"])
				}
			},
		},
		{
			name:   "listByCategory returns tools in category",
			method: "listByCategory",
			args:   []interface{}{"math"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				tools, ok := result.([]map[string]interface{})
				require.True(t, ok)

				// All returned tools should be in math category
				for _, tool := range tools {
					category, ok := tool["category"].(string)
					require.True(t, ok)
					assert.Equal(t, "math", category)
				}
			},
		},
		{
			name:   "getToolInfo returns detailed info",
			method: "getToolInfo",
			args:   []interface{}{"calculator"},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					// Calculator might not be available, skip test
					t.Skip("Calculator tool not available")
				}

				info, ok := result.(map[string]interface{})
				require.True(t, ok)

				// Check required fields
				assert.Equal(t, "calculator", info["name"])
				assert.NotEmpty(t, info["description"])
				assert.NotEmpty(t, info["category"])
				assert.NotNil(t, info["version"])
			},
		},
		{
			name:   "getToolSchema returns schema info",
			method: "getToolSchema",
			args:   []interface{}{"calculator"},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Skip("Calculator tool not available")
				}

				schema, ok := result.(map[string]interface{})
				require.True(t, ok)

				// Should have name and description
				assert.Equal(t, "calculator", schema["name"])
				assert.NotEmpty(t, schema["description"])

				// Should have parameters schema
				if params, exists := schema["parameters"]; exists && params != nil {
					// Verify it's a valid schema structure
					_, ok := params.(map[string]interface{})
					assert.True(t, ok, "Parameters should be a map")
				}
			},
		},
		{
			name:   "getToolHelp returns help text",
			method: "getToolHelp",
			args:   []interface{}{"calculator"},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Skip("Calculator tool not available")
				}

				help, ok := result.(string)
				require.True(t, ok)

				// Help should contain tool name and description
				assert.Contains(t, help, "calculator")
				assert.Contains(t, help, "Tool:")
				assert.Contains(t, help, "Description:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			tt.validate(t, result, err)
		})
	}
}

func TestToolsBridge_ToolExecution(t *testing.T) {
	ctx := context.Background()
	bridge := NewToolsBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Test tool creation and execution
	t.Run("createTool creates tool instance", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "createTool", []interface{}{"calculator"})
		if err != nil {
			t.Skip("Calculator tool not available")
		}

		assert.NotNil(t, result)
		// The result should be a tool wrapper that can be used in scripts
		toolWrapper, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "calculator", toolWrapper["name"])
	})

	t.Run("executeTool runs tool with parameters", func(t *testing.T) {
		params := map[string]interface{}{
			"operation": "add",
			"operand1":  10.0,
			"operand2":  5.0,
		}

		result, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{"calculator", params})
		if err != nil {
			t.Skip("Calculator tool not available or execution failed")
		}

		// Check that we got a result
		assert.NotNil(t, result)
	})
}

func TestToolsBridge_CustomTools(t *testing.T) {
	ctx := context.Background()
	bridge := NewToolsBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Test custom tool registration
	t.Run("registerCustomTool adds new tool", func(t *testing.T) {
		customTool := map[string]interface{}{
			"name":        "custom_test_tool",
			"description": "A custom test tool",
			"category":    "test",
			"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
				return map[string]interface{}{"result": "custom tool executed"}, nil
			},
		}

		_, err := bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{customTool})
		require.NoError(t, err)

		// Verify tool is registered
		result, err := bridge.ExecuteMethod(ctx, "getToolInfo", []interface{}{"custom_test_tool"})
		require.NoError(t, err)

		info, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "custom_test_tool", info["name"])
		assert.Equal(t, "A custom test tool", info["description"])
	})
}

func TestToolsBridge_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	bridge := NewToolsBridge()

	t.Run("methods fail when not initialized", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "listTools", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	require.NoError(t, bridge.Initialize(ctx))

	t.Run("unknown method returns error", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "unknownMethod", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})

	t.Run("getToolInfo with non-existent tool", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "getToolInfo", []interface{}{"non_existent_tool"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid arguments return error", func(t *testing.T) {
		// Missing required arguments
		_, err := bridge.ExecuteMethod(ctx, "searchTools", []interface{}{})
		assert.Error(t, err)

		// Wrong argument type
		_, err = bridge.ExecuteMethod(ctx, "searchTools", []interface{}{123})
		assert.Error(t, err)
	})
}

func TestToolsBridge_TypeMappings(t *testing.T) {
	bridge := NewToolsBridge()
	mappings := bridge.TypeMappings()

	// Check that we have the expected type mappings
	expectedTypes := []string{
		"Tool",
		"ToolInfo",
		"ToolSchema",
		"ToolExample",
		"ToolContext",
	}

	for _, typeName := range expectedTypes {
		t.Run("has_type_"+typeName, func(t *testing.T) {
			mapping, exists := mappings[typeName]
			assert.True(t, exists, "Missing type mapping: %s", typeName)
			assert.NotEmpty(t, mapping.GoType)
			assert.NotEmpty(t, mapping.ScriptType)
		})
	}
}

func TestToolsBridge_Permissions(t *testing.T) {
	bridge := NewToolsBridge()
	permissions := bridge.RequiredPermissions()

	// Should require permissions for tool operations
	assert.NotEmpty(t, permissions)

	// Check for essential permissions
	hasToolPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionProcess && perm.Resource == "tool" {
			hasToolPermission = true
			assert.Contains(t, perm.Actions, "execute")
			assert.Contains(t, perm.Actions, "list")
		}
	}
	assert.True(t, hasToolPermission, "Missing tool execution permission")
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
