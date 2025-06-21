package repl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompleter(t *testing.T) {
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

	completer := NewCompleter(repl)
	assert.NotNil(t, completer)
	assert.Equal(t, repl, completer.repl)
}

func TestCompleter_Do(t *testing.T) {
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

	completer := NewCompleter(repl)

	// Test with partial input
	line := []rune("pri")
	pos := 3
	newLine, length := completer.Do(line, pos)

	assert.Greater(t, len(newLine), 0)
	assert.Equal(t, 3, length)

	// Should contain "print"
	found := false
	for _, completion := range newLine {
		if string(completion) == "print" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should contain 'print' completion")
}

func TestCompleter_GetCompletions_REPLCommands(t *testing.T) {
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

	completer := NewCompleter(repl)

	tests := []struct {
		input    string
		expected []string
	}{
		{".he", []string{".help"}},
		{".e", []string{".engines", ".exit"}},
		{".q", []string{".quit"}},
		{".c", []string{".clear"}},
		{".l", []string{".load"}},
		{".s", []string{".save"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			completions := completer.GetCompletions(tt.input)
			for _, expected := range tt.expected {
				assert.Contains(t, completions, expected)
			}
		})
	}
}

func TestCompleter_GetLuaCompletions(t *testing.T) {
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

	completer := NewCompleter(repl)

	tests := []struct {
		input    string
		expected []string
	}{
		{"pri", []string{"print"}},
		{"loc", []string{"local"}},
		{"fun", []string{"function"}},
		{"tab", []string{"table", "table.insert", "table.remove", "table.concat", "table.sort"}},
		{"string.", []string{"string.len", "string.sub", "string.find", "string.match", "string.gsub", "string.format", "string.upper", "string.lower"}},
		{"math.", []string{"math.abs", "math.ceil", "math.floor", "math.max", "math.min", "math.random", "math.sqrt", "math.sin", "math.cos", "math.pi"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			completions := completer.GetCompletions(tt.input)
			for _, expected := range tt.expected {
				assert.Contains(t, completions, expected, "Should contain completion: %s", expected)
			}
		})
	}
}

func TestCompleter_GetJavaScriptCompletions(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "javascript",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	completer := NewCompleter(repl)

	tests := []struct {
		input    string
		expected []string
	}{
		{"fun", []string{"function"}},
		{"con", []string{"const", "console.log", "console.error", "console.warn", "console.info"}},
		{"let", []string{"let"}},
		{"var", []string{"var"}},
		{"if", []string{"if"}},
		{"for", []string{"for"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			completions := completer.GetCompletions(tt.input)
			for _, expected := range tt.expected {
				assert.Contains(t, completions, expected, "Should contain completion: %s", expected)
			}
		})
	}
}

func TestCompleter_GetTengoCompletions(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "tengo",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	completer := NewCompleter(repl)

	tests := []struct {
		input    string
		expected []string
	}{
		{"fun", []string{"func"}},
		{"if", []string{"if"}},
		{"for", []string{"for"}},
		{"len", []string{"len"}},
		{"pri", []string{"print", "printf"}},
		{"is_", []string{"is_string", "is_int", "is_float", "is_bool", "is_char", "is_bytes", "is_array", "is_map", "is_undefined", "is_function", "is_callable", "is_iterable"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			completions := completer.GetCompletions(tt.input)
			for _, expected := range tt.expected {
				assert.Contains(t, completions, expected, "Should contain completion: %s", expected)
			}
		})
	}
}

func TestCompleter_GetCompletions_UnknownEngine(t *testing.T) {
	var stdin, stdout, stderr bytes.Buffer
	config := REPLConfig{
		Engine: "unknown",
		Prompt: "test> ",
		Input:  &stdin,
		Output: &stdout,
		Error:  &stderr,
	}

	repl, err := NewBaseREPL(config)
	require.NoError(t, err)

	completer := NewCompleter(repl)

	// Should only return REPL commands for unknown engines
	completions := completer.GetCompletions(".he")
	assert.Contains(t, completions, ".help")

	// Should return empty for language completions
	completions = completer.GetCompletions("pri")
	assert.Empty(t, completions)
}

func TestCompleter_GetCompletions_EmptyInput(t *testing.T) {
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

	completer := NewCompleter(repl)

	completions := completer.GetCompletions("")
	// Should return all Lua keywords and built-ins
	assert.Greater(t, len(completions), 50)
	assert.Contains(t, completions, "print")
	assert.Contains(t, completions, "function")
	assert.Contains(t, completions, "local")
}

func TestCompleter_GetCompletions_Sorted(t *testing.T) {
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

	completer := NewCompleter(repl)

	completions := completer.GetCompletions("a")

	// Verify results are sorted
	for i := 1; i < len(completions); i++ {
		assert.LessOrEqual(t, completions[i-1], completions[i], "Completions should be sorted")
	}
}

func TestCompleter_Integration_WithBaseREPL(t *testing.T) {
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

	// Test that BaseREPL.Complete uses the new completer
	completions := repl.Complete("pri")
	assert.Contains(t, completions, "print")

	// Test REPL commands
	completions = repl.Complete(".he")
	assert.Contains(t, completions, ".help")
}
