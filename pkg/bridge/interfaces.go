// ABOUTME: Type aliases for go-llms types used in bridge implementations
// ABOUTME: Only includes aliases needed for script engine bridging

// Package bridge provides type aliases and re-exports from go-llms.
// This package serves as the boundary between go-llms functionality and
// the script engines, ensuring clean separation of concerns and making
// go-llms types available to bridge implementations without direct imports.
package bridge

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
)
