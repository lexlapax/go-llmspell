// ABOUTME: This file implements the script execution logic with proper lifecycle management.
// ABOUTME: It handles parameter passing, progress tracking, and concurrent execution control.

package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// mapProfileToLevelAndFeatureSet maps old security profiles to new security levels and feature sets.
// This provides backward compatibility for existing configurations.
func mapProfileToLevelAndFeatureSet(profile string) (securityLevel, featureSet string) {
	switch profile {
	case "sandbox":
		return "untrusted", "full"
	case "development":
		return "trusted", "full"
	case "production":
		return "trusted", "full"
	case "minimal":
		return "trusted", "minimal"
	case "llm":
		return "trusted", "llm"
	default:
		// Default to trusted/full for unknown profiles
		return "trusted", "full"
	}
}

// mapSecurityLevelToEngineLevel maps new security levels to engine-expected security levels.
// This ensures engines receive the security level names they understand.
func mapSecurityLevelToEngineLevel(securityLevel string) string {
	switch securityLevel {
	case "untrusted":
		return "strict"
	case "trusted":
		return "standard"
	case "privileged":
		return "standard" // Privileged uses standard for now, could be enhanced later
	default:
		// Default to standard for unknown levels
		return "standard"
	}
}

// mapProfileDirectToEngineLevel maps old security profiles directly to engine-expected security levels.
// This provides backward compatibility and handles special cases like "minimal".
func mapProfileDirectToEngineLevel(profile string) string {
	switch profile {
	case "sandbox":
		return "strict"
	case "development":
		return "standard"
	case "production":
		return "standard"
	case "minimal":
		return "minimal" // Special case: minimal maps directly to minimal
	case "":
		return "standard" // Default for empty profile
	default:
		// Default to standard for unknown profiles
		return "standard"
	}
}

// ScriptExecutor implements the Runner interface for executing scripts.
// It manages engine lifecycle, concurrent execution limits, progress tracking,
// and metrics collection across multiple script engines.
type ScriptExecutor struct {
	config        *RunnerConfig
	engineManager *EngineRegistryManager
	selector      *EngineSelector
	loader        *SpellLoader

	// Execution control
	semaphore chan struct{} // Limits concurrent executions
	wg        sync.WaitGroup
	mu        sync.RWMutex

	// Metrics tracking
	metrics          *RunnerMetrics
	metricsLock      sync.RWMutex
	startTime        time.Time
	shutdownComplete bool
}

// NewScriptExecutor creates a new script executor.
// It initializes the executor with the provided configuration, engine manager,
// and selector, setting up concurrency controls and metrics tracking.
func NewScriptExecutor(config *RunnerConfig, engineManager *EngineRegistryManager, selector *EngineSelector) *ScriptExecutor {
	return &ScriptExecutor{
		config:        config,
		engineManager: engineManager,
		selector:      selector,
		loader:        NewSpellLoader(),
		semaphore:     make(chan struct{}, config.MaxConcurrentScripts),
		metrics: &RunnerMetrics{
			EngineMetrics: make(map[string]*EngineMetric),
		},
		startTime: time.Now(),
	}
}

// Initialize prepares the executor for running scripts.
// It ensures the executor hasn't been shut down and initializes
// the engine manager for script execution.
func (e *ScriptExecutor) Initialize(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.shutdownComplete {
		return fmt.Errorf("executor has been shut down")
	}

	// Initialize engine manager
	if err := e.engineManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize engine manager: %w", err)
	}

	return nil
}

// Execute runs a script with the given parameters.
// It uses default options with the provided parameters and delegates
// to ExecuteWithOptions for the actual execution.
func (e *ScriptExecutor) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	options := &RunnerOptions{
		Parameters:    params,
		Engine:        e.config.DefaultEngine,
		SecurityLevel: e.config.DefaultSecurityLevel,
		FeatureSet:    e.config.DefaultFeatureSet,
	}

	result, err := e.ExecuteWithOptions(ctx, script, options)
	if err != nil {
		return nil, err
	}

	return result.Value, result.Error
}

