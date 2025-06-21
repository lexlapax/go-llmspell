// ABOUTME: Tests for error formatter, covering color output, formatting modes, and error display.
// ABOUTME: Ensures proper formatting of errors, context, suggestions, and stack traces.

package errors

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatterOptions(t *testing.T) {
	t.Run("default_options", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		
		assert.False(t, opts.ShowStackTrace)
		assert.True(t, opts.ShowContext)
		assert.True(t, opts.ShowSuggestions)
		assert.False(t, opts.ShowTimestamp)
		assert.True(t, opts.ColorOutput)
		assert.Equal(t, 10, opts.MaxStackFrames)
		assert.Equal(t, 5, opts.MaxContextItems)
		assert.Equal(t, 2, opts.IndentLevel)
		assert.False(t, opts.DebugMode)
	})
	
	t.Run("debug_options", func(t *testing.T) {
		opts := DebugFormatterOptions()
		
		assert.True(t, opts.ShowStackTrace)
		assert.True(t, opts.ShowContext)
		assert.True(t, opts.ShowSuggestions)
		assert.True(t, opts.ShowTimestamp)
		assert.True(t, opts.ColorOutput)
		assert.Equal(t, 20, opts.MaxStackFrames)
		assert.True(t, opts.DebugMode)
	})
}

func TestFormatter_Basic(t *testing.T) {
	t.Run("new_formatter", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		f := NewFormatter(opts)
		
		assert.NotNil(t, f)
		// NewFormatter may modify ColorOutput based on terminal detection
		// so we check individual fields instead
		assert.Equal(t, opts.ShowStackTrace, f.options.ShowStackTrace)
		assert.Equal(t, opts.ShowContext, f.options.ShowContext)
		assert.Equal(t, opts.ShowSuggestions, f.options.ShowSuggestions)
		assert.NotNil(t, f.writer)
	})
	
	t.Run("format_nil_error", func(t *testing.T) {
		f := NewFormatter(DefaultFormatterOptions())
		
		result := f.Format(nil)
		assert.Equal(t, "", result)
	})
	
	t.Run("set_writer", func(t *testing.T) {
		f := NewFormatter(DefaultFormatterOptions())
		var buf bytes.Buffer
		
		f.SetWriter(&buf)
		f.Print(New(CategoryConfig, "test"))
		
		assert.NotEmpty(t, buf.String())
	})
}

func TestFormatter_SpellError(t *testing.T) {
	t.Run("simple_error", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false // Disable colors for easier testing
		f := NewFormatter(opts)
		
		err := ConfigError("invalid configuration")
		result := f.Format(err)
		
		assert.Contains(t, result, "Config Error")
		assert.Contains(t, result, "invalid configuration")
		assert.Contains(t, result, "Check your configuration")
		assert.NotContains(t, result, "Stack trace") // Not in debug mode
	})
	
	t.Run("error_with_context", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := ScriptError("syntax error").
			WithContext("file", "test.lua").
			WithContext("line", 42).
			WithContext("column", 10)
		
		result := f.Format(err)
		
		assert.Contains(t, result, "Script Error")
		assert.Contains(t, result, "syntax error")
		assert.Contains(t, result, "Context:")
		assert.Contains(t, result, "file: \"test.lua\"")
		assert.Contains(t, result, "line: 42")
		assert.Contains(t, result, "column: 10")
	})
	
	t.Run("error_with_suggestions", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := NetworkError("connection timeout").
			WithSuggestion("Check your internet connection").
			WithSuggestion("Verify proxy settings")
		
		result := f.Format(err)
		
		assert.Contains(t, result, "Network Error")
		assert.Contains(t, result, "connection timeout")
		assert.Contains(t, result, "Suggestions:")
		assert.Contains(t, result, "Check your internet connection")
		assert.Contains(t, result, "Verify proxy settings")
	})
	
	t.Run("error_with_cause", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		cause := errors.New("file not found")
		err := Wrap(cause, CategoryIO, "failed to read config")
		
		result := f.Format(err)
		
		assert.Contains(t, result, "Io Error")
		assert.Contains(t, result, "failed to read config")
		assert.Contains(t, result, "file not found")
		assert.Contains(t, result, "└─")
	})
}

