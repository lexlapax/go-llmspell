// ABOUTME: Tests for unified script logger bridge with ScriptValue-based API
// ABOUTME: Validates combined debug/structured logging, context propagation, and configuration

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptLoggerBridgeInitialization(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "script_logger", bridge.GetID())
	assert.False(t, bridge.IsInitialized())

	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Verify sub-bridges are initialized
	assert.True(t, bridge.debugBridge.IsInitialized())
	assert.True(t, bridge.slogBridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestScriptLoggerBridgeMetadata(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "script_logger", metadata.Name)
	assert.Equal(t, "v1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Unified script-friendly logging")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
	assert.Contains(t, metadata.Dependencies, "github.com/lexlapax/go-llmspell/pkg/bridge/util.DebugBridge")
	assert.Contains(t, metadata.Dependencies, "github.com/lexlapax/go-llmspell/pkg/bridge/util.SlogBridge")
}

func TestScriptLoggerBridgeMethods(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := []string{
		// Unified logging
		"log",
		"logWithContext",
		// Context management
		"withContext",
		"setGlobalContext",
		"clearGlobalContext",
		// Configuration
		"configure",
		"getConfiguration",
		// Component management
		"enableComponent",
		"disableComponent",
		"listEnabledComponents",
		// Convenience methods
		"debug",
		"info",
		"warn",
		"error",
		// Bridge integration
		"logBridgeError",
		// Formatting
		"formatMessage",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestScriptLoggerBridgeUnifiedLog(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []engine.ScriptValue
		wantErr bool
	}{
		{
			name: "log with level and message only",
			args: []engine.ScriptValue{
				sv("info"),
				sv("Test message"),
			},
			wantErr: false,
		},
		{
			name: "log with level, message, and attributes",
			args: []engine.ScriptValue{
				sv("warn"),
				sv("Warning message"),
				svMap(map[string]interface{}{
					"code":   404,
					"reason": "not found",
				}),
			},
			wantErr: false,
		},
		{
			name: "log with level, component, and message",
			args: []engine.ScriptValue{
				sv("debug"),
				sv("agent"),
				sv("Debug message"),
			},
			wantErr: false,
		},
		{
			name: "log with all parameters",
			args: []engine.ScriptValue{
				sv("error"),
				sv("workflow"),
				sv("Error occurred"),
				svMap(map[string]interface{}{
					"step":  3,
					"error": "timeout",
				}),
			},
			wantErr: false,
		},
		{
			name:    "missing required arguments",
			args:    []engine.ScriptValue{sv("info")},
			wantErr: true,
		},
		{
			name: "invalid level type",
			args: []engine.ScriptValue{
				sv(123),
				sv("message"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "log", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.IsNil())
			}
		})
	}
}

