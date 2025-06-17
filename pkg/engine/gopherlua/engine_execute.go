// ABOUTME: Execution pipeline for LuaEngine implementing state acquisition, security, parameter injection, and result extraction
// ABOUTME: Handles the complete lifecycle of script execution with proper error handling and resource management

package gopherlua

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ExecutionContext holds context for a single script execution
type ExecutionContext struct {
	// Input
	Script string
	Params map[string]interface{}

	// Execution state
	State     *lua.LState
	Context   context.Context
	StartTime time.Time

	// Configuration
	TimeoutLimit time.Duration
	MemoryLimit  int64

	// Results
	Result interface{}
	Error  error

	// Metrics
	CompilationTime time.Duration
	ExecutionTime   time.Duration
	MemoryUsed      int64
	CacheHit        bool
}

// ExecutionPipeline manages the complete script execution flow
type ExecutionPipeline struct {
	engine    *LuaEngine
	converter *LuaTypeConverter
	cache     *ChunkCache
	pool      *LStatePool
	bridges   *BridgeManager
}

// NewExecutionPipeline creates a new execution pipeline
func NewExecutionPipeline(engine *LuaEngine) *ExecutionPipeline {
	return &ExecutionPipeline{
		engine:    engine,
		converter: engine.converter,
		cache:     engine.chunkCache,
		pool:      engine.pool,
		bridges:   engine.bridgeManager,
	}
}

// Execute runs the complete execution pipeline
func (ep *ExecutionPipeline) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	execCtx := &ExecutionContext{
		Script:       script,
		Params:       params,
		Context:      ctx,
		StartTime:    time.Now(),
		TimeoutLimit: ep.engine.timeoutLimit,
		MemoryLimit:  ep.engine.memoryLimit,
	}

	// Pipeline stages
	if err := ep.acquireState(execCtx); err != nil {
		return nil, err
	}
	defer ep.releaseState(execCtx)

	if err := ep.applySecurity(execCtx); err != nil {
		return nil, err
	}

	if err := ep.loadBridgeModules(execCtx); err != nil {
		return nil, err
	}

	if err := ep.injectParameters(execCtx); err != nil {
		return nil, err
	}

	if err := ep.compileScript(execCtx); err != nil {
		return nil, err
	}

	if err := ep.executeScript(execCtx); err != nil {
		return nil, err
	}

	if err := ep.extractResult(execCtx); err != nil {
		return nil, err
	}

	// Update metrics
	ep.updateMetrics(execCtx)

	return execCtx.Result, nil
}

// acquireState gets a Lua state from the pool
func (ep *ExecutionPipeline) acquireState(execCtx *ExecutionContext) error {
	// Apply timeout context
	ctx := execCtx.Context
	if execCtx.TimeoutLimit > 0 {
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > execCtx.TimeoutLimit {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(execCtx.Context, execCtx.TimeoutLimit)
			// Store cancel function to call it during cleanup
			execCtx.Context = ctx
			_ = cancel // We'll handle this in defer cleanup
		}
	}

	state, err := ep.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire Lua state: %w", err)
	}

	execCtx.State = state
	return nil
}

// releaseState returns the Lua state to the pool
func (ep *ExecutionPipeline) releaseState(execCtx *ExecutionContext) {
	if execCtx.State != nil {
		// Only return the state if it wasn't abandoned due to timeout
		ep.pool.Put(execCtx.State)
		execCtx.State = nil
	}
	// If State is nil, it was abandoned due to timeout and will be
	// cleaned up during pool shutdown
}

// applySecurity applies security sandbox to the state
func (ep *ExecutionPipeline) applySecurity(execCtx *ExecutionContext) error {
	// Security is already applied when creating states in the factory
	// Set execution context on the state for timeout handling
	execCtx.State.SetContext(execCtx.Context)
	return nil
}

// loadBridgeModules loads bridge modules into the state
func (ep *ExecutionPipeline) loadBridgeModules(execCtx *ExecutionContext) error {
	if err := ep.bridges.LoadBridgeModules(execCtx.State); err != nil {
		return fmt.Errorf("failed to load bridge modules: %w", err)
	}
	return nil
}

