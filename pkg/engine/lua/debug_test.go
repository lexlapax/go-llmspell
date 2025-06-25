// ABOUTME: This file contains comprehensive tests for the Lua script debugger.
// ABOUTME: It tests breakpoints, step debugging, variable inspection, and watch expressions.

package gopherlua

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestDebugger_CreateAndDestroy(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)

	if debugger == nil {
		t.Fatal("Expected debugger to be created")
	}

	if err := debugger.Shutdown(); err != nil {
		t.Errorf("Expected clean shutdown, got error: %v", err)
	}
}

func TestDebugger_EngineAttachment(t *testing.T) {
	debugConfig := DefaultDebuggerConfig()
	debugger := NewDebugger(debugConfig)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Create a mock engine
	engine := NewLuaEngine()
	if engine == nil {
		t.Fatal("Failed to create engine")
	}

	// Test attachment without initialization (simplified test)
	if err := debugger.AttachToEngine(engine); err != nil {
		t.Errorf("Failed to attach to engine: %v", err)
	}

	if debugger.engine != engine {
		t.Error("Engine not properly attached")
	}

	// Test detachment
	if err := debugger.DetachFromEngine(); err != nil {
		t.Errorf("Failed to detach from engine: %v", err)
	}

	if debugger.engine != nil {
		t.Error("Engine not properly detached")
	}
}

func TestDebugger_Breakpoints(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	tests := []struct {
		name      string
		file      string
		line      int
		condition string
		expectErr bool
	}{
		{
			name: "simple breakpoint",
			file: "test.lua",
			line: 10,
		},
		{
			name:      "conditional breakpoint",
			file:      "test.lua",
			line:      20,
			condition: "x > 5",
		},
		{
			name:      "duplicate breakpoint",
			file:      "test.lua",
			line:      10,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, err := debugger.AddBreakpoint(tt.file, tt.line, tt.condition)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error for duplicate breakpoint")
				}
				return
			}

			if err != nil {
				t.Errorf("Failed to add breakpoint: %v", err)
				return
			}

			if bp.File != tt.file || bp.Line != tt.line {
				t.Errorf("Breakpoint location mismatch: got %s:%d, want %s:%d",
					bp.File, bp.Line, tt.file, tt.line)
			}

			if bp.Condition != tt.condition {
				t.Errorf("Breakpoint condition mismatch: got %q, want %q",
					bp.Condition, tt.condition)
			}

			if !bp.Enabled {
				t.Error("Breakpoint should be enabled by default")
			}
		})
	}

	// Test breakpoint removal
	bp, _ := debugger.AddBreakpoint("remove.lua", 5, "")
	if err := debugger.RemoveBreakpoint(bp.ID); err != nil {
		t.Errorf("Failed to remove breakpoint: %v", err)
	}

	// Verify removal
	state := debugger.GetState()
	if _, exists := state.Breakpoints[bp.ID]; exists {
		t.Error("Breakpoint was not removed")
	}
}

func TestDebugger_BreakpointManagement(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Add a breakpoint
	bp, err := debugger.AddBreakpoint("test.lua", 15, "")
	if err != nil {
		t.Fatalf("Failed to add breakpoint: %v", err)
	}

	// Test enabling/disabling
	if err := debugger.SetBreakpointEnabled(bp.ID, false); err != nil {
		t.Errorf("Failed to disable breakpoint: %v", err)
	}

	state := debugger.GetState()
	if state.Breakpoints[bp.ID].Enabled {
		t.Error("Breakpoint should be disabled")
	}

	if err := debugger.SetBreakpointEnabled(bp.ID, true); err != nil {
		t.Errorf("Failed to enable breakpoint: %v", err)
	}

	state = debugger.GetState()
	if !state.Breakpoints[bp.ID].Enabled {
		t.Error("Breakpoint should be enabled")
	}

	// Test non-existent breakpoint
	if err := debugger.SetBreakpointEnabled("nonexistent", true); err == nil {
		t.Error("Expected error for non-existent breakpoint")
	}
}

