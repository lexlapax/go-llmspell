// ABOUTME: This file implements engine selection logic based on file extensions and spell metadata.
// ABOUTME: It provides utilities for determining which script engine to use for execution.

package runner

import (
	"fmt"
	"path/filepath"
	"strings"
)

// EngineSelector handles engine selection logic.
// It determines which script engine to use based on file extensions,
// spell metadata, and runtime options.
type EngineSelector struct {
	manager *EngineRegistryManager
}

// NewEngineSelector creates a new engine selector.
// It uses the provided engine registry manager for engine discovery
// and validation.
func NewEngineSelector(manager *EngineRegistryManager) *EngineSelector {
	return &EngineSelector{
		manager: manager,
	}
}

// SelectByExtension selects an engine based on file extension.
// It extracts the extension from the filepath and finds the appropriate
// engine that can handle files with that extension.
func (s *EngineSelector) SelectByExtension(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("empty filepath provided")
	}

	ext := extractExtension(filepath)
	if ext == "" {
		return "", fmt.Errorf("no file extension found in %s", filepath)
	}

	engine, err := s.manager.FindEngineByExtension(ext)
	if err != nil {
		return "", fmt.Errorf("no engine found for extension .%s: %w", ext, err)
	}

	return engine, nil
}

// SelectForSpell selects an engine for a spell based on its metadata.
// It prioritizes explicitly specified engines, then attempts to determine
// the engine from the entry point file extension.
func (s *EngineSelector) SelectForSpell(metadata *SpellMetadata) (string, error) {
	// Priority 1: Explicit engine in metadata
	if metadata.Engine != "" {
		if err := s.ValidateEngineAvailability(metadata.Engine); err != nil {
			return "", fmt.Errorf("specified engine %s: %w", metadata.Engine, err)
		}
		return metadata.Engine, nil
	}

	// Priority 2: Determine from entry point extension
	if metadata.EntryPoint != "" {
		engine, err := s.SelectByExtension(metadata.EntryPoint)
		if err == nil {
			return engine, nil
		}
		// Continue if we can't determine from extension
	}

	return "", fmt.Errorf("unable to determine engine for spell %s: no engine specified and cannot determine from entry point", metadata.Name)
}

// SelectWithOptions selects an engine considering both spell metadata and runtime options.
// Runtime options take precedence over spell metadata, allowing dynamic
// engine selection at execution time.
func (s *EngineSelector) SelectWithOptions(metadata *SpellMetadata, options *RunnerOptions) (string, error) {
	// Options override metadata
	if options != nil && options.Engine != "" {
		if err := s.ValidateEngineAvailability(options.Engine); err != nil {
			return "", fmt.Errorf("option-specified engine %s: %w", options.Engine, err)
		}
		return options.Engine, nil
	}

	// Fall back to spell metadata
	return s.SelectForSpell(metadata)
}

// ValidateEngineAvailability checks if an engine is available.
// It verifies the engine exists and is in an active state,
// returning an error if the engine is unavailable or inactive.
func (s *EngineSelector) ValidateEngineAvailability(engineName string) error {
	info, err := s.manager.GetEngineInfo(engineName)
	if err != nil {
		return fmt.Errorf("not available: %w", err)
	}

	if info.Status == "error" || info.Status == "inactive" {
		return fmt.Errorf("engine %s is %s", engineName, info.Status)
	}

	return nil
}

// GetSupportedExtensions returns all supported file extensions.
// It aggregates extensions from all registered engines, returning
// a unique list of file types that can be executed.
func (s *EngineSelector) GetSupportedExtensions() []string {
	engines := s.manager.ListEngines()
	extensionSet := make(map[string]bool)

	for _, engine := range engines {
		for _, ext := range engine.FileExtensions {
			extensionSet[ext] = true
		}
	}

	extensions := make([]string, 0, len(extensionSet))
	for ext := range extensionSet {
		extensions = append(extensions, ext)
	}

	return extensions
}

// GetEngineExtensionMap returns a map of extensions to engine names.
// For each supported file extension, it maps to the engine that
// handles it (first engine to claim an extension wins).
func (s *EngineSelector) GetEngineExtensionMap() map[string]string {
	engines := s.manager.ListEngines()
	extensionMap := make(map[string]string)

	for _, engine := range engines {
		for _, ext := range engine.FileExtensions {
			// First engine to claim an extension wins
			if _, exists := extensionMap[ext]; !exists {
				extensionMap[ext] = engine.Name
			}
		}
	}

	return extensionMap
}

// extractExtension extracts the file extension from a path.
// It removes the leading dot and converts to lowercase for
// consistent comparison.
func extractExtension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	// Remove the dot and convert to lowercase
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// EngineSelectorError represents an engine selection error.
// It provides detailed information about why engine selection failed,
// including available alternatives.
type EngineSelectorError struct {
	Filepath            string
	RequestedEngine     string
	AvailableEngines    []string
	SupportedExtensions []string
}

// Error returns a human-readable error message.
// It provides specific error messages based on whether a specific
// engine was requested or file extension matching failed.
func (e *EngineSelectorError) Error() string {
	if e.RequestedEngine != "" {
		return fmt.Sprintf("engine '%s' not found. Available engines: %s",
			e.RequestedEngine, strings.Join(e.AvailableEngines, ", "))
	}
	if e.Filepath != "" {
		return fmt.Sprintf("no engine found for file '%s'. Supported extensions: %s",
			e.Filepath, strings.Join(e.SupportedExtensions, ", "))
	}
	return "unable to select engine"
}
