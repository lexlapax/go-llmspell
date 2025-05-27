// ABOUTME: Bridge between script engines and go-llms library
// ABOUTME: Provides script-accessible wrappers for LLM functionality

package bridge

import (
	"context"
	"fmt"
	"os"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// LLMBridge provides script access to LLM functionality
type LLMBridge struct {
	provider domain.Provider
}

// NewLLMBridge creates a new bridge instance
func NewLLMBridge() (*LLMBridge, error) {
	// Detect provider from environment
	providerName := ""
	if os.Getenv("OPENAI_API_KEY") != "" {
		providerName = "openai"
	} else if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providerName = "anthropic"
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		providerName = "gemini"
	} else {
		return nil, fmt.Errorf("no API key found in environment (OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY)")
	}

	// Create provider with config from environment
	config := llmutil.ModelConfig{
		Provider: providerName,
	}

	provider, err := llmutil.CreateProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &LLMBridge{
		provider: provider,
	}, nil
}

// Chat sends a chat message to the LLM
func (b *LLMBridge) Chat(ctx context.Context, prompt string) (string, error) {
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

	response, err := b.provider.GenerateMessage(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	return response.Content, nil
}

// Complete generates text completion
func (b *LLMBridge) Complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// Use Generate method with options
	options := []domain.Option{}
	if maxTokens > 0 {
		options = append(options, domain.WithMaxTokens(maxTokens))
	}

	response, err := b.provider.Generate(ctx, prompt, options...)
	if err != nil {
		return "", fmt.Errorf("completion failed: %w", err)
	}

	return response, nil
}

// StreamChat sends a chat message and streams the response
func (b *LLMBridge) StreamChat(ctx context.Context, prompt string, callback func(chunk string) error) error {
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
	stream, err := b.provider.StreamMessage(ctx, messages)
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