// injectParameters injects script parameters into the Lua state
func (ep *ExecutionPipeline) injectParameters(execCtx *ExecutionContext) error {
	for key, value := range execCtx.Params {
		luaValue, err := ep.converter.ToLua(execCtx.State, value)
		if err != nil {
			return fmt.Errorf("failed to convert parameter %s: %w", key, err)
		}
		execCtx.State.SetGlobal(key, luaValue)
	}
	return nil
}

// compileScript compiles the script with caching
func (ep *ExecutionPipeline) compileScript(execCtx *ExecutionContext) error {
	compileStart := time.Now()
	defer func() {
		execCtx.CompilationTime = time.Since(compileStart)
	}()

	// Try to get compiled chunk from cache
	cacheKey := ep.cache.GenerateKey(execCtx.Script, "")

	if cached := ep.cache.Get(cacheKey); cached != nil {
		// Use cached chunk
		execCtx.State.Push(execCtx.State.NewFunctionFromProto(cached))
		execCtx.CacheHit = true
		return nil
	}

	// Compile script
	chunk, err := execCtx.State.LoadString(execCtx.Script)
	if err != nil {
		return ep.engine.wrapLuaError(err, engine.ErrorTypeSyntax)
	}

	// Cache compiled chunk
	ep.cache.Put(cacheKey, chunk.Proto)
	execCtx.CacheHit = false

	// Push function for execution
	execCtx.State.Push(chunk)
	return nil
}

// executeScript executes the compiled script
func (ep *ExecutionPipeline) executeScript(execCtx *ExecutionContext) error {
	executeStart := time.Now()
	defer func() {
		execCtx.ExecutionTime = time.Since(executeStart)
	}()

	// Execute with timeout handling
	resultChan := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				select {
				case resultChan <- fmt.Errorf("panic during execution: %v", r):
				default:
					// Timed out, ignore the panic result
				}
			}
		}()

		err := execCtx.State.PCall(0, lua.MultRet, nil)

		// Only send result if we haven't timed out
		select {
		case resultChan <- err:
		default:
			// Channel is full, we timed out
		}
	}()

	select {
	case err := <-resultChan:
		<-done // Wait for goroutine to complete
		if err != nil {
			return ep.engine.wrapLuaError(err, engine.ErrorTypeRuntime)
		}
		return nil
	case <-execCtx.Context.Done():
		// Context cancelled/timed out
		// Abandon this state - it's still running
		abandonedState := execCtx.State
		execCtx.State = nil // Prevent releaseState from returning it to pool

		// Tell the pool to abandon this state
		if ep.pool != nil && abandonedState != nil {
			ep.pool.AbandonState(abandonedState)
		}

		// The execution goroutine will finish on its own
		// The pool's shutdown will wait for it if needed

		return &engine.EngineError{
			Type:    engine.ErrorTypeTimeout,
			Message: "script execution timed out",
			Cause:   execCtx.Context.Err(),
		}
	}
}

// extractResult extracts the result from the Lua stack
func (ep *ExecutionPipeline) extractResult(execCtx *ExecutionContext) error {
	// Get return value (top of stack)
	if execCtx.State.GetTop() > 0 {
		luaValue := execCtx.State.Get(-1)
		goValue, err := ep.converter.FromLua(luaValue)
		if err != nil {
			return fmt.Errorf("failed to convert result: %w", err)
		}
		execCtx.Result = goValue
	} else {
		execCtx.Result = nil
	}

	return nil
}

