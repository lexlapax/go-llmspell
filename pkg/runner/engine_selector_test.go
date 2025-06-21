// ABOUTME: Tests for engine selection logic, covering file extension detection and engine matching.
// ABOUTME: Ensures proper engine selection based on file types and explicit configuration.

package runner

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineSelector(t *testing.T) {
	t.Run("new_selector", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		assert.NotNil(t, selector)
		assert.Equal(t, manager, selector.manager)
	})

	t.Run("select_by_extension", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
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
		
		tests := []struct {
			filepath string
			expected string
			wantErr  bool
		}{
			{"script.lua", "lua", false},
			{"path/to/script.lua", "lua", false},
			{"script.js", "javascript", false},
			{"module.mjs", "javascript", false},
			{"SCRIPT.LUA", "lua", false}, // Case insensitive
			{"script.py", "", true},       // No Python engine
			{"script", "", true},          // No extension
			{"", "", true},                // Empty path
		}
		
		for _, tt := range tests {
			t.Run(tt.filepath, func(t *testing.T) {
				engine, err := selector.SelectByExtension(tt.filepath)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Empty(t, engine)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, engine)
				}
			})
		}
	})

	t.Run("select_for_spell", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register test engines
		luaFactory := &mockEngineFactory{
			name:           "lua",
			fileExtensions: []string{"lua"},
		}
		jsFactory := &mockEngineFactory{
			name:           "javascript",
			fileExtensions: []string{"js"},
		}
		
		_ = registry.Register(luaFactory)
		_ = registry.Register(jsFactory)
		
		tests := []struct {
			name     string
			metadata *SpellMetadata
			expected string
			wantErr  bool
		}{
			{
				name: "explicit_engine",
				metadata: &SpellMetadata{
					Name:   "test",
					Engine: "lua",
				},
				expected: "lua",
				wantErr:  false,
			},
			{
				name: "engine_from_entrypoint",
				metadata: &SpellMetadata{
					Name:       "test",
					EntryPoint: "main.js",
				},
				expected: "javascript",
				wantErr:  false,
			},
			{
				name: "explicit_overrides_entrypoint",
				metadata: &SpellMetadata{
					Name:       "test",
					Engine:     "lua",
					EntryPoint: "main.js",
				},
				expected: "lua",
				wantErr:  false,
			},
			{
				name: "nonexistent_engine",
				metadata: &SpellMetadata{
					Name:   "test",
					Engine: "python",
				},
				expected: "",
				wantErr:  true,
			},
			{
				name: "no_engine_info",
				metadata: &SpellMetadata{
					Name: "test",
				},
				expected: "",
				wantErr:  true,
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				engine, err := selector.SelectForSpell(tt.metadata)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Empty(t, engine)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, engine)
				}
			})
		}
	})

	t.Run("select_with_options", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register test engine
		luaFactory := &mockEngineFactory{
			name: "lua",
		}
		jsFactory := &mockEngineFactory{
			name: "javascript",
		}
		
		_ = registry.Register(luaFactory)
		_ = registry.Register(jsFactory)
		
		// Options override spell metadata
		metadata := &SpellMetadata{
			Name:   "test",
			Engine: "lua",
		}
		options := &RunnerOptions{
			Engine: "javascript",
		}
		
		engine, err := selector.SelectWithOptions(metadata, options)
		assert.NoError(t, err)
		assert.Equal(t, "javascript", engine)
		
		// No override in options
		options.Engine = ""
		engine, err = selector.SelectWithOptions(metadata, options)
		assert.NoError(t, err)
		assert.Equal(t, "lua", engine)
	})

	t.Run("validate_engine_availability", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register test engine
		luaFactory := &mockEngineFactory{
			name:     "lua",
			features: []engine.EngineFeature{engine.FeatureAsync},
		}
		_ = registry.Register(luaFactory)
		
		// Validate existing engine
		err = selector.ValidateEngineAvailability("lua")
		assert.NoError(t, err)
		
		// Validate non-existent engine
		err = selector.ValidateEngineAvailability("python")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})

	t.Run("get_supported_extensions", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
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
		
		extensions := selector.GetSupportedExtensions()
		assert.Len(t, extensions, 3)
		assert.Contains(t, extensions, "lua")
		assert.Contains(t, extensions, "js")
		assert.Contains(t, extensions, "mjs")
	})

	t.Run("get_engine_for_extension", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register test engines
		jsFactory := &mockEngineFactory{
			name:           "javascript",
			fileExtensions: []string{"js", "mjs"},
		}
		_ = registry.Register(jsFactory)
		
		// Build extension map
		extMap := selector.GetEngineExtensionMap()
		assert.Len(t, extMap, 2)
		assert.Equal(t, "javascript", extMap["js"])
		assert.Equal(t, "javascript", extMap["mjs"])
	})
}

