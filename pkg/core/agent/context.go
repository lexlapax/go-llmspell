// ABOUTME: Agent execution context that provides resource limits, cancellation, timeout, and distributed tracing
// ABOUTME: Supports multi-engine execution with isolated contexts and comprehensive resource management

package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AgentContext provides execution context for agents with resource limits,
// cancellation support, distributed tracing, and multi-engine execution
type AgentContext interface {
	// Core context methods
	Context() context.Context
	Cancel()
	IsCancelled() bool
	Done() <-chan struct{}
	Deadline() (deadline time.Time, ok bool)
	Value(key interface{}) interface{}
	WithValue(key, value interface{}) AgentContext

	// Resource limits
	Timeout() time.Duration
	MaxMemory() int64
	MaxCPU() int
	CheckMemoryLimit(required int64) bool
	CheckCPULimit(required int) bool
	RecordMemoryUsage(bytes int64)
	RecordCPUUsage(cores int)
	ReleaseMemory(bytes int64)
	ReleaseCPU(cores int)
	CurrentMemory() int64
	CurrentCPU() int

	// Metadata
	Metadata() map[string]interface{}
	GetMetadata(key string) interface{}
	SetMetadata(key string, value interface{})

	// Tracing
	TraceID() string
	SpanID() string
	StartSpan(name string, attributes map[string]interface{}) Span
	SetCurrentSpan(span Span)

	// Multi-engine support
	ForEngine(engine string) AgentContext

	// Hooks
	AddBeforeExecuteHook(hook BeforeExecuteHook)
	AddAfterExecuteHook(hook AfterExecuteHook)
	BeforeExecute(input interface{}) error
	AfterExecute(output interface{}, err error)
}

// Span represents a distributed tracing span
type Span interface {
	Name() string
	TraceID() string
	SpanID() string
	ParentID() string
	Attributes() map[string]interface{}
	SetAttribute(key string, value interface{})
	AddEvent(name string, attributes map[string]interface{})
	Events() []SpanEvent
	RecordError(err error)
	HasError() bool
	Status() string
	End()
	IsEnded() bool
	Duration() time.Duration
}

// SpanEvent represents an event within a span
type SpanEvent struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]interface{}
}

// Hook types for execution lifecycle
type (
	BeforeExecuteHook func(ctx AgentContext, input interface{}) error
	AfterExecuteHook  func(ctx AgentContext, output interface{}, err error)
)

// Default resource limits
const (
	DefaultTimeout   = 30 * time.Second
	DefaultMaxMemory = int64(1024 * 1024 * 512) // 512MB
	DefaultMaxCPU    = 4
)

// agentContext is the default implementation of AgentContext
type agentContext struct {
	mu sync.RWMutex

	// Core context
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool

	// Resource limits
	timeout    time.Duration
	maxMemory  int64
	maxCPU     int
	usedMemory atomic.Int64
	usedCPU    atomic.Int32

	// Metadata
	metadata map[string]interface{}

	// Tracing
	traceID     string
	spanID      string
	currentSpan Span
	spanCounter atomic.Uint64

	// Hooks
	beforeHooks []BeforeExecuteHook
	afterHooks  []AfterExecuteHook

	// Multi-engine support
	engine         string
	engineContexts map[string]*agentContext
	parent         *agentContext
}

// AgentContextOption configures an AgentContext
type AgentContextOption func(*agentContext)

// NewAgentContext creates a new agent execution context
func NewAgentContext(parent context.Context, opts ...AgentContextOption) AgentContext {
	ctx, cancel := context.WithCancel(parent)

	ac := &agentContext{
		ctx:            ctx,
		cancel:         cancel,
		timeout:        DefaultTimeout,
		maxMemory:      DefaultMaxMemory,
		maxCPU:         DefaultMaxCPU,
		metadata:       make(map[string]interface{}),
		engineContexts: make(map[string]*agentContext),
	}

	// Apply options
	for _, opt := range opts {
		opt(ac)
	}

	// Apply timeout if set
	if ac.timeout > 0 {
		ctx, cancel = context.WithTimeout(ac.ctx, ac.timeout)
		ac.ctx = ctx
		ac.cancel = cancel
	}

	return ac
}

// Context returns the underlying context.Context
func (c *agentContext) Context() context.Context {
	return c.ctx
}

// Cancel cancels the context
func (c *agentContext) Cancel() {
	c.cancelled.Store(true)
	c.cancel()
}

