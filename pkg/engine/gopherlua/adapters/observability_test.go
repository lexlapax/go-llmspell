// ABOUTME: Tests for Observability bridge adapter that exposes go-llms guardrails, metrics, and tracing to Lua scripts
// ABOUTME: Validates safety system configuration, metric recording, and distributed tracing capabilities

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

func TestObservabilityAdapter_Creation(t *testing.T) {
	t.Run("create_observability_adapter", func(t *testing.T) {
		// Create observability bridges mock
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Guardrails Bridge",
				Version:     "1.0.0",
				Description: "Safety system with content filtering",
			})

		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Metrics Bridge",
				Version:     "1.0.0",
				Description: "Performance monitoring system",
			})

		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Tracing Bridge",
				Version:     "1.0.0",
				Description: "Distributed tracing system",
			})

		// Create adapter
		adapter := NewObservabilityAdapter(guardrailsBridge, metricsBridge, tracingBridge)
		require.NotNil(t, adapter)

		// Should have observability-specific methods
		methods := adapter.GetMethods()

		// Legacy guardrails methods
		assert.Contains(t, methods, "enableGuardrails")
		assert.Contains(t, methods, "validateContent")
		assert.Contains(t, methods, "addBehavioralConstraint")
		assert.Contains(t, methods, "checkCompliance")

		// Legacy metrics methods
		assert.Contains(t, methods, "createCounter")
		assert.Contains(t, methods, "createGauge")
		assert.Contains(t, methods, "createTimer")
		assert.Contains(t, methods, "recordMetric")
		assert.Contains(t, methods, "getMetrics")

		// Legacy tracing methods
		assert.Contains(t, methods, "startSpan")
		assert.Contains(t, methods, "addSpanEvent")
		assert.Contains(t, methods, "setSpanAttribute")
		assert.Contains(t, methods, "endSpan")

		// Flattened guardrails methods
		assert.Contains(t, methods, "guardrailsRegisterRule")
		assert.Contains(t, methods, "guardrailsCheck")
		assert.Contains(t, methods, "guardrailsEnableRule")
		assert.Contains(t, methods, "guardrailsDisableRule")

		// Flattened metrics methods
		assert.Contains(t, methods, "metricsIncrement")
		assert.Contains(t, methods, "metricsGauge")
		assert.Contains(t, methods, "metricsHistogram")
		assert.Contains(t, methods, "metricsGetAll")

		// Flattened tracing methods
		assert.Contains(t, methods, "tracingStartSpan")
		assert.Contains(t, methods, "tracingEndSpan")
		assert.Contains(t, methods, "tracingAddAttribute")
		assert.Contains(t, methods, "tracingGetTrace")
	})

	t.Run("observability_module_structure", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true)
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true)
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true)

		adapter := NewObservabilityAdapter(guardrailsBridge, metricsBridge, tracingBridge)

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
		L.SetGlobal("observability", module)

		// Test module structure
		err = L.DoString(`
			-- Check basic module properties
			assert(observability._adapter == "observability", "should have correct adapter name")
			assert(observability._version == "1.0.0", "should have correct version")
			
			-- Check namespaces exist (backward compatibility)
			assert(type(observability.guardrails) == "table", "guardrails namespace should exist")
			assert(type(observability.metrics) == "table", "metrics namespace should exist")
			assert(type(observability.tracing) == "table", "tracing namespace should exist")
			
			-- Check flattened methods exist
			assert(type(observability.guardrailsRegisterRule) == "function", "guardrailsRegisterRule should exist")
			assert(type(observability.guardrailsCheck) == "function", "guardrailsCheck should exist")
			assert(type(observability.guardrailsEnableRule) == "function", "guardrailsEnableRule should exist")
			assert(type(observability.guardrailsDisableRule) == "function", "guardrailsDisableRule should exist")
			assert(type(observability.metricsIncrement) == "function", "metricsIncrement should exist")
			assert(type(observability.metricsGauge) == "function", "metricsGauge should exist")
			assert(type(observability.metricsHistogram) == "function", "metricsHistogram should exist")
			assert(type(observability.metricsGetAll) == "function", "metricsGetAll should exist")
			assert(type(observability.tracingStartSpan) == "function", "tracingStartSpan should exist")
			assert(type(observability.tracingEndSpan) == "function", "tracingEndSpan should exist")
			assert(type(observability.tracingAddAttribute) == "function", "tracingAddAttribute should exist")
			assert(type(observability.tracingGetTrace) == "function", "tracingGetTrace should exist")
			
			-- Check metric types
			assert(observability.metrics.COUNTER == "counter", "should have counter type")
			assert(observability.metrics.GAUGE == "gauge", "should have gauge type")
			assert(observability.metrics.TIMER == "timer", "should have timer type")
			
			-- Check compliance levels
			assert(observability.guardrails.COMPLIANCE_STRICT == "strict", "should have strict compliance")
			assert(observability.guardrails.COMPLIANCE_MODERATE == "moderate", "should have moderate compliance")
			assert(observability.guardrails.COMPLIANCE_RELAXED == "relaxed", "should have relaxed compliance")
		`)
		assert.NoError(t, err)
	})
}

