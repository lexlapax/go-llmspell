// ABOUTME: Tests for pool bridge functionality including provider pooling and health monitoring
// ABOUTME: Comprehensive test coverage for load balancing strategies, object pools, and adaptive pooling

package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// go-llms imports for pool functionality
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// Test PoolBridge core functionality
func TestPoolBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *PoolBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "pool", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "provider pool")
			},
		},
		{
			name: "List pools initially empty",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				result, err := bridge.listPools(ctx, []interface{}{})
				require.NoError(t, err)
				assert.NotNil(t, result)

				pools, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.Empty(t, pools)
			},
		},
		{
			name: "Create pool with mock providers",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Register mock providers
				mockProvider1 := mocks.NewMockProvider("mock1")
				mockProvider2 := mocks.NewMockProvider("mock2")
				bridge.RegisterProvider("mock1", mockProvider1)
				bridge.RegisterProvider("mock2", mockProvider2)

				// Create pool
				result, err := bridge.createPool(ctx, []interface{}{
					"test-pool",
					[]interface{}{"mock1", "mock2"},
					"round_robin",
				})
				require.NoError(t, err)
				assert.NotNil(t, result)

				poolInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-pool", poolInfo["name"])
				assert.Equal(t, "round_robin", poolInfo["strategy"])
				assert.Equal(t, 2, poolInfo["providers"])
			},
		},
		{
			name: "Get and remove pool",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Register mock providers
				mockProvider := mocks.NewMockProvider("mock1")
				bridge.RegisterProvider("mock1", mockProvider)

				// Create pool
				_, err = bridge.createPool(ctx, []interface{}{
					"test-pool",
					[]interface{}{"mock1"},
					"failover",
				})
				require.NoError(t, err)

				// Get pool
				result, err := bridge.getPool(ctx, []interface{}{"test-pool"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				poolInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-pool", poolInfo["name"])

				// Remove pool
				err = bridge.removePool(ctx, []interface{}{"test-pool"})
				require.NoError(t, err)

				// Verify pool is gone
				_, err = bridge.getPool(ctx, []interface{}{"test-pool"})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "pool not found")
			},
		},
		{
			name: "Pool metrics",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Register mock provider
				mockProvider := mocks.NewMockProvider("mock1")
				bridge.RegisterProvider("mock1", mockProvider)

				// Create pool
				_, err = bridge.createPool(ctx, []interface{}{
					"test-pool",
					[]interface{}{"mock1"},
					"fastest",
				})
				require.NoError(t, err)

				// Get metrics
				result, err := bridge.getPoolMetrics(ctx, []interface{}{"test-pool"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				metrics, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, metrics, "provider_0")
			},
		},
		{
			name: "Pool strategies validation",
			test: func(t *testing.T, bridge *PoolBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Register mock provider
				mockProvider := mocks.NewMockProvider("mock1")
				bridge.RegisterProvider("mock1", mockProvider)

				strategies := []string{"round_robin", "failover", "fastest"}

				for _, strategy := range strategies {
					poolName := "test-pool-" + strategy
					_, err := bridge.createPool(ctx, []interface{}{
						poolName,
						[]interface{}{"mock1"},
						strategy,
					})
					require.NoError(t, err, "Strategy %s should be valid", strategy)
				}

				// Test invalid strategy
				_, err = bridge.createPool(ctx, []interface{}{
					"invalid-pool",
					[]interface{}{"mock1"},
					"invalid_strategy",
				})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown strategy")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewPoolBridge()
			tt.test(t, bridge)
		})
	}
}

