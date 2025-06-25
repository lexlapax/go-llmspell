// ABOUTME: Factory for creating bridge adapters based on bridge IDs with support for multi-bridge dependencies
// ABOUTME: Ensures proper adapter creation for the execution path: Lua → Adapter → Bridge → go-llms

package factory

import (
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/adapters/impl"
)

// AdapterFactory creates appropriate adapters based on bridge IDs
type AdapterFactory struct {
	mu sync.RWMutex
}

// NewAdapterFactory creates a new adapter factory
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{}
}

// CreateAdapter creates the appropriate adapter for a bridge ID
// It handles complex multi-bridge dependencies by looking up related bridges in the bridgeMap
func (f *AdapterFactory) CreateAdapter(bridgeID string, bridgeMap map[string]engine.Bridge) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Get the primary bridge
	bridge, ok := bridgeMap[bridgeID]
	if !ok {
		return nil, fmt.Errorf("bridge not found for ID: %s", bridgeID)
	}

	// Create adapter based on bridge ID
	switch bridgeID {
	// LLM adapters (requires multiple bridges)
	case "llm_core", "llm_providers", "llm_pool":
		// LLM adapter needs all three bridges
		var coreBridge, providersBridge, poolBridge engine.Bridge
		
		// Determine which bridge we have and look up the others
		switch bridgeID {
		case "llm_core":
			coreBridge = bridge
			providersBridge = bridgeMap["llm_providers"] // May be nil
			poolBridge = bridgeMap["llm_pool"]           // May be nil
		case "llm_providers":
			coreBridge = bridgeMap["llm_core"]
			if coreBridge == nil {
				return nil, fmt.Errorf("llm_core bridge required for llm_providers")
			}
			providersBridge = bridge
			poolBridge = bridgeMap["llm_pool"] // May be nil
		case "llm_pool":
			coreBridge = bridgeMap["llm_core"]
			if coreBridge == nil {
				return nil, fmt.Errorf("llm_core bridge required for llm_pool")
			}
			providersBridge = bridgeMap["llm_providers"] // May be nil
			poolBridge = bridge
		}
		
		return impl.NewLLMAdapter(coreBridge, providersBridge, poolBridge), nil

	// State management
	case "state_manager":
		return impl.NewStateAdapter(bridge), nil

	// Event system
	case "agent_events":
		return impl.NewEventsAdapter(bridge), nil

	// Structured data
	case "structured_schema":
		return impl.NewStructuredAdapter(bridge), nil

	// Agent system
	case "agent_core":
		return impl.NewAgentAdapter(bridge), nil

	// Hooks system
	case "agent_hooks":
		return impl.NewHooksAdapter(bridge), nil

	// Workflow
	case "agent_workflow":
		return impl.NewWorkflowAdapter(bridge), nil

	// Tools
	case "agent_tools":
		// Check if registry bridge is available for enhanced functionality
		registryBridge := bridgeMap["tools_registry"] // May be nil
		if registryBridge != nil {
			return impl.NewToolsAdapterWithRegistry(bridge, registryBridge), nil
		}
		return impl.NewToolsAdapter(bridge), nil

	// Observability (requires multiple bridges)
	case "observability_metrics":
		// Observability adapter needs tracing and guardrails bridges
		tracingBridge := bridgeMap["observability_tracing"]      // May be nil
		guardrailsBridge := bridgeMap["observability_guardrails"] // May be nil
		return impl.NewObservabilityAdapter(bridge, tracingBridge, guardrailsBridge), nil

	// Model info
	case "llm_modelinfo":
		return impl.NewModelInfoAdapter(bridge), nil

	// Utils (requires multiple bridges)
	case "util_core", "auth", "util_debug", "util_errors", "util_json", "llm_utils", "util_script_logger", "util_slog":
		// Utils adapter needs all utility bridges
		// Note: The primary bridge is already validated above, so we pass the full map
		return f.createUtilsAdapter(bridgeMap)

	default:
		return nil, fmt.Errorf("unknown bridge ID: %s", bridgeID)
	}
}

