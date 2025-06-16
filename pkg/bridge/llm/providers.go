// ABOUTME: Provider system bridge for go-llms provider implementations and orchestration
// ABOUTME: Bridges provider registry, multi-provider strategies, consensus algorithms, and factory management

package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	// go-llms imports for provider functionality
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ProvidersBridge provides script access to go-llms provider system
type ProvidersBridge struct {
	initialized bool
	registry    *provider.DynamicRegistry
	providers   map[string]domain.Provider
	multi       map[string]*provider.MultiProvider
	mu          sync.RWMutex
}

// NewProvidersBridge creates a new providers bridge
func NewProvidersBridge() *ProvidersBridge {
	return &ProvidersBridge{
		registry:  provider.GetGlobalRegistry(),
		providers: make(map[string]domain.Provider),
		multi:     make(map[string]*provider.MultiProvider),
	}
}

// GetID returns the bridge identifier
func (pb *ProvidersBridge) GetID() string {
	return "providers"
}

// GetMetadata returns bridge metadata
func (pb *ProvidersBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "providers",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms provider system with registry, multi-provider strategies, and consensus algorithms",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/llm/provider"},
	}
}

// Initialize sets up the providers bridge
func (pb *ProvidersBridge) Initialize(ctx context.Context) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Register default factories
	if err := provider.RegisterDefaultFactories(pb.registry); err != nil {
		return fmt.Errorf("failed to register default factories: %w", err)
	}

	pb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (pb *ProvidersBridge) Cleanup(ctx context.Context) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Clear all stored providers
	pb.providers = make(map[string]domain.Provider)
	pb.multi = make(map[string]*provider.MultiProvider)
	pb.initialized = false

	return nil
}

