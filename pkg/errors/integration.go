// ABOUTME: This file provides integration between error handling and configuration system.
// ABOUTME: It enables debug mode error reporting and configuration-based error formatting.

package errors

import (
	"fmt"
	"io"
	"os"
)

// ErrorHandler provides centralized error handling with configuration support
type ErrorHandler struct {
	formatter *Formatter
	metrics   *ErrorMetrics
	debugMode bool
}

// GlobalErrorHandler is the default error handler instance
var globalErrorHandler *ErrorHandler

// InitializeErrorHandler initializes the global error handler with configuration
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

// GetErrorHandler returns the global error handler
func GetErrorHandler() *ErrorHandler {
	if globalErrorHandler == nil {
		// Initialize with defaults if not yet initialized
		InitializeErrorHandler(false, true)
	}
	return globalErrorHandler
}

// Handle processes an error, recording metrics and formatting for display
func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	// Record in metrics
	h.metrics.RecordError(err)

	// Format and print
	h.formatter.Print(err)
}

// HandleWithExit handles an error and exits with appropriate code
func (h *ErrorHandler) HandleWithExit(err error) {
	if err == nil {
		return
	}

	h.Handle(err)
	os.Exit(GetExitCode(err))
}

// SetOutput sets the output writer for error messages
func (h *ErrorHandler) SetOutput(w io.Writer) {
	h.formatter.SetWriter(w)
}

// SetDebugMode updates the debug mode setting
func (h *ErrorHandler) SetDebugMode(debug bool) {
	h.debugMode = debug

	// Update formatter options
	opts := h.formatter.options
	opts.DebugMode = debug
	opts.ShowStackTrace = debug
	h.formatter.options = opts
}

// GetMetrics returns the error metrics instance
func (h *ErrorHandler) GetMetrics() *ErrorMetrics {
	return h.metrics
}

// FormatError formats an error without recording metrics
func (h *ErrorHandler) FormatError(err error) string {
	return h.formatter.Format(err)
}

// PrintMetrics prints error metrics to the configured output
func (h *ErrorHandler) PrintMetrics(options MetricsFormatterOptions) {
	if h.metrics == nil {
		return
	}

	formatted := FormatMetrics(h.metrics, options)
	_, _ = fmt.Fprint(h.formatter.writer, formatted)
}

// Quick helper functions for common operations

// HandleError is a convenience function that handles an error using the global handler
func HandleError(err error) {
	GetErrorHandler().Handle(err)
}

// HandleErrorWithExit is a convenience function that handles an error and exits
func HandleErrorWithExit(err error) {
	GetErrorHandler().HandleWithExit(err)
}

// DebugError handles an error with debug information regardless of global debug mode
func DebugError(err error) {
	if err == nil {
		return
	}

	// Create a debug formatter
	opts := DebugFormatterOptions()
	debugFormatter := NewFormatter(opts)
	debugFormatter.Print(err)
}

// Must panics if err is not nil, useful for initialization code
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustReturn returns the value or panics if err is not nil
func MustReturn[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Recover recovers from a panic and returns it as an error
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

// RecoverWithHandler recovers from panic and handles the error
func RecoverWithHandler(handler func(error)) {
	if err := Recover(); err != nil {
		handler(err)
	}
}

// ErrorContextKey is used for storing errors in context
type ErrorContextKey struct{}

// ChainHandler handles multiple errors as a chain
type ChainHandler struct {
	chain   *Chain
	handler *ErrorHandler
}

// NewChainHandler creates a new chain handler
func NewChainHandler() *ChainHandler {
	return &ChainHandler{
		chain:   NewChain(),
		handler: GetErrorHandler(),
	}
}

// Add adds an error to the chain
func (ch *ChainHandler) Add(err error) {
	ch.chain.Add(err)
}

// Handle handles all errors in the chain
func (ch *ChainHandler) Handle() {
	if ch.chain.HasErrors() {
		ch.handler.Handle(ch.chain)
	}
}

// HandleWithExit handles all errors and exits if any exist
func (ch *ChainHandler) HandleWithExit() {
	if ch.chain.HasErrors() {
		ch.handler.HandleWithExit(ch.chain.First())
	}
}

// HasErrors returns true if the chain has errors
func (ch *ChainHandler) HasErrors() bool {
	return ch.chain.HasErrors()
}
