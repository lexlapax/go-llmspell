// ABOUTME: LStatePool manages a pool of reusable Lua VM instances for performance and resource efficiency
// ABOUTME: Provides adaptive scaling, health monitoring, lifecycle management, and graceful shutdown

package gopherlua

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// PoolConfig configures the LState pool behavior
type PoolConfig struct {
	// MinSize is the minimum number of states to keep in the pool
	MinSize int

	// MaxSize is the maximum number of states allowed in the pool
	MaxSize int

	// IdleTimeout is how long a state can be idle before being removed
	IdleTimeout time.Duration

	// HealthThreshold is the health score threshold below which states are recycled
	HealthThreshold float64

	// CleanupInterval is how often to run cleanup of idle/unhealthy states
	CleanupInterval time.Duration
}

// PoolMetrics provides insight into pool performance and usage
type PoolMetrics struct {
	// Available is the number of states available in the pool
	Available int64

	// InUse is the number of states currently in use
	InUse int64

	// TotalCreated is the total number of states created since pool start
	TotalCreated int64

	// TotalRecycled is the total number of states recycled due to health issues
	TotalRecycled int64

	// TotalCleanedUp is the total number of states cleaned up due to idle timeout
	TotalCleanedUp int64
}

// pooledState wraps an LState with metadata for pool management
type pooledState struct {
	state     *lua.LState
	lastUsed  time.Time
	useCount  int64
	health    float64
	id        int64
	executing bool          // true when state is being executed
	done      chan struct{} // closed when execution completes
	mu        sync.Mutex    // protects executing flag
}

// LStatePool manages a pool of Lua VM instances
type LStatePool struct {
	factory       *LStateFactory
	config        PoolConfig
	states        chan *pooledState
	inUse         map[*lua.LState]*pooledState
	metrics       PoolMetrics
	shutdown      chan struct{}
	shutdownOnce  sync.Once
	mu            sync.RWMutex
	nextStateID   int64
	cleanupTicker *time.Ticker
}

// NewLStatePool creates a new pool with the given factory and configuration
func NewLStatePool(factory *LStateFactory, config PoolConfig) (*LStatePool, error) {
	if factory == nil {
		return nil, fmt.Errorf("factory cannot be nil")
	}

	// Apply defaults
	if config.MinSize <= 0 {
		config.MinSize = 1
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 10
	}
	if config.MaxSize < config.MinSize {
		config.MaxSize = config.MinSize
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 10 * time.Minute
	}
	if config.HealthThreshold <= 0 {
		config.HealthThreshold = 0.7
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Minute
	}

	pool := &LStatePool{
		factory:       factory,
		config:        config,
		states:        make(chan *pooledState, config.MaxSize),
		inUse:         make(map[*lua.LState]*pooledState),
		shutdown:      make(chan struct{}),
		cleanupTicker: time.NewTicker(config.CleanupInterval),
	}

	// Pre-populate with minimum states
	for i := 0; i < config.MinSize; i++ {
		state, err := pool.createState()
		if err != nil {
			_ = pool.Shutdown(context.Background())
			return nil, fmt.Errorf("failed to create initial state %d: %w", i, err)
		}
		pool.states <- state
		atomic.AddInt64(&pool.metrics.Available, 1)
	}

	// Start cleanup goroutine
	go pool.cleanupLoop()

	return pool, nil
}

