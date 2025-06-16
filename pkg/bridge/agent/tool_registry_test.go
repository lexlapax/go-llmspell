// ABOUTME: Tests for tools registry bridge functionality including discovery, filtering, and MCP export
// ABOUTME: Comprehensive test coverage for tool registry operations, metadata, and versioning

package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// go-llms imports for tools functionality
	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Test ToolsRegistryBridge core functionality
func TestToolsRegistryBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *ToolsRegistryBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *ToolsRegistryBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "tools_registry", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "tools registry")
			},
		},
		{
			name: "List tools initially",
			test: func(t *testing.T, bridge *ToolsRegistryBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.listTools(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				tools, ok := result.([]map[string]interface{})
				require.True(t, ok)
				// Note: The global registry might have tools, so we just check it's an array
				assert.IsType(t, []map[string]interface{}{}, tools)
			},
		},
		{
			name: "Get tool categories",
			test: func(t *testing.T, bridge *ToolsRegistryBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.getToolCategories(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				categories, ok := result.([]string)
				require.True(t, ok)
				assert.IsType(t, []string{}, categories)
			},
		},
		{
			name: "Get registry stats",
			test: func(t *testing.T, bridge *ToolsRegistryBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.getRegistryStats(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				stats, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, stats, "total_tools")
				assert.Contains(t, stats, "total_categories")
				assert.Contains(t, stats, "categories")
				assert.Contains(t, stats, "tools_by_category")
			},
		},
		{
			name: "Search tools with empty query",
			test: func(t *testing.T, bridge *ToolsRegistryBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.searchTools(ctx, []interface{}{""})
				require.NoError(t, err)
				assert.NotNil(t, result)

				tools, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.IsType(t, []map[string]interface{}{}, tools)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewToolsRegistryBridge()
			tt.test(t, bridge)
		})
	}
}

// MockTool implements domain.Tool for testing
type MockTool struct {
	name                 string
	description          string
	category             string
	tags                 []string
	version              string
	usageInstructions    string
	examples             []domain.ToolExample
	constraints          []string
	errorGuidance        map[string]string
	isDeterministic      bool
	isDestructive        bool
	requiresConfirmation bool
	estimatedLatency     string
	parameterSchema      *schemaDomain.Schema
	outputSchema         *schemaDomain.Schema
}

func NewMockTool(name string) *MockTool {
	return &MockTool{
		name:              name,
		description:       "Mock tool for testing",
		category:          "test",
		tags:              []string{"mock", "test"},
		version:           "1.0.0",
		usageInstructions: "Use this tool for testing purposes",
		examples: []domain.ToolExample{
			{
				Name:        "Basic usage",
				Description: "Shows basic usage",
				Scenario:    "When testing",
				Input:       map[string]interface{}{"test": "value"},
				Output:      map[string]interface{}{"result": "success"},
				Explanation: "This demonstrates the tool",
			},
		},
		constraints:          []string{"Testing only"},
		errorGuidance:        map[string]string{"test_error": "This is a test error"},
		isDeterministic:      true,
		isDestructive:        false,
		requiresConfirmation: false,
		estimatedLatency:     "fast",
		parameterSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"test": {Type: "string", Description: "Test parameter"},
			},
		},
		outputSchema: &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"result": {Type: "string", Description: "Test result"},
			},
		},
	}
}

func (m *MockTool) Name() string        { return m.name }
func (m *MockTool) Description() string { return m.description }
func (m *MockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	return map[string]interface{}{"result": "success"}, nil
}
func (m *MockTool) ParameterSchema() *schemaDomain.Schema { return m.parameterSchema }
func (m *MockTool) OutputSchema() *schemaDomain.Schema    { return m.outputSchema }
func (m *MockTool) UsageInstructions() string             { return m.usageInstructions }
func (m *MockTool) Examples() []domain.ToolExample        { return m.examples }
func (m *MockTool) Constraints() []string                 { return m.constraints }
func (m *MockTool) ErrorGuidance() map[string]string      { return m.errorGuidance }
func (m *MockTool) Category() string                      { return m.category }
func (m *MockTool) Tags() []string                        { return m.tags }
func (m *MockTool) Version() string                       { return m.version }
func (m *MockTool) IsDeterministic() bool                 { return m.isDeterministic }
func (m *MockTool) IsDestructive() bool                   { return m.isDestructive }
func (m *MockTool) RequiresConfirmation() bool            { return m.requiresConfirmation }
func (m *MockTool) EstimatedLatency() string              { return m.estimatedLatency }
func (m *MockTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:         m.name,
		Description:  m.description,
		InputSchema:  m.parameterSchema,
		OutputSchema: m.outputSchema,
		Annotations: map[string]interface{}{
			"category": m.category,
			"version":  m.version,
		},
	}
}

