// ABOUTME: Tests for error utilities bridge functionality
// ABOUTME: Verifies SerializableError, recovery strategies, aggregation, and categorization

package util

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Import errors package to verify types
	llmerrors "github.com/lexlapax/go-llms/pkg/errors"

	// Use go-llms testutils for consistency with project patterns
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
)

// Test error scenarios using testutils patterns instead of custom types
// This follows go-llms testutils best practices for error testing

func TestUtilErrorsBridge_InterfaceCompliance(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T, *UtilErrorsBridge)
	}{
		{"GetID", func(t *testing.T, b *UtilErrorsBridge) {
			assert.Equal(t, "util_errors", b.GetID())
		}},
		{"GetMetadata", func(t *testing.T, b *UtilErrorsBridge) {
			metadata := b.GetMetadata()
			assert.Equal(t, "util_errors", metadata.Name)
			assert.Equal(t, "2.0.0", metadata.Version)
			assert.Contains(t, metadata.Description, "Error serialization utilities")
		}},
		{"Initialize", func(t *testing.T, b *UtilErrorsBridge) {
			ctx := context.Background()
			err := b.Initialize(ctx)
			assert.NoError(t, err)
			assert.True(t, b.IsInitialized())
		}},
		{"IsInitialized", func(t *testing.T, b *UtilErrorsBridge) {
			assert.False(t, b.IsInitialized())
			ctx := context.Background()
			err := b.Initialize(ctx)
			assert.NoError(t, err)
			assert.True(t, b.IsInitialized())
		}},
		{"Methods", func(t *testing.T, b *UtilErrorsBridge) {
			methods := b.Methods()
			assert.NotEmpty(t, methods)
			// Check for key methods
			methodNames := make(map[string]bool)
			for _, m := range methods {
				methodNames[m.Name] = true
			}
			assert.True(t, methodNames["createError"])
			assert.True(t, methodNames["wrapError"])
			assert.True(t, methodNames["errorToJSON"])
			assert.True(t, methodNames["createExponentialBackoffStrategy"])
			assert.True(t, methodNames["createErrorAggregator"])
			assert.True(t, methodNames["categorizeError"])
			assert.True(t, methodNames["emitErrorEvent"])
		}},
		{"TypeMappings", func(t *testing.T, b *UtilErrorsBridge) {
			mappings := b.TypeMappings()
			assert.NotEmpty(t, mappings)
			assert.Contains(t, mappings, "SerializableError")
			assert.Contains(t, mappings, "RecoveryStrategy")
			assert.Contains(t, mappings, "ErrorAggregator")
			assert.Contains(t, mappings, "ErrorCategory")
			assert.Contains(t, mappings, "ErrorBuilder")
		}},
		{"ValidateMethod", func(t *testing.T, b *UtilErrorsBridge) {
			err := b.ValidateMethod("createError", []interface{}{"test error"})
			assert.NoError(t, err)
		}},
		{"RequiredPermissions", func(t *testing.T, b *UtilErrorsBridge) {
			perms := b.RequiredPermissions()
			assert.NotEmpty(t, perms)
			// Should have memory and storage permissions
			hasMemory := false
			hasStorage := false
			for _, p := range perms {
				if p.Resource == "errors" {
					if string(p.Type) == "memory" {
						hasMemory = true
					} else if string(p.Type) == "storage" {
						hasStorage = true
					}
				}
			}
			assert.True(t, hasMemory)
			assert.True(t, hasStorage)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewUtilErrorsBridge()
			tt.test(t, bridge)
		})
	}
}

func TestUtilErrorsBridge_ErrorCreation(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		validate func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "createError basic",
			method: "createError",
			args:   []interface{}{"test error message"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				baseErr, ok := result.(*llmerrors.BaseError)
				require.True(t, ok)
				assert.Equal(t, "test error message", baseErr.Message)
			},
		},
		{
			name:   "createError with metadata",
			method: "createError",
			args: []interface{}{
				"test error with metadata",
				map[string]interface{}{"key": "value", "number": 42},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				baseErr, ok := result.(*llmerrors.BaseError)
				require.True(t, ok)
				assert.Equal(t, "test error with metadata", baseErr.Message)
				assert.Equal(t, "value", baseErr.GetContext()["key"])
				assert.Equal(t, 42, baseErr.GetContext()["number"])
			},
		},
		{
			name:   "createErrorWithCode",
			method: "createErrorWithCode",
			args:   []interface{}{"ERR_CODE_001", "error with code"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				baseErr, ok := result.(*llmerrors.BaseError)
				require.True(t, ok)
				assert.Equal(t, "error with code", baseErr.Message)
				assert.Equal(t, "ERR_CODE_001", baseErr.Code)
			},
		},
		{
			name:   "wrapError",
			method: "wrapError",
			args:   []interface{}{errors.New("original error"), "wrapper message"},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				baseErr, ok := result.(*llmerrors.BaseError)
				require.True(t, ok)
				assert.Equal(t, "wrapper message", baseErr.Message)
				assert.NotNil(t, baseErr.Cause)
			},
		},
		{
			name:   "enrichError",
			method: "enrichError",
			args: []interface{}{
				errors.New("base error"),
				map[string]interface{}{"enriched": true, "data": "extra"},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				baseErr, ok := result.(*llmerrors.BaseError)
				require.True(t, ok)
				assert.Equal(t, true, baseErr.GetContext()["enriched"])
				assert.Equal(t, "extra", baseErr.GetContext()["data"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			tt.validate(t, result, err)
		})
	}
}

