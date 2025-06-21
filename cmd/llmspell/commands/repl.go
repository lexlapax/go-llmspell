// ABOUTME: Implementation of the repl command for starting an interactive REPL.
// ABOUTME: Provides an interactive script execution environment.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/repl"
)

// REPLCmd starts an interactive REPL
type REPLCmd struct {
	BaseCommand
	Engine      string `short:"e" help:"Script engine to use" default:"lua"`
	HistoryFile string `short:"f" help:"History file path (default: ~/.llmspell_history)"`
}

// Run executes the command
func (c *REPLCmd) Run(ctx context.Context) error {
	c.Debug(ctx, "Starting REPL with %s engine", c.Engine)

	// Set default history file if not specified
	historyFile := c.HistoryFile
	if historyFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, errors.CategoryIO, "failed to get user home directory")
		}
		historyFile = filepath.Join(homeDir, ".llmspell_history")
	}

	// Create REPL configuration
	replConfig := repl.REPLConfig{
		Engine:       c.Engine,
		Prompt:       c.Engine + "> ",
		HistoryFile:  historyFile,
		SaveHistory:  true,
		Input:        os.Stdin,
		Output:       os.Stdout,
		Error:        os.Stderr,
		AutoComplete: true,
		MultiLine:    true,
	}

	// Create engine-specific REPL
	var replInstance repl.REPL
	var err error

	switch c.Engine {
	case "lua":
		replInstance, err = repl.NewLuaREPL(replConfig)
	default:
		return errors.Newf(errors.CategoryValidation, "unsupported engine: %s", c.Engine)
	}

	if err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "failed to create REPL")
	}
	defer func() { _ = replInstance.Close() }()

	c.Printf("Starting %s REPL. Type .help for commands.\n", c.Engine)

	// Start the interactive REPL
	if err := replInstance.Start(ctx); err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "REPL execution failed")
	}

	return nil
}
