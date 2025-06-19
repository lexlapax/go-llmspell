// ABOUTME: Tests for Agent bridge adapter that exposes go-llms agent functionality to Lua scripts
// ABOUTME: Validates agent lifecycle, communication, state management, events, profiling, and workflow operations

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

func TestAgentAdapter_Creation(t *testing.T) {
	t.Run("create_agent_adapter", func(t *testing.T) {
		// Create agent bridge mock
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "agent",
				Version:     "2.0.0",
				Description: "Agent system bridge with state serialization, event replay, and performance profiling",
			}).
			WithMethod("createAgent", engine.MethodInfo{
				Name:        "createAgent",
				Description: "Create a new agent with configuration",
				ReturnType:  "Agent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock agent creation
				result := map[string]engine.ScriptValue{
					"id":   engine.NewStringValue("agent-123"),
					"type": engine.NewStringValue("basic"),
					"name": engine.NewStringValue("test-agent"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("listAgents", engine.MethodInfo{
				Name:        "listAgents",
				Description: "List all registered agents",
				ReturnType:  "array",
			}, nil).
			WithMethod("getAgent", engine.MethodInfo{
				Name:        "getAgent",
				Description: "Get an agent by ID",
				ReturnType:  "Agent",
			}, nil)

		// Create adapter
		adapter := NewAgentAdapter(agentBridge)
		require.NotNil(t, adapter)

		// Should have flattened agent-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "createAgent")
		assert.Contains(t, methods, "createLLMAgent")
		assert.Contains(t, methods, "listAgents")
		assert.Contains(t, methods, "getAgent")
		assert.Contains(t, methods, "removeAgent")
		assert.Contains(t, methods, "run")
		assert.Contains(t, methods, "runAsync")
		assert.Contains(t, methods, "stateGet")
		assert.Contains(t, methods, "stateSet")
	})

	t.Run("agent_module_structure", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "agent",
			}).
			WithMethod("createAgent", engine.MethodInfo{
				Name: "createAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-agent"), nil
			}).
			WithMethod("listAgents", engine.MethodInfo{
				Name: "listAgents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			}).
			WithMethod("getAgentState", engine.MethodInfo{
				Name: "getAgentState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{}), nil
			}).
			WithMethod("startAgentProfiling", engine.MethodInfo{
				Name: "startAgentProfiling",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("profile-session"), nil
			}).
			WithMethod("replayAgentEvents", engine.MethodInfo{
				Name: "replayAgentEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("workflow-agent"), nil
			}).
			WithMethod("setAgentHook", engine.MethodInfo{
				Name: "setAgentHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check standard methods exist
		assert.NotEqual(t, lua.LNil, module.RawGetString("createAgent"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("listAgents"))

		// Check that old namespaces do NOT exist (flattened structure)
		lifecycle := module.RawGetString("lifecycle")
		assert.Equal(t, lua.LNil, lifecycle, "lifecycle namespace should NOT exist - methods are flattened")

		communication := module.RawGetString("communication")
		assert.Equal(t, lua.LNil, communication, "communication namespace should NOT exist - methods are flattened")

		state := module.RawGetString("state")
		assert.Equal(t, lua.LNil, state, "state namespace should NOT exist - methods are flattened")

		events := module.RawGetString("events")
		assert.Equal(t, lua.LNil, events, "events namespace should NOT exist - methods are flattened")

		profiling := module.RawGetString("profiling")
		assert.Equal(t, lua.LNil, profiling, "profiling namespace should NOT exist - methods are flattened")

		workflow := module.RawGetString("workflow")
		assert.Equal(t, lua.LNil, workflow, "workflow namespace should NOT exist - methods are flattened")

		hooks := module.RawGetString("hooks")
		assert.Equal(t, lua.LNil, hooks, "hooks namespace should NOT exist - methods are flattened")

		// Check flattened methods exist
		// Lifecycle methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleCreate"), "lifecycleCreate should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleCreateLLM"), "lifecycleCreateLLM should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleList"), "lifecycleList should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleGet"), "lifecycleGet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleRemove"), "lifecycleRemove should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("lifecycleGetMetrics"), "lifecycleGetMetrics should exist")

		// Communication methods (simplified naming)
		assert.NotEqual(t, lua.LNil, module.RawGetString("run"), "run should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("runAsync"), "runAsync should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("registerTool"), "registerTool should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("unregisterTool"), "unregisterTool should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("listTools"), "listTools should exist")

		// State methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateGet"), "stateGet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateSet"), "stateSet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateExport"), "stateExport should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateImport"), "stateImport should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateSaveSnapshot"), "stateSaveSnapshot should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateLoadSnapshot"), "stateLoadSnapshot should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("stateListSnapshots"), "stateListSnapshots should exist")

		// Events methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsEmit"), "eventsEmit should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsSubscribe"), "eventsSubscribe should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsUnsubscribe"), "eventsUnsubscribe should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsStartRecording"), "eventsStartRecording should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsStopRecording"), "eventsStopRecording should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("eventsReplay"), "eventsReplay should exist")

		// Profiling methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("profilingStart"), "profilingStart should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("profilingStop"), "profilingStop should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("profilingGetMetrics"), "profilingGetMetrics should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("profilingGetReport"), "profilingGetReport should exist")

		// Workflow methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("workflowCreate"), "workflowCreate should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("workflowExecute"), "workflowExecute should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("workflowAddStep"), "workflowAddStep should exist")

		// Hooks methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("hooksRegister"), "hooksRegister should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("hooksSet"), "hooksSet should exist")
		assert.NotEqual(t, lua.LNil, module.RawGetString("hooksUnregister"), "hooksUnregister should exist")

		// Utils methods
		assert.NotEqual(t, lua.LNil, module.RawGetString("utilsValidateConfig"), "utilsValidateConfig should exist")
	})
}

func TestAgentAdapter_AgentLifecycle(t *testing.T) {
	t.Run("create_basic_agent", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createAgent", engine.MethodInfo{
				Name: "createAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Extract agent config
				if len(args) >= 2 && args[0].Type() == engine.TypeString && args[1].Type() == engine.TypeObject {
					agentID := args[0].(engine.StringValue).Value()
					config := args[1].(engine.ObjectValue).Fields()

					// Mock agent creation with config
					result := map[string]engine.ScriptValue{
						"id":   engine.NewStringValue(agentID),
						"type": engine.NewStringValue("basic"),
						"name": config["name"],
					}
					return engine.NewObjectValue(result), nil
				}
				return nil, fmt.Errorf("invalid parameters")
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Create agent from Lua
		err = L.DoString(`
			local agent = require("agent")
			local newAgent = agent.lifecycleCreate("test-agent", {
				name = "Test Agent",
				type = "basic",
				description = "A test agent"
			})
			assert(newAgent ~= nil)
			assert(newAgent.id == "test-agent")
			assert(newAgent.type == "basic")
			assert(newAgent.name == "Test Agent")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_llm_agent", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createLLMAgent", engine.MethodInfo{
				Name: "createLLMAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeString {
					name := args[0].(engine.StringValue).Value()
					// provider := args[1] // Would be provider object

					result := map[string]engine.ScriptValue{
						"id":       engine.NewStringValue("llm-agent-123"),
						"type":     engine.NewStringValue("llm"),
						"name":     engine.NewStringValue(name),
						"provider": engine.NewStringValue("openai"),
					}
					return engine.NewObjectValue(result), nil
				}
				return nil, fmt.Errorf("invalid parameters")
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Create LLM agent from Lua
		err = L.DoString(`
			local agent = require("agent")
			local llmAgent = agent.lifecycleCreateLLM("Smart Agent", {
				provider = "openai",
				model = "gpt-4"
			}, {
				temperature = 0.7
			})
			assert(llmAgent ~= nil)
			assert(llmAgent.type == "llm")
			assert(llmAgent.provider == "openai")
		`)
		assert.NoError(t, err)
	})

	t.Run("list_and_get_agents", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("listAgents", engine.MethodInfo{
				Name: "listAgents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock agents
				agents := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("agent-1"),
						"type": engine.NewStringValue("basic"),
						"name": engine.NewStringValue("First Agent"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("agent-2"),
						"type": engine.NewStringValue("llm"),
						"name": engine.NewStringValue("Second Agent"),
					}),
				}
				return engine.NewArrayValue(agents), nil
			}).
			WithMethod("getAgent", engine.MethodInfo{
				Name: "getAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if args[0].Type() == engine.TypeString {
					agentID := args[0].(engine.StringValue).Value()
					result := map[string]engine.ScriptValue{
						"id":          engine.NewStringValue(agentID),
						"type":        engine.NewStringValue("basic"),
						"name":        engine.NewStringValue("Retrieved Agent"),
						"description": engine.NewStringValue("An agent retrieved by ID"),
					}
					return engine.NewObjectValue(result), nil
				}
				return nil, fmt.Errorf("invalid agent ID")
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test listing and getting agents (arrays return as multiple values)
		err = L.DoString(`
			local agent = require("agent")
			
			-- List agents - get individual agents as multiple returns
			local agent1, agent2 = agent.lifecycleList()
			assert(agent1.id == "agent-1", "first agent should be agent-1")
			assert(agent2.id == "agent-2", "second agent should be agent-2")
			
			-- Get specific agent
			local retrievedAgent, err = agent.lifecycleGet("test-agent")
			assert(err == nil, "get should not error: " .. tostring(err))
			assert(retrievedAgent.id == "test-agent", "should retrieve correct agent")
			assert(retrievedAgent.name == "Retrieved Agent", "should have correct name")
		`)
		assert.NoError(t, err)
	})

	t.Run("remove_agent", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("removeAgent", engine.MethodInfo{
				Name: "removeAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful removal
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test agent removal
		err = L.DoString(`
			local agent = require("agent")
			
			local result, err = agent.lifecycleRemove("test-agent")
			assert(err == nil, "remove should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_AgentCommunication(t *testing.T) {
	t.Run("run_agent", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("runAgent", engine.MethodInfo{
				Name: "runAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock agent execution
				result := map[string]engine.ScriptValue{
					"output":        engine.NewStringValue("Agent execution completed"),
					"success":       engine.NewBoolValue(true),
					"executionTime": engine.NewNumberValue(150),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test agent execution
		err = L.DoString(`
			local agent = require("agent")
			
			local result, err = agent.run("test-agent", {
				prompt = "Hello, how are you?",
				context = "test conversation"
			})
			assert(err == nil, "run should not error: " .. tostring(err))
			assert(result.success == true, "execution should be successful")
			assert(result.output == "Agent execution completed", "should have output")
		`)
		assert.NoError(t, err)
	})

	t.Run("run_agent_async", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("runAgentAsync", engine.MethodInfo{
				Name: "runAgentAsync",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock async execution - return channel/promise identifier
				result := map[string]engine.ScriptValue{
					"channelID": engine.NewStringValue("async-123"),
					"status":    engine.NewStringValue("running"),
					"async":     engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test async agent execution
		err = L.DoString(`
			local agent = require("agent")
			
			local channel, err = agent.runAsync("test-agent", {
				prompt = "Generate a long story",
				streaming = true
			})
			assert(err == nil, "runAsync should not error: " .. tostring(err))
			assert(channel.async == true, "should be async execution")
			assert(channel.status == "running", "should be running")
		`)
		assert.NoError(t, err)
	})

	t.Run("register_tool", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("registerTool", engine.MethodInfo{
				Name: "registerTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock tool registration
				return engine.NewNilValue(), nil
			}).
			WithMethod("getAgentTools", engine.MethodInfo{
				Name: "getAgentTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock tool list
				tools := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Performs calculations"),
						"type":        engine.NewStringValue("function"),
					}),
				}
				return engine.NewArrayValue(tools), nil
			}).
			WithMethod("registerAgentTool", engine.MethodInfo{
				Name: "registerAgentTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock tool registration (alias)
				return engine.NewNilValue(), nil
			}).
			WithMethod("listAgentTools", engine.MethodInfo{
				Name: "listAgentTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock tool list (alias)
				tools := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Performs calculations"),
						"type":        engine.NewStringValue("function"),
					}),
				}
				return engine.NewArrayValue(tools), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test tool registration and listing
		err = L.DoString(`
			local agent = require("agent")
			
			-- Register a tool
			local regResult, regErr = agent.registerTool("test-agent", {
				name = "calculator",
				description = "Performs mathematical calculations",
				["function"] = function(a, b) return a + b end
			})
			assert(regErr == nil, "tool registration should not error: " .. tostring(regErr))
			
			-- List tools - get individual tools as multiple returns
			local tool = agent.listTools("test-agent")
			assert(tool.name == "calculator", "should have calculator tool")
			assert(tool.type == "function", "should be function type")
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_StateManagement(t *testing.T) {
	t.Run("get_and_set_state", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("getAgentState", engine.MethodInfo{
				Name: "getAgentState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock state retrieval
				state := map[string]engine.ScriptValue{
					"agentID":     engine.NewStringValue("test-agent"),
					"currentStep": engine.NewNumberValue(3),
					"variables": engine.NewObjectValue(map[string]engine.ScriptValue{
						"counter": engine.NewNumberValue(42),
						"mode":    engine.NewStringValue("active"),
					}),
				}
				return engine.NewObjectValue(state), nil
			}).
			WithMethod("setAgentState", engine.MethodInfo{
				Name: "setAgentState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock state setting
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test state operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Get agent state
			local state, getErr = agent.stateGet("test-agent")
			assert(getErr == nil, "get state should not error: " .. tostring(getErr))
			assert(state.agentID == "test-agent", "should have correct agent ID")
			assert(state.currentStep == 3, "should have current step")
			assert(state.variables.counter == 42, "should have variables")
			
			-- Set agent state
			local newState = {
				agentID = "test-agent",
				currentStep = 4,
				variables = { counter = 50, mode = "updated" }
			}
			local setResult, setErr = agent.stateSet("test-agent", newState)
			assert(setErr == nil, "set state should not error: " .. tostring(setErr))
		`)
		assert.NoError(t, err)
	})

	t.Run("export_and_import_state", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("exportAgentState", engine.MethodInfo{
				Name: "exportAgentState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock state export
				result := map[string]engine.ScriptValue{
					"agentID": engine.NewStringValue("test-agent"),
					"format":  engine.NewStringValue("json"),
					"state": engine.NewObjectValue(map[string]engine.ScriptValue{
						"data": engine.NewStringValue("exported state data"),
					}),
					"timestamp": engine.NewStringValue("2024-01-01T00:00:00Z"),
					"version":   engine.NewStringValue("1.0"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("importAgentState", engine.MethodInfo{
				Name: "importAgentState",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock state import
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test state export/import
		err = L.DoString(`
			local agent = require("agent")
			
			-- Export agent state
			local exported, exportErr = agent.stateExport("test-agent", "json")
			assert(exportErr == nil, "export should not error: " .. tostring(exportErr))
			assert(exported.agentID == "test-agent", "should have agent ID")
			assert(exported.format == "json", "should be JSON format")
			assert(exported.version == "1.0", "should have version")
			
			-- Import agent state
			local importResult, importErr = agent.stateImport("test-agent", exported)
			assert(importErr == nil, "import should not error: " .. tostring(importErr))
		`)
		assert.NoError(t, err)
	})

	t.Run("save_and_load_snapshots", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("saveAgentSnapshot", engine.MethodInfo{
				Name: "saveAgentSnapshot",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock snapshot save
				result := map[string]engine.ScriptValue{
					"snapshotName": args[1],
					"agentID":      args[0],
					"created":      engine.NewStringValue("2024-01-01T00:00:00Z"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("loadAgentSnapshot", engine.MethodInfo{
				Name: "loadAgentSnapshot",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock snapshot load
				return engine.NewNilValue(), nil
			}).
			WithMethod("listAgentSnapshots", engine.MethodInfo{
				Name: "listAgentSnapshots",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock snapshot list
				snapshots := []engine.ScriptValue{
					engine.NewStringValue("snapshot-1"),
					engine.NewStringValue("snapshot-2"),
				}
				return engine.NewArrayValue(snapshots), nil
			}).
			WithMethod("createAgentSnapshot", engine.MethodInfo{
				Name: "createAgentSnapshot",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock snapshot save (alias for saveAgentSnapshot)
				result := map[string]engine.ScriptValue{
					"snapshotName": args[1],
					"agentID":      args[0],
					"created":      engine.NewStringValue("2024-01-01T00:00:00Z"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("restoreAgentSnapshot", engine.MethodInfo{
				Name: "restoreAgentSnapshot",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock snapshot load (alias for loadAgentSnapshot)
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test snapshot operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Save snapshot
			local saved, saveErr = agent.stateSaveSnapshot("test-agent", "backup-1")
			assert(saveErr == nil, "save snapshot should not error: " .. tostring(saveErr))
			assert(saved.snapshotName == "backup-1", "should have snapshot name")
			
			-- Load snapshot
			local loadResult, loadErr = agent.stateLoadSnapshot("test-agent", "backup-1")
			assert(loadErr == nil, "load snapshot should not error: " .. tostring(loadErr))
			
			-- List snapshots - get individual snapshots as multiple returns
			local snap1, snap2 = agent.stateListSnapshots("test-agent")
			assert(snap1 == "snapshot-1", "should have first snapshot")
			assert(snap2 == "snapshot-2", "should have second snapshot")
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_Events(t *testing.T) {
	t.Run("emit_and_subscribe_events", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("emitAgentEvent", engine.MethodInfo{
				Name: "emitAgentEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock event emission
				return engine.NewNilValue(), nil
			}).
			WithMethod("subscribeToEvents", engine.MethodInfo{
				Name: "subscribeToEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock event subscription - return subscription ID
				return engine.NewStringValue("sub-123"), nil
			}).
			WithMethod("unsubscribeFromEvents", engine.MethodInfo{
				Name: "unsubscribeFromEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock unsubscribe
				return engine.NewNilValue(), nil
			}).
			WithMethod("subscribeAgentEvent", engine.MethodInfo{
				Name: "subscribeAgentEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock event subscription (alias) - return subscription ID
				return engine.NewStringValue("sub-123"), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test event operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Emit event
			local emitResult, emitErr = agent.eventsEmit("test-agent", "task_completed", {
				task_id = "task-123",
				result = "success"
			})
			assert(emitErr == nil, "emit should not error: " .. tostring(emitErr))
			
			-- Subscribe to events
			local handler = function(event)
				print("Received event:", event.type)
			end
			local subId, subErr = agent.eventsSubscribe({
				agentID = "test-agent",
				eventType = "task_completed"
			}, handler)
			assert(subErr == nil, "subscribe should not error: " .. tostring(subErr))
			assert(subId == "sub-123", "should return subscription ID")
			
			-- Unsubscribe
			local unsubResult, unsubErr = agent.eventsUnsubscribe(subId)
			assert(unsubErr == nil, "unsubscribe should not error: " .. tostring(unsubErr))
		`)
		assert.NoError(t, err)
	})

	t.Run("event_recording_and_replay", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("startEventRecording", engine.MethodInfo{
				Name: "startEventRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock recording start
				result := map[string]engine.ScriptValue{
					"recordingID": engine.NewStringValue("rec-123"),
					"started":     engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("stopEventRecording", engine.MethodInfo{
				Name: "stopEventRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock recording stop
				result := map[string]engine.ScriptValue{
					"recordingID": args[0],
					"stopped":     engine.NewBoolValue(true),
					"eventCount":  engine.NewNumberValue(25),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("replayAgentEvents", engine.MethodInfo{
				Name: "replayAgentEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock event replay
				return engine.NewNilValue(), nil
			}).
			WithMethod("startAgentEventRecording", engine.MethodInfo{
				Name: "startAgentEventRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock recording start (alias)
				result := map[string]engine.ScriptValue{
					"recordingID": engine.NewStringValue("rec-123"),
					"started":     engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("stopAgentEventRecording", engine.MethodInfo{
				Name: "stopAgentEventRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock recording stop (alias)
				result := map[string]engine.ScriptValue{
					"recordingID": args[0],
					"stopped":     engine.NewBoolValue(true),
					"eventCount":  engine.NewNumberValue(25),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test event recording and replay
		err = L.DoString(`
			local agent = require("agent")
			
			-- Start recording
			local recording, startErr = agent.eventsStartRecording("test-agent")
			assert(startErr == nil, "start recording should not error: " .. tostring(startErr))
			assert(recording.started == true, "recording should be started")
			
			-- Stop recording
			local stopped, stopErr = agent.eventsStopRecording(recording.recordingID)
			assert(stopErr == nil, "stop recording should not error: " .. tostring(stopErr))
			assert(stopped.eventCount == 25, "should have recorded events")
			
			-- Replay events
			local replayResult, replayErr = agent.eventsReplay("test-agent", {
				speed = 2.0,
				fromTime = "2024-01-01T00:00:00Z"
			})
			assert(replayErr == nil, "replay should not error: " .. tostring(replayErr))
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_Profiling(t *testing.T) {
	t.Run("agent_profiling", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("startAgentProfiling", engine.MethodInfo{
				Name: "startAgentProfiling",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock profiling start
				return engine.NewStringValue("profile-session-123"), nil
			}).
			WithMethod("stopAgentProfiling", engine.MethodInfo{
				Name: "stopAgentProfiling",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock profiling stop
				result := map[string]engine.ScriptValue{
					"sessionID":  args[0],
					"stopped":    engine.NewBoolValue(true),
					"cpuProfile": engine.NewStringValue("/tmp/cpu.pprof"),
					"memProfile": engine.NewStringValue("/tmp/mem.pprof"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("getAgentPerformanceReport", engine.MethodInfo{
				Name: "getAgentPerformanceReport",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock performance report
				report := map[string]engine.ScriptValue{
					"agentID":     args[0],
					"cpuUsage":    engine.NewStringValue("15%"),
					"memoryUsage": engine.NewStringValue("128MB"),
					"avgLatency":  engine.NewStringValue("45ms"),
					"totalOps":    engine.NewNumberValue(1234),
					"successRate": engine.NewNumberValue(0.987),
				}
				return engine.NewObjectValue(report), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test profiling operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Start profiling
			local sessionID, startErr = agent.profilingStart("test-agent")
			assert(startErr == nil, "start profiling should not error: " .. tostring(startErr))
			assert(sessionID == "profile-session-123", "should return session ID")
			
			-- Stop profiling
			local stopped, stopErr = agent.profilingStop(sessionID)
			assert(stopErr == nil, "stop profiling should not error: " .. tostring(stopErr))
			assert(stopped.stopped == true, "should be stopped")
			assert(stopped.cpuProfile ~= nil, "should have CPU profile")
			
			-- Get performance report
			local report, reportErr = agent.profilingGetReport("test-agent")
			assert(reportErr == nil, "get report should not error: " .. tostring(reportErr))
			assert(report.cpuUsage == "15%", "should have CPU usage")
			assert(report.successRate == 0.987, "should have success rate")
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_Workflow(t *testing.T) {
	t.Run("create_workflow", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createWorkflow", engine.MethodInfo{
				Name: "createWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock workflow creation
				if args[0].Type() == engine.TypeString {
					workflowType := args[0].(engine.StringValue).Value()
					result := map[string]engine.ScriptValue{
						"id":   engine.NewStringValue("workflow-123"),
						"type": engine.NewStringValue(workflowType),
						"name": engine.NewStringValue("Test Workflow"),
					}
					return engine.NewObjectValue(result), nil
				}
				return nil, fmt.Errorf("invalid workflow type")
			}).
			WithMethod("addWorkflowStep", engine.MethodInfo{
				Name: "addWorkflowStep",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock step addition
				return engine.NewNilValue(), nil
			}).
			WithMethod("createAgentWorkflow", engine.MethodInfo{
				Name: "createAgentWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock workflow creation (alias)
				if len(args) >= 2 && args[0].Type() == engine.TypeString {
					agentID := args[0].(engine.StringValue).Value()
					result := map[string]engine.ScriptValue{
						"id":      engine.NewStringValue("workflow-123"),
						"type":    engine.NewStringValue("sequential"),
						"name":    engine.NewStringValue("Test Workflow"),
						"agentID": engine.NewStringValue(agentID),
					}
					return engine.NewObjectValue(result), nil
				}
				return nil, fmt.Errorf("invalid parameters")
			}).
			WithMethod("executeAgentWorkflow", engine.MethodInfo{
				Name: "executeAgentWorkflow",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock workflow execution
				result := map[string]engine.ScriptValue{
					"workflowID": args[0],
					"status":     engine.NewStringValue("completed"),
					"output":     engine.NewStringValue("workflow execution result"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test workflow operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Create workflow
			local workflow, createErr = agent.workflowCreate("sequential", {
				name = "Test Workflow",
				description = "A test sequential workflow"
			})
			assert(createErr == nil, "create workflow should not error: " .. tostring(createErr))
			assert(workflow.type == "sequential", "should be sequential workflow")
			
			-- Add workflow step
			local stepResult, stepErr = agent.workflowAddStep(workflow.id, {
				name = "step1",
				action = "process_input",
				config = { timeout = 30 }
			})
			assert(stepErr == nil, "add step should not error: " .. tostring(stepErr))
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_Hooks(t *testing.T) {
	t.Run("set_agent_hooks", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("setAgentHook", engine.MethodInfo{
				Name: "setAgentHook",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock hook setting
				return engine.NewNilValue(), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test hook operations
		err = L.DoString(`
			local agent = require("agent")
			
			-- Set before run hook
			local beforeHook = function(agentID, input)
				print("Before running agent:", agentID)
				return input
			end
			
			local hookResult, hookErr = agent.hooksSet("test-agent", "beforeRun", beforeHook)
			assert(hookErr == nil, "set hook should not error: " .. tostring(hookErr))
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createAgent", engine.MethodInfo{
				Name: "createAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("agent service unavailable")
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test error handling
		err = L.DoString(`
			local agent = require("agent")
			
			local result, err = agent.createAgent("test-agent", {
				type = "basic",
				name = "Test Agent"
			})
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "agent service unavailable"), "should contain error message")
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_agent_id", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("getAgent", engine.MethodInfo{
				Name: "getAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("agent not found")
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test invalid agent ID handling
		err = L.DoString(`
			local agent = require("agent")
			
			local result, err = agent.getAgent("nonexistent-agent")
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "agent not found"), "should contain not found error")
		`)
		assert.NoError(t, err)
	})
}

func TestAgentAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("agent_constants", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("createAgent", engine.MethodInfo{
				Name: "createAgent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-agent"), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test agent constants
		err = L.DoString(`
			local agent = require("agent")
			
			-- Check that constants are available
			assert(agent.TYPES ~= nil, "TYPES constants should exist")
			assert(agent.TYPES.BASIC == "basic", "basic type should be available")
			assert(agent.TYPES.LLM == "llm", "llm type should be available")
			assert(agent.TYPES.WORKFLOW == "workflow", "workflow type should be available")
			
			assert(agent.HOOKS ~= nil, "HOOKS constants should exist")
			assert(agent.HOOKS.BEFORE_RUN == "beforeRun", "beforeRun hook should be available")
			assert(agent.HOOKS.AFTER_RUN == "afterRun", "afterRun hook should be available")
			
			assert(agent.WORKFLOW_TYPES ~= nil, "WORKFLOW_TYPES constants should exist")
			assert(agent.WORKFLOW_TYPES.SEQUENTIAL == "sequential", "sequential type should be available")
			assert(agent.WORKFLOW_TYPES.PARALLEL == "parallel", "parallel type should be available")
		`)
		assert.NoError(t, err)
	})

	t.Run("metrics_and_utilities", func(t *testing.T) {
		agentBridge := testutils.NewMockBridge("agent").
			WithInitialized(true).
			WithMethod("getAgentMetrics", engine.MethodInfo{
				Name: "getAgentMetrics",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock agent metrics
				metrics := map[string]engine.ScriptValue{
					"execution_count": engine.NewNumberValue(42),
					"success_count":   engine.NewNumberValue(40),
					"error_count":     engine.NewNumberValue(2),
					"avg_duration_ms": engine.NewNumberValue(125),
				}
				return engine.NewObjectValue(metrics), nil
			})

		adapter := NewAgentAdapter(agentBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "agent")
		require.NoError(t, err)

		err = ms.LoadModule(L, "agent")
		require.NoError(t, err)

		// Test metrics and utilities
		err = L.DoString(`
			local agent = require("agent")
			
			local metrics, err = agent.lifecycleGetMetrics("test-agent")
			assert(err == nil, "get metrics should not error: " .. tostring(err))
			assert(metrics.execution_count == 42, "should have execution count")
			assert(metrics.success_count == 40, "should have success count")
			assert(metrics.avg_duration_ms == 125, "should have average duration")
		`)
		assert.NoError(t, err)
	})
}
