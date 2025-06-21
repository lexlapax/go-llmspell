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

// Token represents a piece of text with its type
type Token struct {
	Type  string
	Value string
	Start int
	End   int
}

const (
	TokenTypeString  = "string"
	TokenTypeComment = "comment"
	TokenTypeKeyword = "keyword"
	TokenTypeBuiltin = "builtin"
	TokenTypeNumber  = "number"
	TokenTypeDefault = "default"
)

// BuiltinCategory represents different categories of built-in identifiers
type BuiltinCategory struct {
	Words []string
	Color string
}

// Tokenize breaks the input into tokens (exported for testing)
func Tokenize(input string, keywords []string, builtins []string, commentPrefix string) []Token {
	tokens := []Token{}
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		// Check for strings
		if i < len(runes) && (runes[i] == '"' || runes[i] == '\'') {
			quote := runes[i]
			start := i
			i++
			for i < len(runes) && (runes[i] != quote || (i > 0 && runes[i-1] == '\\')) {
				if runes[i] == '\\' && i+1 < len(runes) {
					i++ // Skip escaped character
				}
				i++
			}
			if i < len(runes) {
				i++ // Include closing quote
			}
			tokens = append(tokens, Token{
				Type:  TokenTypeString,
				Value: string(runes[start:i]),
				Start: start,
				End:   i,
			})
			continue
		}

		// Check for comments
		if commentPrefix != "" && i+len(commentPrefix) <= len(runes) {
			if string(runes[i:i+len(commentPrefix)]) == commentPrefix {
				start := i
				for i < len(runes) && runes[i] != '\n' {
					i++
				}
				tokens = append(tokens, Token{
					Type:  TokenTypeComment,
					Value: string(runes[start:i]),
					Start: start,
					End:   i,
				})
				continue
			}
		}

		// Check for numbers first (before word boundaries)
		if runes[i] >= '0' && runes[i] <= '9' {
			start := i
			// Match integer part
			for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
				i++
			}
			// Check for decimal part
			if i < len(runes) && runes[i] == '.' && i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9' {
				i++ // consume dot
				for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
					i++
				}
			}
			tokens = append(tokens, Token{
				Type:  TokenTypeNumber,
				Value: string(runes[start:i]),
				Start: start,
				End:   i,
			})
			continue
		}

		// Check for word boundaries
		if isWordChar(runes[i]) {
			start := i
			for i < len(runes) && isWordChar(runes[i]) {
				i++
			}
			word := string(runes[start:i])

			// Check if it's a keyword or builtin
			tokenType := TokenTypeDefault
			for _, kw := range keywords {
				if word == kw {
					tokenType = TokenTypeKeyword
					break
				}
			}
			if tokenType == TokenTypeDefault {
				for _, bi := range builtins {
					if word == bi {
						tokenType = TokenTypeBuiltin
						break
					}
				}
			}

			// Check if it's a number
			if tokenType == TokenTypeDefault && isNumber(word) {
				tokenType = TokenTypeNumber
			}

			tokens = append(tokens, Token{
				Type:  tokenType,
				Value: word,
				Start: start,
				End:   i,
			})
			continue
		}

		// Default: single character token
		tokens = append(tokens, Token{
			Type:  TokenTypeDefault,
			Value: string(runes[i]),
			Start: i,
			End:   i + 1,
		})
		i++
	}

	return tokens
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

func isNumber(s string) bool {
	matched, _ := regexp.MatchString(`^\d+(\.\d+)?$`, s)
	return matched
}

// highlightWithTokens applies highlighting using tokenization
func highlightWithTokens(input string, keywords []string, builtins []string, commentPrefix string) string {
	tokens := Tokenize(input, keywords, builtins, commentPrefix)

	var result strings.Builder
	for _, token := range tokens {
		switch token.Type {
		case TokenTypeString:
			result.WriteString(ColorString + token.Value + ColorReset)
		case TokenTypeComment:
			result.WriteString(ColorComment + token.Value + ColorReset)
		case TokenTypeKeyword:
			result.WriteString(ColorKeyword + token.Value + ColorReset)
		case TokenTypeBuiltin:
			result.WriteString(ColorBuiltin + token.Value + ColorReset)
		case TokenTypeNumber:
			result.WriteString(ColorNumber + token.Value + ColorReset)
		default:
			result.WriteString(token.Value)
		}
	}

	return result.String()
}

// ExtendedToken includes metadata for additional information
type ExtendedToken struct {
	Token
	Metadata interface{}
}

