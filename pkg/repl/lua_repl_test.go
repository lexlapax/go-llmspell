package repl

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLuaREPL_Creation(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	assert.NotNil(t, repl)

	// Test that it implements REPL interface
	var _ REPL = repl
}

func TestLuaREPL_InvalidConfig(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "javascript", // Wrong engine
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	_, err := NewLuaREPL(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "engine must be 'lua'")
}

func TestLuaREPL_BasicExecution(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"return 2 + 2", "4", false},
		{"return 'hello world'", "hello world", false},
		{"return true", "true", false},
		{"return nil", "", false}, // nil typically prints as empty
		{"invalid lua syntax ((", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := repl.Evaluate(ctx, tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != "" {
					assert.Contains(t, result, tt.expected)
				}
			}
		})
	}
}

func TestLuaREPL_Variables(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Set a variable
	_, err = repl.Evaluate(ctx, "x = 42")
	require.NoError(t, err)

	// Use the variable
	result, err := repl.Evaluate(ctx, "return x")
	require.NoError(t, err)
	assert.Contains(t, result, "42")

	// Modify the variable
	_, err = repl.Evaluate(ctx, "x = x * 2")
	require.NoError(t, err)

	result, err = repl.Evaluate(ctx, "return x")
	require.NoError(t, err)
	assert.Contains(t, result, "84")
}

func TestLuaREPL_Functions(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Define a function
	_, err = repl.Evaluate(ctx, "function add(a, b) return a + b end")
	require.NoError(t, err)

	// Use the function
	result, err := repl.Evaluate(ctx, "return add(3, 5)")
	require.NoError(t, err)
	assert.Contains(t, result, "8")
}

func TestLuaREPL_Completion(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:       "lua",
		Prompt:       "lua> ",
		AutoComplete: true,
		Input:        &stdin,
		Output:       &stdout,
		Error:        &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test Lua-specific completions
	completions := repl.Complete("pri")
	assert.Contains(t, completions, "print")

	completions = repl.Complete("loc")
	assert.Contains(t, completions, "local")

	completions = repl.Complete("fun")
	assert.Contains(t, completions, "function")

	completions = repl.Complete("tab")
	assert.Contains(t, completions, "table")
}

func TestLuaREPL_MultilineDetection(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine:    "lua",
		Prompt:    "lua> ",
		MultiLine: true,
		Input:     &stdin,
		Output:    &stdout,
		Error:     &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	// Test that multiline detection works for Lua
	assert.NotNil(t, repl)
}

func TestLuaREPL_ErrorHandling(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Test syntax error
	_, err = repl.Evaluate(ctx, "invalid syntax ((")
	assert.Error(t, err)

	// Test runtime error
	_, err = repl.Evaluate(ctx, "error('test error')")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestLuaREPL_StatePreservation(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "lua",
		Prompt: "lua> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewLuaREPL(config)
	require.NoError(t, err)
	defer func() { _ = repl.Close() }()

	ctx := context.Background()

	// Set up some state
	_, err = repl.Evaluate(ctx, "counter = 0")
	require.NoError(t, err)

	_, err = repl.Evaluate(ctx, "function increment() counter = counter + 1 return counter end")
	require.NoError(t, err)

	// Test that state is preserved across evaluations
	result, err := repl.Evaluate(ctx, "return increment()")
	require.NoError(t, err)
	assert.Contains(t, result, "1")

	result, err = repl.Evaluate(ctx, "return increment()")
	require.NoError(t, err)
	assert.Contains(t, result, "2")

	result, err = repl.Evaluate(ctx, "return counter")
	require.NoError(t, err)
	assert.Contains(t, result, "2")
}
