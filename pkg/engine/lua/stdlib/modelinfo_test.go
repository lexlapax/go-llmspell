// ABOUTME: Tests for the modelinfo Lua module with mock bridge
// ABOUTME: Validates model discovery and comparison functionality

package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// setupModelInfoBridge sets up mock model info bridge
func setupModelInfoBridge(t *testing.T, L *lua.LState) {
	t.Helper()

	// Create bridges table
	bridges := L.NewTable()
	L.SetGlobal("bridges", bridges)

	// Create model info bridge
	modelInfoBridge := L.NewTable()

	// Mock model storage
	models := make(map[string]*lua.LTable)
	
	// Create model tables
	gpt4 := L.NewTable()
	gpt4.RawSetString("name", lua.LString("gpt-4"))
	gpt4.RawSetString("provider", lua.LString("openai"))
	gpt4.RawSetString("contextWindow", lua.LNumber(128000))
	gpt4.RawSetString("maxOutput", lua.LNumber(4096))
	models["gpt-4"] = gpt4
	
	claude := L.NewTable()
	claude.RawSetString("name", lua.LString("claude-3-opus"))
	claude.RawSetString("provider", lua.LString("anthropic"))
	claude.RawSetString("contextWindow", lua.LNumber(200000))
	claude.RawSetString("maxOutput", lua.LNumber(4096))
	models["claude-3-opus"] = claude
	
	llama := L.NewTable()
	llama.RawSetString("name", lua.LString("llama-3-70b"))
	llama.RawSetString("provider", lua.LString("meta"))
	llama.RawSetString("contextWindow", lua.LNumber(8192))
	llama.RawSetString("maxOutput", lua.LNumber(2048))
	models["llama-3-70b"] = llama

	// Discovery methods
	modelInfoBridge.RawSetString("discoveryScan", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		result.RawSetString("providers", lua.LNumber(3))
		result.RawSetString("models", lua.LNumber(len(models)))
		result.RawSetString("status", lua.LString("completed"))
		L.Push(result)
		return 1
	}))

	modelInfoBridge.RawSetString("discoveryRefresh", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LTrue)
		return 1
	}))

	modelInfoBridge.RawSetString("discoveryGetProviders", L.NewFunction(func(L *lua.LState) int {
		providers := L.NewTable()
		providers.Append(lua.LString("openai"))
		providers.Append(lua.LString("anthropic"))
		providers.Append(lua.LString("meta"))
		L.Push(providers)
		return 1
	}))

	modelInfoBridge.RawSetString("discoveryGetModels", L.NewFunction(func(L *lua.LState) int {
		modelList := L.NewTable()
		for _, model := range models {
			modelList.Append(model)
		}
		L.Push(modelList)
		return 1
	}))

	// Legacy discovery methods
	modelInfoBridge.RawSetString("listModels", L.NewFunction(func(L *lua.LState) int {
		modelList := L.NewTable()
		for name := range models {
			modelList.Append(lua.LString(name))
		}
		L.Push(modelList)
		return 1
	}))

	modelInfoBridge.RawSetString("fetchModelInventory", L.NewFunction(func(L *lua.LState) int {
		inventory := L.NewTable()
		for name, model := range models {
			inventory.RawSetString(name, model)
		}
		L.Push(inventory)
		return 1
	}))

	modelInfoBridge.RawSetString("getModel", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		if model, ok := models[modelName]; ok {
			L.Push(model)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	modelInfoBridge.RawSetString("listRegistries", L.NewFunction(func(L *lua.LState) int {
		registries := L.NewTable()
		registries.Append(lua.LString("huggingface"))
		registries.Append(lua.LString("ollama"))
		registries.Append(lua.LString("custom"))
		L.Push(registries)
		return 1
	}))

	// Capabilities methods
	modelInfoBridge.RawSetString("capabilitiesCheck", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		
		caps := L.NewTable()
		if modelName == "gpt-4" {
			caps.RawSetString("functionCalling", lua.LTrue)
			caps.RawSetString("streaming", lua.LTrue)
			textCaps := L.NewTable()
			textCaps.RawSetString("read", lua.LTrue)
			textCaps.RawSetString("write", lua.LTrue)
			caps.RawSetString("text", textCaps)
			imageCaps := L.NewTable()
			imageCaps.RawSetString("read", lua.LTrue)
			caps.RawSetString("image", imageCaps)
		} else if modelName == "claude-3-opus" {
			caps.RawSetString("functionCalling", lua.LTrue)
			caps.RawSetString("streaming", lua.LTrue)
			textCaps := L.NewTable()
			textCaps.RawSetString("read", lua.LTrue)
			textCaps.RawSetString("write", lua.LTrue)
			caps.RawSetString("text", textCaps)
			imageCaps := L.NewTable()
			imageCaps.RawSetString("read", lua.LTrue)
			caps.RawSetString("image", imageCaps)
		} else if modelName == "llama-3-70b" {
			caps.RawSetString("streaming", lua.LTrue)
			textCaps := L.NewTable()
			textCaps.RawSetString("read", lua.LTrue)
			textCaps.RawSetString("write", lua.LTrue)
			caps.RawSetString("text", textCaps)
		}
		
		L.Push(caps)
		return 1
	}))

	modelInfoBridge.RawSetString("capabilitiesList", L.NewFunction(func(L *lua.LState) int {
		capsList := L.NewTable()
		capsList.Append(lua.LString("text.read"))
		capsList.Append(lua.LString("text.write"))
		capsList.Append(lua.LString("image.read"))
		capsList.Append(lua.LString("image.write"))
		capsList.Append(lua.LString("functionCalling"))
		capsList.Append(lua.LString("streaming"))
		L.Push(capsList)
		return 1
	}))

	modelInfoBridge.RawSetString("capabilitiesCompare", L.NewFunction(func(L *lua.LState) int {
		capability := L.CheckString(1)
		
		modelsWithCap := L.NewTable()
		if capability == "functionCalling" {
			modelsWithCap.Append(models["gpt-4"])
			modelsWithCap.Append(models["claude-3-opus"])
		} else if capability == "streaming" {
			modelsWithCap.Append(models["gpt-4"])
			modelsWithCap.Append(models["claude-3-opus"])
			modelsWithCap.Append(models["llama-3-70b"])
		}
		
		L.Push(modelsWithCap)
		return 1
	}))

	modelInfoBridge.RawSetString("capabilitiesGetDetails", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		
		details := L.NewTable()
		if model, ok := models[modelName]; ok {
			details.RawSetString("model", model)
			caps := L.NewTable()
			if modelName == "gpt-4" || modelName == "claude-3-opus" {
				caps.RawSetString("functionCalling", lua.LTrue)
			}
			caps.RawSetString("streaming", lua.LTrue)
			details.RawSetString("capabilities", caps)
		}
		
		L.Push(details)
		return 1
	}))

	// Legacy capabilities methods
	modelInfoBridge.RawSetString("getModelCapabilities", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		
		if modelName == "gpt-4" || modelName == "claude-3-opus" {
			caps := L.NewTable()
			caps.Append(lua.LString("text"))
			caps.Append(lua.LString("image"))
			caps.Append(lua.LString("function_calling"))
			caps.Append(lua.LString("streaming"))
			L.Push(caps)
		} else if modelName == "llama-3-70b" {
			caps := L.NewTable()
			caps.Append(lua.LString("text"))
			caps.Append(lua.LString("streaming"))
			L.Push(caps)
		} else {
			L.Push(L.NewTable())
		}
		return 1
	}))

	modelInfoBridge.RawSetString("findModelsByCapability", L.NewFunction(func(L *lua.LState) int {
		capability := L.CheckString(1)
		
		found := L.NewTable()
		if capability == "function_calling" {
			found.Append(lua.LString("gpt-4"))
			found.Append(lua.LString("claude-3-opus"))
		} else if capability == "streaming" {
			found.Append(lua.LString("gpt-4"))
			found.Append(lua.LString("claude-3-opus"))
			found.Append(lua.LString("llama-3-70b"))
		} else if capability == "text" {
			for name := range models {
				found.Append(lua.LString(name))
			}
		}
		
		L.Push(found)
		return 1
	}))

	// Selection methods
	modelInfoBridge.RawSetString("selectionFind", L.NewFunction(func(L *lua.LState) int {
		requirements := L.CheckTable(1)
		
		// Mock implementation - return best model based on priority
		priority := requirements.RawGetString("priority")
		if priority.String() == "cost" {
			L.Push(models["llama-3-70b"])
		} else if priority.String() == "context_window" {
			L.Push(models["claude-3-opus"])
		} else if priority.String() == "capability" {
			L.Push(models["gpt-4"])
		} else {
			L.Push(models["gpt-4"]) // default
		}
		return 1
	}))

	modelInfoBridge.RawSetString("selectionRank", L.NewFunction(func(L *lua.LState) int {
		criteria := L.CheckString(1)
		
		ranked := L.NewTable()
		if criteria == "cost" {
			ranked.Append(models["llama-3-70b"])
			ranked.Append(models["gpt-4"])
			ranked.Append(models["claude-3-opus"])
		} else if criteria == "performance" {
			ranked.Append(models["claude-3-opus"])
			ranked.Append(models["gpt-4"])
			ranked.Append(models["llama-3-70b"])
		} else {
			// Default ranking
			ranked.Append(models["gpt-4"])
			ranked.Append(models["claude-3-opus"])
			ranked.Append(models["llama-3-70b"])
		}
		
		L.Push(ranked)
		return 1
	}))

	modelInfoBridge.RawSetString("selectionFilter", L.NewFunction(func(L *lua.LState) int {
		filters := L.CheckTable(1)
		
		filtered := L.NewTable()
		
		// Check context window filter
		minContext := filters.RawGetString("minContextWindow")
		if minContext != lua.LNil {
			minCtx := int(minContext.(lua.LNumber))
			for _, model := range models {
				ctxWindow := model.RawGetString("contextWindow")
				if ctxWindow != lua.LNil && int(ctxWindow.(lua.LNumber)) >= minCtx {
					filtered.Append(model)
				}
			}
		} else {
			// No filter, return all
			for _, model := range models {
				filtered.Append(model)
			}
		}
		
		L.Push(filtered)
		return 1
	}))

	modelInfoBridge.RawSetString("selectionRecommend", L.NewFunction(func(L *lua.LState) int {
		task := L.CheckString(1)
		
		if task == "function_calling" {
			L.Push(models["gpt-4"])
		} else if task == "text_generation" {
			L.Push(models["claude-3-opus"])
		} else if task == "code_generation" {
			L.Push(models["gpt-4"])
		} else {
			L.Push(models["gpt-4"]) // default recommendation
		}
		return 1
	}))

	// Legacy selection methods
	modelInfoBridge.RawSetString("suggestModel", L.NewFunction(func(L *lua.LState) int {
		requirements := L.CheckTable(1)
		
		// Simple suggestion based on requirements
		caps := requirements.RawGetString("capabilities")
		if caps != lua.LNil && caps.Type() == lua.LTTable {
			capTable := caps.(*lua.LTable)
			hasFunctionCalling := false
			capTable.ForEach(func(_, v lua.LValue) {
				if v.String() == "function_calling" {
					hasFunctionCalling = true
				}
			})
			if hasFunctionCalling {
				L.Push(lua.LString("gpt-4"))
			} else {
				L.Push(lua.LString("claude-3-opus"))
			}
		} else {
			L.Push(lua.LString("gpt-4"))
		}
		return 1
	}))

	modelInfoBridge.RawSetString("compareModels", L.NewFunction(func(L *lua.LState) int {
		modelNames := L.CheckTable(1)
		
		comparison := L.NewTable()
		modelNames.ForEach(func(_, v lua.LValue) {
			if name, ok := v.(lua.LString); ok {
				if model, exists := models[string(name)]; exists {
					modelInfo := L.NewTable()
					modelInfo.RawSetString("name", name)
					modelInfo.RawSetString("contextWindow", model.RawGetString("contextWindow"))
					modelInfo.RawSetString("provider", model.RawGetString("provider"))
					comparison.Append(modelInfo)
				}
			}
		})
		
		L.Push(comparison)
		return 1
	}))

	modelInfoBridge.RawSetString("estimateCost", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		usage := L.CheckTable(2)
		
		inputTokens := usage.RawGetString("inputTokens")
		outputTokens := usage.RawGetString("outputTokens")
		
		// Mock cost calculation
		var costPerMillion float64
		if modelName == "gpt-4" {
			costPerMillion = 30.0
		} else if modelName == "claude-3-opus" {
			costPerMillion = 15.0
		} else {
			costPerMillion = 0.5
		}
		
		totalTokens := 0.0
		if inputTokens != lua.LNil {
			totalTokens += float64(inputTokens.(lua.LNumber))
		}
		if outputTokens != lua.LNil {
			totalTokens += float64(outputTokens.(lua.LNumber))
		}
		
		cost := (totalTokens / 1000000.0) * costPerMillion
		
		result := L.NewTable()
		result.RawSetString("cost", lua.LNumber(cost))
		result.RawSetString("currency", lua.LString("USD"))
		L.Push(result)
		return 1
	}))

	modelInfoBridge.RawSetString("getBestModelForTask", L.NewFunction(func(L *lua.LState) int {
		task := L.CheckString(1)
		
		if task == "function_calling" {
			L.Push(lua.LString("gpt-4"))
		} else if task == "text_generation" {
			L.Push(lua.LString("claude-3-opus"))
		} else if task == "code_generation" {
			L.Push(lua.LString("gpt-4"))
		} else if task == "analysis" {
			L.Push(lua.LString("claude-3-opus"))
		} else {
			L.Push(lua.LString("gpt-4"))
		}
		return 1
	}))

	bridges.RawSetString("llm_modelinfo", modelInfoBridge)
}

