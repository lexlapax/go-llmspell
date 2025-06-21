// ABOUTME: Implementation of the completion command for generating shell completions.
// ABOUTME: Supports bash, zsh, fish, powershell, and outputs installation instructions.

package commands

import (
	"context"
	"os"
	"runtime"

	"github.com/lexlapax/go-llmspell/pkg/shell"
)

// CompletionCmd generates shell completion scripts
type CompletionCmd struct {
	BaseCommand
	Shell string `arg:"" optional:"" help:"Shell type (bash, zsh, fish, powershell, sh)"`
	List  bool   `short:"l" help:"List supported shells"`
}

// Run executes the completion command
func (c *CompletionCmd) Run(ctx context.Context) error {
	if c.List {
		return c.listShells()
	}

	// If no shell specified, try to detect
	if c.Shell == "" {
		detectedShell := c.detectShell()
		if detectedShell == "" {
			c.Error(ctx, "Could not detect shell. Please specify: bash, zsh, fish, powershell, or sh")
			return ctx.Err()
		}
		c.Shell = detectedShell
		c.Info(ctx, "Detected shell: %s", c.Shell)
	}

	// Parse shell type
	shellType, err := shell.ParseShell(c.Shell)
	if err != nil {
		c.Error(ctx, "Invalid shell: %s. Supported shells: bash, zsh, fish, powershell, sh", c.Shell)
		return err
	}

	// Generate completion
	gen := shell.LLMSpellCommands()
	script, err := gen.Generate(shellType)
	if err != nil {
		c.Error(ctx, "Failed to generate completion: %v", err)
		return err
	}

	// Output the script
	c.Println(script)

	// Show installation instructions to stderr if outputting to terminal
	if isTerminal() {
		c.Debug(ctx, "\n# Installation instructions:")
		instructions := shell.InstallInstructions(shellType)
		c.Debug(ctx, instructions)
	}

	return nil
}

// listShells lists all supported shells
func (c *CompletionCmd) listShells() error {
	c.Println("Supported shells:")
	for _, s := range shell.GetSupportedShells() {
		c.Printf("  %s\n", s)
	}
	c.Println("\nUsage:")
	c.Println("  llmspell completion <shell>")
	c.Println("  llmspell completion bash > ~/.local/share/bash-completion/completions/llmspell")
	c.Println("  llmspell completion zsh > ~/.zsh/completions/_llmspell")
	c.Println("  llmspell completion fish > ~/.config/fish/completions/llmspell.fish")
	return nil
}

// detectShell attempts to detect the current shell
func (c *CompletionCmd) detectShell() string {
	// First try SHELL environment variable
	if shellEnv := os.Getenv("SHELL"); shellEnv != "" {
		// Extract just the shell name from path
		if runtime.GOOS == "windows" {
			// On Windows, might be something like C:\Program Files\Git\bin\bash.exe
			for _, shell := range []string{"bash", "zsh", "fish", "sh"} {
				if contains(shellEnv, shell) {
					return shell
				}
			}
		} else {
			// On Unix-like systems, usually /bin/bash, /usr/bin/zsh, etc.
			switch {
			case contains(shellEnv, "bash"):
				return "bash"
			case contains(shellEnv, "zsh"):
				return "zsh"
			case contains(shellEnv, "fish"):
				return "fish"
			case contains(shellEnv, "sh"):
				return "sh"
			}
		}
	}

	// On Windows, check if we're in PowerShell
	if runtime.GOOS == "windows" {
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
	}

	// Note: parent process detection is platform-specific and complex,
	// so we default to bash if shell cannot be detected

	// Check for shell-specific environment variables
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("FISH_VERSION") != "" {
		return "fish"
	}

	return ""
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || contains(s[1:], substr))
}
