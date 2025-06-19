// ABOUTME: Events bridge adapter that exposes go-llms event system functionality to Lua scripts
// ABOUTME: Provides event bus, subscription, emission, filtering, aggregation, recording, and replay operations

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// EventsAdapter specializes BridgeAdapter for event system functionality
type EventsAdapter struct {
	*gopherlua.BridgeAdapter

	// Optional related bridges for enhanced functionality
	storageBridge engine.Bridge // For event storage if separate from main bridge
}

// NewEventsAdapter creates a new events adapter
func NewEventsAdapter(bridge engine.Bridge) *EventsAdapter {
	// Create events adapter
	adapter := &EventsAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = gopherlua.NewBridgeAdapter(bridge)
	}

	// Add events-specific methods if not already present
	adapter.ensureEventMethods()

	return adapter
}

// NewEventsAdapterWithStorage creates a new events adapter with storage bridge
func NewEventsAdapterWithStorage(bridge engine.Bridge, storageBridge engine.Bridge) *EventsAdapter {
	adapter := NewEventsAdapter(bridge)
	adapter.storageBridge = storageBridge
	return adapter
}

// ensureEventMethods ensures event-specific methods are available
func (ea *EventsAdapter) ensureEventMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
	// In production, this could validate that expected event methods exist
}