func TestFormatter_Debug(t *testing.T) {
	t.Run("debug_mode", func(t *testing.T) {
		opts := DebugFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := EngineError("lua", "initialization failed").
			WithContext("version", "5.4")
		
		result := f.Format(err)
		
		assert.Contains(t, result, "Engine Error")
		assert.Contains(t, result, "initialization failed")
		assert.Contains(t, result, "Stack trace:")
		assert.Contains(t, result, "Debug info:")
		assert.Contains(t, result, "Category: engine")
		assert.Contains(t, result, "Exit code:")
	})
	
	t.Run("stack_trace_formatting", func(t *testing.T) {
		opts := DebugFormatterOptions()
		opts.ColorOutput = false
		opts.MaxStackFrames = 3
		f := NewFormatter(opts)
		
		err := New(CategoryScript, "test error")
		result := f.Format(err)
		
		assert.Contains(t, result, "Stack trace:")
		// Should show function names and file:line
		assert.Contains(t, result, ".go:")
		assert.Contains(t, result, "1.")
	})
}

func TestFormatter_GenericError(t *testing.T) {
	t.Run("format_generic_error", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := errors.New("generic error message")
		result := f.Format(err)
		
		assert.Contains(t, result, "Error:")
		assert.Contains(t, result, "generic error message")
		assert.NotContains(t, result, "Context:")
		assert.NotContains(t, result, "Suggestions:")
	})
}

func TestFormatter_Chain(t *testing.T) {
	t.Run("format_error_chain", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		chain := NewChain()
		chain.Add(ConfigError("config error"))
		chain.Add(ScriptError("script error"))
		chain.Add(ValidationError("validation error"))
		
		result := f.FormatChain(chain)
		
		assert.Contains(t, result, "Multiple errors (3)")
		assert.Contains(t, result, "▸ 1.")
		assert.Contains(t, result, "config error")
		assert.Contains(t, result, "▸ 2.")
		assert.Contains(t, result, "script error")
		assert.Contains(t, result, "▸ 3.")
		assert.Contains(t, result, "validation error")
	})
	
	t.Run("format_empty_chain", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		f := NewFormatter(opts)
		
		chain := NewChain()
		result := f.FormatChain(chain)
		
		assert.Equal(t, "", result)
	})
	
	t.Run("chain_with_suggestions", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		chain := NewChain()
		chain.Add(ConfigError("first error").WithSuggestion("Fix config"))
		chain.Add(ScriptError("second error"))
		
		result := f.FormatChain(chain)
		
		// Only first error suggestions shown
		assert.Contains(t, result, "Fix config")
	})
}

func TestFormatter_Colors(t *testing.T) {
	t.Run("color_output_enabled", func(t *testing.T) {
		// Set NO_COLOR to empty to ensure colors work
		oldNoColor := os.Getenv("NO_COLOR")
		_ = os.Setenv("NO_COLOR", "")
		defer func() {
			_ = os.Setenv("NO_COLOR", oldNoColor)
		}()
		
		opts := DefaultFormatterOptions()
		opts.ColorOutput = true
		// Create formatter with struct literal to bypass terminal detection
		f := &Formatter{
			options: opts,
			writer:  os.Stderr,
		}
		
		err := SecurityError("access denied")
		result := f.Format(err)
		
		// Should contain ANSI color codes
		assert.Contains(t, result, "\033[")
		assert.Contains(t, result, "\033[0m") // Reset
	})
	
	t.Run("color_output_disabled", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := SecurityError("access denied")
		result := f.Format(err)
		
		// Should not contain ANSI color codes
		assert.NotContains(t, result, "\033[")
	})
}

func TestFormatter_Timestamp(t *testing.T) {
	t.Run("timestamp_enabled", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ShowTimestamp = true
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := New(CategoryConfig, "test")
		result := f.Format(err)
		
		// Should contain timestamp in format [HH:MM:SS]
		assert.Contains(t, result, "[")
		assert.Contains(t, result, ":")
		assert.Contains(t, result, "]")
	})
	
	t.Run("timestamp_disabled", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ShowTimestamp = false
		f := NewFormatter(opts)
		
		err := New(CategoryConfig, "test")
		result := f.Format(err)
		
		// Should not have timestamp brackets at start
		assert.NotRegexp(t, `^\[.*\]`, result)
	})
}

func TestFormatter_Context(t *testing.T) {
	t.Run("max_context_items", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		opts.MaxContextItems = 3
		f := NewFormatter(opts)
		
		err := New(CategoryConfig, "test").
			WithContext("item1", "value1").
			WithContext("item2", "value2").
			WithContext("item3", "value3").
			WithContext("item4", "value4").
			WithContext("item5", "value5")
		
		result := f.Format(err)
		
		assert.Contains(t, result, "Context:")
		// Should show message about remaining items
		assert.Contains(t, result, "... and 2 more")
	})
	
	t.Run("context_value_formatting", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := New(CategoryConfig, "test").
			WithContext("string", "hello world").
			WithContext("number", 42).
			WithContext("bool", true).
			WithContext("long_string", strings.Repeat("a", 100)).
			WithContext("error", errors.New("nested error"))
		
		result := f.Format(err)
		
		assert.Contains(t, result, `string: "hello world"`)
		assert.Contains(t, result, "number: 42")
		assert.Contains(t, result, "bool: true")
		assert.Contains(t, result, "...") // Long string truncated
		assert.Contains(t, result, "nested error")
	})
}

