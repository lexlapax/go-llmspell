// ABOUTME: Workflow bridge provides access to go-llms workflow functionality for script engines
// ABOUTME: Wraps workflow creation, configuration, step management, and execution without reimplementation

package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for workflow functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

// WorkflowBridge provides script access to go-llms workflow functionality
type WorkflowBridge struct {
	mu          sync.RWMutex
	initialized bool
	workflows   map[string]bridge.BaseAgent // Workflows are agents in go-llms

	// Task 1.4.10.1: Workflow Import/Export
	serializers     map[string]workflow.WorkflowSerializer
	serializerCache map[string][]byte // Cache serialized workflows

	// Task 1.4.10.2: Script Step Handlers
	scriptHandlers map[string]ScriptStepHandler
	scriptRegistry map[string]*workflow.ScriptStep

	// Task 1.4.10.3: Workflow Templates
	templateCache    map[string]*workflow.WorkflowTemplate
	templateRegistry map[string]*workflow.WorkflowTemplate // Local template registry
}

// ScriptStepHandler handles script execution for workflow steps
type ScriptStepHandler struct {
	Language  string
	Validator func(script string) error
	Executor  func(ctx context.Context, script string, env map[string]interface{}) (interface{}, error)
	Debugger  func(script string, breakpoint int) error
	Metadata  map[string]interface{}
}

// NewWorkflowBridge creates a new workflow bridge
func NewWorkflowBridge() *WorkflowBridge {
	return &WorkflowBridge{
		workflows:        make(map[string]bridge.BaseAgent),
		serializers:      make(map[string]workflow.WorkflowSerializer),
		serializerCache:  make(map[string][]byte),
		scriptHandlers:   make(map[string]ScriptStepHandler),
		scriptRegistry:   make(map[string]*workflow.ScriptStep),
		templateCache:    make(map[string]*workflow.WorkflowTemplate),
		templateRegistry: make(map[string]*workflow.WorkflowTemplate),
	}
}

// GetID returns the bridge identifier
func (b *WorkflowBridge) GetID() string {
	return "workflow"
}

