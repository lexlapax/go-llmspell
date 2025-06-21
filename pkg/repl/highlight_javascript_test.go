// ABOUTME: Tests for JavaScript-specific syntax highlighting including keywords and built-in objects.
// ABOUTME: Verifies proper highlighting of JavaScript/ECMAScript syntax elements.

package repl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyntaxHighlighter_HighlightJavaScript(t *testing.T) {
	highlighter := NewSyntaxHighlighter("javascript")

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "keywords",
			input: "function test() { if (true) { console.log('hello'); } }",
			contains: []string{
				ColorKeyword + "function" + ColorReset,
				ColorKeyword + "if" + ColorReset,
				ColorKeyword + "true" + ColorReset,
			},
		},
		{
			name:  "built-in objects",
			input: "console.log(JSON.stringify(Math.PI))",
			contains: []string{
				ColorBuiltin + "console" + ColorReset,
				ColorBuiltin + "JSON" + ColorReset,
				ColorBuiltin + "Math" + ColorReset,
			},
		},
		{
			name:  "comments",
			input: "var x = 5; // this is a comment",
			contains: []string{
				ColorComment + "// this is a comment" + ColorReset,
			},
		},
		{
			name:  "const and let",
			input: "const x = 5; let y = 10; var z = 15;",
			contains: []string{
				ColorKeyword + "const" + ColorReset,
				ColorKeyword + "let" + ColorReset,
				ColorKeyword + "var" + ColorReset,
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
