package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowBridge_Initialize(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestWorkflowBridge_GetID(t *testing.T) {
	bridge := NewWorkflowBridge()
	assert.Equal(t, "workflow", bridge.GetID())
}

func TestWorkflowBridge_GetMetadata(t *testing.T) {
	bridge := NewWorkflowBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "Workflow Bridge", metadata.Name)
	assert.Equal(t, "2.1.0", metadata.Version)
	assert.Contains(t, metadata.Description, "workflow")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestWorkflowBridge_Methods(t *testing.T) {
	bridge := NewWorkflowBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"createWorkflow", "executeWorkflow", "pauseWorkflow", "resumeWorkflow",
		"stopWorkflow", "getWorkflowStatus", "listWorkflows", "getWorkflow",
		"removeWorkflow", "addStep", "removeStep", "updateStep", "getStep",
		"listSteps", "moveStep", "duplicateStep", "validateWorkflow",
		"getWorkflowMetrics", "resetWorkflowMetrics", "scheduleWorkflow",
		"cancelScheduledWorkflow", "listScheduledWorkflows", "createWorkflowTemplate",
		"listWorkflowTemplates", "getWorkflowTemplate", "removeWorkflowTemplate",
		"createWorkflowFromTemplate", "exportWorkflow", "importWorkflow",
		"getWorkflowHistory", "clearWorkflowHistory", "setWorkflowVariable",
		"getWorkflowVariable", "listWorkflowVariables", "removeWorkflowVariable",
	}

	assert.GreaterOrEqual(t, len(methods), len(expectedMethods))

	// Check that key methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}

func TestWorkflowBridge_ValidateMethod(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		expectError bool
	}{
		{
			name:        "valid createWorkflow",
			method:      "createWorkflow",
			args:        []engine.ScriptValue{engine.NewStringValue("test-workflow"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
			expectError: false,
		},
		{
			name:        "invalid createWorkflow - missing args",
			method:      "createWorkflow",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid listWorkflows",
			method:      "listWorkflows",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:        "valid executeWorkflow",
			method:      "executeWorkflow",
			args:        []engine.ScriptValue{engine.NewStringValue("test-workflow"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
			expectError: false,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkflowBridge_ExecuteMethod_CreateWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test createWorkflow
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name":        engine.NewStringValue("Test Workflow"),
		"description": engine.NewStringValue("A test workflow"),
		"steps": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewObjectValue(map[string]engine.ScriptValue{
				"name": engine.NewStringValue("step1"),
				"type": engine.NewStringValue("action"),
			}),
		}),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	result, err := bridge.ExecuteMethod(ctx, "createWorkflow", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (workflow ID) from createWorkflow")
	assert.Equal(t, workflowID, stringValue.Value(), "Workflow ID should match input")
}

func TestWorkflowBridge_ExecuteMethod_ListWorkflows(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listWorkflows - should work even with no workflows
	result, err := bridge.ExecuteMethod(ctx, "listWorkflows", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listWorkflows")

	workflows := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(workflows), 0, "Should return array of workflows")
}

func TestWorkflowBridge_ExecuteMethod_GetWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test getWorkflow
	getArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "getWorkflow", getArgs)
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getWorkflow")

	workflow := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, workflow, "id")
	assert.Contains(t, workflow, "name")

	// Test getWorkflow with non-existent ID
	nonExistentArgs := []engine.ScriptValue{engine.NewStringValue("non-existent")}
	result, err = bridge.ExecuteMethod(ctx, "getWorkflow", nonExistentArgs)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for non-existent workflow")
	assert.Contains(t, errorValue.Error().Error(), "not found")
}

func TestWorkflowBridge_ExecuteMethod_ExecuteWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
		"steps": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewObjectValue(map[string]engine.ScriptValue{
				"name": engine.NewStringValue("step1"),
				"type": engine.NewStringValue("action"),
			}),
		}),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test executeWorkflow
	executeArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"input": engine.NewStringValue("test input"),
		}),
	}

	result, err := bridge.ExecuteMethod(ctx, "executeWorkflow", executeArgs)
	assert.NoError(t, err)

	// Should return execution result
	assert.NotNil(t, result, "Should return workflow execution result")
}

func TestWorkflowBridge_ExecuteMethod_GetWorkflowStatus(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test getWorkflowStatus
	statusArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "getWorkflowStatus", statusArgs)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue from getWorkflowStatus")
	assert.Contains(t, []string{"created", "running", "paused", "completed", "failed"}, stringValue.Value())
}

