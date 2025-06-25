// ABOUTME: LuaEngineFactory implements the EngineFactory interface for creating Lua script engines
// ABOUTME: Provides factory pattern for LuaEngine instances with proper configuration and validation

package gopherlua

import (
	"fmt"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// LuaEngineFactory creates new instances of LuaEngine
type LuaEngineFactory struct {
	version     string
	description string
}

// NewLuaEngineFactory creates a new factory for Lua engines
func NewLuaEngineFactory() *LuaEngineFactory {
	return &LuaEngineFactory{
		version:     "5.1.5", // Gopher-Lua implements Lua 5.1
		description: "Lua 5.1 scripting engine powered by gopher-lua",
	}
}

// Create creates a new LuaEngine instance with the given configuration
func (f *LuaEngineFactory) Create(config engine.EngineConfig) (engine.ScriptEngine, error) {
	// Validate configuration
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create the engine
	luaEngine := NewLuaEngine()

	// Initialize with config
	if err := luaEngine.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize Lua engine: %w", err)
	}

	return luaEngine, nil
}

// Name returns the name of the engine
func (f *LuaEngineFactory) Name() string {
	return "lua"
}

// Version returns the version of the Lua engine
func (f *LuaEngineFactory) Version() string {
	return f.version
}

// Description returns a description of the engine
func (f *LuaEngineFactory) Description() string {
	return f.description
}

// FileExtensions returns the file extensions supported by Lua
func (f *LuaEngineFactory) FileExtensions() []string {
	return []string{"lua"}
}

// Features returns the features supported by the Lua engine
func (f *LuaEngineFactory) Features() []engine.EngineFeature {
	return []engine.EngineFeature{
		engine.FeatureAsync,
		engine.FeatureDebugging,
		engine.FeatureModules,
		engine.FeatureHotReload,
		engine.FeatureCoroutines,
		engine.FeatureCompilation,
		engine.FeatureInteractive,
		engine.FeatureStreaming,
	}
}

// ValidateConfig validates the engine configuration
func (f *LuaEngineFactory) ValidateConfig(config engine.EngineConfig) error {
	// Validate memory limit
	if config.MemoryLimit < 0 {
		return fmt.Errorf("memory limit cannot be negative")
	}

	// Validate timeout
	if config.TimeoutLimit < 0 {
		return fmt.Errorf("timeout limit cannot be negative")
	}

	// Validate goroutine limit
	if config.GoroutineLimit < 0 {
		return fmt.Errorf("goroutine limit cannot be negative")
	}

	// Validate file system mode
	switch config.FileSystemMode {
	case engine.FSModeNone, engine.FSModeReadOnly, engine.FSModeReadWrite, engine.FSModeSandbox, "":
		// Valid modes (empty string is allowed and means default)
	default:
		return fmt.Errorf("invalid file system mode: %v", config.FileSystemMode)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}
	if config.LogLevel != "" && !validLogLevels[config.LogLevel] {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	return nil
}

// GetDefaultConfig returns the default configuration for the Lua engine
func (f *LuaEngineFactory) GetDefaultConfig() engine.EngineConfig {
	return engine.EngineConfig{
		// Resource limits
		MemoryLimit:    64 * 1024 * 1024, // 64MB
		TimeoutLimit:   30 * time.Second,
		GoroutineLimit: 100,

		// Security
		SandboxMode:    true,
		FileSystemMode: engine.FSModeReadOnly,

		// Features
		DebugMode:   false,
		LogLevel:    "info",
		MetricsMode: true,
		TracingMode: false,

		// Engine-specific options
		EngineOptions: map[string]interface{}{
			"registry_size":                   2048,
			"call_stack_size":                 1024,
			"registry_max_size":               0, // 0 means no limit
			"registry_grow_step":              32,
			"min_stack_size":                  20,
			"pool_size":                       5,
			"preload_stdlib":                  true,
			"enable_chunk_cache":              true,
			"chunk_cache_size":                100,
			"enable_compilation_optimization": true,
		},
	}
}
