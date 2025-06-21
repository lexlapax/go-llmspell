// ABOUTME: Base REPL implementation providing common functionality for all script engines.
// ABOUTME: Handles history, completion, commands, and multi-line input detection.

package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// BaseREPL provides common REPL functionality that can be extended by engine-specific implementations
type BaseREPL struct {
	config      REPLConfig
	history     []string
	readline    *readline.Instance
	highlighter *SyntaxHighlighter
	mu          sync.RWMutex
	closed      bool
}

// NewBaseREPL creates a new base REPL instance
func NewBaseREPL(config REPLConfig) (*BaseREPL, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, errors.CategoryConfig, "invalid REPL configuration")
	}

	// Set default streams if not provided
	if config.Input == nil {
		config.Input = os.Stdin
	}
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.Error == nil {
		config.Error = os.Stderr
	}

	repl := &BaseREPL{
		config:      config,
		history:     make([]string, 0, config.HistorySize),
		highlighter: NewSyntaxHighlighter(config.Engine),
	}

	// Load history if file specified
	if config.HistoryFile != "" {
		repl.loadHistory()
	}

	// Set up readline if using real stdin/stdout
	if config.Input == os.Stdin && config.Output == os.Stdout {
		if err := repl.setupReadline(); err != nil {
			return nil, errors.Wrap(err, errors.CategoryConfig, "failed to setup readline")
		}
	}

	return repl, nil
}

// Start begins the interactive REPL session
func (r *BaseREPL) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errors.New(errors.CategoryConfig, "REPL is closed")
	}

	_, _ = fmt.Fprintf(r.config.Output, "Starting %s REPL. Type .help for commands.\n", r.config.Engine)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read input
		var input string
		var err error

		if r.readline != nil {
			r.readline.SetPrompt(r.config.Prompt)
			input, err = r.readline.Readline()
		} else {
			// Fallback for testing
			_, _ = fmt.Fprint(r.config.Output, r.config.Prompt)
			scanner := bufio.NewScanner(r.config.Input)
			if scanner.Scan() {
				input = scanner.Text()
			} else {
				err = io.EOF
			}
		}

		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				_, _ = fmt.Fprintln(r.config.Output, "\nGoodbye!")
				return nil
			}
			return errors.Wrap(err, errors.CategoryIO, "input error")
		}

		// Skip empty lines
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle multi-line input
		if r.config.MultiLine && r.isIncompleteInput(input) {
			input = r.readMultilineInput(input)
		}

		// Show highlighted input if syntax highlighting is enabled
		if r.config.SyntaxHighlight && input != "" {
			highlighted := r.highlightInput(input)
			if highlighted != input {
				// Show the highlighted version on a new line
				_, _ = fmt.Fprintf(r.config.Output, "\033[1A\033[K%s%s\n", r.config.Prompt, highlighted)
			}
		}

		// Add to history
		r.AddHistory(input)

		// Evaluate input
		result, err := r.Evaluate(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "exit requested") {
				_, _ = fmt.Fprintln(r.config.Output, "Goodbye!")
				return nil
			}
			_, _ = fmt.Fprintf(r.config.Error, "Error: %v\n", err)
			continue
		}

		// Print result if not empty
		if result != "" {
			_, _ = fmt.Fprintln(r.config.Output, result)
		}
	}
}

// Evaluate executes a single line of code and returns the result
func (r *BaseREPL) Evaluate(ctx context.Context, input string) (string, error) {
	input = strings.TrimSpace(input)

	// Check if it's a REPL command
	if isCommand, command := parseREPLCommand(input); isCommand {
		return r.executeCommand(ctx, input, command)
	}

	// TODO: Execute script using engine
	// For now, return a placeholder
	return fmt.Sprintf("Executing (%s): %s", r.config.Engine, input), nil
}

// Complete provides auto-completion suggestions for the given input
func (r *BaseREPL) Complete(input string) []string {
	completer := NewCompleter(r)
	return completer.GetCompletions(input)
}

// highlightInput applies syntax highlighting to input if enabled
func (r *BaseREPL) highlightInput(input string) string {
	if !r.config.SyntaxHighlight || r.highlighter == nil {
		return input
	}
	return r.highlighter.Highlight(input)
}