// GetMetadata returns bridge metadata
func (b *WorkflowBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "workflow",
		Version:     "2.0.0",
		Description: "Enhanced workflow engine bridge with serialization, script steps, and templates (v0.3.5)",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *WorkflowBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize default serializers
	b.serializers["json"] = workflow.NewJSONWorkflowSerializer(false)
	b.serializers["json-pretty"] = workflow.NewJSONWorkflowSerializer(true)
	b.serializers["yaml"] = workflow.NewYAMLWorkflowSerializer()

	// Initialize default script handlers
	b.initializeDefaultScriptHandlers()

	// Register default templates
	if err := workflow.RegisterDefaultTemplates(); err != nil {
		return fmt.Errorf("failed to register default templates: %w", err)
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *WorkflowBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clean up any registered workflows
	for id, workflow := range b.workflows {
		if err := workflow.Cleanup(ctx); err != nil {
			// Log error but continue cleanup
			_ = err
		}
		delete(b.workflows, id)
	}

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *WorkflowBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *WorkflowBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *WorkflowBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Workflow creation methods
		{
			Name:        "createSequentialWorkflow",
			Description: "Create a sequential workflow that executes steps one after another",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Workflow name", Required: true},
				{Name: "config", Type: "object", Description: "Sequential workflow configuration", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "listWorkflows",
			Description: "List all registered workflows",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "executeWorkflow",
			Description: "Execute a workflow synchronously",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "input", Type: "any", Description: "Workflow input", Required: false},
			},
			ReturnType: "any",
		},
		// Export/Import functionality
		{
			Name:        "exportWorkflow",
			Description: "Export workflow to specified format",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "format", Type: "string", Description: "Export format", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "importWorkflow",
			Description: "Import workflow from data",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Description: "Workflow data", Required: true},
				{Name: "format", Type: "string", Description: "Data format", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateWorkflowData",
			Description: "Validate workflow data format",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Description: "Workflow data", Required: true},
				{Name: "format", Type: "string", Description: "Data format", Required: false},
			},
			ReturnType: "object",
		},
		// Script step functionality
		{
			Name:        "registerScriptHandler",
			Description: "Register a script handler for a language",
			Parameters: []engine.ParameterInfo{
				{Name: "language", Type: "string", Description: "Script language", Required: true},
				{Name: "handler", Type: "object", Description: "Handler configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "createScriptStep",
			Description: "Create a script step",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Step name", Required: true},
				{Name: "language", Type: "string", Description: "Script language", Required: true},
				{Name: "script", Type: "string", Description: "Script code", Required: true},
				{Name: "config", Type: "object", Description: "Step configuration", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateScriptStep",
			Description: "Validate a script step",
			Parameters: []engine.ParameterInfo{
				{Name: "step", Type: "object", Description: "Script step", Required: true},
			},
			ReturnType: "object",
		},
		// Template functionality
		{
			Name:        "listTemplates",
			Description: "List available templates",
			Parameters: []engine.ParameterInfo{
				{Name: "category", Type: "string", Description: "Template category", Required: false},
				{Name: "tags", Type: "array", Description: "Template tags", Required: false},
			},
			ReturnType: "array",
		},
		{
			Name:        "getTemplate",
			Description: "Get template by ID",
			Parameters: []engine.ParameterInfo{
				{Name: "templateID", Type: "string", Description: "Template ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createFromTemplate",
			Description: "Create workflow from template",
			Parameters: []engine.ParameterInfo{
				{Name: "templateID", Type: "string", Description: "Template ID", Required: true},
				{Name: "variables", Type: "object", Description: "Template variables", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "registerTemplate",
			Description: "Register a new template",
			Parameters: []engine.ParameterInfo{
				{Name: "template", Type: "object", Description: "Template data", Required: true},
			},
			ReturnType: "string",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *WorkflowBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"WorkflowAgent": {
			GoType:     "BaseAgent",
			ScriptType: "object",
		},
		"WorkflowStep": {
			GoType:     "WorkflowStep",
			ScriptType: "object",
		},
		"WorkflowState": {
			GoType:     "*WorkflowState",
			ScriptType: "object",
		},
		"WorkflowDefinition": {
			GoType:     "WorkflowDefinition",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *WorkflowBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	return nil
}

// RequiredPermissions returns required permissions
func (b *WorkflowBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionProcess,
			Resource:    "workflow",
			Actions:     []string{"create", "execute", "manage"},
			Description: "Access to workflow engine",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "state",
			Actions:     []string{"allocate", "manage"},
			Description: "Memory for workflow state and execution",
		},
	}
}

// ExecuteMethod executes a bridge method
func (b *WorkflowBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch name {
	case "createSequentialWorkflow":
		return b.createSequentialWorkflow(args)

	case "listWorkflows":
		return b.listWorkflows()

	case "executeWorkflow":
		return b.executeWorkflow(ctx, args)

	case "exportWorkflow":
		return b.exportWorkflow(args)

	case "importWorkflow":
		return b.importWorkflow(args)

	case "validateWorkflowData":
		return b.validateWorkflowData(args)

	case "registerScriptHandler":
		return b.registerScriptHandler(args)

	case "createScriptStep":
		return b.createScriptStep(args)

	case "validateScriptStep":
		return b.validateScriptStep(args)

	case "listTemplates":
		return b.listTemplates(args)

	case "getTemplate":
		return b.getTemplate(args)

	case "createFromTemplate":
		return b.createFromTemplate(args)

	case "registerTemplate":
		return b.registerTemplate(args)

	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// Method implementations

func (b *WorkflowBridge) createSequentialWorkflow(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return engine.NewErrorValue(fmt.Errorf("createSequentialWorkflow requires name and config parameters")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("name must be string")), nil
	}
	name := args[0].(engine.StringValue).Value()

	if args[1].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("config must be object")), nil
	}
	config := args[1].ToGo().(map[string]interface{})

	// Create workflow result
	result := map[string]engine.ScriptValue{
		"id":     engine.NewStringValue(fmt.Sprintf("workflow-%s", name)),
		"type":   engine.NewStringValue("sequential"),
		"name":   engine.NewStringValue(name),
		"config": convertWorkflowToScriptValue(config),
	}
	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) listWorkflows() (engine.ScriptValue, error) {
	workflows := make([]engine.ScriptValue, 0, len(b.workflows))
	for id, workflow := range b.workflows {
		workflowData := map[string]engine.ScriptValue{
			"id":   engine.NewStringValue(id),
			"type": engine.NewStringValue(string(workflow.Type())),
			"name": engine.NewStringValue(workflow.Name()),
		}
		workflows = append(workflows, engine.NewObjectValue(workflowData))
	}
	return engine.NewArrayValue(workflows), nil
}

func (b *WorkflowBridge) executeWorkflow(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("executeWorkflow requires workflowID parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("workflowID must be string")), nil
	}
	workflowID := args[0].(engine.StringValue).Value()

	workflow, err := b.getWorkflow(workflowID)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Create input state
	inputState := domain.NewState()
	if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeObject {
			inputData := args[1].ToGo().(map[string]interface{})
			for k, v := range inputData {
				inputState.Set(k, v)
			}
		}
	}

	// Execute workflow
	resultState, err := workflow.Run(ctx, inputState)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("workflow execution failed: %w", err)), nil
	}

	// Return result state values
	return convertWorkflowToScriptValue(resultState.Values()), nil
}

func (b *WorkflowBridge) exportWorkflow(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("exportWorkflow requires workflowID parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("workflowID must be string")), nil
	}
	workflowID := args[0].(engine.StringValue).Value()

	format := "json"
	if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeString {
			format = args[1].(engine.StringValue).Value()
		}
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", workflowID, format)
	if cached, exists := b.serializerCache[cacheKey]; exists {
		return engine.NewStringValue(string(cached)), nil
	}

	// Get workflow and serialize
	wf, err := b.getWorkflow(workflowID)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Create workflow definition from agent
	def := &workflow.WorkflowDefinition{
		Name:        wf.Name(),
		Description: wf.Description(),
		Steps:       []workflow.WorkflowStep{},
	}

	serializer := b.serializers[format]
	if serializer == nil {
		return engine.NewErrorValue(fmt.Errorf("unsupported format: %s", format)), nil
	}

	data, err := serializer.Serialize(def)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("serialization failed: %w", err)), nil
	}

	// Cache the result
	b.serializerCache[cacheKey] = data

	return engine.NewStringValue(string(data)), nil
}