// Test tools registry operations with mock tools
func TestToolsRegistryOperations(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a test registry for isolated testing
	testRegistry := tools.NewTestRegistry()
	bridge.registry = testRegistry

	// Register mock tools
	mockTool1 := NewMockTool("test-tool-1")
	mockTool1.category = "math"
	mockTool1.tags = []string{"calculation", "test"}

	mockTool2 := NewMockTool("test-tool-2")
	mockTool2.category = "file"
	mockTool2.tags = []string{"filesystem", "test"}

	metadata1 := tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "test-tool-1",
			Category:    "math",
			Tags:        []string{"calculation", "test"},
			Description: "Test math tool",
			Version:     "1.0.0",
		},
		RequiredPermissions: []string{"math:calculate"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	}

	metadata2 := tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "test-tool-2",
			Category:    "file",
			Tags:        []string{"filesystem", "test"},
			Description: "Test file tool",
			Version:     "1.0.0",
		},
		RequiredPermissions: []string{"file:read", "file:write"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     false,
			FileSystem:  true,
			Concurrency: false,
		},
	}

	err = testRegistry.RegisterTool("test-tool-1", mockTool1, metadata1)
	require.NoError(t, err)
	err = testRegistry.RegisterTool("test-tool-2", mockTool2, metadata2)
	require.NoError(t, err)

	// Test listing tools
	t.Run("List all tools", func(t *testing.T) {
		result, err := bridge.listTools(ctx, []interface{}{})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 2, len(tools))

		// Check tool names are present
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool["name"].(string)] = true
		}
		assert.True(t, toolNames["test-tool-1"])
		assert.True(t, toolNames["test-tool-2"])
	})

	// Test getting specific tool
	t.Run("Get specific tool", func(t *testing.T) {
		result, err := bridge.getTool(ctx, []interface{}{"test-tool-1"})
		require.NoError(t, err)

		tool, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test-tool-1", tool["name"])
		assert.Equal(t, "Mock tool for testing", tool["description"])
		assert.Equal(t, "math", tool["category"])
		assert.True(t, tool["is_deterministic"].(bool))
		assert.False(t, tool["is_destructive"].(bool))
	})

	// Test listing by category
	t.Run("List tools by category", func(t *testing.T) {
		result, err := bridge.listToolsByCategory(ctx, []interface{}{"math"})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(tools))
		assert.Equal(t, "test-tool-1", tools[0]["name"])
	})

	// Test listing by tags
	t.Run("List tools by tags", func(t *testing.T) {
		result, err := bridge.listToolsByTags(ctx, []interface{}{[]interface{}{"test"}})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 2, len(tools)) // Both tools have "test" tag
	})

	// Test searching tools
	t.Run("Search tools", func(t *testing.T) {
		result, err := bridge.searchTools(ctx, []interface{}{"math"})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(tools))
		assert.Equal(t, "test-tool-1", tools[0]["name"])
	})

	// Test getting categories
	t.Run("Get tool categories", func(t *testing.T) {
		result, err := bridge.getToolCategories(ctx, []interface{}{})
		require.NoError(t, err)

		categories, ok := result.([]string)
		require.True(t, ok)
		assert.Contains(t, categories, "math")
		assert.Contains(t, categories, "file")
	})
}

// Test permission and resource filtering
func TestToolsFiltering(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a test registry for isolated testing
	testRegistry := tools.NewTestRegistry()
	bridge.registry = testRegistry

	// Register mock tool with specific permissions and resources
	mockTool := NewMockTool("filtered-tool")
	metadata := tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "filtered-tool",
			Category:    "system",
			Tags:        []string{"network", "api"},
			Description: "Tool requiring network access",
			Version:     "1.0.0",
		},
		RequiredPermissions: []string{"network:access", "api:call"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "high",
			Network:     true,
			FileSystem:  false,
			Concurrency: true,
		},
	}

	err = testRegistry.RegisterTool("filtered-tool", mockTool, metadata)
	require.NoError(t, err)

	// Test filtering by permission
	t.Run("Filter by permission", func(t *testing.T) {
		result, err := bridge.listToolsByPermission(ctx, []interface{}{"network:access"})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(tools))
		assert.Equal(t, "filtered-tool", tools[0]["name"])
	})

	// Test filtering by resource usage
	t.Run("Filter by resource usage", func(t *testing.T) {
		criteria := map[string]interface{}{
			"requiresNetwork": true,
			"maxMemory":       "high",
		}

		result, err := bridge.listToolsByResourceUsage(ctx, []interface{}{criteria})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(tools))
		assert.Equal(t, "filtered-tool", tools[0]["name"])
	})

	// Test filtering with restrictive criteria
	t.Run("Filter with restrictive criteria", func(t *testing.T) {
		criteria := map[string]interface{}{
			"requiresNetwork": false, // Tool requires network, so should be filtered out
		}

		result, err := bridge.listToolsByResourceUsage(ctx, []interface{}{criteria})
		require.NoError(t, err)

		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 0, len(tools))
	})
}

