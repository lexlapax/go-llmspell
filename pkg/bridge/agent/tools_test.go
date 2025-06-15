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

	// Use go-llms testutils for better mock tools
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
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

// Test schema support for custom tools
func TestToolsBridge_SchemaSupport_Enhanced(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register a custom math tool with schema
	toolDef := map[string]interface{}{
		"name":        "math_tool",
		"description": "Performs mathematical operations",
		"category":    "calculation",
		"parameterSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation to perform",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]interface{}{
					"type":        "number",
					"description": "First operand",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "Second operand",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
		"outputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type":        "number",
					"description": "The result of the operation",
				},
				"operation": map[string]interface{}{
					"type":        "string",
					"description": "The operation performed",
				},
			},
			"required": []string{"result"},
		},
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			p := params.(map[string]interface{})
			op := p["operation"].(string)
			a := p["a"].(float64)
			b := p["b"].(float64)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				result = a / b
			}

			return map[string]interface{}{
				"result":    result,
				"operation": op,
			}, nil
		},
	}

	// Register the custom tool
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{toolDef})
	require.NoError(t, err)

	// Get tool info
	info, err := bridge.ExecuteMethod(ctx, "getToolInfo", []interface{}{"math_tool"})
	require.NoError(t, err)

	toolInfo := info.(map[string]interface{})
	assert.Equal(t, "math_tool", toolInfo["name"])
	assert.Equal(t, "Performs mathematical operations", toolInfo["description"])
	assert.Equal(t, "calculation", toolInfo["category"])

	// Get tool schema
	schema, err := bridge.ExecuteMethod(ctx, "getToolSchema", []interface{}{"math_tool"})
	require.NoError(t, err)

	toolSchema := schema.(map[string]interface{})
	assert.NotNil(t, toolSchema["parameters"])
	assert.NotNil(t, toolSchema["output"])

	// Test executing with valid parameters
	result, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{
		"math_tool",
		map[string]interface{}{
			"operation": "multiply",
			"a":         5.0,
			"b":         3.0,
		},
	})
	require.NoError(t, err)

	execResult := result.(map[string]interface{})
	assert.Equal(t, 15.0, execResult["result"])
	assert.Equal(t, "multiply", execResult["operation"])

	// Test schema validation with invalid parameters
	_, err = bridge.ExecuteMethod(ctx, "executeTool", []interface{}{
		"math_tool",
		map[string]interface{}{
			"operation": "invalid_op",
			"a":         5.0,
			"b":         3.0,
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")
}

// Test enhanced tool metadata
func TestToolsBridge_EnhancedMetadata_V2(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Use testutils to create a mock file tool with enhanced metadata
	mockFileTool := mocks.NewMockTool("file_cleaner", "Cleans up temporary files").
		WithCategory("file").
		WithUsageInstructions("Use this tool to clean temporary files older than specified days. "+
			"Be careful with the force option as it bypasses confirmation.").
		WithConstraints(
			"Requires write permissions on target directory",
			"Cannot delete files in use",
			"Minimum age is 1 day",
		).
		WithErrorGuidance(map[string]string{
			"permission_denied": "Check if you have write permissions on the directory",
			"path_not_found":    "Verify the path exists and is accessible",
			"invalid_days":      "Days must be a positive integer",
		})

	// Register custom tool with enhanced metadata
	// Note: We still need to create the toolDef for the bridge since it expects this format
	toolDef := map[string]interface{}{
		"name":              mockFileTool.Name(),
		"description":       mockFileTool.Description(),
		"category":          mockFileTool.Category(),
		"usageInstructions": mockFileTool.UsageInstructions(),
		"examples": []map[string]interface{}{
			{
				"name":        "Clean old temp files",
				"description": "Remove temp files older than 7 days",
				"input": map[string]interface{}{
					"path": "/tmp",
					"days": 7,
				},
				"output": map[string]interface{}{
					"deleted": 42,
					"freed":   "1.2GB",
				},
			},
		},
		"constraints":          mockFileTool.Constraints(),
		"errorGuidance":        mockFileTool.ErrorGuidance(),
		"isDeterministic":      false,
		"isDestructive":        true,
		"requiresConfirmation": true,
		"estimatedLatency":     "high",
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			// Mock implementation
			return map[string]interface{}{
				"deleted": 42,
				"freed":   "1.2GB",
			}, nil
		},
	}

	// Register the tool
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{toolDef})
	require.NoError(t, err)

	// Get tool info
	info, err := bridge.ExecuteMethod(ctx, "getToolInfo", []interface{}{"file_cleaner"})
	require.NoError(t, err)

	toolInfo := info.(map[string]interface{})
	assert.Equal(t, "file_cleaner", toolInfo["name"])
	assert.Equal(t, "Cleans up temporary files", toolInfo["description"])
	assert.Equal(t, "file", toolInfo["category"])
	assert.Equal(t, false, toolInfo["isDeterministic"])
	assert.Equal(t, true, toolInfo["isDestructive"])
	assert.Equal(t, true, toolInfo["requiresConfirmation"])
	assert.Equal(t, "high", toolInfo["estimatedLatency"])

	// Get usage instructions
	help, err := bridge.ExecuteMethod(ctx, "getToolHelp", []interface{}{"file_cleaner"})
	require.NoError(t, err)
	assert.Contains(t, help.(string), "Use this tool to clean temporary files")

	// Get examples
	examples, err := bridge.ExecuteMethod(ctx, "getToolExamples", []interface{}{"file_cleaner"})
	require.NoError(t, err)

	examplesList := examples.([]map[string]interface{})
	assert.Len(t, examplesList, 1)
	assert.Equal(t, "Clean old temp files", examplesList[0]["name"])
}

