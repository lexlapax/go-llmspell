// ABOUTME: Tests for the configuration loader, covering file loading, environment variables, and flag merging.
// ABOUTME: Ensures proper configuration layering and validation in the Koanf v2-based loader.

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("default_options", func(t *testing.T) {
		loader := NewLoader(LoaderOptions{})

		assert.Equal(t, "config", loader.options.ConfigName)
		assert.Equal(t, "yaml", loader.options.ConfigFormat)
		assert.Equal(t, "LLMSPELL_", loader.options.EnvPrefix)
		assert.Equal(t, "_", loader.options.EnvDelimiter)
		assert.NotEmpty(t, loader.options.ConfigPaths)
	})

	t.Run("custom_options", func(t *testing.T) {
		options := LoaderOptions{
			ConfigName:   "custom",
			ConfigFormat: "yml",
			EnvPrefix:    "CUSTOM_",
			EnvDelimiter: "__",
			ConfigPaths:  []string{"/custom/path"},
		}

		loader := NewLoader(options)

		assert.Equal(t, "custom", loader.options.ConfigName)
		assert.Equal(t, "yml", loader.options.ConfigFormat)
		assert.Equal(t, "CUSTOM_", loader.options.EnvPrefix)
		assert.Equal(t, "__", loader.options.EnvDelimiter)
		assert.Equal(t, []string{"/custom/path"}, loader.options.ConfigPaths)
	})
}

func TestLoader_LoadConfig(t *testing.T) {
	t.Run("load_defaults_only", func(t *testing.T) {
		// Create loader with non-existent config paths
		loader := NewLoader(LoaderOptions{
			ConfigPaths: []string{"/nonexistent/path"},
		})

		config, err := loader.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// Should match default config
		defaultConfig := GetDefaultConfig()
		assert.Equal(t, defaultConfig.Version, config.Version)
		assert.Equal(t, defaultConfig.Engine.Default, config.Engine.Default)
		assert.Equal(t, defaultConfig.Security.Profile, config.Security.Profile)
	})

	t.Run("load_with_validation", func(t *testing.T) {
		loader := NewLoader(LoaderOptions{
			ConfigPaths:    []string{"/nonexistent/path"},
			ValidateOnLoad: true,
		})

		config, err := loader.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, config)
	})
}

func TestLoader_LoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
version: "test-version"
debug: true
engine:
  default: "javascript"
  memory_limit: 128000000
security:
  profile: "development"
  network_access: true
logging:
  level: "debug"
  format: "json"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	t.Run("load_from_valid_file", func(t *testing.T) {
		loader := NewLoader(LoaderOptions{})

		config, err := loader.LoadConfigFromFile(configFile)
		require.NoError(t, err)
		require.NotNil(t, config)

		// Check that file values override defaults
		assert.Equal(t, "test-version", config.Version)
		assert.True(t, config.Debug)
		assert.Equal(t, "javascript", config.Engine.Default)
		assert.Equal(t, int64(128000000), config.Engine.MemoryLimit)
		assert.Equal(t, "development", config.Security.Profile)
		assert.True(t, config.Security.NetworkAccess)
		assert.Equal(t, "debug", config.Logging.Level)
		assert.Equal(t, "json", config.Logging.Format)

		// Check that non-overridden values remain defaults
		assert.Equal(t, 30*time.Second, config.Engine.TimeoutLimit)
		assert.False(t, config.Quiet) // Not set in file, should be default
	})

	t.Run("load_from_nonexistent_file", func(t *testing.T) {
		loader := NewLoader(LoaderOptions{})

		_, err := loader.LoadConfigFromFile("/nonexistent/file.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config file")
	})
}