func TestObservabilityAdapter_Guardrails(t *testing.T) {
	t.Run("enable_guardrails", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("enableGuardrails", engine.MethodInfo{
				Name: "enableGuardrails",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				config := args[0].(engine.ObjectValue).Fields()

				// Mock enabling guardrails
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"enabled": engine.NewBoolValue(true),
					"level":   config["level"],
					"rules":   engine.NewNumberValue(5),
				}), nil
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Enable guardrails with configuration
			local result, err = obs.guardrails.enable({
				level = "strict",
				contentFilter = true,
				behavioralLimits = true
			})
			assert(err == nil, "should not error")
			assert(result.enabled == true, "should be enabled")
			assert(result.level == "strict", "should have strict level")
			assert(result.rules == 5, "should have 5 rules")
		`)
		assert.NoError(t, err)
	})

	t.Run("validate_content", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("validateContent", engine.MethodInfo{
				Name: "validateContent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				content := args[0].(engine.StringValue).Value()
				contentType := args[1].(engine.StringValue).Value()

				// Mock content validation
				isValid := !contains(content, "inappropriate")

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid": engine.NewBoolValue(isValid),
					"score": engine.NewNumberValue(0.95),
					"type":  engine.NewStringValue(contentType),
					"violations": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewStringValue("none"),
					}),
				}), nil
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Validate safe content
			local result, err = obs.guardrails.validateContent("Hello world", "text")
			assert(err == nil, "should not error")
			assert(result.valid == true, "should be valid")
			assert(result.score == 0.95, "should have high score")
			assert(result.type == "text", "should have text type")
			
			-- Validate inappropriate content
			local invalid, err2 = obs.guardrails.validateContent("inappropriate content", "text")
			assert(err2 == nil, "should not error")
			assert(invalid.valid == false, "should be invalid")
		`)
		assert.NoError(t, err)
	})

	t.Run("behavioral_constraints", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("addBehavioralConstraint", engine.MethodInfo{
				Name: "addBehavioralConstraint",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				constraint := args[0].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":     engine.NewStringValue("constraint-123"),
					"name":   constraint["name"],
					"type":   constraint["type"],
					"active": engine.NewBoolValue(true),
				}), nil
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Add behavioral constraint
			local constraint, err = obs.guardrails.addBehavioralConstraint({
				name = "rate_limit",
				type = "frequency",
				limit = 100,
				window = "1m"
			})
			assert(err == nil, "should not error")
			assert(constraint.id == "constraint-123", "should have ID")
			assert(constraint.name == "rate_limit", "should have name")
			assert(constraint.active == true, "should be active")
		`)
		assert.NoError(t, err)
	})

	t.Run("check_compliance", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("checkCompliance", engine.MethodInfo{
				Name: "checkCompliance",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Validate request was passed
				_ = args[0].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"compliant": engine.NewBoolValue(true),
					"level":     engine.NewStringValue("high"),
					"details": engine.NewObjectValue(map[string]engine.ScriptValue{
						"contentCheck": engine.NewBoolValue(true),
						"rateCheck":    engine.NewBoolValue(true),
						"policyCheck":  engine.NewBoolValue(true),
					}),
				}), nil
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Check compliance
			local result, err = obs.guardrails.checkCompliance({
				action = "generate",
				content = "test content",
				userId = "user-123"
			})
			assert(err == nil, "should not error")
			assert(result.compliant == true, "should be compliant")
			assert(result.level == "high", "should have high compliance level")
			assert(result.details.contentCheck == true, "content should pass")
			assert(result.details.rateCheck == true, "rate should pass")
			assert(result.details.policyCheck == true, "policy should pass")
		`)
		assert.NoError(t, err)
	})
}

