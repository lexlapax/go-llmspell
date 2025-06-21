// ABOUTME: Tests for error metrics and tracking functionality, covering counters, rates, and buffering.
// ABOUTME: Ensures thread safety and accurate statistics collection for monitoring purposes.

package errors

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorMetrics_Basic(t *testing.T) {
	t.Run("create_metrics", func(t *testing.T) {
		metrics := NewErrorMetrics()

		assert.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.GetTotalErrors())
		assert.NotNil(t, metrics.categoryCounters)
		assert.NotNil(t, metrics.errorRates)
		assert.NotNil(t, metrics.recentErrors)
	})

	t.Run("record_spell_error", func(t *testing.T) {
		metrics := NewErrorMetrics()

		err := New(CategoryConfig, "config error")
		metrics.RecordError(err)

		assert.Equal(t, int64(1), metrics.GetTotalErrors())
		assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryConfig))
		assert.Equal(t, int64(0), metrics.GetCategoryCount(CategoryScript))
	})

	t.Run("record_generic_error", func(t *testing.T) {
		metrics := NewErrorMetrics()

		err := errors.New("generic error")
		metrics.RecordError(err)

		assert.Equal(t, int64(1), metrics.GetTotalErrors())
		assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryUnknown))
	})

	t.Run("record_nil_error", func(t *testing.T) {
		metrics := NewErrorMetrics()

		metrics.RecordError(nil)

		assert.Equal(t, int64(0), metrics.GetTotalErrors())
	})
}

func TestErrorMetrics_Categories(t *testing.T) {
	t.Run("multiple_categories", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Record different error categories
		metrics.RecordError(New(CategoryConfig, "config"))
		metrics.RecordError(New(CategoryConfig, "config2"))
		metrics.RecordError(New(CategoryScript, "script"))
		metrics.RecordError(New(CategoryEngine, "engine"))
		metrics.RecordError(New(CategoryEngine, "engine2"))
		metrics.RecordError(New(CategoryEngine, "engine3"))

		assert.Equal(t, int64(6), metrics.GetTotalErrors())
		assert.Equal(t, int64(2), metrics.GetCategoryCount(CategoryConfig))
		assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryScript))
		assert.Equal(t, int64(3), metrics.GetCategoryCount(CategoryEngine))
		assert.Equal(t, int64(0), metrics.GetCategoryCount(CategoryNetwork))
	})

	t.Run("all_categories_initialized", func(t *testing.T) {
		metrics := NewErrorMetrics()

		categories := []ErrorCategory{
			CategoryUnknown, CategoryUsage, CategoryConfig, CategoryScript,
			CategoryEngine, CategorySecurity, CategoryNetwork, CategoryTimeout,
			CategoryResource, CategoryValidation, CategoryDependency, CategoryIO,
			CategoryInterrupted,
		}

		for _, cat := range categories {
			assert.Equal(t, int64(0), metrics.GetCategoryCount(cat))
			assert.Equal(t, float64(0), metrics.GetErrorRate(cat))
		}
	})
}

func TestErrorMetrics_Stats(t *testing.T) {
	t.Run("get_stats", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Record various errors
		metrics.RecordError(New(CategoryConfig, "config"))
		metrics.RecordError(New(CategoryConfig, "config2"))
		metrics.RecordError(New(CategoryScript, "script"))
		metrics.RecordError(New(CategoryEngine, "engine"))

		stats := metrics.GetStats()

		assert.Equal(t, int64(4), stats.TotalErrors)
		assert.True(t, stats.Uptime > 0)
		assert.NotNil(t, stats.Categories)

		configStats := stats.Categories[CategoryConfig]
		assert.NotNil(t, configStats)
		assert.Equal(t, int64(2), configStats.Count)
		assert.Equal(t, float64(50), configStats.Percentage)

		scriptStats := stats.Categories[CategoryScript]
		assert.NotNil(t, scriptStats)
		assert.Equal(t, int64(1), scriptStats.Count)
		assert.Equal(t, float64(25), scriptStats.Percentage)
	})

	t.Run("empty_stats", func(t *testing.T) {
		metrics := NewErrorMetrics()
		stats := metrics.GetStats()

		assert.Equal(t, int64(0), stats.TotalErrors)
		assert.True(t, stats.Uptime > 0)

		for _, catStats := range stats.Categories {
			assert.Equal(t, int64(0), catStats.Count)
			assert.True(t, catStats.Percentage == 0 || catStats.Percentage != catStats.Percentage) // 0 or NaN
			assert.Equal(t, float64(0), catStats.Rate)
		}
	})
}

