// ABOUTME: Adapter to convert bridge.LLMBridge to LLMBridgeInterface
// ABOUTME: Handles type conversions between bridge types and Lua-friendly types

package bridges

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
)

// LLMBridgeAdapter adapts bridge.LLMBridge to LLMBridgeInterface
type LLMBridgeAdapter struct {
	bridge *bridge.LLMBridge
}

// NewLLMBridgeAdapter creates a new adapter
func NewLLMBridgeAdapter(b *bridge.LLMBridge) *LLMBridgeAdapter {
	return &LLMBridgeAdapter{bridge: b}
}

// Chat sends a chat message to the LLM
func (a *LLMBridgeAdapter) Chat(ctx context.Context, prompt string) (string, error) {
	return a.bridge.Chat(ctx, prompt)
}

// Complete generates text completion
func (a *LLMBridgeAdapter) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	return a.bridge.Complete(ctx, prompt, maxTokens)
}

// StreamChat sends a chat message and streams the response
func (a *LLMBridgeAdapter) StreamChat(ctx context.Context, prompt string, callback func(chunk string) error) error {
	return a.bridge.StreamChat(ctx, prompt, callback)
}

// ListModels returns available models - converts ModelInfo to map[string]interface{}
func (a *LLMBridgeAdapter) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	models, err := a.bridge.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// Convert ModelInfo to map[string]interface{}
	result := make([]map[string]interface{}, 0, len(models))
	for _, model := range models {
		m := map[string]interface{}{
			"id":       model.ID,
			"name":     model.Name,
			"provider": model.Provider,
		}

		if model.ContextSize > 0 {
			m["context_size"] = model.ContextSize
		}

		if len(model.Metadata) > 0 {
			m["metadata"] = model.Metadata
		}

		result = append(result, m)
	}

	return result, nil
}

// ListProviders returns a list of available provider names
func (a *LLMBridgeAdapter) ListProviders() []string {
	return a.bridge.ListProviders()
}

// GetCurrentProvider returns the name of the current provider
func (a *LLMBridgeAdapter) GetCurrentProvider() string {
	return a.bridge.GetCurrentProvider()
}

// SetProvider switches to a different provider
func (a *LLMBridgeAdapter) SetProvider(name string) error {
	return a.bridge.SetProvider(name)
}
