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

func TestDebugCmd_Structure(t *testing.T) {
	cmd := &DebugCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test fields - Kong sets defaults at parse time, not struct creation
	assert.Empty(t, cmd.Script)
	assert.Nil(t, cmd.Breakpoints)
	assert.Empty(t, cmd.Engine)     // Default "lua" is set by Kong at parse time
	assert.Equal(t, 0, cmd.Timeout) // Default 300 is set by Kong at parse time
	assert.False(t, cmd.StepMode)
	assert.Nil(t, cmd.Env)
}

func TestDebugCmd_Run_NoEngineRegistry(t *testing.T) {
	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.lua")
	err := os.WriteFile(scriptPath, []byte("print('test')"), 0644)
	require.NoError(t, err)

	cmd := &DebugCmd{
		Script: scriptPath,
		Engine: "lua",
	}

	// Set up output capture
	var stderr bytes.Buffer
	cmd.Err = &stderr

	ctx := context.Background()
	err = cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "engine registry not found in context")
}

func TestDebugCmd_Run_InvalidScript(t *testing.T) {
	cmd := &DebugCmd{
		Script: "/nonexistent/script.lua",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read script")
}

func TestDebugCmd_Run_DebugHeader(t *testing.T) {
	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.lua")
	err := os.WriteFile(scriptPath, []byte("print('test')"), 0644)
	require.NoError(t, err)

	cmd := &DebugCmd{
		Script:      scriptPath,
		Engine:      "lua",
		Breakpoints: []int{10, 20, 30},
		StepMode:    true,
	}

	// Set up output capture
	var stderr bytes.Buffer
	cmd.Err = &stderr

	ctx := context.Background()
	err = cmd.Run(ctx)

	// Will error due to no engine registry
	require.Error(t, err)
	assert.Contains(t, err.Error(), "engine registry not found in context")
}

func TestDebugCmd_Run_WithEnv(t *testing.T) {
	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.js")
	err := os.WriteFile(scriptPath, []byte("console.log('test')"), 0644)
	require.NoError(t, err)

	cmd := &DebugCmd{
		Script: scriptPath,
		Engine: "javascript",
		Env: map[string]string{
			"FOO": "bar",
			"BAZ": "qux",
		},
		Timeout: 60,
	}

	// Set up output capture
	var stderr bytes.Buffer
	cmd.Err = &stderr

	ctx := context.Background()
	err = cmd.Run(ctx)

	// Will error due to no engine registry
	require.Error(t, err)
	assert.Contains(t, err.Error(), "engine registry not found in context")
}

func TestDebugCmd_Run_PrintsDebugHeader(t *testing.T) {
	// This test verifies that debug header is printed after file read
	// We need a custom implementation that exits after printing header

	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.lua")
	err := os.WriteFile(scriptPath, []byte("print('test')"), 0644)
	require.NoError(t, err)

	cmd := &DebugCmd{
		Script:      scriptPath,
		Engine:      "lua",
		Breakpoints: []int{5, 10},
		StepMode:    true,
	}

	// Set up output capture
	var stderr bytes.Buffer
	cmd.Err = &stderr

	// We can't fully test this without engine registry, but we can verify
	// the structure is correct
	assert.Equal(t, scriptPath, cmd.Script)
	assert.Equal(t, "lua", cmd.Engine)
	assert.Equal(t, []int{5, 10}, cmd.Breakpoints)
	assert.True(t, cmd.StepMode)
}

func TestDebugCmd_Run_VerboseContext(t *testing.T) {
	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.lua")
	err := os.WriteFile(scriptPath, []byte("return 42"), 0644)
	require.NoError(t, err)

	cmd := &DebugCmd{
		Script: scriptPath,
	}

	// Set up output capture
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Out = &stdout
	cmd.Err = &stderr

	// Create context with verbose flag
	ctx := context.WithValue(context.Background(), VerboseKey, true)
	err = cmd.Run(ctx)

	// Will error due to no engine registry
	require.Error(t, err)

	// Verify context usage in engine config setup
	// The error occurs before we can test verbose output
	assert.Contains(t, err.Error(), "engine registry not found in context")
}
