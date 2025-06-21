// ABOUTME: Integration tests for the config command verifying configuration management.
// ABOUTME: Tests config viewing, setting, and file operations.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("view default config", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "view")

		h.AssertSuccess(stdout, stderr, err)
		// Should show default configuration
		h.AssertOutput(stdout, "engine:")
		h.AssertOutput(stdout, "security:")
		h.AssertOutput(stdout, "repl:")
	})

	t.Run("get specific config value", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "get", "engine.default")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "lua") // Default engine
	})

	t.Run("set config value", func(t *testing.T) {
		// Create a test config file
		configFile := filepath.Join(h.TempDir(), "test-config.yaml")

		// Set a value
		stdout, stderr, err := h.RunCommand("config", "set", "engine.timeout", "30", "--config", configFile)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Configuration updated")

		// Verify it was set
		stdout2, stderr2, err2 := h.RunCommand("config", "get", "engine.timeout", "--config", configFile)
		h.AssertSuccess(stdout2, stderr2, err2)
		assert.Contains(t, stdout2, "30")
	})

	t.Run("init config file", func(t *testing.T) {
		configFile := filepath.Join(h.TempDir(), "new-config.yaml")

		stdout, stderr, err := h.RunCommand("config", "init", "--config", configFile)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Configuration file created")

		// Check file exists
		assert.FileExists(t, configFile)

		// Verify it's valid YAML
		content, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "engine:")
		assert.Contains(t, string(content), "security:")
	})

	t.Run("validate config file", func(t *testing.T) {
		// Create valid config
		validConfig := h.CreateConfigFile(`
engine:
  default: lua
  timeout: 60
security:
  profile: sandbox
`)

		stdout, stderr, err := h.RunCommand("config", "validate", "--config", validConfig)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("validate invalid config", func(t *testing.T) {
		// Create invalid config
		invalidConfig := h.CreateConfigFile(`
engine:
  default: python  # Invalid engine
  timeout: -1      # Invalid timeout
`)

		stdout, stderr, err := h.RunCommand("config", "validate", "--config", invalidConfig)

		h.AssertFailure(stdout, stderr, err)
		output := stdout + stderr
		assert.Contains(t, output, "invalid")
	})

	t.Run("config with environment variables", func(t *testing.T) {
		h.SetEnv("LLMSPELL_ENGINE_DEFAULT", "lua")
		h.SetEnv("LLMSPELL_ENGINE_TIMEOUT", "120")

		stdout, stderr, err := h.RunCommand("config", "get", "engine.timeout")

		h.AssertSuccess(stdout, stderr, err)
		assert.Contains(t, stdout, "120") // Should pick up env var
	})

	t.Run("config layering", func(t *testing.T) {
		// Create base config
		baseConfig := h.CreateConfigFile(`
engine:
  default: lua
  timeout: 30
repl:
  prompt: "base> "
`)

		// Set environment variable (higher priority)
		h.SetEnv("LLMSPELL_ENGINE_TIMEOUT", "60")

		// Get value with layering
		stdout, stderr, err := h.RunCommand("config", "get", "engine.timeout", "--config", baseConfig)

		h.AssertSuccess(stdout, stderr, err)
		assert.Contains(t, stdout, "60") // Env var should override file

		// Get value not in env
		stdout2, stderr2, err2 := h.RunCommand("config", "get", "repl.prompt", "--config", baseConfig)
		h.AssertSuccess(stdout2, stderr2, err2)
		assert.Contains(t, stdout2, "base>") // Should use file value
	})

	t.Run("reset config value", func(t *testing.T) {
		configFile := filepath.Join(h.TempDir(), "reset-config.yaml")

		// Set a value
		_, _, err := h.RunCommand("config", "set", "engine.timeout", "90", "--config", configFile)
		require.NoError(t, err)

		// Reset it
		stdout, stderr, err := h.RunCommand("config", "reset", "engine.timeout", "--config", configFile)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "reset to default")

		// Verify it's back to default
		stdout2, stderr2, err2 := h.RunCommand("config", "get", "engine.timeout", "--config", configFile)
		h.AssertSuccess(stdout2, stderr2, err2)
		assert.Contains(t, stdout2, "60") // Default value
	})

	t.Run("list config keys", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "list")

		h.AssertSuccess(stdout, stderr, err)
		// Should list all available config keys
		h.AssertOutput(stdout, "engine.default")
		h.AssertOutput(stdout, "engine.timeout")
		h.AssertOutput(stdout, "security.profile")
		h.AssertOutput(stdout, "repl.prompt")
		h.AssertOutput(stdout, "repl.history_file")
	})

	t.Run("export config", func(t *testing.T) {
		exportFile := filepath.Join(h.TempDir(), "exported.yaml")

		stdout, stderr, err := h.RunCommand("config", "export", exportFile)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Configuration exported")

		// Check exported file
		assert.FileExists(t, exportFile)
		content, err := os.ReadFile(exportFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "engine:")
	})

	t.Run("import config", func(t *testing.T) {
		// Create config to import
		importConfig := h.CreateConfigFile(`
engine:
  default: lua
  timeout: 45
custom:
  setting: value
`)

		targetConfig := filepath.Join(h.TempDir(), "target-config.yaml")

		stdout, stderr, err := h.RunCommand("config", "import", importConfig, "--config", targetConfig)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Configuration imported")

		// Verify imported values
		stdout2, stderr2, err2 := h.RunCommand("config", "get", "engine.timeout", "--config", targetConfig)
		h.AssertSuccess(stdout2, stderr2, err2)
		assert.Contains(t, stdout2, "45")
	})
}

func TestConfigCommandErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("get non-existent key", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "get", "non.existent.key")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "not found")
	})

	t.Run("set invalid value", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "set", "engine.timeout", "not-a-number")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "invalid")
	})

	t.Run("import non-existent file", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "import", "/non/existent/file.yaml")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "no such file")
	})

	t.Run("export to invalid path", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("config", "export", "/invalid/path/export.yaml")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "cannot create")
	})
}
