// ABOUTME: Tests for LLMSpell-specific man page generation.
// ABOUTME: Validates man page content and structure for all commands.

package docs

import (
	"strings"
	"testing"
)

func TestGenerateLLMSpellManPage(t *testing.T) {
	version := "1.0.0"
	man := GenerateLLMSpellManPage(version)

	// Test basic metadata
	if man.Name != "llmspell" {
		t.Errorf("Expected name 'llmspell', got '%s'", man.Name)
	}
	if man.Section != 1 {
		t.Errorf("Expected section 1, got %d", man.Section)
	}
	if man.Version != version {
		t.Errorf("Expected version '%s', got '%s'", version, man.Version)
	}

	// Test description
	if !strings.Contains(man.Description, "scriptable LLM interactions") {
		t.Error("Description should contain 'scriptable LLM interactions'")
	}

	// Test synopsis
	if !strings.Contains(man.Synopsis, "GLOBAL-OPTIONS") {
		t.Error("Synopsis should contain GLOBAL-OPTIONS")
	}

	// Test global options
	if len(man.Options) < 5 {
		t.Errorf("Expected at least 5 global options, got %d", len(man.Options))
	}

	// Verify specific global options exist
	expectedOptions := []string{"help", "debug", "config", "quiet", "verbose", "profile"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected || opt.Short == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected global option '%s' not found", expected)
		}
	}

	// Test commands
	if len(man.Commands) < 10 {
		t.Errorf("Expected at least 10 commands, got %d", len(man.Commands))
	}

	// Verify specific commands exist
	expectedCommands := []string{"run", "repl", "new", "validate", "config", "security", "engines", "debug", "version", "completion"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range man.Commands {
			if cmd.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' not found", expected)
		}
	}

	// Test examples
	if len(man.Examples) < 3 {
		t.Errorf("Expected at least 3 examples, got %d", len(man.Examples))
	}

	// Test files
	if len(man.Files) < 2 {
		t.Errorf("Expected at least 2 files, got %d", len(man.Files))
	}

	// Test see also
	if len(man.SeeAlso) < 1 {
		t.Errorf("Expected at least 1 see also entry, got %d", len(man.SeeAlso))
	}

	// Test bugs field
	if !strings.Contains(man.Bugs, "github.com") {
		t.Error("Bugs field should contain GitHub URL")
	}
}

func TestGenerateCommandManPage_Run(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("run", version)
	if err != nil {
		t.Fatalf("Error generating run command man page: %v", err)
	}

	if man.Name != "llmspell-run" {
		t.Errorf("Expected name 'llmspell-run', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "execute a spell script") {
		t.Error("Description should contain 'execute a spell script'")
	}

	// Test run-specific options
	expectedOptions := []string{"parameters", "engine", "timeout"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected run option '%s' not found", expected)
		}
	}

	// Test examples
	if len(man.Examples) < 2 {
		t.Errorf("Expected at least 2 examples for run command, got %d", len(man.Examples))
	}
}

func TestGenerateCommandManPage_REPL(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("repl", version)
	if err != nil {
		t.Fatalf("Error generating repl command man page: %v", err)
	}

	if man.Name != "llmspell-repl" {
		t.Errorf("Expected name 'llmspell-repl', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "interactive REPL") {
		t.Error("Description should contain 'interactive REPL'")
	}

	// Test REPL-specific options
	expectedOptions := []string{"engine", "no-history", "no-highlight", "history-file"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected repl option '%s' not found", expected)
		}
	}

	// Test REPL commands in description
	if !strings.Contains(man.Description, ".help") {
		t.Error("Description should contain REPL commands like .help")
	}
}