// ExecuteFile runs a script file with the given parameters.
// It automatically selects the appropriate engine based on the file extension
// and executes the file contents with the provided parameters.
func (e *ScriptExecutor) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	// Determine engine from file extension
	engineName, err := e.selector.SelectByExtension(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to select engine: %w", err)
	}

	options := &RunnerOptions{
		Parameters:    params,
		Engine:        engineName,
		SecurityLevel: e.config.DefaultSecurityLevel,
		FeatureSet:    e.config.DefaultFeatureSet,
	}

	// Build engine config
	var engineConfig map[string]interface{}
	if e.config.EngineConfigs != nil {
		engineConfig = e.config.EngineConfigs[engineName]
	}
	config := BuildEngineConfig(e.config, engineConfig)

	// Get engine with bridges loaded lazily
	// Use defaults for file execution
	securityLevel := e.config.DefaultSecurityLevel
	if securityLevel == "" {
		securityLevel = "trusted"
	}
	featureSet := e.config.DefaultFeatureSet
	if featureSet == "" {
		featureSet = "full"
	}

	scriptEngine, err := e.engineManager.GetEngine(engineName, config, security.SecurityLevel(securityLevel), registry.FeatureSet(featureSet))
	if err != nil {
		e.updateMetrics(engineName, 0, err)
		return nil, fmt.Errorf("failed to get engine %s: %w", engineName, err)
	}

	// Execute using the script engine directly
	startTime := time.Now()
	result, err := scriptEngine.ExecuteFile(ctx, filepath, options.Parameters)
	duration := time.Since(startTime)

	e.updateMetrics(engineName, duration, err)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteWithOptions executes a script with custom options.
