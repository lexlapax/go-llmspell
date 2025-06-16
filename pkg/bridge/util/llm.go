// ABOUTME: LLM utilities bridge provides access to go-llms LLM utility functions.
// ABOUTME: Wraps provider creation, typed generation, pooling, and model inventory utilities.

package util

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for LLM utilities
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/outputs"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	llmjson "github.com/lexlapax/go-llms/pkg/util/json"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"
	modelinfoDomain "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// UtilLLMBridge provides script access to go-llms LLM utilities.
type UtilLLMBridge struct {
	mu          sync.RWMutex
	initialized bool

	// Enhanced components from go-llms v0.3.5
	metadataRegistry map[string]provider.ProviderMetadata // Provider capabilities
	modelService     *modelinfo.ModelInfoService          // Model discovery
	eventEmitter     agentDomain.EventEmitter             // For streaming events
	eventBus         *events.EventBus                     // Event bus for streaming
	validator        schemaDomain.Validator               // Schema validation
	costTracker      *CostTracker                         // Per-request cost tracking
}

// CostTracker tracks costs per request
type CostTracker struct {
	mu     sync.RWMutex
	costs  map[string]*RequestCost
	totals map[string]float64 // Total costs per provider
}

// RequestCost represents the cost of a single request
type RequestCost struct {
	RequestID    string
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	InputCost    float64
	OutputCost   float64
	TotalCost    float64
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// NewUtilLLMBridge creates a new LLM utilities bridge.
func NewUtilLLMBridge() *UtilLLMBridge {
	return &UtilLLMBridge{
		metadataRegistry: make(map[string]provider.ProviderMetadata),
		costTracker: &CostTracker{
			costs:  make(map[string]*RequestCost),
			totals: make(map[string]float64),
		},
	}
}

// NewUtilLLMBridgeWithEventEmitter creates a new LLM utilities bridge with event emitter.
func NewUtilLLMBridgeWithEventEmitter(eventEmitter agentDomain.EventEmitter) *UtilLLMBridge {
	return &UtilLLMBridge{
		eventEmitter:     eventEmitter,
		metadataRegistry: make(map[string]provider.ProviderMetadata),
		costTracker: &CostTracker{
			costs:  make(map[string]*RequestCost),
			totals: make(map[string]float64),
		},
	}
}

// GetID returns the bridge identifier.
func (b *UtilLLMBridge) GetID() string {
	return "util_llm"
}

// GetMetadata returns bridge metadata.
func (b *UtilLLMBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_llm",
		Version:     "2.0.0",
		Description: "Enhanced LLM utilities with provider capabilities, model discovery, response parsing, streaming events, and cost tracking",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilLLMBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize enhanced components from go-llms v0.3.5
	// Note: Parser is an interface, not a struct - will be set when needed

	if b.validator == nil {
		b.validator = validation.NewValidator()
	}

	if b.eventBus == nil {
		b.eventBus = events.NewEventBus()
	}

	// Initialize model service for discovery
	if b.modelService == nil {
		// Use the factory function to create service
		b.modelService = modelinfo.NewModelInfoServiceFunc()
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilLLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilLLMBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilLLMBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilLLMBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Provider creation utilities
		{
			Name:        "createProvider",
			Description: "Create an LLM provider from configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Description: "Provider configuration", Required: true},
			},
			ReturnType: "Provider",
		},
		{
			Name:        "createProviderFromEnv",
			Description: "Create an LLM provider from environment variables",
			Parameters: []engine.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "Provider",
		},
		{
			Name:        "withProviderOptions",
			Description: "Create provider-specific options",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
				{Name: "options", Type: "object", Description: "Provider-specific options", Required: true},
			},
			ReturnType: "object",
		},

		// Typed generation utilities
		{
			Name:        "generateTyped",
			Description: "Generate a typed/structured response",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "Provider", Description: "LLM provider", Required: true},
				{Name: "prompt", Type: "string", Description: "Generation prompt", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema for output", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateStructuredOutput",
			Description: "Validate structured output against schema",
			Parameters: []engine.ParameterInfo{
				{Name: "output", Type: "object", Description: "Output to validate", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema", Required: true},
			},
			ReturnType: "boolean",
		},

		// Provider pool utilities
		{
			Name:        "createProviderPool",
			Description: "Create a provider pool for load balancing/failover",
			Parameters: []engine.ParameterInfo{
				{Name: "providers", Type: "array", Description: "Array of providers", Required: true},
				{Name: "strategy", Type: "string", Description: "Pool strategy (roundrobin/failover/fastest)", Required: true},
			},
			ReturnType: "ProviderPool",
		},
		{
			Name:        "addProviderToPool",
			Description: "Add a provider to an existing pool",
			Parameters: []engine.ParameterInfo{
				{Name: "pool", Type: "ProviderPool", Description: "Provider pool", Required: true},
				{Name: "provider", Type: "Provider", Description: "Provider to add", Required: true},
				{Name: "weight", Type: "number", Description: "Provider weight", Required: false},
			},
			ReturnType: "void",
		},

		// Model inventory utilities
		{
			Name:        "createModelInventory",
			Description: "Create a model inventory service",
			Parameters: []engine.ParameterInfo{
				{Name: "fetchers", Type: "object", Description: "Provider fetchers configuration", Required: false},
			},
			ReturnType: "ModelInventory",
		},
		{
			Name:        "fetchModelInfo",
			Description: "Fetch model information for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "inventory", Type: "ModelInventory", Description: "Model inventory", Required: true},
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "cacheModelInfo",
			Description: "Cache model information to file",
			Parameters: []engine.ParameterInfo{
				{Name: "inventory", Type: "ModelInventory", Description: "Model inventory", Required: true},
				{Name: "cachePath", Type: "string", Description: "Cache file path", Required: true},
			},
			ReturnType: "void",
		},

		// Configuration utilities
		{
			Name:        "createModelConfig",
			Description: "Create a model configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
				{Name: "model", Type: "string", Description: "Model name", Required: true},
				{Name: "options", Type: "object", Description: "Model options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "mergeProviderOptions",
			Description: "Merge multiple provider options",
			Parameters: []engine.ParameterInfo{
				{Name: "options1", Type: "object", Description: "First options", Required: true},
				{Name: "options2", Type: "object", Description: "Second options", Required: true},
			},
			ReturnType: "object",
		},

		// Enhanced v0.3.5 features
		{
			Name:        "getProviderCapabilities",
			Description: "Get provider capability metadata",
			Parameters: []engine.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "discoverModels",
			Description: "Discover available models for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
				{Name: "refresh", Type: "boolean", Description: "Force refresh from API", Required: false},
			},
			ReturnType: "array",
		},
		{
			Name:        "parseResponseWithRecovery",
			Description: "Parse LLM response with recovery for malformed output",
			Parameters: []engine.ParameterInfo{
				{Name: "response", Type: "string", Description: "LLM response", Required: true},
				{Name: "format", Type: "string", Description: "Expected format (json/xml/yaml)", Required: false},
				{Name: "schema", Type: "object", Description: "Optional schema for validation", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "streamWithEvents",
			Description: "Stream LLM response with event emission",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "Provider", Description: "LLM provider", Required: true},
				{Name: "prompt", Type: "string", Description: "Generation prompt", Required: true},
				{Name: "eventHandler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "trackRequestCost",
			Description: "Track cost for an LLM request",
			Parameters: []engine.ParameterInfo{
				{Name: "requestID", Type: "string", Description: "Request identifier", Required: true},
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
				{Name: "model", Type: "string", Description: "Model name", Required: true},
				{Name: "usage", Type: "object", Description: "Token usage data", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getCostReport",
			Description: "Get cost tracking report",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Filter criteria", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "createProviderOptions",
			Description: "Create provider-specific options with advanced features",
			Parameters: []engine.ParameterInfo{
				{Name: "providerType", Type: "string", Description: "Provider type", Required: true},
				{Name: "config", Type: "object", Description: "Configuration options", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *UtilLLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"ProviderPool": {
			GoType:     "ProviderPool",
			ScriptType: "object",
		},
		"ModelInventory": {
			GoType:     "ModelInventory",
			ScriptType: "object",
		},
		"ModelConfig": {
			GoType:     "ModelConfig",
			ScriptType: "object",
		},
		"ProviderMetadata": {
			GoType:     "ProviderMetadata",
			ScriptType: "object",
		},
		"RequestCost": {
			GoType:     "RequestCost",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls.
func (b *UtilLLMBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *UtilLLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm",
			Actions:     []string{"create", "access"},
			Description: "Create and access LLM providers",
		},
		{
			Type:        engine.PermissionFileSystem,
			Resource:    "cache",
			Actions:     []string{"read", "write"},
			Description: "Cache model information",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "metadata",
			Actions:     []string{"read", "write"},
			Description: "Store provider metadata and cost tracking",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *UtilLLMBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createProviderPool":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProviderPool requires providers and strategy parameters")
		}

		// Get providers array
		providersArg, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("providers must be an array")
		}

		// Convert to domain.Provider slice
		providers := make([]bridge.Provider, 0, len(providersArg))
		for i, p := range providersArg {
			provider, ok := p.(bridge.Provider)
			if !ok {
				return nil, fmt.Errorf("provider at index %d must be a Provider", i)
			}
			providers = append(providers, provider)
		}

		// Get strategy
		strategyStr, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("strategy must be string")
		}

		// Convert strategy string to enum
		var strategy llmutil.PoolStrategy
		switch strings.ToLower(strategyStr) {
		case "roundrobin":
			strategy = llmutil.StrategyRoundRobin
		case "failover":
			strategy = llmutil.StrategyFailover
		case "fastest":
			strategy = llmutil.StrategyFastest
		default:
			return nil, fmt.Errorf("invalid strategy: %s", strategyStr)
		}

		// Create the provider pool
		pool := llmutil.NewProviderPool(providers, strategy)

		return pool, nil

	case "createModelInventory":
		// Create model info service to fetch model inventory
		// Using the modelinfo package's service to aggregate models
		// Note: The actual model inventory is returned by the service's AggregateModels method
		// For now, return a placeholder as the service requires provider-specific fetchers
		return map[string]interface{}{
			"type": "ModelInventory",
			"id":   "inventory_1",
			"note": "Use fetchModelInfo to retrieve actual model data",
		}, nil

	case "createModelConfig":
		if len(args) < 2 {
			return nil, fmt.Errorf("createModelConfig requires provider and model parameters")
		}
		provider, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("provider must be string")
		}
		model, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("model must be string")
		}

		// Create model config
		config := llmutil.ModelConfig{
			Provider: provider,
			Model:    model,
		}

		// Add options if provided
		if len(args) > 2 && args[2] != nil {
			if options, ok := args[2].(map[string]interface{}); ok {
				// Options will be applied when go-llms ModelConfig supports additional fields
				_ = options
			}
		}

		return map[string]interface{}{
			"provider": config.Provider,
			"model":    config.Model,
		}, nil

	// Enhanced v0.3.5 features
	case "getProviderCapabilities":
		if len(args) < 1 {
			return nil, fmt.Errorf("getProviderCapabilities requires providerName")
		}
		providerName, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("providerName must be string")
		}

		// Check if we have cached metadata
		if metadata, exists := b.metadataRegistry[providerName]; exists {
			return convertProviderMetadataToMap(metadata), nil
		}

		// TODO: Load metadata from provider when available in go-llms
		// For now, return basic capabilities based on provider type
		capabilities := map[string]interface{}{
			"provider": providerName,
			"capabilities": map[string]bool{
				"streaming":       true,
				"functionCalling": providerName == "openai" || providerName == "anthropic",
				"vision":          providerName == "openai" || providerName == "anthropic",
				"embeddings":      providerName == "openai",
			},
			"constraints": map[string]interface{}{
				"maxTokens":     4096,
				"rateLimit":     60,
				"contextWindow": 8192,
			},
		}

		return capabilities, nil

	case "discoverModels":
		if len(args) < 1 {
			return nil, fmt.Errorf("discoverModels requires providerName")
		}
		providerName, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("providerName must be string")
		}

		// Check refresh flag (not used in current implementation)
		// refresh := false
		// if len(args) > 1 {
		// 	refresh, _ = args[1].(bool)
		// }

		// Use model service to aggregate models from all providers
		// Note: ModelInfoService doesn't have provider-specific methods
		inventory, err := b.modelService.AggregateModels()
		if err != nil {
			return nil, fmt.Errorf("failed to discover models: %w", err)
		}

		// Filter models for the requested provider
		var models []modelinfoDomain.Model
		for _, model := range inventory.Models {
			if model.Provider == providerName {
				models = append(models, model)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to discover models: %w", err)
		}

		// Convert models to script-friendly format
		result := make([]map[string]interface{}, 0, len(models))
		for _, model := range models {
			result = append(result, map[string]interface{}{
				"id":            model.Name, // Use Name as ID
				"name":          model.DisplayName,
				"description":   model.Description,
				"inputCost":     model.Pricing.InputPer1kTokens,
				"outputCost":    model.Pricing.OutputPer1kTokens,
				"maxTokens":     model.MaxOutputTokens,
				"contextWindow": model.ContextWindow,
				"capabilities": map[string]interface{}{
					"streaming":       model.Capabilities.Streaming,
					"functionCalling": model.Capabilities.FunctionCalling,
					"vision":          model.Capabilities.Image.Read,
					"jsonMode":        model.Capabilities.JSONMode,
				},
			})
		}

		return result, nil

	case "parseResponseWithRecovery":
		if len(args) < 1 {
			return nil, fmt.Errorf("parseResponseWithRecovery requires response")
		}
		response, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("response must be string")
		}

		// Get optional format
		format := ""
		if len(args) > 1 && args[1] != nil {
			format, _ = args[1].(string)
		}

		// Get optional schema
		var schema *schemaDomain.Schema
		if len(args) > 2 && args[2] != nil {
			if schemaMap, ok := args[2].(map[string]interface{}); ok {
				// Convert to schema
				schemaJSON, _ := llmjson.Marshal(schemaMap)
				schema = &schemaDomain.Schema{}
				if err := llmjson.Unmarshal(schemaJSON, schema); err != nil {
					return nil, fmt.Errorf("invalid schema: %w", err)
				}
			}
		}

		// Parse with recovery
		// Note: go-llms RecoveryOptions has specific fields
		options := &outputs.RecoveryOptions{
			ExtractFromMarkdown: true,
			FixCommonIssues:     true,
			StrictMode:          false,
			MaxAttempts:         3,
			Schema:              nil, // Schema is part of OutputSchema, not RecoveryOptions
		}

		// Auto-detect parser based on format or response content
		var parser outputs.Parser
		var parseErr error

		if format != "" {
			// Get specific parser by format
			parser, parseErr = outputs.GetParser(format)
			if parseErr != nil {
				// Try auto-detection if format parser not found
				parser, parseErr = outputs.AutoDetectParser(response)
				if parseErr != nil {
					return nil, fmt.Errorf("failed to find suitable parser: %w", parseErr)
				}
			}
		} else {
			// Auto-detect parser from response
			parser, parseErr = outputs.AutoDetectParser(response)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to auto-detect parser: %w", parseErr)
			}
		}

		result, err := parser.ParseWithRecovery(ctx, response, options)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return result, nil

	case "streamWithEvents":
		if len(args) < 3 {
			return nil, fmt.Errorf("streamWithEvents requires provider, prompt, and eventHandler")
		}

		provider, ok := args[0].(bridge.Provider)
		if !ok {
			return nil, fmt.Errorf("provider must be Provider")
		}
		prompt, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("prompt must be string")
		}
		eventHandler, ok := args[2].(func(interface{}) error)
		if !ok {
			return nil, fmt.Errorf("eventHandler must be function")
		}

		// Create message for streaming
		message := domain.NewTextMessage(domain.RoleUser, prompt)

		// Call StreamMessage on provider
		// Note: Provider interface has StreamMessage method, not GenerateStream
		responseChan, err := provider.StreamMessage(ctx, []domain.Message{message})
		if err != nil {
			return nil, fmt.Errorf("failed to start streaming: %w", err)
		}

		// Collect full response while emitting events
		var fullContent strings.Builder
		tokenCount := 0

		for token := range responseChan {
			// domain.Token has Text and Finished fields
			if token.Finished {
				break
			}

			// Emit chunk event
			if b.eventEmitter != nil {
				b.eventEmitter.EmitCustom("stream.chunk", map[string]interface{}{
					"content":   token.Text,
					"index":     tokenCount,
					"timestamp": time.Now(),
				})
			}

			// Call script event handler
			if err := eventHandler(map[string]interface{}{
				"type":    "chunk",
				"content": token.Text,
				"index":   tokenCount,
			}); err != nil {
				return nil, fmt.Errorf("event handler error: %w", err)
			}

			fullContent.WriteString(token.Text)
			tokenCount++
		}

		// Emit completion event
		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("stream.complete", map[string]interface{}{
				"totalTokens": tokenCount,
				"content":     fullContent.String(),
				"timestamp":   time.Now(),
			})
		}

		return map[string]interface{}{
			"content":    fullContent.String(),
			"tokenCount": tokenCount,
		}, nil

	case "trackRequestCost":
		if len(args) < 4 {
			return nil, fmt.Errorf("trackRequestCost requires requestID, provider, model, and usage")
		}

		requestID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("requestID must be string")
		}
		provider, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("provider must be string")
		}
		model, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("model must be string")
		}
		usage, ok := args[3].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("usage must be object")
		}

		// Extract token counts
		inputTokens := 0
		outputTokens := 0
		totalTokens := 0

		if val, ok := usage["inputTokens"].(float64); ok {
			inputTokens = int(val)
		}
		if val, ok := usage["outputTokens"].(float64); ok {
			outputTokens = int(val)
		}
		if val, ok := usage["totalTokens"].(float64); ok {
			totalTokens = int(val)
		}

		// Get model pricing (would come from model metadata in real implementation)
		inputCostPer1k := 0.003 // Default pricing
		outputCostPer1k := 0.004

		// Calculate costs
		inputCost := float64(inputTokens) / 1000.0 * inputCostPer1k
		outputCost := float64(outputTokens) / 1000.0 * outputCostPer1k
		totalCost := inputCost + outputCost

		// Create and store request cost
		cost := &RequestCost{
			RequestID:    requestID,
			Provider:     provider,
			Model:        model,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			TotalTokens:  totalTokens,
			InputCost:    inputCost,
			OutputCost:   outputCost,
			TotalCost:    totalCost,
			Timestamp:    time.Now(),
			Metadata:     usage,
		}

		// Store in tracker
		b.costTracker.mu.Lock()
		b.costTracker.costs[requestID] = cost
		b.costTracker.totals[provider] += totalCost
		b.costTracker.mu.Unlock()

		// Emit cost event
		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("cost.tracked", map[string]interface{}{
				"requestID": requestID,
				"provider":  provider,
				"model":     model,
				"cost":      totalCost,
				"timestamp": time.Now(),
			})
		}

		return map[string]interface{}{
			"requestID": cost.RequestID,
			"totalCost": cost.TotalCost,
			"breakdown": map[string]interface{}{
				"inputCost":    cost.InputCost,
				"outputCost":   cost.OutputCost,
				"inputTokens":  cost.InputTokens,
				"outputTokens": cost.OutputTokens,
			},
		}, nil

	case "getCostReport":
		filter := make(map[string]interface{})
		if len(args) > 0 && args[0] != nil {
			if f, ok := args[0].(map[string]interface{}); ok {
				filter = f
			}
		}

		b.costTracker.mu.RLock()
		defer b.costTracker.mu.RUnlock()

		// Build report
		report := map[string]interface{}{
			"totalCosts": b.costTracker.totals,
			"requests":   []map[string]interface{}{},
			"summary": map[string]interface{}{
				"totalRequests": len(b.costTracker.costs),
				"providers":     make(map[string]int),
			},
		}

		// Filter and aggregate
		for _, cost := range b.costTracker.costs {
			// Apply filters
			if provider, ok := filter["provider"].(string); ok && cost.Provider != provider {
				continue
			}
			if model, ok := filter["model"].(string); ok && cost.Model != model {
				continue
			}

			// Add to report
			requests := report["requests"].([]map[string]interface{})
			requests = append(requests, map[string]interface{}{
				"requestID": cost.RequestID,
				"provider":  cost.Provider,
				"model":     cost.Model,
				"totalCost": cost.TotalCost,
				"timestamp": cost.Timestamp,
			})
			report["requests"] = requests

			// Update summary
			providers := report["summary"].(map[string]interface{})["providers"].(map[string]int)
			providers[cost.Provider]++
		}

		return report, nil

	case "createProviderOptions":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProviderOptions requires providerType and config")
		}

		providerType, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("providerType must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Create provider-specific options based on type
		options := map[string]interface{}{
			"type": providerType,
		}

		// Extract common options
		if baseURL, ok := config["baseURL"].(string); ok {
			options["baseURL"] = baseURL
		}
		if apiKey, ok := config["apiKey"].(string); ok {
			options["apiKey"] = apiKey
		}
		if timeout, ok := config["timeout"].(float64); ok {
			options["timeout"] = int(timeout)
		}

		// Add provider-specific options
		switch providerType {
		case "openai":
			if org, ok := config["organization"].(string); ok {
				options["organization"] = org
			}
			if apiVersion, ok := config["apiVersion"].(string); ok {
				options["apiVersion"] = apiVersion
			}

		case "anthropic":
			if apiVersion, ok := config["anthropicVersion"].(string); ok {
				options["anthropicVersion"] = apiVersion
			}

		case "gemini":
			if location, ok := config["location"].(string); ok {
				options["location"] = location
			}
			if projectID, ok := config["projectID"].(string); ok {
				options["projectID"] = projectID
			}
		}

		return options, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper function to convert ProviderMetadata to map
func convertProviderMetadataToMap(metadata provider.ProviderMetadata) map[string]interface{} {
	// Get capabilities
	capabilities := metadata.GetCapabilities()
	capMap := make(map[string]bool)
	for _, cap := range capabilities {
		switch cap {
		case provider.CapabilityStreaming:
			capMap["streaming"] = true
		case provider.CapabilityFunctionCalling:
			capMap["functionCalling"] = true
		case provider.CapabilityVision:
			capMap["vision"] = true
		case provider.CapabilityEmbeddings:
			capMap["embeddings"] = true
		case provider.CapabilityStructuredOutput:
			capMap["structured"] = true
		}
	}

	// Get constraints
	constraints := metadata.GetConstraints()

	return map[string]interface{}{
		"name":         metadata.Name(),
		"description":  metadata.Description(),
		"capabilities": capMap,
		"constraints": map[string]interface{}{
			"maxBatchSize":    constraints.MaxBatchSize,
			"maxConcurrency":  constraints.MaxConcurrency,
			"rateLimit":       constraints.RateLimit,
			"minRequestDelay": constraints.MinRequestDelay,
			"maxRetries":      constraints.MaxRetries,
		},
	}
}
