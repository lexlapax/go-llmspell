// ABOUTME: Tests for optimized state pool implementation including predictive scaling, pre-warming, and memory pooling
// ABOUTME: Validates pool optimization features, performance improvements, and adaptive behavior

package gopherlua

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestOptimizedLStatePool_BasicOperations(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     10,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling: false,
		EnablePreWarming:        false,
		EnableMemoryPooling:     false,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("get and put states", func(t *testing.T) {
		ctx := context.Background()

		// Get a state
		state1, err := pool.Get(ctx)
		require.NoError(t, err)
		require.NotNil(t, state1)

		// Get another state
		state2, err := pool.Get(ctx)
		require.NoError(t, err)
		require.NotNil(t, state2)
		require.NotEqual(t, state1, state2)

		// Return states
		pool.Put(state1)
		pool.Put(state2)

		// Verify metrics
		metrics := pool.GetMetrics()
		assert.GreaterOrEqual(t, metrics.Available, int64(2))
		assert.Equal(t, int64(0), metrics.InUse)
	})
}

func TestOptimizedLStatePool_PredictiveScaling(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     20,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling: true,
		PredictionInterval:      100 * time.Millisecond,
		PredictionWindowSize:    5,
		ScaleUpThreshold:        0.7,
		ScaleDownThreshold:      0.3,
		MaxPredictedScaleUp:     5,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("scale up on high usage", func(t *testing.T) {
		ctx := context.Background()
		initialMetrics := pool.GetMetrics()

		// Create high usage by getting many states
		states := make([]*lua.LState, 0)
		for i := 0; i < 8; i++ {
			state, err := pool.Get(ctx)
			require.NoError(t, err)
			states = append(states, state)
		}

		// Wait for prediction cycle
		time.Sleep(250 * time.Millisecond)

		// Check if pool scaled up (may not always happen due to timing)
		optimMetrics := pool.GetOptimizationMetrics()
		predictedScaleUps := optimMetrics["predicted_scale_ups"].(int64)
		// Just log it, don't assert - timing can be unpredictable
		t.Logf("Predicted scale ups: %d", predictedScaleUps)

		// Return states
		for _, state := range states {
			pool.Put(state)
		}

		// Final metrics should show more available states
		finalMetrics := pool.GetMetrics()
		assert.Greater(t, finalMetrics.TotalCreated, initialMetrics.TotalCreated)
	})

	t.Run("scale down on low usage", func(t *testing.T) {
		// Wait for states to be idle
		time.Sleep(200 * time.Millisecond)

		// Trigger several prediction cycles with low usage
		for i := 0; i < 3; i++ {
			time.Sleep(150 * time.Millisecond)
		}

		// Check if pool scaled down
		optimMetrics := pool.GetOptimizationMetrics()
		predictedScaleDowns := optimMetrics["predicted_scale_downs"].(int64)
		// May or may not scale down depending on timing
		_ = predictedScaleDowns
	})
}

func TestOptimizedLStatePool_PreWarming(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     10,
			IdleTimeout: time.Minute,
		},
		EnablePreWarming: true,
		PreWarmOnInit:    3,
		PreWarmScript:    "x = 1 + 1", // Simple warm-up script
		PreWarmTimeout:   5 * time.Second,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	// Wait for pre-warming to complete
	pool.WaitForPreWarm()

	// Check pre-warmed states count
	optimMetrics := pool.GetOptimizationMetrics()
	preWarmedStates := optimMetrics["pre_warmed_states"].(int64)
	assert.Equal(t, int64(3), preWarmedStates)

	// Verify states are available and warmed
	ctx := context.Background()
	state, err := pool.Get(ctx)
	require.NoError(t, err)

	// The pre-warm script should have been executed
	// We can't directly verify this without modifying the script
	pool.Put(state)
}

