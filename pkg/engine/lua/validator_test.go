// ABOUTME: This file contains comprehensive tests for the Lua script validator.
// ABOUTME: It tests syntax validation, security checks, performance warnings, and linting rules.

package gopherlua

import (
	"fmt"
	"strings"
	"testing"
)

func TestScriptValidator_SyntaxCheck(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expectValid bool
		errorType   string
		errorMsg    string
	}{
		{
			name:        "valid script",
			script:      `local x = 10`,
			expectValid: true,
		},
		{
			name:        "syntax error - missing end",
			script:      `if true then`,
			expectValid: false,
			errorType:   "syntax",
			errorMsg:    "syntax error",
		},
		{
			name:        "syntax error - unexpected token",
			script:      `local x = = 10`,
			expectValid: false,
			errorType:   "syntax",
			errorMsg:    "syntax error",
		},
		{
			name: "valid function",
			script: `
				function add(a, b)
					return a + b
				end
			`,
			expectValid: true,
		},
		{
			name:        "unclosed string",
			script:      `local s = "hello`,
			expectValid: false,
			errorType:   "syntax",
		},
	}

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateScript(tt.script, tt.name+".lua")
			if err != nil {
				t.Fatalf("ValidateScript failed: %v", err)
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if !tt.expectValid && len(result.Errors) > 0 {
				if result.Errors[0].Type != tt.errorType {
					t.Errorf("Expected error type %s, got %s", tt.errorType, result.Errors[0].Type)
				}
				if tt.errorMsg != "" && !strings.Contains(result.Errors[0].Message, tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, result.Errors[0].Message)
				}
			}
		})
	}
}

func TestScriptValidator_SecurityCheck(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		expectErrors      int
		expectSecWarnings int // only count security warnings
		errorPattern      string
	}{
		{
			name:         "forbidden os.execute",
			script:       `os.execute("rm -rf /")`,
			expectErrors: 1,
			errorPattern: "os\\.execute",
		},
		{
			name:         "forbidden io.popen",
			script:       `local f = io.popen("ls")`,
			expectErrors: 1,
			errorPattern: "io\\.popen",
		},
		{
			name:         "forbidden loadstring",
			script:       `local fn = loadstring("return 42")`,
			expectErrors: 1,
			errorPattern: "loadstring",
		},
		{
			name:              "global variable assignment",
			script:            `myGlobal = 42`,
			expectSecWarnings: 1,
		},
		{
			name:   "local variable is fine",
			script: `local myLocal = 42`,
		},
		{
			name:   "allowed global",
			script: `print("hello")`,
		},
		{
			name:         "multiple security issues",
			script:       `os.execute("bad"); io.popen("also bad")`,
			expectErrors: 2,
		},
	}

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateScript(tt.script, tt.name+".lua")
			if err != nil {
				t.Fatalf("ValidateScript failed: %v", err)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
				for _, e := range result.Errors {
					t.Logf("Error: %s - %s", e.Type, e.Message)
				}
			}

			// Count only security warnings
			securityWarnings := 0
			for _, w := range result.Warnings {
				if w.Type == "security" {
					securityWarnings++
				}
			}

			if securityWarnings != tt.expectSecWarnings {
				t.Errorf("Expected %d security warnings, got %d", tt.expectSecWarnings, securityWarnings)
				for _, w := range result.Warnings {
					t.Logf("Warning: %s - %s", w.Type, w.Message)
				}
			}

			if tt.errorPattern != "" && len(result.Errors) > 0 {
				if result.Errors[0].Code != tt.errorPattern {
					t.Errorf("Expected error code %s, got %s", tt.errorPattern, result.Errors[0].Code)
				}
			}
		})
	}
}

func TestScriptValidator_Linting(t *testing.T) {
	tests := []struct {
		name             string
		script           string
		expectedWarnings []string
	}{
		{
			name:             "trailing whitespace",
			script:           "local x = 10  \nlocal y = 20\t",
			expectedWarnings: []string{"Trailing whitespace", "Trailing whitespace"},
		},
		{
			name:             "TODO comment",
			script:           "-- TODO: implement this\n-- FIXME: broken",
			expectedWarnings: []string{"TODO/FIXME comment found", "TODO/FIXME comment found"},
		},
		{
			name:             "long line",
			script:           "local veryLongVariableName = 'this is a very long string that exceeds the recommended line length of 120 characters and should trigger a warning'",
			expectedWarnings: []string{"Line too long"},
		},
		{
			name:             "mixed indentation",
			script:           "\tlocal x = 10    -- spaces after tab",
			expectedWarnings: []string{"Mixed tabs and spaces"},
		},
		{
			name:             "missing documentation",
			script:           "local x = 10",
			expectedWarnings: []string{"Missing module documentation"},
		},
		{
			name:             "well-documented module",
			script:           "--- This module does something\n-- @module mymodule\nlocal x = 10",
			expectedWarnings: []string{}, // No documentation warning
		},
	}

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateScript(tt.script, tt.name+".lua")
			if err != nil {
				t.Fatalf("ValidateScript failed: %v", err)
			}

			foundWarnings := make(map[string]bool)
			for _, w := range result.Warnings {
				foundWarnings[w.Message] = true
			}

			for _, expected := range tt.expectedWarnings {
				found := false
				for _, w := range result.Warnings {
					if strings.Contains(w.Message, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected warning containing %q not found", expected)
					t.Logf("Got warnings: %v", result.Warnings)
				}
			}
		})
	}
}

