-- Test enhanced agent.lua with all AgentAdapter methods
print("=== Testing Enhanced Agent Module ===")

local agent = require("agent")
print("Agent module loaded successfully")

-- Test that new methods exist
print("\n--- Testing New AgentAdapter Methods ---")
local methods_to_test = {
    -- Standard convenience methods
    "createAgent", "createLLMAgent", "listAgents", "getAgent", "removeAgent",
    -- Lifecycle methods
    "lifecycleCreate", "lifecycleCreateLLM", "lifecycleList", "lifecycleGet", "lifecycleRemove", "lifecycleGetMetrics",
    -- Tool methods
    "registerTool", "unregisterTool", "listTools",
    -- State methods
    "stateGet", "stateSet", "stateExport", "stateImport", 
    "stateSaveSnapshot", "stateLoadSnapshot", "stateListSnapshots",
    -- Events methods
    "eventsEmit", "eventsSubscribe", "eventsUnsubscribe", 
    "eventsStartRecording", "eventsStopRecording", "eventsReplay",
    -- Profiling methods
    "profilingStart", "profilingStop", "profilingGetMetrics", "profilingGetReport",
    -- Workflow methods
    "workflowCreate", "workflowExecute", "workflowAddStep",
    -- Hooks methods
    "hooksRegister", "hooksUnregister", "hooksExecute", "hooksList"
}

local missing_methods = {}
local found_methods = {}

for _, method_name in ipairs(methods_to_test) do
    if type(agent[method_name]) == "function" then
        table.insert(found_methods, method_name)
    else
        table.insert(missing_methods, method_name)
    end
end

print("Found methods: " .. #found_methods)
print("Missing methods: " .. #missing_methods)

if #missing_methods > 0 then
    print("Missing methods list:")
    for _, method in ipairs(missing_methods) do
        print("  - " .. method)
    end
end

-- Test namespaces
print("\n--- Testing Namespaces ---")
local namespaces = {"state", "events", "profiling", "hooks"}
local namespace_results = {}

for _, ns in ipairs(namespaces) do
    namespace_results[ns] = type(agent[ns]) == "table"
    print(ns .. " namespace: " .. tostring(namespace_results[ns]))
end

print("\n=== Enhanced Agent Test Complete ===")
return {
    success = true,
    total_methods_tested = #methods_to_test,
    found_methods = #found_methods,
    missing_methods = #missing_methods,
    namespaces_working = namespace_results,
    agent_adapter_complete = #missing_methods == 0
}