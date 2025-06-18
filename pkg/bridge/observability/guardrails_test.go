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
				result, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				guardrailInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_guardrail", guardrailMap["name"])
				assert.Equal(t, "input", guardrailMap["type"])
				assert.NotEmpty(t, guardrailMap["id"])
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
				result, err := bridge.ExecuteMethod(ctx, "createGuardrailChain", params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				chainInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				chainMap := chainInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_chain", chainMap["name"])
				assert.Equal(t, "both", chainMap["type"])
				assert.Equal(t, true, chainMap["fail_fast"])
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
				chainResult, err := bridge.ExecuteMethod(ctx, "createGuardrailChain", chainParams)
				require.NoError(t, err)
				chainInfo := chainResult.(engine.ObjectValue)
				chainMap := chainInfo.ToGo().(map[string]interface{})
				chainID := chainMap["id"].(string)

				// Create guardrail
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewBoolValue(true), nil
				}
				guardrailParams := []engine.ScriptValue{
					engine.NewStringValue("test_guardrail"),
					engine.NewStringValue("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Add to chain
				addParams := []engine.ScriptValue{
					engine.NewStringValue(chainID),
					engine.NewStringValue(guardrailID),
				}
				result, err := bridge.ExecuteMethod(ctx, "addGuardrailToChain", addParams)
				require.NoError(t, err)
				assert.NotNil(t, result)
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
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Test valid state
				validState := map[string]engine.ScriptValue{"test_key": engine.NewStringValue("some_value")}
				validateParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(validState),
				}
				result, err := bridge.ExecuteMethod(ctx, "validateGuardrail", validateParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test invalid state
				invalidState := map[string]engine.ScriptValue{"other_key": engine.NewStringValue("some_value")}
				invalidParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(invalidState),
				}
				_, err = bridge.ExecuteMethod(ctx, "validateGuardrail", invalidParams)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "validation failed")
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
				result, err := bridge.ExecuteMethod(ctx, "createRequiredKeysGuardrail", requiredKeysParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test content moderation guardrail
				wordsArray := []engine.ScriptValue{engine.NewStringValue("bad_word"), engine.NewStringValue("prohibited")}
				contentParams := []engine.ScriptValue{
					engine.NewStringValue("content_filter"),
					engine.NewArrayValue(wordsArray),
				}
				result, err = bridge.ExecuteMethod(ctx, "createContentModerationGuardrail", contentParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test message count guardrail
				messageParams := []engine.ScriptValue{
					engine.NewStringValue("message_limit"),
					engine.NewNumberValue(10),
				}
				result, err = bridge.ExecuteMethod(ctx, "createMessageCountGuardrail", messageParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test max state size guardrail
				sizeParams := []engine.ScriptValue{
					engine.NewStringValue("state_size_limit"),
					engine.NewNumberValue(1024),
				}
				result, err = bridge.ExecuteMethod(ctx, "createMaxStateSizeGuardrail", sizeParams)
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
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Test async validation
				state := map[string]engine.ScriptValue{"test": engine.NewStringValue("data")}
				timeout := 1 * time.Second

				asyncParams := []engine.ScriptValue{
					engine.NewStringValue(guardrailID),
					engine.NewObjectValue(state),
					engine.NewNumberValue(timeout.Seconds()),
				}
				result, err := bridge.ExecuteMethod(ctx, "validateGuardrailAsync", asyncParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				resultInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				resultMap := resultInfo.ToGo().(map[string]interface{})
				assert.Contains(t, resultMap, "channel_id")
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
	_, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.ExecuteMethod(ctx, "createGuardrailFunc", []engine.ScriptValue{})
	assert.Error(t, err)

	invalidParams := []engine.ScriptValue{
		engine.NewStringValue("name"),
		engine.NewStringValue("invalid_type"),
		funcValue,
	}
	_, err = bridge.ExecuteMethod(ctx, "createGuardrailFunc", invalidParams)
	assert.Error(t, err)

	validateParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-id"),
		engine.NewObjectValue(map[string]engine.ScriptValue{}),
	}
	_, err = bridge.ExecuteMethod(ctx, "validateGuardrail", validateParams)
	assert.Error(t, err)

	addParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-chain"),
		engine.NewStringValue("invalid-guardrail"),
	}
	_, err = bridge.ExecuteMethod(ctx, "addGuardrailToChain", addParams)
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

			result, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", params)
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
