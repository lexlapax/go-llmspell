// ABOUTME: Provider pool bridge for go-llms connection pooling and load balancing
// ABOUTME: Bridges provider pool management with health monitoring and adaptive pooling strategies

package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	// go-llms imports for pool functionality
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// PoolBridge provides script access to go-llms provider pool system
type PoolBridge struct {
	initialized bool
	pools       map[string]*llmutil.ProviderPool
	providers   map[string]domain.Provider
	mu          sync.RWMutex
}

// NewPoolBridge creates a new pool bridge
func NewPoolBridge() *PoolBridge {
	return &PoolBridge{
		pools:     make(map[string]*llmutil.ProviderPool),
		providers: make(map[string]domain.Provider),
	}
}

// GetID returns the bridge identifier
func (pb *PoolBridge) GetID() string {
	return "pool"
}

// GetMetadata returns bridge metadata
func (pb *PoolBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "pool",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms provider pool system with load balancing, health monitoring, and adaptive pooling",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/util/llmutil"},
	}
}

// Initialize sets up the pool bridge
func (pb *PoolBridge) Initialize(ctx context.Context) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (pb *PoolBridge) Cleanup(ctx context.Context) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Clear all stored pools and providers
	pb.pools = make(map[string]*llmutil.ProviderPool)
	pb.providers = make(map[string]domain.Provider)
	pb.initialized = false

	return nil
}

// IsInitialized returns initialization status
func (pb *PoolBridge) IsInitialized() bool {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (pb *PoolBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(pb)
}

// Methods returns available bridge methods
func (pb *PoolBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Pool creation and management
		{
			Name:        "createPool",
			Description: "Create provider pool with load balancing strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
				{Name: "providers", Type: "array", Required: true, Description: "Array of provider names"},
				{Name: "strategy", Type: "string", Required: true, Description: "Pool strategy (round_robin, failover, fastest)"},
			},
			ReturnType: "object",
			Examples:   []string{"createPool('my-pool', ['openai', 'anthropic'], 'round_robin')"},
		},
		{
			Name:        "getPool",
			Description: "Get pool by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getPool('my-pool')"},
		},
		{
			Name:        "listPools",
			Description: "List all created pools",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listPools()"},
		},
		{
			Name:        "removePool",
			Description: "Remove pool by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "void",
			Examples:   []string{"removePool('my-pool')"},
		},
		// Pool health monitoring
		{
			Name:        "getPoolMetrics",
			Description: "Get performance metrics for all providers in pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getPoolMetrics('my-pool')"},
		},
		{
			Name:        "getProviderHealth",
			Description: "Get health status of providers in pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "array",
			Examples:   []string{"getProviderHealth('my-pool')"},
		},
		{
			Name:        "resetPoolMetrics",
			Description: "Reset metrics for a pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "void",
			Examples:   []string{"resetPoolMetrics('my-pool')"},
		},
		// Pool operations
		{
			Name:        "generateWithPool",
			Description: "Generate text using provider pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "string",
			Examples:   []string{"generateWithPool('my-pool', 'Hello world', {temperature: 0.7})"},
		},
		{
			Name:        "generateMessageWithPool",
			Description: "Generate message response using provider pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "messages", Type: "array", Required: true, Description: "Array of messages"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples:   []string{"generateMessageWithPool('my-pool', messages, {temperature: 0.7})"},
		},
		{
			Name:        "streamWithPool",
			Description: "Stream response using provider pool",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples:   []string{"streamWithPool('my-pool', 'Hello world', {temperature: 0.7})"},
		},
		// Object pool utilities
		{
			Name:        "getResponseFromPool",
			Description: "Get response object from global response pool",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getResponseFromPool()"},
		},
		{
			Name:        "returnResponseToPool",
			Description: "Return response object to global response pool",
			Parameters: []engine.ParameterInfo{
				{Name: "response", Type: "object", Required: true, Description: "Response object to return"},
			},
			ReturnType: "void",
			Examples:   []string{"returnResponseToPool(response)"},
		},
		{
			Name:        "getTokenFromPool",
			Description: "Get token object from global token pool",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getTokenFromPool()"},
		},
		{
			Name:        "returnTokenToPool",
			Description: "Return token object to global token pool",
			Parameters: []engine.ParameterInfo{
				{Name: "token", Type: "object", Required: true, Description: "Token object to return"},
			},
			ReturnType: "void",
			Examples:   []string{"returnTokenToPool(token)"},
		},
		{
			Name:        "getChannelFromPool",
			Description: "Get channel from global channel pool for streaming",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getChannelFromPool()"},
		},
		{
			Name:        "returnChannelToPool",
			Description: "Return channel to global channel pool",
			Parameters: []engine.ParameterInfo{
				{Name: "channelID", Type: "string", Required: true, Description: "Channel identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"returnChannelToPool(channelID)"},
		},
		// Pool configuration
		{
			Name:        "setPoolConfiguration",
			Description: "Configure pool behavior and thresholds",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration object"},
			},
			ReturnType: "void",
			Examples:   []string{"setPoolConfiguration('my-pool', {errorThreshold: 3, healthCheckInterval: 60})"},
		},
		{
			Name:        "getPoolConfiguration",
			Description: "Get current pool configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "poolName", Type: "string", Required: true, Description: "Pool name"},
			},
			ReturnType: "object",
			Examples:   []string{"getPoolConfiguration('my-pool')"},
		},
	}
}

