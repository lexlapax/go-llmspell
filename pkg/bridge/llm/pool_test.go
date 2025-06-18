// ABOUTME: Tests for the pool bridge that manages provider pools and load balancing
// ABOUTME: Tests pool creation, strategies, metrics, and object pooling functionality

package llm

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolBridge_Initialization(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	assert.NotNil(t, bridge)

	// Test initial state
	assert.False(t, bridge.IsInitialized())
	assert.Equal(t, "pool", bridge.GetID())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "Pool Bridge", metadata.Name)
	assert.NotEmpty(t, metadata.Version)
	assert.NotEmpty(t, metadata.Description)

	// Test initialization
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestPoolBridge_Methods(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	methods := bridge.Methods()

	// Check that we have methods defined
	assert.NotEmpty(t, methods)

	// Check for key methods
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	// Pool management methods
	assert.True(t, methodNames["createPool"])
	assert.True(t, methodNames["getPool"])
	assert.True(t, methodNames["listPools"])
	assert.True(t, methodNames["removePool"])

	// Pool operation methods
	assert.True(t, methodNames["generateWithPool"])
	assert.True(t, methodNames["generateMessageWithPool"])
	assert.True(t, methodNames["streamWithPool"])

	// Metrics methods
	assert.True(t, methodNames["getPoolMetrics"])
	assert.True(t, methodNames["getProviderHealth"])
	assert.True(t, methodNames["resetPoolMetrics"])
}

func TestPoolBridge_CreatePool(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test createPool with round-robin strategy
	providers := []interface{}{"provider1", "provider2", "provider3"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("round_robin"),
	}

	result, err := bridge.ExecuteMethod(ctx, "createPool", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	poolInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "test-pool", poolInfo["name"])
	assert.Equal(t, "round_robin", poolInfo["strategy"])
	providersArray := poolInfo["providers"].([]interface{})
	assert.Len(t, providersArray, 3)
}

func TestPoolBridge_PoolStrategies(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	strategies := []string{"round_robin", "failover", "fastest", "weighted", "least_used"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			providers := []interface{}{"provider1", "provider2"}
			args := []engine.ScriptValue{
				engine.NewStringValue("pool-" + strategy),
				engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
				engine.NewStringValue(strategy),
			}

			result, err := bridge.ExecuteMethod(ctx, "createPool", args)
			assert.NoError(t, err)

			obj, ok := result.(engine.ObjectValue)
			require.True(t, ok)
			poolInfo := obj.ToGo().(map[string]interface{})
			assert.Equal(t, strategy, poolInfo["strategy"])
		})
	}
}

