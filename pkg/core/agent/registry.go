// ABOUTME: Agent registry that manages thread-safe agent registration, discovery, and lifecycle
// ABOUTME: Provides capability-based discovery, template system, and health monitoring

package agent

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	// ErrAgentNotFound is returned when an agent is not found in the registry
	ErrAgentNotFound = errors.New("agent not found")
	// ErrAgentAlreadyRegistered is returned when trying to register an agent with duplicate ID
	ErrAgentAlreadyRegistered = errors.New("agent already registered")
	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("template not found")
	// ErrTemplateAlreadyRegistered is returned when trying to register a template with duplicate ID
	ErrTemplateAlreadyRegistered = errors.New("template already registered")
)

// Registry manages agent registration and discovery
type Registry struct {
	mu        sync.RWMutex
	agents    map[string]Agent
	templates map[string]*AgentTemplate
}

// global registry instance
var globalRegistry *Registry
var globalRegistryOnce sync.Once

// NewRegistry creates a new agent registry
func NewRegistry() *Registry {
	return &Registry{
		agents:    make(map[string]Agent),
		templates: make(map[string]*AgentTemplate),
	}
}

// GlobalRegistry returns the global agent registry singleton
func GlobalRegistry() *Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// Register adds an agent to the registry
func (r *Registry) Register(ctx context.Context, agent Agent) error {
	return r.RegisterWithOptions(ctx, agent, RegisterOptions{})
}

// RegisterWithOptions adds an agent to the registry with options
func (r *Registry) RegisterWithOptions(ctx context.Context, agent Agent, opts RegisterOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := agent.ID()
	if _, exists := r.agents[id]; exists && !opts.Force {
		return fmt.Errorf("%w: %s", ErrAgentAlreadyRegistered, id)
	}

	r.agents[id] = agent

	// Auto-initialize if requested
	if opts.AutoInit {
		if err := agent.Init(ctx); err != nil {
			// Remove from registry if init fails
			delete(r.agents, id)
			return fmt.Errorf("failed to initialize agent: %w", err)
		}
	}

	return nil
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[id]; !exists {
		return fmt.Errorf("%w: %s", ErrAgentNotFound, id)
	}

	delete(r.agents, id)
	return nil
}

// UnregisterWithCleanup removes an agent and performs cleanup
func (r *Registry) UnregisterWithCleanup(ctx context.Context, id string) error {
	r.mu.Lock()
	agent, exists := r.agents[id]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrAgentNotFound, id)
	}
	delete(r.agents, id)
	r.mu.Unlock()

	// Cleanup outside of lock
	if err := agent.Cleanup(ctx); err != nil {
		return fmt.Errorf("failed to cleanup agent: %w", err)
	}

	return nil
}

// Get retrieves an agent by ID
func (r *Registry) Get(id string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrAgentNotFound, id)
	}

	return agent, nil
}

// List returns all registered agents
func (r *Registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

// FindByCapability finds agents with a specific capability
func (r *Registry) FindByCapability(key string, value interface{}) []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []Agent
	for _, agent := range r.agents {
		caps := agent.Capabilities()
		if capValue, exists := caps[key]; exists {
			if matchCapability(capValue, value) {
				matched = append(matched, agent)
			}
		}
	}
	return matched
}

// FindByCapabilities finds agents matching complex capability criteria
func (r *Registry) FindByCapabilities(filter CapabilityFilter) []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []Agent
	for _, agent := range r.agents {
		if matchesFilter(agent, filter) {
			matched = append(matched, agent)
		}
	}
	return matched
}

// matchesFilter checks if an agent matches the capability filter
func matchesFilter(agent Agent, filter CapabilityFilter) bool {
	caps := agent.Capabilities()

	// Check required capabilities
	for key, value := range filter.Required {
		capValue, exists := caps[key]
		if !exists || !matchCapability(capValue, value) {
			return false
		}
	}

	// Check any capabilities (at least one must match)
	if len(filter.Any) > 0 {
		anyMatched := false
		for key, value := range filter.Any {
			if capValue, exists := caps[key]; exists && matchCapability(capValue, value) {
				anyMatched = true
				break
			}
		}
		if !anyMatched {
			return false
		}
	}

	// Check excluded capabilities
	for key, value := range filter.Excluded {
		if capValue, exists := caps[key]; exists && matchCapability(capValue, value) {
			return false
		}
	}

	return true
}

// matchCapability checks if capability values match
func matchCapability(capValue, filterValue interface{}) bool {
	// Direct equality check
	if reflect.DeepEqual(capValue, filterValue) {
		return true
	}

	// Check if filterValue is a slice and capValue contains any element
	if filterSlice, ok := filterValue.([]string); ok {
		if capSlice, ok := capValue.([]string); ok {
			for _, filterItem := range filterSlice {
				for _, capItem := range capSlice {
					if filterItem == capItem {
						return true
					}
				}
			}
		}
	}

	return false
}

