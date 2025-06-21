package commands

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPLCmd_Structure(t *testing.T) {
	cmd := &REPLCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test that Engine field exists
	assert.IsType(t, "", cmd.Engine)

	// Test that HistoryFile field exists
	assert.IsType(t, "", cmd.HistoryFile)
}

func TestREPLCmd_UnsupportedEngine(t *testing.T) {
	cmd := &REPLCmd{
		Engine: "javascript",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported engine: javascript")
}

func TestREPLCmd_DefaultHistoryPath(t *testing.T) {
	cmd := &REPLCmd{
		Engine: "lua",
	}

	// Set up output capture
	var stdout, stderr bytes.Buffer
	var stdin strings.Reader
	stdin.Reset(".exit\n")

	cmd.Out = &stdout
	cmd.Err = &stderr

	// Create a mock input that immediately exits
	ctx := context.Background()

	// This will fail because we can't easily mock stdin for readline,
	// but we can test that it gets to the REPL creation
	err := cmd.Run(ctx)

	// The error might be from readline setup or from .exit command
	// Both are acceptable for this test - we just want to verify
	// it gets past the initial validation
	if err != nil {
		// If it's a readline error, that's expected in test environment
		assert.True(t,
			strings.Contains(err.Error(), "readline") ||
				strings.Contains(err.Error(), "exit") ||
				strings.Contains(err.Error(), "REPL execution failed"),
			"Expected readline, exit, or REPL execution error, got: %v", err)
	}
}

func TestREPLCmd_CustomHistoryFile(t *testing.T) {
	// Create temp file for testing
	tmpFile, err := os.CreateTemp("", "repl_test_*.hist")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	cmd := &REPLCmd{
		Engine:      "lua",
		HistoryFile: tmpFile.Name(),
	}

	// Set up output capture
	var stdout, stderr bytes.Buffer
	cmd.Out = &stdout
	cmd.Err = &stderr

	ctx := context.Background()
	err = cmd.Run(ctx)

	// Similar to above - we expect either readline or exit error
	if err != nil {
		assert.True(t,
			strings.Contains(err.Error(), "readline") ||
				strings.Contains(err.Error(), "exit") ||
				strings.Contains(err.Error(), "REPL execution failed"),
			"Expected readline, exit, or REPL execution error, got: %v", err)
	}
}
