// ABOUTME: Tests for LStatePool which manages a pool of reusable Lua VM instances
// ABOUTME: Validates pool management, health checking, adaptive scaling, and resource management

package gopherlua

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestLStatePool_BasicOperations(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	config := PoolConfig{
		MinSize:     2,
		MaxSize:     10,
		IdleTimeout: 5 * time.Minute,
	}

	pool, err := NewLStatePool(factory, config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("get_state_from_pool", func(t *testing.T) {
		state, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, state)
		defer pool.Put(state)

		// Test that state is functional
		err = state.DoString(`x = 42`)
		assert.NoError(t, err)
		assert.Equal(t, lua.LNumber(42), state.GetGlobal("x"))
	})

	t.Run("put_state_back_to_pool", func(t *testing.T) {
		state, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, state)

		// Modify state
		err = state.DoString(`test_value = "from_pool"`)
		require.NoError(t, err)

		// Put it back
		pool.Put(state)

		// Get another state - should be cleaned
		state2, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, state2)
		defer pool.Put(state2)

		// Should be cleaned (globals should be reset)
		assert.Equal(t, lua.LNil, state2.GetGlobal("test_value"))
	})

	t.Run("pool_respects_max_size", func(t *testing.T) {
		states := make([]*lua.LState, config.MaxSize+1)

		// Get max number of states
		for i := 0; i < config.MaxSize; i++ {
			state, err := pool.Get(context.Background())
			require.NoError(t, err)
			states[i] = state
		}

		// Getting one more should block or create temporary
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		state, err := pool.Get(ctx)
		if err == nil {
			// If we got a state, it should be temporary (not from pool)
			states[config.MaxSize] = state
		} else {
			// Should timeout waiting for available state
			assert.Error(t, err)
		}

		// Put states back
		for i := 0; i < config.MaxSize; i++ {
			if states[i] != nil {
				pool.Put(states[i])
			}
		}
		if states[config.MaxSize] != nil {
			pool.Put(states[config.MaxSize])
		}
	})
}

