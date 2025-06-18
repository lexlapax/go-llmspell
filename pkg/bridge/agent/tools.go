// ABOUTME: Tools bridge providing access to go-llms tool discovery system
// ABOUTME: Wraps go-llms tool discovery API for dynamic tool exploration and execution

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for tool functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/docs"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// ToolsBridge provides access to go-llms tool discovery system with v2.0.0 enhancements
type ToolsBridge struct {
	mu          sync.RWMutex
	initialized bool
	discovery   bridge.ToolDiscovery
	customTools map[string]domain.Tool // For script-registered tools

	// Schema validation (Task 1.4.9.1)
	schemaRepo        schemaDomain.SchemaRepository
	validator         schemaDomain.Validator
	validationCache   map[string]*schemaDomain.ValidationResult // Cache validation results
	validationReports map[string]*ValidationReport              // Store validation reports

	// Documentation generation (Task 1.4.9.2)
	docGenerator *docs.ToolDocumentationIntegrator
	docConfig    docs.GeneratorConfig
	docCache     map[string]*docs.Documentation // Cache generated docs

	// Execution analytics (Task 1.4.9.3)
	profiler         *profiling.Profiler
	executionMetrics map[string]*ExecutionMetrics // Track tool execution metrics
	metricsLock      sync.RWMutex
}

// ValidationReport stores detailed validation results for a tool
type ValidationReport struct {
	ToolName         string                         `json:"toolName"`
	Timestamp        time.Time                      `json:"timestamp"`
	InputValidation  *schemaDomain.ValidationResult `json:"inputValidation,omitempty"`
	OutputValidation *schemaDomain.ValidationResult `json:"outputValidation,omitempty"`
	SchemaIssues     []string                       `json:"schemaIssues,omitempty"`
	Recommendations  []string                       `json:"recommendations,omitempty"`
}

