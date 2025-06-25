// ABOUTME: This file provides integration with the engine registry for managing script engines.
// ABOUTME: It wraps the existing engine registry and provides runner-specific functionality.

package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// EngineRegistryManager wraps the engine registry for use by the runner.
// It provides a higher-level interface for engine management, including
// registration, execution, and statistics gathering.
type EngineRegistryManager struct {
	registry    *engine.Registry
	bridgeCache map[string]bool // tracks which engine+featureset combinations have bridges loaded
	cacheMutex  sync.RWMutex    // protects bridgeCache from concurrent access
	config      *RunnerConfig   // configuration for engine-specific feature sets
}

// NewEngineRegistryManager creates a new engine registry manager.
// It wraps the provided engine registry for runner-specific operations.
func NewEngineRegistryManager(registry *engine.Registry, config *RunnerConfig) *EngineRegistryManager {
	return &EngineRegistryManager{
		registry:    registry,
		bridgeCache: make(map[string]bool),
		config:      config,
	}
}

// Initialize initializes the registry.
// It handles the case where the registry is already initialized,
// treating it as a non-error condition.
func (m *EngineRegistryManager) Initialize() error {
	// If already initialized, that's OK
	err := m.registry.Initialize()
	if err != nil && err.Error() == "registry already initialized" {
		return nil
	}
	return err
}

// RegisterEngines registers multiple engine factories.
// It iterates through the provided factories map and registers
// each engine with the underlying registry.
func (m *EngineRegistryManager) RegisterEngines(factories map[string]engine.EngineFactory) error {
	for name, factory := range factories {
		if err := m.registry.Register(factory); err != nil {
			return fmt.Errorf("failed to register engine %s: %w", name, err)
		}
	}
	return nil
}

// GetEngine gets or creates an engine instance with bridges loaded on-demand.
// It ensures bridges are registered for the engine based on the security level and feature set,
// and caches bridge registration to prevent redundant loading.
func (m *EngineRegistryManager) GetEngine(name string, config engine.EngineConfig, securityLevel security.SecurityLevel, featureSet registry.FeatureSet) (engine.ScriptEngine, error) {
	// Get engine instance first
	scriptEngine, err := m.registry.GetEngine(name, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get engine %s: %w", name, err)
	}

	// Create cache key from engine instance, security level, and feature set
	// Use the engine's address as a unique identifier
	cacheKey := fmt.Sprintf("%p:%s:%s", scriptEngine, securityLevel, featureSet)

	// Check if bridges are already loaded for this engine instance+profile combination
	m.cacheMutex.RLock()
	bridgesLoaded := m.bridgeCache[cacheKey]
	m.cacheMutex.RUnlock()

	// If bridges not loaded for this combination, load them now
	if !bridgesLoaded {
		if err := m.loadBridgesForEngine(scriptEngine, securityLevel, featureSet, cacheKey); err != nil {
			return nil, fmt.Errorf("failed to load bridges for engine %s with security level %s and feature set %s: %w", name, securityLevel, featureSet, err)
		}
	}

	return scriptEngine, nil
}

// loadBridgesForEngine loads bridges for an engine based on security level and feature set.
// It registers the appropriate bridges based on the feature set.
func (m *EngineRegistryManager) loadBridgesForEngine(scriptEngine engine.ScriptEngine, securityLevel security.SecurityLevel, featureSet registry.FeatureSet, cacheKey string) error {
	// TODO: In the future, we might want to filter bridges based on security level
	// For now, security level is enforced at runtime, not at bridge registration time

	// Register bridges using the feature set
	if err := registry.RegisterBridgesByFeatureSet(scriptEngine, featureSet); err != nil {
		// If bridges are already registered, that's OK for our lazy loading approach
		// The engine instance may have been reused from the registry pool
		if bridgeAlreadyRegisteredError(err) {
			// Mark as cached even though we didn't register (since they're already there)
			m.cacheMutex.Lock()
			m.bridgeCache[cacheKey] = true
			m.cacheMutex.Unlock()
			return nil
		}
		return fmt.Errorf("failed to register bridges for feature set %s: %w", featureSet, err)
	}

	// Cache that bridges are loaded for this combination
	m.cacheMutex.Lock()
	m.bridgeCache[cacheKey] = true
	m.cacheMutex.Unlock()

	return nil
}

// FindEngineByExtension finds the best engine for a file extension.
// It returns the name of the engine that can handle files with
// the given extension.
func (m *EngineRegistryManager) FindEngineByExtension(extension string) (string, error) {
	return m.registry.FindEngineByExtension(extension)
}