func TestObservabilityAdapter_Metrics(t *testing.T) {
	t.Run("create_counter", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("createCounter", engine.MethodInfo{
				Name: "createCounter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[0].(engine.StringValue).Value()
				labels := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":     engine.NewStringValue("counter-" + name),
					"name":   engine.NewStringValue(name),
					"type":   engine.NewStringValue("counter"),
					"labels": engine.NewObjectValue(labels),
					"value":  engine.NewNumberValue(0),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Create counter
			local counter, err = obs.metrics.createCounter("requests_total", {
				service = "api",
				method = "GET"
			})
			assert(err == nil, "should not error")
			assert(counter.name == "requests_total", "should have name")
			assert(counter.type == "counter", "should be counter type")
			assert(counter.labels.service == "api", "should have service label")
			assert(counter.value == 0, "should start at 0")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_gauge", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("createGauge", engine.MethodInfo{
				Name: "createGauge",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[0].(engine.StringValue).Value()
				labels := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":     engine.NewStringValue("gauge-" + name),
					"name":   engine.NewStringValue(name),
					"type":   engine.NewStringValue("gauge"),
					"labels": engine.NewObjectValue(labels),
					"value":  engine.NewNumberValue(0),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Create gauge
			local gauge, err = obs.metrics.createGauge("memory_usage", {
				service = "api",
				unit = "bytes"
			})
			assert(err == nil, "should not error")
			assert(gauge.name == "memory_usage", "should have name")
			assert(gauge.type == "gauge", "should be gauge type")
			assert(gauge.labels.unit == "bytes", "should have unit label")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_timer", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("createTimer", engine.MethodInfo{
				Name: "createTimer",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[0].(engine.StringValue).Value()
				labels := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":     engine.NewStringValue("timer-" + name),
					"name":   engine.NewStringValue(name),
					"type":   engine.NewStringValue("timer"),
					"labels": engine.NewObjectValue(labels),
					"count":  engine.NewNumberValue(0),
					"mean":   engine.NewNumberValue(0),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Create timer
			local timer, err = obs.metrics.createTimer("request_duration", {
				service = "api",
				endpoint = "/users"
			})
			assert(err == nil, "should not error")
			assert(timer.name == "request_duration", "should have name")
			assert(timer.type == "timer", "should be timer type")
			assert(timer.labels.endpoint == "/users", "should have endpoint label")
		`)
		assert.NoError(t, err)
	})

	t.Run("record_metric", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("recordMetric", engine.MethodInfo{
				Name: "recordMetric",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[0].(engine.StringValue).Value()
				value := args[1].(engine.NumberValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"recorded":  engine.NewBoolValue(true),
					"metric":    engine.NewStringValue(name),
					"value":     engine.NewNumberValue(value),
					"timestamp": engine.NewStringValue("2024-01-01T00:00:00Z"),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Record metric value
			local result, err = obs.metrics.recordMetric("requests_total", 42, {
				service = "api"
			})
			assert(err == nil, "should not error")
			assert(result.recorded == true, "should be recorded")
			assert(result.metric == "requests_total", "should have metric name")
			assert(result.value == 42, "should have value")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_metrics", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("getMetrics", engine.MethodInfo{
				Name: "getMetrics",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"counters": engine.NewObjectValue(map[string]engine.ScriptValue{
						"requests_total": engine.NewNumberValue(1000),
					}),
					"gauges": engine.NewObjectValue(map[string]engine.ScriptValue{
						"memory_usage": engine.NewNumberValue(1048576),
					}),
					"timers": engine.NewObjectValue(map[string]engine.ScriptValue{
						"request_duration": engine.NewObjectValue(map[string]engine.ScriptValue{
							"count": engine.NewNumberValue(100),
							"mean":  engine.NewNumberValue(50.5),
						}),
					}),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Get all metrics
			local metrics, err = obs.metrics.getMetrics()
			assert(err == nil, "should not error")
			assert(metrics.counters.requests_total == 1000, "should have request count")
			assert(metrics.gauges.memory_usage == 1048576, "should have memory usage")
			assert(metrics.timers.request_duration.count == 100, "should have timer count")
			assert(metrics.timers.request_duration.mean == 50.5, "should have timer mean")
		`)
		assert.NoError(t, err)
	})
}

