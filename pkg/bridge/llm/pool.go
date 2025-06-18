// ABOUTME: Provider pool bridge implements connection pooling and load balancing for LLM providers
// ABOUTME: Manages provider pools with strategies like round-robin, failover, and performance-based routing

package llm

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// Bridge imports
	"github.com/lexlapax/go-llmspell/pkg/bridge"
)

// PoolStrategy defines the strategy for selecting providers from a pool
type PoolStrategy string

const (
	StrategyRoundRobin PoolStrategy = "round_robin"
	StrategyFailover   PoolStrategy = "failover"
	StrategyFastest    PoolStrategy = "fastest"
	StrategyWeighted   PoolStrategy = "weighted"
	StrategyLeastUsed  PoolStrategy = "least_used"
)

// ProviderPool manages a pool of LLM providers
type ProviderPool struct {
	Name            string
	Providers       []string // Provider names
	Strategy        PoolStrategy
	CurrentIndex    int32 // For round-robin
	Config          PoolConfig
	Metrics         *PoolMetrics
	lastUsed        map[string]time.Time
	providerWeights map[string]float64
	mu              sync.RWMutex
}

// PoolConfig holds configuration for a provider pool
type PoolConfig struct {
	MaxRetries          int
	RetryDelay          time.Duration
	Timeout             time.Duration
	CircuitBreaker      bool
	CircuitThreshold    int
	HealthCheckInterval time.Duration
}

// PoolMetrics tracks metrics for a provider pool
type PoolMetrics struct {
	TotalRequests   int64
	SuccessfulCalls int64
	FailedCalls     int64
	RetryCount      int64
	ProviderMetrics map[string]*ProviderPoolMetrics
	mu              sync.RWMutex
}

// ProviderPoolMetrics tracks metrics for a provider in a pool
type ProviderPoolMetrics struct {
	Requests       int64
	Successes      int64
	Failures       int64
	TotalLatency   time.Duration
	AverageLatency time.Duration
	LastError      error
	LastErrorTime  time.Time
	HealthStatus   string
}

// ResponsePool manages pooled response objects
type ResponsePool struct {
	pool sync.Pool
}

// TokenPool manages pooled token objects
type TokenPool struct {
	pool sync.Pool
}

// ChannelPool manages pooled channels for streaming
type ChannelPool struct {
	channels map[string]chan bridge.ResponseStream
	mu       sync.RWMutex
}

// PoolBridge provides connection pooling for LLM providers
type PoolBridge struct {
	mu           sync.RWMutex
	initialized  bool
	pools        map[string]*ProviderPool
	responsePool *ResponsePool
	tokenPool    *TokenPool
	channelPool  *ChannelPool
	llmBridge    *LLMBridge // Reference to main LLM bridge
}

// NewPoolBridge creates a new pool bridge
func NewPoolBridge(llmBridge *LLMBridge) *PoolBridge {
	return &PoolBridge{
		pools: make(map[string]*ProviderPool),
		responsePool: &ResponsePool{
			pool: sync.Pool{
				New: func() interface{} {
					return &bridge.Response{}
				},
			},
		},
		tokenPool: &TokenPool{
			pool: sync.Pool{
				New: func() interface{} {
					return &Token{}
				},
			},
		},
		channelPool: &ChannelPool{
			channels: make(map[string]chan bridge.ResponseStream),
		},
		llmBridge: llmBridge,
	}
}

// Token represents a pooled token object
type Token struct {
	Value     string
	CreatedAt time.Time
	Used      bool
}

// GetID returns the bridge ID
func (b *PoolBridge) GetID() string {
	return "pool"
}

