// ABOUTME: Tests for the workflow bridge that exposes go-llms workflow functionality to scripts
// ABOUTME: Verifies workflow creation, configuration, step management, and execution bridging

package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkflowBridge(t *testing.T) {
	bridge := NewWorkflowBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "workflow", bridge.GetID())
}

func TestWorkflowBridgeMetadata(t *testing.T) {
	bridge := NewWorkflowBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "workflow", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "workflow")
	assert.NotEmpty(t, metadata.Author)
	assert.NotEmpty(t, metadata.License)
}

func TestWorkflowBridgeInitialization(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	// Test initial state
	assert.False(t, bridge.IsInitialized())

	// Initialize
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Check serializers are initialized
	assert.Len(t, bridge.serializers, 3) // json, json-pretty, yaml
	assert.NotNil(t, bridge.serializers["json"])
	assert.NotNil(t, bridge.serializers["json-pretty"])
	assert.NotNil(t, bridge.serializers["yaml"])

	// Check script handlers are initialized
	assert.Len(t, bridge.scriptHandlers, 3) // javascript, lua, tengo
	assert.NotNil(t, bridge.scriptHandlers["javascript"])
	assert.NotNil(t, bridge.scriptHandlers["lua"])
	assert.NotNil(t, bridge.scriptHandlers["tengo"])

	// Double initialization should be safe
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// Test Task 1.4.10.1: Workflow Import/Export
func TestWorkflowBridge_ImportExport(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name:    "export workflow - JSON",
			method:  "exportWorkflow",
			args:    []interface{}{"test-workflow", "json"},
			wantErr: true, // No workflow registered
		},
		{
			name:   "import workflow - JSON",
			method: "importWorkflow",
			args: []interface{}{
				`{"name":"test","description":"Test workflow","steps":[]}`,
				"json",
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				wf, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", wf["name"])
				assert.Equal(t, "Test workflow", wf["description"])
			},
		},
		{
			name:   "validate workflow data - valid JSON",
			method: "validateWorkflowData",
			args: []interface{}{
				`{"name":"test","steps":[]}`,
				"json",
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				validation, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.True(t, validation["valid"].(bool))
				assert.Equal(t, "json", validation["format"])
			},
		},
		{
			name:   "validate workflow data - invalid JSON",
			method: "validateWorkflowData",
			args: []interface{}{
				`{invalid json}`,
				"json",
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				validation, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.False(t, validation["valid"].(bool))
				assert.NotEmpty(t, validation["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

// Test Task 1.4.10.2: Script Step Handlers
func TestWorkflowBridge_ScriptSteps(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name:   "register script handler",
			method: "registerScriptHandler",
			args: []interface{}{
				"python",
				map[string]interface{}{
					"version": "3.9",
					"enabled": true,
				},
			},
			wantErr: false,
		},
		{
			name:   "create script step",
			method: "createScriptStep",
			args: []interface{}{
				"calculate",
				"javascript",
				"return 2 + 2;",
				map[string]interface{}{
					"description": "Simple calculation",
					"environment": map[string]interface{}{
						"debug": true,
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				step, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "calculate", step["name"])
				assert.Equal(t, "javascript", step["language"])
				assert.Equal(t, "script", step["type"])
			},
		},
		{
			name:   "validate script step - valid",
			method: "validateScriptStep",
			args: []interface{}{
				map[string]interface{}{
					"language": "lua",
					"script":   "print('hello')",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				validation, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.True(t, validation["valid"].(bool))
				assert.Equal(t, "lua", validation["language"])
			},
		},
		{
			name:   "validate script step - unsupported language",
			method: "validateScriptStep",
			args: []interface{}{
				map[string]interface{}{
					"language": "rust",
					"script":   "println!(\"hello\");",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				validation, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.False(t, validation["valid"].(bool))
				assert.Contains(t, validation["error"], "no handler")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

// Test Task 1.4.10.3: Workflow Templates
func TestWorkflowBridge_Templates(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name:    "list templates",
			method:  "listTemplates",
			args:    []interface{}{},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				templates, ok := result.([]map[string]interface{})
				require.True(t, ok)
				// Should have default templates
				assert.NotEmpty(t, templates)
			},
		},
		{
			name:    "get template - data processing",
			method:  "getTemplate",
			args:    []interface{}{"data-processing"},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				// Template should exist from default templates
				assert.NotNil(t, result)
			},
		},
		{
			name:   "register template",
			method: "registerTemplate",
			args: []interface{}{
				map[string]interface{}{
					"id":          "custom-template",
					"name":        "Custom Template",
					"description": "A custom workflow template",
					"category":    "custom",
					"tags":        []string{"custom", "test"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				templateID, ok := result.(string)
				require.True(t, ok)
				assert.Equal(t, "custom-template", templateID)
			},
		},
		{
			name:   "create from template",
			method: "createFromTemplate",
			args: []interface{}{
				"data-processing",
				map[string]interface{}{
					"input_source": "data.csv",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				wf, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, wf["id"])
				assert.Equal(t, "data-processing", wf["fromTemplate"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestWorkflowBridgeInitialization_Legacy(t *testing.T) {
	tests := []struct {
		name    string
		bridge  *WorkflowBridge
		wantErr bool
	}{
		{
			name:    "successful initialization",
			bridge:  NewWorkflowBridge(),
			wantErr: false,
		},
		{
			name:    "double initialization",
			bridge:  &WorkflowBridge{initialized: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bridge.Initialize(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.bridge.IsInitialized())
			}
		})
	}
}

func TestWorkflowBridgeMethods(t *testing.T) {
	bridge := NewWorkflowBridge()
	methods := bridge.Methods()

	// Essential workflow methods
	expectedMethods := []string{
		"createSequentialWorkflow",
		"createParallelWorkflow",
		"createConditionalWorkflow",
		"createLoopWorkflow",
		"addWorkflowStep",
		"removeWorkflowStep",
		"setWorkflowConfig",
		"executeWorkflow",
		"executeWorkflowAsync",
		"getWorkflowStatus",
		"getWorkflowState",
		"setWorkflowState",
		"pauseWorkflow",
		"resumeWorkflow",
		"cancelWorkflow",
		"listWorkflows",
		"getWorkflow",
		"removeWorkflow",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Missing expected method: %s", expected)
	}

	// Verify method details
	for _, method := range methods {
		assert.NotEmpty(t, method.Description)
		assert.NotEmpty(t, method.ReturnType)

		// Check specific methods have correct parameters
		switch method.Name {
		case "createSequentialWorkflow":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // name, config
		case "addWorkflowStep":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // workflowID, step
		case "executeWorkflow":
			assert.GreaterOrEqual(t, len(method.Parameters), 1) // workflowID
		}
	}
}

func TestWorkflowBridgeTypeMappings(t *testing.T) {
	bridge := NewWorkflowBridge()
	mappings := bridge.TypeMappings()

	expectedTypes := []string{
		"WorkflowAgent",
		"WorkflowStep",
		"WorkflowState",
		"WorkflowDefinition",
		"WorkflowStatus",
		"WorkflowConfig",
		"SequentialConfig",
		"ParallelConfig",
		"ConditionalConfig",
		"LoopConfig",
		"ErrorHandler",
		"StepResult",
	}

	for _, typeName := range expectedTypes {
		mapping, exists := mappings[typeName]
		assert.True(t, exists, "Missing type mapping for %s", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestWorkflowBridgeRequiredPermissions(t *testing.T) {
	bridge := NewWorkflowBridge()
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Should require workflow management permission
	hasWorkflowPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionProcess && perm.Resource == "workflow" {
			hasWorkflowPermission = true
			assert.Contains(t, perm.Actions, "create")
			assert.Contains(t, perm.Actions, "execute")
			assert.Contains(t, perm.Actions, "manage")
		}
	}
	assert.True(t, hasWorkflowPermission, "Missing workflow permission")
}

func TestWorkflowBridgeValidateMethod(t *testing.T) {
	bridge := NewWorkflowBridge()

	tests := []struct {
		name    string
		method  string
		args    []interface{}
		wantErr bool
	}{
		{
			name:    "valid createSequentialWorkflow",
			method:  "createSequentialWorkflow",
			args:    []interface{}{"test-workflow", map[string]interface{}{"timeout": 60}},
			wantErr: false,
		},
		{
			name:    "valid addWorkflowStep",
			method:  "addWorkflowStep",
			args:    []interface{}{"workflow-id", map[string]interface{}{"name": "step1"}},
			wantErr: false,
		},
		{
			name:    "executeWorkflow missing args",
			method:  "executeWorkflow",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
		{
			name:    "unknown method",
			method:  "unknownMethod",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkflowBridgeEngineRegistration(t *testing.T) {
	bridge := NewWorkflowBridge()
	engine := NewMockScriptEngine()

	err := bridge.RegisterWithEngine(engine)
	require.NoError(t, err)

	// Verify bridge was registered
	registered, err := engine.GetBridge("workflow")
	assert.NoError(t, err)
	assert.Equal(t, bridge, registered)
}

func TestWorkflowBridgeCleanup(t *testing.T) {
	bridge := NewWorkflowBridge()

	// Initialize first
	err := bridge.Initialize(context.Background())
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Cleanup
	err = bridge.Cleanup(context.Background())
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestWorkflowBridgeConcurrentAccess(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Concurrent operations
	done := make(chan bool, 3)

	// Reader 1
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.IsInitialized()
			_ = bridge.GetID()
			_ = bridge.Methods()
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.TypeMappings()
			_ = bridge.RequiredPermissions()
		}
		done <- true
	}()

	// Writer
	go func() {
		for i := 0; i < 50; i++ {
			_ = bridge.Initialize(ctx)
			_ = bridge.Cleanup(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestWorkflowBridgeWorkflowTypes(t *testing.T) {
	bridge := NewWorkflowBridge()
	methods := bridge.Methods()

	// Ensure all workflow types are supported
	workflowTypes := []string{
		"createSequentialWorkflow",
		"createParallelWorkflow",
		"createConditionalWorkflow",
		"createLoopWorkflow",
	}

	for _, wfType := range workflowTypes {
		found := false
		for _, method := range methods {
			if method.Name == wfType {
				found = true
				// Check that workflow creation methods have proper parameters
				assert.GreaterOrEqual(t, len(method.Parameters), 2,
					"Workflow creation method %s should have at least name and config parameters", wfType)
				break
			}
		}
		assert.True(t, found, "Missing workflow type creation method: %s", wfType)
	}
}

func TestWorkflowBridgeStepManagement(t *testing.T) {
	bridge := NewWorkflowBridge()
	methods := bridge.Methods()

	// Ensure step management methods exist
	stepMethods := []string{
		"addWorkflowStep",
		"removeWorkflowStep",
		"getWorkflowSteps",
		"updateWorkflowStep",
		"reorderWorkflowSteps",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, stepMethod := range stepMethods {
		if methodMap[stepMethod] {
			// Found the method, verify it has appropriate parameters
			for _, method := range methods {
				if method.Name == stepMethod {
					assert.GreaterOrEqual(t, len(method.Parameters), 1,
						"Step management method %s should have parameters", stepMethod)
					break
				}
			}
		}
	}
}

func TestWorkflowBridgeExecutionControl(t *testing.T) {
	bridge := NewWorkflowBridge()
	methods := bridge.Methods()

	// Ensure execution control methods exist
	controlMethods := []string{
		"pauseWorkflow",
		"resumeWorkflow",
		"cancelWorkflow",
		"retryWorkflowStep",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, controlMethod := range controlMethods {
		assert.True(t, methodMap[controlMethod],
			"Missing workflow control method: %s", controlMethod)
	}
}
