// ABOUTME: Integration tests for the security command verifying security profile management.
// ABOUTME: Tests profile listing, viewing, and enforcement.

package commands

import (
	"testing"

	"github.com/lexlapax/go-llmspell/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestSecurityCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("list security profiles", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "list")

		h.AssertSuccess(stdout, stderr, err)
		// Should list all available profiles
		h.AssertOutput(stdout, "sandbox")
		h.AssertOutput(stdout, "development")
		h.AssertOutput(stdout, "production")
	})

	t.Run("view specific profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "view", "sandbox")

		h.AssertSuccess(stdout, stderr, err)
		// Should show profile details
		h.AssertOutput(stdout, "sandbox")
		h.AssertOutput(stdout, "file_system:")
		h.AssertOutput(stdout, "network:")
		h.AssertOutput(stdout, "external_commands:")
	})

	t.Run("view development profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "view", "development")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "development")
		// Development should be more permissive
		assert.Contains(t, stdout, "read_write") // File system access
	})

	t.Run("view production profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "view", "production")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "production")
		// Production should have restrictions
		assert.Contains(t, stdout, "restricted") // Limited access
	})

	t.Run("validate security profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "validate", "sandbox")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("invalid profile name", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "view", "non-existent")

		h.AssertFailure(stdout, stderr, err)
		assert.Contains(t, stderr, "unknown profile")
	})

	t.Run("check permissions for profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "check", "sandbox", "file_read")

		h.AssertSuccess(stdout, stderr, err)
		// Should indicate if permission is allowed
		output := stdout + stderr
		assert.Contains(t, output, "denied") // Sandbox denies file operations
	})

	t.Run("compare profiles", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "compare", "sandbox", "development")

		h.AssertSuccess(stdout, stderr, err)
		// Should show differences
		h.AssertOutput(stdout, "sandbox")
		h.AssertOutput(stdout, "development")
		h.AssertOutput(stdout, "file_system")
		h.AssertOutput(stdout, "network")
	})

	t.Run("export profile", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "export", "sandbox")

		h.AssertSuccess(stdout, stderr, err)
		// Should output YAML representation
		h.AssertOutput(stdout, "name: sandbox")
		h.AssertOutput(stdout, "file_system:")
		h.AssertOutput(stdout, "network:")
	})

	t.Run("verbose profile info", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "view", "sandbox", "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose mode should show more details
		h.AssertOutput(stdout, "sandbox")
		output := stdout + stderr
		assert.Contains(t, output, "description") // Should include descriptions
	})
}

func TestSecurityEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("run with sandbox profile", func(t *testing.T) {
		// Create a script that tries to access file system
		script := h.CreateSpell("security-test.lua", `
			-- Try to read a file (should be blocked in sandbox)
			local file = io.open("/etc/passwd", "r")
			if file then
				print("SECURITY BREACH: File access allowed!")
				file:close()
			else
				print("File access properly blocked")
			end
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--profile", "sandbox")

		// Script should run but file access should be blocked
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "File access properly blocked")
		h.AssertNotOutput(stdout, "SECURITY BREACH")
	})

	t.Run("run with development profile", func(t *testing.T) {
		// Create a script that uses development features
		script := h.CreateSpell("dev-test.lua", `
			-- Development profile allows more access
			print("Development mode active")
			-- Would have more permissive access here
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--profile", "development")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Development mode active")
	})

	t.Run("validate script against profile", func(t *testing.T) {
		// Create a script with potential security issues
		script := h.CreateSpell("risky.lua", `
			os.execute("rm -rf /")  -- Dangerous!
			io.popen("curl evil.com")  -- Network access
		`)

		stdout, stderr, _ := h.RunCommand("validate", script, "--profile", "sandbox")

		// Validation should warn about security issues
		output := stdout + stderr
		assert.Contains(t, output, "security")   // Should mention security concerns
		assert.Contains(t, output, "os.execute") // Should identify risky calls
	})

	t.Run("profile from config", func(t *testing.T) {
		// Create config with default profile
		config := h.CreateConfigFile(`
security:
  profile: production
`)

		// Create a simple script
		script := h.CreateSpell("config-profile.lua", `
			print("Using profile from config")
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--config", config)

		h.AssertSuccess(stdout, stderr, err)
		// Should use production profile from config
		h.AssertOutput(stdout, "Using profile from config")
	})
}