func TestGenerateCommandManPage_New(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("new", version)
	if err != nil {
		t.Fatalf("Error generating new command man page: %v", err)
	}

	if man.Name != "llmspell-new" {
		t.Errorf("Expected name 'llmspell-new', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "create a new spell") {
		t.Error("Description should contain 'create a new spell'")
	}

	// Test template types in description
	templateTypes := []string{"basic", "advanced", "agent", "workflow", "interactive"}
	for _, template := range templateTypes {
		if !strings.Contains(man.Description, template) {
			t.Errorf("Description should contain template type '%s'", template)
		}
	}
}

func TestGenerateCommandManPage_Validate(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("validate", version)
	if err != nil {
		t.Fatalf("Error generating validate command man page: %v", err)
	}

	if man.Name != "llmspell-validate" {
		t.Errorf("Expected name 'llmspell-validate', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "validate a spell") {
		t.Error("Description should contain 'validate a spell'")
	}

	// Test validate-specific options
	expectedOptions := []string{"engine", "strict", "security"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected validate option '%s' not found", expected)
		}
	}
}

func TestGenerateCommandManPage_Config(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("config", version)
	if err != nil {
		t.Fatalf("Error generating config command man page: %v", err)
	}

	if man.Name != "llmspell-config" {
		t.Errorf("Expected name 'llmspell-config', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "manage configuration") {
		t.Error("Description should contain 'manage configuration'")
	}

	// Test examples
	if len(man.Examples) < 3 {
		t.Errorf("Expected at least 3 examples for config command, got %d", len(man.Examples))
	}
}

func TestGenerateCommandManPage_Security(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("security", version)
	if err != nil {
		t.Fatalf("Error generating security command man page: %v", err)
	}

	if man.Name != "llmspell-security" {
		t.Errorf("Expected name 'llmspell-security', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "manage security profiles") {
		t.Error("Description should contain 'manage security profiles'")
	}

	// Test examples
	if len(man.Examples) < 2 {
		t.Errorf("Expected at least 2 examples for security command, got %d", len(man.Examples))
	}
}

func TestGenerateCommandManPage_Engines(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("engines", version)
	if err != nil {
		t.Fatalf("Error generating engines command man page: %v", err)
	}

	if man.Name != "llmspell-engines" {
		t.Errorf("Expected name 'llmspell-engines', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "script engines") {
		t.Error("Description should contain 'script engines'")
	}
}

func TestGenerateCommandManPage_Debug(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("debug", version)
	if err != nil {
		t.Fatalf("Error generating debug command man page: %v", err)
	}

	if man.Name != "llmspell-debug" {
		t.Errorf("Expected name 'llmspell-debug', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "debug a spell script") {
		t.Error("Description should contain 'debug a spell script'")
	}

	// Test debug-specific options
	expectedOptions := []string{"breakpoint", "watch"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected debug option '%s' not found", expected)
		}
	}
}

func TestGenerateCommandManPage_Version(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("version", version)
	if err != nil {
		t.Fatalf("Error generating version command man page: %v", err)
	}

	if man.Name != "llmspell-version" {
		t.Errorf("Expected name 'llmspell-version', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "show version information") {
		t.Error("Description should contain 'show version information'")
	}

	// Test version-specific options
	expectedOptions := []string{"short", "build-info", "deps", "check-compat"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version option '%s' not found", expected)
		}
	}
}

func TestGenerateCommandManPage_Completion(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("completion", version)
	if err != nil {
		t.Fatalf("Error generating completion command man page: %v", err)
	}

	if man.Name != "llmspell-completion" {
		t.Errorf("Expected name 'llmspell-completion', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "shell completion scripts") {
		t.Error("Description should contain 'shell completion scripts'")
	}

	// Test examples
	if len(man.Examples) < 2 {
		t.Errorf("Expected at least 2 examples for completion command, got %d", len(man.Examples))
	}
}

func TestGenerateCommandManPage_Man(t *testing.T) {
	version := "1.0.0"
	man, err := GenerateCommandManPage("man", version)
	if err != nil {
		t.Fatalf("Error generating man command man page: %v", err)
	}

	if man.Name != "llmspell-man" {
		t.Errorf("Expected name 'llmspell-man', got '%s'", man.Name)
	}

	if !strings.Contains(man.Description, "generate man pages") {
		t.Error("Description should contain 'generate man pages'")
	}

	// Test man-specific options
	expectedOptions := []string{"output", "dir", "all", "install", "format"}
	for _, expected := range expectedOptions {
		found := false
		for _, opt := range man.Options {
			if opt.Long == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected man option '%s' not found", expected)
		}
	}
}

func TestGenerateCommandManPage_UnknownCommand(t *testing.T) {
	version := "1.0.0"
	_, err := GenerateCommandManPage("unknown", version)
	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' in error, got: %v", err)
	}
}

func TestGetAllCommands(t *testing.T) {
	commands := GetAllCommands()

	// Test minimum number of commands
	if len(commands) < 10 {
		t.Errorf("Expected at least 10 commands, got %d", len(commands))
	}

	// Test that all expected commands are present
	expectedCommands := []string{
		"run", "repl", "new", "validate", "config",
		"security", "engines", "debug", "version", "completion", "man",
	}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' not found in GetAllCommands()", expected)
		}
	}

	// Test that we can generate man pages for all commands
	for _, cmd := range commands {
		_, err := GenerateCommandManPage(cmd, "test-version")
		if err != nil {
			t.Errorf("Failed to generate man page for command '%s': %v", cmd, err)
		}
	}
}

func TestCommandManPageConsistency(t *testing.T) {
	version := "1.0.0"
	commands := GetAllCommands()

	for _, cmd := range commands {
		man, err := GenerateCommandManPage(cmd, version)
		if err != nil {
			t.Errorf("Failed to generate man page for command '%s': %v", cmd, err)
			continue
		}

		// Test that all command man pages have required fields
		if man.Name == "" {
			t.Errorf("Command '%s' man page missing name", cmd)
		}
		if man.Description == "" {
			t.Errorf("Command '%s' man page missing description", cmd)
		}
		if man.Synopsis == "" {
			t.Errorf("Command '%s' man page missing synopsis", cmd)
		}

		// Test that man page name follows convention
		expectedName := "llmspell-" + cmd
		if man.Name != expectedName {
			t.Errorf("Command '%s' man page has incorrect name: expected '%s', got '%s'", cmd, expectedName, man.Name)
		}

		// Test that files section includes config file
		foundConfigFile := false
		for _, file := range man.Files {
			if strings.Contains(file, "config.yaml") {
				foundConfigFile = true
				break
			}
		}
		if !foundConfigFile {
			t.Errorf("Command '%s' man page missing config file in files section", cmd)
		}

		// Test that see also includes main llmspell reference
		foundMainRef := false
		for _, ref := range man.SeeAlso {
			if strings.Contains(ref, "llmspell(1)") {
				foundMainRef = true
				break
			}
		}
		if !foundMainRef {
			t.Errorf("Command '%s' man page missing llmspell(1) reference in see also", cmd)
		}
	}
}
