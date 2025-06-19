// ABOUTME: ModelInfo bridge adapter that exposes go-llms model discovery and comparison functionality to Lua scripts
// ABOUTME: Provides model discovery, capability querying, model comparison, and recommendation functionality

package adapters

import (
	"context"
	"fmt"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// ModelInfoAdapter specializes BridgeAdapter for model information functionality
type ModelInfoAdapter struct {
	*gopherlua.BridgeAdapter
}

// NewModelInfoAdapter creates a new model info adapter
func NewModelInfoAdapter(bridge engine.Bridge) *ModelInfoAdapter {
	// Create model info adapter
	adapter := &ModelInfoAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = gopherlua.NewBridgeAdapter(bridge)
	}

	return adapter
}

// GetAdapterName returns the adapter name
func (mia *ModelInfoAdapter) GetAdapterName() string {
	return "modelinfo"
}

// CreateLuaModule creates a Lua module with model info enhancements
func (mia *ModelInfoAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if mia.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := mia.BridgeAdapter.CreateLuaModule()
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(baseLoader),
				NRet:    1,
				Protect: true,
			})
			if err != nil {
				L.RaiseError("failed to create base module: %v", err)
				return 0
			}

			// Get the base module and copy its methods
			baseModule := L.Get(-1).(*lua.LTable)
			L.Pop(1)

			// Copy base module methods to our module
			baseModule.ForEach(func(k, v lua.LValue) {
				module.RawSet(k, v)
			})
		}

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("modelinfo"))
		L.SetField(module, "_version", lua.LString("1.0.0"))

		// Add discovery namespace
		mia.addDiscoveryMethods(L, module)

		// Add capabilities namespace
		mia.addCapabilityMethods(L, module)

		// Add selection namespace
		mia.addSelectionMethods(L, module)

		// Add constants
		mia.addModelInfoConstants(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addDiscoveryMethods adds model discovery methods
func (mia *ModelInfoAdapter) addDiscoveryMethods(L *lua.LState, module *lua.LTable) {
	// Create discovery namespace
	discovery := L.NewTable()

	// listModels method (enhanced wrapper)
	L.SetField(discovery, "listModels", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := mia.GetBridge().ExecuteMethod(ctx, "listModels", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// fetchInventory method
	L.SetField(discovery, "fetchInventory", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add discovery namespace to module
	L.SetField(module, "discovery", discovery)
}

// addCapabilityMethods adds capability-related methods
func (mia *ModelInfoAdapter) addCapabilityMethods(L *lua.LState, module *lua.LTable) {
	// Create capabilities namespace
	capabilities := L.NewTable()

	// Capability constants
	L.SetField(capabilities, "TEXT_READ", lua.LString("text.read"))
	L.SetField(capabilities, "TEXT_WRITE", lua.LString("text.write"))
	L.SetField(capabilities, "IMAGE_READ", lua.LString("image.read"))
	L.SetField(capabilities, "IMAGE_WRITE", lua.LString("image.write"))
	L.SetField(capabilities, "AUDIO_READ", lua.LString("audio.read"))
	L.SetField(capabilities, "AUDIO_WRITE", lua.LString("audio.write"))
	L.SetField(capabilities, "VIDEO_READ", lua.LString("video.read"))
	L.SetField(capabilities, "VIDEO_WRITE", lua.LString("video.write"))
	L.SetField(capabilities, "FILE_READ", lua.LString("file.read"))
	L.SetField(capabilities, "FILE_WRITE", lua.LString("file.write"))
	L.SetField(capabilities, "FUNCTION_CALLING", lua.LString("functionCalling"))
	L.SetField(capabilities, "STREAMING", lua.LString("streaming"))

	// getModelCapabilities method
	L.SetField(capabilities, "getModelCapabilities", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)

		// Fetch inventory to get model data
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Find the model in the inventory
		inventoryObj, ok := result.(engine.ObjectValue)
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LString("invalid inventory format"))
			return 2
		}

		modelsArray, exists := inventoryObj.Fields()["models"]
		if !exists || modelsArray.Type() != engine.TypeArray {
			L.Push(lua.LNil)
			L.Push(lua.LString("no models in inventory"))
			return 2
		}

		models := modelsArray.(engine.ArrayValue).Elements()
		for _, modelVal := range models {
			if modelVal.Type() != engine.TypeObject {
				continue
			}
			modelObj := modelVal.(engine.ObjectValue)
			nameVal, exists := modelObj.Fields()["name"]
			if !exists || nameVal.Type() != engine.TypeString {
				continue
			}
			if nameVal.(engine.StringValue).Value() == modelName {
				// Found the model, return its capabilities
				capabilitiesVal, exists := modelObj.Fields()["capabilities"]
				if !exists {
					L.Push(lua.LNil)
					L.Push(lua.LString("model has no capabilities"))
					return 2
				}

				luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, capabilitiesVal)
				if err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(err.Error()))
					return 2
				}

				L.Push(luaResult)
				L.Push(lua.LNil)
				return 2
			}
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("model not found: " + modelName))
		return 2
	}))

	// findModelsByCapability method
	L.SetField(capabilities, "findModelsByCapability", L.NewFunction(func(L *lua.LState) int {
		capability := L.CheckString(1)

		// Fetch inventory to get all models
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Filter models by capability
		inventoryObj, ok := result.(engine.ObjectValue)
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LString("invalid inventory format"))
			return 2
		}

		modelsArray, exists := inventoryObj.Fields()["models"]
		if !exists || modelsArray.Type() != engine.TypeArray {
			L.Push(lua.LNil)
			L.Push(lua.LString("no models in inventory"))
			return 2
		}

		var matchingModels []engine.ScriptValue
		models := modelsArray.(engine.ArrayValue).Elements()
		for _, modelVal := range models {
			if modelVal.Type() != engine.TypeObject {
				continue
			}
			modelObj := modelVal.(engine.ObjectValue)
			capabilitiesVal, exists := modelObj.Fields()["capabilities"]
			if !exists || capabilitiesVal.Type() != engine.TypeObject {
				continue
			}

			// Check if model has the requested capability
			capabilities := capabilitiesVal.(engine.ObjectValue).Fields()
			if mia.hasCapability(capabilities, capability) {
				matchingModels = append(matchingModels, modelVal)
			}
		}

		// Convert to Lua table
		luaModels := L.NewTable()
		for i, model := range matchingModels {
			luaModel, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, model)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			luaModels.RawSetInt(i+1, luaModel)
		}

		L.Push(luaModels)
		L.Push(lua.LNil)
		return 2
	}))

	// Add capabilities namespace to module
	L.SetField(module, "capabilities", capabilities)
}