// ExecutionMetrics tracks tool execution statistics
type ExecutionMetrics struct {
	ToolName        string                 `json:"toolName"`
	TotalExecutions int64                  `json:"totalExecutions"`
	SuccessCount    int64                  `json:"successCount"`
	FailureCount    int64                  `json:"failureCount"`
	TotalDuration   time.Duration          `json:"totalDuration"`
	AverageDuration time.Duration          `json:"averageDuration"`
	MinDuration     time.Duration          `json:"minDuration"`
	MaxDuration     time.Duration          `json:"maxDuration"`
	LastExecution   time.Time              `json:"lastExecution"`
	ErrorTypes      map[string]int         `json:"errorTypes"`
	ParameterStats  map[string]interface{} `json:"parameterStats"`
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
		Version:     "2.1.0",
		Description: "Enhanced tools bridge with schema validation, documentation generation, and execution analytics (v0.3.5)",
		Author:      "go-llmspell",
		License:     "MIT",
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

	// Initialize schema validation (Task 1.4.9.1)
	b.schemaRepo = repository.NewInMemorySchemaRepository()
	b.validator = validation.NewValidator(validation.WithCoercion(true))
	b.validationCache = make(map[string]*schemaDomain.ValidationResult)
	b.validationReports = make(map[string]*ValidationReport)

	// Initialize documentation generation (Task 1.4.9.2)
	b.docConfig = docs.GeneratorConfig{
		Title:       "Tools Documentation",
		Version:     "1.0.0",
		Description: "Auto-generated documentation for available tools",
	}
	b.docGenerator = docs.NewToolDocumentationIntegrator(b.discovery, b.docConfig)
	b.docCache = make(map[string]*docs.Documentation)

	// Initialize execution analytics (Task 1.4.9.3)
	b.profiler = profiling.NewProfiler("tools_bridge")
	b.executionMetrics = make(map[string]*ExecutionMetrics)

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
		// Schema validation methods (Task 1.4.9.1)
		{
			Name:        "executeToolValidated",
			Description: "Execute a tool with schema validation of inputs and outputs",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Tool parameters", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateToolInput",
			Description: "Validate tool input parameters against schema",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Parameters to validate", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateToolOutput",
			Description: "Validate tool output against schema",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "output", Type: "any", Description: "Output to validate", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getValidationReport",
			Description: "Get validation report for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		// Documentation generation methods (Task 1.4.9.2)
		{
			Name:        "generateToolDocumentation",
			Description: "Generate comprehensive documentation for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "format", Type: "string", Description: "Documentation format (markdown, openapi, json)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateAllToolsDocs",
			Description: "Generate documentation for all available tools",
			Parameters: []engine.ParameterInfo{
				{Name: "format", Type: "string", Description: "Documentation format", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateToolPlayground",
			Description: "Generate interactive playground HTML for a tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateSDKSnippet",
			Description: "Generate SDK code snippet for tool usage",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "language", Type: "string", Description: "Programming language (go, python, javascript)", Required: true},
			},
			ReturnType: "string",
		},
		// Execution analytics methods (Task 1.4.9.3)
		{
			Name:        "getToolMetrics",
			Description: "Get execution metrics for a specific tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getAllToolsMetrics",
			Description: "Get execution metrics for all tools",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getToolUsageReport",
			Description: "Generate usage report for tools",
			Parameters: []engine.ParameterInfo{
				{Name: "period", Type: "string", Description: "Time period (hour, day, week, month)", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "enableToolProfiling",
			Description: "Enable profiling for a specific tool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getToolAnomalies",
			Description: "Get anomaly alerts for tool execution",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: false},
			},
			ReturnType: "array",
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
		"ValidationResult": {
			GoType:     "ValidationResult",
			ScriptType: "object",
		},
		"ValidationReport": {
			GoType:     "ValidationReport",
			ScriptType: "object",
		},
		"ExecutionMetrics": {
			GoType:     "ExecutionMetrics",
			ScriptType: "object",
		},
		"Documentation": {
			GoType:     "Documentation",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *ToolsBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	switch name {
	case "listTools", "getAllToolsMetrics", "generateAllToolsDocs":
		// No arguments required
		return nil
	case "searchTools", "listByCategory", "getToolInfo", "getToolSchema", "getToolHelp", "getToolExamples", "createTool", "getToolMetrics", "enableToolProfiling":
		if len(args) < 1 {
			return fmt.Errorf("%s requires at least one argument", name)
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return fmt.Errorf("%s requires first argument to be string", name)
		}
		return nil
	case "executeTool", "executeToolValidated", "validateToolInput", "validateToolOutput":
		if len(args) < 2 {
			return fmt.Errorf("%s requires at least two arguments", name)
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return fmt.Errorf("%s requires first argument to be string", name)
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return fmt.Errorf("%s requires second argument to be object", name)
		}
		return nil
	case "registerCustomTool":
		if len(args) < 1 {
			return fmt.Errorf("registerCustomTool requires tool definition")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return fmt.Errorf("registerCustomTool requires argument to be object")
		}
		return nil
	case "getValidationReport", "getToolUsageReport", "getToolAnomalies":
		// Optional arguments
		return nil
	case "generateToolDocumentation", "generateSDKSnippet", "generateToolPlayground":
		if len(args) < 1 {
			return fmt.Errorf("%s requires at least one argument", name)
		}
		return nil
	default:
		return fmt.Errorf("unknown method: %s", name)
	}
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
func (b *ToolsBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "listTools":
		tools := b.discovery.ListTools()
		result := make([]engine.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, engine.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return engine.NewArrayValue(result), nil

	case "searchTools":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("searchTools requires query parameter as string")
		}
		query := args[0].(engine.StringValue).Value()

		tools := b.discovery.SearchTools(query)
		result := make([]engine.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, engine.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return engine.NewArrayValue(result), nil

	case "listByCategory":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("listByCategory requires category parameter as string")
		}
		category := args[0].(engine.StringValue).Value()

		tools := b.discovery.ListByCategory(category)
		result := make([]engine.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, engine.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return engine.NewArrayValue(result), nil

	case "getToolInfo":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("getToolInfo requires name parameter as string")
		}
		name := args[0].(engine.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return engine.NewObjectValue(customToolToScriptValue(name, tool)), nil
		}

		// Get from discovery
		tools := b.discovery.ListTools()
		for _, tool := range tools {
			if tool.Name == name {
				return engine.NewObjectValue(toolInfoToScriptValue(tool)), nil
			}
		}
		return nil, fmt.Errorf("tool not found: %s", name)

	case "getToolSchema":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("getToolSchema requires name parameter as string")
		}
		name := args[0].(engine.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return engine.NewObjectValue(b.toolToSchemaScriptValue(tool)), nil
		}

		// Get from discovery
		schema, err := b.discovery.GetToolSchema(name)
		if err != nil {
			return nil, err
		}

		return engine.NewObjectValue(toolSchemaToScriptValue(schema)), nil

	case "getToolHelp":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("getToolHelp requires name parameter as string")
		}
		name := args[0].(engine.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return engine.NewStringValue(tool.UsageInstructions()), nil
		}

		// Get from discovery
		help, err := b.discovery.GetToolHelp(name)
		if err != nil {
			return nil, err
		}

		return engine.NewStringValue(help), nil

	case "getToolExamples":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("getToolExamples requires name parameter as string")
		}
		name := args[0].(engine.StringValue).Value()

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

		result := make([]engine.ScriptValue, 0, len(examples))
		for _, ex := range examples {
			exampleData := map[string]engine.ScriptValue{
				"name":        engine.NewStringValue(ex.Name),
				"description": engine.NewStringValue(ex.Description),
				"input":       engine.ConvertToScriptValue(ex.Input),
				"output":      engine.ConvertToScriptValue(ex.Output),
			}
			result = append(result, engine.NewObjectValue(exampleData))
		}
		return engine.NewArrayValue(result), nil

	case "createTool":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("createTool requires name parameter as string")
		}
		name := args[0].(engine.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return engine.ConvertToScriptValue(toolToWrapper(name, tool)), nil
		}

		// Create from discovery
		tool, err := b.discovery.CreateTool(name)
		if err != nil {
			return nil, err
		}

		return engine.ConvertToScriptValue(toolToWrapper(name, tool)), nil

	case "executeTool":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeTool requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		params := args[1]

		// Track execution start time
		startTime := time.Now()

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

		// Update metrics
		b.updateExecutionMetrics(name, err == nil, time.Since(startTime), err)

		if err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", err)
		}

		return engine.ConvertToScriptValue(result), nil

	case "registerCustomTool":
		if len(args) < 1 {
			return nil, fmt.Errorf("registerCustomTool requires tool parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("tool must be object")
		}
		toolDefObj := args[0].(engine.ObjectValue)
		toolDef := make(map[string]interface{})
		for k, v := range toolDefObj.Fields() {
			toolDef[k] = convertScriptValueToInterface(v)
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
		return engine.NewNilValue(), nil

	// Schema validation methods (Task 1.4.9.1)
	case "executeToolValidated":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeToolValidated requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		params := convertScriptValueToInterface(args[1])

		// Track execution start time
		startTime := time.Now()

		// Get the tool
		var tool domain.Tool
		var paramSchema, outputSchema *schemaDomain.Schema

		if customTool, exists := b.customTools[name]; exists {
			tool = customTool
			paramSchema = customTool.ParameterSchema()
			outputSchema = customTool.OutputSchema()
		} else {
			var err error
			tool, err = b.discovery.CreateTool(name)
			if err != nil {
				return nil, err
			}
			// Get schemas from discovery
			if schemaInfo, err := b.discovery.GetToolSchema(name); err == nil {
				// Convert bridge.ToolSchema to domain schemas
				paramSchema = b.convertBridgeSchemaToSchema(schemaInfo.Parameters)
				outputSchema = b.convertBridgeSchemaToSchema(schemaInfo.Output)
			}
		}

		// Validate input parameters
		var inputValidation *schemaDomain.ValidationResult
		if paramSchema != nil {
			inputValidation, _ = b.validator.ValidateStruct(paramSchema, params)
			if !inputValidation.Valid {
				// Update metrics
				b.updateExecutionMetrics(name, false, time.Since(startTime), fmt.Errorf("input validation failed"))
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"success":          engine.NewBoolValue(false),
					"error":            engine.NewStringValue("Input validation failed"),
					"validationErrors": engine.ConvertToScriptValue(inputValidation.Errors),
				}), nil
			}
		}

		// Execute the tool
		toolCtx := &domain.ToolContext{
			Context: ctx,
		}
		result, err := tool.Execute(toolCtx, params)

		// Update metrics
		b.updateExecutionMetrics(name, err == nil, time.Since(startTime), err)

		if err != nil {
			return engine.NewObjectValue(map[string]engine.ScriptValue{
				"success": engine.NewBoolValue(false),
				"error":   engine.NewStringValue(err.Error()),
			}), nil
		}

		// Validate output
		var outputValidation *schemaDomain.ValidationResult
		if outputSchema != nil {
			outputValidation, _ = b.validator.ValidateStruct(outputSchema, result)
			if !outputValidation.Valid {
				// Still return the result but with validation warnings
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"success":                  engine.NewBoolValue(true),
					"result":                   engine.ConvertToScriptValue(result),
					"outputValidationWarnings": engine.ConvertToScriptValue(outputValidation.Errors),
				}), nil
			}
		}

		// Store validation report
		b.storeValidationReport(name, inputValidation, outputValidation)

		return engine.NewObjectValue(map[string]engine.ScriptValue{
			"success": engine.NewBoolValue(true),
			"result":  engine.ConvertToScriptValue(result),
		}), nil

	case "validateToolInput":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateToolInput requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		params := convertScriptValueToInterface(args[1])

		// Get parameter schema
		var paramSchema *schemaDomain.Schema
		if tool, exists := b.customTools[name]; exists {
			paramSchema = tool.ParameterSchema()
		} else {
			if schemaInfo, err := b.discovery.GetToolSchema(name); err == nil {
				paramSchema = b.convertBridgeSchemaToSchema(schemaInfo.Parameters)
			}
		}

		if paramSchema == nil {
			return engine.NewObjectValue(map[string]engine.ScriptValue{
				"valid":   engine.NewBoolValue(true),
				"message": engine.NewStringValue("No schema available for validation"),
			}), nil
		}

		result, err := b.validator.ValidateStruct(paramSchema, params)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		return engine.NewObjectValue(map[string]engine.ScriptValue{
			"valid":  engine.NewBoolValue(result.Valid),
			"errors": engine.ConvertToScriptValue(result.Errors),
		}), nil

	case "validateToolOutput":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateToolOutput requires name and output parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		output := convertScriptValueToInterface(args[1])

		// Get output schema
		var outputSchema *schemaDomain.Schema
		if tool, exists := b.customTools[name]; exists {
			outputSchema = tool.OutputSchema()
		} else {
			if schemaInfo, err := b.discovery.GetToolSchema(name); err == nil {
				outputSchema = b.convertBridgeSchemaToSchema(schemaInfo.Output)
			}
		}

		if outputSchema == nil {
			return engine.NewObjectValue(map[string]engine.ScriptValue{
				"valid":   engine.NewBoolValue(true),
				"message": engine.NewStringValue("No schema available for validation"),
			}), nil
		}

		result, err := b.validator.ValidateStruct(outputSchema, output)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		return engine.NewObjectValue(map[string]engine.ScriptValue{
			"valid":  engine.NewBoolValue(result.Valid),
			"errors": engine.ConvertToScriptValue(result.Errors),
		}), nil

	case "getValidationReport":
		if len(args) < 1 {
			return nil, fmt.Errorf("getValidationReport requires name parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()

		if report, exists := b.validationReports[name]; exists {
			return engine.NewObjectValue(map[string]engine.ScriptValue{
				"toolName":         engine.NewStringValue(report.ToolName),
				"timestamp":        engine.NewStringValue(report.Timestamp.Format(time.RFC3339)),
				"inputValidation":  engine.ConvertToScriptValue(report.InputValidation),
				"outputValidation": engine.ConvertToScriptValue(report.OutputValidation),
				"schemaIssues":     engine.ConvertToScriptValue(report.SchemaIssues),
				"recommendations":  engine.ConvertToScriptValue(report.Recommendations),
			}), nil
		}

		return nil, fmt.Errorf("no validation report found for tool: %s", name)

	// Documentation generation methods (Task 1.4.9.2)
	case "generateToolDocumentation":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateToolDocumentation requires name parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()

		format := "markdown" // default
		if len(args) > 1 {
			if args[1] != nil && args[1].Type() == engine.TypeString {
				format = args[1].(engine.StringValue).Value()
			}
		}

		// Get tool info
		var toolInfo bridge.ToolInfo
		found := false

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			toolInfo = b.customToolToToolInfo(name, tool)
			found = true
		} else {
			// Get from discovery
			tools := b.discovery.ListTools()
			for _, ti := range tools {
				if ti.Name == name {
					toolInfo = ti
					found = true
					break
				}
			}
		}

		if !found {
			return nil, fmt.Errorf("tool not found: %s", name)
		}

		// Generate documentation based on format
		switch format {
		case "markdown":
			doc, err := docs.GenerateToolMarkdown(ctx, []bridge.ToolInfo{toolInfo}, b.docConfig)
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(doc), nil
		case "openapi":
			spec, err := docs.GenerateToolOpenAPI(ctx, []bridge.ToolInfo{toolInfo}, b.docConfig)
			if err != nil {
				return nil, err
			}
			// Convert to JSON string
			jsonBytes, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(string(jsonBytes)), nil
		case "json":
			doc, err := docs.GenerateToolDocumentation(toolInfo)
			if err != nil {
				return nil, err
			}
			// Convert to JSON
			jsonBytes, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(string(jsonBytes)), nil
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}

	case "generateAllToolsDocs":
		format := "markdown" // default
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == engine.TypeString {
				format = args[0].(engine.StringValue).Value()
			}
		}

		switch format {
		case "markdown":
			doc, err := b.docGenerator.GenerateMarkdownForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(doc), nil
		case "openapi":
			spec, err := b.docGenerator.GenerateOpenAPIForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			jsonBytes, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(string(jsonBytes)), nil
		case "json":
			docs, err := b.docGenerator.GenerateDocsForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			jsonBytes, err := json.MarshalIndent(docs, "", "  ")
			if err != nil {
				return nil, err
			}
			return engine.NewStringValue(string(jsonBytes)), nil
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}

	case "generateToolPlayground":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateToolPlayground requires name parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()

		// Generate interactive HTML playground
		html, err := b.generatePlaygroundHTML(name)
		if err != nil {
			return nil, err
		}
		return engine.NewStringValue(html), nil

	case "generateSDKSnippet":
		if len(args) < 2 {
			return nil, fmt.Errorf("generateSDKSnippet requires name and language parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("language must be string")
		}
		name := args[0].(engine.StringValue).Value()
		language := args[1].(engine.StringValue).Value()

		snippet, err := b.generateSDKSnippet(name, language)
		if err != nil {
			return nil, err
		}
		return engine.NewStringValue(snippet), nil

	// Execution analytics methods (Task 1.4.9.3)
	case "getToolMetrics":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolMetrics requires name parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()

		b.metricsLock.RLock()
		metrics, exists := b.executionMetrics[name]
		b.metricsLock.RUnlock()

		if !exists {
			return nil, fmt.Errorf("no metrics found for tool: %s", name)
		}

		return engine.ConvertToScriptValue(b.metricsToMap(metrics)), nil

	case "getAllToolsMetrics":
		b.metricsLock.RLock()
		defer b.metricsLock.RUnlock()

		result := make([]engine.ScriptValue, 0, len(b.executionMetrics))
		for _, metrics := range b.executionMetrics {
			result = append(result, engine.ConvertToScriptValue(b.metricsToMap(metrics)))
		}

		return engine.NewArrayValue(result), nil

	case "getToolUsageReport":
		period := "day" // default
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == engine.TypeString {
				period = args[0].(engine.StringValue).Value()
			}
		}

		report, err := b.generateUsageReport(period)
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(report), nil

	case "enableToolProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("enableToolProfiling requires name parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}

		// Enable profiling for all tools
		b.profiler.Enable()
		return engine.NewNilValue(), nil

	case "getToolAnomalies":
		var toolName string
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == engine.TypeString {
				toolName = args[0].(engine.StringValue).Value()
			}
		}

		anomalies, err := b.detectAnomalies(toolName)
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(anomalies), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper functions

