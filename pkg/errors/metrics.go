// ABOUTME: This file implements error metrics and tracking for monitoring and debugging purposes.
// ABOUTME: It provides counters, histograms, and error reporting for operational insights.

package errors

import (
	"sync"
	"sync/atomic"
	"time"
)

// ErrorMetrics tracks error statistics.
// It provides comprehensive error monitoring including counts by category,
// error rates over time, and recent error history for debugging.
type ErrorMetrics struct {
	// Counters by category
	categoryCounters map[ErrorCategory]*int64

	// Error rate tracking
	errorRates map[ErrorCategory]*RateTracker

	// Recent errors for debugging
	recentErrors *RecentErrorsBuffer

	// Total error count
	totalErrors int64

	// Start time for uptime calculation
	startTime time.Time

	mu sync.RWMutex
}

// RateTracker tracks error rates over time.
// It uses a circular buffer of time buckets to calculate
// rolling error rates without storing individual timestamps.
type RateTracker struct {
	window      time.Duration
	buckets     []int64
	bucketCount int
	currentIdx  int
	lastUpdate  time.Time
	mu          sync.Mutex
}

// RecentErrorsBuffer stores recent errors for debugging.
// It maintains a circular buffer of error records with their
// full context for post-mortem analysis.
type RecentErrorsBuffer struct {
	errors     []*ErrorRecord
	maxSize    int
	currentIdx int
	mu         sync.RWMutex
}

// ErrorRecord represents a recorded error.
// It captures the error details, timestamp, and any additional
// context at the time of the error for debugging.
type ErrorRecord struct {
	Error     *SpellError
	Timestamp time.Time
	Context   map[string]interface{}
}

// Global metrics instance.
// Uses sync.Once to ensure thread-safe lazy initialization.
var (
	globalMetrics *ErrorMetrics
	metricsOnce   sync.Once
)

// GetMetrics returns the global error metrics instance.
// It's initialized lazily on first access with default settings
// for all error categories.
func GetMetrics() *ErrorMetrics {
	metricsOnce.Do(func() {
		globalMetrics = NewErrorMetrics()
	})
	return globalMetrics
}

// NewErrorMetrics creates a new error metrics instance.
// It initializes counters and rate trackers for all known error
// categories with a 100-entry recent errors buffer.
func NewErrorMetrics() *ErrorMetrics {
	m := &ErrorMetrics{
		categoryCounters: make(map[ErrorCategory]*int64),
		errorRates:       make(map[ErrorCategory]*RateTracker),
		recentErrors:     NewRecentErrorsBuffer(100),
		startTime:        time.Now(),
	}

	// Initialize counters and rate trackers for all categories
	categories := []ErrorCategory{
		CategoryUnknown, CategoryUsage, CategoryConfig, CategoryScript,
		CategoryEngine, CategorySecurity, CategoryNetwork, CategoryTimeout,
		CategoryResource, CategoryValidation, CategoryDependency, CategoryIO,
		CategoryInterrupted,
	}

	for _, cat := range categories {
		var counter int64
		m.categoryCounters[cat] = &counter
		m.errorRates[cat] = NewRateTracker(time.Minute, 60)
	}

	return m
}

// RecordError records an error in metrics.
// It increments counters, updates rate trackers, and stores
// the error in recent history. Non-SpellErrors are wrapped.
func (m *ErrorMetrics) RecordError(err error) {
	if err == nil || m == nil {
		return
	}

	// Increment total counter
	atomic.AddInt64(&m.totalErrors, 1)

	// Check if it's a SpellError
	var spellErr *SpellError
	if IsSpellError(err) {
		spellErr, _ = err.(*SpellError)
	} else {
		// Create a SpellError wrapper for non-SpellErrors
		spellErr = Wrap(err, CategoryUnknown, err.Error())
	}

	// Ensure we have a valid SpellError
	if spellErr == nil {
		return
	}

	// Record by category
	m.mu.RLock()
	counter, exists := m.categoryCounters[spellErr.Category]
	rateTracker := m.errorRates[spellErr.Category]
	m.mu.RUnlock()

	if exists && counter != nil {
		atomic.AddInt64(counter, 1)
	}

	// Record in rate tracker
	if rateTracker != nil {
		rateTracker.Record()
	}

	// Record in recent errors
	record := &ErrorRecord{
		Error:     spellErr,
		Timestamp: time.Now(),
		Context:   spellErr.Context,
	}
	m.recentErrors.Add(record)
}

