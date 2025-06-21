// ABOUTME: This file implements user-friendly error formatting with context, suggestions, and debug information.
// ABOUTME: It provides different formatting modes for terminal output with color support and structured formats.

package errors

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FormatterOptions configures error formatting
type FormatterOptions struct {
	ShowStackTrace  bool
	ShowContext     bool
	ShowSuggestions bool
	ShowTimestamp   bool
	ColorOutput     bool
	MaxStackFrames  int
	MaxContextItems int
	IndentLevel     int
	DebugMode       bool
}

// DefaultFormatterOptions returns default formatter options
func DefaultFormatterOptions() FormatterOptions {
	return FormatterOptions{
		ShowStackTrace:  false,
		ShowContext:     true,
		ShowSuggestions: true,
		ShowTimestamp:   false,
		ColorOutput:     true,
		MaxStackFrames:  10,
		MaxContextItems: 5,
		IndentLevel:     2,
		DebugMode:       false,
	}
}

// DebugFormatterOptions returns formatter options for debug mode
func DebugFormatterOptions() FormatterOptions {
	opts := DefaultFormatterOptions()
	opts.ShowStackTrace = true
	opts.ShowTimestamp = true
	opts.DebugMode = true
	opts.MaxStackFrames = 20
	return opts
}

// Formatter formats errors for display
type Formatter struct {
	options FormatterOptions
	writer  io.Writer
}

// NewFormatter creates a new error formatter
func NewFormatter(options FormatterOptions) *Formatter {
	// Disable colors if not a terminal or explicitly disabled
	if !isTerminal() || os.Getenv("NO_COLOR") != "" {
		options.ColorOutput = false
	}

	return &Formatter{
		options: options,
		writer:  os.Stderr,
	}
}

// SetWriter sets the output writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Format formats an error for display
func (f *Formatter) Format(err error) string {
	if err == nil {
		return ""
	}

	var b strings.Builder

	// Check if it's a Chain
	if chain, ok := err.(*Chain); ok {
		formatted := f.FormatChain(chain)
		b.WriteString(formatted)
	} else if IsSpellError(err) {
		// Check if it's a SpellError
		spellErr, _ := err.(*SpellError)
		f.formatSpellError(&b, spellErr)
	} else {
		// Format as generic error
		f.formatGenericError(&b, err)
	}

	return b.String()
}

// Print prints an error to the configured writer
func (f *Formatter) Print(err error) {
	if err == nil {
		return
	}

	_, _ = fmt.Fprint(f.writer, f.Format(err))
}

// formatSpellError formats a SpellError
func (f *Formatter) formatSpellError(b *strings.Builder, err *SpellError) {
	// Timestamp if enabled
	if f.options.ShowTimestamp {
		b.WriteString(f.gray(fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))))
	}

	// Error header
	b.WriteString(f.formatErrorHeader(err))
	b.WriteString("\n")

	// Error message
	b.WriteString(f.formatErrorMessage(err))
	b.WriteString("\n")

	// Context if available
	if f.options.ShowContext && len(err.Context) > 0 {
		b.WriteString("\n")
		b.WriteString(f.formatContext(err.Context))
		b.WriteString("\n")
	}

	// Suggestions if available
	if f.options.ShowSuggestions && len(err.Suggestions) > 0 {
		b.WriteString("\n")
		b.WriteString(f.formatSuggestions(err.Suggestions))
		b.WriteString("\n")
	}

	// Stack trace if enabled
	if f.options.ShowStackTrace && len(err.StackTrace) > 0 {
		b.WriteString("\n")
		b.WriteString(f.formatStackTrace(err.StackTrace))
		b.WriteString("\n")
	}

	// Debug info if enabled
	if f.options.DebugMode {
		b.WriteString("\n")
		b.WriteString(f.formatDebugInfo(err))
		b.WriteString("\n")
	}
}

