// ABOUTME: Tests for ModelInfo bridge adapter that exposes go-llms model discovery and comparison functionality to Lua scripts
// ABOUTME: Validates model discovery, capability querying, model comparison, and recommendation functionality

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestModelInfoAdapter_Creation(t *testing.T) {
	t.Run("create_modelinfo_adapter", func(t *testing.T) {
		// Create modelinfo bridge mock
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Model Info Bridge",
				Version:     "1.0.0",
				Description: "Provides access to go-llms ModelRegistry for model discovery",
			}).
			WithMethod("listModels", engine.MethodInfo{
				Name:        "listModels",
				Description: "List all models from all registries",
				ReturnType:  "array",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock model list
				models := []engine.ScriptValue{
					engine.NewStringValue("gpt-4"),
					engine.NewStringValue("claude-3-opus"),
					engine.NewStringValue("llama-3-70b"),
				}
				return engine.NewArrayValue(models), nil
			}).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name:        "fetchModelInventory",
				Description: "Fetch complete model inventory",
				ReturnType:  "object",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock inventory
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{
						"version":     engine.NewStringValue("1.0.0"),
						"lastUpdated": engine.NewStringValue("2024-01-01T00:00:00Z"),
					}),
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":     engine.NewStringValue("gpt-4"),
							"provider": engine.NewStringValue("openai"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"text": engine.NewObjectValue(map[string]engine.ScriptValue{
									"read":  engine.NewBoolValue(true),
									"write": engine.NewBoolValue(true),
								}),
								"functionCalling": engine.NewBoolValue(true),
							}),
							"contextWindow":   engine.NewNumberValue(8192),
							"maxOutputTokens": engine.NewNumberValue(4096),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens":  engine.NewNumberValue(0.03),
								"outputPer1kTokens": engine.NewNumberValue(0.06),
							}),
						}),
					}),
				}), nil
			})

		// Create adapter
		adapter := NewModelInfoAdapter(modelinfoBridge)
		require.NotNil(t, adapter)

		// Should have modelinfo-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "listModels")
		assert.Contains(t, methods, "fetchModelInventory")
		// Legacy namespaced methods
		assert.Contains(t, methods, "getModelCapabilities")
		assert.Contains(t, methods, "findModelsByCapability")
		assert.Contains(t, methods, "suggestModel")
		assert.Contains(t, methods, "compareModels")
		// Flattened discovery methods
		assert.Contains(t, methods, "discoveryScan")
		assert.Contains(t, methods, "discoveryRefresh")
		assert.Contains(t, methods, "discoveryGetProviders")
		assert.Contains(t, methods, "discoveryGetModels")
		// Flattened capabilities methods
		assert.Contains(t, methods, "capabilitiesCheck")
		assert.Contains(t, methods, "capabilitiesList")
		assert.Contains(t, methods, "capabilitiesCompare")
		assert.Contains(t, methods, "capabilitiesGetDetails")
		// Flattened selection methods
		assert.Contains(t, methods, "selectionFind")
		assert.Contains(t, methods, "selectionRank")
		assert.Contains(t, methods, "selectionFilter")
		assert.Contains(t, methods, "selectionRecommend")
	})

	t.Run("modelinfo_module_structure", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Model Info Bridge",
			}).
			WithMethod("listModels", engine.MethodInfo{
				Name: "listModels",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{
					engine.NewStringValue("gpt-4"),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module table
		module := L.Get(-1)
		L.SetGlobal("modelinfo", module)

		// Test module structure
		err = L.DoString(`
			-- Check basic module properties
			assert(modelinfo._adapter == "modelinfo", "should have correct adapter name")
			assert(modelinfo._version == "1.0.0", "should have correct version")
			
			-- Check namespaces exist (backward compatibility)
			assert(type(modelinfo.discovery) == "table", "discovery namespace should exist")
			assert(type(modelinfo.capabilities) == "table", "capabilities namespace should exist")
			assert(type(modelinfo.selection) == "table", "selection namespace should exist")
			
			-- Check flattened methods exist
			assert(type(modelinfo.discoveryScan) == "function", "discoveryScan should exist")
			assert(type(modelinfo.discoveryRefresh) == "function", "discoveryRefresh should exist")
			assert(type(modelinfo.discoveryGetProviders) == "function", "discoveryGetProviders should exist")
			assert(type(modelinfo.discoveryGetModels) == "function", "discoveryGetModels should exist")
			assert(type(modelinfo.capabilitiesCheck) == "function", "capabilitiesCheck should exist")
			assert(type(modelinfo.capabilitiesList) == "function", "capabilitiesList should exist")
			assert(type(modelinfo.capabilitiesCompare) == "function", "capabilitiesCompare should exist")
			assert(type(modelinfo.capabilitiesGetDetails) == "function", "capabilitiesGetDetails should exist")
			assert(type(modelinfo.selectionFind) == "function", "selectionFind should exist")
			assert(type(modelinfo.selectionRank) == "function", "selectionRank should exist")
			assert(type(modelinfo.selectionFilter) == "function", "selectionFilter should exist")
			assert(type(modelinfo.selectionRecommend) == "function", "selectionRecommend should exist")
			
			-- Check capability types
			assert(modelinfo.capabilities.TEXT_READ == "text.read", "should have text read capability")
			assert(modelinfo.capabilities.TEXT_WRITE == "text.write", "should have text write capability")
			assert(modelinfo.capabilities.FUNCTION_CALLING == "functionCalling", "should have function calling capability")
		`)
		assert.NoError(t, err)
	})
}