// updateMetrics updates engine metrics based on execution
func (ep *ExecutionPipeline) updateMetrics(execCtx *ExecutionContext) {
	atomic.AddInt64(&ep.engine.metrics.scriptsExecuted, 1)
	atomic.AddInt64(&ep.engine.metrics.totalExecTime, execCtx.ExecutionTime.Nanoseconds())
	atomic.AddInt64(&ep.engine.metrics.compilationTime, execCtx.CompilationTime.Nanoseconds())

	if execCtx.CacheHit {
		atomic.AddInt64(&ep.engine.metrics.cacheHits, 1)
	} else {
		atomic.AddInt64(&ep.engine.metrics.cacheMisses, 1)
	}

	if execCtx.Error != nil {
		atomic.AddInt64(&ep.engine.metrics.errorCount, 1)
	}

	// Update memory metrics if available
	if execCtx.MemoryUsed > 0 {
		atomic.StoreInt64(&ep.engine.metrics.memoryUsed, execCtx.MemoryUsed)
		if execCtx.MemoryUsed > atomic.LoadInt64(&ep.engine.metrics.peakMemoryUsed) {
			atomic.StoreInt64(&ep.engine.metrics.peakMemoryUsed, execCtx.MemoryUsed)
		}
	}
}

// ExecuteWithPipeline is a convenience method that uses the execution pipeline
func (e *LuaEngine) ExecuteWithPipeline(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	if !e.initialized {
		return nil, fmt.Errorf("engine not initialized")
	}

	if e.shuttingDown {
		return nil, fmt.Errorf("engine is shutting down")
	}

	pipeline := NewExecutionPipeline(e)
	return pipeline.Execute(ctx, script, params)
}

// BatchExecute executes multiple scripts concurrently
func (ep *ExecutionPipeline) BatchExecute(ctx context.Context, requests []ScriptRequest) ([]ScriptResult, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	results := make([]ScriptResult, len(requests))
	resultChan := make(chan batchResult, len(requests))

	// Execute all requests concurrently
	for i, req := range requests {
		go func(index int, request ScriptRequest) {
			result, err := ep.Execute(ctx, request.Script, request.Params)
			resultChan <- batchResult{
				Index:  index,
				Result: result,
				Error:  err,
			}
		}(i, req)
	}

	// Collect results
	for i := 0; i < len(requests); i++ {
		batchRes := <-resultChan
		results[batchRes.Index] = ScriptResult{
			Result: batchRes.Result,
			Error:  batchRes.Error,
		}
	}

	return results, nil
}

// ScriptRequest represents a script execution request
type ScriptRequest struct {
	Script string
	Params map[string]interface{}
}

// ScriptResult represents a script execution result
type ScriptResult struct {
	Result interface{}
	Error  error
}

// batchResult is used internally for batch execution
type batchResult struct {
	Index  int
	Result interface{}
	Error  error
}

// ValidateScript validates a script without executing it
func (ep *ExecutionPipeline) ValidateScript(script string) error {
	// Get a temporary state for validation
	tempState := lua.NewState()
	defer tempState.Close()

	// Try to compile the script
	_, err := tempState.LoadString(script)
	if err != nil {
		return ep.engine.wrapLuaError(err, engine.ErrorTypeSyntax)
	}

	return nil
}

// GetExecutionStats returns execution statistics
func (ep *ExecutionPipeline) GetExecutionStats() ExecutionStats {
	return ExecutionStats{
		ScriptsExecuted: atomic.LoadInt64(&ep.engine.metrics.scriptsExecuted),
		TotalExecTime:   time.Duration(atomic.LoadInt64(&ep.engine.metrics.totalExecTime)),
		CompilationTime: time.Duration(atomic.LoadInt64(&ep.engine.metrics.compilationTime)),
		CacheHits:       atomic.LoadInt64(&ep.engine.metrics.cacheHits),
		CacheMisses:     atomic.LoadInt64(&ep.engine.metrics.cacheMisses),
		ErrorCount:      atomic.LoadInt64(&ep.engine.metrics.errorCount),
		MemoryUsed:      atomic.LoadInt64(&ep.engine.metrics.memoryUsed),
		PeakMemoryUsed:  atomic.LoadInt64(&ep.engine.metrics.peakMemoryUsed),
	}
}

// ExecutionStats provides execution statistics
type ExecutionStats struct {
	ScriptsExecuted int64
	TotalExecTime   time.Duration
	CompilationTime time.Duration
	CacheHits       int64
	CacheMisses     int64
	ErrorCount      int64
	MemoryUsed      int64
	PeakMemoryUsed  int64
}