func TestDebugger_WatchExpressions(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	tests := []struct {
		name       string
		expression string
	}{
		{
			name:       "simple variable",
			expression: "x",
		},
		{
			name:       "arithmetic expression",
			expression: "x + y",
		},
		{
			name:       "function call",
			expression: "math.max(a, b)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watch, err := debugger.AddWatchExpression(tt.expression)
			if err != nil {
				t.Errorf("Failed to add watch expression: %v", err)
				return
			}

			if watch.Expression != tt.expression {
				t.Errorf("Watch expression mismatch: got %q, want %q",
					watch.Expression, tt.expression)
			}

			if watch.ID == "" {
				t.Error("Watch expression should have an ID")
			}
		})
	}

	// Test watch removal
	watch, _ := debugger.AddWatchExpression("test_var")
	if err := debugger.RemoveWatchExpression(watch.ID); err != nil {
		t.Errorf("Failed to remove watch expression: %v", err)
	}

	// Verify removal
	state := debugger.GetState()
	if _, exists := state.Watches[watch.ID]; exists {
		t.Error("Watch expression was not removed")
	}
}

func TestDebugger_StepModes(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Test step modes when not paused
	stepModes := []StepMode{
		StepModeNone,
		StepModeOver,
		StepModeInto,
		StepModeOut,
		StepModeLine,
	}

	for _, mode := range stepModes {
		if err := debugger.Step(mode); err == nil {
			t.Errorf("Expected error when stepping while not paused for mode %s", mode)
		}
	}

	// Test continue when not paused
	if err := debugger.Continue(); err == nil {
		t.Error("Expected error when continuing while not paused")
	}
}

func TestDebugger_StateManagement(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Test initial state
	state := debugger.GetState()
	if state.Running {
		t.Error("Debugger should not be running initially")
	}
	if state.Paused {
		t.Error("Debugger should not be paused initially")
	}
	if len(state.Breakpoints) != 0 {
		t.Error("Should have no breakpoints initially")
	}
	if len(state.Watches) != 0 {
		t.Error("Should have no watch expressions initially")
	}
}

func TestDebugger_EventHandling(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	var mu sync.Mutex
	var events []DebugEvent

	// Add event handler
	debugger.AddEventHandler(func(event DebugEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	})

	// Trigger some events
	_, _ = debugger.AddBreakpoint("test.lua", 10, "")
	_, _ = debugger.AddWatchExpression("x")

	// Give events time to be processed
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}

	// Check event types
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	if !eventTypes["breakpoint_added"] {
		t.Error("Expected breakpoint_added event")
	}
	if !eventTypes["watch_added"] {
		t.Error("Expected watch_added event")
	}
}

func TestDebugger_MaxLimits(t *testing.T) {
	config := DefaultDebuggerConfig()
	config.MaxBreakpoints = 2
	config.MaxWatchExpressions = 2

	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Test breakpoint limit
	for i := 0; i < config.MaxBreakpoints; i++ {
		_, err := debugger.AddBreakpoint("test.lua", i+1, "")
		if err != nil {
			t.Errorf("Failed to add breakpoint %d: %v", i, err)
		}
	}

	// Try to add one more
	_, err := debugger.AddBreakpoint("test.lua", 99, "")
	if err == nil {
		t.Error("Expected error when exceeding max breakpoints")
	}

	// Test watch expression limit
	for i := 0; i < config.MaxWatchExpressions; i++ {
		_, err := debugger.AddWatchExpression("var" + string(rune(i+'1')))
		if err != nil {
			t.Errorf("Failed to add watch expression %d: %v", i, err)
		}
	}

	// Try to add one more
	_, err = debugger.AddWatchExpression("overflow")
	if err == nil {
		t.Error("Expected error when exceeding max watch expressions")
	}
}

