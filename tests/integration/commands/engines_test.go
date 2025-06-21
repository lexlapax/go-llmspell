// ABOUTME: Integration tests for the engines command verifying engine listing and info.
// ABOUTME: Tests engine discovery, capabilities display, and version information.

package commands

import (
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestEnginesCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("list all engines", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "list")

		h.AssertSuccess(stdout, stderr, err)
		// Should list available engines
		h.AssertOutput(stdout, "lua")
		h.AssertOutput(stdout, "GopherLua")
		// Future engines would appear here:
		// h.AssertOutput(stdout, "javascript")
		// h.AssertOutput(stdout, "tengo")
	})

	t.Run("show engine details", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "info", "lua")

		h.AssertSuccess(stdout, stderr, err)
		// Should show engine information
		h.AssertOutput(stdout, "lua")
		h.AssertOutput(stdout, "GopherLua")
		h.AssertOutput(stdout, "version:")
		h.AssertOutput(stdout, "features:")
	})

	t.Run("verbose engine listing", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "list", "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose should show more details
		h.AssertOutput(stdout, "lua")
		h.AssertOutput(stdout, "version")
		h.AssertOutput(stdout, "features")
		h.AssertOutput(stdout, "file extensions")
	})

	t.Run("check engine capabilities", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "capabilities", "lua")

		h.AssertSuccess(stdout, stderr, err)
		// Should list what the engine can do
		h.AssertOutput(stdout, "async")
		h.AssertOutput(stdout, "debug")
		h.AssertOutput(stdout, "validate")
	})

	t.Run("invalid engine name", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "info", "python")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "unknown engine")
	})

	t.Run("compare engines", func(t *testing.T) {
		// When we have multiple engines, this would compare them
		stdout, stderr, err := h.RunCommand("engines", "compare", "lua", "lua")

		// For now, comparing the same engine
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "lua")
	})

	t.Run("engine benchmarks", func(t *testing.T) {
		t.Skip("Benchmarks take too long for integration tests")

		stdout, stderr, err := h.RunCommand("engines", "benchmark", "lua")

		h.AssertSuccess(stdout, stderr, err)
		// Would show performance metrics
		h.AssertOutput(stdout, "operations/sec")
		h.AssertOutput(stdout, "memory usage")
	})

	t.Run("detect engine for file", func(t *testing.T) {
		// Create test files
		luaFile := h.CreateSpell("test.lua", `print("lua")`)
		jsFile := h.CreateSpell("test.js", `console.log("js")`)
		unknownFile := h.CreateSpell("test.xyz", `unknown`)

		// Test Lua detection
		stdout, stderr, err := h.RunCommand("engines", "detect", luaFile)
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "lua")

		// Test JS detection (would work when JS engine is added)
		stdout2, stderr2, err2 := h.RunCommand("engines", "detect", jsFile)
		// Currently fails as JS not implemented
		assert.Error(t, err2)
		output := stdout2 + stderr2
		assert.Contains(t, output, "no engine")

		// Test unknown extension
		stdout3, stderr3, err3 := h.RunCommand("engines", "detect", unknownFile)
		h.AssertFailure(stdout3, stderr3, err3)
		assert.Contains(t, stderr3, "cannot detect")
	})

	t.Run("engine health check", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "check", "lua")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "healthy")
		h.AssertOutput(stdout, "lua")
	})

	t.Run("JSON output format", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "list", "--format", "json")

		h.AssertSuccess(stdout, stderr, err)
		// Should output valid JSON
		assert.Contains(t, stdout, "{")
		assert.Contains(t, stdout, "\"name\"")
		assert.Contains(t, stdout, "\"lua\"")
		assert.Contains(t, stdout, "}")
	})

	t.Run("YAML output format", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "list", "--format", "yaml")

		h.AssertSuccess(stdout, stderr, err)
		// Should output valid YAML
		assert.Contains(t, stdout, "- name: lua")
		assert.Contains(t, stdout, "  version:")
		assert.Contains(t, stdout, "  features:")
	})
}

func TestEnginesWithScripts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("run script with explicit engine", func(t *testing.T) {
		script := h.CreateSpell("generic.script", `
			print("Hello from explicit engine")
		`)

		// Run with explicit engine selection
		stdout, stderr, err := h.RunCommand("run", script, "--engine", "lua")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello from explicit engine")
	})

	t.Run("validate script with engine", func(t *testing.T) {
		script := h.CreateSpell("validate.script", `
			print("Valid script")
		`)

		// Validate with explicit engine
		stdout, stderr, err := h.RunCommand("validate", script, "--engine", "lua")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("engine-specific features", func(t *testing.T) {
		script := h.CreateSpell("features.lua", `
			-- Lua-specific features
			local co = coroutine.create(function()
				print("Coroutine support")
			end)
			coroutine.resume(co)
		`)

		stdout, stderr, err := h.RunCommand("run", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Coroutine support")
	})
}
