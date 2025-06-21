// ABOUTME: This file implements spell loading functionality for parsing spell.yaml files.
// ABOUTME: It handles spell metadata validation and provides utilities for working with spell directories.

package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SpellMetadata represents the metadata from a spell.yaml file
type SpellMetadata struct {
	// Basic information
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`

	// Execution settings
	Engine     string `yaml:"engine"`
	EntryPoint string `yaml:"entry_point"`
	Timeout    string `yaml:"timeout,omitempty"`

	// Dependencies and parameters
	Dependencies []string         `yaml:"dependencies,omitempty"`
	Parameters   []SpellParameter `yaml:"parameters,omitempty"`

	// Security and metadata
	SecurityProfile string                 `yaml:"security_profile,omitempty"`
	Tags            []string               `yaml:"tags,omitempty"`
	Metadata        map[string]interface{} `yaml:"metadata,omitempty"`

	// Runtime information (not from YAML)
	RootDir string `yaml:"-"`
}

// SpellParameter defines a parameter that can be passed to a spell
type SpellParameter struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Validation  string      `yaml:"validation,omitempty"`
}

// Validate checks if the spell metadata is valid
func (m *SpellMetadata) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("spell name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("spell version is required")
	}
	if m.Engine == "" {
		return fmt.Errorf("spell engine is required")
	}
	if m.EntryPoint == "" {
		return fmt.Errorf("spell entry point is required")
	}

	// Validate parameters
	paramNames := make(map[string]bool)
	for _, param := range m.Parameters {
		if err := validateParameter(param); err != nil {
			return fmt.Errorf("invalid parameter %s: %w", param.Name, err)
		}
		if paramNames[param.Name] {
			return fmt.Errorf("duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true
	}

	return nil
}

// validateParameter validates a single parameter
func validateParameter(p SpellParameter) error {
	if p.Name == "" {
		return fmt.Errorf("parameter name is required")
	}
	if p.Type == "" {
		return fmt.Errorf("parameter type is required")
	}

	// Validate type
	validTypes := []string{"string", "number", "boolean", "array", "object"}
	valid := false
	for _, t := range validTypes {
		if p.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid parameter type: %s (must be one of: %s)", p.Type, strings.Join(validTypes, ", "))
	}

	return nil
}

// SpellLoader handles loading spell metadata from files
type SpellLoader struct {
	// Could add caching or other features in the future
}

// NewSpellLoader creates a new spell loader
func NewSpellLoader() *SpellLoader {
	return &SpellLoader{}
}

// LoadFromFile loads spell metadata from a spell.yaml file
func (l *SpellLoader) LoadFromFile(filename string) (*SpellMetadata, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read spell file: %w", err)
	}

	var metadata SpellMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse spell.yaml: %w", err)
	}

	// Set the root directory
	metadata.RootDir = filepath.Dir(filename)

	// Validate the metadata
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid spell metadata: %w", err)
	}

	return &metadata, nil
}

// LoadFromDirectory loads spell metadata from a directory containing spell.yaml
func (l *SpellLoader) LoadFromDirectory(dir string) (*SpellMetadata, error) {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dir)
	}

	// Look for spell.yaml
	spellFile := filepath.Join(dir, "spell.yaml")
	if _, err := os.Stat(spellFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("spell.yaml not found in directory: %s", dir)
		}
		return nil, fmt.Errorf("failed to access spell.yaml: %w", err)
	}

	// Load the spell metadata
	metadata, err := l.LoadFromFile(spellFile)
	if err != nil {
		return nil, err
	}

	// Override root directory to the actual directory
	metadata.RootDir = dir

	return metadata, nil
}

// ResolveEntryPoint resolves the entry point path relative to the spell root
func (l *SpellLoader) ResolveEntryPoint(metadata *SpellMetadata) string {
	// If entry point is absolute, return as-is
	if filepath.IsAbs(metadata.EntryPoint) {
		return metadata.EntryPoint
	}

	// If no root directory, return entry point as-is
	if metadata.RootDir == "" {
		return metadata.EntryPoint
	}

	// Join with root directory
	return filepath.Join(metadata.RootDir, metadata.EntryPoint)
}

// ValidateParameters validates that the provided parameters match the spell's requirements
func (l *SpellLoader) ValidateParameters(metadata *SpellMetadata, params map[string]interface{}) error {
	// Check required parameters
	for _, param := range metadata.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return fmt.Errorf("required parameter missing: %s", param.Name)
			}
		}
	}

	// Could add type validation here in the future

	return nil
}

// ApplyDefaults applies default values to parameters that weren't provided
func (l *SpellLoader) ApplyDefaults(metadata *SpellMetadata, params map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing parameters
	for k, v := range params {
		result[k] = v
	}

	// Apply defaults for missing parameters
	for _, param := range metadata.Parameters {
		if _, ok := result[param.Name]; !ok && param.Default != nil {
			result[param.Name] = param.Default
		}
	}

	return result
}
