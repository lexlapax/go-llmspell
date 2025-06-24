// ABOUTME: Implementation of the validate command for validating spell and script files.
// ABOUTME: Validates spell.yaml files and script syntax using the appropriate engine.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// ValidateCmd validates a spell or script.
// It supports both spell.yaml files and script files,
// performing syntax validation and basic checks.
type ValidateCmd struct {
	BaseCommand
	Path string `arg:"" help:"Path to spell.yaml or script file" type:"existingfile"`
}

// Run executes the command.
// It determines the file type based on extension,
// then validates using the appropriate validator.
func (c *ValidateCmd) Run(ctx context.Context) error {
	// Get runner from context
	runnerInterface := GetRunner(ctx)
	if runnerInterface == nil {
		return errors.New(errors.CategoryConfig, "runner not found in context")
	}

	// Extract engine registry from runner
	var engineRegistry *runner.EngineRegistryManager
	if registryProvider, ok := runnerInterface.(interface{ GetEngineRegistry() interface{} }); ok {
		if em, ok := registryProvider.GetEngineRegistry().(*runner.EngineRegistryManager); ok {
			engineRegistry = em
		} else {
			return errors.New(errors.CategoryConfig, "runner engine registry is not the expected type")
		}
	} else {
		return errors.New(errors.CategoryConfig, "runner does not provide engine registry")
	}

	// Check if it's a spell file or script
	ext := filepath.Ext(c.Path)
	if ext == ".yaml" || ext == ".yml" {
		// Validate spell file
		loader := runner.NewSpellLoader()
		spell, err := loader.LoadFromFile(c.Path)
		if err != nil {
			return errors.Wrap(err, errors.CategoryValidation, "failed to load spell file")
		}

		c.Printf("✓ Spell file is valid\n")
		c.Printf("  Name: %s\n", spell.Name)
		c.Printf("  Version: %s\n", spell.Version)
		c.Printf("  Entry Point: %s\n", spell.EntryPoint)
		if spell.Engine != "" {
			c.Printf("  Engine: %s\n", spell.Engine)
		}

		return nil
	}

	// It's a script file - validate using engine
	selector := runner.NewEngineSelector(engineRegistry)
	engineName, err := selector.SelectByExtension(c.Path)
	if err != nil {
		return errors.Wrap(err, errors.CategoryValidation, "unable to determine script engine")
	}

	// Check if engine is available
	if _, err := engineRegistry.GetEngineInfo(engineName); err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "engine not available")
	}

	// Read the script file
	scriptContent, err := os.ReadFile(c.Path)
	if err != nil {
		return errors.Wrap(err, errors.CategoryIO, "failed to read script file")
	}

	// Get the engine to validate the script
	config := engine.EngineConfig{
		DebugMode: IsDebug(ctx),
	}
	scriptEngine, err := engineRegistry.GetEngine(engineName, config, security.SecurityLevelTrusted, registry.FeatureSetMinimal)
	if err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "failed to get engine")
	}

	// Validate the script
	if validator, ok := scriptEngine.(interface {
		Validate(script string) error
	}); ok {
		if err := validator.Validate(string(scriptContent)); err != nil {
			return errors.Wrap(err, errors.CategoryValidation, "script validation failed")
		}
	}

	c.Printf("✓ Script file is valid (%s engine)\n", engineName)

	// Additional validation info if verbose
	if IsVerbose(ctx) {
		c.Printf("  Path: %s\n", c.Path)
		c.Printf("  Engine: %s\n", engineName)
		c.Printf("  Size: %d bytes\n", len(scriptContent))
	}

	return nil
}