// tokenizeWithCategories breaks the input into tokens with category support
func tokenizeWithCategories(input string, keywords []string, builtinCategories []BuiltinCategory, commentPrefix string) []ExtendedToken {
	tokens := []ExtendedToken{}
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		// Check for strings
		if i < len(runes) && (runes[i] == '"' || runes[i] == '\'') {
			quote := runes[i]
			start := i
			i++
			for i < len(runes) && (runes[i] != quote || (i > 0 && runes[i-1] == '\\')) {
				if runes[i] == '\\' && i+1 < len(runes) {
					i++ // Skip escaped character
				}
				i++
			}
			if i < len(runes) {
				i++ // Include closing quote
			}
			tokens = append(tokens, ExtendedToken{
				Token: Token{
					Type:  TokenTypeString,
					Value: string(runes[start:i]),
					Start: start,
					End:   i,
				},
			})
			continue
		}

		// Check for comments
		if commentPrefix != "" && i+len(commentPrefix) <= len(runes) {
			if string(runes[i:i+len(commentPrefix)]) == commentPrefix {
				start := i
				for i < len(runes) && runes[i] != '\n' {
					i++
				}
				tokens = append(tokens, ExtendedToken{
					Token: Token{
						Type:  TokenTypeComment,
						Value: string(runes[start:i]),
						Start: start,
						End:   i,
					},
				})
				continue
			}
		}

		// Check for numbers first (before word boundaries)
		if runes[i] >= '0' && runes[i] <= '9' {
			start := i
			// Match integer part
			for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
				i++
			}
			// Check for decimal part
			if i < len(runes) && runes[i] == '.' && i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9' {
				i++ // consume dot
				for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
					i++
				}
			}
			tokens = append(tokens, ExtendedToken{
				Token: Token{
					Type:  TokenTypeNumber,
					Value: string(runes[start:i]),
					Start: start,
					End:   i,
				},
			})
			continue
		}

		// Check for word boundaries
		if isWordChar(runes[i]) {
			start := i
			for i < len(runes) && isWordChar(runes[i]) {
				i++
			}
			word := string(runes[start:i])

			// Check if it's a keyword
			tokenType := TokenTypeDefault
			var metadata interface{}

			for _, kw := range keywords {
				if word == kw {
					tokenType = TokenTypeKeyword
					break
				}
			}

			// Check builtin categories
			if tokenType == TokenTypeDefault {
				for _, category := range builtinCategories {
					for _, bi := range category.Words {
						if word == bi {
							tokenType = "builtin"
							metadata = category.Color
							break
						}
					}
					if tokenType == "builtin" {
						break
					}
				}
			}

			// Check if it's a number
			if tokenType == TokenTypeDefault && isNumber(word) {
				tokenType = TokenTypeNumber
			}

			tokens = append(tokens, ExtendedToken{
				Token: Token{
					Type:  tokenType,
					Value: word,
					Start: start,
					End:   i,
				},
				Metadata: metadata,
			})
			continue
		}

		// Default: single character token
		tokens = append(tokens, ExtendedToken{
			Token: Token{
				Type:  TokenTypeDefault,
				Value: string(runes[i]),
				Start: i,
				End:   i + 1,
			},
		})
		i++
	}

	return tokens
}

// highlightWithCategories applies highlighting using tokenization with different builtin categories
func highlightWithCategories(input string, keywords []string, builtinCategories []BuiltinCategory, commentPrefix string) string {
	tokens := tokenizeWithCategories(input, keywords, builtinCategories, commentPrefix)

	var result strings.Builder
	for _, token := range tokens {
		switch token.Type {
		case TokenTypeString:
			result.WriteString(ColorString + token.Value + ColorReset)
		case TokenTypeComment:
			result.WriteString(ColorComment + token.Value + ColorReset)
		case TokenTypeKeyword:
			result.WriteString(ColorKeyword + token.Value + ColorReset)
		case TokenTypeNumber:
			result.WriteString(ColorNumber + token.Value + ColorReset)
		default:
			// Check if it's a categorized builtin
			if color, ok := token.Metadata.(string); ok && token.Type == "builtin" {
				result.WriteString(color + token.Value + ColorReset)
			} else {
				result.WriteString(token.Value)
			}
		}
	}

	return result.String()
}

// StripColors removes ANSI color codes from text
func StripColors(text string) string {
	ansiRe := regexp.MustCompile(`\033\[[0-9;]*m`)
	return ansiRe.ReplaceAllString(text, "")
}
