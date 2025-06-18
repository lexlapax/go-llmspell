// ABOUTME: LuaEngine implements the ScriptEngine interface for Lua script execution using gopher-lua
// ABOUTME: Integrates LStatePool, TypeConverter, SecurityManager for comprehensive Lua scripting support

package gopherlua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// LuaEngine implements the engine.ScriptEngine interface for Lua scripting
type LuaEngine struct {
	// Core components
	pool         *LStatePool
	factory      *LStateFactory
	converter    *LuaTypeConverter
	eventBus     engine.EventBus
	typeRegistry engine.TypeRegistry

	// Configuration
	config       engine.EngineConfig
	initialized  bool
	shuttingDown bool
	mu           sync.RWMutex

	// Bridge management
	bridgeManager *BridgeManager

	// Resource limits
	memoryLimit    int64
	timeoutLimit   time.Duration
	resourceLimits engine.ResourceLimits

	// Metrics
	metrics EngineMetrics

	// Profiling
	profilingEnabled bool
	profilingConfig  engine.ProfilingConfig

	// Chunk caching
	chunkCache *ChunkCache
}

// EngineMetrics tracks Lua engine performance
type EngineMetrics struct {
	scriptsExecuted  int64
	totalExecTime    int64 // nanoseconds, use atomic operations
	errorCount       int64
	memoryUsed       int64
	peakMemoryUsed   int64
	bridgeCallsCount int64
	cacheHits        int64
	cacheMisses      int64
	compilationTime  int64 // nanoseconds, use atomic operations
	gcCollections    int64
}

// NewLuaEngine creates a new Lua script engine
func NewLuaEngine() *LuaEngine {
	converter := NewLuaTypeConverter()
	return &LuaEngine{
		converter:     converter,
		bridgeManager: NewBridgeManager(converter),
		chunkCache: NewChunkCache(ChunkCacheConfig{
			MaxSize:         100,
			TTL:             30 * time.Minute,
			EnableDiskCache: false,
		}),
	}
}

// Initialize initializes the Lua engine with the given configuration
func (e *LuaEngine) Initialize(config engine.EngineConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return fmt.Errorf("engine already initialized")
	}

	e.config = config

	// Apply resource limits
	if config.MemoryLimit > 0 {
		e.memoryLimit = config.MemoryLimit
	} else {
		e.memoryLimit = 64 * 1024 * 1024 // 64MB default
	}

	if config.TimeoutLimit > 0 {
		e.timeoutLimit = config.TimeoutLimit
	} else {
		e.timeoutLimit = 30 * time.Second // 30s default
	}

	// Create SecurityManager based on config
	securityConfig := SecurityConfig{
		Level: SecurityLevelStandard, // Default
	}

	if config.SandboxMode {
		securityConfig.Level = SecurityLevelStrict
	}

	// Override with engine-specific options
	if secLevel, ok := config.EngineOptions["security_level"].(string); ok {
		switch secLevel {
		case "minimal":
			securityConfig.Level = SecurityLevelMinimal
		case "standard":
			securityConfig.Level = SecurityLevelStandard
		case "strict":
			securityConfig.Level = SecurityLevelStrict
		}
	}

	if config.AllowedModules != nil {
		securityConfig.AllowedLibraries = config.AllowedModules
	}
	// Note: DisabledModules would need to be handled differently
	// since SecurityConfig doesn't have DeniedLibraries field

	securityManager := NewSecurityManager(securityConfig)

	// Create factory with security manager
	factoryConfig := FactoryConfig{
		SecurityManager: securityManager,
		Options: lua.Options{
			SkipOpenLibs:        true, // We handle library loading through SecurityManager
			IncludeGoStackTrace: config.DebugMode,
		},
	}

	e.factory = NewLStateFactory(factoryConfig)

	// Create pool configuration from engine options
	poolConfig := PoolConfig{
		MinSize:         2,
		MaxSize:         10,
		IdleTimeout:     10 * time.Minute,
		HealthThreshold: 0.7,
		CleanupInterval: time.Minute,
	}

	// Override with engine-specific options
	if minSize, ok := config.EngineOptions["pool_min_size"].(int); ok {
		poolConfig.MinSize = minSize
	}
	if maxSize, ok := config.EngineOptions["pool_max_size"].(int); ok {
		poolConfig.MaxSize = maxSize
	}
	if idleTimeout, ok := config.EngineOptions["pool_idle_timeout"].(string); ok {
		if duration, err := time.ParseDuration(idleTimeout); err == nil {
			poolConfig.IdleTimeout = duration
		}
	}
	if healthThreshold, ok := config.EngineOptions["health_threshold"].(float64); ok {
		poolConfig.HealthThreshold = healthThreshold
	}
	if cleanupInterval, ok := config.EngineOptions["cleanup_interval"].(string); ok {
		if duration, err := time.ParseDuration(cleanupInterval); err == nil {
			poolConfig.CleanupInterval = duration
		}
	}

	// Create LState pool
	pool, err := NewLStatePool(e.factory, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create LState pool: %w", err)
	}
	e.pool = pool

	e.initialized = true
	return nil
}