// GetTotalErrors returns the total number of errors recorded.
// This is the sum of all category counters since initialization
// or last reset.
func (m *ErrorMetrics) GetTotalErrors() int64 {
	return atomic.LoadInt64(&m.totalErrors)
}

// GetCategoryCount returns the error count for a specific category.
// Returns 0 for unknown categories. The count is since initialization
// or last reset.
func (m *ErrorMetrics) GetCategoryCount(category ErrorCategory) int64 {
	m.mu.RLock()
	counter, exists := m.categoryCounters[category]
	m.mu.RUnlock()

	if !exists {
		return 0
	}

	return atomic.LoadInt64(counter)
}

// GetErrorRate returns the error rate for a specific category.
// The rate is calculated as errors per minute over the configured
// time window (default 1 minute).
func (m *ErrorMetrics) GetErrorRate(category ErrorCategory) float64 {
	m.mu.RLock()
	rateTracker, exists := m.errorRates[category]
	m.mu.RUnlock()

	if !exists {
		return 0
	}

	return rateTracker.GetRate()
}

// GetRecentErrors returns recent error records.
// Returns up to 'limit' most recent errors in reverse chronological
// order for debugging and analysis.
func (m *ErrorMetrics) GetRecentErrors(limit int) []*ErrorRecord {
	return m.recentErrors.GetRecent(limit)
}

// GetStats returns error statistics.
// It provides a comprehensive snapshot including total errors,
// uptime, and per-category statistics with percentages and rates.
func (m *ErrorMetrics) GetStats() *ErrorStats {
	stats := &ErrorStats{
		TotalErrors: m.GetTotalErrors(),
		Uptime:      time.Since(m.startTime),
		Categories:  make(map[ErrorCategory]*CategoryStats),
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for category, counter := range m.categoryCounters {
		count := atomic.LoadInt64(counter)
		rate := 0.0
		if rateTracker, exists := m.errorRates[category]; exists {
			rate = rateTracker.GetRate()
		}

		stats.Categories[category] = &CategoryStats{
			Count:      count,
			Percentage: float64(count) / float64(stats.TotalErrors) * 100,
			Rate:       rate,
		}
	}

	return stats
}

// Reset resets all metrics.
// This clears all counters, rate trackers, and recent errors.
// Useful for testing or periodic metric resets.
func (m *ErrorMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset counters
	for _, counter := range m.categoryCounters {
		atomic.StoreInt64(counter, 0)
	}

	// Reset rate trackers
	for _, tracker := range m.errorRates {
		tracker.Reset()
	}

	// Reset total
	atomic.StoreInt64(&m.totalErrors, 0)

	// Clear recent errors
	m.recentErrors.Clear()

	// Update start time
	m.startTime = time.Now()
}

// ErrorStats represents error statistics.
// It provides a point-in-time snapshot of all error metrics
// for reporting and monitoring.
type ErrorStats struct {
	TotalErrors int64
	Uptime      time.Duration
	Categories  map[ErrorCategory]*CategoryStats
}

// CategoryStats represents statistics for a specific error category.
// It includes the absolute count, percentage of total errors,
// and current error rate per minute.
type CategoryStats struct {
	Count      int64
	Percentage float64
	Rate       float64 // errors per minute
}

// NewRateTracker creates a new rate tracker.
// The window parameter defines the time period for rate calculation,
// and buckets determines the granularity of tracking.
func NewRateTracker(window time.Duration, buckets int) *RateTracker {
	return &RateTracker{
		window:      window,
		buckets:     make([]int64, buckets),
		bucketCount: buckets,
		lastUpdate:  time.Now(),
	}
}

// Record records an event.
// It updates the current time bucket and handles bucket
// rotation based on elapsed time.
func (rt *RateTracker) Record() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.updateBuckets()
	atomic.AddInt64(&rt.buckets[rt.currentIdx], 1)
}

// GetRate returns the current rate (events per minute).
// It sums all buckets within the time window and calculates
// the rate normalized to events per minute.
func (rt *RateTracker) GetRate() float64 {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.updateBuckets()

	// Sum all buckets
	var total int64
	for i := 0; i < rt.bucketCount; i++ {
		total += atomic.LoadInt64(&rt.buckets[i])
	}

	// Calculate rate per minute
	windowMinutes := rt.window.Minutes()
	if windowMinutes > 0 {
		return float64(total) / windowMinutes
	}

	return 0
}

