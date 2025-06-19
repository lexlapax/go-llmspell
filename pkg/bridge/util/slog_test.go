// ABOUTME: Tests for structured logging bridge with ScriptValue-based API
// ABOUTME: Validates slog integration, logging hooks, and structured attribute handling

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlogBridgeInitialization(t *testing.T) {
	bridge := NewSlogBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "slog", bridge.GetID())
	assert.False(t, bridge.IsInitialized())

	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestSlogBridgeMetadata(t *testing.T) {
	bridge := NewSlogBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "slog", metadata.Name)
	assert.Equal(t, "v2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "structured logging")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
	assert.Contains(t, metadata.Dependencies, "log/slog")
}

func TestSlogBridgeMethods(t *testing.T) {
	bridge := NewSlogBridge()
	methods := bridge.Methods()

	// Check that all expected logging methods are present
	expectedMethods := []string{
		// Basic logging
		"info",
		"warn",
		"error",
		"debug",
		// Logging hooks
		"logBeforeGenerate",
		"logAfterGenerate",
		"logBeforeToolCall",
		"logAfterToolCall",
		// Configuration
		"setLogLevel",
		"getLogLevel",
		"configureLogger",
		"withAttributes",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestSlogBridgeBasicLogging(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		method  string
		args    []engine.ScriptValue
		wantErr bool
	}{
		{
			name:   "info with message only",
			method: "info",
			args: []engine.ScriptValue{
				sv("Test info message"),
			},
			wantErr: false,
		},
		{
			name:   "warn with emoji",
			method: "warn",
			args: []engine.ScriptValue{
				sv("Warning message"),
				sv("⚠️"),
			},
			wantErr: false,
		},
		{
			name:   "error with attributes",
			method: "error",
			args: []engine.ScriptValue{
				sv("Error occurred"),
				sv("❌"),
				svMap(map[string]interface{}{
					"error_code": 500,
					"reason":     "internal error",
				}),
			},
			wantErr: false,
		},
		{
			name:   "debug with all params",
			method: "debug",
			args: []engine.ScriptValue{
				sv("Debug info"),
				sv("🐛"),
				svMap(map[string]interface{}{
					"step":  1,
					"value": "test",
				}),
			},
			wantErr: false,
		},
		{
			name:    "missing message",
			method:  "info",
			args:    []engine.ScriptValue{},
			wantErr: true,
		},
		{
			name:   "invalid message type",
			method: "info",
			args: []engine.ScriptValue{
				sv(123),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.IsNil())
			}
		})
	}
}

func TestSlogBridgeLoggingHooks(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	t.Run("logBeforeGenerate", func(t *testing.T) {
		messages := svArray(
			map[string]interface{}{
				"role":    "user",
				"content": "Hello AI",
			},
			map[string]interface{}{
				"role":    "assistant",
				"content": "Hello human",
			},
		)

		result, err := bridge.ExecuteMethod(ctx, "logBeforeGenerate", []engine.ScriptValue{
			messages,
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())
	})

	t.Run("logAfterGenerate", func(t *testing.T) {
		response := svMap(map[string]interface{}{
			"content": "Generated response",
		})

		result, err := bridge.ExecuteMethod(ctx, "logAfterGenerate", []engine.ScriptValue{
			response,
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())

		// Test with error
		result, err = bridge.ExecuteMethod(ctx, "logAfterGenerate", []engine.ScriptValue{
			response,
			sv("API error occurred"),
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())
	})

	t.Run("logBeforeToolCall", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "logBeforeToolCall", []engine.ScriptValue{
			sv("web_search"),
			svMap(map[string]interface{}{
				"query": "golang tutorials",
				"limit": 10,
			}),
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())
	})

	t.Run("logAfterToolCall", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "logAfterToolCall", []engine.ScriptValue{
			sv("web_search"),
			svMap(map[string]interface{}{
				"results": svArray("result1", "result2"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())

		// Test with error
		result, err = bridge.ExecuteMethod(ctx, "logAfterToolCall", []engine.ScriptValue{
			sv("web_search"),
			sv(nil),
			sv("Network timeout"),
		})
		require.NoError(t, err)
		assert.True(t, result.IsNil())
	})
}

func TestSlogBridgeLogLevel(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get initial log level
	result, err := bridge.ExecuteMethod(ctx, "getLogLevel", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, "basic", result.(engine.StringValue).Value())

	// Set log level to detailed
	result, err = bridge.ExecuteMethod(ctx, "setLogLevel", []engine.ScriptValue{
		sv("detailed"),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Verify level changed
	result, err = bridge.ExecuteMethod(ctx, "getLogLevel", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, "detailed", result.(engine.StringValue).Value())

	// Test invalid level
	_, err = bridge.ExecuteMethod(ctx, "setLogLevel", []engine.ScriptValue{
		sv("invalid"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestSlogBridgeConfigureLogger(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name   string
		config engine.ScriptValue
	}{
		{
			name: "json format with debug level",
			config: svMap(map[string]interface{}{
				"format": "json",
				"level":  "debug",
			}),
		},
		{
			name: "text format with error level",
			config: svMap(map[string]interface{}{
				"format": "text",
				"level":  "error",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "configureLogger", []engine.ScriptValue{
				tt.config,
			})
			require.NoError(t, err)
			assert.True(t, result.IsNil())
		})
	}
}

func TestSlogBridgeWithAttributes(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	attributes := svMap(map[string]interface{}{
		"component":  "auth",
		"session_id": "abc123",
		"user_id":    42,
	})

	result, err := bridge.ExecuteMethod(ctx, "withAttributes", []engine.ScriptValue{
		attributes,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, engine.TypeObject, result.Type())

	// Verify returned context object
	contextObj := result.(engine.ObjectValue).Fields()
	assert.Contains(t, contextObj, "logger")
	assert.Contains(t, contextObj, "attributes")
}

func TestSlogBridgeValidateMethod(t *testing.T) {
	bridge := NewSlogBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("info", []engine.ScriptValue{
		sv("test message"),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestSlogBridgeRequiredPermissions(t *testing.T) {
	bridge := NewSlogBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for expected permissions
	hasStorage := false
	hasMemory := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionStorage && perm.Resource == "slog.logging" {
			hasStorage = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
		if perm.Type == engine.PermissionMemory && perm.Resource == "slog.context" {
			hasMemory = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasStorage, "Storage permission not found")
	assert.True(t, hasMemory, "Memory permission not found")
}

func TestSlogBridgeTypeMappings(t *testing.T) {
	bridge := NewSlogBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"slog_logger",
		"log_level",
		"logging_hook",
		"message_array",
		"llm_response",
	}

	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestSlogBridgeErrorHandling(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()

	// Test method execution before initialization
	_, err := bridge.ExecuteMethod(ctx, "info", []engine.ScriptValue{
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
	_, err = bridge.ExecuteMethod(ctx, "logBeforeGenerate", []engine.ScriptValue{
		sv("not an array"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array")
}
