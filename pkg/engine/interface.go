// ABOUTME: This file defines the core interfaces for the multi-engine scripting architecture.
// ABOUTME: It provides engine-agnostic abstractions for script execution, bridging, and type conversion.

package engine

import (
	"context"
	"time"
)

// ScriptEngine defines the common interface that all scripting engines must implement.
// This abstraction allows go-llmspell to support multiple scripting languages
// (Lua, JavaScript, Tengo) through a unified API.
type ScriptEngine interface {
	// Lifecycle management
	Initialize(config EngineConfig) error
	Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error)
	ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error)
	Shutdown() error

	// Bridge management - allows engines to register functionality from Go
	RegisterBridge(bridge Bridge) error
	UnregisterBridge(name string) error
	GetBridge(name string) (Bridge, error)
	ListBridges() []string

	// Type system - handles conversion between engine types and Go types
	ToNative(scriptValue interface{}) (interface{}, error)
	FromNative(goValue interface{}) (interface{}, error)

	// Metadata and capabilities
	Name() string
	Version() string
	FileExtensions() []string
	Features() []EngineFeature

	// Resource management and security
	SetMemoryLimit(bytes int64) error
	SetTimeout(duration time.Duration) error
	SetResourceLimits(limits ResourceLimits) error
	GetMetrics() EngineMetrics

	// Script state management
	CreateContext(options ContextOptions) (ScriptContext, error)
	DestroyContext(ctx ScriptContext) error
	ExecuteScript(ctx context.Context, script string, options ExecutionOptions) (*ExecutionResult, error)
}

// Bridge defines the interface for functionality that can be exposed to scripts.
// Bridges are engine-agnostic and handle the translation between Go functions
// and script-callable methods.
type Bridge interface {
	// Identity and metadata
	GetID() string
	GetMetadata() BridgeMetadata

	// Lifecycle management
	Initialize(ctx context.Context) error
	Cleanup(ctx context.Context) error
	IsInitialized() bool

	// Engine registration
	RegisterWithEngine(engine ScriptEngine) error

	// Method exposure
	Methods() []MethodInfo
	ValidateMethod(name string, args []interface{}) error

	// Type conversion hints for engines
	TypeMappings() map[string]TypeMapping

	// Security
	RequiredPermissions() []Permission
}

// BridgeMetadata contains metadata about a bridge.
type BridgeMetadata struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Author       string   `json:"author"`
	License      string   `json:"license"`
}

// TypeConverter handles conversion between Go types and script engine types.
// Each engine implements this interface to handle its specific type system.
type TypeConverter interface {
	// Basic type conversions
	ToBoolean(v interface{}) (bool, error)
	ToNumber(v interface{}) (float64, error)
	ToString(v interface{}) (string, error)
	ToArray(v interface{}) ([]interface{}, error)
	ToMap(v interface{}) (map[string]interface{}, error)

	// Complex type handling
	ToStruct(v interface{}, target interface{}) error
	FromStruct(v interface{}) (map[string]interface{}, error)

	// Function and callback handling
	ToFunction(v interface{}) (Function, error)
	FromFunction(fn Function) (interface{}, error)

	// Engine-specific type support
	SupportsType(typeName string) bool
	GetTypeInfo(typeName string) TypeInfo
}

// EngineConfig holds configuration parameters for script engine initialization.
type EngineConfig struct {
	// Resource limits
	MemoryLimit    int64         `json:"memory_limit"`
	TimeoutLimit   time.Duration `json:"timeout_limit"`
	GoroutineLimit int           `json:"goroutine_limit"`

	// Security settings
	SandboxMode     bool     `json:"sandbox_mode"`
	AllowedModules  []string `json:"allowed_modules"`
	DisabledModules []string `json:"disabled_modules"`
	FileSystemMode  FSMode   `json:"filesystem_mode"`

	// Engine-specific settings
	EngineOptions map[string]interface{} `json:"engine_options"`

	// Debugging and observability
	DebugMode   bool   `json:"debug_mode"`
	LogLevel    string `json:"log_level"`
	MetricsMode bool   `json:"metrics_mode"`
	TracingMode bool   `json:"tracing_mode"`
}