// ListEngines returns information about all registered engines.
// It provides details about engine capabilities, supported extensions,
// and features for each registered engine.
func (m *EngineRegistryManager) ListEngines() []engine.EngineInfo {
	return m.registry.ListEngines()
}

// GetEngineInfo returns information about a specific engine.
// It retrieves detailed information about the engine's capabilities
// and configuration requirements.
func (m *EngineRegistryManager) GetEngineInfo(name string) (*engine.EngineInfo, error) {
	return m.registry.GetEngineInfo(name)
}

// ExecuteScript executes a script using the specified engine.
// It creates or retrieves an engine instance and executes the provided
// script with the given parameters.
func (m *EngineRegistryManager) ExecuteScript(ctx context.Context, engineName, script string, params map[string]interface{}) (interface{}, error) {
	return m.registry.ExecuteScript(ctx, engineName, script, params)
}

// ExecuteFile executes a script file using the appropriate engine.
// It automatically selects the engine based on the file extension
// and executes the file contents.
func (m *EngineRegistryManager) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	return m.registry.ExecuteFile(ctx, filepath, params)
}

// GetStats returns statistics for all engines.
// It provides execution counts, timing information, and error rates
// for performance monitoring and optimization.
func (m *EngineRegistryManager) GetStats() map[string]*engine.EngineStats {
	return m.registry.GetStats()
}

// Shutdown shuts down all engines and cleans up resources.
// It ensures all engine instances are properly terminated and
// resources are released.
func (m *EngineRegistryManager) Shutdown() error {
	return m.registry.Shutdown()
}

// BuildEngineConfig builds an engine configuration from runner config and engine-specific settings.
// It merges runner-level settings with engine-specific overrides to create
// a complete engine configuration with appropriate defaults.
func BuildEngineConfig(runnerConfig *RunnerConfig, engineConfig map[string]interface{}) engine.EngineConfig {
	// Determine sandbox mode based on security level
	sandboxMode := true // Default to true for security
	if runnerConfig != nil && runnerConfig.DefaultSecurityLevel != "" {
		switch runnerConfig.DefaultSecurityLevel {
		case "trusted", "privileged":
			sandboxMode = false
		case "untrusted":
			sandboxMode = true
		}
	}

	config := engine.EngineConfig{
		MemoryLimit:    64 * 1024 * 1024, // 64MB default
		TimeoutLimit:   30 * time.Second, // 30 seconds default
		GoroutineLimit: 100,              // 100 goroutines default
		SandboxMode:    sandboxMode,
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

// ApplyOptionsToConfig applies RunnerOptions to an engine config.
// It overrides specific configuration values based on the provided
// options, such as timeout and debug settings.
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

// GetEngineForSpell determines the appropriate engine for a spell.
// It first checks for an explicitly specified engine in the metadata,
// then attempts to infer from the entry point file extension.
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

// CreateEngineMetrics creates engine metrics from registry stats.
// It converts engine statistics into runner-specific metrics format
// for consistent reporting and monitoring.
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

// SetupEngineRegistry creates and configures an engine registry with lightweight engine factories.
// This function only registers engine factories for discovery and extension mapping.
// Bridge registration is handled lazily when engines are actually requested.
func SetupEngineRegistry(config *RunnerConfig, profile string) (*EngineRegistryManager, error) {
	if config == nil {
		config = DefaultRunnerConfig()
	}

	// Create engine registry with configuration
	registryConfig := engine.RegistryConfig{
		MaxEngines:        10,
		DefaultTimeout:    30 * time.Second,
		HealthCheckPeriod: 60 * time.Second,
		PoolingEnabled:    true,
		MaxPoolSize:       5,
		IdleTimeout:       10 * time.Minute,
		MetricsEnabled:    config.EnableMetrics,
		LoggingEnabled:    config.EnableDebug,
		TracingEnabled:    config.EnableDebug,
	}
	registry := engine.NewRegistry(registryConfig)

	// Initialize registry
	if err := registry.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize engine registry: %w", err)
	}

	// Register lightweight engine factories only (no bridges)
	luaFactory := lua.NewLuaEngineFactory()
	if err := registry.Register(luaFactory); err != nil {
		return nil, fmt.Errorf("failed to register Lua engine factory: %w", err)
	}

	// TODO: Register JavaScript and Tengo engine factories when implemented

	// Create and return engine registry manager
	// Bridge registration will happen on-demand in GetEngine()
	return NewEngineRegistryManager(registry, config), nil
}

// bridgeAlreadyRegisteredError checks if an error indicates that a bridge is already registered.
// This is used to handle the case where engine instances are reused from a pool.
func bridgeAlreadyRegisteredError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "already registered") || strings.Contains(errMsg, "duplicate")
}