func TestDebugger_VariableInspection(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Create mock call stack
	debugger.state.CallStack = []DebugFrame{
		{
			Index:    0,
			Function: "main",
			File:     "test.lua",
			Line:     10,
			Locals: map[string]interface{}{
				"x": 42,
				"y": "hello",
				"z": true,
			},
			Upvalues: map[string]interface{}{
				"upval1": "test",
			},
		},
		{
			Index:    1,
			Function: "helper",
			File:     "test.lua",
			Line:     5,
			Locals: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
	}

	// Test getting variables for valid frame
	vars, err := debugger.GetVariables(0)
	if err != nil {
		t.Errorf("Failed to get variables: %v", err)
	}

	if len(vars) < 3 {
		t.Errorf("Expected at least 3 variables, got %d", len(vars))
	}

	if vars["x"] != 42 {
		t.Errorf("Expected x=42, got %v", vars["x"])
	}

	// Test invalid frame index
	_, err = debugger.GetVariables(99)
	if err == nil {
		t.Error("Expected error for invalid frame index")
	}
}

func TestDebugger_CallStack(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Test empty call stack
	stack := debugger.GetCallStack()
	if len(stack) != 0 {
		t.Errorf("Expected empty call stack, got %d frames", len(stack))
	}

	// Set up mock call stack
	debugger.state.CallStack = []DebugFrame{
		{
			Index:    0,
			Function: "main",
			File:     "main.lua",
			Line:     15,
		},
		{
			Index:    1,
			Function: "helper",
			File:     "helper.lua",
			Line:     8,
		},
	}

	stack = debugger.GetCallStack()
	if len(stack) != 2 {
		t.Errorf("Expected 2 frames in call stack, got %d", len(stack))
	}

	// Check frame order (should be top to bottom)
	if stack[0].Function != "main" {
		t.Errorf("Expected top frame to be 'main', got %s", stack[0].Function)
	}

	if stack[1].Function != "helper" {
		t.Errorf("Expected second frame to be 'helper', got %s", stack[1].Function)
	}
}

func TestDebugger_LuaValueConversion(t *testing.T) {
	config := DefaultDebuggerConfig()
	config.MaxStringLength = 10
	config.MaxVarDepth = 2

	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	// Create a simple Lua state for testing
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name     string
		setup    func() lua.LValue
		expected interface{}
	}{
		{
			name: "nil value",
			setup: func() lua.LValue {
				return lua.LNil
			},
			expected: nil,
		},
		{
			name: "boolean true",
			setup: func() lua.LValue {
				return lua.LTrue
			},
			expected: true,
		},
		{
			name: "boolean false",
			setup: func() lua.LValue {
				return lua.LFalse
			},
			expected: false,
		},
		{
			name: "number",
			setup: func() lua.LValue {
				return lua.LNumber(42.5)
			},
			expected: 42.5,
		},
		{
			name: "short string",
			setup: func() lua.LValue {
				return lua.LString("hello")
			},
			expected: "hello",
		},
		{
			name: "long string (truncated)",
			setup: func() lua.LValue {
				return lua.LString("this is a very long string that should be truncated")
			},
			expected: "this is a ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lval := tt.setup()
			result := debugger.luaValueToInterface(lval)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test function conversion
	_ = L.DoString("function test() end")
	fn := L.GetGlobal("test")
	result := debugger.luaValueToInterface(fn)
	if !strings.Contains(result.(string), "<function:") {
		t.Errorf("Expected function representation, got %v", result)
	}
}

func TestDebugger_ConcurrentAccess(t *testing.T) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent breakpoint operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Add breakpoint
			bp, err := debugger.AddBreakpoint("test.lua", i+1, "")
			if err != nil {
				errors <- err
				return
			}

			// Modify breakpoint
			if err := debugger.SetBreakpointEnabled(bp.ID, false); err != nil {
				errors <- err
				return
			}

			// Remove breakpoint
			if err := debugger.RemoveBreakpoint(bp.ID); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Concurrent watch operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Add watch
			watch, err := debugger.AddWatchExpression("var" + string(rune(i+'0')))
			if err != nil {
				errors <- err
				return
			}

			// Remove watch
			if err := debugger.RemoveWatchExpression(watch.ID); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	// Concurrent state reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			state := debugger.GetState()
			if state == nil {
				errors <- fmt.Errorf("got nil state")
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}
}

func BenchmarkDebugger_AddRemoveBreakpoint(b *testing.B) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			b.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp, err := debugger.AddBreakpoint("test.lua", i%1000+1, "")
		if err != nil {
			b.Fatal(err)
		}

		if err := debugger.RemoveBreakpoint(bp.ID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDebugger_WatchExpressions(b *testing.B) {
	config := DefaultDebuggerConfig()
	debugger := NewDebugger(config)
	defer func() {
		if err := debugger.Shutdown(); err != nil {
			b.Errorf("Failed to shutdown debugger: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watch, err := debugger.AddWatchExpression("test_var")
		if err != nil {
			b.Fatal(err)
		}

		if err := debugger.RemoveWatchExpression(watch.ID); err != nil {
			b.Fatal(err)
		}
	}
}
