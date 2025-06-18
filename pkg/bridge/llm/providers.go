// ABOUTME: Provider management bridge for dynamic creation and configuration of LLM providers
// ABOUTME: Supports multi-provider configurations, templates, and environment-based setup

package llm

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ProviderTemplate defines a template for creating providers
type ProviderTemplate struct {
	Type            string
	Description     string
	RequiredEnvVars []string
	OptionalEnvVars []string
	DefaultConfig   map[string]interface{}
}

// MultiProvider manages multiple providers with a selection strategy
type MultiProvider struct {
	Name      string
	Providers []MultiProviderEntry
	Strategy  string // "fastest", "primary", "consensus"
	Config    MultiProviderConfig
}

// MultiProviderEntry represents a provider in a multi-provider setup
type MultiProviderEntry struct {
	Name     string
	Provider bridge.Provider
	Weight   float64
	Primary  bool
}

// MultiProviderConfig holds configuration for multi-provider
type MultiProviderConfig struct {
	ConsensusThreshold float64       // For consensus strategy
	Timeout            time.Duration // For fastest strategy
	RetryOnFailure     bool
}

// ProvidersBridge manages provider creation and configuration
type ProvidersBridge struct {
	mu             sync.RWMutex
	initialized    bool
	providers      map[string]bridge.Provider
	multiProviders map[string]*MultiProvider
	templates      map[string]*ProviderTemplate
	metadata       map[string]map[string]interface{}
	llmBridge      *LLMBridge // Reference to main LLM bridge
}

// NewProvidersBridge creates a new providers bridge
func NewProvidersBridge(llmBridge *LLMBridge) *ProvidersBridge {
	return &ProvidersBridge{
		providers:      make(map[string]bridge.Provider),
		multiProviders: make(map[string]*MultiProvider),
		templates:      initializeTemplates(),
		metadata:       make(map[string]map[string]interface{}),
		llmBridge:      llmBridge,
	}
}

// initializeTemplates creates default provider templates
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
func (b *ProvidersBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
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

	b.providers = make(map[string]bridge.Provider)
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
func (b *ProvidersBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *ProvidersBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Provider creation
		{
			Name:        "createProvider",
			Description: "Create a new provider",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "config", Type: "object", Required: true, Description: "Provider configuration"},
			},
			ReturnType: "object",
		},
		{
			Name:        "createProviderFromEnvironment",
			Description: "Create provider from environment variables",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "getProvider",
			Description: "Get a provider by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProviders",
			Description: "List all providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "removeProvider",
			Description: "Remove a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "void",
		},
		// Templates
		{
			Name:        "getProviderTemplate",
			Description: "Get provider template",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProviderTemplates",
			Description: "List available provider templates",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "validateProviderConfig",
			Description: "Validate provider configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Required: true, Description: "Provider type"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration to validate"},
			},
			ReturnType: "object",
		},
		// Multi-provider
		{
			Name:        "createMultiProvider",
			Description: "Create a multi-provider configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "providers", Type: "array", Required: true, Description: "Array of provider configurations"},
				{Name: "strategy", Type: "string", Required: true, Description: "Selection strategy"},
			},
			ReturnType: "object",
		},
		{
			Name:        "configureMultiProvider",
			Description: "Configure multi-provider settings",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
				{Name: "config", Type: "object", Required: true, Description: "Configuration object"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getMultiProvider",
			Description: "Get multi-provider information",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Multi-provider name"},
			},
			ReturnType: "object",
		},
		// Mock provider
		{
			Name:        "createMockProvider",
			Description: "Create a mock provider for testing",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "responses", Type: "array", Required: true, Description: "Array of mock responses"},
			},
			ReturnType: "object",
		},
		// Provider operations
		{
			Name:        "generateWithProvider",
			Description: "Generate text using specific provider",
			Parameters: []engine.ParameterInfo{
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
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "importProviderConfig",
			Description: "Import provider configurations",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Configuration to import"},
			},
			ReturnType: "void",
		},
		// Metadata
		{
			Name:        "setProviderMetadata",
			Description: "Set metadata for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
				{Name: "metadata", Type: "object", Required: true, Description: "Metadata object"},
			},
			ReturnType: "void",
		},
		{
			Name:        "getProviderMetadata",
			Description: "Get metadata for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
		},
		{
			Name:        "listProvidersByCapability",
			Description: "List providers by capability",
			Parameters: []engine.ParameterInfo{
				{Name: "capability", Type: "string", Required: true, Description: "Capability name"},
			},
			ReturnType: "array",
		},
	}
}

