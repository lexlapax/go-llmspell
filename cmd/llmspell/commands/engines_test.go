package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnginesCmd_Structure(t *testing.T) {
	cmd := &EnginesCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test default values
	assert.False(t, cmd.Details)
}

func TestEnginesCmd_Run_NoRegistry(t *testing.T) {
	cmd := &EnginesCmd{}

	// Set up output capture
	var stdout, stderr bytes.Buffer
	cmd.Out = &stdout
	cmd.Err = &stderr

	// Run without engine registry in context
	ctx := context.Background()
	err := cmd.Run(ctx)

	// Should not error, but show fallback list
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Available engines:")
	assert.Contains(t, output, "lua")
	assert.Contains(t, output, "javascript")
	assert.Contains(t, output, "tengo")
}

func TestEnginesCmd_Details(t *testing.T) {
	cmd := &EnginesCmd{
		Details: true,
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Run without registry (fallback mode)
	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// With details flag but no registry, should still show basic info
	output := stdout.String()
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 3)
}