func TestModelInfoAdapter_Discovery(t *testing.T) {
	t.Run("list_models", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("listModels", engine.MethodInfo{
				Name: "listModels",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				models := []engine.ScriptValue{
					engine.NewStringValue("gpt-4"),
					engine.NewStringValue("claude-3-opus"),
					engine.NewStringValue("llama-3-70b"),
				}
				return engine.NewArrayValue(models), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- List all models
			local models, err = modelinfo.discovery.listModels()
			assert(err == nil, "should not error")
			assert(#models == 3, "should have 3 models")
			assert(models[1] == "gpt-4", "should have gpt-4")
			assert(models[2] == "claude-3-opus", "should have claude-3-opus")
			assert(models[3] == "llama-3-70b", "should have llama-3-70b")
		`)
		assert.NoError(t, err)
	})

	t.Run("fetch_model_inventory", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{
						"version":     engine.NewStringValue("1.0.0"),
						"lastUpdated": engine.NewStringValue("2024-01-01T00:00:00Z"),
					}),
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":          engine.NewStringValue("gpt-4"),
							"provider":      engine.NewStringValue("openai"),
							"contextWindow": engine.NewNumberValue(8192),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Fetch inventory
			local inventory, err = modelinfo.discovery.fetchInventory()
			assert(err == nil, "should not error")
			assert(inventory.metadata.version == "1.0.0", "should have version")
			assert(#inventory.models == 1, "should have 1 model")
			assert(inventory.models[1].name == "gpt-4", "should have gpt-4")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_model_capabilities", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("gpt-4"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"text": engine.NewObjectValue(map[string]engine.ScriptValue{
									"read":  engine.NewBoolValue(true),
									"write": engine.NewBoolValue(true),
								}),
								"functionCalling": engine.NewBoolValue(true),
								"streaming":       engine.NewBoolValue(true),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Get model capabilities
			local capabilities, err = modelinfo.capabilities.getModelCapabilities("gpt-4")
			assert(err == nil, "should not error")
			assert(capabilities.text.read == true, "should support text read")
			assert(capabilities.text.write == true, "should support text write")
			assert(capabilities.functionCalling == true, "should support function calling")
			assert(capabilities.streaming == true, "should support streaming")
		`)
		assert.NoError(t, err)
	})

	t.Run("find_models_by_capability", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("gpt-4"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
								"streaming":       engine.NewBoolValue(true),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("claude-3-opus"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(false),
								"streaming":       engine.NewBoolValue(true),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Find models by capability
			local models, err = modelinfo.capabilities.findModelsByCapability("functionCalling")
			assert(err == nil, "should not error")
			assert(#models == 1, "should find 1 model with function calling")
			assert(models[1].name == "gpt-4", "should find gpt-4")
			
			-- Find models by streaming capability
			local streamingModels, err2 = modelinfo.capabilities.findModelsByCapability("streaming")
			assert(err2 == nil, "should not error")
			assert(#streamingModels == 2, "should find 2 models with streaming")
		`)
		assert.NoError(t, err)
	})
}

func TestModelInfoAdapter_Selection(t *testing.T) {
	t.Run("suggest_model", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":            engine.NewStringValue("gpt-4"),
							"provider":        engine.NewStringValue("openai"),
							"contextWindow":   engine.NewNumberValue(8192),
							"maxOutputTokens": engine.NewNumberValue(4096),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
							}),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.03),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":            engine.NewStringValue("llama-3-70b"),
							"provider":        engine.NewStringValue("meta"),
							"contextWindow":   engine.NewNumberValue(4096),
							"maxOutputTokens": engine.NewNumberValue(2048),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(false),
							}),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.001),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Suggest model with function calling requirement
			local suggestion, err = modelinfo.selection.suggestModel({
				capabilities = {"functionCalling"},
				minContextWindow = 8000
			})
			assert(err == nil, "should not error")
			assert(suggestion.model.name == "gpt-4", "should suggest gpt-4")
			assert(suggestion.reason:find("function calling"), "should mention function calling in reason")
			
			-- Suggest cheapest model
			local cheapSuggestion, err2 = modelinfo.selection.suggestModel({
				priority = "cost"
			})
			assert(err2 == nil, "should not error")
			assert(cheapSuggestion.model.name == "llama-3-70b", "should suggest cheapest model")
		`)
		assert.NoError(t, err)
	})

	t.Run("compare_models", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":            engine.NewStringValue("gpt-4"),
							"contextWindow":   engine.NewNumberValue(8192),
							"maxOutputTokens": engine.NewNumberValue(4096),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.03),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":            engine.NewStringValue("claude-3-opus"),
							"contextWindow":   engine.NewNumberValue(200000),
							"maxOutputTokens": engine.NewNumberValue(4096),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.015),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Compare two models
			local comparison, err = modelinfo.selection.compareModels({"gpt-4", "claude-3-opus"})
			assert(err == nil, "should not error")
			assert(#comparison.models == 2, "should compare 2 models")
			assert(comparison.comparison.contextWindow.winner == "claude-3-opus", "claude should win context window")
			assert(comparison.comparison.pricing.winner == "claude-3-opus", "claude should win pricing")
			
			-- Check summary
			assert(type(comparison.summary) == "table", "should have summary")
			assert(type(comparison.summary.strengths) == "table", "should have strengths")
		`)
		assert.NoError(t, err)
	})
}

func TestModelInfoAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("model service unavailable")
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Try to fetch inventory with error
			local result, err = modelinfo.discovery.fetchInventory()
			assert(err ~= nil, "should have error")
			assert(string.find(err, "model service unavailable"), "error should contain message")
			assert(result == nil, "result should be nil on error")
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_missing_model", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Try to get capabilities for non-existent model
			local capabilities, err = modelinfo.capabilities.getModelCapabilities("non-existent-model")
			assert(err ~= nil, "should have error")
			assert(string.find(err, "model not found"), "error should indicate model not found")
			assert(capabilities == nil, "capabilities should be nil")
		`)
		assert.NoError(t, err)
	})
}

func TestModelInfoAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("estimate_cost", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("gpt-4"),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens":  engine.NewNumberValue(0.03),
								"outputPer1kTokens": engine.NewNumberValue(0.06),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Estimate cost for 1000 input tokens and 500 output tokens
			local cost, err = modelinfo.selection.estimateCost("gpt-4", {
				inputTokens = 1000,
				outputTokens = 500
			})
			assert(err == nil, "should not error")
			assert(cost.inputCost == 0.03, "should calculate input cost correctly")
			assert(cost.outputCost == 0.03, "should calculate output cost correctly")  -- 500 tokens * 0.06 / 1000
			assert(cost.totalCost == 0.06, "should calculate total cost correctly")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_best_model_for_task", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("gpt-4"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("claude-3-opus"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(false),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Get best model for function calling task
			local best, err = modelinfo.selection.getBestModelForTask("function_calling")
			assert(err == nil, "should not error")
			assert(best.name == "gpt-4", "should return gpt-4 for function calling")
			assert(best.reason:find("function calling"), "should explain why gpt-4 was chosen")
		`)
		assert.NoError(t, err)
	})
}

// Test flattened methods specifically
func TestModelInfoAdapter_FlattenedMethods(t *testing.T) {
	t.Run("flattened_discovery_methods", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("listModels", engine.MethodInfo{
				Name: "listModels",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{
					engine.NewStringValue("gpt-4"),
					engine.NewStringValue("claude-3-opus"),
				}), nil
			}).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":     engine.NewStringValue("gpt-4"),
							"provider": engine.NewStringValue("openai"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":     engine.NewStringValue("claude-3-opus"),
							"provider": engine.NewStringValue("anthropic"),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test flattened discovery methods
			local models, err = modelinfo.discoveryScan()
			assert(err == nil, "discoveryScan should not error")
			assert(#models == 2, "should have 2 models")
			assert(models[1] == "gpt-4", "should have gpt-4")
			
			local inventory, err2 = modelinfo.discoveryRefresh()
			assert(err2 == nil, "discoveryRefresh should not error")
			assert(#inventory.models == 2, "should have 2 models in inventory")
			
			local providers, err3 = modelinfo.discoveryGetProviders()
			assert(err3 == nil, "discoveryGetProviders should not error")
			assert(#providers == 2, "should have 2 providers")
			
			local modelsAgain, err4 = modelinfo.discoveryGetModels()
			assert(err4 == nil, "discoveryGetModels should not error")
			assert(#modelsAgain == 2, "should have 2 models")
		`)
		assert.NoError(t, err)
	})

	t.Run("flattened_capabilities_methods", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("gpt-4"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
								"streaming":       engine.NewBoolValue(true),
							}),
							"contextWindow": engine.NewNumberValue(8192),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.03),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewStringValue("claude-3-opus"),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(false),
								"streaming":       engine.NewBoolValue(true),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test flattened capabilities methods
			local capabilities, err = modelinfo.capabilitiesCheck("gpt-4")
			assert(err == nil, "capabilitiesCheck should not error")
			assert(capabilities.functionCalling == true, "should have function calling")
			
			local allCaps, err2 = modelinfo.capabilitiesList()
			assert(err2 == nil, "capabilitiesList should not error")
			assert(#allCaps > 0, "should have capabilities list")
			
			local funcModels, err3 = modelinfo.capabilitiesCompare("functionCalling")
			assert(err3 == nil, "capabilitiesCompare should not error")
			assert(#funcModels == 1, "should find 1 model with function calling")
			
			local details, err4 = modelinfo.capabilitiesGetDetails("gpt-4")
			assert(err4 == nil, "capabilitiesGetDetails should not error")
			assert(details.functionCalling == true, "details should have function calling")
			assert(details.contextWindow == 8192, "details should have context window")
		`)
		assert.NoError(t, err)
	})

	t.Run("flattened_selection_methods", func(t *testing.T) {
		modelinfoBridge := testutils.NewMockBridge("modelinfo").
			WithInitialized(true).
			WithMethod("fetchModelInventory", engine.MethodInfo{
				Name: "fetchModelInventory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"models": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":          engine.NewStringValue("gpt-4"),
							"provider":      engine.NewStringValue("openai"),
							"contextWindow": engine.NewNumberValue(8192),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(true),
							}),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.03),
							}),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":          engine.NewStringValue("llama-3-70b"),
							"provider":      engine.NewStringValue("meta"),
							"contextWindow": engine.NewNumberValue(4096),
							"capabilities": engine.NewObjectValue(map[string]engine.ScriptValue{
								"functionCalling": engine.NewBoolValue(false),
							}),
							"pricing": engine.NewObjectValue(map[string]engine.ScriptValue{
								"inputPer1kTokens": engine.NewNumberValue(0.001),
							}),
						}),
					}),
				}), nil
			})

		adapter := NewModelInfoAdapter(modelinfoBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "modelinfo")
		require.NoError(t, err)

		err = ms.LoadModule(L, "modelinfo")
		require.NoError(t, err)

		err = L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test flattened selection methods
			local suggestion, err = modelinfo.selectionFind({
				capabilities = {"functionCalling"},
				minContextWindow = 8000
			})
			assert(err == nil, "selectionFind should not error")
			assert(suggestion.model.name == "gpt-4", "should suggest gpt-4")
			
			local ranked, err2 = modelinfo.selectionRank("cost")
			assert(err2 == nil, "selectionRank should not error")
			assert(#ranked == 2, "should rank both models")
			assert(ranked[1].name == "llama-3-70b", "cheapest should be first")
			
			local filtered, err3 = modelinfo.selectionFilter({
				minContextWindow = 5000
			})
			assert(err3 == nil, "selectionFilter should not error")
			assert(#filtered == 1, "should filter to 1 model")
			assert(filtered[1].name == "gpt-4", "should filter to gpt-4")
			
			local recommendation, err4 = modelinfo.selectionRecommend("function_calling")
			assert(err4 == nil, "selectionRecommend should not error")
			assert(recommendation.name == "gpt-4", "should recommend gpt-4 for function calling")
		`)
		assert.NoError(t, err)
	})
}