func TestWorkflowBridge_ExecuteMethod_PauseResumeWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test pauseWorkflow
	pauseArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "pauseWorkflow", pauseArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from pauseWorkflow")
	assert.True(t, boolValue.Value(), "Pause should succeed")

	// Test resumeWorkflow
	resumeArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err = bridge.ExecuteMethod(ctx, "resumeWorkflow", resumeArgs)
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from resumeWorkflow")
	assert.True(t, boolValue.Value(), "Resume should succeed")
}

func TestWorkflowBridge_ExecuteMethod_AddRemoveStep(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test addStep
	stepConfig := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("new-step"),
		"type": engine.NewStringValue("action"),
	}

	addArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(stepConfig),
	}

	result, err := bridge.ExecuteMethod(ctx, "addStep", addArgs)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (step ID) from addStep")
	stepID := stringValue.Value()

	// Test getStep
	getStepArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(stepID),
	}

	result, err = bridge.ExecuteMethod(ctx, "getStep", getStepArgs)
	assert.NoError(t, err)

	_, ok = result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getStep")

	// Test removeStep
	removeArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(stepID),
	}

	result, err = bridge.ExecuteMethod(ctx, "removeStep", removeArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from removeStep")
	assert.True(t, boolValue.Value(), "Remove should succeed")
}

func TestWorkflowBridge_ExecuteMethod_ListSteps(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
		"steps": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewObjectValue(map[string]engine.ScriptValue{
				"name": engine.NewStringValue("step1"),
			}),
		}),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test listSteps
	listArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "listSteps", listArgs)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listSteps")

	steps := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(steps), 0, "Should return array of steps")
}

func TestWorkflowBridge_ExecuteMethod_ValidateWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test validateWorkflow
	workflowConfig := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
		"steps": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewObjectValue(map[string]engine.ScriptValue{
				"name": engine.NewStringValue("step1"),
				"type": engine.NewStringValue("action"),
			}),
		}),
	}

	args := []engine.ScriptValue{engine.NewObjectValue(workflowConfig)}
	result, err := bridge.ExecuteMethod(ctx, "validateWorkflow", args)
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from validateWorkflow")

	validation := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, validation, "valid")
}

func TestWorkflowBridge_ExecuteMethod_GetWorkflowMetrics(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test getWorkflowMetrics
	metricsArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "getWorkflowMetrics", metricsArgs)
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getWorkflowMetrics")

	metrics := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, metrics, "execution_count")
	assert.Contains(t, metrics, "success_count")
}

func TestWorkflowBridge_ExecuteMethod_SetGetWorkflowVariable(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Test setWorkflowVariable
	setArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue("test_var"),
		engine.NewStringValue("test_value"),
	}

	result, err := bridge.ExecuteMethod(ctx, "setWorkflowVariable", setArgs)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from setWorkflowVariable")

	// Test getWorkflowVariable
	getArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue("test_var"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getWorkflowVariable", getArgs)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue from getWorkflowVariable")
	assert.Equal(t, "test_value", stringValue.Value(), "Variable value should match what was set")
}

func TestWorkflowBridge_ExecuteMethod_RemoveWorkflow(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a workflow first
	workflowID := "test-workflow"
	config := map[string]engine.ScriptValue{
		"name": engine.NewStringValue("Test Workflow"),
	}

	createArgs := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewObjectValue(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createWorkflow", createArgs)
	require.NoError(t, err)

	// Verify workflow exists
	getArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err := bridge.ExecuteMethod(ctx, "getWorkflow", getArgs)
	require.NoError(t, err)
	_, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Workflow should exist")

	// Remove the workflow
	removeArgs := []engine.ScriptValue{engine.NewStringValue(workflowID)}
	result, err = bridge.ExecuteMethod(ctx, "removeWorkflow", removeArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from removeWorkflow")
	assert.True(t, boolValue.Value(), "Remove should succeed")

	// Verify workflow was removed
	result, err = bridge.ExecuteMethod(ctx, "getWorkflow", getArgs)
	assert.NoError(t, err)
	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for removed workflow")
	assert.Contains(t, errorValue.Error().Error(), "not found")
}

func TestWorkflowBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestWorkflowBridge_RequiredPermissions(t *testing.T) {
	bridge := NewWorkflowBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasWorkflowPermission := false
	for _, perm := range permissions {
		if perm.Resource == "workflow" {
			hasWorkflowPermission = true
			break
		}
	}
	assert.True(t, hasWorkflowPermission, "Should have workflow permission")
}

func TestWorkflowBridge_TypeMappings(t *testing.T) {
	bridge := NewWorkflowBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Workflow", "WorkflowStep", "WorkflowTemplate"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestWorkflowBridge_Cleanup(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestWorkflowBridge_NotInitialized(t *testing.T) {
	bridge := NewWorkflowBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listWorkflows", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
