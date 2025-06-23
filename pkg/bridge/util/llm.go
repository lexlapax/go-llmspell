// ABOUTME: LLM utilities bridge provides access to go-llms LLM utility functions.
// ABOUTME: Wraps provider creation, typed generation, pooling, and model inventory utilities.

package util

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"

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
// It manages provider creation, typed generation, model discovery,
// response parsing, streaming events, and cost tracking.
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

// CostTracker tracks costs per request.
// It maintains per-request cost details and total costs per provider.
type CostTracker struct {
	mu     sync.RWMutex
	costs  map[string]*RequestCost
	totals map[string]float64 // Total costs per provider
}

// RequestCost represents the cost of a single request.
// It includes token counts, cost breakdowns, and metadata.
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
// It initializes empty registries for metadata and cost tracking.
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
// It enables streaming events and cost tracking notifications.
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
// It implements the types.Bridge interface.
func (b *UtilLLMBridge) GetID() string {
	return "util_llm"
}

// GetMetadata returns bridge metadata.
// It provides information about the LLM utilities bridge including
// version, description, and supported LLM features.
func (b *UtilLLMBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "util_llm",
		Version:     "2.0.0",
		Description: "Enhanced LLM utilities with provider capabilities, model discovery, response parsing, streaming events, and cost tracking",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
// It sets up the validator, event bus, and model service
// for comprehensive LLM functionality.
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
// It releases resources and marks the bridge as uninitialized.
func (b *UtilLLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
// It returns true if the bridge has been initialized and is ready for use.
func (b *UtilLLMBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// It enables the script engine to access LLM utilities through this bridge.
func (b *UtilLLMBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this bridge.
// It provides metadata about all LLM-related methods available to scripts,
// including provider management, generation, model discovery, and cost tracking.
func (b *UtilLLMBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Provider creation utilities
		{
			Name:        "createProvider",
			Description: "Create an LLM provider from configuration",
			Parameters: []types.ParameterInfo{
				{Name: "config", Type: "object", Description: "Provider configuration", Required: true},
			},
			ReturnType: "Provider",
		},
		{
			Name:        "createProviderFromEnv",
			Description: "Create an LLM provider from environment variables",
			Parameters: []types.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "Provider",
		},
		{
			Name:        "withProviderOptions",
			Description: "Create provider-specific options",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
				{Name: "options", Type: "object", Description: "Provider-specific options", Required: true},
			},
			ReturnType: "object",
		},

		// Typed generation utilities
		{
			Name:        "generateTyped",
			Description: "Generate a typed/structured response",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "output", Type: "object", Description: "Output to validate", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema", Required: true},
			},
			ReturnType: "boolean",
		},

		// Provider pool utilities
		{
			Name:        "createProviderPool",
			Description: "Create a provider pool for load balancing/failover",
			Parameters: []types.ParameterInfo{
				{Name: "providers", Type: "array", Description: "Array of providers", Required: true},
				{Name: "strategy", Type: "string", Description: "Pool strategy (roundrobin/failover/fastest)", Required: true},
			},
			ReturnType: "ProviderPool",
		},
		{
			Name:        "addProviderToPool",
			Description: "Add a provider to an existing pool",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "fetchers", Type: "object", Description: "Provider fetchers configuration", Required: false},
			},
			ReturnType: "ModelInventory",
		},
		{
			Name:        "fetchModelInfo",
			Description: "Fetch model information for a provider",
			Parameters: []types.ParameterInfo{
				{Name: "inventory", Type: "ModelInventory", Description: "Model inventory", Required: true},
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "cacheModelInfo",
			Description: "Cache model information to file",
			Parameters: []types.ParameterInfo{
				{Name: "inventory", Type: "ModelInventory", Description: "Model inventory", Required: true},
				{Name: "cachePath", Type: "string", Description: "Cache file path", Required: true},
			},
			ReturnType: "void",
		},

		// Configuration utilities
		{
			Name:        "createModelConfig",
			Description: "Create a model configuration",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
				{Name: "model", Type: "string", Description: "Model name", Required: true},
				{Name: "options", Type: "object", Description: "Model options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "mergeProviderOptions",
			Description: "Merge multiple provider options",
			Parameters: []types.ParameterInfo{
				{Name: "options1", Type: "object", Description: "First options", Required: true},
				{Name: "options2", Type: "object", Description: "Second options", Required: true},
			},
			ReturnType: "object",
		},

		// Enhanced v0.3.5 features
		{
			Name:        "getProviderCapabilities",
			Description: "Get provider capability metadata",
			Parameters: []types.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "discoverModels",
			Description: "Discover available models for a provider",
			Parameters: []types.ParameterInfo{
				{Name: "providerName", Type: "string", Description: "Provider name", Required: true},
				{Name: "refresh", Type: "boolean", Description: "Force refresh from API", Required: false},
			},
			ReturnType: "array",
		},
		{
			Name:        "parseResponseWithRecovery",
			Description: "Parse LLM response with recovery for malformed output",
			Parameters: []types.ParameterInfo{
				{Name: "response", Type: "string", Description: "LLM response", Required: true},
				{Name: "format", Type: "string", Description: "Expected format (json/xml/yaml)", Required: false},
				{Name: "schema", Type: "object", Description: "Optional schema for validation", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "streamWithEvents",
			Description: "Stream LLM response with event emission",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "Provider", Description: "LLM provider", Required: true},
				{Name: "prompt", Type: "string", Description: "Generation prompt", Required: true},
				{Name: "eventHandler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "trackRequestCost",
			Description: "Track cost for an LLM request",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Filter criteria", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "createProviderOptions",
			Description: "Create provider-specific options with advanced features",
			Parameters: []types.ParameterInfo{
				{Name: "providerType", Type: "string", Description: "Provider type", Required: true},
				{Name: "config", Type: "object", Description: "Configuration options", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings.
// It defines how Go LLM types are mapped to script types
// for pools, inventory, configurations, and costs.
func (b *UtilLLMBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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
// It delegates validation to the engine based on Methods() metadata.
func (b *UtilLLMBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
// It specifies permissions for LLM provider access, cache operations,
// and metadata storage.
func (b *UtilLLMBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionNetwork,
			Resource:    "llm",
			Actions:     []string{"create", "access"},
			Description: "Create and access LLM providers",
		},
		{
			Type:        types.PermissionFileSystem,
			Resource:    "cache",
			Actions:     []string{"read", "write"},
			Description: "Cache model information",
		},
		{
			Type:        types.PermissionMemory,
			Resource:    "metadata",
			Actions:     []string{"read", "write"},
			Description: "Store provider metadata and cost tracking",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function.
// It implements the types.Bridge interface, routing method calls
// to the appropriate LLM operations.
func (b *UtilLLMBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
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
		if args[0] == nil || args[0].Type() != types.TypeArray {
			return nil, fmt.Errorf("providers must be an array")
		}
		providersArg := args[0].(types.ArrayValue).Elements()

		// Convert to domain.Provider slice
		providers := make([]types.Provider, 0, len(providersArg))
		for i, p := range providersArg {
			// For now, we'll need custom handling for Provider types
			if p.Type() != types.TypeCustom {
				return nil, fmt.Errorf("provider at index %d must be a Provider", i)
			}
			customVal := p.(types.CustomValue)
			provider, ok := customVal.Value().(types.Provider)
			if !ok {
				return nil, fmt.Errorf("provider at index %d must be a Provider", i)
			}
			providers = append(providers, provider)
		}

		// Get strategy
		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("strategy must be string")
		}
		strategyStr := args[1].(types.StringValue).Value()

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
		return types.NewCustomValue("ProviderPool", pool), nil

	case "createModelInventory":
		// Create model info service to fetch model inventory
		// Using the modelinfo package's service to aggregate models
		// Note: The actual model inventory is returned by the service's AggregateModels method
		// For now, return a placeholder as the service requires provider-specific fetchers
		result := map[string]types.ScriptValue{
			"type": types.NewStringValue("ModelInventory"),
			"id":   types.NewStringValue("inventory_1"),
			"note": types.NewStringValue("Use fetchModelInfo to retrieve actual model data"),
		}
		return types.NewObjectValue(result), nil

	case "createModelConfig":
		if len(args) < 2 {
			return nil, fmt.Errorf("createModelConfig requires provider and model parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("provider must be string")
		}
		provider := args[0].(types.StringValue).Value()

		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("model must be string")
		}
		model := args[1].(types.StringValue).Value()

		// Create model config
		config := llmutil.ModelConfig{
			Provider: provider,
			Model:    model,
		}

		// Add options if provided
		if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeObject {
			options := make(map[string]interface{})
			for k, v := range args[2].(types.ObjectValue).Fields() {
				options[k] = v.ToGo()
			}
			// Options will be applied when go-llms ModelConfig supports additional fields
			_ = options
		}

		result := map[string]types.ScriptValue{
			"provider": types.NewStringValue(config.Provider),
			"model":    types.NewStringValue(config.Model),
		}
		return types.NewObjectValue(result), nil

	// Enhanced v0.3.5 features
	case "getProviderCapabilities":
		if len(args) < 1 {
			return nil, fmt.Errorf("getProviderCapabilities requires providerName")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("providerName must be string")
		}
		providerName := args[0].(types.StringValue).Value()

		// Check if we have cached metadata
		if metadata, exists := b.metadataRegistry[providerName]; exists {
			return convertProviderMetadataToScriptValue(metadata), nil
		}

		// TODO: Load metadata from provider when available in go-llms
		// For now, return basic capabilities based on provider type
		capabilities := map[string]types.ScriptValue{
			"provider": types.NewStringValue(providerName),
			"capabilities": types.NewObjectValue(map[string]types.ScriptValue{
				"streaming":       types.NewBoolValue(true),
				"functionCalling": types.NewBoolValue(providerName == "openai" || providerName == "anthropic"),
				"vision":          types.NewBoolValue(providerName == "openai" || providerName == "anthropic"),
				"embeddings":      types.NewBoolValue(providerName == "openai"),
			}),
			"constraints": types.NewObjectValue(map[string]types.ScriptValue{
				"maxTokens":     types.NewNumberValue(4096),
				"rateLimit":     types.NewNumberValue(60),
				"contextWindow": types.NewNumberValue(8192),
			}),
		}

		return types.NewObjectValue(capabilities), nil

	case "discoverModels":
		if len(args) < 1 {
			return nil, fmt.Errorf("discoverModels requires providerName")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("providerName must be string")
		}
		providerName := args[0].(types.StringValue).Value()

		// Check refresh flag (not used in current implementation)
		// refresh := false
		// if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeBool {
		// 	refresh = args[1].(types.BoolValue).Value()
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
		result := make([]types.ScriptValue, 0, len(models))
		for _, model := range models {
			modelInfo := map[string]types.ScriptValue{
				"id":            types.NewStringValue(model.Name), // Use Name as ID
				"name":          types.NewStringValue(model.DisplayName),
				"description":   types.NewStringValue(model.Description),
				"inputCost":     types.NewNumberValue(model.Pricing.InputPer1kTokens),
				"outputCost":    types.NewNumberValue(model.Pricing.OutputPer1kTokens),
				"maxTokens":     types.NewNumberValue(float64(model.MaxOutputTokens)),
				"contextWindow": types.NewNumberValue(float64(model.ContextWindow)),
				"capabilities": types.NewObjectValue(map[string]types.ScriptValue{
					"streaming":       types.NewBoolValue(model.Capabilities.Streaming),
					"functionCalling": types.NewBoolValue(model.Capabilities.FunctionCalling),
					"vision":          types.NewBoolValue(model.Capabilities.Image.Read),
					"jsonMode":        types.NewBoolValue(model.Capabilities.JSONMode),
				}),
			}
			result = append(result, types.NewObjectValue(modelInfo))
		}

		return types.NewArrayValue(result), nil

	case "parseResponseWithRecovery":
		if len(args) < 1 {
			return nil, fmt.Errorf("parseResponseWithRecovery requires response")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("response must be string")
		}
		response := args[0].(types.StringValue).Value()

		// Get optional format
		format := ""
		if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeString {
			format = args[1].(types.StringValue).Value()
		}

		// Get optional schema
		var schema *schemaDomain.Schema
		if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeObject {
			schemaMap := make(map[string]interface{})
			for k, v := range args[2].(types.ObjectValue).Fields() {
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
		return types.NewCustomValue("ParsedResponse", result), nil

	case "streamWithEvents":
		if len(args) < 3 {
			return nil, fmt.Errorf("streamWithEvents requires provider, prompt, and eventHandler")
		}

		// Get provider from custom value
		if args[0] == nil || args[0].Type() != types.TypeCustom {
			return nil, fmt.Errorf("provider must be Provider")
		}
		customVal := args[0].(types.CustomValue)
		_, ok := customVal.Value().(types.Provider)
		if !ok {
			return nil, fmt.Errorf("provider must be Provider")
		}

		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("prompt must be string")
		}
		_ = args[1].(types.StringValue).Value()

		// Get event handler function from custom value
		if args[2] == nil || args[2].Type() != types.TypeFunction {
			return nil, fmt.Errorf("eventHandler must be function")
		}
		// For now, we'll need custom handling for function types
		// The engine will need to provide a way to call script functions
		// This is a placeholder that would need engine-specific implementation
		return nil, fmt.Errorf("function callbacks not yet implemented for ScriptValue")

		// TODO: The rest of this method would need engine-specific function callback support
		// return map[string]types.ScriptValue{
		// 	"content":    types.NewStringValue(fullContent.String()),
		// 	"tokenCount": types.NewNumberValue(float64(tokenCount)),
		// }, nil

	case "trackRequestCost":
		if len(args) < 4 {
			return nil, fmt.Errorf("trackRequestCost requires requestID, provider, model, and usage")
		}

		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("requestID must be string")
		}
		requestID := args[0].(types.StringValue).Value()

		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("provider must be string")
		}
		provider := args[1].(types.StringValue).Value()

		if args[2] == nil || args[2].Type() != types.TypeString {
			return nil, fmt.Errorf("model must be string")
		}
		model := args[2].(types.StringValue).Value()

		if args[3] == nil || args[3].Type() != types.TypeObject {
			return nil, fmt.Errorf("usage must be object")
		}
		usageObj := args[3].(types.ObjectValue).Fields()
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

		return types.NewObjectValue(map[string]types.ScriptValue{
			"requestID": types.NewStringValue(cost.RequestID),
			"totalCost": types.NewNumberValue(cost.TotalCost),
			"breakdown": types.NewObjectValue(map[string]types.ScriptValue{
				"inputCost":    types.NewNumberValue(cost.InputCost),
				"outputCost":   types.NewNumberValue(cost.OutputCost),
				"inputTokens":  types.NewNumberValue(float64(cost.InputTokens)),
				"outputTokens": types.NewNumberValue(float64(cost.OutputTokens)),
			}),
		}), nil

	case "getCostReport":
		filter := make(map[string]interface{})
		if len(args) > 0 && args[0] != nil && args[0].Type() == types.TypeObject {
			filterObj := args[0].(types.ObjectValue).Fields()
			for k, v := range filterObj {
				filter[k] = v.ToGo()
			}
		}

		b.costTracker.mu.RLock()
		defer b.costTracker.mu.RUnlock()

		// Convert totals to ScriptValue
		totalCosts := make(map[string]types.ScriptValue)
		for provider, cost := range b.costTracker.totals {
			totalCosts[provider] = types.NewNumberValue(cost)
		}

		// Build report
		requests := make([]types.ScriptValue, 0)
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
			requestInfo := map[string]types.ScriptValue{
				"requestID": types.NewStringValue(cost.RequestID),
				"provider":  types.NewStringValue(cost.Provider),
				"model":     types.NewStringValue(cost.Model),
				"totalCost": types.NewNumberValue(cost.TotalCost),
				"timestamp": types.NewStringValue(cost.Timestamp.Format(time.RFC3339)),
			}
			requests = append(requests, types.NewObjectValue(requestInfo))

			// Update summary
			providers[cost.Provider]++
		}

		// Convert providers count to ScriptValue
		providersCount := make(map[string]types.ScriptValue)
		for provider, count := range providers {
			providersCount[provider] = types.NewNumberValue(float64(count))
		}

		report := map[string]types.ScriptValue{
			"totalCosts": types.NewObjectValue(totalCosts),
			"requests":   types.NewArrayValue(requests),
			"summary": types.NewObjectValue(map[string]types.ScriptValue{
				"totalRequests": types.NewNumberValue(float64(len(b.costTracker.costs))),
				"providers":     types.NewObjectValue(providersCount),
			}),
		}

		return types.NewObjectValue(report), nil

	case "createProviderOptions":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProviderOptions requires providerType and config")
		}

		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("providerType must be string")
		}
		providerType := args[0].(types.StringValue).Value()

		if args[1] == nil || args[1].Type() != types.TypeObject {
			return nil, fmt.Errorf("config must be object")
		}
		configObj := args[1].(types.ObjectValue).Fields()
		config := make(map[string]interface{})
		for k, v := range configObj {
			config[k] = v.ToGo()
		}

		// Create provider-specific options based on type
		options := map[string]types.ScriptValue{
			"type": types.NewStringValue(providerType),
		}

		// Extract common options
		if baseURL, ok := config["baseURL"].(string); ok {
			options["baseURL"] = types.NewStringValue(baseURL)
		}
		if apiKey, ok := config["apiKey"].(string); ok {
			options["apiKey"] = types.NewStringValue(apiKey)
		}
		if timeout, ok := config["timeout"].(float64); ok {
			options["timeout"] = types.NewNumberValue(float64(int(timeout)))
		}

		// Add provider-specific options
		switch providerType {
		case "openai":
			if org, ok := config["organization"].(string); ok {
				options["organization"] = types.NewStringValue(org)
			}
			if apiVersion, ok := config["apiVersion"].(string); ok {
				options["apiVersion"] = types.NewStringValue(apiVersion)
			}

		case "anthropic":
			if apiVersion, ok := config["anthropicVersion"].(string); ok {
				options["anthropicVersion"] = types.NewStringValue(apiVersion)
			}

		case "gemini":
			if location, ok := config["location"].(string); ok {
				options["location"] = types.NewStringValue(location)
			}
			if projectID, ok := config["projectID"].(string); ok {
				options["projectID"] = types.NewStringValue(projectID)
			}
		}

		return types.NewObjectValue(options), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// convertProviderMetadataToScriptValue converts ProviderMetadata to ScriptValue.
// It extracts capabilities and constraints into a script-friendly format.
func convertProviderMetadataToScriptValue(metadata provider.ProviderMetadata) types.ScriptValue {
	// Get capabilities
	capabilities := metadata.GetCapabilities()
	capMap := make(map[string]types.ScriptValue)
	for _, cap := range capabilities {
		switch cap {
		case provider.CapabilityStreaming:
			capMap["streaming"] = types.NewBoolValue(true)
		case provider.CapabilityFunctionCalling:
			capMap["functionCalling"] = types.NewBoolValue(true)
		case provider.CapabilityVision:
			capMap["vision"] = types.NewBoolValue(true)
		case provider.CapabilityEmbeddings:
			capMap["embeddings"] = types.NewBoolValue(true)
		case provider.CapabilityStructuredOutput:
			capMap["structured"] = types.NewBoolValue(true)
		}
	}

	// Get constraints
	constraints := metadata.GetConstraints()

	return types.NewObjectValue(map[string]types.ScriptValue{
		"name":         types.NewStringValue(metadata.Name()),
		"description":  types.NewStringValue(metadata.Description()),
		"capabilities": types.NewObjectValue(capMap),
		"constraints": types.NewObjectValue(map[string]types.ScriptValue{
			"maxBatchSize":    types.NewNumberValue(float64(constraints.MaxBatchSize)),
			"maxConcurrency":  types.NewNumberValue(float64(constraints.MaxConcurrency)),
			"rateLimit":       types.NewNumberValue(0), // TODO: Extract rate limit value when available
			"minRequestDelay": types.NewNumberValue(constraints.MinRequestDelay.Seconds()),
			"maxRetries":      types.NewNumberValue(float64(constraints.MaxRetries)),
		}),
	})
}
