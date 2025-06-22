// ABOUTME: Validator interface wrapper that provides unified validation across different script engines.
// ABOUTME: Integrates security profiles and provides a common validation API for the runner.

// Package validator provides a unified validation interface for script engines.
// It includes syntax checking, security validation, style checking, and type checking
// capabilities with configurable profiles and chain-based validation support.
package validator

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Validator is the interface that all validators must implement.
// It provides methods to validate scripts and files, returning detailed
// validation results including errors, warnings, and metrics.
type Validator interface {
	// ValidateScript validates a script string with an optional filename.
	// It returns validation results including errors, warnings, and metrics.
	ValidateScript(script string, filename string) (*ValidationResult, error)
	
	// ValidateFile validates a script file by reading it from disk.
	// It returns validation results for the file contents.
	ValidateFile(filename string) (*ValidationResult, error)
	
	// GetConfig returns the validator's configuration.
	// This includes feature toggles, limits, and security settings.
	GetConfig() *ValidationConfig
}

// ValidationResult contains the results of validation.
// It includes validation status, errors, warnings, metrics, and timing information.
type ValidationResult struct {
	Valid         bool                `json:"valid"`
	Errors        []ValidationError   `json:"errors,omitempty"`
	Warnings      []ValidationWarning `json:"warnings,omitempty"`
	Metrics       ValidationMetrics   `json:"metrics,omitempty"`
	Duration      time.Duration       `json:"duration,omitempty"`
	ValidatorName string              `json:"validator_name,omitempty"`
}

// ValidationError represents a validation error.
// It contains error type, message, location information, and optional severity/code.
type ValidationError struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity,omitempty"`
	Code     string `json:"code,omitempty"`
}

// ValidationWarning represents a validation warning.
// It contains warning type, message, location, and optional improvement suggestions.
type ValidationWarning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationMetrics contains script metrics.
// It tracks code complexity indicators like lines, functions, nesting depth, and cyclomatic complexity.
type ValidationMetrics struct {
	Lines                int `json:"lines"`
	Functions            int `json:"functions"`
	MaxDepth             int `json:"max_depth"`
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
}

// ValidationConfig configures validation behavior.
// It provides fine-grained control over validation features, limits, security policies, and performance constraints.
type ValidationConfig struct {
	// Feature toggles
	EnableSyntaxCheck   bool `json:"enable_syntax_check"`
	EnableSecurityCheck bool `json:"enable_security_check"`
	EnableStyleCheck    bool `json:"enable_style_check"`
	EnableTypeCheck     bool `json:"enable_type_check"`

	// Limits
	MaxErrors     int `json:"max_errors"`
	MaxWarnings   int `json:"max_warnings"`
	MaxLineLength int `json:"max_line_length"`

	// Security
	SecurityProfile    string   `json:"security_profile"`
	ForbiddenPatterns  []string `json:"forbidden_patterns"`
	AllowedModules     []string `json:"allowed_modules"`
	ForbiddenFunctions []string `json:"forbidden_functions"`

	// Performance
	MaxLoopDepth     int `json:"max_loop_depth"`
	MaxFunctionDepth int `json:"max_function_depth"`
}

// DefaultValidationConfig returns default validation configuration.
// It provides a secure baseline with syntax and security checks enabled,
// reasonable limits, and sandbox security profile.
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		EnableSyntaxCheck:   true,
		EnableSecurityCheck: true,
		EnableStyleCheck:    true,
		EnableTypeCheck:     false,
		MaxErrors:           10,
		MaxWarnings:         20,
		MaxLineLength:       120,
		SecurityProfile:     "sandbox",
		MaxLoopDepth:        10,
		MaxFunctionDepth:    20,
		ForbiddenPatterns: []string{
			`os\.execute`,
			`io\.popen`,
			`loadstring`,
			`dofile`,
		},
		AllowedModules: []string{
			"string", "table", "math", "coroutine",
		},
	}
}

// IsValid returns true if validation passed.
// A result is valid when no errors were found during validation.
func (r *ValidationResult) IsValid() bool {
	return r.Valid
}

// HasErrors returns true if there are errors.
// This is a convenience method to check if any validation errors exist.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are warnings.
// This is a convenience method to check if any validation warnings exist.
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// BaseValidator provides common validation functionality.
// It serves as a foundation for implementing specific validators and
// provides a pass-through implementation for basic use cases.
type BaseValidator struct {
	config *ValidationConfig
}

