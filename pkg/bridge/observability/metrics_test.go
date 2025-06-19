// ABOUTME: Tests for metrics bridge functionality including performance monitoring and aggregation
// ABOUTME: Comprehensive test coverage for counters, gauges, timers, and ratio tracking

package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for metrics functionality
	"github.com/lexlapax/go-llms/pkg/util/metrics"
)

// Test MetricsBridge core functionality
func TestMetricsBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *MetricsBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "metrics", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "performance metrics")
			},
		},
		{
			name: "Create and use counter",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create counter
				result, err := bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{
					sv("test_counter"),
				})
				require.NoError(t, err)
				assert.NotNil(t, result)

				counterInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				counterMap := counterInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_counter", counterMap["name"])
				counterID := counterMap["id"].(string)

				// Increment counter
				_, err = bridge.ExecuteMethod(ctx, "incrementCounter", []engine.ScriptValue{
					sv(counterID),
				})
				require.NoError(t, err)

				// Increment by specific value
				_, err = bridge.ExecuteMethod(ctx, "incrementCounterBy", []engine.ScriptValue{
					sv(counterID),
					sv(5),
				})
				require.NoError(t, err)

				// Get counter value
				result, err = bridge.ExecuteMethod(ctx, "getCounterValue", []engine.ScriptValue{
					sv(counterID),
				})
				require.NoError(t, err)
				numValue, ok := result.(engine.NumberValue)
				require.True(t, ok)
				assert.Equal(t, float64(6), numValue.Value())
			},
		},
		{
			name: "Create and use gauge",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create gauge
				result, err := bridge.ExecuteMethod(ctx, "createGauge", []engine.ScriptValue{
					sv("test_gauge"),
				})
				require.NoError(t, err)
				assert.NotNil(t, result)

				gaugeInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				gaugeMap := gaugeInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_gauge", gaugeMap["name"])
				gaugeID := gaugeMap["id"].(string)

				// Set gauge value
				_, err = bridge.ExecuteMethod(ctx, "setGaugeValue", []engine.ScriptValue{
					sv(gaugeID),
					sv(42.5),
				})
				require.NoError(t, err)

				// Increment gauge
				_, err = bridge.ExecuteMethod(ctx, "incrementGauge", []engine.ScriptValue{
					sv(gaugeID),
				})
				require.NoError(t, err)

				// Add to gauge
				_, err = bridge.ExecuteMethod(ctx, "addToGaugeValue", []engine.ScriptValue{
					sv(gaugeID),
					sv(7.5),
				})
				require.NoError(t, err)

				// Get gauge value
				result, err = bridge.ExecuteMethod(ctx, "getGaugeValue", []engine.ScriptValue{
					sv(gaugeID),
				})
				require.NoError(t, err)
				numValue, ok := result.(engine.NumberValue)
				require.True(t, ok)
				assert.Equal(t, float64(51.0), numValue.Value())
			},
		},
		{
			name: "Create and use ratio counter",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create ratio counter
				result, err := bridge.ExecuteMethod(ctx, "createRatioCounter", []engine.ScriptValue{
					sv("test_ratio"),
				})
				require.NoError(t, err)
				assert.NotNil(t, result)

				ratioInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				ratioMap := ratioInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_ratio", ratioMap["name"])
				ratioID := ratioMap["id"].(string)

				// Increment numerator and denominator
				_, err = bridge.ExecuteMethod(ctx, "incrementRatioNumerator", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "incrementRatioNumerator", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "incrementRatioDenominator", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "incrementRatioDenominator", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "incrementRatioDenominator", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)

				// Get ratio
				result, err = bridge.ExecuteMethod(ctx, "getRatio", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				numValue, ok := result.(engine.NumberValue)
				require.True(t, ok)
				assert.InDelta(t, float64(2.0/3.0), numValue.Value(), 0.001)

				// Get raw values
				result, err = bridge.ExecuteMethod(ctx, "getRatioValues", []engine.ScriptValue{
					sv(ratioID),
				})
				require.NoError(t, err)
				values, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				valuesMap := values.ToGo().(map[string]interface{})
				assert.Equal(t, float64(2), valuesMap["numerator"])
				assert.Equal(t, float64(3), valuesMap["denominator"])
			},
		},
		{
			name: "Create and use timer",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create timer
				result, err := bridge.ExecuteMethod(ctx, "createTimer", []engine.ScriptValue{
					sv("test_timer"),
				})
				require.NoError(t, err)
				assert.NotNil(t, result)

				timerInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				timerMap := timerInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test_timer", timerMap["name"])
				timerID := timerMap["id"].(string)

				// Start timer
				_, err = bridge.ExecuteMethod(ctx, "startTimer", []engine.ScriptValue{
					sv(timerID),
				})
				require.NoError(t, err)

				// Simulate some work
				time.Sleep(10 * time.Millisecond)

				// Stop timer
				result, err = bridge.ExecuteMethod(ctx, "stopTimer", []engine.ScriptValue{
					sv(timerID),
				})
				require.NoError(t, err)
				duration, ok := result.(engine.NumberValue)
				require.True(t, ok)
				assert.Greater(t, duration.Value(), float64(0))

				// Record manual duration
				_, err = bridge.ExecuteMethod(ctx, "recordTimerDuration", []engine.ScriptValue{
					sv(timerID),
					sv(0.05), // 50ms
				})
				require.NoError(t, err)

				// Get timer stats
				result, err = bridge.ExecuteMethod(ctx, "getTimerStats", []engine.ScriptValue{
					sv(timerID),
				})
				require.NoError(t, err)
				stats, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				statsMap := stats.ToGo().(map[string]interface{})
				assert.Equal(t, float64(2), statsMap["count"])
				assert.Greater(t, statsMap["total_duration"].(float64), float64(0))
				assert.Greater(t, statsMap["average_duration"].(float64), float64(0))
			},
		},
		{
			name: "Get all metrics",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create various metrics
				_, err = bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{
					sv("counter1"),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "createGauge", []engine.ScriptValue{
					sv("gauge1"),
				})
				require.NoError(t, err)
				_, err = bridge.ExecuteMethod(ctx, "createTimer", []engine.ScriptValue{
					sv("timer1"),
				})
				require.NoError(t, err)

				// Get all metrics
				result, err := bridge.ExecuteMethod(ctx, "getAllMetrics", []engine.ScriptValue{})
				require.NoError(t, err)
				allMetrics, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				metricsMap := allMetrics.ToGo().(map[string]interface{})

				assert.Contains(t, metricsMap, "counters")
				assert.Contains(t, metricsMap, "gauges")
				assert.Contains(t, metricsMap, "timers")
				assert.Contains(t, metricsMap, "ratio_counters")
			},
		},
		{
			name: "Reset metrics",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create and modify counter
				result, err := bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{
					sv("reset_counter"),
				})
				require.NoError(t, err)
				counterInfo := result.(engine.ObjectValue)
				counterMap := counterInfo.ToGo().(map[string]interface{})
				counterID := counterMap["id"].(string)

				_, err = bridge.ExecuteMethod(ctx, "incrementCounterBy", []engine.ScriptValue{
					sv(counterID),
					sv(10),
				})
				require.NoError(t, err)

				// Verify value
				result, err = bridge.ExecuteMethod(ctx, "getCounterValue", []engine.ScriptValue{
					sv(counterID),
				})
				require.NoError(t, err)
				numValue := result.(engine.NumberValue)
				assert.Equal(t, float64(10), numValue.Value())

				// Reset all metrics
				_, err = bridge.ExecuteMethod(ctx, "resetAllMetrics", []engine.ScriptValue{})
				require.NoError(t, err)

				// Counter should be gone - this will return an error
				_, err = bridge.ExecuteMethod(ctx, "incrementCounter", []engine.ScriptValue{
					sv(counterID),
				})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewMetricsBridge()
			tt.test(t, bridge)
		})
	}
}

