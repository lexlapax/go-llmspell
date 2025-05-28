// ABOUTME: Bridge between script engines and go-llms library
// ABOUTME: Provides script-accessible wrappers for LLM functionality

package bridge

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
	modelinfodomain "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// LLMBridge provides script access to LLM functionality
type LLMBridge struct {
	providers map[string]domain.Provider
	mu        sync.RWMutex
	current   string // current provider name
}

// NewLLMBridge creates a new bridge instance
func NewLLMBridge() (*LLMBridge, error) {
	bridge := &LLMBridge{
		providers: make(map[string]domain.Provider),
	}

	// Auto-detect and initialize available providers from environment
	availableProviders := []string{}

	// Check OpenAI
	if os.Getenv("OPENAI_API_KEY") != "" {
		if err := bridge.initProvider("openai"); err == nil {
			availableProviders = append(availableProviders, "openai")
		}
	}

	// Check Anthropic
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		if err := bridge.initProvider("anthropic"); err == nil {
			availableProviders = append(availableProviders, "anthropic")
		}
	}

	// Check Gemini
	if os.Getenv("GEMINI_API_KEY") != "" {
		if err := bridge.initProvider("gemini"); err == nil {
			availableProviders = append(availableProviders, "gemini")
		}
	}

	if len(availableProviders) == 0 {
		return nil, fmt.Errorf("no API key found in environment (OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY)")
	}

	// Set the first available provider as current
	bridge.current = availableProviders[0]

	return bridge, nil
}

// initProvider initializes a provider by name
func (b *LLMBridge) initProvider(name string) error {
	config := llmutil.ModelConfig{
		Provider: name,
	}

	provider, err := llmutil.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create %s provider: %w", name, err)
	}

	b.mu.Lock()
	b.providers[name] = provider
	b.mu.Unlock()

	return nil
}

// SetProvider switches to a different provider
func (b *LLMBridge) SetProvider(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.providers[name]; !exists {
		return fmt.Errorf("provider '%s' not available", name)
	}

	b.current = name
	return nil
}

// GetCurrentProvider returns the name of the current provider
func (b *LLMBridge) GetCurrentProvider() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current
}

// ListProviders returns a list of available provider names
func (b *LLMBridge) ListProviders() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]string, 0, len(b.providers))
	for name := range b.providers {
		providers = append(providers, name)
	}
	return providers
}

// getProvider returns the current provider
func (b *LLMBridge) getProvider() (domain.Provider, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	provider, exists := b.providers[b.current]
	if !exists {
		return nil, fmt.Errorf("current provider '%s' not found", b.current)
	}
	return provider, nil
}

// Chat sends a chat message to the LLM
func (b *LLMBridge) Chat(ctx context.Context, prompt string) (string, error) {
	provider, err := b.getProvider()
	if err != nil {
		return "", err
	}

	messages := []domain.Message{
		{
			Role: domain.RoleUser,
			Content: []domain.ContentPart{
				{
					Type: domain.ContentTypeText,
					Text: prompt,
				},
			},
		},
	}

	response, err := provider.GenerateMessage(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	return response.Content, nil
}

// Complete generates text completion
func (b *LLMBridge) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	provider, err := b.getProvider()
	if err != nil {
		return "", err
	}

	// Use Generate method with options
	options := []domain.Option{}
	if maxTokens > 0 {
		options = append(options, domain.WithMaxTokens(maxTokens))
	}

	response, err := provider.Generate(ctx, prompt, options...)
	if err != nil {
		return "", fmt.Errorf("completion failed: %w", err)
	}

	return response, nil
}

