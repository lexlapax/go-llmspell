// ABOUTME: This file implements configuration loading using Koanf v2 with support for files, environment variables, and flags.
// ABOUTME: It provides a layered configuration system with defaults → file → env → flags priority order.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	yamlv3 "gopkg.in/yaml.v3"
)

// Loader handles configuration loading from multiple sources
type Loader struct {
	koanf   *koanf.Koanf
	options LoaderOptions
	mu      sync.RWMutex
}

// LoaderOptions configures the loader behavior
type LoaderOptions struct {
	// Configuration file settings
	ConfigFile   string
	ConfigPaths  []string
	ConfigName   string
	ConfigFormat string

	// Environment variable settings
	EnvPrefix    string
	EnvDelimiter string

	// Watch settings
	WatchConfig   bool
	WatchCallback func(*Config) error

	// Validation settings
	ValidateOnLoad bool
	StrictMode     bool
}

// NewLoader creates a new configuration loader
func NewLoader(options LoaderOptions) *Loader {
	if options.ConfigName == "" {
		options.ConfigName = "config"
	}
	if options.ConfigFormat == "" {
		options.ConfigFormat = "yaml"
	}
	if options.EnvPrefix == "" {
		options.EnvPrefix = "LLMSPELL_"
	}
	if options.EnvDelimiter == "" {
		options.EnvDelimiter = "_"
	}
	if len(options.ConfigPaths) == 0 {
		options.ConfigPaths = getDefaultConfigPaths()
	}

	return &Loader{
		koanf:   koanf.New("."),
		options: options,
	}
}