func TestScriptLoggerBridgeContextManagement(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test setGlobalContext
	globalAttrs := svMap(map[string]interface{}{
		"app":     "test-app",
		"version": "1.0.0",
	})
	result, err := bridge.ExecuteMethod(ctx, "setGlobalContext", []engine.ScriptValue{globalAttrs})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test withContext
	contextAttrs := svMap(map[string]interface{}{
		"session_id": "abc123",
		"user_id":    42,
	})
	result, err = bridge.ExecuteMethod(ctx, "withContext", []engine.ScriptValue{contextAttrs})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, engine.TypeObject, result.Type())

	contextObj := result.(engine.ObjectValue).Fields()
	assert.Contains(t, contextObj, "attributes")
	assert.Contains(t, contextObj, "logger")

	// Test logWithContext
	result, err = bridge.ExecuteMethod(ctx, "logWithContext", []engine.ScriptValue{
		sv("info"),
		sv("Context test"),
		svMap(map[string]interface{}{
			"action": "test",
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test clearGlobalContext
	result, err = bridge.ExecuteMethod(ctx, "clearGlobalContext", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.True(t, result.IsNil())
}

func TestScriptLoggerBridgeConfiguration(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get initial configuration
	result, err := bridge.ExecuteMethod(ctx, "getConfiguration", []engine.ScriptValue{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, engine.TypeObject, result.Type())

	config := result.(engine.ObjectValue).Fields()
	assert.Equal(t, "info", config["default_level"].(engine.StringValue).Value())
	assert.True(t, config["enable_debug"].(engine.BoolValue).Value())
	assert.True(t, config["enable_structure"].(engine.BoolValue).Value())

	// Configure logger
	newConfig := svMap(map[string]interface{}{
		"level":        "debug",
		"format":       "json",
		"enable_debug": false,
		"components":   svArray("agent", "workflow"),
	})
	result, err = bridge.ExecuteMethod(ctx, "configure", []engine.ScriptValue{newConfig})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Verify configuration changed
	result, err = bridge.ExecuteMethod(ctx, "getConfiguration", []engine.ScriptValue{})
	require.NoError(t, err)
	config = result.(engine.ObjectValue).Fields()
	assert.Equal(t, "debug", config["default_level"].(engine.StringValue).Value())
	assert.Equal(t, "json", config["format"].(engine.StringValue).Value())
	assert.False(t, config["enable_debug"].(engine.BoolValue).Value())
}

func TestScriptLoggerBridgeComponentManagement(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Enable component
	result, err := bridge.ExecuteMethod(ctx, "enableComponent", []engine.ScriptValue{
		sv("test-component"),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// List enabled components
	result, err = bridge.ExecuteMethod(ctx, "listEnabledComponents", []engine.ScriptValue{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, engine.TypeArray, result.Type())

	components := result.(engine.ArrayValue).Elements()
	found := false
	for _, comp := range components {
		if comp.Type() == engine.TypeString && comp.(engine.StringValue).Value() == "test-component" {
			found = true
			break
		}
	}
	assert.True(t, found, "test-component not found in enabled components")

	// Disable component
	result, err = bridge.ExecuteMethod(ctx, "disableComponent", []engine.ScriptValue{
		sv("test-component"),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())
}

func TestScriptLoggerBridgeConvenienceMethods(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test debug
	result, err := bridge.ExecuteMethod(ctx, "debug", []engine.ScriptValue{
		sv("Debug message"),
		sv("test-component"),
		svMap(map[string]interface{}{
			"step": 1,
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test info
	result, err = bridge.ExecuteMethod(ctx, "info", []engine.ScriptValue{
		sv("Info message"),
		svMap(map[string]interface{}{
			"user": "john",
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test warn
	result, err = bridge.ExecuteMethod(ctx, "warn", []engine.ScriptValue{
		sv("Warning message"),
		svMap(map[string]interface{}{
			"threshold": 90,
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test error
	result, err = bridge.ExecuteMethod(ctx, "error", []engine.ScriptValue{
		sv("Error message"),
		svMap(map[string]interface{}{
			"code": "ERR_001",
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())
}

func TestScriptLoggerBridgeBridgeError(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "logBridgeError", []engine.ScriptValue{
		sv("llm"),
		sv("generate"),
		sv("API timeout"),
		svMap(map[string]interface{}{
			"duration": 5000,
			"retries":  3,
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test without optional context
	result, err = bridge.ExecuteMethod(ctx, "logBridgeError", []engine.ScriptValue{
		sv("tool"),
		sv("execute"),
		sv("Tool not found"),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())
}

func TestScriptLoggerBridgeFormatMessage(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "formatMessage", []engine.ScriptValue{
		sv("User {user} completed task {task} in {duration}ms"),
		svMap(map[string]interface{}{
			"user":     "john",
			"task":     "upload",
			"duration": 1500,
		}),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, engine.TypeString, result.Type())
	assert.Equal(t, "User john completed task upload in 1500ms", result.(engine.StringValue).Value())
}

func TestScriptLoggerBridgeValidateMethod(t *testing.T) {
	bridge := NewScriptLoggerBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("log", []engine.ScriptValue{
		sv("info"),
		sv("test"),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestScriptLoggerBridgeRequiredPermissions(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for expected permissions
	hasMemory := false
	hasStorage := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionMemory && perm.Resource == "script_logger.context" {
			hasMemory = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
		if perm.Type == engine.PermissionStorage && perm.Resource == "script_logger.output" {
			hasStorage = true
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasMemory, "Memory permission not found")
	assert.True(t, hasStorage, "Storage permission not found")
}

func TestScriptLoggerBridgeTypeMappings(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"logger_config",
		"log_context",
		"unified_logger",
	}

	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestScriptLoggerBridgeErrorHandling(t *testing.T) {
	bridge := NewScriptLoggerBridge()
	ctx := context.Background()

	// Test method execution before initialization
	_, err := bridge.ExecuteMethod(ctx, "log", []engine.ScriptValue{
		sv("info"),
		sv("test"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test unknown method
	_, err = bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown method")

	// Test invalid arguments
	_, err = bridge.ExecuteMethod(ctx, "log", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least")
}
