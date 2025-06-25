// ABOUTME: Bridge registry for modular bridge registration across different script engines
// ABOUTME: Provides configurable bridge sets and engine-agnostic registration patterns

package registry

import (
	"fmt"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	"github.com/lexlapax/go-llmspell/pkg/bridge/llm"
	"github.com/lexlapax/go-llmspell/pkg/bridge/observability"
	"github.com/lexlapax/go-llmspell/pkg/bridge/state"
	"github.com/lexlapax/go-llmspell/pkg/bridge/structured"
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
	"github.com/lexlapax/go-llmspell/pkg/bridge/util"
)

// BridgeSet represents a collection of related bridges
type BridgeSet string

const (
	BridgeSetCore          BridgeSet = "core"          // Essential bridges needed by all engines
	BridgeSetLLM           BridgeSet = "llm"           // LLM-related bridges
	BridgeSetUtility       BridgeSet = "utility"       // Utility and helper bridges
	BridgeSetAgent         BridgeSet = "agent"         // Agent and workflow bridges
	BridgeSetObservability BridgeSet = "observability" // Monitoring and debugging bridges
	BridgeSetState         BridgeSet = "state"         // State management bridges
	BridgeSetStructured    BridgeSet = "structured"    // Schema and validation bridges
	BridgeSetAll           BridgeSet = "all"           // All available bridges
)

// RegisterBridgesByFeatureSet registers bridges based on a feature set.
// This replaces the old profile-based system with a more flexible approach.
func RegisterBridgesByFeatureSet(scriptEngine types.ScriptEngine, featureSet FeatureSet) error {
	bridgeSets := GetBridgeSetsForFeature(featureSet)
	return RegisterBridgeSets(scriptEngine, bridgeSets)
}

// BridgeFactory creates bridges for a specific bridge set
type BridgeFactory func() ([]types.Bridge, error)

// bridgeFactories maps bridge sets to their factory functions
var bridgeFactories = map[BridgeSet]BridgeFactory{
	BridgeSetCore:          createCoreBridges,
	BridgeSetLLM:           createLLMBridges,
	BridgeSetUtility:       createUtilityBridges,
	BridgeSetAgent:         createAgentBridges,
	BridgeSetObservability: createObservabilityBridges,
	BridgeSetState:         createStateBridges,
	BridgeSetStructured:    createStructuredBridges,
}

// createCoreBridges creates essential bridges needed by all engines
func createCoreBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	// Model info bridge (essential for all engines)
	modelInfoBridge := bridge.NewModelInfoBridge()
	bridges = append(bridges, modelInfoBridge)

	return bridges, nil
}

// createLLMBridges creates LLM-related bridges
func createLLMBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	// Create core LLM bridge (required by other LLM bridges)
	llmBridge := llm.NewLLMBridge()
	bridges = append(bridges, llmBridge)

	// Create dependent LLM bridges
	providersBridge := llm.NewProvidersBridge(llmBridge)
	poolBridge := llm.NewPoolBridge(llmBridge)
	bridges = append(bridges, providersBridge, poolBridge)

	return bridges, nil
}

// createUtilityBridges creates utility and helper bridges
func createUtilityBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	utilBridge := util.NewUtilBridge()
	utilAuthBridge := util.NewUtilAuthBridge()
	utilDebugBridge := util.NewDebugBridge()
	utilErrorsBridge := util.NewUtilErrorsBridge()
	utilJSONBridge := util.NewUtilJSONBridge()
	utilLLMBridge := util.NewUtilLLMBridge()
	slogBridge := util.NewSlogBridge()
	scriptLoggerBridge := util.NewScriptLoggerBridge()

	bridges = append(bridges, utilBridge, utilAuthBridge, utilDebugBridge,
		utilErrorsBridge, utilJSONBridge, utilLLMBridge, slogBridge, scriptLoggerBridge)

	return bridges, nil
}

// createAgentBridges creates agent and workflow bridges
func createAgentBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	agentBridge := agent.NewAgentBridge()
	eventsBridge := agent.NewEventBridge()
	hooksBridge := agent.NewHooksBridge()
	toolsBridge := agent.NewToolsBridge()
	toolsRegistryBridge := agent.NewToolsRegistryBridge()
	workflowBridge := agent.NewWorkflowBridge()

	bridges = append(bridges, agentBridge, eventsBridge, hooksBridge,
		toolsBridge, toolsRegistryBridge, workflowBridge)

	return bridges, nil
}

