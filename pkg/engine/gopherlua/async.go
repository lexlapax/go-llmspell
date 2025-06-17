// ABOUTME: Async runtime for coroutine management in GopherLua engine
// ABOUTME: Provides promise-coroutine integration, async execution contexts, and cancellation support

package gopherlua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

// AsyncRuntime manages coroutines and async operations in the Lua engine
type AsyncRuntime struct {
	maxCoroutines  int
	activeRoutines map[string]*coroutineInfo
	routineResults map[string]*coroutineResult
	mu             sync.RWMutex
	closed         bool
	closeOnce      sync.Once
}

// coroutineInfo tracks active coroutine state
type coroutineInfo struct {
	ID        string
	StartTime time.Time
	Context   context.Context
	Cancel    context.CancelFunc
	Done      chan struct{}
}

// coroutineResult stores completed coroutine results
type coroutineResult struct {
	Value lua.LValue
	Error error
}

// Promise represents an async operation backed by a coroutine
type Promise struct {
	coroID  string
	runtime *AsyncRuntime
}

// AsyncExecutionContext provides context for async operations
type AsyncExecutionContext struct {
	ID        string
	StartTime time.Time
	Context   context.Context
	LState    *lua.LState
}

const (
	defaultMaxCoroutines = 100
)

// NewAsyncRuntime creates a new async runtime with specified max coroutines
func NewAsyncRuntime(maxCoroutines int) (*AsyncRuntime, error) {
	if maxCoroutines < 0 {
		return nil, fmt.Errorf("maxCoroutines cannot be negative: %d", maxCoroutines)
	}

	if maxCoroutines == 0 {
		maxCoroutines = defaultMaxCoroutines
	}

	return &AsyncRuntime{
		maxCoroutines:  maxCoroutines,
		activeRoutines: make(map[string]*coroutineInfo),
		routineResults: make(map[string]*coroutineResult),
	}, nil
}

// SpawnCoroutine creates and starts a new coroutine
func (ar *AsyncRuntime) SpawnCoroutine(ctx context.Context, L *lua.LState, script string, args ...lua.LValue) (string, error) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if ar.closed {
		return "", fmt.Errorf("async runtime is closed")
	}

	if len(ar.activeRoutines) >= ar.maxCoroutines {
		return "", fmt.Errorf("maximum coroutines (%d) exceeded", ar.maxCoroutines)
	}

	coroID := uuid.New().String()
	coroCtx, cancel := context.WithCancel(ctx)

	info := &coroutineInfo{
		ID:        coroID,
		StartTime: time.Now(),
		Context:   coroCtx,
		Cancel:    cancel,
		Done:      make(chan struct{}),
	}

	ar.activeRoutines[coroID] = info

	// Start the coroutine in a goroutine
	go func() {
		defer func() {
			close(info.Done)
			cancel()

			ar.mu.Lock()
			delete(ar.activeRoutines, coroID)
			ar.mu.Unlock()
		}()

		// Create new LState for this coroutine (thread safety)
		coroL := lua.NewState()
		defer coroL.Close()

		// Push arguments onto stack
		for _, arg := range args {
			coroL.Push(arg)
		}

		var result *coroutineResult

		// Execute script directly with context monitoring
		select {
		case <-coroCtx.Done():
			result = &coroutineResult{
				Value: lua.LNil,
				Error: coroCtx.Err(),
			}
		default:
			err := coroL.DoString(script)
			value := lua.LNil

			if err == nil && coroL.GetTop() > 0 {
				value = coroL.Get(-1)
			}

			result = &coroutineResult{
				Value: value,
				Error: err,
			}
		}

		// Store result
		ar.mu.Lock()
		ar.routineResults[coroID] = result
		ar.mu.Unlock()
	}()

	return coroID, nil
}

// WaitForCoroutine waits for a coroutine to complete and returns its result
func (ar *AsyncRuntime) WaitForCoroutine(ctx context.Context, coroID string) (lua.LValue, error) {
	ar.mu.RLock()
	info, exists := ar.activeRoutines[coroID]
	result, hasResult := ar.routineResults[coroID]
	ar.mu.RUnlock()

	// If already completed, return result
	if hasResult {
		return result.Value, result.Error
	}

	if !exists {
		return lua.LNil, fmt.Errorf("coroutine not found: %s", coroID)
	}

	// Wait for completion or context cancellation
	select {
	case <-info.Done:
		ar.mu.RLock()
		result := ar.routineResults[coroID]
		ar.mu.RUnlock()

		if result == nil {
			return lua.LNil, fmt.Errorf("coroutine completed but no result found")
		}

		return result.Value, result.Error
	case <-ctx.Done():
		// Cancel the coroutine
		info.Cancel()
		return lua.LNil, ctx.Err()
	}
}

// CreatePromise creates a promise backed by a coroutine
func (ar *AsyncRuntime) CreatePromise(ctx context.Context, L *lua.LState, script string, args ...lua.LValue) (*Promise, error) {
	coroID, err := ar.SpawnCoroutine(ctx, L, script, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn coroutine for promise: %w", err)
	}

	return &Promise{
		coroID:  coroID,
		runtime: ar,
	}, nil
}

// CreateExecutionContext creates an async execution context
func (ar *AsyncRuntime) CreateExecutionContext(ctx context.Context, L *lua.LState) (*AsyncExecutionContext, error) {
	if ar.closed {
		return nil, fmt.Errorf("async runtime is closed")
	}

	return &AsyncExecutionContext{
		ID:        uuid.New().String(),
		StartTime: time.Now(),
		Context:   ctx,
		LState:    L,
	}, nil
}

// Close shuts down the async runtime and cancels all active coroutines
func (ar *AsyncRuntime) Close() error {
	var closeErr error

	ar.closeOnce.Do(func() {
		ar.mu.Lock()
		defer ar.mu.Unlock()

		ar.closed = true

		// Cancel all active coroutines
		for _, info := range ar.activeRoutines {
			info.Cancel()
		}

		// Clear maps
		ar.activeRoutines = make(map[string]*coroutineInfo)
		ar.routineResults = make(map[string]*coroutineResult)
	})

	return closeErr
}

// GetActiveCoroutineCount returns the number of currently active coroutines
func (ar *AsyncRuntime) GetActiveCoroutineCount() int {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return len(ar.activeRoutines)
}

// IsCoroutineActive checks if a coroutine is still running
func (ar *AsyncRuntime) IsCoroutineActive(coroID string) bool {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	_, exists := ar.activeRoutines[coroID]
	return exists
}

// Promise methods

// Await waits for the promise to resolve
func (p *Promise) Await(ctx context.Context) (lua.LValue, error) {
	return p.runtime.WaitForCoroutine(ctx, p.coroID)
}

// IsResolved checks if the promise has completed
func (p *Promise) IsResolved() bool {
	p.runtime.mu.RLock()
	defer p.runtime.mu.RUnlock()
	_, hasResult := p.runtime.routineResults[p.coroID]
	return hasResult
}

// Cancel cancels the promise's underlying coroutine
func (p *Promise) Cancel() {
	p.runtime.mu.RLock()
	info, exists := p.runtime.activeRoutines[p.coroID]
	p.runtime.mu.RUnlock()

	if exists {
		info.Cancel()
	}
}
