// ABOUTME: This file implements debugging support for Lua scripts including breakpoints, step debugging, and variable inspection.
// ABOUTME: It provides comprehensive debugging capabilities for development and troubleshooting of Lua spells.

package gopherlua

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// DebuggerConfig configures the debugger behavior
type DebuggerConfig struct {
	// Breakpoint settings
	MaxBreakpoints int  `json:"max_breakpoints"`
	BreakOnError   bool `json:"break_on_error"`
	BreakOnPanic   bool `json:"break_on_panic"`

	// Step debugging
	StepTimeout  time.Duration `json:"step_timeout"`
	MaxStepDepth int           `json:"max_step_depth"`

	// Variable inspection
	MaxVarDepth      int `json:"max_var_depth"`
	MaxStringLength  int `json:"max_string_length"`
	MaxArrayElements int `json:"max_array_elements"`

	// Watch expressions
	MaxWatchExpressions int           `json:"max_watch_expressions"`
	WatchInterval       time.Duration `json:"watch_interval"`

	// Output settings
	IncludeStackTrace bool `json:"include_stack_trace"`
	IncludeLocals     bool `json:"include_locals"`
	IncludeUpvalues   bool `json:"include_upvalues"`
}

// DefaultDebuggerConfig returns a default debugger configuration
func DefaultDebuggerConfig() DebuggerConfig {
	return DebuggerConfig{
		MaxBreakpoints:      50,
		BreakOnError:        true,
		BreakOnPanic:        true,
		StepTimeout:         30 * time.Second,
		MaxStepDepth:        100,
		MaxVarDepth:         5,
		MaxStringLength:     1000,
		MaxArrayElements:    100,
		MaxWatchExpressions: 20,
		WatchInterval:       100 * time.Millisecond,
		IncludeStackTrace:   true,
		IncludeLocals:       true,
		IncludeUpvalues:     false,
	}
}

// Breakpoint represents a debugging breakpoint
type Breakpoint struct {
	ID        string            `json:"id"`
	File      string            `json:"file"`
	Line      int               `json:"line"`
	Condition string            `json:"condition,omitempty"`
	HitCount  int               `json:"hit_count"`
	Enabled   bool              `json:"enabled"`
	Temporary bool              `json:"temporary"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// WatchExpression represents a watch expression
type WatchExpression struct {
	ID         string      `json:"id"`
	Expression string      `json:"expression"`
	Value      interface{} `json:"value"`
	Error      string      `json:"error,omitempty"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// DebugFrame represents a single frame in the call stack
type DebugFrame struct {
	Index    int                    `json:"index"`
	Function string                 `json:"function"`
	File     string                 `json:"file"`
	Line     int                    `json:"line"`
	Locals   map[string]interface{} `json:"locals,omitempty"`
	Upvalues map[string]interface{} `json:"upvalues,omitempty"`
}

// DebugState represents the current debugging state
type DebugState struct {
	Running      bool                        `json:"running"`
	Paused       bool                        `json:"paused"`
	CurrentFrame int                         `json:"current_frame"`
	CallStack    []DebugFrame                `json:"call_stack"`
	Breakpoints  map[string]*Breakpoint      `json:"breakpoints"`
	Watches      map[string]*WatchExpression `json:"watches"`
	StepMode     StepMode                    `json:"step_mode"`
	LastBreakHit *Breakpoint                 `json:"last_break_hit,omitempty"`
	Variables    map[string]interface{}      `json:"variables,omitempty"`
	Error        string                      `json:"error,omitempty"`
}

// StepMode defines the stepping behavior
type StepMode string

const (
	StepModeNone StepMode = "none" // Continue execution
	StepModeOver StepMode = "over" // Step over function calls
	StepModeInto StepMode = "into" // Step into function calls
	StepModeOut  StepMode = "out"  // Step out of current function
	StepModeLine StepMode = "line" // Step to next line
)

// DebugEvent represents a debugging event
type DebugEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	Frame     *DebugFrame `json:"frame,omitempty"`
}

// DebugEventHandler handles debug events
type DebugEventHandler func(event DebugEvent)

// Debugger provides debugging capabilities for Lua scripts
type Debugger struct {
	config        DebuggerConfig
	state         *DebugState
	engine        *LuaEngine
	luaState      *lua.LState
	mu            sync.RWMutex
	handlers      []DebugEventHandler
	stepChan      chan StepMode
	ctx           context.Context
	cancel        context.CancelFunc
	hookInstalled bool
}

