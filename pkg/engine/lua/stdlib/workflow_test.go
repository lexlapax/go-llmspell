// ABOUTME: Tests for the workflow Lua module with mock bridge
// ABOUTME: Validates workflow orchestration functionality

package stdlib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// setupWorkflowBridge sets up mock workflow bridge
func setupWorkflowBridge(t *testing.T, L *lua.LState) {
	t.Helper()

	// Create bridges table
	bridges := L.NewTable()
	L.SetGlobal("bridges", bridges)

	// Create workflow bridge
	workflowBridge := L.NewTable()

	// Mock workflow storage
	workflows := make(map[string]*lua.LTable)
	templates := make(map[string]*lua.LTable)

	// Workflow lifecycle methods
	workflowBridge.RawSetString("createWorkflow", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		config := L.CheckTable(2)

		// Create workflow
		workflow := L.NewTable()
		workflow.RawSetString("id", lua.LString(id))
		workflow.RawSetString("status", lua.LString("created"))
		
		// Copy config fields
		config.ForEach(func(k, v lua.LValue) {
			workflow.RawSet(k, v)
		})

		workflows[id] = workflow
		L.Push(workflow)
		return 1
	}))

	workflowBridge.RawSetString("executeWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		input := L.Get(2) // optional

		if wf, ok := workflows[workflowID]; ok {
			wf.RawSetString("status", lua.LString("running"))
			if input != lua.LNil {
				wf.RawSetString("input", input)
			}
			
			result := L.NewTable()
			result.RawSetString("workflow_id", lua.LString(workflowID))
			result.RawSetString("status", lua.LString("completed"))
			result.RawSetString("output", lua.LString("execution result"))
			
			wf.RawSetString("status", lua.LString("completed"))
			L.Push(result)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("pauseWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			wf.RawSetString("status", lua.LString("paused"))
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("resumeWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			wf.RawSetString("status", lua.LString("running"))
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("stopWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			wf.RawSetString("status", lua.LString("cancelled"))
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("getWorkflowStatus", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			status := wf.RawGetString("status")
			L.Push(status)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("listWorkflows", L.NewFunction(func(L *lua.LState) int {
		list := L.NewTable()
		i := 1
		for _, wf := range workflows {
			list.RawSetInt(i, wf)
			i++
		}
		L.Push(list)
		return 1
	}))

	workflowBridge.RawSetString("getWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			L.Push(wf)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("deleteWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if _, ok := workflows[workflowID]; ok {
			delete(workflows, workflowID)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	// Step management methods
	workflowBridge.RawSetString("addStep", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		step := L.CheckTable(2)
		
		if wf, ok := workflows[workflowID]; ok {
			steps := wf.RawGetString("steps")
			if steps == lua.LNil {
				steps = L.NewTable()
				wf.RawSetString("steps", steps)
			}
			stepList := steps.(*lua.LTable)
			stepList.Append(step)
			
			// Return step with ID
			step.RawSetString("id", lua.LString(fmt.Sprintf("step_%d", stepList.Len())))
			L.Push(step)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("removeStep", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		_ = L.CheckString(2) // stepID
		
		if _, ok := workflows[workflowID]; ok {
			// Mock removal
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("updateStep", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // workflowID
		stepID := L.CheckString(2)
		updates := L.CheckTable(3)
		
		// Mock update
		result := L.NewTable()
		result.RawSetString("id", lua.LString(stepID))
		updates.ForEach(func(k, v lua.LValue) {
			result.RawSet(k, v)
		})
		L.Push(result)
		return 1
	}))

	workflowBridge.RawSetString("getStep", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		stepID := L.CheckString(2)
		
		// Mock step
		step := L.NewTable()
		step.RawSetString("id", lua.LString(stepID))
		step.RawSetString("workflow_id", lua.LString(workflowID))
		step.RawSetString("type", lua.LString("agent"))
		L.Push(step)
		return 1
	}))

	workflowBridge.RawSetString("listSteps", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			steps := wf.RawGetString("steps")
			if steps != lua.LNil {
				L.Push(steps)
			} else {
				L.Push(L.NewTable())
			}
		} else {
			L.Push(L.NewTable())
		}
		return 1
	}))

	workflowBridge.RawSetString("reorderSteps", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // workflowID
		_ = L.CheckTable(2)  // stepIDs
		
		// Mock success
		L.Push(lua.LTrue)
		return 1
	}))

	// Template methods
	workflowBridge.RawSetString("listTemplates", L.NewFunction(func(L *lua.LState) int {
		list := L.NewTable()
		i := 1
		for id, tmpl := range templates {
			item := L.NewTable()
			item.RawSetString("id", lua.LString(id))
			item.RawSetString("name", tmpl.RawGetString("name"))
			list.RawSetInt(i, item)
			i++
		}
		L.Push(list)
		return 1
	}))

	workflowBridge.RawSetString("listWorkflowTemplates", L.NewFunction(func(L *lua.LState) int {
		// Same as listTemplates for mock
		list := L.NewTable()
		i := 1
		for id, tmpl := range templates {
			item := L.NewTable()
			item.RawSetString("id", lua.LString(id))
			item.RawSetString("name", tmpl.RawGetString("name"))
			list.RawSetInt(i, item)
			i++
		}
		L.Push(list)
		return 1
	}))

	workflowBridge.RawSetString("getTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		
		if tmpl, ok := templates[templateID]; ok {
			L.Push(tmpl)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("getWorkflowTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		
		if tmpl, ok := templates[templateID]; ok {
			L.Push(tmpl)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("createWorkflowTemplate", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		templateName := L.CheckString(2)
		
		if wf, ok := workflows[workflowID]; ok {
			tmpl := L.NewTable()
			tmpl.RawSetString("id", lua.LString("tmpl_" + templateName))
			tmpl.RawSetString("source_workflow", lua.LString(workflowID))
			
			// Copy workflow config but preserve template name
			wf.ForEach(func(k, v lua.LValue) {
				if k.String() != "id" && k.String() != "status" && k.String() != "name" {
					tmpl.RawSet(k, v)
				}
			})
			
			// Set template name after copying to ensure it's not overwritten
			tmpl.RawSetString("name", lua.LString(templateName))
			
			templates["tmpl_"+templateName] = tmpl
			L.Push(tmpl)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("removeWorkflowTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		
		if _, ok := templates[templateID]; ok {
			delete(templates, templateID)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("createWorkflowFromTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		workflowID := L.CheckString(2)
		variables := L.Get(3) // optional
		
		if tmpl, ok := templates[templateID]; ok {
			wf := L.NewTable()
			wf.RawSetString("id", lua.LString(workflowID))
			wf.RawSetString("status", lua.LString("created"))
			wf.RawSetString("from_template", lua.LString(templateID))
			
			// Copy template config
			tmpl.ForEach(func(k, v lua.LValue) {
				if k.String() != "id" {
					wf.RawSet(k, v)
				}
			})
			
			// Apply variables if provided
			if variables != lua.LNil && variables.Type() == lua.LTTable {
				wf.RawSetString("variables", variables)
			}
			
			workflows[workflowID] = wf
			L.Push(wf)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("saveAsTemplate", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		templateConfig := L.CheckTable(2)
		
		if wf, ok := workflows[workflowID]; ok {
			templateName := templateConfig.RawGetString("name")
			if templateName == lua.LNil {
				templateName = lua.LString("template_" + workflowID)
			}
			
			tmpl := L.NewTable()
			tmpl.RawSetString("id", lua.LString("tmpl_" + templateName.String()))
			templateConfig.ForEach(func(k, v lua.LValue) {
				tmpl.RawSet(k, v)
			})
			
			// Copy workflow structure
			wf.ForEach(func(k, v lua.LValue) {
				if k.String() != "id" && k.String() != "status" {
					tmpl.RawSet(k, v)
				}
			})
			
			templates["tmpl_"+templateName.String()] = tmpl
			L.Push(tmpl)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	// Import/Export methods
	workflowBridge.RawSetString("exportWorkflow", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		format := L.CheckString(2)
		
		if wf, ok := workflows[workflowID]; ok {
			if format == "json" {
				L.Push(lua.LString(`{"id":"` + workflowID + `","status":"` + wf.RawGetString("status").String() + `"}`))
			} else if format == "yaml" {
				L.Push(lua.LString("id: " + workflowID + "\nstatus: " + wf.RawGetString("status").String()))
			} else {
				L.Push(lua.LNil)
			}
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("importWorkflow", L.NewFunction(func(L *lua.LState) int {
		data := L.CheckString(1)
		format := L.CheckString(2)
		
		// Mock import - create simple workflow
		wf := L.NewTable()
		wf.RawSetString("id", lua.LString("imported_workflow"))
		wf.RawSetString("status", lua.LString("created"))
		wf.RawSetString("imported_from", lua.LString(format))
		wf.RawSetString("data", lua.LString(data))
		
		workflows["imported_workflow"] = wf
		L.Push(wf)
		return 1
	}))

	// Variable management
	workflowBridge.RawSetString("setWorkflowVariable", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		name := L.CheckString(2)
		value := L.Get(3)
		
		if wf, ok := workflows[workflowID]; ok {
			vars := wf.RawGetString("variables")
			if vars == lua.LNil {
				vars = L.NewTable()
				wf.RawSetString("variables", vars)
			}
			varTable := vars.(*lua.LTable)
			varTable.RawSetString(name, value)
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	workflowBridge.RawSetString("getWorkflowVariable", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		name := L.CheckString(2)
		
		if wf, ok := workflows[workflowID]; ok {
			vars := wf.RawGetString("variables")
			if vars != lua.LNil {
				value := vars.(*lua.LTable).RawGetString(name)
				L.Push(value)
			} else {
				L.Push(lua.LNil)
			}
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	workflowBridge.RawSetString("listWorkflowVariables", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			vars := wf.RawGetString("variables")
			if vars != lua.LNil {
				L.Push(vars)
			} else {
				L.Push(L.NewTable())
			}
		} else {
			L.Push(L.NewTable())
		}
		return 1
	}))

	workflowBridge.RawSetString("removeWorkflowVariable", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		name := L.CheckString(2)
		
		if wf, ok := workflows[workflowID]; ok {
			vars := wf.RawGetString("variables")
			if vars != lua.LNil {
				vars.(*lua.LTable).RawSetString(name, lua.LNil)
				L.Push(lua.LTrue)
			} else {
				L.Push(lua.LFalse)
			}
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	// Error handling
	workflowBridge.RawSetString("getWorkflowErrors", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			errors := wf.RawGetString("errors")
			if errors != lua.LNil {
				L.Push(errors)
			} else {
				L.Push(L.NewTable())
			}
		} else {
			L.Push(L.NewTable())
		}
		return 1
	}))

	workflowBridge.RawSetString("clearWorkflowErrors", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		if wf, ok := workflows[workflowID]; ok {
			wf.RawSetString("errors", L.NewTable())
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))

	// Convenience methods
	workflowBridge.RawSetString("createBuilder", L.NewFunction(func(L *lua.LState) int {
		workflowID := L.CheckString(1)
		
		// Create builder (simplified)
		builder := L.NewTable()
		builder.RawSetString("_workflowID", lua.LString(workflowID))
		
		// Add builder methods
		builder.RawSetString("withType", L.NewFunction(func(L *lua.LState) int {
			builder := L.CheckTable(1)
			workflowType := L.CheckString(2)
			builder.RawSetString("_type", lua.LString(workflowType))
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("withName", L.NewFunction(func(L *lua.LState) int {
			builder := L.CheckTable(1)
			name := L.CheckString(2)
			builder.RawSetString("_name", lua.LString(name))
			L.Push(builder)
			return 1
		}))
		
		builder.RawSetString("build", L.NewFunction(func(L *lua.LState) int {
			builder := L.CheckTable(1)
			id := builder.RawGetString("_workflowID").String()
			
			config := L.NewTable()
			if t := builder.RawGetString("_type"); t != lua.LNil {
				config.RawSetString("type", t)
			}
			if n := builder.RawGetString("_name"); n != lua.LNil {
				config.RawSetString("name", n)
			}
			
			// Create workflow
			wf := L.NewTable()
			wf.RawSetString("id", lua.LString(id))
			wf.RawSetString("status", lua.LString("created"))
			config.ForEach(func(k, v lua.LValue) {
				wf.RawSet(k, v)
			})
			
			workflows[id] = wf
			L.Push(wf)
			return 1
		}))
		
		L.Push(builder)
		return 1
	}))

	workflowBridge.RawSetString("validateWorkflow", L.NewFunction(func(L *lua.LState) int {
		arg := L.Get(1)
		
		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		result.RawSetString("errors", L.NewTable())
		
		if arg.Type() == lua.LTString {
			// Validate by ID
			if _, ok := workflows[arg.String()]; !ok {
				result.RawSetString("valid", lua.LFalse)
				errors := L.NewTable()
				errors.Append(lua.LString("Workflow not found"))
				result.RawSetString("errors", errors)
			}
		}
		
		L.Push(result)
		return 1
	}))

	bridges.RawSetString("agent_workflow", workflowBridge)
}

func TestWorkflowModule(t *testing.T) {
	t.Run("module loads with bridge", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			assert(type(workflow) == "table", "workflow should be a table")
			
			-- Check constants exist
			assert(type(workflow.TYPES) == "table", "TYPES should be a table")
			assert(workflow.TYPES.SEQUENTIAL == "sequential", "TYPES.SEQUENTIAL should be 'sequential'")
			assert(workflow.TYPES.PARALLEL == "parallel", "TYPES.PARALLEL should be 'parallel'")
			
			assert(type(workflow.STATUS) == "table", "STATUS should be a table")
			assert(workflow.STATUS.CREATED == "created", "STATUS.CREATED should be 'created'")
			assert(workflow.STATUS.RUNNING == "running", "STATUS.RUNNING should be 'running'")
			
			assert(type(workflow.FORMATS) == "table", "FORMATS should be a table")
			assert(workflow.FORMATS.JSON == "json", "FORMATS.JSON should be 'json'")
			
			assert(type(workflow.STEP_TYPES) == "table", "STEP_TYPES should be a table")
			assert(workflow.STEP_TYPES.AGENT == "agent", "STEP_TYPES.AGENT should be 'agent'")
			
			-- Check methods exist
			assert(type(workflow.create_workflow) == "function", "create_workflow should be a function")
			assert(type(workflow.execute_workflow) == "function", "execute_workflow should be a function")
			assert(type(workflow.pause_workflow) == "function", "pause_workflow should be a function")
			assert(type(workflow.list_templates) == "function", "list_templates should be a function")
			assert(type(workflow.create_builder) == "function", "create_builder should be a function")
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow lifecycle", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("test_workflow", {
				name = "Test Workflow",
				type = workflow.TYPES.SEQUENTIAL,
				description = "A test workflow"
			})
			assert(wf.id == "test_workflow", "workflow id should match")
			assert(wf.status == "created", "workflow should be created")
			assert(wf.name == "Test Workflow", "workflow name should match")
			
			-- Get workflow status
			local status = workflow.get_workflow_status("test_workflow")
			assert(status == "created", "status should be created")
			
			-- Execute workflow
			local result = workflow.execute_workflow("test_workflow", {input = "test data"})
			assert(result.status == "completed", "execution should complete")
			assert(result.output == "execution result", "should have output")
			
			-- Get updated status
			status = workflow.get_workflow_status("test_workflow")
			assert(status == "completed", "status should be completed")
			
			-- List workflows
			local workflows = workflow.list_workflows()
			assert(#workflows >= 1, "should have at least one workflow")
			
			-- Delete workflow
			local deleted = workflow.delete_workflow("test_workflow")
			assert(deleted == true, "should delete workflow")
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow control operations", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create and execute workflow
			local wf = workflow.create_workflow("control_test", {
				name = "Control Test",
				type = workflow.TYPES.SEQUENTIAL
			})
			
			workflow.execute_workflow("control_test")
			
			-- Pause workflow
			local paused = workflow.pause_workflow("control_test")
			assert(paused == true, "should pause workflow")
			
			local status = workflow.get_workflow_status("control_test")
			assert(status == "paused", "status should be paused")
			
			-- Resume workflow
			local resumed = workflow.resume_workflow("control_test")
			assert(resumed == true, "should resume workflow")
			
			status = workflow.get_workflow_status("control_test")
			assert(status == "running", "status should be running")
			
			-- Stop workflow
			local stopped = workflow.stop_workflow("control_test")
			assert(stopped == true, "should stop workflow")
			
			status = workflow.get_workflow_status("control_test")
			assert(status == "cancelled", "status should be cancelled")
		`)
		assert.NoError(t, err)
	})

	t.Run("step management", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("step_test", {
				name = "Step Test"
			})
			
			-- Add steps
			local step1 = workflow.add_step("step_test", {
				type = workflow.STEP_TYPES.AGENT,
				name = "Agent Step",
				config = {agent_id = "agent1"}
			})
			assert(step1.id == "step_1", "step should have id")
			
			local step2 = workflow.add_step("step_test", {
				type = workflow.STEP_TYPES.SCRIPT,
				name = "Script Step",
				script = "return true"
			})
			
			-- List steps
			local steps = workflow.list_steps("step_test")
			assert(#steps == 2, "should have 2 steps")
			
			-- Get specific step
			local step = workflow.get_step("step_test", "step_1")
			assert(step.id == "step_1", "step id should match")
			assert(step.type == "agent", "step type should be agent")
			
			-- Update step
			local updated = workflow.update_step("step_test", "step_1", {
				name = "Updated Agent Step",
				config = {agent_id = "agent2"}
			})
			assert(updated.name == "Updated Agent Step", "step should be updated")
			
			-- Remove step
			local removed = workflow.remove_step("step_test", "step_1")
			assert(removed == true, "step should be removed")
			
			-- Reorder steps
			local reordered = workflow.reorder_steps("step_test", {"step_2", "step_1"})
			assert(reordered == true, "steps should be reordered")
		`)
		assert.NoError(t, err)
	})

	t.Run("template management", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create a workflow to use as template
			local wf = workflow.create_workflow("template_source", {
				name = "Template Source",
				type = workflow.TYPES.SEQUENTIAL,
				steps = {}
			})
			
			-- Create template from workflow
			local template = workflow.create_workflow_template("template_source", "my_template")
			assert(template.id == "tmpl_my_template", "template should have id")
			assert(template.name == "my_template", "template name should match")
			
			-- List templates
			local templates = workflow.list_templates()
			assert(#templates >= 1, "should have at least one template")
			
			-- Get template
			local tmpl = workflow.get_template("tmpl_my_template")
			assert(tmpl.name == "my_template", "template name should match")
			
			-- Create workflow from template
			local new_wf = workflow.create_workflow_from_template("tmpl_my_template", "new_workflow", {
				var1 = "value1",
				var2 = "value2"
			})
			assert(new_wf.id == "new_workflow", "new workflow id should match")
			assert(new_wf.from_template == "tmpl_my_template", "should track template source")
			
			-- Save existing workflow as template
			local saved = workflow.save_as_template("new_workflow", {
				name = "saved_template",
				description = "Saved from workflow"
			})
			assert(saved.id == "tmpl_saved_template", "saved template should have id")
			
			-- Remove template
			local removed = workflow.remove_workflow_template("tmpl_my_template")
			assert(removed == true, "template should be removed")
		`)
		assert.NoError(t, err)
	})

	t.Run("import export", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("export_test", {
				name = "Export Test",
				type = workflow.TYPES.SEQUENTIAL
			})
			
			-- Export as JSON
			local json_export = workflow.export_workflow("export_test", workflow.FORMATS.JSON)
			assert(type(json_export) == "string", "should export as JSON string")
			assert(json_export:find("export_test") ~= nil, "JSON should contain workflow id")
			
			-- Export as YAML
			local yaml_export = workflow.export_workflow("export_test", workflow.FORMATS.YAML)
			assert(type(yaml_export) == "string", "should export as YAML string")
			assert(yaml_export:find("export_test") ~= nil, "YAML should contain workflow id")
			
			-- Import workflow
			local imported = workflow.import_workflow('{"id": "imported", "name": "Imported"}', workflow.FORMATS.JSON)
			assert(imported.id == "imported_workflow", "should import workflow")
			assert(imported.imported_from == "json", "should track import format")
		`)
		assert.NoError(t, err)
	})

	t.Run("variable management", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("var_test", {
				name = "Variable Test"
			})
			
			-- Set variables
			local set1 = workflow.set_workflow_variable("var_test", "api_key", "secret123")
			assert(set1 == true, "should set variable")
			
			local set2 = workflow.set_workflow_variable("var_test", "timeout", 30)
			assert(set2 == true, "should set numeric variable")
			
			local set3 = workflow.set_workflow_variable("var_test", "config", {debug = true})
			assert(set3 == true, "should set table variable")
			
			-- Get variable
			local api_key = workflow.get_workflow_variable("var_test", "api_key")
			assert(api_key == "secret123", "should get variable value")
			
			-- List variables
			local vars = workflow.list_workflow_variables("var_test")
			assert(vars.api_key == "secret123", "should have api_key")
			assert(vars.timeout == 30, "should have timeout")
			assert(type(vars.config) == "table", "should have config table")
			
			-- Remove variable
			local removed = workflow.remove_workflow_variable("var_test", "api_key")
			assert(removed == true, "should remove variable")
			
			-- Verify removal
			local removed_var = workflow.get_workflow_variable("var_test", "api_key")
			assert(removed_var == nil, "variable should be removed")
		`)
		assert.NoError(t, err)
	})

	t.Run("error handling", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("error_test", {
				name = "Error Test"
			})
			
			-- Get errors (initially empty)
			local errors = workflow.get_workflow_errors("error_test")
			assert(type(errors) == "table", "should return errors table")
			assert(#errors == 0, "should have no errors initially")
			
			-- Clear errors
			local cleared = workflow.clear_workflow_errors("error_test")
			assert(cleared == true, "should clear errors")
		`)
		assert.NoError(t, err)
	})

	t.Run("builder pattern", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Use bridge builder
			local builder = workflow.create_builder("builder_test")
			local wf = builder
				:withType(workflow.TYPES.SEQUENTIAL)
				:withName("Built Workflow")
				:build()
			
			assert(wf.id == "builder_test", "workflow id should match")
			assert(wf.type == workflow.TYPES.SEQUENTIAL, "type should match")
			assert(wf.name == "Built Workflow", "name should match")
			
			-- Use Lua builder helper
			local builder2 = workflow.new_builder("lua_builder_test")
			local wf2 = builder2
				:with_type(workflow.TYPES.PARALLEL)
				:with_name("Lua Built Workflow")
				:with_description("Built using Lua builder")
				:add_step({type = workflow.STEP_TYPES.AGENT, name = "Step 1"})
				:add_step({type = workflow.STEP_TYPES.SCRIPT, name = "Step 2"})
				:build()
			
			assert(wf2.id == "lua_builder_test", "workflow id should match")
			assert(wf2.type == workflow.TYPES.PARALLEL, "type should match")
			assert(wf2.description == "Built using Lua builder", "description should match")
			assert(#wf2.steps == 2, "should have 2 steps")
		`)
		assert.NoError(t, err)
	})

	t.Run("workflow validation", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- Create workflow
			local wf = workflow.create_workflow("validate_test", {
				name = "Validation Test"
			})
			
			-- Validate existing workflow
			local result = workflow.validate_workflow("validate_test")
			assert(result.valid == true, "workflow should be valid")
			assert(#result.errors == 0, "should have no errors")
			
			-- Validate non-existent workflow
			result = workflow.validate_workflow("non_existent")
			assert(result.valid == false, "non-existent workflow should be invalid")
			assert(#result.errors > 0, "should have errors")
			
			-- Validate config object
			result = workflow.validate_workflow({
				type = workflow.TYPES.SEQUENTIAL,
				name = "Test"
			})
			assert(result.valid == true, "config should be valid")
		`)
		assert.NoError(t, err)
	})

	t.Run("missing bridge graceful failure", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		// Don't setup bridge
		L.SetGlobal("bridges", L.NewTable())
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			local success, err = pcall(function()
				workflow.create_workflow("test", {})
			end)
			assert(not success, "should fail without bridge")
			assert(err:find("Workflow bridge not available"), "should have correct error message")
		`)
		assert.NoError(t, err)
	})
}

func TestWorkflowIntegration(t *testing.T) {
	t.Run("complete workflow scenario", func(t *testing.T) {
		L := lua.NewState()
		defer L.Close()

		setupWorkflowBridge(t, L)
		LoadModule(t, L, "workflow")

		err := L.DoString(`
			local workflow = require("workflow")
			
			-- 1. Create a workflow using builder
			local wf = workflow.new_builder("integration_test")
				:with_type(workflow.TYPES.SEQUENTIAL)
				:with_name("Integration Test Workflow")
				:with_description("Complete integration test")
				:add_step({
					type = workflow.STEP_TYPES.AGENT,
					name = "Data Collection",
					config = {agent_id = "collector"}
				})
				:add_step({
					type = workflow.STEP_TYPES.SCRIPT,
					name = "Data Processing",
					script = "process_data()"
				})
				:add_step({
					type = workflow.STEP_TYPES.AGENT,
					name = "Report Generation",
					config = {agent_id = "reporter"}
				})
				:build()
			
			-- 2. Set workflow variables
			workflow.set_workflow_variable("integration_test", "input_source", "database")
			workflow.set_workflow_variable("integration_test", "output_format", "pdf")
			workflow.set_workflow_variable("integration_test", "processing_options", {
				validate = true,
				clean = true,
				aggregate = false
			})
			
			-- 3. Validate workflow
			local validation = workflow.validate_workflow("integration_test")
			assert(validation.valid == true, "workflow should be valid")
			
			-- 4. Save as template
			local template = workflow.save_as_template("integration_test", {
				name = "data_pipeline_template",
				description = "Standard data processing pipeline",
				category = "data_processing"
			})
			
			-- 5. Create another instance from template
			local wf2 = workflow.create_workflow_from_template(
				template.id,
				"pipeline_instance_1",
				{input_source = "api", output_format = "json"}
			)
			
			-- 6. Execute the new workflow
			local execution = workflow.execute_workflow("pipeline_instance_1", {
				data = {1, 2, 3, 4, 5},
				timestamp = os.time()
			})
			assert(execution.status == "completed", "execution should complete")
			
			-- 7. Export the workflow
			local exported = workflow.export_workflow("pipeline_instance_1", workflow.FORMATS.JSON)
			assert(type(exported) == "string", "should export workflow")
			
			-- 8. Clean up
			workflow.delete_workflow("integration_test")
			workflow.delete_workflow("pipeline_instance_1")
			workflow.remove_workflow_template(template.id)
		`)
		assert.NoError(t, err)
	})
}