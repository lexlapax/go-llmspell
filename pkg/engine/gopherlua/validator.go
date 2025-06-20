// ABOUTME: This file implements script validation for Lua scripts including syntax, security, and performance checks.
// ABOUTME: It provides comprehensive validation capabilities for use by the spell runner and development tools.

package gopherlua

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/gopher-lua/parse"
)

// ValidatorConfig configures the script validator behavior
type ValidatorConfig struct {
	// Syntax validation
	EnableSyntaxCheck bool `json:"enable_syntax_check"`
	EnableLinting     bool `json:"enable_linting"`

	// Security validation
	EnableSecurityCheck bool     `json:"enable_security_check"`
	ForbiddenPatterns   []string `json:"forbidden_patterns"`
	AllowedGlobals      []string `json:"allowed_globals"`

	// Performance validation
	EnablePerformanceCheck bool `json:"enable_performance_check"`
	MaxLoopDepth           int  `json:"max_loop_depth"`
	MaxFunctionDepth       int  `json:"max_function_depth"`

	// Type checking
	EnableTypeHints        bool `json:"enable_type_hints"`
	RequireTypeAnnotations bool `json:"require_type_annotations"`
}

// DefaultValidatorConfig returns a default validator configuration
func DefaultValidatorConfig() ValidatorConfig {
	return ValidatorConfig{
		EnableSyntaxCheck:      true,
		EnableLinting:          true,
		EnableSecurityCheck:    true,
		EnablePerformanceCheck: true,
		MaxLoopDepth:           10,
		MaxFunctionDepth:       20,
		EnableTypeHints:        false,
		ForbiddenPatterns: []string{
			`os\.execute`,
			`io\.popen`,
			`loadstring`,
			`dofile`,
			`debug\.`,
		},
		AllowedGlobals: []string{
			"print", "pairs", "ipairs", "next", "type",
			"tostring", "tonumber", "string", "table", "math",
			"coroutine", "require", "module", "package",
			// go-llmspell specific globals
			"llm", "agent", "tools", "promise", "state",
			"events", "hooks", "data", "auth", "log",
		},
	}
}

// ValidationResult contains the results of script validation
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
	Metrics  ValidationMetrics   `json:"metrics"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Code    string `json:"code,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationMetrics contains script complexity metrics
type ValidationMetrics struct {
	Lines                int `json:"lines"`
	Functions            int `json:"functions"`
	MaxDepth             int `json:"max_depth"`
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
}

// ScriptValidator validates Lua scripts
type ScriptValidator struct {
	config   ValidatorConfig
	patterns map[string]*regexp.Regexp
}

// NewScriptValidator creates a new script validator
func NewScriptValidator(config ValidatorConfig) *ScriptValidator {
	v := &ScriptValidator{
		config:   config,
		patterns: make(map[string]*regexp.Regexp),
	}

	// Compile forbidden patterns
	for _, pattern := range config.ForbiddenPatterns {
		v.patterns[pattern] = regexp.MustCompile(pattern)
	}

	return v
}

// ValidateScript validates a Lua script
func (v *ScriptValidator) ValidateScript(script string, filename string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Metrics:  ValidationMetrics{},
	}

	// Syntax check
	if v.config.EnableSyntaxCheck {
		if err := v.checkSyntax(script, filename, result); err != nil {
			return nil, fmt.Errorf("syntax check failed: %w", err)
		}
	}

	// Linting
	if v.config.EnableLinting && result.Valid {
		v.performLinting(script, result)
	}

	// Security check
	if v.config.EnableSecurityCheck && result.Valid {
		v.checkSecurity(script, result)
	}

	// Performance check
	if v.config.EnablePerformanceCheck && result.Valid {
		v.checkPerformance(script, result)
	}

	// Calculate metrics
	v.calculateMetrics(script, result)

	// Set final validity
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// checkSyntax performs syntax validation
func (v *ScriptValidator) checkSyntax(script string, filename string, result *ValidationResult) error {
	// Use gopher-lua's parser to check syntax
	_, err := parse.Parse(strings.NewReader(script), filename)
	if err != nil {
		// Extract error details
		if parseErr, ok := err.(*parse.Error); ok {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "syntax",
				Message: parseErr.Error(),
				Line:    parseErr.Pos.Line,
				Column:  parseErr.Pos.Column,
			})
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "syntax",
				Message: err.Error(),
				Line:    0,
				Column:  0,
			})
		}
		result.Valid = false
	}

	return nil
}

