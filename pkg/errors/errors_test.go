// ABOUTME: Tests for the error handling package, covering error types, categories, and context.
// ABOUTME: Ensures proper error wrapping, unwrapping, and exit code handling.

package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpellError_Basic(t *testing.T) {
	t.Run("create_new_error", func(t *testing.T) {
		err := New(CategoryConfig, "configuration is invalid")

		assert.NotNil(t, err)
		assert.Equal(t, CategoryConfig, err.Category)
		assert.Equal(t, "configuration is invalid", err.Message)
		assert.Equal(t, "configuration is invalid", err.Error())
		assert.Nil(t, err.Cause)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("create_formatted_error", func(t *testing.T) {
		err := Newf(CategoryScript, "script failed at line %d", 42)

		assert.Equal(t, CategoryScript, err.Category)
		assert.Equal(t, "script failed at line 42", err.Message)
		assert.Equal(t, "script failed at line 42", err.Error())
	})

	t.Run("wrap_error", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Wrap(cause, CategoryEngine, "engine failed")

		assert.Equal(t, CategoryEngine, err.Category)
		assert.Equal(t, "engine failed", err.Message)
		assert.Equal(t, "engine failed: underlying error", err.Error())
		assert.Equal(t, cause, err.Cause)
		assert.True(t, errors.Is(err, cause))
	})

	t.Run("wrap_nil_error", func(t *testing.T) {
		err := Wrap(nil, CategoryEngine, "engine failed")
		assert.Nil(t, err)
	})

	t.Run("wrapf_error", func(t *testing.T) {
		cause := errors.New("file not found")
		err := Wrapf(cause, CategoryIO, "failed to read %s", "config.yaml")

		assert.Equal(t, CategoryIO, err.Category)
		assert.Equal(t, "failed to read config.yaml", err.Message)
		assert.Contains(t, err.Error(), "file not found")
	})
}

func TestSpellError_Context(t *testing.T) {
	t.Run("add_context", func(t *testing.T) {
		err := New(CategoryScript, "script error").
			WithContext("file", "test.lua").
			WithContext("line", 42).
			WithContext("engine", "lua")

		assert.Len(t, err.Context, 3)
		assert.Equal(t, "test.lua", err.Context["file"])
		assert.Equal(t, 42, err.Context["line"])
		assert.Equal(t, "lua", err.Context["engine"])
	})

	t.Run("add_suggestions", func(t *testing.T) {
		err := New(CategoryConfig, "config error").
			WithSuggestion("Check the syntax").
			WithSuggestion("Run validation").
			WithSuggestion("See documentation")

		assert.Len(t, err.Suggestions, 3)
		assert.Equal(t, "Check the syntax", err.Suggestions[0])
		assert.Equal(t, "Run validation", err.Suggestions[1])
		assert.Equal(t, "See documentation", err.Suggestions[2])
	})
}

func TestSpellError_ExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		code     int
		expected int
	}{
		{"usage_error", CategoryUsage, 0, ExitUsageError},
		{"config_error", CategoryConfig, 0, ExitConfigError},
		{"script_error", CategoryScript, 0, ExitScriptError},
		{"engine_error", CategoryEngine, 0, ExitEngineError},
		{"security_error", CategorySecurity, 0, ExitSecurityError},
		{"network_error", CategoryNetwork, 0, ExitNetworkError},
		{"timeout_error", CategoryTimeout, 0, ExitTimeoutError},
		{"resource_error", CategoryResource, 0, ExitResourceError},
		{"validation_error", CategoryValidation, 0, ExitValidationError},
		{"dependency_error", CategoryDependency, 0, ExitDependencyError},
		{"io_error", CategoryIO, 0, ExitIOError},
		{"interrupted_error", CategoryInterrupted, 0, ExitInterrupted},
		{"unknown_error", CategoryUnknown, 0, ExitGeneralError},
		{"custom_code", CategoryConfig, 99, 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &SpellError{
				Category: tt.category,
				Code:     tt.code,
				Message:  "test error",
			}

			assert.Equal(t, tt.expected, err.ExitCode())
		})
	}
}

