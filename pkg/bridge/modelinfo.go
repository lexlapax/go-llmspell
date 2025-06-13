// ABOUTME: Model info bridge providing access to go-llms ModelRegistry for LLM model discovery
// ABOUTME: Wraps go-llms model registry functionality without reimplementing

package bridge

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ModelInfoBridge provides access to LLM model information via go-llms ModelRegistry
type ModelInfoBridge struct {
	mu          sync.RWMutex
	registries  map[string]ModelRegistry
	initialized bool
}

// NewModelInfoBridge creates a new model info bridge
func NewModelInfoBridge() *ModelInfoBridge {
	return &ModelInfoBridge{
		registries: make(map[string]ModelRegistry),
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
		Description: "Provides access to go-llms ModelRegistry for model discovery",
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

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *ModelInfoBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *ModelInfoBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *ModelInfoBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "registerModelRegistry",
			Description: "Register a model registry",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Registry name", Required: true},
				{Name: "registry", Type: "ModelRegistry", Description: "Model registry instance", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "listModels",
			Description: "List all models from all registries",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "listModelsByRegistry",
			Description: "List models from a specific registry",
			Parameters: []engine.ParameterInfo{
				{Name: "registryName", Type: "string", Description: "Registry name", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getModel",
			Description: "Get a specific model by ID",
			Parameters: []engine.ParameterInfo{
				{Name: "registryName", Type: "string", Description: "Registry name", Required: true},
				{Name: "modelID", Type: "string", Description: "Model ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "listRegistries",
			Description: "List all registered model registries",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *ModelInfoBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"ModelRegistry": {
			GoType:     "ModelRegistry",
			ScriptType: "object",
		},
		"Model": {
			GoType:     "Model",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *ModelInfoBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions
func (b *ModelInfoBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "modelinfo",
			Actions:     []string{"read"},
			Description: "Access to model information",
		},
	}
}

// RegisterModelRegistry registers a model registry
func (b *ModelInfoBridge) RegisterModelRegistry(name string, registry ModelRegistry) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.registries[name] = registry
	return nil
}

// ListRegistries returns all registered registry names
func (b *ModelInfoBridge) ListRegistries() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.registries))
	for name := range b.registries {
		names = append(names, name)
	}
	return names
}

// GetRegistry returns a specific registry
func (b *ModelInfoBridge) GetRegistry(name string) ModelRegistry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.registries[name]
}
