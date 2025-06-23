// ABOUTME: Feature set definitions for controlling available go-llmspell functionality.
// ABOUTME: Single source of truth for all feature set enums and bridge mappings.

package registry

import (
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// FeatureSet represents the set of go-llmspell functionality to enable.
// This is the single source of truth for feature set definitions.
type FeatureSet string

const (
	// FeatureSetMinimal includes only essential bridges (Core + Utility)
	FeatureSetMinimal FeatureSet = "minimal"
	// FeatureSetLLM includes LLM-focused functionality (Core + Utility + LLM + Structured)
	FeatureSetLLM FeatureSet = "llm"
	// FeatureSetAgent includes agent workflow functionality (Core + Utility + LLM + Structured + Agent + State)
	FeatureSetAgent FeatureSet = "agent"
	// FeatureSetObservable includes debugging and monitoring (Core + Utility + LLM + Observability)
	FeatureSetObservable FeatureSet = "observable"
	// FeatureSetFull includes all available bridges
	FeatureSetFull FeatureSet = "full"
)

// IsValidFeatureSet validates if a string represents a valid feature set.
// This prevents redefinition of feature sets across the codebase.
func IsValidFeatureSet(fs string) bool {
	switch FeatureSet(fs) {
	case FeatureSetMinimal, FeatureSetLLM, FeatureSetAgent, FeatureSetObservable, FeatureSetFull:
		return true
	default:
		return false
	}
}

// FeatureSetBridges defines the mapping from feature sets to bridge sets.
// This is the single source of truth for which bridges are included in each feature set.
var FeatureSetBridges = map[FeatureSet][]BridgeSet{
	// Minimal: Just core functionality
	FeatureSetMinimal: {
		BridgeSetCore,
		BridgeSetUtility,
	},

	// LLM: Language model focused
	FeatureSetLLM: {
		BridgeSetCore,
		BridgeSetUtility,
		BridgeSetLLM,
		BridgeSetStructured,
	},

	// Agent: Full agent workflow support
	FeatureSetAgent: {
		BridgeSetCore,
		BridgeSetUtility,
		BridgeSetLLM,
		BridgeSetStructured,
		BridgeSetAgent,
		BridgeSetState,
	},

	// Observable: Development and debugging focus
	FeatureSetObservable: {
		BridgeSetCore,
		BridgeSetUtility,
		BridgeSetLLM,
		BridgeSetObservability,
	},

	// Full: Everything available
	FeatureSetFull: {
		BridgeSetCore,
		BridgeSetUtility,
		BridgeSetLLM,
		BridgeSetStructured,
		BridgeSetAgent,
		BridgeSetState,
		BridgeSetObservability,
	},
}

// GetBridgeSetsForFeature returns the bridge sets to load for a given feature set.
// This is the single source for feature set to bridge set mappings.
func GetBridgeSetsForFeature(fs FeatureSet) []BridgeSet {
	if sets, ok := FeatureSetBridges[fs]; ok {
		return sets
	}
	// Default to minimal if unknown
	return FeatureSetBridges[FeatureSetMinimal]
}

// GetBridgesForFeatureSet returns all bridge instances for a given feature set.
// It creates bridges based on the feature set's bridge set configuration.
func GetBridgesForFeatureSet(fs FeatureSet) ([]types.Bridge, error) {
	bridgeSets := GetBridgeSetsForFeature(fs)

	var bridges []types.Bridge
	for _, set := range bridgeSets {
		if factory, exists := bridgeFactories[set]; exists {
			setBridges, err := factory()
			if err != nil {
				return nil, err
			}
			bridges = append(bridges, setBridges...)
		}
	}

	return bridges, nil
}

// FeatureSetDescription provides human-readable descriptions of each feature set.
// Useful for documentation and help text.
var FeatureSetDescription = map[FeatureSet]string{
	FeatureSetMinimal:    "Essential functionality only (Core + Utility bridges)",
	FeatureSetLLM:        "Language model operations (adds LLM + Structured bridges)",
	FeatureSetAgent:      "Agent workflows and tools (adds Agent + State bridges)",
	FeatureSetObservable: "Development and debugging (adds Observability bridges)",
	FeatureSetFull:       "All available functionality (all bridge sets)",
}

// GetFeatureSetDescription returns a human-readable description of a feature set.
func GetFeatureSetDescription(fs FeatureSet) string {
	if desc, ok := FeatureSetDescription[fs]; ok {
		return desc
	}
	return "Unknown feature set"
}