func TestSpellError_Is(t *testing.T) {
	t.Run("is_same_error", func(t *testing.T) {
		err1 := &SpellError{Category: CategoryConfig, Code: 123, Message: "error"}
		err2 := &SpellError{Category: CategoryConfig, Code: 123, Message: "different message"}

		assert.True(t, err1.Is(err2))
	})

	t.Run("is_different_category", func(t *testing.T) {
		err1 := &SpellError{Category: CategoryConfig, Code: 123}
		err2 := &SpellError{Category: CategoryScript, Code: 123}

		assert.False(t, err1.Is(err2))
	})

	t.Run("is_different_code", func(t *testing.T) {
		err1 := &SpellError{Category: CategoryConfig, Code: 123}
		err2 := &SpellError{Category: CategoryConfig, Code: 456}

		assert.False(t, err1.Is(err2))
	})

	t.Run("is_wrapped_error", func(t *testing.T) {
		cause := errors.New("cause")
		err := Wrap(cause, CategoryEngine, "wrapped")

		assert.True(t, errors.Is(err, cause))
	})

	t.Run("is_nil_target", func(t *testing.T) {
		err := New(CategoryConfig, "error")
		assert.False(t, err.Is(nil))
	})
}

func TestSpellError_Unwrap(t *testing.T) {
	t.Run("unwrap_wrapped_error", func(t *testing.T) {
		cause := errors.New("cause")
		err := Wrap(cause, CategoryEngine, "wrapped")

		assert.Equal(t, cause, err.Unwrap())
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("unwrap_no_cause", func(t *testing.T) {
		err := New(CategoryConfig, "error")
		assert.Nil(t, err.Unwrap())
	})
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("usage_error", func(t *testing.T) {
		err := UsageError("invalid flag")
		assert.Equal(t, CategoryUsage, err.Category)
		assert.Contains(t, err.Suggestions, "Use 'llmspell --help' for usage information")
	})

	t.Run("config_error", func(t *testing.T) {
		err := ConfigError("invalid syntax")
		assert.Equal(t, CategoryConfig, err.Category)
		assert.Contains(t, err.Suggestions[0], "Check your configuration")
		assert.Contains(t, err.Suggestions[1], "validate")
	})

	t.Run("script_error", func(t *testing.T) {
		err := ScriptError("syntax error")
		assert.Equal(t, CategoryScript, err.Category)
		assert.Contains(t, err.Suggestions[0], "Check the script syntax")
	})

	t.Run("engine_error", func(t *testing.T) {
		err := EngineError("lua", "failed to initialize")
		assert.Equal(t, CategoryEngine, err.Category)
		assert.Equal(t, "lua", err.Context["engine"])
	})

	t.Run("security_error", func(t *testing.T) {
		err := SecurityError("access denied")
		assert.Equal(t, CategorySecurity, err.Category)
		assert.Contains(t, err.Suggestions[0], "security profile")
	})

	t.Run("network_error", func(t *testing.T) {
		err := NetworkError("connection failed")
		assert.Equal(t, CategoryNetwork, err.Category)
		assert.Contains(t, err.Suggestions[0], "network connection")
	})

	t.Run("timeout_error", func(t *testing.T) {
		err := TimeoutError("operation timed out")
		assert.Equal(t, CategoryTimeout, err.Category)
		assert.Contains(t, err.Suggestions[0], "timeout")
	})

	t.Run("resource_error", func(t *testing.T) {
		err := ResourceError("out of memory")
		assert.Equal(t, CategoryResource, err.Category)
		assert.Contains(t, err.Suggestions[0], "memory usage")
	})

	t.Run("validation_error", func(t *testing.T) {
		err := ValidationError("invalid format")
		assert.Equal(t, CategoryValidation, err.Category)
	})

	t.Run("dependency_error", func(t *testing.T) {
		err := DependencyError("missing dependency")
		assert.Equal(t, CategoryDependency, err.Category)
		assert.Contains(t, err.Suggestions[0], "dependencies")
	})

	t.Run("io_error", func(t *testing.T) {
		err := IOError("file not found")
		assert.Equal(t, CategoryIO, err.Category)
		assert.Contains(t, err.Suggestions[0], "permissions")
	})

	t.Run("interrupted_error", func(t *testing.T) {
		err := InterruptedError()
		assert.Equal(t, CategoryInterrupted, err.Category)
		assert.Contains(t, err.Message, "interrupted")
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("is_spell_error", func(t *testing.T) {
		spellErr := New(CategoryConfig, "error")
		normalErr := errors.New("normal error")

		assert.True(t, IsSpellError(spellErr))
		assert.False(t, IsSpellError(normalErr))
		assert.False(t, IsSpellError(nil))
	})

	t.Run("get_category", func(t *testing.T) {
		spellErr := New(CategoryScript, "error")
		normalErr := errors.New("normal error")

		assert.Equal(t, CategoryScript, GetCategory(spellErr))
		assert.Equal(t, CategoryUnknown, GetCategory(normalErr))
		assert.Equal(t, CategoryUnknown, GetCategory(nil))
	})

	t.Run("get_exit_code", func(t *testing.T) {
		spellErr := New(CategoryConfig, "error")
		normalErr := errors.New("normal error")

		assert.Equal(t, ExitConfigError, GetExitCode(spellErr))
		assert.Equal(t, ExitGeneralError, GetExitCode(normalErr))
		assert.Equal(t, ExitSuccess, GetExitCode(nil))
	})

	t.Run("get_suggestions", func(t *testing.T) {
		spellErr := New(CategoryConfig, "error").
			WithSuggestion("suggestion 1").
			WithSuggestion("suggestion 2")
		normalErr := errors.New("normal error")

		suggestions := GetSuggestions(spellErr)
		assert.Len(t, suggestions, 2)
		assert.Equal(t, "suggestion 1", suggestions[0])

		assert.Nil(t, GetSuggestions(normalErr))
		assert.Nil(t, GetSuggestions(nil))
	})

	t.Run("get_context", func(t *testing.T) {
		spellErr := New(CategoryConfig, "error").
			WithContext("key", "value")
		normalErr := errors.New("normal error")

		context := GetContext(spellErr)
		assert.Len(t, context, 1)
		assert.Equal(t, "value", context["key"])

		assert.Nil(t, GetContext(normalErr))
		assert.Nil(t, GetContext(nil))
	})
}

func TestErrorChain(t *testing.T) {
	t.Run("new_chain", func(t *testing.T) {
		chain := NewChain()
		assert.NotNil(t, chain)
		assert.False(t, chain.HasErrors())
		assert.Len(t, chain.Errors(), 0)
		assert.Nil(t, chain.First())
		assert.Equal(t, "", chain.Error())
	})

	t.Run("add_errors", func(t *testing.T) {
		chain := NewChain()

		err1 := New(CategoryConfig, "error 1")
		err2 := errors.New("error 2")
		err3 := New(CategoryScript, "error 3")

		chain.Add(err1)
		chain.Add(err2)
		chain.Add(err3)
		chain.Add(nil) // Should be ignored

		assert.True(t, chain.HasErrors())
		assert.Len(t, chain.Errors(), 3)
		assert.Equal(t, err1, chain.First())

		errStr := chain.Error()
		assert.Contains(t, errStr, "error 1")
		assert.Contains(t, errStr, "error 2")
		assert.Contains(t, errStr, "error 3")
	})

	t.Run("add_formatted_error", func(t *testing.T) {
		chain := NewChain()
		chain.Addf(CategoryIO, "failed to read %s", "file.txt")

		assert.True(t, chain.HasErrors())
		assert.Contains(t, chain.Error(), "failed to read file.txt")
	})

	t.Run("merge_chains", func(t *testing.T) {
		chain1 := NewChain()
		chain1.Add(New(CategoryConfig, "error 1"))
		chain1.Add(New(CategoryScript, "error 2"))

		chain2 := NewChain()
		chain2.Add(New(CategoryEngine, "error 3"))
		chain2.Add(New(CategoryIO, "error 4"))

		chain1.Merge(chain2)

		assert.Len(t, chain1.Errors(), 4)
		assert.Len(t, chain2.Errors(), 2) // Original unchanged

		// Test merging nil
		chain1.Merge(nil)
		assert.Len(t, chain1.Errors(), 4)
	})
}

func TestStackTrace(t *testing.T) {
	t.Run("capture_stack_trace", func(t *testing.T) {
		err := New(CategoryConfig, "test error")

		if assert.NotEmpty(t, err.StackTrace) {
			// Check that we have a valid stack trace
			frame := err.StackTrace[0]
			assert.NotEmpty(t, frame.Function)
			assert.NotEmpty(t, frame.File)
			assert.Greater(t, frame.Line, 0)

			// Stack should contain errors_test.go file
			foundTestFile := false
			for _, f := range err.StackTrace {
				if strings.Contains(f.File, "errors_test.go") {
					foundTestFile = true
					break
				}
			}
			// If not found in test file, at least check we have errors.go
			if !foundTestFile {
				for _, f := range err.StackTrace {
					if strings.Contains(f.File, "errors.go") && strings.Contains(f.Function, "New") {
						foundTestFile = true
						break
					}
				}
			}
			assert.True(t, foundTestFile, "Stack trace should contain relevant source files")
		}
	})

	t.Run("wrap_preserves_new_stack", func(t *testing.T) {
		cause := errors.New("cause")
		wrapped := Wrap(cause, CategoryEngine, "wrapped")

		// Should have new stack trace, not from cause
		assert.NotEmpty(t, wrapped.StackTrace)
		assert.NotEmpty(t, wrapped.StackTrace[0].Function)
	})
}

func TestWrapSpellError(t *testing.T) {
	t.Run("wrap_spell_error_preserves_properties", func(t *testing.T) {
		original := New(CategoryConfig, "original").
			WithContext("key", "value").
			WithSuggestion("suggestion")
		original.Code = 123

		wrapped := Wrap(original, CategoryScript, "wrapped")

		assert.Equal(t, CategoryScript, wrapped.Category)
		assert.Equal(t, 123, wrapped.Code)
		assert.Equal(t, "wrapped", wrapped.Message)
		assert.Equal(t, original, wrapped.Cause)
		assert.Equal(t, "value", wrapped.Context["key"])
		assert.Contains(t, wrapped.Suggestions, "suggestion")
	})
}

// Benchmarks
func BenchmarkNewError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(CategoryConfig, "test error")
	}
}

func BenchmarkWrapError(b *testing.B) {
	cause := errors.New("cause")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Wrap(cause, CategoryEngine, "wrapped")
	}
}

func BenchmarkErrorWithContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(CategoryConfig, "error").
			WithContext("file", "test.lua").
			WithContext("line", 42).
			WithSuggestion("Check syntax")
	}
}