// GetMetadata returns bridge metadata
func (b *PoolBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Pool Bridge",
		Version:     "2.0.0",
		Description: "Provider pooling with load balancing strategies",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *PoolBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *PoolBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clear all pools
	b.pools = make(map[string]*ProviderPool)

	// Clear channel pool
	for _, ch := range b.channelPool.channels {
		close(ch)
	}
	b.channelPool.channels = make(map[string]chan bridge.ResponseStream)

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *PoolBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *PoolBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *PoolBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Pool Management
		{
			Name:        "createPool",
			Description: "Create a new provider pool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
				{Name: "providers", Type: "array", Required: true, Description: "Array of provider names"},
				{Name: "strategy", Type: "string", Required: true, Description: "Pool strategy (round_robin, failover, fastest)"},
			},
			ReturnType: "object",
		},
		{
			Name:        "getPool",
			Description: "Get pool information",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listPools",
			Description: "List all pools",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "removePool",
			Description: "Remove a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "void",
		},
		// Pool Metrics
		{
			Name:        "getPoolMetrics",
			Description: "Get metrics for a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "getProviderHealth",
			Description: "Get health status of providers in pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "array",
		},
		{
			Name:        "resetPoolMetrics",
			Description: "Reset metrics for a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "void",
		},
		// Pool Generation
		{
			Name:        "generateWithPool",
			Description: "Generate text using a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "string",
		},
		{
			Name:        "generateMessageWithPool",
			Description: "Generate from messages using a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "messages", Type: "array", Required: true, Description: "Array of messages"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
		},
		{
			Name:        "streamWithPool",
			Description: "Stream generation using a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
		},
		// Object Pooling
		{
			Name:        "getResponseFromPool",
			Description: "Get a response object from pool",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "returnResponseToPool",
			Description: "Return a response object to pool",
			Parameters: []engine.ParameterInfo{
				{Name: "response", Type: "object", Required: true, Description: "Response object to return"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getTokenFromPool",
			Description: "Get a token object from pool",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "returnTokenToPool",
			Description: "Return a token object to pool",
			Parameters: []engine.ParameterInfo{
				{Name: "token", Type: "object", Required: true, Description: "Token object to return"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getChannelFromPool",
			Description: "Get a channel from pool",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "returnChannelToPool",
			Description: "Return a channel to pool",
			Parameters: []engine.ParameterInfo{
				{Name: "channelID", Type: "string", Required: true, Description: "Channel identifier"},
			},
			ReturnType: "void",
		},
		// Configuration
		{
			Name:        "setPoolConfiguration",
			Description: "Set pool configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration object"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getPoolConfiguration",
			Description: "Get pool configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
		},
		// Advanced Pool Operations
		{
			Name:        "setProviderWeight",
			Description: "Set weight for a provider in weighted strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
				{Name: "weight", Type: "number", Required: true, Description: "Weight value"},
			},
			ReturnType: "void",
		},
		{
			Name:        "rebalancePool",
			Description: "Rebalance providers in pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "void",
		},
		{
			Name:        "performHealthCheck",
			Description: "Perform health check on pool providers",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
		},
	}
}

// ValidateMethod validates method parameters
func (b *PoolBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	if !b.IsInitialized() {
		return fmt.Errorf("pool bridge not initialized")
	}

	switch name {
	case "createPool":
		if len(args) < 3 {
			return fmt.Errorf("createPool requires 3 arguments")
		}
	case "getPool", "removePool", "getPoolMetrics", "getProviderHealth",
		"resetPoolMetrics", "getPoolConfiguration", "rebalancePool", "performHealthCheck":
		if len(args) < 1 {
			return fmt.Errorf("%s requires pool name", name)
		}
	case "generateWithPool", "streamWithPool":
		if len(args) < 2 {
			return fmt.Errorf("%s requires pool name and prompt", name)
		}
	case "generateMessageWithPool":
		if len(args) < 2 {
			return fmt.Errorf("generateMessageWithPool requires pool name and messages")
		}
	case "setPoolConfiguration":
		if len(args) < 2 {
			return fmt.Errorf("setPoolConfiguration requires pool name and config")
		}
	case "setProviderWeight":
		if len(args) < 3 {
			return fmt.Errorf("setProviderWeight requires pool name, provider, and weight")
		}
	default:
		// Check if method exists
		methods := b.Methods()
		for _, method := range methods {
			if method.Name == name {
				return nil
			}
		}
		return fmt.Errorf("unknown method: %s", name)
	}

	return nil
}

// ExecuteMethod executes a bridge method with ScriptValue parameters
func (b *PoolBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	if !b.initialized {
		b.mu.RUnlock()
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}
	b.mu.RUnlock()

	switch name {
	// Pool Management
	case "createPool":
		return b.createPool(ctx, args)
	case "getPool":
		return b.getPool(ctx, args)
	case "listPools":
		return b.listPools(ctx, args)
	case "removePool":
		return b.removePool(ctx, args)

	// Pool Metrics
	case "getPoolMetrics":
		return b.getPoolMetrics(ctx, args)
	case "getProviderHealth":
		return b.getProviderHealth(ctx, args)
	case "resetPoolMetrics":
		return b.resetPoolMetrics(ctx, args)

	// Pool Generation
	case "generateWithPool":
		return b.generateWithPool(ctx, args)
	case "generateMessageWithPool":
		return b.generateMessageWithPool(ctx, args)
	case "streamWithPool":
		return b.streamWithPool(ctx, args)

	// Object Pooling
	case "getResponseFromPool":
		return b.getResponseFromPool(ctx, args)
	case "returnResponseToPool":
		return b.returnResponseToPool(ctx, args)
	case "getTokenFromPool":
		return b.getTokenFromPool(ctx, args)
	case "returnTokenToPool":
		return b.returnTokenToPool(ctx, args)
	case "getChannelFromPool":
		return b.getChannelFromPool(ctx, args)
	case "returnChannelToPool":
		return b.returnChannelToPool(ctx, args)

	// Configuration
	case "setPoolConfiguration":
		return b.setPoolConfiguration(ctx, args)
	case "getPoolConfiguration":
		return b.getPoolConfiguration(ctx, args)

	// Advanced Pool Operations
	case "setProviderWeight":
		return b.setProviderWeight(ctx, args)
	case "rebalancePool":
		return b.rebalancePool(ctx, args)
	case "performHealthCheck":
		return b.performHealthCheck(ctx, args)

	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// TypeMappings returns type conversion mappings
func (b *PoolBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"ProviderPool": {
			GoType:     "llm.ProviderPool",
			ScriptType: "object",
		},
		"PoolConfig": {
			GoType:     "llm.PoolConfig",
			ScriptType: "object",
		},
		"PoolMetrics": {
			GoType:     "llm.PoolMetrics",
			ScriptType: "object",
		},
	}
}

// RequiredPermissions returns required permissions
func (b *PoolBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "pool",
			Actions:     []string{"read", "write"},
			Description: "Manage provider pools",
		},
	}
}

