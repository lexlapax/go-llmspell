// ABOUTME: Integration tests for the debug command verifying interactive debugging features.
// ABOUTME: Tests breakpoints, stepping, variable inspection, and debug REPL.

package commands

import (
	"path/filepath"
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

	t.Run("debug with breakpoint", func(t *testing.T) {
		script := h.CreateSpell("debug-test.lua", `
			local x = 1
			print("Before breakpoint")
			-- BREAKPOINT: Line 4
			x = x + 1
			print("x = " .. x)
			print("After breakpoint")
		`)

		// Debug with breakpoint commands
		input := `break 4
continue
print x
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Before breakpoint")
		h.AssertOutput(stdout, "Breakpoint 1 set at line 4")
		h.AssertOutput(stdout, "Hit breakpoint 1")
		h.AssertOutput(stdout, "1") // Value of x
		h.AssertOutput(stdout, "x = 2")
		h.AssertOutput(stdout, "After breakpoint")
	})

	t.Run("step debugging", func(t *testing.T) {
		script := h.CreateSpell("step-test.lua", `
			function add(a, b)
				local sum = a + b
				return sum
			end
			
			local result = add(5, 3)
			print("Result: " .. result)
		`)

		// Step through execution
		input := `step
step
print sum
step
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		// Should show stepping through code
		output := stdout + stderr
		assert.Contains(t, output, "Step")
		assert.Contains(t, output, "Result: 8")
	})

	t.Run("variable inspection", func(t *testing.T) {
		script := h.CreateSpell("vars-test.lua", `
			local data = {
				name = "test",
				value = 42,
				nested = {
					flag = true
				}
			}
			print("Data initialized")
		`)

		// Inspect variables
		input := `break 8
continue
print data
print data.name
print data.nested.flag
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Data initialized")
		output := stdout + stderr
		assert.Contains(t, output, "test") // data.name
		assert.Contains(t, output, "42")   // data.value
		assert.Contains(t, output, "true") // data.nested.flag
	})

	t.Run("call stack inspection", func(t *testing.T) {
		script := h.CreateSpell("stack-test.lua", `
			function outer()
				inner()
			end
			
			function inner()
				print("In inner function")
			end
			
			outer()
		`)

		// Inspect call stack
		input := `break 7
continue
where
locals
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		output := stdout + stderr
		assert.Contains(t, output, "inner")
		assert.Contains(t, output, "outer")
		assert.Contains(t, output, "Call stack")
	})

	t.Run("conditional breakpoints", func(t *testing.T) {
		script := h.CreateSpell("conditional-test.lua", `
			for i = 1, 5 do
				print("i = " .. i)
			end
		`)

		// Set conditional breakpoint
		input := `break 2 i == 3
continue
print i
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "i = 1")
		h.AssertOutput(stdout, "i = 2")
		h.AssertOutput(stdout, "Hit breakpoint")
		h.AssertOutput(stdout, "3") // Should stop when i == 3
	})

	t.Run("debug help", func(t *testing.T) {
		script := h.CreateSpell("help-test.lua", `print("test")`)

		input := `help
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		// Should show debug commands
		h.AssertOutput(stdout, "Debug Commands:")
		h.AssertOutput(stdout, "break")
		h.AssertOutput(stdout, "step")
		h.AssertOutput(stdout, "continue")
		h.AssertOutput(stdout, "print")
		h.AssertOutput(stdout, "where")
		h.AssertOutput(stdout, "quit")
	})

	t.Run("debug with spell.yaml", func(t *testing.T) {
		// Create spell directory with debug config
		spellDir := filepath.Join(h.TempDir(), "debug-spell")
		h.CreateSpellYAML(spellDir, `
name: debug-spell
version: 1.0.0
engine: lua
debug:
  enabled: true
  breakpoints:
    - line: 3
parameters:
  message:
    type: string
    default: "Debug message"
`)

		script := filepath.Join(spellDir, "main.lua")
		h.CreateFile(script, `
			print("Starting debug spell")
			local msg = params.message
			print("Message: " .. msg)
			print("Done")
		`)

		input := `continue
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Starting debug spell")
		h.AssertOutput(stdout, "Hit breakpoint") // From spell.yaml config
		h.AssertOutput(stdout, "Message: Debug message")
		h.AssertOutput(stdout, "Done")
	})

	t.Run("debug with watch expressions", func(t *testing.T) {
		script := h.CreateSpell("watch-test.lua", `
			local counter = 0
			for i = 1, 3 do
				counter = counter + i
				print("Step " .. i)
			end
		`)

		input := `watch counter
break 4
continue
continue
continue
quit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "debug", script)

		h.AssertSuccess(stdout, stderr, err)
		output := stdout + stderr
		// Should show counter value changing
		assert.Contains(t, output, "Watch")
		assert.Contains(t, output, "counter")
		assert.Contains(t, output, "1") // First iteration
		assert.Contains(t, output, "3") // Second iteration (1+2)
		assert.Contains(t, output, "6") // Third iteration (1+2+3)
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
		assert.Contains(t, stderr, "syntax")
	})

	t.Run("invalid breakpoint", func(t *testing.T) {
		script := h.CreateSpell("small.lua", `print("one line")`)

		input := `break 100
quit
`
		stdout, stderr, _ := h.RunCommandWithInput(input, "debug", script)

		// Should handle gracefully
		output := stdout + stderr
		assert.Contains(t, output, "invalid line") // Line 100 doesn't exist
	})
}
