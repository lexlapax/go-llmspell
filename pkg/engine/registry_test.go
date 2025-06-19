// ABOUTME: Test suite for the engine registry that manages multiple script engines.
// ABOUTME: Validates thread-safe registration, discovery, and factory patterns for script engines.

package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementation of EngineFactory for testing
type mockEngineFactory struct {
	name           string
	version        string
	description    string
	fileExtensions []string
	features       []EngineFeature
	createError    error
}

func newMockEngineFactory(name string) *mockEngineFactory {
	return &mockEngineFactory{
		name:           name,
		version:        "1.0.0",
		description:    "Mock engine for " + name,
		fileExtensions: []string{name, "." + name},
		features:       []EngineFeature{FeatureAsync, FeatureDebugging},
	}
}

func (f *mockEngineFactory) Create(config EngineConfig) (ScriptEngine, error) {
	if f.createError != nil {
		return nil, f.createError
	}
	engine := newTestMockScriptEngine(f.name)
	// Configure custom execute behavior for error handling
	engine.withExecuteFunc(func(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error) {
		if script == "error" {
			return nil, &EngineError{
				Type:    ErrorTypeRuntime,
				Message: "runtime error",
			}
		}
		return NewStringValue("executed: " + script), nil
	})
	if err := engine.Initialize(config); err != nil {
		return nil, err
	}
	return engine, nil
}

func (f *mockEngineFactory) Name() string {
	return f.name
}

func (f *mockEngineFactory) Version() string {
	return f.version
}

func (f *mockEngineFactory) Description() string {
	return f.description
}

func (f *mockEngineFactory) FileExtensions() []string {
	return f.fileExtensions
}

func (f *mockEngineFactory) Features() []EngineFeature {
	return f.features
}

func (f *mockEngineFactory) ValidateConfig(config EngineConfig) error {
	if config.MemoryLimit < 0 {
		return errors.New("invalid memory limit")
	}
	if config.TimeoutLimit < 0 {
		return errors.New("invalid timeout limit")
	}
	return nil
}

func (f *mockEngineFactory) GetDefaultConfig() EngineConfig {
	return EngineConfig{
		MemoryLimit:    1024 * 1024,
		TimeoutLimit:   30 * time.Second,
		GoroutineLimit: 10,
		SandboxMode:    true,
	}
}

// NOTE: Using local testMockScriptEngine from test_helpers.go instead of testutils.MockScriptEngine
// to avoid import cycles (testutils imports engine, so engine tests cannot import testutils)