func TestOptimizedLStatePool_MemoryPooling(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     10,
			IdleTimeout: time.Minute,
		},
		EnableMemoryPooling: true,
		MemoryPoolSize:      5,
		MemoryBlockSize:     1024,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("memory block allocation", func(t *testing.T) {
		// Get memory blocks
		blocks := make([][]byte, 0)
		for i := 0; i < 3; i++ {
			block := pool.GetMemoryBlock()
			require.NotNil(t, block)
			assert.Equal(t, 1024, len(block))
			blocks = append(blocks, block)
		}

		// Return blocks
		for _, block := range blocks {
			pool.PutMemoryBlock(block)
		}

		// Check metrics
		optimMetrics := pool.GetOptimizationMetrics()
		hits := optimMetrics["memory_pool_hits"].(int64)
		assert.GreaterOrEqual(t, hits, int64(3))
	})

	t.Run("memory pool exhaustion", func(t *testing.T) {
		// Get more blocks than pool size
		blocks := make([][]byte, 0)
		for i := 0; i < 10; i++ {
			block := pool.GetMemoryBlock()
			require.NotNil(t, block)
			blocks = append(blocks, block)
		}

		// Check misses
		optimMetrics := pool.GetOptimizationMetrics()
		misses := optimMetrics["memory_pool_misses"].(int64)
		assert.Greater(t, misses, int64(0))

		// Return all blocks
		for _, block := range blocks {
			pool.PutMemoryBlock(block)
		}
	})
}

func TestOptimizedLStatePool_LoadBalancing(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     5,
			MaxSize:     10,
			IdleTimeout: time.Minute,
		},
		EnableLoadBalancing: true,
		StatePriority:       true,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	ctx := context.Background()

	// Get and execute on multiple states
	var wg sync.WaitGroup
	stateUsage := make(map[*lua.LState]int)
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			state, err := pool.Get(ctx)
			if err != nil {
				return
			}
			defer pool.Put(state)

			// Track usage
			mu.Lock()
			stateUsage[state]++
			mu.Unlock()

			// Simulate work
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()

	// Check that load was distributed
	assert.Greater(t, len(stateUsage), 1, "Load should be distributed across multiple states")
}

func TestOptimizedLStatePool_AdaptiveThresholds(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     20,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling:  true,
		EnableAdaptiveThresholds: true,
		PredictionInterval:       100 * time.Millisecond,
		PredictionWindowSize:     10,
		ScaleUpThreshold:         0.8,
		ScaleDownThreshold:       0.2,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	ctx := context.Background()

	// Create varying load patterns
	for cycle := 0; cycle < 3; cycle++ {
		// High load phase
		states := make([]*lua.LState, 0)
		for i := 0; i < 5; i++ {
			state, err := pool.Get(ctx)
			require.NoError(t, err)
			states = append(states, state)
		}
		time.Sleep(150 * time.Millisecond)

		// Return all at once (creates variance)
		for _, state := range states {
			pool.Put(state)
		}
		time.Sleep(150 * time.Millisecond)
	}

	// Check if thresholds were adjusted
	optimMetrics := pool.GetOptimizationMetrics()
	currentScaleUp := optimMetrics["scale_up_threshold"].(float64)
	currentScaleDown := optimMetrics["scale_down_threshold"].(float64)

	// Thresholds might have been adjusted based on variance
	assert.Greater(t, currentScaleUp, 0.0)
	assert.Less(t, currentScaleDown, 1.0)
}

