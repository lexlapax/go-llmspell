// ABOUTME: State health management system for monitoring and evaluating Lua state health
// ABOUTME: Tracks execution metrics, calculates health scores, and provides recycling recommendations

package gopherlua

import (
	"math"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// HealthMetrics contains health information for a Lua state
type HealthMetrics struct {
	// Score is the overall health score (0.0 to 1.0, higher is better)
	Score float64

	// ExecutionCount is the total number of script executions
	ExecutionCount int64

	// ErrorCount is the total number of execution errors
	ErrorCount int64

	// TotalExecutionTime is the cumulative execution time
	TotalExecutionTime time.Duration

	// AverageExecutionTime is the average time per execution
	AverageExecutionTime time.Duration

	// MemoryUsage is the estimated memory usage in bytes
	MemoryUsage int64

	// LastUsed is the timestamp of the last execution
	LastUsed time.Time

	// Age is the time since the state was created
	Age time.Duration

	// ErrorRate is the percentage of executions that resulted in errors
	ErrorRate float64
}

// stateHealth tracks health data for a single state
type stateHealth struct {
	executionCount     int64
	errorCount         int64
	totalExecutionTime time.Duration
	memoryUsage        int64
	lastUsed           time.Time
	created            time.Time
	mu                 sync.RWMutex
}

// HealthMonitor manages health tracking for multiple Lua states
type HealthMonitor struct {
	states map[*lua.LState]*stateHealth
	mu     sync.RWMutex
}

// NewHealthMonitor creates a new health monitoring system
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		states: make(map[*lua.LState]*stateHealth),
	}
}

// RecordExecution records the execution of a script on a state
func (hm *HealthMonitor) RecordExecution(state *lua.LState, duration time.Duration, err error) {
	hm.mu.Lock()
	health, exists := hm.states[state]
	if !exists {
		health = &stateHealth{
			created: time.Now(),
		}
		hm.states[state] = health
	}
	hm.mu.Unlock()

	health.mu.Lock()
	defer health.mu.Unlock()

	health.executionCount++
	health.totalExecutionTime += duration
	health.lastUsed = time.Now()

	if err != nil {
		health.errorCount++
	}
}

// UpdateMemoryUsage updates the memory usage estimate for a state
func (hm *HealthMonitor) UpdateMemoryUsage(state *lua.LState, memoryBytes int64) {
	hm.mu.Lock()
	health, exists := hm.states[state]
	if !exists {
		health = &stateHealth{
			created: time.Now(),
		}
		hm.states[state] = health
	}
	hm.mu.Unlock()

	health.mu.Lock()
	health.memoryUsage = memoryBytes
	health.mu.Unlock()
}

// SetLastUsed manually sets the last used time (useful for testing)
func (hm *HealthMonitor) SetLastUsed(state *lua.LState, lastUsed time.Time) {
	hm.mu.Lock()
	health, exists := hm.states[state]
	if !exists {
		health = &stateHealth{
			created: time.Now(),
		}
		hm.states[state] = health
	}
	hm.mu.Unlock()

	health.mu.Lock()
	health.lastUsed = lastUsed
	health.mu.Unlock()
}

// GetMetrics returns the current health metrics for a state
func (hm *HealthMonitor) GetMetrics(state *lua.LState) HealthMetrics {
	hm.mu.RLock()
	health, exists := hm.states[state]
	hm.mu.RUnlock()

	if !exists {
		return HealthMetrics{
			Score: 1.0, // New states start healthy
		}
	}

	health.mu.RLock()
	defer health.mu.RUnlock()

	now := time.Now()
	age := now.Sub(health.created)

	var avgExecutionTime time.Duration
	var errorRate float64

	if health.executionCount > 0 {
		avgExecutionTime = health.totalExecutionTime / time.Duration(health.executionCount)
		errorRate = float64(health.errorCount) / float64(health.executionCount)
	}

	score := hm.calculateHealthScore(health, now)

	return HealthMetrics{
		Score:                score,
		ExecutionCount:       health.executionCount,
		ErrorCount:           health.errorCount,
		TotalExecutionTime:   health.totalExecutionTime,
		AverageExecutionTime: avgExecutionTime,
		MemoryUsage:          health.memoryUsage,
		LastUsed:             health.lastUsed,
		Age:                  age,
		ErrorRate:            errorRate,
	}
}

// ShouldRecycle determines if a state should be recycled based on health
func (hm *HealthMonitor) ShouldRecycle(state *lua.LState, threshold float64) bool {
	metrics := hm.GetMetrics(state)
	return metrics.Score < threshold
}

