// ABOUTME: Implementation of the repl command for starting an interactive REPL.
// ABOUTME: Provides an interactive script execution environment with config integration.

package commands

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	pkgerrors "github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/repl"
)

// REPLCmd starts an interactive REPL
type REPLCmd struct {
	BaseCommand
	Engine      string `short:"e" help:"Script engine to use"`
	HistoryFile string `short:"f" help:"History file path"`
	NoHistory   bool   `help:"Disable history saving"`
	NoHighlight bool   `help:"Disable syntax highlighting"`
	NoComplete  bool   `help:"Disable auto-completion"`
}

// Run executes the command
func (c *REPLCmd) Run(ctx context.Context) error {
	// Get configuration from context
	cfg := GetConfig(ctx)
	c.Debug(ctx, "Starting REPL with config integration")

	// Determine engine to use
	engine := c.Engine
	if engine == "" {
		engine = cfg.REPL.DefaultEngine
		if engine == "" {
			engine = cfg.Engine.Default
		}
	}
	c.Debug(ctx, "Using engine: %s", engine)

	// Create REPL configuration from main config
	replConfig := repl.NewREPLConfigFromConfig(cfg, engine)

	// Override with command-line options if provided
	if c.HistoryFile != "" {
		replConfig.HistoryFile = c.HistoryFile
	}
	if c.NoHistory {
		replConfig.SaveHistory = false
		replConfig.HistoryFile = ""
	}
	if c.NoHighlight {
		replConfig.SyntaxHighlight = false
	}
	if c.NoComplete {
		replConfig.AutoComplete = false
	}

	// Ensure history file has an absolute path
	if replConfig.HistoryFile != "" && !filepath.IsAbs(replConfig.HistoryFile) {
		if !filepath.IsAbs(replConfig.HistoryFile) {
			// Expand ~ if present
			if len(replConfig.HistoryFile) > 0 && replConfig.HistoryFile[0] == '~' {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return pkgerrors.Wrap(err, pkgerrors.CategoryIO, "failed to get user home directory")
				}
				replConfig.HistoryFile = filepath.Join(homeDir, replConfig.HistoryFile[2:])
			}
		}
	}

	// Create engine-specific REPL
	var replInstance repl.REPL
	var err error

	switch engine {
	case "lua":
		replInstance, err = repl.NewLuaREPL(replConfig)
	case "javascript", "js":
		// JavaScript REPL would be created here when implemented
		return pkgerrors.Newf(pkgerrors.CategoryValidation, "JavaScript engine not yet implemented")
	case "tengo":
		// Tengo REPL would be created here when implemented
		return pkgerrors.Newf(pkgerrors.CategoryValidation, "Tengo engine not yet implemented")
	default:
		return pkgerrors.Newf(pkgerrors.CategoryValidation, "unsupported engine: %s", engine)
	}

	if err != nil {
		return pkgerrors.Wrap(err, pkgerrors.CategoryEngine, "failed to create REPL")
	}
	defer func() { _ = replInstance.Close() }()

	// Apply any additional settings from config
	repl.ApplyREPLSettings(replInstance, cfg)

	// Show config-based settings if in debug mode
	if cfg.Debug || IsDebug(ctx) {
		c.Debug(ctx, "REPL Configuration:")
		c.Debug(ctx, "  Engine: %s", replConfig.Engine)
		c.Debug(ctx, "  History File: %s", replConfig.HistoryFile)
		c.Debug(ctx, "  History Size: %d", replConfig.HistorySize)
		c.Debug(ctx, "  Save History: %v", replConfig.SaveHistory)
		c.Debug(ctx, "  Syntax Highlight: %v", replConfig.SyntaxHighlight)
		c.Debug(ctx, "  Auto Complete: %v", replConfig.AutoComplete)
		c.Debug(ctx, "  Multi-line: %v", replConfig.MultiLine)
	}

	// Start the interactive REPL
	if err := replInstance.Start(ctx); err != nil {
		// Don't wrap exit errors
		if errors.Is(err, context.Canceled) || err.Error() == "exit requested" {
			return nil
		}
		return pkgerrors.Wrap(err, pkgerrors.CategoryEngine, "REPL execution failed")
	}

	return nil
}
