// ABOUTME: Base agent implementation that provides common functionality for all agents
// ABOUTME: Includes state management, event emission, error handling, and metrics collection

package agent

import (
	"context"
	"errors"
	"sync"
	"time"
)

// BaseAgentConfig holds configuration for creating a base agent
type BaseAgentConfig struct {
	ID           string
	Name         string
	Description  string
	Version      string
	Capabilities map[string]interface{}
	Metadata     map[string]interface{}
}

// BaseAgent provides a base implementation of the Agent interface
// with common functionality like state management, event emission,
// error handling, and metrics collection
type BaseAgent struct {
	mu sync.RWMutex

	// Basic properties
	id           string
	name         string
	description  string
	version      string
	capabilities map[string]interface{}
	metadata     map[string]interface{}

	// Status tracking
	status    Status
	lastError error

	// State management
	state map[string]interface{}

	// Event handling
	eventBus         *EventBus
	subscribers      map[string][]chan Event
	subscriberStopCh map[chan Event]chan struct{}

	// Metrics
	metrics map[string]float64

	// Configuration
	config map[string]interface{}

	// Run function
	runFunc func(context.Context, interface{}) (interface{}, error)
}

// NewBaseAgent creates a new base agent with the given configuration
func NewBaseAgent(config BaseAgentConfig) (*BaseAgent, error) {
	// Validate configuration
	if config.ID == "" {
		return nil, errors.New("agent ID is required")
	}
	if config.Name == "" {
		return nil, errors.New("agent name is required")
	}

	agent := &BaseAgent{
		id:               config.ID,
		name:             config.Name,
		description:      config.Description,
		version:          config.Version,
		capabilities:     make(map[string]interface{}),
		metadata:         make(map[string]interface{}),
		status:           StatusCreated,
		state:            make(map[string]interface{}),
		eventBus:         NewEventBus(),
		subscribers:      make(map[string][]chan Event),
		subscriberStopCh: make(map[chan Event]chan struct{}),
		metrics:          make(map[string]float64),
		config:           make(map[string]interface{}),
	}

	// Copy capabilities
	if config.Capabilities != nil {
		for k, v := range config.Capabilities {
			agent.capabilities[k] = v
		}
	}

	// Copy metadata
	if config.Metadata != nil {
		for k, v := range config.Metadata {
			agent.metadata[k] = v
		}
	}

	return agent, nil
}

// ID returns the agent's unique identifier
func (a *BaseAgent) ID() string {
	return a.id
}

// Name returns the agent's name
func (a *BaseAgent) Name() string {
	return a.name
}

// Description returns the agent's description
func (a *BaseAgent) Description() string {
	return a.description
}

// Version returns the agent's version
func (a *BaseAgent) Version() string {
	return a.version
}

// Capabilities returns the agent's capabilities
func (a *BaseAgent) Capabilities() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	caps := make(map[string]interface{})
	for k, v := range a.capabilities {
		caps[k] = v
	}
	return caps
}

// Metadata returns the agent's metadata
func (a *BaseAgent) Metadata() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	meta := make(map[string]interface{})
	for k, v := range a.metadata {
		meta[k] = v
	}
	return meta
}

// Status returns the current status of the agent
func (a *BaseAgent) Status() Status {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

// SetStatus updates the agent's status
func (a *BaseAgent) SetStatus(status Status) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.status = status
}

// Init initializes the agent
func (a *BaseAgent) Init(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status == StatusReady || a.status == StatusRunning {
		return nil // Already initialized
	}

	a.status = StatusInitializing

	// Emit initialization event
	a.eventBus.Emit(Event{
		Type:      "agent.init",
		Source:    a.id,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"agent_id": a.id},
	})

	a.status = StatusReady
	return nil
}

// Run executes the agent's main logic
func (a *BaseAgent) Run(ctx context.Context, input interface{}) (interface{}, error) {
	a.mu.Lock()
	if a.runFunc == nil {
		a.mu.Unlock()
		return nil, errors.New("no run function defined")
	}
	a.status = StatusRunning
	runFunc := a.runFunc
	a.mu.Unlock()

	// Execute the run function
	result, err := runFunc(ctx, input)

	a.mu.Lock()
	if err != nil {
		a.status = StatusError
		a.lastError = err
	} else {
		a.status = StatusReady
	}
	a.mu.Unlock()

	return result, err
}

// Cleanup performs cleanup when the agent is done
func (a *BaseAgent) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status == StatusStopped {
		return nil // Already stopped
	}

	a.status = StatusStopping

	// Emit cleanup event
	a.eventBus.Emit(Event{
		Type:      "agent.cleanup",
		Source:    a.id,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"agent_id": a.id},
	})

	// Give time for the event to be delivered
	time.Sleep(10 * time.Millisecond)

	// Stop all event subscriptions
	for _, stopCh := range a.subscriberStopCh {
		close(stopCh)
	}

	// Clear all subscriptions
	a.subscribers = make(map[string][]chan Event)
	a.subscriberStopCh = make(map[chan Event]chan struct{})

	a.status = StatusStopped
	return nil
}

// SetRunFunc sets the function to be executed by Run
func (a *BaseAgent) SetRunFunc(fn func(context.Context, interface{}) (interface{}, error)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runFunc = fn
}

// GetState returns a copy of the current state
func (a *BaseAgent) GetState(ctx context.Context) map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	state := make(map[string]interface{})
	for k, v := range a.state {
		state[k] = v
	}
	return state
}

// SetState replaces the entire state
func (a *BaseAgent) SetState(ctx context.Context, state map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state = make(map[string]interface{})
	for k, v := range state {
		a.state[k] = v
	}
	return nil
}

