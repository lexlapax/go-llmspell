// ABOUTME: Integration tests for the run command verifying script execution functionality.
// ABOUTME: Tests various scenarios including success, failure, parameters, and timeouts.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("basic script execution", func(t *testing.T) {
		script := h.CreateSpell("test.lua", helpers.BasicLuaSpell())

		stdout, stderr, err := h.RunCommand("run", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello from Lua!")
	})

	t.Run("script with parameters", func(t *testing.T) {
		script := h.CreateSpell("params.lua", `
			local message = params.message or "default"
			print("Message: " .. message)
			return message
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--param", "message=Hello World")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Message: Hello World")
	})

	t.Run("script with spell.yaml", func(t *testing.T) {
		// Create a directory with spell.yaml and script
		spellDir := filepath.Join(h.TempDir(), "myspell")
		h.CreateSpellYAML(spellDir, helpers.BasicSpellYAML())

		script := filepath.Join(spellDir, "main.lua")
		require.NoError(t, os.WriteFile(script, []byte(`
			print("Running spell: " .. spell.name)
			print("Message: " .. params.message)
		`), 0644))

		stdout, stderr, err := h.RunCommand("run", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Running spell: test-spell")
		h.AssertOutput(stdout, "Message: Hello")
	})

	t.Run("script execution failure", func(t *testing.T) {
		script := h.CreateSpell("error.lua", helpers.ErrorLuaSpell())

		stdout, stderr, err := h.RunCommand("run", script)

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "This is a test error")
	})

	t.Run("script not found", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("run", "nonexistent.lua")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "no such file")
	})

	t.Run("timeout handling", func(t *testing.T) {
		script := h.CreateSpell("timeout.lua", `
			-- Infinite loop
			while true do
				-- Do nothing
			end
		`)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		stdout, stderr, err := h.RunCommandWithContext(ctx, "run", script, "--timeout", "1")

		h.AssertFailure(stdout, stderr, err)
		// The error might be in stdout or stderr depending on how timeout is handled
		output := stdout + stderr
		assert.Contains(t, output, "timeout") // Should mention timeout
	})

	t.Run("verbose mode", func(t *testing.T) {
		script := h.CreateSpell("verbose.lua", `print("Hello")`)

		stdout, stderr, err := h.RunCommand("run", script, "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello")
		// Verbose mode should show additional info
		output := stdout + stderr
		assert.Contains(t, output, "lua") // Should mention engine
	})

	t.Run("quiet mode", func(t *testing.T) {
		script := h.CreateSpell("quiet.lua", `
			io.stderr:write("Debug info\n")
			print("Result")
			return "Done"
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--quiet")

		h.AssertSuccess(stdout, stderr, err)
		// In quiet mode, only essential output should be shown
		assert.Contains(t, stdout, "Result")
		assert.NotContains(t, stderr, "Debug info")
	})

	t.Run("dry run", func(t *testing.T) {
		script := h.CreateSpell("dryrun.lua", `
			print("This should not execute")
			return "executed"
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--dry-run")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertNotOutput(stdout, "This should not execute")
		assert.Contains(t, stdout, "Would execute") // Should indicate dry run
	})

	t.Run("watch mode", func(t *testing.T) {
		t.Skip("Watch mode requires interactive testing")
		// Watch mode would need special handling for testing
	})
}

func TestRunCommandWithDifferentEngines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("explicit engine selection", func(t *testing.T) {
		script := h.CreateSpell("test.script", `print("Hello")`)

		stdout, stderr, err := h.RunCommand("run", script, "--engine", "lua")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello")
	})

	t.Run("invalid engine", func(t *testing.T) {
		script := h.CreateSpell("test.lua", `print("Hello")`)

		stdout, stderr, err := h.RunCommand("run", script, "--engine", "python")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "python") // Should mention the invalid engine
	})
}

func TestRunCommandEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("environment variable passing", func(t *testing.T) {
		script := h.CreateSpell("env.lua", `
			local env_var = os.getenv("TEST_VAR")
			print("TEST_VAR=" .. (env_var or "not set"))
		`)

		h.SetEnv("TEST_VAR", "test_value")
		stdout, stderr, err := h.RunCommand("run", script, "--env", "CUSTOM_VAR=custom_value")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "TEST_VAR=test_value")
	})
}