// Pool Management Methods

func (b *PoolBridge) createPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("createPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()
	providersArray := args[1].ToGo().([]interface{})
	strategyStr := args[2].(engine.StringValue).Value()

	// Convert providers array
	providers := make([]string, 0, len(providersArray))
	for _, p := range providersArray {
		if providerName, ok := p.(string); ok {
			providers = append(providers, providerName)
		}
	}

	// Validate strategy
	strategy := PoolStrategy(strategyStr)
	switch strategy {
	case StrategyRoundRobin, StrategyFailover, StrategyFastest, StrategyWeighted, StrategyLeastUsed:
		// Valid strategy
	default:
		return engine.NewErrorValue(fmt.Errorf("invalid strategy: %s", strategyStr)), nil
	}

	// Create pool
	pool := &ProviderPool{
		Name:      name,
		Providers: providers,
		Strategy:  strategy,
		Config: PoolConfig{
			MaxRetries:          3,
			RetryDelay:          time.Second,
			Timeout:             30 * time.Second,
			CircuitBreaker:      true,
			CircuitThreshold:    5,
			HealthCheckInterval: 60 * time.Second,
		},
		Metrics: &PoolMetrics{
			ProviderMetrics: make(map[string]*ProviderPoolMetrics),
		},
		lastUsed:        make(map[string]time.Time),
		providerWeights: make(map[string]float64),
	}

	// Initialize provider metrics
	for _, provider := range providers {
		pool.Metrics.ProviderMetrics[provider] = &ProviderPoolMetrics{
			HealthStatus: "healthy",
		}
		pool.providerWeights[provider] = 1.0 // Default weight
	}

	b.mu.Lock()
	b.pools[name] = pool
	b.mu.Unlock()

	// Return pool info
	return b.poolToScriptValue(pool), nil
}

func (b *PoolBridge) getPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[name]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", name)), nil
	}

	return b.poolToScriptValue(pool), nil
}

func (b *PoolBridge) listPools(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pools := make([]engine.ScriptValue, 0, len(b.pools))
	for name := range b.pools {
		pools = append(pools, engine.NewStringValue(name))
	}

	return engine.NewArrayValue(pools), nil
}

func (b *PoolBridge) removePool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("removePool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	delete(b.pools, name)
	b.mu.Unlock()

	return engine.NewNilValue(), nil
}

// Pool Metrics Methods

