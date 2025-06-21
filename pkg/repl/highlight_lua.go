// ABOUTME: Lua-specific syntax highlighting with keywords, built-ins, and standard library functions.
// ABOUTME: Implements highlighting for Lua 5.1 syntax and standard libraries.

package repl

// highlightLua applies Lua-specific syntax highlighting
func (h *SyntaxHighlighter) highlightLua(input string) string {
	// Lua keywords
	luaKeywords := []string{
		"and", "break", "do", "else", "elseif", "end", "false", "for",
		"function", "if", "in", "local", "nil", "not", "or", "repeat",
		"return", "then", "true", "until", "while",
	}

	// Lua built-in categories
	luaBuiltinCategories := []BuiltinCategory{
		{
			Words: []string{
				"print", "type", "tostring", "tonumber", "pairs", "ipairs",
				"next", "rawget", "rawset", "rawlen", "rawequal",
				"getmetatable", "setmetatable", "pcall", "xpcall",
				"error", "assert", "select", "unpack",
			},
			Color: ColorBuiltin,
		},
		{
			Words: []string{
				"table", "string", "math", "io", "os", "debug", "coroutine",
			},
			Color: ColorFunction,
		},
	}

	// Use categorized tokenization approach
	return highlightWithCategories(input, luaKeywords, luaBuiltinCategories, "--")
}
