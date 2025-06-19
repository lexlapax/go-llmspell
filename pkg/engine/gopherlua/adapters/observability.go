// ABOUTME: Observability bridge adapter that exposes go-llms guardrails, metrics, and tracing to Lua scripts
// ABOUTME: Provides safety system configuration, performance monitoring, and distributed tracing capabilities

package adapters

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// ObservabilityAdapter bridges go-llms observability functionality to Lua
type ObservabilityAdapter struct {
	guardrailsBridge engine.Bridge
	metricsBridge    engine.Bridge
	tracingBridge    engine.Bridge
}

// NewObservabilityAdapter creates a new observability adapter
func NewObservabilityAdapter(guardrailsBridge, metricsBridge, tracingBridge engine.Bridge) *ObservabilityAdapter {
	return &ObservabilityAdapter{
		guardrailsBridge: guardrailsBridge,
		metricsBridge:    metricsBridge,
		tracingBridge:    tracingBridge,
	}
}

// GetAdapterName returns the adapter name
func (oa *ObservabilityAdapter) GetAdapterName() string {
	return "observability"
}

// GetBridge returns the primary bridge (guardrails)
func (oa *ObservabilityAdapter) GetBridge() engine.Bridge {
	return oa.guardrailsBridge
}

// CreateLuaModule creates a Lua module for observability
func (oa *ObservabilityAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Set adapter info
		L.SetField(module, "_adapter", lua.LString("observability"))
		L.SetField(module, "_version", lua.LString("1.0.0"))

		// Create guardrails namespace
		guardrails := L.NewTable()

		// Compliance levels
		L.SetField(guardrails, "COMPLIANCE_STRICT", lua.LString("strict"))
		L.SetField(guardrails, "COMPLIANCE_MODERATE", lua.LString("moderate"))
		L.SetField(guardrails, "COMPLIANCE_RELAXED", lua.LString("relaxed"))

		// Guardrails methods
		L.SetField(guardrails, "enable", L.NewFunction(oa.enableGuardrails))
		L.SetField(guardrails, "validateContent", L.NewFunction(oa.validateContent))
		L.SetField(guardrails, "addBehavioralConstraint", L.NewFunction(oa.addBehavioralConstraint))
		L.SetField(guardrails, "checkCompliance", L.NewFunction(oa.checkCompliance))

		L.SetField(module, "guardrails", guardrails)

		// Create metrics namespace
		metrics := L.NewTable()

		// Metric types
		L.SetField(metrics, "COUNTER", lua.LString("counter"))
		L.SetField(metrics, "GAUGE", lua.LString("gauge"))
		L.SetField(metrics, "TIMER", lua.LString("timer"))

		// Metrics methods
		L.SetField(metrics, "createCounter", L.NewFunction(oa.createCounter))
		L.SetField(metrics, "createGauge", L.NewFunction(oa.createGauge))
		L.SetField(metrics, "createTimer", L.NewFunction(oa.createTimer))
		L.SetField(metrics, "recordMetric", L.NewFunction(oa.recordMetric))
		L.SetField(metrics, "getMetrics", L.NewFunction(oa.getMetrics))
		L.SetField(metrics, "builder", L.NewFunction(oa.createMetricBuilder))

		L.SetField(module, "metrics", metrics)

		// Create tracing namespace
		tracing := L.NewTable()

		// Tracing methods
		L.SetField(tracing, "startSpan", L.NewFunction(oa.startSpan))
		L.SetField(tracing, "addSpanEvent", L.NewFunction(oa.addSpanEvent))
		L.SetField(tracing, "setSpanAttribute", L.NewFunction(oa.setSpanAttribute))
		L.SetField(tracing, "endSpan", L.NewFunction(oa.endSpan))
		L.SetField(tracing, "getCurrentSpan", L.NewFunction(oa.getCurrentSpan))

		L.SetField(module, "tracing", tracing)

		// Direct methods at module level
		L.SetField(module, "enableGuardrails", L.NewFunction(oa.enableGuardrails))
		L.SetField(module, "validateContent", L.NewFunction(oa.validateContent))
		L.SetField(module, "addBehavioralConstraint", L.NewFunction(oa.addBehavioralConstraint))
		L.SetField(module, "checkCompliance", L.NewFunction(oa.checkCompliance))
		L.SetField(module, "createCounter", L.NewFunction(oa.createCounter))
		L.SetField(module, "createGauge", L.NewFunction(oa.createGauge))
		L.SetField(module, "createTimer", L.NewFunction(oa.createTimer))
		L.SetField(module, "recordMetric", L.NewFunction(oa.recordMetric))
		L.SetField(module, "getMetrics", L.NewFunction(oa.getMetrics))
		L.SetField(module, "startSpan", L.NewFunction(oa.startSpan))
		L.SetField(module, "addSpanEvent", L.NewFunction(oa.addSpanEvent))
		L.SetField(module, "setSpanAttribute", L.NewFunction(oa.setSpanAttribute))
		L.SetField(module, "endSpan", L.NewFunction(oa.endSpan))

		L.Push(module)
		return 1
	}
}