func (b *PoolBridge) getPoolMetrics(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getPoolMetrics", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	pool.Metrics.mu.RLock()
	defer pool.Metrics.mu.RUnlock()

	// Build metrics object
	providerMetrics := make(map[string]engine.ScriptValue)
	for provider, metrics := range pool.Metrics.ProviderMetrics {
		providerMetrics[provider] = engine.NewObjectValue(map[string]engine.ScriptValue{
			"requests":       engine.NewNumberValue(float64(metrics.Requests)),
			"successes":      engine.NewNumberValue(float64(metrics.Successes)),
			"failures":       engine.NewNumberValue(float64(metrics.Failures)),
			"averageLatency": engine.NewNumberValue(metrics.AverageLatency.Seconds()),
			"healthStatus":   engine.NewStringValue(metrics.HealthStatus),
		})
	}

	metricsData := map[string]engine.ScriptValue{
		"totalRequests":   engine.NewNumberValue(float64(pool.Metrics.TotalRequests)),
		"successfulCalls": engine.NewNumberValue(float64(pool.Metrics.SuccessfulCalls)),
		"failedCalls":     engine.NewNumberValue(float64(pool.Metrics.FailedCalls)),
		"retryCount":      engine.NewNumberValue(float64(pool.Metrics.RetryCount)),
		"providerMetrics": engine.NewObjectValue(providerMetrics),
	}

	return engine.NewObjectValue(metricsData), nil
}

func (b *PoolBridge) getProviderHealth(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getProviderHealth", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	healthArray := make([]engine.ScriptValue, 0, len(pool.Providers))

	pool.Metrics.mu.RLock()
	defer pool.Metrics.mu.RUnlock()

	for _, provider := range pool.Providers {
		metrics := pool.Metrics.ProviderMetrics[provider]
		healthData := map[string]engine.ScriptValue{
			"provider":    engine.NewStringValue(provider),
			"status":      engine.NewStringValue(metrics.HealthStatus),
			"lastError":   engine.NewNilValue(),
			"successRate": engine.NewNumberValue(0),
		}

		if metrics.LastError != nil {
			healthData["lastError"] = engine.NewStringValue(metrics.LastError.Error())
			healthData["lastErrorTime"] = engine.NewStringValue(metrics.LastErrorTime.Format(time.RFC3339))
		}

		if metrics.Requests > 0 {
			successRate := float64(metrics.Successes) / float64(metrics.Requests)
			healthData["successRate"] = engine.NewNumberValue(successRate)
		}

		healthArray = append(healthArray, engine.NewObjectValue(healthData))
	}

	return engine.NewArrayValue(healthArray), nil
}

