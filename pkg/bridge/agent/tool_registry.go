// ABOUTME: Built-in tools registry bridge for go-llms tool system discovery and management
// ABOUTME: Bridges tool registry, discovery, versioning, and MCP export functionality

package agent

import (
	"context"
	"fmt"
	"sync"

	// go-llms imports for tool functionality
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ToolsRegistryBridge provides script access to go-llms built-in tools registry
type ToolsRegistryBridge struct {
	initialized bool
	registry    tools.ToolRegistry
	mu          sync.RWMutex
}

// NewToolsRegistryBridge creates a new tools registry bridge
func NewToolsRegistryBridge() *ToolsRegistryBridge {
	return &ToolsRegistryBridge{
		registry: tools.Tools, // Use global registry
	}
}

// GetID returns the bridge identifier
func (tb *ToolsRegistryBridge) GetID() string {
	return "tools_registry"
}

// GetMetadata returns bridge metadata
func (tb *ToolsRegistryBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "tools_registry",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms built-in tools registry with discovery, versioning, and MCP export",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"},
	}
}

// Initialize sets up the tools registry bridge
func (tb *ToolsRegistryBridge) Initialize(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (tb *ToolsRegistryBridge) Cleanup(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.initialized = false
	return nil
}

// IsInitialized returns initialization status
func (tb *ToolsRegistryBridge) IsInitialized() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (tb *ToolsRegistryBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(tb)
}

// Methods returns available bridge methods
func (tb *ToolsRegistryBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Tool discovery and listing
		{
			Name:        "listTools",
			Description: "List all registered tools",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listTools()"},
		},
		{
			Name:        "getTool",
			Description: "Get tool by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getTool('calculator')"},
		},
		{
			Name:        "searchTools",
			Description: "Search tools by query string",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			ReturnType: "array",
			Examples:   []string{"searchTools('math')"},
		},
		{
			Name:        "listToolsByCategory",
			Description: "List tools in specific category",
			Parameters: []engine.ParameterInfo{
				{Name: "category", Type: "string", Required: true, Description: "Tool category"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByCategory('math')"},
		},
		{
			Name:        "listToolsByTags",
			Description: "List tools matching all provided tags",
			Parameters: []engine.ParameterInfo{
				{Name: "tags", Type: "array", Required: true, Description: "Array of tags"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByTags(['network', 'api'])"},
		},
		{
			Name:        "getToolCategories",
			Description: "Get all available tool categories",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"getToolCategories()"},
		},
		// Tool filtering by permissions and resources
		{
			Name:        "listToolsByPermission",
			Description: "List tools requiring specific permission",
			Parameters: []engine.ParameterInfo{
				{Name: "permission", Type: "string", Required: true, Description: "Required permission"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByPermission('file:read')"},
		},
		{
			Name:        "listToolsByResourceUsage",
			Description: "List tools matching resource criteria",
			Parameters: []engine.ParameterInfo{
				{Name: "criteria", Type: "object", Required: true, Description: "Resource criteria object"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByResourceUsage({maxMemory: 'low', requiresNetwork: false})"},
		},
		// Tool documentation
		{
			Name:        "getToolDocumentation",
			Description: "Get comprehensive documentation for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getToolDocumentation('calculator')"},
		},
		// Tool registration
		{
			Name:        "registerTool",
			Description: "Register a new tool in the registry",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
				{Name: "tool", Type: "object", Required: true, Description: "Tool implementation"},
				{Name: "metadata", Type: "object", Required: true, Description: "Tool metadata"},
			},
			ReturnType: "void",
			Examples:   []string{"registerTool('my-tool', toolImpl, metadata)"},
		},
		// MCP export functionality
		{
			Name:        "exportToolToMCP",
			Description: "Export single tool to MCP format",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"exportToolToMCP('calculator')"},
		},
		{
			Name:        "exportAllToolsToMCP",
			Description: "Export all tools to MCP catalog",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"exportAllToolsToMCP()"},
		},
		// Registry management
		{
			Name:        "clearRegistry",
			Description: "Clear all tools from registry (testing only)",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
			Examples:    []string{"clearRegistry()"},
		},
		{
			Name:        "getRegistryStats",
			Description: "Get registry statistics and metrics",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getRegistryStats()"},
		},
	}
}

// ValidateMethod validates method calls
func (tb *ToolsRegistryBridge) ValidateMethod(name string, args []interface{}) error {
	if !tb.IsInitialized() {
		return fmt.Errorf("tools registry bridge not initialized")
	}

	methods := tb.Methods()
	for _, method := range methods {
		if method.Name == name {
			requiredCount := 0
			for _, param := range method.Parameters {
				if param.Required {
					requiredCount++
				}
			}
			if len(args) < requiredCount {
				return fmt.Errorf("method %s requires at least %d arguments, got %d", name, requiredCount, len(args))
			}
			return nil
		}
	}
	return fmt.Errorf("unknown method: %s", name)
}

// TypeMappings returns type conversion mappings
func (tb *ToolsRegistryBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"tool_registry_entry": {
			GoType:     "tools.RegistryEntry",
			ScriptType: "object",
			Converter:  "toolRegistryEntryConverter",
			Metadata:   map[string]interface{}{"description": "Tool registry entry with metadata"},
		},
		"tool_metadata": {
			GoType:     "tools.ToolMetadata",
			ScriptType: "object",
			Converter:  "toolMetadataConverter",
			Metadata:   map[string]interface{}{"description": "Tool metadata and configuration"},
		},
		"tool_documentation": {
			GoType:     "tools.ToolDocumentation",
			ScriptType: "object",
			Converter:  "toolDocumentationConverter",
			Metadata:   map[string]interface{}{"description": "Comprehensive tool documentation"},
		},
		"mcp_tool_definition": {
			GoType:     "tools.MCPToolDefinition",
			ScriptType: "object",
			Converter:  "mcpToolDefinitionConverter",
			Metadata:   map[string]interface{}{"description": "MCP tool definition"},
		},
		"mcp_catalog": {
			GoType:     "tools.MCPCatalog",
			ScriptType: "object",
			Converter:  "mcpCatalogConverter",
			Metadata:   map[string]interface{}{"description": "MCP tools catalog"},
		},
		"resource_criteria": {
			GoType:     "tools.ResourceCriteria",
			ScriptType: "object",
			Converter:  "resourceCriteriaConverter",
			Metadata:   map[string]interface{}{"description": "Tool resource filtering criteria"},
		},
	}
}

// RequiredPermissions returns required permissions
func (tb *ToolsRegistryBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionStorage,
			Resource:    "tools.registry",
			Actions:     []string{"read", "write", "export"},
			Description: "Access to tools registry for discovery and management",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "registry.metadata",
			Actions:     []string{"read", "write"},
			Description: "Access to tool metadata and documentation",
		},
	}
}