func toolInfoToScriptValue(info bridge.ToolInfo) map[string]engine.ScriptValue {
	result := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue(info.Name),
		"description": engine.NewStringValue(info.Description),
		"category":    engine.NewStringValue(info.Category),
		"tags":        convertStringSliceToScriptValue(info.Tags),
		"version":     engine.NewStringValue(info.Version),
		"usageHint":   engine.NewStringValue(info.UsageHint),
		"package":     engine.NewStringValue(info.Package),
	}

	// Parse schemas if available
	if len(info.ParameterSchema) > 0 {
		var params interface{}
		if err := json.Unmarshal(info.ParameterSchema, &params); err == nil {
			result["parameterSchema"] = engine.ConvertToScriptValue(params)
		}
	}

	if len(info.OutputSchema) > 0 {
		var output interface{}
		if err := json.Unmarshal(info.OutputSchema, &output); err == nil {
			result["outputSchema"] = engine.ConvertToScriptValue(output)
		}
	}

	return result
}

func convertStringSliceToScriptValue(slice []string) engine.ScriptValue {
	values := make([]engine.ScriptValue, len(slice))
	for i, s := range slice {
		values[i] = engine.NewStringValue(s)
	}
	return engine.NewArrayValue(values)
}

// convertScriptValueToInterface converts ScriptValue to interface{} for go-llms compatibility
func convertScriptValueToInterface(v engine.ScriptValue) interface{} {
	switch v.Type() {
	case engine.TypeString:
		return v.(engine.StringValue).Value()
	case engine.TypeNumber:
		return v.(engine.NumberValue).Value()
	case engine.TypeBool:
		return v.(engine.BoolValue).Value()
	case engine.TypeNil:
		return nil
	case engine.TypeArray:
		arr := v.(engine.ArrayValue).Elements()
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			result[i] = convertScriptValueToInterface(item)
		}
		return result
	case engine.TypeObject:
		obj := v.(engine.ObjectValue).Fields()
		result := make(map[string]interface{})
		for k, val := range obj {
			result[k] = convertScriptValueToInterface(val)
		}
		return result
	default:
		return v.ToGo()
	}
}

