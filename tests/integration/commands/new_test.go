// ABOUTME: Integration tests for the new command verifying spell creation from templates.
// ABOUTME: Tests template generation, file creation, and various template types.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("create basic spell", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "myspell", "--type", "basic")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "✓ Spell created successfully")

		// Check created files
		spellDir := filepath.Join(h.TempDir(), "myspell")
		assert.DirExists(t, spellDir)
		assert.FileExists(t, filepath.Join(spellDir, "spell.yaml"))
		assert.FileExists(t, filepath.Join(spellDir, "main.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "README.md"))
	})

	t.Run("create advanced spell", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "advanced-spell", "--type", "advanced", "--engine", "lua")

		h.AssertSuccess(stdout, stderr, err)

		spellDir := filepath.Join(h.TempDir(), "advanced-spell")
		assert.FileExists(t, filepath.Join(spellDir, "lib", "utils.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "lib", "prompts.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "config", "default.yaml"))
	})

	t.Run("create agent spell", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "agent-spell", "--type", "agent")

		h.AssertSuccess(stdout, stderr, err)

		spellDir := filepath.Join(h.TempDir(), "agent-spell")
		assert.FileExists(t, filepath.Join(spellDir, "tools", "calculator.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "tools", "web_search.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "tools", "file_reader.lua"))
	})

	t.Run("create workflow spell", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "workflow-spell", "--type", "workflow")

		h.AssertSuccess(stdout, stderr, err)

		spellDir := filepath.Join(h.TempDir(), "workflow-spell")
		assert.FileExists(t, filepath.Join(spellDir, "workflows", "process_document.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "workflows", "generate_report.lua"))
		assert.FileExists(t, filepath.Join(spellDir, "workflows", "analyze_data.lua"))
	})

	t.Run("create javascript spell", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "js-spell", "--engine", "javascript")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "llmspell run main.js")

		spellDir := filepath.Join(h.TempDir(), "js-spell")
		assert.FileExists(t, filepath.Join(spellDir, "main.js"))
	})

	t.Run("list templates", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "--list")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Available Templates:")
		h.AssertOutput(stdout, "basic")
		h.AssertOutput(stdout, "advanced")
		h.AssertOutput(stdout, "agent")
		h.AssertOutput(stdout, "workflow")
		h.AssertOutput(stdout, "interactive")
	})

	t.Run("existing directory without force", func(t *testing.T) {
		// Create directory first
		spellDir := filepath.Join(h.TempDir(), "existing")
		require.NoError(t, os.MkdirAll(spellDir, 0755))

		stdout, stderr, err := h.RunCommand("new", "existing")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "already exists")
	})

	t.Run("existing directory with force", func(t *testing.T) {
		// Create directory first
		spellDir := filepath.Join(h.TempDir(), "forced")
		require.NoError(t, os.MkdirAll(spellDir, 0755))

		stdout, stderr, err := h.RunCommand("new", "forced", "--force")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "✓ Spell created successfully")
	})

	t.Run("custom metadata", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("new", "custom-spell",
			"--author", "Test Author",
			"--license", "Apache-2.0",
			"--description", "Custom test spell")

		h.AssertSuccess(stdout, stderr, err)

		// Check spell.yaml contains custom metadata
		spellYaml := filepath.Join(h.TempDir(), "custom-spell", "spell.yaml")
		content, err := os.ReadFile(spellYaml)
		require.NoError(t, err)

		assert.Contains(t, string(content), "author: Test Author")
		assert.Contains(t, string(content), "license: Apache-2.0")
		assert.Contains(t, string(content), "description: Custom test spell")
	})
}
