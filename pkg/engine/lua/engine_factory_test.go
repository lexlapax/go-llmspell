// ABOUTME: Tests for LuaEngineFactory implementation, ensuring proper factory pattern functionality
// ABOUTME: Validates engine creation, configuration handling, and metadata correctness

package gopherlua

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestLuaEngineFactory_Creation(t *testing.T) {
	factory := NewLuaEngineFactory()

	assert.NotNil(t, factory)
	assert.Equal(t, "5.1.5", factory.version)
	assert.Equal(t, "Lua 5.1 scripting engine powered by gopher-lua", factory.description)
}

func TestLuaEngineFactory_Metadata(t *testing.T) {
	factory := NewLuaEngineFactory()

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "lua", factory.Name())
	})

	t.Run("version", func(t *testing.T) {
		assert.Equal(t, "5.1.5", factory.Version())
	})

	t.Run("description", func(t *testing.T) {
		assert.Equal(t, "Lua 5.1 scripting engine powered by gopher-lua", factory.Description())
	})

	t.Run("file_extensions", func(t *testing.T) {
		exts := factory.FileExtensions()
		assert.Equal(t, []string{"lua"}, exts)
	})

	t.Run("features", func(t *testing.T) {
		features := factory.Features()
		assert.NotEmpty(t, features)

		// Check for key features
		hasAsync := false
		hasDebugging := false
		hasModules := false

		for _, f := range features {
			switch f {
			case engine.FeatureAsync:
				hasAsync = true
			case engine.FeatureDebugging:
				hasDebugging = true
			case engine.FeatureModules:
				hasModules = true
			}
		}

		assert.True(t, hasAsync, "should have async feature")
		assert.True(t, hasDebugging, "should have debugging feature")
		assert.True(t, hasModules, "should have modules feature")
	})
}

func TestLuaEngineFactory_ValidateConfig(t *testing.T) {
	factory := NewLuaEngineFactory()

	tests := []struct {
		name    string
		config  engine.EngineConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_default_config",
			config:  factory.GetDefaultConfig(),
			wantErr: false,
		},
		{
			name: "valid_custom_config",
			config: engine.EngineConfig{
				MemoryLimit:    128 * 1024 * 1024,
				TimeoutLimit:   60 * time.Second,
				GoroutineLimit: 200,
				FileSystemMode: engine.FSModeReadWrite,
				LogLevel:       "debug",
			},
			wantErr: false,
		},
		{
			name: "negative_memory_limit",
			config: engine.EngineConfig{
				MemoryLimit: -1,
			},
			wantErr: true,
			errMsg:  "memory limit cannot be negative",
		},
		{
			name: "negative_timeout",
			config: engine.EngineConfig{
				TimeoutLimit: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout limit cannot be negative",
		},
		{
			name: "negative_goroutine_limit",
			config: engine.EngineConfig{
				GoroutineLimit: -1,
			},
			wantErr: true,
			errMsg:  "goroutine limit cannot be negative",
		},
		{
			name: "invalid_filesystem_mode",
			config: engine.EngineConfig{
				FileSystemMode: engine.FSMode("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid file system mode",
		},
		{
			name: "invalid_log_level",
			config: engine.EngineConfig{
				LogLevel: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "empty_log_level_ok",
			config: engine.EngineConfig{
				LogLevel: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLuaEngineFactory_GetDefaultConfig(t *testing.T) {
	factory := NewLuaEngineFactory()
	config := factory.GetDefaultConfig()

	// Check resource limits
	assert.Equal(t, int64(64*1024*1024), config.MemoryLimit)
	assert.Equal(t, 30*time.Second, config.TimeoutLimit)
	assert.Equal(t, 100, config.GoroutineLimit)

	// Check security settings
	assert.True(t, config.SandboxMode)
	assert.Equal(t, engine.FSModeReadOnly, config.FileSystemMode)

	// Check feature flags
	assert.False(t, config.DebugMode)
	assert.Equal(t, "info", config.LogLevel)
	assert.True(t, config.MetricsMode)
	assert.False(t, config.TracingMode)

	// Check engine-specific options
	assert.NotNil(t, config.EngineOptions)
	assert.Equal(t, 2048, config.EngineOptions["registry_size"])
	assert.Equal(t, 1024, config.EngineOptions["call_stack_size"])
	assert.Equal(t, 5, config.EngineOptions["pool_size"])
	assert.Equal(t, true, config.EngineOptions["preload_stdlib"])
	assert.Equal(t, true, config.EngineOptions["enable_chunk_cache"])
	assert.Equal(t, 100, config.EngineOptions["chunk_cache_size"])
	assert.Equal(t, true, config.EngineOptions["enable_compilation_optimization"])
}

func TestLuaEngineFactory_Create(t *testing.T) {
	factory := NewLuaEngineFactory()

	t.Run("create_with_default_config", func(t *testing.T) {
		config := factory.GetDefaultConfig()
		engine, err := factory.Create(config)

		require.NoError(t, err)
		require.NotNil(t, engine)

		// Verify it's actually a LuaEngine
		luaEngine, ok := engine.(*LuaEngine)
		assert.True(t, ok, "should be a LuaEngine instance")
		assert.NotNil(t, luaEngine)

		// Clean up
		err = engine.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("create_with_custom_config", func(t *testing.T) {
		config := engine.EngineConfig{
			MemoryLimit:    32 * 1024 * 1024,
			TimeoutLimit:   10 * time.Second,
			GoroutineLimit: 50,
			SandboxMode:    true,
			FileSystemMode: engine.FSModeNone,
			DebugMode:      true,
			LogLevel:       "debug",
			MetricsMode:    false,
			EngineOptions: map[string]interface{}{
				"pool_size": 3,
			},
		}

		engine, err := factory.Create(config)

		require.NoError(t, err)
		require.NotNil(t, engine)

		// Clean up
		err = engine.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("create_with_invalid_config", func(t *testing.T) {
		config := engine.EngineConfig{
			MemoryLimit: -1,
		}

		engine, err := factory.Create(config)

		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "invalid configuration")
		assert.Contains(t, err.Error(), "memory limit cannot be negative")
	})
}

func TestLuaEngineFactory_InterfaceCompliance(t *testing.T) {
	// Ensure LuaEngineFactory implements engine.EngineFactory
	var _ engine.EngineFactory = (*LuaEngineFactory)(nil)
}

// Benchmark tests
func BenchmarkLuaEngineFactory_Create(b *testing.B) {
	factory := NewLuaEngineFactory()
	config := factory.GetDefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine, err := factory.Create(config)
		if err != nil {
			b.Fatal(err)
		}
		_ = engine.Shutdown()
	}
}

func BenchmarkLuaEngineFactory_ValidateConfig(b *testing.B) {
	factory := NewLuaEngineFactory()
	config := factory.GetDefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = factory.ValidateConfig(config)
	}
}

func BenchmarkLuaEngineFactory_GetDefaultConfig(b *testing.B) {
	factory := NewLuaEngineFactory()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = factory.GetDefaultConfig()
	}
}
