// ABOUTME: Tests for error utilities bridge with ScriptValue-based API
// ABOUTME: Validates error serialization, recovery strategies, and categorization

package util

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUtilErrorsBridgeInitialization(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_errors", bridge.GetID())
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

func TestUtilErrorsBridgeMetadata(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_errors", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Error serialization utilities")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilErrorsBridgeMethods(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := []string{
		// Error creation
		"createError",
		"wrapError",
		"createErrorWithCode",
		// Serialization
		"errorToJSON",
		"errorFromJSON",
		// Recovery strategies
		"createExponentialBackoffStrategy",
		"createLinearBackoffStrategy",
		"applyRecoveryStrategy",
		// Categorization
		"categorizeError",
		"registerErrorCategory",
		"getErrorCategories",
		// Aggregation
		"createErrorAggregator",
		"addError",
		"aggregateErrors",
		"getAggregatedErrors",
		// Events
		"emitErrorEvent",
		"subscribeToErrorEvents",
		// Handlers
		"registerErrorHandler",
		"applyErrorHandler",
		// Inspection
		"isRetryableError",
		"isFatalError",
		"getErrorMetadata",
		"getErrorStackTrace",
		// Building
		"createErrorBuilder",
		"buildError",
		// Context
		"enrichError",
		"getErrorContext",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestUtilErrorsBridgeCreateError(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []engine.ScriptValue
		wantErr bool
	}{
		{
			name: "simple error",
			args: []engine.ScriptValue{
				engine.NewStringValue("test error message"),
			},
			wantErr: false,
		},
		{
			name: "error with metadata",
			args: []engine.ScriptValue{
				engine.NewStringValue("test error with metadata"),
				engine.NewObjectValue(map[string]engine.ScriptValue{
					"code":   engine.NewNumberValue(500),
					"module": engine.NewStringValue("auth"),
				}),
			},
			wantErr: false,
		},
		{
			name:    "missing message",
			args:    []engine.ScriptValue{},
			wantErr: true,
		},
		{
			name: "invalid message type",
			args: []engine.ScriptValue{
				engine.NewNumberValue(123),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "createError", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, engine.TypeCustom, result.Type())
			}
		})
	}
}

func TestUtilErrorsBridgeWrapError(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// First create an error to wrap
	origErr, err := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("original error"),
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    []engine.ScriptValue
		wantErr bool
	}{
		{
			name: "wrap error simple",
			args: []engine.ScriptValue{
				origErr,
				engine.NewStringValue("wrapper message"),
			},
			wantErr: false,
		},
		{
			name: "wrap error with metadata",
			args: []engine.ScriptValue{
				origErr,
				engine.NewStringValue("wrapper with metadata"),
				engine.NewObjectValue(map[string]engine.ScriptValue{
					"wrap_level": engine.NewNumberValue(1),
				}),
			},
			wantErr: false,
		},
		{
			name: "missing parameters",
			args: []engine.ScriptValue{
				origErr,
			},
			wantErr: true,
		},
		{
			name: "invalid error type",
			args: []engine.ScriptValue{
				engine.NewStringValue("not an error"),
				engine.NewStringValue("wrapper"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "wrapError", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, engine.TypeCustom, result.Type())
			}
		})
	}
}

func TestUtilErrorsBridgeErrorSerialization(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create an error with metadata
	testErr, err := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("test serialization error"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"code":     engine.NewStringValue("ERR_001"),
			"severity": engine.NewNumberValue(5),
		}),
	})
	require.NoError(t, err)

	// Serialize to JSON
	jsonResult, err := bridge.ExecuteMethod(ctx, "errorToJSON", []engine.ScriptValue{testErr})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, jsonResult.Type())

	jsonStr := jsonResult.(engine.StringValue).Value()
	assert.Contains(t, jsonStr, "test serialization error")
	assert.Contains(t, jsonStr, "ERR_001")

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &jsonData)
	assert.NoError(t, err)

	// Deserialize from JSON
	deserializedErr, err := bridge.ExecuteMethod(ctx, "errorFromJSON", []engine.ScriptValue{
		engine.NewStringValue(jsonStr),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, deserializedErr.Type())
}

func TestUtilErrorsBridgeBackoffStrategies(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	t.Run("exponential backoff", func(t *testing.T) {
		strategy, err := bridge.ExecuteMethod(ctx, "createExponentialBackoffStrategy", []engine.ScriptValue{
			engine.NewNumberValue(100),  // baseDelay
			engine.NewNumberValue(5000), // maxDelay
			engine.NewNumberValue(3),    // maxRetries
		})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeCustom, strategy.Type())
	})

	t.Run("linear backoff", func(t *testing.T) {
		strategy, err := bridge.ExecuteMethod(ctx, "createLinearBackoffStrategy", []engine.ScriptValue{
			engine.NewNumberValue(500), // delay
			engine.NewNumberValue(5),   // maxRetries
		})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeCustom, strategy.Type())
	})

	t.Run("invalid parameters", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "createExponentialBackoffStrategy", []engine.ScriptValue{
			engine.NewStringValue("not a number"),
			engine.NewNumberValue(5000),
			engine.NewNumberValue(3),
		})
		assert.Error(t, err)
	})
}