func toolSchemaToScriptValue(schema *bridge.ToolSchema) map[string]engine.ScriptValue {
	return map[string]engine.ScriptValue{
		"name":          engine.NewStringValue(schema.Name),
		"description":   engine.NewStringValue(schema.Description),
		"parameters":    engine.ConvertToScriptValue(schema.Parameters),
		"output":        engine.ConvertToScriptValue(schema.Output),
		"examples":      engine.ConvertToScriptValue(schema.Examples),
		"constraints":   engine.ConvertToScriptValue(schema.Constraints),
		"errorGuidance": engine.ConvertToScriptValue(schema.ErrorGuidance),
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

func customToolToScriptValue(name string, tool domain.Tool) map[string]engine.ScriptValue {
	return map[string]engine.ScriptValue{
		"name":                 engine.NewStringValue(name),
		"description":          engine.NewStringValue(tool.Description()),
		"category":             engine.NewStringValue(tool.Category()),
		"tags":                 convertStringSliceToScriptValue(tool.Tags()),
		"version":              engine.NewStringValue(tool.Version()),
		"custom":               engine.NewBoolValue(true),
		"isDeterministic":      engine.NewBoolValue(tool.IsDeterministic()),
		"isDestructive":        engine.NewBoolValue(tool.IsDestructive()),
		"requiresConfirmation": engine.NewBoolValue(tool.RequiresConfirmation()),
		"estimatedLatency":     engine.NewStringValue(tool.EstimatedLatency()),
		"usageInstructions":    engine.NewStringValue(tool.UsageInstructions()),
		"constraints":          engine.ConvertToScriptValue(tool.Constraints()),
		"errorGuidance":        engine.ConvertToScriptValue(tool.ErrorGuidance()),
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
func (b *ToolsBridge) toolToSchemaScriptValue(tool domain.Tool) map[string]engine.ScriptValue {
	schema := &bridge.ToolSchema{
		Name:          tool.Name(),
		Description:   tool.Description(),
		Parameters:    tool.ParameterSchema(),
		Output:        tool.OutputSchema(),
		Examples:      tool.Examples(),
		Constraints:   tool.Constraints(),
		ErrorGuidance: tool.ErrorGuidance(),
	}
	return toolSchemaToScriptValue(schema)
}

// Helper methods for enhanced features

// updateExecutionMetrics updates execution metrics for a tool
func (b *ToolsBridge) updateExecutionMetrics(toolName string, success bool, duration time.Duration, err error) {
	b.metricsLock.Lock()
	defer b.metricsLock.Unlock()

	metrics, exists := b.executionMetrics[toolName]
	if !exists {
		metrics = &ExecutionMetrics{
			ToolName:       toolName,
			ErrorTypes:     make(map[string]int),
			ParameterStats: make(map[string]interface{}),
			MinDuration:    duration,
			MaxDuration:    duration,
		}
		b.executionMetrics[toolName] = metrics
	}

	// Update counters
	metrics.TotalExecutions++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
		// Track error types
		if err != nil {
			errorType := fmt.Sprintf("%T", err)
			metrics.ErrorTypes[errorType]++
		}
	}

	// Update durations
	metrics.TotalDuration += duration
	metrics.AverageDuration = metrics.TotalDuration / time.Duration(metrics.TotalExecutions)
	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}
	metrics.LastExecution = time.Now()
}

// storeValidationReport stores a validation report for a tool
func (b *ToolsBridge) storeValidationReport(toolName string, inputValidation, outputValidation *schemaDomain.ValidationResult) {
	report := &ValidationReport{
		ToolName:         toolName,
		Timestamp:        time.Now(),
		InputValidation:  inputValidation,
		OutputValidation: outputValidation,
		SchemaIssues:     []string{},
		Recommendations:  []string{},
	}

	// Add recommendations based on validation results
	if inputValidation != nil && !inputValidation.Valid {
		report.Recommendations = append(report.Recommendations,
			"Consider updating input schema to better match usage patterns")
	}
	if outputValidation != nil && !outputValidation.Valid {
		report.Recommendations = append(report.Recommendations,
			"Output schema may need adjustment to reflect actual tool outputs")
	}

	b.validationReports[toolName] = report
}

// convertBridgeSchemaToSchema converts bridge.ToolSchema fields to domain.Schema
func (b *ToolsBridge) convertBridgeSchemaToSchema(schemaData interface{}) *schemaDomain.Schema {
	if schemaData == nil {
		return nil
	}

	// If it's already a schema, return it
	if schema, ok := schemaData.(*schemaDomain.Schema); ok {
		return schema
	}

	// Try to convert from map
	if schemaMap, ok := schemaData.(map[string]interface{}); ok {
		schema, err := b.convertToSchema(schemaMap)
		if err != nil {
			return nil
		}
		return schema
	}

	// Try to convert from JSON
	if jsonData, ok := schemaData.(json.RawMessage); ok {
		var schema schemaDomain.Schema
		if err := json.Unmarshal(jsonData, &schema); err == nil {
			return &schema
		}
	}

	return nil
}

// customToolToToolInfo converts a custom tool to ToolInfo
func (b *ToolsBridge) customToolToToolInfo(name string, tool domain.Tool) bridge.ToolInfo {
	info := bridge.ToolInfo{
		Name:        name,
		Description: tool.Description(),
		Category:    tool.Category(),
		Tags:        tool.Tags(),
		Version:     tool.Version(),
		UsageHint:   tool.UsageInstructions(),
	}

	// Convert schemas to JSON
	if paramSchema := tool.ParameterSchema(); paramSchema != nil {
		if jsonData, err := json.Marshal(paramSchema); err == nil {
			info.ParameterSchema = jsonData
		}
	}

	if outputSchema := tool.OutputSchema(); outputSchema != nil {
		if jsonData, err := json.Marshal(outputSchema); err == nil {
			info.OutputSchema = jsonData
		}
	}

	// Convert examples
	examples := tool.Examples()
	info.Examples = make([]tools.Example, len(examples))
	for i, ex := range examples {
		inputJSON, _ := json.Marshal(ex.Input)
		outputJSON, _ := json.Marshal(ex.Output)
		info.Examples[i] = tools.Example{
			Name:        ex.Name,
			Description: ex.Description,
			Input:       inputJSON,
			Output:      outputJSON,
		}
	}

	return info
}

// metricsToMap converts ExecutionMetrics to a map for script consumption
func (b *ToolsBridge) metricsToMap(metrics *ExecutionMetrics) map[string]interface{} {
	return map[string]interface{}{
		"toolName":        metrics.ToolName,
		"totalExecutions": metrics.TotalExecutions,
		"successCount":    metrics.SuccessCount,
		"failureCount":    metrics.FailureCount,
		"successRate":     float64(metrics.SuccessCount) / float64(metrics.TotalExecutions),
		"totalDuration":   metrics.TotalDuration.String(),
		"averageDuration": metrics.AverageDuration.String(),
		"minDuration":     metrics.MinDuration.String(),
		"maxDuration":     metrics.MaxDuration.String(),
		"lastExecution":   metrics.LastExecution.Format(time.RFC3339),
		"errorTypes":      metrics.ErrorTypes,
		"parameterStats":  metrics.ParameterStats,
	}
}

// generateUsageReport generates a usage report for tools
func (b *ToolsBridge) generateUsageReport(period string) (map[string]interface{}, error) {
	b.metricsLock.RLock()
	defer b.metricsLock.RUnlock()

	// Calculate time window based on period
	var since time.Time
	now := time.Now()
	switch period {
	case "hour":
		since = now.Add(-time.Hour)
	case "day":
		since = now.Add(-24 * time.Hour)
	case "week":
		since = now.Add(-7 * 24 * time.Hour)
	case "month":
		since = now.Add(-30 * 24 * time.Hour)
	default:
		since = now.Add(-24 * time.Hour) // Default to day
	}

	// Aggregate metrics
	totalExecutions := int64(0)
	totalSuccess := int64(0)
	totalFailure := int64(0)
	toolUsage := make([]map[string]interface{}, 0)

	for _, metrics := range b.executionMetrics {
		if metrics.LastExecution.After(since) {
			totalExecutions += metrics.TotalExecutions
			totalSuccess += metrics.SuccessCount
			totalFailure += metrics.FailureCount

			toolUsage = append(toolUsage, map[string]interface{}{
				"toolName":    metrics.ToolName,
				"executions":  metrics.TotalExecutions,
				"successRate": float64(metrics.SuccessCount) / float64(metrics.TotalExecutions),
				"avgDuration": metrics.AverageDuration.String(),
			})
		}
	}

	return map[string]interface{}{
		"period":          period,
		"since":           since.Format(time.RFC3339),
		"totalExecutions": totalExecutions,
		"totalSuccess":    totalSuccess,
		"totalFailure":    totalFailure,
		"overallSuccessRate": func() float64 {
			if totalExecutions == 0 {
				return 0
			}
			return float64(totalSuccess) / float64(totalExecutions)
		}(),
		"toolUsage": toolUsage,
	}, nil
}

// detectAnomalies detects anomalies in tool execution
func (b *ToolsBridge) detectAnomalies(toolName string) ([]map[string]interface{}, error) {
	b.metricsLock.RLock()
	defer b.metricsLock.RUnlock()

	anomalies := make([]map[string]interface{}, 0)

	checkToolAnomalies := func(metrics *ExecutionMetrics) {
		// Check for high failure rate
		if metrics.TotalExecutions > 10 {
			failureRate := float64(metrics.FailureCount) / float64(metrics.TotalExecutions)
			if failureRate > 0.3 { // More than 30% failure
				anomalies = append(anomalies, map[string]interface{}{
					"toolName": metrics.ToolName,
					"type":     "high_failure_rate",
					"severity": "warning",
					"details": map[string]interface{}{
						"failureRate": failureRate,
						"failures":    metrics.FailureCount,
						"total":       metrics.TotalExecutions,
					},
				})
			}
		}

		// Check for performance degradation
		if metrics.TotalExecutions > 5 && metrics.MaxDuration > 10*metrics.AverageDuration {
			anomalies = append(anomalies, map[string]interface{}{
				"toolName": metrics.ToolName,
				"type":     "performance_outlier",
				"severity": "info",
				"details": map[string]interface{}{
					"maxDuration": metrics.MaxDuration.String(),
					"avgDuration": metrics.AverageDuration.String(),
					"ratio":       float64(metrics.MaxDuration) / float64(metrics.AverageDuration),
				},
			})
		}

		// Check for recent spike in errors
		recentWindow := time.Now().Add(-time.Hour)
		if metrics.LastExecution.After(recentWindow) && metrics.FailureCount > metrics.SuccessCount {
			anomalies = append(anomalies, map[string]interface{}{
				"toolName": metrics.ToolName,
				"type":     "error_spike",
				"severity": "critical",
				"details": map[string]interface{}{
					"recentFailures":  metrics.FailureCount,
					"recentSuccesses": metrics.SuccessCount,
					"errorTypes":      metrics.ErrorTypes,
				},
			})
		}
	}

	if toolName != "" {
		// Check specific tool
		if metrics, exists := b.executionMetrics[toolName]; exists {
			checkToolAnomalies(metrics)
		}
	} else {
		// Check all tools
		for _, metrics := range b.executionMetrics {
			checkToolAnomalies(metrics)
		}
	}

	return anomalies, nil
}

// generatePlaygroundHTML generates an interactive HTML playground for a tool
func (b *ToolsBridge) generatePlaygroundHTML(toolName string) (string, error) {
	// Get tool info
	var toolInfo bridge.ToolInfo
	var paramSchema interface{}

	if tool, exists := b.customTools[toolName]; exists {
		toolInfo = b.customToolToToolInfo(toolName, tool)
		if ps := tool.ParameterSchema(); ps != nil {
			paramSchema = ps
		}
	} else {
		tools := b.discovery.ListTools()
		found := false
		for _, ti := range tools {
			if ti.Name == toolName {
				toolInfo = ti
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("tool not found: %s", toolName)
		}

		// Get parameter schema
		if schemaInfo, err := b.discovery.GetToolSchema(toolName); err == nil && schemaInfo.Parameters != nil {
			paramSchema = schemaInfo.Parameters
		}
	}

	// Generate HTML with form based on schema
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s Tool Playground</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .tool-info { background: #f0f0f0; padding: 20px; margin-bottom: 20px; border-radius: 5px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, textarea, select { width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 3px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
        button:hover { background: #0056b3; }
        .output { background: #f8f9fa; padding: 15px; margin-top: 20px; border-radius: 5px; white-space: pre-wrap; }
        .error { color: #dc3545; }
        .success { color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s Tool Playground</h1>
        <div class="tool-info">
            <h2>Description</h2>
            <p>%s</p>
            <p><strong>Category:</strong> %s</p>
            <p><strong>Version:</strong> %s</p>
        </div>
        
        <form id="toolForm">
            <h2>Parameters</h2>
            <div id="formFields"></div>
            <button type="submit">Execute Tool</button>
        </form>
        
        <div id="output" class="output" style="display:none;"></div>
    </div>
    
    <script>
        const paramSchema = %s;
        
        // Generate form fields based on schema
        function generateFormFields(schema, container, prefix = '') {
            if (!schema || !schema.properties) return;
            
            for (const [key, prop] of Object.entries(schema.properties)) {
                const fieldId = prefix + key;
                const div = document.createElement('div');
                div.className = 'form-group';
                
                const label = document.createElement('label');
                label.textContent = key + (schema.required && schema.required.includes(key) ? ' *' : '');
                label.htmlFor = fieldId;
                div.appendChild(label);
                
                let input;
                switch (prop.type) {
                    case 'string':
                        if (prop.enum) {
                            input = document.createElement('select');
                            input.id = fieldId;
                            prop.enum.forEach(val => {
                                const option = document.createElement('option');
                                option.value = val;
                                option.textContent = val;
                                input.appendChild(option);
                            });
                        } else {
                            input = document.createElement('input');
                            input.type = 'text';
                            input.id = fieldId;
                        }
                        break;
                    case 'number':
                    case 'integer':
                        input = document.createElement('input');
                        input.type = 'number';
                        input.id = fieldId;
                        if (prop.minimum !== undefined) input.min = prop.minimum;
                        if (prop.maximum !== undefined) input.max = prop.maximum;
                        break;
                    case 'boolean':
                        input = document.createElement('input');
                        input.type = 'checkbox';
                        input.id = fieldId;
                        break;
                    case 'array':
                    case 'object':
                        input = document.createElement('textarea');
                        input.id = fieldId;
                        input.placeholder = 'Enter JSON';
                        input.rows = 5;
                        break;
                    default:
                        input = document.createElement('input');
                        input.type = 'text';
                        input.id = fieldId;
                }
                
                if (prop.description) {
                    const desc = document.createElement('small');
                    desc.textContent = prop.description;
                    desc.style.color = '#666';
                    div.appendChild(desc);
                    div.appendChild(document.createElement('br'));
                }
                
                div.appendChild(input);
                container.appendChild(div);
            }
        }
        
        // Initialize form
        const formFields = document.getElementById('formFields');
        generateFormFields(paramSchema, formFields);
        
        // Handle form submission
        document.getElementById('toolForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const params = {};
            const inputs = formFields.querySelectorAll('input, textarea, select');
            
            inputs.forEach(input => {
                const key = input.id;
                let value = input.value;
                
                if (input.type === 'checkbox') {
                    value = input.checked;
                } else if (input.type === 'number') {
                    value = parseFloat(value);
                } else if (input.tagName === 'TEXTAREA') {
                    try {
                        value = JSON.parse(value);
                    } catch (e) {
                        // Keep as string if not valid JSON
                    }
                }
                
                if (value !== '' && value !== null) {
                    params[key] = value;
                }
            });
            
            const output = document.getElementById('output');
            output.style.display = 'block';
            output.innerHTML = '<div>Executing tool...</div>';
            
            // Note: In a real implementation, this would call your tool execution endpoint
            output.innerHTML = '<div class="success">Tool execution would happen here with parameters:</div>' +
                              '<pre>' + JSON.stringify(params, null, 2) + '</pre>';
        });
    </script>
</body>
</html>`, toolInfo.Name, toolInfo.Name, toolInfo.Description, toolInfo.Category, toolInfo.Version,
		func() string {
			if paramSchema != nil {
				jsonBytes, _ := json.Marshal(paramSchema)
				return string(jsonBytes)
			}
			return "{}"
		}())

	return html, nil
}

// generateSDKSnippet generates SDK code snippets for tool usage
func (b *ToolsBridge) generateSDKSnippet(toolName string, language string) (string, error) {
	// Get tool info
	var toolInfo bridge.ToolInfo
	found := false

	if tool, exists := b.customTools[toolName]; exists {
		toolInfo = b.customToolToToolInfo(toolName, tool)
		found = true
	} else {
		tools := b.discovery.ListTools()
		for _, ti := range tools {
			if ti.Name == toolName {
				toolInfo = ti
				found = true
				break
			}
		}
	}

	if !found {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	// Generate code snippet based on language
	switch language {
	case "go":
		return fmt.Sprintf(`// Using %s tool
import (
    "context"
    "fmt"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
)

// Create tool discovery
discovery := tools.NewDiscovery()

// Create the tool
tool, err := discovery.CreateTool("%s")
if err != nil {
    return fmt.Errorf("failed to create tool: %%w", err)
}

// Prepare parameters
params := map[string]interface{}{
    // Add your parameters here based on the tool's schema
}

// Create tool context
ctx := &domain.ToolContext{
    Context: context.Background(),
}

// Execute the tool
result, err := tool.Execute(ctx, params)
if err != nil {
    return fmt.Errorf("tool execution failed: %%w", err)
}

fmt.Printf("Result: %%v\n", result)`, toolInfo.Name, toolInfo.Name), nil

	case "python":
		return fmt.Sprintf(`# Using %s tool
from go_llms import ToolDiscovery, ToolContext

# Create tool discovery
discovery = ToolDiscovery()

# Create the tool
tool = discovery.create_tool("%s")

# Prepare parameters
params = {
    # Add your parameters here based on the tool's schema
}

# Create tool context
ctx = ToolContext()

# Execute the tool
try:
    result = tool.execute(ctx, params)
    print(f"Result: {result}")
except Exception as e:
    print(f"Tool execution failed: {e}")`, toolInfo.Name, toolInfo.Name), nil

	case "javascript":
		return fmt.Sprintf(`// Using %s tool
const { ToolDiscovery, ToolContext } = require('go-llms');

async function useTool() {
    // Create tool discovery
    const discovery = new ToolDiscovery();
    
    // Create the tool
    const tool = await discovery.createTool("%s");
    
    // Prepare parameters
    const params = {
        // Add your parameters here based on the tool's schema
    };
    
    // Create tool context
    const ctx = new ToolContext();
    
    try {
        // Execute the tool
        const result = await tool.execute(ctx, params);
        console.log('Result:', result);
    } catch (error) {
        console.error('Tool execution failed:', error);
    }
}

useTool();`, toolInfo.Name, toolInfo.Name), nil

	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}