// NewBaseValidator creates a new base validator.
// It initializes with the provided configuration for validation behavior.
func NewBaseValidator(config *ValidationConfig) *BaseValidator {
	return &BaseValidator{
		config: config,
	}
}

// ValidateScript validates a script (base implementation).
// The base implementation always passes validation, serving as a template
// for more specific validator implementations.
func (v *BaseValidator) ValidateScript(script string, filename string) (*ValidationResult, error) {
	// Base implementation - always passes
	return &ValidationResult{
		Valid:         true,
		ValidatorName: "base",
	}, nil
}

// ValidateFile validates a file (base implementation).
// The base implementation does not support file validation and returns an error.
// Subclasses should override this method to implement file validation.
func (v *BaseValidator) ValidateFile(filename string) (*ValidationResult, error) {
	// Base implementation - not supported
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the validator configuration.
// It provides access to the validator's configuration settings.
func (v *BaseValidator) GetConfig() *ValidationConfig {
	return v.config
}

// ValidationChain chains multiple validators together.
// It allows running multiple validators in sequence, combining their results,
// and optionally short-circuiting on the first error.
type ValidationChain struct {
	validators   []Validator
	shortCircuit bool
	mu           sync.RWMutex
}

// NewValidationChain creates a new validation chain.
// It initializes an empty chain ready to accept validators.
func NewValidationChain() *ValidationChain {
	return &ValidationChain{
		validators: make([]Validator, 0),
	}
}

// AddValidator adds a validator to the chain.
// Validators are executed in the order they were added.
func (c *ValidationChain) AddValidator(validator Validator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.validators = append(c.validators, validator)
}

// SetShortCircuit sets whether to stop on first error.
// When enabled, validation stops at the first validator that returns an error.
func (c *ValidationChain) SetShortCircuit(shortCircuit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shortCircuit = shortCircuit
}

// Validate runs all validators in the chain.
// It merges results from all validators unless short-circuit is enabled
// and a validator fails. Returns combined validation results.
func (c *ValidationChain) Validate(script string, filename string) (*ValidationResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	for _, validator := range c.validators {
		vResult, err := validator.ValidateScript(script, filename)
		if err != nil {
			return nil, err
		}

		// Merge results
		result.Errors = append(result.Errors, vResult.Errors...)
		result.Warnings = append(result.Warnings, vResult.Warnings...)

		if !vResult.Valid {
			result.Valid = false
			if c.shortCircuit {
				break
			}
		}
	}

	return result, nil
}

// ValidatorRegistry manages registered validators.
// It provides a thread-safe registry for storing and retrieving validators
// by name, supporting dynamic validator management.
type ValidatorRegistry struct {
	validators map[string]Validator
	mu         sync.RWMutex
}

// NewValidatorRegistry creates a new validator registry.
// It initializes an empty registry for managing validators.
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[string]Validator),
	}
}

// Register registers a validator.
// It adds a validator to the registry with the given name.
// Returns an error if a validator with the same name already exists.
func (r *ValidatorRegistry) Register(name string, validator Validator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; exists {
		return fmt.Errorf("validator %s already registered", name)
	}

	r.validators[name] = validator
	return nil
}

// Get retrieves a validator.
// It returns the validator registered with the given name.
// Returns an error if no validator is found with that name.
func (r *ValidatorRegistry) Get(name string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	if !exists {
		return nil, fmt.Errorf("validator %s not found", name)
	}

	return validator, nil
}

// Unregister removes a validator.
// It removes the validator with the given name from the registry.
// Returns an error if no validator is found with that name.
func (r *ValidatorRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; !exists {
		return fmt.Errorf("validator %s not found", name)
	}

	delete(r.validators, name)
	return nil
}

// List returns all registered validator names.
// It returns a slice containing the names of all registered validators.
func (r *ValidatorRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.validators))
	for name := range r.validators {
		names = append(names, name)
	}
	return names
}

// ValidationContext provides context for validation.
// It carries metadata about the validation session including filename,
// options, custom metadata, and timing information.
type ValidationContext struct {
	Filename  string
	Options   map[string]interface{}
	Metadata  map[string]string
	StartTime time.Time
}

// NewValidationContext creates a new validation context.
// It initializes context with filename, options, and current timestamp
// for tracking validation session metadata.
func NewValidationContext(filename string, options map[string]interface{}) *ValidationContext {
	return &ValidationContext{
		Filename:  filename,
		Options:   options,
		Metadata:  make(map[string]string),
		StartTime: time.Now(),
	}
}