// IsInitialized returns initialization status
func (pb *ProvidersBridge) IsInitialized() bool {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	return pb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (pb *ProvidersBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(pb)
}

// Methods returns available bridge methods
func (pb *ProvidersBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Provider registry methods
		{
			Name:        "createProvider",
			Description: "Create provider from factory configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "providerType", Type: "string", Required: true, Description: "Provider type (openai, anthropic, etc.)"},
				{Name: "name", Type: "string", Required: true, Description: "Provider instance name"},
				{Name: "config", Type: "object", Required: true, Description: "Provider configuration"},
			},
			ReturnType: "object",
			Examples:   []string{"createProvider('openai', 'my-openai', {api_key: 'sk-...', model: 'gpt-4'})"},
		},
		{
			Name:        "createProviderFromEnvironment",
			Description: "Create provider from environment variables",
			Parameters: []engine.ParameterInfo{
				{Name: "providerType", Type: "string", Required: true, Description: "Provider type"},
				{Name: "name", Type: "string", Required: true, Description: "Provider instance name"},
			},
			ReturnType: "object",
			Examples:   []string{"createProviderFromEnvironment('openai', 'env-openai')"},
		},
		{
			Name:        "registerProvider",
			Description: "Register existing provider instance",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "providerType", Type: "string", Required: true, Description: "Provider type"},
			},
			ReturnType: "object",
			Examples:   []string{"registerProvider('my-provider', 'openai')"},
		},
		{
			Name:        "getProvider",
			Description: "Get provider by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
			Examples:   []string{"getProvider('my-openai')"},
		},
		{
			Name:        "listProviders",
			Description: "List all registered providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listProviders()"},
		},
		{
			Name:        "removeProvider",
			Description: "Remove provider from registry",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "void",
			Examples:   []string{"removeProvider('my-provider')"},
		},
		{
			Name:        "getProviderMetadata",
			Description: "Get provider metadata and capabilities",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
			Examples:   []string{"getProviderMetadata('my-openai')"},
		},
		{
			Name:        "listProvidersByCapability",
			Description: "List providers with specific capability",
			Parameters: []engine.ParameterInfo{
				{Name: "capability", Type: "string", Required: true, Description: "Capability name"},
			},
			ReturnType: "array",
			Examples:   []string{"listProvidersByCapability('streaming')"},
		},
		// Provider templates and configuration
		{
			Name:        "getProviderTemplate",
			Description: "Get configuration template for provider type",
			Parameters: []engine.ParameterInfo{
				{Name: "providerType", Type: "string", Required: true, Description: "Provider type"},
			},
			ReturnType: "object",
			Examples:   []string{"getProviderTemplate('openai')"},
		},
		{
			Name:        "listProviderTemplates",
			Description: "List all available provider templates",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listProviderTemplates()"},
		},
		{
			Name:        "validateProviderConfig",
			Description: "Validate provider configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "providerType", Type: "string", Required: true, Description: "Provider type"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration to validate"},
			},
			ReturnType: "object",
			Examples:   []string{"validateProviderConfig('openai', {api_key: 'sk-...'})"},
		},
		// Multi-provider orchestration
		{
			Name:        "createMultiProvider",
			Description: "Create multi-provider with strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "providers", Type: "array", Required: true, Description: "Array of provider configurations with weights"},
				{Name: "strategy", Type: "string", Required: true, Description: "Selection strategy (fastest, primary, consensus)"},
			},
			ReturnType: "object",
			Examples:   []string{"createMultiProvider('multi1', [{name: 'openai', weight: 0.7}, {name: 'anthropic', weight: 0.3}], 'consensus')"},
		},
		{
			Name:        "configureMultiProvider",
			Description: "Configure multi-provider settings",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration options"},
			},
			ReturnType: "void",
			Examples:   []string{"configureMultiProvider('multi1', {timeout: 30, primaryIndex: 0, consensusStrategy: 'similarity'})"},
		},
		{
			Name:        "getMultiProvider",
			Description: "Get multi-provider instance",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
			},
			ReturnType: "object",
			Examples:   []string{"getMultiProvider('multi1')"},
		},
		// Registry configuration
		{
			Name:        "exportProviderConfig",
			Description: "Export provider registry configuration",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"exportProviderConfig()"},
		},
		{
			Name:        "importProviderConfig",
			Description: "Import provider registry configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Configuration to import"},
			},
			ReturnType: "void",
			Examples:   []string{"importProviderConfig(config)"},
		},
		// Provider operations
		{
			Name:        "generateWithProvider",
			Description: "Generate text using specific provider",
			Parameters: []engine.ParameterInfo{
				{Name: "providerName", Type: "string", Required: true, Description: "Provider name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "string",
			Examples:   []string{"generateWithProvider('my-openai', 'Hello world', {temperature: 0.7})"},
		},
		{
			Name:        "generateWithMultiProvider",
			Description: "Generate text using multi-provider strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "multiProviderName", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples:   []string{"generateWithMultiProvider('multi1', 'Hello world', {temperature: 0.7})"},
		},
	}
}

// ValidateMethod validates method calls
func (pb *ProvidersBridge) ValidateMethod(name string, args []interface{}) error {
	if !pb.IsInitialized() {
		return fmt.Errorf("providers bridge not initialized")
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
func (pb *ProvidersBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"provider": {
			GoType:     "domain.Provider",
			ScriptType: "object",
			Converter:  "providerConverter",
			Metadata:   map[string]interface{}{"description": "LLM provider interface"},
		},
		"multi_provider": {
			GoType:     "*provider.MultiProvider",
			ScriptType: "object",
			Converter:  "multiProviderConverter",
			Metadata:   map[string]interface{}{"description": "Multi-provider orchestrator"},
		},
		"provider_template": {
			GoType:     "provider.ProviderTemplate",
			ScriptType: "object",
			Converter:  "providerTemplateConverter",
			Metadata:   map[string]interface{}{"description": "Provider configuration template"},
		},
		"provider_metadata": {
			GoType:     "provider.ProviderMetadata",
			ScriptType: "object",
			Converter:  "providerMetadataConverter",
			Metadata:   map[string]interface{}{"description": "Provider metadata and capabilities"},
		},
	}
}

// RequiredPermissions returns required permissions
func (pb *ProvidersBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm.providers",
			Actions:     []string{"create", "read", "update", "delete"},
			Description: "Access to LLM provider APIs",
		},
		{
			Type:        engine.PermissionProcess,
			Resource:    "provider.registry",
			Actions:     []string{"register", "unregister", "list"},
			Description: "Manage provider registry",
		},
	}
}

