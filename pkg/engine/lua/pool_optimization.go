// ABOUTME: Optimized state pool with predictive scaling, pre-warming, memory pooling, and adaptive configuration
// ABOUTME: Provides intelligent pool management with machine learning-inspired prediction and resource optimization

package gopherlua

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// OptimizedPoolConfig extends PoolConfig with optimization features
type OptimizedPoolConfig struct {
	PoolConfig

	// Predictive scaling
	EnablePredictiveScaling bool
	PredictionWindowSize    int           // Number of recent periods to analyze
	PredictionInterval      time.Duration // How often to predict
	ScaleUpThreshold        float64       // Usage percentage to trigger scale up
	ScaleDownThreshold      float64       // Usage percentage to trigger scale down
	MaxPredictedScaleUp     int           // Max states to add in one prediction cycle

	// Pre-warming
	EnablePreWarming bool
	PreWarmOnInit    int           // Number of states to pre-warm on init
	PreWarmScript    string        // Optional script to run during pre-warm
	PreWarmTimeout   time.Duration // Timeout for pre-warm operations

	// Memory pooling
	EnableMemoryPooling bool
	MemoryPoolSize      int // Number of memory blocks to pool
	MemoryBlockSize     int // Size of each memory block

	// Advanced features
	EnableAdaptiveThresholds bool // Dynamically adjust thresholds
	EnableLoadBalancing      bool // Balance load across states
	StatePriority            bool // Prioritize healthier states
}

// UsagePattern tracks usage patterns for prediction
type UsagePattern struct {
	timestamp   time.Time
	inUse       int
	available   int
	requestRate float64
}

// StateLoadInfo tracks load information for a state
type StateLoadInfo struct {
	executionCount int64
	totalDuration  time.Duration
	lastExecution  time.Time
	avgDuration    time.Duration
}

// OptimizedLStatePool extends LStatePool with optimization features
type OptimizedLStatePool struct {
	*LStatePool

	// Configuration
	config OptimizedPoolConfig

	// Predictive scaling
	usageHistory     []UsagePattern
	historyMu        sync.RWMutex
	predictionTicker *time.Ticker

	// Pre-warming
	preWarmChan chan struct{}
	preWarmWg   sync.WaitGroup

	// Memory pooling
	memoryPool      *sync.Pool
	memoryBlockPool [][]byte
	memoryMu        sync.Mutex

	// Load tracking
	stateLoad   map[int64]*StateLoadInfo
	stateLoadMu sync.RWMutex

	// Metrics
	predictedScaleUps   int64
	predictedScaleDowns int64
	preWarmedStates     int64
	memoryPoolHits      int64
	memoryPoolMisses    int64

	// Request tracking for prediction
	requestTimes   []time.Time
	requestTimesMu sync.Mutex
}

// NewOptimizedLStatePool creates an optimized state pool
func NewOptimizedLStatePool(factory *LStateFactory, config OptimizedPoolConfig) (*OptimizedLStatePool, error) {
	// Create base pool
	basePool, err := NewLStatePool(factory, config.PoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create base pool: %w", err)
	}

	// Apply optimization defaults
	if config.PredictionWindowSize <= 0 {
		config.PredictionWindowSize = 10
	}
	if config.PredictionInterval <= 0 {
		config.PredictionInterval = 30 * time.Second
	}
	if config.ScaleUpThreshold <= 0 {
		config.ScaleUpThreshold = 0.8
	}
	if config.ScaleDownThreshold <= 0 {
		config.ScaleDownThreshold = 0.2
	}
	if config.MaxPredictedScaleUp <= 0 {
		config.MaxPredictedScaleUp = 5
	}
	if config.PreWarmTimeout <= 0 {
		config.PreWarmTimeout = 5 * time.Second
	}
	if config.MemoryBlockSize <= 0 {
		config.MemoryBlockSize = 1024 * 1024 // 1MB
	}
	if config.MemoryPoolSize <= 0 {
		config.MemoryPoolSize = 10
	}

	pool := &OptimizedLStatePool{
		LStatePool:   basePool,
		config:       config,
		usageHistory: make([]UsagePattern, 0, config.PredictionWindowSize),
		stateLoad:    make(map[int64]*StateLoadInfo),
		preWarmChan:  make(chan struct{}, config.MaxSize),
		requestTimes: make([]time.Time, 0, 1000),
	}

	// Initialize memory pool if enabled
	if config.EnableMemoryPooling {
		pool.initMemoryPool()
	}

	// Start predictive scaling if enabled
	if config.EnablePredictiveScaling {
		pool.predictionTicker = time.NewTicker(config.PredictionInterval)
		go pool.predictionLoop()
	}

	// Pre-warm states if configured
	if config.EnablePreWarming && config.PreWarmOnInit > 0 {
		pool.preWarmStates(config.PreWarmOnInit)
	}

	return pool, nil
}

