// ABOUTME: Implements the tool registry for managing and discovering tools
// ABOUTME: Provides thread-safe registration, lookup, and listing of tools

package tools

import (
	"fmt"
	"sync"
)

// registry is the default implementation of the Registry interface
type registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() Registry {
	return &registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *registry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool must have a non-empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool with name %q already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}

	return tool, nil
}

// List returns all registered tools
func (r *registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Remove unregisters a tool
func (r *registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool %q not found", name)
	}

	delete(r.tools, name)
	return nil
}

// DefaultRegistry is the global default registry
var DefaultRegistry = NewRegistry()

// Register adds a tool to the default registry
func Register(tool Tool) error {
	return DefaultRegistry.Register(tool)
}

// Get retrieves a tool from the default registry
func Get(name string) (Tool, error) {
	return DefaultRegistry.Get(name)
}

// List returns all tools from the default registry
func List() []Tool {
	return DefaultRegistry.List()
}

// Remove unregisters a tool from the default registry
func Remove(name string) error {
	return DefaultRegistry.Remove(name)
}
