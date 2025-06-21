// ABOUTME: Integration tests for the validate command verifying spell and script validation.
// ABOUTME: Tests syntax checking, security validation, and spell.yaml validation.

package commands

import (
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("validate valid lua script", func(t *testing.T) {
		script := h.CreateSpell("valid.lua", helpers.BasicLuaSpell())

		stdout, stderr, err := h.RunCommand("validate", script)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("validate invalid lua script", func(t *testing.T) {
		script := h.CreateSpell("invalid.lua", `
			-- Invalid syntax
			function missing_end()
				print("no end")
		`)

		stdout, stderr, err := h.RunCommand("validate", script)

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "syntax")
	})

	t.Run("validate spell.yaml", func(t *testing.T) {
		spellDir := filepath.Join(h.TempDir(), "myspell")
		h.CreateSpellYAML(spellDir, helpers.BasicSpellYAML())

		stdout, stderr, err := h.RunCommand("validate", spellDir)

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("validate invalid spell.yaml", func(t *testing.T) {
		spellDir := filepath.Join(h.TempDir(), "badspell")
		h.CreateSpellYAML(spellDir, `
name: 123invalid-name
version: not-a-version
engine: unsupported
`)

		stdout, stderr, err := h.RunCommand("validate", spellDir)

		h.AssertFailure(stdout, stderr, err)
		// Should report validation errors
		output := stdout + stderr
		assert.Contains(t, output, "invalid")
	})

	t.Run("validate security profile", func(t *testing.T) {
		script := h.CreateSpell("secure.lua", `
			-- Attempting to access restricted functionality
			os.execute("rm -rf /")
		`)

		stdout, stderr, _ := h.RunCommand("validate", script, "--profile", "sandbox")

		// Should warn about security issues
		output := stdout + stderr
		assert.Contains(t, output, "security") // Should mention security concern
	})

	t.Run("validate with verbose output", func(t *testing.T) {
		script := h.CreateSpell("verbose.lua", helpers.BasicLuaSpell())

		stdout, stderr, err := h.RunCommand("validate", script, "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose mode should show more details
		output := stdout + stderr
		assert.Contains(t, output, "lua") // Should mention engine
	})

	t.Run("validate multiple files", func(t *testing.T) {
		script1 := h.CreateSpell("script1.lua", helpers.BasicLuaSpell())
		script2 := h.CreateSpell("script2.lua", `print("Script 2")`)

		stdout, stderr, err := h.RunCommand("validate", script1, script2)

		h.AssertSuccess(stdout, stderr, err)
		// Should validate both files
		h.AssertOutput(stdout, "script1.lua")
		h.AssertOutput(stdout, "script2.lua")
	})
}
