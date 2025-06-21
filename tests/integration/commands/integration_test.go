// ABOUTME: Integration tests for cross-command functionality and complex scenarios.
// ABOUTME: Tests command interactions, configuration layering, and edge cases.

package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrossCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("validate then run", func(t *testing.T) {
		script := h.CreateSpell("test.lua", helpers.BasicLuaSpell())

		// First validate
		stdout1, stderr1, err1 := h.RunCommand("validate", script)
		h.AssertSuccess(stdout1, stderr1, err1)
		h.AssertOutput(stdout1, "valid")

		// Then run
		stdout2, stderr2, err2 := h.RunCommand("run", script)
		h.AssertSuccess(stdout2, stderr2, err2)
		h.AssertOutput(stdout2, "Hello from Lua!")
	})

	t.Run("new spell then validate and run", func(t *testing.T) {
		// Create a new spell
		stdout1, stderr1, err1 := h.RunCommand("new", "myspell", "--type", "basic")
		h.AssertSuccess(stdout1, stderr1, err1)

		spellPath := filepath.Join(h.TempDir(), "myspell", "main.lua")

		// Validate it
		stdout2, stderr2, err2 := h.RunCommand("validate", spellPath)
		h.AssertSuccess(stdout2, stderr2, err2)

		// Run it
		stdout3, stderr3, err3 := h.RunCommand("run", spellPath)
		h.AssertSuccess(stdout3, stderr3, err3)
	})

	t.Run("config affects run behavior", func(t *testing.T) {
		// Create config with custom timeout
		config := h.CreateConfigFile(`
engine:
  timeout: 2
`)

		// Create a slow script
		script := h.CreateSpell("slow.lua", `
			local start = os.time()
			while os.time() - start < 5 do
				-- Wait
			end
			print("Should not reach here")
		`)

		// Run with config (should timeout)
		stdout, stderr, err := h.RunCommand("run", script, "--config", config)
		h.AssertFailure(stdout, stderr, err)
		output := stdout + stderr
		assert.Contains(t, output, "timeout")
	})

	t.Run("security profile affects validation", func(t *testing.T) {
		// Create risky script
		script := h.CreateSpell("risky.lua", `
			os.execute("ls")
			io.popen("whoami")
		`)

		// Validate with sandbox profile
		stdout1, stderr1, _ := h.RunCommand("validate", script, "--profile", "sandbox")
		output1 := stdout1 + stderr1
		assert.Contains(t, output1, "security") // Should warn

		// Validate with development profile
		_, _, _ = h.RunCommand("validate", script, "--profile", "development")
		// Development might be more permissive
	})
}

func TestConfigurationLayering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("environment overrides file", func(t *testing.T) {
		// Create config file
		config := h.CreateConfigFile(`
engine:
  default: lua
  timeout: 30
repl:
  prompt: "file> "
`)

		// Set environment variables
		h.SetEnv("LLMSPELL_ENGINE_TIMEOUT", "60")
		h.SetEnv("LLMSPELL_REPL_PROMPT", "env> ")

		// Check layered values
		stdout1, _, err1 := h.RunCommand("config", "get", "engine.timeout", "--config", config)
		require.NoError(t, err1)
		assert.Contains(t, stdout1, "60") // Env override

		stdout2, _, err2 := h.RunCommand("config", "get", "repl.prompt", "--config", config)
		require.NoError(t, err2)
		assert.Contains(t, stdout2, "env>") // Env override

		stdout3, _, err3 := h.RunCommand("config", "get", "engine.default", "--config", config)
		require.NoError(t, err3)
		assert.Contains(t, stdout3, "lua") // File value (no env override)
	})

	t.Run("command flags override everything", func(t *testing.T) {
		config := h.CreateConfigFile(`
security:
  profile: sandbox
`)
		h.SetEnv("LLMSPELL_SECURITY_PROFILE", "development")

		script := h.CreateSpell("test.lua", `print("test")`)

		// Run with explicit profile flag
		stdout, stderr, err := h.RunCommand("run", script,
			"--config", config,
			"--profile", "production")

		h.AssertSuccess(stdout, stderr, err)
		// Would use production profile from flag
	})
}

func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("graceful error messages", func(t *testing.T) {
		// Various error scenarios
		scenarios := []struct {
			name   string
			args   []string
			errMsg string
		}{
			{
				name:   "invalid command",
				args:   []string{"invalidcmd"},
				errMsg: "unknown command",
			},
			{
				name:   "missing required arg",
				args:   []string{"run"},
				errMsg: "expected",
			},
			{
				name:   "invalid flag value",
				args:   []string{"run", "test.lua", "--timeout", "not-a-number"},
				errMsg: "invalid",
			},
			{
				name:   "conflicting flags",
				args:   []string{"config", "set", "key", "value", "--dry-run", "--force"},
				errMsg: "cannot use",
			},
		}

		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				stdout, stderr, err := h.RunCommand(sc.args...)
				assert.Error(t, err)
				output := stdout + stderr
				assert.Contains(t, output, sc.errMsg)
			})
		}
	})

	t.Run("debug mode error details", func(t *testing.T) {
		script := h.CreateSpell("error.lua", `
			error("Something went wrong")
		`)

		// Normal mode
		stdout1, stderr1, _ := h.RunCommand("run", script)
		output1 := stdout1 + stderr1
		assert.Contains(t, output1, "Something went wrong")

		// Debug mode
		stdout2, stderr2, _ := h.RunCommand("run", script, "--debug")
		output2 := stdout2 + stderr2
		assert.Contains(t, output2, "Something went wrong")
		assert.Contains(t, output2, "stack") // Should include stack trace
	})
}

func TestSignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip on Windows where signal handling is different
	if runtime.GOOS == "windows" {
		t.Skip("Skipping signal tests on Windows")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("graceful shutdown on interrupt", func(t *testing.T) {
		script := h.CreateSpell("long-running.lua", `
			print("Started")
			for i = 1, 100 do
				print("Working " .. i)
				os.execute("sleep 0.1")
			end
			print("Should not reach here")
		`)

		// Start the command
		cmd := h.StartCommand("run", script)

		// Wait a bit for it to start
		time.Sleep(500 * time.Millisecond)

		// Send interrupt signal
		cmd.Process.Signal(os.Interrupt)

		// Wait for graceful shutdown
		stdout, stderr, err := h.WaitCommand(cmd)

		// Should have been interrupted
		assert.Error(t, err)
		output := stdout + stderr
		assert.Contains(t, output, "Started")
		assert.Contains(t, output, "Working")
		assert.NotContains(t, output, "Should not reach here")
	})
}

func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("startup time", func(t *testing.T) {
		// Measure how long it takes to show help
		start := time.Now()
		stdout, stderr, err := h.RunCommand("--help")
		duration := time.Since(start)

		h.AssertSuccess(stdout, stderr, err)

		// Should start quickly
		assert.Less(t, duration, 500*time.Millisecond, "Startup took too long: %v", duration)
	})

	t.Run("script execution overhead", func(t *testing.T) {
		script := h.CreateSpell("minimal.lua", `print("test")`)

		// Run multiple times and measure
		var totalDuration time.Duration
		runs := 5

		for i := 0; i < runs; i++ {
			start := time.Now()
			stdout, stderr, err := h.RunCommand("run", script)
			duration := time.Since(start)

			h.AssertSuccess(stdout, stderr, err)
			totalDuration += duration
		}

		avgDuration := totalDuration / time.Duration(runs)
		assert.Less(t, avgDuration, 200*time.Millisecond, "Average execution too slow: %v", avgDuration)
	})
}

func TestCrossPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("path handling", func(t *testing.T) {
		// Create nested directory structure
		deepPath := filepath.Join(h.TempDir(), "a", "b", "c", "d")
		require.NoError(t, os.MkdirAll(deepPath, 0755))

		script := filepath.Join(deepPath, "test.lua")
		h.CreateFile(script, `print("Deep path test")`)

		// Should handle deep paths
		stdout, stderr, err := h.RunCommand("run", script)
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Deep path test")
	})

	t.Run("file permissions", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		script := h.CreateSpell("readonly.lua", `print("test")`)

		// Make read-only
		require.NoError(t, os.Chmod(script, 0444))

		// Should still be able to run
		stdout, stderr, err := h.RunCommand("run", script)
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "test")
	})

	t.Run("unicode handling", func(t *testing.T) {
		script := h.CreateSpell("unicode.lua", `
			print("Hello 世界")
			print("Émojis: 🚀 🌟 ✨")
			local msg = "Ñiño José"
			print(msg)
		`)

		stdout, stderr, err := h.RunCommand("run", script)
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello 世界")
		h.AssertOutput(stdout, "Émojis: 🚀 🌟 ✨")
		h.AssertOutput(stdout, "Ñiño José")
	})
}

func TestComplexScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("multi-file spell", func(t *testing.T) {
		// Create a complex spell structure
		spellDir := filepath.Join(h.TempDir(), "complex-spell")
		h.CreateSpellYAML(spellDir, `
name: complex-spell
version: 1.0.0
engine: lua
main: src/main.lua
parameters:
  count:
    type: integer
    default: 3
`)

		// Create source files
		srcDir := filepath.Join(spellDir, "src")
		require.NoError(t, os.MkdirAll(srcDir, 0755))

		// Main file
		h.CreateFile(filepath.Join(srcDir, "main.lua"), `
			-- Load utility module
			package.path = package.path .. ";src/?.lua"
			local utils = require("utils")
			
			local count = params.count or 3
			for i = 1, count do
				print(utils.format_message(i))
			end
		`)

		// Utility module
		h.CreateFile(filepath.Join(srcDir, "utils.lua"), `
			local M = {}
			
			function M.format_message(n)
				return string.format("Message #%d from complex spell", n)
			end
			
			return M
		`)

		// Run the complex spell
		stdout, stderr, err := h.RunCommand("run", spellDir)
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Message #1 from complex spell")
		h.AssertOutput(stdout, "Message #2 from complex spell")
		h.AssertOutput(stdout, "Message #3 from complex spell")
	})

	t.Run("pipeline of commands", func(t *testing.T) {
		// Create template, modify it, validate, and run

		// Step 1: Create from template
		stdout1, stderr1, err1 := h.RunCommand("new", "pipeline-spell", "--type", "basic")
		h.AssertSuccess(stdout1, stderr1, err1)

		// Step 2: Modify the script
		scriptPath := filepath.Join(h.TempDir(), "pipeline-spell", "main.lua")
		content, err := os.ReadFile(scriptPath)
		require.NoError(t, err)

		// Add custom logic
		newContent := string(content) + `
-- Custom addition
print("Pipeline test successful!")
`
		require.NoError(t, os.WriteFile(scriptPath, []byte(newContent), 0644))

		// Step 3: Validate modified script
		stdout3, stderr3, err3 := h.RunCommand("validate", scriptPath)
		h.AssertSuccess(stdout3, stderr3, err3)

		// Step 4: Run with parameters
		stdout4, stderr4, err4 := h.RunCommand("run", scriptPath, "--param", "message=Pipeline")
		h.AssertSuccess(stdout4, stderr4, err4)
		h.AssertOutput(stdout4, "Pipeline test successful!")
	})
}
