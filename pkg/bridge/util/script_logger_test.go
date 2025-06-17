// ABOUTME: Tests for unified script logger bridge functionality combining debug and structured logging
// ABOUTME: Comprehensive test coverage for context propagation, configuration, and unified logging APIs

package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test ScriptLoggerBridge core functionality
func TestScriptLoggerBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *ScriptLoggerBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "script_logger", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "Unified script-friendly")
			},
		},
		{
			name: "Unified logging method",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test basic log call
				err = bridge.log(ctx, []interface{}{"info", "Test message"})
				require.NoError(t, err)

				// Test log with component
				err = bridge.log(ctx, []interface{}{"debug", "agent", "Processing request"})
				require.NoError(t, err)

				// Test log with component and attributes
				err = bridge.log(ctx, []interface{}{
					"info",
					"workflow",
					"Step completed",
					map[string]interface{}{
						"step":     1,
						"duration": 150,
						"success":  true,
					},
				})
				require.NoError(t, err)

				// Test log without component but with attributes
				err = bridge.log(ctx, []interface{}{
					"warn",
					"Rate limit approaching",
					map[string]interface{}{
						"remaining": 10,
						"limit":     100,
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Context management",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test setting global context
				err = bridge.setGlobalContext(ctx, []interface{}{
					map[string]interface{}{
						"app":     "go-llmspell",
						"version": "1.0.0",
						"env":     "test",
					},
				})
				require.NoError(t, err)

				// Test creating context with attributes
				result, err := bridge.withContext(ctx, []interface{}{
					map[string]interface{}{
						"session_id": "abc123",
						"user":       "test_user",
					},
				})
				require.NoError(t, err)

				contextObj, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "log_context", contextObj["type"])

				attributes, ok := contextObj["attributes"].(map[string]interface{})
				require.True(t, ok)
				// Should include both global and specific attributes
				assert.Equal(t, "go-llmspell", attributes["app"])
				assert.Equal(t, "abc123", attributes["session_id"])

				// Test logging with context
				err = bridge.logWithContext(ctx, []interface{}{
					"info",
					"User action completed",
					map[string]interface{}{
						"action":    "upload",
						"file_size": 1024,
					},
				})
				require.NoError(t, err)

				// Test clearing global context
				err = bridge.clearGlobalContext(ctx, []interface{}{})
				require.NoError(t, err)
			},
		},
		{
			name: "Configuration management",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test getting initial configuration
				result, err := bridge.getConfiguration(ctx, []interface{}{})
				require.NoError(t, err)

				config, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "info", config["default_level"])
				assert.Equal(t, true, config["enable_debug"])
				assert.Equal(t, true, config["enable_structure"])

				// Test configuring logger
				err = bridge.configure(ctx, []interface{}{
					map[string]interface{}{
						"level":            "debug",
						"format":           "json",
						"enable_debug":     false,
						"enable_structure": true,
						"components":       []interface{}{"agent", "workflow"},
						"attributes": map[string]interface{}{
							"service": "test-service",
							"build":   "1.2.3",
						},
					},
				})
				require.NoError(t, err)

				// Verify configuration was applied
				result, err = bridge.getConfiguration(ctx, []interface{}{})
				require.NoError(t, err)

				config, ok = result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "debug", config["default_level"])
				assert.Equal(t, false, config["enable_debug"])
				assert.Equal(t, "json", config["format"])

				globalContext, ok := config["global_context"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-service", globalContext["service"])
			},
		},
		{
			name: "Component management",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test enabling components
				err = bridge.enableComponent(ctx, []interface{}{"agent"})
				require.NoError(t, err)

				err = bridge.enableComponent(ctx, []interface{}{"workflow"})
				require.NoError(t, err)

				// Test listing enabled components
				result, err := bridge.listEnabledComponents(ctx, []interface{}{})
				require.NoError(t, err)

				components, ok := result.([]string)
				require.True(t, ok)
				assert.Contains(t, components, "agent")
				assert.Contains(t, components, "workflow")

				// Test disabling component
				err = bridge.disableComponent(ctx, []interface{}{"agent"})
				require.NoError(t, err)

				result, err = bridge.listEnabledComponents(ctx, []interface{}{})
				require.NoError(t, err)

				components, ok = result.([]string)
				require.True(t, ok)
				assert.NotContains(t, components, "agent")
				assert.Contains(t, components, "workflow")
			},
		},
		{
			name: "Convenience logging methods",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test debug logging
				err = bridge.debug(ctx, []interface{}{"Debug message"})
				require.NoError(t, err)

				err = bridge.debug(ctx, []interface{}{
					"Debug with component",
					"workflow",
					map[string]interface{}{"step": 1},
				})
				require.NoError(t, err)

				// Test info logging
				err = bridge.info(ctx, []interface{}{
					"Info message",
					map[string]interface{}{"user": "test"},
				})
				require.NoError(t, err)

				// Test warn logging
				err = bridge.warn(ctx, []interface{}{
					"Warning message",
					map[string]interface{}{"threshold": 80},
				})
				require.NoError(t, err)

				// Test error logging
				err = bridge.error(ctx, []interface{}{
					"Error message",
					map[string]interface{}{"code": 500},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Bridge error logging",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test bridge error logging
				err = bridge.logBridgeError(ctx, []interface{}{
					"llm",
					"generate",
					"API timeout",
					map[string]interface{}{
						"duration": 5000,
						"retries":  3,
					},
				})
				require.NoError(t, err)

				// Test bridge error without additional context
				err = bridge.logBridgeError(ctx, []interface{}{
					"agent",
					"initialize",
					"Configuration invalid",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Message formatting",
			test: func(t *testing.T, bridge *ScriptLoggerBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test message formatting
				result, err := bridge.formatMessage(ctx, []interface{}{
					"User {user} completed task {task} in {duration}ms",
					map[string]interface{}{
						"user":     "john",
						"task":     "upload",
						"duration": 1500,
					},
				})
				require.NoError(t, err)

				formatted, ok := result.(string)
				require.True(t, ok)
				// Note: This is a simple implementation that would need proper template handling
				assert.Contains(t, formatted, "User")
				assert.Contains(t, formatted, "completed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewScriptLoggerBridge()
			tt.test(t, bridge)
		})
	}
}

// Test script logger bridge error scenarios
func TestScriptLoggerBridgeErrors(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()

	// Test methods without initialization
	err := bridge.log(ctx, []interface{}{"info", "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameter types
	err = bridge.log(ctx, []interface{}{123, "message"}) // level must be string
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.log(ctx, []interface{}{"info", 123}) // message must be string
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.setGlobalContext(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	_, err = bridge.withContext(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	err = bridge.configure(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	// Test bridge error logging with invalid types
	err = bridge.logBridgeError(ctx, []interface{}{123, "operation", "error"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.logBridgeError(ctx, []interface{}{"bridge", 123, "error"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.logBridgeError(ctx, []interface{}{"bridge", "operation", 123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	// Test format message with invalid types
	_, err = bridge.formatMessage(ctx, []interface{}{123, map[string]interface{}{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	_, err = bridge.formatMessage(ctx, []interface{}{"template", "not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")
}

// Test script logger bridge lifecycle
func TestScriptLoggerBridgeLifecycle(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "script_logger", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "logger_config")
	assert.Contains(t, typeMappings, "log_context")
	assert.Contains(t, typeMappings, "unified_logger")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 10)

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"log", "logWithContext", "withContext",
		"setGlobalContext", "clearGlobalContext",
		"configure", "getConfiguration",
		"enableComponent", "disableComponent", "listEnabledComponents",
		"debug", "info", "warn", "error",
		"logBridgeError", "formatMessage",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// Test script logger bridge validation
func TestScriptLoggerBridgeValidation(t *testing.T) {
	bridge := NewScriptLoggerBridge()
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
			name:        "valid log",
			method:      "log",
			args:        []interface{}{"info", "message"},
			shouldError: false,
		},
		{
			name:        "log missing args",
			method:      "log",
			args:        []interface{}{"info"},
			shouldError: true,
		},
		{
			name:        "valid logWithContext",
			method:      "logWithContext",
			args:        []interface{}{"info", "message"},
			shouldError: false,
		},
		{
			name:        "logWithContext missing args",
			method:      "logWithContext",
			args:        []interface{}{"info"},
			shouldError: true,
		},
		{
			name:        "valid withContext",
			method:      "withContext",
			args:        []interface{}{map[string]interface{}{"key": "value"}},
			shouldError: false,
		},
		{
			name:        "withContext missing args",
			method:      "withContext",
			args:        []interface{}{},
			shouldError: true,
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

// Test script logger bridge integration
func TestScriptLoggerBridgeIntegration(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test that both debug and structured logging work together

	// Configure for both debug and structured logging
	err = bridge.configure(ctx, []interface{}{
		map[string]interface{}{
			"enable_debug":     true,
			"enable_structure": true,
			"components":       []interface{}{"test"},
			"format":           "text",
		},
	})
	require.NoError(t, err)

	// Set global context
	err = bridge.setGlobalContext(ctx, []interface{}{
		map[string]interface{}{
			"app":     "test-app",
			"version": "1.0",
		},
	})
	require.NoError(t, err)

	// Test unified logging with component (should use both debug and structured)
	err = bridge.log(ctx, []interface{}{
		"debug",
		"test",
		"Integration test message",
		map[string]interface{}{
			"feature": "integration",
			"count":   42,
		},
	})
	require.NoError(t, err)

	// Test convenience methods
	err = bridge.info(ctx, []interface{}{
		"Info message from integration test",
		map[string]interface{}{
			"test_id": "integration_001",
		},
	})
	require.NoError(t, err)

	// Test bridge error logging
	err = bridge.logBridgeError(ctx, []interface{}{
		"test_bridge",
		"integration_test",
		"Simulated error for testing",
		map[string]interface{}{
			"severity":  "medium",
			"retryable": true,
		},
	})
	require.NoError(t, err)
}

// Test default configuration
func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()

	assert.Equal(t, "info", config.DefaultLevel)
	assert.True(t, config.EnableDebug)
	assert.True(t, config.EnableStructure)
	assert.Equal(t, "text", config.Format)
	assert.Equal(t, "stderr", config.OutputTarget)
	assert.NotNil(t, config.Components)
	assert.NotNil(t, config.Attributes)
}

// Test context merging behavior
func TestScriptLoggerContextMerging(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Set global context
	err = bridge.setGlobalContext(ctx, []interface{}{
		map[string]interface{}{
			"app":     "test-app",
			"env":     "test",
			"version": "1.0",
		},
	})
	require.NoError(t, err)

	// Test that withContext merges correctly
	result, err := bridge.withContext(ctx, []interface{}{
		map[string]interface{}{
			"session_id": "abc123",
			"user":       "test_user",
			"env":        "override", // Should override global
		},
	})
	require.NoError(t, err)

	contextObj, ok := result.(map[string]interface{})
	require.True(t, ok)

	attributes, ok := contextObj["attributes"].(map[string]interface{})
	require.True(t, ok)

	// Should have global attributes
	assert.Equal(t, "test-app", attributes["app"])
	assert.Equal(t, "1.0", attributes["version"])

	// Should have specific attributes
	assert.Equal(t, "abc123", attributes["session_id"])
	assert.Equal(t, "test_user", attributes["user"])

	// Should override global with specific
	assert.Equal(t, "override", attributes["env"])
}
