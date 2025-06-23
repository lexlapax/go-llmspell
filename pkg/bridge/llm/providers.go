// ABOUTME: Provider management bridge for dynamic creation and configuration of LLM providers
// ABOUTME: Supports multi-provider configurations, templates, and environment-based setup

package llm

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// ProviderTemplate defines a template for creating providers.
// It specifies required configuration, environment variables,
// and default settings for different provider types.
type ProviderTemplate struct {
	Type            string
	Description     string
	RequiredEnvVars []string
	OptionalEnvVars []string
	DefaultConfig   map[string]interface{}
}

// MultiProvider manages multiple providers with a selection strategy.
// It supports strategies like fastest response, primary fallback,
// and consensus-based decision making.
type MultiProvider struct {
	Name      string
	Providers []MultiProviderEntry
	Strategy  string // "fastest", "primary", "consensus"
	Config    MultiProviderConfig
}

// MultiProviderEntry represents a provider in a multi-provider setup.
// It includes the provider instance, weight for weighted strategies,
// and primary designation for fallback scenarios.
type MultiProviderEntry struct {
	Name     string
	Provider types.Provider
	Weight   float64
	Primary  bool
}

// MultiProviderConfig holds configuration for multi-provider.
// It defines consensus thresholds, timeouts, and retry behavior
// for multi-provider strategies.
type MultiProviderConfig struct {
	ConsensusThreshold float64       // For consensus strategy
	Timeout            time.Duration // For fastest strategy
	RetryOnFailure     bool
}

// ProvidersBridge manages provider creation and configuration.
// It handles dynamic provider creation, multi-provider setups,
// template management, and provider metadata tracking.
type ProvidersBridge struct {
	mu             sync.RWMutex
	initialized    bool
	providers      map[string]types.Provider
	multiProviders map[string]*MultiProvider
	templates      map[string]*ProviderTemplate
	metadata       map[string]map[string]interface{}
	llmBridge      *LLMBridge // Reference to main LLM bridge
}

// NewProvidersBridge creates a new providers bridge.
// It initializes provider registries, templates, and metadata storage
// with a reference to the main LLM bridge.
func NewProvidersBridge(llmBridge *LLMBridge) *ProvidersBridge {
	return &ProvidersBridge{
		providers:      make(map[string]types.Provider),
		multiProviders: make(map[string]*MultiProvider),
		templates:      initializeTemplates(),
		metadata:       make(map[string]map[string]interface{}),
		llmBridge:      llmBridge,
	}
}

// initializeTemplates creates default provider templates.
// It returns templates for common providers like OpenAI, Anthropic,
// and mock providers with their configuration requirements.
func initializeTemplates() map[string]*ProviderTemplate {
	return map[string]*ProviderTemplate{
		"openai": {
			Type:            "openai",
			Description:     "OpenAI GPT models",
			RequiredEnvVars: []string{"OPENAI_API_KEY"},
			OptionalEnvVars: []string{"OPENAI_ORG_ID", "OPENAI_BASE_URL"},
			DefaultConfig: map[string]interface{}{
				"model":       "gpt-3.5-turbo",
				"temperature": 0.7,
			},
		},
		"anthropic": {
			Type:            "anthropic",
			Description:     "Anthropic Claude models",
			RequiredEnvVars: []string{"ANTHROPIC_API_KEY"},
			OptionalEnvVars: []string{"ANTHROPIC_BASE_URL"},
			DefaultConfig: map[string]interface{}{
				"model":       "claude-3-sonnet-20240229",
				"temperature": 0.7,
			},
		},
		"mock": {
			Type:            "mock",
			Description:     "Mock provider for testing",
			RequiredEnvVars: []string{},
			OptionalEnvVars: []string{},
			DefaultConfig: map[string]interface{}{
				"responses": []string{"Mock response"},
			},
		},
	}
}

// GetID returns the bridge ID
func (b *ProvidersBridge) GetID() string {
	return "providers"
}