// Get retrieves a state with load balancing
func (op *OptimizedLStatePool) Get(ctx context.Context) (*lua.LState, error) {
	// Track request for prediction
	op.trackRequest()

	// Use load balancing if enabled
	if op.config.EnableLoadBalancing {
		return op.getWithLoadBalancing(ctx)
	}

	// Otherwise use base implementation
	return op.LStatePool.Get(ctx)
}

// getWithLoadBalancing selects the least loaded state
func (op *OptimizedLStatePool) getWithLoadBalancing(ctx context.Context) (*lua.LState, error) {
	// Try to get multiple states and pick the least loaded
	candidates := make([]*pooledState, 0, 3)

	// Collect up to 3 candidates
	for i := 0; i < 3; i++ {
		select {
		case pooledState := <-op.states:
			candidates = append(candidates, pooledState)
		default:
			i = 3 // Exit the loop
		}
	}

	if len(candidates) == 0 {
		// No states available, use base Get
		return op.LStatePool.Get(ctx)
	}

	// Find least loaded state
	var bestState *pooledState
	var lowestLoad int64 = math.MaxInt64

	op.stateLoadMu.RLock()
	for _, ps := range candidates {
		load, exists := op.stateLoad[ps.id]
		if !exists || load.executionCount < lowestLoad {
			if bestState != nil {
				// Return previous best to pool
				op.states <- bestState
			}
			bestState = ps
			if exists {
				lowestLoad = load.executionCount
			} else {
				lowestLoad = 0
			}
		} else {
			// Return to pool
			op.states <- ps
		}
	}
	op.stateLoadMu.RUnlock()

	// Update metrics and tracking
	atomic.AddInt64(&op.metrics.Available, -1)
	atomic.AddInt64(&op.metrics.InUse, 1)

	bestState.mu.Lock()
	bestState.executing = true
	bestState.done = make(chan struct{})
	bestState.mu.Unlock()

	op.mu.Lock()
	op.inUse[bestState.state] = bestState
	op.mu.Unlock()

	op.resetState(bestState)

	// Track execution start
	op.trackExecutionStart(bestState.id)

	return bestState.state, nil
}

// Put returns a state and updates load information
func (op *OptimizedLStatePool) Put(state *lua.LState) {
	// Get pooled state info
	op.mu.RLock()
	pooledState, exists := op.inUse[state]
	op.mu.RUnlock()

	if exists {
		// Track execution end
		op.trackExecutionEnd(pooledState.id)
	}

	// Use base implementation
	op.LStatePool.Put(state)
}

// initMemoryPool initializes the memory pool
func (op *OptimizedLStatePool) initMemoryPool() {
	op.memoryPool = &sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&op.memoryPoolMisses, 1)
			b := make([]byte, op.config.MemoryBlockSize)
			return &b
		},
	}

	// Pre-allocate memory blocks
	op.memoryBlockPool = make([][]byte, 0, op.config.MemoryPoolSize)
	for i := 0; i < op.config.MemoryPoolSize; i++ {
		block := make([]byte, op.config.MemoryBlockSize)
		op.memoryBlockPool = append(op.memoryBlockPool, block)
		op.memoryPool.Put(&block)
	}
}