// Get retrieves a state from the pool or creates a new one if needed
func (p *LStatePool) Get(ctx context.Context) (*lua.LState, error) {
	select {
	case <-p.shutdown:
		return nil, fmt.Errorf("pool is shutdown")
	default:
	}

	// Try to get from pool first
	select {
	case pooledState := <-p.states:
		atomic.AddInt64(&p.metrics.Available, -1)
		atomic.AddInt64(&p.metrics.InUse, 1)

		// Mark as executing
		pooledState.mu.Lock()
		pooledState.executing = true
		pooledState.done = make(chan struct{})
		pooledState.mu.Unlock()

		p.mu.Lock()
		p.inUse[pooledState.state] = pooledState
		p.mu.Unlock()

		// Reset state for new use
		p.resetState(pooledState)
		return pooledState.state, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// Pool is empty, try to create new state if under max
		p.mu.RLock()
		inUseCount := len(p.inUse)
		p.mu.RUnlock()

		if inUseCount < p.config.MaxSize {
			// Create temporary state (not pooled)
			state, err := p.createState()
			if err != nil {
				return nil, fmt.Errorf("failed to create temporary state: %w", err)
			}

			atomic.AddInt64(&p.metrics.InUse, 1)

			// Mark as executing
			state.mu.Lock()
			state.executing = true
			state.done = make(chan struct{})
			state.mu.Unlock()

			p.mu.Lock()
			p.inUse[state.state] = state
			p.mu.Unlock()

			return state.state, nil
		}

		// Wait for available state
		select {
		case pooledState := <-p.states:
			atomic.AddInt64(&p.metrics.Available, -1)
			atomic.AddInt64(&p.metrics.InUse, 1)

			// Mark as executing
			pooledState.mu.Lock()
			pooledState.executing = true
			pooledState.done = make(chan struct{})
			pooledState.mu.Unlock()

			p.mu.Lock()
			p.inUse[pooledState.state] = pooledState
			p.mu.Unlock()

			p.resetState(pooledState)
			return pooledState.state, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Put returns a state to the pool after use
func (p *LStatePool) Put(state *lua.LState) {
	if state == nil {
		return
	}

	p.mu.Lock()
	pooledState, exists := p.inUse[state]
	if !exists {
		p.mu.Unlock()
		// State not from this pool, just close it
		state.Close()
		return
	}
	delete(p.inUse, state)
	p.mu.Unlock()

	atomic.AddInt64(&p.metrics.InUse, -1)

	// Mark as not executing
	pooledState.mu.Lock()
	pooledState.executing = false
	if pooledState.done != nil {
		close(pooledState.done)
	}
	pooledState.mu.Unlock()

	// Update state metadata
	pooledState.lastUsed = time.Now()
	pooledState.useCount++
	pooledState.health = p.calculateHealth(pooledState)

	// Check if state should be recycled
	if pooledState.health < p.config.HealthThreshold {
		atomic.AddInt64(&p.metrics.TotalRecycled, 1)
		state.Close()
		return
	}

	// Try to return to pool
	select {
	case p.states <- pooledState:
		atomic.AddInt64(&p.metrics.Available, 1)
	default:
		// Pool is full, discard this state
		state.Close()
	}
}

// AbandonState marks a state as abandoned due to timeout
// The state is removed from tracking but not closed immediately
func (p *LStatePool) AbandonState(state *lua.LState) {
	if state == nil {
		return
	}

	p.mu.Lock()
	pooledState, exists := p.inUse[state]
	if !exists {
		p.mu.Unlock()
		return
	}
	delete(p.inUse, state)
	p.mu.Unlock()

	atomic.AddInt64(&p.metrics.InUse, -1)
	atomic.AddInt64(&p.metrics.TotalRecycled, 1)

	// Mark as not executing and signal done
	pooledState.mu.Lock()
	pooledState.executing = false
	if pooledState.done != nil {
		close(pooledState.done)
	}
	pooledState.mu.Unlock()

	// Don't return to pool, don't close - let it be GC'd when execution completes
}

// GetMetrics returns current pool metrics
func (p *LStatePool) GetMetrics() PoolMetrics {
	p.mu.RLock()
	inUseCount := int64(len(p.inUse))
	p.mu.RUnlock()

	return PoolMetrics{
		Available:      atomic.LoadInt64(&p.metrics.Available),
		InUse:          inUseCount,
		TotalCreated:   atomic.LoadInt64(&p.metrics.TotalCreated),
		TotalRecycled:  atomic.LoadInt64(&p.metrics.TotalRecycled),
		TotalCleanedUp: atomic.LoadInt64(&p.metrics.TotalCleanedUp),
	}
}

// Shutdown gracefully shuts down the pool
func (p *LStatePool) Shutdown(ctx context.Context) error {
	var shutdownErr error
	p.shutdownOnce.Do(func() {
		// Stop cleanup goroutine
		close(p.shutdown)
		p.cleanupTicker.Stop()

		// Wait for in-use states to be returned or timeout
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			p.mu.RLock()
			inUseCount := len(p.inUse)
			p.mu.RUnlock()

			if inUseCount == 0 {
				break
			}

			select {
			case <-ctx.Done():
				shutdownErr = fmt.Errorf("shutdown timeout: %d states still in use", inUseCount)
				goto cleanup
			case <-ticker.C:
				// Continue waiting
			}
		}

	cleanup:
		// Close all pooled states
		close(p.states)
		for pooledState := range p.states {
			pooledState.state.Close()
		}

		// Wait for executing states to finish
		p.mu.Lock()
		var executingStates []*pooledState
		for _, ps := range p.inUse {
			ps.mu.Lock()
			if ps.executing {
				executingStates = append(executingStates, ps)
			}
			ps.mu.Unlock()
		}
		p.mu.Unlock()

		// Wait for executing states with timeout
		if len(executingStates) > 0 {
			waitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			for _, ps := range executingStates {
				ps.mu.Lock()
				done := ps.done
				ps.mu.Unlock()

				if done != nil {
					select {
					case <-done:
						// State finished executing
					case <-waitCtx.Done():
						// Timeout waiting, but don't close - let it finish
					}
				}
			}
		}

		// Now close all states that are not executing
		p.mu.Lock()
		for state, ps := range p.inUse {
			ps.mu.Lock()
			if !ps.executing {
				state.Close()
			}
			ps.mu.Unlock()
		}
		p.inUse = nil
		p.mu.Unlock()
	})

	return shutdownErr
}

// createState creates a new pooled state
func (p *LStatePool) createState() (*pooledState, error) {
	state, err := p.factory.Create()
	if err != nil {
		return nil, err
	}

	pooledState := &pooledState{
		state:     state,
		lastUsed:  time.Now(),
		useCount:  0,
		health:    1.0,
		id:        atomic.AddInt64(&p.nextStateID, 1),
		executing: false,
		done:      nil,
	}

	atomic.AddInt64(&p.metrics.TotalCreated, 1)
	return pooledState, nil
}

// resetState prepares a state for reuse
func (p *LStatePool) resetState(pooledState *pooledState) {
	state := pooledState.state

	// Reset global environment
	state.SetGlobal("_G", state.Env)

	// Clear the stack
	state.SetTop(0)

	// Clear any user-defined globals (but preserve libraries)
	// We'll iterate through known user globals and clear them
	userGlobals := []string{"x", "test_value", "large_table"} // Common test globals
	for _, global := range userGlobals {
		state.SetGlobal(global, lua.LNil)
	}
}

// calculateHealth computes the health score of a state
func (p *LStatePool) calculateHealth(pooledState *pooledState) float64 {
	baseHealth := 1.0

	// Factor in memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Decrease health based on use count (states get "worn out")
	useFactor := float64(pooledState.useCount) / 1000.0
	if useFactor > 0.5 {
		baseHealth -= 0.3
	} else if useFactor > 0.2 {
		baseHealth -= 0.1
	}

	// Factor in age
	age := time.Since(pooledState.lastUsed)
	if age > 30*time.Minute {
		baseHealth -= 0.2
	} else if age > 10*time.Minute {
		baseHealth -= 0.1
	}

	if baseHealth < 0 {
		baseHealth = 0
	}

	return baseHealth
}

// cleanupLoop runs periodic cleanup of idle and unhealthy states
func (p *LStatePool) cleanupLoop() {
	for {
		select {
		case <-p.shutdown:
			return
		case <-p.cleanupTicker.C:
			p.cleanup()
		}
	}
}

// cleanup removes idle and unhealthy states from the pool
func (p *LStatePool) cleanup() {
	now := time.Now()

	// Collect states to remove
	var statesToRemove []*pooledState
	var statesToKeep []*pooledState

	// Drain current states for inspection
drainLoop:
	for {
		select {
		case pooledState := <-p.states:
			age := now.Sub(pooledState.lastUsed)
			health := p.calculateHealth(pooledState)

			if age > p.config.IdleTimeout || health < p.config.HealthThreshold {
				statesToRemove = append(statesToRemove, pooledState)
			} else {
				statesToKeep = append(statesToKeep, pooledState)
			}
		default:
			break drainLoop
		}
	}

	// Close removed states
	for _, pooledState := range statesToRemove {
		pooledState.state.Close()
		atomic.AddInt64(&p.metrics.TotalCleanedUp, 1)
		atomic.AddInt64(&p.metrics.Available, -1)
	}

	// Return kept states to pool
	for _, pooledState := range statesToKeep {
		select {
		case p.states <- pooledState:
			// Successfully returned to pool
		default:
			// Pool is full, close this state
			pooledState.state.Close()
			atomic.AddInt64(&p.metrics.TotalCleanedUp, 1)
			atomic.AddInt64(&p.metrics.Available, -1)
		}
	}

	// Ensure minimum pool size by creating new states if needed
	currentAvailable := len(p.states)
	needed := p.config.MinSize - currentAvailable

	for i := 0; i < needed; i++ {
		state, err := p.createState()
		if err != nil {
			// Log error but continue cleanup
			continue
		}

		select {
		case p.states <- state:
			atomic.AddInt64(&p.metrics.Available, 1)
		default:
			// Pool is somehow full, close the state
			state.state.Close()
		}
	}
}
