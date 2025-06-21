// ABOUTME: Tests for base syntax highlighting infrastructure including color constants and utility methods.
// ABOUTME: Covers tokenization-based highlighting and ANSI color stripping.

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

func TestTokenization(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		keywords      []string
		builtins      []string
		commentPrefix string
		expectedTypes []string
	}{
		{
			name:          "basic tokenization",
			input:         `print("hello")`,
			keywords:      []string{},
			builtins:      []string{"print"},
			commentPrefix: "--",
			expectedTypes: []string{TokenTypeBuiltin, TokenTypeDefault, TokenTypeString, TokenTypeDefault},
		},
		{
			name:          "keywords and numbers",
			input:         "if x > 42.5 then",
			keywords:      []string{"if", "then"},
			builtins:      []string{},
			commentPrefix: "--",
			expectedTypes: []string{TokenTypeKeyword, TokenTypeDefault, TokenTypeDefault, TokenTypeDefault, TokenTypeDefault, TokenTypeDefault, TokenTypeNumber, TokenTypeDefault, TokenTypeKeyword},
		},
		{
			name:          "comments",
			input:         "x = 1 -- comment",
			keywords:      []string{},
			builtins:      []string{},
			commentPrefix: "--",
			expectedTypes: []string{TokenTypeDefault, TokenTypeDefault, TokenTypeDefault, TokenTypeDefault, TokenTypeNumber, TokenTypeDefault, TokenTypeComment},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input, tt.keywords, tt.builtins, tt.commentPrefix)
			var actualTypes []string
			for _, token := range tokens {
				actualTypes = append(actualTypes, token.Type)
			}
			assert.Equal(t, tt.expectedTypes, actualTypes)
		})
	}
}

func TestHighlightWithTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		keywords []string
		builtins []string
		contains []string
	}{
		{
			name:     "highlight keywords",
			input:    "if true then",
			keywords: []string{"if", "true", "then"},
			builtins: []string{},
			contains: []string{
				ColorKeyword + "if" + ColorReset,
				ColorKeyword + "true" + ColorReset,
				ColorKeyword + "then" + ColorReset,
			},
		},
		{
			name:     "highlight strings",
			input:    `print("hello world")`,
			keywords: []string{},
			builtins: []string{"print"},
			contains: []string{
				ColorBuiltin + "print" + ColorReset,
				ColorString + `"hello world"` + ColorReset,
			},
		},
		{
			name:     "highlight numbers",
			input:    "x = 42.5",
			keywords: []string{},
			builtins: []string{},
			contains: []string{
				ColorNumber + "42.5" + ColorReset,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightWithTokens(tt.input, tt.keywords, tt.builtins, "--")
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestHighlightWithCategories(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		keywords   []string
		categories []BuiltinCategory
		contains   []string
	}{
		{
			name:     "different builtin categories",
			input:    "table.insert(arr, print(x))",
			keywords: []string{},
			categories: []BuiltinCategory{
				{Words: []string{"print"}, Color: ColorBuiltin},
				{Words: []string{"table"}, Color: ColorFunction},
			},
			contains: []string{
				ColorFunction + "table" + ColorReset,
				ColorBuiltin + "print" + ColorReset,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightWithCategories(tt.input, tt.keywords, tt.categories, "--")
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}