func TestScriptValidator_PerformanceCheck(t *testing.T) {
	tests := []struct {
		name          string
		script        string
		expectWarning bool
		warningType   string
	}{
		{
			name: "deeply nested loops",
			script: `
				for i = 1, 10 do
					for j = 1, 10 do
						for k = 1, 10 do
							for l = 1, 10 do
								for m = 1, 10 do
									for n = 1, 10 do
										for o = 1, 10 do
											for p = 1, 10 do
												for q = 1, 10 do
													for r = 1, 10 do
														for s = 1, 10 do
															print(s)
														end
													end
												end
											end
										end
									end
								end
							end
						end
					end
				end
			`,
			expectWarning: true,
			warningType:   "performance",
		},
		{
			name: "string concatenation in loop",
			script: `
				local result = ""
				for i = 1, 100 do
					result = result .. tostring(i)
				end
			`,
			expectWarning: true,
			warningType:   "performance",
		},
		{
			name: "repeated table lookup",
			script: `
				local t = {foo = "bar"}
				print(t["foo"])
				print(t["foo"])
				print(t["foo"])
				print(t["foo"])
				print(t["foo"])
				print(t["foo"])
			`,
			expectWarning: true,
			warningType:   "performance",
		},
		{
			name: "efficient table concat",
			script: `
				local parts = {}
				for i = 1, 100 do
					parts[#parts + 1] = tostring(i)
				end
				local result = table.concat(parts)
			`,
			expectWarning: false,
		},
	}

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateScript(tt.script, tt.name+".lua")
			if err != nil {
				t.Fatalf("ValidateScript failed: %v", err)
			}

			hasPerformanceWarning := false
			for _, w := range result.Warnings {
				if w.Type == tt.warningType {
					hasPerformanceWarning = true
					break
				}
			}

			if hasPerformanceWarning != tt.expectWarning {
				t.Errorf("Expected performance warning: %v, got %v", tt.expectWarning, hasPerformanceWarning)
				for _, w := range result.Warnings {
					t.Logf("Warning: %s - %s", w.Type, w.Message)
				}
			}
		})
	}
}

func TestScriptValidator_Metrics(t *testing.T) {
	script := `
		-- Test module
		local M = {}
		
		function M.add(a, b)
			if a > 0 and b > 0 then
				return a + b
			elseif a < 0 or b < 0 then
				return 0
			else
				return -1
			end
		end
		
		function M.multiply(a, b)
			local result = 0
			for i = 1, b do
				result = result + a
			end
			return result
		end
		
		return M
	`

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	result, err := validator.ValidateScript(script, "metrics_test.lua")
	if err != nil {
		t.Fatalf("ValidateScript failed: %v", err)
	}

	// Check metrics
	if result.Metrics.Functions != 2 {
		t.Errorf("Expected 2 functions, got %d", result.Metrics.Functions)
	}

	if result.Metrics.Lines < 20 {
		t.Errorf("Expected at least 20 lines, got %d", result.Metrics.Lines)
	}

	if result.Metrics.CyclomaticComplexity < 5 {
		t.Errorf("Expected complexity >= 5, got %d", result.Metrics.CyclomaticComplexity)
	}

	if result.Metrics.MaxDepth < 2 {
		t.Errorf("Expected max depth >= 2, got %d", result.Metrics.MaxDepth)
	}
}

func TestScriptValidator_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		config         ValidatorConfig
		script         string
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "syntax check disabled",
			config: ValidatorConfig{
				EnableSyntaxCheck: false,
				EnableLinting:     false, // Also disable linting to avoid doc warning
			},
			script:       `if true then`, // syntax error
			expectErrors: 0,              // Should not catch syntax error
		},
		{
			name: "security check disabled",
			config: ValidatorConfig{
				EnableSyntaxCheck:   true,
				EnableSecurityCheck: false,
			},
			script:       `os.execute("bad")`,
			expectErrors: 0, // Should not catch security issue
		},
		{
			name: "custom forbidden patterns",
			config: ValidatorConfig{
				EnableSyntaxCheck:   true,
				EnableSecurityCheck: true,
				ForbiddenPatterns:   []string{`require`},
			},
			script:       `local m = require("module")`,
			expectErrors: 1,
		},
		{
			name: "custom allowed globals",
			config: ValidatorConfig{
				EnableSyntaxCheck:   true,
				EnableSecurityCheck: true,
				AllowedGlobals:      []string{"myGlobal"},
			},
			script:         `myGlobal = 42`,
			expectWarnings: 0, // Should not warn about allowed global
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewScriptValidator(tt.config)
			result, err := validator.ValidateScript(tt.script, tt.name+".lua")
			if err != nil {
				t.Fatalf("ValidateScript failed: %v", err)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
			}

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
			}
		})
	}
}

