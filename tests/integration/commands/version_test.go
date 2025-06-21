// ABOUTME: Integration tests for the version command verifying version information display.
// ABOUTME: Tests version output formats and build information.

package commands

import (
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("basic version", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version")

		h.AssertSuccess(stdout, stderr, err)
		// Should show version information
		h.AssertOutput(stdout, "llmspell")
		// Version format: vX.Y.Z or dev
		assert.Regexp(t, `(v\d+\.\d+\.\d+|dev)`, stdout)
	})

	t.Run("verbose version", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose should show more details
		h.AssertOutput(stdout, "llmspell")
		h.AssertOutput(stdout, "version:")
		h.AssertOutput(stdout, "go version:")
		h.AssertOutput(stdout, "built:")
		h.AssertOutput(stdout, "platform:")
	})

	t.Run("version with build info", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--build-info")

		h.AssertSuccess(stdout, stderr, err)
		// Should show build details
		h.AssertOutput(stdout, "version")
		h.AssertOutput(stdout, "commit")
		h.AssertOutput(stdout, "built")
		h.AssertOutput(stdout, "go version")
	})

	t.Run("JSON format", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--format", "json")

		h.AssertSuccess(stdout, stderr, err)
		// Should output valid JSON
		assert.Contains(t, stdout, "{")
		assert.Contains(t, stdout, "\"version\"")
		assert.Contains(t, stdout, "\"go_version\"")
		assert.Contains(t, stdout, "}")

		// Validate it's proper JSON structure
		assert.Regexp(t, `"version"\s*:\s*"[^"]+"`, stdout)
	})

	t.Run("short version", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--short")

		h.AssertSuccess(stdout, stderr, err)
		// Short format should be just the version
		assert.Regexp(t, `^(v\d+\.\d+\.\d+|dev)\s*$`, stdout)
	})

	t.Run("check dependencies", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--deps")

		h.AssertSuccess(stdout, stderr, err)
		// Should list key dependencies
		h.AssertOutput(stdout, "dependencies:")
		h.AssertOutput(stdout, "github.com/yuin/gopher-lua")
		h.AssertOutput(stdout, "github.com/alecthomas/kong")
		h.AssertOutput(stdout, "github.com/knadh/koanf")
	})

	t.Run("check compatibility", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--check-compat")

		h.AssertSuccess(stdout, stderr, err)
		// Should show compatibility info
		h.AssertOutput(stdout, "go-llms compatibility:")
		h.AssertOutput(stdout, "minimum version:")
		h.AssertOutput(stdout, "current version:")
	})
}

func TestVersionHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("version help", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("version", "--help")

		h.AssertSuccess(stdout, stderr, err)
		// Should show help for version command
		h.AssertOutput(stdout, "Show version information")
		h.AssertOutput(stdout, "--verbose")
		h.AssertOutput(stdout, "--format")
		h.AssertOutput(stdout, "--short")
	})
}