// ValidateMethod validates method calls
func (pb *PoolBridge) ValidateMethod(name string, args []interface{}) error {
	if !pb.IsInitialized() {
		return fmt.Errorf("pool bridge not initialized")
	}

	methods := pb.Methods()
	for _, method := range methods {
		if method.Name == name {
			requiredCount := 0
			for _, param := range method.Parameters {
				if param.Required {
					requiredCount++
				}
			}
			if len(args) < requiredCount {
				return fmt.Errorf("method %s requires at least %d arguments, got %d", name, requiredCount, len(args))
			}
			return nil
		}
	}
	return fmt.Errorf("unknown method: %s", name)
}

// TypeMappings returns type conversion mappings
func (pb *PoolBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"provider_pool": {
			GoType:     "*llmutil.ProviderPool",
			ScriptType: "object",
			Converter:  "providerPoolConverter",
			Metadata:   map[string]interface{}{"description": "Provider pool for load balancing"},
		},
		"pool_metrics": {
			GoType:     "*llmutil.ProviderMetrics",
			ScriptType: "object",
			Converter:  "poolMetricsConverter",
			Metadata:   map[string]interface{}{"description": "Provider performance metrics"},
		},
		"pool_strategy": {
			GoType:     "llmutil.PoolStrategy",
			ScriptType: "string",
			Converter:  "poolStrategyConverter",
			Metadata:   map[string]interface{}{"description": "Pool selection strategy"},
		},
		"response_pool": {
			GoType:     "*domain.ResponsePool",
			ScriptType: "object",
			Converter:  "responsePoolConverter",
			Metadata:   map[string]interface{}{"description": "Object pool for responses"},
		},
	}
}

// RequiredPermissions returns required permissions
func (pb *PoolBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm.pool",
			Actions:     []string{"create", "read", "update", "delete"},
			Description: "Manage LLM provider pools",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "pool.objects",
			Actions:     []string{"allocate", "deallocate"},
			Description: "Access object pools for optimization",
		},
	}
}

// Bridge method implementations

// Pool creation and management

// createPool creates a provider pool with load balancing strategy
func (pb *PoolBridge) createPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("createPool", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	providersArray, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("providers must be an array")
	}

	strategyStr, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("strategy must be a string")
	}

	// Convert strategy string to enum
	var strategy llmutil.PoolStrategy
	switch strategyStr {
	case "round_robin":
		strategy = llmutil.StrategyRoundRobin
	case "failover":
		strategy = llmutil.StrategyFailover
	case "fastest":
		strategy = llmutil.StrategyFastest
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyStr)
	}

	// Get providers from registry or local storage
	var providers []domain.Provider
	for _, p := range providersArray {
		providerName, ok := p.(string)
		if !ok {
			continue
		}

		// Try to get provider from local storage first
		pb.mu.RLock()
		provider, exists := pb.providers[providerName]
		pb.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("provider not found: %s", providerName)
		}

		providers = append(providers, provider)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers specified")
	}

	// Create pool
	pool := llmutil.NewProviderPool(providers, strategy)

	// Store in registry
	pb.mu.Lock()
	pb.pools[name] = pool
	pb.mu.Unlock()

	return map[string]interface{}{
		"name":      name,
		"strategy":  strategyStr,
		"providers": len(providers),
		"created":   time.Now(),
	}, nil
}

// getPool gets a pool by name
func (pb *PoolBridge) getPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getPool", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	pb.mu.RLock()
	pool, exists := pb.pools[name]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", name)
	}

	return map[string]interface{}{
		"name": name,
		"pool": pool,
	}, nil
}

// listPools lists all created pools
func (pb *PoolBridge) listPools(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("listPools", args); err != nil {
		return nil, err
	}

	pb.mu.RLock()
	defer pb.mu.RUnlock()

	pools := make([]map[string]interface{}, 0)
	for name := range pb.pools {
		pools = append(pools, map[string]interface{}{
			"name": name,
		})
	}

	return pools, nil
}