func TestErrorMetrics_Reset(t *testing.T) {
	t.Run("reset_metrics", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Record some errors
		metrics.RecordError(New(CategoryConfig, "error1"))
		metrics.RecordError(New(CategoryScript, "error2"))
		metrics.RecordError(New(CategoryEngine, "error3"))

		assert.Equal(t, int64(3), metrics.GetTotalErrors())

		// Reset
		metrics.Reset()

		assert.Equal(t, int64(0), metrics.GetTotalErrors())
		assert.Equal(t, int64(0), metrics.GetCategoryCount(CategoryConfig))
		assert.Equal(t, int64(0), metrics.GetCategoryCount(CategoryScript))
		assert.Equal(t, int64(0), metrics.GetCategoryCount(CategoryEngine))

		// Recent errors should be cleared
		recent := metrics.GetRecentErrors(10)
		assert.Len(t, recent, 0)
	})
}

func TestRateTracker(t *testing.T) {
	t.Run("basic_rate", func(t *testing.T) {
		tracker := NewRateTracker(time.Minute, 60)

		// Record some events
		for i := 0; i < 10; i++ {
			tracker.Record()
		}

		rate := tracker.GetRate()
		assert.True(t, rate > 0, "Rate should be positive after recording events")
	})

	t.Run("reset_tracker", func(t *testing.T) {
		tracker := NewRateTracker(time.Minute, 60)

		// Record events
		tracker.Record()
		tracker.Record()
		assert.True(t, tracker.GetRate() > 0)

		// Reset
		tracker.Reset()
		assert.Equal(t, float64(0), tracker.GetRate())
	})

	t.Run("bucket_rotation", func(t *testing.T) {
		// Use a very short window for testing
		tracker := NewRateTracker(100*time.Millisecond, 10)

		// Record event
		tracker.Record()

		// Wait for bucket rotation
		time.Sleep(15 * time.Millisecond)

		// Record another event
		tracker.Record()

		rate := tracker.GetRate()
		assert.True(t, rate > 0)
	})
}