// Execute executes a Lua script with the given parameters
func (e *LuaEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error) {
	// Use the execution pipeline for cleaner, more maintainable code
	result, err := e.ExecuteWithPipeline(ctx, script, params)
	if err != nil {
		return engine.NewErrorValue(err), err
	}

	// Convert result to ScriptValue
	return e.converter.ToScriptValue(result)
}

// ExecuteFile executes a Lua script from a file
func (e *LuaEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (engine.ScriptValue, error) {
	if !e.initialized {
		return engine.NewErrorValue(fmt.Errorf("engine not initialized")), fmt.Errorf("engine not initialized")
	}

	// Check if file exists and is readable
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewErrorValue(err), fmt.Errorf("cannot access file %s: %w", path, err)
	}
	if info.IsDir() {
		return engine.NewErrorValue(fmt.Errorf("path %s is a directory", path)), fmt.Errorf("path %s is a directory", path)
	}

	// Check file extension
	ext := filepath.Ext(path)
	validExts := e.FileExtensions()
	isValidExt := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValidExt = true
			break
		}
	}
	if !isValidExt {
		return engine.NewErrorValue(fmt.Errorf("unsupported file extension %s", ext)), fmt.Errorf("unsupported file extension %s", ext)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return engine.NewErrorValue(err), fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Execute script content
	return e.Execute(ctx, string(content), params)
}

// Shutdown gracefully shuts down the engine
func (e *LuaEngine) Shutdown() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return nil
	}

	e.shuttingDown = true

	// Shutdown pool
	if e.pool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.pool.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown pool: %w", err)
		}
	}

	// Clear chunk cache
	if e.chunkCache != nil {
		e.chunkCache.Clear()
	}

	// Cleanup bridges
	if e.bridgeManager != nil {
		_ = e.bridgeManager.Cleanup()
	}

	e.initialized = false
	e.shuttingDown = false
	return nil
}

// RegisterBridge registers a bridge with the engine
func (e *LuaEngine) RegisterBridge(bridge engine.Bridge) error {
	if !e.initialized {
		return fmt.Errorf("engine not initialized")
	}

	// Register with engine
	if err := bridge.RegisterWithEngine(e); err != nil {
		return fmt.Errorf("failed to register bridge %s with engine: %w", bridge.GetID(), err)
	}

	return e.bridgeManager.RegisterBridge(bridge)
}

// UnregisterBridge unregisters a bridge from the engine
func (e *LuaEngine) UnregisterBridge(name string) error {
	return e.bridgeManager.UnregisterBridge(name)
}

// GetBridge retrieves a bridge by name
func (e *LuaEngine) GetBridge(name string) (engine.Bridge, error) {
	return e.bridgeManager.GetBridge(name)
}

// ListBridges returns a list of registered bridge names
func (e *LuaEngine) ListBridges() []string {
	return e.bridgeManager.ListBridges()
}

// ToNative converts a ScriptValue to a Go value
func (e *LuaEngine) ToNative(scriptValue engine.ScriptValue) (interface{}, error) {
	if scriptValue == nil || scriptValue.IsNil() {
		return nil, nil
	}
	return e.converter.FromScriptValue(scriptValue), nil
}

// FromNative converts a Go value to a ScriptValue
func (e *LuaEngine) FromNative(goValue interface{}) (engine.ScriptValue, error) {
	return e.converter.ToScriptValue(goValue)
}

// Name returns the engine name
func (e *LuaEngine) Name() string {
	return "lua"
}

// Version returns the engine version
func (e *LuaEngine) Version() string {
	return "1.0.0" // Our engine version
}

// FileExtensions returns supported file extensions
func (e *LuaEngine) FileExtensions() []string {
	return []string{".lua"}
}

// Features returns supported engine features
func (e *LuaEngine) Features() []engine.EngineFeature {
	return []engine.EngineFeature{
		engine.FeatureCoroutines,
		engine.FeatureModules,
		engine.FeatureCompilation,
	}
}

// SetMemoryLimit sets the memory limit for script execution
func (e *LuaEngine) SetMemoryLimit(bytes int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.memoryLimit = bytes
	return nil
}

// SetTimeout sets the timeout limit for script execution
func (e *LuaEngine) SetTimeout(duration time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timeoutLimit = duration
	return nil
}

// SetResourceLimits sets comprehensive resource limits
func (e *LuaEngine) SetResourceLimits(limits engine.ResourceLimits) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.resourceLimits = limits
	if limits.MaxMemory > 0 {
		e.memoryLimit = limits.MaxMemory
	}
	if limits.MaxExecTime > 0 {
		e.timeoutLimit = limits.MaxExecTime
	}
	return nil
}