// It handles concurrency control, progress reporting, engine selection,
// timeout enforcement, and metrics tracking for the execution.
func (e *ScriptExecutor) ExecuteWithOptions(ctx context.Context, script string, options *RunnerOptions) (*ExecutionResult, error) {
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
	}

	// Report progress
	if options.ProgressHandler != nil {
		options.ProgressHandler(Progress{
			Stage:       "initialization",
			Message:     "Preparing script execution",
			Percentage:  10,
			CurrentStep: 1,
			TotalSteps:  4,
			StartTime:   startTime,
		})
	}

	// Determine engine
	engineName := options.Engine
	if engineName == "" {
		engineName = e.config.DefaultEngine
	}

	// Build engine config
	var engineConfig map[string]interface{}
	if e.config.EngineConfigs != nil {
		engineConfig = e.config.EngineConfigs[engineName]
	}
	config := BuildEngineConfig(e.config, engineConfig)
	config = ApplyOptionsToConfig(config, options)

	// Determine security level and feature set
	securityLevel := options.SecurityLevel
	featureSet := options.FeatureSet
	usingProfile := false

	// Handle backward compatibility with SecurityProfile
	if securityLevel == "" && options.SecurityProfile != "" {
		// Map old profile to new system
		securityLevel, featureSet = mapProfileToLevelAndFeatureSet(options.SecurityProfile)
		usingProfile = true
	}

	// Use defaults if not specified
	if securityLevel == "" {
		securityLevel = e.config.DefaultSecurityLevel
		if securityLevel == "" {
			// Fallback for old configs
			securityLevel = "trusted"
		}
	}

	if featureSet == "" {
		featureSet = e.config.DefaultFeatureSet
		if featureSet == "" {
			// Fallback for old configs
			featureSet = "full"
		}
	}

	// Validate security level
	if !security.IsValidLevel(securityLevel) {
		result.Error = fmt.Errorf("invalid security level: %s", securityLevel)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		e.updateMetrics(engineName, result.Duration, result.Error)
		return result, result.Error
	}

	// Validate feature set
	if !registry.IsValidFeatureSet(featureSet) {
		result.Error = fmt.Errorf("invalid feature set: %s", featureSet)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		e.updateMetrics(engineName, result.Duration, result.Error)
		return result, result.Error
	}

	// Store security level in engine options for runtime enforcement
	var engineSecurityLevel string
	if usingProfile {
		// For backward compatibility, map profile directly to engine level
		engineSecurityLevel = mapProfileDirectToEngineLevel(options.SecurityProfile)
	} else {
		// Map new security level to engine-expected level format
		engineSecurityLevel = mapSecurityLevelToEngineLevel(securityLevel)
	}
	config.EngineOptions["security_level"] = engineSecurityLevel

	// Get engine with bridges loaded lazily
	engine, err := e.engineManager.GetEngine(engineName, config, security.SecurityLevel(securityLevel), registry.FeatureSet(featureSet))
	if err != nil {
		result.Error = fmt.Errorf("failed to get engine %s: %w", engineName, err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		e.updateMetrics(engineName, result.Duration, result.Error)
		return result, result.Error
	}

	result.Engine = engineName

	// Report progress
	if options.ProgressHandler != nil {
		options.ProgressHandler(Progress{
			Stage:       "execution",
			Message:     fmt.Sprintf("Executing script with %s engine", engineName),
			Percentage:  50,
			CurrentStep: 2,
			TotalSteps:  4,
			StartTime:   startTime,
		})
	}

	// Execute script
	execCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	value, err := engine.Execute(execCtx, script, options.Parameters)

	// Complete result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Value = value
	result.Error = err

	// Report progress
	if options.ProgressHandler != nil {
		options.ProgressHandler(Progress{
			Stage:       "cleanup",
			Message:     "Execution complete",
			Percentage:  90,
			CurrentStep: 3,
			TotalSteps:  4,
			StartTime:   startTime,
		})
	}

	// Update metrics
	e.updateMetrics(engineName, result.Duration, err)

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

// ExecuteSpell executes a spell with its metadata.
// It applies parameter defaults, validates inputs, selects the appropriate
// engine, and executes the spell's entry point with proper metadata tracking.
func (e *ScriptExecutor) ExecuteSpell(ctx context.Context, spell *SpellMetadata, params map[string]interface{}) (*ExecutionResult, error) {
	// Apply parameter defaults
	params = e.loader.ApplyDefaults(spell, params)

	// Validate parameters
	if err := e.loader.ValidateParameters(spell, params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Determine engine
	engineName, err := e.selector.SelectForSpell(spell)
	if err != nil {
		return nil, fmt.Errorf("failed to select engine: %w", err)
	}

	// Resolve entry point
	entryPoint := e.loader.ResolveEntryPoint(spell)

	// Create options
	options := &RunnerOptions{
		Parameters:      params,
		Engine:          engineName,
		SecurityProfile: spell.SecurityProfile,
	}

	// Execute the entry point file
	result, err := e.ExecuteWithOptions(ctx, entryPoint, options)
	if err != nil {
		return nil, err
	}

	// Add spell metadata to result
	result.Metadata["spell_name"] = spell.Name
	result.Metadata["spell_version"] = spell.Version

	return result, nil
}

// Validate checks if a script is valid without executing it.
// It performs syntax validation using engine-specific validators.
// Currently a placeholder for future implementation.
func (e *ScriptExecutor) Validate(script string) error {
	// This would integrate with engine-specific validators
	// For now, just a placeholder
	return nil
}

// Shutdown cleanly shuts down the executor.
// It waits for all running executions to complete (with timeout)
// and shuts down the engine manager, ensuring graceful termination.
func (e *ScriptExecutor) Shutdown() error {
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

	// Shutdown engine manager
	if err := e.engineManager.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown engine manager: %w", err)
	}

	return nil
}

// GetMetrics returns execution metrics.
// It creates a copy of the current metrics to avoid race conditions
// and merges data from the engine registry for comprehensive statistics.
func (e *ScriptExecutor) GetMetrics() *RunnerMetrics {
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

	// Get latest stats from engine registry
	registryStats := e.engineManager.GetStats()
	engineMetrics := CreateEngineMetrics(registryStats)

	// Merge with our metrics
	for name, em := range engineMetrics {
		if existing, ok := metrics.EngineMetrics[name]; ok {
			existing.ExecutionCount = em.ExecutionCount
			existing.TotalDuration = em.TotalDuration
			existing.ErrorCount = em.ErrorCount
		} else {
			metrics.EngineMetrics[name] = em
		}
	}

	return metrics
}

// updateMetrics updates execution metrics.
// It tracks success/error counts, execution duration, and maintains
// per-engine statistics for performance monitoring.
func (e *ScriptExecutor) updateMetrics(engineName string, duration time.Duration, err error) {
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

// GetEngineRegistry returns the engine registry manager.
// This is primarily used by components that need access to the engine registry,
// such as the REPL for loading bridges into its persistent Lua state.
func (e *ScriptExecutor) GetEngineRegistry() interface{} {
	return e.engineManager
}
