// ABOUTME: This file provides integration with the engine registry for managing script engines.
// ABOUTME: It wraps the existing engine registry and provides runner-specific functionality.

package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// EngineRegistryManager wraps the engine registry for use by the runner
type EngineRegistryManager struct {
	registry *engine.Registry
}

// NewEngineRegistryManager creates a new engine registry manager
func NewEngineRegistryManager(registry *engine.Registry) *EngineRegistryManager {
	return &EngineRegistryManager{
		registry: registry,
	}
}

// Initialize initializes the registry
func (m *EngineRegistryManager) Initialize() error {
	// If already initialized, that's OK
	err := m.registry.Initialize()
	if err != nil && err.Error() == "registry already initialized" {
		return nil
	}
	return err
}

// RegisterEngines registers multiple engine factories
func (m *EngineRegistryManager) RegisterEngines(factories map[string]engine.EngineFactory) error {
	for name, factory := range factories {
		if err := m.registry.Register(factory); err != nil {
			return fmt.Errorf("failed to register engine %s: %w", name, err)
		}
	}
	return nil
}

// GetEngine gets or creates an engine instance
func (m *EngineRegistryManager) GetEngine(name string, config engine.EngineConfig) (engine.ScriptEngine, error) {
	return m.registry.GetEngine(name, config)
}

// FindEngineByExtension finds the best engine for a file extension
func (m *EngineRegistryManager) FindEngineByExtension(extension string) (string, error) {
	return m.registry.FindEngineByExtension(extension)
}

// ListEngines returns information about all registered engines
func (m *EngineRegistryManager) ListEngines() []engine.EngineInfo {
	return m.registry.ListEngines()
}

// GetEngineInfo returns information about a specific engine
func (m *EngineRegistryManager) GetEngineInfo(name string) (*engine.EngineInfo, error) {
	return m.registry.GetEngineInfo(name)
}

// ExecuteScript executes a script using the specified engine
func (m *EngineRegistryManager) ExecuteScript(ctx context.Context, engineName, script string, params map[string]interface{}) (interface{}, error) {
	return m.registry.ExecuteScript(ctx, engineName, script, params)
}

// ExecuteFile executes a script file using the appropriate engine
func (m *EngineRegistryManager) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	return m.registry.ExecuteFile(ctx, filepath, params)
}

// GetStats returns statistics for all engines
func (m *EngineRegistryManager) GetStats() map[string]*engine.EngineStats {
	return m.registry.GetStats()
}

// Shutdown shuts down all engines and cleans up resources
func (m *EngineRegistryManager) Shutdown() error {
	return m.registry.Shutdown()
}