// GetMethods returns available adapter methods
func (oa *ObservabilityAdapter) GetMethods() []string {
	return []string{
		// Guardrails methods
		"enableGuardrails", "validateContent", "addBehavioralConstraint", "checkCompliance",
		// Metrics methods
		"createCounter", "createGauge", "createTimer", "recordMetric", "getMetrics",
		// Tracing methods
		"startSpan", "addSpanEvent", "setSpanAttribute", "endSpan",
	}
}

// RegisterAsModule registers the adapter as a module in the module system
func (oa *ObservabilityAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if oa.guardrailsBridge != nil {
		bridgeMetadata = oa.guardrailsBridge.GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "Observability Adapter",
			Description: "Guardrails, metrics, and tracing functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // Observability module has no dependencies by default
		LoadFunc:     oa.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// Guardrails methods

func (oa *ObservabilityAdapter) enableGuardrails(L *lua.LState) int {
	if oa.guardrailsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("guardrails bridge not initialized"))
		return 2
	}

	config := L.CheckTable(1)

	args := []engine.ScriptValue{
		luaToScriptValue(config),
	}

	result, err := oa.guardrailsBridge.ExecuteMethod(context.Background(), "enableGuardrails", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) validateContent(L *lua.LState) int {
	if oa.guardrailsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("guardrails bridge not initialized"))
		return 2
	}

	content := L.CheckString(1)
	contentType := L.CheckString(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(content),
		engine.NewStringValue(contentType),
	}

	result, err := oa.guardrailsBridge.ExecuteMethod(context.Background(), "validateContent", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) addBehavioralConstraint(L *lua.LState) int {
	if oa.guardrailsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("guardrails bridge not initialized"))
		return 2
	}

	constraint := L.CheckTable(1)

	args := []engine.ScriptValue{
		luaToScriptValue(constraint),
	}

	result, err := oa.guardrailsBridge.ExecuteMethod(context.Background(), "addBehavioralConstraint", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) checkCompliance(L *lua.LState) int {
	if oa.guardrailsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("guardrails bridge not initialized"))
		return 2
	}

	request := L.CheckTable(1)

	args := []engine.ScriptValue{
		luaToScriptValue(request),
	}

	result, err := oa.guardrailsBridge.ExecuteMethod(context.Background(), "checkCompliance", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// Metrics methods

func (oa *ObservabilityAdapter) createCounter(L *lua.LState) int {
	if oa.metricsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("metrics bridge not initialized"))
		return 2
	}

	name := L.CheckString(1)
	labels := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(name),
		luaToScriptValue(labels),
	}

	result, err := oa.metricsBridge.ExecuteMethod(context.Background(), "createCounter", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) createGauge(L *lua.LState) int {
	if oa.metricsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("metrics bridge not initialized"))
		return 2
	}

	name := L.CheckString(1)
	labels := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(name),
		luaToScriptValue(labels),
	}

	result, err := oa.metricsBridge.ExecuteMethod(context.Background(), "createGauge", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) createTimer(L *lua.LState) int {
	if oa.metricsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("metrics bridge not initialized"))
		return 2
	}

	name := L.CheckString(1)
	labels := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(name),
		luaToScriptValue(labels),
	}

	result, err := oa.metricsBridge.ExecuteMethod(context.Background(), "createTimer", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) recordMetric(L *lua.LState) int {
	if oa.metricsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("metrics bridge not initialized"))
		return 2
	}

	name := L.CheckString(1)
	value := L.CheckNumber(2)

	// Optional labels
	var labels engine.ScriptValue
	if L.GetTop() >= 3 {
		labelsTable := L.CheckTable(3)
		labels = luaToScriptValue(labelsTable)
	} else {
		labels = engine.NewObjectValue(map[string]engine.ScriptValue{})
	}

	args := []engine.ScriptValue{
		engine.NewStringValue(name),
		engine.NewNumberValue(float64(value)),
		labels,
	}

	result, err := oa.metricsBridge.ExecuteMethod(context.Background(), "recordMetric", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) getMetrics(L *lua.LState) int {
	if oa.metricsBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("metrics bridge not initialized"))
		return 2
	}

	result, err := oa.metricsBridge.ExecuteMethod(context.Background(), "getMetrics", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

// Metric builder support
func (oa *ObservabilityAdapter) createMetricBuilder(L *lua.LState) int {
	name := L.CheckString(1)

	// Create builder table with methods
	builder := L.NewTable()

	// Store metric definition
	metricDef := L.NewTable()
	L.SetField(metricDef, "name", lua.LString(name))
	L.SetField(metricDef, "type", lua.LString("counter")) // Default type
	L.SetField(metricDef, "labels", L.NewTable())
	L.SetField(builder, "_metricDef", metricDef)
	L.SetField(builder, "_adapter", L.NewUserData()) // Store adapter reference

	// Builder methods
	L.SetField(builder, "withType", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		metricType := L.CheckString(2)

		metricDef := L.GetField(self, "_metricDef").(*lua.LTable)
		L.SetField(metricDef, "type", lua.LString(metricType))

		L.Push(self) // Return self for chaining
		return 1
	}))

	L.SetField(builder, "withLabels", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		labels := L.CheckTable(2)

		metricDef := L.GetField(self, "_metricDef").(*lua.LTable)
		L.SetField(metricDef, "labels", labels)

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "withDescription", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		description := L.CheckString(2)

		metricDef := L.GetField(self, "_metricDef").(*lua.LTable)
		L.SetField(metricDef, "description", lua.LString(description))

		L.Push(self)
		return 1
	}))

	L.SetField(builder, "build", L.NewFunction(func(L *lua.LState) int {
		self := L.CheckTable(1)
		metricDef := L.GetField(self, "_metricDef").(*lua.LTable)

		// Get metric properties
		name := string(L.GetField(metricDef, "name").(lua.LString))
		metricType := string(L.GetField(metricDef, "type").(lua.LString))
		labels := L.GetField(metricDef, "labels").(*lua.LTable)

		// Create the metric based on type
		var methodName string
		switch metricType {
		case "counter":
			methodName = "createCounter"
		case "gauge":
			methodName = "createGauge"
		case "timer":
			methodName = "createTimer"
		default:
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("unknown metric type: %s", metricType)))
			return 2
		}

		if oa.metricsBridge == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("metrics bridge not initialized"))
			return 2
		}

		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			luaToScriptValue(labels),
		}

		result, err := oa.metricsBridge.ExecuteMethod(context.Background(), methodName, args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(scriptValueToLua(L, result))
		return 1
	}))

	L.Push(builder)
	return 1
}