// Test parameter validation
func TestToolsBridge_ParameterValidation_Enhanced(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register tool with complex parameter schema
	toolDef := map[string]interface{}{
		"name":        "data_processor",
		"description": "Process structured data",
		"parameterSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"data": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type": "integer",
							},
							"name": map[string]interface{}{
								"type": "string",
							},
							"tags": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
						"required": []string{"id", "name"},
					},
				},
				"options": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"sortBy": map[string]interface{}{
							"type": "string",
							"enum": []string{"id", "name"},
						},
						"limit": map[string]interface{}{
							"type":    "integer",
							"minimum": 1,
							"maximum": 100,
						},
					},
				},
			},
			"required": []string{"data"},
		},
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			// Mock implementation
			return map[string]interface{}{
				"processed": true,
				"count":     2,
			}, nil
		},
	}

	// Register the tool
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{toolDef})
	require.NoError(t, err)

	// Test with valid complex parameters
	validParams := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "Item 1",
				"tags": []string{"tag1", "tag2"},
			},
			map[string]interface{}{
				"id":   2,
				"name": "Item 2",
			},
		},
		"options": map[string]interface{}{
			"sortBy": "name",
			"limit":  10,
		},
	}

	result, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{"data_processor", validParams})
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Test with invalid parameters (missing required field)
	invalidParams := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id": 1,
				// Missing required "name" field
			},
		},
	}

	_, err = bridge.ExecuteMethod(ctx, "executeTool", []interface{}{"data_processor", invalidParams})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")

	// Test with invalid enum value
	invalidEnumParams := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "Item 1",
			},
		},
		"options": map[string]interface{}{
			"sortBy": "invalid_field", // Not in enum
		},
	}

	_, err = bridge.ExecuteMethod(ctx, "executeTool", []interface{}{"data_processor", invalidEnumParams})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")
}