// Bridge method implementations

// Tool discovery and listing

// listTools lists all registered tools
func (tb *ToolsRegistryBridge) listTools(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("listTools", args); err != nil {
		return nil, err
	}

	entries := tb.registry.List()
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":         entry.Metadata.Name,
			"description":  entry.Metadata.Description,
			"category":     entry.Metadata.Category,
			"tags":         entry.Metadata.Tags,
			"version":      entry.Metadata.Version,
			"deprecated":   entry.Metadata.Deprecated,
			"experimental": entry.Metadata.Experimental,
		})
	}

	return result, nil
}

// getTool gets a tool by name
func (tb *ToolsRegistryBridge) getTool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("getTool", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("tool name must be a string")
	}

	tool, found := tb.registry.Get(name)
	if !found {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return map[string]interface{}{
		"name":                  tool.Name(),
		"description":           tool.Description(),
		"category":              tool.Category(),
		"tags":                  tool.Tags(),
		"version":               tool.Version(),
		"usage_instructions":    tool.UsageInstructions(),
		"examples":              tool.Examples(),
		"constraints":           tool.Constraints(),
		"error_guidance":        tool.ErrorGuidance(),
		"is_deterministic":      tool.IsDeterministic(),
		"is_destructive":        tool.IsDestructive(),
		"requires_confirmation": tool.RequiresConfirmation(),
		"estimated_latency":     tool.EstimatedLatency(),
		"parameter_schema":      tool.ParameterSchema(),
		"output_schema":         tool.OutputSchema(),
	}, nil
}

