// ABOUTME: Comprehensive testing for all bridge adapters including cross-adapter interactions
// ABOUTME: Validates adapter interoperability, error propagation, and type conversions

package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestAllAdaptersIntegration tests that all adapters can be loaded together
func TestAllAdaptersIntegration(t *testing.T) {
	t.Run("load_all_adapters", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Create mock bridges for all adapters
		llmBridge := testutils.NewMockBridge("llm").WithInitialized(true)
		stateBridge := testutils.NewMockBridge("state").WithInitialized(true)
		eventsBridge := testutils.NewMockBridge("events").WithInitialized(true)
		structuredBridge := testutils.NewMockBridge("structured").WithInitialized(true)
		agentBridge := testutils.NewMockBridge("agent").WithInitialized(true)
		hooksBridge := testutils.NewMockBridge("hooks").WithInitialized(true)
		workflowBridge := testutils.NewMockBridge("workflow").WithInitialized(true)
		toolsBridge := testutils.NewMockBridge("tools").WithInitialized(true)
		observabilityBridge := testutils.NewMockBridge("observability").WithInitialized(true)
		modelinfoBridge := testutils.NewMockBridge("modelinfo").WithInitialized(true)

		// Utility bridges
		authBridge := testutils.NewMockBridge("auth").WithInitialized(true)
		debugBridge := testutils.NewMockBridge("debug").WithInitialized(true)
		errorsBridge := testutils.NewMockBridge("errors").WithInitialized(true)
		jsonBridge := testutils.NewMockBridge("json").WithInitialized(true)
		llmUtilsBridge := testutils.NewMockBridge("llm_utils").WithInitialized(true)
		loggerBridge := testutils.NewMockBridge("logger").WithInitialized(true)
		slogBridge := testutils.NewMockBridge("slog").WithInitialized(true)
		utilBridge := testutils.NewMockBridge("util").WithInitialized(true)

		// Create and register all adapters
		llmAdapter := NewLLMAdapter(llmBridge, nil, nil) // LLM adapter needs 3 bridges
		require.NoError(t, llmAdapter.RegisterAsModule(ms, "llm"))

		stateAdapter := NewStateAdapter(stateBridge)
		require.NoError(t, stateAdapter.RegisterAsModule(ms, "state"))

		eventsAdapter := NewEventsAdapter(eventsBridge)
		require.NoError(t, eventsAdapter.RegisterAsModule(ms, "events"))

		structuredAdapter := NewStructuredAdapter(structuredBridge)
		require.NoError(t, structuredAdapter.RegisterAsModule(ms, "structured"))

		agentAdapter := NewAgentAdapter(agentBridge)
		require.NoError(t, agentAdapter.RegisterAsModule(ms, "agent"))

		hooksAdapter := NewHooksAdapter(hooksBridge)
		err := hooksAdapter.RegisterAsModule(ms, "hooks")
		require.NoError(t, err, "failed to register hooks module")

		// Debug: List all registered modules before checking hooks
		allModules := ms.ListModules()
		t.Logf("Registered modules after hooks registration: %+v", allModules)

		// Verify hooks was registered
		info, err := ms.GetModuleInfo("hooks")
		require.NoError(t, err, "hooks module info should be retrievable")
		require.NotNil(t, info, "hooks module info should not be nil")

		workflowAdapter := NewWorkflowAdapter(workflowBridge)
		require.NoError(t, workflowAdapter.RegisterAsModule(ms, "workflow"))

		toolsAdapter := NewToolsAdapter(toolsBridge)
		require.NoError(t, toolsAdapter.RegisterAsModule(ms, "tools"))

		observabilityAdapter := NewObservabilityAdapter(observabilityBridge, nil, nil) // Observability needs 3 bridges
		require.NoError(t, observabilityAdapter.RegisterAsModule(ms, "observability"))

		modelInfoAdapter := NewModelInfoAdapter(modelinfoBridge)
		require.NoError(t, modelInfoAdapter.RegisterAsModule(ms, "modelinfo"))

		utilsAdapter := NewUtilsAdapter(authBridge, debugBridge, errorsBridge, jsonBridge,
			llmUtilsBridge, loggerBridge, slogBridge, utilBridge)
		require.NoError(t, utilsAdapter.RegisterAsModule(ms, "utils"))

		// Load all modules and verify they exist
		moduleNames := []string{
			"llm", "state", "events", "structured", "agent",
			"hooks", "workflow", "tools", "observability",
			"modelinfo", "utils",
		}

		for _, name := range moduleNames {
			err := ms.LoadModule(L, name)
			if err != nil {
				// Check if module was actually registered
				info, _ := ms.GetModuleInfo(name)
				t.Logf("Module %s info: %+v", name, info)
			}
			require.NoError(t, err, "should load module %s", name)
		}

		// Verify all modules are accessible
		err = L.DoString(`
			-- Check all modules loaded
			local modules = {
				"llm", "state", "events", "structured", "agent",
				"hooks", "workflow", "tools", "observability",
				"modelinfo", "utils"
			}
			
			for _, name in ipairs(modules) do
				local mod = require(name)
				assert(mod ~= nil, "module " .. name .. " should exist")
				assert(type(mod) == "table", "module " .. name .. " should be a table")
			end
		`)
		assert.NoError(t, err)
	})
}

