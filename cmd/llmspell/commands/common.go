// ABOUTME: Common types and utilities shared across all CLI commands.
// ABOUTME: Provides context keys, base command functionality, and output helpers.

// Package commands implements all CLI commands for llmspell.
// It provides command implementations for script execution, validation,
// REPL interaction, and various management tasks.
package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llmspell/pkg/config"
)

// contextKey is the type for context value keys
type contextKey string

// Context keys for command execution
const (
	// ConfigKey stores the application configuration
	ConfigKey contextKey = "config"
	// DebugKey stores the debug mode flag
	DebugKey contextKey = "debug"
	// VerboseKey stores the verbose output flag
	VerboseKey contextKey = "verbose"
	// ProfileKey stores the security profile name
	ProfileKey contextKey = "profile"
	// EngineRegistryKey stores the engine registry instance
	EngineRegistryKey contextKey = "engineRegistry"
)

// BaseCommand provides common functionality for all commands.
// It includes output writers for consistent command output handling
// across all command implementations.
type BaseCommand struct {
	// Output writer (defaults to stdout)
	Out io.Writer `kong:"-"`
	// Error writer (defaults to stderr)
	Err io.Writer `kong:"-"`
}

// GetConfig extracts config from context.
// Returns default configuration if not found in context.
func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(ConfigKey).(*config.Config); ok {
		return cfg
	}
	// Return a basic config for now
	return &config.Config{
		Debug: false,
	}
}

// IsDebug checks if debug mode is enabled.
// Returns false if not found in context.
func IsDebug(ctx context.Context) bool {
	if debug, ok := ctx.Value(DebugKey).(bool); ok {
		return debug
	}
	return false
}

// IsVerbose checks if verbose mode is enabled.
// Returns false if not found in context.
func IsVerbose(ctx context.Context) bool {
	if verbose, ok := ctx.Value(VerboseKey).(bool); ok {
		return verbose
	}
	return false
}

// GetProfile gets the security profile from context.
// Returns "sandbox" as default if not found.
func GetProfile(ctx context.Context) string {
	if profile, ok := ctx.Value(ProfileKey).(string); ok {
		return profile
	}
	return "sandbox"
}

// GetEngineRegistry gets the engine registry from context.
// Returns nil if not found in context.
func GetEngineRegistry(ctx context.Context) interface{} {
	return ctx.Value(EngineRegistryKey)
}

// Printf prints formatted output to stdout.
// Uses the configured output writer or defaults to os.Stdout.
func (b *BaseCommand) Printf(format string, args ...interface{}) {
	out := b.Out
	if out == nil {
		out = os.Stdout
	}
	_, _ = fmt.Fprintf(out, format, args...)
}

// Println prints a line to stdout.
// Uses the configured output writer or defaults to os.Stdout.
func (b *BaseCommand) Println(args ...interface{}) {
	out := b.Out
	if out == nil {
		out = os.Stdout
	}
	_, _ = fmt.Fprintln(out, args...)
}

// Errorf prints formatted error to stderr.
// Uses the configured error writer or defaults to os.Stderr.
func (b *BaseCommand) Errorf(format string, args ...interface{}) {
	err := b.Err
	if err == nil {
		err = os.Stderr
	}
	_, _ = fmt.Fprintf(err, format, args...)
}

// Info prints an info message.
// Adds a newline automatically to the formatted output.
func (b *BaseCommand) Info(ctx context.Context, format string, args ...interface{}) {
	b.Printf(format+"\n", args...)
}

// Debug prints a debug message if debug mode is enabled.
// Messages are prefixed with [DEBUG] for clarity.
func (b *BaseCommand) Debug(ctx context.Context, format string, args ...interface{}) {
	if IsDebug(ctx) {
		b.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Error prints an error message.
// Messages are prefixed with [ERROR] and sent to stderr.
func (b *BaseCommand) Error(ctx context.Context, format string, args ...interface{}) {
	b.Errorf("[ERROR] "+format+"\n", args...)
}

// Errorln prints error line to stderr.
// Uses the configured error writer or defaults to os.Stderr.
func (b *BaseCommand) Errorln(args ...interface{}) {
	err := b.Err
	if err == nil {
		err = os.Stderr
	}
	_, _ = fmt.Fprintln(err, args...)
}

// Verbose prints verbose message if verbose mode is enabled.
// Only outputs when verbose flag is set in context.
func (b *BaseCommand) Verbose(ctx context.Context, format string, args ...interface{}) {
	if IsVerbose(ctx) {
		b.Printf(format+"\n", args...)
	}
}

// TableWriter helps format tabular output.
// It provides a simple way to display data in aligned columns
// with headers and consistent formatting.
type TableWriter struct {
	headers []string
	rows    [][]string
	out     io.Writer
}

// NewTableWriter creates a new table writer.
// The headers parameter defines the column headers for the table.
func NewTableWriter(out io.Writer, headers ...string) *TableWriter {
	return &TableWriter{
		headers: headers,
		rows:    [][]string{},
		out:     out,
	}
}

// AddRow adds a row to the table.
// The number of values should match the number of headers.
func (t *TableWriter) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Render outputs the table.
// It calculates column widths and formats the output
// with proper alignment and spacing.
func (t *TableWriter) Render() {
	if t.out == nil {
		t.out = os.Stdout
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, v := range row {
			if i < len(widths) && len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	// Print headers
	for i, h := range t.headers {
		_, _ = fmt.Fprintf(t.out, "%-*s", widths[i]+2, h)
	}
	_, _ = fmt.Fprintln(t.out)

	// Print separator
	for i := range t.headers {
		for j := 0; j < widths[i]+2; j++ {
			_, _ = fmt.Fprint(t.out, "-")
		}
	}
	_, _ = fmt.Fprintln(t.out)

	// Print rows
	for _, row := range t.rows {
		for i, v := range row {
			if i < len(widths) {
				_, _ = fmt.Fprintf(t.out, "%-*s", widths[i]+2, v)
			}
		}
		_, _ = fmt.Fprintln(t.out)
	}
}

// getDefaultConfigPath returns the default config file path.
// It follows XDG Base Directory specification, using XDG_CONFIG_HOME
// if set, otherwise defaulting to ~/.config/llmspell/config.yaml.
func getDefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "llmspell", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "llmspell", "config.yaml")
}