// performLinting checks for code style and best practices
func (v *ScriptValidator) performLinting(script string, result *ValidationResult) {
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check for trailing whitespace
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "style",
				Message:    "Trailing whitespace",
				Line:       lineNum,
				Column:     len(line),
				Suggestion: "Remove trailing whitespace",
			})
		}

		// Check for TODO/FIXME comments
		if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "todo",
				Message: "TODO/FIXME comment found",
				Line:    lineNum,
				Column:  strings.Index(line, "TODO"),
			})
		}

		// Check for very long lines
		if len(line) > 120 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "style",
				Message:    fmt.Sprintf("Line too long (%d characters)", len(line)),
				Line:       lineNum,
				Column:     120,
				Suggestion: "Break line into multiple lines",
			})
		}

		// Check for inconsistent indentation
		if strings.HasPrefix(line, "\t") && strings.Contains(line, "    ") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "style",
				Message:    "Mixed tabs and spaces for indentation",
				Line:       lineNum,
				Column:     0,
				Suggestion: "Use consistent indentation (tabs or spaces)",
			})
		}
	}

	// Check for missing module documentation
	if !strings.Contains(script, "-- @module") && !strings.Contains(script, "---") {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:       "documentation",
			Message:    "Missing module documentation",
			Line:       1,
			Column:     0,
			Suggestion: "Add module documentation at the beginning of the file",
		})
	}
}

// checkSecurity validates security constraints
func (v *ScriptValidator) checkSecurity(script string, result *ValidationResult) {
	// Check for forbidden patterns
	for pattern, regex := range v.patterns {
		if matches := regex.FindAllStringIndex(script, -1); len(matches) > 0 {
			// Find line number for first match
			lines := strings.Split(script[:matches[0][0]], "\n")
			lineNum := len(lines)

			result.Errors = append(result.Errors, ValidationError{
				Type:    "security",
				Message: fmt.Sprintf("Forbidden pattern '%s' detected", pattern),
				Line:    lineNum,
				Column:  matches[0][0] - strings.LastIndex(script[:matches[0][0]], "\n") - 1,
				Code:    pattern,
			})
		}
	}

	// Check for global variable assignments
	globalAssignPattern := regexp.MustCompile(`^[^-\n]*\b([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	lines := strings.Split(script, "\n")

	for i, line := range lines {
		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		if matches := globalAssignPattern.FindStringSubmatch(line); len(matches) > 0 {
			globalName := matches[1]

			// Check if it's a local declaration
			if !strings.Contains(line, "local "+globalName) {
				// Check if it's an allowed global
				allowed := false
				for _, allowedGlobal := range v.config.AllowedGlobals {
					if globalName == allowedGlobal {
						allowed = true
						break
					}
				}

				if !allowed && !strings.Contains(line, "."+globalName) {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Type:       "security",
						Message:    fmt.Sprintf("Global variable assignment '%s' detected", globalName),
						Line:       i + 1,
						Column:     strings.Index(line, globalName),
						Suggestion: fmt.Sprintf("Use 'local %s' or add to allowed globals", globalName),
					})
				}
			}
		}
	}
}

// checkPerformance validates performance concerns
func (v *ScriptValidator) checkPerformance(script string, result *ValidationResult) {
	// Check for deeply nested loops
	loopPattern := regexp.MustCompile(`\b(for|while|repeat)\b`)
	lines := strings.Split(script, "\n")

	loopDepth := 0
	maxLoopDepth := 0

	for i, line := range lines {
		// Track loop depth
		if loopPattern.MatchString(line) {
			loopDepth++
			if loopDepth > maxLoopDepth {
				maxLoopDepth = loopDepth
			}

			if loopDepth > v.config.MaxLoopDepth {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:       "performance",
					Message:    fmt.Sprintf("Deeply nested loop (depth: %d)", loopDepth),
					Line:       i + 1,
					Column:     0,
					Suggestion: "Consider refactoring to reduce nesting",
				})
			}
		}

		// Check for end statements
		if strings.Contains(line, "end") || strings.Contains(line, "until") {
			if loopDepth > 0 {
				loopDepth--
			}
		}
	}

	// Check for string concatenation in loops - look for .. inside loop blocks
	inLoop := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if entering a loop
		if loopPattern.MatchString(trimmed) {
			inLoop = true
		}

		// Check for string concatenation inside loop
		if inLoop && strings.Contains(line, "..") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "performance",
				Message:    "String concatenation in loop detected",
				Line:       i + 1,
				Column:     strings.Index(line, ".."),
				Suggestion: "Use table.concat() for better performance",
			})
			// Only warn once per loop
			inLoop = false
		}

		// Check if exiting loop
		if strings.HasPrefix(trimmed, "end") {
			inLoop = false
		}
	}

	// Check for repeated table lookups
	tableLookupPattern := regexp.MustCompile(`(\w+)\[["'](\w+)["']\]`)
	lookupCounts := make(map[string]int)

	for _, match := range tableLookupPattern.FindAllStringSubmatch(script, -1) {
		lookup := match[0]
		lookupCounts[lookup]++

		if lookupCounts[lookup] > 5 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "performance",
				Message:    fmt.Sprintf("Repeated table lookup '%s' (%d times)", lookup, lookupCounts[lookup]),
				Line:       0, // Would need more complex tracking for exact line
				Column:     0,
				Suggestion: "Cache the value in a local variable",
			})
			delete(lookupCounts, lookup) // Only warn once
		}
	}
}

