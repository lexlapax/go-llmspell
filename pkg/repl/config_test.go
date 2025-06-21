// ABOUTME: Tests for config integration utilities ensuring proper conversion from main config.
// ABOUTME: Validates that REPL settings are correctly mapped from the application configuration.

package repl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewREPLConfigFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		engine   string
		expected REPLConfig
	}{
		{
			name:   "default configuration",
			cfg:    config.GetDefaultConfig(),
			engine: "",
			expected: REPLConfig{
				Engine:          "lua",
				Prompt:          "lua> ",
				ContinuePrompt:  "... ",
				HistoryFile:     filepath.Join(os.Getenv("HOME"), ".llmspell_history"),
				HistorySize:     1000,
				SaveHistory:     true,
				SyntaxHighlight: true,
				AutoComplete:    true,
				MultiLine:       true,
				Input:           os.Stdin,
				Output:          os.Stdout,
				Error:           os.Stderr,
			},
		},
		{
			name:   "custom engine specified",
			cfg:    config.GetDefaultConfig(),
			engine: "javascript",
			expected: REPLConfig{
				Engine:          "javascript",
				Prompt:          "javascript> ",
				ContinuePrompt:  "... ",
				HistoryFile:     filepath.Join(os.Getenv("HOME"), ".llmspell_history"),
				HistorySize:     1000,
				SaveHistory:     true,
				SyntaxHighlight: true,
				AutoComplete:    true,
				MultiLine:       true,
				Input:           os.Stdin,
				Output:          os.Stdout,
				Error:           os.Stderr,
			},
		},
		{
			name: "disabled features",
			cfg: &config.Config{
				Engine: config.EngineConfig{
					Default: "tengo",
				},
				REPL: config.REPLConfig{
					Prompt:          "custom> ",
					ContinuePrompt:  ">>> ",
					HistoryFile:     "/tmp/history",
					HistorySize:     500,
					SaveHistory:     false,
					SyntaxHighlight: false,
					AutoComplete:    false,
					MultiLine:       false,
					DefaultEngine:   "lua",
				},
			},
			engine: "",
			expected: REPLConfig{
				Engine:          "lua",
				Prompt:          "custom> ",
				ContinuePrompt:  ">>> ",
				HistoryFile:     "/tmp/history",
				HistorySize:     500,
				SaveHistory:     false,
				SyntaxHighlight: false,
				AutoComplete:    false,
				MultiLine:       false,
				Input:           os.Stdin,
				Output:          os.Stdout,
				Error:           os.Stderr,
			},
		},
		{
			name: "prompt with engine name",
			cfg: &config.Config{
				Engine: config.EngineConfig{
					Default: "lua",
				},
				REPL: config.REPLConfig{
					Prompt:          "llmspell [lua]> ",
					ContinuePrompt:  "... ",
					HistoryFile:     "~/.llmspell_history",
					HistorySize:     1000,
					SaveHistory:     true,
					SyntaxHighlight: true,
					AutoComplete:    true,
					MultiLine:       true,
				},
			},
			engine: "lua",
			expected: REPLConfig{
				Engine:          "lua",
				Prompt:          "llmspell [lua]> ",
				ContinuePrompt:  "... ",
				HistoryFile:     filepath.Join(os.Getenv("HOME"), ".llmspell_history"),
				HistorySize:     1000,
				SaveHistory:     true,
				SyntaxHighlight: true,
				AutoComplete:    true,
				MultiLine:       true,
				Input:           os.Stdin,
				Output:          os.Stdout,
				Error:           os.Stderr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewREPLConfigFromConfig(tt.cfg, tt.engine)

			assert.Equal(t, tt.expected.Engine, result.Engine)
			assert.Equal(t, tt.expected.Prompt, result.Prompt)
			assert.Equal(t, tt.expected.ContinuePrompt, result.ContinuePrompt)
			assert.Equal(t, tt.expected.HistoryFile, result.HistoryFile)
			assert.Equal(t, tt.expected.HistorySize, result.HistorySize)
			assert.Equal(t, tt.expected.SaveHistory, result.SaveHistory)
			assert.Equal(t, tt.expected.SyntaxHighlight, result.SyntaxHighlight)
			assert.Equal(t, tt.expected.AutoComplete, result.AutoComplete)
			assert.Equal(t, tt.expected.MultiLine, result.MultiLine)
			assert.Equal(t, tt.expected.Input, result.Input)
			assert.Equal(t, tt.expected.Output, result.Output)
			assert.Equal(t, tt.expected.Error, result.Error)
		})
	}
}

func TestExpandPath(t *testing.T) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "absolute path",
			input:    "/tmp/history",
			expected: "/tmp/history",
		},
		{
			name:     "tilde path",
			input:    "~/.llmspell_history",
			expected: filepath.Join(homeDir, ".llmspell_history"),
		},
		{
			name:     "tilde with subdirectory",
			input:    "~/.config/llmspell/history",
			expected: filepath.Join(homeDir, ".config/llmspell/history"),
		},
		{
			name:     "relative path",
			input:    "./history",
			expected: "./history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
