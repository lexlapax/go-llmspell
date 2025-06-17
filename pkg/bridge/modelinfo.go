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

// ModelInfoBridge provides access to LLM model information via go-llms ModelRegistry
type ModelInfoBridge struct {
	mu          sync.RWMutex
	registries  map[string]llmdomain.ModelRegistry
	initialized bool
}

// NewModelInfoBridge creates a new model info bridge
func NewModelInfoBridge() *ModelInfoBridge {
	return &ModelInfoBridge{
		registries: make(map[string]llmdomain.ModelRegistry),
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
func (b *ModelInfoBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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
func (b *ModelInfoBridge) RegisterModelRegistry(name string, registry llmdomain.ModelRegistry) error {
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
func (b *ModelInfoBridge) GetRegistry(name string) llmdomain.ModelRegistry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.registries[name]
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
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

// Helper function to convert models to ScriptValue
func convertModelsToScriptValue(models []domain.Model) engine.ScriptValue {
	values := make([]engine.ScriptValue, len(models))
	for i, m := range models {
		values[i] = convertModelToScriptValue(m)
	}
	return engine.NewArrayValue(values)
}

// Helper function to convert models to script format (kept for compatibility)
func convertModelsToScript(models []domain.Model) []map[string]interface{} {
	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		result[i] = convertModelToScript(m)
	}
	return result
}

// Helper function to convert a single model to ScriptValue
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

// Helper function to convert a single model to script format
func convertModelToScript(m domain.Model) map[string]interface{} {
	return map[string]interface{}{
		"provider":         m.Provider,
		"name":             m.Name,
		"displayName":      m.DisplayName,
		"description":      m.Description,
		"documentationURL": m.DocumentationURL,
		"contextWindow":    m.ContextWindow,
		"maxOutputTokens":  m.MaxOutputTokens,
		"trainingCutoff":   m.TrainingCutoff,
		"modelFamily":      m.ModelFamily,
		"lastUpdated":      m.LastUpdated,
		"pricing": map[string]interface{}{
			"inputPer1kTokens":  m.Pricing.InputPer1kTokens,
			"outputPer1kTokens": m.Pricing.OutputPer1kTokens,
		},
		"capabilities": convertCapabilitiesToScript(m.Capabilities),
	}
}

// Helper function to convert capabilities to ScriptValue
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

// Helper function to convert capabilities to script format
func convertCapabilitiesToScript(c domain.Capabilities) map[string]interface{} {
	return map[string]interface{}{
		"text": map[string]interface{}{
			"read":  c.Text.Read,
			"write": c.Text.Write,
		},
		"image": map[string]interface{}{
			"read":  c.Image.Read,
			"write": c.Image.Write,
		},
		"audio": map[string]interface{}{
			"read":  c.Audio.Read,
			"write": c.Audio.Write,
		},
		"video": map[string]interface{}{
			"read":  c.Video.Read,
			"write": c.Video.Write,
		},
		"file": map[string]interface{}{
			"read":  c.File.Read,
			"write": c.File.Write,
		},
		"functionCalling": c.FunctionCalling,
		"streaming":       c.Streaming,
	}
}
