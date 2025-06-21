// ABOUTME: Tests for error handling integration with configuration and global error handler.
// ABOUTME: Ensures proper error handling, metrics recording, and debug mode functionality.

package errors

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandler_Basic(t *testing.T) {
	t.Run("initialize_handler", func(t *testing.T) {
		InitializeErrorHandler(false, false)
		handler := GetErrorHandler()

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.formatter)
		assert.NotNil(t, handler.metrics)
		assert.False(t, handler.debugMode)
	})

	t.Run("initialize_debug_mode", func(t *testing.T) {
		InitializeErrorHandler(true, true)
		handler := GetErrorHandler()

		assert.True(t, handler.debugMode)
		assert.True(t, handler.formatter.options.ShowStackTrace)
		assert.True(t, handler.formatter.options.DebugMode)
	})

	t.Run("handle_error", func(t *testing.T) {
		// Reset metrics
		metrics := GetMetrics()
		metrics.Reset()

		InitializeErrorHandler(false, false)
		handler := GetErrorHandler()

		// Capture output
		var buf bytes.Buffer
		handler.SetOutput(&buf)

		// Handle error
		err := ConfigError("test error")
		handler.Handle(err)

		// Check output
		output := buf.String()
		assert.Contains(t, output, "test error")
		assert.Contains(t, output, "Config")

		// Check metrics
		assert.Equal(t, int64(1), metrics.GetTotalErrors())
		assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryConfig))
	})

	t.Run("handle_nil_error", func(t *testing.T) {
		handler := GetErrorHandler()
		metrics := handler.GetMetrics()
		metrics.Reset()

		var buf bytes.Buffer
		handler.SetOutput(&buf)

		handler.Handle(nil)

		assert.Empty(t, buf.String())
		assert.Equal(t, int64(0), metrics.GetTotalErrors())
	})
}

func TestErrorHandler_DebugMode(t *testing.T) {
	t.Run("set_debug_mode", func(t *testing.T) {
		InitializeErrorHandler(false, false)
		handler := GetErrorHandler()

		assert.False(t, handler.debugMode)

		// Enable debug mode
		handler.SetDebugMode(true)

		assert.True(t, handler.debugMode)
		assert.True(t, handler.formatter.options.ShowStackTrace)
		assert.True(t, handler.formatter.options.DebugMode)

		// Disable debug mode
		handler.SetDebugMode(false)

		assert.False(t, handler.debugMode)
		assert.False(t, handler.formatter.options.ShowStackTrace)
		assert.False(t, handler.formatter.options.DebugMode)
	})

	t.Run("debug_output", func(t *testing.T) {
		InitializeErrorHandler(true, false)
		handler := GetErrorHandler()

		var buf bytes.Buffer
		handler.SetOutput(&buf)

		err := New(CategoryScript, "debug test")
		handler.Handle(err)

		output := buf.String()
		assert.Contains(t, output, "Stack trace:")
		assert.Contains(t, output, "Debug info:")
	})
}

func TestErrorHandler_Formatting(t *testing.T) {
	t.Run("format_error", func(t *testing.T) {
		handler := GetErrorHandler()

		err := ScriptError("format test").
			WithContext("file", "test.lua").
			WithSuggestion("Check syntax")

		formatted := handler.FormatError(err)

		assert.Contains(t, formatted, "format test")
		assert.Contains(t, formatted, "test.lua")
		assert.Contains(t, formatted, "Check syntax")
	})

	t.Run("format_chain", func(t *testing.T) {
		handler := GetErrorHandler()

		chain := NewChain()
		chain.Add(ConfigError("error 1"))
		chain.Add(ScriptError("error 2"))

		formatted := handler.FormatError(chain)

		assert.Contains(t, formatted, "error 1")
		assert.Contains(t, formatted, "error 2")
	})
}

func TestGlobalHelpers(t *testing.T) {
	t.Run("handle_error", func(t *testing.T) {
		metrics := GetMetrics()
		metrics.Reset()

		var buf bytes.Buffer
		GetErrorHandler().SetOutput(&buf)

		err := NetworkError("connection failed")
		HandleError(err)

		assert.Contains(t, buf.String(), "connection failed")
		assert.Equal(t, int64(1), metrics.GetTotalErrors())
	})

	t.Run("debug_error", func(t *testing.T) {
		// Set non-debug mode
		InitializeErrorHandler(false, false)

		var buf bytes.Buffer
		GetErrorHandler().SetOutput(&buf)

		// DebugError should show debug info regardless
		err := ValidationError("validation failed")

		// Capture debug output separately
		debugBuf := bytes.Buffer{}
		opts := DebugFormatterOptions()
		formatter := NewFormatter(opts)
		formatter.SetWriter(&debugBuf)
		formatter.Print(err)

		output := debugBuf.String()
		assert.Contains(t, output, "Stack trace:")
	})
}