// Tests for Registry
func TestRegistry(t *testing.T) {
	t.Run("NewRegistry", func(t *testing.T) {
		config := RegistryConfig{
			MaxEngines:        5,
			DefaultTimeout:    10 * time.Second,
			HealthCheckPeriod: 30 * time.Second,
			PoolingEnabled:    true,
			MaxPoolSize:       3,
		}

		registry := NewRegistry(config)
		assert.NotNil(t, registry)
	})

	t.Run("Initialize", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})

		err := registry.Initialize()
		assert.NoError(t, err)

		// Test double initialization
		err = registry.Initialize()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
	})

	t.Run("Register", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			MaxEngines: 3,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		// Register engines
		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		assert.NoError(t, err)

		js := newMockEngineFactory("javascript")
		err = registry.Register(js)
		assert.NoError(t, err)

		// Test duplicate registration
		err = registry.Register(lua)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")

		// Test empty name
		emptyFactory := newMockEngineFactory("")
		err = registry.Register(emptyFactory)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")

		// Test max engines limit
		tengo := newMockEngineFactory("tengo")
		err = registry.Register(tengo)
		assert.NoError(t, err)

		python := newMockEngineFactory("python")
		err = registry.Register(python)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maximum number of engines")
	})

	t.Run("Register with Allowed/Disallowed", func(t *testing.T) {
		// Test allowed engines
		registry := NewRegistry(RegistryConfig{
			AllowedEngines: []string{"lua", "javascript"},
		})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		assert.NoError(t, err)

		python := newMockEngineFactory("python")
		err = registry.Register(python)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in allowed list")

		// Test disallowed engines
		registry2 := NewRegistry(RegistryConfig{
			DisallowedEngines: []string{"python", "ruby"},
		})
		err = registry2.Initialize()
		require.NoError(t, err)

		lua2 := newMockEngineFactory("lua")
		err = registry2.Register(lua2)
		assert.NoError(t, err)

		python2 := newMockEngineFactory("python")
		err = registry2.Register(python2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "disallowed")
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Create an instance
		eng, err := registry.GetEngine("lua", EngineConfig{})
		require.NoError(t, err)
		require.NotNil(t, eng)

		// Unregister
		err = registry.Unregister("lua")
		assert.NoError(t, err)

		// Verify engine is gone
		_, err = registry.GetEngine("lua", EngineConfig{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test unregistering non-existent engine
		err = registry.Unregister("python")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetEngine without pooling", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			PoolingEnabled: false,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Get engine
		engine1, err := registry.GetEngine("lua", EngineConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, engine1)

		// Get another instance - should be different
		engine2, err := registry.GetEngine("lua", EngineConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, engine2)

		// Verify they are different instances (different memory addresses)
		assert.NotSame(t, engine1, engine2)

		// Check metrics
		stats := registry.GetStats()
		assert.Equal(t, int64(2), stats["lua"].InstancesCreated)
		assert.Equal(t, int64(2), stats["lua"].SuccessCount)

		// Test non-existent engine
		_, err = registry.GetEngine("python", EngineConfig{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetEngine with pooling", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			PoolingEnabled: true,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Get engine
		engine1, err := registry.GetEngine("lua", EngineConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, engine1)

		// Get same instance
		engine2, err := registry.GetEngine("lua", EngineConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, engine2)
		assert.Equal(t, engine1, engine2) // Should be same instance with pooling

		// Check metrics
		stats := registry.GetStats()
		assert.Equal(t, int64(1), stats["lua"].InstancesCreated)
		assert.Equal(t, int64(1), stats["lua"].InstancesActive)
	})

	t.Run("GetEngine with creation error", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		badFactory := newMockEngineFactory("bad")
		badFactory.createError = errors.New("creation failed")
		err = registry.Register(badFactory)
		require.NoError(t, err)

		_, err = registry.GetEngine("bad", EngineConfig{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create engine")
		assert.Contains(t, err.Error(), "creation failed")

		// Check error count in metrics
		stats := registry.GetStats()
		assert.Equal(t, int64(1), stats["bad"].ErrorCount)
	})

	t.Run("ListEngines", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			MetricsEnabled: true,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		// Register engines
		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		js := newMockEngineFactory("javascript")
		err = registry.Register(js)
		require.NoError(t, err)

		// List engines
		engines := registry.ListEngines()
		assert.Len(t, engines, 2)

		// Find lua engine info
		var luaInfo *EngineInfo
		for i := range engines {
			if engines[i].Name == "lua" {
				luaInfo = &engines[i]
				break
			}
		}

		require.NotNil(t, luaInfo)
		assert.Equal(t, "lua", luaInfo.Name)
		assert.Equal(t, "1.0.0", luaInfo.Version)
		assert.Equal(t, "Mock engine for lua", luaInfo.Description)
		assert.Contains(t, luaInfo.FileExtensions, "lua")
		assert.Contains(t, luaInfo.Features, FeatureAsync)
		assert.Equal(t, EngineStatusRegistered, luaInfo.Status)
		assert.NotNil(t, luaInfo.Stats)
	})

	t.Run("FindEngineByExtension", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		lua.fileExtensions = []string{"lua", "luac"}
		err = registry.Register(lua)
		require.NoError(t, err)

		js := newMockEngineFactory("javascript")
		js.fileExtensions = []string{"js", "mjs", "jsx"}
		err = registry.Register(js)
		require.NoError(t, err)

		// Test finding engines
		eng, err := registry.FindEngineByExtension("lua")
		assert.NoError(t, err)
		assert.Equal(t, "lua", eng)

		eng, err = registry.FindEngineByExtension(".js")
		assert.NoError(t, err)
		assert.Equal(t, "javascript", eng)

		eng, err = registry.FindEngineByExtension("JSX") // Test case insensitive
		assert.NoError(t, err)
		assert.Equal(t, "javascript", eng)

		// Test not found
		_, err = registry.FindEngineByExtension("py")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no engine found")
	})

	t.Run("FindEngineByFeature", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		lua.features = []EngineFeature{FeatureAsync, FeatureCoroutines}
		err = registry.Register(lua)
		require.NoError(t, err)

		js := newMockEngineFactory("javascript")
		js.features = []EngineFeature{FeatureAsync, FeatureModules}
		err = registry.Register(js)
		require.NoError(t, err)

		tengo := newMockEngineFactory("tengo")
		tengo.features = []EngineFeature{FeatureCompilation}
		err = registry.Register(tengo)
		require.NoError(t, err)

		// Find engines with async
		engines := registry.FindEngineByFeature(FeatureAsync)
		assert.Len(t, engines, 2)
		assert.Contains(t, engines, "lua")
		assert.Contains(t, engines, "javascript")

		// Find engines with coroutines
		engines = registry.FindEngineByFeature(FeatureCoroutines)
		assert.Len(t, engines, 1)
		assert.Contains(t, engines, "lua")

		// Find engines with streaming (none)
		engines = registry.FindEngineByFeature(FeatureStreaming)
		assert.Len(t, engines, 0)
	})

	t.Run("GetEngineInfo", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			MetricsEnabled: true,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Get info
		info, err := registry.GetEngineInfo("lua")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "lua", info.Name)
		assert.Equal(t, "1.0.0", info.Version)
		assert.NotNil(t, info.Stats)

		// Test non-existent engine
		_, err = registry.GetEngineInfo("python")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ExecuteScript", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Execute script
		result, err := registry.ExecuteScript(context.Background(), "lua", "print('hello')", nil)
		assert.NoError(t, err)
		// Extract value from ScriptValue
		if sv, ok := result.(ScriptValue); ok {
			assert.Equal(t, "executed: print('hello')", sv.ToGo())
		} else {
			assert.Equal(t, "executed: print('hello')", result)
		}

		// Check metrics updated
		stats := registry.GetStats()
		assert.Equal(t, int64(2), stats["lua"].SuccessCount) // 1 from GetEngine + 1 from ExecuteScript
		assert.Greater(t, stats["lua"].TotalExecTime, time.Duration(0))

		// Test with non-existent engine
		_, err = registry.ExecuteScript(context.Background(), "python", "print('hello')", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test with error script
		_, err = registry.ExecuteScript(context.Background(), "lua", "error", nil)
		assert.Error(t, err)
		stats = registry.GetStats()
		assert.Equal(t, int64(1), stats["lua"].ErrorCount)
	})

	t.Run("ExecuteFile", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Execute file
		result, err := registry.ExecuteFile(context.Background(), "test.lua", nil)
		assert.NoError(t, err)
		// Extract value from ScriptValue
		if sv, ok := result.(ScriptValue); ok {
			assert.Equal(t, "executed file: test.lua", sv.ToGo())
		} else {
			assert.Equal(t, "executed file: test.lua", result)
		}

		// Test file without extension
		_, err = registry.ExecuteFile(context.Background(), "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "without extension")

		// Test unknown extension
		_, err = registry.ExecuteFile(context.Background(), "test.py", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no engine found")
	})

	t.Run("Shutdown", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{
			PoolingEnabled: true,
		})
		err := registry.Initialize()
		require.NoError(t, err)

		lua := newMockEngineFactory("lua")
		err = registry.Register(lua)
		require.NoError(t, err)

		// Create instance
		eng, err := registry.GetEngine("lua", EngineConfig{})
		require.NoError(t, err)
		require.NotNil(t, eng)

		// Shutdown
		err = registry.Shutdown()
		assert.NoError(t, err)

		// Verify we can't use registry after shutdown
		_, err = registry.GetEngine("lua", EngineConfig{})
		assert.Error(t, err)
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		registry := NewRegistry(RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)

		// Register multiple engines
		engines := []string{"lua", "javascript", "tengo"}
		for _, name := range engines {
			factory := newMockEngineFactory(name)
			err := registry.Register(factory)
			require.NoError(t, err)
		}

		// Concurrent operations
		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Concurrent engine creation
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				engineName := engines[id%len(engines)]
				_, err := registry.GetEngine(engineName, EngineConfig{})
				if err != nil {
					errors <- err
				}
			}(i)
		}

		// Concurrent script execution
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				engineName := engines[id%len(engines)]
				_, err := registry.ExecuteScript(context.Background(), engineName, "test", nil)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		// Concurrent listing
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				list := registry.ListEngines()
				if len(list) != len(engines) {
					errors <- fmt.Errorf("incorrect engine count")
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		var errCount int
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
			errCount++
		}
		assert.Equal(t, 0, errCount)
	})
}
