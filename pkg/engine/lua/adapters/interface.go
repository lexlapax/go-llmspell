// ABOUTME: Common interface definitions for Lua bridge adapters ensuring consistency and interoperability.
// ABOUTME: Defines LuaModuleProvider interface that all adapters must implement for factory pattern integration.

package adapters

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// LuaModuleProvider defines the interface that all Lua adapters must implement.
// This ensures consistent behavior across all adapter types and enables
// the factory pattern to create adapters uniformly.
type LuaModuleProvider interface {
	// CreateLuaModule returns a Lua module loader function.
	// This function will be called by the engine to create a Lua module
	// that exposes the adapter's functionality to Lua scripts.
	CreateLuaModule() lua.LGFunction
}

// BridgeProvider defines the interface for adapters that wrap engine bridges.
// This allows access to the underlying bridge for metadata and validation purposes.
type BridgeProvider interface {
	// GetBridge returns the primary wrapped bridge.
	// For multi-bridge adapters, this returns the main bridge.
	GetBridge() engine.Bridge
	
	// GetID returns the bridge ID.
	// This should delegate to the wrapped bridge's GetID method.
	GetID() string
	
	// GetMetadata returns the bridge metadata.
	// This should delegate to the wrapped bridge's GetMetadata method.
	GetMetadata() engine.BridgeMetadata
}

// TypeConversionProvider defines the interface for adapters that support custom type conversion.
// This allows adapters to handle complex data transformations between Go and Lua types.
type TypeConversionProvider interface {
	// GetMethods returns the available method names.
	// This provides introspection capabilities for the adapter.
	GetMethods() []string
	
	// GetMethodInfo returns information about a specific method.
	// This supports validation and documentation generation.
	GetMethodInfo(name string) (engine.MethodInfo, error)
}

// FullAdapter combines all adapter interfaces for comprehensive adapter functionality.
// Adapters should implement this interface to ensure full compatibility with
// the engine and factory systems.
type FullAdapter interface {
	LuaModuleProvider
	BridgeProvider
	TypeConversionProvider
}

// ValidationProvider defines the interface for adapters that support validation.
// This allows adapters to enable or disable method argument validation.
type ValidationProvider interface {
	// EnableValidation enables or disables method argument validation.
	// When enabled, adapters should validate arguments before calling bridge methods.
	EnableValidation(enable bool)
}

// AdapterMetadata provides metadata about an adapter implementation.
// This supports documentation generation and adapter introspection.
type AdapterMetadata struct {
	// Name is the human-readable name of the adapter
	Name string
	
	// Type describes the category of functionality (e.g., "state", "llm", "observability")
	Type string
	
	// Description provides a detailed explanation of the adapter's purpose
	Description string
	
	// BridgeIDs lists the bridge IDs that this adapter can handle
	BridgeIDs []string
	
	// RequiredBridges lists bridge IDs that are required for this adapter to function
	RequiredBridges []string
	
	// OptionalBridges lists bridge IDs that enhance functionality but are not required
	OptionalBridges []string
	
	// SupportsMultiBridge indicates if this adapter handles multiple bridges
	SupportsMultiBridge bool
}

// MetadataProvider defines the interface for adapters that provide metadata.
// This supports adapter discovery and documentation generation.
type MetadataProvider interface {
	// GetAdapterMetadata returns metadata about this adapter implementation.
	// This provides information about supported bridges, requirements, and capabilities.
	GetAdapterMetadata() AdapterMetadata
}

// ErrorHandlingProvider defines the interface for adapters with custom error handling.
// This allows adapters to provide domain-specific error formatting and recovery.
type ErrorHandlingProvider interface {
	// FormatError formats an error for display to Lua scripts.
	// This allows adapters to provide context-specific error messages.
	FormatError(err error) string
	
	// HandleMethodError handles errors that occur during method execution.
	// This allows adapters to implement recovery strategies or detailed logging.
	HandleMethodError(methodName string, err error) error
}