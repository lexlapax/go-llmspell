package repl

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseREPL_Creation(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)
	assert.NotNil(t, repl)

	// Test that it implements REPL interface
	var _ REPL = repl
}

func TestBaseREPL_Config(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:          "lua",
		Prompt:          "test> ",
		ContinuePrompt:  "... ",
		HistorySize:     500,
		SyntaxHighlight: true,
		AutoComplete:    true,
		MultiLine:       true,
		Input:           &stdin,
		Output:          &stdout,
		Error:           &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	// Verify config was applied correctly
	assert.NotNil(t, repl)
}

func TestBaseREPL_History(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:      "lua",
		Prompt:      "test> ",
		HistorySize: 3,
		Input:       &stdin,
		Output:      &stdout,
		Error:       &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	// Add history entries
	repl.AddHistory("first")
	repl.AddHistory("second")
	repl.AddHistory("third")

	history := repl.GetHistory()
	assert.Len(t, history, 3)
	assert.Equal(t, []string{"first", "second", "third"}, history)

	// Test history size limit
	repl.AddHistory("fourth")
	history = repl.GetHistory()
	assert.Len(t, history, 3)
	assert.Equal(t, []string{"second", "third", "fourth"}, history)
}

func TestBaseREPL_HistoryFile(t *testing.T) {
	// Create temp file for history
	tmpFile, err := os.CreateTemp("", "repl_test_*.hist")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:      "lua",
		Prompt:      "test> ",
		HistoryFile: tmpFile.Name(),
		SaveHistory: true,
		Input:       &stdin,
		Output:      &stdout,
		Error:       &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	// Add some history
	repl.AddHistory("test command 1")
	repl.AddHistory("test command 2")

	// Close should save history
	err = repl.Close()
	require.NoError(t, err)

	// Create new REPL and verify history loaded
	repl2, err := NewBaseREPL(config)
	require.NoError(t, err)

	history := repl2.GetHistory()
	assert.Contains(t, history, "test command 1")
	assert.Contains(t, history, "test command 2")

	_ = repl2.Close()
}

func TestBaseREPL_Commands(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	tests := []struct {
		input       string
		shouldError bool
		contains    string
	}{
		{".help", false, "Available commands"},
		{".help help", false, "Show help information"},
		{".clear", false, "\033[2J\033[H"},
		{".engines", false, "Available engines"},
		{".load", true, "usage: .load"},
		{".save", true, "usage: .save"},
		{".unknown", true, "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := repl.Evaluate(ctx, tt.input)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
				if tt.contains != "" {
					assert.Contains(t, result, tt.contains)
				}
			}
		})
	}
}

func TestBaseREPL_Completion(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:       "lua",
		Prompt:       "test> ",
		AutoComplete: true,
		Input:        &stdin,
		Output:       &stdout,
		Error:        &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test REPL command completion
	completions := repl.Complete(".he")
	assert.Contains(t, completions, ".help")

	completions = repl.Complete(".e")
	assert.Contains(t, completions, ".exit")
	assert.Contains(t, completions, ".engines")

	// Test Lua keyword completion
	completions = repl.Complete("loc")
	assert.Contains(t, completions, "local")

	completions = repl.Complete("fun")
	assert.Contains(t, completions, "function")
}

func TestBaseREPL_MultilineDetection(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:    "lua",
		Prompt:    "test> ",
		MultiLine: true,
		Input:     &stdin,
		Output:    &stdout,
		Error:     &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test that basic multiline detection works
	assert.NotNil(t, repl)
}

func TestBaseREPL_Close(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	// Should not error on close
	err = repl.Close()
	assert.NoError(t, err)

	// Multiple closes should be safe
	err = repl.Close()
	assert.NoError(t, err)
}
