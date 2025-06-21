// ABOUTME: Tests for the new command ensuring proper spell generation from templates.
// ABOUTME: Validates command execution, template selection, and error handling.

package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmd_Structure(t *testing.T) {
	cmd := &NewCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test default values - Kong sets these at parse time, not struct creation
	assert.Empty(t, cmd.Type)
	assert.Empty(t, cmd.Engine)
	assert.Empty(t, cmd.Description)
	assert.Empty(t, cmd.License)
	assert.Empty(t, cmd.OutputDir)
	assert.False(t, cmd.Force)
	assert.False(t, cmd.List)
}

func TestNewCmd_Run_ListTemplates(t *testing.T) {
	var stdout bytes.Buffer

	cmd := &NewCmd{
		List: true,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Available Templates:")
	assert.Contains(t, output, "basic")
	assert.Contains(t, output, "advanced")
	assert.Contains(t, output, "agent")
	assert.Contains(t, output, "workflow")
	assert.Contains(t, output, "interactive")
	assert.Contains(t, output, "Usage: llmspell new <name>")
}

func TestNewCmd_Run_MissingName(t *testing.T) {
	cmd := &NewCmd{
		Name: "",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "spell name is required")
}

func TestNewCmd_Run_BasicTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:        "test-spell",
		Type:        "basic",
		Engine:      "lua",
		Description: "Test spell",
		Author:      "Test Author",
		License:     "MIT",
		OutputDir:   tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check output messages
	output := stdout.String()
	assert.Contains(t, output, "Creating new basic spell: test-spell")
	assert.Contains(t, output, "✓ Spell created successfully")
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "cd test-spell")
	assert.Contains(t, output, "llmspell run main.lua")

	// Check generated files
	spellDir := filepath.Join(tmpDir, "test-spell")
	assert.DirExists(t, spellDir)
	assert.FileExists(t, filepath.Join(spellDir, "spell.yaml"))
	assert.FileExists(t, filepath.Join(spellDir, "main.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "README.md"))
}

func TestNewCmd_Run_JavaScriptEngine(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "js-spell",
		Type:      "basic",
		Engine:    "javascript",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check output shows .js extension
	output := stdout.String()
	assert.Contains(t, output, "llmspell run main.js")

	// Check generated files
	spellDir := filepath.Join(tmpDir, "js-spell")
	assert.FileExists(t, filepath.Join(spellDir, "main.js"))
}

func TestNewCmd_Run_TengoEngine(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "tengo-spell",
		Type:      "basic",
		Engine:    "tengo",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check output shows .tengo extension
	output := stdout.String()
	assert.Contains(t, output, "llmspell run main.tengo")

	// Check generated files
	spellDir := filepath.Join(tmpDir, "tengo-spell")
	assert.FileExists(t, filepath.Join(spellDir, "main.tengo"))
}

func TestNewCmd_Run_AdvancedTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "advanced-spell",
		Type:      "advanced",
		Engine:    "lua",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check generated files specific to advanced template
	spellDir := filepath.Join(tmpDir, "advanced-spell")
	assert.FileExists(t, filepath.Join(spellDir, "lib", "utils.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "lib", "prompts.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "config", "default.yaml"))
}

func TestNewCmd_Run_AgentTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "agent-spell",
		Type:      "agent",
		Engine:    "lua",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check generated files specific to agent template
	spellDir := filepath.Join(tmpDir, "agent-spell")
	assert.FileExists(t, filepath.Join(spellDir, "tools", "calculator.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "tools", "web_search.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "tools", "file_reader.lua"))
}

func TestNewCmd_Run_WorkflowTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "workflow-spell",
		Type:      "workflow",
		Engine:    "lua",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check generated files specific to workflow template
	spellDir := filepath.Join(tmpDir, "workflow-spell")
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "process_document.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "generate_report.lua"))
	assert.FileExists(t, filepath.Join(spellDir, "workflows", "analyze_data.lua"))
}

func TestNewCmd_Run_InteractiveTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "interactive-spell",
		Type:      "interactive",
		Engine:    "lua",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check main script contains interactive elements
	spellDir := filepath.Join(tmpDir, "interactive-spell")
	content, err := os.ReadFile(filepath.Join(spellDir, "main.lua"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Command handlers")
	assert.Contains(t, string(content), "/help")
}

func TestNewCmd_Run_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing directory
	spellDir := filepath.Join(tmpDir, "existing-spell")
	err := os.MkdirAll(spellDir, 0755)
	require.NoError(t, err)

	cmd := &NewCmd{
		Name:      "existing-spell",
		Type:      "basic",
		OutputDir: tmpDir,
		Force:     false,
	}

	ctx := context.Background()
	err = cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory already exists")
}

func TestNewCmd_Run_Force(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	// Create existing directory
	spellDir := filepath.Join(tmpDir, "existing-spell")
	err := os.MkdirAll(spellDir, 0755)
	require.NoError(t, err)

	cmd := &NewCmd{
		Name:      "existing-spell",
		Type:      "basic",
		OutputDir: tmpDir,
		Force:     true,
	}
	cmd.Out = &stdout

	ctx := context.Background()
	err = cmd.Run(ctx)

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "✓ Spell created successfully")
}

func TestNewCmd_Run_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &NewCmd{
		Name:      "verbose-spell",
		Type:      "basic",
		Engine:    "lua",
		Author:    "Test Author",
		License:   "Apache-2.0",
		OutputDir: tmpDir,
	}
	cmd.Out = &stdout

	ctx := context.WithValue(context.Background(), VerboseKey, true)
	ctx = context.WithValue(ctx, DebugKey, true)
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Check verbose output
	output := stdout.String()
	assert.Contains(t, output, "Template: basic")
	assert.Contains(t, output, "Engine: lua")
	assert.Contains(t, output, "Author: Test Author")
	assert.Contains(t, output, "License: Apache-2.0")
}

func TestNewCmd_GetExtension(t *testing.T) {
	tests := []struct {
		engine   string
		expected string
	}{
		{"lua", "lua"},
		{"javascript", "js"},
		{"js", "js"},
		{"tengo", "tengo"},
		{"unknown", "lua"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			cmd := &NewCmd{Engine: tt.engine}
			assert.Equal(t, tt.expected, cmd.getExtension())
		})
	}
}

func TestNewCmd_GetGitAuthor(t *testing.T) {
	cmd := &NewCmd{}

	// This test is environment-dependent
	// Just ensure it returns something
	author := cmd.getGitAuthor()
	assert.NotEmpty(t, author)
}

func TestNewCmd_SplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"single line", []string{"single line"}},
		{"", []string{}},
		{"line1\nline2\n", []string{"line1", "line2", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
