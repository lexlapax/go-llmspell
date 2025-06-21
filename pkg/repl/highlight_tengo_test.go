// ABOUTME: Tests for Tengo-specific syntax highlighting including keywords and built-in functions.
// ABOUTME: Verifies proper highlighting of Tengo scripting language syntax elements.

package repl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyntaxHighlighter_HighlightTengo(t *testing.T) {
	highlighter := NewSyntaxHighlighter("tengo")

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "keywords",
			input: "func test() { if true { print('hello'); } }",
			contains: []string{
				ColorKeyword + "func" + ColorReset,
				ColorKeyword + "if" + ColorReset,
				ColorKeyword + "true" + ColorReset,
			},
		},
		{
			name:  "built-in functions",
			input: "print(len(array)) is_string(value)",
			contains: []string{
				ColorBuiltin + "print" + ColorReset,
				ColorBuiltin + "len" + ColorReset,
				ColorBuiltin + "is_string" + ColorReset,
			},
		},
		{
			name:  "comments",
			input: "x := 5 // this is a comment",
			contains: []string{
				ColorComment + "// this is a comment" + ColorReset,
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
