// ABOUTME: Validator interface wrapper that provides unified validation across different script engines.
// ABOUTME: Integrates security profiles and provides a common validation API for the runner.

package validator

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Validator is the interface that all validators must implement
type Validator interface {
	ValidateScript(script string, filename string) (*ValidationResult, error)
	ValidateFile(filename string) (*ValidationResult, error)
	GetConfig() *ValidationConfig
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Valid         bool                `json:"valid"`
	Errors        []ValidationError   `json:"errors,omitempty"`
	Warnings      []ValidationWarning `json:"warnings,omitempty"`
	Metrics       ValidationMetrics   `json:"metrics,omitempty"`
	Duration      time.Duration       `json:"duration,omitempty"`
	ValidatorName string              `json:"validator_name,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity,omitempty"`
	Code     string `json:"code,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationMetrics contains script metrics
type ValidationMetrics struct {
	Lines                int `json:"lines"`
	Functions            int `json:"functions"`
	MaxDepth             int `json:"max_depth"`
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
}

// ValidationConfig configures validation behavior
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

// DefaultValidationConfig returns default validation configuration
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

// IsValid returns true if validation passed
func (r *ValidationResult) IsValid() bool {
	return r.Valid
}

// HasErrors returns true if there are errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// BaseValidator provides common validation functionality
type BaseValidator struct {
	config *ValidationConfig
}

// NewBaseValidator creates a new base validator
func NewBaseValidator(config *ValidationConfig) *BaseValidator {
	return &BaseValidator{
		config: config,
	}
}

// ValidateScript validates a script (base implementation)
func (v *BaseValidator) ValidateScript(script string, filename string) (*ValidationResult, error) {
	// Base implementation - always passes
	return &ValidationResult{
		Valid:         true,
		ValidatorName: "base",
	}, nil
}

// ValidateFile validates a file (base implementation)
func (v *BaseValidator) ValidateFile(filename string) (*ValidationResult, error) {
	// Base implementation - not supported
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the validator configuration
func (v *BaseValidator) GetConfig() *ValidationConfig {
	return v.config
}

// ValidationChain chains multiple validators together
type ValidationChain struct {
	validators   []Validator
	shortCircuit bool
	mu           sync.RWMutex
}

// NewValidationChain creates a new validation chain
func NewValidationChain() *ValidationChain {
	return &ValidationChain{
		validators: make([]Validator, 0),
	}
}

// AddValidator adds a validator to the chain
func (c *ValidationChain) AddValidator(validator Validator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.validators = append(c.validators, validator)
}

// SetShortCircuit sets whether to stop on first error
func (c *ValidationChain) SetShortCircuit(shortCircuit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shortCircuit = shortCircuit
}

// Validate runs all validators in the chain
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

// ValidatorRegistry manages registered validators
type ValidatorRegistry struct {
	validators map[string]Validator
	mu         sync.RWMutex
}

// NewValidatorRegistry creates a new validator registry
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[string]Validator),
	}
}

// Register registers a validator
func (r *ValidatorRegistry) Register(name string, validator Validator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; exists {
		return fmt.Errorf("validator %s already registered", name)
	}

	r.validators[name] = validator
	return nil
}

// Get retrieves a validator
func (r *ValidatorRegistry) Get(name string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	if !exists {
		return nil, fmt.Errorf("validator %s not found", name)
	}

	return validator, nil
}

// Unregister removes a validator
func (r *ValidatorRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; !exists {
		return fmt.Errorf("validator %s not found", name)
	}

	delete(r.validators, name)
	return nil
}

// List returns all registered validator names
func (r *ValidatorRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.validators))
	for name := range r.validators {
		names = append(names, name)
	}
	return names
}

// ValidationContext provides context for validation
type ValidationContext struct {
	Filename  string
	Options   map[string]interface{}
	Metadata  map[string]string
	StartTime time.Time
}

// NewValidationContext creates a new validation context
func NewValidationContext(filename string, options map[string]interface{}) *ValidationContext {
	return &ValidationContext{
		Filename:  filename,
		Options:   options,
		Metadata:  make(map[string]string),
		StartTime: time.Now(),
	}
}

// SecurityValidator validates security constraints
type SecurityValidator struct {
	config   *ValidationConfig
	patterns map[string]*regexp.Regexp
}

// NewSecurityValidator creates a new security validator
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

// ValidateScript validates security constraints
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

// ValidateFile validates a file
func (v *SecurityValidator) ValidateFile(filename string) (*ValidationResult, error) {
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the configuration
func (v *SecurityValidator) GetConfig() *ValidationConfig {
	return v.config
}

// StyleValidator validates code style
type StyleValidator struct {
	config *ValidationConfig
}

// NewStyleValidator creates a new style validator
func NewStyleValidator(config *ValidationConfig) *StyleValidator {
	return &StyleValidator{
		config: config,
	}
}

// ValidateScript validates code style
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

// ValidateFile validates a file
func (v *StyleValidator) ValidateFile(filename string) (*ValidationResult, error) {
	return nil, fmt.Errorf("file validation not implemented")
}

// GetConfig returns the configuration
func (v *StyleValidator) GetConfig() *ValidationConfig {
	return v.config
}
