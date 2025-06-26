// ABOUTME: LuaEngine implements the ScriptEngine interface for Lua script execution using gopher-lua
// ABOUTME: Integrates LStatePool, TypeConverter, SecurityManager for comprehensive Lua scripting support

// Package gopherlua provides a Lua scripting engine implementation for go-llmspell.
// It wraps the gopher-lua library to provide a complete Lua 5.1 compatible scripting
// environment with support for bridges, async operations, debugging, and security sandboxing.
//
// The package includes:
//   - Full Lua 5.1 compatibility via gopher-lua
//   - Bridge system for exposing Go functionality to Lua
//   - Async/coroutine support with channels
//   - Script validation and security sandboxing
//   - Performance profiling and optimization
//   - Comprehensive standard library
//
// Example usage:
//
//	engine := lua.NewEngine()
//	err := engine.Initialize(engine.EngineConfig{
//	    SandboxMode: true,
//	    MemoryLimit: 100 * 1024 * 1024, // 100MB
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer engine.Shutdown()
//
//	result, err := engine.Execute(ctx, "return 'Hello from Lua!'", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Output: Hello from Lua!
package lua

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
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/factory"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/converters"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// LuaEngine implements the engine.ScriptEngine interface for Lua scripting.
// It provides a complete Lua 5.1 environment with extensions for async operations,
// bridge integration, and resource management.
type LuaEngine struct {
	// Core components
	pool         *LStatePool
	factory      *LStateFactory
	converter    *converters.LuaTypeConverter
	eventBus     engine.EventBus
	typeRegistry engine.TypeRegistry

	// Configuration
	config       engine.EngineConfig
	initialized  bool
	shuttingDown bool
	mu           sync.RWMutex

	// Bridge management
	bridgeManager *BridgeManager
	bridgeMap     map[string]engine.Bridge
	bridgeMapMu   sync.RWMutex

	// Adapter management  
	adapterFactory AdapterFactoryInterface
	adapters       map[string]interface{}
	adapterMu      sync.RWMutex

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
	chunkCache *converters.ChunkCache

	// Performance profiling
	profiler ProfilerInterface
}

// EngineMetrics tracks Lua engine performance metrics.
// All fields use atomic operations for thread-safe updates.
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

// NewLuaEngine creates a new Lua script engine with default configuration.
// The engine must be initialized with Initialize() before use.
func NewLuaEngine() *LuaEngine {
	converter := converters.NewLuaTypeConverter()
	
	e := &LuaEngine{
		converter:      converter,
		bridgeMap:      make(map[string]engine.Bridge),
		adapterFactory: factory.NewAdapterFactory(),
		adapters:       make(map[string]interface{}),
		chunkCache: converters.NewChunkCache(converters.ChunkCacheConfig{
			MaxSize:         100,
			TTL:             30 * time.Minute,
			EnableDiskCache: false,
		}),
		profiler: NewProfiler(), // Initialize with default profiler
	}
	
	// Create moduleCreator closure that accesses adapters
	moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
		adapter := e.GetAdapter(bridgeID)
		if adapter == nil {
			return nil, fmt.Errorf("no adapter found for bridge %s", bridgeID)
		}
		
		// Type assert to get CreateLuaModule method
		type moduleProvider interface {
			CreateLuaModule() lua.LGFunction
		}
		
		provider, ok := adapter.(moduleProvider)
		if !ok {
			return nil, fmt.Errorf("adapter for bridge %s does not implement CreateLuaModule", bridgeID)
		}
		
		return provider.CreateLuaModule(), nil
	}
	
	// Create BridgeManager with moduleCreator
	e.bridgeManager = NewBridgeManagerWithModuleCreator(converter, moduleCreator)
	
	return e
}

// AdapterFactoryInterface defines the interface for adapter factories.
// This allows dependency injection of custom factories for testing.
type AdapterFactoryInterface interface {
	CreateAdapter(bridgeID string, bridgeMap map[string]engine.Bridge) (interface{}, error)
}

