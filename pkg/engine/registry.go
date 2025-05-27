// ABOUTME: Engine registry for managing and discovering script engines
// ABOUTME: Provides thread-safe registration and factory pattern support

package engine

import (
	"fmt"
	"strings"
	"sync"
)

// EngineFactory is a function that creates a new engine instance
type EngineFactory func(config Config) (Engine, error)

// EngineMetadata contains additional information about an engine
type EngineMetadata struct {
	// Description provides a human-readable description of the engine
	Description string

	// FileExtensions lists the file extensions this engine handles (e.g., [".lua"])
	FileExtensions []string

	// MimeTypes lists the MIME types this engine handles
	MimeTypes []string

	// Version indicates the engine version
	Version string
}

// engineEntry holds an engine factory and its metadata
type engineEntry struct {
	factory  EngineFactory
	metadata EngineMetadata
}

// Registry manages engine factories and provides thread-safe access
type Registry struct {
	mu      sync.RWMutex
	engines map[string]engineEntry
}

// NewRegistry creates a new engine registry
func NewRegistry() *Registry {
	return &Registry{
		engines: make(map[string]engineEntry),
	}
}

// Register adds a new engine factory to the registry
func (r *Registry) Register(name string, factory EngineFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.engines[name]; exists {
		return fmt.Errorf("engine %q already registered", name)
	}

	r.engines[name] = engineEntry{
		factory: factory,
	}

	return nil
}

// RegisterWithMetadata adds a new engine factory with metadata to the registry
func (r *Registry) RegisterWithMetadata(name string, factory EngineFactory, metadata EngineMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.engines[name]; exists {
		return fmt.Errorf("engine %q already registered", name)
	}

	r.engines[name] = engineEntry{
		factory:  factory,
		metadata: metadata,
	}

	return nil
}

// GetFactory retrieves an engine factory by name
func (r *Registry) GetFactory(name string) (EngineFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.engines[name]
	if !exists {
		return nil, fmt.Errorf("engine %q not found", name)
	}

	return entry.factory, nil
}

// List returns the names of all registered engines
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.engines))
	for name := range r.engines {
		names = append(names, name)
	}

	return names
}

// Unregister removes an engine from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.engines[name]; !exists {
		return fmt.Errorf("engine %q not found", name)
	}

	delete(r.engines, name)
	return nil
}

// DiscoverByExtension finds an engine that handles the given file extension
func (r *Registry) DiscoverByExtension(ext string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext = strings.ToLower(ext)

	for name, entry := range r.engines {
		for _, supportedExt := range entry.metadata.FileExtensions {
			if strings.ToLower(supportedExt) == ext {
				return name, nil
			}
		}
	}

	return "", fmt.Errorf("no engine found for extension %q", ext)
}

// DiscoverByMimeType finds an engine that handles the given MIME type
func (r *Registry) DiscoverByMimeType(mimeType string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mimeType = strings.ToLower(mimeType)

	for name, entry := range r.engines {
		for _, supportedType := range entry.metadata.MimeTypes {
			if strings.ToLower(supportedType) == mimeType {
				return name, nil
			}
		}
	}

	return "", fmt.Errorf("no engine found for MIME type %q", mimeType)
}

// GetMetadata retrieves metadata for a registered engine
func (r *Registry) GetMetadata(name string) (EngineMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.engines[name]
	if !exists {
		return EngineMetadata{}, fmt.Errorf("engine %q not found", name)
	}

	return entry.metadata, nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// RegisterEngine registers an engine factory in the global registry
func RegisterEngine(name string, factory EngineFactory) error {
	return globalRegistry.Register(name, factory)
}

// RegisterEngineWithMetadata registers an engine factory with metadata in the global registry
func RegisterEngineWithMetadata(name string, factory EngineFactory, metadata EngineMetadata) error {
	return globalRegistry.RegisterWithMetadata(name, factory, metadata)
}

// CreateEngine creates a new engine instance using the global registry
func CreateEngine(name string, config Config) (Engine, error) {
	factory, err := globalRegistry.GetFactory(name)
	if err != nil {
		return nil, err
	}

	return factory(config)
}

// ListEngines returns the names of all engines in the global registry
func ListEngines() []string {
	return globalRegistry.List()
}

// UnregisterEngine removes an engine from the global registry
func UnregisterEngine(name string) error {
	return globalRegistry.Unregister(name)
}

// DiscoverEngineByExtension finds an engine in the global registry by file extension
func DiscoverEngineByExtension(ext string) (string, error) {
	return globalRegistry.DiscoverByExtension(ext)
}

// DiscoverEngineByMimeType finds an engine in the global registry by MIME type
func DiscoverEngineByMimeType(mimeType string) (string, error) {
	return globalRegistry.DiscoverByMimeType(mimeType)
}
