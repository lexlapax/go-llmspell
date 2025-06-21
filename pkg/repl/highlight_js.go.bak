// ABOUTME: JavaScript-specific syntax highlighting with keywords, built-in objects, and global functions.
// ABOUTME: Implements highlighting for JavaScript/ECMAScript syntax and standard objects.

package repl

// highlightJavaScript applies JavaScript-specific syntax highlighting
func (h *SyntaxHighlighter) highlightJavaScript(input string) string {
	result := input

	// JavaScript keywords
	jsKeywords := []string{
		"break", "case", "catch", "class", "const", "continue", "debugger",
		"default", "delete", "do", "else", "export", "extends", "finally",
		"for", "function", "if", "import", "in", "instanceof", "let", "new",
		"return", "super", "switch", "this", "throw", "try", "typeof",
		"var", "void", "while", "with", "yield", "true", "false", "null",
		"undefined",
	}

	// JavaScript built-in objects and functions
	jsBuiltins := []string{
		"Array", "Boolean", "Date", "Error", "Function", "JSON", "Math",
		"Number", "Object", "Promise", "RegExp", "String", "Symbol",
		"console", "parseInt", "parseFloat", "isNaN", "isFinite",
		"setTimeout", "setInterval", "clearTimeout", "clearInterval",
	}

	// Apply highlighting patterns in order: strings and comments first
	result = h.highlightStrings(result)
	result = h.highlightComments(result, "//")
	result = h.highlightKeywordsCarefully(result, jsKeywords, ColorKeyword)
	result = h.highlightKeywordsCarefully(result, jsBuiltins, ColorBuiltin)
	result = h.highlightNumbersCarefully(result)
	// Skip operators and brackets for now to avoid conflicts
	// result = h.highlightOperatorsCarefully(result)
	// result = h.highlightBracketsCarefully(result)

	return result
}