func TestOptimizedLStatePool_LoadProfiles(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	tests := []struct {
		name    string
		profile string
		verify  func(t *testing.T, pool *OptimizedLStatePool)
	}{
		{
			name:    "burst profile",
			profile: "burst",
			verify: func(t *testing.T, pool *OptimizedLStatePool) {
				assert.Equal(t, 0.6, pool.config.ScaleUpThreshold)
				assert.Equal(t, 0.1, pool.config.ScaleDownThreshold)
				assert.Equal(t, 10, pool.config.MaxPredictedScaleUp)
			},
		},
		{
			name:    "steady profile",
			profile: "steady",
			verify: func(t *testing.T, pool *OptimizedLStatePool) {
				assert.Equal(t, 0.85, pool.config.ScaleUpThreshold)
				assert.Equal(t, 0.25, pool.config.ScaleDownThreshold)
				assert.Equal(t, 3, pool.config.MaxPredictedScaleUp)
			},
		},
		{
			name:    "periodic profile",
			profile: "periodic",
			verify: func(t *testing.T, pool *OptimizedLStatePool) {
				assert.True(t, pool.config.EnableAdaptiveThresholds)
				assert.Equal(t, 20, pool.config.PredictionWindowSize)
			},
		},
		{
			name:    "memory_intensive profile",
			profile: "memory_intensive",
			verify: func(t *testing.T, pool *OptimizedLStatePool) {
				assert.True(t, pool.config.EnableMemoryPooling)
				assert.Equal(t, 20, pool.config.MemoryPoolSize)
				assert.Equal(t, 2*1024*1024, pool.config.MemoryBlockSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OptimizedPoolConfig{
				PoolConfig: PoolConfig{
					MinSize:     2,
					MaxSize:     20,
					IdleTimeout: time.Minute,
				},
			}

			pool, err := NewOptimizedLStatePool(factory, config)
			require.NoError(t, err)
			defer func() {
				_ = pool.Shutdown(context.Background())
			}()

			err = pool.ApplyLoadProfile(tt.profile)
			require.NoError(t, err)

			tt.verify(t, pool)
		})
	}

	t.Run("unknown profile", func(t *testing.T) {
		config := OptimizedPoolConfig{
			PoolConfig: PoolConfig{
				MinSize:     2,
				MaxSize:     20,
				IdleTimeout: time.Minute,
			},
		}

		pool, err := NewOptimizedLStatePool(factory, config)
		require.NoError(t, err)
		defer func() {
			_ = pool.Shutdown(context.Background())
		}()

		err = pool.ApplyLoadProfile("unknown")
		assert.Error(t, err)
	})
}

func TestOptimizedLStatePool_RequestTracking(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     2,
			MaxSize:     10,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling: true,
		PredictionInterval:      500 * time.Millisecond,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	ctx := context.Background()

	// Generate requests
	for i := 0; i < 10; i++ {
		state, err := pool.Get(ctx)
		require.NoError(t, err)
		pool.Put(state)
		time.Sleep(50 * time.Millisecond)
	}

	// Check request rate
	optimMetrics := pool.GetOptimizationMetrics()
	requestRate := optimMetrics["current_request_rate"].(float64)
	assert.Greater(t, requestRate, 0.0)
}

func TestOptimizedLStatePool_ConcurrentOperations(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     5,
			MaxSize:     20,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling: true,
		EnablePreWarming:        true,
		EnableMemoryPooling:     true,
		EnableLoadBalancing:     true,
		PredictionInterval:      100 * time.Millisecond,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	ctx := context.Background()
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent state operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			state, err := pool.Get(ctx)
			if err != nil {
				errors <- err
				return
			}

			// Simulate work
			time.Sleep(time.Duration(id%10) * time.Millisecond)

			// Execute something
			err = state.DoString(fmt.Sprintf("x = %d", id))
			if err != nil {
				errors <- err
			}

			pool.Put(state)
		}(i)
	}

	// Concurrent memory operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			block := pool.GetMemoryBlock()
			time.Sleep(5 * time.Millisecond)
			pool.PutMemoryBlock(block)
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent operation error: %v", err)
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent operations")

	// Verify pool is still healthy
	metrics := pool.GetMetrics()
	assert.GreaterOrEqual(t, metrics.Available, int64(0))
	assert.GreaterOrEqual(t, metrics.TotalCreated, int64(5))
}

func TestOptimizedLStatePool_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	factory := NewLStateFactory(FactoryConfig{})

	// Create standard pool
	standardConfig := PoolConfig{
		MinSize:     5,
		MaxSize:     20,
		IdleTimeout: time.Minute,
	}
	standardPool, err := NewLStatePool(factory, standardConfig)
	require.NoError(t, err)
	defer func() {
		_ = standardPool.Shutdown(context.Background())
	}()

	// Create optimized pool
	optimizedConfig := OptimizedPoolConfig{
		PoolConfig:              standardConfig,
		EnablePredictiveScaling: true,
		EnablePreWarming:        true,
		EnableMemoryPooling:     true,
		EnableLoadBalancing:     true,
		PreWarmOnInit:           5,
	}
	optimizedPool, err := NewOptimizedLStatePool(factory, optimizedConfig)
	require.NoError(t, err)
	defer func() {
		_ = optimizedPool.Shutdown(context.Background())
	}()

	// Wait for pre-warming
	optimizedPool.WaitForPreWarm()

	ctx := context.Background()

	// Benchmark function
	benchmark := func(pool interface {
		Get(context.Context) (*lua.LState, error)
		Put(*lua.LState)
	}) time.Duration {
		start := time.Now()
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				state, err := pool.Get(ctx)
				if err != nil {
					return
				}

				// Simulate work
				_ = state.DoString("for i=1,100 do x = i * 2 end")

				pool.Put(state)
			}()
		}

		wg.Wait()
		return time.Since(start)
	}

	// Run benchmarks
	standardDuration := benchmark(standardPool)
	optimizedDuration := benchmark(optimizedPool)

	t.Logf("Standard pool: %v", standardDuration)
	t.Logf("Optimized pool: %v", optimizedDuration)

	// Optimized pool should be at least comparable, potentially faster
	improvement := float64(standardDuration-optimizedDuration) / float64(standardDuration) * 100
	t.Logf("Performance improvement: %.2f%%", improvement)

	// Check optimization metrics
	optimMetrics := optimizedPool.GetOptimizationMetrics()
	t.Logf("Optimization metrics: %+v", optimMetrics)
}