func TestModelInfoModule(t *testing.T) {
	t.Run("module loads with bridge", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			assert(type(modelinfo) == "table", "modelinfo should be a table")
			
			-- Check constants exist
			assert(type(modelinfo.CAPABILITIES) == "table", "CAPABILITIES should be a table")
			assert(modelinfo.CAPABILITIES.TEXT_READ == "text.read", "CAPABILITIES.TEXT_READ should be 'text.read'")
			assert(modelinfo.CAPABILITIES.FUNCTION_CALLING == "functionCalling", "CAPABILITIES.FUNCTION_CALLING should be 'functionCalling'")
			
			assert(type(modelinfo.RANKING) == "table", "RANKING should be a table")
			assert(modelinfo.RANKING.COST == "cost", "RANKING.COST should be 'cost'")
			assert(modelinfo.RANKING.PERFORMANCE == "performance", "RANKING.PERFORMANCE should be 'performance'")
			
			assert(type(modelinfo.PRIORITIES) == "table", "PRIORITIES should be a table")
			assert(modelinfo.PRIORITIES.COST == "cost", "PRIORITIES.COST should be 'cost'")
			
			assert(type(modelinfo.TASKS) == "table", "TASKS should be a table")
			assert(modelinfo.TASKS.FUNCTION_CALLING == "function_calling", "TASKS.FUNCTION_CALLING should be 'function_calling'")
			
			-- Check methods exist
			assert(type(modelinfo.discovery_scan) == "function", "discovery_scan should be a function")
			assert(type(modelinfo.capabilities_check) == "function", "capabilities_check should be a function")
			assert(type(modelinfo.selection_find) == "function", "selection_find should be a function")
			assert(type(modelinfo.find_cheapest_model) == "function", "find_cheapest_model should be a function")
		`)
		assert.NoError(t, err)
	})

	t.Run("discovery methods", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test discovery scan
			local scan_result = modelinfo.discovery_scan()
			assert(scan_result.providers == 3, "should have 3 providers")
			assert(scan_result.models == 3, "should have 3 models")
			assert(scan_result.status == "completed", "scan status should be completed")
			
			-- Test discovery refresh
			local refresh_result = modelinfo.discovery_refresh()
			assert(refresh_result == true, "refresh should return true")
			
			-- Test get providers
			local providers = modelinfo.discovery_get_providers()
			assert(#providers == 3, "should have 3 providers")
			assert(providers[1] == "openai", "first provider should be openai")
			
			-- Test get models
			local models = modelinfo.discovery_get_models()
			assert(#models == 3, "should have 3 models")
			
			-- Test legacy methods
			local model_list = modelinfo.list_models()
			assert(#model_list == 3, "should have 3 model names")
			
			local inventory = modelinfo.fetch_inventory()
			assert(type(inventory) == "table", "inventory should be a table")
			
			local model = modelinfo.get_model("gpt-4")
			assert(model.name == "gpt-4", "model name should match")
			assert(model.provider == "openai", "model provider should be openai")
			
			local registries = modelinfo.list_registries()
			assert(#registries == 3, "should have 3 registries")
		`)
		assert.NoError(t, err)
	})

	t.Run("capabilities methods", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test capabilities check
			local caps = modelinfo.capabilities_check("gpt-4")
			assert(caps.functionCalling == true, "gpt-4 should support function calling")
			assert(caps.streaming == true, "gpt-4 should support streaming")
			assert(type(caps.text) == "table", "gpt-4 should have text capabilities")
			assert(caps.text.read == true, "gpt-4 should support text read")
			
			-- Test capabilities list
			local cap_list = modelinfo.capabilities_list()
			assert(#cap_list >= 6, "should have at least 6 capabilities")
			
			-- Test capabilities compare
			local fc_models = modelinfo.capabilities_compare("functionCalling")
			assert(#fc_models == 2, "should have 2 models with function calling")
			
			-- Test capabilities get details
			local details = modelinfo.capabilities_get_details("claude-3-opus")
			assert(type(details.model) == "table", "should have model details")
			assert(type(details.capabilities) == "table", "should have capabilities")
			
			-- Test legacy methods
			local model_caps = modelinfo.get_model_capabilities("gpt-4")
			assert(#model_caps >= 4, "gpt-4 should have at least 4 capabilities")
			
			local fc_models_legacy = modelinfo.find_models_by_capability("function_calling")
			assert(#fc_models_legacy == 2, "should find 2 models with function calling")
		`)
		assert.NoError(t, err)
	})

	t.Run("selection methods", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test selection find
			local found = modelinfo.selection_find({
				priority = "cost"
			})
			assert(found.name == "llama-3-70b", "cheapest model should be llama-3-70b")
			
			-- Test selection rank
			local ranked = modelinfo.selection_rank("cost")
			assert(#ranked == 3, "should rank all 3 models")
			assert(ranked[1].name == "llama-3-70b", "cheapest should be first")
			
			-- Test selection filter
			local filtered = modelinfo.selection_filter({
				minContextWindow = 100000
			})
			assert(#filtered >= 1, "should find models with large context windows")
			
			-- Test selection recommend
			local recommended = modelinfo.selection_recommend("function_calling")
			assert(recommended.name == "gpt-4", "should recommend gpt-4 for function calling")
			
			-- Test legacy methods
			local suggested = modelinfo.suggest_model({
				capabilities = {"function_calling"}
			})
			assert(suggested == "gpt-4", "should suggest gpt-4")
			
			local comparison = modelinfo.compare_models({"gpt-4", "claude-3-opus"})
			assert(#comparison == 2, "should compare 2 models")
			
			local cost_estimate = modelinfo.estimate_cost("gpt-4", {
				inputTokens = 1000,
				outputTokens = 500
			})
			assert(type(cost_estimate.cost) == "number", "should return cost")
			assert(cost_estimate.currency == "USD", "should be in USD")
			
			local best = modelinfo.get_best_model_for_task("code_generation")
			assert(best == "gpt-4", "should recommend gpt-4 for code generation")
		`)
		assert.NoError(t, err)
	})

	t.Run("helper functions", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test find cheapest model
			local cheapest = modelinfo.find_cheapest_model()
			assert(cheapest.name == "llama-3-70b", "should find llama-3-70b as cheapest")
			
			-- Test find largest context model
			local largest = modelinfo.find_largest_context_model(100000)
			assert(largest.name == "claude-3-opus", "should find claude-3-opus with largest context")
			
			-- Test find most capable model
			local capable = modelinfo.find_most_capable_model({"functionCalling"})
			assert(capable.name == "gpt-4", "should find gpt-4 as most capable")
			
			-- Test quick check functions
			assert(modelinfo.supports_function_calling("gpt-4") == true, "gpt-4 should support function calling")
			assert(modelinfo.supports_function_calling("llama-3-70b") == false, "llama-3-70b should not support function calling")
			
			assert(modelinfo.supports_streaming("claude-3-opus") == true, "claude-3-opus should support streaming")
			
			assert(modelinfo.supports_images("gpt-4") == true, "gpt-4 should support images")
			assert(modelinfo.supports_images("llama-3-70b") == false, "llama-3-70b should not support images")
			
			-- Test cost estimation helper
			local conv_cost = modelinfo.estimate_conversation_cost("gpt-4", 10, 100)
			assert(type(conv_cost.cost) == "number", "should estimate conversation cost")
			assert(conv_cost.cost > 0, "cost should be positive")
			
			-- Test compare all by capability
			local streaming_models = modelinfo.compare_all_by_capability("streaming")
			assert(#streaming_models == 3, "should compare all 3 models with streaming")
			
			-- Test get top N models
			local top_2 = modelinfo.get_top_n_models("cost", 2)
			assert(#top_2 == 2, "should return top 2 models")
			assert(top_2[1].name == "llama-3-70b", "first should be cheapest")
			
			-- Test find models matching filters
			local matched = modelinfo.find_models_matching({
				min_context = 50000
			})
			assert(#matched >= 2, "should find at least 2 models with large context")
		`)
		assert.NoError(t, err)
	})

	t.Run("namespace API compatibility", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- Test discovery namespace
			assert(type(modelinfo.discovery) == "table", "discovery namespace should exist")
			assert(type(modelinfo.discovery.scan) == "function", "discovery.scan should be a function")
			assert(type(modelinfo.discovery.list_models) == "function", "discovery.list_models should be a function")
			
			local scan = modelinfo.discovery.scan()
			assert(scan.models == 3, "discovery.scan should work")
			
			-- Test capabilities namespace
			assert(type(modelinfo.capabilities) == "table", "capabilities namespace should exist")
			assert(type(modelinfo.capabilities.check) == "function", "capabilities.check should be a function")
			
			local caps = modelinfo.capabilities.check("gpt-4")
			assert(caps.functionCalling == true, "capabilities.check should work")
			
			-- Test selection namespace
			assert(type(modelinfo.selection) == "table", "selection namespace should exist")
			assert(type(modelinfo.selection.find) == "function", "selection.find should be a function")
			
			local found = modelinfo.selection.find({priority = "cost"})
			assert(found.name == "llama-3-70b", "selection.find should work")
		`)
		assert.NoError(t, err)
	})

	t.Run("missing bridge graceful failure", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Don't setup bridge
		L.SetGlobal("bridges", L.NewTable())
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			local success, err = pcall(function()
				modelinfo.discovery_scan()
			end)
			assert(not success, "should fail without bridge")
			assert(err:find("ModelInfo bridge not available"), "should have correct error message")
		`)
		assert.NoError(t, err)
	})
}

func TestModelInfoIntegration(t *testing.T) {
	t.Run("complete model selection scenario", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupModelInfoBridge(t, L)
		LoadModule(t, L, "modelinfo")

		err := L.DoString(`
			local modelinfo = require("modelinfo")
			
			-- 1. Discover available models
			local scan = modelinfo.discovery_scan()
			print(string.format("Found %d models from %d providers", scan.models, scan.providers))
			
			-- 2. Get all models and their details
			local models = modelinfo.discovery_get_models()
			local model_details = {}
			
			for i, model in ipairs(models) do
				local caps = modelinfo.capabilities_check(model.name)
				model_details[model.name] = {
					provider = model.provider,
					context_window = model.contextWindow,
					capabilities = caps
				}
			end
			
			-- 3. Find model for specific use case
			print("\n=== Use Case: Function Calling API ===")
			
			-- Find models with function calling
			local fc_models = modelinfo.find_models_by_capability("function_calling")
			print(string.format("Models with function calling: %d", #fc_models))
			
			-- Compare their costs
			local cost_comparison = {}
			for _, model_name in ipairs(fc_models) do
				local cost = modelinfo.estimate_cost(model_name, {
					inputTokens = 10000,
					outputTokens = 2000
				})
				cost_comparison[model_name] = cost.cost
				print(string.format("%s: $%.4f", model_name, cost.cost))
			end
			
			-- 4. Select best model based on requirements
			local requirements = {
				capabilities = {"function_calling"},
				priority = modelinfo.PRIORITIES.COST
			}
			
			local best_cheap = modelinfo.selection_find(requirements)
			print(string.format("\nCheapest with function calling: %s", best_cheap.name))
			
			-- 5. Get alternative recommendations
			requirements.priority = modelinfo.PRIORITIES.PERFORMANCE
			local best_performance = modelinfo.selection_find(requirements)
			print(string.format("Best performance with function calling: %s", best_performance.name))
			
			-- 6. Rank all models by different criteria
			print("\n=== Rankings ===")
			
			local cost_ranking = modelinfo.selection_rank("cost")
			print("By cost:")
			for i, model in ipairs(cost_ranking) do
				print(string.format("  %d. %s", i, model.name))
			end
			
			local perf_ranking = modelinfo.selection_rank("performance")
			print("\nBy performance:")
			for i, model in ipairs(perf_ranking) do
				print(string.format("  %d. %s", i, model.name))
			end
			
			-- 7. Filter models by context window
			print("\n=== Context Window Filter ===")
			local large_context = modelinfo.find_models_matching({
				min_context = 100000
			})
			print(string.format("Models with 100k+ context: %d", #large_context))
			for _, model in ipairs(large_context) do
				print(string.format("  - %s: %d tokens", model.name, model.contextWindow))
			end
			
			-- 8. Task-specific recommendations
			print("\n=== Task Recommendations ===")
			local tasks = {
				"function_calling",
				"text_generation",
				"code_generation",
				"analysis"
			}
			
			for _, task in ipairs(tasks) do
				local recommended = modelinfo.get_best_model_for_task(task)
				print(string.format("%s: %s", task, recommended))
			end
			
			-- 9. Estimate costs for different usage patterns
			print("\n=== Cost Estimates ===")
			
			-- Small conversation
			local small_cost = modelinfo.estimate_conversation_cost("gpt-4", 5, 50)
			print(string.format("Small conversation (5 msgs, 50 tokens): $%.4f", small_cost.cost))
			
			-- Large conversation
			local large_cost = modelinfo.estimate_conversation_cost("gpt-4", 100, 200)
			print(string.format("Large conversation (100 msgs, 200 tokens): $%.4f", large_cost.cost))
			
			-- Compare with cheaper model
			local cheap_cost = modelinfo.estimate_conversation_cost("llama-3-70b", 100, 200)
			print(string.format("Same with llama-3-70b: $%.4f (%.1f%% savings)", 
				cheap_cost.cost, 
				(1 - cheap_cost.cost / large_cost.cost) * 100
			))
			
			-- 10. Quick capability checks
			print("\n=== Quick Checks ===")
			local check_models = {"gpt-4", "claude-3-opus", "llama-3-70b"}
			
			for _, model in ipairs(check_models) do
				print(string.format("\n%s:", model))
				print(string.format("  Function calling: %s", 
					modelinfo.supports_function_calling(model) and "Yes" or "No"))
				print(string.format("  Streaming: %s", 
					modelinfo.supports_streaming(model) and "Yes" or "No"))
				print(string.format("  Images: %s", 
					modelinfo.supports_images(model) and "Yes" or "No"))
			end
			
			-- Verify we completed all steps
			assert(#models == 3, "should have discovered 3 models")
			assert(#fc_models >= 2, "should have at least 2 models with function calling")
			assert(best_cheap.name ~= best_performance.name or #fc_models == 1, 
				"different priorities should recommend different models unless only one option")
		`)
		assert.NoError(t, err)
	})
}