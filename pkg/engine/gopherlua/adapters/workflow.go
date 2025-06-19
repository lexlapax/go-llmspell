// ABOUTME: Workflow bridge adapter that exposes go-llms workflow functionality to Lua scripts
// ABOUTME: Provides workflow creation, execution, step management, templates, and serialization capabilities

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// WorkflowAdapter exposes workflow bridge functionality to Lua
type WorkflowAdapter struct {
	*gopherlua.BridgeAdapter
}

// NewWorkflowAdapter creates a new workflow adapter
func NewWorkflowAdapter(bridge engine.Bridge) *WorkflowAdapter {
	adapter := &WorkflowAdapter{
		BridgeAdapter: gopherlua.NewBridgeAdapter(bridge),
	}
	return adapter
}

// GetMethods returns the list of methods exposed by the underlying bridge
func (wa *WorkflowAdapter) GetMethods() []string {
	if wa.BridgeAdapter == nil || wa.GetBridge() == nil {
		return []string{}
	}

	bridgeMethods := wa.BridgeAdapter.GetBridge().Methods()
	methods := make([]string, len(bridgeMethods))
	for i, method := range bridgeMethods {
		methods[i] = method.Name
	}
	return methods
}

// CreateLuaModule creates a Lua module with workflow-specific enhancements
func (wa *WorkflowAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if wa.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := wa.BridgeAdapter.CreateLuaModule()
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
		L.SetField(module, "_adapter", lua.LString("workflow"))
		L.SetField(module, "_version", lua.LString("2.1.0"))

		// Add workflow-specific enhancements
		wa.addWorkflowConstants(L, module)
		wa.addWorkflowMethods(L, module)

		// Push module
		L.Push(module)
		return 1
	}
}

// addWorkflowConstants adds workflow-related constants
func (wa *WorkflowAdapter) addWorkflowConstants(L *lua.LState, module *lua.LTable) {
	// Add workflow type constants
	types := L.NewTable()
	L.SetField(types, "SEQUENTIAL", lua.LString("sequential"))
	L.SetField(types, "PARALLEL", lua.LString("parallel"))
	L.SetField(types, "CONDITIONAL", lua.LString("conditional"))
	L.SetField(types, "LOOP", lua.LString("loop"))
	L.SetField(types, "CUSTOM", lua.LString("custom"))
	L.SetField(module, "TYPES", types)

	// Add workflow states (the test expects STATUS not STATES)
	states := L.NewTable()
	L.SetField(states, "CREATED", lua.LString("created"))
	L.SetField(states, "PENDING", lua.LString("pending"))
	L.SetField(states, "RUNNING", lua.LString("running"))
	L.SetField(states, "PAUSED", lua.LString("paused"))
	L.SetField(states, "COMPLETED", lua.LString("completed"))
	L.SetField(states, "FAILED", lua.LString("failed"))
	L.SetField(states, "CANCELLED", lua.LString("cancelled"))
	L.SetField(module, "STATUS", states)

	// Add export/import formats
	formats := L.NewTable()
	L.SetField(formats, "JSON", lua.LString("json"))
	L.SetField(formats, "YAML", lua.LString("yaml"))
	L.SetField(module, "FORMATS", formats)

	// Add step types
	stepTypes := L.NewTable()
	L.SetField(stepTypes, "AGENT", lua.LString("agent"))
	L.SetField(stepTypes, "SCRIPT", lua.LString("script"))
	L.SetField(stepTypes, "CONDITION", lua.LString("condition"))
	L.SetField(stepTypes, "PARALLEL", lua.LString("parallel"))
	L.SetField(stepTypes, "WAIT", lua.LString("wait"))
	L.SetField(module, "STEP_TYPES", stepTypes)
}