func (b *WorkflowBridge) importWorkflow(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("importWorkflow requires data parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("data must be string")), nil
	}
	data := args[0].(engine.StringValue).Value()

	format := "json"
	if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeString {
			format = args[1].(engine.StringValue).Value()
		}
	}

	serializer := b.serializers[format]
	if serializer == nil {
		return engine.NewErrorValue(fmt.Errorf("unsupported format: %s", format)), nil
	}

	def, err := serializer.Deserialize([]byte(data))
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("deserialization failed: %w", err)), nil
	}

	// Create workflow from definition
	result := map[string]engine.ScriptValue{
		"id":          engine.NewStringValue(fmt.Sprintf("workflow-%s", def.Name)),
		"name":        engine.NewStringValue(def.Name),
		"description": engine.NewStringValue(def.Description),
		"steps":       engine.NewNumberValue(float64(len(def.Steps))),
	}
	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) validateWorkflowData(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("validateWorkflowData requires data parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("data must be string")), nil
	}
	data := args[0].(engine.StringValue).Value()

	format := "json"
	if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeString {
			format = args[1].(engine.StringValue).Value()
		}
	}

	serializer := b.serializers[format]
	if serializer == nil {
		return engine.NewErrorValue(fmt.Errorf("unsupported format: %s", format)), nil
	}

	_, err := serializer.Deserialize([]byte(data))

	result := map[string]engine.ScriptValue{
		"valid":  engine.NewBoolValue(err == nil),
		"format": engine.NewStringValue(format),
	}

	if err != nil {
		result["error"] = engine.NewStringValue(err.Error())
	}

	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) registerScriptHandler(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return engine.NewErrorValue(fmt.Errorf("registerScriptHandler requires language and handler parameters")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("language must be string")), nil
	}
	language := args[0].(engine.StringValue).Value()

	if args[1].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("handler must be object")), nil
	}
	handlerConfig := args[1].ToGo().(map[string]interface{})

	// Create handler from config
	handler := ScriptStepHandler{
		Language: language,
		Metadata: handlerConfig,
	}

	b.scriptHandlers[language] = handler
	return engine.NewNilValue(), nil
}