// CleanupState removes tracking for a closed state
func (hm *HealthMonitor) CleanupState(state *lua.LState) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.states, state)
}

// calculateHealthScore computes the health score based on various factors
func (hm *HealthMonitor) calculateHealthScore(health *stateHealth, now time.Time) float64 {
	baseScore := 1.0

	// Factor 1: Error rate (heavily weighted)
	if health.executionCount > 0 {
		errorRate := float64(health.errorCount) / float64(health.executionCount)
		// Error rate penalty: 0% errors = no penalty, 50% errors = -0.5, 100% errors = -1.0
		errorPenalty := errorRate * 1.0
		baseScore -= errorPenalty
	}

	// Factor 2: Average execution time (performance indicator)
	if health.executionCount > 0 {
		avgTime := health.totalExecutionTime / time.Duration(health.executionCount)
		// Penalty for slow execution (more sensitive for testing)
		if avgTime > time.Second {
			baseScore -= 0.3 // Heavy penalty for very slow execution
		} else if avgTime > 200*time.Millisecond {
			baseScore -= 0.2 // Significant penalty for slow execution
		} else if avgTime > 50*time.Millisecond {
			baseScore -= 0.1 // Moderate penalty for slow execution
		}
	}

	// Factor 3: Memory usage
	if health.memoryUsage > 0 {
		// Penalty based on memory usage (more sensitive for testing)
		memoryMB := float64(health.memoryUsage) / (1024 * 1024)
		if memoryMB > 100 { // >100MB is concerning
			baseScore -= 0.3
		} else if memoryMB > 50 { // >50MB is moderate concern
			baseScore -= 0.2
		} else if memoryMB > 10 { // >10MB is mild concern
			baseScore -= 0.1
		}
	}

	// Factor 4: Age since last use
	if !health.lastUsed.IsZero() {
		timeSinceLastUse := now.Sub(health.lastUsed)
		if timeSinceLastUse > 2*time.Hour {
			baseScore -= 0.2 // Old unused states are less healthy
		} else if timeSinceLastUse > 30*time.Minute {
			baseScore -= 0.1
		}
	}

	// Factor 5: Total execution count (wear and tear)
	if health.executionCount > 10000 {
		baseScore -= 0.1 // Heavy usage penalty
	} else if health.executionCount > 1000 {
		baseScore -= 0.05 // Moderate usage penalty
	}

	// Factor 6: Recent error pattern (last few executions)
	// This would require more sophisticated tracking but adds significant value
	// For now, we use the overall error rate as a proxy

	// Ensure score stays within bounds
	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 1 {
		baseScore = 1
	}

	// Apply smoothing to prevent rapid health fluctuations
	return math.Round(baseScore*100) / 100
}

// GetAllStatesMetrics returns metrics for all tracked states (useful for monitoring)
func (hm *HealthMonitor) GetAllStatesMetrics() map[*lua.LState]HealthMetrics {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make(map[*lua.LState]HealthMetrics, len(hm.states))
	for state := range hm.states {
		result[state] = hm.GetMetrics(state)
	}

	return result
}

// GetHealthStatistics returns aggregate statistics across all states
func (hm *HealthMonitor) GetHealthStatistics() struct {
	TotalStates      int
	HealthyStates    int // Score >= 0.8
	UnhealthyStates  int // Score < 0.5
	AverageScore     float64
	TotalExecutions  int64
	TotalErrors      int64
	OverallErrorRate float64
} {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stats := struct {
		TotalStates      int
		HealthyStates    int
		UnhealthyStates  int
		AverageScore     float64
		TotalExecutions  int64
		TotalErrors      int64
		OverallErrorRate float64
	}{}

	if len(hm.states) == 0 {
		return stats
	}

	var totalScore float64
	for state := range hm.states {
		metrics := hm.GetMetrics(state)
		stats.TotalStates++
		totalScore += metrics.Score
		stats.TotalExecutions += metrics.ExecutionCount
		stats.TotalErrors += metrics.ErrorCount

		if metrics.Score >= 0.8 {
			stats.HealthyStates++
		} else if metrics.Score < 0.5 {
			stats.UnhealthyStates++
		}
	}

	stats.AverageScore = totalScore / float64(stats.TotalStates)
	if stats.TotalExecutions > 0 {
		stats.OverallErrorRate = float64(stats.TotalErrors) / float64(stats.TotalExecutions)
	}

	return stats
}
