package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugCmd_Structure(t *testing.T) {
	cmd := &DebugCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test fields
	assert.Empty(t, cmd.Script)
	assert.Nil(t, cmd.Breakpoints)
}

func TestDebugCmd_Run_NotImplemented(t *testing.T) {
	cmd := &DebugCmd{
		Script: "test.lua",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "debug command not implemented yet")

	// Should show debug info
	output := stdout.String()
	assert.Contains(t, output, "Debugging script: test.lua")
}

func TestDebugCmd_Run_WithBreakpoints(t *testing.T) {
	cmd := &DebugCmd{
		Script:      "test.lua",
		Breakpoints: []int{10, 20, 30},
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "debug command not implemented yet")

	// Should show breakpoints
	output := stdout.String()
	assert.Contains(t, output, "Debugging script: test.lua")
	assert.Contains(t, output, "Breakpoints at lines: [10 20 30]")
}
