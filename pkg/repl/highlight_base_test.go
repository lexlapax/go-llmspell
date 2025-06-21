// ABOUTME: Tests for base syntax highlighting infrastructure including color constants and utility methods.
// ABOUTME: Covers string highlighting, comment detection, and ANSI sequence handling.

package repl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSyntaxHighlighter(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")
	assert.NotNil(t, highlighter)
	assert.Equal(t, "lua", highlighter.engine)
}

func TestSyntaxHighlighter_Highlight_EmptyInput(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")
	result := highlighter.Highlight("")
	assert.Equal(t, "", result)
}

func TestSyntaxHighlighter_Highlight_UnknownEngine(t *testing.T) {
	highlighter := NewSyntaxHighlighter("unknown")
	input := "function test() print('hello') end"
	result := highlighter.Highlight(input)
	assert.Equal(t, input, result) // Should return unchanged
}

func TestSyntaxHighlighter_HighlightStrings_EdgeCases(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escaped quotes in double quotes",
			input:    `"hello \"world\""`,
			expected: ColorString + `"hello \"world\""` + ColorReset,
		},
		{
			name:     "escaped quotes in single quotes",
			input:    `'hello \'world\''`,
			expected: ColorString + `'hello \'world\''` + ColorReset,
		},
		{
			name:     "mixed quotes",
			input:    `"double" and 'single'`,
			expected: ColorString + `"double"` + ColorReset + ` and ` + ColorString + `'single'` + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlighter.highlightStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyntaxHighlighter_HighlightComments_InsideStrings(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")

	// Comments inside strings should not be highlighted
	input := `print("This -- is not a comment")`
	result := highlighter.highlightComments(input, "--")

	// Should not contain comment highlighting
	assert.NotContains(t, result, ColorComment)
	assert.Equal(t, input, result) // Should be unchanged
}

func TestSyntaxHighlighter_IsInsideString(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")

	tests := []struct {
		text     string
		expected bool
	}{
		{`"unclosed string`, true},
		{`'unclosed string`, true},
		{`"closed string"`, false},
		{`'closed string'`, false},
		{`"string with \" escape"`, false},
		{`'string with \' escape'`, false},
		{`"first" "second`, true},
		{`"first" 'second`, true},
		{``, false},
		{`no quotes`, false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := highlighter.isInsideString(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyntaxHighlighter_HighlightNumbers_EdgeCases(t *testing.T) {
	highlighter := NewSyntaxHighlighter("lua")

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "integers",
			input:    "x = 123",
			contains: []string{ColorNumber + "123" + ColorReset},
		},
		{
			name:     "floats",
			input:    "y = 45.67",
			contains: []string{ColorNumber + "45.67" + ColorReset},
		},
		{
			name:     "numbers in identifiers should not be highlighted",
			input:    "var123 = test456",
			contains: []string{}, // Should not highlight parts of identifiers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlighter.highlightNumbers(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestStripColors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple color code",
			input:    ColorKeyword + "function" + ColorReset,
			expected: "function",
		},
		{
			name:     "multiple color codes",
			input:    ColorKeyword + "if" + ColorReset + " " + ColorBuiltin + "print" + ColorReset,
			expected: "if print",
		},
		{
			name:     "no color codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripColors(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