// formatGenericError formats a generic error
func (f *Formatter) formatGenericError(b *strings.Builder, err error) {
	if f.options.ShowTimestamp {
		b.WriteString(f.gray(fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))))
	}

	b.WriteString(f.red("Error: "))
	b.WriteString(err.Error())
	b.WriteString("\n")
}

// formatErrorHeader formats the error header with category
func (f *Formatter) formatErrorHeader(err *SpellError) string {
	icon := f.getErrorIcon(err.Category)
	category := f.formatCategory(err.Category)

	return fmt.Sprintf("%s %s", icon, category)
}

// formatErrorMessage formats the main error message
func (f *Formatter) formatErrorMessage(err *SpellError) string {
	indent := strings.Repeat(" ", f.options.IndentLevel)

	if err.Cause != nil {
		return fmt.Sprintf("%s%s\n%s%s %s",
			indent,
			f.bold(err.Message),
			indent,
			f.gray("└─"),
			err.Cause.Error())
	}

	return indent + f.bold(err.Message)
}

// formatContext formats error context
func (f *Formatter) formatContext(context map[string]interface{}) string {
	var b strings.Builder

	b.WriteString(f.yellow("Context:"))
	b.WriteString("\n")

	indent := strings.Repeat(" ", f.options.IndentLevel)
	count := 0

	for key, value := range context {
		if count >= f.options.MaxContextItems {
			remaining := len(context) - count
			b.WriteString(fmt.Sprintf("%s%s ... and %d more\n",
				indent,
				f.gray("•"),
				remaining))
			break
		}

		b.WriteString(fmt.Sprintf("%s%s %s: %v\n",
			indent,
			f.gray("•"),
			f.cyan(key),
			f.formatValue(value)))
		count++
	}

	return b.String()
}

// formatSuggestions formats error suggestions
func (f *Formatter) formatSuggestions(suggestions []string) string {
	var b strings.Builder

	b.WriteString(f.green("Suggestions:"))
	b.WriteString("\n")

	indent := strings.Repeat(" ", f.options.IndentLevel)

	for i, suggestion := range suggestions {
		icon := "•"
		if i == 0 {
			icon = "→"
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n",
			indent,
			f.green(icon),
			suggestion))
	}

	return b.String()
}

// formatStackTrace formats the stack trace
func (f *Formatter) formatStackTrace(frames []StackFrame) string {
	var b strings.Builder

	b.WriteString(f.magenta("Stack trace:"))
	b.WriteString("\n")

	indent := strings.Repeat(" ", f.options.IndentLevel)
	maxFrames := f.options.MaxStackFrames
	if maxFrames > len(frames) {
		maxFrames = len(frames)
	}

	for i := 0; i < maxFrames; i++ {
		frame := frames[i]

		// Format function name
		funcName := frame.Function
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}

		// Format file path
		filePath := frame.File
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, filePath); err == nil {
				filePath = rel
			}
		}

		b.WriteString(fmt.Sprintf("%s%d. %s\n%s   %s:%d\n",
			indent,
			i+1,
			f.bold(funcName),
			indent,
			f.gray(filePath),
			frame.Line))
	}

	if len(frames) > maxFrames {
		b.WriteString(fmt.Sprintf("%s   ... %d more frames\n",
			indent,
			len(frames)-maxFrames))
	}

	return b.String()
}

// formatDebugInfo formats debug information
func (f *Formatter) formatDebugInfo(err *SpellError) string {
	var b strings.Builder

	b.WriteString(f.gray("Debug info:"))
	b.WriteString("\n")

	indent := strings.Repeat(" ", f.options.IndentLevel)

	b.WriteString(fmt.Sprintf("%sCategory: %s\n", indent, err.Category))
	b.WriteString(fmt.Sprintf("%sExit code: %d\n", indent, err.ExitCode()))

	if err.Cause != nil {
		b.WriteString(fmt.Sprintf("%sUnderlying type: %T\n", indent, err.Cause))
	}

	return b.String()
}