// StreamChat sends a chat message and streams the response
func (b *LLMBridge) StreamChat(ctx context.Context, prompt string, callback func(chunk string) error) error {
	provider, err := b.getProvider()
	if err != nil {
		return err
	}

	// Create message for streaming
	messages := []domain.Message{
		{
			Role: domain.RoleUser,
			Content: []domain.ContentPart{
				{
					Type: domain.ContentTypeText,
					Text: prompt,
				},
			},
		},
	}

	// Start streaming
	stream, err := provider.StreamMessage(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Process stream chunks from channel
	for token := range stream {
		if err := callback(token.Text); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}

		if token.Finished {
			break
		}
	}

	return nil
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    string            `json:"provider"`
	ContextSize int               `json:"context_size,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ListModels returns a list of available models from all providers
func (b *LLMBridge) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// Get model inventory from go-llms
	opts := &llmutil.GetAvailableModelsOptions{
		UseCache: true,
	}

	inventory, err := llmutil.GetAvailableModels(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}

	models := []ModelInfo{}

	// Convert inventory models to our ModelInfo format
	for _, model := range inventory.Models {
		models = append(models, b.convertToModelInfo(model))
	}

	return models, nil
}

// convertToModelInfo converts a modelinfodomain.Model to ModelInfo
func (b *LLMBridge) convertToModelInfo(model modelinfodomain.Model) ModelInfo {
	info := ModelInfo{
		ID:       model.Name,
		Name:     model.DisplayName,
		Provider: model.Provider,
		Metadata: make(map[string]string),
	}

	// Add context size if available
	if model.ContextWindow > 0 {
		info.ContextSize = model.ContextWindow
	}

	// Add other metadata
	if model.Description != "" {
		info.Metadata["description"] = model.Description
	}

	if model.TrainingCutoff != "" {
		info.Metadata["training_cutoff"] = model.TrainingCutoff
	}

	if model.ModelFamily != "" {
		info.Metadata["model_family"] = model.ModelFamily
	}

	// Add capabilities info
	caps := []string{}
	if model.Capabilities.FunctionCalling {
		caps = append(caps, "function_calling")
	}
	if model.Capabilities.Streaming {
		caps = append(caps, "streaming")
	}
	if model.Capabilities.JSONMode {
		caps = append(caps, "json_mode")
	}
	if len(caps) > 0 {
		info.Metadata["capabilities"] = strings.Join(caps, ",")
	}

	return info
}

// ListModelsForProvider returns models for a specific provider
func (b *LLMBridge) ListModelsForProvider(ctx context.Context, provider string) ([]ModelInfo, error) {
	models, err := b.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// Filter models by provider
	filtered := []ModelInfo{}
	for _, model := range models {
		if model.Provider == provider {
			filtered = append(filtered, model)
		}
	}

	return filtered, nil
}

// Implement Bridge interface

// Name returns the name of this bridge
func (b *LLMBridge) Name() string {
	return "llm"
}

// Methods returns information about all methods exposed by this bridge
func (b *LLMBridge) Methods() []MethodInfo {
	return []MethodInfo{
		{
			Name:        "chat",
			Description: "Send a chat message to the LLM and get a response",
			Parameters: []ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "The message to send"},
			},
			ReturnType: "string",
			IsAsync:    false,
		},
		{
			Name:        "complete",
			Description: "Generate text completion with optional token limit",
			Parameters: []ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "The text prompt"},
				{Name: "maxTokens", Type: "number", Required: false, Description: "Maximum tokens to generate"},
			},
			ReturnType: "string",
			IsAsync:    false,
		},
		{
			Name:        "streamChat",
			Description: "Send a chat message and stream the response",
			Parameters: []ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "The message to send"},
				{Name: "callback", Type: "function", Required: true, Description: "Function to handle stream chunks"},
			},
			ReturnType: "void",
			IsAsync:    true,
		},
		{
			Name:        "setProvider",
			Description: "Switch to a different LLM provider",
			Parameters: []ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name (openai, anthropic, gemini)"},
			},
			ReturnType: "void",
			IsAsync:    false,
		},
		{
			Name:        "getCurrentProvider",
			Description: "Get the name of the current provider",
			Parameters:  []ParameterInfo{},
			ReturnType:  "string",
			IsAsync:     false,
		},
		{
			Name:        "listProviders",
			Description: "List all available providers",
			Parameters:  []ParameterInfo{},
			ReturnType:  "string[]",
			IsAsync:     false,
		},
		{
			Name:        "listModels",
			Description: "List all available models from all providers",
			Parameters:  []ParameterInfo{},
			ReturnType:  "ModelInfo[]",
			IsAsync:     false,
		},
		{
			Name:        "listModelsForProvider",
			Description: "List models for a specific provider",
			Parameters: []ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "ModelInfo[]",
			IsAsync:    false,
		},
	}
}

// Initialize prepares the bridge for use
func (b *LLMBridge) Initialize(ctx context.Context) error {
	// Bridge is already initialized in NewLLMBridge
	return nil
}

// Cleanup releases any resources held by the bridge
func (b *LLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clear providers
	b.providers = nil
	b.current = ""

	return nil
}
