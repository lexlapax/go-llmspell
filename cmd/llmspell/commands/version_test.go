package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd_Structure(t *testing.T) {
	cmd := &VersionCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test default values
	assert.False(t, cmd.Short)
}

func TestVersionCmd_Run_Short(t *testing.T) {
	cmd := &VersionCmd{
		Short: true,
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Run command
	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Short version should only show version string
	output := strings.TrimSpace(stdout.String())
	assert.Equal(t, "dev", output)
}

func TestVersionCmd_Run_Full(t *testing.T) {
	// Save original values
	origVersion := Version
	origBuildDate := BuildDate
	origGitCommit := GitCommit

	// Set test values
	Version = "1.0.0"
	BuildDate = "2024-01-01"
	GitCommit = "abc123"

	// Restore after test
	defer func() {
		Version = origVersion
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()

	cmd := &VersionCmd{
		Short: false,
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Run command
	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Full version should show all info
	output := stdout.String()
	assert.Contains(t, output, "llmspell version 1.0.0")
	assert.Contains(t, output, "Commit: abc123")
	assert.Contains(t, output, "Built: 2024-01-01")
}

func TestVersionCmd_Run_MinimalInfo(t *testing.T) {
	// Save original values
	origBuildDate := BuildDate
	origGitCommit := GitCommit

	// Clear optional values
	BuildDate = ""
	GitCommit = ""

	// Restore after test
	defer func() {
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()

	cmd := &VersionCmd{}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Run command
	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Should only show version line
	output := stdout.String()
	assert.Contains(t, output, "llmspell version")
	assert.NotContains(t, output, "Commit:")
	assert.NotContains(t, output, "Built:")
}