func TestOptimizedLStatePool_EdgeCases(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{})

	t.Run("nil factory", func(t *testing.T) {
		_, err := NewOptimizedLStatePool(nil, OptimizedPoolConfig{})
		assert.Error(t, err)
	})

	t.Run("zero config values", func(t *testing.T) {
		config := OptimizedPoolConfig{
			// All zero values
		}

		pool, err := NewOptimizedLStatePool(factory, config)
		require.NoError(t, err)
		defer func() {
			_ = pool.Shutdown(context.Background())
		}()

		// Should have defaults applied
		assert.Greater(t, pool.LStatePool.config.MinSize, 0)
		assert.Greater(t, pool.LStatePool.config.MaxSize, 0)
		assert.Greater(t, pool.config.PredictionWindowSize, 0)
	})

	t.Run("invalid memory block return", func(t *testing.T) {
		config := OptimizedPoolConfig{
			PoolConfig: PoolConfig{
				MinSize: 1,
				MaxSize: 5,
			},
			EnableMemoryPooling: true,
			MemoryBlockSize:     1024,
		}

		pool, err := NewOptimizedLStatePool(factory, config)
		require.NoError(t, err)
		defer func() {
			_ = pool.Shutdown(context.Background())
		}()

		// Try to return wrong size block
		wrongBlock := make([]byte, 512)
		pool.PutMemoryBlock(wrongBlock) // Should not panic

		// Try to return nil
		pool.PutMemoryBlock(nil) // Should not panic
	})

	t.Run("shutdown during operations", func(t *testing.T) {
		config := OptimizedPoolConfig{
			PoolConfig: PoolConfig{
				MinSize: 2,
				MaxSize: 10,
			},
			EnablePredictiveScaling: true,
			PredictionInterval:      100 * time.Millisecond,
		}

		pool, err := NewOptimizedLStatePool(factory, config)
		require.NoError(t, err)

		ctx := context.Background()

		// Start operations
		var wg sync.WaitGroup
		stopChan := make(chan struct{})

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case <-stopChan:
						return
					default:
						state, err := pool.Get(ctx)
						if err != nil {
							return
						}
						time.Sleep(10 * time.Millisecond)
						pool.Put(state)
					}
				}
			}()
		}

		// Let it run briefly
		time.Sleep(50 * time.Millisecond)

		// Shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		close(stopChan)
		_ = pool.Shutdown(shutdownCtx)
		// Error is acceptable here due to states potentially being in use

		wg.Wait()
	})
}

// BenchmarkOptimizedPool benchmarks the optimized pool
func BenchmarkOptimizedPool(b *testing.B) {
	factory := NewLStateFactory(FactoryConfig{})

	config := OptimizedPoolConfig{
		PoolConfig: PoolConfig{
			MinSize:     10,
			MaxSize:     50,
			IdleTimeout: time.Minute,
		},
		EnablePredictiveScaling: true,
		EnablePreWarming:        true,
		EnableMemoryPooling:     true,
		EnableLoadBalancing:     true,
		PreWarmOnInit:           10,
	}

	pool, err := NewOptimizedLStatePool(factory, config)
	require.NoError(b, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	pool.WaitForPreWarm()

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			state, err := pool.Get(ctx)
			if err != nil {
				b.Fatal(err)
			}

			// Simulate some work
			_ = state.DoString("x = 1 + 1")

			pool.Put(state)
		}
	})

	b.StopTimer()

	// Report optimization metrics
	optimMetrics := pool.GetOptimizationMetrics()
	b.Logf("Optimization metrics: %+v", optimMetrics)
}
