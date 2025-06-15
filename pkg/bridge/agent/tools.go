// ABOUTME: Tools bridge providing access to go-llms tool discovery system
// ABOUTME: Wraps go-llms tool discovery API for dynamic tool exploration and execution

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for tool functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// ToolsBridge provides access to go-llms tool discovery system
type ToolsBridge struct {
	mu          sync.RWMutex
	initialized bool
	discovery   bridge.ToolDiscovery
	customTools map[string]domain.Tool // For script-registered tools
	schemaRepo  schemaDomain.SchemaRepository
	validator   schemaDomain.Validator
}

// NewToolsBridge creates a new tools bridge
func NewToolsBridge() *ToolsBridge {
	return &ToolsBridge{}
}

// GetID returns the bridge ID
func (b *ToolsBridge) GetID() string {
	return "tools"
}

// GetMetadata returns bridge metadata
func (b *ToolsBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Tools Bridge",
		Version:     "2.0.0",
		Description: "Provides access to go-llms tool discovery system for dynamic tool exploration",
		Author:      "go-llmspell",
	}
}

// Initialize initializes the bridge
func (b *ToolsBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize discovery system
	b.discovery = tools.NewDiscovery()
	b.customTools = make(map[string]domain.Tool)

	// Initialize schema repository and validator
	b.schemaRepo = repository.NewInMemorySchemaRepository()
	b.validator = validation.NewValidator()

	b.initialized = true
	return nil
}

// Cleanup performs cleanup
func (b *ToolsBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	b.discovery = nil
	b.customTools = nil
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *ToolsBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *ToolsBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *ToolsBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Tool discovery methods
		{
			Name:        "listTools",
			Description: "List all available tools with metadata",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "searchTools",
			Description: "Search tools by keyword in name, description, or tags",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "string", Description: "Search query", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "listByCategory",
			Description: "List tools in a specific category",
			Parameters: []engine.ParameterInfo{
				{Name: "category", Type: "string", Description: "Tool category", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getToolInfo",
			Description: "Get detailed information about a specific tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getToolSchema",
			Description: "Get parameter and output schemas for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getToolHelp",
			Description: "Get help text for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "getToolExamples",
			Description: "Get usage examples for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "array",
		},
		// Tool creation and execution
		{
			Name:        "createTool",
			Description: "Create a tool instance by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "executeTool",
			Description: "Execute a tool with parameters",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Tool parameters", Required: true},
			},
			ReturnType: "any",
		},
		// Custom tool registration
		{
			Name:        "registerCustomTool",
			Description: "Register a custom tool implementation",
			Parameters: []engine.ParameterInfo{
				{Name: "tool", Type: "object", Description: "Tool definition", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *ToolsBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Tool": {
			GoType:     "Tool",
			ScriptType: "object",
		},
		"ToolInfo": {
			GoType:     "ToolInfo",
			ScriptType: "object",
		},
		"ToolSchema": {
			GoType:     "ToolSchema",
			ScriptType: "object",
		},
		"ToolExample": {
			GoType:     "ToolExample",
			ScriptType: "object",
		},
		"ToolContext": {
			GoType:     "ToolContext",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *ToolsBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions
func (b *ToolsBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionProcess,
			Resource:    "tool",
			Actions:     []string{"execute", "register", "list"},
			Description: "Tool execution and management",
		},
		{
			Type:        engine.PermissionFileSystem,
			Resource:    "*",
			Actions:     []string{"read", "write"},
			Description: "File system access for file tools",
		},
		{
			Type:        engine.PermissionNetwork,
			Resource:    "*",
			Actions:     []string{"http"},
			Description: "Network access for web tools",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *ToolsBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "listTools":
		tools := b.discovery.ListTools()
		result := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			result = append(result, toolInfoToMap(tool))
		}
		return result, nil

	case "searchTools":
		if len(args) < 1 {
			return nil, fmt.Errorf("searchTools requires query parameter")
		}
		query, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("query must be string")
		}

		tools := b.discovery.SearchTools(query)
		result := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			result = append(result, toolInfoToMap(tool))
		}
		return result, nil

	case "listByCategory":
		if len(args) < 1 {
			return nil, fmt.Errorf("listByCategory requires category parameter")
		}
		category, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("category must be string")
		}

		tools := b.discovery.ListByCategory(category)
		result := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			result = append(result, toolInfoToMap(tool))
		}
		return result, nil

	case "getToolInfo":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolInfo requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return customToolToInfo(name, tool), nil
		}

		// Get from discovery
		tools := b.discovery.ListTools()
		for _, tool := range tools {
			if tool.Name == name {
				return toolInfoToMap(tool), nil
			}
		}
		return nil, fmt.Errorf("tool not found: %s", name)

	case "getToolSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolSchema requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return b.toolToSchemaMap(tool), nil
		}

		// Get from discovery
		schema, err := b.discovery.GetToolSchema(name)
		if err != nil {
			return nil, err
		}

		return toolSchemaToMap(schema), nil

	case "getToolHelp":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolHelp requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return tool.UsageInstructions(), nil
		}

		// Get from discovery
		help, err := b.discovery.GetToolHelp(name)
		if err != nil {
			return nil, err
		}

		return help, nil

	case "getToolExamples":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolExamples requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		var examples []domain.ToolExample

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			examples = tool.Examples()
		} else {
			// Get from discovery
			var err error
			examples, err = b.discovery.GetToolExamples(name)
			if err != nil {
				return nil, err
			}
		}

		result := make([]map[string]interface{}, 0, len(examples))
		for _, ex := range examples {
			result = append(result, map[string]interface{}{
				"name":        ex.Name,
				"description": ex.Description,
				"input":       ex.Input,
				"output":      ex.Output,
			})
		}
		return result, nil

	case "createTool":
		if len(args) < 1 {
			return nil, fmt.Errorf("createTool requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return toolToWrapper(name, tool), nil
		}

		// Create from discovery
		tool, err := b.discovery.CreateTool(name)
		if err != nil {
			return nil, err
		}

		return toolToWrapper(name, tool), nil

	case "executeTool":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeTool requires name and params parameters")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}
		params := args[1]

		var tool domain.Tool

		// Check custom tools first
		if customTool, exists := b.customTools[name]; exists {
			tool = customTool
		} else {
			// Create from discovery
			var err error
			tool, err = b.discovery.CreateTool(name)
			if err != nil {
				return nil, err
			}
		}

		// Create a basic tool context
		toolCtx := &domain.ToolContext{
			Context: ctx,
		}

		// Execute the tool
		result, err := tool.Execute(toolCtx, params)
		if err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", err)
		}

		return result, nil

	case "registerCustomTool":
		if len(args) < 1 {
			return nil, fmt.Errorf("registerCustomTool requires tool parameter")
		}
		toolDef, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tool must be object")
		}

		name, ok := toolDef["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("tool must have name")
		}

		// Create enhanced custom tool using ToolBuilder
		tool, err := b.createEnhancedCustomTool(toolDef)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom tool: %w", err)
		}

		b.customTools[name] = tool
		return nil, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper functions

