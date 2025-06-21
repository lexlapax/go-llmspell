// ABOUTME: Lua-specific syntax highlighting with keywords, built-ins, and standard library functions.
// ABOUTME: Implements highlighting for Lua 5.1 syntax and standard libraries.

package repl

// highlightLua applies Lua-specific syntax highlighting
func (h *SyntaxHighlighter) highlightLua(input string) string {
	result := input

	// Lua keywords
	luaKeywords := []string{
		"and", "break", "do", "else", "elseif", "end", "false", "for",
		"function", "if", "in", "local", "nil", "not", "or", "repeat",
		"return", "then", "true", "until", "while",
	}

	// Lua built-in functions
	luaBuiltins := []string{
		"print", "type", "tostring", "tonumber", "pairs", "ipairs",
		"next", "rawget", "rawset", "rawlen", "rawequal",
		"getmetatable", "setmetatable", "pcall", "xpcall",
		"error", "assert", "select", "unpack",
	}

	// Lua standard libraries
	luaLibraries := []string{
		"table", "string", "math", "io", "os", "debug", "coroutine",
	}

	// Apply highlighting patterns in order: strings and comments first (preserve their content)
	// then keywords, then decorative elements (operators, brackets, numbers)
	result = h.highlightStrings(result)
	result = h.highlightComments(result, "--")
	result = h.highlightKeywordsCarefully(result, luaKeywords, ColorKeyword)
	result = h.highlightKeywordsCarefully(result, luaBuiltins, ColorBuiltin)
	result = h.highlightKeywordsCarefully(result, luaLibraries, ColorFunction)
	result = h.highlightNumbersCarefully(result)
	// Skip operators and brackets for now to avoid conflicts - basic highlighting is sufficient
	// result = h.highlightOperatorsCarefully(result)
	// result = h.highlightBracketsCarefully(result)

	return result
}
