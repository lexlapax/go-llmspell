// ABOUTME: Tests for engine registry integration, covering engine discovery and management.
// ABOUTME: Ensures proper integration with the existing engine registry from pkg/engine.

package runner

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock engine factory for testing
type mockEngineFactory struct {
	name           string
	version        string
	description    string
	fileExtensions []string
	features       []engine.EngineFeature
	createError    error
}

func (f *mockEngineFactory) Create(config engine.EngineConfig) (engine.ScriptEngine, error) {
	if f.createError != nil {
		return nil, f.createError
	}
	return &mockEngine{name: f.name}, nil
}

func (f *mockEngineFactory) Name() string                                    { return f.name }
func (f *mockEngineFactory) Version() string                                 { return f.version }
func (f *mockEngineFactory) Description() string                             { return f.description }
func (f *mockEngineFactory) FileExtensions() []string                        { return f.fileExtensions }
func (f *mockEngineFactory) Features() []engine.EngineFeature                { return f.features }
func (f *mockEngineFactory) ValidateConfig(config engine.EngineConfig) error { return nil }
func (f *mockEngineFactory) GetDefaultConfig() engine.EngineConfig {
	return engine.EngineConfig{}
}

// Mock engine for testing - implements the full ScriptEngine interface
type mockEngine struct {
	name string
}

func (e *mockEngine) Initialize(config engine.EngineConfig) error { return nil }
func (e *mockEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error) {
	return engine.NewStringValue("mock result"), nil
}
func (e *mockEngine) ExecuteFile(ctx context.Context, filepath string, params map[string]interface{}) (engine.ScriptValue, error) {
	return engine.NewStringValue("mock file result"), nil
}
func (e *mockEngine) Shutdown() error { return nil }

// Bridge management
func (e *mockEngine) RegisterBridge(bridge engine.Bridge) error    { return nil }
func (e *mockEngine) UnregisterBridge(name string) error           { return nil }
func (e *mockEngine) GetBridge(name string) (engine.Bridge, error) { return nil, nil }
func (e *mockEngine) ListBridges() []string                        { return []string{} }

// Type system
func (e *mockEngine) ToNative(scriptValue engine.ScriptValue) (interface{}, error) { return nil, nil }
func (e *mockEngine) FromNative(goValue interface{}) (engine.ScriptValue, error)   { return nil, nil }

// Metadata
func (e *mockEngine) Name() string                     { return e.name }
func (e *mockEngine) Version() string                  { return "1.0.0" }
func (e *mockEngine) FileExtensions() []string         { return []string{".mock"} }
func (e *mockEngine) Features() []engine.EngineFeature { return []engine.EngineFeature{} }

// Resource management
func (e *mockEngine) SetMemoryLimit(bytes int64) error                     { return nil }
func (e *mockEngine) SetTimeout(duration time.Duration) error              { return nil }
func (e *mockEngine) SetResourceLimits(limits engine.ResourceLimits) error { return nil }
func (e *mockEngine) GetMetrics() engine.EngineMetrics {
	return engine.EngineMetrics{}
}

// Script state management
func (e *mockEngine) CreateContext(options engine.ContextOptions) (engine.ScriptContext, error) {
	return nil, nil
}
func (e *mockEngine) DestroyContext(ctx engine.ScriptContext) error { return nil }
func (e *mockEngine) ExecuteScript(ctx context.Context, script string, options engine.ExecutionOptions) (*engine.ExecutionResult, error) {
	return &engine.ExecutionResult{}, nil
}

// Event bus
func (e *mockEngine) GetEventBus() engine.EventBus { return nil }

// Type conversion registry
func (e *mockEngine) RegisterTypeConverter(fromType, toType string, converter engine.TypeConverterFunc) error {
	return nil
}
func (e *mockEngine) GetTypeRegistry() engine.TypeRegistry { return nil }

// Profiling
func (e *mockEngine) EnableProfiling(config engine.ProfilingConfig) error  { return nil }
func (e *mockEngine) DisableProfiling() error                              { return nil }
func (e *mockEngine) GetProfilingReport() (*engine.ProfilingReport, error) { return nil, nil }