func BenchmarkErrorChain(b *testing.B) {
	errors := []error{
		New(CategoryConfig, "error 1"),
		New(CategoryScript, "error 2"),
		New(CategoryEngine, "error 3"),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		chain := NewChain()
		for _, err := range errors {
			chain.Add(err)
		}
		_ = chain.Error()
	}
}

// Example error for testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestErrorCompatibility(t *testing.T) {
	t.Run("wrap_custom_error", func(t *testing.T) {
		custom := &customError{msg: "custom error"}
		wrapped := Wrap(custom, CategoryEngine, "wrapped custom")

		assert.Equal(t, custom, wrapped.Cause)
		assert.Contains(t, wrapped.Error(), "custom error")

		// Should work with errors.As
		var target *customError
		assert.True(t, errors.As(wrapped, &target))
		assert.Equal(t, custom, target)
	})

	t.Run("error_is_chain", func(t *testing.T) {
		err1 := errors.New("base")
		err2 := fmt.Errorf("wrapped: %w", err1)
		err3 := Wrap(err2, CategoryConfig, "spell wrapped")

		assert.True(t, errors.Is(err3, err1))
		assert.True(t, errors.Is(err3, err2))
	})
}

// Test that exit codes are unique and within expected range
func TestExitCodeUniqueness(t *testing.T) {
	exitCodes := map[int]string{
		ExitSuccess:         "Success",
		ExitGeneralError:    "GeneralError",
		ExitUsageError:      "UsageError",
		ExitConfigError:     "ConfigError",
		ExitScriptError:     "ScriptError",
		ExitEngineError:     "EngineError",
		ExitSecurityError:   "SecurityError",
		ExitNetworkError:    "NetworkError",
		ExitTimeoutError:    "TimeoutError",
		ExitResourceError:   "ResourceError",
		ExitValidationError: "ValidationError",
		ExitDependencyError: "DependencyError",
		ExitIOError:         "IOError",
		ExitInterrupted:     "Interrupted",
	}

	// Check uniqueness
	seen := make(map[int]bool)
	for code, name := range exitCodes {
		if seen[code] {
			t.Errorf("Duplicate exit code %d for %s", code, name)
		}
		seen[code] = true
	}

	// Check reasonable range (0-255 for Unix exit codes)
	for code, name := range exitCodes {
		if code < 0 || code > 255 {
			t.Errorf("Exit code %d for %s is outside valid range (0-255)", code, name)
		}
	}
}

