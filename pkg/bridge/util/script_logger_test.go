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
				engine.NewStringValue("info"),
				engine.NewStringValue("Test message"),
			},
			wantErr: false,
		},
		{
			name: "log with level, message, and attributes",
			args: []engine.ScriptValue{
				engine.NewStringValue("warn"),
				engine.NewStringValue("Warning message"),
				engine.NewObjectValue(map[string]engine.ScriptValue{
					"code":   engine.NewNumberValue(404),
					"reason": engine.NewStringValue("not found"),
				}),
			},
			wantErr: false,
		},
		{
			name: "log with level, component, and message",
			args: []engine.ScriptValue{
				engine.NewStringValue("debug"),
				engine.NewStringValue("agent"),
				engine.NewStringValue("Debug message"),
			},
			wantErr: false,
		},
		{
			name: "log with all parameters",
			args: []engine.ScriptValue{
				engine.NewStringValue("error"),
				engine.NewStringValue("workflow"),
				engine.NewStringValue("Error occurred"),
				engine.NewObjectValue(map[string]engine.ScriptValue{
					"step":  engine.NewNumberValue(3),
					"error": engine.NewStringValue("timeout"),
				}),
			},
			wantErr: false,
		},
		{
			name:    "missing required arguments",
			args:    []engine.ScriptValue{engine.NewStringValue("info")},
			wantErr: true,
		},
		{
			name: "invalid level type",
			args: []engine.ScriptValue{
				engine.NewNumberValue(123),
				engine.NewStringValue("message"),
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
	globalAttrs := engine.NewObjectValue(map[string]engine.ScriptValue{
		"app":     engine.NewStringValue("test-app"),
		"version": engine.NewStringValue("1.0.0"),
	})
	result, err := bridge.ExecuteMethod(ctx, "setGlobalContext", []engine.ScriptValue{globalAttrs})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test withContext
	contextAttrs := engine.NewObjectValue(map[string]engine.ScriptValue{
		"session_id": engine.NewStringValue("abc123"),
		"user_id":    engine.NewNumberValue(42),
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
		engine.NewStringValue("info"),
		engine.NewStringValue("Context test"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"action": engine.NewStringValue("test"),
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
	newConfig := engine.NewObjectValue(map[string]engine.ScriptValue{
		"level":        engine.NewStringValue("debug"),
		"format":       engine.NewStringValue("json"),
		"enable_debug": engine.NewBoolValue(false),
		"components": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("agent"),
			engine.NewStringValue("workflow"),
		}),
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
		engine.NewStringValue("test-component"),
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
		engine.NewStringValue("test-component"),
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
		engine.NewStringValue("Debug message"),
		engine.NewStringValue("test-component"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"step": engine.NewNumberValue(1),
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test info
	result, err = bridge.ExecuteMethod(ctx, "info", []engine.ScriptValue{
		engine.NewStringValue("Info message"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"user": engine.NewStringValue("john"),
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test warn
	result, err = bridge.ExecuteMethod(ctx, "warn", []engine.ScriptValue{
		engine.NewStringValue("Warning message"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"threshold": engine.NewNumberValue(90),
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test error
	result, err = bridge.ExecuteMethod(ctx, "error", []engine.ScriptValue{
		engine.NewStringValue("Error message"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"code": engine.NewStringValue("ERR_001"),
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
		engine.NewStringValue("llm"),
		engine.NewStringValue("generate"),
		engine.NewStringValue("API timeout"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"duration": engine.NewNumberValue(5000),
			"retries":  engine.NewNumberValue(3),
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test without optional context
	result, err = bridge.ExecuteMethod(ctx, "logBridgeError", []engine.ScriptValue{
		engine.NewStringValue("tool"),
		engine.NewStringValue("execute"),
		engine.NewStringValue("Tool not found"),
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
		engine.NewStringValue("User {user} completed task {task} in {duration}ms"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"user":     engine.NewStringValue("john"),
			"task":     engine.NewStringValue("upload"),
			"duration": engine.NewNumberValue(1500),
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
		engine.NewStringValue("info"),
		engine.NewStringValue("test"),
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
		engine.NewStringValue("info"),
		engine.NewStringValue("test"),
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
