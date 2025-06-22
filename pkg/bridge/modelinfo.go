// ABOUTME: Model info bridge providing access to go-llms ModelRegistry for LLM model discovery
// ABOUTME: Wraps go-llms model registry functionality without reimplementing

package bridge

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for model info functionality
	"fmt"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// ModelInfoBridge provides access to LLM model information via go-llms ModelRegistry.
// It wraps go-llms model discovery and registry functionality, exposing model
// metadata, capabilities, and pricing information to script engines.
type ModelInfoBridge struct {
	mu          sync.RWMutex
	registries  map[string]llmdomain.ModelRegistry
	initialized bool
}

// NewModelInfoBridge creates a new model info bridge.
// The bridge starts uninitialized with an empty registry map.
func NewModelInfoBridge() *ModelInfoBridge {
	return &ModelInfoBridge{
		registries: make(map[string]llmdomain.ModelRegistry),
	}
}

// GetID returns the bridge ID.
// Always returns "modelinfo" for this bridge.
func (b *ModelInfoBridge) GetID() string {
	return "modelinfo"
}

// GetMetadata returns bridge metadata.
// Provides information about the bridge including name, version,
// and description for documentation and discovery.
func (b *ModelInfoBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Model Info Bridge",
		Version:     "1.0.0",
		Description: "Provides access to go-llms ModelRegistry for model discovery",
		Author:      "go-llmspell",
	}
}

// Initialize initializes the bridge.
// Currently performs minimal initialization. Can be extended to
// pre-load model registries or connect to external services.
func (b *ModelInfoBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup performs cleanup.
// Marks the bridge as uninitialized. Registries are preserved
// and can be used after re-initialization.
func (b *ModelInfoBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
// Thread-safe check of initialization status.
func (b *ModelInfoBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
// Delegates to the engine's RegisterBridge method for proper integration.
func (b *ModelInfoBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
// Defines the script-accessible API for model information including
// registry management and model queries.
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

// TypeMappings returns type conversion mappings.
// Maps go-llms types to script types for proper type conversion
// during method execution.
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

// ValidateMethod validates method calls.
// Currently delegates validation to the engine based on Methods() metadata.
// Can be extended for custom validation logic.
func (b *ModelInfoBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
// Defines that scripts need read access to model information
// for security sandboxing.
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

// RegisterModelRegistry registers a model registry.
// Associates a named registry with the bridge for later queries.
// Thread-safe for concurrent registry additions.
func (b *ModelInfoBridge) RegisterModelRegistry(name string, registry llmdomain.ModelRegistry) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.registries[name] = registry
	return nil
}

// ListRegistries returns all registered registry names.
// Returns a copy of registry names to prevent external modification.
// Thread-safe for concurrent access.
func (b *ModelInfoBridge) ListRegistries() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.registries))
	for name := range b.registries {
		names = append(names, name)
	}
	return names
}

// GetRegistry returns a specific registry.
// Returns nil if registry not found. Thread-safe for concurrent access.
func (b *ModelInfoBridge) GetRegistry(name string) llmdomain.ModelRegistry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.registries[name]
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function.
// Handles all script-callable methods including model inventory fetching,
// registry management, and model queries. Returns script-compatible values.
func (b *ModelInfoBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "fetchModelInventory":
		// Create model info service and fetch inventory
		service := modelinfo.NewModelInfoServiceFunc()
		inventory, err := service.AggregateModels()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch model inventory: %w", err)
		}

		// Convert to script-friendly format
		metadata := map[string]engine.ScriptValue{
			"version":       engine.NewStringValue(inventory.Metadata.Version),
			"lastUpdated":   engine.NewStringValue(inventory.Metadata.LastUpdated),
			"description":   engine.NewStringValue(inventory.Metadata.Description),
			"schemaVersion": engine.NewStringValue(inventory.Metadata.SchemaVersion),
		}
		models := convertModelsToScriptValue(inventory.Models)

		result := map[string]engine.ScriptValue{
			"metadata": engine.NewObjectValue(metadata),
			"models":   models,
		}
		return engine.NewObjectValue(result), nil

	case "fetchProviderModels":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("fetchProviderModels requires provider parameter")
		}
		provider := args[0].(engine.StringValue).Value()

		// Create service and fetch models for specific provider
		service := modelinfo.NewModelInfoServiceFunc()
		inventory, err := service.AggregateModels()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch models: %w", err)
		}

		// Filter models by provider
		var providerModels []domain.Model
		for _, m := range inventory.Models {
			if m.Provider == provider {
				providerModels = append(providerModels, m)
			}
		}

		if len(providerModels) == 0 {
			return nil, fmt.Errorf("provider %s not found", provider)
		}

		return convertModelsToScriptValue(providerModels), nil

	case "listRegistries":
		registries := b.ListRegistries()
		values := make([]engine.ScriptValue, len(registries))
		for i, reg := range registries {
			values[i] = engine.NewStringValue(reg)
		}
		return engine.NewArrayValue(values), nil

	case "registerModelRegistry":
		if len(args) < 2 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("registerModelRegistry requires name and registry parameters")
		}
		name := args[0].(engine.StringValue).Value()
		// Note: registry parameter would need special handling as it's a Go type
		// For now, this would need to be passed as a custom value
		if args[1] == nil || args[1].Type() != engine.TypeCustom {
			return nil, fmt.Errorf("registry must be ModelRegistry")
		}
		customVal := args[1].(engine.CustomValue)
		registry, ok := customVal.Value().(llmdomain.ModelRegistry)
		if !ok {
			return nil, fmt.Errorf("registry must be ModelRegistry")
		}
		err := b.RegisterModelRegistry(name, registry)
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "listModels":
		// List all models from all registries
		var allModels []string
		for _, registry := range b.registries {
			models := registry.ListModels()
			allModels = append(allModels, models...)
		}
		values := make([]engine.ScriptValue, len(allModels))
		for i, model := range allModels {
			values[i] = engine.NewStringValue(model)
		}
		return engine.NewArrayValue(values), nil

	case "listModelsByRegistry":
		if len(args) < 1 || args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("listModelsByRegistry requires registryName parameter")
		}
		registryName := args[0].(engine.StringValue).Value()
		registry := b.GetRegistry(registryName)
		if registry == nil {
			return nil, fmt.Errorf("registry not found: %s", registryName)
		}
		models := registry.ListModels()
		values := make([]engine.ScriptValue, len(models))
		for i, model := range models {
			values[i] = engine.NewStringValue(model)
		}
		return engine.NewArrayValue(values), nil

	case "getModel":
		if len(args) < 2 || args[0] == nil || args[0].Type() != engine.TypeString ||
			args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("getModel requires registryName and modelID parameters")
		}
		registryName := args[0].(engine.StringValue).Value()
		modelID := args[1].(engine.StringValue).Value()
		registry := b.GetRegistry(registryName)
		if registry == nil {
			return nil, fmt.Errorf("registry not found: %s", registryName)
		}
		provider, err := registry.GetModel(modelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get model: %w", err)
		}
		// Return provider as a custom value wrapped in an object
		result := map[string]engine.ScriptValue{
			"provider": engine.NewCustomValue("Provider", provider),
			"modelID":  engine.NewStringValue(modelID),
		}
		return engine.NewObjectValue(result), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// convertModelsToScriptValue converts an array of domain.Model to ScriptValue.