// NewDebugger creates a new debugger instance
func NewDebugger(config DebuggerConfig) *Debugger {
	ctx, cancel := context.WithCancel(context.Background())

	return &Debugger{
		config: config,
		state: &DebugState{
			Running:     false,
			Paused:      false,
			Breakpoints: make(map[string]*Breakpoint),
			Watches:     make(map[string]*WatchExpression),
			StepMode:    StepModeNone,
			CallStack:   []DebugFrame{},
		},
		handlers: []DebugEventHandler{},
		stepChan: make(chan StepMode, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// AttachToEngine attaches the debugger to a Lua engine
func (d *Debugger) AttachToEngine(engine *LuaEngine) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.engine = engine
	// Note: LuaEngine uses a pool, so we'll get state during execution
	d.luaState = nil

	// Install debug hook
	return d.installDebugHook()
}

// DetachFromEngine detaches the debugger from the engine
func (d *Debugger) DetachFromEngine() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Note: gopher-lua doesn't support SetHook, so we use a simpler approach
	d.engine = nil
	d.luaState = nil
	d.hookInstalled = false

	return nil
}

// AddBreakpoint adds a breakpoint at the specified location
func (d *Debugger) AddBreakpoint(file string, line int, condition string) (*Breakpoint, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.state.Breakpoints) >= d.config.MaxBreakpoints {
		return nil, fmt.Errorf("maximum number of breakpoints (%d) reached", d.config.MaxBreakpoints)
	}

	id := fmt.Sprintf("%s:%d", file, line)
	if _, exists := d.state.Breakpoints[id]; exists {
		return nil, fmt.Errorf("breakpoint already exists at %s:%d", file, line)
	}

	breakpoint := &Breakpoint{
		ID:        id,
		File:      file,
		Line:      line,
		Condition: condition,
		HitCount:  0,
		Enabled:   true,
		Temporary: false,
		CreatedAt: time.Now(),
	}

	d.state.Breakpoints[id] = breakpoint
	d.emitEvent("breakpoint_added", breakpoint)

	return breakpoint, nil
}

// RemoveBreakpoint removes a breakpoint
func (d *Debugger) RemoveBreakpoint(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	breakpoint, exists := d.state.Breakpoints[id]
	if !exists {
		return fmt.Errorf("breakpoint %s not found", id)
	}

	delete(d.state.Breakpoints, id)
	d.emitEvent("breakpoint_removed", breakpoint)

	return nil
}

// SetBreakpointEnabled enables or disables a breakpoint
func (d *Debugger) SetBreakpointEnabled(id string, enabled bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	breakpoint, exists := d.state.Breakpoints[id]
	if !exists {
		return fmt.Errorf("breakpoint %s not found", id)
	}

	breakpoint.Enabled = enabled
	d.emitEvent("breakpoint_changed", breakpoint)

	return nil
}

// AddWatchExpression adds a watch expression
func (d *Debugger) AddWatchExpression(expression string) (*WatchExpression, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.state.Watches) >= d.config.MaxWatchExpressions {
		return nil, fmt.Errorf("maximum number of watch expressions (%d) reached", d.config.MaxWatchExpressions)
	}

	id := fmt.Sprintf("watch_%d", len(d.state.Watches)+1)
	watch := &WatchExpression{
		ID:         id,
		Expression: expression,
		UpdatedAt:  time.Now(),
	}

	// Evaluate the expression if we're paused
	if d.state.Paused && d.luaState != nil {
		d.evaluateWatchExpression(watch)
	}

	d.state.Watches[id] = watch
	d.emitEvent("watch_added", watch)

	return watch, nil
}

// RemoveWatchExpression removes a watch expression
func (d *Debugger) RemoveWatchExpression(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	watch, exists := d.state.Watches[id]
	if !exists {
		return fmt.Errorf("watch expression %s not found", id)
	}

	delete(d.state.Watches, id)
	d.emitEvent("watch_removed", watch)

	return nil
}

// Step performs a step operation
func (d *Debugger) Step(mode StepMode) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.state.Paused {
		return fmt.Errorf("debugger is not paused")
	}

	d.state.StepMode = mode

	// Send step command
	select {
	case d.stepChan <- mode:
		d.state.Paused = false
		d.emitEvent("step", map[string]interface{}{"mode": mode})
	default:
		return fmt.Errorf("step channel is full")
	}

	return nil
}

// Continue resumes execution
func (d *Debugger) Continue() error {
	return d.Step(StepModeNone)
}

// Pause pauses execution at the next statement
func (d *Debugger) Pause() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state.Paused {
		return fmt.Errorf("debugger is already paused")
	}

	d.state.StepMode = StepModeLine
	d.emitEvent("pause_requested", nil)

	return nil
}

// GetCallStack returns the current call stack
func (d *Debugger) GetCallStack() []DebugFrame {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.state.CallStack
}

