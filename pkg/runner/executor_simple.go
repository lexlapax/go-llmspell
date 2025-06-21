// ABOUTME: This file provides a simplified executor implementation for the runner package.
// ABOUTME: It focuses on the core execution logic without direct engine dependencies.

package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SimpleExecutor provides basic script execution functionality
type SimpleExecutor struct {
	config           *RunnerConfig
	loader           *SpellLoader
	semaphore        chan struct{}
	wg               sync.WaitGroup
	mu               sync.RWMutex
	metrics          *RunnerMetrics
	metricsLock      sync.RWMutex
	startTime        time.Time
	shutdownComplete bool
}

// NewSimpleExecutor creates a new simple executor
func NewSimpleExecutor(config *RunnerConfig) *SimpleExecutor {
	return &SimpleExecutor{
		config:    config,
		loader:    NewSpellLoader(),
		semaphore: make(chan struct{}, config.MaxConcurrentScripts),
		metrics: &RunnerMetrics{
			EngineMetrics: make(map[string]*EngineMetric),
		},
		startTime: time.Now(),
	}
}

// Initialize prepares the executor
func (e *SimpleExecutor) Initialize(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.shutdownComplete {
		return fmt.Errorf("executor has been shut down")
	}

	return nil
}

// Execute runs a script
func (e *SimpleExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	options := &RunnerOptions{
		Parameters: params,
		Engine:     e.config.DefaultEngine,
	}

	result, err := e.ExecuteWithOptions(ctx, script, options)
	if err != nil {
		return nil, err
	}

	return result.Value, result.Error
}

// ExecuteFile runs a script file
func (e *SimpleExecutor) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	// For now, just return a placeholder
	e.updateMetrics(e.config.DefaultEngine, 10*time.Millisecond, nil)
	return fmt.Sprintf("Executed file: %s", filepath), nil
}

// ExecuteWithOptions executes with custom options
func (e *SimpleExecutor) ExecuteWithOptions(ctx context.Context, script string, options *RunnerOptions) (*ExecutionResult, error) {
	// Check if shutdown in progress
	e.mu.RLock()
	if e.shutdownComplete {
		e.mu.RUnlock()
		return nil, fmt.Errorf("executor is shutting down")
	}
	e.mu.RUnlock()

	// Acquire semaphore for concurrency control
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Track execution
	e.wg.Add(1)
	defer e.wg.Done()

	startTime := time.Now()
	result := &ExecutionResult{
		StartTime: startTime,
		Metadata:  make(map[string]interface{}),
		Engine:    options.Engine,
	}

	if result.Engine == "" {
		result.Engine = e.config.DefaultEngine
	}

	// Report progress
	if options.ProgressHandler != nil {
		options.ProgressHandler(Progress{
			Stage:       "execution",
			Message:     fmt.Sprintf("Executing script with %s engine", result.Engine),
			Percentage:  50,
			CurrentStep: 2,
			TotalSteps:  4,
			StartTime:   startTime,
		})
	}

	// Simulate execution
	time.Sleep(10 * time.Millisecond)

	// Complete result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Value = fmt.Sprintf("Executed: %s", script)

	// Update metrics
	e.updateMetrics(result.Engine, result.Duration, nil)

	// Final progress
	if options.ProgressHandler != nil {
		options.ProgressHandler(Progress{
			Stage:       "complete",
			Message:     "Done",
			Percentage:  100,
			CurrentStep: 4,
			TotalSteps:  4,
			StartTime:   startTime,
		})
	}

	return result, nil
}

// Validate checks if a script is valid
func (e *SimpleExecutor) Validate(script string) error {
	// Placeholder validation
	if script == "" {
		return fmt.Errorf("empty script")
	}
	return nil
}

// Shutdown cleanly shuts down the executor
func (e *SimpleExecutor) Shutdown() error {
	e.mu.Lock()
	if e.shutdownComplete {
		e.mu.Unlock()
		return nil
	}
	e.shutdownComplete = true
	e.mu.Unlock()

	// Wait for all executions to complete
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		// All executions completed
	case <-time.After(30 * time.Second):
		return fmt.Errorf("shutdown timeout: some executions did not complete")
	}

	return nil
}

// GetMetrics returns execution metrics
func (e *SimpleExecutor) GetMetrics() *RunnerMetrics {
	e.metricsLock.RLock()
	defer e.metricsLock.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &RunnerMetrics{
		ScriptsExecuted:   e.metrics.ScriptsExecuted,
		TotalDuration:     e.metrics.TotalDuration,
		AverageDuration:   e.metrics.AverageDuration,
		SuccessCount:      e.metrics.SuccessCount,
		ErrorCount:        e.metrics.ErrorCount,
		LastExecutionTime: e.metrics.LastExecutionTime,
		EngineMetrics:     make(map[string]*EngineMetric),
	}

	// Copy engine metrics
	for name, em := range e.metrics.EngineMetrics {
		metrics.EngineMetrics[name] = &EngineMetric{
			ExecutionCount: em.ExecutionCount,
			TotalDuration:  em.TotalDuration,
			ErrorCount:     em.ErrorCount,
		}
	}

	return metrics
}

// updateMetrics updates execution metrics
func (e *SimpleExecutor) updateMetrics(engineName string, duration time.Duration, err error) {
	e.metricsLock.Lock()
	defer e.metricsLock.Unlock()

	e.metrics.ScriptsExecuted++
	e.metrics.TotalDuration += duration
	e.metrics.LastExecutionTime = time.Now()

	if err == nil {
		e.metrics.SuccessCount++
	} else {
		e.metrics.ErrorCount++
	}

	// Update average duration
	if e.metrics.ScriptsExecuted > 0 {
		e.metrics.AverageDuration = e.metrics.TotalDuration / time.Duration(e.metrics.ScriptsExecuted)
	}

	// Update engine-specific metrics
	if engineName != "" {
		if _, ok := e.metrics.EngineMetrics[engineName]; !ok {
			e.metrics.EngineMetrics[engineName] = &EngineMetric{}
		}

		em := e.metrics.EngineMetrics[engineName]
		em.ExecutionCount++
		em.TotalDuration += duration
		if err != nil {
			em.ErrorCount++
		}
	}
}