// Test metrics integration with go-llms
func TestMetricsBridgeIntegration(t *testing.T) {
	bridge := NewMetricsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test direct go-llms metrics usage
	registry := metrics.GetRegistry()

	// Create metrics directly through go-llms
	counter := registry.GetOrCreateCounter("integration_counter")
	gauge := registry.GetOrCreateGauge("integration_gauge")
	timer := registry.GetOrCreateTimer("integration_timer")

	// Use them
	counter.IncrementBy(5)
	gauge.Set(42.0)

	timer.Start()
	time.Sleep(1 * time.Millisecond)
	timer.Stop()

	// Verify values
	assert.Equal(t, int64(5), counter.GetValue())
	assert.Equal(t, float64(42.0), gauge.GetValue())
	assert.Equal(t, int64(1), timer.GetCount())
	assert.Greater(t, timer.GetLastDuration(), time.Duration(0))
}

// Test error scenarios
func TestMetricsBridgeErrors(t *testing.T) {
	bridge := NewMetricsBridge()
	ctx := context.Background()

	// Test methods without initialization
	_, err := bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{
		sv("test"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters - empty args
	_, err = bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least")

	// Test with non-existent counter ID
	_, err = bridge.ExecuteMethod(ctx, "incrementCounter", []engine.ScriptValue{
		sv("invalid-id"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = bridge.ExecuteMethod(ctx, "setGaugeValue", []engine.ScriptValue{
		sv("invalid-id"),
		sv(42),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = bridge.ExecuteMethod(ctx, "startTimer", []engine.ScriptValue{
		sv("invalid-id"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test concurrent operations
func TestMetricsBridgeConcurrency(t *testing.T) {
	bridge := NewMetricsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create counter
	result, err := bridge.ExecuteMethod(ctx, "createCounter", []engine.ScriptValue{
		sv("concurrent_counter"),
	})
	require.NoError(t, err)
	counterInfo := result.(engine.ObjectValue)
	counterMap := counterInfo.ToGo().(map[string]interface{})
	counterID := counterMap["id"].(string)

	// Increment concurrently
	numRoutines := 100
	done := make(chan bool, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func() {
			_, err := bridge.ExecuteMethod(ctx, "incrementCounter", []engine.ScriptValue{
				sv(counterID),
			})
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all increments
	for i := 0; i < numRoutines; i++ {
		<-done
	}

	// Check final value
	result, err = bridge.ExecuteMethod(ctx, "getCounterValue", []engine.ScriptValue{
		sv(counterID),
	})
	require.NoError(t, err)
	numValue := result.(engine.NumberValue)
	assert.Equal(t, float64(numRoutines), numValue.Value())
}