// Test tool documentation and MCP export
func TestToolDocumentationAndMCP(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a test registry for isolated testing
	testRegistry := tools.NewTestRegistry()
	bridge.registry = testRegistry

	// Register mock tool
	mockTool := NewMockTool("documented-tool")
	mockTool.category = "utility"
	mockTool.tags = []string{"documentation", "example"}
	mockTool.version = "2.0.0"
	metadata := tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "documented-tool",
			Category:    "utility",
			Tags:        []string{"documentation", "example"},
			Description: "Well-documented test tool",
			Version:     "2.0.0",
		},
		RequiredPermissions: []string{"util:access"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	}

	err = testRegistry.RegisterTool("documented-tool", mockTool, metadata)
	require.NoError(t, err)

	// Test getting tool documentation
	t.Run("Get tool documentation", func(t *testing.T) {
		result, err := bridge.getToolDocumentation(ctx, []interface{}{"documented-tool"})
		require.NoError(t, err)

		doc, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "documented-tool", doc["name"])
		assert.Equal(t, "Mock tool for testing", doc["description"])
		assert.Equal(t, "utility", doc["category"])
		assert.Contains(t, doc["tags"], "documentation")
		assert.Equal(t, "2.0.0", doc["version"])
		assert.Equal(t, "Use this tool for testing purposes", doc["usage_instructions"])
		assert.NotNil(t, doc["examples"])
		assert.NotNil(t, doc["parameter_schema"])
		assert.NotNil(t, doc["output_schema"])
	})

	// Test MCP export for single tool
	t.Run("Export tool to MCP", func(t *testing.T) {
		result, err := bridge.exportToolToMCP(ctx, []interface{}{"documented-tool"})
		require.NoError(t, err)

		mcp, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "documented-tool", mcp["name"])
		assert.Equal(t, "Mock tool for testing", mcp["description"])
		assert.NotNil(t, mcp["inputSchema"])
		assert.NotNil(t, mcp["outputSchema"])
		assert.NotNil(t, mcp["annotations"])
	})

	// Test MCP export for all tools
	t.Run("Export all tools to MCP", func(t *testing.T) {
		result, err := bridge.exportAllToolsToMCP(ctx, []interface{}{})
		require.NoError(t, err)

		catalog, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "1.0.0", catalog["version"])
		assert.Equal(t, "Go-LLMs Tool Catalog", catalog["description"])

		toolsArray, ok := catalog["tools"].([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(toolsArray))
		assert.Equal(t, "documented-tool", toolsArray[0]["name"])

		metadata, ok := catalog["metadata"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, metadata["tool_count"])
	})
}

// Test error scenarios
func TestToolsRegistryBridgeErrors(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()

	// Test methods without initialization
	_, err := bridge.listTools(ctx, []interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create test registry
	testRegistry := tools.NewTestRegistry()
	bridge.registry = testRegistry

	// Test invalid parameters
	_, err = bridge.getTool(ctx, []interface{}{123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	_, err = bridge.listToolsByCategory(ctx, []interface{}{123})
	assert.Error(t, err)

	_, err = bridge.listToolsByTags(ctx, []interface{}{123})
	assert.Error(t, err)

	_, err = bridge.searchTools(ctx, []interface{}{123})
	assert.Error(t, err)

	// Test getting nonexistent tool
	_, err = bridge.getTool(ctx, []interface{}{"nonexistent-tool"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool not found")

	// Test getting documentation for nonexistent tool
	_, err = bridge.getToolDocumentation(ctx, []interface{}{"nonexistent-tool"})
	assert.Error(t, err)

	// Test MCP export for nonexistent tool
	_, err = bridge.exportToolToMCP(ctx, []interface{}{"nonexistent-tool"})
	assert.Error(t, err)

	// Test tool registration (should fail as not implemented)
	err = bridge.registerTool(ctx, []interface{}{"test-tool", map[string]interface{}{}, map[string]interface{}{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

// Test bridge lifecycle
func TestToolsRegistryBridgeLifecycle(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "tools_registry", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "tool_registry_entry")
	assert.Contains(t, typeMappings, "tool_metadata")
	assert.Contains(t, typeMappings, "mcp_tool_definition")
	assert.Contains(t, typeMappings, "mcp_catalog")
	assert.Contains(t, typeMappings, "resource_criteria")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 10)

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"listTools",
		"getTool",
		"searchTools",
		"getToolDocumentation",
		"exportToolToMCP",
		"exportAllToolsToMCP",
		"getRegistryStats",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// Test registry management operations
func TestRegistryManagement(t *testing.T) {
	bridge := NewToolsRegistryBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create test registry
	testRegistry := tools.NewTestRegistry()
	bridge.registry = testRegistry

	// Register a test tool
	mockTool := NewMockTool("management-test-tool")
	metadata := tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "management-test-tool",
			Category:    "test",
			Description: "Tool for testing registry management",
			Version:     "1.0.0",
		},
	}
	err = testRegistry.RegisterTool("management-test-tool", mockTool, metadata)
	require.NoError(t, err)

	// Test clearing registry
	t.Run("Clear registry", func(t *testing.T) {
		// Verify tool exists
		result, err := bridge.listTools(ctx, []interface{}{})
		require.NoError(t, err)
		tools, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 1, len(tools))

		// Clear registry
		err = bridge.clearRegistry(ctx, []interface{}{})
		require.NoError(t, err)

		// Verify registry is empty
		result, err = bridge.listTools(ctx, []interface{}{})
		require.NoError(t, err)
		tools, ok = result.([]map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 0, len(tools))
	})
}
