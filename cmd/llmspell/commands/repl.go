// ABOUTME: Implementation of the repl command for starting an interactive REPL.
// ABOUTME: Provides an interactive script execution environment.

package commands

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/errors"
)

// REPLCmd starts an interactive REPL
type REPLCmd struct {
	BaseCommand
	Engine string `short:"e" help:"Script engine to use" default:"lua"`
}

// Run executes the command
func (c *REPLCmd) Run(ctx context.Context) error {
	c.Printf("Starting REPL with %s engine...\n", c.Engine)
	return errors.New(errors.CategoryUsage, "REPL not implemented yet")
}
