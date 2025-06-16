// ABOUTME: Tests for metrics bridge functionality including performance monitoring and aggregation
// ABOUTME: Comprehensive test coverage for counters, gauges, timers, and ratio tracking

package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				result, err := bridge.createCounter(ctx, []interface{}{"test_counter"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				counterInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_counter", counterInfo["name"])
				counterID := counterInfo["id"].(string)

				// Increment counter
				err = bridge.incrementCounter(ctx, []interface{}{counterID})
				require.NoError(t, err)

				// Increment by specific value
				err = bridge.incrementCounterBy(ctx, []interface{}{counterID, float64(5)})
				require.NoError(t, err)

				// Get counter value
				result, err = bridge.getCounterValue(ctx, []interface{}{counterID})
				require.NoError(t, err)
				assert.Equal(t, int64(6), result)
			},
		},
		{
			name: "Create and use gauge",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create gauge
				result, err := bridge.createGauge(ctx, []interface{}{"test_gauge"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				gaugeInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_gauge", gaugeInfo["name"])
				gaugeID := gaugeInfo["id"].(string)

				// Set gauge value
				err = bridge.setGaugeValue(ctx, []interface{}{gaugeID, float64(42.5)})
				require.NoError(t, err)

				// Increment gauge
				err = bridge.incrementGauge(ctx, []interface{}{gaugeID})
				require.NoError(t, err)

				// Add to gauge
				err = bridge.addToGauge(ctx, []interface{}{gaugeID, float64(7.5)})
				require.NoError(t, err)

				// Get gauge value
				result, err = bridge.getGaugeValue(ctx, []interface{}{gaugeID})
				require.NoError(t, err)
				assert.Equal(t, float64(51.0), result)
			},
		},
		{
			name: "Create and use ratio counter",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create ratio counter
				result, err := bridge.createRatioCounter(ctx, []interface{}{"test_ratio"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				ratioInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_ratio", ratioInfo["name"])
				ratioID := ratioInfo["id"].(string)

				// Increment numerator and denominator
				err = bridge.incrementRatioNumerator(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				err = bridge.incrementRatioNumerator(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				err = bridge.incrementRatioDenominator(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				err = bridge.incrementRatioDenominator(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				err = bridge.incrementRatioDenominator(ctx, []interface{}{ratioID})
				require.NoError(t, err)

				// Get ratio
				result, err = bridge.getRatio(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				assert.InDelta(t, float64(2.0/3.0), result, 0.001)

				// Get raw values
				result, err = bridge.getRatioValues(ctx, []interface{}{ratioID})
				require.NoError(t, err)
				values, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, int64(2), values["numerator"])
				assert.Equal(t, int64(3), values["denominator"])
			},
		},
		{
			name: "Create and use timer",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create timer
				result, err := bridge.createTimer(ctx, []interface{}{"test_timer"})
				require.NoError(t, err)
				assert.NotNil(t, result)

				timerInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test_timer", timerInfo["name"])
				timerID := timerInfo["id"].(string)

				// Start timer
				err = bridge.startTimer(ctx, []interface{}{timerID})
				require.NoError(t, err)

				// Simulate some work
				time.Sleep(10 * time.Millisecond)

				// Stop timer
				result, err = bridge.stopTimer(ctx, []interface{}{timerID})
				require.NoError(t, err)
				duration, ok := result.(float64)
				require.True(t, ok)
				assert.Greater(t, duration, float64(0))

				// Record manual duration
				err = bridge.recordTimerDuration(ctx, []interface{}{timerID, float64(0.05)}) // 50ms
				require.NoError(t, err)

				// Get timer stats
				result, err = bridge.getTimerStats(ctx, []interface{}{timerID})
				require.NoError(t, err)
				stats, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, int64(2), stats["count"])
				assert.Greater(t, stats["total_duration"].(float64), float64(0))
				assert.Greater(t, stats["average_duration"].(float64), float64(0))
			},
		},
		{
			name: "Get all metrics",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create various metrics
				_, err = bridge.createCounter(ctx, []interface{}{"counter1"})
				require.NoError(t, err)
				_, err = bridge.createGauge(ctx, []interface{}{"gauge1"})
				require.NoError(t, err)
				_, err = bridge.createTimer(ctx, []interface{}{"timer1"})
				require.NoError(t, err)

				// Get all metrics
				result, err := bridge.getAllMetrics(ctx, []interface{}{})
				require.NoError(t, err)
				allMetrics, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.Contains(t, allMetrics, "counters")
				assert.Contains(t, allMetrics, "gauges")
				assert.Contains(t, allMetrics, "timers")
				assert.Contains(t, allMetrics, "ratio_counters")
			},
		},
		{
			name: "Reset metrics",
			test: func(t *testing.T, bridge *MetricsBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create and modify counter
				result, err := bridge.createCounter(ctx, []interface{}{"reset_counter"})
				require.NoError(t, err)
				counterID := result.(map[string]interface{})["id"].(string)

				err = bridge.incrementCounterBy(ctx, []interface{}{counterID, float64(10)})
				require.NoError(t, err)

				// Verify value
				result, err = bridge.getCounterValue(ctx, []interface{}{counterID})
				require.NoError(t, err)
				assert.Equal(t, int64(10), result)

				// Reset all metrics
				err = bridge.resetAllMetrics(ctx, []interface{}{})
				require.NoError(t, err)

				// Counter should be gone
				err = bridge.incrementCounter(ctx, []interface{}{counterID})
				assert.Error(t, err)
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
	_, err := bridge.createCounter(ctx, []interface{}{"test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createCounter(ctx, []interface{}{})
	assert.Error(t, err)

	err = bridge.incrementCounter(ctx, []interface{}{"invalid-id"})
	assert.Error(t, err)

	err = bridge.setGaugeValue(ctx, []interface{}{"invalid-id", float64(42)})
	assert.Error(t, err)

	err = bridge.startTimer(ctx, []interface{}{"invalid-id"})
	assert.Error(t, err)
}

// Test concurrent operations
func TestMetricsBridgeConcurrency(t *testing.T) {
	bridge := NewMetricsBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create counter
	result, err := bridge.createCounter(ctx, []interface{}{"concurrent_counter"})
	require.NoError(t, err)
	counterID := result.(map[string]interface{})["id"].(string)

	// Increment concurrently
	numRoutines := 100
	done := make(chan bool, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func() {
			err := bridge.incrementCounter(ctx, []interface{}{counterID})
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all increments
	for i := 0; i < numRoutines; i++ {
		<-done
	}

	// Check final value
	result, err = bridge.getCounterValue(ctx, []interface{}{counterID})
	require.NoError(t, err)
	assert.Equal(t, int64(numRoutines), result)
}
