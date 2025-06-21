// ABOUTME: Helper utilities for integration tests including test spell creation and cleanup.
// ABOUTME: Provides common functionality for setting up and tearing down test environments.

package helpers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHelper provides utilities for integration tests
type TestHelper struct {
	t       *testing.T
	tempDir string
	binPath string
	cleanup []func()
	env     []string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	// Get binary path
	binPath := os.Getenv("LLMSPELL_TEST_BIN")
	if binPath == "" {
		// Build the binary if not provided
		binPath = filepath.Join(t.TempDir(), "llmspell")

		// Find project root by looking for go.mod
		projectRoot := findProjectRoot(t)

		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/llmspell")
		cmd.Dir = projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build llmspell: %v\n%s", err, output)
		}
	}

	return &TestHelper{
		t:       t,
		tempDir: t.TempDir(),
		binPath: binPath,
		cleanup: []func(){},
		env:     os.Environ(),
	}
}

// TempDir returns the test's temporary directory
func (h *TestHelper) TempDir() string {
	return h.tempDir
}

// BinPath returns the path to the llmspell binary
func (h *TestHelper) BinPath() string {
	return h.binPath
}

// SetEnv sets an environment variable for commands
func (h *TestHelper) SetEnv(key, value string) {
	h.env = append(h.env, fmt.Sprintf("%s=%s", key, value))
}

// RunCommand runs llmspell with the given arguments
func (h *TestHelper) RunCommand(args ...string) (string, string, error) {
	h.t.Helper()

	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.tempDir
	cmd.Env = h.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// RunCommandWithInput runs llmspell with stdin input
func (h *TestHelper) RunCommandWithInput(input string, args ...string) (string, string, error) {
	h.t.Helper()

	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.tempDir
	cmd.Env = h.env
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// RunCommandWithContext runs llmspell with a context
func (h *TestHelper) RunCommandWithContext(ctx context.Context, args ...string) (string, string, error) {
	h.t.Helper()

	cmd := exec.CommandContext(ctx, h.binPath, args...)
	cmd.Dir = h.tempDir
	cmd.Env = h.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// CreateSpell creates a test spell file
func (h *TestHelper) CreateSpell(name, content string) string {
	h.t.Helper()

	spellPath := filepath.Join(h.tempDir, name)
	require.NoError(h.t, os.WriteFile(spellPath, []byte(content), 0644))

	h.cleanup = append(h.cleanup, func() {
		os.Remove(spellPath)
	})

	return spellPath
}

// CreateSpellYAML creates a spell.yaml file
func (h *TestHelper) CreateSpellYAML(dir string, content string) string {
	h.t.Helper()

	if dir == "" {
		dir = h.tempDir
	}

	spellPath := filepath.Join(dir, "spell.yaml")
	require.NoError(h.t, os.MkdirAll(dir, 0755))
	require.NoError(h.t, os.WriteFile(spellPath, []byte(content), 0644))

	return spellPath
}

// CreateConfigFile creates a configuration file
func (h *TestHelper) CreateConfigFile(content string) string {
	h.t.Helper()

	configPath := filepath.Join(h.tempDir, "config.yaml")
	require.NoError(h.t, os.WriteFile(configPath, []byte(content), 0644))

	h.cleanup = append(h.cleanup, func() {
		os.Remove(configPath)
	})

	return configPath
}

// AssertSuccess asserts that a command succeeded
func (h *TestHelper) AssertSuccess(stdout, stderr string, err error) {
	h.t.Helper()

	if err != nil {
		h.t.Fatalf("Command failed: %v\nSTDOUT:\n%s\nSTDERR:\n%s", err, stdout, stderr)
	}
}

// AssertFailure asserts that a command failed
func (h *TestHelper) AssertFailure(stdout, stderr string, err error) {
	h.t.Helper()

	if err == nil {
		h.t.Fatalf("Expected command to fail but it succeeded\nSTDOUT:\n%s\nSTDERR:\n%s", stdout, stderr)
	}
}

// AssertOutput asserts that output contains expected string
func (h *TestHelper) AssertOutput(output, expected string) {
	h.t.Helper()

	if !strings.Contains(output, expected) {
		h.t.Fatalf("Output does not contain expected string\nExpected: %s\nActual:\n%s", expected, output)
	}
}

// AssertNotOutput asserts that output does not contain string
func (h *TestHelper) AssertNotOutput(output, unexpected string) {
	h.t.Helper()

	if strings.Contains(output, unexpected) {
		h.t.Fatalf("Output contains unexpected string: %s\nActual:\n%s", unexpected, output)
	}
}

// WaitForFile waits for a file to exist
func (h *TestHelper) WaitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for file: %s", path)
}

// Cleanup runs all cleanup functions
func (h *TestHelper) Cleanup() {
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}
}

// Basic spell templates for testing

// BasicLuaSpell returns a basic Lua spell for testing
func BasicLuaSpell() string {
	return `-- Test spell
local result = "Hello from Lua!"
print(result)
return result
`
}

// BasicSpellYAML returns a basic spell.yaml for testing
func BasicSpellYAML() string {
	return `name: test-spell
description: A test spell
author: Test Author
version: 1.0.0
engine: lua

security:
  profile: sandbox
  permissions:
    - file:read

parameters:
  message:
    type: string
    description: Test message
    default: "Hello"
`
}

// ErrorLuaSpell returns a Lua spell that errors
func ErrorLuaSpell() string {
	return `-- Error spell
error("This is a test error")
`
}

// ConfigYAML returns a test configuration
func ConfigYAML() string {
	return `debug: true
engine:
  default: lua
  timeout: 30s
  
repl:
  prompt: "test> "
  save_history: false
  
security:
  default_profile: development
`
}

// CreateFile creates a file with the given content at the specified path
func (h *TestHelper) CreateFile(path string, content string) {
	h.t.Helper()

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		require.NoError(h.t, err)
	}
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(h.t, err)
}

// StartCommand starts a command and returns it without waiting
func (h *TestHelper) StartCommand(args ...string) *exec.Cmd {
	h.t.Helper()

	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.tempDir
	cmd.Env = h.env

	err := cmd.Start()
	require.NoError(h.t, err)

	return cmd
}

// WaitCommand waits for a started command to complete
func (h *TestHelper) WaitCommand(cmd *exec.Cmd) (string, string, error) {
	h.t.Helper()

	err := cmd.Wait()
	return "", "", err
}

// findProjectRoot finds the project root by looking for go.mod
func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
