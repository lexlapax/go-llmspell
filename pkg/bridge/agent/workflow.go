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
	// _ "github.com/lexlapax/go-llms/pkg/agent/workflow" // Will be used when workflow package is available
)

// WorkflowBridge provides script access to go-llms workflow functionality
type WorkflowBridge struct {
	mu          sync.RWMutex
	initialized bool
	workflows   map[string]bridge.BaseAgent // Workflows are agents in go-llms
}

// NewWorkflowBridge creates a new workflow bridge
func NewWorkflowBridge() *WorkflowBridge {
	return &WorkflowBridge{
		workflows: make(map[string]bridge.BaseAgent),
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
		Version:     "1.0.0",
		Description: "Workflow engine bridge wrapping go-llms workflow functionality",
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
			ReturnType: "WorkflowAgent",
		},
		{
			Name:        "createParallelWorkflow",
			Description: "Create a parallel workflow that executes steps concurrently",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Workflow name", Required: true},
				{Name: "config", Type: "object", Description: "Parallel workflow configuration", Required: true},
			},
			ReturnType: "WorkflowAgent",
		},
		{
			Name:        "createConditionalWorkflow",
			Description: "Create a conditional workflow with branching logic",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Workflow name", Required: true},
				{Name: "config", Type: "object", Description: "Conditional workflow configuration", Required: true},
			},
			ReturnType: "WorkflowAgent",
		},
		{
			Name:        "createLoopWorkflow",
			Description: "Create a loop workflow for iterative processing",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Workflow name", Required: true},
				{Name: "config", Type: "object", Description: "Loop workflow configuration", Required: true},
			},
			ReturnType: "WorkflowAgent",
		},
		// Step management
		{
			Name:        "addWorkflowStep",
			Description: "Add a step to a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "step", Type: "object", Description: "Step configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "removeWorkflowStep",
			Description: "Remove a step from a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "stepID", Type: "string", Description: "Step ID to remove", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getWorkflowSteps",
			Description: "Get all steps in a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "updateWorkflowStep",
			Description: "Update a workflow step configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "stepID", Type: "string", Description: "Step ID", Required: true},
				{Name: "config", Type: "object", Description: "Updated step configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "reorderWorkflowSteps",
			Description: "Reorder steps in a sequential workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "stepOrder", Type: "array", Description: "New step order (array of step IDs)", Required: true},
			},
			ReturnType: "void",
		},
		// Workflow configuration
		{
			Name:        "setWorkflowConfig",
			Description: "Update workflow configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "config", Type: "object", Description: "Workflow configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getWorkflowConfig",
			Description: "Get workflow configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "object",
		},
		// Execution methods
		{
			Name:        "executeWorkflow",
			Description: "Execute a workflow synchronously",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "input", Type: "any", Description: "Workflow input", Required: false},
			},
			ReturnType: "any",
		},
		{
			Name:        "executeWorkflowAsync",
			Description: "Execute a workflow asynchronously",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "input", Type: "any", Description: "Workflow input", Required: false},
			},
			ReturnType: "channel",
		},
		// Workflow state and status
		{
			Name:        "getWorkflowStatus",
			Description: "Get the current status of a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "WorkflowStatus",
		},
		{
			Name:        "getWorkflowState",
			Description: "Get the current state of a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "WorkflowState",
		},
		{
			Name:        "setWorkflowState",
			Description: "Set the state of a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "state", Type: "WorkflowState", Description: "New workflow state", Required: true},
			},
			ReturnType: "void",
		},
		// Execution control
		{
			Name:        "pauseWorkflow",
			Description: "Pause a running workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "resumeWorkflow",
			Description: "Resume a paused workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "cancelWorkflow",
			Description: "Cancel a running workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "retryWorkflowStep",
			Description: "Retry a failed workflow step",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "stepID", Type: "string", Description: "Step ID to retry", Required: true},
			},
			ReturnType: "void",
		},
		// Workflow management
		{
			Name:        "listWorkflows",
			Description: "List all registered workflows",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getWorkflow",
			Description: "Get a workflow by ID",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "WorkflowAgent",
		},
		{
			Name:        "removeWorkflow",
			Description: "Remove a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
			},
			ReturnType: "void",
		},
		// Error handling
		{
			Name:        "setWorkflowErrorHandler",
			Description: "Set error handler for a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "handler", Type: "function", Description: "Error handler function", Required: true},
			},
			ReturnType: "void",
		},
		// Hooks and events
		{
			Name:        "setWorkflowHook",
			Description: "Set a lifecycle hook for a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "hookType", Type: "string", Description: "Hook type (beforeStep, afterStep, etc.)", Required: true},
				{Name: "handler", Type: "function", Description: "Hook handler function", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribeToWorkflowEvents",
			Description: "Subscribe to workflow events",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "eventType", Type: "string", Description: "Event type to subscribe to", Required: false},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string", // subscription ID
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
		"WorkflowStatus": {
			GoType:     "WorkflowStatus",
			ScriptType: "object",
		},
		"WorkflowConfig": {
			GoType:     "WorkflowConfig",
			ScriptType: "object",
		},
		"SequentialConfig": {
			GoType:     "SequentialConfig",
			ScriptType: "object",
		},
		"ParallelConfig": {
			GoType:     "ParallelConfig",
			ScriptType: "object",
		},
		"ConditionalConfig": {
			GoType:     "ConditionalConfig",
			ScriptType: "object",
		},
		"LoopConfig": {
			GoType:     "LoopConfig",
			ScriptType: "object",
		},
		"ErrorHandler": {
			GoType:     "ErrorHandler",
			ScriptType: "function",
		},
		"StepResult": {
			GoType:     "StepResult",
			ScriptType: "object",
		},
		"ErrorAction": {
			GoType:     "ErrorAction",
			ScriptType: "string",
		},
		"BranchCondition": {
			GoType:     "BranchCondition",
			ScriptType: "object",
		},
		"LoopCondition": {
			GoType:     "LoopCondition",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *WorkflowBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
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

// Helper methods for type conversion and workflow management

// getWorkflow retrieves a workflow by ID
//
//nolint:unused // will be used when implementing workflow methods
func (b *WorkflowBridge) getWorkflow(id string) (bridge.BaseAgent, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	workflow, exists := b.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}
	return workflow, nil
}

// registerWorkflow registers a workflow in the bridge
//
//nolint:unused // will be used when implementing workflow creation methods
func (b *WorkflowBridge) registerWorkflow(workflow bridge.BaseAgent) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.workflows[workflow.ID()]; exists {
		return fmt.Errorf("workflow %s already registered", workflow.ID())
	}

	b.workflows[workflow.ID()] = workflow
	return nil
}

