package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPLCmd_Structure(t *testing.T) {
	cmd := &REPLCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test that Engine field exists (Kong will set default later)
	assert.IsType(t, "", cmd.Engine)
}

func TestREPLCmd_Run_NotImplemented(t *testing.T) {
	cmd := &REPLCmd{
		Engine: "lua",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "REPL not implemented yet")

	// Should show startup message
	output := stdout.String()
	assert.Contains(t, output, "Starting REPL with lua engine...")
}

func TestREPLCmd_CustomEngine(t *testing.T) {
	cmd := &REPLCmd{
		Engine: "javascript",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)

	// Should show engine in message
	output := stdout.String()
	assert.Contains(t, output, "Starting REPL with javascript engine...")
}