// GetMetadata returns bridge metadata
func (b *ProvidersBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "providers",
		Version:     "1.0.0",
		Description: "Provider management and configuration bridge",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *ProvidersBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *ProvidersBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.providers = make(map[string]types.Provider)
	b.multiProviders = make(map[string]*MultiProvider)
	b.metadata = make(map[string]map[string]interface{})
	b.initialized = false

	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *ProvidersBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *ProvidersBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this bridge
func (b *ProvidersBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Provider creation
		{
			Name:        "createProvider",
			Description: "Create a new provider",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "config", Type: "object", Required: true, Description: "Provider configuration"},
			},
			ReturnType: "object",
		},
		{
			Name:        "createProviderFromEnvironment",
			Description: "Create provider from environment variables",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "getProvider",
			Description: "Get a provider by name",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProviders",
			Description: "List all providers",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "removeProvider",
			Description: "Remove a provider",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "void",
		},
		// Templates
		{
			Name:        "getProviderTemplate",
			Description: "Get provider template",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProviderTemplates",
			Description: "List available provider templates",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "validateProviderConfig",
			Description: "Validate provider configuration",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration to validate"},
			},
			ReturnType: "object",
		},
		// Multi-provider
		{
			Name:        "createMultiProvider",
			Description: "Create a multi-provider configuration",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "providers", Type: "array", Required: true, Description: "Array of provider configurations"},
				{Name: "strategy", Type: "string", Required: true, Description: "Selection strategy"},
			},
			ReturnType: "object",
		},
		{
			Name:        "configureMultiProvider",
			Description: "Configure multi-provider settings",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration object"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getMultiProvider",
			Description: "Get multi-provider information",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
			},
			ReturnType: "object",
		},
		// Mock provider
		{
			Name:        "createMockProvider",
			Description: "Create a mock provider for testing",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "responses", Type: "array", Required: true, Description: "Array of mock responses"},
			},
			ReturnType: "object",
		},
		// Provider operations
		{
			Name:        "generateWithProvider",
			Description: "Generate text using specific provider",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
				{Name: "prompt", Type: "string", Required: true, Description: "Prompt text"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "string",
		},
		// Export/Import
		{
			Name:        "exportProviderConfig",
			Description: "Export all provider configurations",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "importProviderConfig",
			Description: "Import provider configurations",
			Parameters: []types.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Configuration to import"},
			},
			ReturnType: "void",
		},
		// Metadata
		{
			Name:        "setProviderMetadata",
			Description: "Set metadata for a provider",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
				{Name: "metadata", Type: "object", Required: true, Description: "Metadata object"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getProviderMetadata",
			Description: "Get metadata for a provider",
			Parameters: []types.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProvidersByCapability",
			Description: "List providers by capability",
			Parameters: []types.ParameterInfo{
				{Name: "capability", Type: "string", Required: true, Description: "Capability name"},
			},
			ReturnType: "array",
		},
	}
}

// ValidateMethod validates method parameters
func (b *ProvidersBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	if !b.IsInitialized() {
		return fmt.Errorf("providers bridge not initialized")
	}

	switch name {
	case "createProvider":
		if len(args) < 3 {
			return fmt.Errorf("createProvider requires type, name, and config")
		}
	case "createProviderFromEnvironment", "createMultiProvider":
		if len(args) < 2 {
			return fmt.Errorf("%s requires at least 2 arguments", name)
		}
	case "getProvider", "removeProvider", "getProviderTemplate", "getMultiProvider",
		"getProviderMetadata", "listProvidersByCapability":
		if len(args) < 1 {
			return fmt.Errorf("%s requires name/type argument", name)
		}
	case "validateProviderConfig", "configureMultiProvider", "setProviderMetadata",
		"createMockProvider":
		if len(args) < 2 {
			return fmt.Errorf("%s requires 2 arguments", name)
		}
	case "generateWithProvider":
		if len(args) < 2 {
			return fmt.Errorf("generateWithProvider requires provider and prompt")
		}
	case "importProviderConfig":
		if len(args) < 1 {
			return fmt.Errorf("importProviderConfig requires config")
		}
	}

	return nil
}