// searchTools searches tools by query string
func (tb *ToolsRegistryBridge) searchTools(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("searchTools", args); err != nil {
		return nil, err
	}

	query, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("search query must be a string")
	}

	entries := tb.registry.Search(query)
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":        entry.Metadata.Name,
			"description": entry.Metadata.Description,
			"category":    entry.Metadata.Category,
			"tags":        entry.Metadata.Tags,
			"version":     entry.Metadata.Version,
		})
	}

	return result, nil
}

// listToolsByCategory lists tools in specific category
func (tb *ToolsRegistryBridge) listToolsByCategory(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("listToolsByCategory", args); err != nil {
		return nil, err
	}

	category, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("category must be a string")
	}

	entries := tb.registry.ListByCategory(category)
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":        entry.Metadata.Name,
			"description": entry.Metadata.Description,
			"category":    entry.Metadata.Category,
			"tags":        entry.Metadata.Tags,
			"version":     entry.Metadata.Version,
		})
	}

	return result, nil
}

// listToolsByTags lists tools matching all provided tags
func (tb *ToolsRegistryBridge) listToolsByTags(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("listToolsByTags", args); err != nil {
		return nil, err
	}

	tagsArray, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("tags must be an array")
	}

	// Convert interface{} slice to string slice
	tags := make([]string, 0, len(tagsArray))
	for _, tag := range tagsArray {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	entries := tb.registry.ListByTags(tags...)
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":        entry.Metadata.Name,
			"description": entry.Metadata.Description,
			"category":    entry.Metadata.Category,
			"tags":        entry.Metadata.Tags,
			"version":     entry.Metadata.Version,
		})
	}

	return result, nil
}

// getToolCategories gets all available tool categories
func (tb *ToolsRegistryBridge) getToolCategories(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("getToolCategories", args); err != nil {
		return nil, err
	}

	categories := tb.registry.Categories()
	return categories, nil
}

// Tool filtering by permissions and resources

// listToolsByPermission lists tools requiring specific permission
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) listToolsByPermission(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("listToolsByPermission", args); err != nil {
		return nil, err
	}

	permission, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("permission must be a string")
	}

	entries := tb.registry.ListByPermission(permission)
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":        entry.Metadata.Name,
			"description": entry.Metadata.Description,
			"category":    entry.Metadata.Category,
			"tags":        entry.Metadata.Tags,
			"version":     entry.Metadata.Version,
		})
	}

	return result, nil
}

// listToolsByResourceUsage lists tools matching resource criteria
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) listToolsByResourceUsage(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("listToolsByResourceUsage", args); err != nil {
		return nil, err
	}

	criteriaMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("criteria must be an object")
	}

	// Convert criteria map to ResourceCriteria struct
	criteria := tools.ResourceCriteria{}

	if maxMemory, ok := criteriaMap["maxMemory"].(string); ok {
		criteria.MaxMemory = maxMemory
	}

	if requiresNetwork, ok := criteriaMap["requiresNetwork"].(bool); ok {
		criteria.RequiresNetwork = &requiresNetwork
	}

	if requiresFileSystem, ok := criteriaMap["requiresFileSystem"].(bool); ok {
		criteria.RequiresFileSystem = &requiresFileSystem
	}

	if requiresConcurrent, ok := criteriaMap["requiresConcurrent"].(bool); ok {
		criteria.RequiresConcurrent = &requiresConcurrent
	}

	entries := tb.registry.ListByResourceUsage(criteria)
	result := make([]map[string]interface{}, 0, len(entries))

	for _, entry := range entries {
		result = append(result, map[string]interface{}{
			"name":        entry.Metadata.Name,
			"description": entry.Metadata.Description,
			"category":    entry.Metadata.Category,
			"tags":        entry.Metadata.Tags,
			"version":     entry.Metadata.Version,
		})
	}

	return result, nil
}

// Tool documentation