// API export
func (e *mockEngine) ExportAPI(format engine.ExportFormat) ([]byte, error) { return nil, nil }
func (e *mockEngine) GenerateClientLibrary(language string, options engine.ClientLibraryOptions) ([]byte, error) {
	return nil, nil
}

func TestEngineRegistryManager(t *testing.T) {
	t.Run("new_manager", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		manager := NewEngineRegistryManager(registry)

		assert.NotNil(t, manager)
		assert.Equal(t, registry, manager.registry)
	})

	t.Run("initialize", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		manager := NewEngineRegistryManager(registry)

		err := manager.Initialize()
		assert.NoError(t, err)
	})

	t.Run("register_engines", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register test engines
		factories := map[string]engine.EngineFactory{
			"lua": &mockEngineFactory{
				name:           "lua",
				version:        "1.0.0",
				fileExtensions: []string{"lua"},
			},
			"javascript": &mockEngineFactory{
				name:           "javascript",
				version:        "1.0.0",
				fileExtensions: []string{"js", "mjs"},
			},
		}

		err = manager.RegisterEngines(factories)
		assert.NoError(t, err)

		// Verify engines are registered
		engines := manager.ListEngines()
		assert.Len(t, engines, 2)

		// Find engine names
		engineNames := make(map[string]bool)
		for _, info := range engines {
			engineNames[info.Name] = true
		}
		assert.True(t, engineNames["lua"])
		assert.True(t, engineNames["javascript"])
	})

	t.Run("get_engine", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register a test engine
		factory := &mockEngineFactory{
			name:    "lua",
			version: "1.0.0",
		}
		err = registry.Register(factory)
		require.NoError(t, err)

		// Get the engine
		eng, err := manager.GetEngine("lua", engine.EngineConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, eng)
	})

	t.Run("get_nonexistent_engine", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		eng, err := manager.GetEngine("nonexistent", engine.EngineConfig{})
		assert.Error(t, err)
		assert.Nil(t, eng)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("find_engine_by_extension", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register test engines
		luaFactory := &mockEngineFactory{
			name:           "lua",
			fileExtensions: []string{"lua"},
		}
		jsFactory := &mockEngineFactory{
			name:           "javascript",
			fileExtensions: []string{"js", "mjs"},
		}

		_ = registry.Register(luaFactory)
		_ = registry.Register(jsFactory)

		// Test finding engines
		name, err := manager.FindEngineByExtension(".lua")
		assert.NoError(t, err)
		assert.Equal(t, "lua", name)

		name, err = manager.FindEngineByExtension("js")
		assert.NoError(t, err)
		assert.Equal(t, "javascript", name)

		name, err = manager.FindEngineByExtension(".mjs")
		assert.NoError(t, err)
		assert.Equal(t, "javascript", name)

		_, err = manager.FindEngineByExtension(".py")
		assert.Error(t, err)
	})

	t.Run("execute_script", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register test engine
		factory := &mockEngineFactory{
			name: "lua",
		}
		err = registry.Register(factory)
		require.NoError(t, err)

		// Execute script
		ctx := context.Background()
		params := map[string]interface{}{"test": true}
		result, err := manager.ExecuteScript(ctx, "lua", "test script", params)

		assert.NoError(t, err)
		// Result is a ScriptValue, need to check its string value
		sv, ok := result.(engine.ScriptValue)
		assert.True(t, ok, "result should be a ScriptValue")
		assert.Equal(t, "mock result", sv.String())
	})

	t.Run("execute_file", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register test engine
		factory := &mockEngineFactory{
			name:           "lua",
			fileExtensions: []string{"lua"},
		}
		err = registry.Register(factory)
		require.NoError(t, err)

		// Execute file
		ctx := context.Background()
		params := map[string]interface{}{"test": true}
		result, err := manager.ExecuteFile(ctx, "test.lua", params)

		assert.NoError(t, err)
		// Result is a ScriptValue, need to check its string value
		sv, ok := result.(engine.ScriptValue)
		assert.True(t, ok, "result should be a ScriptValue")
		assert.Equal(t, "mock file result", sv.String())
	})

	t.Run("get_engine_info", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		// Register test engine
		factory := &mockEngineFactory{
			name:        "lua",
			version:     "5.4.0",
			description: "Lua scripting engine",
			features: []engine.EngineFeature{
				engine.FeatureAsync,
				engine.FeatureDebugging,
			},
		}
		err = registry.Register(factory)
		require.NoError(t, err)

		// Get engine info
		info, err := manager.GetEngineInfo("lua")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "lua", info.Name)
		assert.Equal(t, "5.4.0", info.Version)
		assert.Equal(t, "Lua scripting engine", info.Description)
		assert.Contains(t, info.Features, engine.FeatureAsync)
		assert.Contains(t, info.Features, engine.FeatureDebugging)
	})

	t.Run("get_stats", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{
			MetricsEnabled: true,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		manager := NewEngineRegistryManager(registry)

		// Register and use an engine
		factory := &mockEngineFactory{name: "lua"}
		err = registry.Register(factory)
		require.NoError(t, err)

		ctx := context.Background()
		_, _ = manager.ExecuteScript(ctx, "lua", "test", nil)

		// Get stats
		stats := manager.GetStats()
		assert.NotNil(t, stats)
		assert.Contains(t, stats, "lua")
		// SuccessCount is 2: 1 for GetEngine and 1 for ExecuteScript
		assert.Equal(t, int64(2), stats["lua"].SuccessCount)
	})

	t.Run("shutdown", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)

		err = manager.Shutdown()
		assert.NoError(t, err)
	})
}

