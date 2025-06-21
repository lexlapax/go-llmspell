// ABOUTME: Auto-completion functionality for REPL including command and language-specific completions.
// ABOUTME: Implements readline.AutoCompleter interface for integration with readline library.

package repl

import (
	"sort"
	"strings"
)

// Completer implements readline.AutoCompleter for REPL auto-completion
type Completer struct {
	repl *BaseREPL
}

// NewCompleter creates a new completer instance for the given REPL
func NewCompleter(repl *BaseREPL) *Completer {
	return &Completer{repl: repl}
}

// Do implements the readline.AutoCompleter interface
// It processes the current line and position to provide completion suggestions
func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	input := string(line[:pos])
	completions := c.GetCompletions(input)

	var results [][]rune
	for _, completion := range completions {
		results = append(results, []rune(completion))
	}

	return results, len(input)
}

// GetCompletions provides auto-completion suggestions for the given input
func (c *Completer) GetCompletions(input string) []string {
	var completions []string

	// REPL command completion
	if strings.HasPrefix(input, ".") {
		completions = append(completions, c.getREPLCommandCompletions(input)...)
	} else {
		// Language-specific completion
		completions = append(completions, c.getLanguageCompletions(input)...)
	}

	sort.Strings(completions)
	return completions
}

// getREPLCommandCompletions returns completions for REPL commands (starting with .)
func (c *Completer) getREPLCommandCompletions(input string) []string {
	var completions []string

	commands := GetBuiltinCommands()
	for name := range commands {
		cmdName := "." + name
		if strings.HasPrefix(cmdName, input) {
			completions = append(completions, cmdName)
		}
	}

	return completions
}

// getLanguageCompletions returns language-specific completions based on the engine
func (c *Completer) getLanguageCompletions(input string) []string {
	switch c.repl.config.Engine {
	case "lua":
		return c.getLuaCompletions(input)
	case "javascript", "js":
		return c.getJavaScriptCompletions(input)
	case "tengo":
		return c.getTengoCompletions(input)
	default:
		return []string{}
	}
}

// getLuaCompletions returns Lua-specific keyword and built-in completions
func (c *Completer) getLuaCompletions(input string) []string {
	var completions []string

	luaKeywords := []string{
		// Lua keywords
		"and", "break", "do", "else", "elseif", "end", "false", "for",
		"function", "if", "in", "local", "nil", "not", "or", "repeat",
		"return", "then", "true", "until", "while",
		// Built-in functions
		"print", "type", "tostring", "tonumber", "pairs", "ipairs",
		"next", "rawget", "rawset", "rawlen", "rawequal",
		"getmetatable", "setmetatable", "pcall", "xpcall",
		"error", "assert", "select", "unpack",
		// Standard libraries
		"table", "string", "math", "io", "os", "debug", "coroutine",
		// Table library
		"table.insert", "table.remove", "table.concat", "table.sort",
		// String library
		"string.len", "string.sub", "string.find", "string.match",
		"string.gsub", "string.format", "string.upper", "string.lower",
		// Math library
		"math.abs", "math.ceil", "math.floor", "math.max", "math.min",
		"math.random", "math.sqrt", "math.sin", "math.cos", "math.pi",
	}

	for _, keyword := range luaKeywords {
		if strings.HasPrefix(keyword, input) {
			completions = append(completions, keyword)
		}
	}

	return completions
}

// getJavaScriptCompletions returns JavaScript-specific completions
// Currently returns basic keywords - will be expanded when JS engine is implemented
func (c *Completer) getJavaScriptCompletions(input string) []string {
	var completions []string

	jsKeywords := []string{
		// JavaScript keywords
		"break", "case", "catch", "class", "const", "continue", "debugger",
		"default", "delete", "do", "else", "export", "extends", "finally",
		"for", "function", "if", "import", "in", "instanceof", "let", "new",
		"return", "super", "switch", "this", "throw", "try", "typeof",
		"var", "void", "while", "with", "yield",
		// Built-in objects
		"Array", "Boolean", "Date", "Error", "Function", "JSON", "Math",
		"Number", "Object", "Promise", "RegExp", "String", "Symbol",
		// Global functions
		"console.log", "console.error", "console.warn", "console.info",
		"parseInt", "parseFloat", "isNaN", "isFinite", "setTimeout", "setInterval",
	}

	for _, keyword := range jsKeywords {
		if strings.HasPrefix(keyword, input) {
			completions = append(completions, keyword)
		}
	}

	return completions
}

// getTengoCompletions returns Tengo-specific completions
// Currently returns basic keywords - will be expanded when Tengo engine is implemented
func (c *Completer) getTengoCompletions(input string) []string {
	var completions []string

	tengoKeywords := []string{
		// Tengo keywords
		"break", "continue", "else", "for", "func", "if", "return",
		"true", "false", "undefined", "import", "in",
		// Built-in functions
		"len", "copy", "append", "string", "int", "float", "bool",
		"char", "bytes", "time", "is_string", "is_int", "is_float",
		"is_bool", "is_char", "is_bytes", "is_array", "is_map",
		"is_undefined", "is_function", "is_callable", "is_iterable",
		"type_name", "format", "range", "printf", "sprintf", "print",
	}

	for _, keyword := range tengoKeywords {
		if strings.HasPrefix(keyword, input) {
			completions = append(completions, keyword)
		}
	}

	return completions
}