// LoadConfig loads configuration from all sources in priority order
func (l *Loader) LoadConfig() (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Step 1: Load defaults
	defaultConfig := GetDefaultConfig()
	if err := l.koanf.Load(structs.Provider(defaultConfig, "yaml"), nil); err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Step 2: Load from configuration file
	configFile, err := l.findConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find config file: %w", err)
	}

	if configFile != "" {
		if err := l.loadConfigFile(configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	}

	// Step 3: Load from environment variables
	if err := l.loadEnvironmentVars(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Step 4: Unmarshal to config struct
	var config Config
	if err := l.koanf.UnmarshalWithConf("", &config, koanf.UnmarshalConf{
		Tag: "yaml",
	}); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Step 5: Validate configuration if requested
	if l.options.ValidateOnLoad {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	// Step 6: Set up file watching if requested
	if l.options.WatchConfig && configFile != "" {
		if err := l.setupFileWatcher(configFile, &config); err != nil {
			return nil, fmt.Errorf("failed to setup config file watcher: %w", err)
		}
	}

	return &config, nil
}

// LoadConfigFromFile loads configuration from a specific file
func (l *Loader) LoadConfigFromFile(filepath string) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Reset koanf instance
	l.koanf = koanf.New(".")

	// Load defaults first
	defaultConfig := GetDefaultConfig()
	if err := l.koanf.Load(structs.Provider(defaultConfig, "yaml"), nil); err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Load from specified file
	if err := l.loadConfigFile(filepath); err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filepath, err)
	}

	// Load environment variables
	if err := l.loadEnvironmentVars(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Unmarshal to config struct
	var config Config
	if err := l.koanf.UnmarshalWithConf("", &config, koanf.UnmarshalConf{
		Tag: "yaml",
	}); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if l.options.ValidateOnLoad {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return &config, nil
}

// MergeFlags merges command-line flags into the configuration
func (l *Loader) MergeFlags(flagsMap map[string]interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Set flags directly in koanf
	for key, value := range flagsMap {
		_ = l.koanf.Set(key, value)
	}

	return nil
}

// GetCurrentConfig returns the current configuration state
func (l *Loader) GetCurrentConfig() (*Config, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var config Config
	if err := l.koanf.UnmarshalWithConf("", &config, koanf.UnmarshalConf{
		Tag: "yaml",
	}); err != nil {
		return nil, fmt.Errorf("failed to unmarshal current config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the current configuration to a file
func (l *Loader) SaveConfig(config *Config, configPath string) error {
	// Marshal config directly to YAML using yaml.v3 package
	yamlData, err := yamlv3.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Watch starts watching the configuration file for changes
func (l *Loader) Watch(callback func(*Config) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	configFile, err := l.findConfigFile()
	if err != nil {
		return fmt.Errorf("failed to find config file for watching: %w", err)
	}

	if configFile == "" {
		return fmt.Errorf("no config file found to watch")
	}

	l.options.WatchConfig = true
	l.options.WatchCallback = callback

	return l.setupFileWatcher(configFile, nil)
}

// GetRaw returns the raw configuration value at the given key
func (l *Loader) GetRaw(key string) interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.koanf.Get(key)
}

// SetRaw sets a raw configuration value at the given key
func (l *Loader) SetRaw(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.koanf.Set(key, value)
}

// Keys returns all configuration keys
func (l *Loader) Keys() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.koanf.Keys()
}

// findConfigFile searches for a configuration file in the configured paths
func (l *Loader) findConfigFile() (string, error) {
	// If a specific config file is provided, use it
	if l.options.ConfigFile != "" {
		if _, err := os.Stat(l.options.ConfigFile); err == nil {
			return l.options.ConfigFile, nil
		}
		return "", fmt.Errorf("specified config file not found: %s", l.options.ConfigFile)
	}

	// Search in configured paths
	filename := fmt.Sprintf("%s.%s", l.options.ConfigName, l.options.ConfigFormat)

	for _, path := range l.options.ConfigPaths {
		// Expand home directory
		expandedPath := expandPath(path)
		configPath := filepath.Join(expandedPath, filename)

		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// No config file found - this is not an error
	return "", nil
}

// loadConfigFile loads configuration from a file
func (l *Loader) loadConfigFile(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configFile)
	}

	var parser koanf.Parser
	ext := strings.ToLower(filepath.Ext(configFile))

	switch ext {
	case ".yaml", ".yml":
		parser = yaml.Parser()
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err := l.koanf.Load(file.Provider(configFile), parser); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	return nil
}

// loadEnvironmentVars loads configuration from environment variables
func (l *Loader) loadEnvironmentVars() error {
	return l.koanf.Load(env.Provider(l.options.EnvPrefix, ".", func(s string) string {
		// Convert LLMSPELL_ENGINE_MEMORY_LIMIT to engine.memory_limit
		s = strings.ToLower(strings.TrimPrefix(s, l.options.EnvPrefix))
		return strings.ReplaceAll(s, "_", ".")
	}), nil)
}

// setupFileWatcher sets up file watching for configuration changes
func (l *Loader) setupFileWatcher(configFile string, config *Config) error {
	// File watching implementation would go here
	// For now, we'll just return nil as file watching is complex
	// and might require additional dependencies like fsnotify
	return nil
}

// getDefaultConfigPaths returns the default configuration search paths
func getDefaultConfigPaths() []string {
	return []string{
		".",                  // Current directory
		"~/.llmspell",        // User config directory
		"~/.config/llmspell", // XDG config directory
		"/etc/llmspell",      // System config directory
	}
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return homeDir
	}

	return filepath.Join(homeDir, path[2:])
}

// setNestedValue sets a value in a nested map using dot notation
func setNestedValue(m map[string]interface{}, key string, value interface{}) {
	keys := strings.Split(key, ".")
	current := m

	for _, k := range keys[:len(keys)-1] {
		if _, exists := current[k]; !exists {
			current[k] = make(map[string]interface{})
		}

		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			// Override non-map value with map
			current[k] = make(map[string]interface{})
			current = current[k].(map[string]interface{})
		}
	}

	current[keys[len(keys)-1]] = value
}

// GetDefaultConfigFile returns the default configuration file path
func GetDefaultConfigFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.yaml"
	}
	return filepath.Join(homeDir, ".llmspell", "config.yaml")
}

// CreateDefaultConfigFile creates a default configuration file
func CreateDefaultConfigFile(configPath string) error {
	config := GetDefaultConfig()

	// Marshal config directly to YAML
	yamlData, err := yamlv3.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
