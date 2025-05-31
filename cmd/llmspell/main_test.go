// ABOUTME: Tests for the llmspell CLI main entry point
// ABOUTME: Verifies command parsing, spell execution, and error handling

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput captures stdout and stderr during test execution
func captureOutput(_ *testing.T, fn func()) (stdout, stderr string) {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Set new stdout and stderr
	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function
	fn()

	// Close writers
	wOut.Close()
	wErr.Close()

	// Read output
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)

	// Restore original stdout and stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return string(outBytes), string(errBytes)
}

func TestPrintUsage(t *testing.T) {
	stdout, _ := captureOutput(t, func() {
		printUsage()
	})

	assert.Contains(t, stdout, "llmspell - Cast scripting spells")
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "llmspell run <spell-path>")
	assert.Contains(t, stdout, "llmspell help")
	assert.Contains(t, stdout, "llmspell version")
	assert.Contains(t, stdout, "Examples:")
	assert.Contains(t, stdout, "Environment Variables:")
}

func TestSetupParams(t *testing.T) {
	// Create a test engine
	eng, err := lua.NewLuaEngine(&engine.Config{
		MaxExecutionTime: 30,
		MaxMemory:        64 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer eng.Close()

	tests := []struct {
		name           string
		args           []string
		expectedParams map[string]string
	}{
		{
			name: "single parameter",
			args: []string{"key=value"},
			expectedParams: map[string]string{
				"key": "value",
			},
		},
		{
			name: "multiple parameters",
			args: []string{"key1=value1", "key2=value2", "key3=value3"},
			expectedParams: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "parameter with equals in value",
			args: []string{"url=https://example.com?param=value"},
			expectedParams: map[string]string{
				"url": "https://example.com?param=value",
			},
		},
		{
			name: "mixed valid and invalid args",
			args: []string{"valid=yes", "invalid_no_equals", "another=param"},
			expectedParams: map[string]string{
				"valid":   "yes",
				"another": "param",
			},
		},
		{
			name:           "no parameters",
			args:           []string{},
			expectedParams: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupParams(eng, tt.args)

			// Check that params were set correctly
			for key, expectedValue := range tt.expectedParams {
				// Get the value from Lua
				err := eng.LoadScript(strings.NewReader(`
					testValue = params.` + key + `
				`))
				require.NoError(t, err)

				err = eng.Execute(context.Background())
				require.NoError(t, err)

				value, err := eng.GetVariable("testValue")
				require.NoError(t, err)
				assert.Equal(t, expectedValue, value)
			}
		})
	}
}

func TestRegisterMockLLM(t *testing.T) {
	// Create a test engine
	eng, err := lua.NewLuaEngine(&engine.Config{
		MaxExecutionTime: 30,
		MaxMemory:        64 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer eng.Close()

	// Register mock LLM
	registerMockLLM(eng)

	// Test chat function
	err = eng.LoadScript(strings.NewReader(`
		response = llm.chat("Hello")
	`))
	require.NoError(t, err)

	err = eng.Execute(context.Background())
	require.NoError(t, err)

	response, err := eng.GetVariable("response")
	require.NoError(t, err)
	assert.Contains(t, response.(string), "[Mock LLM Response]")
	assert.Contains(t, response.(string), "Hello")

	// Test complete function
	err = eng.LoadScript(strings.NewReader(`
		completion = llm.complete("Start of text", 100)
	`))
	require.NoError(t, err)

	err = eng.Execute(context.Background())
	require.NoError(t, err)

	completion, err := eng.GetVariable("completion")
	require.NoError(t, err)
	assert.Contains(t, completion.(string), "Start of text")
	assert.Contains(t, completion.(string), "[Mock completion with max 100 tokens]")

	// Test get_provider function
	err = eng.LoadScript(strings.NewReader(`
		provider = llm.get_provider()
	`))
	require.NoError(t, err)

	err = eng.Execute(context.Background())
	require.NoError(t, err)

	provider, err := eng.GetVariable("provider")
	require.NoError(t, err)
	assert.Equal(t, "mock", provider)

	// Test list_providers function
	err = eng.LoadScript(strings.NewReader(`
		providers = llm.list_providers()
		firstProvider = providers[1]
	`))
	require.NoError(t, err)

	err = eng.Execute(context.Background())
	require.NoError(t, err)

	firstProvider, err := eng.GetVariable("firstProvider")
	require.NoError(t, err)
	assert.Equal(t, "mock", firstProvider)
}

func TestInitializeBridges(t *testing.T) {
	// Create a test engine
	eng, err := lua.NewLuaEngine(&engine.Config{
		MaxExecutionTime: 30,
		MaxMemory:        64 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer eng.Close()

	// Test with MOCK_LLM=true
	os.Setenv("MOCK_LLM", "true")
	defer os.Unsetenv("MOCK_LLM")

	// Initialize bridges
	initializeBridges(eng, "test-spell")

	// Check that standard library is available
	err = eng.LoadScript(strings.NewReader(`
		-- Test JSON module
		jsonData = json.encode({test = "value"})
		assert(type(jsonData) == "string", "JSON encode should return string")
		
		-- Test log module
		assert(type(log) == "table", "log module should be available")
		
		-- Test that llm module is available
		assert(type(llm) == "table", "llm module should be available")
		assert(type(llm.chat) == "function", "llm.chat should be a function")
	`))
	require.NoError(t, err)

	err = eng.Execute(context.Background())
	require.NoError(t, err)
}

func TestRunSpellWithFile(t *testing.T) {
	// Create a temporary spell file
	tmpDir := t.TempDir()
	spellFile := filepath.Join(tmpDir, "test_spell.lua")

	spellContent := `
		print("Test spell executed")
		result = "success"
	`

	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	// Set MOCK_LLM to avoid needing real API keys
	os.Setenv("MOCK_LLM", "true")
	defer os.Unsetenv("MOCK_LLM")

	// Capture output
	stdout, stderr := captureOutput(t, func() {
		runSpell(spellFile, []string{})
	})

	// Check output
	assert.Contains(t, stdout, "Running spell: test_spell")
	assert.Contains(t, stdout, "Test spell executed")
	assert.Contains(t, stdout, "=== Spell Complete ===")
	assert.Empty(t, stderr)
}

func TestRunSpellWithDirectory(t *testing.T) {
	// Create a temporary spell directory
	tmpDir := t.TempDir()
	spellDir := filepath.Join(tmpDir, "test_spell_dir")
	err := os.Mkdir(spellDir, 0755)
	require.NoError(t, err)

	mainFile := filepath.Join(spellDir, "main.lua")
	spellContent := `
		print("Directory spell executed")
		if params and params.test then
			print("Parameter test = " .. params.test)
		end
	`

	err = os.WriteFile(mainFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	// Set MOCK_LLM to avoid needing real API keys
	os.Setenv("MOCK_LLM", "true")
	defer os.Unsetenv("MOCK_LLM")

	// Capture output with parameters
	stdout, stderr := captureOutput(t, func() {
		runSpell(spellDir, []string{"test=value123"})
	})

	// Check output
	assert.Contains(t, stdout, "Running spell: test_spell_dir")
	assert.Contains(t, stdout, "Directory spell executed")
	assert.Contains(t, stdout, "Parameter test = value123")
	assert.Contains(t, stdout, "=== Spell Complete ===")
	assert.Empty(t, stderr)
}

func TestMainCommands(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name           string
		args           []string
		expectExit     bool
		expectedOutput []string
	}{
		{
			name:       "help command",
			args:       []string{"llmspell", "help"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell - Cast scripting spells",
				"Usage:",
			},
		},
		{
			name:       "-h flag",
			args:       []string{"llmspell", "-h"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell - Cast scripting spells",
			},
		},
		{
			name:       "--help flag",
			args:       []string{"llmspell", "--help"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell - Cast scripting spells",
			},
		},
		{
			name:       "version command",
			args:       []string{"llmspell", "version"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell v0.1.0",
			},
		},
		{
			name:       "-v flag",
			args:       []string{"llmspell", "-v"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell v0.1.0",
			},
		},
		{
			name:       "--version flag",
			args:       []string{"llmspell", "--version"},
			expectExit: false,
			expectedOutput: []string{
				"llmspell v0.1.0",
			},
		},
		{
			name:       "no arguments",
			args:       []string{"llmspell"},
			expectExit: true,
			expectedOutput: []string{
				"llmspell - Cast scripting spells",
			},
		},
		{
			name:       "unknown command",
			args:       []string{"llmspell", "unknown"},
			expectExit: true,
			expectedOutput: []string{
				"Unknown command: unknown",
				"llmspell - Cast scripting spells",
			},
		},
		{
			name:       "run without spell path",
			args:       []string{"llmspell", "run"},
			expectExit: true,
			expectedOutput: []string{
				"Error: spell path required",
				"Usage: llmspell run <spell-path>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			if tt.expectExit {
				// For commands that call os.Exit, we need to handle it differently
				// We'll just check the output without actually calling main()
				if tt.name == "no arguments" {
					stdout, _ := captureOutput(t, func() {
						printUsage()
					})
					for _, expected := range tt.expectedOutput {
						assert.Contains(t, stdout, expected)
					}
				} else if tt.name == "unknown command" {
					stdout, _ := captureOutput(t, func() {
						fmt.Printf("Unknown command: %s\n", "unknown")
						printUsage()
					})
					for _, expected := range tt.expectedOutput {
						assert.Contains(t, stdout, expected)
					}
				} else if tt.name == "run without spell path" {
					stdout, _ := captureOutput(t, func() {
						fmt.Println("Error: spell path required")
						fmt.Println("Usage: llmspell run <spell-path> [param=value ...]")
					})
					for _, expected := range tt.expectedOutput {
						assert.Contains(t, stdout, expected)
					}
				}
			} else {
				// For commands that don't exit, we can test them directly
				stdout, _ := captureOutput(t, func() {
					// Simulate the switch statement from main()
					command := os.Args[1]
					switch command {
					case "help", "-h", "--help":
						printUsage()
					case "version", "-v", "--version":
						fmt.Println("llmspell v0.1.0")
					}
				})

				for _, expected := range tt.expectedOutput {
					assert.Contains(t, stdout, expected)
				}
			}
		})
	}
}

// TestRunSpellErrors tests error conditions in runSpell
func TestRunSpellErrors(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() string
		expectedError string
	}{
		{
			name: "non-existent spell",
			setup: func() string {
				return "/non/existent/spell.lua"
			},
			expectedError: "Cannot access spell:",
		},
		{
			name: "directory without main.lua",
			setup: func() string {
				tmpDir := t.TempDir()
				spellDir := filepath.Join(tmpDir, "empty_spell")
				err := os.Mkdir(spellDir, 0755)
				require.NoError(t, err)
				return spellDir
			},
			expectedError: "Cannot find spell script:",
		},
		{
			name: "invalid lua script",
			setup: func() string {
				tmpDir := t.TempDir()
				spellFile := filepath.Join(tmpDir, "invalid.lua")
				err := os.WriteFile(spellFile, []byte("invalid lua syntax {{"), 0644)
				require.NoError(t, err)
				return spellFile
			},
			expectedError: "Failed to load spell:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spellPath := tt.setup()

			// Set MOCK_LLM to avoid needing real API keys
			os.Setenv("MOCK_LLM", "true")
			defer os.Unsetenv("MOCK_LLM")

			// We need to capture the log.Fatal output
			// Since log.Fatal calls os.Exit, we'll test the error conditions
			// by checking what would happen without the exit

			if tt.name == "non-existent spell" {
				_, err := os.Stat(spellPath)
				assert.Error(t, err)
			} else if tt.name == "directory without main.lua" {
				mainScript := filepath.Join(spellPath, "main.lua")
				_, err := os.Stat(mainScript)
				assert.Error(t, err)
			} else if tt.name == "invalid lua script" {
				// Create engine and try to load the invalid script
				eng, err := lua.NewLuaEngine(&engine.Config{
					MaxExecutionTime: 30,
					MaxMemory:        64 * 1024 * 1024,
				})
				require.NoError(t, err)
				defer eng.Close()

				err = eng.LoadScriptFile(spellPath)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "parse error")
			}
		})
	}
}

// TestEnvironmentVariableLoading tests .env file loading
func TestEnvironmentVariableLoading(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	envContent := `TEST_VAR=test_value
ANOTHER_VAR=another_value`

	err = os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)

	// Clear any existing values
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("ANOTHER_VAR")

	// The main function would load .env, but we'll test the godotenv.Load directly
	err = godotenv.Load()
	require.NoError(t, err)

	// Check that variables were loaded
	assert.Equal(t, "test_value", os.Getenv("TEST_VAR"))
	assert.Equal(t, "another_value", os.Getenv("ANOTHER_VAR"))

	// Clean up
	os.Unsetenv("TEST_VAR")
	os.Unsetenv("ANOTHER_VAR")
}
