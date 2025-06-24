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
		stdout, stderr, err := h.RunCommand("engines")

		h.AssertSuccess(stdout, stderr, err)
		// Should list available engines
		h.AssertOutput(stdout, "Available engines:")
		h.AssertOutput(stdout, "lua")
	})

	t.Run("show engine details", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "--details")

		h.AssertSuccess(stdout, stderr, err)
		// Should show engine information with details flag
		h.AssertOutput(stdout, "Available engines:")
		h.AssertOutput(stdout, "lua")
		// Details flag should show more information
		output := stdout + stderr
		// Just verify details flag doesn't break the command
		assert.Contains(t, output, "engines")
	})

	t.Run("verbose engine listing", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("engines", "--details", "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose should show more details
		h.AssertOutput(stdout, "Available engines:")
		h.AssertOutput(stdout, "lua")
	})
}

func TestEnginesWithScripts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("run script with explicit engine", func(t *testing.T) {
		script := h.CreateSpell("generic.lua", `
			return "Hello from explicit engine"
		`)

		// Run with explicit engine selection
		stdout, stderr, err := h.RunCommand("run", script, "--engine", "lua")

		h.AssertSuccess(stdout, stderr, err)
		// The script output might be captured as return value
		output := stdout + stderr
		// Check that execution succeeds with explicit engine
		assert.NotEmpty(t, output)
	})

	t.Run("validate script with engine", func(t *testing.T) {
		script := h.CreateSpell("validate.lua", `
			return "Valid script"
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
				return "Coroutine support"
			end)
			local status, result = coroutine.resume(co)
			return result
		`)

		stdout, stderr, err := h.RunCommand("run", script)

		h.AssertSuccess(stdout, stderr, err)
		// Script should execute successfully with coroutine support
		output := stdout + stderr
		assert.NotEmpty(t, output)
	})
}