// CreateLuaModule creates a Lua module with event-specific enhancements
func (ea *EventsAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if ea.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := ea.BridgeAdapter.CreateLuaModule()
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
		L.SetField(module, "_adapter", lua.LString("events"))
		L.SetField(module, "_version", lua.LString("2.0.0"))

		// Add event-specific enhancements
		ea.addEventEnhancements(L, module)

		// Add bus methods
		ea.addBusMethods(L, module)

		// Add filter methods
		ea.addFilterMethods(L, module)

		// Add recording methods
		ea.addRecordingMethods(L, module)

		// Add replay methods
		ea.addReplayMethods(L, module)

		// Add aggregation methods
		ea.addAggregationMethods(L, module)

		// Add convenience methods
		ea.addConvenienceMethods(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addEventEnhancements adds event-specific enhancements to the module
func (ea *EventsAdapter) addEventEnhancements(L *lua.LState, module *lua.LTable) {
	// Add event constants
	ea.addEventConstants(L, module)
}

// addBusMethods adds event bus-related methods
func (ea *EventsAdapter) addBusMethods(L *lua.LState, module *lua.LTable) {
	// Create bus namespace
	bus := L.NewTable()

	// publish method (alias for publishEvent)
	L.SetField(bus, "publish", L.NewFunction(func(L *lua.LState) int {
		event := L.CheckTable(1)

		eventMap := ea.tableToMap(L, event)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(eventMap)}

		_, err := ea.GetBridge().ExecuteMethod(ctx, "publishEvent", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// subscribe method (enhanced wrapper)
	L.SetField(bus, "subscribe", L.NewFunction(func(L *lua.LState) int {
		pattern := L.CheckString(1)
		handler := L.CheckFunction(2)

		// Store handler reference (simplified for mock)
		_ = handler

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(pattern),
			engine.NewCustomValue("function", handler),
		}

		result, err := ea.GetBridge().ExecuteMethod(ctx, "subscribe", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// unsubscribe method
	L.SetField(bus, "unsubscribe", L.NewFunction(func(L *lua.LState) int {
		subID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(subID)}

		_, err := ea.GetBridge().ExecuteMethod(ctx, "unsubscribe", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// Add bus namespace to module
	L.SetField(module, "bus", bus)
}

// addFilterMethods adds filter-related methods
func (ea *EventsAdapter) addFilterMethods(L *lua.LState, module *lua.LTable) {
	// Create filters namespace
	filters := L.NewTable()

	// create method
	L.SetField(filters, "create", L.NewFunction(func(L *lua.LState) int {
		filterConfig := L.CheckTable(1)

		filterMap := ea.tableToMap(L, filterConfig)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(filterMap)}

		result, err := ea.GetBridge().ExecuteMethod(ctx, "createFilter", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// createComposite method
	L.SetField(filters, "createComposite", L.NewFunction(func(L *lua.LState) int {
		filterIDs := L.CheckTable(1)
		operator := L.CheckString(2)

		// Convert filter IDs table to array
		var filterIDValues []engine.ScriptValue
		filterIDs.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTString {
				filterIDValues = append(filterIDValues, engine.NewStringValue(string(v.(lua.LString))))
			}
		})

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewArrayValue(filterIDValues),
			engine.NewStringValue(operator),
		}

		result, err := ea.GetBridge().ExecuteMethod(ctx, "createCompositeFilter", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add filters namespace to module
	L.SetField(module, "filters", filters)
}

// addRecordingMethods adds recording-related methods
func (ea *EventsAdapter) addRecordingMethods(L *lua.LState, module *lua.LTable) {
	// Create recording namespace
	recording := L.NewTable()

	// start method
	L.SetField(recording, "start", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := ea.GetBridge().ExecuteMethod(ctx, "startRecording", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// stop method
	L.SetField(recording, "stop", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := ea.GetBridge().ExecuteMethod(ctx, "stopRecording", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// isRecording method
	L.SetField(recording, "isRecording", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := ea.GetBridge().ExecuteMethod(ctx, "isRecording", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add recording namespace to module
	L.SetField(module, "recording", recording)
}

// addReplayMethods adds replay-related methods
func (ea *EventsAdapter) addReplayMethods(L *lua.LState, module *lua.LTable) {
	// Create replay namespace
	replay := L.NewTable()

	// start method (alias for replayEvents)
	L.SetField(replay, "start", L.NewFunction(func(L *lua.LState) int {
		query := L.CheckTable(1)
		options := L.OptTable(2, L.NewTable())

		queryMap := ea.tableToMap(L, query)
		optionsMap := ea.tableToMap(L, options)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(queryMap),
			engine.NewObjectValue(optionsMap),
		}

		_, err := ea.GetBridge().ExecuteMethod(ctx, "replayEvents", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// pause method
	L.SetField(replay, "pause", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := ea.GetBridge().ExecuteMethod(ctx, "pauseReplay", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// resume method
	L.SetField(replay, "resume", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := ea.GetBridge().ExecuteMethod(ctx, "resumeReplay", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// stop method
	L.SetField(replay, "stop", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		_, err := ea.GetBridge().ExecuteMethod(ctx, "stopReplay", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// Add replay namespace to module
	L.SetField(module, "replay", replay)
}

// addAggregationMethods adds aggregation-related methods
func (ea *EventsAdapter) addAggregationMethods(L *lua.LState, module *lua.LTable) {
	// Create aggregation namespace
	aggregation := L.NewTable()

	// create method
	L.SetField(aggregation, "create", L.NewFunction(func(L *lua.LState) int {
		aggType := L.CheckString(1)
		config := L.CheckTable(2)

		configMap := ea.tableToMap(L, config)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(aggType),
			engine.NewObjectValue(configMap),
		}

		result, err := ea.GetBridge().ExecuteMethod(ctx, "createAggregator", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// getData method
	L.SetField(aggregation, "getData", L.NewFunction(func(L *lua.LState) int {
		aggID := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(aggID)}

		result, err := ea.GetBridge().ExecuteMethod(ctx, "getAggregatedData", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Add aggregation namespace to module
	L.SetField(module, "aggregation", aggregation)
}

// addConvenienceMethods adds convenience methods to the module
func (ea *EventsAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// Add correlateEvents method if not already present
	if module.RawGetString("correlateEvents") == lua.LNil {
		L.SetField(module, "correlateEvents", L.NewFunction(func(L *lua.LState) int {
			config := L.CheckTable(1)

			configMap := ea.tableToMap(L, config)

			ctx := context.Background()
			args := []engine.ScriptValue{engine.NewObjectValue(configMap)}

			result, err := ea.GetBridge().ExecuteMethod(ctx, "correlateEvents", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			luaResult, err := ea.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))
	}
}

// addEventConstants adds event-related constants to the module
func (ea *EventsAdapter) addEventConstants(L *lua.LState, module *lua.LTable) {
	// Add event types
	eventTypes := L.NewTable()
	L.SetField(eventTypes, "USER_ACTION", lua.LString("user_action"))
	L.SetField(eventTypes, "SYSTEM_EVENT", lua.LString("system_event"))
	L.SetField(eventTypes, "ERROR_EVENT", lua.LString("error_event"))
	L.SetField(eventTypes, "METRIC_EVENT", lua.LString("metric_event"))
	L.SetField(module, "EVENT_TYPES", eventTypes)

	// Add filter types
	filterTypes := L.NewTable()
	L.SetField(filterTypes, "PATTERN", lua.LString("pattern"))
	L.SetField(filterTypes, "TYPE", lua.LString("type"))
	L.SetField(filterTypes, "TIME_RANGE", lua.LString("time_range"))
	L.SetField(module, "FILTER_TYPES", filterTypes)

	// Add aggregation types
	aggTypes := L.NewTable()
	L.SetField(aggTypes, "COUNT", lua.LString("count"))
	L.SetField(aggTypes, "SUM", lua.LString("sum"))
	L.SetField(aggTypes, "AVERAGE", lua.LString("average"))
	L.SetField(aggTypes, "RATE", lua.LString("rate"))
	L.SetField(module, "AGGREGATION_TYPES", aggTypes)
}

// WrapMethod wraps a bridge method with event-specific handling
func (ea *EventsAdapter) WrapMethod(methodName string) lua.LGFunction {
	// Get base wrapped method if available
	if ea.BridgeAdapter != nil {
		baseWrapped := ea.BridgeAdapter.WrapMethod(methodName)

		// Add event-specific handling for certain methods
		switch methodName {
		case "publishEvent", "subscribe", "unsubscribe":
			return ea.wrapEventOperation(methodName, baseWrapped)
		default:
			return baseWrapped
		}
	}

	// Return a simple function that returns an error when no bridge is available
	return func(L *lua.LState) int {
		L.Push(lua.LNil)
		L.Push(lua.LString("method not available - no bridge adapter"))
		return 2
	}
}

// wrapEventOperation adds event operation handling
func (ea *EventsAdapter) wrapEventOperation(_ string, baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least one parameter is provided for event operations
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("event operation requires parameters"))
			return 2
		}

		return baseFn(L)
	}
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (ea *EventsAdapter) tableToMap(L *lua.LState, table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *gopherlua.LuaTypeConverter
			if ea.BridgeAdapter != nil {
				converter = ea.GetTypeConverter()
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
func (ea *EventsAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if ea.GetBridge() != nil {
		bridgeMetadata = ea.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "Events Adapter",
			Description: "Event system functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // Events module has no dependencies by default
		LoadFunc:     ea.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (ea *EventsAdapter) GetBridge() engine.Bridge {
	if ea.BridgeAdapter != nil {
		return ea.BridgeAdapter.GetBridge()
	}
	return nil
}

// GetMethods returns the available methods
func (ea *EventsAdapter) GetMethods() []string {
	// Get base methods if bridge adapter exists
	var methods []string
	if ea.BridgeAdapter != nil {
		methods = ea.BridgeAdapter.GetMethods()
	}

	// Add event-specific methods if not already present
	eventMethods := []string{
		"publishEvent", "subscribe", "subscribeWithFilter", "unsubscribe",
		"storeEvent", "queryEvents", "getEventHistory",
		"createFilter", "createCompositeFilter",
		"replayEvents", "pauseReplay", "resumeReplay", "stopReplay",
		"serializeEvent", "deserializeEvent",
		"createAggregator", "getAggregatedData",
		"startRecording", "stopRecording", "isRecording",
		"getSubscriptionCount", "getSubscriptionInfo",
		"correlateEvents",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range eventMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