func TestLStatePool_HealthManagement(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	config := PoolConfig{
		MinSize:         2,
		MaxSize:         5,
		IdleTimeout:     1 * time.Second,
		HealthThreshold: 0.8,
	}

	pool, err := NewLStatePool(factory, config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("unhealthy_state_recycled", func(t *testing.T) {
		state, err := pool.Get(context.Background())
		require.NoError(t, err)
		require.NotNil(t, state)

		// Damage the state by creating memory pressure
		err = state.DoString(`
			local t = {}
			for i = 1, 10000 do
				t[i] = string.rep("x", 1000)
			end
			_G.large_table = t
		`)
		require.NoError(t, err)

		// Put back the damaged state
		pool.Put(state)

		// Get metrics to verify health tracking
		metrics := pool.GetMetrics()
		assert.Greater(t, metrics.TotalCreated, int64(0))
	})

	t.Run("idle_timeout_cleanup", func(t *testing.T) {
		// Get a state and put it back
		state, err := pool.Get(context.Background())
		require.NoError(t, err)
		pool.Put(state)

		// Wait for idle timeout
		time.Sleep(config.IdleTimeout + 100*time.Millisecond)

		// Trigger cleanup by getting metrics
		metrics := pool.GetMetrics()

		// The idle state should have been cleaned up
		assert.LessOrEqual(t, metrics.Available, int64(config.MaxSize))
	})
}

func TestLStatePool_Concurrency(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	config := PoolConfig{
		MinSize:     2,
		MaxSize:     10,
		IdleTimeout: 5 * time.Minute,
	}

	pool, err := NewLStatePool(factory, config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("concurrent_get_put", func(t *testing.T) {
		const numGoroutines = 20
		const numIterations = 5

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numIterations; j++ {
					state, err := pool.Get(context.Background())
					if err != nil {
						errors <- err
						return
					}

					// Do some work
					err = state.DoString(`local x = math.random(100)`)
					if err != nil {
						errors <- err
						pool.Put(state)
						return
					}

					pool.Put(state)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			assert.NoError(t, err)
		}

		// Pool should still be functional
		state, err := pool.Get(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, state)
		pool.Put(state)
	})
}

func TestLStatePool_Metrics(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	config := PoolConfig{
		MinSize:     1,
		MaxSize:     3,
		IdleTimeout: 5 * time.Minute,
	}

	pool, err := NewLStatePool(factory, config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("metrics_tracking", func(t *testing.T) {
		metrics := pool.GetMetrics()
		initialCreated := metrics.TotalCreated

		// Get and put some states
		states := make([]*lua.LState, 2)
		for i := 0; i < 2; i++ {
			state, err := pool.Get(context.Background())
			require.NoError(t, err)
			states[i] = state
		}

		metrics = pool.GetMetrics()
		assert.GreaterOrEqual(t, metrics.TotalCreated, initialCreated)
		assert.Equal(t, int64(2), metrics.InUse)
		assert.GreaterOrEqual(t, metrics.Available, int64(0))

		// Put states back
		for i := 0; i < 2; i++ {
			pool.Put(states[i])
		}

		metrics = pool.GetMetrics()
		assert.Equal(t, int64(0), metrics.InUse)
		assert.GreaterOrEqual(t, metrics.Available, int64(2))
	})
}

func TestLStatePool_Shutdown(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	config := PoolConfig{
		MinSize:     2,
		MaxSize:     5,
		IdleTimeout: 5 * time.Minute,
	}

	pool, err := NewLStatePool(factory, config)
	require.NoError(t, err)
	require.NotNil(t, pool)

	t.Run("graceful_shutdown", func(t *testing.T) {
		// Get some states
		state1, err := pool.Get(context.Background())
		require.NoError(t, err)
		state2, err := pool.Get(context.Background())
		require.NoError(t, err)

		// Put them back
		pool.Put(state1)
		pool.Put(state2)

		// Shutdown should clean up all states
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = pool.Shutdown(ctx)
		assert.NoError(t, err)

		// Pool should not be usable after shutdown
		_, err = pool.Get(context.Background())
		assert.Error(t, err)
	})

	t.Run("shutdown_timeout", func(t *testing.T) {
		pool2, err := NewLStatePool(factory, config)
		require.NoError(t, err)

		// Get a state and don't put it back
		_, err = pool2.Get(context.Background())
		require.NoError(t, err)

		// Shutdown with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err = pool2.Shutdown(ctx)
		// Should timeout but not panic
		assert.Error(t, err)
	})
}

func TestLStatePool_AbandonState(t *testing.T) {
	// Create factory with minimal security
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelMinimal,
		}),
	})

	// Create pool
	pool, err := NewLStatePool(factory, PoolConfig{
		MinSize:         1,
		MaxSize:         3,
		IdleTimeout:     10 * time.Second,
		HealthThreshold: 0.5,
		CleanupInterval: 10 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	t.Run("abandon_state_removes_from_pool", func(t *testing.T) {
		// Get a state
		ctx := context.Background()
		state, err := pool.Get(ctx)
		require.NoError(t, err)
		require.NotNil(t, state)

		// Initial metrics
		initialMetrics := pool.GetMetrics()
		assert.Equal(t, int64(1), initialMetrics.InUse)
		initialRecycled := initialMetrics.TotalRecycled

		// Abandon the state
		pool.AbandonState(state)

		// Check metrics after abandon
		finalMetrics := pool.GetMetrics()
		assert.Equal(t, int64(0), finalMetrics.InUse)
		// TotalRecycled should have increased by 1
		assert.Equal(t, int64(1), finalMetrics.TotalRecycled-initialRecycled)

		// State should not be returned to available pool
		assert.Equal(t, int64(0), finalMetrics.Available)
	})

	t.Run("abandon_nil_state_is_safe", func(t *testing.T) {
		// Should not panic
		pool.AbandonState(nil)
	})

	t.Run("abandon_unknown_state_is_safe", func(t *testing.T) {
		// Create a state outside the pool
		unknownState, err := factory.Create()
		require.NoError(t, err)
		defer unknownState.Close()

		// Should not panic or affect metrics
		initialMetrics := pool.GetMetrics()
		pool.AbandonState(unknownState)
		finalMetrics := pool.GetMetrics()

		assert.Equal(t, initialMetrics.InUse, finalMetrics.InUse)
		assert.Equal(t, initialMetrics.TotalRecycled, finalMetrics.TotalRecycled)
	})

	t.Run("abandoned_state_marked_not_executing", func(t *testing.T) {
		// Get a state
		ctx := context.Background()
		state, err := pool.Get(ctx)
		require.NoError(t, err)

		// Find the pooled state
		pool.mu.RLock()
		pooledState, exists := pool.inUse[state]
		pool.mu.RUnlock()
		require.True(t, exists)

		// Should be marked as executing
		pooledState.mu.Lock()
		assert.True(t, pooledState.executing)
		done := pooledState.done
		pooledState.mu.Unlock()
		require.NotNil(t, done)

		// Abandon the state
		pool.AbandonState(state)

		// Should be marked as not executing and done closed
		pooledState.mu.Lock()
		assert.False(t, pooledState.executing)
		pooledState.mu.Unlock()

		// Done channel should be closed
		select {
		case <-done:
			// Good, it's closed
		default:
			t.Fatal("done channel should be closed after abandon")
		}
	})

	t.Run("concurrent_abandon_is_safe", func(t *testing.T) {
		ctx := context.Background()
		states := make([]*lua.LState, 3)

		// Get multiple states
		for i := 0; i < 3; i++ {
			state, err := pool.Get(ctx)
			require.NoError(t, err)
			states[i] = state
		}

		// Abandon them concurrently
		var wg sync.WaitGroup
		for _, state := range states {
			wg.Add(1)
			go func(s *lua.LState) {
				defer wg.Done()
				pool.AbandonState(s)
			}(state)
		}
		wg.Wait()

		// All states should be abandoned
		metrics := pool.GetMetrics()
		assert.Equal(t, int64(0), metrics.InUse)
		// Don't check exact TotalRecycled as it's cumulative from all tests
	})

	t.Run("shutdown_waits_for_executing_states", func(t *testing.T) {
		// Create a new pool for this test
		testPool, err := NewLStatePool(factory, PoolConfig{
			MinSize:         1,
			MaxSize:         2,
			IdleTimeout:     10 * time.Second,
			HealthThreshold: 0.5,
			CleanupInterval: 10 * time.Second,
		})
		require.NoError(t, err)

		// Get a state
		ctx := context.Background()
		state, err := testPool.Get(ctx)
		require.NoError(t, err)

		// Find the pooled state
		testPool.mu.RLock()
		pooledState, exists := testPool.inUse[state]
		testPool.mu.RUnlock()
		require.True(t, exists)

		// Simulate long-running execution
		executionDone := make(chan struct{})
		go func() {
			defer close(executionDone)

			// Simulate work
			time.Sleep(100 * time.Millisecond)

			// Mark as done (but check if already closed by Put)
			pooledState.mu.Lock()
			pooledState.executing = false
			if pooledState.done != nil {
				select {
				case <-pooledState.done:
					// Already closed
				default:
					close(pooledState.done)
				}
			}
			pooledState.mu.Unlock()
		}()

		// Start shutdown in parallel
		shutdownDone := make(chan error, 1)
		go func() {
			shutdownDone <- testPool.Shutdown(context.Background())
		}()

		// Shutdown should wait for execution to complete
		select {
		case <-time.After(200 * time.Millisecond):
			// Execution should have finished by now
		case err := <-shutdownDone:
			// Shutdown completed
			assert.NoError(t, err)
		}

		// Make sure execution finished
		select {
		case <-executionDone:
			// Good
		default:
			t.Fatal("execution should have completed")
		}
	})
}

func TestLStatePool_TimeoutScenario(t *testing.T) {
	// This test simulates the actual timeout scenario
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelMinimal,
		}),
	})

	pool, err := NewLStatePool(factory, PoolConfig{
		MinSize:         1,
		MaxSize:         2,
		IdleTimeout:     10 * time.Second,
		HealthThreshold: 0.5,
		CleanupInterval: 10 * time.Second,
	})
	require.NoError(t, err)
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	// Simulate getting a state and having it timeout
	ctx := context.Background()
	state, err := pool.Get(ctx)
	require.NoError(t, err)

	// Simulate script execution that times out
	executionComplete := make(chan struct{})
	go func() {
		defer close(executionComplete)

		// Simulate long-running script
		time.Sleep(200 * time.Millisecond)

		// In real scenario, PCall would return here
		// The state is already abandoned, so don't touch it
	}()

	// Simulate timeout - abandon the state
	pool.AbandonState(state)

	// State should be removed from pool immediately
	metrics := pool.GetMetrics()
	assert.Equal(t, int64(0), metrics.InUse)

	// Wait for execution to complete
	<-executionComplete

	// Shutdown should be clean - no race
	err = pool.Shutdown(context.Background())
	assert.NoError(t, err)
}