func TestUtilErrorsBridgeErrorCategorization(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test default categories
	categories, err := bridge.ExecuteMethod(ctx, "getErrorCategories", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, categories.Type())

	catMap := categories.(engine.ObjectValue).Fields()
	assert.Contains(t, catMap, "network")
	assert.Contains(t, catMap, "validation")
	assert.Contains(t, catMap, "authentication")
	assert.Contains(t, catMap, "system")

	// Register custom category
	result, err := bridge.ExecuteMethod(ctx, "registerErrorCategory", []engine.ScriptValue{
		engine.NewStringValue("custom"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"description": engine.NewStringValue("Custom error category"),
			"retryable":   engine.NewBoolValue(true),
			"fatal":       engine.NewBoolValue(false),
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Verify custom category was added
	categories, err = bridge.ExecuteMethod(ctx, "getErrorCategories", []engine.ScriptValue{})
	require.NoError(t, err)
	catMap = categories.(engine.ObjectValue).Fields()
	assert.Contains(t, catMap, "custom")
}

func TestUtilErrorsBridgeErrorAggregation(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create aggregator
	aggregator, err := bridge.ExecuteMethod(ctx, "createErrorAggregator", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, aggregator.Type())

	// Create some errors to aggregate
	err1, _ := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("error 1"),
	})
	err2, _ := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("error 2"),
	})

	// Add errors to aggregator
	result, err := bridge.ExecuteMethod(ctx, "addError", []engine.ScriptValue{
		aggregator,
		err1,
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	result, err = bridge.ExecuteMethod(ctx, "addError", []engine.ScriptValue{
		aggregator,
		err2,
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Test aggregateErrors method
	aggregatedErr, err := bridge.ExecuteMethod(ctx, "aggregateErrors", []engine.ScriptValue{
		engine.NewArrayValue([]engine.ScriptValue{err1, err2}),
		engine.NewStringValue("Multiple errors occurred"),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, aggregatedErr.Type())
}

func TestUtilErrorsBridgeErrorInspection(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create test error
	testErr, err := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("test error"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"retryable": engine.NewBoolValue(true),
			"fatal":     engine.NewBoolValue(false),
		}),
	})
	require.NoError(t, err)

	t.Run("isRetryableError", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "isRetryableError", []engine.ScriptValue{testErr})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeBool, result.Type())
	})

	t.Run("isFatalError", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "isFatalError", []engine.ScriptValue{testErr})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeBool, result.Type())
	})

	t.Run("categorizeError", func(t *testing.T) {
		// Create network error
		networkErr, _ := bridge.ExecuteMethod(ctx, "createErrorWithCode", []engine.ScriptValue{
			engine.NewStringValue("NETWORK_ERROR"),
			engine.NewStringValue("Connection failed"),
		})

		result, err := bridge.ExecuteMethod(ctx, "categorizeError", []engine.ScriptValue{networkErr})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeString, result.Type())
		assert.Equal(t, "network", result.(engine.StringValue).Value())
	})
}

func TestUtilErrorsBridgeErrorContext(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create error with context
	testErr, err := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
		engine.NewStringValue("context test error"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"user_id":    engine.NewNumberValue(123),
			"request_id": engine.NewStringValue("req-456"),
		}),
	})
	require.NoError(t, err)

	// Get error context
	context, err := bridge.ExecuteMethod(ctx, "getErrorContext", []engine.ScriptValue{testErr})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, context.Type())

	contextMap := context.(engine.ObjectValue).Fields()
	assert.Contains(t, contextMap, "user_id")
	assert.Contains(t, contextMap, "request_id")

	// Enrich error with additional context
	enrichedErr, err := bridge.ExecuteMethod(ctx, "enrichError", []engine.ScriptValue{
		testErr,
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"timestamp": engine.NewStringValue("2024-01-01"),
			"severity":  engine.NewNumberValue(5),
		}),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, enrichedErr.Type())
}

func TestUtilErrorsBridgeErrorBuilder(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create error builder
	builder, err := bridge.ExecuteMethod(ctx, "createErrorBuilder", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, builder.Type())

	// Build error
	builtErr, err := bridge.ExecuteMethod(ctx, "buildError", []engine.ScriptValue{builder})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, builtErr.Type())
}

func TestUtilErrorsBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilErrorsBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("createError", []engine.ScriptValue{
		engine.NewStringValue("test"),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestUtilErrorsBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for expected permissions
	hasMemory := false
	hasStorage := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionMemory && perm.Resource == "errors" {
			hasMemory = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
		if perm.Type == engine.PermissionStorage && perm.Resource == "errors" {
			hasStorage = true
			assert.Contains(t, perm.Actions, "emit")
			assert.Contains(t, perm.Actions, "subscribe")
		}
	}

	assert.True(t, hasMemory, "Memory permission not found")
	assert.True(t, hasStorage, "Storage permission not found")
}

func TestUtilErrorsBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"SerializableError",
		"RecoveryStrategy",
		"ErrorAggregator",
		"ErrorCategory",
		"ErrorBuilder",
	}

	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestUtilErrorsBridgeErrorHandling(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()

	// Test method execution before initialization
	_, err := bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{
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
	assert.Contains(t, err.Error(), "method not found")

	// Test invalid arguments
	_, err = bridge.ExecuteMethod(ctx, "createError", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires")
}