// addSelectionMethods adds model selection and comparison methods
func (mia *ModelInfoAdapter) addSelectionMethods(L *lua.LState, module *lua.LTable) {
	// Create selection namespace
	selection := L.NewTable()

	// suggestModel method
	L.SetField(selection, "suggestModel", L.NewFunction(func(L *lua.LState) int {
		requirements := L.CheckTable(1)

		// Fetch inventory to get all models
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Parse requirements
		reqMap := mia.tableToMap(L, requirements)
		suggestion, reason, err := mia.suggestModelFromInventory(result, reqMap)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Create suggestion result
		resultMap := map[string]engine.ScriptValue{
			"model":  suggestion,
			"reason": engine.NewStringValue(reason),
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, engine.NewObjectValue(resultMap))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// compareModels method
	L.SetField(selection, "compareModels", L.NewFunction(func(L *lua.LState) int {
		modelNamesTable := L.CheckTable(1)

		// Convert table to string slice
		var modelNames []string
		modelNamesTable.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTString {
				modelNames = append(modelNames, string(v.(lua.LString)))
			}
		})

		// Fetch inventory to get model data
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		comparison, err := mia.compareModelsFromInventory(result, modelNames)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, comparison)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// estimateCost method
	L.SetField(selection, "estimateCost", L.NewFunction(func(L *lua.LState) int {
		modelName := L.CheckString(1)
		usageTable := L.CheckTable(2)

		// Parse usage parameters
		usageMap := mia.tableToMap(L, usageTable)
		inputTokens := 0.0
		outputTokens := 0.0

		if inputVal, exists := usageMap["inputTokens"]; exists && inputVal.Type() == engine.TypeNumber {
			inputTokens = inputVal.(engine.NumberValue).Value()
		}
		if outputVal, exists := usageMap["outputTokens"]; exists && outputVal.Type() == engine.TypeNumber {
			outputTokens = outputVal.(engine.NumberValue).Value()
		}

		// Fetch model pricing
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		cost, err := mia.estimateModelCost(result, modelName, inputTokens, outputTokens)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, cost)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getBestModelForTask method
	L.SetField(selection, "getBestModelForTask", L.NewFunction(func(L *lua.LState) int {
		task := L.CheckString(1)

		// Fetch inventory
		ctx := context.Background()
		result, err := mia.GetBridge().ExecuteMethod(ctx, "fetchModelInventory", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		best, reason, err := mia.getBestModelForTask(result, task)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Create result
		resultMap := map[string]engine.ScriptValue{
			"name":   engine.NewStringValue(best),
			"reason": engine.NewStringValue(reason),
		}

		luaResult, err := mia.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, engine.NewObjectValue(resultMap))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add selection namespace to module
	L.SetField(module, "selection", selection)
}

// addModelInfoConstants adds model info related constants
func (mia *ModelInfoAdapter) addModelInfoConstants(L *lua.LState, module *lua.LTable) {
	// Add priority types
	priorities := L.NewTable()
	L.SetField(priorities, "COST", lua.LString("cost"))
	L.SetField(priorities, "PERFORMANCE", lua.LString("performance"))
	L.SetField(priorities, "CONTEXT_WINDOW", lua.LString("context_window"))
	L.SetField(priorities, "CAPABILITY", lua.LString("capability"))
	L.SetField(module, "PRIORITIES", priorities)

	// Add task types
	tasks := L.NewTable()
	L.SetField(tasks, "FUNCTION_CALLING", lua.LString("function_calling"))
	L.SetField(tasks, "TEXT_GENERATION", lua.LString("text_generation"))
	L.SetField(tasks, "CODE_GENERATION", lua.LString("code_generation"))
	L.SetField(tasks, "ANALYSIS", lua.LString("analysis"))
	L.SetField(module, "TASKS", tasks)
}

// Helper methods for model selection logic

func (mia *ModelInfoAdapter) hasCapability(capabilities map[string]engine.ScriptValue, capability string) bool {
	// Handle nested capabilities like "text.read"
	if strings.Contains(capability, ".") {
		parts := strings.Split(capability, ".")
		if len(parts) != 2 {
			return false
		}

		categoryVal, exists := capabilities[parts[0]]
		if !exists || categoryVal.Type() != engine.TypeObject {
			return false
		}

		category := categoryVal.(engine.ObjectValue).Fields()
		fieldVal, exists := category[parts[1]]
		return exists && fieldVal.Type() == engine.TypeBool && fieldVal.(engine.BoolValue).Value()
	}

	// Handle direct capabilities like "functionCalling"
	val, exists := capabilities[capability]
	return exists && val.Type() == engine.TypeBool && val.(engine.BoolValue).Value()
}

func (mia *ModelInfoAdapter) suggestModelFromInventory(inventory engine.ScriptValue, requirements map[string]engine.ScriptValue) (engine.ScriptValue, string, error) {
	inventoryObj, ok := inventory.(engine.ObjectValue)
	if !ok {
		return nil, "", fmt.Errorf("invalid inventory format")
	}

	modelsArray, exists := inventoryObj.Fields()["models"]
	if !exists || modelsArray.Type() != engine.TypeArray {
		return nil, "", fmt.Errorf("no models in inventory")
	}

	models := modelsArray.(engine.ArrayValue).Elements()

	// Parse requirements
	var requiredCapabilities []string
	var minContextWindow float64
	priority := "capability" // default

	if capVal, exists := requirements["capabilities"]; exists && capVal.Type() == engine.TypeArray {
		capArray := capVal.(engine.ArrayValue).Elements()
		for _, cap := range capArray {
			if cap.Type() == engine.TypeString {
				requiredCapabilities = append(requiredCapabilities, cap.(engine.StringValue).Value())
			}
		}
	}

	if contextVal, exists := requirements["minContextWindow"]; exists && contextVal.Type() == engine.TypeNumber {
		minContextWindow = contextVal.(engine.NumberValue).Value()
	}

	if priorityVal, exists := requirements["priority"]; exists && priorityVal.Type() == engine.TypeString {
		priority = priorityVal.(engine.StringValue).Value()
	}

	// Filter and score models
	var candidates []modelCandidate
	for _, modelVal := range models {
		if modelVal.Type() != engine.TypeObject {
			continue
		}

		modelObj := modelVal.(engine.ObjectValue)
		fields := modelObj.Fields()

		// Check basic requirements
		if minContextWindow > 0 {
			contextVal, exists := fields["contextWindow"]
			if !exists || contextVal.Type() != engine.TypeNumber || contextVal.(engine.NumberValue).Value() < minContextWindow {
				continue
			}
		}

		// Check capabilities
		capabilitiesVal, exists := fields["capabilities"]
		if exists && capabilitiesVal.Type() == engine.TypeObject {
			capabilities := capabilitiesVal.(engine.ObjectValue).Fields()
			allCapsMet := true
			for _, reqCap := range requiredCapabilities {
				if !mia.hasCapability(capabilities, reqCap) {
					allCapsMet = false
					break
				}
			}
			if !allCapsMet {
				continue
			}
		} else if len(requiredCapabilities) > 0 {
			continue // Required capabilities but model has none
		}

		// Calculate score based on priority
		score := mia.calculateModelScore(fields, priority)
		candidates = append(candidates, modelCandidate{
			model: modelVal,
			score: score,
		})
	}

	if len(candidates) == 0 {
		return nil, "", fmt.Errorf("no models match the requirements")
	}

	// Sort by score (higher is better for most priorities, lower for cost)
	if priority == "cost" {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].score < candidates[j].score
		})
	} else {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].score > candidates[j].score
		})
	}

	best := candidates[0]
	reason := mia.generateSuggestionReason(best.model, priority, requiredCapabilities)

	return best.model, reason, nil
}

type modelCandidate struct {
	model engine.ScriptValue
	score float64
}

func (mia *ModelInfoAdapter) calculateModelScore(fields map[string]engine.ScriptValue, priority string) float64 {
	switch priority {
	case "cost":
		if pricingVal, exists := fields["pricing"]; exists && pricingVal.Type() == engine.TypeObject {
			pricing := pricingVal.(engine.ObjectValue).Fields()
			if inputPriceVal, exists := pricing["inputPer1kTokens"]; exists && inputPriceVal.Type() == engine.TypeNumber {
				return inputPriceVal.(engine.NumberValue).Value() // Lower is better for cost
			}
		}
		return 1.0 // Default high cost
	case "context_window":
		if contextVal, exists := fields["contextWindow"]; exists && contextVal.Type() == engine.TypeNumber {
			return contextVal.(engine.NumberValue).Value()
		}
		return 0
	default: // capability, performance
		score := 0.0
		if capabilitiesVal, exists := fields["capabilities"]; exists && capabilitiesVal.Type() == engine.TypeObject {
			capabilities := capabilitiesVal.(engine.ObjectValue).Fields()
			// Count number of capabilities
			for _, capVal := range capabilities {
				if capVal.Type() == engine.TypeBool && capVal.(engine.BoolValue).Value() {
					score += 1.0
				} else if capVal.Type() == engine.TypeObject {
					// Count nested capabilities
					nested := capVal.(engine.ObjectValue).Fields()
					for _, nestedVal := range nested {
						if nestedVal.Type() == engine.TypeBool && nestedVal.(engine.BoolValue).Value() {
							score += 0.5
						}
					}
				}
			}
		}
		return score
	}
}

func (mia *ModelInfoAdapter) generateSuggestionReason(model engine.ScriptValue, priority string, requiredCapabilities []string) string {
	modelObj := model.(engine.ObjectValue)
	fields := modelObj.Fields()

	nameVal := fields["name"]
	name := nameVal.(engine.StringValue).Value()

	var reasons []string

	switch priority {
	case "cost":
		reasons = append(reasons, "lowest cost per token")
	case "context_window":
		if contextVal, exists := fields["contextWindow"]; exists {
			context := contextVal.(engine.NumberValue).Value()
			reasons = append(reasons, fmt.Sprintf("largest context window (%.0f tokens)", context))
		}
	default:
		reasons = append(reasons, "best overall capabilities")
	}

	if len(requiredCapabilities) > 0 {
		// Transform capability names to be more readable
		readableCapabilities := make([]string, len(requiredCapabilities))
		for i, cap := range requiredCapabilities {
			switch cap {
			case "functionCalling":
				readableCapabilities[i] = "function calling"
			case "streaming":
				readableCapabilities[i] = "streaming"
			default:
				readableCapabilities[i] = strings.ReplaceAll(cap, ".", " ")
			}
		}
		reasons = append(reasons, "supports "+strings.Join(readableCapabilities, ", "))
	}

	return fmt.Sprintf("%s recommended: %s", name, strings.Join(reasons, ", "))
}

func (mia *ModelInfoAdapter) compareModelsFromInventory(inventory engine.ScriptValue, modelNames []string) (engine.ScriptValue, error) {
	inventoryObj, ok := inventory.(engine.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("invalid inventory format")
	}

	modelsArray, exists := inventoryObj.Fields()["models"]
	if !exists || modelsArray.Type() != engine.TypeArray {
		return nil, fmt.Errorf("no models in inventory")
	}

	// Find requested models
	allModels := modelsArray.(engine.ArrayValue).Elements()
	var foundModels []engine.ScriptValue

	for _, name := range modelNames {
		for _, modelVal := range allModels {
			if modelVal.Type() != engine.TypeObject {
				continue
			}
			modelObj := modelVal.(engine.ObjectValue)
			if nameVal, exists := modelObj.Fields()["name"]; exists && nameVal.Type() == engine.TypeString {
				if nameVal.(engine.StringValue).Value() == name {
					foundModels = append(foundModels, modelVal)
					break
				}
			}
		}
	}

	if len(foundModels) == 0 {
		return nil, fmt.Errorf("no matching models found")
	}

	// Generate comparison
	comparison := mia.generateComparison(foundModels)
	summary := mia.generateComparisonSummary(foundModels)

	result := map[string]engine.ScriptValue{
		"models":     engine.NewArrayValue(foundModels),
		"comparison": comparison,
		"summary":    summary,
	}

	return engine.NewObjectValue(result), nil
}

func (mia *ModelInfoAdapter) generateComparison(models []engine.ScriptValue) engine.ScriptValue {
	if len(models) < 2 {
		return engine.NewObjectValue(map[string]engine.ScriptValue{})
	}

	// Compare context windows
	contextComparison := mia.compareMetric(models, "contextWindow", true)             // higher is better
	pricingComparison := mia.compareMetric(models, "pricing.inputPer1kTokens", false) // lower is better

	result := map[string]engine.ScriptValue{
		"contextWindow": contextComparison,
		"pricing":       pricingComparison,
	}

	return engine.NewObjectValue(result)
}

func (mia *ModelInfoAdapter) compareMetric(models []engine.ScriptValue, metric string, higherIsBetter bool) engine.ScriptValue {
	var bestModel string
	var bestValue float64
	var values []engine.ScriptValue

	first := true
	for _, modelVal := range models {
		modelObj := modelVal.(engine.ObjectValue)
		fields := modelObj.Fields()

		nameVal := fields["name"]
		name := nameVal.(engine.StringValue).Value()

		var value float64
		if strings.Contains(metric, ".") {
			// Handle nested fields like "pricing.inputPer1kTokens"
			parts := strings.Split(metric, ".")
			if parentVal, exists := fields[parts[0]]; exists && parentVal.Type() == engine.TypeObject {
				parent := parentVal.(engine.ObjectValue).Fields()
				if childVal, exists := parent[parts[1]]; exists && childVal.Type() == engine.TypeNumber {
					value = childVal.(engine.NumberValue).Value()
				}
			}
		} else {
			if val, exists := fields[metric]; exists && val.Type() == engine.TypeNumber {
				value = val.(engine.NumberValue).Value()
			}
		}

		if first || (higherIsBetter && value > bestValue) || (!higherIsBetter && value < bestValue) {
			bestModel = name
			bestValue = value
			first = false
		}

		values = append(values, engine.NewObjectValue(map[string]engine.ScriptValue{
			"model": engine.NewStringValue(name),
			"value": engine.NewNumberValue(value),
		}))
	}

	return engine.NewObjectValue(map[string]engine.ScriptValue{
		"winner": engine.NewStringValue(bestModel),
		"values": engine.NewArrayValue(values),
	})
}

func (mia *ModelInfoAdapter) generateComparisonSummary(models []engine.ScriptValue) engine.ScriptValue {
	strengths := make(map[string][]string)

	for _, modelVal := range models {
		modelObj := modelVal.(engine.ObjectValue)
		fields := modelObj.Fields()

		nameVal := fields["name"]
		name := nameVal.(engine.StringValue).Value()

		var modelStrengths []string

		// Check context window
		if contextVal, exists := fields["contextWindow"]; exists && contextVal.Type() == engine.TypeNumber {
			context := contextVal.(engine.NumberValue).Value()
			if context > 50000 {
				modelStrengths = append(modelStrengths, "large context window")
			}
		}

		// Check capabilities
		if capVal, exists := fields["capabilities"]; exists && capVal.Type() == engine.TypeObject {
			capabilities := capVal.(engine.ObjectValue).Fields()
			if funcVal, exists := capabilities["functionCalling"]; exists && funcVal.Type() == engine.TypeBool && funcVal.(engine.BoolValue).Value() {
				modelStrengths = append(modelStrengths, "function calling")
			}
		}

		// Check pricing
		if pricingVal, exists := fields["pricing"]; exists && pricingVal.Type() == engine.TypeObject {
			pricing := pricingVal.(engine.ObjectValue).Fields()
			if inputPriceVal, exists := pricing["inputPer1kTokens"]; exists && inputPriceVal.Type() == engine.TypeNumber {
				price := inputPriceVal.(engine.NumberValue).Value()
				if price < 0.01 {
					modelStrengths = append(modelStrengths, "low cost")
				}
			}
		}

		if len(modelStrengths) > 0 {
			strengths[name] = modelStrengths
		}
	}

	// Convert to ScriptValue format
	strengthsMap := make(map[string]engine.ScriptValue)
	for model, strList := range strengths {
		strValues := make([]engine.ScriptValue, len(strList))
		for i, str := range strList {
			strValues[i] = engine.NewStringValue(str)
		}
		strengthsMap[model] = engine.NewArrayValue(strValues)
	}

	return engine.NewObjectValue(map[string]engine.ScriptValue{
		"strengths": engine.NewObjectValue(strengthsMap),
	})
}

func (mia *ModelInfoAdapter) estimateModelCost(inventory engine.ScriptValue, modelName string, inputTokens, outputTokens float64) (engine.ScriptValue, error) {
	inventoryObj, ok := inventory.(engine.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("invalid inventory format")
	}

	modelsArray, exists := inventoryObj.Fields()["models"]
	if !exists || modelsArray.Type() != engine.TypeArray {
		return nil, fmt.Errorf("no models in inventory")
	}

	// Find the model
	models := modelsArray.(engine.ArrayValue).Elements()
	for _, modelVal := range models {
		if modelVal.Type() != engine.TypeObject {
			continue
		}
		modelObj := modelVal.(engine.ObjectValue)
		fields := modelObj.Fields()

		nameVal, exists := fields["name"]
		if !exists || nameVal.Type() != engine.TypeString || nameVal.(engine.StringValue).Value() != modelName {
			continue
		}

		// Found the model, get pricing
		pricingVal, exists := fields["pricing"]
		if !exists || pricingVal.Type() != engine.TypeObject {
			return nil, fmt.Errorf("model %s has no pricing information", modelName)
		}

		pricing := pricingVal.(engine.ObjectValue).Fields()

		var inputPrice, outputPrice float64
		if inputPriceVal, exists := pricing["inputPer1kTokens"]; exists && inputPriceVal.Type() == engine.TypeNumber {
			inputPrice = inputPriceVal.(engine.NumberValue).Value()
		}
		if outputPriceVal, exists := pricing["outputPer1kTokens"]; exists && outputPriceVal.Type() == engine.TypeNumber {
			outputPrice = outputPriceVal.(engine.NumberValue).Value()
		}

		// Calculate costs
		inputCost := (inputTokens / 1000) * inputPrice
		outputCost := (outputTokens / 1000) * outputPrice
		totalCost := inputCost + outputCost

		result := map[string]engine.ScriptValue{
			"inputCost":  engine.NewNumberValue(inputCost),
			"outputCost": engine.NewNumberValue(outputCost),
			"totalCost":  engine.NewNumberValue(totalCost),
		}

		return engine.NewObjectValue(result), nil
	}

	return nil, fmt.Errorf("model not found: %s", modelName)
}

func (mia *ModelInfoAdapter) getBestModelForTask(inventory engine.ScriptValue, task string) (string, string, error) {
	inventoryObj, ok := inventory.(engine.ObjectValue)
	if !ok {
		return "", "", fmt.Errorf("invalid inventory format")
	}

	modelsArray, exists := inventoryObj.Fields()["models"]
	if !exists || modelsArray.Type() != engine.TypeArray {
		return "", "", fmt.Errorf("no models in inventory")
	}

	models := modelsArray.(engine.ArrayValue).Elements()

	// Task-specific logic
	switch task {
	case "function_calling":
		for _, modelVal := range models {
			if modelVal.Type() != engine.TypeObject {
				continue
			}
			modelObj := modelVal.(engine.ObjectValue)
			fields := modelObj.Fields()

			// Check if model supports function calling
			capVal, exists := fields["capabilities"]
			if !exists || capVal.Type() != engine.TypeObject {
				continue
			}

			capabilities := capVal.(engine.ObjectValue).Fields()
			if funcVal, exists := capabilities["functionCalling"]; exists && funcVal.Type() == engine.TypeBool && funcVal.(engine.BoolValue).Value() {
				nameVal := fields["name"]
				name := nameVal.(engine.StringValue).Value()
				return name, fmt.Sprintf("%s supports function calling capabilities", name), nil
			}
		}
		return "", "", fmt.Errorf("no models found with function calling capability")

	default:
		// For other tasks, return the first model with good capabilities
		for _, modelVal := range models {
			if modelVal.Type() != engine.TypeObject {
				continue
			}
			modelObj := modelVal.(engine.ObjectValue)
			fields := modelObj.Fields()

			nameVal := fields["name"]
			name := nameVal.(engine.StringValue).Value()
			return name, fmt.Sprintf("%s recommended for %s tasks", name, task), nil
		}
		return "", "", fmt.Errorf("no suitable models found")
	}
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (mia *ModelInfoAdapter) tableToMap(L *lua.LState, table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *gopherlua.LuaTypeConverter
			if mia.BridgeAdapter != nil {
				converter = mia.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			sv, err := converter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return result
}

// RegisterAsModule registers the adapter as a module in the module system
func (mia *ModelInfoAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if mia.GetBridge() != nil {
		bridgeMetadata = mia.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "ModelInfo Adapter",
			Description: "Model discovery and comparison functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},            // ModelInfo module has no dependencies by default
		LoadFunc:     mia.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (mia *ModelInfoAdapter) GetBridge() engine.Bridge {
	if mia.BridgeAdapter != nil {
		return mia.BridgeAdapter.GetBridge()
	}
	return nil
}

// GetMethods returns the available methods
func (mia *ModelInfoAdapter) GetMethods() []string {
	// Get base methods if bridge adapter exists
	var methods []string
	if mia.BridgeAdapter != nil {
		methods = mia.BridgeAdapter.GetMethods()
	}

	// Add modelinfo-specific methods if not already present
	modelinfoMethods := []string{
		"listModels", "fetchModelInventory", "getModel", "listRegistries",
		"getModelCapabilities", "findModelsByCapability",
		"suggestModel", "compareModels", "estimateCost", "getBestModelForTask",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range modelinfoMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
