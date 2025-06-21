// ABOUTME: Tests for the runner package, covering the core Runner interface and execution logic.
// ABOUTME: Ensures proper script execution, parameter passing, and lifecycle management.

package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Interface(t *testing.T) {
	t.Run("runner_interface_compliance", func(t *testing.T) {
		// Ensure the interface is properly defined
		var _ Runner = (*testRunner)(nil)
	})
}

// testRunner is a mock implementation for testing
type testRunner struct {
	initCalled     bool
	executeCalled  bool
	shutdownCalled bool
	validateCalled bool
	lastScript     string
	lastParams     map[string]interface{}
	executeResult  interface{}
	executeError   error
}

func (r *testRunner) Initialize(ctx context.Context) error {
	r.initCalled = true
	return nil
}

func (r *testRunner) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	r.executeCalled = true
	r.lastScript = script
	r.lastParams = params
	return r.executeResult, r.executeError
}

func (r *testRunner) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	r.executeCalled = true
	r.lastScript = filepath
	r.lastParams = params
	return r.executeResult, r.executeError
}

func (r *testRunner) Validate(script string) error {
	r.validateCalled = true
	r.lastScript = script
	return nil
}

func (r *testRunner) Shutdown() error {
	r.shutdownCalled = true
	return nil
}

func (r *testRunner) GetMetrics() *RunnerMetrics {
	return &RunnerMetrics{
		ScriptsExecuted: 1,
		TotalDuration:   time.Second,
		SuccessCount:    1,
		ErrorCount:      0,
	}
}