func TestLoader_MergeFlags(t *testing.T) {
	loader := NewLoader(LoaderOptions{
		ConfigPaths: []string{"/nonexistent/path"},
	})

	// Load base config first
	_, err := loader.LoadConfig()
	require.NoError(t, err)

	t.Run("merge_simple_flags", func(t *testing.T) {
		flags := map[string]interface{}{
			"debug":   true,
			"verbose": true,
			"quiet":   false,
		}

		err := loader.MergeFlags(flags)
		require.NoError(t, err)

		config, err := loader.GetCurrentConfig()
		require.NoError(t, err)

		assert.True(t, config.Debug)
		assert.True(t, config.Verbose)
		assert.False(t, config.Quiet)
	})

	t.Run("merge_nested_flags", func(t *testing.T) {
		flags := map[string]interface{}{
			"engine.default":      "tengo",
			"engine.memory_limit": 256000000,
			"security.profile":    "production",
			"logging.level":       "error",
		}

		err := loader.MergeFlags(flags)
		require.NoError(t, err)

		config, err := loader.GetCurrentConfig()
		require.NoError(t, err)

		assert.Equal(t, "tengo", config.Engine.Default)
		assert.Equal(t, int64(256000000), config.Engine.MemoryLimit)
		assert.Equal(t, "production", config.Security.Profile)
		assert.Equal(t, "error", config.Logging.Level)
	})
}

func TestLoader_EnvironmentVariables(t *testing.T) {
	// Save original env vars
	originalDebug := os.Getenv("LLMSPELL_DEBUG")
	originalEngine := os.Getenv("LLMSPELL_ENGINE_DEFAULT")
	originalProfile := os.Getenv("LLMSPELL_SECURITY_PROFILE")

	// Clean up after test
	defer func() {
		if originalDebug == "" {
			_ = os.Unsetenv("LLMSPELL_DEBUG")
		} else {
			_ = os.Setenv("LLMSPELL_DEBUG", originalDebug)
		}
		if originalEngine == "" {
			_ = os.Unsetenv("LLMSPELL_ENGINE_DEFAULT")
		} else {
			_ = os.Setenv("LLMSPELL_ENGINE_DEFAULT", originalEngine)
		}
		if originalProfile == "" {
			_ = os.Unsetenv("LLMSPELL_SECURITY_PROFILE")
		} else {
			_ = os.Setenv("LLMSPELL_SECURITY_PROFILE", originalProfile)
		}
	}()

	t.Run("load_from_environment", func(t *testing.T) {
		// Set test environment variables
		require.NoError(t, os.Setenv("LLMSPELL_DEBUG", "true"))
		require.NoError(t, os.Setenv("LLMSPELL_ENGINE_DEFAULT", "javascript"))
		require.NoError(t, os.Setenv("LLMSPELL_SECURITY_PROFILE", "production"))

		loader := NewLoader(LoaderOptions{
			ConfigPaths: []string{"/nonexistent/path"},
		})

		config, err := loader.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.True(t, config.Debug)
		assert.Equal(t, "javascript", config.Engine.Default)
		assert.Equal(t, "production", config.Security.Profile)
	})
}

func TestLoader_SaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "saved_config.yaml")

	loader := NewLoader(LoaderOptions{})
	config := GetDefaultConfig()

	// Modify some values
	config.Debug = true
	config.Engine.Default = "tengo"
	config.Security.Profile = "development"

	t.Run("save_config_successfully", func(t *testing.T) {
		err := loader.SaveConfig(config, configFile)
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(configFile)
		assert.NoError(t, err)

		// Create a new loader to load the saved config and verify
		newLoader := NewLoader(LoaderOptions{})
		loadedConfig, err := newLoader.LoadConfigFromFile(configFile)
		require.NoError(t, err)

		assert.True(t, loadedConfig.Debug)
		assert.Equal(t, "tengo", loadedConfig.Engine.Default)
		assert.Equal(t, "development", loadedConfig.Security.Profile)
	})
}

func TestLoader_GetRawAndSetRaw(t *testing.T) {
	t.Run("get_and_set_raw_values", func(t *testing.T) {
		// Create isolated loader for this test
		loader := NewLoader(LoaderOptions{
			ConfigPaths: []string{"/nonexistent/path"},
		})

		// Load base config
		_, err := loader.LoadConfig()
		require.NoError(t, err)

		// Test getting existing value
		debug := loader.GetRaw("debug")
		assert.Equal(t, false, debug) // Default is false

		// Test setting value
		loader.SetRaw("debug", true)
		debug = loader.GetRaw("debug")
		assert.Equal(t, true, debug)

		// Test nested value
		loader.SetRaw("engine.default", "custom")
		engineDefault := loader.GetRaw("engine.default")
		assert.Equal(t, "custom", engineDefault)
	})

	t.Run("get_nonexistent_key", func(t *testing.T) {
		// Create isolated loader for this test
		loader := NewLoader(LoaderOptions{
			ConfigPaths: []string{"/nonexistent/path"},
		})

		// Load base config
		_, err := loader.LoadConfig()
		require.NoError(t, err)

		value := loader.GetRaw("nonexistent.key")
		assert.Nil(t, value)
	})
}