func TestObservabilityAdapter_Tracing(t *testing.T) {
	t.Run("start_span", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("startSpan", engine.MethodInfo{
				Name: "startSpan",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				name := args[0].(engine.StringValue).Value()
				options := args[1].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":        engine.NewStringValue("span-123"),
					"name":      engine.NewStringValue(name),
					"traceId":   engine.NewStringValue("trace-456"),
					"parentId":  options["parentId"],
					"startTime": engine.NewStringValue("2024-01-01T00:00:00Z"),
					"status":    engine.NewStringValue("active"),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Start a span
			local span, err = obs.tracing.startSpan("processRequest", {
				kind = "server",
				attributes = {
					["http.method"] = "GET",
					["http.url"] = "/api/users"
				}
			})
			assert(err == nil, "should not error")
			assert(span.id == "span-123", "should have span ID")
			assert(span.name == "processRequest", "should have name")
			assert(span.traceId == "trace-456", "should have trace ID")
			assert(span.status == "active", "should be active")
		`)
		assert.NoError(t, err)
	})

	t.Run("add_span_event", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("addSpanEvent", engine.MethodInfo{
				Name: "addSpanEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				spanId := args[0].(engine.StringValue).Value()
				eventName := args[1].(engine.StringValue).Value()
				attributes := args[2].(engine.ObjectValue).Fields()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"added":      engine.NewBoolValue(true),
					"spanId":     engine.NewStringValue(spanId),
					"eventName":  engine.NewStringValue(eventName),
					"timestamp":  engine.NewStringValue("2024-01-01T00:00:01Z"),
					"attributes": engine.NewObjectValue(attributes),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Add span event
			local result, err = obs.tracing.addSpanEvent("span-123", "cache_hit", {
				key = "user:123",
				ttl = 3600
			})
			assert(err == nil, "should not error")
			assert(result.added == true, "should be added")
			assert(result.eventName == "cache_hit", "should have event name")
			assert(result.attributes.key == "user:123", "should have key attribute")
		`)
		assert.NoError(t, err)
	})

	t.Run("set_span_attribute", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("setSpanAttribute", engine.MethodInfo{
				Name: "setSpanAttribute",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				spanId := args[0].(engine.StringValue).Value()
				key := args[1].(engine.StringValue).Value()
				value := args[2]

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"set":    engine.NewBoolValue(true),
					"spanId": engine.NewStringValue(spanId),
					"key":    engine.NewStringValue(key),
					"value":  value,
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Set span attribute
			local result, err = obs.tracing.setSpanAttribute("span-123", "user.id", "user-456")
			assert(err == nil, "should not error")
			assert(result.set == true, "should be set")
			assert(result.key == "user.id", "should have key")
			assert(result.value == "user-456", "should have value")
			
			-- Set numeric attribute
			local numResult, numErr = obs.tracing.setSpanAttribute("span-123", "response.size", 1024)
			assert(numErr == nil, "should not error")
			assert(numResult.value == 1024, "should have numeric value")
		`)
		assert.NoError(t, err)
	})

	t.Run("end_span", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("endSpan", engine.MethodInfo{
				Name: "endSpan",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				spanId := args[0].(engine.StringValue).Value()

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"ended":    engine.NewBoolValue(true),
					"spanId":   engine.NewStringValue(spanId),
					"endTime":  engine.NewStringValue("2024-01-01T00:00:10Z"),
					"duration": engine.NewNumberValue(10000), // milliseconds
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- End span
			local result, err = obs.tracing.endSpan("span-123")
			assert(err == nil, "should not error")
			assert(result.ended == true, "should be ended")
			assert(result.spanId == "span-123", "should have span ID")
			assert(result.duration == 10000, "should have duration")
		`)
		assert.NoError(t, err)
	})
}

func TestObservabilityAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("validateContent", engine.MethodInfo{
				Name: "validateContent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("validation service unavailable")
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Try to validate content with error
			local result, err = obs.guardrails.validateContent("test", "text")
			assert(err ~= nil, "should have error")
			assert(string.find(err, "validation service unavailable"), "error should contain message")
			assert(result == nil, "result should be nil on error")
		`)
		assert.NoError(t, err)
	})
}

func TestObservabilityAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("metric_builder", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("createCounter", engine.MethodInfo{
				Name: "createCounter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":   engine.NewStringValue("counter-built"),
					"name": args[0], // Already a ScriptValue
					"type": engine.NewStringValue("counter"),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Use metric builder
			local counter = obs.metrics.builder("my_counter")
				:withType("counter")
				:withLabels({service = "api", env = "prod"})
				:withDescription("My custom counter")
				:build()
			
			assert(counter ~= nil, "should create counter")
			assert(counter.name == "my_counter", "should have name")
			assert(counter.type == "counter", "should have type")
		`)
		assert.NoError(t, err)
	})

	t.Run("span_context", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("getCurrentSpan", engine.MethodInfo{
				Name: "getCurrentSpan",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":      engine.NewStringValue("current-span"),
					"traceId": engine.NewStringValue("current-trace"),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Get current span context
			local span, err = obs.tracing.getCurrentSpan()
			assert(err == nil, "should not error")
			assert(span.id == "current-span", "should have current span")
			assert(span.traceId == "current-trace", "should have trace ID")
		`)
		assert.NoError(t, err)
	})
}

// Test flattened methods specifically
func TestObservabilityAdapter_FlattenedMethods(t *testing.T) {
	t.Run("flattened_guardrails_methods", func(t *testing.T) {
		guardrailsBridge := testutils.NewMockBridge("guardrails").
			WithInitialized(true).
			WithMethod("registerRule", engine.MethodInfo{
				Name: "registerRule",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"ruleId": engine.NewStringValue("rule-123"),
					"status": engine.NewStringValue("registered"),
				}), nil
			}).
			WithMethod("validateContent", engine.MethodInfo{
				Name: "validateContent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid": engine.NewBoolValue(true),
					"score": engine.NewNumberValue(0.95),
				}), nil
			}).
			WithMethod("setRuleEnabled", engine.MethodInfo{
				Name: "setRuleEnabled",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				ruleId := args[0].(engine.StringValue).Value()
				enabled := args[1].(engine.BoolValue).Value()
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"ruleId":  engine.NewStringValue(ruleId),
					"enabled": engine.NewBoolValue(enabled),
				}), nil
			})

		adapter := NewObservabilityAdapter(guardrailsBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Test flattened guardrails methods
			local rule, err = obs.guardrailsRegisterRule({
				name = "content_filter",
				type = "text_safety"
			})
			assert(err == nil, "guardrailsRegisterRule should not error")
			assert(rule.ruleId == "rule-123", "should register rule")
			
			local check, err2 = obs.guardrailsCheck("safe content")
			assert(err2 == nil, "guardrailsCheck should not error")
			assert(check.valid == true, "should validate content")
			
			local enable, err3 = obs.guardrailsEnableRule("rule-123")
			assert(err3 == nil, "guardrailsEnableRule should not error")
			assert(enable.enabled == true, "should enable rule")
			
			local disable, err4 = obs.guardrailsDisableRule("rule-123")
			assert(err4 == nil, "guardrailsDisableRule should not error")
			assert(disable.enabled == false, "should disable rule")
		`)
		assert.NoError(t, err)
	})

	t.Run("flattened_metrics_methods", func(t *testing.T) {
		metricsBridge := testutils.NewMockBridge("metrics").
			WithInitialized(true).
			WithMethod("incrementMetric", engine.MethodInfo{
				Name: "incrementMetric",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"incremented": engine.NewBoolValue(true),
					"value":       engine.NewNumberValue(1),
				}), nil
			}).
			WithMethod("setGaugeValue", engine.MethodInfo{
				Name: "setGaugeValue",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"set":   engine.NewBoolValue(true),
					"value": args[2], // value from args
				}), nil
			}).
			WithMethod("recordHistogram", engine.MethodInfo{
				Name: "recordHistogram",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"recorded": engine.NewBoolValue(true),
					"value":    args[2], // value from args
				}), nil
			}).
			WithMethod("getMetrics", engine.MethodInfo{
				Name: "getMetrics",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"counters": engine.NewObjectValue(map[string]engine.ScriptValue{
						"requests": engine.NewNumberValue(100),
					}),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, metricsBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Test flattened metrics methods
			local inc, err = obs.metricsIncrement("requests", 1)
			assert(err == nil, "metricsIncrement should not error")
			assert(inc.incremented == true, "should increment metric")
			
			local gauge, err2 = obs.metricsGauge("memory", 1024)
			assert(err2 == nil, "metricsGauge should not error")
			assert(gauge.set == true, "should set gauge")
			assert(gauge.value == 1024, "should have correct value")
			
			local hist, err3 = obs.metricsHistogram("latency", 50.5)
			assert(err3 == nil, "metricsHistogram should not error")
			assert(hist.recorded == true, "should record histogram")
			assert(hist.value == 50.5, "should have correct value")
			
			local all, err4 = obs.metricsGetAll()
			assert(err4 == nil, "metricsGetAll should not error")
			assert(all.counters.requests == 100, "should get all metrics")
		`)
		assert.NoError(t, err)
	})

	t.Run("flattened_tracing_methods", func(t *testing.T) {
		tracingBridge := testutils.NewMockBridge("tracing").
			WithInitialized(true).
			WithMethod("startSpan", engine.MethodInfo{
				Name: "startSpan",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"spanId": engine.NewStringValue("span-123"),
					"name":   args[0], // name from args
				}), nil
			}).
			WithMethod("endSpan", engine.MethodInfo{
				Name: "endSpan",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"ended":  engine.NewBoolValue(true),
					"spanId": args[0], // spanId from args
				}), nil
			}).
			WithMethod("setSpanAttribute", engine.MethodInfo{
				Name: "setSpanAttribute",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"set":   engine.NewBoolValue(true),
					"key":   args[1], // key from args
					"value": args[2], // value from args
				}), nil
			}).
			WithMethod("getTrace", engine.MethodInfo{
				Name: "getTrace",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"traceId": args[0], // traceId from args
					"spans":   engine.NewArrayValue([]engine.ScriptValue{}),
				}), nil
			})

		adapter := NewObservabilityAdapter(nil, nil, tracingBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "observability")
		require.NoError(t, err)

		err = ms.LoadModule(L, "observability")
		require.NoError(t, err)

		err = L.DoString(`
			local obs = require("observability")
			
			-- Test flattened tracing methods
			local span, err = obs.tracingStartSpan("test_operation")
			assert(err == nil, "tracingStartSpan should not error")
			assert(span.spanId == "span-123", "should start span")
			assert(span.name == "test_operation", "should have correct name")
			
			local attr, err2 = obs.tracingAddAttribute("span-123", "user.id", "user-456")
			assert(err2 == nil, "tracingAddAttribute should not error")
			assert(attr.set == true, "should set attribute")
			assert(attr.key == "user.id", "should have correct key")
			assert(attr.value == "user-456", "should have correct value")
			
			local ended, err3 = obs.tracingEndSpan("span-123")
			assert(err3 == nil, "tracingEndSpan should not error")
			assert(ended.ended == true, "should end span")
			
			local trace, err4 = obs.tracingGetTrace("trace-456")
			assert(err4 == nil, "tracingGetTrace should not error")
			assert(trace.traceId == "trace-456", "should get trace")
		`)
		assert.NoError(t, err)
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