// createObservabilityBridges creates monitoring and debugging bridges
func createObservabilityBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	metricsBridge := observability.NewMetricsBridge()
	tracingBridge := observability.NewTracingBridge()
	guardrailsBridge := observability.NewGuardrailsBridge()

	bridges = append(bridges, metricsBridge, tracingBridge, guardrailsBridge)

	return bridges, nil
}

// createStateBridges creates state management bridges
func createStateBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	stateContextBridge, err := state.NewStateContextBridge()
	if err != nil {
		return nil, fmt.Errorf("failed to create state context bridge: %w", err)
	}
	bridges = append(bridges, stateContextBridge)

	// Note: StateManagerBridge requires a types.StateManager instance
	// which would typically come from go-llms. Since we don't have a 
	// default implementation available, scripts that need state_manager
	// will need to create their own or use state_context instead.

	return bridges, nil
}

// createStructuredBridges creates schema and validation bridges
func createStructuredBridges() ([]types.Bridge, error) {
	var bridges []types.Bridge

	schemaBridge := structured.NewSchemaBridge()
	bridges = append(bridges, schemaBridge)

	return bridges, nil
}

// RegisterBridgeSets registers specific bridge sets with an engine
func RegisterBridgeSets(scriptEngine types.ScriptEngine, bridgeSets []BridgeSet) error {
	for _, bridgeSet := range bridgeSets {
		if bridgeSet == BridgeSetAll {
			// Register all available bridge sets
			allSets := []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetAgent, BridgeSetObservability, BridgeSetState, BridgeSetStructured}
			return RegisterBridgeSets(scriptEngine, allSets)
		}

		factory, exists := bridgeFactories[bridgeSet]
		if !exists {
			return fmt.Errorf("unknown bridge set: %s", bridgeSet)
		}

		bridges, err := factory()
		if err != nil {
			return fmt.Errorf("failed to create bridges for set %s: %w", bridgeSet, err)
		}

		for _, bridge := range bridges {
			if err := scriptEngine.RegisterBridge(bridge); err != nil {
				return fmt.Errorf("failed to register bridge %s from set %s: %w", bridge.GetID(), bridgeSet, err)
			}
		}
	}

	return nil
}

// RegisterStandardBridges registers all standard bridges (backward compatibility)
// This ensures that bridges are available to scripts via the global bridges table.
func RegisterStandardBridges(scriptEngine types.ScriptEngine) error {
	// Use full feature set as the standard
	return RegisterBridgesByFeatureSet(scriptEngine, FeatureSetFull)
}

// RegisterBridgesWithRegistry registers bridges with an engine obtained from the registry using a feature set
func RegisterBridgesWithRegistry(registry *types.Registry, engineName string, config types.EngineConfig, featureSet FeatureSet) error {
	// Get engine instance
	scriptEngine, err := registry.GetEngine(engineName, config)
	if err != nil {
		return fmt.Errorf("failed to get engine %s: %w", engineName, err)
	}

	// Register bridges using feature set
	return RegisterBridgesByFeatureSet(scriptEngine, featureSet)
}

// RegisterStandardBridgesWithRegistry registers standard bridges with an engine obtained from the registry.
// This is a convenience function for backward compatibility.
func RegisterStandardBridgesWithRegistry(registry *types.Registry, engineName string, config types.EngineConfig) error {
	return RegisterBridgesWithRegistry(registry, engineName, config, FeatureSetFull)
}

// RegisterBridgesWithAllEngines registers bridges with all engines in the registry using engine-specific feature sets
func RegisterBridgesWithAllEngines(registry *types.Registry, engineFeatureSets map[string]FeatureSet) error {
	engines := registry.ListEngines()
	for _, engineInfo := range engines {
		featureSet, exists := engineFeatureSets[engineInfo.Name]
		if !exists {
			// Use full feature set as default
			featureSet = FeatureSetFull
		}

		// Create default config for bridge registration
		config := types.EngineConfig{
			SandboxMode: true,
			DebugMode:   false,
		}

		if err := RegisterBridgesWithRegistry(registry, engineInfo.Name, config, featureSet); err != nil {
			return fmt.Errorf("failed to register bridges with engine %s: %w", engineInfo.Name, err)
		}
	}

	return nil
}
