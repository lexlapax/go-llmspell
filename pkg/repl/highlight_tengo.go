// ABOUTME: Tengo-specific syntax highlighting with keywords and built-in functions.
// ABOUTME: Implements highlighting for Tengo scripting language syntax and standard functions.

package repl

// highlightTengo applies Tengo-specific syntax highlighting
func (h *SyntaxHighlighter) highlightTengo(input string) string {
	// Tengo keywords
	tengoKeywords := []string{
		"break", "continue", "else", "for", "func", "if", "return",
		"true", "false", "undefined", "import", "in",
	}

	// Tengo built-in functions
	tengoBuiltins := []string{
		"len", "copy", "append", "string", "int", "float", "bool",
		"char", "bytes", "time", "is_string", "is_int", "is_float",
		"is_bool", "is_char", "is_bytes", "is_array", "is_map",
		"is_undefined", "is_function", "is_callable", "is_iterable",
		"type_name", "format", "range", "printf", "sprintf", "print",
	}

	// Use tokenization approach for proper highlighting
	return highlightWithTokens(input, tengoKeywords, tengoBuiltins, "//")
}
