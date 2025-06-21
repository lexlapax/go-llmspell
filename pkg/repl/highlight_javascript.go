// ABOUTME: JavaScript-specific syntax highlighting with keywords, built-in objects, and global functions.
// ABOUTME: Implements highlighting for JavaScript/ECMAScript syntax and standard objects.

package repl

// highlightJavaScript applies JavaScript-specific syntax highlighting
func (h *SyntaxHighlighter) highlightJavaScript(input string) string {
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

	// Use tokenization approach for proper highlighting
	return highlightWithTokens(input, jsKeywords, jsBuiltins, "//")
}
