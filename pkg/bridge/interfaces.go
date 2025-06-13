// ABOUTME: Type aliases for go-llms types used in bridge implementations
// ABOUTME: Only includes aliases needed for script engine bridging

package bridge

import (
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
	modelinfodomain "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// Type aliases for go-llms types - we use these directly
type (
	// Agent domain types
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

	// Agent core types
	StateManager   = *core.StateManager
	StateTransform = core.StateTransform
	StateValidator = core.StateValidator
	AgentRegistry  = *core.AgentRegistry
	LLMAgent       = *core.LLMAgent

	// LLM domain types
	Provider        = llmdomain.Provider
	ContentPart     = llmdomain.ContentPart
	Response        = llmdomain.Response
	ResponseStream  = llmdomain.ResponseStream
	Token           = llmdomain.Token
	ProviderOptions = llmdomain.ProviderOptions
	ModelRegistry   = llmdomain.ModelRegistry

	// Util types - auth
	AuthConfig = auth.AuthConfig
	AuthScheme = auth.AuthScheme

	// Util types - llmutil
	ModelConfig    = llmutil.ModelConfig
	ProviderPool   = *llmutil.ProviderPool
	ModelInventory = *modelinfodomain.ModelInventory
)

// Re-export constants from go-llms
const (
	// Artifact types
	ArtifactTypeData     = domain.ArtifactTypeData
	ArtifactTypeImage    = domain.ArtifactTypeImage
	ArtifactTypeDocument = domain.ArtifactTypeDocument
	ArtifactTypeCode     = domain.ArtifactTypeCode

	// Merge strategies
	MergeStrategyLast     = domain.MergeStrategyLast
	MergeStrategyMergeAll = domain.MergeStrategyMergeAll
	MergeStrategyUnion    = domain.MergeStrategyUnion

	// Agent types
	AgentTypeLLM         = domain.AgentTypeLLM
	AgentTypeSequential  = domain.AgentTypeSequential
	AgentTypeParallel    = domain.AgentTypeParallel
	AgentTypeConditional = domain.AgentTypeConditional
	AgentTypeLoop        = domain.AgentTypeLoop
	AgentTypeCustom      = domain.AgentTypeCustom

	// Event types - Lifecycle
	EventAgentStart    = domain.EventAgentStart
	EventAgentComplete = domain.EventAgentComplete
	EventAgentError    = domain.EventAgentError

	// Event types - Execution
	EventStateUpdate = domain.EventStateUpdate
	EventProgress    = domain.EventProgress
	EventMessage     = domain.EventMessage

	// Event types - Tool
	EventToolCall   = domain.EventToolCall
	EventToolResult = domain.EventToolResult
	EventToolError  = domain.EventToolError

	// Event types - Workflow
	EventSubAgentStart = domain.EventSubAgentStart
	EventSubAgentEnd   = domain.EventSubAgentEnd
	EventWorkflowStep  = domain.EventWorkflowStep
	EventWorkflowStart = domain.EventWorkflowStart
)