// NewLuaEngineWithFactory creates a new Lua script engine with a custom adapter factory.
// This constructor allows dependency injection for testing with mock factories.
func NewLuaEngineWithFactory(adapterFactory AdapterFactoryInterface) *LuaEngine {
	converter := converters.NewLuaTypeConverter()
	
	e := &LuaEngine{
		converter:      converter,
		bridgeMap:      make(map[string]engine.Bridge),
		adapterFactory: adapterFactory,
		adapters:       make(map[string]interface{}),
		chunkCache: converters.NewChunkCache(converters.ChunkCacheConfig{
			MaxSize:         100,
			TTL:             30 * time.Minute,
			EnableDiskCache: false,
		}),
		profiler: NewProfiler(), // Initialize with default profiler
	}
	
	// Create moduleCreator closure that accesses adapters
	moduleCreator := func(bridgeID string) (lua.LGFunction, error) {
		adapter := e.GetAdapter(bridgeID)
		if adapter == nil {
			return nil, fmt.Errorf("no adapter found for bridge %s", bridgeID)
		}
		
		// Type assert to get CreateLuaModule method
		type moduleProvider interface {
			CreateLuaModule() lua.LGFunction
		}
		
		provider, ok := adapter.(moduleProvider)
		if !ok {
			return nil, fmt.Errorf("adapter for bridge %s does not implement CreateLuaModule", bridgeID)
		}
		
		return provider.CreateLuaModule(), nil
	}
	
	// Create BridgeManager with moduleCreator
	e.bridgeManager = NewBridgeManagerWithModuleCreator(converter, moduleCreator)
	
	return e
}

// SetProfiler sets the profiler for the engine.
// This should be called before initialization if custom profiling is needed.
func (e *LuaEngine) SetProfiler(profiler ProfilerInterface) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.profiler = profiler
}

// GetProfiler returns the current profiler instance.
func (e *LuaEngine) GetProfiler() ProfilerInterface {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.profiler
}

