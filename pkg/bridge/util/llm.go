// ABOUTME: LLM utilities bridge provides access to go-llms LLM utility functions.
// ABOUTME: Wraps provider creation, typed generation, pooling, and model inventory utilities.

package util

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// UtilLLMBridge provides script access to go-llms LLM utilities.
type UtilLLMBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilLLMBridge creates a new LLM utilities bridge.
func NewUtilLLMBridge() *UtilLLMBridge {
	return &UtilLLMBridge{}
}

// GetID returns the bridge identifier.
func (b *UtilLLMBridge) GetID() string {
	return "util_llm"
}

// GetMetadata returns bridge metadata.
func (b *UtilLLMBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_llm",
		Version:     "1.0.0",
		Description: "LLM utilities bridge for provider creation, typed generation, and pooling",
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
	}
}

// The actual method implementations would be provided by the script engine
// which would call the appropriate go-llms/pkg/util/llmutil functions.
// For example:
// - createProvider would call llmutil.CreateProvider
// - generateTyped would call llmutil.GenerateTyped
// - createProviderPool would create a new llmutil.ProviderPool
// etc.
