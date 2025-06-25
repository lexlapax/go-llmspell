// ABOUTME: Tools bridge providing access to go-llms tool discovery system
// ABOUTME: Wraps go-llms tool discovery API for dynamic tool exploration and execution

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"

	// go-llms imports for tool functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	builtintools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/docs"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// ToolsBridge provides access to go-llms tool discovery system with v2.0.0 enhancements.
// It bridges tool functionality to scripts, enabling discovery, validation,
// documentation generation, and execution analytics for agent tools.
type ToolsBridge struct {
	mu          sync.RWMutex
	initialized bool
	discovery   types.ToolDiscovery
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

// ValidationReport stores detailed validation results for a tool.
// It includes input/output validation results, schema issues,
// and recommendations for fixing validation problems.
type ValidationReport struct {
	ToolName         string                         `json:"toolName"`
	Timestamp        time.Time                      `json:"timestamp"`
	InputValidation  *schemaDomain.ValidationResult `json:"inputValidation,omitempty"`
	OutputValidation *schemaDomain.ValidationResult `json:"outputValidation,omitempty"`
	SchemaIssues     []string                       `json:"schemaIssues,omitempty"`
	Recommendations  []string                       `json:"recommendations,omitempty"`
}

// ExecutionMetrics tracks tool execution statistics.
// It provides comprehensive metrics including execution counts,
// durations, error types, and parameter usage statistics.
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

// NewToolsBridge creates a new tools types.
// It initializes an empty tools bridge ready for configuration
// with tool discovery, validation, and documentation systems.
func NewToolsBridge() *ToolsBridge {
	return &ToolsBridge{}
}

// GetID returns the bridge ID.
// It implements the types.Bridge interface.
func (b *ToolsBridge) GetID() string {
	return "agent_tools"
}

// GetMetadata returns bridge metadata.
// It provides information about the enhanced tools bridge including
// version, description, and supported features.
func (b *ToolsBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "agent_tools",
		Version:     "2.1.0",
		Description: "Enhanced tools bridge with schema validation, documentation generation, and execution analytics (v0.3.5)",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the types.
// It sets up tool discovery, schema validation, documentation generation,
// and execution analytics systems.
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

// Cleanup performs cleanup.
// It releases all resources and resets the bridge to an uninitialized state.
func (b *ToolsBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	b.discovery = nil
	b.customTools = nil
	return nil
}

// IsInitialized checks if the bridge is initialized.
// It returns true if the bridge has been initialized and is ready for use.
func (b *ToolsBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// It enables the script engine to access tool functionality through this types.
func (b *ToolsBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this types.
// It provides metadata about all tool-related methods available to scripts,
// including discovery, validation, documentation, and execution operations.
func (b *ToolsBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Tool discovery methods
		{
			Name:        "listTools",
			Description: "List all available tools with metadata",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "searchTools",
			Description: "Search tools by keyword in name, description, or tags",
			Parameters: []types.ParameterInfo{
				{Name: "query", Type: "string", Description: "Search query", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "listByCategory",
			Description: "List tools in a specific category",
			Parameters: []types.ParameterInfo{
				{Name: "category", Type: "string", Description: "Tool category", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getToolInfo",
			Description: "Get detailed information about a specific tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getToolSchema",
			Description: "Get parameter and output schemas for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getToolHelp",
			Description: "Get help text for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "getToolExamples",
			Description: "Get usage examples for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "array",
		},
		// Tool creation and execution
		{
			Name:        "createTool",
			Description: "Create a tool instance by name",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "executeTool",
			Description: "Execute a tool with parameters",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Tool parameters", Required: true},
			},
			ReturnType: "any",
		},
		// Custom tool registration
		{
			Name:        "registerCustomTool",
			Description: "Register a custom tool implementation",
			Parameters: []types.ParameterInfo{
				{Name: "tool", Type: "object", Description: "Tool definition", Required: true},
			},
			ReturnType: "void",
		},
		// Schema validation methods (Task 1.4.9.1)
		{
			Name:        "executeToolValidated",
			Description: "Execute a tool with schema validation of inputs and outputs",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Tool parameters", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateToolInput",
			Description: "Validate tool input parameters against schema",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "params", Type: "object", Description: "Parameters to validate", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateToolOutput",
			Description: "Validate tool output against schema",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "output", Type: "any", Description: "Output to validate", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getValidationReport",
			Description: "Get validation report for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		// Documentation generation methods (Task 1.4.9.2)
		{
			Name:        "generateToolDocumentation",
			Description: "Generate comprehensive documentation for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "format", Type: "string", Description: "Documentation format (markdown, openapi, json)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateAllToolsDocs",
			Description: "Generate documentation for all available tools",
			Parameters: []types.ParameterInfo{
				{Name: "format", Type: "string", Description: "Documentation format", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateToolPlayground",
			Description: "Generate interactive playground HTML for a tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateSDKSnippet",
			Description: "Generate SDK code snippet for tool usage",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
				{Name: "language", Type: "string", Description: "Programming language (go, python, javascript)", Required: true},
			},
			ReturnType: "string",
		},
		// Execution analytics methods (Task 1.4.9.3)
		{
			Name:        "getToolMetrics",
			Description: "Get execution metrics for a specific tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getAllToolsMetrics",
			Description: "Get execution metrics for all tools",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getToolUsageReport",
			Description: "Generate usage report for tools",
			Parameters: []types.ParameterInfo{
				{Name: "period", Type: "string", Description: "Time period (hour, day, week, month)", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "enableToolProfiling",
			Description: "Enable profiling for a specific tool",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getToolAnomalies",
			Description: "Get anomaly alerts for tool execution",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Tool name", Required: false},
			},
			ReturnType: "array",
		},
	}
}

// TypeMappings returns type conversion mappings.
// It defines how Go types are mapped to script types for tools,
// validation results, execution metrics, and documentation.
func (b *ToolsBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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

// ValidateMethod validates method calls.
// It ensures that each method receives the correct number and
// types of arguments before execution.
func (b *ToolsBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	switch name {
	case "listTools", "getAllToolsMetrics", "generateAllToolsDocs":
		// No arguments required
		return nil
	case "searchTools", "listByCategory", "getToolInfo", "getToolSchema", "getToolHelp", "getToolExamples", "createTool", "getToolMetrics", "enableToolProfiling":
		if len(args) < 1 {
			return fmt.Errorf("%s requires at least one argument", name)
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return fmt.Errorf("%s requires first argument to be string", name)
		}
		return nil
	case "executeTool", "executeToolValidated", "validateToolInput", "validateToolOutput":
		if len(args) < 2 {
			return fmt.Errorf("%s requires at least two arguments", name)
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return fmt.Errorf("%s requires first argument to be string", name)
		}
		if args[1] == nil || args[1].Type() != types.TypeObject {
			return fmt.Errorf("%s requires second argument to be object", name)
		}
		return nil
	case "registerCustomTool":
		if len(args) < 1 {
			return fmt.Errorf("registerCustomTool requires tool definition")
		}
		if args[0] == nil || args[0].Type() != types.TypeObject {
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

// RequiredPermissions returns required permissions.
// It specifies the permissions needed for tool execution,
// including process, file system, and network access.
func (b *ToolsBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionProcess,
			Resource:    "tool",
			Actions:     []string{"execute", "register", "list"},
			Description: "Tool execution and management",
		},
		{
			Type:        types.PermissionFileSystem,
			Resource:    "*",
			Actions:     []string{"read", "write"},
			Description: "File system access for file tools",
		},
		{
			Type:        types.PermissionNetwork,
			Resource:    "*",
			Actions:     []string{"http"},
			Description: "Network access for web tools",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function.
// It implements the types.Bridge interface, routing method calls to tool discovery,
// validation, documentation, and execution functionality.
func (b *ToolsBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "listTools":
		tools := b.discovery.ListTools()
		result := make([]types.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, types.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return types.NewArrayValue(result), nil

	case "searchTools":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("searchTools requires query parameter as string")
		}
		query := args[0].(types.StringValue).Value()

		tools := b.discovery.SearchTools(query)
		result := make([]types.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, types.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return types.NewArrayValue(result), nil

	case "listByCategory":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("listByCategory requires category parameter as string")
		}
		category := args[0].(types.StringValue).Value()

		tools := b.discovery.ListByCategory(category)
		result := make([]types.ScriptValue, 0, len(tools))
		for _, tool := range tools {
			result = append(result, types.NewObjectValue(toolInfoToScriptValue(tool)))
		}
		return types.NewArrayValue(result), nil

	case "getToolInfo":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("getToolInfo requires name parameter as string")
		}
		name := args[0].(types.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return types.NewObjectValue(customToolToScriptValue(name, tool)), nil
		}

		// Get from discovery
		tools := b.discovery.ListTools()
		for _, tool := range tools {
			if tool.Name == name {
				return types.NewObjectValue(toolInfoToScriptValue(tool)), nil
			}
		}
		return nil, fmt.Errorf("tool not found: %s", name)

	case "getToolSchema":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("getToolSchema requires name parameter as string")
		}
		name := args[0].(types.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return types.NewObjectValue(b.toolToSchemaScriptValue(tool)), nil
		}

		// Get from discovery
		schema, err := b.discovery.GetToolSchema(name)
		if err != nil {
			return nil, err
		}

		return types.NewObjectValue(toolSchemaToScriptValue(schema)), nil

	case "getToolHelp":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("getToolHelp requires name parameter as string")
		}
		name := args[0].(types.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return types.NewStringValue(tool.UsageInstructions()), nil
		}

		// Get from discovery
		help, err := b.discovery.GetToolHelp(name)
		if err != nil {
			return nil, err
		}

		return types.NewStringValue(help), nil

	case "getToolExamples":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("getToolExamples requires name parameter as string")
		}
		name := args[0].(types.StringValue).Value()

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

		result := make([]types.ScriptValue, 0, len(examples))
		for _, ex := range examples {
			exampleData := map[string]types.ScriptValue{
				"name":        types.NewStringValue(ex.Name),
				"description": types.NewStringValue(ex.Description),
				"input":       types.ConvertToScriptValue(ex.Input),
				"output":      types.ConvertToScriptValue(ex.Output),
			}
			result = append(result, types.NewObjectValue(exampleData))
		}
		return types.NewArrayValue(result), nil

	case "createTool":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("createTool requires name parameter as string")
		}
		name := args[0].(types.StringValue).Value()

		// Check custom tools first
		if tool, exists := b.customTools[name]; exists {
			return types.ConvertToScriptValue(toolToWrapper(name, tool)), nil
		}

		// Try getting from tools registry first (where real tools are registered)
		if tool, found := builtintools.GetTool(name); found {
			return types.ConvertToScriptValue(toolToWrapper(name, tool)), nil
		}

		// Fall back to discovery (which might have placeholders)
		tool, err := b.discovery.CreateTool(name)
		if err != nil {
			return nil, err
		}

		return types.ConvertToScriptValue(toolToWrapper(name, tool)), nil

	case "executeTool":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeTool requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()
		params := args[1]

		// Track execution start time
		startTime := time.Now()

		var tool domain.Tool

		// Check custom tools first
		if customTool, exists := b.customTools[name]; exists {
			tool = customTool
		} else {
			// Try getting from tools registry first (where real tools are registered)
			if registryTool, found := builtintools.GetTool(name); found {
				tool = registryTool
			} else {
				// Fall back to discovery (which might have placeholders)
				var err error
				tool, err = b.discovery.CreateTool(name)
				if err != nil {
					// If it's a "not yet loaded" error, check registry again
					if err.Error() == fmt.Sprintf("tool %s not yet loaded - import the tool package to use it", name) {
						if registryTool, found := builtintools.GetTool(name); found {
							tool = registryTool
						} else {
							return nil, err
						}
					} else {
						return nil, err
					}
				}
			}
		}

		// Create a basic tool context
		toolCtx := &domain.ToolContext{
			Context: ctx,
		}

		// Convert ScriptValue params to native Go types
		nativeParams := convertScriptValueToInterface(params)
		
		// Execute the tool
		result, err := tool.Execute(toolCtx, nativeParams)

		// Update metrics
		b.updateExecutionMetrics(name, err == nil, time.Since(startTime), err)

		if err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", err)
		}

		return types.ConvertToScriptValue(result), nil

	case "registerCustomTool":
		if len(args) < 1 {
			return nil, fmt.Errorf("registerCustomTool requires tool parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeObject {
			return nil, fmt.Errorf("tool must be object, got: %v", args[0].Type())
		}
		toolDefObj := args[0].(types.ObjectValue)
		toolDef := make(map[string]interface{})
		fields := toolDefObj.Fields()
		
		for k, v := range fields {
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
		return types.NewBoolValue(true), nil

	// Schema validation methods (Task 1.4.9.1)
	case "executeToolValidated":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeToolValidated requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()
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
				// Convert types.ToolSchema to domain schemas
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
				return types.NewObjectValue(map[string]types.ScriptValue{
					"success":          types.NewBoolValue(false),
					"error":            types.NewStringValue("Input validation failed"),
					"validationErrors": types.ConvertToScriptValue(inputValidation.Errors),
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
			return types.NewObjectValue(map[string]types.ScriptValue{
				"success": types.NewBoolValue(false),
				"error":   types.NewStringValue(err.Error()),
			}), nil
		}

		// Validate output
		var outputValidation *schemaDomain.ValidationResult
		if outputSchema != nil {
			outputValidation, _ = b.validator.ValidateStruct(outputSchema, result)
			if !outputValidation.Valid {
				// Still return the result but with validation warnings
				return types.NewObjectValue(map[string]types.ScriptValue{
					"success":                  types.NewBoolValue(true),
					"result":                   types.ConvertToScriptValue(result),
					"outputValidationWarnings": types.ConvertToScriptValue(outputValidation.Errors),
				}), nil
			}
		}

		// Store validation report
		b.storeValidationReport(name, inputValidation, outputValidation)

		return types.NewObjectValue(map[string]types.ScriptValue{
			"success": types.NewBoolValue(true),
			"result":  types.ConvertToScriptValue(result),
		}), nil

	case "validateToolInput":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateToolInput requires name and params parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()
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
			return types.NewObjectValue(map[string]types.ScriptValue{
				"valid":   types.NewBoolValue(true),
				"message": types.NewStringValue("No schema available for validation"),
			}), nil
		}

		result, err := b.validator.ValidateStruct(paramSchema, params)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		return types.NewObjectValue(map[string]types.ScriptValue{
			"valid":  types.NewBoolValue(result.Valid),
			"errors": types.ConvertToScriptValue(result.Errors),
		}), nil

	case "validateToolOutput":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateToolOutput requires name and output parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()
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
			return types.NewObjectValue(map[string]types.ScriptValue{
				"valid":   types.NewBoolValue(true),
				"message": types.NewStringValue("No schema available for validation"),
			}), nil
		}

		result, err := b.validator.ValidateStruct(outputSchema, output)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		return types.NewObjectValue(map[string]types.ScriptValue{
			"valid":  types.NewBoolValue(result.Valid),
			"errors": types.ConvertToScriptValue(result.Errors),
		}), nil

	case "getValidationReport":
		if len(args) < 1 {
			return nil, fmt.Errorf("getValidationReport requires name parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()

		if report, exists := b.validationReports[name]; exists {
			return types.NewObjectValue(map[string]types.ScriptValue{
				"toolName":         types.NewStringValue(report.ToolName),
				"timestamp":        types.NewStringValue(report.Timestamp.Format(time.RFC3339)),
				"inputValidation":  types.ConvertToScriptValue(report.InputValidation),
				"outputValidation": types.ConvertToScriptValue(report.OutputValidation),
				"schemaIssues":     types.ConvertToScriptValue(report.SchemaIssues),
				"recommendations":  types.ConvertToScriptValue(report.Recommendations),
			}), nil
		}

		return nil, fmt.Errorf("no validation report found for tool: %s", name)

	// Documentation generation methods (Task 1.4.9.2)
	case "generateToolDocumentation":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateToolDocumentation requires name parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()

		format := "markdown" // default
		if len(args) > 1 {
			if args[1] != nil && args[1].Type() == types.TypeString {
				format = args[1].(types.StringValue).Value()
			}
		}

		// Get tool info
		var toolInfo types.ToolInfo
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
			doc, err := docs.GenerateToolMarkdown(ctx, []types.ToolInfo{toolInfo}, b.docConfig)
			if err != nil {
				return nil, err
			}
			return types.NewStringValue(doc), nil
		case "openapi":
			spec, err := docs.GenerateToolOpenAPI(ctx, []types.ToolInfo{toolInfo}, b.docConfig)
			if err != nil {
				return nil, err
			}
			// Convert to JSON string
			jsonBytes, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return nil, err
			}
			return types.NewStringValue(string(jsonBytes)), nil
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
			return types.NewStringValue(string(jsonBytes)), nil
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}

	case "generateAllToolsDocs":
		format := "markdown" // default
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == types.TypeString {
				format = args[0].(types.StringValue).Value()
			}
		}

		switch format {
		case "markdown":
			doc, err := b.docGenerator.GenerateMarkdownForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			return types.NewStringValue(doc), nil
		case "openapi":
			spec, err := b.docGenerator.GenerateOpenAPIForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			jsonBytes, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				return nil, err
			}
			return types.NewStringValue(string(jsonBytes)), nil
		case "json":
			docs, err := b.docGenerator.GenerateDocsForAllTools(ctx)
			if err != nil {
				return nil, err
			}
			jsonBytes, err := json.MarshalIndent(docs, "", "  ")
			if err != nil {
				return nil, err
			}
			return types.NewStringValue(string(jsonBytes)), nil
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}

	case "generateToolPlayground":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateToolPlayground requires name parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()

		// Generate interactive HTML playground
		html, err := b.generatePlaygroundHTML(name)
		if err != nil {
			return nil, err
		}
		return types.NewStringValue(html), nil

	case "generateSDKSnippet":
		if len(args) < 2 {
			return nil, fmt.Errorf("generateSDKSnippet requires name and language parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("language must be string")
		}
		name := args[0].(types.StringValue).Value()
		language := args[1].(types.StringValue).Value()

		snippet, err := b.generateSDKSnippet(name, language)
		if err != nil {
			return nil, err
		}
		return types.NewStringValue(snippet), nil

	// Execution analytics methods (Task 1.4.9.3)
	case "getToolMetrics":
		if len(args) < 1 {
			return nil, fmt.Errorf("getToolMetrics requires name parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(types.StringValue).Value()

		b.metricsLock.RLock()
		metrics, exists := b.executionMetrics[name]
		b.metricsLock.RUnlock()

		if !exists {
			return nil, fmt.Errorf("no metrics found for tool: %s", name)
		}

		return types.ConvertToScriptValue(b.metricsToMap(metrics)), nil

	case "getAllToolsMetrics":
		b.metricsLock.RLock()
		defer b.metricsLock.RUnlock()

		result := make([]types.ScriptValue, 0, len(b.executionMetrics))
		for _, metrics := range b.executionMetrics {
			result = append(result, types.ConvertToScriptValue(b.metricsToMap(metrics)))
		}

		return types.NewArrayValue(result), nil

	case "getToolUsageReport":
		period := "day" // default
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == types.TypeString {
				period = args[0].(types.StringValue).Value()
			}
		}

		report, err := b.generateUsageReport(period)
		if err != nil {
			return nil, err
		}
		return types.ConvertToScriptValue(report), nil

	case "enableToolProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("enableToolProfiling requires name parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("name must be string")
		}

		// Enable profiling for all tools
		b.profiler.Enable()
		return types.NewNilValue(), nil

	case "getToolAnomalies":
		var toolName string
		if len(args) > 0 {
			if args[0] != nil && args[0].Type() == types.TypeString {
				toolName = args[0].(types.StringValue).Value()
			}
		}

		anomalies, err := b.detectAnomalies(toolName)
		if err != nil {
			return nil, err
		}
		return types.ConvertToScriptValue(anomalies), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper functions

func toolInfoToScriptValue(info types.ToolInfo) map[string]types.ScriptValue {
	result := map[string]types.ScriptValue{
		"name":        types.NewStringValue(info.Name),
		"description": types.NewStringValue(info.Description),
		"category":    types.NewStringValue(info.Category),
		"tags":        convertStringSliceToScriptValue(info.Tags),
		"version":     types.NewStringValue(info.Version),
		"usageHint":   types.NewStringValue(info.UsageHint),
		"package":     types.NewStringValue(info.Package),
	}

	// Parse schemas if available
	if len(info.ParameterSchema) > 0 {
		var params interface{}
		if err := json.Unmarshal(info.ParameterSchema, &params); err == nil {
			result["parameterSchema"] = types.ConvertToScriptValue(params)
		}
	}

	if len(info.OutputSchema) > 0 {
		var output interface{}
		if err := json.Unmarshal(info.OutputSchema, &output); err == nil {
			result["outputSchema"] = types.ConvertToScriptValue(output)
		}
	}

	return result
}

func convertStringSliceToScriptValue(slice []string) types.ScriptValue {
	values := make([]types.ScriptValue, len(slice))
	for i, s := range slice {
		values[i] = types.NewStringValue(s)
	}
	return types.NewArrayValue(values)
}

// convertScriptValueToInterface converts ScriptValue to interface{} for go-llms compatibility
func convertScriptValueToInterface(v types.ScriptValue) interface{} {
	switch v.Type() {
	case types.TypeString:
		return v.(types.StringValue).Value()
	case types.TypeNumber:
		return v.(types.NumberValue).Value()
	case types.TypeBool:
		return v.(types.BoolValue).Value()
	case types.TypeNil:
		return nil
	case types.TypeArray:
		arr := v.(types.ArrayValue).Elements()
		result := make([]interface{}, len(arr))
		for i, item := range arr {
			result[i] = convertScriptValueToInterface(item)
		}
		return result
	case types.TypeObject:
		obj := v.(types.ObjectValue).Fields()
		result := make(map[string]interface{})
		for k, val := range obj {
			result[k] = convertScriptValueToInterface(val)
		}
		return result
	default:
		return v.ToGo()
	}
}

func toolSchemaToScriptValue(schema *types.ToolSchema) map[string]types.ScriptValue {
	return map[string]types.ScriptValue{
		"name":          types.NewStringValue(schema.Name),
		"description":   types.NewStringValue(schema.Description),
		"parameters":    types.ConvertToScriptValue(schema.Parameters),
		"output":        types.ConvertToScriptValue(schema.Output),
		"examples":      types.ConvertToScriptValue(schema.Examples),
		"constraints":   types.ConvertToScriptValue(schema.Constraints),
		"errorGuidance": types.ConvertToScriptValue(schema.ErrorGuidance),
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

func customToolToScriptValue(name string, tool domain.Tool) map[string]types.ScriptValue {
	return map[string]types.ScriptValue{
		"name":                 types.NewStringValue(name),
		"description":          types.NewStringValue(tool.Description()),
		"category":             types.NewStringValue(tool.Category()),
		"tags":                 convertStringSliceToScriptValue(tool.Tags()),
		"version":              types.NewStringValue(tool.Version()),
		"custom":               types.NewBoolValue(true),
		"isDeterministic":      types.NewBoolValue(tool.IsDeterministic()),
		"isDestructive":        types.NewBoolValue(tool.IsDestructive()),
		"requiresConfirmation": types.NewBoolValue(tool.RequiresConfirmation()),
		"estimatedLatency":     types.NewStringValue(tool.EstimatedLatency()),
		"usageInstructions":    types.NewStringValue(tool.UsageInstructions()),
		"constraints":          types.ConvertToScriptValue(tool.Constraints()),
		"errorGuidance":        types.ConvertToScriptValue(tool.ErrorGuidance()),
	}
}

// getStringField safely extracts a string field from a map.
// It returns the string value if present, or an empty string if not found or not a string.
func getStringField(m map[string]interface{}, field string) string {
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}

// createEnhancedCustomTool creates a custom tool with full schema support using ToolBuilder.
// It constructs a tool from a definition map that includes schemas, examples,
// metadata, constraints, and execution handlers.
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
			// For script-defined tools, we'll need to handle execution differently
			// For now, just return a placeholder result
			result = map[string]interface{}{
				"message": "Script tool execution not yet implemented",
				"tool": name,
			}
			err = nil
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
func (b *ToolsBridge) toolToSchemaScriptValue(tool domain.Tool) map[string]types.ScriptValue {
	schema := &types.ToolSchema{
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

// convertBridgeSchemaToSchema converts types.ToolSchema fields to domain.Schema
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
func (b *ToolsBridge) customToolToToolInfo(name string, tool domain.Tool) types.ToolInfo {
	info := types.ToolInfo{
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
	var toolInfo types.ToolInfo
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
	var toolInfo types.ToolInfo
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
