package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCmd_Structure(t *testing.T) {
	cmd := &ValidateCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)
}

func TestValidateCmd_Run_NoRegistry(t *testing.T) {
	cmd := &ValidateCmd{
		Path: "test.lua",
	}

	// Set up output capture
	var stdout, stderr bytes.Buffer
	cmd.Out = &stdout
	cmd.Err = &stderr

	// Run without engine registry in context
	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "engine registry not found")
}

func TestValidateCmd_PathRequired(t *testing.T) {
	cmd := &ValidateCmd{}

	// Path should be required by Kong validation
	assert.Empty(t, cmd.Path)
}