// IsCancelled returns true if the context has been cancelled
func (c *agentContext) IsCancelled() bool {
	return c.cancelled.Load() || c.ctx.Err() != nil
}

// Done returns a channel that's closed when the context is cancelled
func (c *agentContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Deadline returns the deadline for this context
func (c *agentContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Value returns a value from the context
func (c *agentContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// WithValue returns a new context with the given key-value pair
func (c *agentContext) WithValue(key, value interface{}) AgentContext {
	// We need to update the existing context rather than copying
	// to avoid copying the mutex
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = context.WithValue(c.ctx, key, value)
	return c
}

// Resource limit methods

// Timeout returns the configured timeout
func (c *agentContext) Timeout() time.Duration {
	return c.timeout
}

// MaxMemory returns the maximum memory limit
func (c *agentContext) MaxMemory() int64 {
	return c.maxMemory
}

// MaxCPU returns the maximum CPU cores limit
func (c *agentContext) MaxCPU() int {
	return c.maxCPU
}

// CheckMemoryLimit checks if the required memory is available
func (c *agentContext) CheckMemoryLimit(required int64) bool {
	current := c.usedMemory.Load()
	return current+required <= c.maxMemory
}

// CheckCPULimit checks if the required CPU cores are available
func (c *agentContext) CheckCPULimit(required int) bool {
	current := c.usedCPU.Load()
	return int(current)+required <= c.maxCPU
}

// RecordMemoryUsage records memory usage
func (c *agentContext) RecordMemoryUsage(bytes int64) {
	c.usedMemory.Add(bytes)
	// If this is an engine context, also record in parent
	if c.parent != nil {
		c.parent.RecordMemoryUsage(bytes)
	}
}

// RecordCPUUsage records CPU usage
func (c *agentContext) RecordCPUUsage(cores int) {
	c.usedCPU.Add(int32(cores))
	// If this is an engine context, also record in parent
	if c.parent != nil {
		c.parent.RecordCPUUsage(cores)
	}
}

// ReleaseMemory releases memory
func (c *agentContext) ReleaseMemory(bytes int64) {
	c.usedMemory.Add(-bytes)
	// If this is an engine context, also release in parent
	if c.parent != nil {
		c.parent.ReleaseMemory(bytes)
	}
}

// ReleaseCPU releases CPU cores
func (c *agentContext) ReleaseCPU(cores int) {
	c.usedCPU.Add(-int32(cores))
	// If this is an engine context, also release in parent
	if c.parent != nil {
		c.parent.ReleaseCPU(cores)
	}
}

// CurrentMemory returns current memory usage
func (c *agentContext) CurrentMemory() int64 {
	return c.usedMemory.Load()
}

// CurrentCPU returns current CPU usage
func (c *agentContext) CurrentCPU() int {
	return int(c.usedCPU.Load())
}

// Metadata methods

// Metadata returns a copy of all metadata
func (c *agentContext) Metadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	meta := make(map[string]interface{})
	for k, v := range c.metadata {
		meta[k] = v
	}
	return meta
}

// GetMetadata gets a metadata value
func (c *agentContext) GetMetadata(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metadata[key]
}

// SetMetadata sets a metadata value
func (c *agentContext) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metadata[key] = value
}

// Tracing methods

// TraceID returns the trace ID
func (c *agentContext) TraceID() string {
	return c.traceID
}

// SpanID returns the current span ID
func (c *agentContext) SpanID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.currentSpan != nil {
		return c.currentSpan.SpanID()
	}
	return c.spanID
}

// StartSpan starts a new span
func (c *agentContext) StartSpan(name string, attributes map[string]interface{}) Span {
	c.mu.Lock()
	defer c.mu.Unlock()

	parentID := c.spanID
	if c.currentSpan != nil {
		parentID = c.currentSpan.SpanID()
	}

	spanID := fmt.Sprintf("span-%d", c.spanCounter.Add(1))

	span := &span{
		name:       name,
		traceID:    c.traceID,
		spanID:     spanID,
		parentID:   parentID,
		attributes: make(map[string]interface{}),
		events:     make([]SpanEvent, 0),
		startTime:  time.Now(),
		status:     "ok",
	}

	for k, v := range attributes {
		span.attributes[k] = v
	}

	return span
}