// Reset resets the rate tracker.
// This clears all buckets and resets the time tracking,
// effectively starting rate calculation from zero.
func (rt *RateTracker) Reset() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	for i := 0; i < rt.bucketCount; i++ {
		atomic.StoreInt64(&rt.buckets[i], 0)
	}
	rt.currentIdx = 0
	rt.lastUpdate = time.Now()
}

// updateBuckets updates bucket indices based on elapsed time.
// It advances the current bucket index and clears expired buckets
// to maintain accurate rolling window calculations.
func (rt *RateTracker) updateBuckets() {
	elapsed := time.Since(rt.lastUpdate)
	bucketDuration := rt.window / time.Duration(rt.bucketCount)
	bucketsToAdvance := int(elapsed / bucketDuration)

	if bucketsToAdvance > 0 {
		// Clear old buckets
		for i := 0; i < bucketsToAdvance && i < rt.bucketCount; i++ {
			nextIdx := (rt.currentIdx + i + 1) % rt.bucketCount
			atomic.StoreInt64(&rt.buckets[nextIdx], 0)
		}

		// Advance current index
		rt.currentIdx = (rt.currentIdx + bucketsToAdvance) % rt.bucketCount
		rt.lastUpdate = time.Now()
	}
}

// NewRecentErrorsBuffer creates a new recent errors buffer.
// The maxSize parameter determines how many recent errors to keep
// in the circular buffer for debugging purposes.
func NewRecentErrorsBuffer(maxSize int) *RecentErrorsBuffer {
	return &RecentErrorsBuffer{
		errors:  make([]*ErrorRecord, maxSize),
		maxSize: maxSize,
	}
}

// Add adds an error record to the buffer.
// It overwrites the oldest entry when the buffer is full,
// maintaining a fixed-size history of recent errors.
func (reb *RecentErrorsBuffer) Add(record *ErrorRecord) {
	reb.mu.Lock()
	defer reb.mu.Unlock()

	reb.errors[reb.currentIdx] = record
	reb.currentIdx = (reb.currentIdx + 1) % reb.maxSize
}

// GetRecent returns the most recent error records.
// Records are returned in reverse chronological order (newest first)
// up to the specified limit.
func (reb *RecentErrorsBuffer) GetRecent(limit int) []*ErrorRecord {
	reb.mu.RLock()
	defer reb.mu.RUnlock()

	if limit > reb.maxSize {
		limit = reb.maxSize
	}

	result := make([]*ErrorRecord, 0, limit)

	// Start from the most recent and work backwards
	for i := 0; i < limit; i++ {
		idx := (reb.currentIdx - 1 - i + reb.maxSize) % reb.maxSize
		if reb.errors[idx] != nil {
			result = append(result, reb.errors[idx])
		}
	}

	return result
}

// Clear clears the buffer.
// This removes all stored error records and resets the buffer
// to its initial empty state.
func (reb *RecentErrorsBuffer) Clear() {
	reb.mu.Lock()
	defer reb.mu.Unlock()

	for i := range reb.errors {
		reb.errors[i] = nil
	}
	reb.currentIdx = 0
}

// MetricsFormatterOptions configures metrics formatting.
// It controls which metrics are displayed and how they're formatted
// for different output scenarios (terminal, logs, monitoring).
type MetricsFormatterOptions struct {
	ShowCategories   bool
	ShowRates        bool
	ShowRecentErrors bool
	RecentErrorLimit int
	ColorOutput      bool
}

// DefaultMetricsFormatterOptions returns default options.
// The defaults show all available metrics with color output
// and the 10 most recent errors.
func DefaultMetricsFormatterOptions() MetricsFormatterOptions {
	return MetricsFormatterOptions{
		ShowCategories:   true,
		ShowRates:        true,
		ShowRecentErrors: true,
		RecentErrorLimit: 10,
		ColorOutput:      true,
	}
}

// FormatMetrics formats error metrics for display.
// It creates a human-readable representation of error statistics
// based on the provided formatting options.
func FormatMetrics(metrics *ErrorMetrics, options MetricsFormatterOptions) string {
	// Implementation would format metrics for display
	// This is a placeholder for brevity
	stats := metrics.GetStats()
	return formatStats(stats, options)
}

// formatStats formats error statistics.
// This is a placeholder for the actual implementation that would
// create formatted tables, apply colors, and structure the output.
func formatStats(stats *ErrorStats, options MetricsFormatterOptions) string {
	// Placeholder implementation
	// Would format stats with colors, tables, etc.
	return "Error metrics formatting not yet implemented"
}