func TestRecentErrorsBuffer(t *testing.T) {
	t.Run("add_and_get", func(t *testing.T) {
		buffer := NewRecentErrorsBuffer(5)

		// Add errors
		for i := 0; i < 3; i++ {
			record := &ErrorRecord{
				Error:     Newf(CategoryConfig, "error %d", i),
				Timestamp: time.Now(),
			}
			buffer.Add(record)
		}

		recent := buffer.GetRecent(10)
		assert.Len(t, recent, 3)

		// Most recent should be first
		assert.Contains(t, recent[0].Error.Message, "error 2")
		assert.Contains(t, recent[1].Error.Message, "error 1")
		assert.Contains(t, recent[2].Error.Message, "error 0")
	})

	t.Run("buffer_overflow", func(t *testing.T) {
		buffer := NewRecentErrorsBuffer(3)

		// Add more than buffer size
		for i := 0; i < 5; i++ {
			record := &ErrorRecord{
				Error:     Newf(CategoryConfig, "error %d", i),
				Timestamp: time.Now(),
			}
			buffer.Add(record)
		}

		recent := buffer.GetRecent(10)
		assert.Len(t, recent, 3)

		// Should only have the 3 most recent
		assert.Contains(t, recent[0].Error.Message, "error 4")
		assert.Contains(t, recent[1].Error.Message, "error 3")
		assert.Contains(t, recent[2].Error.Message, "error 2")
	})

	t.Run("get_limited", func(t *testing.T) {
		buffer := NewRecentErrorsBuffer(10)

		// Add 5 errors
		for i := 0; i < 5; i++ {
			record := &ErrorRecord{
				Error:     Newf(CategoryConfig, "error %d", i),
				Timestamp: time.Now(),
			}
			buffer.Add(record)
		}

		// Get only 3
		recent := buffer.GetRecent(3)
		assert.Len(t, recent, 3)
	})

	t.Run("clear_buffer", func(t *testing.T) {
		buffer := NewRecentErrorsBuffer(5)

		// Add errors
		for i := 0; i < 3; i++ {
			record := &ErrorRecord{
				Error:     New(CategoryConfig, "error"),
				Timestamp: time.Now(),
			}
			buffer.Add(record)
		}

		assert.Len(t, buffer.GetRecent(10), 3)

		// Clear
		buffer.Clear()
		assert.Len(t, buffer.GetRecent(10), 0)
	})

	t.Run("empty_buffer", func(t *testing.T) {
		buffer := NewRecentErrorsBuffer(5)

		recent := buffer.GetRecent(10)
		assert.Len(t, recent, 0)
	})
}

func TestGlobalMetrics(t *testing.T) {
	t.Run("get_global_metrics", func(t *testing.T) {
		metrics1 := GetMetrics()
		metrics2 := GetMetrics()

		assert.Same(t, metrics1, metrics2, "Should return same instance")
	})
}

func TestMetricsThreadSafety(t *testing.T) {
	t.Run("concurrent_recording", func(t *testing.T) {
		metrics := NewErrorMetrics()

		var wg sync.WaitGroup
		workers := 10
		errorsPerWorker := 100

		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < errorsPerWorker; j++ {
					category := []ErrorCategory{
						CategoryConfig, CategoryScript, CategoryEngine,
					}[j%3]

					err := Newf(category, "error from worker %d", workerID)
					metrics.RecordError(err)
				}
			}(i)
		}

		wg.Wait()

		expectedTotal := int64(workers * errorsPerWorker)
		assert.Equal(t, expectedTotal, metrics.GetTotalErrors())

		// Check category distribution
		configCount := metrics.GetCategoryCount(CategoryConfig)
		scriptCount := metrics.GetCategoryCount(CategoryScript)
		engineCount := metrics.GetCategoryCount(CategoryEngine)

		assert.Equal(t, expectedTotal, configCount+scriptCount+engineCount)
	})

	t.Run("concurrent_stats", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Record some initial errors
		for i := 0; i < 10; i++ {
			metrics.RecordError(New(CategoryConfig, "error"))
		}

		var wg sync.WaitGroup
		wg.Add(3)

		// Concurrent stats reading
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				stats := metrics.GetStats()
				assert.True(t, stats.TotalErrors >= 10)
			}
		}()

		// Concurrent error recording
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				metrics.RecordError(New(CategoryScript, "error"))
			}
		}()

		// Concurrent recent errors reading
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				recent := metrics.GetRecentErrors(5)
				assert.NotNil(t, recent)
			}
		}()

		wg.Wait()
	})
}

func TestErrorRecord(t *testing.T) {
	t.Run("error_record_context", func(t *testing.T) {
		err := New(CategoryConfig, "test error").
			WithContext("file", "config.yaml").
			WithContext("line", 42)

		record := &ErrorRecord{
			Error:     err,
			Timestamp: time.Now(),
			Context:   err.Context,
		}

		assert.Equal(t, "config.yaml", record.Context["file"])
		assert.Equal(t, 42, record.Context["line"])
		assert.NotNil(t, record.Timestamp)
	})
}

