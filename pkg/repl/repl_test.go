package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPLInterface(t *testing.T) {
	// Test that REPL interface can be implemented
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewREPL(config)
	require.NoError(t, err)
	assert.NotNil(t, repl)

	// Verify it implements the REPL interface
	_ = REPL(repl)
}

func TestNewREPL(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:      "lua",
		Prompt:      "test> ",
		HistoryFile: "",
		Input:       &stdin,
		Output:      &stdout,
		Error:       &stderr,
	}

	repl, err := NewREPL(config)
	require.NoError(t, err)
	assert.NotNil(t, repl)
}

func TestREPLConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  REPLConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: REPLConfig{
				Engine: "lua",
				Prompt: "test> ",
			},
			wantErr: false,
		},
		{
			name: "empty engine",
			config: REPLConfig{
				Engine: "",
				Prompt: "test> ",
			},
			wantErr: true,
		},
		{
			name: "empty prompt",
			config: REPLConfig{
				Engine: "lua",
				Prompt: "",
			},
			wantErr: false, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestREPLCommands(t *testing.T) {
	// Test built-in REPL commands
	tests := []struct {
		input     string
		isCommand bool
		command   string
	}{
		{".help", true, "help"},
		{".exit", true, "exit"},
		{".quit", true, "quit"},
		{".clear", true, "clear"},
		{".load test.lua", true, "load"},
		{".save test.lua", true, "save"},
		{".engines", true, "engines"},
		{"print('hello')", false, ""},
		{"local x = 5", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			isCmd, cmd := parseREPLCommand(tt.input)
			assert.Equal(t, tt.isCommand, isCmd)
			if tt.isCommand {
				assert.Equal(t, tt.command, cmd)
			}
		})
	}
}

func TestREPLEvaluate(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	// Use LuaREPL for actual execution tests
	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Test simple evaluation
	result, err := repl.Evaluate(ctx, "return 2 + 2")
	require.NoError(t, err)
	assert.Equal(t, "4", strings.TrimSpace(result))
}

func TestREPLHistory(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:      "lua",
		Prompt:      "test> ",
		HistoryFile: "", // In-memory only for test
		Input:       &stdin,
		Output:      &stdout,
		Error:       &stderr,
	}

	repl, err := NewREPL(config)
	require.NoError(t, err)

	// Add some history
	repl.AddHistory("print('hello')")
	repl.AddHistory("local x = 5")
	repl.AddHistory("print(x)")

	history := repl.GetHistory()
	assert.Len(t, history, 3)
	assert.Equal(t, "print('hello')", history[0])
	assert.Equal(t, "local x = 5", history[1])
	assert.Equal(t, "print(x)", history[2])
}

func TestREPLCompletion(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewREPL(config)
	require.NoError(t, err)

	// Test basic completion
	completions := repl.Complete("pri")
	assert.Contains(t, completions, "print")

	// Test REPL command completion
	completions = repl.Complete(".he")
	assert.Contains(t, completions, ".help")
}
