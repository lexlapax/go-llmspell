// ABOUTME: Tests for guardrails bridge functionality including content filtering and behavioral constraints
// ABOUTME: Comprehensive test coverage for guardrail validation and safety systems

package observability

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// go-llms imports for guardrails functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Test GuardrailsBridge core functionality
func TestGuardrailsBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *GuardrailsBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "guardrails", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "safety system")
			},
		},
		{
			name: "Create function guardrail",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create a simple validation function
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					if len(args) > 0 && args[0].Type() == engine.TypeObject {
						obj := args[0].(engine.ObjectValue)
						if val, exists := obj.Fields()["test_key"]; exists {
							if val.Type() == engine.TypeString && val.(engine.StringValue).Value() == "valid_value" {
								return engine.NewBoolValue(true), nil
							}
						}
					}
					return engine.NewBoolValue(false), nil
				}

				params := []engine.ScriptValue{
					engine.NewStringValue("test_guardrail"),
					engine.NewStringValue("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				result, err := bridge.createGuardrailFunc(ctx, params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				guardrailInfo, ok := result.(map[string]engine.ScriptValue)
				require.True(t, ok)
				assert.Equal(t, "test_guardrail", guardrailInfo["name"].(engine.StringValue).Value())
				assert.Equal(t, "input", guardrailInfo["type"].(engine.StringValue).Value())
				assert.NotEmpty(t, guardrailInfo["id"].(engine.StringValue).Value())
			},
		},
		{
			name: "Create guardrail chain",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				params := []engine.ScriptValue{
					engine.NewStringValue("test_chain"),
					engine.NewStringValue("both"),
					engine.NewBoolValue(true),
				}
				result, err := bridge.createGuardrailChain(ctx, params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				chainInfo, ok := result.(map[string]engine.ScriptValue)
				require.True(t, ok)
				assert.Equal(t, "test_chain", chainInfo["name"].(engine.StringValue).Value())
				assert.Equal(t, "both", chainInfo["type"].(engine.StringValue).Value())
				assert.Equal(t, true, chainInfo["fail_fast"].(engine.BoolValue).Value())
			},
		},
		{
			name: "Add guardrail to chain",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create chain
				chainParams := []engine.ScriptValue{
					engine.NewStringValue("test_chain"),
					engine.NewStringValue("both"),
					engine.NewBoolValue(true),
				}
				chainResult, err := bridge.createGuardrailChain(ctx, chainParams)
				require.NoError(t, err)
				chainInfo := chainResult.(map[string]engine.ScriptValue)
				chainID := chainInfo["id"].(engine.StringValue).Value()

				// Create guardrail
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewBoolValue(true), nil
				}
				guardrailParams := []engine.ScriptValue{
					engine.NewStringValue("test_guardrail"),
					engine.NewStringValue("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.createGuardrailFunc(ctx, guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]engine.ScriptValue)
				guardrailID := guardrailInfo["id"].(engine.StringValue).Value()

				// Add to chain
				addParams := []engine.ScriptValue{
					engine.NewStringValue(chainID),
					engine.NewStringValue(guardrailID),
				}
				err = bridge.addGuardrailToChain(ctx, addParams)
				require.NoError(t, err)
			},
		},
		{
			name: "Validate with guardrail",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create guardrail that validates presence of test_key
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					if len(args) > 0 && args[0].Type() == engine.TypeObject {
						obj := args[0].(engine.ObjectValue)
						_, exists := obj.Fields()["test_key"]
						return engine.NewBoolValue(exists), nil
					}
					return engine.NewBoolValue(false), nil
				}

				guardrailParams := []engine.ScriptValue{
					engine.NewStringValue("test_guardrail"),
					engine.NewStringValue("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.createGuardrailFunc(ctx, guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]engine.ScriptValue)
				guardrailID := guardrailInfo["id"].(engine.StringValue).Value()

				// Test valid state
				validState := map[string]engine.ScriptValue{"test_key": engine.NewStringValue("some_value")}
				validateParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(validState),
				}
				err = bridge.validateGuardrail(ctx, validateParams)
				require.NoError(t, err)

				// Test invalid state
				invalidState := map[string]engine.ScriptValue{"other_key": engine.NewStringValue("some_value")}
				invalidParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(invalidState),
				}
				err = bridge.validateGuardrail(ctx, invalidParams)
				assert.Error(t, err)
			},
		},
		{
			name: "Built-in guardrails",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test required keys guardrail
				keysArray := []engine.ScriptValue{engine.NewStringValue("key1"), engine.NewStringValue("key2")}
				requiredKeysParams := []engine.ScriptValue{
					engine.NewStringValue("required_keys"),
					engine.NewArrayValue(keysArray),
				}
				result, err := bridge.createRequiredKeysGuardrail(ctx, requiredKeysParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test content moderation guardrail
				wordsArray := []engine.ScriptValue{engine.NewStringValue("bad_word"), engine.NewStringValue("prohibited")}
				contentParams := []engine.ScriptValue{
					engine.NewStringValue("content_filter"),
					engine.NewArrayValue(wordsArray),
				}
				result, err = bridge.createContentModerationGuardrail(ctx, contentParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test message count guardrail
				messageParams := []engine.ScriptValue{
					engine.NewStringValue("message_limit"),
					engine.NewNumberValue(10),
				}
				result, err = bridge.createMessageCountGuardrail(ctx, messageParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test max state size guardrail
				sizeParams := []engine.ScriptValue{
					engine.NewStringValue("state_size_limit"),
					engine.NewNumberValue(1024),
				}
				result, err = bridge.createMaxStateSizeGuardrail(ctx, sizeParams)
				require.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "Async validation",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create guardrail
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					// Simulate some processing time
					time.Sleep(10 * time.Millisecond)
					return engine.NewBoolValue(true), nil
				}

				guardrailParams := []engine.ScriptValue{
					engine.NewStringValue("async_guardrail"),
					engine.NewStringValue("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.createGuardrailFunc(ctx, guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]engine.ScriptValue)
				guardrailID := guardrailInfo["id"].(engine.StringValue).Value()

				// Test async validation
				state := map[string]engine.ScriptValue{"test": engine.NewStringValue("data")}
				timeout := 1 * time.Second

				asyncParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(state),
					engine.NewNumberValue(timeout.Seconds()),
				}
				result, err := bridge.validateGuardrailAsync(ctx, asyncParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				resultInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, resultInfo, "channel_id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewGuardrailsBridge()
			tt.test(t, bridge)
		})
	}
}

// Test guardrails integration with go-llms domain
func TestGuardrailsBridgeIntegration(t *testing.T) {
	bridge := NewGuardrailsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a mock state for testing
	mockState := mocks.NewMockState()
	mockState.Set("test_key", "test_value")

	// Test with go-llms RequiredKeysGuardrail
	requiredKeysGuardrail := domain.RequiredKeysGuardrail("test_required", "test_key")
	err = requiredKeysGuardrail.Validate(ctx, mockState.State)
	require.NoError(t, err)

	// Test with missing key
	mockState2 := mocks.NewMockState()
	err = requiredKeysGuardrail.Validate(ctx, mockState2.State)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required keys")

	// Test guardrail chain
	chain := domain.NewGuardrailChain("test_chain", domain.GuardrailTypeInput, true)
	chain.Add(requiredKeysGuardrail)

	err = chain.Validate(ctx, mockState.State)
	require.NoError(t, err)

	err = chain.Validate(ctx, mockState2.State)
	assert.Error(t, err)
}

// Test error scenarios
func TestGuardrailsBridgeErrors(t *testing.T) {
	bridge := NewGuardrailsBridge()
	ctx := context.Background()

	// Test methods without initialization
	funcValue := engine.NewFunctionValue("test", func([]engine.ScriptValue) (engine.ScriptValue, error) {
		return engine.NewBoolValue(true), nil
	})
	params := []engine.ScriptValue{
		engine.NewStringValue("test"),
		engine.NewStringValue("input"),
		funcValue,
	}
	_, err := bridge.createGuardrailFunc(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createGuardrailFunc(ctx, []engine.ScriptValue{})
	assert.Error(t, err)

	invalidParams := []engine.ScriptValue{
		engine.NewStringValue("name"),
		engine.NewStringValue("invalid_type"),
		funcValue,
	}
	_, err = bridge.createGuardrailFunc(ctx, invalidParams)
	assert.Error(t, err)

	validateParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-id"),
		engine.NewObjectValue(map[string]engine.ScriptValue{}),
	}
	err = bridge.validateGuardrail(ctx, validateParams)
	assert.Error(t, err)

	addParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-chain"),
		engine.NewStringValue("invalid-guardrail"),
	}
	err = bridge.addGuardrailToChain(ctx, addParams)
	assert.Error(t, err)
}

// Test concurrent operations
func TestGuardrailsBridgeConcurrency(t *testing.T) {
	bridge := NewGuardrailsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create multiple guardrails concurrently
	numGuardrails := 10
	done := make(chan bool, numGuardrails)

	for i := 0; i < numGuardrails; i++ {
		go func(guardrailNum int) {
			validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil // Always pass
			}

			guardrailName := fmt.Sprintf("concurrent_guardrail_%d", guardrailNum)
			params := []engine.ScriptValue{
				engine.NewStringValue(guardrailName),
				engine.NewStringValue("input"),
				engine.NewFunctionValue("validationFunc", validationFunc),
			}

			result, err := bridge.createGuardrailFunc(ctx, params)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			done <- true
		}(i)
	}

	// Wait for all guardrails to be created
	for i := 0; i < numGuardrails; i++ {
		<-done
	}
}