func TestEngineConfigBuilder(t *testing.T) {
	t.Run("build_default_config", func(t *testing.T) {
		config := BuildEngineConfig(nil, nil)

		assert.NotNil(t, config)
		assert.Equal(t, int64(64*1024*1024), config.MemoryLimit)
		assert.Equal(t, 30*time.Second, config.TimeoutLimit)
		assert.True(t, config.SandboxMode)
	})

	t.Run("build_with_runner_config", func(t *testing.T) {
		runnerConfig := &RunnerConfig{
			Timeout: 60 * time.Second,
			EngineConfigs: map[string]map[string]interface{}{
				"lua": {
					"memory_limit": 128 * 1024 * 1024,
					"sandbox_mode": false,
				},
			},
		}

		config := BuildEngineConfig(runnerConfig, nil)
		assert.Equal(t, 60*time.Second, config.TimeoutLimit)

		// With engine-specific config
		luaConfig := BuildEngineConfig(runnerConfig, runnerConfig.EngineConfigs["lua"])
		assert.Equal(t, int64(128*1024*1024), luaConfig.MemoryLimit)
		assert.False(t, luaConfig.SandboxMode)
	})

	t.Run("engine_config_overrides_runner", func(t *testing.T) {
		runnerConfig := &RunnerConfig{
			Timeout: 60 * time.Second,
		}
		engineConfig := map[string]interface{}{
			"timeout": 120,
		}

		config := BuildEngineConfig(runnerConfig, engineConfig)
		assert.Equal(t, 120*time.Second, config.TimeoutLimit)
	})
}

// Benchmark tests
func BenchmarkEngineRegistryManager_GetEngine(b *testing.B) {
	registry := engine.NewRegistry(engine.RegistryConfig{})
	_ = registry.Initialize()
	manager := NewEngineRegistryManager(registry)

	factory := &mockEngineFactory{name: "lua"}
	_ = registry.Register(factory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GetEngine("lua", engine.EngineConfig{})
	}
}

func BenchmarkEngineRegistryManager_FindByExtension(b *testing.B) {
	registry := engine.NewRegistry(engine.RegistryConfig{})
	_ = registry.Initialize()
	manager := NewEngineRegistryManager(registry)

	factory := &mockEngineFactory{
		name:           "lua",
		fileExtensions: []string{"lua"},
	}
	_ = registry.Register(factory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.FindEngineByExtension(".lua")
	}
}
