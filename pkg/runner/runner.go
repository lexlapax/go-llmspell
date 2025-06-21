// ABOUTME: This file defines the core Runner interface for executing scripts with different engines.
// ABOUTME: It provides lifecycle management, metrics tracking, and execution options.

package runner

import (
	"context"
	"fmt"
	"time"
)

// Runner defines the interface for script execution
type Runner interface {
	// Initialize prepares the runner for execution
	Initialize(ctx context.Context) error

	// Execute runs a script with the given parameters
	Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error)

	// ExecuteFile runs a script file with the given parameters
	ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error)

	// Validate checks if a script is valid without executing it
	Validate(script string) error

	// Shutdown cleanly shuts down the runner and releases resources
	Shutdown() error

	// GetMetrics returns execution metrics
	GetMetrics() *RunnerMetrics
}

// RunnerConfig configures the behavior of a runner
type RunnerConfig struct {
	// Execution settings
	Timeout              time.Duration     `json:"timeout" yaml:"timeout"`
	MaxConcurrentScripts int               `json:"max_concurrent_scripts" yaml:"max_concurrent_scripts"`
	WorkingDirectory     string            `json:"working_directory" yaml:"working_directory"`
	Environment          map[string]string `json:"environment" yaml:"environment"`

	// Feature toggles
	EnableMetrics      bool `json:"enable_metrics" yaml:"enable_metrics"`
	EnableValidation   bool `json:"enable_validation" yaml:"enable_validation"`
	EnableDebug        bool `json:"enable_debug" yaml:"enable_debug"`
	EnableProgressBars bool `json:"enable_progress_bars" yaml:"enable_progress_bars"`

	// Engine settings
	DefaultEngine string                            `json:"default_engine" yaml:"default_engine"`
	EngineConfigs map[string]map[string]interface{} `json:"engine_configs" yaml:"engine_configs"`

	// Security settings
	DefaultSecurityProfile string                 `json:"default_security_profile" yaml:"default_security_profile"`
	SecurityProfiles       map[string]interface{} `json:"security_profiles" yaml:"security_profiles"`
}

// Validate checks if the configuration is valid
func (c *RunnerConfig) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
	}
	if c.MaxConcurrentScripts <= 0 {
		return fmt.Errorf("max_concurrent_scripts must be positive, got %d", c.MaxConcurrentScripts)
	}
	return nil
}

// DefaultRunnerConfig returns a configuration with sensible defaults
func DefaultRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		Timeout:                30 * time.Second,
		MaxConcurrentScripts:   10,
		EnableMetrics:          true,
		EnableValidation:       true,
		EnableDebug:            false,
		EnableProgressBars:     true,
		DefaultEngine:          "lua",
		EngineConfigs:          make(map[string]map[string]interface{}),
		DefaultSecurityProfile: "sandbox",
		SecurityProfiles:       make(map[string]interface{}),
		Environment:            make(map[string]string),
	}
}

// RunnerOptions contains options for a single execution
type RunnerOptions struct {
	// Timeout overrides the default timeout for this execution
	Timeout time.Duration

	// Parameters to pass to the script
	Parameters map[string]interface{}

	// Engine overrides the default engine
	Engine string

	// SecurityProfile overrides the default security profile
	SecurityProfile string

	// ProgressHandler receives progress updates during execution
	ProgressHandler func(Progress)

	// Debug enables debug mode for this execution
	Debug bool
}

// RunnerOption is a function that configures RunnerOptions
type RunnerOption func(*RunnerOptions)

// WithTimeout sets a custom timeout for the execution
func WithTimeout(timeout time.Duration) RunnerOption {
	return func(opts *RunnerOptions) {
		opts.Timeout = timeout
	}
}

// WithParameters sets the parameters for the execution
func WithParameters(params map[string]interface{}) RunnerOption {
	return func(opts *RunnerOptions) {
		opts.Parameters = params
	}
}

// WithEngine sets a specific engine for the execution
func WithEngine(engine string) RunnerOption {
	return func(opts *RunnerOptions) {
		opts.Engine = engine
	}
}

// WithSecurityProfile sets a specific security profile for the execution
func WithSecurityProfile(profile string) RunnerOption {
	return func(opts *RunnerOptions) {
		opts.SecurityProfile = profile
	}
}

// WithProgressHandler sets a progress handler for the execution
func WithProgressHandler(handler func(Progress)) RunnerOption {
	return func(opts *RunnerOptions) {
		opts.ProgressHandler = handler
	}
}

// Progress represents the progress of script execution
type Progress struct {
	Stage       string    // Current stage (e.g., "parsing", "executing", "cleanup")
	Message     string    // Human-readable message
	Percentage  int       // Progress percentage (0-100)
	CurrentStep int       // Current step number
	TotalSteps  int       // Total number of steps
	StartTime   time.Time // When this stage started
}

// RunnerMetrics tracks execution statistics
type RunnerMetrics struct {
	ScriptsExecuted   int64                    `json:"scripts_executed"`
	TotalDuration     time.Duration            `json:"total_duration"`
	AverageDuration   time.Duration            `json:"average_duration"`
	SuccessCount      int64                    `json:"success_count"`
	ErrorCount        int64                    `json:"error_count"`
	LastExecutionTime time.Time                `json:"last_execution_time"`
	EngineMetrics     map[string]*EngineMetric `json:"engine_metrics"`
}

// EngineMetric tracks statistics for a specific engine
type EngineMetric struct {
	ExecutionCount int64         `json:"execution_count"`
	TotalDuration  time.Duration `json:"total_duration"`
	ErrorCount     int64         `json:"error_count"`
}

// SuccessRate calculates the success rate as a percentage
func (m *RunnerMetrics) SuccessRate() float64 {
	total := m.SuccessCount + m.ErrorCount
	if total == 0 {
		return 0.0
	}
	return float64(m.SuccessCount) / float64(total)
}

// ExecutionResult contains the result of a script execution
type ExecutionResult struct {
	Value     interface{}            // The return value from the script
	Error     error                  // Any error that occurred
	Duration  time.Duration          // How long the execution took
	Engine    string                 // Which engine was used
	StartTime time.Time              // When execution started
	EndTime   time.Time              // When execution ended
	Metadata  map[string]interface{} // Additional metadata
}

// IsSuccess returns true if the execution was successful
func (r *ExecutionResult) IsSuccess() bool {
	return r.Error == nil
}

// IsError returns true if the execution resulted in an error
func (r *ExecutionResult) IsError() bool {
	return r.Error != nil
}
