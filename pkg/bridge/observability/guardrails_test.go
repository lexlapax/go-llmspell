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
				validationFunc := func(state interface{}) bool {
					if stateMap, ok := state.(map[string]interface{}); ok {
						if value, exists := stateMap["test_key"]; exists {
							return value == "valid_value"
						}
					}
					return false
				}

				params := []interface{}{"test_guardrail", "input", validationFunc}
				result, err := bridge.createGuardrailFunc(ctx, params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				guardrailInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_guardrail", guardrailInfo["name"])
				assert.Equal(t, "input", guardrailInfo["type"])
				assert.NotEmpty(t, guardrailInfo["id"])
			},
		},
		{
			name: "Create guardrail chain",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				params := []interface{}{"test_chain", "both", true}
				result, err := bridge.createGuardrailChain(ctx, params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				chainInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_chain", chainInfo["name"])
				assert.Equal(t, "both", chainInfo["type"])
				assert.Equal(t, true, chainInfo["fail_fast"])
			},
		},
		{
			name: "Add guardrail to chain",
			test: func(t *testing.T, bridge *GuardrailsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create chain
				chainResult, err := bridge.createGuardrailChain(ctx, []interface{}{"test_chain", "both", true})
				require.NoError(t, err)
				chainInfo := chainResult.(map[string]interface{})
				chainID := chainInfo["id"].(string)

				// Create guardrail
				validationFunc := func(state interface{}) bool { return true }
				guardrailResult, err := bridge.createGuardrailFunc(ctx, []interface{}{"test_guardrail", "input", validationFunc})
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]interface{})
				guardrailID := guardrailInfo["id"].(string)

				// Add to chain
				err = bridge.addGuardrailToChain(ctx, []interface{}{chainID, guardrailID})
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
				validationFunc := func(state interface{}) bool {
					if stateMap, ok := state.(map[string]interface{}); ok {
						_, exists := stateMap["test_key"]
						return exists
					}
					return false
				}

				guardrailResult, err := bridge.createGuardrailFunc(ctx, []interface{}{"test_guardrail", "input", validationFunc})
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]interface{})
				guardrailID := guardrailInfo["id"].(string)

				// Test valid state
				validState := map[string]interface{}{"test_key": "some_value"}
				err = bridge.validateGuardrail(ctx, []interface{}{guardrailID, validState})
				require.NoError(t, err)

				// Test invalid state
				invalidState := map[string]interface{}{"other_key": "some_value"}
				err = bridge.validateGuardrail(ctx, []interface{}{guardrailID, invalidState})
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
				result, err := bridge.createRequiredKeysGuardrail(ctx, []interface{}{"required_keys", []interface{}{"key1", "key2"}})
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test content moderation guardrail
				result, err = bridge.createContentModerationGuardrail(ctx, []interface{}{"content_filter", []interface{}{"bad_word", "prohibited"}})
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test message count guardrail
				result, err = bridge.createMessageCountGuardrail(ctx, []interface{}{"message_limit", float64(10)})
				require.NoError(t, err)
				assert.NotNil(t, result)

				// Test max state size guardrail
				result, err = bridge.createMaxStateSizeGuardrail(ctx, []interface{}{"state_size_limit", float64(1024)})
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
				validationFunc := func(state interface{}) bool {
					// Simulate some processing time
					time.Sleep(10 * time.Millisecond)
					return true
				}

				guardrailResult, err := bridge.createGuardrailFunc(ctx, []interface{}{"async_guardrail", "input", validationFunc})
				require.NoError(t, err)
				guardrailInfo := guardrailResult.(map[string]interface{})
				guardrailID := guardrailInfo["id"].(string)

				// Test async validation
				state := map[string]interface{}{"test": "data"}
				timeout := 1 * time.Second

				result, err := bridge.validateGuardrailAsync(ctx, []interface{}{guardrailID, state, timeout.Seconds()})
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
	_, err := bridge.createGuardrailFunc(ctx, []interface{}{"test", "input", func(interface{}) bool { return true }})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createGuardrailFunc(ctx, []interface{}{})
	assert.Error(t, err)

	_, err = bridge.createGuardrailFunc(ctx, []interface{}{"name", "invalid_type", func(interface{}) bool { return true }})
	assert.Error(t, err)

	err = bridge.validateGuardrail(ctx, []interface{}{"invalid-id", map[string]interface{}{}})
	assert.Error(t, err)

	err = bridge.addGuardrailToChain(ctx, []interface{}{"invalid-chain", "invalid-guardrail"})
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
			validationFunc := func(state interface{}) bool {
				return true // Always pass
			}

			guardrailName := fmt.Sprintf("concurrent_guardrail_%d", guardrailNum)
			params := []interface{}{guardrailName, "input", validationFunc}

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
