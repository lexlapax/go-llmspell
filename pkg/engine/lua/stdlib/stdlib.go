// ABOUTME: Main entry point for registering all standard library modules
// ABOUTME: Provides RegisterAll() to register json, log, storage, http modules

package stdlib

import (
	"log/slog"

	lua "github.com/yuin/gopher-lua"
)

// Config holds configuration for all stdlib modules
type Config struct {
	Storage   *StorageConfig
	HTTP      *HTTPConfig
	LogLevel  slog.Level
	SpellName string
}

// DefaultConfig returns a default stdlib configuration
func DefaultConfig() *Config {
	return &Config{
		Storage:   DefaultStorageConfig(),
		HTTP:      DefaultHTTPConfig(),
		LogLevel:  slog.LevelInfo,
		SpellName: "spell",
	}
}

// RegisterAll registers all standard library modules
func RegisterAll(L *lua.LState, config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	// Register JSON module
	RegisterJSON(L)

	// Register Log module
	logger := NewLogger(config.SpellName, config.LogLevel)
	RegisterLog(L, logger)

	// Register Storage module
	storage, err := NewStorage(config.Storage)
	if err != nil {
		return err
	}
	RegisterStorage(L, storage)

	// Register HTTP module
	httpClient := NewHTTPClient(config.HTTP)
	RegisterHTTP(L, httpClient)

	// Register Promise module for async operations
	RegisterPromise(L)
	
	// Register Async callback module
	RegisterAsyncCallback(L)
	
	// Register Promise-Async integration
	RegisterPromiseAsync(L)

	return nil
}

// RegisterMinimal registers only the essential modules without external dependencies
// This is useful for testing or restricted environments
func RegisterMinimal(L *lua.LState) {
	// Register JSON module
	RegisterJSON(L)

	// Register simple log module (prints to stdout/stderr)
	RegisterSimpleLog(L)

	// Register simple HTTP module
	RegisterSimpleHTTP(L)
}
