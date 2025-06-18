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
func (tb *ToolsRegistryBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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

// ExecuteMethod executes bridge methods with ScriptValue parameters
func (tb *ToolsRegistryBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if !tb.initialized {
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
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
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
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

// Bridge method implementations with ScriptValue

// Tool discovery and listing

// listTools lists all registered tools
func (tb *ToolsRegistryBridge) listTools(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("listTools", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	entries := tb.registry.List()
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":         engine.NewStringValue(entry.Metadata.Name),
			"description":  engine.NewStringValue(entry.Metadata.Description),
			"category":     engine.NewStringValue(entry.Metadata.Category),
			"tags":         convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":      engine.NewStringValue(entry.Metadata.Version),
			"deprecated":   engine.NewBoolValue(entry.Metadata.Deprecated),
			"experimental": engine.NewBoolValue(entry.Metadata.Experimental),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// getTool gets a tool by name
func (tb *ToolsRegistryBridge) getTool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("getTool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	tool, found := tb.registry.Get(name)
	if !found {
		return engine.NewErrorValue(fmt.Errorf("tool not found: %s", name)), nil
	}

	toolData := map[string]engine.ScriptValue{
		"name":                  engine.NewStringValue(tool.Name()),
		"description":           engine.NewStringValue(tool.Description()),
		"category":              engine.NewStringValue(tool.Category()),
		"tags":                  convertTagsToScriptValueRegistry(tool.Tags()),
		"version":               engine.NewStringValue(tool.Version()),
		"usage_instructions":    engine.NewStringValue(tool.UsageInstructions()),
		"examples":              convertToolExamplesToScriptValue(tool.Examples()),
		"constraints":           convertConstraintsToScriptValue(tool.Constraints()),
		"error_guidance":        convertErrorGuidanceToScriptValue(tool.ErrorGuidance()),
		"is_deterministic":      engine.NewBoolValue(tool.IsDeterministic()),
		"is_destructive":        engine.NewBoolValue(tool.IsDestructive()),
		"requires_confirmation": engine.NewBoolValue(tool.RequiresConfirmation()),
		"estimated_latency":     engine.NewStringValue(tool.EstimatedLatency()),
		"parameter_schema":      convertSchemaToScriptValue(tool.ParameterSchema()),
		"output_schema":         convertSchemaToScriptValue(tool.OutputSchema()),
	}

	return engine.NewObjectValue(toolData), nil
}

// searchTools searches tools by query string
func (tb *ToolsRegistryBridge) searchTools(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("searchTools", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	query := args[0].(engine.StringValue).Value()

	entries := tb.registry.Search(query)
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(entry.Metadata.Name),
			"description": engine.NewStringValue(entry.Metadata.Description),
			"category":    engine.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     engine.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// listToolsByCategory lists tools in specific category
func (tb *ToolsRegistryBridge) listToolsByCategory(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByCategory", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	category := args[0].(engine.StringValue).Value()

	entries := tb.registry.ListByCategory(category)
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(entry.Metadata.Name),
			"description": engine.NewStringValue(entry.Metadata.Description),
			"category":    engine.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     engine.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// listToolsByTags lists tools matching all provided tags
func (tb *ToolsRegistryBridge) listToolsByTags(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByTags", args); err != nil {
		return engine.NewErrorValue(err), nil
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
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(entry.Metadata.Name),
			"description": engine.NewStringValue(entry.Metadata.Description),
			"category":    engine.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     engine.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// getToolCategories gets all available tool categories
func (tb *ToolsRegistryBridge) getToolCategories(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("getToolCategories", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	categories := tb.registry.Categories()
	result := make([]engine.ScriptValue, len(categories))
	for i, category := range categories {
		result[i] = engine.NewStringValue(category)
	}

	return engine.NewArrayValue(result), nil
}

// Tool filtering by permissions and resources

// listToolsByPermission lists tools requiring specific permission
func (tb *ToolsRegistryBridge) listToolsByPermission(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByPermission", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	permission := args[0].(engine.StringValue).Value()

	entries := tb.registry.ListByPermission(permission)
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(entry.Metadata.Name),
			"description": engine.NewStringValue(entry.Metadata.Description),
			"category":    engine.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     engine.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// listToolsByResourceUsage lists tools matching resource criteria
func (tb *ToolsRegistryBridge) listToolsByResourceUsage(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("listToolsByResourceUsage", args); err != nil {
		return engine.NewErrorValue(err), nil
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
	result := make([]engine.ScriptValue, 0, len(entries))

	for _, entry := range entries {
		toolData := map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(entry.Metadata.Name),
			"description": engine.NewStringValue(entry.Metadata.Description),
			"category":    engine.NewStringValue(entry.Metadata.Category),
			"tags":        convertTagsToScriptValueRegistry(entry.Metadata.Tags),
			"version":     engine.NewStringValue(entry.Metadata.Version),
		}
		result = append(result, engine.NewObjectValue(toolData))
	}

	return engine.NewArrayValue(result), nil
}

// Tool documentation

// getToolDocumentation gets comprehensive documentation for a tool
func (tb *ToolsRegistryBridge) getToolDocumentation(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("getToolDocumentation", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	doc, err := tb.registry.GetToolDocumentation(name)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to get tool documentation: %w", err)), nil
	}

	docData := map[string]engine.ScriptValue{
		"name":                  engine.NewStringValue(doc.Name),
		"description":           engine.NewStringValue(doc.Description),
		"category":              engine.NewStringValue(doc.Category),
		"tags":                  convertTagsToScriptValueRegistry(doc.Tags),
		"version":               engine.NewStringValue(doc.Version),
		"usage_instructions":    engine.NewStringValue(doc.UsageInstructions),
		"examples":              convertToolExamplesToScriptValue(doc.Examples),
		"constraints":           convertConstraintsToScriptValue(doc.Constraints),
		"error_guidance":        convertErrorGuidanceToScriptValue(doc.ErrorGuidance),
		"required_permissions":  convertPermissionsToScriptValue(doc.RequiredPermissions),
		"resource_usage":        convertResourceUsageToScriptValue(doc.ResourceUsage),
		"is_deterministic":      engine.NewBoolValue(doc.IsDeterministic),
		"is_destructive":        engine.NewBoolValue(doc.IsDestructive),
		"requires_confirmation": engine.NewBoolValue(doc.RequiresConfirmation),
		"estimated_latency":     engine.NewStringValue(doc.EstimatedLatency),
		"parameter_schema":      convertSchemaToScriptValue(doc.ParameterSchema),
		"output_schema":         convertSchemaToScriptValue(doc.OutputSchema),
	}

	return engine.NewObjectValue(docData), nil
}

// Tool registration

// registerTool registers a new tool in the registry (simplified interface)
func (tb *ToolsRegistryBridge) registerTool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("registerTool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	// For now, we'll return an error indicating this requires a proper tool implementation
	// In a real bridge, we'd need to convert the script tool object to a domain.Tool
	_ = name
	return engine.NewErrorValue(fmt.Errorf("tool registration from scripts not yet implemented - tools must be registered in Go code")), nil
}

// MCP export functionality

// exportToolToMCP exports single tool to MCP format
func (tb *ToolsRegistryBridge) exportToolToMCP(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("exportToolToMCP", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	mcp, err := tb.registry.ExportToMCP(name)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to export tool to MCP: %w", err)), nil
	}

	mcpData := map[string]engine.ScriptValue{
		"name":         engine.NewStringValue(mcp.Name),
		"description":  engine.NewStringValue(mcp.Description),
		"inputSchema":  convertSchemaToScriptValue(mcp.InputSchema),
		"outputSchema": convertSchemaToScriptValue(mcp.OutputSchema),
		"annotations":  convertAnnotationsToScriptValue(mcp.Annotations),
	}

	return engine.NewObjectValue(mcpData), nil
}

// exportAllToolsToMCP exports all tools to MCP catalog
func (tb *ToolsRegistryBridge) exportAllToolsToMCP(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("exportAllToolsToMCP", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	catalog, err := tb.registry.ExportAllToMCP()
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to export tools to MCP catalog: %w", err)), nil
	}

	// Convert tools to script-friendly format
	toolsArray := make([]engine.ScriptValue, 0, len(catalog.Tools))
	for _, tool := range catalog.Tools {
		toolData := map[string]engine.ScriptValue{
			"name":         engine.NewStringValue(tool.Name),
			"description":  engine.NewStringValue(tool.Description),
			"inputSchema":  convertSchemaToScriptValue(tool.InputSchema),
			"outputSchema": convertSchemaToScriptValue(tool.OutputSchema),
			"annotations":  convertAnnotationsToScriptValue(tool.Annotations),
		}
		toolsArray = append(toolsArray, engine.NewObjectValue(toolData))
	}

	catalogData := map[string]engine.ScriptValue{
		"version":     engine.NewStringValue(catalog.Version),
		"description": engine.NewStringValue(catalog.Description),
		"tools":       engine.NewArrayValue(toolsArray),
		"metadata":    convertMetadataToScriptValue(catalog.Metadata),
	}

	return engine.NewObjectValue(catalogData), nil
}

// Registry management

// clearRegistry clears all tools from registry (testing only)
func (tb *ToolsRegistryBridge) clearRegistry(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("clearRegistry", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	tb.registry.Clear()
	return engine.NewNilValue(), nil
}

// getRegistryStats gets registry statistics and metrics
func (tb *ToolsRegistryBridge) getRegistryStats(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := tb.ValidateMethod("getRegistryStats", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	allTools := tb.registry.List()
	categories := tb.registry.Categories()

	// Count tools by category
	categoryCount := make(map[string]engine.ScriptValue)
	var deprecatedCount, experimentalCount int

	for _, entry := range allTools {
		if entry.Metadata.Category != "" {
			if existing, ok := categoryCount[entry.Metadata.Category]; ok {
				if existingNum, ok := existing.(engine.NumberValue); ok {
					categoryCount[entry.Metadata.Category] = engine.NewNumberValue(existingNum.Value() + 1)
				}
			} else {
				categoryCount[entry.Metadata.Category] = engine.NewNumberValue(1)
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
	categoriesSV := make([]engine.ScriptValue, len(categories))
	for i, cat := range categories {
		categoriesSV[i] = engine.NewStringValue(cat)
	}

	statsData := map[string]engine.ScriptValue{
		"total_tools":        engine.NewNumberValue(float64(len(allTools))),
		"total_categories":   engine.NewNumberValue(float64(len(categories))),
		"categories":         engine.NewArrayValue(categoriesSV),
		"tools_by_category":  engine.NewObjectValue(categoryCount),
		"deprecated_tools":   engine.NewNumberValue(float64(deprecatedCount)),
		"experimental_tools": engine.NewNumberValue(float64(experimentalCount)),
	}

	return engine.NewObjectValue(statsData), nil
}

// Helper functions for type conversions

func convertTagsToScriptValueRegistry(tags []string) engine.ScriptValue {
	result := make([]engine.ScriptValue, len(tags))
	for i, tag := range tags {
		result[i] = engine.NewStringValue(tag)
	}
	return engine.NewArrayValue(result)
}

func convertExamplesToScriptValue(examples []string) engine.ScriptValue {
	result := make([]engine.ScriptValue, len(examples))
	for i, example := range examples {
		result[i] = engine.NewStringValue(example)
	}
	return engine.NewArrayValue(result)
}

func convertToolExamplesToScriptValue(examples interface{}) engine.ScriptValue {
	if examples == nil {
		return engine.NewArrayValue([]engine.ScriptValue{})
	}
	// Convert to string representation for now
	return engine.NewStringValue(fmt.Sprintf("%v", examples))
}

func convertErrorGuidanceToScriptValue(guidance interface{}) engine.ScriptValue {
	if guidance == nil {
		return engine.NewStringValue("")
	}
	// Convert to string representation for now
	return engine.NewStringValue(fmt.Sprintf("%v", guidance))
}

func convertConstraintsToScriptValue(constraints interface{}) engine.ScriptValue {
	// Convert constraints to JSON-like map
	if constraints == nil {
		return engine.NewNilValue()
	}
	// For now, convert to string representation
	return engine.NewStringValue(fmt.Sprintf("%v", constraints))
}

func convertSchemaToScriptValue(schema interface{}) engine.ScriptValue {
	if schema == nil {
		return engine.NewNilValue()
	}
	// For now, convert to string representation
	return engine.NewStringValue(fmt.Sprintf("%v", schema))
}

func convertPermissionsToScriptValue(permissions []string) engine.ScriptValue {
	result := make([]engine.ScriptValue, len(permissions))
	for i, perm := range permissions {
		result[i] = engine.NewStringValue(perm)
	}
	return engine.NewArrayValue(result)
}

func convertResourceUsageToScriptValue(usage interface{}) engine.ScriptValue {
	if usage == nil {
		return engine.NewNilValue()
	}
	return engine.NewStringValue(fmt.Sprintf("%v", usage))
}

func convertAnnotationsToScriptValue(annotations map[string]interface{}) engine.ScriptValue {
	if annotations == nil {
		return engine.NewNilValue()
	}
	result := make(map[string]engine.ScriptValue)
	for k, v := range annotations {
		result[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
	}
	return engine.NewObjectValue(result)
}

func convertMetadataToScriptValue(metadata map[string]interface{}) engine.ScriptValue {
	if metadata == nil {
		return engine.NewNilValue()
	}
	result := make(map[string]engine.ScriptValue)
	for k, v := range metadata {
		result[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
	}
	return engine.NewObjectValue(result)
}