// SecurityValidator validates security constraints.
// It checks for forbidden patterns, unauthorized module usage,
// and enforces security policies defined in the configuration.
type SecurityValidator struct {
	config   *ValidationConfig
	patterns map[string]*regexp.Regexp
}

// NewSecurityValidator creates a new security validator.
// It compiles forbidden patterns from the configuration for efficient
// pattern matching during validation.
func NewSecurityValidator(config *ValidationConfig) *SecurityValidator {
	v := &SecurityValidator{
		config:   config,
		patterns: make(map[string]*regexp.Regexp),
	}

	// Compile forbidden patterns
	for _, pattern := range config.ForbiddenPatterns {
		v.patterns[pattern] = regexp.MustCompile(pattern)
	}

	return v
}

// ValidateScript validates security constraints.
// It checks the script for forbidden patterns and unauthorized module usage,
// returning errors for any security violations found.
func (v *SecurityValidator) ValidateScript(script string, filename string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:         true,
		Errors:        make([]ValidationError, 0),
		ValidatorName: "security",
	}

	// Check forbidden patterns
	for pattern, regex := range v.patterns {
		if matches := regex.FindAllStringIndex(script, -1); len(matches) > 0 {
			lines := strings.Split(script[:matches[0][0]], "\n")
			result.Errors = append(result.Errors, ValidationError{
				Type:    "security",
				Message: fmt.Sprintf("forbidden pattern detected: %s", pattern),
				Line:    len(lines),
				Column:  matches[0][0] - strings.LastIndex(script[:matches[0][0]], "\n") - 1,
			})
			result.Valid = false
		}
	}

	// Check module usage - look for module.function patterns
	modulePattern := regexp.MustCompile(`\b(\w+)\.[\w_]+`)
	matches := modulePattern.FindAllStringSubmatch(script, -1)

	knownLocals := map[string]bool{
		"local":    true,
		"self":     true,
		"result":   true,
		"response": true,
		"f":        true,
		// Common variable names to ignore
	}

	for _, match := range matches {
		module := match[1]

		// Skip if it's a known local variable
		if knownLocals[module] {
			continue
		}

		// Check if it's a number (like table[1].something)
		if _, err := fmt.Sscanf(module, "%d", new(int)); err == nil {
			continue
		}

		allowed := false
		for _, allowedModule := range v.config.AllowedModules {
			if module == allowedModule {
				allowed = true
				break
			}
		}

		if !allowed {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "security",
				Message: fmt.Sprintf("forbidden module usage: %s", module),
				Line:    1, // Would need more complex tracking for exact line
			})
			result.Valid = false
			break // Only report first forbidden module
		}
	}

	return result, nil
}

// ValidateFile validates a file.
// The security validator does not implement file validation and returns an error.
// Use ValidateScript after reading the file contents instead.
func (v *SecurityValidator) ValidateFile(filename string) (*ValidationResult, error) {
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the configuration.
// It provides access to the security validator's configuration settings.
func (v *SecurityValidator) GetConfig() *ValidationConfig {
	return v.config
}

// StyleValidator validates code style.
// It checks for style issues like line length violations, trailing whitespace,
// and other formatting concerns defined in the configuration.
type StyleValidator struct {
	config *ValidationConfig
}

// NewStyleValidator creates a new style validator.
// It initializes with the provided configuration for style checking rules.
func NewStyleValidator(config *ValidationConfig) *StyleValidator {
	return &StyleValidator{
		config: config,
	}
}

// ValidateScript validates code style.
// It analyzes the script for style violations including line length
// and trailing whitespace, returning warnings for any issues found.
func (v *StyleValidator) ValidateScript(script string, filename string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:         true,
		Warnings:      make([]ValidationWarning, 0),
		ValidatorName: "style",
	}

	lines := strings.Split(script, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// Check line length
		if len(line) > v.config.MaxLineLength {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "style",
				Message: fmt.Sprintf("line too long (%d > %d)", len(line), v.config.MaxLineLength),
				Line:    lineNum,
			})
		}

		// Check trailing whitespace
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:    "style",
				Message: "trailing whitespace",
				Line:    lineNum,
			})
		}
	}

	return result, nil
}

// ValidateFile validates a file.
// The style validator does not implement file validation and returns an error.
// Use ValidateScript after reading the file contents instead.
func (v *StyleValidator) ValidateFile(filename string) (*ValidationResult, error) {
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the configuration.
// It provides access to the style validator's configuration settings.
func (v *StyleValidator) GetConfig() *ValidationConfig {
	return v.config
}