// removeWorkflowInternal removes a workflow from the bridge
//
//nolint:unused // will be used when implementing removeWorkflow method
func (b *WorkflowBridge) removeWorkflowInternal(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *WorkflowBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createSequentialWorkflow":
		if len(args) < 2 {
			return nil, fmt.Errorf("createSequentialWorkflow requires name and config parameters")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Placeholder until go-llms workflow package is available
		// Will create actual workflow using workflow.NewSequentialWorkflow
		return map[string]interface{}{
			"id":     fmt.Sprintf("workflow-%s", name),
			"type":   "sequential",
			"name":   name,
			"config": config,
		}, nil

	case "listWorkflows":
		workflows := make([]map[string]interface{}, 0, len(b.workflows))
		for id, workflow := range b.workflows {
			workflows = append(workflows, map[string]interface{}{
				"id":   id,
				"type": workflow.Type(),
				"name": workflow.Name(),
			})
		}
		return workflows, nil

	case "executeWorkflow":
		if len(args) < 1 {
			return nil, fmt.Errorf("executeWorkflow requires workflowID parameter")
		}
		workflowID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("workflowID must be string")
		}

		workflow, err := b.getWorkflow(workflowID)
		if err != nil {
			return nil, err
		}

		// Create input state
		inputState := domain.NewState()
		if len(args) > 1 && args[1] != nil {
			if inputData, ok := args[1].(map[string]interface{}); ok {
				for k, v := range inputData {
					inputState.Set(k, v)
				}
			}
		}

		// Execute workflow
		resultState, err := workflow.Run(ctx, inputState)
		if err != nil {
			return nil, fmt.Errorf("workflow execution failed: %w", err)
		}

		// Return result state values
		return resultState.Values(), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
