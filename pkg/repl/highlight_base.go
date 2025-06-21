// ABOUTME: Core syntax highlighting infrastructure with ANSI color codes and generic highlighting methods.
// ABOUTME: Provides base functionality that language-specific highlighters can build upon.

package repl

import (
	"regexp"
	"strings"
)

// ANSI color codes for syntax highlighting
const (
	ColorReset    = "\033[0m"
	ColorKeyword  = "\033[94m" // Blue
	ColorString   = "\033[92m" // Green
	ColorComment  = "\033[90m" // Dark gray
	ColorNumber   = "\033[96m" // Cyan
	ColorOperator = "\033[93m" // Yellow
	ColorFunction = "\033[95m" // Magenta
	ColorBuiltin  = "\033[91m" // Red
	ColorBracket  = "\033[97m" // White (bright)
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

// highlightKeywordsCarefully highlights keywords while avoiding ANSI escape sequences
func (h *SyntaxHighlighter) highlightKeywordsCarefully(input string, keywords []string, color string) string {
	result := input
	for _, keyword := range keywords {
		// Use word boundaries and negative lookbehind/ahead to avoid ANSI sequences
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)

		// Find all matches and check each one
		matches := re.FindAllStringIndex(result, -1)
		for i := len(matches) - 1; i >= 0; i-- { // Process in reverse to maintain positions
			start, end := matches[i][0], matches[i][1]
			match := result[start:end]

			// Check if this match is inside an ANSI escape sequence (already highlighted content)
			if h.isInsideANSISequence(result, start) {
				continue // Don't highlight if inside ANSI sequence
			}

			// Replace this specific match
			highlighted := color + match + ColorReset
			result = result[:start] + highlighted + result[end:]
		}
	}
	return result
}

// highlightNumbersCarefully highlights numbers while avoiding ANSI escape sequences
func (h *SyntaxHighlighter) highlightNumbersCarefully(input string) string {
	numberRe := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	return numberRe.ReplaceAllStringFunc(input, func(match string) string {
		matchPos := strings.Index(input, match)
		if h.isInsideANSISequence(input, matchPos) {
			return match
		}
		return ColorNumber + match + ColorReset
	})
}

// isInsideANSISequence checks if a position is inside an ANSI-highlighted content
func (h *SyntaxHighlighter) isInsideANSISequence(text string, pos int) bool {
	// Look backwards from pos to find the most recent color start
	colorStart := -1
	colorEnd := -1

	// Find the most recent color sequence before our position
	for i := range pos {
		if i+1 < len(text) && text[i] == '\033' && text[i+1] == '[' {
			// Found ANSI escape start
			for j := i + 2; j < len(text); j++ {
				if text[j] == 'm' {
					if colorStart == -1 || i > colorStart {
						colorStart = i
						// Look for the corresponding reset sequence
						for k := j + 1; k < len(text); k++ {
							if k+3 < len(text) && text[k:k+4] == "\033[0m" {
								colorEnd = k + 4
								break
							}
						}
					}
					break
				}
			}
		}
	}

	// Check if our position is between a color start and color end
	if colorStart != -1 && colorEnd != -1 {
		return pos > colorStart && pos < colorEnd
	}

	return false
}

// StripColors removes ANSI color codes from text
func StripColors(text string) string {
	ansiRe := regexp.MustCompile(`\033\[[0-9;]*m`)
	return ansiRe.ReplaceAllString(text, "")
}
