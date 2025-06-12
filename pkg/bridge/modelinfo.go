// ABOUTME: Model info bridge providing LLM model discovery, metadata, and inventory management
// ABOUTME: Supports provider-specific model fetchers, caching, and filtering capabilities

package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ModelInfo represents information about an LLM model
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Description  string   `json:"description,omitempty"`
	MaxTokens    int      `json:"max_tokens"`
	Capabilities []string `json:"capabilities,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

// ModelFilter represents criteria for filtering models
type ModelFilter struct {
	Provider     string   `json:"provider,omitempty"`
	MinTokens    int      `json:"min_tokens,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ModelInfoProvider interface for model information sources
type ModelInfoProvider interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
	GetModel(ctx context.Context, modelID string) (*ModelInfo, error)
}

// modelCache represents cached model information
type modelCache struct {
	models    []ModelInfo
	timestamp time.Time
}

// ModelInfoBridge provides access to LLM model information
type ModelInfoBridge struct {
	mu          sync.RWMutex
	providers   map[string]ModelInfoProvider
	cache       map[string]*modelCache
	cacheTTL    time.Duration
	initialized bool
}

// NewModelInfoBridge creates a new model info bridge
func NewModelInfoBridge() *ModelInfoBridge {
	return &ModelInfoBridge{
		providers: make(map[string]ModelInfoProvider),
		cache:     make(map[string]*modelCache),
		cacheTTL:  5 * time.Minute, // Default cache TTL
	}
}

// GetID returns the bridge ID
func (b *ModelInfoBridge) GetID() string {
	return "modelinfo"
}

// GetMetadata returns bridge metadata
func (b *ModelInfoBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Model Info Bridge",
		Version:     "1.0.0",
		Description: "Provides LLM model discovery and metadata",
		Author:      "go-llmspell",
	}
}

// Initialize initializes the bridge
func (b *ModelInfoBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup performs cleanup
func (b *ModelInfoBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clear cache
	b.cache = make(map[string]*modelCache)
	b.initialized = false

	return nil
}

// IsInitialized checks if bridge is initialized
func (b *ModelInfoBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterProvider registers a model info provider
func (b *ModelInfoBridge) RegisterProvider(name string, provider ModelInfoProvider) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	b.providers[name] = provider
	return nil
}

// ListProviders returns list of registered providers
func (b *ModelInfoBridge) ListProviders() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]string, 0, len(b.providers))
	for name := range b.providers {
		providers = append(providers, name)
	}
	return providers
}

// ListAllModels returns all models from all providers
func (b *ModelInfoBridge) ListAllModels(ctx context.Context) ([]ModelInfo, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	// Get a copy of providers to avoid holding lock during provider calls
	b.mu.RLock()
	providersCopy := make(map[string]ModelInfoProvider)
	for name, provider := range b.providers {
		providersCopy[name] = provider
	}
	b.mu.RUnlock()

	var allModels []ModelInfo

	for providerName, provider := range providersCopy {
		models, err := b.getModelsFromProvider(ctx, providerName, provider)
		if err != nil {
			continue // Skip failed providers
		}
		allModels = append(allModels, models...)
	}

	return allModels, nil
}

// ListModelsByProvider returns models for a specific provider
func (b *ModelInfoBridge) ListModelsByProvider(ctx context.Context, providerName string) ([]ModelInfo, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	b.mu.RLock()
	provider, exists := b.providers[providerName]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return b.getModelsFromProvider(ctx, providerName, provider)
}

// GetModel returns information about a specific model
func (b *ModelInfoBridge) GetModel(ctx context.Context, providerName, modelID string) (*ModelInfo, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	b.mu.RLock()
	provider, exists := b.providers[providerName]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return provider.GetModel(ctx, modelID)
}

// FilterModels returns models matching the filter criteria
func (b *ModelInfoBridge) FilterModels(ctx context.Context, filter ModelFilter) ([]ModelInfo, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	allModels, err := b.ListAllModels(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []ModelInfo

	for _, model := range allModels {
		if b.matchesFilter(model, filter) {
			filtered = append(filtered, model)
		}
	}

	return filtered, nil
}

// SetCacheTTL sets the cache time-to-live
func (b *ModelInfoBridge) SetCacheTTL(ttl time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cacheTTL = ttl
}

// ClearCache clears the model cache
func (b *ModelInfoBridge) ClearCache() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache = make(map[string]*modelCache)
}

