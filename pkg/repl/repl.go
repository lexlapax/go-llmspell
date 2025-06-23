// ABOUTME: Main REPL interface and configuration for interactive script execution.
// ABOUTME: Provides engine-agnostic REPL functionality with history, completion, and command support.

// Package repl provides interactive Read-Eval-Print Loop functionality.
// It offers engine-agnostic REPL interfaces with features including syntax highlighting,
// auto-completion, history management, and built-in commands for all supported script engines.
package repl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// REPL represents an interactive Read-Eval-Print Loop interface.
// It provides methods for interactive script evaluation, command completion,
// history management, and session control.
type REPL interface {
	// Start begins the interactive REPL session
	Start(ctx context.Context) error

	// Evaluate executes a single line of code and returns the result
	Evaluate(ctx context.Context, input string) (string, error)

	// Complete provides auto-completion suggestions for the given input
	Complete(input string) []string

	// AddHistory adds a line to the command history
	AddHistory(line string)

	// GetHistory returns the command history
	GetHistory() []string

	// Close shuts down the REPL and cleans up resources
	Close() error
}

// REPLConfig holds configuration for a REPL instance.
// It defines behavior settings including prompts, history management,
// I/O streams, and interactive features like syntax highlighting.
type REPLConfig struct {
	// Engine to use for script execution
	Engine string

	// Prompt string to display
	Prompt string

	// Continue prompt for multi-line input
	ContinuePrompt string

	// History file path (empty for in-memory only)
	HistoryFile string

	// Maximum history size
	HistorySize int

	// Whether to save history on exit
	SaveHistory bool

	// Input/Output streams
	Input  io.Reader
	Output io.Writer
	Error  io.Writer

	// Whether to enable syntax highlighting
	SyntaxHighlight bool

	// Whether to enable auto-completion
	AutoComplete bool

	// Whether to support multi-line input
	MultiLine bool

	// Optional engine registry for bridge support
	// If provided, REPL will use engines from this registry with lazy bridge loading
	EngineRegistry interface{}
}

// Validate checks if the REPL configuration is valid.
// It ensures required fields are present and sets default values
// for optional fields when not specified.
func (c *REPLConfig) Validate() error {
	if c.Engine == "" {
		return errors.New(errors.CategoryConfig, "engine cannot be empty")
	}

	// Set defaults if not specified
	if c.Prompt == "" {
		c.Prompt = "llmspell> "
	}
	if c.ContinuePrompt == "" {
		c.ContinuePrompt = "... "
	}
	if c.HistorySize <= 0 {
		c.HistorySize = 1000
	}

	return nil
}

// NewREPL creates a new REPL instance with the given configuration.
// It validates the configuration and returns an appropriate REPL
// implementation based on the specified engine.
func NewREPL(config REPLConfig) (REPL, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, errors.CategoryConfig, "invalid REPL config")
	}

	return NewBaseREPL(config)
}

// parseREPLCommand checks if input is a REPL command and returns the command name.
// REPL commands start with a dot (.) prefix. Returns true and the command name
// if it's a command, or false and empty string otherwise.
func parseREPLCommand(input string) (bool, string) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, ".") {
		return false, ""
	}

	// Extract command name (first word after .)
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return false, ""
	}

	command := strings.TrimPrefix(parts[0], ".")
	return true, command
}

// REPLCommand represents a built-in REPL command.
// It defines the command's metadata and handler function for
// processing command invocations.
type REPLCommand struct {
	Name        string
	Description string
	Usage       string
	Handler     func(ctx context.Context, args []string) (string, error)
}

// GetBuiltinCommands returns the list of built-in REPL commands.
// These commands provide session management, file operations,
// and information display functionality.
func GetBuiltinCommands() map[string]REPLCommand {
	return map[string]REPLCommand{
		"help": {
			Name:        "help",
			Description: "Show help information",
			Usage:       ".help [command]",
			Handler:     helpCommand,
		},
		"exit": {
			Name:        "exit",
			Description: "Exit the REPL",
			Usage:       ".exit",
			Handler:     exitCommand,
		},
		"quit": {
			Name:        "quit",
			Description: "Exit the REPL (alias for .exit)",
			Usage:       ".quit",
			Handler:     exitCommand,
		},
		"clear": {
			Name:        "clear",
			Description: "Clear the screen",
			Usage:       ".clear",
			Handler:     clearCommand,
		},
		"load": {
			Name:        "load",
			Description: "Load and execute a script file",
			Usage:       ".load <filename>",
			Handler:     loadCommand,
		},
		"save": {
			Name:        "save",
			Description: "Save current session to a file",
			Usage:       ".save <filename>",
			Handler:     saveCommand,
		},
		"engines": {
			Name:        "engines",
			Description: "List available engines",
			Usage:       ".engines",
			Handler:     enginesCommand,
		},
	}
}

// helpCommand displays help information for REPL commands.
// If a specific command is provided, shows detailed help for that command.
// Otherwise, lists all available commands.
func helpCommand(ctx context.Context, args []string) (string, error) {
	commands := GetBuiltinCommands()

	if len(args) > 1 {
		// Show help for specific command
		cmdName := args[1]
		if cmd, exists := commands[cmdName]; exists {
			return fmt.Sprintf("%s - %s\nUsage: %s", cmd.Name, cmd.Description, cmd.Usage), nil
		}
		return fmt.Sprintf("Unknown command: %s", cmdName), nil
	}

	// Show all commands
	var result strings.Builder
	result.WriteString("Available commands:\n")
	for _, cmd := range commands {
		result.WriteString(fmt.Sprintf("  %-10s %s\n", "."+cmd.Name, cmd.Description))
	}
	result.WriteString("\nType .help <command> for more information about a specific command.")

	return result.String(), nil
}

// exitCommand handles the .exit and .quit commands to terminate the REPL session.
// Returns a special error that signals the REPL to gracefully shutdown.
func exitCommand(ctx context.Context, args []string) (string, error) {
	return "exit", errors.New(errors.CategoryValidation, "exit requested")
}

// clearCommand clears the terminal screen using ANSI escape sequences.
// Returns the appropriate escape sequence to clear the screen and reset cursor.
func clearCommand(ctx context.Context, args []string) (string, error) {
	return "\033[2J\033[H", nil // ANSI clear screen
}

// loadCommand loads and executes a script file in the current REPL session.
// Requires a filename argument. The file contents are read and evaluated
// as if typed directly into the REPL.
func loadCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New(errors.CategoryValidation, "usage: .load <filename>")
	}
	// TODO: Implement file loading
	return fmt.Sprintf("Loading file: %s (not implemented)", args[1]), nil
}

// saveCommand saves the current REPL session history to a file.
// Requires a filename argument. The session history is written to the
// specified file for later replay or review.
func saveCommand(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New(errors.CategoryValidation, "usage: .save <filename>")
	}
	// TODO: Implement session saving
	return fmt.Sprintf("Saving session to: %s (not implemented)", args[1]), nil
}

// enginesCommand lists all available script engines.
// Shows engine names, versions, and current availability status.
func enginesCommand(ctx context.Context, args []string) (string, error) {
	// TODO: Integrate with engine registry
	return "Available engines:\n  - lua (Lua 5.1)\n  - javascript (not implemented)\n  - tengo (not implemented)", nil
}
