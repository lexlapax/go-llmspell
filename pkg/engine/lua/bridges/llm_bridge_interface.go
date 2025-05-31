// ABOUTME: Interface definition for LLM bridge operations used by Lua bridge
// ABOUTME: Allows for easier testing by defining the contract needed

package bridges

import (
	"context"
)

// LLMBridgeInterface defines the methods needed by the Lua LLM bridge
type LLMBridgeInterface interface {
	// Chat sends a chat message to the LLM
	Chat(ctx context.Context, prompt string) (string, error)

	// Complete generates text completion
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)

	// StreamChat sends a chat message and streams the response
	StreamChat(ctx context.Context, prompt string, callback func(chunk string) error) error

	// ListModels returns available models
	ListModels(ctx context.Context) ([]map[string]interface{}, error)

	// ListProviders returns a list of available provider names
	ListProviders() []string

	// GetCurrentProvider returns the name of the current provider
	GetCurrentProvider() string

	// SetProvider switches to a different provider
	SetProvider(name string) error
}
