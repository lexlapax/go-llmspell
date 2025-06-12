// ABOUTME: Core LLM bridge provides access to language model providers through a unified interface.
// ABOUTME: Supports multiple providers, streaming responses, message handling, and provider switching.

package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// LLMMessage represents a message in an LLM conversation.
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMProvider defines the interface for language model providers.
type LLMProvider interface {
	Complete(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (string, error)
	CompleteStream(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (<-chan string, <-chan error)
	GetName() string
	GetModel() string
	IsAvailable() bool
}

// LLMBridge provides script access to language model functionality.
type LLMBridge struct {
	mu             sync.RWMutex
	providers      map[string]LLMProvider
	activeProvider string
	initialized    bool
	messageHistory []LLMMessage
	maxHistorySize int
}

// NewLLMBridge creates a new LLM bridge.
func NewLLMBridge() *LLMBridge {
	return &LLMBridge{
		providers:      make(map[string]LLMProvider),
		maxHistorySize: 100,
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
		Description: "Core LLM provider access bridge for language model interactions",
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
	b.messageHistory = nil
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
			Name:        "complete",
			Description: "Complete a conversation with the LLM",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Required: true, Description: "Array of message objects"},
				{Name: "options", Type: "object", Required: false, Description: "Completion options"},
			},
			ReturnType: "string",
		},
		{
			Name:        "completeStream",
			Description: "Complete a conversation with streaming response",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Required: true, Description: "Array of message objects"},
				{Name: "options", Type: "object", Required: false, Description: "Completion options"},
			},
			ReturnType: "stream",
		},
		{
			Name:        "setProvider",
			Description: "Set the active LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "void",
		},
		{
			Name:        "listProviders",
			Description: "List available LLM providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "createMessage",
			Description: "Create a new message object",
			Parameters: []engine.ParameterInfo{
				{Name: "role", Type: "string", Required: true, Description: "Message role (system/user/assistant)"},
				{Name: "content", Type: "string", Required: true, Description: "Message content"},
			},
			ReturnType: "object",
		},
	}
}

// ValidateMethod validates method parameters.
func (b *LLMBridge) ValidateMethod(name string, args []interface{}) error {
	switch name {
	case "complete", "completeStream":
		if len(args) < 1 {
			return fmt.Errorf("missing messages argument")
		}
		// TODO: Validate message structure
		return nil
	case "setProvider":
		if len(args) < 1 {
			return fmt.Errorf("missing provider argument")
		}
		return nil
	case "createMessage":
		if len(args) < 2 {
			return fmt.Errorf("missing role or content argument")
		}
		return nil
	}
	return nil
}

// TypeMappings returns type conversion mappings.
func (b *LLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"LLMMessage": {
			GoType:     "bridge.LLMMessage",
			ScriptType: "object",
			Converter:  "standard",
		},
		"CompletionOptions": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
			Converter:  "standard",
		},
	}
}

// RequiredPermissions returns required permissions.
func (b *LLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm-api",
			Actions:     []string{"connect", "send", "receive"},
			Description: "Network access for LLM API calls",
		},
	}
}

// RegisterProvider registers an LLM provider.
func (b *LLMBridge) RegisterProvider(name string, provider LLMProvider) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	b.providers[name] = provider
	return nil
}

// GetProvider retrieves a provider by name.
func (b *LLMBridge) GetProvider(name string) (LLMProvider, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	provider, exists := b.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns a list of registered provider names.
func (b *LLMBridge) ListProviders() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.providers))
	for name := range b.providers {
		names = append(names, name)
	}

	return names
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

// GetActiveProvider returns the active provider name.
func (b *LLMBridge) GetActiveProvider() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.activeProvider
}

// Complete performs LLM completion.
func (b *LLMBridge) Complete(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages provided")
	}

	b.mu.RLock()
	providerName := b.activeProvider
	provider := b.providers[providerName]
	b.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("no active provider set")
	}

	// Validate messages
	for _, msg := range messages {
		if err := b.ValidateMessage(msg); err != nil {
			return "", err
		}
	}

	// Call provider
	response, err := provider.Complete(ctx, messages, options)
	if err != nil {
		return "", fmt.Errorf("provider error: %w", err)
	}

	// Update message history
	b.updateHistory(messages)

	return response, nil
}

// CompleteStream performs streaming LLM completion.
func (b *LLMBridge) CompleteStream(ctx context.Context, messages []LLMMessage, options map[string]interface{}) (<-chan string, <-chan error) {
	chunkChan := make(chan string)
	errorChan := make(chan error, 1)

	if len(messages) == 0 {
		errorChan <- fmt.Errorf("no messages provided")
		close(chunkChan)
		close(errorChan)
		return chunkChan, errorChan
	}

	b.mu.RLock()
	providerName := b.activeProvider
	provider := b.providers[providerName]
	b.mu.RUnlock()

	if provider == nil {
		errorChan <- fmt.Errorf("no active provider set")
		close(chunkChan)
		close(errorChan)
		return chunkChan, errorChan
	}

	// Validate messages
	for _, msg := range messages {
		if err := b.ValidateMessage(msg); err != nil {
			errorChan <- err
			close(chunkChan)
			close(errorChan)
			return chunkChan, errorChan
		}
	}

	// Call provider
	providerChunks, providerErrors := provider.CompleteStream(ctx, messages, options)

	// Forward chunks and errors
	go func() {
		defer close(chunkChan)
		defer close(errorChan)

		for {
			select {
			case chunk, ok := <-providerChunks:
				if !ok {
					return
				}
				select {
				case chunkChan <- chunk:
				case <-ctx.Done():
					errorChan <- ctx.Err()
					return
				}
			case err := <-providerErrors:
				if err != nil {
					errorChan <- err
					return
				}
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			}
		}
	}()

	// Update message history
	b.updateHistory(messages)

	return chunkChan, errorChan
}

// CreateMessage creates a new message.
func (b *LLMBridge) CreateMessage(role, content string) LLMMessage {
	return LLMMessage{
		Role:    role,
		Content: content,
	}
}

// ValidateMessage validates a message.
func (b *LLMBridge) ValidateMessage(msg LLMMessage) error {
	validRoles := map[string]bool{
		"system":    true,
		"user":      true,
		"assistant": true,
	}

	if !validRoles[msg.Role] {
		return fmt.Errorf("invalid role: %s", msg.Role)
	}

	if msg.Content == "" {
		return fmt.Errorf("empty content for role %s", msg.Role)
	}

	return nil
}

// IsProviderAvailable checks if a provider is available.
func (b *LLMBridge) IsProviderAvailable(name string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	provider, exists := b.providers[name]
	if !exists {
		return false
	}

	return provider.IsAvailable()
}

// updateHistory updates the message history.
func (b *LLMBridge) updateHistory(messages []LLMMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messageHistory = append(b.messageHistory, messages...)

	// Trim history if it exceeds max size
	if len(b.messageHistory) > b.maxHistorySize {
		start := len(b.messageHistory) - b.maxHistorySize
		b.messageHistory = b.messageHistory[start:]
	}
}