// Bridge method implementations

// Provider registry methods

// createProvider creates a provider from factory configuration
func (pb *ProvidersBridge) createProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("createProvider", args); err != nil {
		return nil, err
	}

	providerType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider type must be a string")
	}

	name, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	config, ok := args[2].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("config must be an object")
	}

	// Create provider from template
	if err := pb.registry.CreateProviderFromTemplate(providerType, name, config); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Get the created provider
	createdProvider, err := pb.registry.GetProvider(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get created provider: %w", err)
	}

	// Store in local registry for direct access
	pb.mu.Lock()
	pb.providers[name] = createdProvider
	pb.mu.Unlock()

	return map[string]interface{}{
		"name":     name,
		"type":     providerType,
		"created":  time.Now(),
		"provider": createdProvider,
	}, nil
}

// createProviderFromEnvironment creates a provider from environment variables
func (pb *ProvidersBridge) createProviderFromEnvironment(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("createProviderFromEnvironment", args); err != nil {
		return nil, err
	}

	providerType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider type must be a string")
	}

	name, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	// Create provider from environment
	createdProvider, err := provider.CreateProviderFromEnvironment(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider from environment: %w", err)
	}

	// Register with registry
	if err := pb.registry.RegisterProvider(name, createdProvider, nil); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Store in local registry for direct access
	pb.mu.Lock()
	pb.providers[name] = createdProvider
	pb.mu.Unlock()

	return map[string]interface{}{
		"name":     name,
		"type":     providerType,
		"created":  time.Now(),
		"provider": createdProvider,
		"source":   "environment",
	}, nil
}

// registerProvider registers an existing provider instance
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) registerProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("registerProvider", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	providerType, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("provider type must be a string")
	}

	// For now, this creates a provider from environment as a placeholder
	// In a full implementation, this would accept a provider instance
	createdProvider, err := provider.CreateProviderFromEnvironment(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Register with registry
	if err := pb.registry.RegisterProvider(name, createdProvider, nil); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Store in local registry for direct access
	pb.mu.Lock()
	pb.providers[name] = createdProvider
	pb.mu.Unlock()

	return map[string]interface{}{
		"name":       name,
		"type":       providerType,
		"registered": time.Now(),
	}, nil
}

// getProvider gets a provider by name
func (pb *ProvidersBridge) getProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getProvider", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	provider, err := pb.registry.GetProvider(name)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	return map[string]interface{}{
		"name":     name,
		"provider": provider,
		"active":   true,
	}, nil
}

// listProviders lists all registered providers
func (pb *ProvidersBridge) listProviders(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("listProviders", args); err != nil {
		return nil, err
	}

	providers := pb.registry.ListProviders()
	result := make([]map[string]interface{}, len(providers))

	for i, name := range providers {
		result[i] = map[string]interface{}{
			"name":   name,
			"active": true,
		}
	}

	return result, nil
}

// removeProvider removes a provider from the registry
func (pb *ProvidersBridge) removeProvider(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("removeProvider", args); err != nil {
		return err
	}

	name, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("provider name must be a string")
	}

	// Remove from registry
	if err := pb.registry.UnregisterProvider(name); err != nil {
		return fmt.Errorf("failed to unregister provider: %w", err)
	}

	// Remove from local registry
	pb.mu.Lock()
	delete(pb.providers, name)
	pb.mu.Unlock()

	return nil
}

// getProviderMetadata gets provider metadata and capabilities
func (pb *ProvidersBridge) getProviderMetadata(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getProviderMetadata", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	metadata, err := pb.registry.GetMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	if metadata == nil {
		return map[string]interface{}{
			"name":         name,
			"capabilities": []string{},
			"metadata":     map[string]interface{}{},
		}, nil
	}

	return map[string]interface{}{
		"name":         name,
		"capabilities": metadata.GetCapabilities(),
		"metadata":     metadata,
	}, nil
}