func TestUtilErrorsBridge_ErrorSerialization(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Create an error
	result, err := bridge.ExecuteMethod(ctx, "createErrorWithCode", []interface{}{
		"TEST_ERROR",
		"test serialization",
		map[string]interface{}{"severity": "high"},
	})
	require.NoError(t, err)
	testErr := result.(*llmerrors.BaseError)

	// Serialize to JSON
	jsonResult, err := bridge.ExecuteMethod(ctx, "errorToJSON", []interface{}{testErr})
	require.NoError(t, err)
	jsonStr, ok := jsonResult.(string)
	require.True(t, ok)

	// Verify JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	require.NoError(t, err)
	assert.Equal(t, "TEST_ERROR", parsed["code"])
	assert.Equal(t, "test serialization", parsed["message"])
	assert.NotNil(t, parsed["context"])

	// Deserialize back
	deserResult, err := bridge.ExecuteMethod(ctx, "errorFromJSON", []interface{}{jsonStr})
	require.NoError(t, err)
	require.NotNil(t, deserResult)
}

func TestUtilErrorsBridge_RecoveryStrategies(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		validate func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "createExponentialBackoffStrategy",
			method: "createExponentialBackoffStrategy",
			args:   []interface{}{100.0, 5000.0, 5.0}, // baseDelay, maxDelay, maxRetries
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				strategy, ok := result.(*llmerrors.ExponentialBackoffStrategy)
				require.True(t, ok)
				assert.Equal(t, "exponential_backoff", strategy.Name())
			},
		},
		{
			name:   "createLinearBackoffStrategy",
			method: "createLinearBackoffStrategy",
			args:   []interface{}{1000.0, 3.0}, // delay, maxRetries
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				strategy, ok := result.(*llmerrors.LinearBackoffStrategy)
				require.True(t, ok)
				assert.Equal(t, "linear_backoff", strategy.Name())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			tt.validate(t, result, err)
		})
	}
}

func TestUtilErrorsBridge_ErrorCategorization(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Test default categories
	categories, err := bridge.ExecuteMethod(ctx, "getErrorCategories", []interface{}{})
	require.NoError(t, err)
	catMap, ok := categories.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, catMap, "network")
	assert.Contains(t, catMap, "validation")
	assert.Contains(t, catMap, "authentication")

	// Test categorization
	testCases := []struct {
		errorMsg string
		expected string
	}{
		{"network connection failed", "network"},
		{"invalid input provided", "validation"},
		{"unauthorized access", "authentication"},
		{"rate limit exceeded", "ratelimit"},
		{"internal system error", "system"},
		{"unknown error type", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.errorMsg, func(t *testing.T) {
			err := errors.New(tc.errorMsg)
			result, execErr := bridge.ExecuteMethod(ctx, "categorizeError", []interface{}{err})
			require.NoError(t, execErr)
			category, ok := result.(string)
			require.True(t, ok)
			assert.Equal(t, tc.expected, category)
		})
	}

	// Test custom category registration
	_, err = bridge.ExecuteMethod(ctx, "registerErrorCategory", []interface{}{
		"custom",
		map[string]interface{}{
			"description": "Custom error category",
			"retryable":   false,
			"fatal":       true,
		},
	})
	require.NoError(t, err)

	// Verify custom category was added
	categories, err = bridge.ExecuteMethod(ctx, "getErrorCategories", []interface{}{})
	require.NoError(t, err)
	catMap, ok = categories.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, catMap, "custom")
}