// Tracing methods

func (oa *ObservabilityAdapter) startSpan(L *lua.LState) int {
	if oa.tracingBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("tracing bridge not initialized"))
		return 2
	}

	name := L.CheckString(1)
	options := L.CheckTable(2)

	args := []engine.ScriptValue{
		engine.NewStringValue(name),
		luaToScriptValue(options),
	}

	result, err := oa.tracingBridge.ExecuteMethod(context.Background(), "startSpan", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) addSpanEvent(L *lua.LState) int {
	if oa.tracingBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("tracing bridge not initialized"))
		return 2
	}

	spanId := L.CheckString(1)
	eventName := L.CheckString(2)
	attributes := L.CheckTable(3)

	args := []engine.ScriptValue{
		engine.NewStringValue(spanId),
		engine.NewStringValue(eventName),
		luaToScriptValue(attributes),
	}

	result, err := oa.tracingBridge.ExecuteMethod(context.Background(), "addSpanEvent", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) setSpanAttribute(L *lua.LState) int {
	if oa.tracingBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("tracing bridge not initialized"))
		return 2
	}

	spanId := L.CheckString(1)
	key := L.CheckString(2)
	value := L.Get(3)

	args := []engine.ScriptValue{
		engine.NewStringValue(spanId),
		engine.NewStringValue(key),
		luaToScriptValue(value),
	}

	result, err := oa.tracingBridge.ExecuteMethod(context.Background(), "setSpanAttribute", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) endSpan(L *lua.LState) int {
	if oa.tracingBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("tracing bridge not initialized"))
		return 2
	}

	spanId := L.CheckString(1)

	args := []engine.ScriptValue{
		engine.NewStringValue(spanId),
	}

	result, err := oa.tracingBridge.ExecuteMethod(context.Background(), "endSpan", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}

func (oa *ObservabilityAdapter) getCurrentSpan(L *lua.LState) int {
	if oa.tracingBridge == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("tracing bridge not initialized"))
		return 2
	}

	result, err := oa.tracingBridge.ExecuteMethod(context.Background(), "getCurrentSpan", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(scriptValueToLua(L, result))
	L.Push(lua.LNil)
	return 2
}
