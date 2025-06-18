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
func (b *UtilLLMBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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
func (b *UtilLLMBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
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
		if args[0] == nil || args[0].Type() != engine.TypeArray {
			return nil, fmt.Errorf("providers must be an array")
		}
		providersArg := args[0].(engine.ArrayValue).Elements()

		// Convert to domain.Provider slice
		providers := make([]bridge.Provider, 0, len(providersArg))
		for i, p := range providersArg {
			// For now, we'll need custom handling for Provider types
			if p.Type() != engine.TypeCustom {
				return nil, fmt.Errorf("provider at index %d must be a Provider", i)
			}
			customVal := p.(engine.CustomValue)
			provider, ok := customVal.Value().(bridge.Provider)
			if !ok {
				return nil, fmt.Errorf("provider at index %d must be a Provider", i)
			}
			providers = append(providers, provider)
		}

		// Get strategy
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("strategy must be string")
		}
		strategyStr := args[1].(engine.StringValue).Value()

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

		// Return as custom value since pool is a complex type
		return engine.NewCustomValue("ProviderPool", pool), nil

	case "createModelInventory":
		// Create model info service to fetch model inventory
		// Using the modelinfo package's service to aggregate models
		// Note: The actual model inventory is returned by the service's AggregateModels method
		// For now, return a placeholder as the service requires provider-specific fetchers
		result := map[string]engine.ScriptValue{
			"type": engine.NewStringValue("ModelInventory"),
			"id":   engine.NewStringValue("inventory_1"),
			"note": engine.NewStringValue("Use fetchModelInfo to retrieve actual model data"),
		}
		return engine.NewObjectValue(result), nil

	case "createModelConfig":
		if len(args) < 2 {
			return nil, fmt.Errorf("createModelConfig requires provider and model parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("provider must be string")
		}
		provider := args[0].(engine.StringValue).Value()

		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("model must be string")
		}
		model := args[1].(engine.StringValue).Value()

		// Create model config
		config := llmutil.ModelConfig{
			Provider: provider,
			Model:    model,
		}

		// Add options if provided
		if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
			options := make(map[string]interface{})
			for k, v := range args[2].(engine.ObjectValue).Fields() {
				options[k] = v.ToGo()
			}
			// Options will be applied when go-llms ModelConfig supports additional fields
			_ = options
		}

		result := map[string]engine.ScriptValue{
			"provider": engine.NewStringValue(config.Provider),
			"model":    engine.NewStringValue(config.Model),
		}
		return engine.NewObjectValue(result), nil

	// Enhanced v0.3.5 features
	case "getProviderCapabilities":
		if len(args) < 1 {
			return nil, fmt.Errorf("getProviderCapabilities requires providerName")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("providerName must be string")
		}
		providerName := args[0].(engine.StringValue).Value()

		// Check if we have cached metadata
		if metadata, exists := b.metadataRegistry[providerName]; exists {
			return convertProviderMetadataToScriptValue(metadata), nil
		}

		// TODO: Load metadata from provider when available in go-llms
		// For now, return basic capabilities based on provider type
		capabilities := map[string]engine.ScriptValue{
			"provider": engine.NewStringValue(providerName),
			"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
				"streaming":       engine.NewBoolValue(true),
				"functionCalling": engine.NewBoolValue(providerName == "openai" || providerName == "anthropic"),
				"vision":          engine.NewBoolValue(providerName == "openai" || providerName == "anthropic"),
				"embeddings":      engine.NewBoolValue(providerName == "openai"),
			}),
			"constraints": engine.NewObjectValue(map[string]engine.ScriptValue{
				"maxTokens":     engine.NewNumberValue(4096),
				"rateLimit":     engine.NewNumberValue(60),
				"contextWindow": engine.NewNumberValue(8192),
			}),
		}

		return engine.NewObjectValue(capabilities), nil

	case "discoverModels":
		if len(args) < 1 {
			return nil, fmt.Errorf("discoverModels requires providerName")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("providerName must be string")
		}
		providerName := args[0].(engine.StringValue).Value()

		// Check refresh flag (not used in current implementation)
		// refresh := false
		// if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeBool {
		// 	refresh = args[1].(engine.BoolValue).Value()
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
		result := make([]engine.ScriptValue, 0, len(models))
		for _, model := range models {
			modelInfo := map[string]engine.ScriptValue{
				"id":            engine.NewStringValue(model.Name), // Use Name as ID
				"name":          engine.NewStringValue(model.DisplayName),
				"description":   engine.NewStringValue(model.Description),
				"inputCost":     engine.NewNumberValue(model.Pricing.InputPer1kTokens),
				"outputCost":    engine.NewNumberValue(model.Pricing.OutputPer1kTokens),
				"maxTokens":     engine.NewNumberValue(float64(model.MaxOutputTokens)),
				"contextWindow": engine.NewNumberValue(float64(model.ContextWindow)),
				"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
					"streaming":       engine.NewBoolValue(model.Capabilities.Streaming),
					"functionCalling": engine.NewBoolValue(model.Capabilities.FunctionCalling),
					"vision":          engine.NewBoolValue(model.Capabilities.Image.Read),
					"jsonMode":        engine.NewBoolValue(model.Capabilities.JSONMode),
				}),
			}
			result = append(result, engine.NewObjectValue(modelInfo))
		}

		return engine.NewArrayValue(result), nil

	case "parseResponseWithRecovery":
		if len(args) < 1 {
			return nil, fmt.Errorf("parseResponseWithRecovery requires response")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("response must be string")
		}
		response := args[0].(engine.StringValue).Value()

		// Get optional format
		format := ""
		if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
			format = args[1].(engine.StringValue).Value()
		}

		// Get optional schema
		var schema *schemaDomain.Schema
		if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
			schemaMap := make(map[string]interface{})
			for k, v := range args[2].(engine.ObjectValue).Fields() {
				schemaMap[k] = v.ToGo()
			}
			// Convert to schema
			schemaJSON, _ := llmjson.Marshal(schemaMap)
			schema = &schemaDomain.Schema{}
			if err := llmjson.Unmarshal(schemaJSON, schema); err != nil {
				return nil, fmt.Errorf("invalid schema: %w", err)
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

		// Convert result to ScriptValue
		return engine.NewCustomValue("ParsedResponse", result), nil

	case "streamWithEvents":
		if len(args) < 3 {
			return nil, fmt.Errorf("streamWithEvents requires provider, prompt, and eventHandler")
		}

		// Get provider from custom value
		if args[0] == nil || args[0].Type() != engine.TypeCustom {
			return nil, fmt.Errorf("provider must be Provider")
		}
		customVal := args[0].(engine.CustomValue)
		_, ok := customVal.Value().(bridge.Provider)
		if !ok {
			return nil, fmt.Errorf("provider must be Provider")
		}

		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("prompt must be string")
		}
		_ = args[1].(engine.StringValue).Value()

		// Get event handler function from custom value
		if args[2] == nil || args[2].Type() != engine.TypeFunction {
			return nil, fmt.Errorf("eventHandler must be function")
		}
		// For now, we'll need custom handling for function types
		// The engine will need to provide a way to call script functions
		// This is a placeholder that would need engine-specific implementation
		return nil, fmt.Errorf("function callbacks not yet implemented for ScriptValue")

		// TODO: The rest of this method would need engine-specific function callback support
		// return map[string]engine.ScriptValue{
		// 	"content":    engine.NewStringValue(fullContent.String()),
		// 	"tokenCount": engine.NewNumberValue(float64(tokenCount)),
		// }, nil

	case "trackRequestCost":
		if len(args) < 4 {
			return nil, fmt.Errorf("trackRequestCost requires requestID, provider, model, and usage")
		}

		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("requestID must be string")
		}
		requestID := args[0].(engine.StringValue).Value()

		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("provider must be string")
		}
		provider := args[1].(engine.StringValue).Value()

		if args[2] == nil || args[2].Type() != engine.TypeString {
			return nil, fmt.Errorf("model must be string")
		}
		model := args[2].(engine.StringValue).Value()

		if args[3] == nil || args[3].Type() != engine.TypeObject {
			return nil, fmt.Errorf("usage must be object")
		}
		usageObj := args[3].(engine.ObjectValue).Fields()
		usage := make(map[string]interface{})
		for k, v := range usageObj {
			usage[k] = v.ToGo()
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

		return engine.NewObjectValue(map[string]engine.ScriptValue{
			"requestID": engine.NewStringValue(cost.RequestID),
			"totalCost": engine.NewNumberValue(cost.TotalCost),
			"breakdown": engine.NewObjectValue(map[string]engine.ScriptValue{
				"inputCost":    engine.NewNumberValue(cost.InputCost),
				"outputCost":   engine.NewNumberValue(cost.OutputCost),
				"inputTokens":  engine.NewNumberValue(float64(cost.InputTokens)),
				"outputTokens": engine.NewNumberValue(float64(cost.OutputTokens)),
			}),
		}), nil

	case "getCostReport":
		filter := make(map[string]interface{})
		if len(args) > 0 && args[0] != nil && args[0].Type() == engine.TypeObject {
			filterObj := args[0].(engine.ObjectValue).Fields()
			for k, v := range filterObj {
				filter[k] = v.ToGo()
			}
		}

		b.costTracker.mu.RLock()
		defer b.costTracker.mu.RUnlock()

		// Convert totals to ScriptValue
		totalCosts := make(map[string]engine.ScriptValue)
		for provider, cost := range b.costTracker.totals {
			totalCosts[provider] = engine.NewNumberValue(cost)
		}

		// Build report
		requests := make([]engine.ScriptValue, 0)
		providers := make(map[string]int)

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
			requestInfo := map[string]engine.ScriptValue{
				"requestID": engine.NewStringValue(cost.RequestID),
				"provider":  engine.NewStringValue(cost.Provider),
				"model":     engine.NewStringValue(cost.Model),
				"totalCost": engine.NewNumberValue(cost.TotalCost),
				"timestamp": engine.NewStringValue(cost.Timestamp.Format(time.RFC3339)),
			}
			requests = append(requests, engine.NewObjectValue(requestInfo))

			// Update summary
			providers[cost.Provider]++
		}

		// Convert providers count to ScriptValue
		providersCount := make(map[string]engine.ScriptValue)
		for provider, count := range providers {
			providersCount[provider] = engine.NewNumberValue(float64(count))
		}

		report := map[string]engine.ScriptValue{
			"totalCosts": engine.NewObjectValue(totalCosts),
			"requests":   engine.NewArrayValue(requests),
			"summary": engine.NewObjectValue(map[string]engine.ScriptValue{
				"totalRequests": engine.NewNumberValue(float64(len(b.costTracker.costs))),
				"providers":     engine.NewObjectValue(providersCount),
			}),
		}

		return engine.NewObjectValue(report), nil

	case "createProviderOptions":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProviderOptions requires providerType and config")
		}

		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("providerType must be string")
		}
		providerType := args[0].(engine.StringValue).Value()

		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return nil, fmt.Errorf("config must be object")
		}
		configObj := args[1].(engine.ObjectValue).Fields()
		config := make(map[string]interface{})
		for k, v := range configObj {
			config[k] = v.ToGo()
		}

		// Create provider-specific options based on type
		options := map[string]engine.ScriptValue{
			"type": engine.NewStringValue(providerType),
		}

		// Extract common options
		if baseURL, ok := config["baseURL"].(string); ok {
			options["baseURL"] = engine.NewStringValue(baseURL)
		}
		if apiKey, ok := config["apiKey"].(string); ok {
			options["apiKey"] = engine.NewStringValue(apiKey)
		}
		if timeout, ok := config["timeout"].(float64); ok {
			options["timeout"] = engine.NewNumberValue(float64(int(timeout)))
		}

		// Add provider-specific options
		switch providerType {
		case "openai":
			if org, ok := config["organization"].(string); ok {
				options["organization"] = engine.NewStringValue(org)
			}
			if apiVersion, ok := config["apiVersion"].(string); ok {
				options["apiVersion"] = engine.NewStringValue(apiVersion)
			}

		case "anthropic":
			if apiVersion, ok := config["anthropicVersion"].(string); ok {
				options["anthropicVersion"] = engine.NewStringValue(apiVersion)
			}

		case "gemini":
			if location, ok := config["location"].(string); ok {
				options["location"] = engine.NewStringValue(location)
			}
			if projectID, ok := config["projectID"].(string); ok {
				options["projectID"] = engine.NewStringValue(projectID)
			}
		}

		return engine.NewObjectValue(options), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper function to convert ProviderMetadata to ScriptValue
