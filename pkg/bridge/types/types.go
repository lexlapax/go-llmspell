// ABOUTME: Type aliases for go-llms types used in bridge implementations
// ABOUTME: Only includes aliases needed for script engine bridging

// Package types provides type aliases and re-exports from go-llms.
// This package serves as the boundary between go-llms functionality and
// the script engines, ensuring clean separation of concerns and making
// go-llms types available to bridge implementations without direct imports.
package types

import (
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/docs"
	"github.com/lexlapax/go-llms/pkg/errors"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/outputs"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
	modelinfodomain "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"

	// Import engine types for bridge interface compliance
	// This import is needed for bridge implementations to access engine types
	// without creating circular dependencies
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Type aliases for go-llms types - we use these directly.
// These aliases allow bridge implementations to use go-llms types
// without importing the go-llms packages directly, maintaining
// architectural boundaries.
type (
	// Agent domain types from go-llms/pkg/agent/domain.
	// These types define the core agent system abstractions.
	State              = *domain.State
	Artifact           = *domain.Artifact
	Message            = domain.Message
	ArtifactType       = domain.ArtifactType
	SharedStateContext = *domain.SharedStateContext
	StateReader        = domain.StateReader
	MergeStrategy      = domain.MergeStrategy
	Tool               = domain.Tool
	AgentError         = *domain.AgentError
	ToolError          = *domain.ToolError
	BaseAgent          = domain.BaseAgent
	Agent              = domain.Agent
	AgentType          = domain.AgentType
	AgentConfig        = domain.AgentConfig
	LLMConfig          = domain.LLMConfig
	Event              = domain.Event
	EventType          = domain.EventType
	Hook               = domain.Hook
	ToolContext        = domain.ToolContext
	ToolExample        = domain.ToolExample
	MCPToolDefinition  = domain.MCPToolDefinition
	Handoff            = domain.Handoff
	RetryStrategy      = domain.RetryStrategy

	// Agent core types from go-llms/pkg/agent/core.
	// These provide concrete implementations of agent functionality.
	StateManager   = *core.StateManager
	StateTransform = core.StateTransform
	StateValidator = core.StateValidator
	AgentRegistry  = *core.AgentRegistry
	LLMAgent       = *core.LLMAgent

	// Workflow types from go-llms/pkg/agent/workflow.
	// These enable complex multi-step agent orchestration.
	WorkflowAgent          = workflow.WorkflowAgent
	WorkflowStep           = workflow.WorkflowStep
	WorkflowState          = *workflow.WorkflowState
	WorkflowStatus         = workflow.WorkflowStatus
	WorkflowDefinition     = *workflow.WorkflowDefinition
	WorkflowStateType      = workflow.WorkflowStateType
	StepStateType          = workflow.StepStateType
	StepStatus             = workflow.StepStatus
	ErrorAction            = workflow.ErrorAction
	ErrorHandler           = workflow.ErrorHandler
	DefaultErrorHandler    = *workflow.DefaultErrorHandler
	AgentStep              = *workflow.AgentStep
	BranchingWorkflowAgent = workflow.BranchingWorkflowAgent
	WorkflowOrchestrator   = workflow.WorkflowOrchestrator
	CoordinationStrategy   = workflow.CoordinationStrategy

	// LLM domain types from go-llms/pkg/llm/domain.
	// Core types for LLM provider interactions.
	Provider        = llmdomain.Provider
	ContentPart     = llmdomain.ContentPart
	Response        = llmdomain.Response
	ResponseStream  = llmdomain.ResponseStream
	Token           = llmdomain.Token
	ProviderOptions = llmdomain.ProviderOptions
	ModelRegistry   = llmdomain.ModelRegistry

	// Util types - auth from go-llms/pkg/util/auth.
	// Authentication and authorization types.
	AuthConfig = auth.AuthConfig
	AuthScheme = auth.AuthScheme

	// Util types - llmutil from go-llms/pkg/util/llmutil.
	// Utility types for LLM configuration and pooling.
	ModelConfig    = llmutil.ModelConfig
	ProviderPool   = *llmutil.ProviderPool
	ModelInventory = *modelinfodomain.ModelInventory

	// Tool discovery types from go-llms/pkg/agent/tools.
	// Types for dynamic tool discovery and registration.
	ToolDiscovery = tools.ToolDiscovery
	ToolInfo      = tools.ToolInfo
	ToolSchema    = tools.ToolSchema
	ToolFactory   = tools.ToolFactory

	// Schema types from go-llms/pkg/schema/domain.
	// JSON Schema validation and generation types.
	Schema           = schemaDomain.Schema
	Property         = schemaDomain.Property
	ValidationResult = schemaDomain.ValidationResult

	// Schema system types (v0.3.5) from go-llms/pkg/schema/domain.
	// Extended schema repository and generator functionality.
	SchemaRepository = schemaDomain.SchemaRepository
	SchemaGenerator  = schemaDomain.SchemaGenerator

	// Structured output types (v0.3.5) from go-llms/pkg/llm/outputs.
	// Parsers for structured LLM responses in various formats.
	OutputParser = outputs.Parser
	JSONParser   = *outputs.JSONParser
	XMLParser    = *outputs.XMLParser
	YAMLParser   = *outputs.YAMLParser

	// Event system types (v0.3.5) from go-llms/pkg/agent/events.
	// Event storage, filtering, and replay capabilities.
	EventStore      = events.EventStorage
	EventFilter     = events.EventFilter
	EventReplayer   = *events.EventReplayer
	EventSerializer = events.EventSerializer

	// Documentation types (v0.3.5) from go-llms/pkg/docs.
	// API documentation generation capabilities.
	DocGenerator     = docs.Generator
	OpenAPIGenerator = *docs.OpenAPIGenerator

	// Error types (v0.3.5) from go-llms/pkg/errors.
	// Enhanced error handling with serialization and recovery.
	SerializableError = errors.SerializableError
	ErrorRecovery     = errors.RecoveryStrategy

	// Engine interface types from go-llmspell/pkg/engine.
	// These types enable bridges to implement the engine.Bridge interface
	// without importing pkg/engine directly, preventing circular dependencies.
	ScriptEngine     = engine.ScriptEngine
	Bridge           = engine.Bridge
	BridgeMetadata   = engine.BridgeMetadata
	Registry         = engine.Registry
	TypeConverter    = engine.TypeConverter
	EngineConfig     = engine.EngineConfig
	ContextOptions   = engine.ContextOptions
	ExecutionOptions = engine.ExecutionOptions
	ExecutionResult  = engine.ExecutionResult
	ResourceLimits   = engine.ResourceLimits
	EngineMetrics    = engine.EngineMetrics
	ScriptContext    = engine.ScriptContext
	MethodInfo       = engine.MethodInfo
	ParameterInfo    = engine.ParameterInfo
	TypeMapping      = engine.TypeMapping
	TypeInfo         = engine.TypeInfo
	Function         = engine.Function
	FunctionSignature = engine.FunctionSignature
	Permission       = engine.Permission
	ScriptValue      = engine.ScriptValue
	EventBus         = engine.EventBus
	EventHandler     = engine.EventHandler
	EngineEvent      = engine.EngineEvent
	SubscriptionInfo = engine.SubscriptionInfo
	TypeRegistry     = engine.TypeRegistry
	TypeConverterFunc = engine.TypeConverterFunc
	ProfilingConfig  = engine.ProfilingConfig
	ProfilingReport  = engine.ProfilingReport
	MemoryStats      = engine.MemoryStats
	Hotspot          = engine.Hotspot
	OptimizationHint = engine.OptimizationHint
	ClientLibraryOptions = engine.ClientLibraryOptions
	EngineError      = engine.EngineError

	// ScriptValue types from go-llmspell/pkg/engine.
	// These enable bridges to work with script values without importing engine directly.
	ScriptValueType = engine.ScriptValueType
	NilValue        = engine.NilValue
	BoolValue       = engine.BoolValue
	NumberValue     = engine.NumberValue
	StringValue     = engine.StringValue
	ArrayValue      = engine.ArrayValue
	ObjectValue     = engine.ObjectValue
	FunctionValue   = engine.FunctionValue
	ErrorValue      = engine.ErrorValue
	ChannelValue    = engine.ChannelValue
	CustomValue     = engine.CustomValue
)

// Re-export constants from go-llms.
// These constants define enumerated values used throughout
// the bridge implementations for type safety.
const (
	// Artifact types define the kinds of data agents can produce.
	// Used for type-safe artifact handling in workflows.
	ArtifactTypeData     = domain.ArtifactTypeData
	ArtifactTypeImage    = domain.ArtifactTypeImage
	ArtifactTypeDocument = domain.ArtifactTypeDocument
	ArtifactTypeCode     = domain.ArtifactTypeCode

	// Merge strategies control how state updates are combined.
	// Essential for managing concurrent agent state changes.
	MergeStrategyLast     = domain.MergeStrategyLast
	MergeStrategyMergeAll = domain.MergeStrategyMergeAll
	MergeStrategyUnion    = domain.MergeStrategyUnion

	// Agent types enumerate the different agent implementations.
	// Each type has specific execution semantics.
	AgentTypeLLM         = domain.AgentTypeLLM
	AgentTypeSequential  = domain.AgentTypeSequential
	AgentTypeParallel    = domain.AgentTypeParallel
	AgentTypeConditional = domain.AgentTypeConditional
	AgentTypeLoop        = domain.AgentTypeLoop
	AgentTypeCustom      = domain.AgentTypeCustom

	// Event types - Lifecycle events track agent execution phases.
	// Used for monitoring and debugging agent behavior.
	EventAgentStart    = domain.EventAgentStart
	EventAgentComplete = domain.EventAgentComplete
	EventAgentError    = domain.EventAgentError

	// Event types - Execution events track runtime activities.
	// Provide visibility into agent internal operations.
	EventStateUpdate = domain.EventStateUpdate
	EventProgress    = domain.EventProgress
	EventMessage     = domain.EventMessage

	// Event types - Tool events track tool invocations.
	// Critical for debugging tool usage and performance.
	EventToolCall   = domain.EventToolCall
	EventToolResult = domain.EventToolResult
	EventToolError  = domain.EventToolError

	// Event types - Workflow events track multi-agent coordination.
	// Essential for understanding complex workflow execution.
	EventSubAgentStart = domain.EventSubAgentStart
	EventSubAgentEnd   = domain.EventSubAgentEnd
	EventWorkflowStep  = domain.EventWorkflowStep
	EventWorkflowStart = domain.EventWorkflowStart

	// Workflow states represent the lifecycle of a workflow.
	// Used to track and control workflow execution.
	WorkflowStatePending   = workflow.WorkflowStatePending
	WorkflowStateRunning   = workflow.WorkflowStateRunning
	WorkflowStatePaused    = workflow.WorkflowStatePaused
	WorkflowStateCompleted = workflow.WorkflowStateCompleted
	WorkflowStateFailed    = workflow.WorkflowStateFailed
	WorkflowStateCanceled  = workflow.WorkflowStateCanceled

	// Step states represent individual step status within workflows.
	// Enable fine-grained workflow progress tracking.
	StepStatePending   = workflow.StepStatePending
	StepStateRunning   = workflow.StepStateRunning
	StepStateCompleted = workflow.StepStateCompleted
	StepStateFailed    = workflow.StepStateFailed
	StepStateSkipped   = workflow.StepStateSkipped

	// Error actions define how workflows handle step failures.
	// Critical for building resilient agent workflows.
	ErrorActionRetry    = workflow.ErrorActionRetry
	ErrorActionSkip     = workflow.ErrorActionSkip
	ErrorActionAbort    = workflow.ErrorActionAbort
	ErrorActionContinue = workflow.ErrorActionContinue

	// Engine feature constants from go-llmspell/pkg/engine.
	// These define capabilities supported by script engines.
	FeatureAsync       = engine.FeatureAsync
	FeatureCoroutines  = engine.FeatureCoroutines
	FeatureModules     = engine.FeatureModules
	FeatureDebugging   = engine.FeatureDebugging
	FeatureHotReload   = engine.FeatureHotReload
	FeatureCompilation = engine.FeatureCompilation
	FeatureInteractive = engine.FeatureInteractive
	FeatureStreaming   = engine.FeatureStreaming

	// Filesystem access modes for script security.
	FSModeReadOnly  = engine.FSModeReadOnly
	FSModeReadWrite = engine.FSModeReadWrite
	FSModeNone      = engine.FSModeNone
	FSModeSandbox   = engine.FSModeSandbox

	// Type categories for script type system.
	TypeCategoryPrimitive = engine.TypeCategoryPrimitive
	TypeCategoryObject    = engine.TypeCategoryObject
	TypeCategoryFunction  = engine.TypeCategoryFunction
	TypeCategoryArray     = engine.TypeCategoryArray
	TypeCategoryMap       = engine.TypeCategoryMap
	TypeCategoryCustom    = engine.TypeCategoryCustom

	// Permission types for bridge security.
	PermissionFileSystem = engine.PermissionFileSystem
	PermissionNetwork    = engine.PermissionNetwork
	PermissionProcess    = engine.PermissionProcess
	PermissionMemory     = engine.PermissionMemory
	PermissionTime       = engine.PermissionTime
	PermissionCrypto     = engine.PermissionCrypto
	PermissionStorage    = engine.PermissionStorage

	// Export formats for API documentation.
	ExportFormatOpenAPI  = engine.ExportFormatOpenAPI
	ExportFormatMarkdown = engine.ExportFormatMarkdown
	ExportFormatJSON     = engine.ExportFormatJSON
	ExportFormatGraphQL  = engine.ExportFormatGraphQL
	ExportFormatProtobuf = engine.ExportFormatProtobuf

	// Engine error types for error categorization.
	ErrorTypeSyntax     = engine.ErrorTypeSyntax
	ErrorTypeRuntime    = engine.ErrorTypeRuntime
	ErrorTypeType       = engine.ErrorTypeType
	ErrorTypeResource   = engine.ErrorTypeResource
	ErrorTypeSecurity   = engine.ErrorTypeSecurity
	ErrorTypeBridge     = engine.ErrorTypeBridge
	ErrorTypeTimeout    = engine.ErrorTypeTimeout
	ErrorTypeMemory     = engine.ErrorTypeMemory
	ErrorTypePermission = engine.ErrorTypePermission

	// ScriptValue type constants for type checking.
	TypeNil      = engine.TypeNil
	TypeBool     = engine.TypeBool
	TypeNumber   = engine.TypeNumber
	TypeString   = engine.TypeString
	TypeArray    = engine.TypeArray
	TypeObject   = engine.TypeObject
	TypeFunction = engine.TypeFunction
	TypeError    = engine.TypeError
	TypeChannel  = engine.TypeChannel
	TypeCustom   = engine.TypeCustom
)

// Re-export ScriptValue constructor functions.
// These functions enable bridges to create ScriptValues without importing engine directly.
var (
	NewNilValue      = engine.NewNilValue
	NewBoolValue     = engine.NewBoolValue
	NewNumberValue   = engine.NewNumberValue
	NewStringValue   = engine.NewStringValue
	NewArrayValue    = engine.NewArrayValue
	NewObjectValue   = engine.NewObjectValue
	NewFunctionValue = engine.NewFunctionValue
	NewErrorValue    = engine.NewErrorValue
	NewChannelValue  = engine.NewChannelValue
	NewCustomValue   = engine.NewCustomValue

	// Conversion functions
	ConvertToScriptValue      = engine.ConvertToScriptValue
	ConvertMapToScriptValue   = engine.ConvertMapToScriptValue
	ConvertSliceToScriptValue = engine.ConvertSliceToScriptValue
	IsTrue                    = engine.IsTrue
	ConvertToString           = engine.ConvertToString
	ConvertToNumber           = engine.ConvertToNumber
	ConvertToBool             = engine.ConvertToBool
)