// addWorkflowMethods adds workflow-specific methods to the module
func (wa *WorkflowAdapter) addWorkflowMethods(L *lua.LState, module *lua.LTable) {

	// Workflow lifecycle methods
	L.SetField(module, "createWorkflow", L.NewFunction(wa.createWorkflow))
	L.SetField(module, "executeWorkflow", L.NewFunction(wa.executeWorkflow))
	L.SetField(module, "pauseWorkflow", L.NewFunction(wa.pauseWorkflow))
	L.SetField(module, "resumeWorkflow", L.NewFunction(wa.resumeWorkflow))
	L.SetField(module, "stopWorkflow", L.NewFunction(wa.stopWorkflow))
	L.SetField(module, "getWorkflowStatus", L.NewFunction(wa.getWorkflowStatus))
	L.SetField(module, "listWorkflows", L.NewFunction(wa.listWorkflows))
	L.SetField(module, "getWorkflow", L.NewFunction(wa.getWorkflow))
	L.SetField(module, "deleteWorkflow", L.NewFunction(wa.deleteWorkflow))

	// Step management methods
	L.SetField(module, "addStep", L.NewFunction(wa.addStep))
	L.SetField(module, "removeStep", L.NewFunction(wa.removeStep))
	L.SetField(module, "updateStep", L.NewFunction(wa.updateStep))
	L.SetField(module, "getStep", L.NewFunction(wa.getStep))
	L.SetField(module, "listSteps", L.NewFunction(wa.listSteps))
	L.SetField(module, "reorderSteps", L.NewFunction(wa.reorderSteps))

	// Template methods
	L.SetField(module, "listTemplates", L.NewFunction(wa.listTemplates))
	L.SetField(module, "listWorkflowTemplates", L.NewFunction(wa.listWorkflowTemplates))
	L.SetField(module, "getTemplate", L.NewFunction(wa.getTemplate))
	L.SetField(module, "getWorkflowTemplate", L.NewFunction(wa.getWorkflowTemplate))
	L.SetField(module, "createWorkflowTemplate", L.NewFunction(wa.createWorkflowTemplate))
	L.SetField(module, "removeWorkflowTemplate", L.NewFunction(wa.removeWorkflowTemplate))
	L.SetField(module, "createWorkflowFromTemplate", L.NewFunction(wa.createWorkflowFromTemplate))
	L.SetField(module, "saveAsTemplate", L.NewFunction(wa.saveAsTemplate))

	// Import/Export methods
	L.SetField(module, "exportWorkflow", L.NewFunction(wa.exportWorkflow))
	L.SetField(module, "importWorkflow", L.NewFunction(wa.importWorkflow))

	// Variable management
	L.SetField(module, "setWorkflowVariable", L.NewFunction(wa.setWorkflowVariable))
	L.SetField(module, "getWorkflowVariable", L.NewFunction(wa.getWorkflowVariable))
	L.SetField(module, "listWorkflowVariables", L.NewFunction(wa.listWorkflowVariables))
	L.SetField(module, "removeWorkflowVariable", L.NewFunction(wa.removeWorkflowVariable))

	// Error handling
	L.SetField(module, "getWorkflowErrors", L.NewFunction(wa.getWorkflowErrors))
	L.SetField(module, "clearWorkflowErrors", L.NewFunction(wa.clearWorkflowErrors))

	// Convenience methods
	L.SetField(module, "createBuilder", L.NewFunction(wa.createBuilder))
	L.SetField(module, "validateWorkflow", L.NewFunction(wa.validateWorkflow))
}

// Workflow lifecycle methods