func TestFormatter_Icons(t *testing.T) {
	t.Run("error_category_icons", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		testCases := []struct {
			err  *SpellError
			icon string
		}{
			{UsageError("test"), "⚠"},
			{ConfigError("test"), "⚙"},
			{ScriptError("test"), "📜"},
			{EngineError("lua", "test"), "⚡"},
			{SecurityError("test"), "🔒"},
			{NetworkError("test"), "🌐"},
			{TimeoutError("test"), "⏱"},
			{ResourceError("test"), "💾"},
			{ValidationError("test"), "✓"},
			{DependencyError("test"), "📦"},
			{IOError("test"), "💾"},
			{InterruptedError(), "⛔"},
		}
		
		for _, tc := range testCases {
			result := f.Format(tc.err)
			assert.Contains(t, result, tc.icon, "Icon for %s", tc.err.Category)
		}
	})
}

func TestFormatter_Print(t *testing.T) {
	t.Run("print_to_writer", func(t *testing.T) {
		var buf bytes.Buffer
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		f.SetWriter(&buf)
		
		err := ConfigError("test error")
		f.Print(err)
		
		output := buf.String()
		assert.Contains(t, output, "Config Error")
		assert.Contains(t, output, "test error")
	})
	
	t.Run("print_nil_error", func(t *testing.T) {
		var buf bytes.Buffer
		f := NewFormatter(DefaultFormatterOptions())
		f.SetWriter(&buf)
		
		f.Print(nil)
		
		assert.Empty(t, buf.String())
	})
}

// Benchmarks
func BenchmarkFormatter_SimpleError(b *testing.B) {
	opts := DefaultFormatterOptions()
	opts.ColorOutput = false
	f := NewFormatter(opts)
	err := New(CategoryConfig, "benchmark error")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = f.Format(err)
	}
}

func BenchmarkFormatter_ComplexError(b *testing.B) {
	opts := DebugFormatterOptions()
	opts.ColorOutput = false
	f := NewFormatter(opts)
	
	err := New(CategoryConfig, "benchmark error").
		WithContext("file", "config.yaml").
		WithContext("line", 42).
		WithContext("section", "database").
		WithSuggestion("Check syntax").
		WithSuggestion("Validate schema")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = f.Format(err)
	}
}

func BenchmarkFormatter_Chain(b *testing.B) {
	opts := DefaultFormatterOptions()
	f := NewFormatter(opts)
	
	chain := NewChain()
	for i := 0; i < 10; i++ {
		chain.Add(Newf(CategoryConfig, "error %d", i))
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = f.FormatChain(chain)
	}
}

// Test formatting edge cases
func TestFormatter_EdgeCases(t *testing.T) {
	t.Run("nil_formatter", func(t *testing.T) {
		// Ensure NewFormatter handles edge cases
		f := NewFormatter(FormatterOptions{})
		assert.NotNil(t, f)
		assert.NotNil(t, f.writer)
	})
	
	t.Run("very_long_message", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		longMessage := strings.Repeat("error ", 100)
		err := New(CategoryConfig, longMessage)
		
		result := f.Format(err)
		assert.Contains(t, result, longMessage)
	})
	
	t.Run("special_characters", func(t *testing.T) {
		opts := DefaultFormatterOptions()
		opts.ColorOutput = false
		f := NewFormatter(opts)
		
		err := New(CategoryConfig, "error with special chars: \n\t\r\"'<>&")
		result := f.Format(err)
		
		assert.Contains(t, result, "error with special chars")
	})
}

// Example of custom formatter usage
func ExampleFormatter() {
	// Create custom formatter options
	opts := FormatterOptions{
		ShowStackTrace:  false,
		ShowContext:     true,
		ShowSuggestions: true,
		ColorOutput:     false, // Disable for example output
		IndentLevel:     2,
	}
	
	formatter := NewFormatter(opts)
	
	// Create an error with rich information
	err := ConfigError("invalid YAML syntax").
		WithContext("file", "config.yaml").
		WithContext("line", 10).
		WithSuggestion("Check for proper indentation").
		WithSuggestion("Validate YAML syntax online")
	
	// Format and print
	formatter.Print(err)
}