// ABOUTME: Comprehensive test suite for Agent Management Library in Lua standard library
// ABOUTME: Tests agent lifecycle, communication, tools, workflows, and integration with promises

package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// MockAgentBridge represents a mock agent bridge for testing
type MockAgentBridge struct {
	agents    map[string]interface{}
	tools     map[string][]interface{}
	callLog   []string
	responses map[string]interface{}
}

// NewMockAgentBridge creates a new mock agent bridge
func NewMockAgentBridge() *MockAgentBridge {
	return &MockAgentBridge{
		agents:    make(map[string]interface{}),
		tools:     make(map[string][]interface{}),
		callLog:   []string{},
		responses: make(map[string]interface{}),
	}
}

// MockWorkflowBridge represents a mock workflow bridge for testing
type MockWorkflowBridge struct {
	workflows map[string]interface{}
	steps     map[string][]interface{}
	callLog   []string
}

// NewMockWorkflowBridge creates a new mock workflow bridge
func NewMockWorkflowBridge() *MockWorkflowBridge {
	return &MockWorkflowBridge{
		workflows: make(map[string]interface{}),
		steps:     make(map[string][]interface{}),
		callLog:   []string{},
	}
}

// setupMockAgentBridges sets up mock agent and workflow bridges in the Lua state
func setupMockAgentBridges(L *lua.LState, agentBridge *MockAgentBridge, workflowBridge *MockWorkflowBridge) {
	// Create mock agent bridge table
	agentBridgeTable := L.NewTable()

	// Agent lifecycle methods
	agentBridgeTable.RawSetString("lifecycleCreate", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		_ = L.OptTable(2, L.NewTable()) // config parameter
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleCreate:"+name)

		agentID := "agent_" + name
		agent := L.NewTable()
		agent.RawSetString("id", lua.LString(agentID))
		agent.RawSetString("name", lua.LString(name))
		agent.RawSetString("state", lua.LString("created"))

		agentBridge.agents[agentID] = agent
		L.Push(agent)
		return 1
	}))

	agentBridgeTable.RawSetString("lifecycleGet", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleGet:"+agentID)

		if agentData, exists := agentBridge.agents[agentID]; exists {
			L.Push(agentData.(*lua.LTable))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	agentBridgeTable.RawSetString("lifecycleList", L.NewFunction(func(L *lua.LState) int {
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleList")

		agents := L.NewTable()
		i := 1
		for _, agentData := range agentBridge.agents {
			agents.RawSetInt(i, agentData.(*lua.LTable))
			i++
		}
		L.Push(agents)
		return 1
	}))

	agentBridgeTable.RawSetString("lifecycleRemove", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleRemove:"+agentID)

		if _, exists := agentBridge.agents[agentID]; exists {
			delete(agentBridge.agents, agentID)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	// Agent execution methods
	agentBridgeTable.RawSetString("run", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		_ = L.CheckAny(2)               // input parameter
		_ = L.OptTable(3, L.NewTable()) // options parameter
		agentBridge.callLog = append(agentBridge.callLog, "run:"+agentID)

		response := L.NewTable()
		response.RawSetString("content", lua.LString("Mock agent response from "+agentID))
		response.RawSetString("success", lua.LTrue)

		L.Push(response)
		return 1
	}))

	agentBridgeTable.RawSetString("runAsync", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		_ = L.CheckAny(2)               // input parameter
		_ = L.OptTable(3, L.NewTable()) // options parameter
		agentBridge.callLog = append(agentBridge.callLog, "runAsync:"+agentID)

		response := L.NewTable()
		response.RawSetString("content", lua.LString("Mock async agent response from "+agentID))
		response.RawSetString("success", lua.LTrue)

		L.Push(response)
		return 1
	}))

	// Agent state methods
	agentBridgeTable.RawSetString("stateGet", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		agentBridge.callLog = append(agentBridge.callLog, "stateGet:"+agentID)

		state := L.NewTable()
		state.RawSetString("agent_id", lua.LString(agentID))
		state.RawSetString("status", lua.LString("running"))

		L.Push(state)
		return 1
	}))

	agentBridgeTable.RawSetString("stateSet", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		_ = L.CheckTable(2) // state parameter
		agentBridge.callLog = append(agentBridge.callLog, "stateSet:"+agentID)

		L.Push(lua.LTrue)
		return 1
	}))

	// Agent tool methods
	agentBridgeTable.RawSetString("registerTool", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		_ = L.CheckAny(2) // tool parameter (could be name or table)
		agentBridge.callLog = append(agentBridge.callLog, "registerTool:"+agentID)

		if agentBridge.tools[agentID] == nil {
			agentBridge.tools[agentID] = []interface{}{}
		}
		agentBridge.tools[agentID] = append(agentBridge.tools[agentID], "mock_tool")

		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("unregisterTool", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		toolName := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "unregisterTool:"+agentID+":"+toolName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("listTools", L.NewFunction(func(L *lua.LState) int {
		agentID := L.CheckString(1)
		agentBridge.callLog = append(agentBridge.callLog, "listTools:"+agentID)

		tools := L.NewTable()
		if agentTools, exists := agentBridge.tools[agentID]; exists {
			for i, tool := range agentTools {
				toolTable := L.NewTable()
				toolTable.RawSetString("name", lua.LString(tool.(string)))
				tools.RawSetInt(i+1, toolTable)
			}
		}

		L.Push(tools)
		return 1
	}))

	// Add missing AgentAdapter bridge methods for complete testing coverage

	// Lifecycle methods
	agentBridgeTable.RawSetString("lifecycleCreateLLM", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		name := L.CheckString(2)
		_ = L.OptTable(3, L.NewTable()) // config parameter
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleCreateLLM:"+name)

		agentID := "llm_agent_" + name
		agent := L.NewTable()
		agent.RawSetString("id", lua.LString(agentID))
		agent.RawSetString("name", lua.LString(name))
		agent.RawSetString("type", lua.LString("llm"))
		agent.RawSetString("state", lua.LString("created"))

		agentBridge.agents[agentID] = agent
		L.Push(agent)
		return 1
	}))

	agentBridgeTable.RawSetString("lifecycleGetMetrics", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentBridge.callLog = append(agentBridge.callLog, "lifecycleGetMetrics")
		metrics := L.NewTable()
		metrics.RawSetString("total_agents", lua.LNumber(len(agentBridge.agents)))
		metrics.RawSetString("active_agents", lua.LNumber(1))
		L.Push(metrics)
		return 1
	}))

	// State methods
	agentBridgeTable.RawSetString("stateExport", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "stateExport:"+agentID)
		exportData := L.NewTable()
		exportData.RawSetString("agent_id", lua.LString(agentID))
		exportData.RawSetString("exported_at", lua.LNumber(1234567890))
		L.Push(exportData)
		return 1
	}))

	agentBridgeTable.RawSetString("stateImport", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		_ = L.CheckTable(3) // state_data parameter
		agentBridge.callLog = append(agentBridge.callLog, "stateImport:"+agentID)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("stateSaveSnapshot", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		snapshotName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "stateSaveSnapshot:"+agentID+":"+snapshotName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("stateLoadSnapshot", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		snapshotName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "stateLoadSnapshot:"+agentID+":"+snapshotName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("stateListSnapshots", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "stateListSnapshots:"+agentID)
		snapshots := L.NewTable()
		snapshots.RawSetInt(1, lua.LString("snapshot1"))
		snapshots.RawSetInt(2, lua.LString("snapshot2"))
		L.Push(snapshots)
		return 1
	}))

	// Events methods
	agentBridgeTable.RawSetString("eventsEmit", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		eventName := L.CheckString(3)
		_ = L.CheckAny(4) // event_data parameter
		agentBridge.callLog = append(agentBridge.callLog, "eventsEmit:"+agentID+":"+eventName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("eventsSubscribe", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		eventName := L.CheckString(3)
		_ = L.CheckFunction(4) // handler parameter
		agentBridge.callLog = append(agentBridge.callLog, "eventsSubscribe:"+agentID+":"+eventName)
		L.Push(lua.LString("subscription_id_123"))
		return 1
	}))

	agentBridgeTable.RawSetString("eventsUnsubscribe", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		eventName := L.CheckString(3)
		handlerID := L.CheckString(4)
		agentBridge.callLog = append(agentBridge.callLog, "eventsUnsubscribe:"+agentID+":"+eventName+":"+handlerID)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("eventsStartRecording", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "eventsStartRecording:"+agentID)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("eventsStopRecording", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "eventsStopRecording:"+agentID)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("eventsReplay", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		_ = L.CheckTable(3) // events parameter
		agentBridge.callLog = append(agentBridge.callLog, "eventsReplay:"+agentID)
		L.Push(lua.LTrue)
		return 1
	}))

	// Profiling methods
	agentBridgeTable.RawSetString("profilingStart", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		profileName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "profilingStart:"+agentID+":"+profileName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("profilingStop", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		profileName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "profilingStop:"+agentID+":"+profileName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("profilingGetMetrics", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "profilingGetMetrics:"+agentID)
		metrics := L.NewTable()
		metrics.RawSetString("cpu_usage", lua.LNumber(25.5))
		metrics.RawSetString("memory_usage", lua.LNumber(128))
		L.Push(metrics)
		return 1
	}))

	agentBridgeTable.RawSetString("profilingGetReport", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		profileName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "profilingGetReport:"+agentID+":"+profileName)
		report := L.NewTable()
		report.RawSetString("profile_name", lua.LString(profileName))
		report.RawSetString("duration", lua.LNumber(1500))
		L.Push(report)
		return 1
	}))

	// Hooks methods
	agentBridgeTable.RawSetString("hooksRegister", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		hookName := L.CheckString(3)
		_ = L.CheckFunction(4) // hook_function parameter
		agentBridge.callLog = append(agentBridge.callLog, "hooksRegister:"+agentID+":"+hookName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("hooksUnregister", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		hookName := L.CheckString(3)
		agentBridge.callLog = append(agentBridge.callLog, "hooksUnregister:"+agentID+":"+hookName)
		L.Push(lua.LTrue)
		return 1
	}))

	agentBridgeTable.RawSetString("hooksExecute", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		hookName := L.CheckString(3)
		_ = L.CheckTable(4) // context parameter
		agentBridge.callLog = append(agentBridge.callLog, "hooksExecute:"+agentID+":"+hookName)
		result := L.NewTable()
		result.RawSetString("hook_executed", lua.LTrue)
		L.Push(result)
		return 1
	}))

	agentBridgeTable.RawSetString("hooksList", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		agentID := L.CheckString(2)
		agentBridge.callLog = append(agentBridge.callLog, "hooksList:"+agentID)
		hooks := L.NewTable()
		hooks.RawSetInt(1, lua.LString("pre_run"))
		hooks.RawSetInt(2, lua.LString("post_run"))
		L.Push(hooks)
		return 1
	}))

	// Add workflow bridge methods to agent bridge for workflow testing
	agentBridgeTable.RawSetString("workflowAddStep", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		workflowID := L.CheckString(2)
		_ = L.CheckTable(3) // step parameter
		_ = L.OptNumber(4, 0) // position parameter
		agentBridge.callLog = append(agentBridge.callLog, "workflowAddStep:"+workflowID)
		L.Push(lua.LTrue)
		return 1
	}))

	// Create mock workflow bridge table
	workflowBridgeTable := L.NewTable()

	workflowBridgeTable.RawSetString("create", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		_ = L.OptTable(2, L.NewTable()) // config parameter
		workflowBridge.callLog = append(workflowBridge.callLog, "create:"+name)

		workflowID := "workflow_" + name
		workflow := L.NewTable()
		workflow.RawSetString("id", lua.LString(workflowID))
		workflow.RawSetString("name", lua.LString(name))
		workflow.RawSetString("state", lua.LString("created"))

		workflowBridge.workflows[workflowID] = workflow
		L.Push(workflow)
		return 1
	}))

	workflowBridgeTable.RawSetString("addStep", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		_ = L.CheckTable(2) // step parameter
		workflowBridge.callLog = append(workflowBridge.callLog, "addStep:"+workflowID)

		if workflowBridge.steps[workflowID] == nil {
			workflowBridge.steps[workflowID] = []interface{}{}
		}
		workflowBridge.steps[workflowID] = append(workflowBridge.steps[workflowID], "mock_step")

		L.Push(lua.LTrue)
		return 1
	}))

	workflowBridgeTable.RawSetString("execute", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		_ = L.CheckAny(2)               // input parameter
		_ = L.OptTable(3, L.NewTable()) // options parameter
		workflowBridge.callLog = append(workflowBridge.callLog, "execute:"+workflowID)

		response := L.NewTable()
		response.RawSetString("workflow_id", lua.LString(workflowID))
		response.RawSetString("result", lua.LString("Mock workflow execution result"))
		response.RawSetString("success", lua.LTrue)

		L.Push(response)
		return 1
	}))

	workflowBridgeTable.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		workflowBridge.callLog = append(workflowBridge.callLog, "get:"+workflowID)

		if workflowData, exists := workflowBridge.workflows[workflowID]; exists {
			L.Push(workflowData.(*lua.LTable))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	// Set up mock bridges in bridges table
	bridgesTable := L.GetGlobal("bridges")
	if bridgesTable == lua.LNil {
		bridgesTable = L.NewTable()
		L.SetGlobal("bridges", bridgesTable)
	}
	bridgesTable.(*lua.LTable).RawSetString("agent_core", agentBridgeTable)
	bridgesTable.(*lua.LTable).RawSetString("agent_workflow", workflowBridgeTable)
}