func TestExtractExtension(t *testing.T) {
	tests := []struct {
		filepath string
		expected string
	}{
		{"script.lua", "lua"},
		{"path/to/script.js", "js"},
		{"file.test.mjs", "mjs"},
		{"UPPERCASE.LUA", "lua"},
		{"no_extension", ""},
		{"", ""},
		{".hidden", "hidden"},
		{"path/.hidden.js", "js"},
	}
	
	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			ext := extractExtension(tt.filepath)
			assert.Equal(t, tt.expected, ext)
		})
	}
}

func TestEngineSelectorPriority(t *testing.T) {
	t.Run("priority_order", func(t *testing.T) {
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register all engines
		engines := []string{"lua", "javascript", "tengo"}
		for _, name := range engines {
			factory := &mockEngineFactory{
				name:           name,
				fileExtensions: []string{name[:2]}, // lu, ja, te
			}
			err = registry.Register(factory)
			require.NoError(t, err)
		}
		
		// Test priority: options > metadata.Engine > extension
		metadata := &SpellMetadata{
			Name:       "test",
			Engine:     "lua",
			EntryPoint: "script.ja", // Would select javascript
		}
		
		// 1. Options override everything
		options := &RunnerOptions{Engine: "tengo"}
		engine, err := selector.SelectWithOptions(metadata, options)
		assert.NoError(t, err)
		assert.Equal(t, "tengo", engine)
		
		// 2. Metadata engine overrides extension
		options.Engine = ""
		engine, err = selector.SelectWithOptions(metadata, options)
		assert.NoError(t, err)
		assert.Equal(t, "lua", engine)
		
		// 3. Extension used when no explicit engine
		metadata.Engine = ""
		engine, err = selector.SelectWithOptions(metadata, options)
		assert.NoError(t, err)
		assert.Equal(t, "javascript", engine)
	})
}

// Benchmark tests
func BenchmarkEngineSelector_SelectByExtension(b *testing.B) {
	registry := engine.NewRegistry(engine.RegistryConfig{})
	manager := NewEngineRegistryManager(registry)
	selector := NewEngineSelector(manager)
	
	// Register test engine
	factory := &mockEngineFactory{
		name:           "lua",
		fileExtensions: []string{"lua"},
	}
	_ = registry.Register(factory)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.SelectByExtension("script.lua")
	}
}

func BenchmarkExtractExtension(b *testing.B) {
	testPath := "/path/to/deeply/nested/script.lua"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractExtension(testPath)
	}
}

func TestEngineSelection_Integration(t *testing.T) {
	t.Run("full_selection_flow", func(t *testing.T) {
		// Create registry with multiple engines
		registry := engine.NewRegistry(engine.RegistryConfig{})
		err := registry.Initialize()
		require.NoError(t, err)
		
		manager := NewEngineRegistryManager(registry)
		selector := NewEngineSelector(manager)
		
		// Register engines
		engines := map[string][]string{
			"lua":        {"lua"},
			"javascript": {"js", "mjs"},
			"tengo":      {"tengo"},
		}
		
		for name, exts := range engines {
			factory := &mockEngineFactory{
				name:           name,
				fileExtensions: exts,
				version:        "1.0.0",
			}
			err := registry.Register(factory)
			require.NoError(t, err)
		}
		
		// Test various selection scenarios
		scenarios := []struct {
			name     string
			file     string
			metadata *SpellMetadata
			options  *RunnerOptions
			expected string
		}{
			{
				name:     "simple_lua_file",
				file:     "hello.lua",
				expected: "lua",
			},
			{
				name:     "javascript_module",
				file:     "app.mjs",
				expected: "javascript",
			},
			{
				name: "spell_with_explicit_engine",
				metadata: &SpellMetadata{
					Name:       "my-spell",
					Engine:     "tengo",
					EntryPoint: "main.lua", // Would normally be lua
				},
				expected: "tengo",
			},
			{
				name: "options_override",
				metadata: &SpellMetadata{
					Name:   "test",
					Engine: "lua",
				},
				options: &RunnerOptions{
					Engine: "javascript",
				},
				expected: "javascript",
			},
		}
		
		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				var engine string
				var err error
				
				if sc.metadata != nil && sc.options != nil {
					engine, err = selector.SelectWithOptions(sc.metadata, sc.options)
				} else if sc.metadata != nil {
					engine, err = selector.SelectForSpell(sc.metadata)
				} else {
					engine, err = selector.SelectByExtension(sc.file)
				}
				
				assert.NoError(t, err)
				assert.Equal(t, sc.expected, engine)
			})
		}
	})
}