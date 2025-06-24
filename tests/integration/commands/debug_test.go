// ABOUTME: Integration tests for the debug command verifying debug mode execution.
// ABOUTME: Tests debug mode script execution with appropriate debug output.

package commands

import (
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestDebugCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("debug basic execution", func(t *testing.T) {
		script := h.CreateSpell("debug-test.lua", `
			local x = 1
			x = x + 1
			return {
				message = "Processing complete",
				value = x
			}
		`)

		stdout, stderr, err := h.RunCommand("debug", script)

		h.AssertSuccess(stdout, stderr, err)
		// Check debug session markers
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		h.AssertOutput(stderr, "=== Debug Session Complete ===")
		// The debug command should at least succeed and show debug markers
		output := stdout + stderr
		assert.Contains(t, output, "Debug Session")
	})

	t.Run("debug with engine specification", func(t *testing.T) {
		script := h.CreateSpell("engine-test.lua", `
			return "engine test complete"
		`)

		stdout, stderr, err := h.RunCommand("debug", "--engine", "lua", script)

		h.AssertSuccess(stdout, stderr, err)
		// Should show lua engine in debug info
		h.AssertOutput(stderr, "Engine: lua")
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		h.AssertOutput(stderr, "=== Debug Session Complete ===")
	})

	t.Run("debug with timeout", func(t *testing.T) {
		script := h.CreateSpell("timeout-test.lua", `
			return "timeout test complete"
		`)

		stdout, stderr, err := h.RunCommand("debug", "--timeout", "10", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		h.AssertOutput(stderr, "=== Debug Session Complete ===")
	})

	t.Run("debug with step mode", func(t *testing.T) {
		script := h.CreateSpell("step-test.lua", `
			return "step test complete"
		`)

		stdout, stderr, err := h.RunCommand("debug", "--step-mode", script)

		h.AssertSuccess(stdout, stderr, err)
		// Should show step mode in debug info
		h.AssertOutput(stderr, "Mode: Step-by-step")
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		h.AssertOutput(stderr, "=== Debug Session Complete ===")
	})

	t.Run("debug with breakpoints", func(t *testing.T) {
		script := h.CreateSpell("breakpoint-test.lua", `
			return "breakpoint test complete"
		`)

		stdout, stderr, err := h.RunCommand("debug", "-b", "2", "-b", "4", script)

		h.AssertSuccess(stdout, stderr, err)
		// Should show breakpoints in debug info
		h.AssertOutput(stderr, "Breakpoints at lines: [2 4]")
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		h.AssertOutput(stderr, "=== Debug Session Complete ===")
	})
}

func TestDebugCommandErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("debug non-existent file", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("debug", "non-existent.lua")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "no such file")
	})

	t.Run("debug invalid script", func(t *testing.T) {
		script := h.CreateSpell("invalid.lua", `
			-- Syntax error
			function missing_end()
				print("no end")
		`)

		stdout, stderr, err := h.RunCommand("debug", script)

		h.AssertFailure(stdout, stderr, err)
		// Should show debug session start even for invalid script
		h.AssertOutput(stderr, "=== Debug Session Started ===")
		// Should contain error information
		output := stdout + stderr
		assert.Contains(t, output, "error") // Some form of error indication
	})

	t.Run("debug with invalid timeout", func(t *testing.T) {
		script := h.CreateSpell("small.lua", `return "complete"`)

		stdout, stderr, err := h.RunCommand("debug", "--timeout", "0", script)

		// Should handle gracefully or succeed with minimal timeout
		if err != nil {
			// If it fails, should be a reasonable error
			output := stdout + stderr
			assert.Contains(t, output, "timeout")
		} else {
			// If it succeeds, should have debug markers
			h.AssertOutput(stderr, "=== Debug Session Started ===")
		}
	})
}