func TestMustFunctions(t *testing.T) {
	t.Run("must_no_error", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Must(nil)
		})
	})

	t.Run("must_with_error", func(t *testing.T) {
		err := errors.New("test error")
		assert.Panics(t, func() {
			Must(err)
		})
	})

	t.Run("must_return_no_error", func(t *testing.T) {
		result := MustReturn("value", nil)
		assert.Equal(t, "value", result)
	})

	t.Run("must_return_with_error", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = MustReturn("value", errors.New("error"))
		})
	})
}

func TestRecover(t *testing.T) {
	t.Run("recover_error", func(t *testing.T) {
		var recovered error

		func() {
			defer func() {
				if r := recover(); r != nil {
					// Need to call Recover in the same goroutine as recover()
					switch v := r.(type) {
					case error:
						recovered = Wrap(v, CategoryUnknown, "panic recovered")
					case string:
						recovered = New(CategoryUnknown, v)
					default:
						recovered = Newf(CategoryUnknown, "panic recovered: %v", v)
					}
				}
			}()

			panic(errors.New("panic error"))
		}()

		assert.NotNil(t, recovered)
		assert.Contains(t, recovered.Error(), "panic error")
		assert.Contains(t, recovered.Error(), "panic recovered")
	})

	t.Run("recover_string", func(t *testing.T) {
		var recovered error

		func() {
			defer func() {
				if r := recover(); r != nil {
					switch v := r.(type) {
					case error:
						recovered = Wrap(v, CategoryUnknown, "panic recovered")
					case string:
						recovered = New(CategoryUnknown, v)
					default:
						recovered = Newf(CategoryUnknown, "panic recovered: %v", v)
					}
				}
			}()

			panic("string panic")
		}()

		assert.NotNil(t, recovered)
		assert.Equal(t, "string panic", recovered.Error())
	})

	t.Run("recover_other", func(t *testing.T) {
		var recovered error

		func() {
			defer func() {
				if r := recover(); r != nil {
					switch v := r.(type) {
					case error:
						recovered = Wrap(v, CategoryUnknown, "panic recovered")
					case string:
						recovered = New(CategoryUnknown, v)
					default:
						recovered = Newf(CategoryUnknown, "panic recovered: %v", v)
					}
				}
			}()

			panic(42)
		}()

		assert.NotNil(t, recovered)
		assert.Contains(t, recovered.Error(), "42")
	})

	t.Run("recover_no_panic", func(t *testing.T) {
		// Test Recover when there's no panic
		assert.NotPanics(t, func() {
			err := Recover()
			assert.Nil(t, err)
		})
	})

	t.Run("recover_with_handler", func(t *testing.T) {
		var handled error

		func() {
			defer func() {
				if r := recover(); r != nil {
					var err error
					switch v := r.(type) {
					case error:
						err = Wrap(v, CategoryUnknown, "panic recovered")
					case string:
						err = New(CategoryUnknown, v)
					default:
						err = Newf(CategoryUnknown, "panic recovered: %v", v)
					}
					if err != nil {
						handled = err
					}
				}
			}()

			panic("test panic")
		}()

		assert.NotNil(t, handled)
		assert.Contains(t, handled.Error(), "test panic")
	})
}

func TestChainHandler(t *testing.T) {
	t.Run("create_chain_handler", func(t *testing.T) {
		ch := NewChainHandler()

		assert.NotNil(t, ch)
		assert.NotNil(t, ch.chain)
		assert.NotNil(t, ch.handler)
		assert.False(t, ch.HasErrors())
	})

	t.Run("add_and_handle", func(t *testing.T) {
		metrics := GetMetrics()
		metrics.Reset()

		ch := NewChainHandler()

		var buf bytes.Buffer
		ch.handler.SetOutput(&buf)

		// Add errors
		ch.Add(ConfigError("config error"))
		ch.Add(ScriptError("script error"))
		ch.Add(nil) // Should be ignored

		assert.True(t, ch.HasErrors())

		// Handle all
		ch.Handle()

		output := buf.String()
		assert.Contains(t, output, "Multiple errors")
		assert.Contains(t, output, "config error")
		assert.Contains(t, output, "script error")

		// Check metrics
		assert.Equal(t, int64(1), metrics.GetTotalErrors()) // Chain counted as one
	})

	t.Run("empty_chain", func(t *testing.T) {
		ch := NewChainHandler()

		var buf bytes.Buffer
		ch.handler.SetOutput(&buf)

		ch.Handle()

		assert.Empty(t, buf.String())
		assert.False(t, ch.HasErrors())
	})
}

