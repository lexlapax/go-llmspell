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
								return sv(true), nil
							}
						}
					}
					return sv(false), nil
				}

				params := []engine.ScriptValue{
					sv("test_guardrail"),
					sv("input"),
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
					sv("test_chain"),
					sv("both"),
					sv(true),
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
					sv("test_chain"),
					sv("both"),
					sv(true),
				}
				chainResult, err := bridge.ExecuteMethod(ctx, "createGuardrailChain", chainParams)
				require.NoError(t, err)
				chainInfo := chainResult.(engine.ObjectValue)
				chainMap := chainInfo.ToGo().(map[string]interface{})
				chainID := chainMap["id"].(string)

				// Create guardrail
				validationFunc := func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return sv(true), nil
				}
				guardrailParams := []engine.ScriptValue{
					sv("test_guardrail"),
					sv("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Add to chain
				addParams := []engine.ScriptValue{
					sv(chainID),
					sv(guardrailID),
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
						return sv(exists), nil
					}
					return sv(false), nil
				}

				guardrailParams := []engine.ScriptValue{
					sv("test_guardrail"),
					sv("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Test valid state
				validState := map[string]interface{}{"test_key": "some_value"}
				validateParams := []engine.ScriptValue{
					sv(guardrailID),
					svMap(validState),
				}
				result, err := bridge.ExecuteMethod(ctx, "validateGuardrail", validateParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test invalid state
				invalidState := map[string]interface{}{"other_key": "some_value"}
				invalidParams := []engine.ScriptValue{
					sv(guardrailID),
					svMap(invalidState),
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
				requiredKeysParams := []engine.ScriptValue{
					sv("required_keys"),
					svArray("key1", "key2"),
				}
				result, err := bridge.ExecuteMethod(ctx, "createRequiredKeysGuardrail", requiredKeysParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test content moderation guardrail
				contentParams := []engine.ScriptValue{
					sv("content_filter"),
					svArray("bad_word", "prohibited"),
				}
				result, err = bridge.ExecuteMethod(ctx, "createContentModerationGuardrail", contentParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test message count guardrail
				messageParams := []engine.ScriptValue{
					sv("message_limit"),
					sv(10),
				}
				result, err = bridge.ExecuteMethod(ctx, "createMessageCountGuardrail", messageParams)
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test max state size guardrail
				sizeParams := []engine.ScriptValue{
					sv("state_size_limit"),
					sv(1024),
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
					return sv(true), nil
				}

				guardrailParams := []engine.ScriptValue{
					sv("async_guardrail"),
					sv("input"),
					engine.NewFunctionValue("validationFunc", validationFunc),
				}
				guardrailResult, err := bridge.ExecuteMethod(ctx, "createGuardrailFunc", guardrailParams)
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(engine.ObjectValue)
				guardrailMap := guardrailInfo.ToGo().(map[string]interface{})
				guardrailID := guardrailMap["id"].(string)

				// Test async validation
				state := map[string]interface{}{"test": "data"}
				timeout := 1 * time.Second

				asyncParams := []engine.ScriptValue{
					sv(guardrailID),
					svMap(state),
					sv(timeout.Seconds()),
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
		return sv(true), nil
	})
	params := []engine.ScriptValue{
		sv("test"),
		sv("input"),
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
		sv("name"),
		sv("invalid_type"),
		funcValue,
	}
	_, err = bridge.ExecuteMethod(ctx, "createGuardrailFunc", invalidParams)
	assert.Error(t, err)

	validateParams := []engine.ScriptValue{
		sv("invalid-id"),
		svMap(map[string]interface{}{}),
	}
	_, err = bridge.ExecuteMethod(ctx, "validateGuardrail", validateParams)
	assert.Error(t, err)

	addParams := []engine.ScriptValue{
		sv("invalid-chain"),
		sv("invalid-guardrail"),
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
				return sv(true), nil // Always pass
			}

			guardrailName := fmt.Sprintf("concurrent_guardrail_%d", guardrailNum)
			params := []engine.ScriptValue{
				sv(guardrailName),
				sv("input"),
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