// GetMemoryBlock gets a memory block from the pool
func (op *OptimizedLStatePool) GetMemoryBlock() []byte {
	if !op.config.EnableMemoryPooling {
		return make([]byte, op.config.MemoryBlockSize)
	}

	blockPtr := op.memoryPool.Get().(*[]byte)
	atomic.AddInt64(&op.memoryPoolHits, 1)
	return *blockPtr
}

// PutMemoryBlock returns a memory block to the pool
func (op *OptimizedLStatePool) PutMemoryBlock(block []byte) {
	if !op.config.EnableMemoryPooling || len(block) != op.config.MemoryBlockSize {
		return
	}

	// Clear the block before returning to pool
	for i := range block {
		block[i] = 0
	}

	op.memoryPool.Put(&block)
}

// preWarmStates pre-warms the specified number of states
func (op *OptimizedLStatePool) preWarmStates(count int) {
	op.preWarmWg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer op.preWarmWg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), op.config.PreWarmTimeout)
			defer cancel()

			state, err := op.LStatePool.Get(ctx)
			if err != nil {
				return
			}

			// Run pre-warm script if configured
			if op.config.PreWarmScript != "" {
				_ = state.DoString(op.config.PreWarmScript)
			}

			// Return to pool
			op.LStatePool.Put(state)
			atomic.AddInt64(&op.preWarmedStates, 1)
		}()
	}
}

// WaitForPreWarm waits for pre-warming to complete
func (op *OptimizedLStatePool) WaitForPreWarm() {
	op.preWarmWg.Wait()
}

// predictionLoop runs the predictive scaling loop
func (op *OptimizedLStatePool) predictionLoop() {
	for {
		select {
		case <-op.shutdown:
			op.predictionTicker.Stop()
			return
		case <-op.predictionTicker.C:
			op.performPrediction()
		}
	}
}

// performPrediction analyzes usage and scales the pool
func (op *OptimizedLStatePool) performPrediction() {
	metrics := op.GetMetrics()

	// Calculate current usage
	total := metrics.Available + metrics.InUse
	if total == 0 {
		return
	}

	usage := float64(metrics.InUse) / float64(total)

	// Calculate request rate
	requestRate := op.calculateRequestRate()

	// Record usage pattern
	pattern := UsagePattern{
		timestamp:   time.Now(),
		inUse:       int(metrics.InUse),
		available:   int(metrics.Available),
		requestRate: requestRate,
	}

	op.historyMu.Lock()
	op.usageHistory = append(op.usageHistory, pattern)
	if len(op.usageHistory) > op.config.PredictionWindowSize {
		op.usageHistory = op.usageHistory[1:]
	}
	op.historyMu.Unlock()

	// Predict future usage
	predictedUsage := op.predictUsage()

	// Scale based on prediction
	if predictedUsage > op.config.ScaleUpThreshold {
		op.scaleUp(predictedUsage)
	} else if predictedUsage < op.config.ScaleDownThreshold && usage < op.config.ScaleDownThreshold {
		op.scaleDown()
	}

	// Adaptive threshold adjustment
	if op.config.EnableAdaptiveThresholds {
		op.adjustThresholds()
	}
}

