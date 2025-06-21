// ABOUTME: Tests for the man command functionality.
// ABOUTME: Verifies man page generation and installation features.

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

func TestManCmd_Run_MainPage(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, ".TH LLMSPELL 1")
	assert.Contains(t, output, "scriptable LLM interactions")
	assert.Contains(t, output, ".SH COMMANDS")
}

func TestManCmd_Run_CommandPage(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
		Command: "run",
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, ".TH LLMSPELL-RUN 1")
	assert.Contains(t, output, "execute a spell script")
}

func TestManCmd_Run_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
			Err: &stderr,
		},
		Command: "unknown",
	}

	err := cmd.Run(context.Background())
	assert.Error(t, err)
}

func TestManCmd_Run_OutputToFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test.1")

	cmd := &ManCmd{
		Output: outputFile,
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	// Check file was created
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".TH LLMSPELL 1")
}

func TestManCmd_GenerateAll(t *testing.T) {
	tempDir := t.TempDir()
	var stdout bytes.Buffer

	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
		Dir: tempDir,
		All: true,
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	// Check main page was created
	mainFile := filepath.Join(tempDir, "llmspell.1")
	assert.FileExists(t, mainFile)

	// Check some command pages were created
	assert.FileExists(t, filepath.Join(tempDir, "llmspell-run.1"))
	assert.FileExists(t, filepath.Join(tempDir, "llmspell-repl.1"))
	assert.FileExists(t, filepath.Join(tempDir, "llmspell-new.1"))

	// Check output message
	output := stdout.String()
	assert.Contains(t, output, "All man pages generated successfully!")
}

func TestManCmd_Format_Text(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
		Format: "text",
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	output := stdout.String()
	// Should not contain troff commands
	assert.NotContains(t, output, ".TH")
	assert.NotContains(t, output, ".SH")
	assert.NotContains(t, output, "\\fB")
	// Should contain content
	assert.Contains(t, output, "scriptable LLM interactions")
}

func TestManCmd_Format_HTML(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &ManCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
		Format: "html",
	}

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "<html>")
	assert.Contains(t, output, "</html>")
	assert.Contains(t, output, "<pre>")
}

func TestManCmd_findManDirectory(t *testing.T) {
	cmd := &ManCmd{}

	dir := cmd.findManDirectory()
	// Should find some directory (even if it's the user directory)
	assert.NotEmpty(t, dir)
}

func TestManCmd_checkWritePermission(t *testing.T) {
	cmd := &ManCmd{}

	// Test with temp directory (should have permission)
	tempDir := t.TempDir()
	err := cmd.checkWritePermission(tempDir)
	assert.NoError(t, err)

	// Test with non-existent directory (should fail)
	err = cmd.checkWritePermission("/nonexistent/directory")
	assert.Error(t, err)
}

func TestManCmd_convertToText(t *testing.T) {
	cmd := &ManCmd{}

	troff := `.TH TEST 1 "Date" "Version"
.SH NAME
test \- test program
.B bold text
\fIitalic\fR text`

	text := cmd.convertToText(troff)

	assert.Contains(t, text, "MANUAL PAGE:")
	assert.Contains(t, text, "NAME")
	assert.Contains(t, text, "test - test program")
	assert.NotContains(t, text, "\\fI")
	assert.NotContains(t, text, "\\fR")
	assert.NotContains(t, text, ".B")
}

func TestManCmd_writeToFile_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	cmd := &ManCmd{}

	// Write to nested directory that doesn't exist
	filePath := filepath.Join(tempDir, "subdir", "test.txt")
	err := cmd.writeToFile(filePath, "test content")
	require.NoError(t, err)

	// Check file exists and has content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}
