// ABOUTME: LLM bridge provides access to language model providers through go-llms interfaces.
// ABOUTME: Wraps go-llms Provider interface for script engine access without reimplementation.

package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for LLM functionality
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// LLMBridge provides script access to language model functionality via go-llms.
type LLMBridge struct {
	mu             sync.RWMutex
	providers      map[string]bridge.Provider
	activeProvider string
	initialized    bool
}

// NewLLMBridge creates a new LLM bridge.
func NewLLMBridge() *LLMBridge {
	return &LLMBridge{
		providers: make(map[string]bridge.Provider),
	}
}

// GetID returns the bridge identifier.
func (b *LLMBridge) GetID() string {
	return "llm"
}

// GetMetadata returns bridge metadata.
func (b *LLMBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "llm",
		Version:     "1.0.0",
		Description: "LLM provider access bridge wrapping go-llms functionality",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *LLMBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *LLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *LLMBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *LLMBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *LLMBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "registerProvider",
			Description: "Register an LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Provider name", Required: true},
				{Name: "provider", Type: "Provider", Description: "Provider instance", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "setActiveProvider",
			Description: "Set the active LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "generate",
			Description: "Generate text using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Input prompt", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "generateMessage",
			Description: "Generate response from messages using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Description: "Input messages", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "stream",
			Description: "Stream text generation using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Input prompt", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "channel",
		},
		{
			Name:        "streamMessage",
			Description: "Stream response from messages using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Description: "Input messages", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "channel",
		},
		{
			Name:        "listProviders",
			Description: "List all registered providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getActiveProvider",
			Description: "Get the name of the active provider",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *LLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Provider": {
			GoType:     "Provider",
			ScriptType: "object",
		},
		"Message": {
			GoType:     "Message",
			ScriptType: "object",
		},
		"Response": {
			GoType:     "Response",
			ScriptType: "object",
		},
		"ProviderOptions": {
			GoType:     "ProviderOptions",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls.
func (b *LLMBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *LLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm",
			Actions:     []string{"access"},
			Description: "Access to LLM providers",
		},
	}
}

// RegisterProvider registers an LLM provider.
func (b *LLMBridge) RegisterProvider(name string, provider bridge.Provider) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.providers[name] = provider
	return nil
}

// SetActiveProvider sets the active provider.
func (b *LLMBridge) SetActiveProvider(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	b.activeProvider = name
	return nil
}

// GetActiveProvider returns the active provider.
func (b *LLMBridge) GetActiveProvider() bridge.Provider {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.activeProvider == "" {
		return nil
	}

	return b.providers[b.activeProvider]
}

// ListProviders returns all registered provider names.
func (b *LLMBridge) ListProviders() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.providers))
	for name := range b.providers {
		names = append(names, name)
	}
	return names
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *LLMBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createProvider":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProvider requires name and config parameters")
		}
		providerName, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("provider name must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Extract provider type
		providerType, ok := config["type"].(string)
		if !ok {
			return nil, fmt.Errorf("provider type required in config")
		}

		// Create provider based on type
		var llmProvider bridge.Provider

		// Get API key and model
		apiKey, _ := config["apiKey"].(string)
		model, _ := config["model"].(string)
		if model == "" {
			model = "gpt-4" // default
		}

		switch providerType {
		case "openai":
			// Create OpenAI provider
			llmProvider = provider.NewOpenAIProvider(apiKey, model)
		case "anthropic":
			// Create Anthropic provider
			llmProvider = provider.NewAnthropicProvider(apiKey, model)
		default:
			return nil, fmt.Errorf("unsupported provider type: %s", providerType)
		}

		// Store provider
		b.providers[providerName] = llmProvider
		if b.activeProvider == "" {
			b.activeProvider = providerName
		}

		return map[string]interface{}{
			"name": providerName,
			"type": providerType,
		}, nil

	case "generate":
		if len(args) < 1 {
			return nil, fmt.Errorf("generate requires prompt parameter")
		}

		// Get active provider
		if b.activeProvider == "" {
			return nil, fmt.Errorf("no active provider set")
		}
		provider, exists := b.providers[b.activeProvider]
		if !exists {
			return nil, fmt.Errorf("active provider not found: %s", b.activeProvider)
		}

		// Get prompt
		prompt, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("prompt must be string")
		}

		// Generate response
		response, err := provider.Generate(ctx, prompt, nil)
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// Return response
		return map[string]interface{}{
			"content": response,
		}, nil

	case "generateMessage":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateMessage requires messages parameter")
		}

		// Get active provider
		if b.activeProvider == "" {
			return nil, fmt.Errorf("no active provider set")
		}
		provider, exists := b.providers[b.activeProvider]
		if !exists {
			return nil, fmt.Errorf("active provider not found: %s", b.activeProvider)
		}

		// Convert messages
		var messages []llmdomain.Message
		if msgList, ok := args[0].([]interface{}); ok {
			for _, msg := range msgList {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					role, _ := msgMap["role"].(string)
					content, _ := msgMap["content"].(string)
					messages = append(messages, llmdomain.NewTextMessage(llmdomain.Role(role), content))
				}
			}
		}

		// Generate response
		response, err := provider.GenerateMessage(ctx, messages, nil)
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// Return response
		return map[string]interface{}{
			"content": response.Content,
		}, nil

	case "setActiveProvider":
		if len(args) < 1 {
			return nil, fmt.Errorf("setActiveProvider requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		if _, exists := b.providers[name]; !exists {
			return nil, fmt.Errorf("provider %s not found", name)
		}

		b.activeProvider = name
		return nil, nil

	case "getActiveProvider":
		return b.activeProvider, nil

	case "listProviders":
		providers := make([]map[string]interface{}, 0, len(b.providers))
		for name := range b.providers {
			providers = append(providers, map[string]interface{}{
				"name":   name,
				"active": name == b.activeProvider,
			})
		}
		return providers, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