func toolInfoToMap(info bridge.ToolInfo) map[string]interface{} {
	result := map[string]interface{}{
		"name":        info.Name,
		"description": info.Description,
		"category":    info.Category,
		"tags":        info.Tags,
		"version":     info.Version,
		"usageHint":   info.UsageHint,
		"package":     info.Package,
	}

	// Parse schemas if available
	if len(info.ParameterSchema) > 0 {
		var params interface{}
		if err := json.Unmarshal(info.ParameterSchema, &params); err == nil {
			result["parameterSchema"] = params
		}
	}

	if len(info.OutputSchema) > 0 {
		var output interface{}
		if err := json.Unmarshal(info.OutputSchema, &output); err == nil {
			result["outputSchema"] = output
		}
	}

	return result
}

func toolSchemaToMap(schema *bridge.ToolSchema) map[string]interface{} {
	return map[string]interface{}{
		"name":          schema.Name,
		"description":   schema.Description,
		"parameters":    schema.Parameters,
		"output":        schema.Output,
		"examples":      schema.Examples,
		"constraints":   schema.Constraints,
		"errorGuidance": schema.ErrorGuidance,
	}
}

func toolToWrapper(name string, tool domain.Tool) map[string]interface{} {
	return map[string]interface{}{
		"name":                 name,
		"description":          tool.Description(),
		"category":             tool.Category(),
		"tags":                 tool.Tags(),
		"version":              tool.Version(),
		"isDeterministic":      tool.IsDeterministic(),
		"isDestructive":        tool.IsDestructive(),
		"requiresConfirmation": tool.RequiresConfirmation(),
		"estimatedLatency":     tool.EstimatedLatency(),
		"usageInstructions":    tool.UsageInstructions(),
		"constraints":          tool.Constraints(),
	}
}