// ExecuteMethod executes a bridge method with ScriptValue parameters
func (b *ProvidersBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	if !b.initialized {
		b.mu.RUnlock()
		return types.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}
	b.mu.RUnlock()

	switch name {
	case "createProvider":
		return b.createProvider(ctx, args)
	case "createProviderFromEnvironment":
		return b.createProviderFromEnvironment(ctx, args)
	case "getProvider":
		return b.getProvider(ctx, args)
	case "listProviders":
		return b.listProviders(ctx, args)
	case "removeProvider":
		return b.removeProvider(ctx, args)
	case "getProviderTemplate":
		return b.getProviderTemplate(ctx, args)
	case "listProviderTemplates":
		return b.listProviderTemplates(ctx, args)
	case "validateProviderConfig":
		return b.validateProviderConfig(ctx, args)
	case "createMultiProvider":
		return b.createMultiProvider(ctx, args)
	case "configureMultiProvider":
		return b.configureMultiProvider(ctx, args)
	case "getMultiProvider":
		return b.getMultiProvider(ctx, args)
	case "createMockProvider":
		return b.createMockProvider(ctx, args)
	case "generateWithProvider":
		return b.generateWithProvider(ctx, args)
	case "exportProviderConfig":
		return b.exportProviderConfig(ctx, args)
	case "importProviderConfig":
		return b.importProviderConfig(ctx, args)
	case "setProviderMetadata":
		return b.setProviderMetadata(ctx, args)
	case "getProviderMetadata":
		return b.getProviderMetadata(ctx, args)
	case "listProvidersByCapability":
		return b.listProvidersByCapability(ctx, args)
	default:
		return types.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// TypeMappings returns type conversion mappings
func (b *ProvidersBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
		"provider": {
			GoType:     "types.Provider",
			ScriptType: "object",
		},
		"provider_template": {
			GoType:     "ProviderTemplate",
			ScriptType: "object",
		},
		"multi_provider": {
			GoType:     "MultiProvider",
			ScriptType: "object",
		},
	}
}

// RequiredPermissions returns required permissions
func (b *ProvidersBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionNetwork,
			Resource:    "llm.providers",
			Actions:     []string{"create", "read", "write", "delete"},
			Description: "Manage LLM providers",
		},
		{
			Type:        types.PermissionProcess,
			Resource:    "provider.registry",
			Actions:     []string{"read", "write"},
			Description: "Access provider registry",
		},
	}
}

// Provider Management Methods

