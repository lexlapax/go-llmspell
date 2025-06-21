package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityCmd_Structure(t *testing.T) {
	cmd := &SecurityCmd{}

	// Test that it embeds BaseCommand
	assert.NotNil(t, cmd.BaseCommand)

	// Test that Action field exists (Kong will set default later)
	assert.IsType(t, "", cmd.Action)
}

func TestSecurityCmd_Run_List(t *testing.T) {
	cmd := &SecurityCmd{
		Action: "list",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Available security profiles:")
	assert.Contains(t, output, "sandbox")
	assert.Contains(t, output, "development")
	assert.Contains(t, output, "production")
}

func TestSecurityCmd_Run_Show_Default(t *testing.T) {
	cmd := &SecurityCmd{
		Action:  "show",
		Profile: "",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Context with profile
	ctx := context.WithValue(context.Background(), ProfileKey, "development")
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Profile: development")
	assert.Contains(t, output, "Balanced for development")
	assert.Contains(t, output, "Permissions:")
}

func TestSecurityCmd_Run_Show_Specific(t *testing.T) {
	tests := []struct {
		profile string
		desc    string
		perms   []string
	}{
		{
			profile: "sandbox",
			desc:    "Maximum security restrictions",
			perms:   []string{"read:script", "execute:llm"},
		},
		{
			profile: "development",
			desc:    "Balanced for development",
			perms:   []string{"read:*", "write:temp", "execute:*", "network:llm"},
		},
		{
			profile: "production",
			desc:    "Production security settings",
			perms:   []string{"read:*", "write:output", "execute:*", "network:*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			cmd := &SecurityCmd{
				Action:  "show",
				Profile: tt.profile,
			}

			var stdout bytes.Buffer
			cmd.Out = &stdout

			ctx := context.Background()
			err := cmd.Run(ctx)

			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Profile: "+tt.profile)
			assert.Contains(t, output, tt.desc)
			for _, perm := range tt.perms {
				assert.Contains(t, output, perm)
			}
		})
	}
}

func TestSecurityCmd_Run_Show_Unknown(t *testing.T) {
	cmd := &SecurityCmd{
		Action:  "show",
		Profile: "invalid",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown profile: invalid")
}

func TestSecurityCmd_Run_Validate_NoProfile(t *testing.T) {
	cmd := &SecurityCmd{
		Action:  "validate",
		Profile: "",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile name required")
}

func TestSecurityCmd_Run_Validate_Valid(t *testing.T) {
	profiles := []string{"sandbox", "development", "production"}

	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			cmd := &SecurityCmd{
				Action:  "validate",
				Profile: profile,
			}

			var stdout bytes.Buffer
			cmd.Out = &stdout

			ctx := context.Background()
			err := cmd.Run(ctx)

			require.NoError(t, err)

			output := strings.TrimSpace(stdout.String())
			assert.Equal(t, "✓ Profile '"+profile+"' is valid", output)
		})
	}
}

func TestSecurityCmd_Run_Validate_Invalid(t *testing.T) {
	cmd := &SecurityCmd{
		Action:  "validate",
		Profile: "invalid",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid profile: invalid")
}

func TestSecurityCmd_Run_UnknownAction(t *testing.T) {
	cmd := &SecurityCmd{
		Action: "invalid",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action: invalid")
}
