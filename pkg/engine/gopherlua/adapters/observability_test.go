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

		// Guardrails methods
		assert.Contains(t, methods, "enableGuardrails")
		assert.Contains(t, methods, "validateContent")
		assert.Contains(t, methods, "addBehavioralConstraint")
		assert.Contains(t, methods, "checkCompliance")

		// Metrics methods
		assert.Contains(t, methods, "createCounter")
		assert.Contains(t, methods, "createGauge")
		assert.Contains(t, methods, "createTimer")
		assert.Contains(t, methods, "recordMetric")
		assert.Contains(t, methods, "getMetrics")

		// Tracing methods
		assert.Contains(t, methods, "startSpan")
		assert.Contains(t, methods, "addSpanEvent")
		assert.Contains(t, methods, "setSpanAttribute")
		assert.Contains(t, methods, "endSpan")
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
			
			-- Check namespaces exist
			assert(type(observability.guardrails) == "table", "guardrails namespace should exist")
			assert(type(observability.metrics) == "table", "metrics namespace should exist")
			assert(type(observability.tracing) == "table", "tracing namespace should exist")
			
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

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