// formatCategory formats the error category
func (f *Formatter) formatCategory(category ErrorCategory) string {
	categoryStr := string(category)
	// Capitalize first letter
	if len(categoryStr) > 0 {
		categoryStr = strings.ToUpper(categoryStr[:1]) + categoryStr[1:]
	}

	switch category {
	case CategoryUsage:
		return f.yellow(categoryStr + " Error")
	case CategoryConfig:
		return f.blue(categoryStr + " Error")
	case CategoryScript:
		return f.cyan(categoryStr + " Error")
	case CategoryEngine:
		return f.magenta(categoryStr + " Error")
	case CategorySecurity:
		return f.red(categoryStr + " Error")
	case CategoryNetwork:
		return f.yellow(categoryStr + " Error")
	case CategoryTimeout:
		return f.yellow(categoryStr + " Error")
	case CategoryResource:
		return f.red(categoryStr + " Error")
	case CategoryValidation:
		return f.yellow(categoryStr + " Error")
	case CategoryDependency:
		return f.blue(categoryStr + " Error")
	case CategoryIO:
		return f.yellow(categoryStr + " Error")
	case CategoryInterrupted:
		return f.gray(categoryStr)
	default:
		return f.red("Error")
	}
}

// getErrorIcon returns an icon for the error category
func (f *Formatter) getErrorIcon(category ErrorCategory) string {
	switch category {
	case CategoryUsage:
		return f.yellow("⚠")
	case CategoryConfig:
		return f.blue("⚙")
	case CategoryScript:
		return f.cyan("📜")
	case CategoryEngine:
		return f.magenta("⚡")
	case CategorySecurity:
		return f.red("🔒")
	case CategoryNetwork:
		return f.yellow("🌐")
	case CategoryTimeout:
		return f.yellow("⏱")
	case CategoryResource:
		return f.red("💾")
	case CategoryValidation:
		return f.yellow("✓")
	case CategoryDependency:
		return f.blue("📦")
	case CategoryIO:
		return f.yellow("💾")
	case CategoryInterrupted:
		return f.gray("⛔")
	default:
		return f.red("✗")
	}
}

// formatValue formats a context value
func (f *Formatter) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 50 {
			return fmt.Sprintf("%q...", v[:47])
		}
		return fmt.Sprintf("%q", v)
	case error:
		return v.Error()
	default:
		s := fmt.Sprintf("%v", v)
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	}
}

// Color functions
func (f *Formatter) red(s string) string {
	if f.options.ColorOutput {
		return "\033[31m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) green(s string) string {
	if f.options.ColorOutput {
		return "\033[32m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) yellow(s string) string {
	if f.options.ColorOutput {
		return "\033[33m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) blue(s string) string {
	if f.options.ColorOutput {
		return "\033[34m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) magenta(s string) string {
	if f.options.ColorOutput {
		return "\033[35m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) cyan(s string) string {
	if f.options.ColorOutput {
		return "\033[36m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) gray(s string) string {
	if f.options.ColorOutput {
		return "\033[90m" + s + "\033[0m"
	}
	return s
}

func (f *Formatter) bold(s string) string {
	if f.options.ColorOutput {
		return "\033[1m" + s + "\033[0m"
	}
	return s
}

// isTerminal checks if output is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// FormatChain formats an error chain
func (f *Formatter) FormatChain(chain *Chain) string {
	if chain == nil || !chain.HasErrors() {
		return ""
	}

	var b strings.Builder

	b.WriteString(f.red(fmt.Sprintf("Multiple errors (%d):\n", len(chain.errors))))

	for i, err := range chain.errors {
		b.WriteString(fmt.Sprintf("\n%s %d. ", f.gray("▸"), i+1))

		// Format each error without the full header
		if spellErr, ok := err.(*SpellError); ok {
			b.WriteString(f.formatErrorMessage(spellErr))

			if f.options.ShowSuggestions && len(spellErr.Suggestions) > 0 && i == 0 {
				b.WriteString("\n")
				b.WriteString(f.formatSuggestions(spellErr.Suggestions))
			}
		} else {
			b.WriteString(err.Error())
		}
	}

	b.WriteString("\n")

	return b.String()
}