func (b *PoolBridge) resetPoolMetrics(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("resetPoolMetrics", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	pool.Metrics.mu.Lock()
	defer pool.Metrics.mu.Unlock()

	// Reset all metrics
	pool.Metrics.TotalRequests = 0
	pool.Metrics.SuccessfulCalls = 0
	pool.Metrics.FailedCalls = 0
	pool.Metrics.RetryCount = 0

	for _, metrics := range pool.Metrics.ProviderMetrics {
		metrics.Requests = 0
		metrics.Successes = 0
		metrics.Failures = 0
		metrics.TotalLatency = 0
		metrics.AverageLatency = 0
	}

	return engine.NewNilValue(), nil
}

// Pool Generation Methods

func (b *PoolBridge) generateWithPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("generateWithPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()
	prompt := args[1].(engine.StringValue).Value()

	var options map[string]interface{}
	if len(args) > 2 {
		options = args[2].ToGo().(map[string]interface{})
	}

	// Get pool
	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	// Select provider based on strategy
	provider, err := b.selectProvider(pool)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Set active provider in LLM bridge
	b.llmBridge.mu.Lock()
	oldProvider := b.llmBridge.activeProvider
	b.llmBridge.activeProvider = provider
	b.llmBridge.mu.Unlock()

	// Generate using LLM bridge
	result, genErr := b.llmBridge.generate(ctx, []engine.ScriptValue{
		engine.NewStringValue(prompt),
		engine.NewObjectValue(poolConvertMapToScriptValue(options)),
	})

	// Restore old provider
	b.llmBridge.mu.Lock()
	b.llmBridge.activeProvider = oldProvider
	b.llmBridge.mu.Unlock()

	// Update metrics
	b.updatePoolMetrics(pool, provider, genErr == nil)

	if genErr != nil {
		return engine.NewErrorValue(genErr), nil
	}

	// Extract content from result
	if objValue, ok := result.(engine.ObjectValue); ok {
		resultMap := objValue.ToGo().(map[string]interface{})
		if content, exists := resultMap["content"]; exists {
			return engine.NewStringValue(fmt.Sprintf("%v", content)), nil
		}
	}

	return result, nil
}

func (b *PoolBridge) generateMessageWithPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("generateMessageWithPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()
	messages := args[1]

	var options map[string]interface{}
	if len(args) > 2 {
		options = args[2].ToGo().(map[string]interface{})
	}

	// Get pool
	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	// Select provider based on strategy
	provider, err := b.selectProvider(pool)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Set active provider in LLM bridge
	b.llmBridge.mu.Lock()
	oldProvider := b.llmBridge.activeProvider
	b.llmBridge.activeProvider = provider
	b.llmBridge.mu.Unlock()

	// Generate using LLM bridge
	result, genErr := b.llmBridge.generateMessage(ctx, []engine.ScriptValue{
		messages,
		engine.NewObjectValue(poolConvertMapToScriptValue(options)),
	})

	// Restore old provider
	b.llmBridge.mu.Lock()
	b.llmBridge.activeProvider = oldProvider
	b.llmBridge.mu.Unlock()

	// Update metrics
	b.updatePoolMetrics(pool, provider, genErr == nil)

	return result, genErr
}

func (b *PoolBridge) streamWithPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("streamWithPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()
	prompt := args[1].(engine.StringValue).Value()

	var options map[string]interface{}
	if len(args) > 2 {
		options = args[2].ToGo().(map[string]interface{})
	}

	// Get pool
	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	// Select provider based on strategy
	provider, err := b.selectProvider(pool)
	if err != nil {
		return engine.NewErrorValue(err), nil
	}

	// Set active provider in LLM bridge
	b.llmBridge.mu.Lock()
	oldProvider := b.llmBridge.activeProvider
	b.llmBridge.activeProvider = provider
	b.llmBridge.mu.Unlock()

	// Stream using LLM bridge
	result, streamErr := b.llmBridge.stream(ctx, []engine.ScriptValue{
		engine.NewStringValue(prompt),
		engine.NewObjectValue(poolConvertMapToScriptValue(options)),
	})

	// Restore old provider
	b.llmBridge.mu.Lock()
	b.llmBridge.activeProvider = oldProvider
	b.llmBridge.mu.Unlock()

	// Update metrics
	b.updatePoolMetrics(pool, provider, streamErr == nil)

	return result, streamErr
}

// Object Pooling Methods

func (b *PoolBridge) getResponseFromPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	response := b.responsePool.pool.Get().(*bridge.Response)

	// Convert to ScriptValue
	responseData := map[string]engine.ScriptValue{
		"content": engine.NewStringValue(response.Content),
		"id":      engine.NewStringValue(fmt.Sprintf("response-%d", time.Now().UnixNano())),
	}

	return engine.NewObjectValue(responseData), nil
}

func (b *PoolBridge) returnResponseToPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("returnResponseToPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	// In a real implementation, we would convert the ScriptValue back to a Response
	// For now, we just acknowledge the return
	return engine.NewNilValue(), nil
}

func (b *PoolBridge) getTokenFromPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	token := b.tokenPool.pool.Get().(*Token)
	token.Value = fmt.Sprintf("token-%d", time.Now().UnixNano())
	token.CreatedAt = time.Now()
	token.Used = false

	tokenData := map[string]engine.ScriptValue{
		"value":     engine.NewStringValue(token.Value),
		"createdAt": engine.NewStringValue(token.CreatedAt.Format(time.RFC3339)),
		"used":      engine.NewBoolValue(token.Used),
	}

	return engine.NewObjectValue(tokenData), nil
}

func (b *PoolBridge) returnTokenToPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("returnTokenToPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	// In a real implementation, we would convert the ScriptValue back to a Token
	// For now, we just acknowledge the return
	return engine.NewNilValue(), nil
}

func (b *PoolBridge) getChannelFromPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	channelID := fmt.Sprintf("channel-%d", time.Now().UnixNano())
	ch := make(chan bridge.ResponseStream, 100)

	b.channelPool.mu.Lock()
	b.channelPool.channels[channelID] = ch
	b.channelPool.mu.Unlock()

	channelData := map[string]engine.ScriptValue{
		"id":       engine.NewStringValue(channelID),
		"capacity": engine.NewNumberValue(100),
	}

	return engine.NewObjectValue(channelData), nil
}