func (wa *WorkflowAdapter) createWorkflow(L *lua.LState) int {
	id := L.CheckString(1)
	config := L.CheckTable(2)

	// Convert config to ScriptValue
	configValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, config)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(id),
		configValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "createWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) executeWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)
	input := L.Get(2) // Optional input

	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	if input != lua.LNil {
		inputValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, input)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		args = append(args, inputValue)
	}

	ctx := context.Background()
	result, err := wa.GetBridge().ExecuteMethod(ctx, "executeWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) pauseWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "pauseWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) resumeWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "resumeWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) stopWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "stopWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getWorkflowStatus(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getWorkflowStatus", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) listWorkflows(L *lua.LState) int {
	ctx := context.Background()
	result, err := wa.GetBridge().ExecuteMethod(ctx, "listWorkflows", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayValue, ok := result.(engine.ArrayValue); ok {
		for _, item := range arrayValue.Elements() {
			luaItem, err := wa.GetTypeConverter().FromLuaScriptValue(L, item)
			if err != nil {
				L.Push(lua.LNil)
				continue
			}
			L.Push(luaItem)
		}
		return len(arrayValue.Elements())
	}

	// Fallback to table
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) deleteWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "deleteWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Step management methods

func (wa *WorkflowAdapter) addStep(L *lua.LState) int {
	workflowID := L.CheckString(1)
	step := L.CheckTable(2)

	// Convert step to ScriptValue
	stepValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, step)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		stepValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "addStep", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) removeStep(L *lua.LState) int {
	workflowID := L.CheckString(1)
	stepID := L.CheckString(2)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(stepID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "removeStep", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) updateStep(L *lua.LState) int {
	workflowID := L.CheckString(1)
	stepID := L.CheckString(2)
	updates := L.CheckTable(3)

	// Convert updates to ScriptValue
	updatesValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, updates)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(stepID),
		updatesValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "updateStep", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getStep(L *lua.LState) int {
	workflowID := L.CheckString(1)
	stepID := L.CheckString(2)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(stepID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getStep", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) listSteps(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "listSteps", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayValue, ok := result.(engine.ArrayValue); ok {
		for _, item := range arrayValue.Elements() {
			luaItem, err := wa.GetTypeConverter().FromLuaScriptValue(L, item)
			if err != nil {
				L.Push(lua.LNil)
				continue
			}
			L.Push(luaItem)
		}
		return len(arrayValue.Elements())
	}

	// Fallback to table
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) reorderSteps(L *lua.LState) int {
	workflowID := L.CheckString(1)
	stepIDs := L.CheckTable(2)

	// Convert step IDs to ScriptValue
	stepIDsValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, stepIDs)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		stepIDsValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "reorderSteps", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Template methods

func (wa *WorkflowAdapter) listWorkflowTemplates(L *lua.LState) int {
	ctx := context.Background()
	result, err := wa.GetBridge().ExecuteMethod(ctx, "listWorkflowTemplates", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayValue, ok := result.(engine.ArrayValue); ok {
		for _, item := range arrayValue.Elements() {
			luaItem, err := wa.GetTypeConverter().FromLuaScriptValue(L, item)
			if err != nil {
				L.Push(lua.LNil)
				continue
			}
			L.Push(luaItem)
		}
		return len(arrayValue.Elements())
	}

	// Fallback to table
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) listTemplates(L *lua.LState) int {
	ctx := context.Background()
	result, err := wa.GetBridge().ExecuteMethod(ctx, "listTemplates", []engine.ScriptValue{})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayValue, ok := result.(engine.ArrayValue); ok {
		for _, item := range arrayValue.Elements() {
			luaItem, err := wa.GetTypeConverter().FromLuaScriptValue(L, item)
			if err != nil {
				L.Push(lua.LNil)
				continue
			}
			L.Push(luaItem)
		}
		return len(arrayValue.Elements())
	}

	// Fallback to table
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getWorkflowTemplate(L *lua.LState) int {
	templateID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(templateID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getWorkflowTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getTemplate(L *lua.LState) int {
	templateID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(templateID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) createWorkflowTemplate(L *lua.LState) int {
	workflowID := L.CheckString(1)
	templateName := L.CheckString(2)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(templateName),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "createWorkflowTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) removeWorkflowTemplate(L *lua.LState) int {
	templateID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(templateID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "removeWorkflowTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) createWorkflowFromTemplate(L *lua.LState) int {
	templateID := L.CheckString(1)
	workflowID := L.CheckString(2)
	variables := L.Get(3) // Optional variables

	args := []engine.ScriptValue{
		engine.NewStringValue(templateID),
		engine.NewStringValue(workflowID),
	}

	if variables != lua.LNil {
		varsValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, variables)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		args = append(args, varsValue)
	}

	ctx := context.Background()
	result, err := wa.GetBridge().ExecuteMethod(ctx, "createWorkflowFromTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) saveAsTemplate(L *lua.LState) int {
	workflowID := L.CheckString(1)
	templateConfig := L.CheckTable(2)

	// Convert template config to ScriptValue
	configValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, templateConfig)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		configValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "saveAsTemplate", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Import/Export methods

func (wa *WorkflowAdapter) exportWorkflow(L *lua.LState) int {
	workflowID := L.CheckString(1)
	format := L.OptString(2, "json")

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(format),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "exportWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) importWorkflow(L *lua.LState) int {
	data := L.CheckString(1)
	format := L.OptString(2, "json")

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(data),
		engine.NewStringValue(format),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "importWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Variable management

func (wa *WorkflowAdapter) setWorkflowVariable(L *lua.LState) int {
	workflowID := L.CheckString(1)
	name := L.CheckString(2)
	value := L.Get(3)

	// Convert value to ScriptValue
	valueScriptValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(name),
		valueScriptValue,
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "setWorkflowVariable", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) getWorkflowVariable(L *lua.LState) int {
	workflowID := L.CheckString(1)
	name := L.CheckString(2)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(name),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getWorkflowVariable", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) listWorkflowVariables(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "listWorkflowVariables", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) removeWorkflowVariable(L *lua.LState) int {
	workflowID := L.CheckString(1)
	name := L.CheckString(2)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
		engine.NewStringValue(name),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "removeWorkflowVariable", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Error handling

func (wa *WorkflowAdapter) getWorkflowErrors(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "getWorkflowErrors", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert array result to multiple return values
	if arrayValue, ok := result.(engine.ArrayValue); ok {
		for _, item := range arrayValue.Elements() {
			luaItem, err := wa.GetTypeConverter().FromLuaScriptValue(L, item)
			if err != nil {
				L.Push(lua.LNil)
				continue
			}
			L.Push(luaItem)
		}
		return len(arrayValue.Elements())
	}

	// Fallback to table
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

func (wa *WorkflowAdapter) clearWorkflowErrors(L *lua.LState) int {
	workflowID := L.CheckString(1)

	ctx := context.Background()
	args := []engine.ScriptValue{
		engine.NewStringValue(workflowID),
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "clearWorkflowErrors", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}

// Convenience methods

func (wa *WorkflowAdapter) createBuilder(L *lua.LState) int {
	workflowID := L.CheckString(1)

	// Create builder table with chaining methods
	builder := L.NewTable()

	// Store workflow ID in builder
	L.SetField(builder, "_workflowID", lua.LString(workflowID))

	// Add type method
	L.SetField(builder, "withType", L.NewFunction(func(L *lua.LState) int {
		builder := L.CheckTable(1)
		workflowType := L.CheckString(2)
		L.SetField(builder, "_type", lua.LString(workflowType))
		L.Push(builder) // Return self for chaining
		return 1
	}))

	// Add name method
	L.SetField(builder, "withName", L.NewFunction(func(L *lua.LState) int {
		builder := L.CheckTable(1)
		name := L.CheckString(2)
		L.SetField(builder, "_name", lua.LString(name))
		L.Push(builder) // Return self for chaining
		return 1
	}))

	// Add description method
	L.SetField(builder, "withDescription", L.NewFunction(func(L *lua.LState) int {
		builder := L.CheckTable(1)
		description := L.CheckString(2)
		L.SetField(builder, "_description", lua.LString(description))
		L.Push(builder) // Return self for chaining
		return 1
	}))

	// Add step method
	L.SetField(builder, "addStep", L.NewFunction(func(L *lua.LState) int {
		builder := L.CheckTable(1)
		step := L.CheckTable(2)

		// Get or create steps array
		stepsValue := L.GetField(builder, "_steps")
		var steps *lua.LTable
		if stepsValue == lua.LNil {
			steps = L.NewTable()
			L.SetField(builder, "_steps", steps)
		} else {
			steps = stepsValue.(*lua.LTable)
		}

		// Add step to array
		steps.Append(step)

		L.Push(builder) // Return self for chaining
		return 1
	}))

	// Add build method
	L.SetField(builder, "build", L.NewFunction(func(L *lua.LState) int {
		builder := L.CheckTable(1)

		// Extract values from builder
		workflowIDValue := L.GetField(builder, "_workflowID")
		typeValue := L.GetField(builder, "_type")
		nameValue := L.GetField(builder, "_name")
		descValue := L.GetField(builder, "_description")
		stepsValue := L.GetField(builder, "_steps")

		// Build config table
		config := L.NewTable()

		if typeValue != lua.LNil {
			L.SetField(config, "type", typeValue)
		}
		if nameValue != lua.LNil {
			L.SetField(config, "name", nameValue)
		}
		if descValue != lua.LNil {
			L.SetField(config, "description", descValue)
		}
		if stepsValue != lua.LNil {
			L.SetField(config, "steps", stepsValue)
		}

		// Call createWorkflow directly
		ctx := context.Background()
		createArgs := []engine.ScriptValue{
			engine.NewStringValue(workflowIDValue.String()),
		}

		// Convert config to ScriptValue
		configValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, config)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		createArgs = append(createArgs, configValue)

		result, err := wa.GetBridge().ExecuteMethod(ctx, "createWorkflow", createArgs)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Convert result back to Lua
		luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	L.Push(builder)
	return 1
}

func (wa *WorkflowAdapter) validateWorkflow(L *lua.LState) int {
	// Check if it's a config object or workflowID
	arg1 := L.Get(1)

	ctx := context.Background()
	var args []engine.ScriptValue

	if arg1.Type() == lua.LTTable {
		// It's a config object
		configValue, err := wa.GetTypeConverter().ToLuaScriptValue(L, arg1)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		args = []engine.ScriptValue{configValue}
	} else {
		// It's a workflowID
		workflowID := L.CheckString(1)
		args = []engine.ScriptValue{engine.NewStringValue(workflowID)}
	}

	result, err := wa.GetBridge().ExecuteMethod(ctx, "validateWorkflow", args)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result back to Lua
	luaResult, err := wa.GetTypeConverter().FromLuaScriptValue(L, result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(luaResult)
	return 1
}
