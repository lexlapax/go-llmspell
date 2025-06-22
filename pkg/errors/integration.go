// ABOUTME: This file provides integration between error handling and configuration system.
// ABOUTME: It enables debug mode error reporting and configuration-based error formatting.

package errors

import (
	"fmt"
	"io"
	"os"
)

// ErrorHandler provides centralized error handling with configuration support.
// It combines error formatting, metrics recording, and debug mode handling
// into a single configurable component used throughout the application.
type ErrorHandler struct {
	formatter *Formatter
	metrics   *ErrorMetrics
	debugMode bool
}

// globalErrorHandler is the default error handler instance.
// It's initialized lazily with default settings if not explicitly configured.
var globalErrorHandler *ErrorHandler

// InitializeErrorHandler initializes the global error handler with configuration.
// This should be called early in application startup to configure error
// handling behavior based on runtime settings.
func InitializeErrorHandler(debugMode bool, colorOutput bool) {
	opts := DefaultFormatterOptions()
	opts.DebugMode = debugMode
	opts.ShowStackTrace = debugMode
	opts.ColorOutput = colorOutput

	globalErrorHandler = &ErrorHandler{
		formatter: NewFormatter(opts),
		metrics:   GetMetrics(),
		debugMode: debugMode,
	}
}

// GetErrorHandler returns the global error handler.
// If not yet initialized, it creates one with default settings
// (no debug mode, color output enabled).
func GetErrorHandler() *ErrorHandler {
	if globalErrorHandler == nil {
		// Initialize with defaults if not yet initialized
		InitializeErrorHandler(false, true)
	}
	return globalErrorHandler
}

// Handle processes an error, recording metrics and formatting for display.
// It records the error in metrics for monitoring and prints formatted
// output to the configured writer (default stderr).
func (h *ErrorHandler) Handle(err error) {
	if err == nil || h == nil {
		return
	}

	// Record in metrics
	if h.metrics != nil {
		h.metrics.RecordError(err)
	}

	// Format and print
	if h.formatter != nil {
		h.formatter.Print(err)
	}
}

// HandleWithExit handles an error and exits with appropriate code.
// It processes the error normally then exits the process with the
// error's exit code (determined by error category).
func (h *ErrorHandler) HandleWithExit(err error) {
	if err == nil {
		return
	}

	h.Handle(err)
	os.Exit(GetExitCode(err))
}

// SetOutput sets the output writer for error messages.
// By default errors go to stderr, but this allows redirection
// for testing or alternative output handling.
func (h *ErrorHandler) SetOutput(w io.Writer) {
	h.formatter.SetWriter(w)
}

// SetDebugMode updates the debug mode setting.
// When enabled, errors include stack traces and additional
// diagnostic information for troubleshooting.
func (h *ErrorHandler) SetDebugMode(debug bool) {
	h.debugMode = debug

	// Update formatter options
	opts := h.formatter.options
	opts.DebugMode = debug
	opts.ShowStackTrace = debug
	h.formatter.options = opts
}

// GetMetrics returns the error metrics instance.
// The metrics track error counts by category and other
// statistics for monitoring and alerting.
func (h *ErrorHandler) GetMetrics() *ErrorMetrics {
	return h.metrics
}

// FormatError formats an error without recording metrics.
// This is useful for preview or logging without affecting
// error statistics.
func (h *ErrorHandler) FormatError(err error) string {
	return h.formatter.Format(err)
}

// PrintMetrics prints error metrics to the configured output.
// It formats metrics according to the provided options and
// writes to the same writer as error messages.
func (h *ErrorHandler) PrintMetrics(options MetricsFormatterOptions) {
	if h.metrics == nil {
		return
	}

	formatted := FormatMetrics(h.metrics, options)
	_, _ = fmt.Fprint(h.formatter.writer, formatted)
}

// Quick helper functions for common operations.
// These provide convenient access to global error handler functionality.

// HandleError is a convenience function that handles an error using the global handler.
// It's equivalent to GetErrorHandler().Handle(err).
func HandleError(err error) {
	GetErrorHandler().Handle(err)
}

// HandleErrorWithExit is a convenience function that handles an error and exits.
// It's equivalent to GetErrorHandler().HandleWithExit(err).
func HandleErrorWithExit(err error) {
	GetErrorHandler().HandleWithExit(err)
}

// DebugError handles an error with debug information regardless of global debug mode.
// This is useful for critical errors where you always want full diagnostics
// even in production.
func DebugError(err error) {
	if err == nil {
		return
	}

	// Create a debug formatter
	opts := DebugFormatterOptions()
	debugFormatter := NewFormatter(opts)
	debugFormatter.Print(err)
}

// Must panics if err is not nil, useful for initialization code.
// Use sparingly, only for errors that indicate programming bugs
// or unrecoverable initialization failures.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustReturn returns the value or panics if err is not nil.
// This generic helper is useful for initialization where errors
// indicate programming bugs rather than runtime failures.
func MustReturn[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Recover recovers from a panic and returns it as an error.
// It wraps the panic value in a SpellError with appropriate
// context for error handling and reporting.
func Recover() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return Wrap(v, CategoryUnknown, "panic recovered")
		case string:
			return New(CategoryUnknown, v)
		default:
			return Newf(CategoryUnknown, "panic recovered: %v", v)
		}
	}
	return nil
}

// RecoverWithHandler recovers from panic and handles the error.
// This is useful in defer statements to convert panics into
// handled errors with proper formatting and metrics.
func RecoverWithHandler(handler func(error)) {
	if err := Recover(); err != nil {
		handler(err)
	}
}

// ErrorContextKey is used for storing errors in context.
// This allows passing errors through context.Context for
// middleware and handler patterns.
type ErrorContextKey struct{}

// ChainHandler handles multiple errors as a chain.
// It accumulates errors during an operation and handles
// them together at the end, useful for batch operations.
type ChainHandler struct {
	chain   *Chain
	handler *ErrorHandler
}

// NewChainHandler creates a new chain handler.
// The handler uses the global error handler for formatting
// and metrics recording.
func NewChainHandler() *ChainHandler {
	return &ChainHandler{
		chain:   NewChain(),
		handler: GetErrorHandler(),
	}
}

// Add adds an error to the chain.
// Nil errors are ignored, allowing unconditional adds
// without nil checks.
func (ch *ChainHandler) Add(err error) {
	ch.chain.Add(err)
}

// Handle handles all errors in the chain.
// If the chain has any errors, they're formatted as a
// group and recorded in metrics.
func (ch *ChainHandler) Handle() {
	if ch.chain.HasErrors() {
		ch.handler.Handle(ch.chain)
	}
}

// HandleWithExit handles all errors and exits if any exist.
// It uses the exit code from the first error in the chain
// for process termination.
func (ch *ChainHandler) HandleWithExit() {
	if ch.chain.HasErrors() {
		ch.handler.HandleWithExit(ch.chain.First())
	}
}

// HasErrors returns true if the chain has errors.
// This allows checking before handling or for conditional logic
// based on error presence.
func (ch *ChainHandler) HasErrors() bool {
	return ch.chain.HasErrors()
}