func TestMetricsFormatter(t *testing.T) {
	t.Run("default_options", func(t *testing.T) {
		opts := DefaultMetricsFormatterOptions()

		assert.True(t, opts.ShowCategories)
		assert.True(t, opts.ShowRates)
		assert.True(t, opts.ShowRecentErrors)
		assert.Equal(t, 10, opts.RecentErrorLimit)
		assert.True(t, opts.ColorOutput)
	})

	t.Run("format_metrics_placeholder", func(t *testing.T) {
		metrics := NewErrorMetrics()
		opts := DefaultMetricsFormatterOptions()

		result := FormatMetrics(metrics, opts)
		assert.Contains(t, result, "not yet implemented")
	})
}

// Benchmarks
func BenchmarkMetricsRecordError(b *testing.B) {
	metrics := NewErrorMetrics()
	err := New(CategoryConfig, "benchmark error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metrics.RecordError(err)
	}
}

func BenchmarkMetricsGetStats(b *testing.B) {
	metrics := NewErrorMetrics()

	// Pre-populate with errors
	for i := 0; i < 1000; i++ {
		metrics.RecordError(New(ErrorCategory(fmt.Sprintf("cat%d", i%10)), "error"))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = metrics.GetStats()
	}
}

func BenchmarkRateTracker(b *testing.B) {
	tracker := NewRateTracker(time.Minute, 60)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tracker.Record()
		if i%100 == 0 {
			_ = tracker.GetRate()
		}
	}
}

func BenchmarkRecentErrorsBuffer(b *testing.B) {
	buffer := NewRecentErrorsBuffer(100)
	err := New(CategoryConfig, "benchmark error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		record := &ErrorRecord{
			Error:     err,
			Timestamp: time.Now(),
		}
		buffer.Add(record)

		if i%100 == 0 {
			_ = buffer.GetRecent(10)
		}
	}
}

// Test error metrics integration with formatter
func TestErrorMetricsIntegration(t *testing.T) {
	t.Run("record_and_format", func(t *testing.T) {
		metrics := NewErrorMetrics()

		// Record various errors
		err1 := ConfigError("invalid syntax").
			WithContext("file", "config.yaml").
			WithContext("line", 10)

		err2 := ScriptError("undefined variable").
			WithContext("script", "test.lua").
			WithContext("line", 42)

		err3 := NetworkError("connection timeout").
			WithContext("host", "api.example.com").
			WithContext("port", 443)

		metrics.RecordError(err1)
		metrics.RecordError(err2)
		metrics.RecordError(err3)

		// Get stats
		stats := metrics.GetStats()
		assert.Equal(t, int64(3), stats.TotalErrors)

		// Get recent errors
		recent := metrics.GetRecentErrors(5)
		assert.Len(t, recent, 3)

		// Most recent first
		assert.Equal(t, CategoryNetwork, recent[0].Error.Category)
		assert.Equal(t, CategoryScript, recent[1].Error.Category)
		assert.Equal(t, CategoryConfig, recent[2].Error.Category)

		// Check context preservation
		assert.Equal(t, "api.example.com", recent[0].Context["host"])
		assert.Equal(t, "test.lua", recent[1].Context["script"])
		assert.Equal(t, "config.yaml", recent[2].Context["file"])
	})
}

// Example of using metrics in production
func ExampleErrorMetrics() {
	metrics := GetMetrics()

	// Record an error
	err := ConfigError("invalid configuration").
		WithContext("file", "app.yaml").
		WithContext("section", "database")

	metrics.RecordError(err)

	// Get statistics
	stats := metrics.GetStats()
	fmt.Printf("Total errors: %d\n", stats.TotalErrors)

	// Get recent errors for debugging
	recent := metrics.GetRecentErrors(5)
	for _, record := range recent {
		fmt.Printf("Error: %s (Category: %s)\n",
			record.Error.Message,
			record.Error.Category)
	}
}
