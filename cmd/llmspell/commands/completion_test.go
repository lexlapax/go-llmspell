// ABOUTME: Tests for the completion command functionality.
// ABOUTME: Verifies shell detection and completion script generation.

package commands

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		list      bool
		wantError bool
		contains  []string
	}{
		{
			name:     "list shells",
			list:     true,
			contains: []string{"bash", "zsh", "fish", "powershell", "sh"},
		},
		{
			name:     "bash completion",
			shell:    "bash",
			contains: []string{"_llmspell_complete", "complete -F"},
		},
		{
			name:     "zsh completion",
			shell:    "zsh",
			contains: []string{"#compdef llmspell", "_llmspell"},
		},
		{
			name:     "fish completion",
			shell:    "fish",
			contains: []string{"complete -c llmspell"},
		},
		{
			name:     "powershell completion",
			shell:    "powershell",
			contains: []string{"Register-ArgumentCompleter"},
		},
		{
			name:     "sh reference",
			shell:    "sh",
			contains: []string{"This shell doesn't support programmable completion"},
		},
		{
			name:      "invalid shell",
			shell:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := &CompletionCmd{
				BaseCommand: BaseCommand{
					Out: &stdout,
					Err: &stderr,
				},
				Shell: tt.shell,
				List:  tt.list,
			}

			err := cmd.Run(context.Background())
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				output := stdout.String()
				for _, expected := range tt.contains {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

func TestCompletionCmd_detectShell(t *testing.T) {
	cmd := &CompletionCmd{}

	t.Run("detect from SHELL env", func(t *testing.T) {
		tests := []struct {
			shellEnv string
			want     string
		}{
			{"/bin/bash", "bash"},
			{"/usr/bin/zsh", "zsh"},
			{"/usr/local/bin/fish", "fish"},
			{"/bin/sh", "sh"},
		}

		for _, tt := range tests {
			t.Run(tt.shellEnv, func(t *testing.T) {
				oldShell := os.Getenv("SHELL")
				os.Setenv("SHELL", tt.shellEnv)
				defer os.Setenv("SHELL", oldShell)

				got := cmd.detectShell()
				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("detect from version env vars", func(t *testing.T) {
		// Clear SHELL env var first
		oldShell := os.Getenv("SHELL")
		os.Unsetenv("SHELL")
		defer os.Setenv("SHELL", oldShell)

		// Test BASH_VERSION
		oldBash := os.Getenv("BASH_VERSION")
		os.Setenv("BASH_VERSION", "5.0.0")
		defer os.Setenv("BASH_VERSION", oldBash)

		got := cmd.detectShell()
		assert.Equal(t, "bash", got)
	})

	t.Run("windows detection", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows-specific test")
		}

		// Test PowerShell detection
		oldPSPath := os.Getenv("PSModulePath")
		os.Setenv("PSModulePath", "C:\\something")
		defer os.Setenv("PSModulePath", oldPSPath)

		got := cmd.detectShell()
		assert.Equal(t, "powershell", got)
	})
}

func TestCompletionCmd_ListShells(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &CompletionCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
		},
	}

	err := cmd.listShells()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Supported shells:")
	assert.Contains(t, output, "bash")
	assert.Contains(t, output, "zsh")
	assert.Contains(t, output, "fish")
	assert.Contains(t, output, "powershell")
	assert.Contains(t, output, "sh")
	assert.Contains(t, output, "Usage:")
}

func TestCompletionCmd_AutoDetect(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := &CompletionCmd{
		BaseCommand: BaseCommand{
			Out: &stdout,
			Err: &stderr,
		},
		Shell: "", // No shell specified
	}

	// Set a known shell environment
	oldShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "/bin/bash")
	defer os.Setenv("SHELL", oldShell)

	err := cmd.Run(context.Background())
	require.NoError(t, err)

	// Should detect bash and generate completion
	output := stdout.String()
	assert.Contains(t, output, "_llmspell_complete")

	// The info message goes to stdout in this implementation
	assert.Contains(t, output, "Detected shell: bash")
}

func TestIsTerminal(t *testing.T) {
	// This is hard to test properly without mocking os.Stdout
	// Just ensure the function doesn't panic
	_ = isTerminal()
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "foo", false},
		{"foo", "", true},
		{"foo", "foo", true},
		{"/bin/bash", "bash", true},
		{"/usr/local/bin/zsh", "zsh", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			assert.Equal(t, tt.want, got)
		})
	}
}
