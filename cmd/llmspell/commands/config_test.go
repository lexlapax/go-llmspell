package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd_Structure(t *testing.T) {
	cmd := &ConfigCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test that Action field exists (Kong will set default later)
	// We can't test default values in unit tests since Kong sets them
	assert.IsType(t, "", cmd.Action)
}

func TestConfigCmd_Run_Show(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "show",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Create context with config
	ctx := context.WithValue(context.Background(), ConfigKey, &config.Config{
		Debug: true,
		Engine: config.EngineConfig{
			Default:      "lua",
			MemoryLimit:  1024,
			TimeoutLimit: 30,
		},
		Security: config.SecurityConfig{
			Profile: "sandbox",
		},
	})

	err := cmd.Run(ctx)
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Configuration:")
	assert.Contains(t, output, "debug: true")
	assert.Contains(t, output, "default: lua")
	assert.Contains(t, output, "profile: sandbox")
}

func TestConfigCmd_Run_Path(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "path",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := strings.TrimSpace(stdout.String())
	assert.Contains(t, output, "llmspell")
	assert.Contains(t, output, "config.yaml")
}

func TestConfigCmd_Run_Get_NoKey(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "get",
		Key:    "",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "key required for get action")
}

func TestConfigCmd_Run_Set_NoKey(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "set",
		Key:    "",
		Value:  "test",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "key required for set action")
}

func TestConfigCmd_Run_Set_NoValue(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "set",
		Key:    "test.key",
		Value:  "",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "value required for set action")
}

func TestConfigCmd_Run_Set_Success(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "set",
		Key:    "test.key",
		Value:  "test-value",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	// Should show instructions
	output := stdout.String()
	assert.Contains(t, output, "To set 'test.key' to 'test-value'")
	assert.Contains(t, output, "edit the config file")
	assert.Contains(t, output, "Example:")
}

func TestConfigCmd_Run_UnknownAction(t *testing.T) {
	cmd := &ConfigCmd{
		Action: "invalid",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action: invalid")
}