func (b *WorkflowBridge) createScriptStep(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return engine.NewErrorValue(fmt.Errorf("createScriptStep requires name, language, and script parameters")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("name must be string")), nil
	}
	name := args[0].(engine.StringValue).Value()

	if args[1].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("language must be string")), nil
	}
	language := args[1].(engine.StringValue).Value()

	if args[2].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("script must be string")), nil
	}
	script := args[2].(engine.StringValue).Value()

	// Check if handler exists
	if _, exists := b.scriptHandlers[language]; !exists {
		return engine.NewErrorValue(fmt.Errorf("no handler for language: %s", language)), nil
	}

	// Create script step
	stepID := fmt.Sprintf("script-%s-%s", language, name)

	stepConfig := map[string]engine.ScriptValue{
		"id":       engine.NewStringValue(stepID),
		"name":     engine.NewStringValue(name),
		"language": engine.NewStringValue(language),
		"script":   engine.NewStringValue(script),
		"type":     engine.NewStringValue("script"),
	}

	if len(args) > 3 && args[3] != nil {
		if args[3].Type() == engine.TypeObject {
			config := args[3].ToGo().(map[string]interface{})
			if desc, ok := config["description"].(string); ok {
				stepConfig["description"] = engine.NewStringValue(desc)
			}
			if env, ok := config["environment"].(map[string]interface{}); ok {
				stepConfig["environment"] = convertWorkflowToScriptValue(env)
			}
		}
	}

	return engine.NewObjectValue(stepConfig), nil
}

func (b *WorkflowBridge) validateScriptStep(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("validateScriptStep requires step parameter")), nil
	}

	if args[0].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("step must be object")), nil
	}
	stepData := args[0].ToGo().(map[string]interface{})

	language, _ := stepData["language"].(string)
	script, _ := stepData["script"].(string)

	handler, exists := b.scriptHandlers[language]
	if !exists {
		result := map[string]engine.ScriptValue{
			"valid": engine.NewBoolValue(false),
			"error": engine.NewStringValue(fmt.Sprintf("no handler for language: %s", language)),
		}
		return engine.NewObjectValue(result), nil
	}

	var err error
	if handler.Validator != nil {
		err = handler.Validator(script)
	}

	result := map[string]engine.ScriptValue{
		"valid":    engine.NewBoolValue(err == nil),
		"language": engine.NewStringValue(language),
	}

	if err != nil {
		result["error"] = engine.NewStringValue(err.Error())
	}

	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) listTemplates(args []engine.ScriptValue) (engine.ScriptValue, error) {
	var templates []*workflow.WorkflowTemplate

	if len(args) > 0 && args[0] != nil {
		if args[0].Type() == engine.TypeString {
			category := args[0].(engine.StringValue).Value()
			templates = workflow.ListTemplatesByCategory(category)
		}
	} else if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeArray {
			tagsArray := args[1].ToGo().([]interface{})
			tags := make([]string, len(tagsArray))
			for i, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags[i] = tagStr
				}
			}
			templates = workflow.SearchTemplates(tags)
		}
	} else {
		templates = workflow.ListTemplates()
	}

	// Include local templates
	for _, tmpl := range b.templateRegistry {
		templates = append(templates, tmpl)
	}

	result := make([]engine.ScriptValue, len(templates))
	for i, tmpl := range templates {
		templateData := map[string]engine.ScriptValue{
			"id":          engine.NewStringValue(tmpl.ID),
			"name":        engine.NewStringValue(tmpl.Name),
			"description": engine.NewStringValue(tmpl.Description),
			"category":    engine.NewStringValue(tmpl.Category),
		}

		// Convert tags to ScriptValue array
		tags := make([]engine.ScriptValue, len(tmpl.Tags))
		for j, tag := range tmpl.Tags {
			tags[j] = engine.NewStringValue(tag)
		}
		templateData["tags"] = engine.NewArrayValue(tags)

		result[i] = engine.NewObjectValue(templateData)
	}

	return engine.NewArrayValue(result), nil
}

