// ABOUTME: Implementation of the debug command for debugging spell scripts.
// ABOUTME: Provides debugging capabilities with breakpoints and step execution.

package commands

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// DebugCmd debugs a spell script
type DebugCmd struct {
	BaseCommand
	Script      string `arg:"" help:"Script file to debug" type:"existingfile"`
	Breakpoints []int  `short:"b" help:"Line numbers for breakpoints"`
}

// Run executes the command
func (c *DebugCmd) Run(ctx context.Context) error {
	c.Printf("Debugging script: %s\n", c.Script)
	if len(c.Breakpoints) > 0 {
		c.Printf("Breakpoints at lines: %v\n", c.Breakpoints)
	}
	return errors.New(errors.CategoryUsage, "debug command not implemented yet")
}