func customToolToInfo(name string, tool domain.Tool) map[string]interface{} {
	return map[string]interface{}{
		"name":                 name,
		"description":          tool.Description(),
		"category":             tool.Category(),
		"tags":                 tool.Tags(),
		"version":              tool.Version(),
		"custom":               true,
		"isDeterministic":      tool.IsDeterministic(),
		"isDestructive":        tool.IsDestructive(),
		"requiresConfirmation": tool.RequiresConfirmation(),
		"estimatedLatency":     tool.EstimatedLatency(),
		"usageInstructions":    tool.UsageInstructions(),
		"constraints":          tool.Constraints(),
		"errorGuidance":        tool.ErrorGuidance(),
	}
}

func getStringField(m map[string]interface{}, field string) string {
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}

// createEnhancedCustomTool creates a custom tool with full schema support using ToolBuilder
func (b *ToolsBridge) createEnhancedCustomTool(toolDef map[string]interface{}) (domain.Tool, error) {
	name := getStringField(toolDef, "name")
	description := getStringField(toolDef, "description")

	// Create tool using ToolBuilder
	builder := tools.NewToolBuilder(name, description)

	// Convert and set parameter schema
	if paramSchemaData, ok := toolDef["parameterSchema"]; ok {
		paramSchema, err := b.convertToSchema(paramSchemaData)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter schema: %w", err)
		}
		builder.WithParameterSchema(paramSchema)
	}

	// Convert and set output schema
	if outputSchemaData, ok := toolDef["outputSchema"]; ok {
		outputSchema, err := b.convertToSchema(outputSchemaData)
		if err != nil {
			return nil, fmt.Errorf("invalid output schema: %w", err)
		}
		builder.WithOutputSchema(outputSchema)
	}

	// Set metadata
	if category := getStringField(toolDef, "category"); category != "" {
		builder.WithCategory(category)
	}

	if usageInstructions := getStringField(toolDef, "usageInstructions"); usageInstructions != "" {
		builder.WithUsageInstructions(usageInstructions)
	}

	// Convert and set examples
	if examplesData, ok := toolDef["examples"]; ok {
		var examples []domain.ToolExample
		switch data := examplesData.(type) {
		case []interface{}:
			examples = b.convertToExamples(data)
		case []map[string]interface{}:
			// Convert []map[string]interface{} to []interface{}
			interfaceSlice := make([]interface{}, len(data))
			for i, m := range data {
				interfaceSlice[i] = m
			}
			examples = b.convertToExamples(interfaceSlice)
		}
		if len(examples) > 0 {
			builder.WithExamples(examples)
		}
	}

	// Set constraints
	if constraintsData, ok := toolDef["constraints"].([]interface{}); ok {
		constraints := make([]string, 0, len(constraintsData))
		for _, c := range constraintsData {
			if str, ok := c.(string); ok {
				constraints = append(constraints, str)
			}
		}
		builder.WithConstraints(constraints)
	}

	// Set error guidance
	if errorGuidanceData, ok := toolDef["errorGuidance"].(map[string]interface{}); ok {
		errorGuidance := make(map[string]string)
		for k, v := range errorGuidanceData {
			if str, ok := v.(string); ok {
				errorGuidance[k] = str
			}
		}
		builder.WithErrorGuidance(errorGuidance)
	}

	// Set behavioral flags
	isDeterministic := getBoolField(toolDef, "isDeterministic", true)
	isDestructive := getBoolField(toolDef, "isDestructive", false)
	requiresConfirmation := getBoolField(toolDef, "requiresConfirmation", false)
	estimatedLatency := getStringField(toolDef, "estimatedLatency")
	if estimatedLatency == "" {
		estimatedLatency = "medium"
	}

	builder.WithBehavior(isDeterministic, isDestructive, requiresConfirmation, estimatedLatency)

	// Set version
	if version := getStringField(toolDef, "version"); version != "" {
		builder.WithVersion(version)
	}

	// Set tags
	if tagsData, ok := toolDef["tags"].([]interface{}); ok {
		tags := make([]string, 0, len(tagsData))
		for _, t := range tagsData {
			if str, ok := t.(string); ok {
				tags = append(tags, str)
			}
		}
		builder.WithTags(tags)
	}

	// Store the parameter and output schemas for validation
	var paramSchema *schemaDomain.Schema
	var outputSchema *schemaDomain.Schema

	if paramSchemaData, ok := toolDef["parameterSchema"]; ok {
		paramSchema, _ = b.convertToSchema(paramSchemaData)
	}

	if outputSchemaData, ok := toolDef["outputSchema"]; ok {
		outputSchema, _ = b.convertToSchema(outputSchemaData)
	}

	// Create wrapper function that validates inputs/outputs against schemas
	scriptExecute := toolDef["execute"]
	wrapperFunc := func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
		// Validate input parameters if schema exists
		if paramSchema != nil {
			result, err := b.validator.ValidateStruct(paramSchema, params)
			if err != nil {
				return nil, fmt.Errorf("parameter validation error: %w", err)
			}
			if !result.Valid {
				return nil, fmt.Errorf("parameter validation failed: %v", result.Errors)
			}
		}

		// Execute the script function
		var result interface{}
		var err error

		if execFn, ok := scriptExecute.(func(interface{}, interface{}) (interface{}, error)); ok {
			result, err = execFn(ctx, params)
		} else {
			return nil, fmt.Errorf("custom tool execute function not valid")
		}

		if err != nil {
			return nil, err
		}

		// Validate output if schema exists
		if outputSchema != nil {
			validationResult, err := b.validator.ValidateStruct(outputSchema, result)
			if err != nil {
				return nil, fmt.Errorf("output validation error: %w", err)
			}
			if !validationResult.Valid {
				return nil, fmt.Errorf("output validation failed: %v", validationResult.Errors)
			}
		}

		return result, nil
	}

	builder.WithFunction(wrapperFunc)

	return builder.Build(), nil
}