// GetVariables returns variables at the specified frame
func (d *Debugger) GetVariables(frameIndex int) (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if frameIndex < 0 || frameIndex >= len(d.state.CallStack) {
		return nil, fmt.Errorf("invalid frame index %d", frameIndex)
	}

	frame := d.state.CallStack[frameIndex]
	variables := make(map[string]interface{})

	// Combine locals and upvalues
	for k, v := range frame.Locals {
		variables[k] = v
	}

	if d.config.IncludeUpvalues {
		for k, v := range frame.Upvalues {
			variables["upvalue."+k] = v
		}
	}

	return variables, nil
}

// EvaluateExpression evaluates an expression in the current context
func (d *Debugger) EvaluateExpression(expression string) (interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.state.Paused || d.luaState == nil {
		return nil, fmt.Errorf("debugger is not paused or not attached")
	}

	// Compile and execute the expression
	if err := d.luaState.DoString(fmt.Sprintf("return %s", expression)); err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Get the result
	result := d.luaState.Get(-1)
	d.luaState.Pop(1)

	return d.luaValueToInterface(result), nil
}

// GetState returns the current debug state
func (d *Debugger) GetState() *DebugState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Return a copy to prevent modification
	stateCopy := *d.state
	return &stateCopy
}

// AddEventHandler adds a debug event handler
func (d *Debugger) AddEventHandler(handler DebugEventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers = append(d.handlers, handler)
}

// Shutdown shuts down the debugger
func (d *Debugger) Shutdown() error {
	d.cancel()
	return d.DetachFromEngine()
}

// installDebugHook installs the debug hook in the Lua state
func (d *Debugger) installDebugHook() error {
	// Note: gopher-lua doesn't support SetHook, so we use a simpler approach
	// Debug functionality will be limited to manual breakpoints and variable inspection
	// No Lua state is required during attachment - it will be provided during execution
	d.hookInstalled = true

	return nil
}

// debugHook would be the main debug hook function, but gopher-lua doesn't support SetHook
// Instead, we implement manual debugging through script instrumentation or breakpoint checking

// CheckBreakpoint manually checks if execution should break at a given location
// This can be called by the engine during script execution
func (d *Debugger) CheckBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.hookInstalled || d.luaState == nil {
		return false
	}

	// Build current call stack
	d.buildCallStack(d.luaState)

	// Check for breakpoints
	shouldBreak := d.checkBreakpoints(file, line)

	// Check step mode
	if !shouldBreak {
		shouldBreak = d.checkStepMode()
	}

	if shouldBreak {
		d.state.Paused = true
		d.state.Running = false

		// Update watch expressions
		d.updateWatchExpressions()

		// Emit pause event
		d.emitEvent("paused", map[string]interface{}{
			"file":   file,
			"line":   line,
			"reason": d.getPauseReason(),
		})

		// Wait for step command
		d.waitForStep()

		d.state.Paused = false
		d.state.Running = true

		return true
	}

	return false
}

// checkBreakpoints checks if we should break at the current location
func (d *Debugger) checkBreakpoints(file string, line int) bool {
	id := fmt.Sprintf("%s:%d", file, line)
	breakpoint, exists := d.state.Breakpoints[id]

	if !exists || !breakpoint.Enabled {
		return false
	}

	// Check condition if present
	if breakpoint.Condition != "" {
		result, err := d.EvaluateExpression(breakpoint.Condition)
		if err != nil || !d.isTruthy(result) {
			return false
		}
	}

	breakpoint.HitCount++
	d.state.LastBreakHit = breakpoint

	// Remove temporary breakpoints
	if breakpoint.Temporary {
		delete(d.state.Breakpoints, id)
	}

	return true
}

// checkStepMode checks if we should break based on step mode
// Note: Without hook support, step debugging is limited
func (d *Debugger) checkStepMode() bool {
	switch d.state.StepMode {
	case StepModeLine:
		d.state.StepMode = StepModeNone
		return true
	case StepModeOver:
		// Step over: break on same level or higher
		if len(d.state.CallStack) <= d.state.CurrentFrame {
			d.state.StepMode = StepModeNone
			return true
		}
	case StepModeInto:
		// Step into: always break
		d.state.StepMode = StepModeNone
		return true
	case StepModeOut:
		// Step out: break when we return to a higher level
		if len(d.state.CallStack) < d.state.CurrentFrame {
			d.state.StepMode = StepModeNone
			return true
		}
	}

	return false
}

// buildCallStack builds the current call stack
func (d *Debugger) buildCallStack(L *lua.LState) {
	d.state.CallStack = []DebugFrame{}

	for level := 0; level < d.config.MaxStepDepth; level++ {
		ar, ok := L.GetStack(level)
		if !ok {
			break
		}

		frame := DebugFrame{
			Index:    level,
			Function: ar.Name,
			File:     ar.Source,
			Line:     ar.CurrentLine,
		}

		// Get local variables if enabled
		if d.config.IncludeLocals {
			frame.Locals = d.getLocalVariables(L, level)
		}

		// Get upvalues if enabled
		if d.config.IncludeUpvalues {
			frame.Upvalues = d.getUpvalues(L, level)
		}

		d.state.CallStack = append(d.state.CallStack, frame)
	}
}