// getModelsFromProvider gets models from provider with caching
func (b *ModelInfoBridge) getModelsFromProvider(ctx context.Context, providerName string, provider ModelInfoProvider) ([]ModelInfo, error) {
	// Check cache first
	b.mu.RLock()
	cached, exists := b.cache[providerName]
	cacheTTL := b.cacheTTL
	b.mu.RUnlock()

	if exists && time.Since(cached.timestamp) < cacheTTL {
		return cached.models, nil
	}

	// Fetch from provider
	models, err := provider.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	b.mu.Lock()
	b.cache[providerName] = &modelCache{
		models:    models,
		timestamp: time.Now(),
	}
	b.mu.Unlock()

	return models, nil
}

// matchesFilter checks if a model matches the filter criteria
func (b *ModelInfoBridge) matchesFilter(model ModelInfo, filter ModelFilter) bool {
	// Check provider
	if filter.Provider != "" && model.Provider != filter.Provider {
		return false
	}

	// Check token limits
	if filter.MinTokens > 0 && model.MaxTokens < filter.MinTokens {
		return false
	}
	if filter.MaxTokens > 0 && model.MaxTokens > filter.MaxTokens {
		return false
	}

	// Check capabilities
	if len(filter.Capabilities) > 0 {
		modelCaps := make(map[string]bool)
		for _, cap := range model.Capabilities {
			modelCaps[cap] = true
		}

		for _, requiredCap := range filter.Capabilities {
			if !modelCaps[requiredCap] {
				return false
			}
		}
	}

	return true
}

// RegisterWithEngine registers the bridge with a script engine
func (b *ModelInfoBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *ModelInfoBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "list_all_models",
			Description: "List all available models from all providers",
			Parameters: []engine.ParameterInfo{
				{Name: "ctx", Type: "context", Required: true},
			},
			ReturnType: "[]ModelInfo",
		},
		{
			Name:        "list_models_by_provider",
			Description: "List models from a specific provider",
			Parameters: []engine.ParameterInfo{
				{Name: "ctx", Type: "context", Required: true},
				{Name: "provider", Type: "string", Required: true},
			},
			ReturnType: "[]ModelInfo",
		},
		{
			Name:        "get_model",
			Description: "Get information about a specific model",
			Parameters: []engine.ParameterInfo{
				{Name: "ctx", Type: "context", Required: true},
				{Name: "provider", Type: "string", Required: true},
				{Name: "model_id", Type: "string", Required: true},
			},
			ReturnType: "ModelInfo",
		},
		{
			Name:        "filter_models",
			Description: "Filter models based on criteria",
			Parameters: []engine.ParameterInfo{
				{Name: "ctx", Type: "context", Required: true},
				{Name: "filter", Type: "ModelFilter", Required: true},
			},
			ReturnType: "[]ModelInfo",
		},
		{
			Name:        "list_providers",
			Description: "List all registered model providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "[]string",
		},
		{
			Name:        "clear_cache",
			Description: "Clear the model information cache",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "set_cache_ttl",
			Description: "Set cache time-to-live duration",
			Parameters: []engine.ParameterInfo{
				{Name: "ttl_seconds", Type: "number", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// ValidateMethod validates method parameters
func (b *ModelInfoBridge) ValidateMethod(name string, args []interface{}) error {
	switch name {
	case "list_all_models":
		if len(args) != 1 {
			return errors.New("list_all_models requires 1 argument")
		}
	case "list_models_by_provider":
		if len(args) != 2 {
			return errors.New("list_models_by_provider requires 2 arguments")
		}
	case "get_model":
		if len(args) != 3 {
			return errors.New("get_model requires 3 arguments")
		}
	case "filter_models":
		if len(args) != 2 {
			return errors.New("filter_models requires 2 arguments")
		}
	case "list_providers":
		if len(args) != 0 {
			return errors.New("list_providers requires no arguments")
		}
	case "clear_cache":
		if len(args) != 0 {
			return errors.New("clear_cache requires no arguments")
		}
	case "set_cache_ttl":
		if len(args) != 1 {
			return errors.New("set_cache_ttl requires 1 argument")
		}
	default:
		return fmt.Errorf("unknown method: %s", name)
	}
	return nil
}

// TypeMappings returns type conversion hints for engines
func (b *ModelInfoBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"ModelInfo": {
			ScriptType: "object",
			GoType:     "ModelInfo",
		},
		"[]ModelInfo": {
			ScriptType: "array",
			GoType:     "[]ModelInfo",
		},
		"ModelFilter": {
			ScriptType: "object",
			GoType:     "ModelFilter",
		},
	}
}

// RequiredPermissions returns required permissions
func (b *ModelInfoBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "model",
			Actions:     []string{"list", "get", "filter"},
			Description: "Read model information",
		},
	}
}
