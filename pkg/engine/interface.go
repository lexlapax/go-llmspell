// ABOUTME: This file defines the core interfaces for the multi-engine scripting architecture.
// ABOUTME: It provides engine-agnostic abstractions for script execution, bridging, and type conversion.

// Package engine provides the core interfaces and types for the multi-engine scripting architecture.
// It defines engine-agnostic abstractions that allow go-llmspell to support multiple scripting
// languages (Lua, JavaScript, Tengo) through a unified API.
//
// The package includes:
//   - ScriptEngine interface for engine implementations
//   - Bridge interface for exposing Go functionality to scripts
//   - TypeConverter interface for handling type conversions
//   - Common types and constants used across all engines
//
// Example usage:
//
//	engine := lua.NewEngine()
//	err := engine.Initialize(engine.EngineConfig{
//	    SandboxMode: true,
//	    MemoryLimit: 100 * 1024 * 1024, // 100MB
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer engine.Shutdown()
//
//	result, err := engine.Execute(ctx, "return 2 + 2", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Output: 4
package engine

import (
	"context"
	"time"
)

// ScriptEngine defines the common interface that all scripting engines must implement.
// This abstraction allows go-llmspell to support multiple scripting languages
// (Lua, JavaScript, Tengo) through a unified API.
type ScriptEngine interface {
	// Initialize prepares the engine for use with the given configuration.
	// This must be called before any script execution.
	Initialize(config EngineConfig) error
	
	// Execute runs a script string with optional parameters and returns the result.
	// The params map is made available to the script as global variables.
	Execute(ctx context.Context, script string, params map[string]interface{}) (ScriptValue, error)
	
	// ExecuteFile loads and executes a script from a file path.
	// The params map is made available to the script as global variables.
	ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (ScriptValue, error)
	
	// Shutdown cleanly shuts down the engine and releases all resources.
	// After calling Shutdown, the engine cannot be used again.
	Shutdown() error

	// RegisterBridge registers a bridge with the engine, making its methods available to scripts.
	RegisterBridge(bridge Bridge) error
	
	// UnregisterBridge removes a previously registered bridge by name.
	UnregisterBridge(name string) error
	
	// GetBridge retrieves a registered bridge by name.
	GetBridge(name string) (Bridge, error)
	
	// ListBridges returns the names of all registered bridges.
	ListBridges() []string

	// ToNative converts a script value to a native Go value.
	// This is used when retrieving values from scripts.
	ToNative(scriptValue ScriptValue) (interface{}, error)
	
	// FromNative converts a native Go value to a script value.
	// This is used when passing values to scripts.
	FromNative(goValue interface{}) (ScriptValue, error)

	// Name returns the name of the scripting engine (e.g., "lua", "javascript", "tengo").
	Name() string
	
	// Version returns the version of the scripting engine implementation.
	Version() string
	
	// FileExtensions returns the file extensions this engine supports (e.g., [".lua"]).
	FileExtensions() []string
	
	// Features returns the list of features supported by this engine.
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

	// Task 1.4.11.1: Engine Event Bus
	GetEventBus() EventBus

	// Task 1.4.11.2: Type Conversion Registry
	RegisterTypeConverter(fromType, toType string, converter TypeConverterFunc) error
	GetTypeRegistry() TypeRegistry

	// Task 1.4.11.3: Engine Profiling
	EnableProfiling(config ProfilingConfig) error
	DisableProfiling() error
	GetProfilingReport() (*ProfilingReport, error)

	// Task 1.4.11.4: Engine API Export
	ExportAPI(format ExportFormat) ([]byte, error)
	GenerateClientLibrary(language string, options ClientLibraryOptions) ([]byte, error)
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
	ValidateMethod(name string, args []ScriptValue) error
	ExecuteMethod(ctx context.Context, name string, args []ScriptValue) (ScriptValue, error)

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
	// Basic type conversions from ScriptValue
	ToBoolean(v ScriptValue) (bool, error)
	ToNumber(v ScriptValue) (float64, error)
	ToString(v ScriptValue) (string, error)
	ToArray(v ScriptValue) ([]ScriptValue, error)
	ToMap(v ScriptValue) (map[string]ScriptValue, error)

	// Complex type handling
	ToStruct(v ScriptValue, target interface{}) error
	FromStruct(v interface{}) (ScriptValue, error)

	// Function and callback handling
	ToFunction(v ScriptValue) (Function, error)
	FromFunction(fn Function) (ScriptValue, error)

	// ScriptValue creation from Go types
	FromInterface(v interface{}) (ScriptValue, error)
	ToInterface(v ScriptValue) (interface{}, error)

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
	Value    ScriptValue            `json:"value"`
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
	Call(args ...ScriptValue) (ScriptValue, error)
	Bind(thisArg ScriptValue) Function
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

// Task 1.4.11.1: EventBus interface for engine-level events
type EventBus interface {
	// Subscribe to events with optional filters
	Subscribe(pattern string, handler EventHandler) (string, error)
	Unsubscribe(subscriptionID string) error

	// Publish events
	Publish(event EngineEvent) error
	PublishAsync(event EngineEvent) error

	// Event routing and management
	SetPriority(subscriptionID string, priority int) error
	GetSubscriptions() []SubscriptionInfo
	Clear() error
}

// EventHandler for engine events
type EventHandler func(event EngineEvent) error

// EngineEvent represents an event in the engine
type EngineEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SubscriptionInfo contains information about an event subscription
type SubscriptionInfo struct {
	ID       string    `json:"id"`
	Pattern  string    `json:"pattern"`
	Priority int       `json:"priority"`
	Created  time.Time `json:"created"`
}

// Task 1.4.11.2: Type conversion registry
type TypeRegistry interface {
	// Register converters
	Register(fromType, toType string, converter TypeConverterFunc) error
	RegisterBidirectional(type1, type2 string, forward, reverse TypeConverterFunc) error

	// Conversion operations
	Convert(value interface{}, fromType, toType string) (interface{}, error)
	CanConvert(fromType, toType string) bool

	// Registry management
	GetConverters() map[string][]string
	ClearCache() error
	ExportDocumentation() ([]byte, error)
}

// TypeConverterFunc converts values between types
type TypeConverterFunc func(value interface{}) (interface{}, error)

// Task 1.4.11.3: Profiling support
type ProfilingConfig struct {
	Enabled        bool          `json:"enabled"`
	CPUProfiling   bool          `json:"cpu_profiling"`
	MemProfiling   bool          `json:"mem_profiling"`
	TraceProfiling bool          `json:"trace_profiling"`
	SampleRate     int           `json:"sample_rate"`
	OutputDir      string        `json:"output_dir"`
	Duration       time.Duration `json:"duration"`
}

// ProfilingReport contains profiling results
type ProfilingReport struct {
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Executions      int64                  `json:"executions"`
	AverageExecTime time.Duration          `json:"average_exec_time"`
	MemoryStats     MemoryStats            `json:"memory_stats"`
	Hotspots        []Hotspot              `json:"hotspots"`
	Optimizations   []OptimizationHint     `json:"optimizations"`
	Metrics         map[string]interface{} `json:"metrics"`
}

// MemoryStats contains memory usage statistics
type MemoryStats struct {
	Allocated      uint64 `json:"allocated"`
	TotalAllocated uint64 `json:"total_allocated"`
	Sys            uint64 `json:"sys"`
	NumGC          uint32 `json:"num_gc"`
	PauseTotal     uint64 `json:"pause_total"`
}

// Hotspot represents a performance hotspot
type Hotspot struct {
	Location    string        `json:"location"`
	Count       int64         `json:"count"`
	TotalTime   time.Duration `json:"total_time"`
	AverageTime time.Duration `json:"average_time"`
	Percentage  float64       `json:"percentage"`
}

// OptimizationHint suggests performance improvements
type OptimizationHint struct {
	Type        string `json:"type"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Priority    int    `json:"priority"`
}

// Task 1.4.11.4: API Export formats
type ExportFormat string

const (
	ExportFormatOpenAPI  ExportFormat = "openapi"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatGraphQL  ExportFormat = "graphql"
	ExportFormatProtobuf ExportFormat = "protobuf"
)

// ClientLibraryOptions for generating client libraries
type ClientLibraryOptions struct {
	PackageName  string                 `json:"package_name"`
	Version      string                 `json:"version"`
	IncludeTypes bool                   `json:"include_types"`
	IncludeDocs  bool                   `json:"include_docs"`
	CustomConfig map[string]interface{} `json:"custom_config"`
}

// Errors that engines can return

// EngineError represents an error from a script engine.
// It includes detailed information about the error location and type
// to help with debugging script execution issues.
type EngineError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	ScriptLine int       `json:"script_line,omitempty"`
	ScriptCol  int       `json:"script_col,omitempty"`
	StackTrace []string  `json:"stack_trace,omitempty"`
	Cause      error     `json:"-"`
}

// Error implements the error interface, returning the error message.
func (e *EngineError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error, if any.
// This allows the error to work with errors.Is and errors.As.
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