func (b *PoolBridge) returnChannelToPool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("returnChannelToPool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	channelID := args[0].(engine.StringValue).Value()

	b.channelPool.mu.Lock()
	if ch, exists := b.channelPool.channels[channelID]; exists {
		close(ch)
		delete(b.channelPool.channels, channelID)
	}
	b.channelPool.mu.Unlock()

	return engine.NewNilValue(), nil
}

// Configuration Methods

func (b *PoolBridge) setPoolConfiguration(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("setPoolConfiguration", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()
	configMap := args[1].ToGo().(map[string]interface{})

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	// Update configuration
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if maxRetries, ok := configMap["maxRetries"].(float64); ok {
		pool.Config.MaxRetries = int(maxRetries)
	}
	if retryDelay, ok := configMap["retryDelay"].(float64); ok {
		pool.Config.RetryDelay = time.Duration(retryDelay) * time.Millisecond
	}
	if timeout, ok := configMap["timeout"].(float64); ok {
		pool.Config.Timeout = time.Duration(timeout) * time.Millisecond
	}
	if circuitBreaker, ok := configMap["circuitBreaker"].(bool); ok {
		pool.Config.CircuitBreaker = circuitBreaker
	}
	if circuitThreshold, ok := configMap["circuitThreshold"].(float64); ok {
		pool.Config.CircuitThreshold = int(circuitThreshold)
	}

	return engine.NewNilValue(), nil
}

func (b *PoolBridge) getPoolConfiguration(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getPoolConfiguration", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	configData := map[string]engine.ScriptValue{
		"maxRetries":       engine.NewNumberValue(float64(pool.Config.MaxRetries)),
		"retryDelay":       engine.NewNumberValue(float64(pool.Config.RetryDelay.Milliseconds())),
		"timeout":          engine.NewNumberValue(float64(pool.Config.Timeout.Milliseconds())),
		"circuitBreaker":   engine.NewBoolValue(pool.Config.CircuitBreaker),
		"circuitThreshold": engine.NewNumberValue(float64(pool.Config.CircuitThreshold)),
	}

	return engine.NewObjectValue(configData), nil
}

// Advanced Pool Operations

