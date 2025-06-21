// ABOUTME: Implementation of the validate command for validating spell and script files.
// ABOUTME: Validates spell.yaml files and script syntax using the appropriate engine.

package commands

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// ValidateCmd validates a spell or script
type ValidateCmd struct {
	BaseCommand
	Path string `arg:"" help:"Path to spell.yaml or script file" type:"existingfile"`
}

// Run executes the command
func (c *ValidateCmd) Run(ctx context.Context) error {
	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		return errors.New(errors.CategoryConfig, "engine registry not found in context")
	}

	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
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
	scriptEngine, err := engineRegistry.GetEngine(engineName, config)
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
