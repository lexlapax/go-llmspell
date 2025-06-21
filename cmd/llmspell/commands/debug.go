// ABOUTME: Implementation of the debug command for debugging spell scripts.
// ABOUTME: Provides debugging capabilities with breakpoints and step execution.

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// DebugCmd debugs a spell script
type DebugCmd struct {
	BaseCommand
	Script      string            `arg:"" help:"Script file to debug" type:"existingfile"`
	Breakpoints []int             `short:"b" help:"Line numbers for breakpoints"`
	Engine      string            `help:"Script engine to use" default:"lua"`
	Env         map[string]string `help:"Environment variables to set"`
	StepMode    bool              `help:"Enable step-by-step execution"`
	Timeout     int               `short:"t" help:"Execution timeout in seconds" default:"300"`
}

// Run executes the debug command
func (c *DebugCmd) Run(ctx context.Context) error {
	// Read script content first to validate file exists
	scriptContent, err := os.ReadFile(c.Script)
	if err != nil {
		return errors.Wrap(err, errors.CategoryIO, "failed to read script")
	}

	// Get engine registry from context
	engineRegistryInterface := GetEngineRegistry(ctx)
	if engineRegistryInterface == nil {
		return errors.New(errors.CategoryConfig, "engine registry not found in context")
	}

	engineRegistry, ok := engineRegistryInterface.(*runner.EngineRegistryManager)
	if !ok {
		return errors.New(errors.CategoryConfig, "invalid engine registry type")
	}

	// Print debug header
	c.Errorf("=== Debug Session Started ===\n")
	c.Errorf("Script: %s\n", c.Script)
	c.Errorf("Engine: %s\n", c.Engine)
	if c.StepMode {
		c.Errorf("Mode: Step-by-step\n")
	}
	if len(c.Breakpoints) > 0 {
		c.Errorf("Breakpoints at lines: %v\n", c.Breakpoints)
	}
	c.Errorf("===========================\n\n")

	// Create runner config with debug enabled
	runnerConfig := &runner.RunnerConfig{
		MaxConcurrentScripts: 1,
		Timeout:              time.Duration(c.Timeout) * time.Second,
		EngineConfigs:        make(map[string]map[string]interface{}),
		DefaultEngine:        c.Engine,
	}

	// Add debug configuration for the engine
	runnerConfig.EngineConfigs[c.Engine] = map[string]interface{}{
		"debug":       true,
		"verbose":     IsVerbose(ctx),
		"stepMode":    c.StepMode,
		"breakpoints": c.Breakpoints,
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

	// Set up debug environment
	debugParams := make(map[string]interface{})
	for k, v := range c.Env {
		debugParams[k] = v
	}

	// Add debug-specific parameters
	debugParams["DEBUG"] = true
	debugParams["STEP_MODE"] = c.StepMode
	if len(c.Breakpoints) > 0 {
		// Convert int breakpoints to strings
		breakpointStrs := make([]string, len(c.Breakpoints))
		for i, bp := range c.Breakpoints {
			breakpointStrs[i] = fmt.Sprintf("%d", bp)
		}
		debugParams["BREAKPOINTS"] = strings.Join(breakpointStrs, ",")
	}

	// Create runner options with debug settings
	runnerOptions := &runner.RunnerOptions{
		Engine:     c.Engine,
		Parameters: debugParams,
		Timeout:    time.Duration(c.Timeout) * time.Second,
		Debug:      true,
	}

	// Execute script in debug mode
	c.Debug(ctx, "Executing script in debug mode: %s", c.Script)

	execResult, err := executor.ExecuteWithOptions(ctx, string(scriptContent), runnerOptions)
	if err != nil {
		// In debug mode, provide more detailed error information
		c.Errorf("\n=== Debug Error Details ===\n")
		c.Errorf("Error: %v\n", err)

		// Try to extract more details from the error
		if spellErr, ok := err.(*errors.SpellError); ok {
			c.Errorf("Category: %s\n", spellErr.Category)
			if spellErr.Context != nil {
				for k, v := range spellErr.Context {
					c.Errorf("%s: %v\n", k, v)
				}
			}
			// Print stack trace if available
			if len(spellErr.StackTrace) > 0 && IsVerbose(ctx) {
				c.Errorf("\nStack Trace:\n")
				for _, frame := range spellErr.StackTrace {
					c.Errorf("  %s\n    %s:%d\n", frame.Function, frame.File, frame.Line)
				}
			}
		}

		c.Errorf("==========================\n")
		return errors.Wrap(err, errors.CategoryScript, "debug execution failed")
	}

	// Print debug footer with results
	c.Errorf("\n=== Debug Session Complete ===\n")

	// Handle execution result
	if execResult.IsError() {
		c.Errorf("Script completed with error: %v\n", execResult.Error)
	} else if execResult.Value != nil {
		c.Printf("%v\n", execResult.Value)
	}

	// Print execution metadata if verbose
	if IsVerbose(ctx) && execResult.Metadata != nil {
		c.Errorf("\nExecution Metadata:\n")
		for k, v := range execResult.Metadata {
			c.Errorf("  %s: %v\n", k, v)
		}
	}

	c.Errorf("=============================\n")

	// Return any error from the execution result
	if execResult.IsError() {
		return execResult.Error
	}

	return nil
}
