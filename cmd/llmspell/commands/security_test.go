// ABOUTME: Unit tests for the security command implementation.
// ABOUTME: Tests list, show, and validate actions for security levels and feature sets.

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
	assert.Contains(t, output, "Available security levels:")
	assert.Contains(t, output, "untrusted")
	assert.Contains(t, output, "trusted")
	assert.Contains(t, output, "privileged")

	assert.Contains(t, output, "Available feature sets:")
	assert.Contains(t, output, "minimal")
	assert.Contains(t, output, "llm")
	assert.Contains(t, output, "agent")
	assert.Contains(t, output, "observable")
	assert.Contains(t, output, "full")
}

func TestSecurityCmd_Run_Show_Default(t *testing.T) {
	cmd := &SecurityCmd{
		Action:        "show",
		SecurityLevel: "",
		FeatureSet:    "full",
	}

	// Set up output capture
	var stdout bytes.Buffer
	cmd.Out = &stdout

	// Context with security level
	ctx := context.WithValue(context.Background(), SecurityLevelKey, "trusted")
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Security Level: trusted")
	assert.Contains(t, output, "Permissions:")
	assert.Contains(t, output, "Network: true")
	assert.Contains(t, output, "Filesystem: true")

	assert.Contains(t, output, "Feature Set: full")
	assert.Contains(t, output, "Enabled Bridge Sets:")
}

func TestSecurityCmd_Run_Show_Specific(t *testing.T) {
	tests := []struct {
		level      string
		featureSet string
		network    string
		filesystem string
		exec       string
	}{
		{
			level:      "untrusted",
			featureSet: "minimal",
			network:    "Network: false",
			filesystem: "Filesystem: false",
			exec:       "Execute: false",
		},
		{
			level:      "trusted",
			featureSet: "full",
			network:    "Network: true",
			filesystem: "Filesystem: true",
			exec:       "Execute: false",
		},
		{
			level:      "privileged",
			featureSet: "full",
			network:    "Network: true",
			filesystem: "Filesystem: true",
			exec:       "Execute: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cmd := &SecurityCmd{
				Action:        "show",
				SecurityLevel: tt.level,
				FeatureSet:    tt.featureSet,
			}

			var stdout bytes.Buffer
			cmd.Out = &stdout

			ctx := context.Background()
			err := cmd.Run(ctx)

			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Security Level: "+tt.level)
			assert.Contains(t, output, tt.network)
			assert.Contains(t, output, tt.filesystem)
			assert.Contains(t, output, tt.exec)

			assert.Contains(t, output, "Feature Set: "+tt.featureSet)
		})
	}
}

func TestSecurityCmd_Run_Validate_NoInput(t *testing.T) {
	cmd := &SecurityCmd{
		Action:        "validate",
		SecurityLevel: "",
		FeatureSet:    "",
	}

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "security level or feature set required")
}

func TestSecurityCmd_Run_Validate_Valid(t *testing.T) {
	levels := []string{"untrusted", "trusted", "privileged"}
	featureSets := []string{"minimal", "llm", "agent", "observable", "full"}

	// Test security levels
	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			cmd := &SecurityCmd{
				Action:        "validate",
				SecurityLevel: level,
				FeatureSet:    "",
			}

			var stdout bytes.Buffer
			cmd.Out = &stdout

			ctx := context.Background()
			err := cmd.Run(ctx)

			require.NoError(t, err)

			output := strings.TrimSpace(stdout.String())
			assert.Equal(t, "✓ Security level '"+level+"' is valid", output)
		})
	}

	// Test feature sets
	for _, fs := range featureSets {
		t.Run("featureset_"+fs, func(t *testing.T) {
			cmd := &SecurityCmd{
				Action:        "validate",
				SecurityLevel: "",
				FeatureSet:    fs,
			}

			var stdout bytes.Buffer
			cmd.Out = &stdout

			ctx := context.Background()
			err := cmd.Run(ctx)

			require.NoError(t, err)

			output := strings.TrimSpace(stdout.String())
			assert.Equal(t, "✓ Feature set '"+fs+"' is valid", output)
		})
	}
}

func TestSecurityCmd_Run_Validate_Invalid(t *testing.T) {
	t.Run("invalid_security_level", func(t *testing.T) {
		cmd := &SecurityCmd{
			Action:        "validate",
			SecurityLevel: "invalid",
			FeatureSet:    "",
		}

		ctx := context.Background()
		err := cmd.Run(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid security level")
	})

	t.Run("invalid_feature_set", func(t *testing.T) {
		cmd := &SecurityCmd{
			Action:        "validate",
			SecurityLevel: "",
			FeatureSet:    "invalid",
		}

		ctx := context.Background()
		err := cmd.Run(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid feature set")
	})
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

func TestSecurityCmd_Run_Validate_BothValid(t *testing.T) {
	cmd := &SecurityCmd{
		Action:        "validate",
		SecurityLevel: "trusted",
		FeatureSet:    "full",
	}

	var stdout bytes.Buffer
	cmd.Out = &stdout

	ctx := context.Background()
	err := cmd.Run(ctx)

	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "✓ Security level 'trusted' is valid")
	assert.Contains(t, output, "✓ Feature set 'full' is valid")
}