func (b *WorkflowBridge) getTemplate(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("getTemplate requires templateID parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("templateID must be string")), nil
	}
	templateID := args[0].(engine.StringValue).Value()

	// Check local registry first
	if tmpl, exists := b.templateRegistry[templateID]; exists {
		result := map[string]engine.ScriptValue{
			"id":          engine.NewStringValue(tmpl.ID),
			"name":        engine.NewStringValue(tmpl.Name),
			"description": engine.NewStringValue(tmpl.Description),
			"category":    engine.NewStringValue(tmpl.Category),
		}

		// Convert tags
		tags := make([]engine.ScriptValue, len(tmpl.Tags))
		for i, tag := range tmpl.Tags {
			tags[i] = engine.NewStringValue(tag)
		}
		result["tags"] = engine.NewArrayValue(tags)

		return engine.NewObjectValue(result), nil
	}

	// Check global registry
	tmpl, err := workflow.GetTemplate(templateID)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	result := map[string]engine.ScriptValue{
		"id":          engine.NewStringValue(tmpl.ID),
		"name":        engine.NewStringValue(tmpl.Name),
		"description": engine.NewStringValue(tmpl.Description),
		"category":    engine.NewStringValue(tmpl.Category),
	}

	// Convert tags
	tags := make([]engine.ScriptValue, len(tmpl.Tags))
	for i, tag := range tmpl.Tags {
		tags[i] = engine.NewStringValue(tag)
	}
	result["tags"] = engine.NewArrayValue(tags)

	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) createFromTemplate(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("createFromTemplate requires templateID parameter")), nil
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("templateID must be string")), nil
	}
	templateID := args[0].(engine.StringValue).Value()

	variables := make(map[string]interface{})
	if len(args) > 1 && args[1] != nil {
		if args[1].Type() == engine.TypeObject {
			variables = args[1].ToGo().(map[string]interface{})
		}
	}

	def, err := workflow.ApplyTemplate(templateID, variables)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to apply template: %w", err)), nil
	}

	result := map[string]engine.ScriptValue{
		"id":           engine.NewStringValue(fmt.Sprintf("workflow-%s", def.Name)),
		"name":         engine.NewStringValue(def.Name),
		"description":  engine.NewStringValue(def.Description),
		"fromTemplate": engine.NewStringValue(templateID),
	}
	return engine.NewObjectValue(result), nil
}

func (b *WorkflowBridge) registerTemplate(args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return engine.NewErrorValue(fmt.Errorf("registerTemplate requires template parameter")), nil
	}

	if args[0].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("template must be object")), nil
	}
	templateData := args[0].ToGo().(map[string]interface{})

	// Create template from data
	id, _ := templateData["id"].(string)
	name, _ := templateData["name"].(string)
	description, _ := templateData["description"].(string)

	tmpl := &workflow.WorkflowTemplate{
		ID:          id,
		Name:        name,
		Description: description,
	}

	if category, ok := templateData["category"].(string); ok {
		tmpl.Category = category
	}

	if tagsInterface, ok := templateData["tags"].([]interface{}); ok {
		tags := make([]string, len(tagsInterface))
		for i, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags[i] = tagStr
			}
		}
		tmpl.Tags = tags
	}

	// Store in local registry
	b.templateRegistry[tmpl.ID] = tmpl

	return engine.NewStringValue(tmpl.ID), nil
}