// calculateMetrics calculates script complexity metrics
func (v *ScriptValidator) calculateMetrics(script string, result *ValidationResult) {
	lines := strings.Split(script, "\n")
	result.Metrics.Lines = len(lines)

	// Count functions
	functionPattern := regexp.MustCompile(`\bfunction\b`)
	result.Metrics.Functions = len(functionPattern.FindAllString(script, -1))

	// Calculate cyclomatic complexity (simplified)
	// Count decision points
	decisionPattern := regexp.MustCompile(`\b(if|elseif|for|while|repeat|and|or)\b`)
	result.Metrics.CyclomaticComplexity = len(decisionPattern.FindAllString(script, -1)) + 1

	// Max depth already calculated in performance check
	result.Metrics.MaxDepth = v.calculateMaxDepth(script)
}

// calculateMaxDepth calculates the maximum nesting depth
func (v *ScriptValidator) calculateMaxDepth(script string) int {
	lines := strings.Split(script, "\n")
	depth := 0
	maxDepth := 0

	depthKeywords := regexp.MustCompile(`\b(function|if|for|while|repeat|do)\b`)
	endKeywords := regexp.MustCompile(`\b(end|until)\b`)

	for _, line := range lines {
		if depthKeywords.MatchString(line) {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		if endKeywords.MatchString(line) {
			if depth > 0 {
				depth--
			}
		}
	}

	return maxDepth
}

// ValidateFile validates a Lua script file
func (v *ScriptValidator) ValidateFile(filename string) (*ValidationResult, error) {
	// For now, return an error as we don't have file system access in the validator
	// The caller should read the file and use ValidateScript
	return nil, fmt.Errorf("use ValidateScript with file contents instead")
}

// GetLintRules returns the active linting rules
func (v *ScriptValidator) GetLintRules() []string {
	rules := []string{}

	if v.config.EnableLinting {
		rules = append(rules,
			"no-trailing-whitespace",
			"no-mixed-indentation",
			"max-line-length",
			"require-module-doc",
		)
	}

	if v.config.EnableSecurityCheck {
		rules = append(rules,
			"no-forbidden-patterns",
			"no-unauthorized-globals",
		)
	}

	if v.config.EnablePerformanceCheck {
		rules = append(rules,
			"max-loop-depth",
			"no-string-concat-in-loop",
			"cache-repeated-lookups",
		)
	}

	return rules
}