// predictUsage predicts future usage based on history
func (op *OptimizedLStatePool) predictUsage() float64 {
	op.historyMu.RLock()
	defer op.historyMu.RUnlock()

	if len(op.usageHistory) < 3 {
		// Not enough history
		return 0.5
	}

	// Simple linear regression for trend
	n := float64(len(op.usageHistory))
	var sumX, sumY, sumXY, sumX2 float64

	for i, pattern := range op.usageHistory {
		x := float64(i)
		y := float64(pattern.inUse) / float64(pattern.inUse+pattern.available)

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate intercept
	intercept := (sumY - slope*sumX) / n

	// Predict next value
	nextX := n
	predicted := slope*nextX + intercept

	// Consider request rate trend
	if len(op.usageHistory) >= 2 {
		recent := op.usageHistory[len(op.usageHistory)-1]
		previous := op.usageHistory[len(op.usageHistory)-2]

		if recent.requestRate > previous.requestRate*1.2 {
			// Request rate increasing significantly
			predicted *= 1.1
		}
	}

	// Clamp to [0, 1]
	if predicted < 0 {
		predicted = 0
	} else if predicted > 1 {
		predicted = 1
	}

	return predicted
}

// scaleUp adds more states to the pool
func (op *OptimizedLStatePool) scaleUp(predictedUsage float64) {
	metrics := op.GetMetrics()
	current := int(metrics.Available + metrics.InUse)

	// Calculate how many states to add
	targetTotal := int(float64(current) * (1 + (predictedUsage - op.config.ScaleUpThreshold)))
	toAdd := targetTotal - current

	if toAdd > op.config.MaxPredictedScaleUp {
		toAdd = op.config.MaxPredictedScaleUp
	}

	if toAdd <= 0 || current+toAdd > op.config.MaxSize {
		return
	}

	// Add states asynchronously
	for i := 0; i < toAdd; i++ {
		go func() {
			state, err := op.createState()
			if err != nil {
				return
			}

			select {
			case op.states <- state:
				atomic.AddInt64(&op.metrics.Available, 1)
				atomic.AddInt64(&op.predictedScaleUps, 1)
			default:
				// Pool is full
				state.state.Close()
			}
		}()
	}
}

// scaleDown removes excess states from the pool
func (op *OptimizedLStatePool) scaleDown() {
	metrics := op.GetMetrics()

	// Only scale down if we have more than minimum
	if metrics.Available <= int64(op.config.MinSize) {
		return
	}

	// Remove up to 20% of available states
	toRemove := int(float64(metrics.Available) * 0.2)
	if toRemove == 0 {
		toRemove = 1
	}

	removed := 0
	for i := 0; i < toRemove; i++ {
		select {
		case pooledState := <-op.states:
			pooledState.state.Close()
			atomic.AddInt64(&op.metrics.Available, -1)
			atomic.AddInt64(&op.predictedScaleDowns, 1)
			removed++
		default:
			// No more states available
			i = toRemove // Exit the loop
		}
	}
}

// adjustThresholds dynamically adjusts scaling thresholds
func (op *OptimizedLStatePool) adjustThresholds() {
	op.historyMu.RLock()
	defer op.historyMu.RUnlock()

	if len(op.usageHistory) < op.config.PredictionWindowSize {
		return
	}

	// Calculate average wait times and usage variance
	var avgUsage, usageVariance float64
	usages := make([]float64, len(op.usageHistory))

	for i, pattern := range op.usageHistory {
		usage := float64(pattern.inUse) / float64(pattern.inUse+pattern.available)
		usages[i] = usage
		avgUsage += usage
	}
	avgUsage /= float64(len(op.usageHistory))

	// Calculate variance
	for _, usage := range usages {
		diff := usage - avgUsage
		usageVariance += diff * diff
	}
	usageVariance /= float64(len(op.usageHistory))

	// Adjust thresholds based on variance
	if usageVariance > 0.1 {
		// High variance - be more conservative
		op.config.ScaleUpThreshold = math.Max(0.7, op.config.ScaleUpThreshold-0.05)
		op.config.ScaleDownThreshold = math.Min(0.3, op.config.ScaleDownThreshold+0.05)
	} else if usageVariance < 0.05 {
		// Low variance - can be more aggressive
		op.config.ScaleUpThreshold = math.Min(0.9, op.config.ScaleUpThreshold+0.05)
		op.config.ScaleDownThreshold = math.Max(0.1, op.config.ScaleDownThreshold-0.05)
	}
}

// trackRequest tracks a request for rate calculation
func (op *OptimizedLStatePool) trackRequest() {
	op.requestTimesMu.Lock()
	defer op.requestTimesMu.Unlock()

	now := time.Now()
	op.requestTimes = append(op.requestTimes, now)

	// Keep only recent requests (last minute)
	cutoff := now.Add(-time.Minute)
	i := 0
	for i < len(op.requestTimes) && op.requestTimes[i].Before(cutoff) {
		i++
	}
	op.requestTimes = op.requestTimes[i:]
}

// calculateRequestRate calculates requests per second
func (op *OptimizedLStatePool) calculateRequestRate() float64 {
	op.requestTimesMu.Lock()
	defer op.requestTimesMu.Unlock()

	if len(op.requestTimes) < 2 {
		return 0
	}

	duration := op.requestTimes[len(op.requestTimes)-1].Sub(op.requestTimes[0])
	if duration <= 0 {
		return 0
	}

	return float64(len(op.requestTimes)) / duration.Seconds()
}

// trackExecutionStart tracks the start of execution for a state
func (op *OptimizedLStatePool) trackExecutionStart(stateID int64) {
	op.stateLoadMu.Lock()
	defer op.stateLoadMu.Unlock()

	load, exists := op.stateLoad[stateID]
	if !exists {
		load = &StateLoadInfo{}
		op.stateLoad[stateID] = load
	}

	load.lastExecution = time.Now()
	load.executionCount++
}

// trackExecutionEnd tracks the end of execution for a state
func (op *OptimizedLStatePool) trackExecutionEnd(stateID int64) {
	op.stateLoadMu.Lock()
	defer op.stateLoadMu.Unlock()

	load, exists := op.stateLoad[stateID]
	if !exists {
		return
	}

	duration := time.Since(load.lastExecution)
	load.totalDuration += duration
	load.avgDuration = load.totalDuration / time.Duration(load.executionCount)
}

// GetOptimizationMetrics returns optimization-specific metrics
func (op *OptimizedLStatePool) GetOptimizationMetrics() map[string]interface{} {
	base := op.GetMetrics()

	return map[string]interface{}{
		"base_metrics":          base,
		"predicted_scale_ups":   atomic.LoadInt64(&op.predictedScaleUps),
		"predicted_scale_downs": atomic.LoadInt64(&op.predictedScaleDowns),
		"pre_warmed_states":     atomic.LoadInt64(&op.preWarmedStates),
		"memory_pool_hits":      atomic.LoadInt64(&op.memoryPoolHits),
		"memory_pool_misses":    atomic.LoadInt64(&op.memoryPoolMisses),
		"current_request_rate":  op.calculateRequestRate(),
		"scale_up_threshold":    op.config.ScaleUpThreshold,
		"scale_down_threshold":  op.config.ScaleDownThreshold,
	}
}

// Shutdown gracefully shuts down the optimized pool
func (op *OptimizedLStatePool) Shutdown(ctx context.Context) error {
	// Stop prediction ticker
	if op.predictionTicker != nil {
		op.predictionTicker.Stop()
	}

	// Wait for pre-warming to complete
	op.WaitForPreWarm()

	// Clear memory pool
	if op.config.EnableMemoryPooling {
		op.memoryMu.Lock()
		op.memoryBlockPool = nil
		op.memoryMu.Unlock()
	}

	// Shutdown base pool
	return op.LStatePool.Shutdown(ctx)
}

// ApplyLoadProfile applies a predefined load profile for optimization
func (op *OptimizedLStatePool) ApplyLoadProfile(profile string) error {
	switch profile {
	case "burst":
		// Optimize for burst traffic
		op.config.ScaleUpThreshold = 0.6
		op.config.ScaleDownThreshold = 0.1
		op.config.MaxPredictedScaleUp = 10

	case "steady":
		// Optimize for steady traffic
		op.config.ScaleUpThreshold = 0.85
		op.config.ScaleDownThreshold = 0.25
		op.config.MaxPredictedScaleUp = 3

	case "periodic":
		// Optimize for periodic traffic
		op.config.EnableAdaptiveThresholds = true
		op.config.PredictionWindowSize = 20

	case "memory_intensive":
		// Optimize for memory-intensive workloads
		op.config.EnableMemoryPooling = true
		op.config.MemoryPoolSize = 20
		op.config.MemoryBlockSize = 2 * 1024 * 1024 // 2MB

	default:
		return fmt.Errorf("unknown load profile: %s", profile)
	}

	return nil
}