// Test object pools functionality
func TestObjectPools(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test response pool
	t.Run("Response pool", func(t *testing.T) {
		result, err := bridge.getResponseFromPool(ctx, []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, result)

		responseInfo, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.True(t, responseInfo["pooled"].(bool))
		assert.NotNil(t, responseInfo["response"])

		// Test returning to pool
		err = bridge.returnResponseToPool(ctx, []interface{}{responseInfo})
		require.NoError(t, err)
	})

	// Test token pool
	t.Run("Token pool", func(t *testing.T) {
		result, err := bridge.getTokenFromPool(ctx, []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, result)

		tokenInfo, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.True(t, tokenInfo["pooled"].(bool))
		assert.NotNil(t, tokenInfo["token"])

		// Test returning to pool
		err = bridge.returnTokenToPool(ctx, []interface{}{tokenInfo})
		require.NoError(t, err)
	})

	// Test channel pool
	t.Run("Channel pool", func(t *testing.T) {
		result, err := bridge.getChannelFromPool(ctx, []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, result)

		channelInfo, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.True(t, channelInfo["pooled"].(bool))
		assert.NotEmpty(t, channelInfo["channel_id"])

		// Test returning to pool
		channelID := channelInfo["channel_id"].(string)
		err = bridge.returnChannelToPool(ctx, []interface{}{channelID})
		require.NoError(t, err)
	})
}

// Test pool operations with mock providers
func TestPoolOperations(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create mock provider with predefined response
	mockProvider := mocks.NewMockProvider("mock1")
	mockProvider.OnGenerate = func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "Mock response", nil
	}
	mockProvider.OnGenerateMessage = func(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
		return domain.Response{Content: "Mock message response"}, nil
	}

	bridge.RegisterProvider("mock1", mockProvider)

	// Create pool
	_, err = bridge.createPool(ctx, []interface{}{
		"ops-pool",
		[]interface{}{"mock1"},
		"round_robin",
	})
	require.NoError(t, err)

	// Test text generation
	t.Run("Generate with pool", func(t *testing.T) {
		result, err := bridge.generateWithPool(ctx, []interface{}{
			"ops-pool",
			"Hello world",
			map[string]interface{}{"temperature": 0.7},
		})
		require.NoError(t, err)
		assert.Equal(t, "Mock response", result)
	})

	// Test message generation
	t.Run("Generate message with pool", func(t *testing.T) {
		messages := []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		}

		result, err := bridge.generateMessageWithPool(ctx, []interface{}{
			"ops-pool",
			messages,
			map[string]interface{}{"temperature": 0.5},
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Mock message response", response["content"])
	})
}

// Test pool health monitoring
func TestPoolHealthMonitoring(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create mock providers
	healthyProvider := mocks.NewMockProvider("healthy")
	healthyProvider.OnGenerate = func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "Healthy response", nil
	}

	bridge.RegisterProvider("healthy", healthyProvider)

	// Create pool
	_, err = bridge.createPool(ctx, []interface{}{
		"health-pool",
		[]interface{}{"healthy"},
		"fastest",
	})
	require.NoError(t, err)

	// Generate some traffic to create metrics
	_, err = bridge.generateWithPool(ctx, []interface{}{
		"health-pool",
		"Test prompt",
	})
	require.NoError(t, err)

	// Check health
	result, err := bridge.getProviderHealth(ctx, []interface{}{"health-pool"})
	require.NoError(t, err)
	assert.NotNil(t, result)

	health, ok := result.([]map[string]interface{})
	require.True(t, ok)
	assert.Greater(t, len(health), 0)

	// First provider should be healthy
	firstProvider := health[0]
	assert.Equal(t, "healthy", firstProvider["status"])
	assert.Equal(t, 0, firstProvider["consecutive_errors"])
}

// Test pool integration with go-llms
func TestPoolIntegration(t *testing.T) {
	// Test direct go-llms pool usage
	providers := []domain.Provider{
		mocks.NewMockProvider("test1"),
		mocks.NewMockProvider("test2"),
	}

	pool := llmutil.NewProviderPool(providers, llmutil.StrategyRoundRobin)
	assert.NotNil(t, pool)

	// Test pool metrics
	metrics := pool.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, len(metrics)) // Should have metrics for both providers

	// Test object pools
	responsePool := domain.GetResponsePool()
	assert.NotNil(t, responsePool)

	response := responsePool.Get()
	assert.NotNil(t, response)
	responsePool.Put(response)

	tokenPool := domain.GetTokenPool()
	assert.NotNil(t, tokenPool)

	token := tokenPool.Get()
	assert.NotNil(t, token)
	tokenPool.Put(token)

	channelPool := domain.GetChannelPool()
	assert.NotNil(t, channelPool)

	stream, channel := channelPool.GetResponseStream()
	assert.NotNil(t, stream)
	assert.NotNil(t, channel)
}