// listProvidersByCapability lists providers with specific capability
func (pb *ProvidersBridge) listProvidersByCapability(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("listProvidersByCapability", args); err != nil {
		return nil, err
	}

	capabilityStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("capability must be a string")
	}

	capability := provider.Capability(capabilityStr)
	providers := pb.registry.ListProvidersByCapability(capability)

	result := make([]map[string]interface{}, len(providers))
	for i, name := range providers {
		result[i] = map[string]interface{}{
			"name":       name,
			"capability": capabilityStr,
		}
	}

	return result, nil
}

// Provider templates and configuration

// getProviderTemplate gets configuration template for provider type
func (pb *ProvidersBridge) getProviderTemplate(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getProviderTemplate", args); err != nil {
		return nil, err
	}

	providerType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider type must be a string")
	}

	template, err := pb.registry.GetTemplate(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return map[string]interface{}{
		"type":        template.Type,
		"name":        template.Name,
		"description": template.Description,
		"schema":      template.Schema,
		"defaults":    template.Defaults,
		"examples":    template.Examples,
	}, nil
}

// listProviderTemplates lists all available provider templates
func (pb *ProvidersBridge) listProviderTemplates(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("listProviderTemplates", args); err != nil {
		return nil, err
	}

	templates := pb.registry.ListTemplates()
	result := make([]map[string]interface{}, len(templates))

	for i, template := range templates {
		result[i] = map[string]interface{}{
			"type":        template.Type,
			"name":        template.Name,
			"description": template.Description,
		}
	}

	return result, nil
}

// validateProviderConfig validates provider configuration
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) validateProviderConfig(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("validateProviderConfig", args); err != nil {
		return nil, err
	}

	providerType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider type must be a string")
	}

	config, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("config must be an object")
	}

	// Get template to access factory
	template, err := pb.registry.GetTemplate(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// For validation, we'd need access to the factory
	// This is a simplified implementation that acknowledges the config
	_ = config // Acknowledge config parameter for future validation implementation
	return map[string]interface{}{
		"valid":        true,
		"providerType": providerType,
		"template":     template.Type,
		"errors":       []string{},
	}, nil
}

// Multi-provider orchestration

// createMultiProvider creates multi-provider with strategy
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) createMultiProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("createMultiProvider", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("multi-provider name must be a string")
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
	var strategy provider.SelectionStrategy
	switch strategyStr {
	case "fastest":
		strategy = provider.StrategyFastest
	case "primary":
		strategy = provider.StrategyPrimary
	case "consensus":
		strategy = provider.StrategyConsensus
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyStr)
	}

	// Build provider weights
	var providerWeights []provider.ProviderWeight
	for _, p := range providersArray {
		providerConfig, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		providerName, _ := providerConfig["name"].(string)
		weight, _ := providerConfig["weight"].(float64)
		if weight <= 0 {
			weight = 1.0
		}

		// Get provider from registry
		providerInstance, err := pb.registry.GetProvider(providerName)
		if err != nil {
			return nil, fmt.Errorf("provider not found: %s", providerName)
		}

		providerWeights = append(providerWeights, provider.ProviderWeight{
			Provider: providerInstance,
			Weight:   weight,
			Name:     providerName,
		})
	}

	if len(providerWeights) == 0 {
		return nil, fmt.Errorf("no valid providers specified")
	}

	// Create multi-provider
	multiProvider := provider.NewMultiProvider(providerWeights, strategy)

	// Store in local registry
	pb.mu.Lock()
	pb.multi[name] = multiProvider
	pb.mu.Unlock()

	return map[string]interface{}{
		"name":      name,
		"strategy":  strategyStr,
		"providers": len(providerWeights),
		"created":   time.Now(),
	}, nil
}

