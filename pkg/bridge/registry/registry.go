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
	"github.com/lexlapax/go-llmspell/pkg/bridge/util"
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
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

// BridgeProfile defines which bridge sets should be registered for an engine type
type BridgeProfile struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	BridgeSets  []BridgeSet `json:"bridge_sets"`
}

// Standard bridge profiles for different use cases
var (
	// StandardProfile includes all bridges needed for normal operation
	StandardProfile = BridgeProfile{
		Name:        "standard",
		Description: "Standard bridge set for full-featured script execution",
		BridgeSets:  []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetAgent, BridgeSetObservability, BridgeSetState, BridgeSetStructured},
	}

	// MinimalProfile includes only essential bridges
	MinimalProfile = BridgeProfile{
		Name:        "minimal",
		Description: "Minimal bridge set for lightweight execution",
		BridgeSets:  []BridgeSet{BridgeSetCore, BridgeSetUtility},
	}

	// LLMProfile optimized for LLM operations
	LLMProfile = BridgeProfile{
		Name:        "llm",
		Description: "Bridge set optimized for LLM operations",
		BridgeSets:  []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetStructured},
	}

	// DevelopmentProfile includes debugging and observability bridges
	DevelopmentProfile = BridgeProfile{
		Name:        "development",
		Description: "Bridge set for development and debugging",
		BridgeSets:  []BridgeSet{BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetObservability},
	}
)

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

// RegisterBridgeProfile registers bridges according to a predefined profile
func RegisterBridgeProfile(scriptEngine types.ScriptEngine, profile BridgeProfile) error {
	return RegisterBridgeSets(scriptEngine, profile.BridgeSets)
}

// RegisterStandardBridges registers all standard bridges (backward compatibility)
// This ensures that bridges are available to scripts via the global bridges table.
func RegisterStandardBridges(scriptEngine types.ScriptEngine) error {
	return RegisterBridgeProfile(scriptEngine, StandardProfile)
}

// RegisterBridgesWithRegistry registers bridges with an engine obtained from the registry using a profile
func RegisterBridgesWithRegistry(registry *types.Registry, engineName string, config types.EngineConfig, profile BridgeProfile) error {
	// Get engine instance
	scriptEngine, err := registry.GetEngine(engineName, config)
	if err != nil {
		return fmt.Errorf("failed to get engine %s: %w", engineName, err)
	}

	// Register bridges using profile
	return RegisterBridgeProfile(scriptEngine, profile)
}

// RegisterStandardBridgesWithRegistry registers standard bridges with an engine obtained from the registry.
// This is a convenience function for backward compatibility.
func RegisterStandardBridgesWithRegistry(registry *types.Registry, engineName string, config types.EngineConfig) error {
	return RegisterBridgesWithRegistry(registry, engineName, config, StandardProfile)
}

// RegisterBridgesWithAllEngines registers bridges with all engines in the registry using engine-specific profiles
func RegisterBridgesWithAllEngines(registry *types.Registry, engineProfiles map[string]BridgeProfile) error {
	engines := registry.ListEngines()
	for _, engineInfo := range engines {
		profile, exists := engineProfiles[engineInfo.Name]
		if !exists {
			// Use standard profile as default
			profile = StandardProfile
		}

		// Create default config for bridge registration
		config := types.EngineConfig{
			SandboxMode: true,
			DebugMode:   false,
		}

		if err := RegisterBridgesWithRegistry(registry, engineInfo.Name, config, profile); err != nil {
			return fmt.Errorf("failed to register bridges with engine %s: %w", engineInfo.Name, err)
		}
	}

	return nil
}

// GetAvailableBridgeProfiles returns all available bridge profiles
func GetAvailableBridgeProfiles() []BridgeProfile {
	return []BridgeProfile{
		StandardProfile,
		MinimalProfile,
		LLMProfile,
		DevelopmentProfile,
	}
}

// GetBridgeProfileByName returns a bridge profile by name
func GetBridgeProfileByName(name string) (BridgeProfile, error) {
	profiles := GetAvailableBridgeProfiles()
	for _, profile := range profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return BridgeProfile{}, fmt.Errorf("bridge profile %s not found", name)
}