func convertProviderMetadataToScriptValue(metadata provider.ProviderMetadata) engine.ScriptValue {
	// Get capabilities
	capabilities := metadata.GetCapabilities()
	capMap := make(map[string]engine.ScriptValue)
	for _, cap := range capabilities {
		switch cap {
		case provider.CapabilityStreaming:
			capMap["streaming"] = engine.NewBoolValue(true)
		case provider.CapabilityFunctionCalling:
			capMap["functionCalling"] = engine.NewBoolValue(true)
		case provider.CapabilityVision:
			capMap["vision"] = engine.NewBoolValue(true)
		case provider.CapabilityEmbeddings:
			capMap["embeddings"] = engine.NewBoolValue(true)
		case provider.CapabilityStructuredOutput:
			capMap["structured"] = engine.NewBoolValue(true)
		}
	}

	// Get constraints
	constraints := metadata.GetConstraints()

	return engine.NewObjectValue(map[string]engine.ScriptValue{
		"name":         engine.NewStringValue(metadata.Name()),
		"description":  engine.NewStringValue(metadata.Description()),
		"capabilities": engine.NewObjectValue(capMap),
		"constraints": engine.NewObjectValue(map[string]engine.ScriptValue{
			"maxBatchSize":    engine.NewNumberValue(float64(constraints.MaxBatchSize)),
			"maxConcurrency":  engine.NewNumberValue(float64(constraints.MaxConcurrency)),
			"rateLimit":       engine.NewNumberValue(0), // TODO: Extract rate limit value when available
			"minRequestDelay": engine.NewNumberValue(constraints.MinRequestDelay.Seconds()),
			"maxRetries":      engine.NewNumberValue(float64(constraints.MaxRetries)),
		}),
	})
}