// getToolDocumentation gets comprehensive documentation for a tool
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) getToolDocumentation(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("getToolDocumentation", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("tool name must be a string")
	}

	doc, err := tb.registry.GetToolDocumentation(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool documentation: %w", err)
	}

	return map[string]interface{}{
		"name":                  doc.Name,
		"description":           doc.Description,
		"category":              doc.Category,
		"tags":                  doc.Tags,
		"version":               doc.Version,
		"usage_instructions":    doc.UsageInstructions,
		"examples":              doc.Examples,
		"constraints":           doc.Constraints,
		"error_guidance":        doc.ErrorGuidance,
		"required_permissions":  doc.RequiredPermissions,
		"resource_usage":        doc.ResourceUsage,
		"is_deterministic":      doc.IsDeterministic,
		"is_destructive":        doc.IsDestructive,
		"requires_confirmation": doc.RequiresConfirmation,
		"estimated_latency":     doc.EstimatedLatency,
		"parameter_schema":      doc.ParameterSchema,
		"output_schema":         doc.OutputSchema,
	}, nil
}

// Tool registration

// registerTool registers a new tool in the registry (simplified interface)
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) registerTool(ctx context.Context, args []interface{}) error {
	if err := tb.ValidateMethod("registerTool", args); err != nil {
		return err
	}

	name, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("tool name must be a string")
	}

	// For now, we'll return an error indicating this requires a proper tool implementation
	// In a real bridge, we'd need to convert the script tool object to a domain.Tool
	_ = name
	return fmt.Errorf("tool registration from scripts not yet implemented - tools must be registered in Go code")
}

// MCP export functionality

// exportToolToMCP exports single tool to MCP format
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) exportToolToMCP(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("exportToolToMCP", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("tool name must be a string")
	}

	mcp, err := tb.registry.ExportToMCP(name)
	if err != nil {
		return nil, fmt.Errorf("failed to export tool to MCP: %w", err)
	}

	return map[string]interface{}{
		"name":         mcp.Name,
		"description":  mcp.Description,
		"inputSchema":  mcp.InputSchema,
		"outputSchema": mcp.OutputSchema,
		"annotations":  mcp.Annotations,
	}, nil
}

// exportAllToolsToMCP exports all tools to MCP catalog
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) exportAllToolsToMCP(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("exportAllToolsToMCP", args); err != nil {
		return nil, err
	}

	catalog, err := tb.registry.ExportAllToMCP()
	if err != nil {
		return nil, fmt.Errorf("failed to export tools to MCP catalog: %w", err)
	}

	// Convert tools to script-friendly format
	toolsArray := make([]map[string]interface{}, 0, len(catalog.Tools))
	for _, tool := range catalog.Tools {
		toolsArray = append(toolsArray, map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"inputSchema":  tool.InputSchema,
			"outputSchema": tool.OutputSchema,
			"annotations":  tool.Annotations,
		})
	}

	return map[string]interface{}{
		"version":     catalog.Version,
		"description": catalog.Description,
		"tools":       toolsArray,
		"metadata":    catalog.Metadata,
	}, nil
}

// Registry management

// clearRegistry clears all tools from registry (testing only)
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) clearRegistry(ctx context.Context, args []interface{}) error {
	if err := tb.ValidateMethod("clearRegistry", args); err != nil {
		return err
	}

	tb.registry.Clear()
	return nil
}

// getRegistryStats gets registry statistics and metrics
//
//nolint:unused // Bridge method called via reflection
func (tb *ToolsRegistryBridge) getRegistryStats(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := tb.ValidateMethod("getRegistryStats", args); err != nil {
		return nil, err
	}

	allTools := tb.registry.List()
	categories := tb.registry.Categories()

	// Count tools by category
	categoryCount := make(map[string]int)
	var deprecatedCount, experimentalCount int

	for _, entry := range allTools {
		if entry.Metadata.Category != "" {
			categoryCount[entry.Metadata.Category]++
		}
		if entry.Metadata.Deprecated {
			deprecatedCount++
		}
		if entry.Metadata.Experimental {
			experimentalCount++
		}
	}

	return map[string]interface{}{
		"total_tools":        len(allTools),
		"total_categories":   len(categories),
		"categories":         categories,
		"tools_by_category":  categoryCount,
		"deprecated_tools":   deprecatedCount,
		"experimental_tools": experimentalCount,
	}, nil
}
