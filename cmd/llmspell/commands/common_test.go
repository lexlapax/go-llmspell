// ABOUTME: Unit tests for common command utilities and validation helpers.
// ABOUTME: Tests context extraction, validation functions, and output helpers.

package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/lexlapax/go-llmspell/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	t.Run("returns config from context", func(t *testing.T) {
		cfg := &config.Config{Debug: true}
		ctx := context.WithValue(context.Background(), ConfigKey, cfg)

		result := GetConfig(ctx)
		assert.Equal(t, cfg, result)
	})

	t.Run("returns default config when not in context", func(t *testing.T) {
		ctx := context.Background()

		result := GetConfig(ctx)
		assert.NotNil(t, result)
		assert.False(t, result.Debug)
	})
}

func TestIsDebug(t *testing.T) {
	t.Run("returns true when debug is enabled", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), DebugKey, true)
		assert.True(t, IsDebug(ctx))
	})

	t.Run("returns false when debug is disabled", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), DebugKey, false)
		assert.False(t, IsDebug(ctx))
	})

	t.Run("returns false when not in context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, IsDebug(ctx))
	})
}

func TestIsVerbose(t *testing.T) {
	t.Run("returns true when verbose is enabled", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), VerboseKey, true)
		assert.True(t, IsVerbose(ctx))
	})

	t.Run("returns false when verbose is disabled", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), VerboseKey, false)
		assert.False(t, IsVerbose(ctx))
	})

	t.Run("returns false when not in context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, IsVerbose(ctx))
	})
}

func TestGetSecurityLevel(t *testing.T) {
	t.Run("returns security level from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SecurityLevelKey, security.SecurityLevelPrivileged)
		assert.Equal(t, security.SecurityLevelPrivileged, GetSecurityLevel(ctx))
	})

	t.Run("returns security level from string in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SecurityLevelKey, "untrusted")
		assert.Equal(t, security.SecurityLevelUntrusted, GetSecurityLevel(ctx))
	})

	t.Run("returns trusted as default when not in context", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, security.SecurityLevelTrusted, GetSecurityLevel(ctx))
	})
}

func TestGetFeatureSet(t *testing.T) {
	t.Run("returns feature set from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), FeatureSetKey, registry.FeatureSetMinimal)
		assert.Equal(t, registry.FeatureSetMinimal, GetFeatureSet(ctx))
	})

	t.Run("returns feature set from string in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), FeatureSetKey, "agent")
		assert.Equal(t, registry.FeatureSetAgent, GetFeatureSet(ctx))
	})

	t.Run("returns full as default when not in context", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, registry.FeatureSetFull, GetFeatureSet(ctx))
	})
}

func TestValidateSecurityLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  security.SecurityLevel
		wantError bool
	}{
		{"valid untrusted", "untrusted", security.SecurityLevelUntrusted, false},
		{"valid trusted", "trusted", security.SecurityLevelTrusted, false},
		{"valid privileged", "privileged", security.SecurityLevelPrivileged, false},
		{"invalid empty", "", "", true},
		{"invalid unknown", "unknown", "", true},
		{"invalid old name", "sandbox", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSecurityLevel(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid security level")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateFeatureSet(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  registry.FeatureSet
		wantError bool
	}{
		{"valid minimal", "minimal", registry.FeatureSetMinimal, false},
		{"valid llm", "llm", registry.FeatureSetLLM, false},
		{"valid agent", "agent", registry.FeatureSetAgent, false},
		{"valid observable", "observable", registry.FeatureSetObservable, false},
		{"valid full", "full", registry.FeatureSetFull, false},
		{"invalid empty", "", "", true},
		{"invalid unknown", "unknown", "", true},
		{"invalid old name", "standard", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateFeatureSet(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid feature set")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetAvailableSecurityLevels(t *testing.T) {
	levels := GetAvailableSecurityLevels()
	assert.Equal(t, 3, len(levels))
	assert.Contains(t, levels, "untrusted")
	assert.Contains(t, levels, "trusted")
	assert.Contains(t, levels, "privileged")
}

func TestGetAvailableFeatureSets(t *testing.T) {
	sets := GetAvailableFeatureSets()
	assert.Equal(t, 5, len(sets))
	assert.Contains(t, sets, "minimal")
	assert.Contains(t, sets, "llm")
	assert.Contains(t, sets, "agent")
	assert.Contains(t, sets, "observable")
	assert.Contains(t, sets, "full")
}

func TestGetSecurityLevelDescription(t *testing.T) {
	t.Run("returns description for valid levels", func(t *testing.T) {
		desc := GetSecurityLevelDescription(security.SecurityLevelUntrusted)
		assert.NotEmpty(t, desc)
		assert.Contains(t, desc, "untrusted")

		desc = GetSecurityLevelDescription(security.SecurityLevelTrusted)
		assert.NotEmpty(t, desc)
		assert.Contains(t, desc, "trusted")

		desc = GetSecurityLevelDescription(security.SecurityLevelPrivileged)
		assert.NotEmpty(t, desc)
		assert.Contains(t, desc, "system")
	})

	t.Run("returns unknown for invalid level", func(t *testing.T) {
		desc := GetSecurityLevelDescription(security.SecurityLevel("invalid"))
		assert.Equal(t, "Unknown security level", desc)
	})
}

func TestGetFeatureSetDescription(t *testing.T) {
	t.Run("returns description for valid feature sets", func(t *testing.T) {
		desc := GetFeatureSetDescription(registry.FeatureSetMinimal)
		assert.NotEmpty(t, desc)

		desc = GetFeatureSetDescription(registry.FeatureSetFull)
		assert.NotEmpty(t, desc)
	})

	t.Run("returns unknown for invalid feature set", func(t *testing.T) {
		desc := GetFeatureSetDescription(registry.FeatureSet("invalid"))
		assert.Equal(t, "Unknown feature set", desc)
	})
}

func TestBaseCommand(t *testing.T) {
	t.Run("Printf writes to configured output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Out: buf}
		cmd.Printf("Hello %s", "World")
		assert.Equal(t, "Hello World", buf.String())
	})

	t.Run("Println writes to configured output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Out: buf}
		cmd.Println("Hello", "World")
		assert.Equal(t, "Hello World\n", buf.String())
	})

	t.Run("Errorf writes to configured error output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Err: buf}
		cmd.Errorf("Error: %s", "test")
		assert.Equal(t, "Error: test", buf.String())
	})

	t.Run("Info adds newline", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Out: buf}
		cmd.Info(context.Background(), "Info: %s", "test")
		assert.Equal(t, "Info: test\n", buf.String())
	})

	t.Run("Debug only prints when debug is enabled", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Out: buf}

		// Debug disabled
		ctx := context.Background()
		cmd.Debug(ctx, "Debug message")
		assert.Empty(t, buf.String())

		// Debug enabled
		buf.Reset()
		ctx = context.WithValue(ctx, DebugKey, true)
		cmd.Debug(ctx, "Debug message")
		assert.Equal(t, "[DEBUG] Debug message\n", buf.String())
	})

	t.Run("Error writes to error output with prefix", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Err: buf}
		cmd.Error(context.Background(), "Error: %s", "test")
		assert.Equal(t, "[ERROR] Error: test\n", buf.String())
	})

	t.Run("Verbose only prints when verbose is enabled", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd := &BaseCommand{Out: buf}

		// Verbose disabled
		ctx := context.Background()
		cmd.Verbose(ctx, "Verbose message")
		assert.Empty(t, buf.String())

		// Verbose enabled
		buf.Reset()
		ctx = context.WithValue(ctx, VerboseKey, true)
		cmd.Verbose(ctx, "Verbose message")
		assert.Equal(t, "Verbose message\n", buf.String())
	})
}

func TestTableWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	table := NewTableWriter(buf, "Name", "Age", "City")

	table.AddRow("Alice", "30", "New York")
	table.AddRow("Bob", "25", "San Francisco")
	table.AddRow("Charlie", "35", "Los Angeles")

	table.Render()

	output := buf.String()
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Age")
	assert.Contains(t, output, "City")
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "30")
	assert.Contains(t, output, "New York")
	assert.Contains(t, output, "-------") // Separator line
}
