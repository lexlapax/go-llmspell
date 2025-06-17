// ABOUTME: Tests for structured logging bridge functionality including slog integration and logging hooks
// ABOUTME: Comprehensive test coverage for go-llms LoggingHook system and structured key-value logging

package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test SlogBridge core functionality
func TestSlogBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *SlogBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "slog", metadata.Name)
				assert.Equal(t, "v2.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "structured logging")
			},
		},
		{
			name: "Basic logging methods",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test info logging
				err = bridge.info(ctx, []interface{}{"Test info message"})
				require.NoError(t, err)

				// Test info with emoji
				err = bridge.info(ctx, []interface{}{"Test info message", "ℹ️"})
				require.NoError(t, err)

				// Test info with emoji and attributes
				err = bridge.info(ctx, []interface{}{
					"Test info message",
					"ℹ️",
					map[string]interface{}{"user": "test", "count": 42},
				})
				require.NoError(t, err)

				// Test warn logging
				err = bridge.warn(ctx, []interface{}{"Test warning", "⚠️"})
				require.NoError(t, err)

				// Test error logging
				err = bridge.error(ctx, []interface{}{"Test error", "❌"})
				require.NoError(t, err)

				// Test debug logging
				err = bridge.debug(ctx, []interface{}{"Test debug", "🐛"})
				require.NoError(t, err)
			},
		},
		{
			name: "Log level management",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test initial log level
				result, err := bridge.getLogLevel(ctx, []interface{}{})
				require.NoError(t, err)
				level, ok := result.(string)
				require.True(t, ok)
				assert.Equal(t, "basic", level)

				// Test setting log level to detailed
				err = bridge.setLogLevel(ctx, []interface{}{"detailed"})
				require.NoError(t, err)

				result, err = bridge.getLogLevel(ctx, []interface{}{})
				require.NoError(t, err)
				level, ok = result.(string)
				require.True(t, ok)
				assert.Equal(t, "detailed", level)

				// Test setting log level to debug
				err = bridge.setLogLevel(ctx, []interface{}{"debug"})
				require.NoError(t, err)

				result, err = bridge.getLogLevel(ctx, []interface{}{})
				require.NoError(t, err)
				level, ok = result.(string)
				require.True(t, ok)
				assert.Equal(t, "debug", level)
			},
		},
		{
			name: "Logger configuration",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test text format configuration
				err = bridge.configureLogger(ctx, []interface{}{
					map[string]interface{}{
						"format": "text",
						"level":  "info",
					},
				})
				require.NoError(t, err)

				// Test JSON format configuration
				err = bridge.configureLogger(ctx, []interface{}{
					map[string]interface{}{
						"format": "json",
						"level":  "debug",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "Structured attributes",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test creating context with attributes
				result, err := bridge.withAttributes(ctx, []interface{}{
					map[string]interface{}{
						"component":  "test",
						"session_id": "abc123",
						"user_id":    42,
						"is_active":  true,
					},
				})
				require.NoError(t, err)

				contextObj, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "logging_context", contextObj["type"])

				attributes, ok := contextObj["attributes"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", attributes["component"])
				assert.Equal(t, "abc123", attributes["session_id"])
				assert.Equal(t, 42, attributes["user_id"])
				assert.Equal(t, true, attributes["is_active"])
			},
		},
		{
			name: "Logging hook integration",
			test: func(t *testing.T, bridge *SlogBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test BeforeGenerate hook
				err = bridge.logBeforeGenerate(ctx, []interface{}{
					[]interface{}{
						map[string]interface{}{
							"role":    "user",
							"content": "Hello, world!",
						},
						map[string]interface{}{
							"role":    "assistant",
							"content": "Hi there!",
						},
					},
				})
				require.NoError(t, err)

				// Test AfterGenerate hook with success
				err = bridge.logAfterGenerate(ctx, []interface{}{
					map[string]interface{}{
						"content": "Generated response text",
					},
					nil, // no error
				})
				require.NoError(t, err)

				// Test AfterGenerate hook with error
				err = bridge.logAfterGenerate(ctx, []interface{}{
					map[string]interface{}{
						"content": "Partial response",
					},
					"API timeout error",
				})
				require.NoError(t, err)

				// Test BeforeToolCall hook
				err = bridge.logBeforeToolCall(ctx, []interface{}{
					"web_search",
					map[string]interface{}{
						"query":   "golang tutorial",
						"limit":   10,
						"timeout": 5000,
					},
				})
				require.NoError(t, err)

				// Test AfterToolCall hook with success
				err = bridge.logAfterToolCall(ctx, []interface{}{
					"web_search",
					map[string]interface{}{
						"results": []interface{}{
							map[string]interface{}{
								"title": "Go Tutorial",
								"url":   "https://example.com",
							},
						},
						"count": 1,
					},
					nil, // no error
				})
				require.NoError(t, err)

				// Test AfterToolCall hook with error
				err = bridge.logAfterToolCall(ctx, []interface{}{
					"web_search",
					nil, // no result
					"Network timeout",
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewSlogBridge()
			tt.test(t, bridge)
		})
	}
}

// Test slog bridge error scenarios
func TestSlogBridgeErrors(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()

	// Test methods without initialization
	err := bridge.info(ctx, []interface{}{"test message"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameter types
	err = bridge.info(ctx, []interface{}{123}) // message must be string
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.setLogLevel(ctx, []interface{}{123}) // level must be string
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.setLogLevel(ctx, []interface{}{"invalid"}) // invalid level
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")

	err = bridge.configureLogger(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	_, err = bridge.withAttributes(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	// Test logging hook errors
	err = bridge.logBeforeGenerate(ctx, []interface{}{"not an array"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an array")

	err = bridge.logBeforeGenerate(ctx, []interface{}{
		[]interface{}{"not an object"}, // message should be object
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	err = bridge.logAfterGenerate(ctx, []interface{}{"not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	err = bridge.logBeforeToolCall(ctx, []interface{}{123, map[string]interface{}{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")

	err = bridge.logBeforeToolCall(ctx, []interface{}{"tool", "not an object"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an object")

	err = bridge.logAfterToolCall(ctx, []interface{}{123})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

// Test slog bridge lifecycle
func TestSlogBridgeLifecycle(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "slog", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "slog_logger")
	assert.Contains(t, typeMappings, "log_level")
	assert.Contains(t, typeMappings, "logging_hook")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 8)

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"info", "warn", "error", "debug",
		"logBeforeGenerate", "logAfterGenerate",
		"logBeforeToolCall", "logAfterToolCall",
		"setLogLevel", "getLogLevel",
		"configureLogger", "withAttributes",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// Test slog bridge method validation
func TestSlogBridgeValidation(t *testing.T) {
	bridge := NewSlogBridge()
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
			name:        "valid info",
			method:      "info",
			args:        []interface{}{"message"},
			shouldError: false,
		},
		{
			name:        "info missing args",
			method:      "info",
			args:        []interface{}{},
			shouldError: true,
		},
		{
			name:        "valid setLogLevel",
			method:      "setLogLevel",
			args:        []interface{}{"debug"},
			shouldError: false,
		},
		{
			name:        "setLogLevel missing args",
			method:      "setLogLevel",
			args:        []interface{}{},
			shouldError: true,
		},
		{
			name:        "valid getLogLevel",
			method:      "getLogLevel",
			args:        []interface{}{},
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

// Test concurrent slog operations
func TestSlogBridgeConcurrency(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent logging operations
	numOperations := 10
	done := make(chan bool, numOperations*4)

	// Concurrent logging at different levels
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			defer func() { done <- true }()
			err := bridge.info(ctx, []interface{}{fmt.Sprintf("Info message %d", index)})
			assert.NoError(t, err)
		}(i)

		go func(index int) {
			defer func() { done <- true }()
			err := bridge.warn(ctx, []interface{}{fmt.Sprintf("Warn message %d", index)})
			assert.NoError(t, err)
		}(i)

		go func(index int) {
			defer func() { done <- true }()
			err := bridge.error(ctx, []interface{}{fmt.Sprintf("Error message %d", index)})
			assert.NoError(t, err)
		}(i)

		go func(index int) {
			defer func() { done <- true }()
			err := bridge.debug(ctx, []interface{}{fmt.Sprintf("Debug message %d", index)})
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numOperations*4; i++ {
		<-done
	}

	// Verify bridge still works
	result, err := bridge.getLogLevel(ctx, []interface{}{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test structured logging with complex attributes
func TestSlogBridgeStructuredLogging(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test logging with complex attributes
	complexAttributes := map[string]interface{}{
		"string_field":  "value",
		"number_field":  42,
		"boolean_field": true,
		"array_field":   []interface{}{"item1", "item2", "item3"},
		"object_field": map[string]interface{}{
			"nested_field": "nested_value",
			"nested_num":   123,
		},
	}

	// Test info with complex attributes
	err = bridge.info(ctx, []interface{}{
		"Complex structured log",
		"📊",
		complexAttributes,
	})
	require.NoError(t, err)

	// Test creating context with complex attributes
	result, err := bridge.withAttributes(ctx, []interface{}{complexAttributes})
	require.NoError(t, err)

	contextObj, ok := result.(map[string]interface{})
	require.True(t, ok)

	attributes, ok := contextObj["attributes"].(map[string]interface{})
	require.True(t, ok)

	// Verify all attribute types are preserved
	assert.Equal(t, "value", attributes["string_field"])
	assert.Equal(t, 42, attributes["number_field"])
	assert.Equal(t, true, attributes["boolean_field"])

	arrayField, ok := attributes["array_field"].([]interface{})
	require.True(t, ok)
	assert.Len(t, arrayField, 3)

	objectField, ok := attributes["object_field"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "nested_value", objectField["nested_field"])
}

// Test log level changes affect logging hook
func TestSlogBridgeLogLevelIntegration(t *testing.T) {
	bridge := NewSlogBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test that log level changes are applied
	levels := []string{"basic", "detailed", "debug"}

	for _, level := range levels {
		err = bridge.setLogLevel(ctx, []interface{}{level})
		require.NoError(t, err)

		result, err := bridge.getLogLevel(ctx, []interface{}{})
		require.NoError(t, err)
		currentLevel, ok := result.(string)
		require.True(t, ok)
		assert.Equal(t, level, currentLevel)

		// Test that logging still works at the new level
		err = bridge.info(ctx, []interface{}{fmt.Sprintf("Test at %s level", level)})
		require.NoError(t, err)

		// Test logging hook integration
		err = bridge.logBeforeGenerate(ctx, []interface{}{
			[]interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": fmt.Sprintf("Test message at %s level", level),
				},
			},
		})
		require.NoError(t, err)
	}
}