// Helper methods

// getWorkflow retrieves a workflow by ID
func (b *WorkflowBridge) getWorkflow(id string) (bridge.BaseAgent, error) {
	workflow, exists := b.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}
	return workflow, nil
}

// registerWorkflow registers a workflow in the bridge
func (b *WorkflowBridge) registerWorkflow(workflow bridge.BaseAgent) error {
	if _, exists := b.workflows[workflow.ID()]; exists {
		return fmt.Errorf("workflow %s already registered", workflow.ID())
	}

	b.workflows[workflow.ID()] = workflow
	return nil
}

// initializeDefaultScriptHandlers sets up default script handlers
func (b *WorkflowBridge) initializeDefaultScriptHandlers() {
	// JavaScript handler
	b.scriptHandlers["javascript"] = ScriptStepHandler{
		Language: "javascript",
		Validator: func(script string) error {
			if script == "" {
				return fmt.Errorf("empty script")
			}
			return nil
		},
		Executor: func(ctx context.Context, script string, env map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": "JavaScript execution placeholder",
				"script": script,
				"env":    env,
			}, nil
		},
		Metadata: map[string]interface{}{
			"supported": true,
			"version":   "ES6",
		},
	}

	// Lua handler
	b.scriptHandlers["lua"] = ScriptStepHandler{
		Language: "lua",
		Validator: func(script string) error {
			if script == "" {
				return fmt.Errorf("empty script")
			}
			return nil
		},
		Executor: func(ctx context.Context, script string, env map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": "Lua execution placeholder",
				"script": script,
				"env":    env,
			}, nil
		},
		Metadata: map[string]interface{}{
			"supported": true,
			"version":   "5.4",
		},
	}

	// Tengo handler
	b.scriptHandlers["tengo"] = ScriptStepHandler{
		Language: "tengo",
		Validator: func(script string) error {
			if script == "" {
				return fmt.Errorf("empty script")
			}
			return nil
		},
		Executor: func(ctx context.Context, script string, env map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": "Tengo execution placeholder",
				"script": script,
				"env":    env,
			}, nil
		},
		Metadata: map[string]interface{}{
			"supported": true,
			"version":   "2.0",
		},
	}
}

// removeWorkflowInternal removes a workflow from the bridge
func (b *WorkflowBridge) removeWorkflowInternal(id string) error {
	workflow, exists := b.workflows[id]
	if !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	// Cleanup the workflow
	if err := workflow.Cleanup(context.Background()); err != nil {
		return fmt.Errorf("failed to cleanup workflow %s: %w", id, err)
	}

	delete(b.workflows, id)
	return nil
}

// convertWorkflowToScriptValue converts a Go interface{} to engine.ScriptValue
func convertWorkflowToScriptValue(v interface{}) engine.ScriptValue {
	if v == nil {
		return engine.NewNilValue()
	}

	switch val := v.(type) {
	case string:
		return engine.NewStringValue(val)
	case bool:
		return engine.NewBoolValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case float32:
		return engine.NewNumberValue(float64(val))
	case map[string]interface{}:
		result := make(map[string]engine.ScriptValue)
		for k, mv := range val {
			result[k] = convertWorkflowToScriptValue(mv)
		}
		return engine.NewObjectValue(result)
	case []interface{}:
		result := make([]engine.ScriptValue, len(val))
		for i, av := range val {
			result[i] = convertWorkflowToScriptValue(av)
		}
		return engine.NewArrayValue(result)
	default:
		// For unknown types, convert to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", val))
	}
}
