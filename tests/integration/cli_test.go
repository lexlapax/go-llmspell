// ABOUTME: Integration tests for the llmspell CLI, testing script execution and engine registration.
// ABOUTME: Verifies end-to-end functionality of the spell runner with real scripts.

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIBasicExecution tests basic script execution through the CLI
func TestCLIBasicExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the CLI binary
	binPath := buildCLI(t)

	t.Run("execute_lua_script", func(t *testing.T) {
		// Create a simple Lua script
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "hello.lua")
		scriptContent := `return "Hello from Lua!"`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		// Execute the script
		cmd := exec.Command(binPath, "run", scriptPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "Hello from Lua!")
	})

	t.Run("execute_with_parameters", func(t *testing.T) {
		// Create a Lua script that uses parameters
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "params.lua")
		scriptContent := `
-- Access parameters passed from CLI
if name then
    return "Hello, " .. name .. "!"
else
    return "Hello, stranger!"
end
`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		// Execute with parameters
		cmd := exec.Command(binPath, "run", scriptPath, "-p", "name=World")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "Hello, World!")
	})

	t.Run("execute_with_timeout", func(t *testing.T) {
		// Create a long-running script
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "slow.lua")
		scriptContent := `
io.write("Starting...\n")
io.flush()
-- Simulate long operation
local start = os.time()
while os.time() - start < 5 do
    -- busy wait
end
return "Done!"
`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		// Execute with 1 second timeout
		cmd := exec.Command(binPath, "run", scriptPath, "-t", "1")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		start := time.Now()
		err = cmd.Run()
		duration := time.Since(start)

		// Should timeout
		assert.Error(t, err)
		assert.Less(t, duration, 3*time.Second, "Should timeout quickly")
		// The script might not output anything if it times out immediately
		// Just verify it timed out quickly
		_ = stdout.String() + stderr.String() // Combined output for debugging if needed
	})
}

// TestCLIEngineRegistration tests that engines are properly registered
func TestCLIEngineRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildCLI(t)

	t.Run("list_engines", func(t *testing.T) {
		cmd := exec.Command(binPath, "engines")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "lua", "Should list Lua engine")
	})

	t.Run("list_engines_detailed", func(t *testing.T) {
		cmd := exec.Command(binPath, "engines", "--details")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		output := stdout.String()
		assert.Contains(t, output, "lua", "Should list Lua engine")
		assert.Contains(t, output, "Lua 5.1 scripting engine", "Should show description")
	})

	t.Run("execute_with_explicit_engine", func(t *testing.T) {
		// Create a script
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "test.script") // Generic extension
		scriptContent := `return "Executed with explicit engine"`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		// Execute with explicit Lua engine
		cmd := exec.Command(binPath, "run", scriptPath, "-e", "lua")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "Executed with explicit engine")
	})
}

// TestCLIValidation tests script validation
func TestCLIValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildCLI(t)

	t.Run("validate_lua_script", func(t *testing.T) {
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "valid.lua")
		scriptContent := `return "Valid Lua"`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "validate", scriptPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		assert.Contains(t, stdout.String(), "✓", "Should show success")
		assert.Contains(t, stdout.String(), "lua", "Should identify Lua engine")
	})

	t.Run("validate_spell_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		spellPath := filepath.Join(tmpDir, "spell.yaml")
		spellContent := `
name: test-spell
version: 1.0.0
description: Test spell for validation
entry_point: main.lua
engine: lua
`
		err := os.WriteFile(spellPath, []byte(spellContent), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "validate", spellPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err, "stderr: %s", stderr.String())
		output := stdout.String()
		assert.Contains(t, output, "✓", "Should show success")
		assert.Contains(t, output, "test-spell", "Should show spell name")
		assert.Contains(t, output, "1.0.0", "Should show version")
	})
}

// TestCLIErrorHandling tests error handling and reporting
func TestCLIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildCLI(t)

	t.Run("script_not_found", func(t *testing.T) {
		cmd := exec.Command(binPath, "run", "/nonexistent/script.lua")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		assert.Error(t, err)
		// Should have user-friendly error message
		output := stderr.String() + stdout.String()
		assert.NotEmpty(t, output)
	})

	t.Run("invalid_engine", func(t *testing.T) {
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(scriptPath, []byte("test"), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "run", scriptPath, "-e", "nonexistent")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.Error(t, err)
		output := stderr.String() + stdout.String()
		assert.Contains(t, strings.ToLower(output), "engine")
	})
}

// TestCLIGlobalFlags tests global flag handling
func TestCLIGlobalFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	binPath := buildCLI(t)

	t.Run("debug_mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		scriptPath := filepath.Join(tmpDir, "debug.lua")
		scriptContent := `return "Debug test"`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "--debug", "run", scriptPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		assert.NoError(t, err)
		// In debug mode, might have additional output
		assert.Contains(t, stdout.String(), "Debug test")
	})

	t.Run("quiet_mode", func(t *testing.T) {
		cmd := exec.Command(binPath, "--quiet", "engines")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		assert.NoError(t, err)
		// In quiet mode, should still show essential output
		assert.Contains(t, stdout.String(), "lua")
	})
}

// buildCLI builds the llmspell CLI binary for testing
func buildCLI(t *testing.T) string {
	t.Helper()

	// Check if we're in the right directory
	_, err := os.Stat("go.mod")
	if err != nil {
		// Try to change to project root
		err = os.Chdir("../..")
		require.NoError(t, err, "Failed to change to project root")
	}

	// Build the binary
	binPath := filepath.Join(t.TempDir(), "llmspell")
	if os.Getenv("GOOS") == "windows" {
		binPath += ".exe"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./cmd/llmspell")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	require.NoError(t, err, "Failed to build CLI: %s", stderr.String())

	// Verify binary exists
	_, err = os.Stat(binPath)
	require.NoError(t, err, "Binary not found after build")

	return binPath
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Change to project root if needed
	if _, err := os.Stat("go.mod"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			panic("Failed to find project root")
		}
	}

	os.Exit(m.Run())
}
