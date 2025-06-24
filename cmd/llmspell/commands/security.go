// ABOUTME: Implementation of the security command for managing security levels and feature sets.
// ABOUTME: Supports list, show, and validate actions for security configuration.

package commands

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/bridge/registry"
	"github.com/lexlapax/go-llmspell/pkg/errors"
	"github.com/lexlapax/go-llmspell/pkg/security"
)

// SecurityCmd manages security levels and feature sets.
// It provides commands to list available options, show details,
// and validate configurations.
type SecurityCmd struct {
	BaseCommand
	Action        string `arg:"" help:"Action to perform: list (show all options), show (display details), validate (check validity)" enum:"list,show,validate" default:"list"`
	SecurityLevel string `arg:"" optional:"" help:"Security level name (untrusted, trusted, privileged)"`
}

// Run executes the command.
// It performs the requested security action: list all options,
// show details of a specific configuration, or validate settings.
func (c *SecurityCmd) Run(ctx context.Context) error {
	switch c.Action {
	case "list":
		c.Println("Available security levels:")
		for _, level := range GetAvailableSecurityLevels() {
			desc := GetSecurityLevelDescription(security.SecurityLevel(level))
			c.Printf("  - %s: %s\n", level, desc)
		}

		c.Println("\nAvailable feature sets:")
		for _, fs := range GetAvailableFeatureSets() {
			desc := GetFeatureSetDescription(registry.FeatureSet(fs))
			c.Printf("  - %s: %s\n", fs, desc)
		}
		return nil

	case "show":
		if c.SecurityLevel == "" {
			c.SecurityLevel = string(GetSecurityLevel(ctx))
		}

		// Validate security level
		if !security.IsValidLevel(c.SecurityLevel) {
			return errors.Newf(errors.CategoryUsage, "invalid security level: %s", c.SecurityLevel)
		}

		// Show security level details
		c.Printf("Security Level: %s\n", c.SecurityLevel)
		config := security.GetLevelConfig(security.SecurityLevel(c.SecurityLevel))
		if config != nil {
			c.Printf("Description: %s\n", config.Description)
			c.Println("\nPermissions:")
			c.Printf("  - Network: %v\n", config.AllowNetwork)
			c.Printf("  - Filesystem: %v\n", config.AllowFilesystem)
			c.Printf("  - Environment: %v\n", config.AllowEnvironment)
			c.Printf("  - Execute: %v\n", config.AllowExec)
			c.Printf("  - Unsafe: %v\n", config.AllowUnsafe)

			c.Println("\nResource Limits:")
			c.Printf("  - Memory: %d MB\n", config.MemoryLimit/(1024*1024))
			c.Printf("  - CPU: %d%%\n", config.CPULimit)
			c.Printf("  - Timeout: %d seconds\n", config.TimeoutSeconds)
		}

		// Show feature set details
		featureSet := GetFeatureSet(ctx)
		c.Printf("\nFeature Set: %s\n", featureSet)
		desc := registry.GetFeatureSetDescription(featureSet)
		c.Printf("Description: %s\n", desc)

		bridgeSets := registry.GetBridgeSetsForFeature(featureSet)
		if len(bridgeSets) > 0 {
			c.Println("\nEnabled Bridge Sets:")
			for _, bs := range bridgeSets {
				c.Printf("  - %s\n", bs)
			}
		}

		return nil

	case "validate":
		featureSet := GetFeatureSet(ctx)
		if c.SecurityLevel == "" && string(featureSet) == "" {
			return errors.New(errors.CategoryUsage, "security level or feature set required")
		}

		// Validate security level if provided
		if c.SecurityLevel != "" {
			if _, err := ValidateSecurityLevel(c.SecurityLevel); err != nil {
				return errors.Wrap(err, errors.CategoryValidation, "invalid security level")
			}
			c.Printf("✓ Security level '%s' is valid\n", c.SecurityLevel)
		}

		// Validate feature set if provided
		if string(featureSet) != "" {
			if _, err := ValidateFeatureSet(string(featureSet)); err != nil {
				return errors.Wrap(err, errors.CategoryValidation, "invalid feature set")
			}
			c.Printf("✓ Feature set '%s' is valid\n", featureSet)
		}

		return nil

	default:
		return errors.Newf(errors.CategoryUsage, "unknown action: %s", c.Action)
	}
}
