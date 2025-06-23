// ABOUTME: Implementation of the run command for executing spell scripts.
// ABOUTME: Handles script execution with parameter passing and engine selection.

package commands

import (
	"context"
	"os"

	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// RunCmd executes a spell script.
// It handles script loading, engine selection, parameter passing,
// and execution with configurable timeout.
type RunCmd struct {
	BaseCommand
	Script     string            `arg:"" help:"Script file to execute" type:"existingfile"`
	Parameters map[string]string `short:"p" help:"Parameters to pass to the script (key=value)"`
	Engine     string            `short:"e" help:"Script engine to use (auto-detected if not specified)"`
	Timeout    int               `short:"t" help:"Execution timeout in seconds" default:"300"`
}

// Run executes the command.
// It gets the script runner from context and executes the script
// with the provided parameters and timeout.
func (c *RunCmd) Run(ctx context.Context) error {
	// Get script runner from context
	runnerInterface := GetRunner(ctx)
	if runnerInterface == nil {
		return errors.New(errors.CategoryConfig, "script runner not found in context")
	}

	scriptRunner, ok := runnerInterface.(runner.Runner)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid script runner type")
	}

	// Convert string parameters to interface{}
	params := make(map[string]interface{})
	for k, v := range c.Parameters {
		params[k] = v
	}

	// Execute the script
	c.Debug(ctx, "Executing script: %s", c.Script)

	// If engine is specified, we need to read the file and use Execute
	if c.Engine != "" {
		// Read the script file
		scriptContent, err := os.ReadFile(c.Script)
		if err != nil {
			return errors.Wrap(err, errors.CategoryIO, "failed to read script file")
		}

		// Execute the script content directly
		result, err := scriptRunner.Execute(ctx, string(scriptContent), params)
		if err != nil {
			return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
		}

		// Print result if not nil
		if result != nil {
			c.Printf("%v\n", result)
		}
	} else {
		// Use ExecuteFile which will auto-detect the engine
		result, err := scriptRunner.ExecuteFile(ctx, c.Script, params)
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