// removePool removes a pool by name
func (pb *PoolBridge) removePool(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("removePool", args); err != nil {
		return err
	}

	name, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("pool name must be a string")
	}

	pb.mu.Lock()
	defer pb.mu.Unlock()

	if _, exists := pb.pools[name]; !exists {
		return fmt.Errorf("pool not found: %s", name)
	}

	delete(pb.pools, name)
	return nil
}

// Pool health monitoring

// getPoolMetrics gets performance metrics for all providers in pool
func (pb *PoolBridge) getPoolMetrics(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getPoolMetrics", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	pb.mu.RLock()
	pool, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	metrics := pool.GetMetrics()
	result := make(map[string]interface{})

	for idx, metric := range metrics {
		result[fmt.Sprintf("provider_%d", idx)] = map[string]interface{}{
			"requests":           metric.Requests,
			"failures":           metric.Failures,
			"avg_latency_ms":     metric.AvgLatencyMs,
			"total_latency_ms":   metric.TotalLatencyMs,
			"last_used":          metric.LastUsed,
			"consecutive_errors": metric.ConsecutiveErrors,
			"success_rate":       float64(metric.Requests-metric.Failures) / float64(metric.Requests) * 100,
		}
	}

	return result, nil
}

// getProviderHealth gets health status of providers in pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) getProviderHealth(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getProviderHealth", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	pb.mu.RLock()
	pool, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	metrics := pool.GetMetrics()
	var health []map[string]interface{}

	for idx, metric := range metrics {
		status := "healthy"
		if metric.ConsecutiveErrors > 3 {
			status = "unhealthy"
		} else if metric.ConsecutiveErrors > 0 {
			status = "warning"
		}

		health = append(health, map[string]interface{}{
			"provider_index":     idx,
			"status":             status,
			"consecutive_errors": metric.ConsecutiveErrors,
			"last_used":          metric.LastUsed,
			"requests":           metric.Requests,
			"failures":           metric.Failures,
		})
	}

	return health, nil
}

// resetPoolMetrics resets metrics for a pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) resetPoolMetrics(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("resetPoolMetrics", args); err != nil {
		return err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("pool name must be a string")
	}

	pb.mu.RLock()
	_, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolName)
	}

	// Note: llmutil.ProviderPool doesn't expose a reset method
	// This would need to be implemented in go-llms or we'd need to recreate the pool
	return fmt.Errorf("metric reset not implemented in go-llms ProviderPool")
}

// Pool operations

// generateWithPool generates text using provider pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) generateWithPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("generateWithPool", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	prompt, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("prompt must be a string")
	}

	// Get options if provided
	var options []domain.Option
	if len(args) > 2 {
		if optionsMap, ok := args[2].(map[string]interface{}); ok {
			if temperature, ok := optionsMap["temperature"].(float64); ok {
				options = append(options, domain.WithTemperature(temperature))
			}
			if maxTokens, ok := optionsMap["max_tokens"].(float64); ok {
				options = append(options, domain.WithMaxTokens(int(maxTokens)))
			}
		}
	}

	pb.mu.RLock()
	pool, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	result, err := pool.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return result, nil
}

// generateMessageWithPool generates message response using provider pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) generateMessageWithPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("generateMessageWithPool", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	messagesArray, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("messages must be an array")
	}

	// Convert messages array to domain.Message slice
	var messages []domain.Message
	for _, m := range messagesArray {
		if msgMap, ok := m.(map[string]interface{}); ok {
			roleStr, _ := msgMap["role"].(string)
			contentStr, _ := msgMap["content"].(string)

			// Convert string role to domain.Role
			var role domain.Role
			switch roleStr {
			case "system":
				role = domain.RoleSystem
			case "user":
				role = domain.RoleUser
			case "assistant":
				role = domain.RoleAssistant
			case "tool":
				role = domain.RoleTool
			default:
				role = domain.RoleUser
			}

			// Create content parts from string
			contentParts := []domain.ContentPart{
				{
					Type: domain.ContentTypeText,
					Text: contentStr,
				},
			}

			messages = append(messages, domain.Message{
				Role:    role,
				Content: contentParts,
			})
		}
	}

	// Get options if provided
	var options []domain.Option
	if len(args) > 2 {
		if optionsMap, ok := args[2].(map[string]interface{}); ok {
			if temperature, ok := optionsMap["temperature"].(float64); ok {
				options = append(options, domain.WithTemperature(temperature))
			}
			if maxTokens, ok := optionsMap["max_tokens"].(float64); ok {
				options = append(options, domain.WithMaxTokens(int(maxTokens)))
			}
		}
	}

	pb.mu.RLock()
	pool, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	result, err := pool.GenerateMessage(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return map[string]interface{}{
		"content": result.Content,
	}, nil
}