func (b *ProvidersBridge) createProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("createProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerType := args[0].(types.StringValue).Value()
	providerName := args[1].(types.StringValue).Value()
	configMap := args[2].ToGo().(map[string]interface{})

	// Check if provider already exists
	b.mu.RLock()
	if _, exists := b.providers[providerName]; exists {
		b.mu.RUnlock()
		return types.NewErrorValue(fmt.Errorf("provider %s already exists", providerName)), nil
	}
	b.mu.RUnlock()

	// Get template
	template, exists := b.templates[providerType]
	if !exists {
		return types.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
	}

	// Merge with default config
	mergedConfig := make(map[string]interface{})
	for k, v := range template.DefaultConfig {
		mergedConfig[k] = v
	}
	for k, v := range configMap {
		mergedConfig[k] = v
	}

	// Add provider to LLM bridge
	if b.llmBridge != nil {
		b.llmBridge.mu.Lock()
		// In a real implementation, this would create the actual provider
		// For now, we'll just register the name
		b.llmBridge.providers[providerName] = nil
		b.llmBridge.mu.Unlock()
	}

	b.mu.Lock()
	// Store a placeholder - in real implementation this would be the actual provider
	b.providers[providerName] = nil
	b.metadata[providerName] = mergedConfig
	b.mu.Unlock()

	result := map[string]types.ScriptValue{
		"name":    types.NewStringValue(providerName),
		"type":    types.NewStringValue(providerType),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) createProviderFromEnvironment(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("createProviderFromEnvironment", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerType := args[0].(types.StringValue).Value()
	providerName := args[1].(types.StringValue).Value()

	template, exists := b.templates[providerType]
	if !exists {
		return types.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
	}

	// Check required env vars
	config := make(map[string]interface{})
	for _, envVar := range template.RequiredEnvVars {
		value := os.Getenv(envVar)
		if value == "" {
			return types.NewErrorValue(fmt.Errorf("required environment variable %s not set", envVar)), nil
		}
		config[envVar] = value
	}

	// Add optional env vars
	for _, envVar := range template.OptionalEnvVars {
		if value := os.Getenv(envVar); value != "" {
			config[envVar] = value
		}
	}

	// Add defaults
	for k, v := range template.DefaultConfig {
		if _, exists := config[k]; !exists {
			config[k] = v
		}
	}

	// Add provider to LLM bridge
	if b.llmBridge != nil {
		b.llmBridge.mu.Lock()
		b.llmBridge.providers[providerName] = nil
		b.llmBridge.mu.Unlock()
	}

	b.mu.Lock()
	b.providers[providerName] = nil
	b.metadata[providerName] = config
	b.mu.Unlock()

	result := map[string]types.ScriptValue{
		"name":    types.NewStringValue(providerName),
		"type":    types.NewStringValue(providerType),
		"source":  types.NewStringValue("environment"),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) getProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("getProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	b.mu.RLock()
	_, exists := b.providers[name]
	metadata := b.metadata[name]
	b.mu.RUnlock()

	if !exists {
		return types.NewErrorValue(fmt.Errorf("provider not found: %s", name)), nil
	}

	// Find provider type from metadata
	providerType := "unknown"
	if metadata != nil {
		for pType, template := range b.templates {
			matches := true
			for _, reqVar := range template.RequiredEnvVars {
				if _, ok := metadata[reqVar]; !ok {
					matches = false
					break
				}
			}
			if matches {
				providerType = pType
				break
			}
		}
	}

	result := map[string]types.ScriptValue{
		"name": types.NewStringValue(name),
		"type": types.NewStringValue(providerType),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) listProviders(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]types.ScriptValue, 0, len(b.providers))
	for name := range b.providers {
		providers = append(providers, types.NewStringValue(name))
	}

	return types.NewArrayValue(providers), nil
}

func (b *ProvidersBridge) removeProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("removeProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	// Remove from LLM bridge
	if b.llmBridge != nil {
		b.llmBridge.mu.Lock()
		delete(b.llmBridge.providers, name)
		b.llmBridge.mu.Unlock()
	}

	b.mu.Lock()
	delete(b.providers, name)
	delete(b.metadata, name)
	b.mu.Unlock()

	return types.NewNilValue(), nil
}

// Template Methods

func (b *ProvidersBridge) getProviderTemplate(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("getProviderTemplate", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerType := args[0].(types.StringValue).Value()

	template, exists := b.templates[providerType]
	if !exists {
		return types.NewErrorValue(fmt.Errorf("template not found: %s", providerType)), nil
	}

	return b.templateToScriptValue(template), nil
}

func (b *ProvidersBridge) listProviderTemplates(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	templates := make([]types.ScriptValue, 0, len(b.templates))
	for _, template := range b.templates {
		templates = append(templates, b.templateToScriptValue(template))
	}

	return types.NewArrayValue(templates), nil
}

func (b *ProvidersBridge) validateProviderConfig(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("validateProviderConfig", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerType := args[0].(types.StringValue).Value()
	configMap := args[1].ToGo().(map[string]interface{})

	template, exists := b.templates[providerType]
	if !exists {
		return types.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
	}

	errors := []string{}

	// Check required vars
	for _, reqVar := range template.RequiredEnvVars {
		if _, ok := configMap[reqVar]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", reqVar))
		}
	}

	valid := len(errors) == 0

	errorsArray := make([]types.ScriptValue, len(errors))
	for i, err := range errors {
		errorsArray[i] = types.NewStringValue(err)
	}

	result := map[string]types.ScriptValue{
		"valid":  types.NewBoolValue(valid),
		"errors": types.NewArrayValue(errorsArray),
	}

	return types.NewObjectValue(result), nil
}

// Multi-Provider Methods

func (b *ProvidersBridge) createMultiProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("createMultiProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()
	providersArray := args[1].ToGo().([]interface{})
	strategy := args[2].(types.StringValue).Value()

	// Validate strategy
	validStrategies := map[string]bool{
		"fastest":   true,
		"primary":   true,
		"consensus": true,
	}
	if !validStrategies[strategy] {
		return types.NewErrorValue(fmt.Errorf("invalid strategy: %s", strategy)), nil
	}

	// Parse providers
	entries := make([]MultiProviderEntry, 0, len(providersArray))
	for _, p := range providersArray {
		providerMap := p.(map[string]interface{})
		entry := MultiProviderEntry{
			Name:   providerMap["name"].(string),
			Weight: 1.0,
		}
		if weight, ok := providerMap["weight"].(float64); ok {
			entry.Weight = weight
		}
		if primary, ok := providerMap["primary"].(bool); ok {
			entry.Primary = primary
		}

		// Check if provider exists
		b.mu.RLock()
		if _, exists := b.providers[entry.Name]; !exists {
			b.mu.RUnlock()
			return types.NewErrorValue(fmt.Errorf("provider not found: %s", entry.Name)), nil
		}
		b.mu.RUnlock()

		entries = append(entries, entry)
	}

	// Create multi-provider
	multi := &MultiProvider{
		Name:      name,
		Providers: entries,
		Strategy:  strategy,
		Config: MultiProviderConfig{
			ConsensusThreshold: 0.5,
			Timeout:            30 * time.Second,
			RetryOnFailure:     true,
		},
	}

	b.mu.Lock()
	b.multiProviders[name] = multi
	b.mu.Unlock()

	result := map[string]types.ScriptValue{
		"name":      types.NewStringValue(name),
		"strategy":  types.NewStringValue(strategy),
		"providers": types.NewNumberValue(float64(len(entries))),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) configureMultiProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("configureMultiProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()
	configMap := args[1].ToGo().(map[string]interface{})

	b.mu.Lock()
	multi, exists := b.multiProviders[name]
	if !exists {
		b.mu.Unlock()
		return types.NewErrorValue(fmt.Errorf("multi-provider not found: %s", name)), nil
	}

	// Update config
	if threshold, ok := configMap["consensusThreshold"].(float64); ok {
		multi.Config.ConsensusThreshold = threshold
	}
	if timeout, ok := configMap["timeout"].(float64); ok {
		multi.Config.Timeout = time.Duration(timeout) * time.Second
	}
	if retry, ok := configMap["retryOnFailure"].(bool); ok {
		multi.Config.RetryOnFailure = retry
	}
	b.mu.Unlock()

	return types.NewNilValue(), nil
}

func (b *ProvidersBridge) getMultiProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("getMultiProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()

	b.mu.RLock()
	multi, exists := b.multiProviders[name]
	b.mu.RUnlock()

	if !exists {
		return types.NewErrorValue(fmt.Errorf("multi-provider not found: %s", name)), nil
	}

	providers := make([]types.ScriptValue, len(multi.Providers))
	for i, entry := range multi.Providers {
		providers[i] = types.NewObjectValue(map[string]types.ScriptValue{
			"name":    types.NewStringValue(entry.Name),
			"weight":  types.NewNumberValue(entry.Weight),
			"primary": types.NewBoolValue(entry.Primary),
		})
	}

	config := map[string]types.ScriptValue{
		"consensusThreshold": types.NewNumberValue(multi.Config.ConsensusThreshold),
		"timeout":            types.NewNumberValue(multi.Config.Timeout.Seconds()),
		"retryOnFailure":     types.NewBoolValue(multi.Config.RetryOnFailure),
	}

	result := map[string]types.ScriptValue{
		"name":      types.NewStringValue(multi.Name),
		"strategy":  types.NewStringValue(multi.Strategy),
		"providers": types.NewArrayValue(providers),
		"config":    types.NewObjectValue(config),
	}
	return types.NewObjectValue(result), nil
}

// Mock Provider Methods

func (b *ProvidersBridge) createMockProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("createMockProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	name := args[0].(types.StringValue).Value()
	responsesArray := args[1].ToGo().([]interface{})

	// Convert responses
	responses := make([]string, 0, len(responsesArray))
	for _, r := range responsesArray {
		responses = append(responses, fmt.Sprintf("%v", r))
	}

	// Add to providers
	b.mu.Lock()
	b.providers[name] = nil // In real implementation, this would be a mock provider
	b.metadata[name] = map[string]interface{}{
		"type":      "mock",
		"responses": responses,
	}
	b.mu.Unlock()

	// Add to LLM bridge
	if b.llmBridge != nil {
		b.llmBridge.mu.Lock()
		b.llmBridge.providers[name] = nil
		b.llmBridge.mu.Unlock()
	}

	result := map[string]types.ScriptValue{
		"name":      types.NewStringValue(name),
		"type":      types.NewStringValue("mock"),
		"responses": types.NewNumberValue(float64(len(responses))),
	}
	return types.NewObjectValue(result), nil
}

// Provider Operations

func (b *ProvidersBridge) generateWithProvider(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("generateWithProvider", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerName := args[0].(types.StringValue).Value()
	prompt := args[1].(types.StringValue).Value()

	// Check if provider exists
	b.mu.RLock()
	_, exists := b.providers[providerName]
	metadata := b.metadata[providerName]
	b.mu.RUnlock()

	if !exists {
		return types.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	// Mock response for mock providers
	if metadata != nil {
		if providerType, ok := metadata["type"].(string); ok && providerType == "mock" {
			if responses, ok := metadata["responses"].([]string); ok && len(responses) > 0 {
				// Simple cycling through responses
				responseIndex := len(prompt) % len(responses)
				return types.NewStringValue(responses[responseIndex]), nil
			}
		}
	}

	// For other providers, use LLM bridge
	if b.llmBridge != nil {
		b.llmBridge.mu.Lock()
		oldProvider := b.llmBridge.activeProvider
		b.llmBridge.activeProvider = providerName
		b.llmBridge.mu.Unlock()

		var options map[string]interface{}
		if len(args) > 2 {
			options = args[2].ToGo().(map[string]interface{})
		}

		// Use the generate method from LLM bridge
		result, err := b.llmBridge.generate(ctx, []types.ScriptValue{
			types.NewStringValue(prompt),
			types.NewObjectValue(types.ConvertMapToScriptValue(options)),
		})

		// Restore old provider
		b.llmBridge.mu.Lock()
		b.llmBridge.activeProvider = oldProvider
		b.llmBridge.mu.Unlock()

		if err != nil {
			return types.NewErrorValue(err), nil
		}

		// Extract content from result if it's an object
		if objValue, ok := result.(types.ObjectValue); ok {
			resultMap := objValue.ToGo().(map[string]interface{})
			if content, exists := resultMap["content"]; exists {
				return types.NewStringValue(fmt.Sprintf("%v", content)), nil
			}
		}

		return result, nil
	}

	return types.NewStringValue(fmt.Sprintf("Generated from %s: %s", providerName, prompt)), nil
}

// Export/Import Methods

func (b *ProvidersBridge) exportProviderConfig(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Export providers
	providers := make(map[string]types.ScriptValue)
	for name, metadata := range b.metadata {
		providers[name] = types.NewObjectValue(types.ConvertMapToScriptValue(metadata))
	}

	// Export templates
	templates := make(map[string]types.ScriptValue)
	for name, template := range b.templates {
		templates[name] = b.templateToScriptValue(template)
	}

	result := map[string]types.ScriptValue{
		"providers": types.NewObjectValue(providers),
		"templates": types.NewObjectValue(templates),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) importProviderConfig(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("importProviderConfig", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	configMap := args[0].ToGo().(map[string]interface{})

	// Import providers
	if providers, ok := configMap["providers"].(map[string]interface{}); ok {
		for name, metadata := range providers {
			b.mu.Lock()
			b.providers[name] = nil
			b.metadata[name] = metadata.(map[string]interface{})
			b.mu.Unlock()
		}
	}

	return types.NewNilValue(), nil
}

// Metadata Methods

func (b *ProvidersBridge) setProviderMetadata(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("setProviderMetadata", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerName := args[0].(types.StringValue).Value()
	metadata := args[1].ToGo().(map[string]interface{})

	b.mu.Lock()
	if _, exists := b.providers[providerName]; !exists {
		b.mu.Unlock()
		return types.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	if b.metadata[providerName] == nil {
		b.metadata[providerName] = make(map[string]interface{})
	}

	// Merge metadata
	for k, v := range metadata {
		b.metadata[providerName][k] = v
	}
	b.mu.Unlock()

	return types.NewNilValue(), nil
}

func (b *ProvidersBridge) getProviderMetadata(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("getProviderMetadata", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	providerName := args[0].(types.StringValue).Value()

	b.mu.RLock()
	metadata := b.metadata[providerName]
	b.mu.RUnlock()

	if metadata == nil {
		return types.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	result := map[string]types.ScriptValue{
		"name":     types.NewStringValue(providerName),
		"metadata": types.NewObjectValue(types.ConvertMapToScriptValue(metadata)),
	}
	return types.NewObjectValue(result), nil
}

func (b *ProvidersBridge) listProvidersByCapability(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := b.ValidateMethod("listProvidersByCapability", args); err != nil {
		return types.NewErrorValue(err), nil
	}

	capability := args[0].(types.StringValue).Value()

	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]types.ScriptValue, 0)

	// For simplicity, all providers support "generate" capability
	if capability == "generate" {
		for name := range b.providers {
			providerInfo := map[string]types.ScriptValue{
				"name":       types.NewStringValue(name),
				"capability": types.NewStringValue(capability),
			}
			providers = append(providers, types.NewObjectValue(providerInfo))
		}
	}

	return types.NewArrayValue(providers), nil
}

// Helper Methods

func (b *ProvidersBridge) templateToScriptValue(template *ProviderTemplate) types.ScriptValue {
	requiredVars := make([]types.ScriptValue, len(template.RequiredEnvVars))
	for i, v := range template.RequiredEnvVars {
		requiredVars[i] = types.NewStringValue(v)
	}

	optionalVars := make([]types.ScriptValue, len(template.OptionalEnvVars))
	for i, v := range template.OptionalEnvVars {
		optionalVars[i] = types.NewStringValue(v)
	}

	result := map[string]types.ScriptValue{
		"type":            types.NewStringValue(template.Type),
		"description":     types.NewStringValue(template.Description),
		"requiredEnvVars": types.NewArrayValue(requiredVars),
		"optionalEnvVars": types.NewArrayValue(optionalVars),
		"defaultConfig":   types.NewObjectValue(types.ConvertMapToScriptValue(template.DefaultConfig)),
	}
	return types.NewObjectValue(result)
}

// NOTE: Duplicate conversion functions removed - using centralized types.ConvertToScriptValue() instead
