package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd_Structure(t *testing.T) {
	cmd := &RunCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test that Timeout field exists (Kong will set default later)
	assert.IsType(t, 0, cmd.Timeout)
}

func TestRunCmd_Run_NoRegistry(t *testing.T) {
	cmd := &RunCmd{
		Script: "test.lua",
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

func TestRunCmd_ParameterConversion(t *testing.T) {
	cmd := &RunCmd{
		Parameters: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Test that parameters are properly set
	assert.Equal(t, "value1", cmd.Parameters["key1"])
	assert.Equal(t, "value2", cmd.Parameters["key2"])
}
