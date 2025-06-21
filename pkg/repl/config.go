// ABOUTME: Config integration utilities for converting between main config and REPL config.
// ABOUTME: Provides functions to map configuration settings from the main config structure.

package repl

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llmspell/pkg/config"
)

// NewREPLConfigFromConfig creates a REPL configuration from the main config
func NewREPLConfigFromConfig(cfg *config.Config, engine string) REPLConfig {
	replCfg := REPLConfig{
		Engine:          engine,
		Prompt:          cfg.REPL.Prompt,
		ContinuePrompt:  cfg.REPL.ContinuePrompt,
		HistoryFile:     expandPath(cfg.REPL.HistoryFile),
		HistorySize:     cfg.REPL.HistorySize,
		SaveHistory:     cfg.REPL.SaveHistory,
		SyntaxHighlight: cfg.REPL.SyntaxHighlight,
		AutoComplete:    cfg.REPL.AutoComplete,
		MultiLine:       cfg.REPL.MultiLine,
		Input:           os.Stdin,
		Output:          os.Stdout,
		Error:           os.Stderr,
	}

	// If engine not specified, use default from config
	if engine == "" {
		replCfg.Engine = cfg.REPL.DefaultEngine
		if replCfg.Engine == "" {
			replCfg.Engine = cfg.Engine.Default
		}
	}

	// Customize prompt to include engine name if it's the default prompt
	if replCfg.Prompt == "llmspell> " {
		replCfg.Prompt = replCfg.Engine + "> "
	}

	return replCfg
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}

	return path
}

// ApplyREPLSettings applies additional REPL settings from the main config
// This can be used to update REPL behavior based on config settings
func ApplyREPLSettings(replInstance REPL, cfg *config.Config) {
	// Additional settings can be applied here in the future
	// For now, most settings are handled during creation

	// Example: Update history duration, vim mode, etc.
	// These would require extending the REPL interface
}