// createUtilsAdapter creates a utils adapter with all available utility bridges
func (f *AdapterFactory) createUtilsAdapter(bridgeMap map[string]engine.Bridge) (interface{}, error) {
	// Utils adapter can work with partial bridges (graceful degradation)
	authBridge := bridgeMap["auth"]
	debugBridge := bridgeMap["util_debug"]
	errorsBridge := bridgeMap["util_errors"]
	jsonBridge := bridgeMap["util_json"]
	llmUtilsBridge := bridgeMap["llm_utils"]
	loggerBridge := bridgeMap["util_script_logger"]
	slogBridge := bridgeMap["util_slog"]
	utilBridge := bridgeMap["util_core"]

	// At least one utility bridge must be present
	if authBridge == nil && debugBridge == nil && errorsBridge == nil && 
	   jsonBridge == nil && llmUtilsBridge == nil && loggerBridge == nil && 
	   slogBridge == nil && utilBridge == nil {
		return nil, fmt.Errorf("no utility bridges available for utils adapter")
	}

	return impl.NewUtilsAdapter(
		authBridge,
		debugBridge,
		errorsBridge,
		jsonBridge,
		llmUtilsBridge,
		loggerBridge,
		slogBridge,
		utilBridge,
	), nil
}

// GetSupportedBridgeIDs returns all bridge IDs supported by the factory
func (f *AdapterFactory) GetSupportedBridgeIDs() []string {
	return []string{
		// Core LLM
		"llm_core",
		"llm_providers",
		"llm_pool",
		"llm_modelinfo",
		"llm_utils",

		// Agent system
		"agent_core",
		"agent_events",
		"agent_hooks",
		"agent_workflow",
		"agent_tools",

		// Data management
		"state_manager",
		"structured_schema",

		// Tools
		"tools_registry",

		// Observability
		"observability_metrics",
		"observability_tracing",
		"observability_guardrails",

		// Utilities
		"auth",
		"util_core",
		"util_debug",
		"util_errors",
		"util_json",
		"util_script_logger",
		"util_slog",
	}
}

// IsMultiBridgeAdapter returns true if the adapter requires multiple bridges
func (f *AdapterFactory) IsMultiBridgeAdapter(bridgeID string) bool {
	switch bridgeID {
	case "llm_core", "llm_providers", "llm_pool", "observability_metrics":
		return true
	case "util_core", "auth", "util_debug", "util_errors", "util_json", "llm_utils", "util_script_logger", "util_slog":
		return true // Any util bridge triggers multi-bridge utils adapter
	default:
		return false
	}
}

// GetRequiredBridges returns the required bridge IDs for a multi-bridge adapter
func (f *AdapterFactory) GetRequiredBridges(bridgeID string) []string {
	switch bridgeID {
	case "llm_core":
		return []string{"llm_core"} // providers and pool are optional
	case "observability_metrics":
		return []string{"observability_metrics"} // tracing and guardrails are optional
	case "util_core", "auth", "util_debug", "util_errors", "util_json", "llm_utils", "util_script_logger", "util_slog":
		// At least one utility bridge is required
		return []string{} // No specific bridge required, just at least one
	default:
		return []string{bridgeID}
	}
}

// GetOptionalBridges returns the optional bridge IDs that enhance an adapter
func (f *AdapterFactory) GetOptionalBridges(bridgeID string) []string {
	switch bridgeID {
	case "llm_core":
		return []string{"llm_providers", "llm_pool"}
	case "observability_metrics":
		return []string{"observability_tracing", "observability_guardrails"}
	case "agent_tools":
		return []string{"tools_registry"}
	case "util_core", "auth", "util_debug", "util_errors", "util_json", "llm_utils", "util_script_logger", "util_slog":
		// All utility bridges are optional as long as at least one is present
		return []string{
			"auth", "util_core", "util_debug", "util_errors",
			"util_json", "llm_utils", "util_script_logger", "util_slog",
		}
	default:
		return []string{}
	}
}