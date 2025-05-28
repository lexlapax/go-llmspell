// ABOUTME: Logging module for Lua scripts using slog
// ABOUTME: Provides log.info(), error(), debug(), warn() functions

package stdlib

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// Logger provides logging functionality for Lua scripts
type Logger struct {
	logger *slog.Logger
	ctx    context.Context
}

// NewLogger creates a new logger instance
func NewLogger(name string, level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler).With("spell", name)

	return &Logger{
		logger: logger,
		ctx:    context.Background(),
	}
}

// RegisterLog registers the log module with all functions
func RegisterLog(L *lua.LState, logger *Logger) {
	// Create log module table
	logModule := L.NewTable()

	// Register functions
	L.SetField(logModule, "debug", L.NewClosure(logger.debug))
	L.SetField(logModule, "info", L.NewClosure(logger.info))
	L.SetField(logModule, "warn", L.NewClosure(logger.warn))
	L.SetField(logModule, "error", L.NewClosure(logger.error))

	// Register the module
	L.SetGlobal("log", logModule)
}

// formatMessage formats log arguments into a single message and extracts attributes
func (l *Logger) formatMessage(L *lua.LState) (string, []slog.Attr) {
	n := L.GetTop()
	if n == 0 {
		return "", nil
	}

	// First argument is the message
	msg := lua.LVAsString(L.Get(1))

	// Additional arguments can be key-value pairs for structured logging
	attrs := []slog.Attr{}
	for i := 2; i <= n; i += 2 {
		if i+1 <= n {
			key := lua.LVAsString(L.Get(i))
			value := lua.LVAsString(L.Get(i + 1))
			attrs = append(attrs, slog.String(key, value))
		}
	}

	return msg, attrs
}

// debug logs a debug message
func (l *Logger) debug(L *lua.LState) int {
	msg, attrs := l.formatMessage(L)
	l.logger.LogAttrs(l.ctx, slog.LevelDebug, msg, attrs...)
	return 0
}

// info logs an info message
func (l *Logger) info(L *lua.LState) int {
	msg, attrs := l.formatMessage(L)
	l.logger.LogAttrs(l.ctx, slog.LevelInfo, msg, attrs...)
	return 0
}

// warn logs a warning message
func (l *Logger) warn(L *lua.LState) int {
	msg, attrs := l.formatMessage(L)
	l.logger.LogAttrs(l.ctx, slog.LevelWarn, msg, attrs...)
	return 0
}

// error logs an error message
func (l *Logger) error(L *lua.LState) int {
	msg, attrs := l.formatMessage(L)
	l.logger.LogAttrs(l.ctx, slog.LevelError, msg, attrs...)
	return 0
}

// RegisterSimpleLog registers a simplified log module (used in examples)
func RegisterSimpleLog(L *lua.LState) {
	logModule := L.NewTable()

	// Simple info function that prints to stdout
	L.SetField(logModule, "info", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		fmt.Printf("[INFO] %s\n", msg)
		return 0
	}))

	// Simple error function that prints to stderr
	L.SetField(logModule, "error", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
		return 0
	}))

	L.SetGlobal("log", logModule)
}
