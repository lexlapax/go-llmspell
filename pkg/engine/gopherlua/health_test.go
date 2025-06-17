// ABOUTME: Tests for state health management system - monitors and evaluates Lua state health
// ABOUTME: Validates health metrics, scoring algorithms, recycling decisions, and monitoring functionality

package gopherlua

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestHealthMetrics_Basic(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	state, err := factory.Create()
	require.NoError(t, err)
	defer state.Close()

	t.Run("initial_health_metrics", func(t *testing.T) {
		monitor := NewHealthMonitor()
		metrics := monitor.GetMetrics(state)

		assert.Greater(t, metrics.Score, 0.8) // Should start healthy
		assert.Equal(t, int64(0), metrics.ExecutionCount)
		assert.Equal(t, int64(0), metrics.ErrorCount)
		assert.Equal(t, time.Duration(0), metrics.TotalExecutionTime)
		assert.True(t, metrics.LastUsed.IsZero())
	})

	t.Run("track_execution", func(t *testing.T) {
		monitor := NewHealthMonitor()

		// Execute some scripts
		start := time.Now()
		err := state.DoString(`x = 42`)
		require.NoError(t, err)

		monitor.RecordExecution(state, time.Since(start), nil)

		metrics := monitor.GetMetrics(state)
		assert.Equal(t, int64(1), metrics.ExecutionCount)
		assert.Equal(t, int64(0), metrics.ErrorCount)
		assert.Greater(t, metrics.TotalExecutionTime, time.Duration(0))
		assert.False(t, metrics.LastUsed.IsZero())
	})

	t.Run("track_errors", func(t *testing.T) {
		monitor := NewHealthMonitor()

		// Execute script with error
		start := time.Now()
		err := state.DoString(`error("test error")`)
		assert.Error(t, err)

		monitor.RecordExecution(state, time.Since(start), err)

		metrics := monitor.GetMetrics(state)
		assert.Equal(t, int64(1), metrics.ExecutionCount)
		assert.Equal(t, int64(1), metrics.ErrorCount)
		assert.Less(t, metrics.Score, 1.0) // Health should decrease
	})
}

func TestHealthScoring_Algorithm(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	state, err := factory.Create()
	require.NoError(t, err)
	defer state.Close()

	monitor := NewHealthMonitor()

	t.Run("healthy_state_high_score", func(t *testing.T) {
		// Simulate successful executions
		for i := 0; i < 10; i++ {
			monitor.RecordExecution(state, 10*time.Millisecond, nil)
		}

		metrics := monitor.GetMetrics(state)
		assert.Greater(t, metrics.Score, 0.8)
		assert.Equal(t, int64(10), metrics.ExecutionCount)
		assert.Equal(t, int64(0), metrics.ErrorCount)
	})

	t.Run("error_prone_state_low_score", func(t *testing.T) {
		// Reset state
		state2, err := factory.Create()
		require.NoError(t, err)
		defer state2.Close()

		// Simulate many errors
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				monitor.RecordExecution(state2, 10*time.Millisecond, nil)
			} else {
				monitor.RecordExecution(state2, 10*time.Millisecond, assert.AnError)
			}
		}

		metrics := monitor.GetMetrics(state2)
		assert.Less(t, metrics.Score, 0.7) // Should be unhealthy
		assert.Equal(t, int64(10), metrics.ExecutionCount)
		assert.Equal(t, int64(5), metrics.ErrorCount)
	})

	t.Run("slow_execution_affects_score", func(t *testing.T) {
		state3, err := factory.Create()
		require.NoError(t, err)
		defer state3.Close()

		// Simulate slow executions
		for i := 0; i < 5; i++ {
			monitor.RecordExecution(state3, 500*time.Millisecond, nil)
		}

		metrics := monitor.GetMetrics(state3)
		assert.Less(t, metrics.Score, 0.9) // Should be somewhat less healthy
		assert.Greater(t, metrics.TotalExecutionTime, 2*time.Second)
	})

	t.Run("age_affects_score", func(t *testing.T) {
		state4, err := factory.Create()
		require.NoError(t, err)
		defer state4.Close()

		// Record old execution
		monitor.RecordExecution(state4, 10*time.Millisecond, nil)

		// Manually set last used to simulate age
		monitor.SetLastUsed(state4, time.Now().Add(-2*time.Hour))

		metrics := monitor.GetMetrics(state4)
		assert.Less(t, metrics.Score, 1.0) // Age should affect score
	})
}