// UpdateState updates specific state values
func (a *BaseAgent) UpdateState(ctx context.Context, updates map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for k, v := range updates {
		a.state[k] = v
	}
	return nil
}

// ClearState removes all state
func (a *BaseAgent) ClearState(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.state = make(map[string]interface{})
	return nil
}

// Subscribe subscribes to events of a specific type
func (a *BaseAgent) Subscribe(ctx context.Context, eventType string) <-chan Event {
	a.mu.Lock()
	defer a.mu.Unlock()

	ch := make(chan Event, 100)
	stopCh := make(chan struct{})

	a.subscribers[eventType] = append(a.subscribers[eventType], ch)
	a.subscriberStopCh[ch] = stopCh

	// Subscribe to event bus
	a.eventBus.Subscribe(eventType, func(event Event) {
		select {
		case <-stopCh:
			// Subscription cancelled, don't send
			return
		default:
			select {
			case ch <- event:
			default:
				// Channel full, drop event
			}
		}
	})

	return ch
}

// Unsubscribe removes a subscription
func (a *BaseAgent) Unsubscribe(ctx context.Context, ch <-chan Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Find and remove the channel
	for eventType, chans := range a.subscribers {
		for i, c := range chans {
			if c == ch {
				// Remove from subscribers
				a.subscribers[eventType] = append(chans[:i], chans[i+1:]...)

				// Signal stop to the subscription handler
				if stopCh, ok := a.subscriberStopCh[c]; ok {
					close(stopCh)
					delete(a.subscriberStopCh, c)
				}

				return nil
			}
		}
	}
	return nil
}

// EmitEvent emits an event
func (a *BaseAgent) EmitEvent(ctx context.Context, event Event) error {
	a.eventBus.Emit(event)
	return nil
}

// HandleError handles an error
func (a *BaseAgent) HandleError(ctx context.Context, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastError = err
	a.status = StatusError

	// Emit error event
	a.eventBus.Emit(Event{
		Type:      "agent.error",
		Source:    a.id,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

// LastError returns the last error
func (a *BaseAgent) LastError() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastError
}

// Recover attempts to recover from error state
func (a *BaseAgent) Recover(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.status != StatusError {
		return nil
	}

	a.lastError = nil
	a.status = StatusReady
	return nil
}

// ExecuteWithRetry executes an operation with retries
func (a *BaseAgent) ExecuteWithRetry(ctx context.Context, operation func() error, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries {
				time.Sleep(retryDelay)
			}
		}
	}

	// All retries failed
	a.HandleError(ctx, lastErr)
	return lastErr
}

// RecordMetric records a metric value
func (a *BaseAgent) RecordMetric(ctx context.Context, name string, value float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics[name] = value
}

// IncrementMetric increments a metric value
func (a *BaseAgent) IncrementMetric(ctx context.Context, name string, delta float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics[name] += delta
}

// GetMetrics returns a copy of all metrics
func (a *BaseAgent) GetMetrics(ctx context.Context) map[string]float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	metrics := make(map[string]float64)
	for k, v := range a.metrics {
		metrics[k] = v
	}
	return metrics
}

// ResetMetrics clears all metrics
func (a *BaseAgent) ResetMetrics(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics = make(map[string]float64)
}

// Configure applies configuration options
func (a *BaseAgent) Configure(options ...AgentOption) error {
	for _, option := range options {
		if err := option(a); err != nil {
			return err
		}
	}
	return nil
}

// GetConfig returns the current configuration
func (a *BaseAgent) GetConfig() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	config := make(map[string]interface{})
	for k, v := range a.config {
		config[k] = v
	}
	return config
}

// RunAsync executes the agent asynchronously
func (a *BaseAgent) RunAsync(ctx context.Context, input interface{}) (<-chan interface{}, <-chan error) {
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := a.Run(ctx, input)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
		close(resultChan)
		close(errChan)
	}()

	return resultChan, errChan
}

// Configuration options

// WithTimeout sets the agent timeout
func WithTimeout(timeout time.Duration) AgentOption {
	return func(agent Agent) error {
		if ba, ok := agent.(*BaseAgent); ok {
			ba.mu.Lock()
			defer ba.mu.Unlock()
			ba.config["timeout"] = timeout
		}
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) AgentOption {
	return func(agent Agent) error {
		if ba, ok := agent.(*BaseAgent); ok {
			ba.mu.Lock()
			defer ba.mu.Unlock()
			ba.config["maxRetries"] = maxRetries
		}
		return nil
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) AgentOption {
	return func(agent Agent) error {
		if ba, ok := agent.(*BaseAgent); ok {
			ba.mu.Lock()
			defer ba.mu.Unlock()
			ba.config["debug"] = debug
		}
		return nil
	}
}

// Event represents an agent event
type Event struct {
	Type      string
	Source    string
	Timestamp time.Time
	Data      map[string]interface{}
}

// EventBus manages event subscriptions and emissions
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(Event)
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(Event)),
	}
}

// Subscribe adds a subscriber for an event type
func (eb *EventBus) Subscribe(eventType string, handler func(Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Emit sends an event to all subscribers
func (eb *EventBus) Emit(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Notify specific subscribers
	if handlers, ok := eb.subscribers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}

	// Notify wildcard subscribers
	if handlers, ok := eb.subscribers["*"]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// Ensure BaseAgent implements the required interfaces
var (
	_ Agent             = (*BaseAgent)(nil)
	_ ExtendedAgent     = (*BaseAgent)(nil)
	_ AsyncAgent        = (*BaseAgent)(nil)
	_ ConfigurableAgent = (*BaseAgent)(nil)
)