// BuildEngineConfig builds an engine configuration from runner config and engine-specific settings
func BuildEngineConfig(runnerConfig *RunnerConfig, engineConfig map[string]interface{}) engine.EngineConfig {
	config := engine.EngineConfig{
		MemoryLimit:    64 * 1024 * 1024, // 64MB default
		TimeoutLimit:   30 * time.Second, // 30 seconds default
		GoroutineLimit: 100,              // 100 goroutines default
		SandboxMode:    true,
		FileSystemMode: engine.FSModeReadOnly,
		EngineOptions:  make(map[string]interface{}),
		DebugMode:      false,
		LogLevel:       "info",
		MetricsMode:    true,
		TracingMode:    false,
	}

	// Apply runner config if provided
	if runnerConfig != nil {
		if runnerConfig.Timeout > 0 {
			config.TimeoutLimit = runnerConfig.Timeout
		}
		if runnerConfig.EnableDebug {
			config.DebugMode = true
		}
		if runnerConfig.EnableMetrics {
			config.MetricsMode = true
		}
		// Store working directory and environment in engine options
		if runnerConfig.WorkingDirectory != "" {
			config.EngineOptions["working_directory"] = runnerConfig.WorkingDirectory
		}
		if len(runnerConfig.Environment) > 0 {
			config.EngineOptions["environment"] = runnerConfig.Environment
		}
	}

	// Apply engine-specific config if provided
	if engineConfig != nil {
		if memLimit, ok := engineConfig["memory_limit"].(int64); ok {
			config.MemoryLimit = memLimit
		} else if memLimit, ok := engineConfig["memory_limit"].(int); ok {
			config.MemoryLimit = int64(memLimit)
		}

		if timeout, ok := engineConfig["timeout_limit"].(time.Duration); ok {
			config.TimeoutLimit = timeout
		} else if timeoutSecs, ok := engineConfig["timeout"].(int); ok {
			config.TimeoutLimit = time.Duration(timeoutSecs) * time.Second
		}

		if goroutineLimit, ok := engineConfig["goroutine_limit"].(int); ok {
			config.GoroutineLimit = goroutineLimit
		}

		if sandboxed, ok := engineConfig["sandbox_mode"].(bool); ok {
			config.SandboxMode = sandboxed
		}

		if fsMode, ok := engineConfig["filesystem_mode"].(string); ok {
			switch fsMode {
			case "readonly":
				config.FileSystemMode = engine.FSModeReadOnly
			case "readwrite":
				config.FileSystemMode = engine.FSModeReadWrite
			case "none":
				config.FileSystemMode = engine.FSModeNone
			}
		}

		if debug, ok := engineConfig["debug_mode"].(bool); ok {
			config.DebugMode = debug
		}

		if logLevel, ok := engineConfig["log_level"].(string); ok {
			config.LogLevel = logLevel
		}

		if metrics, ok := engineConfig["metrics_mode"].(bool); ok {
			config.MetricsMode = metrics
		}

		if tracing, ok := engineConfig["tracing_mode"].(bool); ok {
			config.TracingMode = tracing
		}

		// Copy engine-specific options
		for k, v := range engineConfig {
			switch k {
			case "memory_limit", "timeout_limit", "timeout", "goroutine_limit",
				"sandbox_mode", "filesystem_mode", "debug_mode", "log_level",
				"metrics_mode", "tracing_mode":
				// Already handled above
			default:
				config.EngineOptions[k] = v
			}
		}
	}

	return config
}

// ApplyOptionsToConfig applies RunnerOptions to an engine config
func ApplyOptionsToConfig(config engine.EngineConfig, options *RunnerOptions) engine.EngineConfig {
	if options == nil {
		return config
	}

	if options.Timeout > 0 {
		config.TimeoutLimit = options.Timeout
	}
	if options.Debug {
		config.DebugMode = true
	}

	return config
}

// GetEngineForSpell determines the appropriate engine for a spell
func GetEngineForSpell(manager *EngineRegistryManager, metadata *SpellMetadata) (string, error) {
	// If engine is explicitly specified in metadata, use it
	if metadata.Engine != "" {
		// Verify the engine exists
		if _, err := manager.GetEngineInfo(metadata.Engine); err != nil {
			return "", fmt.Errorf("specified engine %s not found: %w", metadata.Engine, err)
		}
		return metadata.Engine, nil
	}

	// Try to determine engine from entry point extension
	if metadata.EntryPoint != "" {
		engine, err := manager.FindEngineByExtension(metadata.EntryPoint)
		if err == nil {
			return engine, nil
		}
	}

	return "", fmt.Errorf("unable to determine engine for spell %s", metadata.Name)
}

// CreateEngineMetrics creates engine metrics from registry stats
func CreateEngineMetrics(stats map[string]*engine.EngineStats) map[string]*EngineMetric {
	metrics := make(map[string]*EngineMetric)

	for name, stat := range stats {
		metrics[name] = &EngineMetric{
			ExecutionCount: stat.SuccessCount + stat.ErrorCount,
			TotalDuration:  stat.TotalExecTime,
			ErrorCount:     stat.ErrorCount,
		}
	}

	return metrics
}