func TestHealthMonitor_StateTracking(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	monitor := NewHealthMonitor()

	t.Run("track_multiple_states", func(t *testing.T) {
		states := make([]*lua.LState, 3)
		for i := 0; i < 3; i++ {
			state, err := factory.Create()
			require.NoError(t, err)
			states[i] = state
			defer state.Close()
		}

		// Record different patterns for each state
		monitor.RecordExecution(states[0], 10*time.Millisecond, nil)
		monitor.RecordExecution(states[1], 50*time.Millisecond, assert.AnError)
		monitor.RecordExecution(states[2], 5*time.Millisecond, nil)

		// Verify each state has independent metrics
		metrics0 := monitor.GetMetrics(states[0])
		metrics1 := monitor.GetMetrics(states[1])
		metrics2 := monitor.GetMetrics(states[2])

		assert.Equal(t, int64(1), metrics0.ExecutionCount)
		assert.Equal(t, int64(0), metrics0.ErrorCount)

		assert.Equal(t, int64(1), metrics1.ExecutionCount)
		assert.Equal(t, int64(1), metrics1.ErrorCount)

		assert.Equal(t, int64(1), metrics2.ExecutionCount)
		assert.Equal(t, int64(0), metrics2.ErrorCount)

		// Health scores should differ
		assert.Greater(t, metrics0.Score, metrics1.Score)
		assert.Greater(t, metrics2.Score, metrics1.Score)
	})

	t.Run("cleanup_closed_state", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)

		monitor.RecordExecution(state, 10*time.Millisecond, nil)
		metrics := monitor.GetMetrics(state)
		assert.Equal(t, int64(1), metrics.ExecutionCount)

		// Close state and cleanup
		state.Close()
		monitor.CleanupState(state)

		// Getting metrics for closed state should return default
		metrics = monitor.GetMetrics(state)
		assert.Equal(t, int64(0), metrics.ExecutionCount)
	})
}

func TestHealthMonitor_HealthDecision(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	monitor := NewHealthMonitor()

	t.Run("healthy_state_should_continue", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)
		defer state.Close()

		// Healthy pattern
		for i := 0; i < 5; i++ {
			monitor.RecordExecution(state, 10*time.Millisecond, nil)
		}

		shouldRecycle := monitor.ShouldRecycle(state, 0.7)
		assert.False(t, shouldRecycle)
	})

	t.Run("unhealthy_state_should_recycle", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)
		defer state.Close()

		// Unhealthy pattern - many errors
		for i := 0; i < 10; i++ {
			monitor.RecordExecution(state, 10*time.Millisecond, assert.AnError)
		}

		shouldRecycle := monitor.ShouldRecycle(state, 0.7)
		assert.True(t, shouldRecycle)
	})

	t.Run("threshold_boundary_testing", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)
		defer state.Close()

		// Create state right at threshold
		monitor.RecordExecution(state, 10*time.Millisecond, nil)
		monitor.RecordExecution(state, 10*time.Millisecond, assert.AnError)
		monitor.RecordExecution(state, 10*time.Millisecond, nil)

		metrics := monitor.GetMetrics(state)
		threshold := metrics.Score + 0.01 // Slightly above current score

		shouldRecycle := monitor.ShouldRecycle(state, threshold)
		assert.True(t, shouldRecycle)

		shouldRecycle = monitor.ShouldRecycle(state, threshold-0.02)
		assert.False(t, shouldRecycle)
	})
}

func TestHealthMonitor_MemoryTracking(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	monitor := NewHealthMonitor()

	t.Run("memory_usage_affects_health", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)
		defer state.Close()

		// Create memory pressure
		err = state.DoString(`
			local t = {}
			for i = 1, 1000 do
				t[i] = string.rep("x", 100)
			end
			large_data = t
		`)
		require.NoError(t, err)

		monitor.RecordExecution(state, 10*time.Millisecond, nil)
		monitor.UpdateMemoryUsage(state, 50*1024*1024) // 50MB

		metrics := monitor.GetMetrics(state)
		assert.Greater(t, metrics.MemoryUsage, int64(0))
		// High memory usage should affect health score
		assert.Less(t, metrics.Score, 1.0)
	})

	t.Run("memory_cleanup_improves_health", func(t *testing.T) {
		state, err := factory.Create()
		require.NoError(t, err)
		defer state.Close()

		// Create and then clear memory
		err = state.DoString(`
			local t = {}
			for i = 1, 1000 do
				t[i] = string.rep("x", 100)
			end
			t = nil
			collectgarbage("collect")
		`)
		require.NoError(t, err)

		monitor.UpdateMemoryUsage(state, 1024*1024) // 1MB after cleanup
		monitor.RecordExecution(state, 10*time.Millisecond, nil)

		metrics := monitor.GetMetrics(state)
		assert.Greater(t, metrics.Score, 0.8) // Should be healthy again
	})
}

func TestHealthMonitor_Concurrency(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	monitor := NewHealthMonitor()

	t.Run("concurrent_state_tracking", func(t *testing.T) {
		const numStates = 10
		const numOperations = 20

		states := make([]*lua.LState, numStates)
		for i := 0; i < numStates; i++ {
			state, err := factory.Create()
			require.NoError(t, err)
			states[i] = state
			defer state.Close()
		}

		// Concurrent operations on different states
		done := make(chan bool, numStates)
		for i := 0; i < numStates; i++ {
			go func(stateIdx int) {
				state := states[stateIdx]
				for j := 0; j < numOperations; j++ {
					var err error
					if j%5 == 0 {
						err = assert.AnError // Simulate occasional errors
					}
					monitor.RecordExecution(state, time.Millisecond, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numStates; i++ {
			<-done
		}

		// Verify all states have correct metrics
		for i := 0; i < numStates; i++ {
			metrics := monitor.GetMetrics(states[i])
			assert.Equal(t, int64(numOperations), metrics.ExecutionCount)
			assert.Equal(t, int64(4), metrics.ErrorCount) // 20/5 = 4 errors per state
			assert.Greater(t, metrics.Score, 0.0)
		}
	})
}