// Initialize initializes the Lua engine with the given configuration.
// This must be called before any script execution. It sets up the LState pool,
// security manager, and loads the standard library.
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
		Level: SecurityLevelStandard, // Default - still using local constants for now, will be updated in security.go
	}

	if config.SandboxMode {
		securityConfig.Level = SecurityLevelStrict
	}

	// Override with engine-specific options using centralized security levels
	if secLevel, ok := config.EngineOptions["security_level"].(security.SecurityLevel); ok {
		switch secLevel {
		case security.SecurityLevelUntrusted:
			securityConfig.Level = SecurityLevelStrict // Map untrusted to strict
		case security.SecurityLevelTrusted:
			securityConfig.Level = SecurityLevelStandard // Map trusted to standard
		case security.SecurityLevelPrivileged:
			securityConfig.Level = SecurityLevelMinimal // Map privileged to minimal
		}
	} else if secLevelStr, ok := config.EngineOptions["security_level"].(string); ok && security.IsValidLevel(secLevelStr) {
		// Handle string values using centralized validation
		switch security.SecurityLevel(secLevelStr) {
		case security.SecurityLevelUntrusted:
			securityConfig.Level = SecurityLevelStrict
		case security.SecurityLevelTrusted:
			securityConfig.Level = SecurityLevelStandard
		case security.SecurityLevelPrivileged:
			securityConfig.Level = SecurityLevelMinimal
		}
	}

	if config.AllowedModules != nil {
		securityConfig.AllowedLibraries = config.AllowedModules
		// If stdlib is enabled (default) and package is not explicitly allowed, add it
		if disableStdlib, ok := config.EngineOptions["disable_stdlib"].(bool); !ok || !disableStdlib {
			hasPackage := false
			for _, lib := range securityConfig.AllowedLibraries {
				if lib == "package" {
					hasPackage = true
					break
				}
			}
			if !hasPackage {
				securityConfig.AllowedLibraries = append(securityConfig.AllowedLibraries, "package")
			}
		}
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

	// Check if stdlib loading is disabled
	if val, ok := config.EngineOptions["disable_stdlib"].(bool); ok && val {
		// User explicitly disabled stdlib
		factoryConfig.DisableStdlib = true
	}

	// Create factory (stdlib is loaded by default unless disabled)
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

// Validate validates a Lua script for syntax errors
func (e *LuaEngine) Validate(script string) error {
	// Use the script validator to check syntax
	validator := NewScriptValidator(DefaultValidatorConfig())
	result, err := validator.ValidateScript(script, "<script>")
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// If there are validation errors, return the first one
	if !result.Valid && len(result.Errors) > 0 {
		firstError := result.Errors[0]
		return fmt.Errorf("%s at line %d, column %d", firstError.Message, firstError.Line, firstError.Column)
	}
	
	return nil
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

	// Cleanup adapters (we don't clear the map to avoid breaking references)
	// Adapters will be garbage collected when no longer referenced

	e.initialized = false
	e.shuttingDown = false
	return nil
}

// Cleanup performs cleanup of the engine resources
func (e *LuaEngine) Cleanup(ctx context.Context) error {
	return e.Shutdown()
}

// RegisterBridge registers a bridge with the engine and creates its adapter
func (e *LuaEngine) RegisterBridge(bridge engine.Bridge) error {
	if !e.initialized {
		return fmt.Errorf("engine not initialized")
	}

	if bridge == nil {
		return fmt.Errorf("bridge cannot be nil")
	}

	bridgeID := bridge.GetID()

	// Store bridge in bridge map
	e.bridgeMapMu.Lock()
	e.bridgeMap[bridgeID] = bridge
	e.bridgeMapMu.Unlock()

	// Create adapter for the bridge using factory
	if e.adapterFactory != nil {
		adapter, err := e.adapterFactory.CreateAdapter(bridgeID, e.bridgeMap)
		if err != nil {
			// Remove bridge from map on failure
			e.bridgeMapMu.Lock()
			delete(e.bridgeMap, bridgeID)
			e.bridgeMapMu.Unlock()
			return fmt.Errorf("failed to create adapter for bridge %s: %w", bridgeID, err)
		}

		// Store adapter
		e.adapterMu.Lock()
		e.adapters[bridgeID] = adapter
		e.adapterMu.Unlock()

		// Check if this bridge is an optional dependency for existing multi-bridge adapters
		// and recreate them if needed
		e.recreateRelatedMultiBridgeAdapters(bridgeID)
	}

	// Register with engine
	if err := bridge.RegisterWithEngine(e); err != nil {
		// Clean up adapter on failure
		e.adapterMu.Lock()
		delete(e.adapters, bridgeID)
		e.adapterMu.Unlock()
		return fmt.Errorf("failed to register bridge %s with engine: %w", bridgeID, err)
	}

	return e.bridgeManager.RegisterBridge(bridge)
}

// UnregisterBridge unregisters a bridge from the engine
func (e *LuaEngine) UnregisterBridge(name string) error {
	// Remove adapter
	e.adapterMu.Lock()
	delete(e.adapters, name)
	e.adapterMu.Unlock()
	
	// Remove bridge from bridge map
	e.bridgeMapMu.Lock()
	delete(e.bridgeMap, name)
	e.bridgeMapMu.Unlock()
	
	return e.bridgeManager.UnregisterBridge(name)
}

// GetAdapter retrieves an adapter by bridge ID
func (e *LuaEngine) GetAdapter(bridgeID string) interface{} {
	if bridgeID == "" {
		return nil
	}
	
	e.adapterMu.RLock()
	defer e.adapterMu.RUnlock()
	
	return e.adapters[bridgeID]
}


// GetBridge retrieves a bridge by name
func (e *LuaEngine) GetBridge(name string) (engine.Bridge, error) {
	return e.bridgeManager.GetBridge(name)
}

// ListBridges returns a list of registered bridge names
func (e *LuaEngine) ListBridges() []string {
	return e.bridgeManager.ListBridges()
}

// LoadBridgeModulesIntoState loads all registered bridge modules into a Lua state
// This is primarily used by the REPL to make bridges available in its persistent state
func (e *LuaEngine) LoadBridgeModulesIntoState(L *lua.LState) error {
	if e.bridgeManager == nil {
		return fmt.Errorf("bridge manager not initialized")
	}
	return e.bridgeManager.LoadBridgeModules(L)
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

// recreateRelatedMultiBridgeAdapters checks if the newly registered bridge
// is an optional dependency for any existing multi-bridge adapters and recreates them
func (e *LuaEngine) recreateRelatedMultiBridgeAdapters(newBridgeID string) {
	// Define multi-bridge adapter groups
	utilBridges := []string{"util_core", "auth", "util_auth", "util_debug", "util_errors", "util_json", "llm_utils", "util_llm", "util_script_logger", "util_slog"}
	llmBridges := []string{"llm_core", "llm_providers", "llm_pool"}
	observabilityBridges := []string{"observability_metrics", "observability_tracing", "observability_guardrails"}

	// Helper to check if bridge is in a group
	isInGroup := func(bridgeID string, group []string) bool {
		for _, id := range group {
			if id == bridgeID {
				return true
			}
		}
		return false
	}

	// Helper to recreate adapters for a group
	recreateGroupAdapters := func(group []string) {
		e.adapterMu.Lock()
		defer e.adapterMu.Unlock()

		// Check which bridges in the group have existing adapters
		for _, bridgeID := range group {
			if _, exists := e.adapters[bridgeID]; exists {
				// This adapter exists, try to recreate it with updated bridge map
				e.bridgeMapMu.RLock()
				adapter, err := e.adapterFactory.CreateAdapter(bridgeID, e.bridgeMap)
				e.bridgeMapMu.RUnlock()
				
				if err == nil {
					// Successfully created new adapter, replace the old one
					e.adapters[bridgeID] = adapter
				}
			}
		}
	}

	// Check which group the new bridge belongs to and recreate related adapters
	if isInGroup(newBridgeID, utilBridges) {
		recreateGroupAdapters(utilBridges)
	} else if isInGroup(newBridgeID, llmBridges) {
		recreateGroupAdapters(llmBridges)
	} else if isInGroup(newBridgeID, observabilityBridges) {
		recreateGroupAdapters(observabilityBridges)
	}
}