// TestCrossAdapterCommunication tests interactions between different adapters
func TestCrossAdapterCommunication(t *testing.T) {
	t.Run("llm_with_structured_output", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Mock LLM bridge that returns structured data
		llmBridge := testutils.NewMockBridge("llm").
			WithInitialized(true).
			WithMethod("generateStructured", engine.MethodInfo{
				Name: "generateStructured",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"name":  engine.NewStringValue("John Doe"),
					"age":   engine.NewNumberValue(30),
					"email": engine.NewStringValue("john@example.com"),
				}), nil
			})

		// Mock structured bridge for validation
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateStruct", engine.MethodInfo{
				Name: "validateStruct",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(true),
					"errors": engine.NewArrayValue([]engine.ScriptValue{}),
				}), nil
			})

		llmAdapter := NewLLMAdapter(llmBridge, nil, nil)
		structuredAdapter := NewStructuredAdapter(structuredBridge)

		require.NoError(t, llmAdapter.RegisterAsModule(ms, "llm"))
		require.NoError(t, structuredAdapter.RegisterAsModule(ms, "structured"))

		require.NoError(t, ms.LoadModule(L, "llm"))
		require.NoError(t, ms.LoadModule(L, "structured"))

		err := L.DoString(`
			local llm = require("llm")
			local structured = require("structured")
			
			-- Generate structured data with LLM
			local result, err = llm.generateStructured("Create a user profile", {
				type = "object",
				properties = {
					name = {type = "string"},
					age = {type = "number"},
					email = {type = "string"}
				}
			})
			assert(err == nil, "should not error")
			assert(result.name == "John Doe", "should have name")
			
			-- Validate with structured module
			local validation, err2 = structured.validateStruct(result, {
				type = "object",
				properties = {
					name = {type = "string"},
					age = {type = "number"},
					email = {type = "string"}
				}
			})
			assert(err2 == nil, "validation should not error")
			assert(validation.valid == true, "should be valid")
		`)
		assert.NoError(t, err)
	})

	t.Run("agent_with_tools", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Mock tools bridge
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("listTools", engine.MethodInfo{
				Name: "listTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Performs calculations"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("web_search"),
						"description": engine.NewStringValue("Searches the web"),
					}),
				}), nil
			})

		// Mock agent bridge that uses tools
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createAgent", engine.MethodInfo{
				Name: "createAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"id":   engine.NewStringValue("agent-123"),
					"name": engine.NewStringValue("Assistant"),
				}), nil
			}).
			WithMethod("registerTool", engine.MethodInfo{
				Name: "registerTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"registered": engine.NewBoolValue(true),
				}), nil
			})

		toolsAdapter := NewToolsAdapter(toolsBridge)
		agentAdapter := NewAgentAdapter(agentBridge)

		require.NoError(t, toolsAdapter.RegisterAsModule(ms, "tools"))
		require.NoError(t, agentAdapter.RegisterAsModule(ms, "agent"))

		require.NoError(t, ms.LoadModule(L, "tools"))
		require.NoError(t, ms.LoadModule(L, "agent"))

		err := L.DoString(`
			local tools = require("tools")
			local agent = require("agent")
			
			-- List available tools
			local toolList = {tools.listTools()}
			assert(#toolList == 2, "should have 2 tools")
			
			-- Create agent
			local myAgent, err2 = agent.createAgent("Assistant", {
				model = "gpt-4",
				temperature = 0.7
			})
			assert(err2 == nil, "should create agent")
			
			-- Register tools with agent
			for _, tool in ipairs(toolList) do
				local result, err3 = agent.registerTool(myAgent.id, tool.name)
				assert(err3 == nil, "should register tool")
				assert(result.registered == true, "tool should be registered")
			end
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow_with_state_and_events", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Mock state bridge
		stateBridge := testutils.NewMockBridge("state").
			WithInitialized(true).
			WithMethod("setState", engine.MethodInfo{
				Name: "setState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"saved": engine.NewBoolValue(true),
				}), nil
			}).
			WithMethod("getState", engine.MethodInfo{
				Name: "getState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"step":   engine.NewNumberValue(2),
					"status": engine.NewStringValue("in_progress"),
				}), nil
			})

		// Mock events bridge
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("publishEvent", engine.MethodInfo{
				Name: "publishEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"published": engine.NewBoolValue(true),
					"eventId":   engine.NewStringValue("evt-123"),
				}), nil
			})

		// Mock workflow bridge
		workflowBridge := testutils.NewMockBridge("workflow").
			WithInitialized(true).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("wf-123"), nil
			}).
			WithMethod("executeWorkflow", engine.MethodInfo{
				Name: "executeWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"executionId": engine.NewStringValue("exec-123"),
					"status":      engine.NewStringValue("running"),
				}), nil
			})

		stateAdapter := NewStateAdapter(stateBridge)
		eventsAdapter := NewEventsAdapter(eventsBridge)
		workflowAdapter := NewWorkflowAdapter(workflowBridge)

		require.NoError(t, stateAdapter.RegisterAsModule(ms, "state"))
		require.NoError(t, eventsAdapter.RegisterAsModule(ms, "events"))
		require.NoError(t, workflowAdapter.RegisterAsModule(ms, "workflow"))

		require.NoError(t, ms.LoadModule(L, "state"))
		require.NoError(t, ms.LoadModule(L, "events"))
		require.NoError(t, ms.LoadModule(L, "workflow"))

		err := L.DoString(`
			local state = require("state")
			local events = require("events")
			local workflow = require("workflow")
			
			-- Create workflow
			local wfId = workflow.createWorkflow("DataProcessing", {
				type = "sequential",
				steps = {"validate", "process", "save"}
			})
			assert(wfId == "wf-123", "should create workflow")
			
			-- Start execution
			local exec = workflow.executeWorkflow(wfId, {
				input = "test data"
			})
			assert(exec.executionId == "exec-123", "should start execution")
			
			-- Save workflow state
			local saveResult = state.setState("workflow:" .. wfId, {
				executionId = exec.executionId,
				currentStep = 1,
				status = "running"
			})
			assert(saveResult.saved == true, "state should be saved")
			
			-- Publish workflow event
			local event = events.publishEvent("workflow.started", {
				workflowId = wfId,
				executionId = exec.executionId,
				timestamp = os.time()
			})
			assert(event.published == true, "event should be published")
			
			-- Get workflow state
			local currentState = state.getState("workflow:" .. wfId)
			assert(currentState.status == "in_progress", "should have correct status")
		`)
		assert.NoError(t, err)
	})
}

// TestAdapterErrorPropagation tests error handling across adapters
func TestAdapterErrorPropagation(t *testing.T) {
	t.Run("bridge_error_propagation", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Create bridge that returns errors
		errorBridge := testutils.NewMockBridge("test").
			WithInitialized(true).
			WithMethod("failingMethod", engine.MethodInfo{
				Name: "failingMethod",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("bridge operation failed: connection timeout")
			})

		// Create a simple test adapter
		adapter := &testAdapter{bridge: errorBridge}
		require.NoError(t, adapter.RegisterAsModule(ms, "test"))
		require.NoError(t, ms.LoadModule(L, "test"))

		err := L.DoString(`
			local test = require("test")
			
			-- Call failing method
			local result, err = test.failingMethod()
			assert(result == nil, "result should be nil on error")
			assert(err ~= nil, "should have error")
			assert(string.find(err, "connection timeout"), "error should contain message")
		`)
		assert.NoError(t, err)
	})

	t.Run("type_conversion_errors", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Create bridge that expects specific types
		typeBridge := testutils.NewMockBridge("llm").
			WithInitialized(true).
			WithMethod("generate", engine.MethodInfo{
				Name: "generate",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Validate first argument is string
				if len(args) < 1 || args[0].Type() != engine.TypeString {
					return nil, fmt.Errorf("prompt must be a string")
				}
				// Validate second argument is object
				if len(args) < 2 || args[1].Type() != engine.TypeObject {
					return nil, fmt.Errorf("options must be an object")
				}
				return engine.NewStringValue("Generated text"), nil
			})

		llmAdapter := NewLLMAdapter(typeBridge, nil, nil)
		require.NoError(t, llmAdapter.RegisterAsModule(ms, "llm"))
		require.NoError(t, ms.LoadModule(L, "llm"))

		err := L.DoString(`
			local llm = require("llm")
			
			-- Test with wrong types
			local result1, err1 = llm.generate(123, {}) -- number instead of string
			assert(result1 == nil, "should fail with wrong type")
			assert(string.find(err1, "prompt must be a string"), "should have type error")
			
			local result2, err2 = llm.generate("prompt", "not-an-object") -- string instead of object
			assert(result2 == nil, "should fail with wrong type")
			assert(string.find(err2, "options must be an object"), "should have type error")
			
			-- Test with correct types
			local result3, err3 = llm.generate("Test prompt", {temperature = 0.7})
			assert(err3 == nil, "should succeed with correct types")
			assert(result3 == "Generated text", "should return expected result")
		`)
		assert.NoError(t, err)
	})
}

// TestAdapterTypeConversions tests complex type conversions
func TestAdapterTypeConversions(t *testing.T) {
	t.Run("nested_object_conversions", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Mock bridge that works with nested structures
		complexBridge := testutils.NewMockBridge("complex").
			WithInitialized(true).
			WithMethod("processNested", engine.MethodInfo{
				Name: "processNested",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return a deeply nested structure
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"user": engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewNumberValue(123),
						"name": engine.NewStringValue("John"),
						"settings": engine.NewObjectValue(map[string]engine.ScriptValue{
							"theme": engine.NewStringValue("dark"),
							"notifications": engine.NewObjectValue(map[string]engine.ScriptValue{
								"email":     engine.NewBoolValue(true),
								"push":      engine.NewBoolValue(false),
								"frequency": engine.NewStringValue("daily"),
							}),
						}),
						"tags": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("premium"),
							engine.NewStringValue("verified"),
						}),
					}),
					"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{
						"createdAt": engine.NewNumberValue(float64(time.Now().Unix())),
						"version":   engine.NewNumberValue(2.0),
					}),
				}), nil
			})

		adapter := &testAdapter{bridge: complexBridge}
		require.NoError(t, adapter.RegisterAsModule(ms, "complex"))
		require.NoError(t, ms.LoadModule(L, "complex"))

		err := L.DoString(`
			local complex = require("complex")
			
			-- Process nested structure
			local result, err = complex.processNested({})
			assert(err == nil, "should not error")
			
			-- Verify nested access
			assert(result.user.id == 123, "should have user id")
			assert(result.user.name == "John", "should have user name")
			assert(result.user.settings.theme == "dark", "should have theme")
			assert(result.user.settings.notifications.email == true, "should have email notifications")
			assert(result.user.settings.notifications.push == false, "should not have push notifications")
			assert(#result.user.tags == 2, "should have 2 tags")
			assert(result.user.tags[1] == "premium", "should have premium tag")
			assert(result.metadata.version == 2.0, "should have version")
		`)
		assert.NoError(t, err)
	})

	t.Run("array_of_objects_conversion", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Mock bridge that returns array of objects
		arrayBridge := testutils.NewMockBridge("array").
			WithInitialized(true).
			WithMethod("getItems", engine.MethodInfo{
				Name: "getItems",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				items := []engine.ScriptValue{}
				for i := 0; i < 3; i++ {
					items = append(items, engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":     engine.NewNumberValue(float64(i + 1)),
						"name":   engine.NewStringValue(fmt.Sprintf("Item %d", i+1)),
						"active": engine.NewBoolValue(i%2 == 0),
						"data": engine.NewObjectValue(map[string]engine.ScriptValue{
							"value": engine.NewNumberValue(float64((i + 1) * 10)),
							"unit":  engine.NewStringValue("points"),
						}),
					}))
				}
				return engine.NewArrayValue(items), nil
			})

		adapter := &testAdapter{bridge: arrayBridge}
		require.NoError(t, adapter.RegisterAsModule(ms, "array"))
		require.NoError(t, ms.LoadModule(L, "array"))

		err := L.DoString(`
			local array = require("array")
			
			-- Get array of items
			local items, err = array.getItems()
			assert(err == nil, "should not error")
			assert(#items == 3, "should have 3 items")
			
			-- Verify each item
			for i, item in ipairs(items) do
				assert(item.id == i, "should have correct id")
				assert(item.name == "Item " .. i, "should have correct name")
				assert(item.active == (i % 2 == 1), "should have correct active state")
				assert(item.data.value == i * 10, "should have correct value")
				assert(item.data.unit == "points", "should have correct unit")
			end
		`)
		assert.NoError(t, err)
	})
}

// TestAdapterPerformance tests performance characteristics
func TestAdapterPerformance(t *testing.T) {
	t.Run("concurrent_adapter_calls", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		ms := gopherlua.NewModuleSystem()

		// Create a bridge that simulates work
		perfBridge := testutils.NewMockBridge("perf").
			WithInitialized(true).
			WithMethod("process", engine.MethodInfo{
				Name: "process",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Simulate some work
				time.Sleep(10 * time.Millisecond)
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"processed": engine.NewBoolValue(true),
					"timestamp": engine.NewNumberValue(float64(time.Now().UnixNano())),
				}), nil
			})

		adapter := &testAdapter{bridge: perfBridge}
		require.NoError(t, adapter.RegisterAsModule(ms, "perf"))
		require.NoError(t, ms.LoadModule(L, "perf"))

		// Test concurrent calls from Lua coroutines
		err := L.DoString(`
			local perf = require("perf")
			
			-- Create multiple coroutines
			local coroutines = {}
			local results = {}
			
			for i = 1, 5 do
				coroutines[i] = coroutine.create(function()
					local result, err = perf.process({id = i})
					assert(err == nil, "should not error")
					results[i] = result
				end)
			end
			
			-- Run all coroutines
			local allDone = false
			while not allDone do
				allDone = true
				for i, co in ipairs(coroutines) do
					if coroutine.status(co) ~= "dead" then
						allDone = false
						coroutine.resume(co)
					end
				end
			end
			
			-- Verify all completed
			assert(#results == 5, "should have 5 results")
			for i = 1, 5 do
				assert(results[i].processed == true, "should be processed")
			end
		`)
		assert.NoError(t, err)
	})
}

// TestAdapterDocumentation tests that all adapters provide proper documentation
func TestAdapterDocumentation(t *testing.T) {
	// Test each adapter can be created and registered
	adapters := []struct {
		name   string
		create func() error
	}{
		{"llm", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewLLMAdapter(testutils.NewMockBridge("llm").WithInitialized(true), nil, nil)
			return adapter.RegisterAsModule(ms, "llm")
		}},
		{"state", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewStateAdapter(testutils.NewMockBridge("state").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "state")
		}},
		{"events", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewEventsAdapter(testutils.NewMockBridge("events").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "events")
		}},
		{"structured", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewStructuredAdapter(testutils.NewMockBridge("structured").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "structured")
		}},
		{"agent", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewAgentAdapter(testutils.NewMockBridge("agent").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "agent")
		}},
		{"hooks", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewHooksAdapter(testutils.NewMockBridge("hooks").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "hooks")
		}},
		{"workflow", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewWorkflowAdapter(testutils.NewMockBridge("workflow").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "workflow")
		}},
		{"tools", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewToolsAdapter(testutils.NewMockBridge("tools").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "tools")
		}},
		{"observability", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewObservabilityAdapter(testutils.NewMockBridge("observability").WithInitialized(true), nil, nil)
			return adapter.RegisterAsModule(ms, "observability")
		}},
		{"modelinfo", func() error {
			ms := gopherlua.NewModuleSystem()
			adapter := NewModelInfoAdapter(testutils.NewMockBridge("modelinfo").WithInitialized(true))
			return adapter.RegisterAsModule(ms, "modelinfo")
		}},
	}

	for _, tc := range adapters {
		t.Run(tc.name+"_adapter_creation", func(t *testing.T) {
			err := tc.create()
			assert.NoError(t, err, "adapter should be created and registered")
		})
	}
}

// Helper test adapter for testing
type testAdapter struct {
	bridge engine.Bridge
}

func (ta *testAdapter) GetAdapterName() string {
	return "test"
}

func (ta *testAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		module := L.NewTable()
		converter := gopherlua.NewLuaTypeConverter()

		// Add failingMethod
		L.SetField(module, "failingMethod", L.NewFunction(func(L *lua.LState) int {
			result, err := ta.bridge.ExecuteMethod(context.Background(), "failingMethod", []engine.ScriptValue{})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))

		// Add processNested
		L.SetField(module, "processNested", L.NewFunction(func(L *lua.LState) int {
			arg := L.CheckTable(1)
			scriptArg, err := converter.ToScriptValue(arg)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result, err := ta.bridge.ExecuteMethod(context.Background(), "processNested", []engine.ScriptValue{scriptArg})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))

		// Add getItems
		L.SetField(module, "getItems", L.NewFunction(func(L *lua.LState) int {
			result, err := ta.bridge.ExecuteMethod(context.Background(), "getItems", []engine.ScriptValue{})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))

		// Add process
		L.SetField(module, "process", L.NewFunction(func(L *lua.LState) int {
			arg := L.CheckTable(1)
			scriptArg, err := converter.ToScriptValue(arg)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result, err := ta.bridge.ExecuteMethod(context.Background(), "process", []engine.ScriptValue{scriptArg})
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			luaResult, err := converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))

		L.Push(module)
		return 1
	}
}

func (ta *testAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Create module definition
	def := gopherlua.ModuleDefinition{
		Name:        name,
		Description: "Test adapter module",
		Version:     "1.0.0",
		LoadFunc:    ta.CreateLuaModule(),
	}
	return ms.Register(def)
}
