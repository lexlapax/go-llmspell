// ABOUTME: Temporary bridge interfaces for async bridge development
// ABOUTME: This will be updated during the ScriptValue refactoring in Task 2.3.2.0

package bridge

import (
	"context"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Bridge interface extension for ScriptValue support
// TODO: This will be merged into the main Bridge interface during refactoring
type Bridge interface {
	GetID() string
	GetMetadata() BridgeMetadata
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	GetMethods() map[string]MethodInfo
	GetTypeMappings() map[string]TypeMapping
	IsInitialized() bool
	GetRequiredPermissions() []string
	ValidateMethod(method string, args []engine.ScriptValue) error
	ExecuteMethod(ctx context.Context, method string, args []engine.ScriptValue) (engine.ScriptValue, error)
}

// BridgeMetadata contains metadata about a bridge
type BridgeMetadata struct {
	Name        string
	Version     string
	Description string
}

// MethodInfo describes a method exposed by a bridge
type MethodInfo struct {
	Description string
	Parameters  []ParameterInfo
	Returns     []ReturnInfo
}

// ParameterInfo describes a method parameter
type ParameterInfo struct {
	Name     string
	Type     string
	Required bool
}

// ReturnInfo describes a method return value
type ReturnInfo struct {
	Type        string
	Description string
}

// TypeMapping defines type conversion hints
type TypeMapping struct {
	GoType     string
	ScriptType string
}