func TestLoader_Keys(t *testing.T) {
	loader := NewLoader(LoaderOptions{
		ConfigPaths: []string{"/nonexistent/path"},
	})

	// Load base config
	_, err := loader.LoadConfig()
	require.NoError(t, err)

	t.Run("get_all_keys", func(t *testing.T) {
		keys := loader.Keys()

		// Should have top-level configuration keys
		assert.Contains(t, keys, "version")
		assert.Contains(t, keys, "debug")

		// Should have nested keys (Koanf flattens all keys)
		assert.Contains(t, keys, "engine.default")
		assert.Contains(t, keys, "security.profile")
		assert.Contains(t, keys, "logging.level")
		assert.Contains(t, keys, "cli.color_output")
		assert.Contains(t, keys, "repl.prompt")
		assert.Contains(t, keys, "runner.default_timeout")
		assert.Contains(t, keys, "templates.builtin_path")

		// Verify we have a reasonable number of keys
		assert.Greater(t, len(keys), 50, "Should have many flattened configuration keys")
	})
}

func TestExpandPath(t *testing.T) {
	t.Run("expand_home_directory", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expanded := expandPath("~/test/path")
		expected := filepath.Join(homeDir, "test/path")
		assert.Equal(t, expected, expanded)
	})

	t.Run("expand_home_only", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expanded := expandPath("~")
		assert.Equal(t, homeDir, expanded)
	})

	t.Run("no_expansion_needed", func(t *testing.T) {
		path := "/absolute/path"
		expanded := expandPath(path)
		assert.Equal(t, path, expanded)

		path = "relative/path"
		expanded = expandPath(path)
		assert.Equal(t, path, expanded)
	})
}

func TestSetNestedValue(t *testing.T) {
	t.Run("set_simple_value", func(t *testing.T) {
		m := make(map[string]interface{})
		setNestedValue(m, "key", "value")

		assert.Equal(t, "value", m["key"])
	})

	t.Run("set_nested_value", func(t *testing.T) {
		m := make(map[string]interface{})
		setNestedValue(m, "level1.level2.key", "value")

		level1, ok := m["level1"].(map[string]interface{})
		require.True(t, ok)

		level2, ok := level1["level2"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "value", level2["key"])
	})

	t.Run("override_existing_value", func(t *testing.T) {
		m := map[string]interface{}{
			"level1": "existing_value",
		}

		setNestedValue(m, "level1.level2.key", "new_value")

		level1, ok := m["level1"].(map[string]interface{})
		require.True(t, ok)

		level2, ok := level1["level2"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "new_value", level2["key"])
	})
}

func TestGetDefaultConfigFile(t *testing.T) {
	t.Run("returns_reasonable_path", func(t *testing.T) {
		configFile := GetDefaultConfigFile()

		assert.NotEmpty(t, configFile)
		assert.Contains(t, configFile, "config.yaml")

		// Should be an absolute path or reasonable fallback
		if !filepath.IsAbs(configFile) {
			assert.Equal(t, "./config.yaml", configFile)
		}
	})
}

func TestCreateDefaultConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "default_config.yaml")

	t.Run("create_default_config", func(t *testing.T) {
		err := CreateDefaultConfigFile(configFile)
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(configFile)
		assert.NoError(t, err)

		// Load and verify it's a valid config
		loader := NewLoader(LoaderOptions{})
		config, err := loader.LoadConfigFromFile(configFile)
		require.NoError(t, err)

		// Should match default config
		defaultConfig := GetDefaultConfig()
		assert.Equal(t, defaultConfig.Version, config.Version)
		assert.Equal(t, defaultConfig.Engine.Default, config.Engine.Default)
	})
}

func TestGetDefaultConfigPaths(t *testing.T) {
	t.Run("returns_standard_paths", func(t *testing.T) {
		paths := getDefaultConfigPaths()

		assert.Contains(t, paths, ".")
		assert.Contains(t, paths, "~/.llmspell")
		assert.Contains(t, paths, "~/.config/llmspell")
		assert.Contains(t, paths, "/etc/llmspell")
	})
}
