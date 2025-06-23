// ABOUTME: Model info bridge providing access to go-llms ModelRegistry for LLM model discovery
// ABOUTME: Wraps go-llms model registry functionality without reimplementing

package bridge

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"

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
func (b *ModelInfoBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
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

// RegisterWithEngine registers the bridge with a script types.
// This method is called by the engine during registration - no delegation needed.
func (b *ModelInfoBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this bridge.
// Defines the script-accessible API for model information including
// registry management and model queries.
func (b *ModelInfoBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		{
			Name:        "registerModelRegistry",
			Description: "Register a model registry",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Description: "Registry name", Required: true},
				{Name: "registry", Type: "ModelRegistry", Description: "Model registry instance", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "listModels",
			Description: "List all models from all registries",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "listModelsByRegistry",
			Description: "List models from a specific registry",
			Parameters: []types.ParameterInfo{
				{Name: "registryName", Type: "string", Description: "Registry name", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getModel",
			Description: "Get a specific model by ID",
			Parameters: []types.ParameterInfo{
				{Name: "registryName", Type: "string", Description: "Registry name", Required: true},
				{Name: "modelID", Type: "string", Description: "Model ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "listRegistries",
			Description: "List all registered model registries",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
	}
}

// TypeMappings returns type conversion mappings.
// Maps go-llms types to script types for proper type conversion
// during method execution.
func (b *ModelInfoBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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
func (b *ModelInfoBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
// Defines that scripts need read access to model information
// for security sandboxing.
func (b *ModelInfoBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionMemory,
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
func (b *ModelInfoBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
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
		metadata := map[string]types.ScriptValue{
			"version":       types.NewStringValue(inventory.Metadata.Version),
			"lastUpdated":   types.NewStringValue(inventory.Metadata.LastUpdated),
			"description":   types.NewStringValue(inventory.Metadata.Description),
			"schemaVersion": types.NewStringValue(inventory.Metadata.SchemaVersion),
		}
		models := convertModelsToScriptValue(inventory.Models)

		result := map[string]types.ScriptValue{
			"metadata": types.NewObjectValue(metadata),
			"models":   models,
		}
		return types.NewObjectValue(result), nil

	case "fetchProviderModels":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("fetchProviderModels requires provider parameter")
		}
		provider := args[0].(types.StringValue).Value()

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
		values := make([]types.ScriptValue, len(registries))
		for i, reg := range registries {
			values[i] = types.NewStringValue(reg)
		}
		return types.NewArrayValue(values), nil

	case "registerModelRegistry":
		if len(args) < 2 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("registerModelRegistry requires name and registry parameters")
		}
		name := args[0].(types.StringValue).Value()
		// Note: registry parameter would need special handling as it's a Go type
		// For now, this would need to be passed as a custom value
		if args[1] == nil || args[1].Type() != types.TypeCustom {
			return nil, fmt.Errorf("registry must be ModelRegistry")
		}
		customVal := args[1].(types.CustomValue)
		registry, ok := customVal.Value().(llmdomain.ModelRegistry)
		if !ok {
			return nil, fmt.Errorf("registry must be ModelRegistry")
		}
		err := b.RegisterModelRegistry(name, registry)
		if err != nil {
			return nil, err
		}
		return types.NewNilValue(), nil

	case "listModels":
		// List all models from all registries
		var allModels []string
		for _, registry := range b.registries {
			models := registry.ListModels()
			allModels = append(allModels, models...)
		}
		values := make([]types.ScriptValue, len(allModels))
		for i, model := range allModels {
			values[i] = types.NewStringValue(model)
		}
		return types.NewArrayValue(values), nil

	case "listModelsByRegistry":
		if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("listModelsByRegistry requires registryName parameter")
		}
		registryName := args[0].(types.StringValue).Value()
		registry := b.GetRegistry(registryName)
		if registry == nil {
			return nil, fmt.Errorf("registry not found: %s", registryName)
		}
		models := registry.ListModels()
		values := make([]types.ScriptValue, len(models))
		for i, model := range models {
			values[i] = types.NewStringValue(model)
		}
		return types.NewArrayValue(values), nil

	case "getModel":
		if len(args) < 2 || args[0] == nil || args[0].Type() != types.TypeString ||
			args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("getModel requires registryName and modelID parameters")
		}
		registryName := args[0].(types.StringValue).Value()
		modelID := args[1].(types.StringValue).Value()
		registry := b.GetRegistry(registryName)
		if registry == nil {
			return nil, fmt.Errorf("registry not found: %s", registryName)
		}
		provider, err := registry.GetModel(modelID)
		if err != nil {
			return nil, fmt.Errorf("failed to get model: %w", err)
		}
		// Return provider as a custom value wrapped in an object
		result := map[string]types.ScriptValue{
			"provider": types.NewCustomValue("Provider", provider),
			"modelID":  types.NewStringValue(modelID),
		}
		return types.NewObjectValue(result), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// convertModelsToScriptValue converts an array of domain.Model to ScriptValue.
// Transforms go-llms model structs into script-compatible array format.
func convertModelsToScriptValue(models []domain.Model) types.ScriptValue {
	values := make([]types.ScriptValue, len(models))
	for i, m := range models {
		values[i] = convertModelToScriptValue(m)
	}
	return types.NewArrayValue(values)
}

// convertModelToScriptValue converts a single domain.Model to ScriptValue.
// Creates a comprehensive object representation including all model metadata,
// pricing information, and capability flags.
func convertModelToScriptValue(m domain.Model) types.ScriptValue {
	pricingFields := map[string]types.ScriptValue{
		"inputPer1kTokens":  types.NewNumberValue(m.Pricing.InputPer1kTokens),
		"outputPer1kTokens": types.NewNumberValue(m.Pricing.OutputPer1kTokens),
	}

	fields := map[string]types.ScriptValue{
		"provider":         types.NewStringValue(m.Provider),
		"name":             types.NewStringValue(m.Name),
		"displayName":      types.NewStringValue(m.DisplayName),
		"description":      types.NewStringValue(m.Description),
		"documentationURL": types.NewStringValue(m.DocumentationURL),
		"contextWindow":    types.NewNumberValue(float64(m.ContextWindow)),
		"maxOutputTokens":  types.NewNumberValue(float64(m.MaxOutputTokens)),
		"trainingCutoff":   types.NewStringValue(m.TrainingCutoff),
		"modelFamily":      types.NewStringValue(m.ModelFamily),
		"lastUpdated":      types.NewStringValue(m.LastUpdated),
		"pricing":          types.NewObjectValue(pricingFields),
		"capabilities":     convertCapabilitiesToScriptValue(m.Capabilities),
	}
	return types.NewObjectValue(fields)
}

// convertCapabilitiesToScriptValue converts domain.Capabilities to ScriptValue.
// Creates nested object structure representing model capabilities across
// different modalities (text, image, audio, video, file) and features.
func convertCapabilitiesToScriptValue(c domain.Capabilities) types.ScriptValue {
	textFields := map[string]types.ScriptValue{
		"read":  types.NewBoolValue(c.Text.Read),
		"write": types.NewBoolValue(c.Text.Write),
	}
	imageFields := map[string]types.ScriptValue{
		"read":  types.NewBoolValue(c.Image.Read),
		"write": types.NewBoolValue(c.Image.Write),
	}
	audioFields := map[string]types.ScriptValue{
		"read":  types.NewBoolValue(c.Audio.Read),
		"write": types.NewBoolValue(c.Audio.Write),
	}
	videoFields := map[string]types.ScriptValue{
		"read":  types.NewBoolValue(c.Video.Read),
		"write": types.NewBoolValue(c.Video.Write),
	}
	fileFields := map[string]types.ScriptValue{
		"read":  types.NewBoolValue(c.File.Read),
		"write": types.NewBoolValue(c.File.Write),
	}

	fields := map[string]types.ScriptValue{
		"text":            types.NewObjectValue(textFields),
		"image":           types.NewObjectValue(imageFields),
		"audio":           types.NewObjectValue(audioFields),
		"video":           types.NewObjectValue(videoFields),
		"file":            types.NewObjectValue(fileFields),
		"functionCalling": types.NewBoolValue(c.FunctionCalling),
		"streaming":       types.NewBoolValue(c.Streaming),
	}
	return types.NewObjectValue(fields)
}