// setupAgentLibrary loads the agent library and its dependencies
func setupAgentLibrary(t *testing.T, L *lua.LState) (*MockAgentBridge, *MockWorkflowBridge) {
	t.Helper()

	// Setup mock bridges first
	agentBridge := NewMockAgentBridge()
	workflowBridge := NewMockWorkflowBridge()
	setupMockAgentBridges(L, agentBridge, workflowBridge)

	// Load promise library (required dependency)
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	// Load agent library
	agentPath := filepath.Join(".", "agent.lua")
	err = L.DoFile(agentPath)
	if err != nil {
		t.Fatalf("Failed to load agent library: %v", err)
	}
	agentLib := L.Get(-1)
	L.SetGlobal("agent", agentLib)

	return agentBridge, workflowBridge
}

// TestAgentLibraryLoading tests that the agent library can be loaded
func TestAgentLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	agentBridge, workflowBridge := setupAgentLibrary(t, L)

	// Check that agent table exists and has expected structure
	script := `
		if type(agent) ~= "table" then
			error("Agent module should be a table")
		end
		
		if type(agent.create) ~= "function" then
			error("agent.create function should be available")
		end
		
		if type(agent.run) ~= "function" then
			error("agent.run function should be available")
		end
		
		if type(agent.conversation) ~= "function" then
			error("agent.conversation function should be available")
		end
		
		if type(agent.workflow_create) ~= "function" then
			error("agent.workflow_create function should be available")
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Agent library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}

	_ = agentBridge    // Use to avoid unused variable warning
	_ = workflowBridge // Use to avoid unused variable warning
}

// TestAgentLifecycle tests agent creation, configuration, and removal
func TestAgentLifecycle(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge)
	}{
		{
			name: "create_agent",
			script: `
				local new_agent = agent.create("test_agent", {role = "assistant"})
				return new_agent and new_agent.id
			`,
			check: func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge) {
				if result.String() != "agent_test_agent" {
					t.Errorf("Expected agent ID 'agent_test_agent', got %v", result.String())
				}
			},
		},
		{
			name: "configure_agent",
			script: `
				local new_agent = agent.create("test_agent", {role = "assistant"})
				local success = agent.configure(new_agent.id, {temperature = 0.8})
				return success
			`,
			check: func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected agent configuration to succeed")
				}
			},
		},
		{
			name: "clone_agent",
			script: `
				local source_agent = agent.create("source_agent", {role = "assistant"})
				local cloned_agent = agent.clone(source_agent.id, "cloned_agent", {temperature = 0.9})
				return cloned_agent and cloned_agent.id
			`,
			check: func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge) {
				if result.String() != "agent_cloned_agent" {
					t.Errorf("Expected cloned agent ID 'agent_cloned_agent', got %v", result.String())
				}
			},
		},
		{
			name: "list_agents",
			script: `
				agent.create("agent1")
				agent.create("agent2")
				local agents = agent.list()
				return type(agents) == "table" and #agents >= 2
			`,
			check: func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected agents list to contain at least 2 agents")
				}
			},
		},
		{
			name: "remove_agent",
			script: `
				local new_agent = agent.create("temp_agent")
				local success = agent.remove(new_agent.id)
				return success
			`,
			check: func(t *testing.T, result lua.LValue, agentBridge *MockAgentBridge) {
				if result != lua.LTrue {
					t.Errorf("Expected agent removal to succeed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			agentBridge, _ := setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result, agentBridge)
		})
	}
}

// TestAgentExecution tests agent running functionality
func TestAgentExecution(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "run_agent_sync",
			script: `
				local test_agent = agent.create("test_agent")
				local response = agent.run(test_agent.id, "test input")
				return response and response.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				expectedContent := "Mock agent response from agent_test_agent"
				if result.String() != expectedContent {
					t.Errorf("Expected '%s', got %v", expectedContent, result.String())
				}
			},
		},
		{
			name: "run_agent_async",
			script: `
				local test_agent = agent.create("test_agent")
				local promise = agent.run_async(test_agent.id, "test input")
				return type(promise) == "table" and type(promise.andThen) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected async run to return a promise")
				}
			},
		},
		{
			name: "agent_status",
			script: `
				local test_agent = agent.create("test_agent")
				agent.run(test_agent.id, "test input")
				local status = agent.get_status(test_agent.id)
				return status and status.exists
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected agent status to show agent exists")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAgentCommunication tests conversation and collaboration functionality
func TestAgentCommunication(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_conversation",
			script: `
				local test_agent = agent.create("chat_agent")
				local conversation = agent.conversation(test_agent.id, "You are helpful")
				return type(conversation) == "table" and type(conversation.send) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected conversation to be created with send method")
				}
			},
		},
		{
			name: "conversation_send",
			script: `
				local test_agent = agent.create("chat_agent")
				local conversation = agent.conversation(test_agent.id)
				local response = conversation:send("Hello!")
				return response and response.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				expectedContent := "Mock agent response from agent_chat_agent"
				if result.String() != expectedContent {
					t.Errorf("Expected '%s', got %v", expectedContent, result.String())
				}
			},
		},
		{
			name: "delegate_task",
			script: `
				local agent1 = agent.create("agent1")
				local agent2 = agent.create("agent2")
				local result = agent.delegate(agent1.id, agent2.id, "test task")
				return result and result.content
			`,
			check: func(t *testing.T, result lua.LValue) {
				expectedContent := "Mock agent response from agent_agent2"
				if result.String() != expectedContent {
					t.Errorf("Expected '%s', got %v", expectedContent, result.String())
				}
			},
		},
		{
			name: "collaborate_sequential",
			script: `
				local agent1 = agent.create("agent1")
				local agent2 = agent.create("agent2")
				local result = agent.collaborate({agent1.id, agent2.id}, "collaborative task")
				return result and result.coordination_method == "sequential"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected sequential collaboration to work")
				}
			},
		},
		{
			name: "collaborate_parallel",
			script: `
				local agent1 = agent.create("agent1")
				local agent2 = agent.create("agent2")
				local result = agent.collaborate({agent1.id, agent2.id}, "parallel task", {coordination_method = "parallel"})
				return type(result) == "table" and result.coordination_method == "parallel"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parallel collaboration to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAgentTools tests tool integration functionality
func TestAgentTools(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_tool",
			script: `
				local tool = agent.create_tool("test_tool", function(input) return "result" end, {description = "Test tool"})
				return tool and tool.name == "test_tool" and type(tool.execute) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tool to be created with correct properties")
				}
			},
		},
		{
			name: "add_tools_to_agent",
			script: `
				local test_agent = agent.create("tool_agent")
				local tool1 = agent.create_tool("tool1", function() return "1" end)
				local tool2 = agent.create_tool("tool2", function() return "2" end)
				local result = agent.add_tools(test_agent.id, {tool1, tool2})
				return result and result.success and result.total_tools == 2
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tools to be added successfully")
				}
			},
		},
		{
			name: "get_agent_tools",
			script: `
				local test_agent = agent.create("tool_agent")
				local tool = agent.create_tool("test_tool", function() return "result" end)
				agent.add_tools(test_agent.id, {tool})
				local tools = agent.get_tools(test_agent.id)
				return type(tools) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected to get agent tools")
				}
			},
		},
		{
			name: "tool_chain",
			script: `
				local tool1 = function(input) return input .. "_1" end
				local tool2 = function(input) return input .. "_2" end
				local result = agent.tool_chain({tool1, tool2}, "start")
				return result and result.final_output == "start_1_2"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tool chain to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestWorkflowOrchestration tests workflow functionality
func TestWorkflowOrchestration(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_workflow",
			script: `
				local steps = {
					{type = "agent", agent_id = "test_agent"},
					{type = "function", func = function() return "step2" end}
				}
				local workflow = agent.workflow_create("test_workflow", steps)
				return workflow and workflow.id
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "workflow_test_workflow" {
					t.Errorf("Expected workflow ID 'workflow_test_workflow', got %v", result.String())
				}
			},
		},
		{
			name: "run_workflow",
			script: `
				local steps = {{type = "simple"}}
				local workflow = agent.workflow_create("test_workflow", steps)
				local result = agent.workflow_run(workflow.id, "test input")
				return result and result.success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected workflow execution to succeed")
				}
			},
		},
		{
			name: "workflow_parallel",
			script: `
				local step1 = function(input) return input .. "_1" end
				local step2 = function(input) return input .. "_2" end
				local promise = agent.workflow_parallel({step1, step2}, "test")
				return type(promise) == "table" and type(promise.andThen) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parallel workflow to return a promise")
				}
			},
		},
		{
			name: "workflow_conditional",
			script: `
				local condition = function(input) return input == "yes" end
				local then_step = function(input) return "executed_then" end
				local else_step = function(input) return "executed_else" end
				
				local result1 = agent.workflow_conditional(condition, then_step, else_step, "yes")
				local result2 = agent.workflow_conditional(condition, then_step, else_step, "no")
				
				return result1.result == "executed_then" and result2.result == "executed_else"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected conditional workflow to work correctly")
				}
			},
		},
		{
			name: "workflow_status",
			script: `
				local steps = {{type = "simple"}}
				local workflow = agent.workflow_create("status_test", steps)
				local status = agent.get_workflow_status(workflow.id)
				return status and status.exists
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected workflow status to show workflow exists")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAgentErrorHandling tests error handling scenarios
func TestAgentErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_agent_name_error",
			script: `
				local success, err = pcall(function()
					agent.create()
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing agent name")
				}
			},
		},
		{
			name: "invalid_agent_id_error",
			script: `
				local success, err = pcall(function()
					agent.run()
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing agent ID")
				}
			},
		},
		{
			name: "invalid_tools_type_error",
			script: `
				local test_agent = agent.create("test_agent")
				local success, err = pcall(function()
					agent.add_tools(test_agent.id, "not a table")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid tools type")
				}
			},
		},
		{
			name: "invalid_tool_function_error",
			script: `
				local success, err = pcall(function()
					agent.create_tool("test", "not a function")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid tool function")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAgentLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// BenchmarkAgentOperations benchmarks basic agent operations
func BenchmarkAgentOperations(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	// Setup agent library for benchmarking
	agentBridge := NewMockAgentBridge()
	workflowBridge := NewMockWorkflowBridge()
	setupMockAgentBridges(L, agentBridge, workflowBridge)

	// Load promise library (required dependency)
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		b.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	// Load agent library
	agentPath := filepath.Join(".", "agent.lua")
	err = L.DoFile(agentPath)
	if err != nil {
		b.Fatalf("Failed to load agent library: %v", err)
	}
	agentLib := L.Get(-1)
	L.SetGlobal("agent", agentLib)

	script := `
		local test_agent = agent.create("benchmark_agent")
		local response = agent.run(test_agent.id, "benchmark input")
		return response
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := L.DoString(script)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		L.Pop(1) // Clean stack
	}
}

// TestAgentPackageRequire tests that the agent module can be required as a package
func TestAgentPackageRequire(t *testing.T) {
	// Change to the stdlib directory for testing
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(filepath.Dir(filepath.Join(wd, "agent.lua")))
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	// Setup mock bridges and dependencies
	agentBridge := NewMockAgentBridge()
	workflowBridge := NewMockWorkflowBridge()
	setupMockAgentBridges(L, agentBridge, workflowBridge)

	// Load promise library first
	promisePath := filepath.Join(".", "promise.lua")
	err = L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	script := `
		local agent = require('agent')
		
		-- Test that the module loads correctly
		if type(agent) ~= "table" then
			error("Agent module should return a table")
		end
		
		if type(agent.create) ~= "function" then
			error("create function should be available")
		end
		
		if type(agent.run) ~= "function" then
			error("run function should be available")
		end
		
		if type(agent.conversation) ~= "function" then
			error("conversation function should be available")
		end
		
		return true
	`

	err = L.DoString(script)
	if err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected require test to pass, got %v", result)
	}
}