// Transforms go-llms model structs into script-compatible array format.
func convertModelsToScriptValue(models []domain.Model) engine.ScriptValue {
	values := make([]engine.ScriptValue, len(models))
	for i, m := range models {
		values[i] = convertModelToScriptValue(m)
	}
	return engine.NewArrayValue(values)
}

// convertModelToScriptValue converts a single domain.Model to ScriptValue.
// Creates a comprehensive object representation including all model metadata,
// pricing information, and capability flags.
func convertModelToScriptValue(m domain.Model) engine.ScriptValue {
	pricingFields := map[string]engine.ScriptValue{
		"inputPer1kTokens":  engine.NewNumberValue(m.Pricing.InputPer1kTokens),
		"outputPer1kTokens": engine.NewNumberValue(m.Pricing.OutputPer1kTokens),
	}

	fields := map[string]engine.ScriptValue{
		"provider":         engine.NewStringValue(m.Provider),
		"name":             engine.NewStringValue(m.Name),
		"displayName":      engine.NewStringValue(m.DisplayName),
		"description":      engine.NewStringValue(m.Description),
		"documentationURL": engine.NewStringValue(m.DocumentationURL),
		"contextWindow":    engine.NewNumberValue(float64(m.ContextWindow)),
		"maxOutputTokens":  engine.NewNumberValue(float64(m.MaxOutputTokens)),
		"trainingCutoff":   engine.NewStringValue(m.TrainingCutoff),
		"modelFamily":      engine.NewStringValue(m.ModelFamily),
		"lastUpdated":      engine.NewStringValue(m.LastUpdated),
		"pricing":          engine.NewObjectValue(pricingFields),
		"capabilities":     convertCapabilitiesToScriptValue(m.Capabilities),
	}
	return engine.NewObjectValue(fields)
}


// convertCapabilitiesToScriptValue converts domain.Capabilities to ScriptValue.
// Creates nested object structure representing model capabilities across
// different modalities (text, image, audio, video, file) and features.
func convertCapabilitiesToScriptValue(c domain.Capabilities) engine.ScriptValue {
	textFields := map[string]engine.ScriptValue{
		"read":  engine.NewBoolValue(c.Text.Read),
		"write": engine.NewBoolValue(c.Text.Write),
	}
	imageFields := map[string]engine.ScriptValue{
		"read":  engine.NewBoolValue(c.Image.Read),
		"write": engine.NewBoolValue(c.Image.Write),
	}
	audioFields := map[string]engine.ScriptValue{
		"read":  engine.NewBoolValue(c.Audio.Read),
		"write": engine.NewBoolValue(c.Audio.Write),
	}
	videoFields := map[string]engine.ScriptValue{
		"read":  engine.NewBoolValue(c.Video.Read),
		"write": engine.NewBoolValue(c.Video.Write),
	}
	fileFields := map[string]engine.ScriptValue{
		"read":  engine.NewBoolValue(c.File.Read),
		"write": engine.NewBoolValue(c.File.Write),
	}

	fields := map[string]engine.ScriptValue{
		"text":            engine.NewObjectValue(textFields),
		"image":           engine.NewObjectValue(imageFields),
		"audio":           engine.NewObjectValue(audioFields),
		"video":           engine.NewObjectValue(videoFields),
		"file":            engine.NewObjectValue(fileFields),
		"functionCalling": engine.NewBoolValue(c.FunctionCalling),
		"streaming":       engine.NewBoolValue(c.Streaming),
	}
	return engine.NewObjectValue(fields)
}
