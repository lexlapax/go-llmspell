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

	// Cast to ScriptExecutor to access ExecuteWithOptions
	scriptExecutor, ok := runnerInterface.(*runner.ScriptExecutor)
	if !ok {
		return errors.New(errors.CategoryConfig, "runner is not a ScriptExecutor")
	}

	// Convert string parameters to interface{}
	params := make(map[string]interface{})
	for k, v := range c.Parameters {
		params[k] = v
	}

	// Get security profile from context
	securityProfile := GetProfile(ctx)

	// Execute the script
	c.Debug(ctx, "Executing script: %s with profile: %s", c.Script, securityProfile)

	// If engine is specified, we need to read the file and use Execute
	if c.Engine != "" {
		// Read the script file
		scriptContent, err := os.ReadFile(c.Script)
		if err != nil {
			return errors.Wrap(err, errors.CategoryIO, "failed to read script file")
		}

		// Create options with security profile
		options := &runner.RunnerOptions{
			Parameters:      params,
			Engine:          c.Engine,
			SecurityProfile: securityProfile,
		}

		// Execute the script content directly with options
		result, err := scriptExecutor.ExecuteWithOptions(ctx, string(scriptContent), options)
		if err != nil {
			return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
		}

		// Print result if not nil
		if result.Value != nil {
			c.Printf("%v\n", result.Value)
		}
	} else {
		// For file execution, read the file and use ExecuteWithOptions
		scriptContent, err := os.ReadFile(c.Script)
		if err != nil {
			return errors.Wrap(err, errors.CategoryIO, "failed to read script file")
		}

		// Create options with security profile for file execution
		options := &runner.RunnerOptions{
			Parameters:      params,
			SecurityProfile: securityProfile,
		}

		// Execute the file content with options
		result, err := scriptExecutor.ExecuteWithOptions(ctx, string(scriptContent), options)
		if err != nil {
			return errors.Wrap(err, errors.CategoryScript, "failed to execute script")
		}

		// Print result if not nil
		if result.Value != nil {
			c.Printf("%v\n", result.Value)
		}
	}

	return nil
}
