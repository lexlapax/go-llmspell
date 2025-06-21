// ABOUTME: Integration tests for the REPL command verifying interactive execution.
// ABOUTME: Tests REPL commands, multi-line input, history, and error handling.

package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestREPLCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("basic REPL interaction", func(t *testing.T) {
		input := `print("Hello from REPL")
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Hello from REPL")
		h.AssertOutput(stdout, "lua>") // Default prompt
	})

	t.Run("REPL with custom engine", func(t *testing.T) {
		input := `.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--engine", "lua", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "lua>") // Engine-specific prompt
	})

	t.Run("REPL commands", func(t *testing.T) {
		input := `.help
.engines
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Available commands:")
		h.AssertOutput(stdout, ".help")
		h.AssertOutput(stdout, ".exit")
		h.AssertOutput(stdout, ".clear")
		h.AssertOutput(stdout, "lua") // Should list available engines
	})

	t.Run("multi-line input", func(t *testing.T) {
		input := `function test()
  return "multi-line"
end
print(test())
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "multi-line")
		h.AssertOutput(stdout, "...") // Continuation prompt
	})

	t.Run("error handling in REPL", func(t *testing.T) {
		input := `error("test error")
print("still working")
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err) // REPL itself should not exit on script errors
		output := stdout + stderr
		assert.Contains(t, output, "test error")
		assert.Contains(t, output, "still working") // Should continue after error
	})

	t.Run("variable persistence", func(t *testing.T) {
		input := `x = 42
print("x = " .. x)
y = x + 8
print("y = " .. y)
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "x = 42")
		h.AssertOutput(stdout, "y = 50")
	})

	t.Run("clear command", func(t *testing.T) {
		input := `x = 42
.clear
print(x)
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		output := stdout + stderr
		assert.Contains(t, output, "nil") // x should be nil after clear
	})

	t.Run("save and load session", func(t *testing.T) {
		// First session - save
		input1 := `x = 42
function greet(name)
  return "Hello, " .. name
end
.save test_session.lua
.exit
`
		stdout1, stderr1, err1 := h.RunCommandWithInput(input1, "repl", "--no-history")
		h.AssertSuccess(stdout1, stderr1, err1)
		h.AssertOutput(stdout1, "Session saved")

		// Second session - load
		input2 := `.load test_session.lua
print(x)
print(greet("World"))
.exit
`
		stdout2, stderr2, err2 := h.RunCommandWithInput(input2, "repl", "--no-history")
		h.AssertSuccess(stdout2, stderr2, err2)
		h.AssertOutput(stdout2, "42")
		h.AssertOutput(stdout2, "Hello, World")
	})

	t.Run("REPL with config", func(t *testing.T) {
		config := h.CreateConfigFile(`
repl:
  prompt: "custom> "
  save_history: false
  syntax_highlight: false
`)

		input := `print("custom prompt")
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--config", config)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "custom>")
		h.AssertOutput(stdout, "custom prompt")
	})

	t.Run("invalid engine", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("repl", "--engine", "python")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "not yet implemented")
	})
}

func TestREPLHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("history persistence", func(t *testing.T) {
		// Skip if we can't test history properly
		t.Skip("History testing requires terminal emulation")

		// First session
		input1 := `print("first command")
.exit
`
		_, _, err1 := h.RunCommandWithInput(input1, "repl")
		assert.NoError(t, err1)

		// Wait a bit for history to be written
		time.Sleep(100 * time.Millisecond)

		// Second session - history should be available
		// This would require terminal emulation to test properly
	})
}

func TestREPLSyntaxHighlighting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("syntax highlighting enabled", func(t *testing.T) {
		input := `print("test")
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history")

		h.AssertSuccess(stdout, stderr, err)
		// With highlighting, we might see ANSI codes
		// This is hard to test without a proper terminal
		output := stdout

		// At minimum, the output should contain the text
		assert.Contains(t, strings.ReplaceAll(output, "\x1b[0m", ""), "test")
	})

	t.Run("syntax highlighting disabled", func(t *testing.T) {
		input := `print("test")
.exit
`
		stdout, stderr, err := h.RunCommandWithInput(input, "repl", "--no-history", "--no-highlight")

		h.AssertSuccess(stdout, stderr, err)
		// Should not contain ANSI escape codes
		assert.NotContains(t, stdout, "\x1b[")
	})
}