// Test error scenarios
func TestPoolBridgeErrors(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()

	// Test methods without initialization
	_, err := bridge.createPool(ctx, []interface{}{"test", []interface{}{}, "round_robin"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createPool(ctx, []interface{}{})
	assert.Error(t, err)

	_, err = bridge.createPool(ctx, []interface{}{123, []interface{}{}, "round_robin"})
	assert.Error(t, err)

	_, err = bridge.getPool(ctx, []interface{}{"nonexistent"})
	assert.Error(t, err)

	err = bridge.removePool(ctx, []interface{}{"nonexistent"})
	assert.Error(t, err)

	// Test pool creation with nonexistent providers
	_, err = bridge.createPool(ctx, []interface{}{
		"test-pool",
		[]interface{}{"nonexistent-provider"},
		"round_robin",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

// Test concurrent operations
func TestPoolBridgeConcurrency(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register multiple mock providers
	for i := 0; i < 5; i++ {
		providerName := fmt.Sprintf("mock%d", i)
		mockProvider := mocks.NewMockProvider(providerName)
		bridge.RegisterProvider(providerName, mockProvider)
	}

	// Create multiple pools concurrently
	numPools := 5
	done := make(chan bool, numPools)

	for i := 0; i < numPools; i++ {
		go func(poolIndex int) {
			defer func() { done <- true }()

			poolName := fmt.Sprintf("concurrent-pool-%d", poolIndex)
			providerName := fmt.Sprintf("mock%d", poolIndex)

			_, err := bridge.createPool(ctx, []interface{}{
				poolName,
				[]interface{}{providerName},
				"round_robin",
			})
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numPools; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Verify all pools were created
	result, err := bridge.listPools(ctx, []interface{}{})
	require.NoError(t, err)

	pools, ok := result.([]map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, numPools, len(pools))
}

// Test pool configuration
func TestPoolConfiguration(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register mock provider
	mockProvider := mocks.NewMockProvider("mock1")
	bridge.RegisterProvider("mock1", mockProvider)

	// Create pool
	_, err = bridge.createPool(ctx, []interface{}{
		"config-pool",
		[]interface{}{"mock1"},
		"round_robin",
	})
	require.NoError(t, err)

	// Test getting configuration (should return defaults)
	result, err := bridge.getPoolConfiguration(ctx, []interface{}{"config-pool"})
	require.NoError(t, err)
	assert.NotNil(t, result)

	config, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "config-pool", config["pool_name"])
	assert.Equal(t, 3, config["error_threshold"])

	// Test setting configuration (not implemented in go-llms, should return error)
	err = bridge.setPoolConfiguration(ctx, []interface{}{
		"config-pool",
		map[string]interface{}{
			"error_threshold": 5,
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

// Test bridge lifecycle
func TestPoolBridgeLifecycle(t *testing.T) {
	bridge := NewPoolBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test metadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "pool", metadata.Name)
	assert.NotEmpty(t, metadata.Dependencies)

	// Test type mappings
	typeMappings := bridge.TypeMappings()
	assert.Contains(t, typeMappings, "provider_pool")
	assert.Contains(t, typeMappings, "pool_metrics")
	assert.Contains(t, typeMappings, "pool_strategy")

	// Test required permissions
	permissions := bridge.RequiredPermissions()
	assert.Greater(t, len(permissions), 0)

	// Test method listing
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 10)

	// Verify specific methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	expectedMethods := []string{
		"createPool",
		"getPoolMetrics",
		"generateWithPool",
		"getResponseFromPool",
	}

	for _, expectedMethod := range expectedMethods {
		assert.True(t, methodNames[expectedMethod], "Method %s should exist", expectedMethod)
	}

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}
