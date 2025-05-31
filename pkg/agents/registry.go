// ABOUTME: Implements a thread-safe registry for agent factories and instances
// ABOUTME: Follows the established registry pattern used throughout the codebase

package agents

import (
	"context"
	"fmt"
	"sync"
)

// registry is the default implementation of the Registry interface
type registry struct {
	factories map[string]Factory
	agents    map[string]Agent
	mu        sync.RWMutex
}

// NewRegistry creates a new agent registry
func NewRegistry() Registry {
	return &registry{
		factories: make(map[string]Factory),
		agents:    make(map[string]Agent),
	}
}

// Register adds a new agent factory
func (r *registry) Register(name string, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("agent factory %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Create creates a new agent instance
func (r *registry) Create(config Config) (Agent, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	r.mu.RLock()
	factory, exists := r.factories[config.Provider]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent factory %s not found", config.Provider)
	}

	// Create the agent
	agent, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Initialize the agent
	ctx := context.Background()
	if err := agent.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize agent: %w", err)
	}

	// Store the agent
	r.mu.Lock()
	r.agents[config.Name] = agent
	r.mu.Unlock()

	return agent, nil
}

// Get retrieves an existing agent by name
func (r *registry) Get(name string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", name)
	}

	return agent, nil
}

// List returns all registered factory names
func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

// Remove removes an agent from the registry
func (r *registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[name]
	if !exists {
		return fmt.Errorf("agent %s not found", name)
	}

	// Cleanup the agent
	if err := agent.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup agent: %w", err)
	}

	delete(r.agents, name)
	return nil
}

// Global registry instance
var (
	globalRegistry Registry
	once           sync.Once
)

// DefaultRegistry returns the global agent registry
func DefaultRegistry() Registry {
	once.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// RegisterAgentFactory registers an agent factory with the global registry
func RegisterAgentFactory(name string, factory Factory) error {
	return DefaultRegistry().Register(name, factory)
}

// CreateAgent creates a new agent using the global registry
func CreateAgent(config Config) (Agent, error) {
	return DefaultRegistry().Create(config)
}

// GetAgent retrieves an agent from the global registry
func GetAgent(name string) (Agent, error) {
	return DefaultRegistry().Get(name)
}

// ListAgentFactories lists all registered agent factories in the global registry
func ListAgentFactories() []string {
	return DefaultRegistry().List()
}

// RemoveAgent removes an agent from the global registry
func RemoveAgent(name string) error {
	return DefaultRegistry().Remove(name)
}