// Test thread safety of error creation
func TestErrorThreadSafety(t *testing.T) {
	// Run multiple goroutines creating errors
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			err := Newf(CategoryEngine, "error from goroutine %d", id).
				WithContext("goroutine", id).
				WithSuggestion(fmt.Sprintf("suggestion %d", id))

			// Verify error properties
			assert.Contains(t, err.Error(), fmt.Sprintf("goroutine %d", id))
			assert.Equal(t, id, err.Context["goroutine"])

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// Test nil safety
func TestNilSafety(t *testing.T) {
	t.Run("nil_spell_error_methods", func(t *testing.T) {
		var err *SpellError

		// These should not panic
		assert.Equal(t, "", err.Error())
		assert.Nil(t, err.Unwrap())
		assert.False(t, err.Is(New(CategoryConfig, "test")))
		assert.Equal(t, 0, err.ExitCode())
	})

	t.Run("nil_chain_methods", func(t *testing.T) {
		var chain *Chain

		// These should not panic
		assert.False(t, chain.HasErrors())
		assert.Nil(t, chain.Errors())
		assert.Nil(t, chain.First())
		assert.Equal(t, "", chain.Error())
	})
}

// Test error message formatting
func TestErrorMessageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "simple_error",
			err:      New(CategoryConfig, "simple error"),
			expected: "simple error",
		},
		{
			name:     "error_with_cause",
			err:      Wrap(errors.New("cause"), CategoryEngine, "wrapper"),
			expected: "wrapper: cause",
		},
		{
			name: "error_chain",
			err: func() error {
				chain := NewChain()
				chain.Add(New(CategoryConfig, "error 1"))
				chain.Add(New(CategoryScript, "error 2"))
				return chain
			}(),
			expected: "error 1; error 2",
		},
		{
			name:     "formatted_error",
			err:      Newf(CategoryIO, "failed to open %s: %v", "file.txt", "permission denied"),
			expected: "failed to open file.txt: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// Test category string representation
func TestCategoryString(t *testing.T) {
	categories := []ErrorCategory{
		CategoryUnknown,
		CategoryUsage,
		CategoryConfig,
		CategoryScript,
		CategoryEngine,
		CategorySecurity,
		CategoryNetwork,
		CategoryTimeout,
		CategoryResource,
		CategoryValidation,
		CategoryDependency,
		CategoryIO,
		CategoryInterrupted,
	}

	for _, cat := range categories {
		// Category should be a non-empty string
		assert.NotEmpty(t, string(cat))
		// Should be lowercase
		assert.Equal(t, strings.ToLower(string(cat)), string(cat))
	}
}
