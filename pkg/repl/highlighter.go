// ABOUTME: Syntax highlighting functionality for REPL input with ANSI color codes.
// ABOUTME: Provides engine-specific syntax highlighting for Lua, JavaScript, and Tengo.

package repl

import (
	"regexp"
	"strings"
)

// ANSI color codes for syntax highlighting
const (
	ColorReset     = "\033[0m"
	ColorKeyword   = "\033[94m"  // Blue
	ColorString    = "\033[92m"  // Green
	ColorComment   = "\033[90m"  // Dark gray
	ColorNumber    = "\033[96m"  // Cyan
	ColorOperator  = "\033[93m"  // Yellow
	ColorFunction  = "\033[95m"  // Magenta
	ColorBuiltin   = "\033[91m"  // Red
	ColorBracket   = "\033[97m"  // White (bright)
)

// SyntaxHighlighter provides syntax highlighting for different script languages
type SyntaxHighlighter struct {
	engine string
}

// NewSyntaxHighlighter creates a new syntax highlighter for the specified engine
func NewSyntaxHighlighter(engine string) *SyntaxHighlighter {
	return &SyntaxHighlighter{engine: engine}
}

// Highlight applies syntax highlighting to the input text
func (h *SyntaxHighlighter) Highlight(input string) string {
	if input == "" {
		return input
	}

	switch h.engine {
	case "lua":
		return h.highlightLua(input)
	case "javascript", "js":
		return h.highlightJavaScript(input)
	case "tengo":
		return h.highlightTengo(input)
	default:
		return input // No highlighting for unknown engines
	}
}

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

	// Apply highlighting patterns
	result = h.highlightKeywords(result, luaKeywords, ColorKeyword)
	result = h.highlightKeywords(result, luaBuiltins, ColorBuiltin)
	result = h.highlightKeywords(result, luaLibraries, ColorFunction)
	result = h.highlightStrings(result)
	result = h.highlightComments(result, "--")
	result = h.highlightNumbers(result)
	result = h.highlightOperators(result)
	result = h.highlightBrackets(result)

	return result
}

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

	// Apply highlighting patterns
	result = h.highlightKeywords(result, jsKeywords, ColorKeyword)
	result = h.highlightKeywords(result, jsBuiltins, ColorBuiltin)
	result = h.highlightStrings(result)
	result = h.highlightComments(result, "//")
	result = h.highlightNumbers(result)
	result = h.highlightOperators(result)
	result = h.highlightBrackets(result)

	return result
}

// highlightTengo applies Tengo-specific syntax highlighting
func (h *SyntaxHighlighter) highlightTengo(input string) string {
	result := input

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

	// Apply highlighting patterns
	result = h.highlightKeywords(result, tengoKeywords, ColorKeyword)
	result = h.highlightKeywords(result, tengoBuiltins, ColorBuiltin)
	result = h.highlightStrings(result)
	result = h.highlightComments(result, "//")
	result = h.highlightNumbers(result)
	result = h.highlightOperators(result)
	result = h.highlightBrackets(result)

	return result
}

// highlightKeywords highlights specific keywords with the given color
func (h *SyntaxHighlighter) highlightKeywords(input string, keywords []string, color string) string {
	result := input
	for _, keyword := range keywords {
		// Use word boundaries to avoid partial matches
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, color+keyword+ColorReset)
	}
	return result
}

// highlightStrings highlights string literals (both single and double quotes)
func (h *SyntaxHighlighter) highlightStrings(input string) string {
	// Double-quoted strings
	doubleQuoteRe := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	result := doubleQuoteRe.ReplaceAllStringFunc(input, func(match string) string {
		return ColorString + match + ColorReset
	})

	// Single-quoted strings
	singleQuoteRe := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	result = singleQuoteRe.ReplaceAllStringFunc(result, func(match string) string {
		return ColorString + match + ColorReset
	})

	return result
}

// highlightComments highlights comments with the specified comment prefix
func (h *SyntaxHighlighter) highlightComments(input string, commentPrefix string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if idx := strings.Index(line, commentPrefix); idx != -1 {
			// Check if comment is inside a string
			beforeComment := line[:idx]
			if h.isInsideString(beforeComment) {
				result = append(result, line)
				continue
			}
			
			// Highlight the comment part
			highlighted := line[:idx] + ColorComment + line[idx:] + ColorReset
			result = append(result, highlighted)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// highlightNumbers highlights numeric literals
func (h *SyntaxHighlighter) highlightNumbers(input string) string {
	// Match integers and floating point numbers
	numberRe := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	return numberRe.ReplaceAllStringFunc(input, func(match string) string {
		return ColorNumber + match + ColorReset
	})
}

// highlightOperators highlights common operators
func (h *SyntaxHighlighter) highlightOperators(input string) string {
	operators := []string{
		"==", "!=", "<=", ">=", "&&", "||", "++", "--",
		"+=", "-=", "*=", "/=", "%=", "=>",
		"+", "-", "*", "/", "%", "=", "<", ">", "!", "&", "|",
	}

	result := input
	for _, op := range operators {
		escaped := regexp.QuoteMeta(op)
		re := regexp.MustCompile(escaped)
		result = re.ReplaceAllString(result, ColorOperator+op+ColorReset)
	}

	return result
}

// highlightBrackets highlights brackets, parentheses, and braces
func (h *SyntaxHighlighter) highlightBrackets(input string) string {
	brackets := []string{"(", ")", "[", "]", "{", "}"}
	result := input

	for _, bracket := range brackets {
		escaped := regexp.QuoteMeta(bracket)
		re := regexp.MustCompile(escaped)
		result = re.ReplaceAllString(result, ColorBracket+bracket+ColorReset)
	}

	return result
}

// isInsideString checks if a position in the text is inside a string literal
func (h *SyntaxHighlighter) isInsideString(text string) bool {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, char := range text {
		if escaped {
			escaped = false
			continue
		}

		switch char {
		case '\\':
			escaped = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		}
	}

	return inSingleQuote || inDoubleQuote
}

// StripColors removes ANSI color codes from text
func StripColors(text string) string {
	ansiRe := regexp.MustCompile(`\033\[[0-9;]*m`)
	return ansiRe.ReplaceAllString(text, "")
}