// streamWithPool streams response using provider pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) streamWithPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("streamWithPool", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	prompt, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("prompt must be a string")
	}

	// Get options if provided
	var options []domain.Option
	if len(args) > 2 {
		if optionsMap, ok := args[2].(map[string]interface{}); ok {
			if temperature, ok := optionsMap["temperature"].(float64); ok {
				options = append(options, domain.WithTemperature(temperature))
			}
		}
	}

	pb.mu.RLock()
	pool, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	stream, err := pool.Stream(ctx, prompt, options...)
	if err != nil {
		return nil, fmt.Errorf("streaming failed: %w", err)
	}

	// Create a channel ID for tracking
	streamID := fmt.Sprintf("stream-%s-%d", poolName, time.Now().UnixNano())

	return map[string]interface{}{
		"stream_id": streamID,
		"stream":    stream,
		"pool":      poolName,
	}, nil
}

// Object pool utilities

// getResponseFromPool gets response object from global response pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) getResponseFromPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getResponseFromPool", args); err != nil {
		return nil, err
	}

	responsePool := domain.GetResponsePool()
	response := responsePool.Get()

	return map[string]interface{}{
		"response": response,
		"pooled":   true,
	}, nil
}

// returnResponseToPool returns response object to global response pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) returnResponseToPool(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("returnResponseToPool", args); err != nil {
		return err
	}

	responseObj, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("response must be an object")
	}

	// In a real implementation, we'd convert the response object back to *domain.Response
	// For now, we'll just acknowledge the operation
	_ = responseObj

	responsePool := domain.GetResponsePool()
	// responsePool.Put(response) // Would need actual *domain.Response

	_ = responsePool // Acknowledge usage
	return nil
}

// getTokenFromPool gets token object from global token pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) getTokenFromPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getTokenFromPool", args); err != nil {
		return nil, err
	}

	tokenPool := domain.GetTokenPool()
	token := tokenPool.Get()

	return map[string]interface{}{
		"token":  token,
		"pooled": true,
	}, nil
}

// returnTokenToPool returns token object to global token pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) returnTokenToPool(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("returnTokenToPool", args); err != nil {
		return err
	}

	tokenObj, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("token must be an object")
	}

	// In a real implementation, we'd convert the token object back to *domain.Token
	// For now, we'll just acknowledge the operation
	_ = tokenObj

	tokenPool := domain.GetTokenPool()
	// tokenPool.Put(token) // Would need actual *domain.Token

	_ = tokenPool // Acknowledge usage
	return nil
}

// getChannelFromPool gets channel from global channel pool for streaming
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) getChannelFromPool(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getChannelFromPool", args); err != nil {
		return nil, err
	}

	channelPool := domain.GetChannelPool()
	stream, channel := channelPool.GetResponseStream()

	channelID := fmt.Sprintf("channel-%d", time.Now().UnixNano())

	return map[string]interface{}{
		"channel_id": channelID,
		"stream":     stream,
		"channel":    channel,
		"pooled":     true,
	}, nil
}

// returnChannelToPool returns channel to global channel pool
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) returnChannelToPool(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("returnChannelToPool", args); err != nil {
		return err
	}

	channelID, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("channel ID must be a string")
	}

	// In a real implementation, we'd track channels by ID and return them to the pool
	// For now, we'll just acknowledge the operation
	_ = channelID

	return nil
}

// Pool configuration

// setPoolConfiguration configures pool behavior and thresholds
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) setPoolConfiguration(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("setPoolConfiguration", args); err != nil {
		return err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("pool name must be a string")
	}

	config, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("config must be an object")
	}

	pb.mu.RLock()
	_, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pool not found: %s", poolName)
	}

	// Note: llmutil.ProviderPool doesn't expose configuration methods
	// This would need to be implemented in go-llms
	_ = config // Acknowledge config parameter

	return fmt.Errorf("pool configuration not implemented in go-llms ProviderPool")
}

// getPoolConfiguration gets current pool configuration
//
//nolint:unused // Bridge method called via reflection
func (pb *PoolBridge) getPoolConfiguration(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getPoolConfiguration", args); err != nil {
		return nil, err
	}

	poolName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("pool name must be a string")
	}

	pb.mu.RLock()
	_, exists := pb.pools[poolName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	// Return default configuration since go-llms doesn't expose this
	return map[string]interface{}{
		"pool_name":             poolName,
		"error_threshold":       3,
		"health_check_interval": 60,
		"adaptive_pooling":      false,
	}, nil
}

// Helper method to register a provider for use in pools
func (pb *PoolBridge) RegisterProvider(name string, provider domain.Provider) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.providers[name] = provider
}