// GetMetrics returns engine performance metrics
func (e *LuaEngine) GetMetrics() engine.EngineMetrics {
	scriptsExecuted := atomic.LoadInt64(&e.metrics.scriptsExecuted)
	totalExecTimeNs := atomic.LoadInt64(&e.metrics.totalExecTime)
	compilationTimeNs := atomic.LoadInt64(&e.metrics.compilationTime)

	totalExecTime := time.Duration(totalExecTimeNs)
	compilationTime := time.Duration(compilationTimeNs)

	var avgExecTime time.Duration
	if scriptsExecuted > 0 {
		avgExecTime = totalExecTime / time.Duration(scriptsExecuted)
	}

	return engine.EngineMetrics{
		ScriptsExecuted:  scriptsExecuted,
		TotalExecTime:    totalExecTime,
		AverageExecTime:  avgExecTime,
		ErrorCount:       atomic.LoadInt64(&e.metrics.errorCount),
		MemoryUsed:       atomic.LoadInt64(&e.metrics.memoryUsed),
		PeakMemoryUsed:   atomic.LoadInt64(&e.metrics.peakMemoryUsed),
		BridgeCallsCount: atomic.LoadInt64(&e.metrics.bridgeCallsCount),
		CacheHits:        atomic.LoadInt64(&e.metrics.cacheHits),
		CacheMisses:      atomic.LoadInt64(&e.metrics.cacheMisses),
		CompilationTime:  compilationTime,
		GCCollections:    atomic.LoadInt64(&e.metrics.gcCollections),
	}
}

// wrapLuaError wraps a Lua error into an EngineError
func (e *LuaEngine) wrapLuaError(err error, errorType engine.ErrorType) error {
	if err == nil {
		return nil
	}

	engineErr := &engine.EngineError{
		Type:    errorType,
		Message: err.Error(),
		Cause:   err,
	}

	// Try to extract line/column information
	if luaErr, ok := err.(*lua.ApiError); ok {
		// Parse line information from Lua error
		if luaErr.Object != nil {
			if errStr, ok := luaErr.Object.(lua.LString); ok {
				engineErr.Message = string(errStr)
			}
		}
	}

	return engineErr
}

// Placeholder implementations for extended interface methods
// These will be implemented in later tasks

func (e *LuaEngine) CreateContext(options engine.ContextOptions) (engine.ScriptContext, error) {
	return nil, fmt.Errorf("CreateContext not implemented yet")
}

func (e *LuaEngine) DestroyContext(ctx engine.ScriptContext) error {
	return fmt.Errorf("DestroyContext not implemented yet")
}

func (e *LuaEngine) ExecuteScript(ctx context.Context, script string, options engine.ExecutionOptions) (*engine.ExecutionResult, error) {
	startTime := time.Now()

	// Convert variables from ExecutionOptions to map[string]interface{}
	params := make(map[string]interface{})
	for k, v := range options.Variables {
		params[k] = v
	}

	// Execute the script
	result, err := e.Execute(ctx, script, params)
	duration := time.Since(startTime)

	// Create ExecutionResult
	execResult := &engine.ExecutionResult{
		Value:    result,
		Duration: duration,
		Metadata: make(map[string]interface{}),
	}

	// Add execution metadata
	execResult.Metadata["engine"] = e.Name()
	execResult.Metadata["script_length"] = len(script)

	if err != nil {
		execResult.Error = err
		// If we have an error value, use it; otherwise create one
		if result != nil && result.Type() == engine.TypeError {
			execResult.Value = result
		} else {
			execResult.Value = engine.NewErrorValue(err)
		}
	}

	return execResult, err
}

func (e *LuaEngine) GetEventBus() engine.EventBus {
	return e.eventBus
}

func (e *LuaEngine) RegisterTypeConverter(fromType, toType string, converter engine.TypeConverterFunc) error {
	return fmt.Errorf("RegisterTypeConverter not implemented yet")
}

func (e *LuaEngine) GetTypeRegistry() engine.TypeRegistry {
	return e.typeRegistry
}

func (e *LuaEngine) EnableProfiling(config engine.ProfilingConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.profilingEnabled = true
	e.profilingConfig = config
	return nil
}

func (e *LuaEngine) DisableProfiling() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.profilingEnabled = false
	return nil
}

func (e *LuaEngine) GetProfilingReport() (*engine.ProfilingReport, error) {
	return nil, fmt.Errorf("GetProfilingReport not implemented yet")
}

func (e *LuaEngine) ExportAPI(format engine.ExportFormat) ([]byte, error) {
	return nil, fmt.Errorf("ExportAPI not implemented yet")
}

func (e *LuaEngine) GenerateClientLibrary(language string, options engine.ClientLibraryOptions) ([]byte, error) {
	return nil, fmt.Errorf("GenerateClientLibrary not implemented yet")
}
