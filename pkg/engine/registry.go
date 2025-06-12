// ABOUTME: This file implements the engine registry for managing multiple script engines.
// ABOUTME: It provides thread-safe registration, discovery, and factory patterns for script engines.

package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Registry manages multiple script engines and provides factory functionality.
type Registry struct {
	mu          sync.RWMutex
	engines     map[string]EngineFactory
	instances   map[string]ScriptEngine
	metrics     map[string]*EngineStats
	config      RegistryConfig
	initialized bool
}

// EngineFactory creates new instances of a script engine.
type EngineFactory interface {
	// Create a new engine instance
	Create(config EngineConfig) (ScriptEngine, error)
	
	// Engine metadata
	Name() string
	Version() string
	Description() string
	FileExtensions() []string
	Features() []EngineFeature
	
	// Validation
	ValidateConfig(config EngineConfig) error
	GetDefaultConfig() EngineConfig
}

// RegistryConfig configures the engine registry behavior.
type RegistryConfig struct {
	// Engine management
	MaxEngines        int           `json:"max_engines"`
	DefaultTimeout    time.Duration `json:"default_timeout"`
	HealthCheckPeriod time.Duration `json:"health_check_period"`
	
	// Instance management
	PoolingEnabled    bool          `json:"pooling_enabled"`
	MaxPoolSize       int           `json:"max_pool_size"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	
	// Security
	RequireSignature  bool     `json:"require_signature"`
	AllowedEngines    []string `json:"allowed_engines"`
	DisallowedEngines []string `json:"disallowed_engines"`
	
	// Observability
	MetricsEnabled    bool `json:"metrics_enabled"`
	TracingEnabled    bool `json:"tracing_enabled"`
	LoggingEnabled    bool `json:"logging_enabled"`
}

// EngineStats tracks statistics for an engine.
type EngineStats struct {
	Name           string        `json:"name"`
	InstancesCreated int64       `json:"instances_created"`
	InstancesActive  int64       `json:"instances_active"`
	TotalExecTime    time.Duration `json:"total_exec_time"`
	SuccessCount     int64       `json:"success_count"`
	ErrorCount       int64       `json:"error_count"`
	LastUsed         time.Time   `json:"last_used"`
	HealthStatus     HealthStatus `json:"health_status"`
}

// HealthStatus represents the health of an engine.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// EngineInfo provides information about a registered engine.
type EngineInfo struct {
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	Description    string          `json:"description"`
	FileExtensions []string        `json:"file_extensions"`
	Features       []EngineFeature `json:"features"`
	Status         EngineStatus    `json:"status"`
	Stats          *EngineStats    `json:"stats,omitempty"`
}

// EngineStatus represents the status of an engine in the registry.
type EngineStatus string

const (
	EngineStatusRegistered EngineStatus = "registered"
	EngineStatusActive     EngineStatus = "active"
	EngineStatusInactive   EngineStatus = "inactive"
	EngineStatusError      EngineStatus = "error"
)

// Global registry instance
var globalRegistry = &Registry{
	engines:   make(map[string]EngineFactory),
	instances: make(map[string]ScriptEngine),
	metrics:   make(map[string]*EngineStats),
}

// GetRegistry returns the global engine registry.
func GetRegistry() *Registry {
	return globalRegistry
}

// NewRegistry creates a new engine registry with the given configuration.
func NewRegistry(config RegistryConfig) *Registry {
	return &Registry{
		engines:   make(map[string]EngineFactory),
		instances: make(map[string]ScriptEngine),
		metrics:   make(map[string]*EngineStats),
		config:    config,
	}
}

// Initialize initializes the registry with default configuration.
func (r *Registry) Initialize() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.initialized {
		return fmt.Errorf("registry already initialized")
	}
	
	// Set default configuration if not provided
	if r.config.MaxEngines == 0 {
		r.config.MaxEngines = 10
	}
	if r.config.DefaultTimeout == 0 {
		r.config.DefaultTimeout = 30 * time.Second
	}
	if r.config.HealthCheckPeriod == 0 {
		r.config.HealthCheckPeriod = 60 * time.Second
	}
	if r.config.MaxPoolSize == 0 {
		r.config.MaxPoolSize = 5
	}
	if r.config.IdleTimeout == 0 {
		r.config.IdleTimeout = 10 * time.Minute
	}
	
	r.initialized = true
	
	// Start health check routine if enabled
	if r.config.MetricsEnabled {
		go r.healthCheckRoutine()
	}
	
	return nil
}

// Register registers a new engine factory.
func (r *Registry) Register(factory EngineFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := factory.Name()
	if name == "" {
		return fmt.Errorf("engine name cannot be empty")
	}
	
	// Check if engine is allowed
	if len(r.config.AllowedEngines) > 0 {
		allowed := false
		for _, allowed_engine := range r.config.AllowedEngines {
			if allowed_engine == name {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("engine %s is not in allowed list", name)
		}
	}
	
	// Check if engine is disallowed
	for _, disallowed := range r.config.DisallowedEngines {
		if disallowed == name {
			return fmt.Errorf("engine %s is disallowed", name)
		}
	}
	
	// Check maximum engines limit
	if len(r.engines) >= r.config.MaxEngines {
		return fmt.Errorf("maximum number of engines (%d) reached", r.config.MaxEngines)
	}
	
	if _, exists := r.engines[name]; exists {
		return fmt.Errorf("engine %s already registered", name)
	}
	
	r.engines[name] = factory
	r.metrics[name] = &EngineStats{
		Name:         name,
		HealthStatus: HealthStatusUnknown,
		LastUsed:     time.Now(),
	}
	
	return nil
}

// Unregister removes an engine factory from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.engines[name]; !exists {
		return fmt.Errorf("engine %s not found", name)
	}
	
	// Shutdown any active instances
	if instance, exists := r.instances[name]; exists {
		_ = instance.Shutdown()
		delete(r.instances, name)
	}
	
	delete(r.engines, name)
	delete(r.metrics, name)
	
	return nil
}

// GetEngine gets or creates an engine instance.
func (r *Registry) GetEngine(name string, config EngineConfig) (ScriptEngine, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	factory, exists := r.engines[name]
	if !exists {
		return nil, fmt.Errorf("engine %s not found", name)
	}
	
	// Return existing instance if pooling is disabled or not found
	if !r.config.PoolingEnabled {
		stats := r.metrics[name]
		stats.InstancesCreated++
		stats.LastUsed = time.Now()
		
		engine, err := factory.Create(config)
		if err != nil {
			stats.ErrorCount++
			return nil, fmt.Errorf("failed to create engine %s: %w", name, err)
		}
		
		stats.InstancesActive++
		stats.SuccessCount++
		return engine, nil
	}
	
	// For pooling, check if we have an existing instance
	if instance, exists := r.instances[name]; exists {
		stats := r.metrics[name]
		stats.LastUsed = time.Now()
		stats.SuccessCount++
		return instance, nil
	}
	
	// Create new instance
	engine, err := factory.Create(config)
	if err != nil {
		r.metrics[name].ErrorCount++
		return nil, fmt.Errorf("failed to create engine %s: %w", name, err)
	}
	
	r.instances[name] = engine
	stats := r.metrics[name]
	stats.InstancesCreated++
	stats.InstancesActive++
	stats.LastUsed = time.Now()
	stats.SuccessCount++
	
	return engine, nil
}

// ListEngines returns information about all registered engines.
func (r *Registry) ListEngines() []EngineInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	engines := make([]EngineInfo, 0, len(r.engines))
	for name, factory := range r.engines {
		status := EngineStatusRegistered
		if _, exists := r.instances[name]; exists {
			status = EngineStatusActive
		}
		
		info := EngineInfo{
			Name:           factory.Name(),
			Version:        factory.Version(),
			Description:    factory.Description(),
			FileExtensions: factory.FileExtensions(),
			Features:       factory.Features(),
			Status:         status,
		}
		
		if r.config.MetricsEnabled {
			if stats, exists := r.metrics[name]; exists {
				info.Stats = stats
			}
		}
		
		engines = append(engines, info)
	}
	
	return engines
}

// FindEngineByExtension finds the best engine for a file extension.
func (r *Registry) FindEngineByExtension(extension string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	extension = strings.TrimPrefix(extension, ".")
	extension = strings.ToLower(extension)
	
	for name, factory := range r.engines {
		for _, ext := range factory.FileExtensions() {
			if strings.ToLower(ext) == extension {
				return name, nil
			}
		}
	}
	
	return "", fmt.Errorf("no engine found for extension .%s", extension)
}

// FindEngineByFeature finds engines that support a specific feature.
func (r *Registry) FindEngineByFeature(feature EngineFeature) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var engines []string
	for name, factory := range r.engines {
		for _, f := range factory.Features() {
			if f == feature {
				engines = append(engines, name)
				break
			}
		}
	}
	
	return engines
}

// GetEngineInfo returns information about a specific engine.
func (r *Registry) GetEngineInfo(name string) (*EngineInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	factory, exists := r.engines[name]
	if !exists {
		return nil, fmt.Errorf("engine %s not found", name)
	}
	
	status := EngineStatusRegistered
	if _, exists := r.instances[name]; exists {
		status = EngineStatusActive
	}
	
	info := &EngineInfo{
		Name:           factory.Name(),
		Version:        factory.Version(),
		Description:    factory.Description(),
		FileExtensions: factory.FileExtensions(),
		Features:       factory.Features(),
		Status:         status,
	}
	
	if r.config.MetricsEnabled {
		if stats, exists := r.metrics[name]; exists {
			info.Stats = stats
		}
	}
	
	return info, nil
}

// GetStats returns statistics for all engines.
func (r *Registry) GetStats() map[string]*EngineStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stats := make(map[string]*EngineStats)
	for name, stat := range r.metrics {
		// Create a copy to avoid race conditions
		statCopy := *stat
		stats[name] = &statCopy
	}
	
	return stats
}

// ExecuteScript executes a script using the appropriate engine.
func (r *Registry) ExecuteScript(ctx context.Context, engineName, script string, params map[string]interface{}) (interface{}, error) {
	engine, err := r.GetEngine(engineName, EngineConfig{})
	if err != nil {
		return nil, err
	}
	
	start := time.Now()
	result, err := engine.Execute(ctx, script, params)
	duration := time.Since(start)
	
	// Update metrics
	r.mu.Lock()
	if stats, exists := r.metrics[engineName]; exists {
		stats.TotalExecTime += duration
		stats.LastUsed = time.Now()
		if err != nil {
			stats.ErrorCount++
		} else {
			stats.SuccessCount++
		}
	}
	r.mu.Unlock()
	
	return result, err
}

// ExecuteFile executes a script file using the appropriate engine based on extension.
func (r *Registry) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (interface{}, error) {
	// Determine engine by file extension
	parts := strings.Split(filepath, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("cannot determine engine for file without extension: %s", filepath)
	}
	
	extension := parts[len(parts)-1]
	engineName, err := r.FindEngineByExtension(extension)
	if err != nil {
		return nil, err
	}
	
	engine, err := r.GetEngine(engineName, EngineConfig{})
	if err != nil {
		return nil, err
	}
	
	start := time.Now()
	result, err := engine.ExecuteFile(ctx, filepath, params)
	duration := time.Since(start)
	
	// Update metrics
	r.mu.Lock()
	if stats, exists := r.metrics[engineName]; exists {
		stats.TotalExecTime += duration
		stats.LastUsed = time.Now()
		if err != nil {
			stats.ErrorCount++
		} else {
			stats.SuccessCount++
		}
	}
	r.mu.Unlock()
	
	return result, err
}

// Shutdown shuts down all engines and cleans up resources.
func (r *Registry) Shutdown() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var errors []string
	
	// Shutdown all active instances
	for name, instance := range r.instances {
		if err := instance.Shutdown(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}
	
	// Clear all data
	r.engines = make(map[string]EngineFactory)
	r.instances = make(map[string]ScriptEngine)
	r.metrics = make(map[string]*EngineStats)
	r.initialized = false
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// healthCheckRoutine runs periodic health checks on engines.
func (r *Registry) healthCheckRoutine() {
	ticker := time.NewTicker(r.config.HealthCheckPeriod)
	defer ticker.Stop()
	
	for range ticker.C {
		r.performHealthChecks()
	}
}

// performHealthChecks checks the health of all engine instances.
func (r *Registry) performHealthChecks() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for name, instance := range r.instances {
		stats := r.metrics[name]
		
		// Simple health check - try to get metrics
		if metrics := instance.GetMetrics(); metrics.ErrorCount > stats.ErrorCount*2 {
			stats.HealthStatus = HealthStatusDegraded
		} else {
			stats.HealthStatus = HealthStatusHealthy
		}
		
		// Check if instance has been idle too long
		if time.Since(stats.LastUsed) > r.config.IdleTimeout {
			_ = instance.Shutdown()
			delete(r.instances, name)
			stats.InstancesActive--
		}
	}
}