// ValidateMethod validates method parameters
func (b *ProvidersBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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
func (b *ProvidersBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	if !b.initialized {
		b.mu.RUnlock()
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
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
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// TypeMappings returns type conversion mappings
func (b *ProvidersBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"provider": {
			GoType:     "bridge.Provider",
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
func (b *ProvidersBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm.providers",
			Actions:     []string{"create", "read", "write", "delete"},
			Description: "Manage LLM providers",
		},
		{
			Type:        engine.PermissionProcess,
			Resource:    "provider.registry",
			Actions:     []string{"read", "write"},
			Description: "Access provider registry",
		},
	}
}

// Provider Management Methods

func (b *ProvidersBridge) createProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("createProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerType := args[0].(engine.StringValue).Value()
	providerName := args[1].(engine.StringValue).Value()
	configMap := args[2].ToGo().(map[string]interface{})

	// Check if provider already exists
	b.mu.RLock()
	if _, exists := b.providers[providerName]; exists {
		b.mu.RUnlock()
		return engine.NewErrorValue(fmt.Errorf("provider %s already exists", providerName)), nil
	}
	b.mu.RUnlock()

	// Get template
	template, exists := b.templates[providerType]
	if !exists {
		return engine.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
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

	result := map[string]engine.ScriptValue{
		"name":    engine.NewStringValue(providerName),
		"type":    engine.NewStringValue(providerType),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) createProviderFromEnvironment(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("createProviderFromEnvironment", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerType := args[0].(engine.StringValue).Value()
	providerName := args[1].(engine.StringValue).Value()

	template, exists := b.templates[providerType]
	if !exists {
		return engine.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
	}

	// Check required env vars
	config := make(map[string]interface{})
	for _, envVar := range template.RequiredEnvVars {
		value := os.Getenv(envVar)
		if value == "" {
			return engine.NewErrorValue(fmt.Errorf("required environment variable %s not set", envVar)), nil
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

	result := map[string]engine.ScriptValue{
		"name":    engine.NewStringValue(providerName),
		"type":    engine.NewStringValue(providerType),
		"source":  engine.NewStringValue("environment"),
		"created": engine.NewStringValue(time.Now().Format(time.RFC3339)),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) getProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	_, exists := b.providers[name]
	metadata := b.metadata[name]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("provider not found: %s", name)), nil
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

	result := map[string]engine.ScriptValue{
		"name": engine.NewStringValue(name),
		"type": engine.NewStringValue(providerType),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) listProviders(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]engine.ScriptValue, 0, len(b.providers))
	for name := range b.providers {
		providers = append(providers, engine.NewStringValue(name))
	}

	return engine.NewArrayValue(providers), nil
}

func (b *ProvidersBridge) removeProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("removeProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

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

	return engine.NewNilValue(), nil
}

// Template Methods

func (b *ProvidersBridge) getProviderTemplate(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getProviderTemplate", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerType := args[0].(engine.StringValue).Value()

	template, exists := b.templates[providerType]
	if !exists {
		return engine.NewErrorValue(fmt.Errorf("template not found: %s", providerType)), nil
	}

	return b.templateToScriptValue(template), nil
}

func (b *ProvidersBridge) listProviderTemplates(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	templates := make([]engine.ScriptValue, 0, len(b.templates))
	for _, template := range b.templates {
		templates = append(templates, b.templateToScriptValue(template))
	}

	return engine.NewArrayValue(templates), nil
}

func (b *ProvidersBridge) validateProviderConfig(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("validateProviderConfig", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerType := args[0].(engine.StringValue).Value()
	configMap := args[1].ToGo().(map[string]interface{})

	template, exists := b.templates[providerType]
	if !exists {
		return engine.NewErrorValue(fmt.Errorf("unknown provider type: %s", providerType)), nil
	}

	errors := []string{}

	// Check required vars
	for _, reqVar := range template.RequiredEnvVars {
		if _, ok := configMap[reqVar]; !ok {
			errors = append(errors, fmt.Sprintf("missing required field: %s", reqVar))
		}
	}

	valid := len(errors) == 0

	errorsArray := make([]engine.ScriptValue, len(errors))
	for i, err := range errors {
		errorsArray[i] = engine.NewStringValue(err)
	}

	result := map[string]engine.ScriptValue{
		"valid":  engine.NewBoolValue(valid),
		"errors": engine.NewArrayValue(errorsArray),
	}

	return engine.NewObjectValue(result), nil
}

// Multi-Provider Methods

func (b *ProvidersBridge) createMultiProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("createMultiProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()
	providersArray := args[1].ToGo().([]interface{})
	strategy := args[2].(engine.StringValue).Value()

	// Validate strategy
	validStrategies := map[string]bool{
		"fastest":   true,
		"primary":   true,
		"consensus": true,
	}
	if !validStrategies[strategy] {
		return engine.NewErrorValue(fmt.Errorf("invalid strategy: %s", strategy)), nil
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
			return engine.NewErrorValue(fmt.Errorf("provider not found: %s", entry.Name)), nil
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

	result := map[string]engine.ScriptValue{
		"name":      engine.NewStringValue(name),
		"strategy":  engine.NewStringValue(strategy),
		"providers": engine.NewNumberValue(float64(len(entries))),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) configureMultiProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("configureMultiProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()
	configMap := args[1].ToGo().(map[string]interface{})

	b.mu.Lock()
	multi, exists := b.multiProviders[name]
	if !exists {
		b.mu.Unlock()
		return engine.NewErrorValue(fmt.Errorf("multi-provider not found: %s", name)), nil
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

	return engine.NewNilValue(), nil
}

func (b *ProvidersBridge) getMultiProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getMultiProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	multi, exists := b.multiProviders[name]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("multi-provider not found: %s", name)), nil
	}

	providers := make([]engine.ScriptValue, len(multi.Providers))
	for i, entry := range multi.Providers {
		providers[i] = engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":    engine.NewStringValue(entry.Name),
			"weight":  engine.NewNumberValue(entry.Weight),
			"primary": engine.NewBoolValue(entry.Primary),
		})
	}

	config := map[string]engine.ScriptValue{
		"consensusThreshold": engine.NewNumberValue(multi.Config.ConsensusThreshold),
		"timeout":            engine.NewNumberValue(multi.Config.Timeout.Seconds()),
		"retryOnFailure":     engine.NewBoolValue(multi.Config.RetryOnFailure),
	}

	result := map[string]engine.ScriptValue{
		"name":      engine.NewStringValue(multi.Name),
		"strategy":  engine.NewStringValue(multi.Strategy),
		"providers": engine.NewArrayValue(providers),
		"config":    engine.NewObjectValue(config),
	}
	return engine.NewObjectValue(result), nil
}

// Mock Provider Methods

func (b *ProvidersBridge) createMockProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("createMockProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	name := args[0].(engine.StringValue).Value()
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

	result := map[string]engine.ScriptValue{
		"name":      engine.NewStringValue(name),
		"type":      engine.NewStringValue("mock"),
		"responses": engine.NewNumberValue(float64(len(responses))),
	}
	return engine.NewObjectValue(result), nil
}

// Provider Operations

func (b *ProvidersBridge) generateWithProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("generateWithProvider", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerName := args[0].(engine.StringValue).Value()
	prompt := args[1].(engine.StringValue).Value()

	// Check if provider exists
	b.mu.RLock()
	_, exists := b.providers[providerName]
	metadata := b.metadata[providerName]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	// Mock response for mock providers
	if metadata != nil {
		if providerType, ok := metadata["type"].(string); ok && providerType == "mock" {
			if responses, ok := metadata["responses"].([]string); ok && len(responses) > 0 {
				// Simple cycling through responses
				responseIndex := len(prompt) % len(responses)
				return engine.NewStringValue(responses[responseIndex]), nil
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
		result, err := b.llmBridge.generate(ctx, []engine.ScriptValue{
			engine.NewStringValue(prompt),
			engine.NewObjectValue(providersConvertMapToScriptValue(options)),
		})

		// Restore old provider
		b.llmBridge.mu.Lock()
		b.llmBridge.activeProvider = oldProvider
		b.llmBridge.mu.Unlock()

		if err != nil {
			return engine.NewErrorValue(err), nil
		}

		// Extract content from result if it's an object
		if objValue, ok := result.(engine.ObjectValue); ok {
			resultMap := objValue.ToGo().(map[string]interface{})
			if content, exists := resultMap["content"]; exists {
				return engine.NewStringValue(fmt.Sprintf("%v", content)), nil
			}
		}

		return result, nil
	}

	return engine.NewStringValue(fmt.Sprintf("Generated from %s: %s", providerName, prompt)), nil
}

// Export/Import Methods

func (b *ProvidersBridge) exportProviderConfig(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Export providers
	providers := make(map[string]engine.ScriptValue)
	for name, metadata := range b.metadata {
		providers[name] = engine.NewObjectValue(providersConvertMapToScriptValue(metadata))
	}

	// Export templates
	templates := make(map[string]engine.ScriptValue)
	for name, template := range b.templates {
		templates[name] = b.templateToScriptValue(template)
	}

	result := map[string]engine.ScriptValue{
		"providers": engine.NewObjectValue(providers),
		"templates": engine.NewObjectValue(templates),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) importProviderConfig(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("importProviderConfig", args); err != nil {
		return engine.NewErrorValue(err), nil
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

	return engine.NewNilValue(), nil
}

// Metadata Methods

func (b *ProvidersBridge) setProviderMetadata(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("setProviderMetadata", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerName := args[0].(engine.StringValue).Value()
	metadata := args[1].ToGo().(map[string]interface{})

	b.mu.Lock()
	if _, exists := b.providers[providerName]; !exists {
		b.mu.Unlock()
		return engine.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	if b.metadata[providerName] == nil {
		b.metadata[providerName] = make(map[string]interface{})
	}

	// Merge metadata
	for k, v := range metadata {
		b.metadata[providerName][k] = v
	}
	b.mu.Unlock()

	return engine.NewNilValue(), nil
}

func (b *ProvidersBridge) getProviderMetadata(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("getProviderMetadata", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	providerName := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	metadata := b.metadata[providerName]
	b.mu.RUnlock()

	if metadata == nil {
		return engine.NewErrorValue(fmt.Errorf("provider not found: %s", providerName)), nil
	}

	result := map[string]engine.ScriptValue{
		"name":     engine.NewStringValue(providerName),
		"metadata": engine.NewObjectValue(providersConvertMapToScriptValue(metadata)),
	}
	return engine.NewObjectValue(result), nil
}

func (b *ProvidersBridge) listProvidersByCapability(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod("listProvidersByCapability", args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	capability := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]engine.ScriptValue, 0)

	// For simplicity, all providers support "generate" capability
	if capability == "generate" {
		for name := range b.providers {
			providerInfo := map[string]engine.ScriptValue{
				"name":       engine.NewStringValue(name),
				"capability": engine.NewStringValue(capability),
			}
			providers = append(providers, engine.NewObjectValue(providerInfo))
		}
	}

	return engine.NewArrayValue(providers), nil
}

// Helper Methods

func (b *ProvidersBridge) templateToScriptValue(template *ProviderTemplate) engine.ScriptValue {
	requiredVars := make([]engine.ScriptValue, len(template.RequiredEnvVars))
	for i, v := range template.RequiredEnvVars {
		requiredVars[i] = engine.NewStringValue(v)
	}

	optionalVars := make([]engine.ScriptValue, len(template.OptionalEnvVars))
	for i, v := range template.OptionalEnvVars {
		optionalVars[i] = engine.NewStringValue(v)
	}

	result := map[string]engine.ScriptValue{
		"type":            engine.NewStringValue(template.Type),
		"description":     engine.NewStringValue(template.Description),
		"requiredEnvVars": engine.NewArrayValue(requiredVars),
		"optionalEnvVars": engine.NewArrayValue(optionalVars),
		"defaultConfig":   engine.NewObjectValue(providersConvertMapToScriptValue(template.DefaultConfig)),
	}
	return engine.NewObjectValue(result)
}

// providersConvertMapToScriptValue converts map[string]interface{} to map[string]engine.ScriptValue
func providersConvertMapToScriptValue(m map[string]interface{}) map[string]engine.ScriptValue {
	if m == nil {
		return make(map[string]engine.ScriptValue)
	}

	result := make(map[string]engine.ScriptValue)
	for k, v := range m {
		result[k] = providersConvertToScriptValue(v)
	}
	return result
}

// providersConvertToScriptValue converts interface{} to ScriptValue
func providersConvertToScriptValue(v interface{}) engine.ScriptValue {
	if v == nil {
		return engine.NewNilValue()
	}

	switch val := v.(type) {
	case bool:
		return engine.NewBoolValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int32:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case float32:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case string:
		return engine.NewStringValue(val)
	case []string:
		// Convert string slice
		arr := make([]engine.ScriptValue, len(val))
		for i, s := range val {
			arr[i] = engine.NewStringValue(s)
		}
		return engine.NewArrayValue(arr)
	case []interface{}:
		arr := make([]engine.ScriptValue, len(val))
		for i, item := range val {
			arr[i] = providersConvertToScriptValue(item)
		}
		return engine.NewArrayValue(arr)
	case map[string]interface{}:
		return engine.NewObjectValue(providersConvertMapToScriptValue(val))
	default:
		// Convert to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}
