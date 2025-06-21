// ABOUTME: Common types and utilities shared across all CLI commands.
// ABOUTME: Provides context keys, base command functionality, and output helpers.

package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lexlapax/go-llmspell/pkg/config"
)

// Context keys for command execution
type contextKey string

const (
	ConfigKey         contextKey = "config"
	DebugKey          contextKey = "debug"
	VerboseKey        contextKey = "verbose"
	ProfileKey        contextKey = "profile"
	EngineRegistryKey contextKey = "engineRegistry"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	// Output writer (defaults to stdout)
	Out io.Writer `kong:"-"`
	// Error writer (defaults to stderr)
	Err io.Writer `kong:"-"`
}

// GetConfig extracts config from context
func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(ConfigKey).(*config.Config); ok {
		return cfg
	}
	// Return a basic config for now
	return &config.Config{
		Debug: false,
	}
}

// IsDebug checks if debug mode is enabled
func IsDebug(ctx context.Context) bool {
	if debug, ok := ctx.Value(DebugKey).(bool); ok {
		return debug
	}
	return false
}

// IsVerbose checks if verbose mode is enabled
func IsVerbose(ctx context.Context) bool {
	if verbose, ok := ctx.Value(VerboseKey).(bool); ok {
		return verbose
	}
	return false
}

// GetProfile gets the security profile from context
func GetProfile(ctx context.Context) string {
	if profile, ok := ctx.Value(ProfileKey).(string); ok {
		return profile
	}
	return "sandbox"
}

// GetEngineRegistry gets the engine registry from context
func GetEngineRegistry(ctx context.Context) interface{} {
	return ctx.Value(EngineRegistryKey)
}

// Printf prints formatted output to stdout
func (b *BaseCommand) Printf(format string, args ...interface{}) {
	out := b.Out
	if out == nil {
		out = os.Stdout
	}
	_, _ = fmt.Fprintf(out, format, args...)
}

// Println prints a line to stdout
func (b *BaseCommand) Println(args ...interface{}) {
	out := b.Out
	if out == nil {
		out = os.Stdout
	}
	_, _ = fmt.Fprintln(out, args...)
}

// Errorf prints formatted error to stderr
func (b *BaseCommand) Errorf(format string, args ...interface{}) {
	err := b.Err
	if err == nil {
		err = os.Stderr
	}
	_, _ = fmt.Fprintf(err, format, args...)
}

// Errorln prints error line to stderr
func (b *BaseCommand) Errorln(args ...interface{}) {
	err := b.Err
	if err == nil {
		err = os.Stderr
	}
	_, _ = fmt.Fprintln(err, args...)
}

// Debug prints debug message if debug mode is enabled
func (b *BaseCommand) Debug(ctx context.Context, format string, args ...interface{}) {
	if IsDebug(ctx) {
		b.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Verbose prints verbose message if verbose mode is enabled
func (b *BaseCommand) Verbose(ctx context.Context, format string, args ...interface{}) {
	if IsVerbose(ctx) {
		b.Printf(format+"\n", args...)
	}
}

// TableWriter helps format tabular output
type TableWriter struct {
	headers []string
	rows    [][]string
	out     io.Writer
}

// NewTableWriter creates a new table writer
func NewTableWriter(out io.Writer, headers ...string) *TableWriter {
	return &TableWriter{
		headers: headers,
		rows:    [][]string{},
		out:     out,
	}
}

// AddRow adds a row to the table
func (t *TableWriter) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Render outputs the table
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
