// ABOUTME: Tests for Lua-specific syntax highlighting including keywords, built-ins, and standard libraries.
// ABOUTME: Verifies proper highlighting of Lua 5.1 syntax elements.

package repl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyntaxHighlighter_HighlightLua(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")

	tests := []struct {
		name     string
		input    string
		contains []string // Substrings that should be in the highlighted output
	}{
		{
			name:  "keywords",
			input: "function test() if true then print('hello') end end",
			contains: []string{
				ColorKeyword + "function" + ColorReset,
				ColorKeyword + "if" + ColorReset,
				ColorKeyword + "true" + ColorReset,
				ColorKeyword + "then" + ColorReset,
				ColorKeyword + "end" + ColorReset,
			},
		},
		{
			name:  "built-in functions",
			input: "print(type(tostring(42)))",
			contains: []string{
				ColorBuiltin + "print" + ColorReset,
				ColorBuiltin + "type" + ColorReset,
				ColorBuiltin + "tostring" + ColorReset,
			},
		},
		{
			name:  "strings",
			input: `print("hello world") print('single quotes')`,
			contains: []string{
				ColorString + `"hello world"` + ColorReset,
				ColorString + `'single quotes'` + ColorReset,
			},
		},
		{
			name:  "comments",
			input: "print('test') -- this is a comment",
			contains: []string{
				ColorComment + "-- this is a comment" + ColorReset,
			},
		},
		{
			name:  "numbers",
			input: "local x = 42.5 local y = 123",
			contains: []string{
				ColorNumber + "42.5" + ColorReset,
				ColorNumber + "123" + ColorReset,
			},
		},
		{
			name:  "standard libraries",
			input: "table.insert(arr, string.upper(text))",
			contains: []string{
				ColorFunction + "table" + ColorReset,
				ColorFunction + "string" + ColorReset,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlighter.Highlight(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Should contain highlighted: %s", expected)
			}
		})
	}
}

func TestSyntaxHighlighter_Integration_WithBaseREPL(t *testing.T) {
	// Test that BaseREPL integrates properly with syntax highlighter
	config := REPLConfig{
		Engine:          "lua",
		Prompt:          "test> ",
		SyntaxHighlight: true,
	}

	repl, err := NewBaseREPL(config)
	assert.NoError(t, err)
	assert.NotNil(t, repl.highlighter)
	assert.Equal(t, "lua", repl.highlighter.engine)

	// Test highlightInput method with simpler expectations
	input := "function test() print('hello') end"
	highlighted := repl.highlightInput(input)

	// At minimum, we should have some highlighting (keywords or strings)
	assert.NotEqual(t, input, highlighted, "Input should be highlighted")
	assert.Contains(t, highlighted, ColorKeyword+"function"+ColorReset)
	assert.Contains(t, highlighted, ColorString+"'hello'"+ColorReset)
	assert.Contains(t, highlighted, ColorKeyword+"end"+ColorReset)
}

func TestSyntaxHighlighter_Integration_Disabled(t *testing.T) {
	// Test with syntax highlighting disabled
	config := REPLConfig{
		Engine:          "lua",
		Prompt:          "test> ",
		SyntaxHighlight: false,
	}

	repl, err := NewBaseREPL(config)
	assert.NoError(t, err)
	assert.NotNil(t, repl.highlighter) // Still created but not used

	// Test highlightInput method returns unchanged input
	input := "function test() print('hello') end"
	highlighted := repl.highlightInput(input)
	assert.Equal(t, input, highlighted) // Should be unchanged
}