func TestScriptValidator_ComplexScript(t *testing.T) {
	script := `
		--- LLM interaction module
		-- @module llm_helper
		
		local llm = require("llm")
		local promise = require("promise")
		
		local M = {}
		
		-- Configuration
		M.config = {
			timeout = 30,
			retries = 3,
			model = "gpt-4"
		}
		
		--- Send a prompt to the LLM
		-- @param prompt string The prompt to send
		-- @return promise A promise that resolves with the response
		function M.query(prompt)
			-- Validate input
			if type(prompt) ~= "string" or #prompt == 0 then
				return promise.reject("Invalid prompt")
			end
			
			-- Create options
			local options = {
				model = M.config.model,
				timeout = M.config.timeout
			}
			
			-- Send request with retry
			local attempts = 0
			local function tryRequest()
				attempts = attempts + 1
				
				return llm.complete(prompt, options):andThen(
					function(response)
						return response.content
					end,
					function(error)
						if attempts < M.config.retries then
							return tryRequest()
						else
							return promise.reject("Failed after " .. attempts .. " attempts: " .. error)
						end
					end
				)
			end
			
			return tryRequest()
		end
		
		--- Batch process multiple prompts
		-- @param prompts table Array of prompts
		-- @return promise A promise that resolves with all responses
		function M.batchQuery(prompts)
			local promises = {}
			
			for i, prompt in ipairs(prompts) do
				promises[i] = M.query(prompt)
			end
			
			return promise.all(promises)
		end
		
		return M
	`

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	result, err := validator.ValidateScript(script, "complex_test.lua")
	if err != nil {
		t.Fatalf("ValidateScript failed: %v", err)
	}

	// Should be valid
	if !result.Valid {
		t.Errorf("Expected script to be valid")
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Type, e.Message)
		}
	}

	// Check metrics
	if result.Metrics.Functions < 3 { // M.query, tryRequest, M.batchQuery
		t.Errorf("Expected at least 3 functions, got %d", result.Metrics.Functions)
	}

	// Should have some complexity
	if result.Metrics.CyclomaticComplexity < 5 {
		t.Errorf("Expected complexity >= 5, got %d", result.Metrics.CyclomaticComplexity)
	}

	// Check for specific warnings
	hasDocWarning := false
	for _, w := range result.Warnings {
		if w.Type == "documentation" {
			hasDocWarning = true
		}
	}

	if hasDocWarning {
		t.Error("Should not have documentation warning for well-documented module")
	}
}

func TestScriptValidator_GetLintRules(t *testing.T) {
	tests := []struct {
		name     string
		config   ValidatorConfig
		expected []string
	}{
		{
			name:   "all enabled",
			config: DefaultValidatorConfig(),
			expected: []string{
				"no-trailing-whitespace",
				"no-mixed-indentation",
				"max-line-length",
				"require-module-doc",
				"no-forbidden-patterns",
				"no-unauthorized-globals",
				"max-loop-depth",
				"no-string-concat-in-loop",
				"cache-repeated-lookups",
			},
		},
		{
			name: "only linting",
			config: ValidatorConfig{
				EnableLinting:          true,
				EnableSecurityCheck:    false,
				EnablePerformanceCheck: false,
			},
			expected: []string{
				"no-trailing-whitespace",
				"no-mixed-indentation",
				"max-line-length",
				"require-module-doc",
			},
		},
		{
			name: "only security",
			config: ValidatorConfig{
				EnableLinting:          false,
				EnableSecurityCheck:    true,
				EnablePerformanceCheck: false,
			},
			expected: []string{
				"no-forbidden-patterns",
				"no-unauthorized-globals",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewScriptValidator(tt.config)
			rules := validator.GetLintRules()

			if len(rules) != len(tt.expected) {
				t.Errorf("Expected %d rules, got %d", len(tt.expected), len(rules))
				t.Logf("Got rules: %v", rules)
			}

			// Check each expected rule is present
			for _, expected := range tt.expected {
				found := false
				for _, rule := range rules {
					if rule == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected rule %q not found", expected)
				}
			}
		})
	}
}

func BenchmarkScriptValidator_ValidateScript(b *testing.B) {
	script := `
		local function fibonacci(n)
			if n <= 1 then
				return n
			end
			return fibonacci(n-1) + fibonacci(n-2)
		end
		
		for i = 1, 10 do
			print(fibonacci(i))
		end
	`

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateScript(script, "bench.lua")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScriptValidator_LargeScript(b *testing.B) {
	// Generate a large script
	var sb strings.Builder
	sb.WriteString("-- Large test script\n")
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("local var%d = %d\n", i, i))
		sb.WriteString(fmt.Sprintf("function func%d(x)\n", i))
		sb.WriteString("  return x * 2\n")
		sb.WriteString("end\n")
	}
	script := sb.String()

	config := DefaultValidatorConfig()
	validator := NewScriptValidator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateScript(script, "large.lua")
		if err != nil {
			b.Fatal(err)
		}
	}
}
