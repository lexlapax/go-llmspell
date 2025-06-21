// ABOUTME: Implementation of the run command for executing spell scripts.
// ABOUTME: Handles script execution with parameter passing and engine selection.

package commands

import (
	"context"
	"os"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// RunCmd executes a spell script
type RunCmd struct {
	BaseCommand
	Script     string            `arg:"" help:"Script file to execute" type:"existingfile"`
	Parameters map[string]string `short:"p" help:"Parameters to pass to the script (key=value)"`
	Engine     string            `short:"e" help:"Script engine to use (auto-detected if not specified)"`
	Timeout    int               `short:"t" help:"Execution timeout in seconds" default:"300"`
}

// Run executes the command
func (c *RunCmd) Run(ctx context.Context) error {
	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		return errors.New(errors.CategoryConfig, "engine registry not found in context")
	}

	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
	}

	// Create runner config
	runnerConfig := &runner.RunnerConfig{
		MaxConcurrentScripts: 1,
		Timeout:              time.Duration(c.Timeout) * time.Second,
		EngineConfigs:        make(map[string]map[string]interface{}),
		DefaultEngine:        "lua",
	}

	// Create engine selector
	selector := runner.NewEngineSelector(engineRegistry)

	// Create script executor
	executor := runner.NewScriptExecutor(runnerConfig, engineRegistry, selector)

	// Initialize executor
	if err := executor.Initialize(ctx); err != nil {
		return errors.Wrap(err, errors.CategoryEngine, "failed to initialize executor")
	}
	defer func() {
		if err := executor.Shutdown(); err != nil {
			c.Errorf("failed to shutdown executor: %v\n", err)
		}
	}()

	// Convert string parameters to interface{}
	params := make(map[string]interface{})
	for k, v := range c.Parameters {
		params[k] = v
	}

	// Execute the script
	c.Debug(ctx, "Executing script: %s", c.Script)

	// If engine is specified, we need to read the file and use ExecuteWithOptions
	if c.Engine != "" {
		// Read the script file
		scriptContent, err := os.ReadFile(c.Script)
		if err != nil {
			return errors.Wrap(err, errors.CategoryIO, "failed to read script file")
		}

		// Create runner options with specified engine
		runnerOptions := &runner.RunnerOptions{
			Engine:     c.Engine,
			Parameters: params,
			Timeout:    time.Duration(c.Timeout) * time.Second,
		}

		// Execute with options
		execResult, err := executor.ExecuteWithOptions(ctx, string(scriptContent), runnerOptions)
		if err != nil {
			return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
		}

		// Handle execution result
		if execResult.IsError() {
			return errors.Wrap(execResult.Error, errors.CategoryScript, "script execution failed")
		}

		// Print result if not nil
		if execResult.Value != nil {
			c.Printf("%v\n", execResult.Value)
		}
	} else {
		// Use ExecuteFile which will auto-detect the engine
		result, err := executor.ExecuteFile(ctx, c.Script, params)
		if err != nil {
			return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
		}

		// Print result if not nil
		if result != nil {
			c.Printf("%v\n", result)
		}
	}

	return nil
}
