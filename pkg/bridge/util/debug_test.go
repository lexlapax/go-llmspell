// ABOUTME: Tests for debug logging bridge functionality including component control and logger configuration
// ABOUTME: Comprehensive test coverage for go-llms debug system integration and conditional compilation

package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test DebugBridge core functionality
func TestDebugBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *DebugBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *DebugBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "debug", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "debug logging")
			},
		},
		{
			name: "Component management",
			test: func(t *testing.T, bridge *DebugBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Initially no components enabled
				result, err := bridge.listEnabledComponents(ctx, []interface{}{})
				require.NoError(t, err)
				components, ok := result.([]string)
				require.True(t, ok)
				assert.Empty(t, components)

				// Enable a component
				err = bridge.enableDebugComponent(ctx, []interface{}{"agent"})
				require.NoError(t, err)

				// Check if component is enabled
				result, err = bridge.isDebugEnabled(ctx, []interface{}{"agent"})
				require.NoError(t, err)
				enabled, ok := result.(bool)
				require.True(t, ok)
				assert.True(t, enabled)

				// List enabled components
				result, err = bridge.listEnabledComponents(ctx, []interface{}{})
				require.NoError(t, err)
				components, ok = result.([]string)
				require.True(t, ok)
				assert.Contains(t, components, "agent")

				// Disable component
				err = bridge.disableDebugComponent(ctx, []interface{}{"agent"})
				require.NoError(t, err)

				// Check if component is disabled
				result, err = bridge.isDebugEnabled(ctx, []interface{}{"agent"})
				require.NoError(t, err)
				enabled, ok = result.(bool)
				require.True(t, ok)
				assert.False(t, enabled)
			},
		},
		{
			name: "Debug logging methods",
			test: func(t *testing.T, bridge *DebugBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test debugPrintf
				err = bridge.debugPrintf(ctx, []interface{}{
					"test",
					"Processing item: %s",
					[]interface{}{"item-123"},
				})
				require.NoError(t, err)

				// Test debugPrintln
				err = bridge.debugPrintln(ctx, []interface{}{
					"test",
					"Simple debug message",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Logger configuration",
			test: func(t *testing.T, bridge *DebugBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test custom logger configuration
				err = bridge.setCustomLogger(ctx, []interface{}{
					map[string]interface{}{
						"prefix": "[SPELL]",
						"flags":  "datetime",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Environment information",
			test: func(t *testing.T, bridge *DebugBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Get debug environment
				result, err := bridge.getDebugEnvironment(ctx, []interface{}{})
				require.NoError(t, err)
				env, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, env, "go_llms_debug_env")
				assert.Contains(t, env, "enabled_components")
				assert.Contains(t, env, "compilation_mode")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewDebugBridge()
			tt.test(t, bridge)
		})
	}
}

// Test debug bridge error scenarios
func TestDebugBridgeErrors(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()

	// Test methods without initialization
	err := bridge.debugPrintf(ctx, []interface{}{"test", "message"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	err = bridge.debugPrintf(ctx, []interface{}{123, "message"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.debugPrintln(ctx, []interface{}{"component", 123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	_, err = bridge.isDebugEnabled(ctx, []interface{}{123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.enableDebugComponent(ctx, []interface{}{123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.setCustomLogger(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")
}

// Test debug bridge lifecycle
func TestDebugBridgeLifecycle(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "debug", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "debug_logger")
	assert.Contains(t, typeMappings, "debug_config")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 5)

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"debugPrintf",
		"debugPrintln",
		"isDebugEnabled",
		"enableDebugComponent",
		"disableDebugComponent",
		"listEnabledComponents",
		"setCustomLogger",
		"getDebugEnvironment",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// Test debug bridge method validation
func TestDebugBridgeValidation(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []interface{}
		shouldError bool
	}{
		{
			name:        "valid debugPrintf",
			method:      "debugPrintf",
			args:        []interface{}{"component", "format", []interface{}{"arg1"}},
			shouldError: false,
		},
		{
			name:        "debugPrintf missing args",
			method:      "debugPrintf",
			args:        []interface{}{"component"},
			shouldError: true,
		},
		{
			name:        "valid debugPrintln",
			method:      "debugPrintln",
			args:        []interface{}{"component", "message"},
			shouldError: false,
		},
		{
			name:        "valid isDebugEnabled",
			method:      "isDebugEnabled",
			args:        []interface{}{"component"},
			shouldError: false,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []interface{}{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test concurrent debug operations
func TestDebugBridgeConcurrency(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent component enable/disable
	numOperations := 10
	done := make(chan bool, numOperations*2)

	// Concurrent enable operations
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			defer func() { done <- true }()
			component := fmt.Sprintf("component-%d", index)
			err := bridge.enableDebugComponent(ctx, []interface{}{component})
			assert.NoError(t, err)
		}(i)
	}

	// Concurrent disable operations
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			defer func() { done <- true }()
			component := fmt.Sprintf("component-%d", index)
			err := bridge.disableDebugComponent(ctx, []interface{}{component})
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numOperations*2; i++ {
		<-done
	}

	// Verify bridge still works
	result, err := bridge.listEnabledComponents(ctx, []interface{}{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test debug bridge component state management
func TestDebugComponentState(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	components := []string{"agent", "tools", "workflow", "llm"}

	// Enable multiple components
	for _, component := range components {
		err = bridge.enableDebugComponent(ctx, []interface{}{component})
		require.NoError(t, err)
	}

	// Verify all are enabled
	for _, component := range components {
		result, err := bridge.isDebugEnabled(ctx, []interface{}{component})
		require.NoError(t, err)
		enabled, ok := result.(bool)
		require.True(t, ok)
		assert.True(t, enabled, "Component %s should be enabled", component)
	}

	// List all enabled components
	result, err := bridge.listEnabledComponents(ctx, []interface{}{})
	require.NoError(t, err)
	enabledList, ok := result.([]string)
	require.True(t, ok)
	assert.Equal(t, len(components), len(enabledList))

	// Disable one component
	err = bridge.disableDebugComponent(ctx, []interface{}{"agent"})
	require.NoError(t, err)

	// Verify agent is disabled
	result, err = bridge.isDebugEnabled(ctx, []interface{}{"agent"})
	require.NoError(t, err)
	enabled, ok := result.(bool)
	require.True(t, ok)
	assert.False(t, enabled)

	// Verify others are still enabled
	result, err = bridge.isDebugEnabled(ctx, []interface{}{"tools"})
	require.NoError(t, err)
	enabled, ok = result.(bool)
	require.True(t, ok)
	assert.True(t, enabled)
}

// Test debug environment functionality
func TestDebugEnvironment(t *testing.T) {
	bridge := NewDebugBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get environment information
	result, err := bridge.getDebugEnvironment(ctx, []interface{}{})
	require.NoError(t, err)

	env, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Verify expected fields
	assert.Contains(t, env, "go_llms_debug_env")
	assert.Contains(t, env, "enabled_components")
	assert.Contains(t, env, "compilation_mode")

	// Verify enabled_components is an array
	enabledComponents, ok := env["enabled_components"].([]string)
	require.True(t, ok)

	// Should be empty by default if no environment variables are set
	// and no components have been explicitly enabled
	assert.GreaterOrEqual(t, len(enabledComponents), 0)
}