func TestRunnerConfig(t *testing.T) {
	t.Run("default_config", func(t *testing.T) {
		config := DefaultRunnerConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 10, config.MaxConcurrentScripts)
		assert.True(t, config.EnableMetrics)
		assert.True(t, config.EnableValidation)
		assert.False(t, config.EnableDebug)
		assert.NotNil(t, config.EngineConfigs)
		assert.NotNil(t, config.SecurityProfiles)
	})

	t.Run("config_validation", func(t *testing.T) {
		tests := []struct {
			name    string
			config  *RunnerConfig
			wantErr bool
		}{
			{
				name:    "valid_config",
				config:  DefaultRunnerConfig(),
				wantErr: false,
			},
			{
				name: "zero_timeout",
				config: &RunnerConfig{
					Timeout:              0,
					MaxConcurrentScripts: 10,
				},
				wantErr: true,
			},
			{
				name: "negative_max_concurrent",
				config: &RunnerConfig{
					Timeout:              time.Second,
					MaxConcurrentScripts: -1,
				},
				wantErr: true,
			},
			{
				name: "zero_max_concurrent",
				config: &RunnerConfig{
					Timeout:              time.Second,
					MaxConcurrentScripts: 0,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.config.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestRunnerOptions(t *testing.T) {
	t.Run("with_timeout", func(t *testing.T) {
		opts := &RunnerOptions{}
		WithTimeout(5 * time.Second)(opts)
		assert.Equal(t, 5*time.Second, opts.Timeout)
	})

	t.Run("with_parameters", func(t *testing.T) {
		params := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		opts := &RunnerOptions{}
		WithParameters(params)(opts)
		assert.Equal(t, params, opts.Parameters)
	})

	t.Run("with_engine", func(t *testing.T) {
		opts := &RunnerOptions{}
		WithEngine("lua")(opts)
		assert.Equal(t, "lua", opts.Engine)
	})

	t.Run("with_security_profile", func(t *testing.T) {
		opts := &RunnerOptions{}
		WithSecurityProfile("sandbox")(opts)
		assert.Equal(t, "sandbox", opts.SecurityProfile)
	})

	t.Run("with_progress_handler", func(t *testing.T) {
		called := false
		handler := func(progress Progress) {
			called = true
		}
		opts := &RunnerOptions{}
		WithProgressHandler(handler)(opts)
		assert.NotNil(t, opts.ProgressHandler)

		// Test that handler is callable
		opts.ProgressHandler(Progress{})
		assert.True(t, called)
	})
}

func TestProgress(t *testing.T) {
	t.Run("progress_creation", func(t *testing.T) {
		p := Progress{
			Stage:       "initialization",
			Message:     "Initializing engine",
			Percentage:  25,
			CurrentStep: 1,
			TotalSteps:  4,
		}

		assert.Equal(t, "initialization", p.Stage)
		assert.Equal(t, "Initializing engine", p.Message)
		assert.Equal(t, 25, p.Percentage)
		assert.Equal(t, 1, p.CurrentStep)
		assert.Equal(t, 4, p.TotalSteps)
	})
}

func TestRunnerMetrics(t *testing.T) {
	t.Run("metrics_creation", func(t *testing.T) {
		metrics := &RunnerMetrics{
			ScriptsExecuted:   10,
			TotalDuration:     5 * time.Minute,
			AverageDuration:   30 * time.Second,
			SuccessCount:      8,
			ErrorCount:        2,
			LastExecutionTime: time.Now(),
			EngineMetrics: map[string]*EngineMetric{
				"lua": {
					ExecutionCount: 6,
					TotalDuration:  3 * time.Minute,
					ErrorCount:     1,
				},
				"javascript": {
					ExecutionCount: 4,
					TotalDuration:  2 * time.Minute,
					ErrorCount:     1,
				},
			},
		}

		assert.Equal(t, int64(10), metrics.ScriptsExecuted)
		assert.Equal(t, int64(8), metrics.SuccessCount)
		assert.Equal(t, int64(2), metrics.ErrorCount)
		assert.Len(t, metrics.EngineMetrics, 2)
		assert.Equal(t, int64(6), metrics.EngineMetrics["lua"].ExecutionCount)
	})

	t.Run("success_rate", func(t *testing.T) {
		metrics := &RunnerMetrics{
			SuccessCount: 75,
			ErrorCount:   25,
		}

		rate := metrics.SuccessRate()
		assert.Equal(t, 0.75, rate)
	})

	t.Run("success_rate_no_executions", func(t *testing.T) {
		metrics := &RunnerMetrics{
			SuccessCount: 0,
			ErrorCount:   0,
		}

		rate := metrics.SuccessRate()
		assert.Equal(t, 0.0, rate)
	})
}

func TestExecutionResult(t *testing.T) {
	t.Run("successful_result", func(t *testing.T) {
		result := &ExecutionResult{
			Value:     "test result",
			Duration:  100 * time.Millisecond,
			Engine:    "lua",
			StartTime: time.Now().Add(-100 * time.Millisecond),
			EndTime:   time.Now(),
			Metadata: map[string]interface{}{
				"version": "1.0",
			},
		}

		assert.Equal(t, "test result", result.Value)
		assert.Equal(t, "lua", result.Engine)
		assert.NotNil(t, result.Metadata)
	})

	t.Run("error_result", func(t *testing.T) {
		result := &ExecutionResult{
			Error:     assert.AnError,
			Duration:  50 * time.Millisecond,
			Engine:    "javascript",
			StartTime: time.Now().Add(-50 * time.Millisecond),
			EndTime:   time.Now(),
		}

		assert.Error(t, result.Error)
		assert.Nil(t, result.Value)
		assert.Equal(t, "javascript", result.Engine)
	})
}

// Integration test placeholder
func TestRunnerLifecycle(t *testing.T) {
	t.Run("full_lifecycle", func(t *testing.T) {
		runner := &testRunner{}
		ctx := context.Background()

		// Initialize
		err := runner.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, runner.initCalled)

		// Execute
		params := map[string]interface{}{"test": true}
		runner.executeResult = "success"
		result, err := runner.Execute(ctx, "test script", params)
		require.NoError(t, err)
		assert.True(t, runner.executeCalled)
		assert.Equal(t, "success", result)
		assert.Equal(t, "test script", runner.lastScript)
		assert.Equal(t, params, runner.lastParams)

		// Validate
		err = runner.Validate("test script")
		require.NoError(t, err)
		assert.True(t, runner.validateCalled)

		// Get metrics
		metrics := runner.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.ScriptsExecuted)

		// Shutdown
		err = runner.Shutdown()
		require.NoError(t, err)
		assert.True(t, runner.shutdownCalled)
	})

	t.Run("context_cancellation", func(t *testing.T) {
		runner := &testRunner{}
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context before execution
		cancel()

		// Initialize should respect context
		err := runner.Initialize(ctx)
		// Mock doesn't check context, but real implementation should
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkRunnerExecution(b *testing.B) {
	runner := &testRunner{
		executeResult: "benchmark result",
	}
	ctx := context.Background()
	params := map[string]interface{}{"benchmark": true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.Execute(ctx, "benchmark script", params)
	}
}

func BenchmarkMetricsCalculation(b *testing.B) {
	metrics := &RunnerMetrics{
		SuccessCount: 1000,
		ErrorCount:   50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.SuccessRate()
	}
}