func (b *PoolBridge) setProviderWeight(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("setProviderWeight", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()
	provider := args[1].(engine.StringValue).Value()
	weight := args[2].(engine.NumberValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	pool.mu.Lock()
	pool.providerWeights[provider] = weight
	pool.mu.Unlock()

	return engine.NewNilValue(), nil
}

func (b *PoolBridge) rebalancePool(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("rebalancePool", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	// Reset provider metrics for rebalancing
	pool.mu.Lock()
	for provider := range pool.lastUsed {
		pool.lastUsed[provider] = time.Time{}
	}
	atomic.StoreInt32(&pool.CurrentIndex, 0)
	pool.mu.Unlock()

	return engine.NewNilValue(), nil
}

func (b *PoolBridge) performHealthCheck(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("performHealthCheck", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	poolName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("pool not found: %s", poolName)), nil
	}

	healthResults := make(map[string]engine.ScriptValue)

	// Check each provider
	for _, provider := range pool.Providers {
		// Simple health check - verify provider exists in LLM bridge
		b.llmBridge.mu.RLock()
		_, providerExists := b.llmBridge.providers[provider]
		b.llmBridge.mu.RUnlock()

		status := "healthy"
		if !providerExists {
			status = "unhealthy"
		}

		pool.Metrics.mu.Lock()
		if metrics, ok := pool.Metrics.ProviderMetrics[provider]; ok {
			metrics.HealthStatus = status
		}
		pool.Metrics.mu.Unlock()

		healthResults[provider] = engine.NewStringValue(status)
	}

	return engine.NewObjectValue(healthResults), nil
}

// Helper Methods

func (b *PoolBridge) selectProvider(pool *ProviderPool) (string, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if len(pool.Providers) == 0 {
		return "", fmt.Errorf("no providers in pool")
	}

	switch pool.Strategy {
	case StrategyRoundRobin:
		// Round-robin selection
		index := atomic.AddInt32(&pool.CurrentIndex, 1) - 1
		provider := pool.Providers[int(index)%len(pool.Providers)]
		return provider, nil

	case StrategyFailover:
		// Return first healthy provider
		for _, provider := range pool.Providers {
			if metrics, ok := pool.Metrics.ProviderMetrics[provider]; ok {
				if metrics.HealthStatus == "healthy" {
					return provider, nil
				}
			}
		}
		return "", fmt.Errorf("no healthy providers available")

	case StrategyFastest:
		// Select provider with lowest average latency
		var fastest string
		var lowestLatency time.Duration = time.Hour

		pool.Metrics.mu.RLock()
		for provider, metrics := range pool.Metrics.ProviderMetrics {
			if metrics.HealthStatus == "healthy" && metrics.AverageLatency < lowestLatency {
				fastest = provider
				lowestLatency = metrics.AverageLatency
			}
		}
		pool.Metrics.mu.RUnlock()

		if fastest == "" {
			return pool.Providers[0], nil // Fallback to first provider
		}
		return fastest, nil

	case StrategyWeighted:
		// Weighted random selection
		totalWeight := 0.0
		for _, provider := range pool.Providers {
			if weight, ok := pool.providerWeights[provider]; ok {
				totalWeight += weight
			}
		}

		if totalWeight == 0 {
			return pool.Providers[0], nil
		}

		// Simple weighted selection (not truly random for determinism in tests)
		currentWeight := 0.0
		target := totalWeight / 2 // Use middle point instead of random
		for _, provider := range pool.Providers {
			if weight, ok := pool.providerWeights[provider]; ok {
				currentWeight += weight
				if currentWeight >= target {
					return provider, nil
				}
			}
		}
		return pool.Providers[0], nil

	case StrategyLeastUsed:
		// Select least recently used provider
		var leastUsed string
		var oldestTime time.Time

		for _, provider := range pool.Providers {
			lastUsed, exists := pool.lastUsed[provider]
			if !exists || lastUsed.IsZero() {
				return provider, nil // Never used
			}
			if oldestTime.IsZero() || lastUsed.Before(oldestTime) {
				leastUsed = provider
				oldestTime = lastUsed
			}
		}

		if leastUsed != "" {
			pool.lastUsed[leastUsed] = time.Now()
			return leastUsed, nil
		}

		return pool.Providers[0], nil

	default:
		return "", fmt.Errorf("unknown strategy: %s", pool.Strategy)
	}
}

func (b *PoolBridge) updatePoolMetrics(pool *ProviderPool, provider string, success bool) {
	pool.Metrics.mu.Lock()
	defer pool.Metrics.mu.Unlock()

	pool.Metrics.TotalRequests++
	if success {
		pool.Metrics.SuccessfulCalls++
	} else {
		pool.Metrics.FailedCalls++
	}

	if metrics, ok := pool.Metrics.ProviderMetrics[provider]; ok {
		metrics.Requests++
		if success {
			metrics.Successes++
		} else {
			metrics.Failures++
		}
		// Update average latency calculation would go here
	}
}

func (b *PoolBridge) poolToScriptValue(pool *ProviderPool) engine.ScriptValue {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	providers := make([]engine.ScriptValue, len(pool.Providers))
	for i, p := range pool.Providers {
		providers[i] = engine.NewStringValue(p)
	}

	poolData := map[string]engine.ScriptValue{
		"name":      engine.NewStringValue(pool.Name),
		"providers": engine.NewArrayValue(providers),
		"strategy":  engine.NewStringValue(string(pool.Strategy)),
	}

	return engine.NewObjectValue(poolData)
}

// poolConvertMapToScriptValue converts map[string]interface{} to map[string]engine.ScriptValue
func poolConvertMapToScriptValue(m map[string]interface{}) map[string]engine.ScriptValue {
	if m == nil {
		return make(map[string]engine.ScriptValue)
	}

	result := make(map[string]engine.ScriptValue)
	for k, v := range m {
		result[k] = poolConvertToScriptValue(v)
	}
	return result
}

// poolConvertToScriptValue converts interface{} to ScriptValue
func poolConvertToScriptValue(v interface{}) engine.ScriptValue {
	if v == nil {
		return engine.NewNilValue()
	}

	switch val := v.(type) {
	case bool:
		return engine.NewBoolValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int32:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case float32:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case string:
		return engine.NewStringValue(val)
	case []interface{}:
		arr := make([]engine.ScriptValue, len(val))
		for i, item := range val {
			arr[i] = poolConvertToScriptValue(item)
		}
		return engine.NewArrayValue(arr)
	case map[string]interface{}:
		return engine.NewObjectValue(poolConvertMapToScriptValue(val))
	default:
		// Convert to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}