// ContextOptions defines options for creating a script context.
type ContextOptions struct {
	ID           string                 `json:"id"`
	MemoryLimit  int64                  `json:"memory_limit"`
	TimeoutLimit time.Duration          `json:"timeout_limit"`
	Variables    map[string]interface{} `json:"variables"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ExecutionOptions defines options for script execution.
type ExecutionOptions struct {
	Timeout         time.Duration          `json:"timeout"`
	MemoryLimit     int64                  `json:"memory_limit"`
	Context         ScriptContext          `json:"-"`
	Variables       map[string]interface{} `json:"variables"`
	CaptureOutput   bool                   `json:"capture_output"`
	ReturnLastValue bool                   `json:"return_last_value"`
}

// ExecutionResult contains the result of script execution.
type ExecutionResult struct {
	Value    interface{}            `json:"value"`
	Output   string                 `json:"output"`
	Error    error                  `json:"-"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ResourceLimits defines resource constraints for script execution.
type ResourceLimits struct {
	MaxMemory     int64         `json:"max_memory"`
	MaxGoroutines int           `json:"max_goroutines"`
	MaxExecTime   time.Duration `json:"max_exec_time"`
	MaxFileSize   int64         `json:"max_file_size"`
	MaxNetworkOps int           `json:"max_network_ops"`
}

// EngineMetrics provides runtime metrics for script engine performance.
type EngineMetrics struct {
	// Execution metrics
	ScriptsExecuted int64         `json:"scripts_executed"`
	TotalExecTime   time.Duration `json:"total_exec_time"`
	AverageExecTime time.Duration `json:"average_exec_time"`
	ErrorCount      int64         `json:"error_count"`

	// Resource usage
	MemoryUsed       int64 `json:"memory_used"`
	PeakMemoryUsed   int64 `json:"peak_memory_used"`
	GoroutinesActive int   `json:"goroutines_active"`
	BridgeCallsCount int64 `json:"bridge_calls_count"`

	// Performance counters
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	CompilationTime time.Duration `json:"compilation_time"`
	GCCollections   int64         `json:"gc_collections"`
}

// ScriptContext represents an isolated execution context within an engine.
type ScriptContext interface {
	ID() string
	SetVariable(name string, value interface{}) error
	GetVariable(name string) (interface{}, error)
	Execute(script string) (interface{}, error)
	Destroy() error
}

// MethodInfo describes a method exposed by a bridge.
type MethodInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  []ParameterInfo        `json:"parameters"`
	ReturnType  string                 `json:"return_type"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ParameterInfo describes a method parameter.
type ParameterInfo struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
}

// TypeMapping defines how to convert between Go and script types.
type TypeMapping struct {
	GoType     string                 `json:"go_type"`
	ScriptType string                 `json:"script_type"`
	Converter  string                 `json:"converter"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TypeInfo provides information about a supported type.
type TypeInfo struct {
	Name        string                 `json:"name"`
	Category    TypeCategory           `json:"category"`
	Description string                 `json:"description"`
	Methods     []string               `json:"methods"`
	Properties  []string               `json:"properties"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Function represents a callable function in the script engine.
type Function interface {
	Call(args ...interface{}) (interface{}, error)
	Bind(thisArg interface{}) Function
	GetSignature() FunctionSignature
}

// FunctionSignature describes a function's signature.
type FunctionSignature struct {
	Name       string          `json:"name"`
	Parameters []ParameterInfo `json:"parameters"`
	ReturnType string          `json:"return_type"`
	IsAsync    bool            `json:"is_async"`
	IsVariadic bool            `json:"is_variadic"`
}

// Permission represents a security permission required by a bridge.
type Permission struct {
	Type        PermissionType `json:"type"`
	Resource    string         `json:"resource"`
	Actions     []string       `json:"actions"`
	Description string         `json:"description"`
}

// Enums and constants

// EngineFeature represents a feature supported by an engine.
type EngineFeature string

const (
	FeatureAsync       EngineFeature = "async"
	FeatureCoroutines  EngineFeature = "coroutines"
	FeatureModules     EngineFeature = "modules"
	FeatureDebugging   EngineFeature = "debugging"
	FeatureHotReload   EngineFeature = "hot_reload"
	FeatureCompilation EngineFeature = "compilation"
	FeatureInteractive EngineFeature = "interactive"
	FeatureStreaming   EngineFeature = "streaming"
)

// FSMode defines filesystem access modes for scripts.
type FSMode string

const (
	FSModeReadOnly  FSMode = "readonly"
	FSModeReadWrite FSMode = "readwrite"
	FSModeNone      FSMode = "none"
	FSModeSandbox   FSMode = "sandbox"
)

// TypeCategory categorizes script types.
type TypeCategory string

const (
	TypeCategoryPrimitive TypeCategory = "primitive"
	TypeCategoryObject    TypeCategory = "object"
	TypeCategoryFunction  TypeCategory = "function"
	TypeCategoryArray     TypeCategory = "array"
	TypeCategoryMap       TypeCategory = "map"
	TypeCategoryCustom    TypeCategory = "custom"
)

// PermissionType defines the type of permission required.
type PermissionType string

const (
	PermissionFileSystem PermissionType = "filesystem"
	PermissionNetwork    PermissionType = "network"
	PermissionProcess    PermissionType = "process"
	PermissionMemory     PermissionType = "memory"
	PermissionTime       PermissionType = "time"
	PermissionCrypto     PermissionType = "crypto"
	PermissionStorage    PermissionType = "storage"
)

// Errors that engines can return

// EngineError represents an error from a script engine.
type EngineError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	ScriptLine int       `json:"script_line,omitempty"`
	ScriptCol  int       `json:"script_col,omitempty"`
	StackTrace []string  `json:"stack_trace,omitempty"`
	Cause      error     `json:"-"`
}

func (e *EngineError) Error() string {
	return e.Message
}

func (e *EngineError) Unwrap() error {
	return e.Cause
}

// ErrorType categorizes engine errors.
type ErrorType string

const (
	ErrorTypeSyntax     ErrorType = "syntax"
	ErrorTypeRuntime    ErrorType = "runtime"
	ErrorTypeType       ErrorType = "type"
	ErrorTypeResource   ErrorType = "resource"
	ErrorTypeSecurity   ErrorType = "security"
	ErrorTypeBridge     ErrorType = "bridge"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeMemory     ErrorType = "memory"
	ErrorTypePermission ErrorType = "permission"
)