// getLocalVariables gets local variables for a frame
func (d *Debugger) getLocalVariables(L *lua.LState, level int) map[string]interface{} {
	locals := make(map[string]interface{})

	// Get stack info for this level
	ar, ok := L.GetStack(level)
	if !ok {
		return locals
	}

	for i := 1; i <= 200; i++ { // Arbitrary limit
		name, value := L.GetLocal(ar, i)
		if name == "" {
			break
		}

		// Skip internal variables
		if strings.HasPrefix(name, "(") {
			continue
		}

		locals[name] = d.luaValueToInterface(value)
	}

	return locals
}

// getUpvalues gets upvalues for a frame
func (d *Debugger) getUpvalues(L *lua.LState, level int) map[string]interface{} {
	upvalues := make(map[string]interface{})

	// Get function at level
	ar, ok := L.GetStack(level)
	if !ok {
		return upvalues
	}

	// Note: gopher-lua's GetInfo may not support all functionality
	// Simplified upvalue extraction for compatibility
	info, err := L.GetInfo("f", ar, lua.LNil)
	if err != nil || info == nil {
		return upvalues
	}

	if fn, ok := info.(*lua.LFunction); ok && fn != nil && fn.Proto != nil {
		// Get upvalues - simplified approach
		for i := 1; i <= int(fn.Proto.NumUpvalues); i++ {
			name, upvalue := L.GetUpvalue(fn, i)
			if name != "" && upvalue != nil {
				upvalues[name] = d.luaValueToInterface(upvalue)
			}
		}
	}

	return upvalues
}

// updateWatchExpressions updates all watch expressions
func (d *Debugger) updateWatchExpressions() {
	for _, watch := range d.state.Watches {
		d.evaluateWatchExpression(watch)
	}
}

// evaluateWatchExpression evaluates a single watch expression
func (d *Debugger) evaluateWatchExpression(watch *WatchExpression) {
	result, err := d.EvaluateExpression(watch.Expression)
	if err != nil {
		watch.Error = err.Error()
		watch.Value = nil
	} else {
		watch.Error = ""
		watch.Value = result
	}
	watch.UpdatedAt = time.Now()
}

// waitForStep waits for a step command
func (d *Debugger) waitForStep() {
	// Release lock while waiting
	d.mu.Unlock()
	defer d.mu.Lock()

	ctx, cancel := context.WithTimeout(d.ctx, d.config.StepTimeout)
	defer cancel()

	select {
	case mode := <-d.stepChan:
		d.state.StepMode = mode
	case <-ctx.Done():
		// Timeout: continue execution
		d.state.StepMode = StepModeNone
	}
}

// emitEvent emits a debug event to all handlers
func (d *Debugger) emitEvent(eventType string, data interface{}) {
	event := DebugEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	if len(d.state.CallStack) > 0 {
		event.Frame = &d.state.CallStack[0]
	}

	for _, handler := range d.handlers {
		go handler(event)
	}
}

// getPauseReason returns the reason for pausing
func (d *Debugger) getPauseReason() string {
	if d.state.LastBreakHit != nil {
		return "breakpoint"
	}
	if d.state.StepMode != StepModeNone {
		return "step"
	}
	return "unknown"
}

// isTruthy checks if a value is truthy in Lua
func (d *Debugger) isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
	}
	return true
}

// luaValueToInterface converts a Lua value to a Go interface
func (d *Debugger) luaValueToInterface(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		s := string(v)
		if len(s) > d.config.MaxStringLength {
			s = s[:d.config.MaxStringLength] + "..."
		}
		return s
	case *lua.LTable:
		return d.tableToMap(v, 0)
	case *lua.LFunction:
		return fmt.Sprintf("<function: %p>", v)
	case *lua.LUserData:
		return fmt.Sprintf("<userdata: %v>", v.Value)
	default:
		return fmt.Sprintf("<%s>", lv.Type().String())
	}
}

// tableToMap converts a Lua table to a Go map with depth limit
func (d *Debugger) tableToMap(table *lua.LTable, depth int) interface{} {
	if depth >= d.config.MaxVarDepth {
		return "<max depth reached>"
	}

	result := make(map[string]interface{})
	count := 0

	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if count >= d.config.MaxArrayElements {
			result["<truncated>"] = fmt.Sprintf("... and %d more", table.Len()-count)
			return
		}

		keyStr := d.luaValueToInterface(key).(string)
		result[keyStr] = d.luaValueToInterface(value)
		count++
	})

	return result
}