func TestUtilErrorsBridge_ErrorAggregation(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Create aggregator
	aggResult, err := bridge.ExecuteMethod(ctx, "createErrorAggregator", []interface{}{})
	require.NoError(t, err)
	aggregator, ok := aggResult.(llmerrors.ErrorAggregator)
	require.True(t, ok)

	// Add errors
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := errors.New("third error")

	_, err = bridge.ExecuteMethod(ctx, "addError", []interface{}{aggregator, err1})
	require.NoError(t, err)
	_, err = bridge.ExecuteMethod(ctx, "addError", []interface{}{aggregator, err2})
	require.NoError(t, err)
	_, err = bridge.ExecuteMethod(ctx, "addError", []interface{}{aggregator, err3})
	require.NoError(t, err)

	// Test aggregateErrors helper
	aggErr, err := bridge.ExecuteMethod(ctx, "aggregateErrors", []interface{}{
		[]interface{}{err1, err2, err3},
		"Multiple errors occurred during processing",
	})
	require.NoError(t, err)
	require.NotNil(t, aggErr)

	// Verify it's a wrapped error with our message
	baseErr, ok := aggErr.(*llmerrors.BaseError)
	require.True(t, ok)
	assert.Contains(t, baseErr.Message, "Multiple errors occurred during processing")
}

func TestUtilErrorsBridge_ErrorInspection(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Create a retryable error
	result, err := bridge.ExecuteMethod(ctx, "createError", []interface{}{"test error"})
	require.NoError(t, err)
	baseErr := result.(*llmerrors.BaseError)
	baseErr = baseErr.SetRetryable(true)
	baseErr = baseErr.SetFatal(false)

	// Test isRetryableError
	retryable, err := bridge.ExecuteMethod(ctx, "isRetryableError", []interface{}{baseErr})
	require.NoError(t, err)
	assert.True(t, retryable.(bool))

	// Test isFatalError
	fatal, err := bridge.ExecuteMethod(ctx, "isFatalError", []interface{}{baseErr})
	require.NoError(t, err)
	assert.False(t, fatal.(bool))

	// Test getErrorContext
	baseErr = baseErr.WithContext("testKey", "testValue")
	context, err := bridge.ExecuteMethod(ctx, "getErrorContext", []interface{}{baseErr})
	require.NoError(t, err)
	ctxMap, ok := context.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "testValue", ctxMap["testKey"])
}

func TestUtilErrorsBridge_ErrorBuilder(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Create error builder
	builderResult, err := bridge.ExecuteMethod(ctx, "createErrorBuilder", []interface{}{})
	require.NoError(t, err)
	builder, ok := builderResult.(*ErrorBuilder)
	require.True(t, ok)

	// Modify the error through builder
	builder.err.Message = "Built error"
	builder.err.Code = "BUILT_001"
	builder.err = builder.err.SetRetryable(true)
	builder.err = builder.err.WithContext("built", true)

	// Build the error
	builtErr, err := bridge.ExecuteMethod(ctx, "buildError", []interface{}{builder})
	require.NoError(t, err)
	baseErr, ok := builtErr.(*llmerrors.BaseError)
	require.True(t, ok)

	assert.Equal(t, "Built error", baseErr.Message)
	assert.Equal(t, "BUILT_001", baseErr.Code)
	assert.True(t, baseErr.Retryable)
	assert.Equal(t, true, baseErr.GetContext()["built"])
}