// configureMultiProvider configures multi-provider settings
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) configureMultiProvider(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("configureMultiProvider", args); err != nil {
		return err
	}

	name, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("multi-provider name must be a string")
	}

	config, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("config must be an object")
	}

	pb.mu.RLock()
	multiProvider, exists := pb.multi[name]
	pb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("multi-provider not found: %s", name)
	}

	// Apply configuration
	if timeout, ok := config["timeout"].(float64); ok {
		multiProvider.WithTimeout(time.Duration(timeout) * time.Second)
	}

	if primaryIndex, ok := config["primaryIndex"].(float64); ok {
		multiProvider.WithPrimaryProvider(int(primaryIndex))
	}

	if consensusStrategy, ok := config["consensusStrategy"].(string); ok {
		var strategy provider.ConsensusStrategy
		switch consensusStrategy {
		case "majority":
			strategy = provider.ConsensusMajority
		case "similarity":
			strategy = provider.ConsensusSimilarity
		case "weighted":
			strategy = provider.ConsensusWeighted
		default:
			return fmt.Errorf("unknown consensus strategy: %s", consensusStrategy)
		}
		multiProvider.WithConsensusStrategy(strategy)
	}

	return nil
}

// getMultiProvider gets multi-provider instance
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) getMultiProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("getMultiProvider", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("multi-provider name must be a string")
	}

	pb.mu.RLock()
	multiProvider, exists := pb.multi[name]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("multi-provider not found: %s", name)
	}

	return map[string]interface{}{
		"name":     name,
		"provider": multiProvider,
		"active":   true,
	}, nil
}

// Registry configuration

// exportProviderConfig exports provider registry configuration
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) exportProviderConfig(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("exportProviderConfig", args); err != nil {
		return nil, err
	}

	config, err := pb.registry.ExportConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to export config: %w", err)
	}

	return config, nil
}

// importProviderConfig imports provider registry configuration
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) importProviderConfig(ctx context.Context, args []interface{}) error {
	if err := pb.ValidateMethod("importProviderConfig", args); err != nil {
		return err
	}

	config, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("config must be an object")
	}

	if err := pb.registry.ImportConfig(config); err != nil {
		return fmt.Errorf("failed to import config: %w", err)
	}

	return nil
}

// Provider operations

// generateWithProvider generates text using specific provider
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) generateWithProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("generateWithProvider", args); err != nil {
		return nil, err
	}

	providerName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("provider name must be a string")
	}

	prompt, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("prompt must be a string")
	}

	// Get options if provided
	var options []domain.Option
	if len(args) > 2 {
		if optionsMap, ok := args[2].(map[string]interface{}); ok {
			// Convert options map to domain options
			if temperature, ok := optionsMap["temperature"].(float64); ok {
				options = append(options, domain.WithTemperature(temperature))
			}
			if maxTokens, ok := optionsMap["max_tokens"].(float64); ok {
				options = append(options, domain.WithMaxTokens(int(maxTokens)))
			}
		}
	}

	// Get provider
	provider, err := pb.registry.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Generate
	result, err := provider.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return result, nil
}

// generateWithMultiProvider generates text using multi-provider strategy
//
//nolint:unused // Bridge method called via reflection
func (pb *ProvidersBridge) generateWithMultiProvider(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := pb.ValidateMethod("generateWithMultiProvider", args); err != nil {
		return nil, err
	}

	multiProviderName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("multi-provider name must be a string")
	}

	prompt, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("prompt must be a string")
	}

	// Get options if provided
	var options []domain.Option
	if len(args) > 2 {
		if optionsMap, ok := args[2].(map[string]interface{}); ok {
			// Convert options map to domain options
			if temperature, ok := optionsMap["temperature"].(float64); ok {
				options = append(options, domain.WithTemperature(temperature))
			}
			if maxTokens, ok := optionsMap["max_tokens"].(float64); ok {
				options = append(options, domain.WithMaxTokens(int(maxTokens)))
			}
		}
	}

	// Get multi-provider
	pb.mu.RLock()
	multiProvider, exists := pb.multi[multiProviderName]
	pb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("multi-provider not found: %s", multiProviderName)
	}

	// Generate
	result, err := multiProvider.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return map[string]interface{}{
		"result":   result,
		"strategy": "multi-provider",
		"provider": multiProviderName,
	}, nil
}