// convertToSchema converts script JSON Schema to domain.Schema
func (b *ToolsBridge) convertToSchema(schemaData interface{}) (*schemaDomain.Schema, error) {
	schemaMap, ok := schemaData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema must be an object")
	}

	// Convert the schema map to JSON
	schemaJSON, err := json.Marshal(schemaMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Parse JSON schema
	schema := &schemaDomain.Schema{}
	if err := json.Unmarshal(schemaJSON, schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return schema, nil
}

// convertToExamples converts script examples to domain.ToolExample
func (b *ToolsBridge) convertToExamples(examplesData []interface{}) []domain.ToolExample {
	examples := make([]domain.ToolExample, 0, len(examplesData))

	for _, exData := range examplesData {
		if exMap, ok := exData.(map[string]interface{}); ok {
			example := domain.ToolExample{
				Name:        getStringField(exMap, "name"),
				Description: getStringField(exMap, "description"),
				Input:       exMap["input"],
				Output:      exMap["output"],
			}
			examples = append(examples, example)
		}
	}

	return examples
}

// getBoolField gets a boolean field from a map with a default value
func getBoolField(m map[string]interface{}, field string, defaultValue bool) bool {
	if v, ok := m[field].(bool); ok {
		return v
	}
	return defaultValue
}

// toolToSchemaMap converts a tool's schemas to a map
func (b *ToolsBridge) toolToSchemaMap(tool domain.Tool) map[string]interface{} {
	result := map[string]interface{}{
		"name":          tool.Name(),
		"description":   tool.Description(),
		"constraints":   tool.Constraints(),
		"errorGuidance": tool.ErrorGuidance(),
	}

	// Convert parameter schema
	if paramSchema := tool.ParameterSchema(); paramSchema != nil {
		result["parameters"] = b.schemaToMap(paramSchema)
	}

	// Convert output schema
	if outputSchema := tool.OutputSchema(); outputSchema != nil {
		result["output"] = b.schemaToMap(outputSchema)
	}

	// Convert examples
	examples := tool.Examples()
	if len(examples) > 0 {
		exampleMaps := make([]map[string]interface{}, 0, len(examples))
		for _, ex := range examples {
			exampleMaps = append(exampleMaps, map[string]interface{}{
				"name":        ex.Name,
				"description": ex.Description,
				"input":       ex.Input,
				"output":      ex.Output,
			})
		}
		result["examples"] = exampleMaps
	}

	return result
}

// schemaToMap converts a domain.Schema to a map
func (b *ToolsBridge) schemaToMap(schema *schemaDomain.Schema) map[string]interface{} {
	// Marshal schema to JSON then unmarshal to map
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaJSON, &schemaMap); err != nil {
		return nil
	}

	return schemaMap
}