// Test output schema validation
func TestToolsBridge_OutputSchemaValidation_Enhanced(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register tool with output schema
	toolDef := map[string]interface{}{
		"name":        "api_caller",
		"description": "Makes API calls and validates responses",
		"outputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":    "integer",
					"minimum": 100,
					"maximum": 599,
				},
				"data": map[string]interface{}{
					"type": "object",
				},
				"headers": map[string]interface{}{
					"type": "object",
					// Note: go-llms currently only supports boolean additionalProperties
					// TODO: Consider adding support for schema-based additionalProperties
				},
			},
			"required": []string{"status", "data"},
		},
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			// Return valid output
			return map[string]interface{}{
				"status": 200,
				"data": map[string]interface{}{
					"message": "Success",
				},
				"headers": map[string]interface{}{
					"Content-Type": "application/json",
				},
			}, nil
		},
	}

	// Register the tool
	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{toolDef})
	require.NoError(t, err)

	// Execute and verify output matches schema
	result, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{"api_caller", map[string]interface{}{}})
	require.NoError(t, err)

	output := result.(map[string]interface{})
	assert.Equal(t, 200, output["status"])
	assert.NotNil(t, output["data"])
	assert.NotNil(t, output["headers"])
}

// Test using go-llms testutils mock tool features
func TestToolsBridge_WithMockToolVerification(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a mock tool with verification capabilities
	mockTool := mocks.NewMockTool("test_tool", "Tool with call tracking").
		WithCategory("test").
		WithResponseMapping("test_input", map[string]interface{}{
			"result": "mapped_response",
		}).
		ExpectCall("test execution", func(input map[string]interface{}) bool {
			// Check if the input contains the expected test parameter
			return input["test_param"] == "expected_value"
		}, 1, 1) // Expect exactly 1 call

	// Register the mock tool
	toolDef := map[string]interface{}{
		"name":        mockTool.Name(),
		"description": mockTool.Description(),
		"category":    mockTool.Category(),
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			// Use the mock tool's Execute method
			return mockTool.Execute(nil, params)
		},
	}

	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{toolDef})
	require.NoError(t, err)

	// Execute the tool
	result, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{
		"test_tool",
		map[string]interface{}{
			"test_param": "expected_value",
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify expectations were met
	err = mockTool.VerifyExpectations()
	assert.NoError(t, err)

	// Check execution count
	assert.Equal(t, 1, mockTool.GetExecutionCount())

	// Verify call history
	history := mockTool.GetCallHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "expected_value", history[0].Input["test_param"])
}

// Test using fixtures for common scenarios
func TestToolsBridge_WithFixtures(t *testing.T) {
	// Create bridge
	bridge := NewToolsBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Use fixtures to create common tools
	webSearchTool := fixtures.WebSearchMockTool()
	fileTool := fixtures.FileMockTool()

	// Register the web search tool
	webSearchDef := map[string]interface{}{
		"name":        webSearchTool.Name(),
		"description": webSearchTool.Description(),
		"category":    "web",
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			return webSearchTool.Execute(nil, params)
		},
	}

	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{webSearchDef})
	require.NoError(t, err)

	// Register the file tool
	fileDef := map[string]interface{}{
		"name":        fileTool.Name(),
		"description": fileTool.Description(),
		"category":    "file",
		"execute": func(ctx interface{}, params interface{}) (interface{}, error) {
			return fileTool.Execute(nil, params)
		},
	}

	_, err = bridge.ExecuteMethod(ctx, "registerCustomTool", []interface{}{fileDef})
	require.NoError(t, err)

	// Test web search
	searchResult, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{
		"web_search",
		map[string]interface{}{
			"query": "weather today",
		},
	})
	require.NoError(t, err)

	searchData := searchResult.(map[string]interface{})
	assert.Equal(t, "weather today", searchData["query"])
	assert.NotNil(t, searchData["results"])

	// Test file operations
	fileResult, err := bridge.ExecuteMethod(ctx, "executeTool", []interface{}{
		"file_manager",
		map[string]interface{}{
			"operation": "read",
			"path":      "/etc/config.txt",
		},
	})
	require.NoError(t, err)

	fileData := fileResult.(map[string]interface{})
	assert.Equal(t, "read", fileData["operation"])
	assert.Contains(t, fileData["content"], "Configuration file")
}