// AddHistory adds a line to the command history
func (r *BaseREPL) AddHistory(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Don't add duplicate consecutive entries
	if len(r.history) > 0 && r.history[len(r.history)-1] == line {
		return
	}

	// Add to history
	r.history = append(r.history, line)

	// Trim history to size limit
	if len(r.history) > r.config.HistorySize {
		r.history = r.history[len(r.history)-r.config.HistorySize:]
	}

	// Add to readline history if available
	if r.readline != nil {
		_ = r.readline.SaveHistory(line)
	}
}

// GetHistory returns the command history
func (r *BaseREPL) GetHistory() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent modification
	result := make([]string, len(r.history))
	copy(result, r.history)
	return result
}

// Close shuts down the REPL and cleans up resources
func (r *BaseREPL) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	// Save history if configured
	if r.config.SaveHistory && r.config.HistoryFile != "" {
		r.saveHistory()
	}

	// Close readline
	if r.readline != nil {
		_ = r.readline.Close()
	}

	r.closed = true
	return nil
}

// setupReadline configures the readline instance
func (r *BaseREPL) setupReadline() error {
	cfg := &readline.Config{
		Prompt:      r.config.Prompt,
		HistoryFile: r.config.HistoryFile,
	}

	// Set up auto-completion
	if r.config.AutoComplete {
		cfg.AutoComplete = NewCompleter(r)
	}

	rl, err := readline.NewEx(cfg)
	if err != nil {
		return err
	}

	r.readline = rl
	return nil
}

// executeCommand executes a built-in REPL command
func (r *BaseREPL) executeCommand(ctx context.Context, input, command string) (string, error) {
	commands := GetBuiltinCommands()

	cmd, exists := commands[command]
	if !exists {
		return "", errors.Newf(errors.CategoryValidation, "unknown command: .%s", command)
	}

	// Parse arguments
	args := strings.Fields(input)

	return cmd.Handler(ctx, args)
}

// isIncompleteInput checks if the input appears to be incomplete (for multi-line support)
func (r *BaseREPL) isIncompleteInput(input string) bool {
	input = strings.TrimSpace(input)

	// Lua-specific incomplete input detection
	if r.config.Engine == "lua" {
		// Simple heuristics for incomplete Lua input
		if strings.HasSuffix(input, "then") ||
			strings.HasSuffix(input, "do") ||
			strings.HasSuffix(input, "{") ||
			strings.HasPrefix(input, "function") && !strings.Contains(input, "end") {
			return true
		}
	}

	return false
}

// readMultilineInput reads additional lines for multi-line input
func (r *BaseREPL) readMultilineInput(initial string) string {
	var lines []string
	lines = append(lines, initial)

	for {
		var input string
		var err error

		if r.readline != nil {
			r.readline.SetPrompt(r.config.ContinuePrompt)
			input, err = r.readline.Readline()
		} else {
			// Fallback for testing
			_, _ = fmt.Fprint(r.config.Output, r.config.ContinuePrompt)
			scanner := bufio.NewScanner(r.config.Input)
			if scanner.Scan() {
				input = scanner.Text()
			} else {
				break
			}
		}

		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			break // Empty line ends multi-line input
		}

		lines = append(lines, input)

		// Check if input is now complete
		combined := strings.Join(lines, "\n")
		if !r.isIncompleteInput(combined) {
			break
		}
	}

	return strings.Join(lines, "\n")
}

// loadHistory loads command history from file
func (r *BaseREPL) loadHistory() {
	if r.config.HistoryFile == "" {
		return
	}

	file, err := os.Open(r.config.HistoryFile)
	if err != nil {
		return // File doesn't exist yet, that's okay
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			r.history = append(r.history, line)
		}
	}

	// Trim to size limit
	if len(r.history) > r.config.HistorySize {
		r.history = r.history[len(r.history)-r.config.HistorySize:]
	}
}

// saveHistory saves command history to file
func (r *BaseREPL) saveHistory() {
	if r.config.HistoryFile == "" {
		return
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(r.config.HistoryFile), 0755); err != nil {
		return
	}

	file, err := os.Create(r.config.HistoryFile)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	for _, line := range r.history {
		_, _ = fmt.Fprintln(file, line)
	}
}