// SetCurrentSpan sets the current span
func (c *agentContext) SetCurrentSpan(s Span) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentSpan = s
}

// Multi-engine support

// ForEngine creates an engine-specific context
func (c *agentContext) ForEngine(engine string) AgentContext {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we already have a context for this engine
	if engineCtx, exists := c.engineContexts[engine]; exists {
		return engineCtx
	}

	// Create new engine context
	engineCtx := &agentContext{
		ctx:       c.ctx,
		cancel:    c.cancel,
		timeout:   c.timeout,
		maxMemory: c.maxMemory,
		maxCPU:    c.maxCPU,
		metadata:  make(map[string]interface{}),
		traceID:   c.traceID,
		spanID:    c.spanID,
		engine:    engine,
		parent:    c,
	}

	// Set engine in metadata
	engineCtx.metadata["engine"] = engine

	c.engineContexts[engine] = engineCtx
	return engineCtx
}

// Hook methods

// AddBeforeExecuteHook adds a before-execute hook
func (c *agentContext) AddBeforeExecuteHook(hook BeforeExecuteHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeHooks = append(c.beforeHooks, hook)
}

// AddAfterExecuteHook adds an after-execute hook
func (c *agentContext) AddAfterExecuteHook(hook AfterExecuteHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterHooks = append(c.afterHooks, hook)
}

// BeforeExecute runs all before-execute hooks
func (c *agentContext) BeforeExecute(input interface{}) error {
	c.mu.RLock()
	hooks := make([]BeforeExecuteHook, len(c.beforeHooks))
	copy(hooks, c.beforeHooks)
	c.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(c, input); err != nil {
			return err
		}
	}
	return nil
}

// AfterExecute runs all after-execute hooks
func (c *agentContext) AfterExecute(output interface{}, err error) {
	c.mu.RLock()
	hooks := make([]AfterExecuteHook, len(c.afterHooks))
	copy(hooks, c.afterHooks)
	c.mu.RUnlock()

	for _, hook := range hooks {
		hook(c, output, err)
	}
}

// Context option functions

// WithContextTimeout sets the timeout for the context
func WithContextTimeout(timeout time.Duration) AgentContextOption {
	return func(c *agentContext) {
		c.timeout = timeout
	}
}

// WithContextMaxMemory sets the maximum memory limit
func WithContextMaxMemory(bytes int64) AgentContextOption {
	return func(c *agentContext) {
		c.maxMemory = bytes
	}
}

// WithContextMaxCPU sets the maximum CPU cores limit
func WithContextMaxCPU(cores int) AgentContextOption {
	return func(c *agentContext) {
		c.maxCPU = cores
	}
}

// WithContextMetadata sets initial metadata
func WithContextMetadata(metadata map[string]interface{}) AgentContextOption {
	return func(c *agentContext) {
		for k, v := range metadata {
			c.metadata[k] = v
		}
	}
}

// WithContextTracing sets tracing information
func WithContextTracing(traceID, spanID string) AgentContextOption {
	return func(c *agentContext) {
		c.traceID = traceID
		c.spanID = spanID
	}
}

// span is the default implementation of Span
type span struct {
	mu         sync.RWMutex
	name       string
	traceID    string
	spanID     string
	parentID   string
	attributes map[string]interface{}
	events     []SpanEvent
	startTime  time.Time
	endTime    time.Time
	ended      bool
	hasError   bool
	status     string
}

func (s *span) Name() string     { return s.name }
func (s *span) TraceID() string  { return s.traceID }
func (s *span) SpanID() string   { return s.spanID }
func (s *span) ParentID() string { return s.parentID }

func (s *span) Attributes() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	attrs := make(map[string]interface{})
	for k, v := range s.attributes {
		attrs[k] = v
	}
	return attrs
}

func (s *span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes[key] = value
}

func (s *span) AddEvent(name string, attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, SpanEvent{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attributes,
	})
}

func (s *span) Events() []SpanEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]SpanEvent, len(s.events))
	copy(events, s.events)
	return events
}

func (s *span) RecordError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hasError = true
	s.status = "error"
	s.attributes["error.message"] = err.Error()
}

func (s *span) HasError() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasError
}

func (s *span) Status() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ended {
		s.ended = true
		s.endTime = time.Now()
	}
}

func (s *span) IsEnded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ended
}

func (s *span) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ended {
		return s.endTime.Sub(s.startTime)
	}
	return time.Since(s.startTime)
}
