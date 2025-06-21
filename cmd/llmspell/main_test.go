// ABOUTME: Tests for the llmspell CLI main entry point and Kong setup.
// ABOUTME: Verifies CLI structure, command parsing, and flag handling.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	t.Run("parse_run_command", func(t *testing.T) {
		// Create temp script file
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "test.lua")
		err := os.WriteFile(scriptFile, []byte("print('test')"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err = parser.Parse([]string{"run", scriptFile})
		require.NoError(t, err)

		assert.Equal(t, scriptFile, cli.Run.Script)
	})

	t.Run("parse_validate_command", func(t *testing.T) {
		// Create temp spell file
		tmpDir := t.TempDir()
		spellFile := filepath.Join(tmpDir, "spell.yaml")
		err := os.WriteFile(spellFile, []byte("name: test"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err = parser.Parse([]string{"validate", spellFile})
		require.NoError(t, err)

		assert.Equal(t, spellFile, cli.Validate.Path)
	})

	t.Run("parse_global_flags", func(t *testing.T) {
		// Create temp script file
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "test.lua")
		err := os.WriteFile(scriptFile, []byte("print('test')"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err = parser.Parse([]string{"--debug", "--config", "custom.yaml", "run", scriptFile})
		require.NoError(t, err)

		assert.True(t, cli.DebugMode)
		// Kong converts relative paths to absolute
		assert.Contains(t, cli.ConfigFile, "custom.yaml")
	})

	t.Run("parse_engines_command", func(t *testing.T) {
		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err := parser.Parse([]string{"engines", "--details"})
		require.NoError(t, err)

		assert.True(t, cli.Engines.Details)
	})

	t.Run("parse_version_command", func(t *testing.T) {
		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err := parser.Parse([]string{"version"})
		require.NoError(t, err)
		// Version command has no specific fields to check
	})

	t.Run("parse_repl_command", func(t *testing.T) {
		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err := parser.Parse([]string{"repl", "--engine", "javascript"})
		require.NoError(t, err)

		assert.Equal(t, "javascript", cli.REPL.Engine)
	})

	t.Run("parse_config_command", func(t *testing.T) {
		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err := parser.Parse([]string{"config", "show"})
		require.NoError(t, err)

		assert.Equal(t, "show", cli.Config.Action)
	})

	t.Run("parse_security_command", func(t *testing.T) {
		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err := parser.Parse([]string{"security", "list"})
		require.NoError(t, err)

		assert.Equal(t, "list", cli.Security.Action)
	})

	t.Run("parse_debug_command", func(t *testing.T) {
		// Create temp script file
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "test.lua")
		err := os.WriteFile(scriptFile, []byte("print('test')"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err = parser.Parse([]string{"debug", scriptFile, "--breakpoints", "10"})
		require.NoError(t, err)

		assert.Equal(t, scriptFile, cli.Debug.Script)
		assert.Contains(t, cli.Debug.Breakpoints, 10)
	})
}

func TestCLIHelp(t *testing.T) {
	t.Run("help_output", func(t *testing.T) {
		cli := &CLI{}

		var buf bytes.Buffer
		parser := mustNewParserWithOutput(t, cli, &buf)

		_, err := parser.Parse([]string{"--help"})
		// Help causes an expected error
		assert.Error(t, err)

		output := buf.String()
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "run")
		assert.Contains(t, output, "validate")
		assert.Contains(t, output, "engines")
	})

	t.Run("command_help", func(t *testing.T) {
		cli := &CLI{}

		var buf bytes.Buffer
		parser := mustNewParserWithOutput(t, cli, &buf)

		_, err := parser.Parse([]string{"run", "--help"})
		assert.Error(t, err) // Help causes expected error

		output := buf.String()
		assert.Contains(t, output, "run <script>")
		assert.Contains(t, output, "Execute a spell script")
	})
}

func TestGlobalFlags(t *testing.T) {
	t.Run("verbosity_flags", func(t *testing.T) {
		// Create temp script file
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "test.lua")
		err := os.WriteFile(scriptFile, []byte("print('test')"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		// Test quiet
		_, err = parser.Parse([]string{"--quiet", "run", scriptFile})
		require.NoError(t, err)
		assert.True(t, cli.Quiet)
		assert.False(t, cli.Verbose)

		// Test verbose
		cli = &CLI{}
		parser = mustNewParser(t, cli)

		_, err = parser.Parse([]string{"--verbose", "run", scriptFile})
		require.NoError(t, err)
		assert.True(t, cli.Verbose)
		assert.False(t, cli.Quiet)
	})

	t.Run("profile_flag", func(t *testing.T) {
		// Create temp script file
		tmpDir := t.TempDir()
		scriptFile := filepath.Join(tmpDir, "test.lua")
		err := os.WriteFile(scriptFile, []byte("print('test')"), 0644)
		require.NoError(t, err)

		cli := &CLI{}
		parser := mustNewParser(t, cli)

		_, err = parser.Parse([]string{"--profile", "development", "run", scriptFile})
		require.NoError(t, err)

		assert.Equal(t, "development", cli.Profile)
	})
}

func TestConfigFile(t *testing.T) {
	t.Run("config_file_location", func(t *testing.T) {
		// Test default config location
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		expectedDefault := filepath.Join(home, ".config", "llmspell", "config.yaml")
		assert.Equal(t, expectedDefault, defaultConfigPath())

		// Test XDG_CONFIG_HOME
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		expectedXDG := filepath.Join("/custom/config", "llmspell", "config.yaml")
		assert.Equal(t, expectedXDG, defaultConfigPath())
	})
}

func TestVersionInfo(t *testing.T) {
	t.Run("version_vars", func(t *testing.T) {
		// These are set during build
		assert.NotEmpty(t, version)

		versionInfo := formatVersion()
		assert.Contains(t, versionInfo, version)
		if gitCommit != "" && len(gitCommit) >= 7 {
			assert.Contains(t, versionInfo, gitCommit[:7])
		}
		if buildDate != "" {
			assert.Contains(t, versionInfo, buildDate)
		}
	})
}

func TestHelpers(t *testing.T) {
	t.Run("expand_path", func(t *testing.T) {
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		// Test tilde expansion
		expanded := expandPath("~/test")
		assert.Equal(t, filepath.Join(home, "test"), expanded)

		// Test absolute path
		expanded = expandPath("/absolute/path")
		assert.Equal(t, "/absolute/path", expanded)

		// Test relative path
		expanded = expandPath("relative/path")
		assert.Equal(t, "relative/path", expanded)
	})

	t.Run("file_exists", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(tmpFile, []byte("test"), 0644)
		require.NoError(t, err)

		assert.True(t, fileExists(tmpFile))
		assert.False(t, fileExists(filepath.Join(tmpDir, "nonexistent.txt")))
	})
}

// Helper functions for testing
func mustNewParser(t *testing.T, cli *CLI) *kong.Kong {
	parser, err := kong.New(cli)
	require.NoError(t, err)
	return parser
}

func mustNewParserWithOutput(t *testing.T, cli *CLI, w *bytes.Buffer) *kong.Kong {
	parser, err := kong.New(cli,
		kong.Writers(w, w),
		kong.Exit(func(int) {}),
	)
	require.NoError(t, err)
	return parser
}

// Benchmark CLI parsing
func BenchmarkCLIParsing(b *testing.B) {
	cli := &CLI{}
	parser, _ := kong.New(cli)
	args := []string{"--debug", "--config", "test.yaml", "run", "script.lua", "-p", "key=value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(args)
	}
}
