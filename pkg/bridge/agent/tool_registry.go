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
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// ToolsRegistryBridge provides script access to go-llms built-in tools registry.
// It enables tool discovery, filtering, documentation retrieval, and MCP export
// functionality without reimplementing the underlying tool system.
type ToolsRegistryBridge struct {
	initialized bool
	registry    tools.ToolRegistry
	mu          sync.RWMutex
}

// NewToolsRegistryBridge creates a new tools registry bridge.
// It wraps the global go-llms tools registry for script access.
func NewToolsRegistryBridge() *ToolsRegistryBridge {
	return &ToolsRegistryBridge{
		registry: tools.Tools, // Use global registry
	}
}

// GetID returns the bridge identifier.
// Always returns "tools_registry" for this bridge.
func (tb *ToolsRegistryBridge) GetID() string {
	return "tools_registry"
}

// GetMetadata returns bridge metadata.
// Provides comprehensive information about the bridge including
// dependencies on go-llms tools package.
func (tb *ToolsRegistryBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:         "tools_registry",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms built-in tools registry with discovery, versioning, and MCP export",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"},
	}
}

// Initialize sets up the tools registry bridge.
// Currently performs minimal initialization as the registry
// is already initialized by go-llms.
func (tb *ToolsRegistryBridge) Initialize(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup.
// Marks the bridge as uninitialized. The underlying registry
// remains intact as it's managed by go-llms.
func (tb *ToolsRegistryBridge) Cleanup(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.initialized = false
	return nil
}

// IsInitialized returns initialization status.
// Thread-safe check of bridge initialization state.
func (tb *ToolsRegistryBridge) IsInitialized() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// Delegates to the engine's RegisterBridge method for integration.
func (tb *ToolsRegistryBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns available bridge methods.
// Provides comprehensive tool discovery, filtering, documentation,
// and MCP export capabilities for script environments.
func (tb *ToolsRegistryBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Tool discovery and listing
		{
			Name:        "listTools",
			Description: "List all registered tools",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listTools()"},
		},
		{
			Name:        "getTool",
			Description: "Get tool by name",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getTool('calculator')"},
		},
		{
			Name:        "searchTools",
			Description: "Search tools by query string",
			Parameters: []types.ParameterInfo{
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			ReturnType: "array",
			Examples:   []string{"searchTools('math')"},
		},
		{
			Name:        "listToolsByCategory",
			Description: "List tools in specific category",
			Parameters: []types.ParameterInfo{
				{Name: "category", Type: "string", Required: true, Description: "Tool category"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByCategory('math')"},
		},
		{
			Name:        "listToolsByTags",
			Description: "List tools matching all provided tags",
			Parameters: []types.ParameterInfo{
				{Name: "tags", Type: "array", Required: true, Description: "Array of tags"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByTags(['network', 'api'])"},
		},
		{
			Name:        "getToolCategories",
			Description: "Get all available tool categories",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"getToolCategories()"},
		},
		// Tool filtering by permissions and resources
		{
			Name:        "listToolsByPermission",
			Description: "List tools requiring specific permission",
			Parameters: []types.ParameterInfo{
				{Name: "permission", Type: "string", Required: true, Description: "Required permission"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByPermission('file:read')"},
		},
		{
			Name:        "listToolsByResourceUsage",
			Description: "List tools matching resource criteria",
			Parameters: []types.ParameterInfo{
				{Name: "criteria", Type: "object", Required: true, Description: "Resource criteria object"},
			},
			ReturnType: "array",
			Examples:   []string{"listToolsByResourceUsage({maxMemory: 'low', requiresNetwork: false})"},
		},
		// Tool documentation
		{
			Name:        "getToolDocumentation",
			Description: "Get comprehensive documentation for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getToolDocumentation('calculator')"},
		},
		// Tool registration
		{
			Name:        "registerTool",
			Description: "Register a new tool in the registry",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tool name"},
			},
			ReturnType: "object",
			Examples:   []string{"exportToolToMCP('calculator')"},
		},
		{
			Name:        "exportAllToolsToMCP",
			Description: "Export all tools to MCP catalog",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"exportAllToolsToMCP()"},
		},
		// Registry management
		{
			Name:        "clearRegistry",
			Description: "Clear all tools from registry (testing only)",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
			Examples:    []string{"clearRegistry()"},
		},
		{
			Name:        "getRegistryStats",
			Description: "Get registry statistics and metrics",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getRegistryStats()"},
		},
	}
}

// ValidateMethod validates method calls.
// Ensures bridge is initialized and validates parameter counts
// against method definitions.
func (tb *ToolsRegistryBridge) ValidateMethod(name string, args []types.ScriptValue) error {
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

// ExecuteMethod executes bridge methods with ScriptValue parameters.
// Routes method calls to appropriate implementations and returns
// script-compatible values wrapped in ScriptValue types.
func (tb *ToolsRegistryBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if !tb.initialized {
		return types.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch name {
	case "listTools":
		return tb.listTools(ctx, args)
	case "getTool":
		return tb.getTool(ctx, args)
	case "searchTools":
		return tb.searchTools(ctx, args)
	case "listToolsByCategory":
		return tb.listToolsByCategory(ctx, args)
	case "listToolsByTags":
		return tb.listToolsByTags(ctx, args)
	case "getToolCategories":
		return tb.getToolCategories(ctx, args)
	case "listToolsByPermission":
		return tb.listToolsByPermission(ctx, args)
	case "listToolsByResourceUsage":
		return tb.listToolsByResourceUsage(ctx, args)
	case "getToolDocumentation":
		return tb.getToolDocumentation(ctx, args)
	case "registerTool":
		return tb.registerTool(ctx, args)
	case "exportToolToMCP":
		return tb.exportToolToMCP(ctx, args)
	case "exportAllToolsToMCP":
		return tb.exportAllToolsToMCP(ctx, args)
	case "clearRegistry":
		return tb.clearRegistry(ctx, args)
	case "getRegistryStats":
		return tb.getRegistryStats(ctx, args)
	default:
		return types.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// TypeMappings returns type conversion mappings.
// Defines mappings between go-llms tool types and script types
// for proper data conversion during method execution.
func (tb *ToolsRegistryBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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

// RequiredPermissions returns required permissions.
// Specifies that scripts need storage access for registry operations
// and memory access for metadata handling.
func (tb *ToolsRegistryBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionStorage,
			Resource:    "tools.registry",
			Actions:     []string{"read", "write", "export"},
			Description: "Access to tools registry for discovery and management",
		},
		{
			Type:        types.PermissionMemory,
			Resource:    "registry.metadata",
			Actions:     []string{"read", "write"},
			Description: "Access to tool metadata and documentation",
		},
	}
}

// Bridge method implementations with ScriptValue

// Tool discovery and listing

// listTools lists all registered tools.
// Returns an array of tool metadata including name, description,
// category, tags, version, and status flags.
func (tb *ToolsRegistryBridge) listTools(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("listTools", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	entries := tb.registry.List()
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":         types.NewStringValue(entry.Metadata.Name),
			"description":  types.NewStringValue(entry.Metadata.Description),
			"category":     types.NewStringValue(entry.Metadata.Category),
			"tags":         convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":      types.NewStringValue(entry.Metadata.Version),
			"deprecated":   types.NewBoolValue(entry.Metadata.Deprecated),
			"experimental": types.NewBoolValue(entry.Metadata.Experimental),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// getTool gets a tool by name.
// Returns comprehensive tool information including schemas,
// constraints, examples, and operational characteristics.
func (tb *ToolsRegistryBridge) getTool(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("getTool", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	tool, found := tb.registry.Get(name)
	if !found {
		return types.NewErrorValue(fmt.Errorf("tool not found: %s", name)), nil
	}

	toolData := map[string]types.ScriptValue{
		"name":                  types.NewStringValue(tool.Name()),
		"description":           types.NewStringValue(tool.Description()),
		"category":              types.NewStringValue(tool.Category()),
		"tags":                  convertTagsToScriptValueRegistry(tool.Tags()),
		"version":               types.NewStringValue(tool.Version()),
		"usage_instructions":    types.NewStringValue(tool.UsageInstructions()),
		"examples":              convertToolExamplesToScriptValue(tool.Examples()),
		"constraints":           convertConstraintsToScriptValue(tool.Constraints()),
		"error_guidance":        convertErrorGuidanceToScriptValue(tool.ErrorGuidance()),
		"is_deterministic":      types.NewBoolValue(tool.IsDeterministic()),
		"is_destructive":        types.NewBoolValue(tool.IsDestructive()),
		"requires_confirmation": types.NewBoolValue(tool.RequiresConfirmation()),
		"estimated_latency":     types.NewStringValue(tool.EstimatedLatency()),
		"parameter_schema":      convertSchemaToScriptValue(tool.ParameterSchema()),
		"output_schema":         convertSchemaToScriptValue(tool.OutputSchema()),
	}

	return types.NewObjectValue(toolData), nil
}

// searchTools searches tools by query string.
// Performs text-based search across tool names, descriptions,
// and metadata to find matching tools.
func (tb *ToolsRegistryBridge) searchTools(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("searchTools", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	query := args[0].(types.StringValue).Value()

	entries := tb.registry.Search(query)
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":        types.NewStringValue(entry.Metadata.Name),
			"description": types.NewStringValue(entry.Metadata.Description),
			"category":    types.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     types.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// listToolsByCategory lists tools in specific category.
// Filters tools by their assigned category for organized discovery.
func (tb *ToolsRegistryBridge) listToolsByCategory(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByCategory", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	category := args[0].(types.StringValue).Value()

	entries := tb.registry.ListByCategory(category)
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":        types.NewStringValue(entry.Metadata.Name),
			"description": types.NewStringValue(entry.Metadata.Description),
			"category":    types.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     types.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// listToolsByTags lists tools matching all provided tags.
// Returns tools that have all specified tags, enabling precise
// filtering for specific capabilities.
func (tb *ToolsRegistryBridge) listToolsByTags(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByTags", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	tagsArray := args[0].ToGo().([]interface{})

	// Convert interface{} slice to string slice
	tags := make([]string, 0, len(tagsArray))
	for _, tag := range tagsArray {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	entries := tb.registry.ListByTags(tags...)
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":        types.NewStringValue(entry.Metadata.Name),
			"description": types.NewStringValue(entry.Metadata.Description),
			"category":    types.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     types.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// getToolCategories gets all available tool categories.
// Returns a list of unique categories used across all registered tools.
func (tb *ToolsRegistryBridge) getToolCategories(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("getToolCategories", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	categories := tb.registry.Categories()
	result := make([]types.ScriptValue, len(categories))
	for i, category := range categories {
		result[i] = types.NewStringValue(category)
	}

	return types.NewArrayValue(result), nil
}

// Tool filtering by permissions and resources

// listToolsByPermission lists tools requiring specific permission.
// Filters tools based on their permission requirements for security-aware
// tool discovery.
func (tb *ToolsRegistryBridge) listToolsByPermission(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByPermission", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	permission := args[0].(types.StringValue).Value()

	entries := tb.registry.ListByPermission(permission)
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":        types.NewStringValue(entry.Metadata.Name),
			"description": types.NewStringValue(entry.Metadata.Description),
			"category":    types.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     types.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// listToolsByResourceUsage lists tools matching resource criteria.
// Filters tools based on resource requirements like memory usage,
// network access, file system access, and concurrency needs.
func (tb *ToolsRegistryBridge) listToolsByResourceUsage(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByResourceUsage", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	criteriaMap := args[0].ToGo().(map[string]interface{})

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
	result := make([]types.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]types.ScriptValue{
			"name":        types.NewStringValue(entry.Metadata.Name),
			"description": types.NewStringValue(entry.Metadata.Description),
			"category":    types.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     types.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, types.NewObjectValue(toolData))
	}

	return types.NewArrayValue(result), nil
}

// Tool documentation

// getToolDocumentation gets comprehensive documentation for a tool.
// Returns complete documentation including usage instructions, examples,
// constraints, error guidance, and schema definitions.
func (tb *ToolsRegistryBridge) getToolDocumentation(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("getToolDocumentation", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	doc, err := tb.registry.GetToolDocumentation(name)
	if err != nil {
		return types.NewErrorValue(fmt.Errorf("failed to get tool documentation: %w", err)), nil
	}

	docData := map[string]types.ScriptValue{
		"name":                  types.NewStringValue(doc.Name),
		"description":           types.NewStringValue(doc.Description),
		"category":              types.NewStringValue(doc.Category),
		"tags":                  convertTagsToScriptValueRegistry(doc.Tags),
		"version":               types.NewStringValue(doc.Version),
		"usage_instructions":    types.NewStringValue(doc.UsageInstructions),
		"examples":              convertToolExamplesToScriptValue(doc.Examples),
		"constraints":           convertConstraintsToScriptValue(doc.Constraints),
		"error_guidance":        convertErrorGuidanceToScriptValue(doc.ErrorGuidance),
		"required_permissions":  convertPermissionsToScriptValue(doc.RequiredPermissions),
		"resource_usage":        convertResourceUsageToScriptValue(doc.ResourceUsage),
		"is_deterministic":      types.NewBoolValue(doc.IsDeterministic),
		"is_destructive":        types.NewBoolValue(doc.IsDestructive),
		"requires_confirmation": types.NewBoolValue(doc.RequiresConfirmation),
		"estimated_latency":     types.NewStringValue(doc.EstimatedLatency),
		"parameter_schema":      convertSchemaToScriptValue(doc.ParameterSchema),
		"output_schema":         convertSchemaToScriptValue(doc.OutputSchema),
	}

	return types.NewObjectValue(docData), nil
}

// Tool registration

// registerTool registers a new tool in the registry (simplified interface).
// Currently returns an error as tool registration from scripts requires
// proper domain.Tool implementation. Tools should be registered in Go code.
func (tb *ToolsRegistryBridge) registerTool(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("registerTool", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	// For now, we'll return an error indicating this requires a proper tool implementation
	// In a real bridge, we'd need to convert the script tool object to a domain.Tool
	_ = name
	return types.NewErrorValue(fmt.Errorf("tool registration from scripts not yet implemented - tools must be registered in Go code")), nil
}

// MCP export functionality

// exportToolToMCP exports single tool to MCP format.
// Converts a tool to Model Context Protocol format for integration
// with MCP-compatible systems.
func (tb *ToolsRegistryBridge) exportToolToMCP(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("exportToolToMCP", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	mcp, err := tb.registry.ExportToMCP(name)
	if err != nil {
		return types.NewErrorValue(fmt.Errorf("failed to export tool to MCP: %w", err)), nil
	}

	mcpData := map[string]types.ScriptValue{
		"name":         types.NewStringValue(mcp.Name),
		"description":  types.NewStringValue(mcp.Description),
		"inputSchema":  convertSchemaToScriptValue(mcp.InputSchema),
		"outputSchema": convertSchemaToScriptValue(mcp.OutputSchema),
		"annotations":  convertAnnotationsToScriptValue(mcp.Annotations),
	}

	return types.NewObjectValue(mcpData), nil
}

// exportAllToolsToMCP exports all tools to MCP catalog.
// Creates a complete MCP catalog of all registered tools for
// bulk export and integration.
func (tb *ToolsRegistryBridge) exportAllToolsToMCP(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("exportAllToolsToMCP", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	catalog, err := tb.registry.ExportAllToMCP()
	if err != nil {
		return types.NewErrorValue(fmt.Errorf("failed to export tools to MCP catalog: %w", err)), nil
	}

	// Convert tools to script-friendly format
	toolsArray := make([]types.ScriptValue, 0, len(catalog.Tools))
	for _, tool := range catalog.Tools {
		toolData := map[string]types.ScriptValue{
			"name":         types.NewStringValue(tool.Name),
			"description":  types.NewStringValue(tool.Description),
			"inputSchema":  convertSchemaToScriptValue(tool.InputSchema),
			"outputSchema": convertSchemaToScriptValue(tool.OutputSchema),
			"annotations":  convertAnnotationsToScriptValue(tool.Annotations),
		}
		toolsArray = append(toolsArray, types.NewObjectValue(toolData))
	}

	catalogData := map[string]types.ScriptValue{
		"version":     types.NewStringValue(catalog.Version),
		"description": types.NewStringValue(catalog.Description),
		"tools":       types.NewArrayValue(toolsArray),
		"metadata":    convertMetadataToScriptValue(catalog.Metadata),
	}

	return types.NewObjectValue(catalogData), nil
}

// Registry management

// clearRegistry clears all tools from registry (testing only).
// Removes all tools from the registry. Should only be used in
// testing scenarios as it affects the global registry.
func (tb *ToolsRegistryBridge) clearRegistry(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("clearRegistry", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	tb.registry.Clear()
	return types.NewNilValue(), nil
}

// getRegistryStats gets registry statistics and metrics.
// Returns comprehensive statistics including tool counts by category,
// deprecated/experimental tool counts, and category distribution.
func (tb *ToolsRegistryBridge) getRegistryStats(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := tb.ValidateMethod("getRegistryStats", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	allTools := tb.registry.List()
	categories := tb.registry.Categories()

	// Count tools by category
	categoryCount := make(map[string]types.ScriptValue)
	var deprecatedCount, experimentalCount int

	for _, entry := range allTools {
		if entry.Metadata.Category != "" {
			if existing, ok := categoryCount[entry.Metadata.Category]; ok {
				if existingNum, ok := existing.(types.NumberValue); ok {
					categoryCount[entry.Metadata.Category] = types.NewNumberValue(existingNum.Value() + 1)
				}
			} else {
				categoryCount[entry.Metadata.Category] = types.NewNumberValue(1)
			}
		}
		if entry.Metadata.Deprecated {
			deprecatedCount++
		}
		if entry.Metadata.Experimental {
			experimentalCount++
		}
	}

	// Convert categories to ScriptValue array
	categoriesSV := make([]types.ScriptValue, len(categories))
	for i, cat := range categories {
		categoriesSV[i] = types.NewStringValue(cat)
	}

	statsData := map[string]types.ScriptValue{
		"total_tools":        types.NewNumberValue(float64(len(allTools))),
		"total_categories":   types.NewNumberValue(float64(len(categories))),
		"categories":         types.NewArrayValue(categoriesSV),
		"tools_by_category":  types.NewObjectValue(categoryCount),
		"deprecated_tools":   types.NewNumberValue(float64(deprecatedCount)),
		"experimental_tools": types.NewNumberValue(float64(experimentalCount)),
	}

	return types.NewObjectValue(statsData), nil
}

// Helper functions for type conversions

func convertTagsToScriptValueRegistry(tags []string) types.ScriptValue {
	result := make([]types.ScriptValue, len(tags))
	for i, tag := range tags {
		result[i] = types.NewStringValue(tag)
	}
	return types.NewArrayValue(result)
}

func convertToolExamplesToScriptValue(examples interface{}) types.ScriptValue {
	if examples == nil {
		return types.NewArrayValue([]types.ScriptValue{})
	}
	// Convert to string representation for now
	return types.NewStringValue(fmt.Sprintf("%v", examples))
}

func convertErrorGuidanceToScriptValue(guidance interface{}) types.ScriptValue {
	if guidance == nil {
		return types.NewStringValue("")
	}
	// Convert to string representation for now
	return types.NewStringValue(fmt.Sprintf("%v", guidance))
}

func convertConstraintsToScriptValue(constraints interface{}) types.ScriptValue {
	// Convert constraints to JSON-like map
	if constraints == nil {
		return types.NewNilValue()
	}
	// For now, convert to string representation
	return types.NewStringValue(fmt.Sprintf("%v", constraints))
}

func convertSchemaToScriptValue(schema interface{}) types.ScriptValue {
	if schema == nil {
		return types.NewNilValue()
	}
	// For now, convert to string representation
	return types.NewStringValue(fmt.Sprintf("%v", schema))
}

func convertPermissionsToScriptValue(permissions []string) types.ScriptValue {
	result := make([]types.ScriptValue, len(permissions))
	for i, perm := range permissions {
		result[i] = types.NewStringValue(perm)
	}
	return types.NewArrayValue(result)
}

func convertResourceUsageToScriptValue(usage interface{}) types.ScriptValue {
	if usage == nil {
		return types.NewNilValue()
	}
	return types.NewStringValue(fmt.Sprintf("%v", usage))
}

func convertAnnotationsToScriptValue(annotations map[string]interface{}) types.ScriptValue {
	if annotations == nil {
		return types.NewNilValue()
	}
	result := make(map[string]types.ScriptValue)
	for k, v := range annotations {
		result[k] = types.NewStringValue(fmt.Sprintf("%v", v))
	}
	return types.NewObjectValue(result)
}

func convertMetadataToScriptValue(metadata map[string]interface{}) types.ScriptValue {
	if metadata == nil {
		return types.NewNilValue()
	}
	result := make(map[string]types.ScriptValue)
	for k, v := range metadata {
		result[k] = types.NewStringValue(fmt.Sprintf("%v", v))
	}
	return types.NewObjectValue(result)
}
