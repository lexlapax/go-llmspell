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
	bridges   map[string]engine.Bridge
	bridgesMu sync.RWMutex

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
	return &LuaEngine{
		bridges:   make(map[string]engine.Bridge),
		converter: NewLuaTypeConverter(),
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
func (e *LuaEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	if !e.initialized {
		return nil, fmt.Errorf("engine not initialized")
	}

	if e.shuttingDown {
		return nil, fmt.Errorf("engine is shutting down")
	}

	startTime := time.Now()
	defer func() {
		atomic.AddInt64(&e.metrics.scriptsExecuted, 1)
		atomic.AddInt64(&e.metrics.totalExecTime, time.Since(startTime).Nanoseconds())
	}()

	// Apply timeout from context or engine default
	execCtx := ctx
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > e.timeoutLimit {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, e.timeoutLimit)
		defer cancel()
	}

	// Get LState from pool
	state, err := e.pool.Get(execCtx)
	if err != nil {
		atomic.AddInt64(&e.metrics.errorCount, 1)
		return nil, fmt.Errorf("failed to get Lua state: %w", err)
	}
	defer e.pool.Put(state)

	// Set parameters in state
	for key, value := range params {
		luaValue, err := e.converter.ToLua(state, value)
		if err != nil {
			atomic.AddInt64(&e.metrics.errorCount, 1)
			return nil, fmt.Errorf("failed to convert parameter %s: %w", key, err)
		}
		state.SetGlobal(key, luaValue)
	}

	// Try to get compiled chunk from cache
	var compiledChunk *lua.FunctionProto
	cacheKey := e.chunkCache.GenerateKey(script, "")

	if cached := e.chunkCache.Get(cacheKey); cached != nil {
		compiledChunk = cached
		atomic.AddInt64(&e.metrics.cacheHits, 1)
	} else {
		// Compile script
		compileStart := time.Now()
		chunk, err := state.LoadString(script)
		if err != nil {
			atomic.AddInt64(&e.metrics.errorCount, 1)
			return nil, e.wrapLuaError(err, engine.ErrorTypeSyntax)
		}
		compiledChunk = chunk.Proto
		atomic.AddInt64(&e.metrics.compilationTime, time.Since(compileStart).Nanoseconds())
		atomic.AddInt64(&e.metrics.cacheMisses, 1)

		// Cache compiled chunk
		e.chunkCache.Put(cacheKey, compiledChunk)
	}

	// Execute with timeout
	resultChan := make(chan executionResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- executionResult{
					error: fmt.Errorf("panic during execution: %v", r),
				}
			}
		}()

		// Push compiled function and call it
		state.Push(state.NewFunctionFromProto(compiledChunk))
		err := state.PCall(0, lua.MultRet, nil)
		if err != nil {
			resultChan <- executionResult{
				error: e.wrapLuaError(err, engine.ErrorTypeRuntime),
			}
			return
		}

		// Get return value (top of stack)
		var result interface{}
		if state.GetTop() > 0 {
			luaValue := state.Get(-1)
			goValue, err := e.converter.FromLua(luaValue)
			if err != nil {
				resultChan <- executionResult{
					error: fmt.Errorf("failed to convert result: %w", err),
				}
				return
			}
			result = goValue
		}

		resultChan <- executionResult{
			value: result,
		}
	}()

	// Wait for execution or timeout
	select {
	case result := <-resultChan:
		if result.error != nil {
			atomic.AddInt64(&e.metrics.errorCount, 1)
		}
		return result.value, result.error
	case <-execCtx.Done():
		atomic.AddInt64(&e.metrics.errorCount, 1)
		return nil, &engine.EngineError{
			Type:    engine.ErrorTypeTimeout,
			Message: "script execution timed out",
			Cause:   execCtx.Err(),
		}
	}
}

// ExecuteFile executes a Lua script from a file
func (e *LuaEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error) {
	if !e.initialized {
		return nil, fmt.Errorf("engine not initialized")
	}

	// Check if file exists and is readable
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access file %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path %s is a directory", path)
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
		return nil, fmt.Errorf("unsupported file extension %s", ext)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
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
	e.bridgesMu.Lock()
	for _, bridge := range e.bridges {
		if bridge.IsInitialized() {
			ctx := context.Background()
			_ = bridge.Cleanup(ctx)
		}
	}
	e.bridges = make(map[string]engine.Bridge)
	e.bridgesMu.Unlock()

	e.initialized = false
	e.shuttingDown = false
	return nil
}

// RegisterBridge registers a bridge with the engine
func (e *LuaEngine) RegisterBridge(bridge engine.Bridge) error {
	if !e.initialized {
		return fmt.Errorf("engine not initialized")
	}

	e.bridgesMu.Lock()
	defer e.bridgesMu.Unlock()

	id := bridge.GetID()
	if _, exists := e.bridges[id]; exists {
		return fmt.Errorf("bridge %s already registered", id)
	}

	// Initialize bridge if needed
	if !bridge.IsInitialized() {
		ctx := context.Background()
		if err := bridge.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize bridge %s: %w", id, err)
		}
	}

	// Register with engine
	if err := bridge.RegisterWithEngine(e); err != nil {
		return fmt.Errorf("failed to register bridge %s with engine: %w", id, err)
	}

	e.bridges[id] = bridge
	return nil
}

// UnregisterBridge unregisters a bridge from the engine
func (e *LuaEngine) UnregisterBridge(name string) error {
	e.bridgesMu.Lock()
	defer e.bridgesMu.Unlock()

	bridge, exists := e.bridges[name]
	if !exists {
		return fmt.Errorf("bridge %s not found", name)
	}

	// Cleanup bridge
	if bridge.IsInitialized() {
		ctx := context.Background()
		if err := bridge.Cleanup(ctx); err != nil {
			return fmt.Errorf("failed to cleanup bridge %s: %w", name, err)
		}
	}

	delete(e.bridges, name)
	return nil
}

// GetBridge retrieves a bridge by name
func (e *LuaEngine) GetBridge(name string) (engine.Bridge, error) {
	e.bridgesMu.RLock()
	defer e.bridgesMu.RUnlock()

	bridge, exists := e.bridges[name]
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", name)
	}

	return bridge, nil
}

// ListBridges returns a list of registered bridge names
func (e *LuaEngine) ListBridges() []string {
	e.bridgesMu.RLock()
	defer e.bridgesMu.RUnlock()

	names := make([]string, 0, len(e.bridges))
	for name := range e.bridges {
		names = append(names, name)
	}
	return names
}

// ToNative converts a Lua value to a Go value
func (e *LuaEngine) ToNative(scriptValue interface{}) (interface{}, error) {
	if luaValue, ok := scriptValue.(lua.LValue); ok {
		return e.converter.FromLua(luaValue)
	}
	return scriptValue, nil
}

// FromNative converts a Go value to a Lua value
func (e *LuaEngine) FromNative(goValue interface{}) (interface{}, error) {
	// For standalone conversion, we need a temporary LState
	// This is not ideal for performance but required by the interface
	tempState, err := e.factory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary state: %w", err)
	}
	defer tempState.Close()

	return e.converter.ToLua(tempState, goValue)
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

// executionResult holds the result of script execution
type executionResult struct {
	value interface{}
	error error
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
	return nil, fmt.Errorf("ExecuteScript not implemented yet")
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