func TestPoolBridge_GetPool(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a pool first
	providers := []interface{}{"provider1", "provider2"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("round_robin"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createPool", args)
	require.NoError(t, err)

	// Test getPool
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getPool", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	poolInfo := obj.ToGo().(map[string]interface{})
	assert.Equal(t, "test-pool", poolInfo["name"])

	// Test non-existent pool
	args = []engine.ScriptValue{
		engine.NewStringValue("non-existent"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getPool", args)
	assert.NoError(t, err)

	errVal, ok := result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "not found")
}

func TestPoolBridge_ListPools(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test empty list
	args := []engine.ScriptValue{}
	result, err := bridge.ExecuteMethod(ctx, "listPools", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	pools := arr.ToGo().([]interface{})
	assert.Len(t, pools, 0)

	// Create some pools
	for i := 0; i < 3; i++ {
		providers := []interface{}{"provider1", "provider2"}
		args := []engine.ScriptValue{
			engine.NewStringValue("pool-" + string(rune('a'+i))),
			engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
			engine.NewStringValue("round_robin"),
		}
		_, err := bridge.ExecuteMethod(ctx, "createPool", args)
		require.NoError(t, err)
	}

	// Test list again
	args = []engine.ScriptValue{}
	result, err = bridge.ExecuteMethod(ctx, "listPools", args)
	assert.NoError(t, err)

	arr, ok = result.(engine.ArrayValue)
	require.True(t, ok)
	pools = arr.ToGo().([]interface{})
	assert.Len(t, pools, 3)
}

func TestPoolBridge_RemovePool(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a pool
	providers := []interface{}{"provider1", "provider2"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("round_robin"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createPool", args)
	require.NoError(t, err)

	// Remove the pool
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err := bridge.ExecuteMethod(ctx, "removePool", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify pool is removed
	result, err = bridge.ExecuteMethod(ctx, "getPool", args)
	assert.NoError(t, err)

	errVal, ok := result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "not found")
}

func TestPoolBridge_PoolMetrics(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a pool
	providers := []interface{}{"provider1", "provider2"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("round_robin"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createPool", args)
	require.NoError(t, err)

	// Test getPoolMetrics
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getPoolMetrics", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	metrics := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, metrics)
}

func TestPoolBridge_ProviderHealth(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a pool
	providers := []interface{}{"provider1", "provider2", "provider3"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("failover"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createPool", args)
	require.NoError(t, err)

	// Test getProviderHealth
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getProviderHealth", args)
	assert.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	health := arr.ToGo().([]interface{})
	assert.Len(t, health, 3)
}

func TestPoolBridge_ObjectPools(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test response pool
	t.Run("ResponsePool", func(t *testing.T) {
		// Get response from pool
		args := []engine.ScriptValue{}
		result, err := bridge.ExecuteMethod(ctx, "getResponseFromPool", args)
		assert.NoError(t, err)

		obj, ok := result.(engine.ObjectValue)
		require.True(t, ok)
		response := obj.ToGo().(map[string]interface{})
		assert.NotEmpty(t, response["id"])
		assert.Contains(t, response["id"], "response-")

		// Return response to pool
		args = []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(response)),
		}
		_, err = bridge.ExecuteMethod(ctx, "returnResponseToPool", args)
		assert.NoError(t, err)
	})

	// Test token pool
	t.Run("TokenPool", func(t *testing.T) {
		// Get token from pool
		args := []engine.ScriptValue{}
		result, err := bridge.ExecuteMethod(ctx, "getTokenFromPool", args)
		assert.NoError(t, err)

		obj, ok := result.(engine.ObjectValue)
		require.True(t, ok)
		token := obj.ToGo().(map[string]interface{})
		assert.False(t, token["used"].(bool))
		assert.NotEmpty(t, token["value"])
		assert.Contains(t, token["value"], "token-")

		// Return token to pool
		args = []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(token)),
		}
		_, err = bridge.ExecuteMethod(ctx, "returnTokenToPool", args)
		assert.NoError(t, err)
	})

	// Test channel pool
	t.Run("ChannelPool", func(t *testing.T) {
		// Get channel from pool
		args := []engine.ScriptValue{}
		result, err := bridge.ExecuteMethod(ctx, "getChannelFromPool", args)
		assert.NoError(t, err)

		obj, ok := result.(engine.ObjectValue)
		require.True(t, ok)
		channel := obj.ToGo().(map[string]interface{})
		assert.NotEmpty(t, channel["id"])
		assert.Equal(t, float64(100), channel["capacity"])

		// Return channel to pool
		args = []engine.ScriptValue{
			engine.NewStringValue(channel["id"].(string)),
		}
		_, err = bridge.ExecuteMethod(ctx, "returnChannelToPool", args)
		assert.NoError(t, err)
	})
}

func TestPoolBridge_PoolConfiguration(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create a pool
	providers := []interface{}{"provider1", "provider2"}
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("round_robin"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createPool", args)
	require.NoError(t, err)

	// Test setPoolConfiguration
	config := map[string]interface{}{
		"maxRetries":          3,
		"retryDelay":          1.5,
		"timeout":             30.0,
		"circuitBreaker":      true,
		"circuitThreshold":    5,
		"healthCheckInterval": 60.0,
	}

	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewObjectValue(engine.ConvertMapToScriptValue(config)),
	}

	result, err := bridge.ExecuteMethod(ctx, "setPoolConfiguration", args)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test getPoolConfiguration
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err = bridge.ExecuteMethod(ctx, "getPoolConfiguration", args)
	assert.NoError(t, err)

	obj, ok := result.(engine.ObjectValue)
	require.True(t, ok)
	poolConfig := obj.ToGo().(map[string]interface{})
	assert.NotNil(t, poolConfig)
}

func TestPoolBridge_ErrorHandling(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Test method call before initialization
	args := []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
	}

	result, err := bridge.ExecuteMethod(ctx, "getPool", args)
	assert.NoError(t, err)

	errVal, ok := result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "not initialized")

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test invalid method
	result, err = bridge.ExecuteMethod(ctx, "invalidMethod", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "unknown method")

	// Test invalid pool strategy
	providers := []interface{}{"provider1"}
	args = []engine.ScriptValue{
		engine.NewStringValue("test-pool"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
		engine.NewStringValue("invalid_strategy"),
	}

	result, err = bridge.ExecuteMethod(ctx, "createPool", args)
	assert.NoError(t, err)

	errVal, ok = result.(engine.ErrorValue)
	require.True(t, ok)
	assert.Contains(t, errVal.Error().Error(), "invalid strategy")
}

func TestPoolBridge_ValidateMethod(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Test valid method with correct args
	args := []engine.ScriptValue{
		engine.NewStringValue("pool-name"),
		engine.NewArrayValue(engine.ConvertSliceToScriptValue([]interface{}{"provider1"})),
		engine.NewStringValue("round_robin"),
	}
	err := bridge.ValidateMethod("createPool", args)
	assert.NoError(t, err)

	// Test invalid method
	err = bridge.ValidateMethod("invalidMethod", args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown method")

	// Test insufficient args
	args = []engine.ScriptValue{
		engine.NewStringValue("pool-name"),
	}
	err = bridge.ValidateMethod("createPool", args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 3 arguments")
}

func TestPoolBridge_TypeMappings(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	mappings := bridge.TypeMappings()

	assert.NotEmpty(t, mappings)

	// Check for key type mappings
	assert.Contains(t, mappings, "ProviderPool")
	assert.Contains(t, mappings, "PoolMetrics")
	assert.Contains(t, mappings, "PoolConfig")

	// Verify mapping structure
	poolMapping := mappings["ProviderPool"]
	assert.NotEmpty(t, poolMapping.GoType)
	assert.Equal(t, "object", poolMapping.ScriptType)
}

func TestPoolBridge_RequiredPermissions(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Check for permissions
	hasNetwork := false
	hasMemory := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionNetwork {
			hasNetwork = true
			assert.Equal(t, "llm.pool", perm.Resource)
		}
		if perm.Type == engine.PermissionMemory {
			hasMemory = true
			assert.Equal(t, "pool", perm.Resource)
		}
	}

	assert.False(t, hasNetwork) // Pool bridge only needs memory permission
	assert.True(t, hasMemory)
}

func TestPoolBridge_Concurrency(t *testing.T) {
	llmBridge := NewLLMBridge()
	bridge := NewPoolBridge(llmBridge)
	ctx := context.Background()

	// Initialize bridge
	require.NoError(t, bridge.Initialize(ctx))

	// Create multiple pools concurrently
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(n int) {
			defer func() { done <- true }()

			providers := []interface{}{"provider1", "provider2"}
			args := []engine.ScriptValue{
				engine.NewStringValue("pool-" + string(rune('a'+n))),
				engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)),
				engine.NewStringValue("round_robin"),
			}

			result, err := bridge.ExecuteMethod(ctx, "createPool", args)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}(i)
	}

	// Wait for all goroutines
	timeout := time.After(5 * time.Second)
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all pools were created
	args := []engine.ScriptValue{}
	result, err := bridge.ExecuteMethod(ctx, "listPools", args)
	require.NoError(t, err)

	arr, ok := result.(engine.ArrayValue)
	require.True(t, ok)
	pools := arr.ToGo().([]interface{})
	assert.Len(t, pools, 5)
}
