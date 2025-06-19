package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentBridge_Initialize(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestAgentBridge_GetID(t *testing.T) {
	bridge := NewAgentBridge()
	assert.Equal(t, "agent", bridge.GetID())
}

func TestAgentBridge_GetMetadata(t *testing.T) {
	bridge := NewAgentBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "agent", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Agent system bridge")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestAgentBridge_Methods(t *testing.T) {
	bridge := NewAgentBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"createAgent", "createLLMAgent", "registerTool", "runAgent", "runAgentAsync",
		"addSubAgent", "getAgentState", "setAgentState", "listAgents", "getAgent",
		"removeAgent", "setAgentHook", "emitAgentEvent", "subscribeToEvents",
		"unsubscribeFromEvents", "getAgentMetrics", "createWorkflow", "addWorkflowStep",
		"getAgentTools", "configureAgent", "exportAgentState", "importAgentState",
		"saveAgentSnapshot", "loadAgentSnapshot", "listAgentSnapshots", "deleteAgentSnapshot",
		"replayAgentEvents", "startEventRecording", "stopEventRecording", "getEventHistory",
		"clearEventHistory", "startAgentProfiling", "stopAgentProfiling",
		"getAgentPerformanceReport", "clearAgentProfilingData", "exportAgentProfilingData",
		"setAgentProfilingConfig",
	}

	assert.GreaterOrEqual(t, len(methods), len(expectedMethods))

	// Check that some key methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}

func TestAgentBridge_ValidateMethod(t *testing.T) {
	bridge := NewAgentBridge()
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
			name:        "valid createAgent",
			method:      "createAgent",
			args:        []engine.ScriptValue{sv("agent1"), svMap(map[string]interface{}{})},
			expectError: false,
		},
		{
			name:        "invalid createAgent - missing args",
			method:      "createAgent",
			args:        []engine.ScriptValue{sv("agent1")},
			expectError: true,
		},
		{
			name:        "valid listAgents",
			method:      "listAgents",
			args:        []engine.ScriptValue{},
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

func TestAgentBridge_ExecuteMethod_ListAgents(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listAgents - should work even with no agents
	result, err := bridge.ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue")
	assert.Equal(t, 0, len(arrayValue.ToGo().([]interface{})), "Expected empty array")
}

func TestAgentBridge_ExecuteMethod_CreateAgent(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test createAgent
	agentID := "test-agent"
	config := map[string]interface{}{
		"name":        "Test Agent",
		"description": "A test agent",
	}

	args := []engine.ScriptValue{
		sv(agentID),
		svMap(config),
	}

	result, err := bridge.ExecuteMethod(ctx, "createAgent", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify agent was created by listing agents
	listResult, err := bridge.ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := listResult.(engine.ArrayValue)
	assert.True(t, ok)
	agents := arrayValue.ToGo().([]interface{})
	assert.Equal(t, 1, len(agents), "Expected one agent")
}

func TestAgentBridge_ExecuteMethod_GetAgent(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create an agent first
	agentID := "test-agent"
	config := map[string]interface{}{
		"name": "Test Agent",
	}

	createArgs := []engine.ScriptValue{
		sv(agentID),
		svMap(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createAgent", createArgs)
	require.NoError(t, err)

	// Test getAgent
	getArgs := []engine.ScriptValue{sv(agentID)}
	result, err := bridge.ExecuteMethod(ctx, "getAgent", getArgs)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getAgent with non-existent ID
	nonExistentArgs := []engine.ScriptValue{sv("non-existent")}
	result, err = bridge.ExecuteMethod(ctx, "getAgent", nonExistentArgs)
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for non-existent agent")
	assert.Contains(t, errorValue.Error().Error(), "not found")
}

func TestAgentBridge_ExecuteMethod_RemoveAgent(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create an agent first
	agentID := "test-agent"
	config := map[string]interface{}{
		"name": "Test Agent",
	}

	createArgs := []engine.ScriptValue{
		sv(agentID),
		svMap(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createAgent", createArgs)
	require.NoError(t, err)

	// Verify agent exists
	listResult, err := bridge.ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
	require.NoError(t, err)
	arrayValue := listResult.(engine.ArrayValue)
	assert.Equal(t, 1, len(arrayValue.ToGo().([]interface{})))

	// Remove the agent
	removeArgs := []engine.ScriptValue{sv(agentID)}
	result, err := bridge.ExecuteMethod(ctx, "removeAgent", removeArgs)
	assert.NoError(t, err)

	nilValue, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from removeAgent")
	assert.True(t, nilValue.IsNil())

	// Verify agent was removed
	listResult, err = bridge.ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
	require.NoError(t, err)
	arrayValue = listResult.(engine.ArrayValue)
	assert.Equal(t, 0, len(arrayValue.ToGo().([]interface{})))
}

func TestAgentBridge_ExecuteMethod_GetAgentMetrics(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create an agent first
	agentID := "test-agent"
	config := map[string]interface{}{
		"name": "Test Agent",
	}

	createArgs := []engine.ScriptValue{
		sv(agentID),
		svMap(config),
	}

	_, err = bridge.ExecuteMethod(ctx, "createAgent", createArgs)
	require.NoError(t, err)

	// Get metrics
	metricsArgs := []engine.ScriptValue{sv(agentID)}
	result, err := bridge.ExecuteMethod(ctx, "getAgentMetrics", metricsArgs)
	assert.NoError(t, err)

	// Should return an object with metrics
	objectValue, ok := result.(engine.ObjectValue)
	if !assert.True(t, ok, "Expected ObjectValue from getAgentMetrics") {
		t.Logf("Got type: %T, value: %v", result, result)
		return
	}

	metrics := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, metrics, "execution_count")
	assert.Contains(t, metrics, "success_count")
}

func TestAgentBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestAgentBridge_RequiredPermissions(t *testing.T) {
	bridge := NewAgentBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasAgentPermission := false
	for _, perm := range permissions {
		if perm.Resource == "agent" {
			hasAgentPermission = true
			break
		}
	}
	assert.True(t, hasAgentPermission, "Should have agent permission")
}

func TestAgentBridge_TypeMappings(t *testing.T) {
	bridge := NewAgentBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Agent", "AgentConfig", "AgentState"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestAgentBridge_Cleanup(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestAgentBridge_NotInitialized(t *testing.T) {
	bridge := NewAgentBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listAgents", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