func TestUtilErrorsBridge_ThreadSafety(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	// Run concurrent operations
	done := make(chan bool, 3)

	// Concurrent error creation
	go func() {
		for i := 0; i < 10; i++ {
			_, err := bridge.ExecuteMethod(ctx, "createError", []interface{}{
				"concurrent error",
			})
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Concurrent category registration
	go func() {
		for i := 0; i < 10; i++ {
			_, err := bridge.ExecuteMethod(ctx, "registerErrorCategory", []interface{}{
				"test" + string(rune(i)),
				map[string]interface{}{"description": "test"},
			})
			assert.NoError(t, err)
		}
		done <- true
	}()

	// Concurrent aggregation with realistic errors using standard errors
	go func() {
		aggregator, _ := bridge.ExecuteMethod(ctx, "createErrorAggregator", []interface{}{})
		for i := 0; i < 10; i++ {
			// Use different error scenarios for variety
			var err error
			switch i % 3 {
			case 0:
				err = errors.New("Network timeout")
			case 1:
				err = errors.New("Rate limit hit")
			default:
				err = errors.New("Auth failed")
			}
			_, execErr := bridge.ExecuteMethod(ctx, "addError", []interface{}{aggregator, err})
			assert.NoError(t, execErr)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestUtilErrorsBridge_RealisticErrorScenarios(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	t.Run("rate limit scenario", func(t *testing.T) {
		// Use testutils pattern for realistic error creation
		testState := fixtures.BasicTestState()
		testState.Set("error_type", "rate_limit")
		testState.Set("provider", "openai")

		// Create realistic rate limit error using standard errors package
		rateLimitErr := errors.New("rate limit exceeded. Please retry after 60 seconds.")

		// Test categorization
		category, err := bridge.ExecuteMethod(ctx, "categorizeError", []interface{}{rateLimitErr})
		require.NoError(t, err)
		assert.Equal(t, "ratelimit", category.(string))

		// Test retryability (standard errors aren't marked retryable by default)
		retryable, err := bridge.ExecuteMethod(ctx, "isRetryableError", []interface{}{rateLimitErr})
		require.NoError(t, err)
		assert.False(t, retryable.(bool)) // Standard error types aren't retryable by default

		// Test wrapping with context from testutils state
		provider, _ := testState.Get("provider")
		wrapped, err := bridge.ExecuteMethod(ctx, "wrapError", []interface{}{
			rateLimitErr,
			"Rate limit handling",
			map[string]interface{}{
				"retry_after": 60,
				"provider":    provider,
			},
		})
		require.NoError(t, err)
		wrappedErr := wrapped.(*llmerrors.BaseError)
		assert.Equal(t, "Rate limit handling", wrappedErr.Message)
		assert.Equal(t, 60, wrappedErr.GetContext()["retry_after"])
		assert.Equal(t, provider, wrappedErr.GetContext()["provider"])
	})

	t.Run("auth failure scenario", func(t *testing.T) {
		// Use testutils pattern for agent context simulation
		testContext := helpers.CreateTestToolContext()
		requestID := testContext.RunID // Use actual run ID from testutils

		// Create realistic auth error using standard errors package
		authErr := errors.New("Invalid API key provided. Please check your authentication credentials.")

		// Test categorization
		category, err := bridge.ExecuteMethod(ctx, "categorizeError", []interface{}{authErr})
		require.NoError(t, err)
		assert.Equal(t, "authentication", category.(string))

		// Test fatality (auth errors are typically fatal)
		fatal, err := bridge.ExecuteMethod(ctx, "isFatalError", []interface{}{authErr})
		require.NoError(t, err)
		assert.False(t, fatal.(bool)) // Auth errors shouldn't be fatal by default

		// Test enriching with realistic request context from testutils
		enriched, err := bridge.ExecuteMethod(ctx, "enrichError", []interface{}{
			authErr,
			map[string]interface{}{
				"request_id": requestID,
				"endpoint":   "/v1/chat/completions",
				"agent_id":   testContext.Agent.ID,
				"user_id":    "user_789",
			},
		})
		require.NoError(t, err)
		enrichedErr := enriched.(*llmerrors.BaseError)
		assert.Equal(t, requestID, enrichedErr.GetContext()["request_id"])
		assert.Equal(t, testContext.Agent.ID, enrichedErr.GetContext()["agent_id"])
	})

	t.Run("network timeout scenario", func(t *testing.T) {
		// Create realistic network errors using standard errors package
		networkErr := errors.New("Request timeout: Unable to connect to the API endpoint.")
		networkErr2 := errors.New("DNS resolution failed")
		networkErr3 := errors.New("Connection refused")

		// Test categorization
		category, err := bridge.ExecuteMethod(ctx, "categorizeError", []interface{}{networkErr})
		require.NoError(t, err)
		assert.Equal(t, "network", category.(string))

		// Test recovery strategy with realistic parameters
		strategy, err := bridge.ExecuteMethod(ctx, "createExponentialBackoffStrategy", []interface{}{
			1000.0,  // 1 second base delay
			30000.0, // 30 second max delay
			5.0,     // 5 max retries
		})
		require.NoError(t, err)
		require.NotNil(t, strategy)

		// Test aggregating multiple network errors
		aggregated, err := bridge.ExecuteMethod(ctx, "aggregateErrors", []interface{}{
			[]interface{}{networkErr, networkErr2, networkErr3},
			"Multiple network issues encountered",
		})
		require.NoError(t, err)
		require.NotNil(t, aggregated)
	})
}

func TestUtilErrorsBridge_EdgeCases(t *testing.T) {
	bridge := NewUtilErrorsBridge()
	ctx := context.Background()
	require.NoError(t, bridge.Initialize(ctx))

	t.Run("nil error handling", func(t *testing.T) {
		// Categorize nil error
		_, err := bridge.ExecuteMethod(ctx, "categorizeError", []interface{}{nil})
		require.Error(t, err) // Should error on nil
	})

	t.Run("empty aggregator", func(t *testing.T) {
		// Create empty aggregator
		_, _ = bridge.ExecuteMethod(ctx, "createErrorAggregator", []interface{}{})

		// Try to aggregate empty
		result, err := bridge.ExecuteMethod(ctx, "aggregateErrors", []interface{}{
			[]interface{}{},
			"No errors",
		})
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid JSON deserialization", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "errorFromJSON", []interface{}{
			"invalid json string",
		})
		require.Error(t, err)
	})

	t.Run("method not initialized", func(t *testing.T) {
		uninitBridge := NewUtilErrorsBridge()
		_, err := uninitBridge.ExecuteMethod(ctx, "createError", []interface{}{"test"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}