// InitAll initializes all registered agents
func (r *Registry) InitAll(ctx context.Context) error {
	r.mu.RLock()
	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	r.mu.RUnlock()

	for _, agent := range agents {
		if err := agent.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize agent %s: %w", agent.ID(), err)
		}
	}
	return nil
}

// CleanupAll cleans up all registered agents
func (r *Registry) CleanupAll(ctx context.Context) error {
	r.mu.RLock()
	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	r.mu.RUnlock()

	var firstErr error
	for _, agent := range agents {
		if err := agent.Cleanup(ctx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to cleanup agent %s: %w", agent.ID(), err)
		}
	}
	return firstErr
}

// RegisterTemplate registers an agent template
func (r *Registry) RegisterTemplate(template *AgentTemplate) error {
	if template == nil || template.ID == "" {
		return errors.New("invalid template")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[template.ID]; exists {
		return fmt.Errorf("%w: %s", ErrTemplateAlreadyRegistered, template.ID)
	}

	r.templates[template.ID] = template
	return nil
}

// GetTemplate retrieves a template by ID
func (r *Registry) GetTemplate(id string) *AgentTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates[id]
}

// ListTemplates returns all registered templates
func (r *Registry) ListTemplates() []*AgentTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]*AgentTemplate, 0, len(r.templates))
	for _, template := range r.templates {
		templates = append(templates, template)
	}
	return templates
}

// CreateFromTemplate creates and registers an agent from a template
func (r *Registry) CreateFromTemplate(ctx context.Context, templateID, agentID string, params map[string]interface{}) (Agent, error) {
	r.mu.RLock()
	template, exists := r.templates[templateID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, templateID)
	}

	// Create agent using template factory
	agent, err := template.Factory(agentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent from template: %w", err)
	}

	// Register the new agent
	if err := r.Register(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Stats returns registry statistics
func (r *Registry) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RegistryStats{
		TotalAgents: len(r.agents),
	}

	for _, agent := range r.agents {
		// Check if agent implements ExtendedAgent interface
		if ea, ok := agent.(ExtendedAgent); ok {
			switch ea.Status() {
			case StatusCreated:
				stats.CreatedAgents++
			case StatusReady:
				stats.ReadyAgents++
			case StatusRunning:
				stats.RunningAgents++
			case StatusError:
				stats.ErrorAgents++
			case StatusStopped:
				stats.StoppedAgents++
			}
		}
	}

	return stats
}

// HealthCheck performs a health check on all agents
func (r *Registry) HealthCheck(ctx context.Context) HealthCheckReport {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report := HealthCheckReport{
		Healthy:   true,
		Timestamp: time.Now(),
	}

	for id, agent := range r.agents {
		// Check if agent implements ExtendedAgent interface
		if ea, ok := agent.(ExtendedAgent); ok {
			status := ea.Status()
			if status == StatusReady || status == StatusRunning {
				report.HealthyAgents = append(report.HealthyAgents, id)
			} else if status == StatusError {
				report.UnhealthyAgents = append(report.UnhealthyAgents, id)
				report.Healthy = false
			}
		}
	}

	// Add issues
	if len(report.UnhealthyAgents) > 0 {
		report.Issues = append(report.Issues, fmt.Sprintf("%d agent(s) in error state", len(report.UnhealthyAgents)))
	}

	return report
}

// CapabilityFilter defines criteria for finding agents by capabilities
type CapabilityFilter struct {
	Required map[string]interface{} // All must match
	Any      map[string]interface{} // At least one must match
	Excluded map[string]interface{} // None must match
}

// RegisterOptions provides options for agent registration
type RegisterOptions struct {
	Force    bool // Replace existing agent
	AutoInit bool // Initialize after registration
}

// AgentTemplate defines a template for creating agents
type AgentTemplate struct {
	ID          string
	Name        string
	Description string
	Config      BaseAgentConfig
	Factory     func(id string, params map[string]interface{}) (Agent, error)
}

// RegistryStats provides statistics about the registry
type RegistryStats struct {
	TotalAgents   int
	ReadyAgents   int
	RunningAgents int
	ErrorAgents   int
	StoppedAgents int
	CreatedAgents int
}

// HealthCheckReport provides health information about agents
type HealthCheckReport struct {
	Healthy         bool
	HealthyAgents   []string
	UnhealthyAgents []string
	Issues          []string
	Timestamp       time.Time
}