func TestPrintMetrics(t *testing.T) {
	t.Run("print_metrics", func(t *testing.T) {
		metrics := GetMetrics()
		metrics.Reset()

		// Record some errors
		metrics.RecordError(ConfigError("error1"))
		metrics.RecordError(ScriptError("error2"))

		handler := GetErrorHandler()

		var buf bytes.Buffer
		handler.SetOutput(&buf)

		opts := DefaultMetricsFormatterOptions()
		handler.PrintMetrics(opts)

		output := buf.String()
		assert.Contains(t, output, "not yet implemented")
	})
}

// Example of using the error handler in production
func ExampleErrorHandler() {
	// Initialize with debug mode from config
	debugMode := false  // Would come from config
	colorOutput := true // Would come from config

	InitializeErrorHandler(debugMode, colorOutput)

	// Handle an error
	err := ConfigError("invalid configuration").
		WithContext("file", "config.yaml").
		WithSuggestion("Check YAML syntax")

	HandleError(err)

	// For fatal errors
	// HandleErrorWithExit(err)
}

// Example of chain handling
func ExampleChainHandler() {
	ch := NewChainHandler()

	// Collect errors during processing
	ch.Add(ValidationError("field1 is required"))
	ch.Add(ValidationError("field2 must be positive"))

	// Handle all at once
	if ch.HasErrors() {
		ch.Handle()
		// Or for fatal: ch.HandleWithExit()
	}
}

// Test error context key
func TestErrorContextKey(t *testing.T) {
	t.Run("context_key_type", func(t *testing.T) {
		key := ErrorContextKey{}
		assert.NotNil(t, key)
	})
}

// Benchmark error handling
func BenchmarkErrorHandler(b *testing.B) {
	InitializeErrorHandler(false, false)
	handler := GetErrorHandler()

	// Disable output
	handler.SetOutput(&bytes.Buffer{})

	err := ConfigError("benchmark error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		handler.Handle(err)
	}
}

// Test thread safety of global handler
func TestErrorHandlerThreadSafety(t *testing.T) {
	InitializeErrorHandler(false, false)
	handler := GetErrorHandler()

	// Reset metrics before test
	handler.GetMetrics().Reset()

	// Use io.Discard for thread-safe discarding of output
	handler.SetOutput(io.Discard)

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			err := Newf(CategoryEngine, "error from goroutine %d", id)
			handler.Handle(err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Check metrics
	assert.Equal(t, int64(100), handler.GetMetrics().GetTotalErrors())
}

// Test output redirection
func TestOutputRedirection(t *testing.T) {
	t.Run("set_custom_output", func(t *testing.T) {
		handler := GetErrorHandler()

		// Custom buffer
		var customBuf bytes.Buffer
		handler.SetOutput(&customBuf)

		err := IOError("test error")
		handler.Handle(err)

		assert.Contains(t, customBuf.String(), "test error")
	})
}

// Integration test with all error types
func TestAllErrorTypes(t *testing.T) {
	handler := GetErrorHandler()
	metrics := handler.GetMetrics()
	metrics.Reset()

	var buf bytes.Buffer
	handler.SetOutput(&buf)

	// Test all error constructors
	errors := []error{
		UsageError("usage"),
		ConfigError("config"),
		ScriptError("script"),
		EngineError("lua", "engine"),
		SecurityError("security"),
		NetworkError("network"),
		TimeoutError("timeout"),
		ResourceError("resource"),
		ValidationError("validation"),
		DependencyError("dependency"),
		IOError("io"),
		InterruptedError(),
	}

	for _, err := range errors {
		handler.Handle(err)
		buf.Reset() // Clear buffer for next error
	}

	// Check all were recorded
	assert.Equal(t, int64(len(errors)), metrics.GetTotalErrors())

	// Check specific categories
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryUsage))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryConfig))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryScript))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryEngine))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategorySecurity))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryNetwork))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryTimeout))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryResource))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryValidation))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryDependency))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryIO))
	assert.Equal(t, int64(1), metrics.GetCategoryCount(CategoryInterrupted))
}
