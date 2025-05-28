// ABOUTME: Storage module for Lua scripts with file-based persistence
// ABOUTME: Provides storage.get(), set(), exists(), read(), write() functions

package stdlib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// StorageConfig holds configuration for the storage module
type StorageConfig struct {
	BaseDir     string
	MaxFileSize int64
	AllowedExts []string
}

// DefaultStorageConfig returns a default storage configuration
func DefaultStorageConfig() *StorageConfig {
	homeDir, _ := os.UserHomeDir()
	return &StorageConfig{
		BaseDir:     filepath.Join(homeDir, ".llmspell", "storage"),
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		AllowedExts: []string{".txt", ".json", ".yaml", ".yml", ".md"},
	}
}

// Storage provides file-based storage for Lua scripts
type Storage struct {
	config *StorageConfig
	memory map[string]string
	mu     sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(config *StorageConfig) (*Storage, error) {
	if config == nil {
		config = DefaultStorageConfig()
	}

	// Ensure base directory exists
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		config: config,
		memory: make(map[string]string),
	}, nil
}

// RegisterStorage registers the storage module with all functions
func RegisterStorage(L *lua.LState, storage *Storage) {
	// Create storage module table
	storageModule := L.NewTable()

	// In-memory key-value functions
	L.SetField(storageModule, "get", L.NewClosure(storage.get))
	L.SetField(storageModule, "set", L.NewClosure(storage.set))

	// File-based functions
	L.SetField(storageModule, "exists", L.NewClosure(storage.exists))
	L.SetField(storageModule, "read", L.NewClosure(storage.read))
	L.SetField(storageModule, "write", L.NewClosure(storage.write))
	L.SetField(storageModule, "delete", L.NewClosure(storage.delete))
	L.SetField(storageModule, "list", L.NewClosure(storage.list))

	// Register the module
	L.SetGlobal("storage", storageModule)
}

// get retrieves a value from in-memory storage
func (s *Storage) get(L *lua.LState) int {
	key := L.CheckString(1)

	s.mu.RLock()
	value, exists := s.memory[key]
	s.mu.RUnlock()

	if exists {
		L.Push(lua.LString(value))
	} else {
		L.Push(lua.LNil)
	}
	return 1
}

// set stores a value in in-memory storage
func (s *Storage) set(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)

	s.mu.Lock()
	s.memory[key] = value
	s.mu.Unlock()

	return 0
}

// validatePath ensures the path is safe and within bounds
func (s *Storage) validatePath(filename string) (string, error) {
	// Clean the filename
	filename = filepath.Clean(filename)

	// Ensure no directory traversal
	if strings.Contains(filename, "..") {
		return "", fmt.Errorf("invalid filename: directory traversal not allowed")
	}

	// Check extension if specified
	if len(s.config.AllowedExts) > 0 {
		ext := filepath.Ext(filename)
		allowed := false
		for _, allowedExt := range s.config.AllowedExts {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("file extension %s not allowed", ext)
		}
	}

	// Build full path
	fullPath := filepath.Join(s.config.BaseDir, filename)

	// Ensure it's still within base directory
	if !strings.HasPrefix(fullPath, s.config.BaseDir) {
		return "", fmt.Errorf("invalid path: outside storage directory")
	}

	return fullPath, nil
}

// exists checks if a file exists
func (s *Storage) exists(L *lua.LState) int {
	filename := L.CheckString(1)

	fullPath, err := s.validatePath(filename)
	if err != nil {
		L.Push(lua.LBool(false))
		return 1
	}

	_, err = os.Stat(fullPath)
	L.Push(lua.LBool(err == nil))
	return 1
}

// read reads a file from storage
func (s *Storage) read(L *lua.LState) int {
	filename := L.CheckString(1)

	fullPath, err := s.validatePath(filename)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Check file size
	info, err := os.Stat(fullPath)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("file not found: %s", filename)))
		return 2
	}

	if info.Size() > s.config.MaxFileSize {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("file too large: %d bytes (max %d)", info.Size(), s.config.MaxFileSize)))
		return 2
	}

	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(content)))
	return 1
}

// write writes content to a file
func (s *Storage) write(L *lua.LState) int {
	filename := L.CheckString(1)
	content := L.CheckString(2)

	fullPath, err := s.validatePath(filename)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	// Check content size
	if int64(len(content)) > s.config.MaxFileSize {
		L.Push(lua.LString(fmt.Sprintf("content too large: %d bytes (max %d)", len(content), s.config.MaxFileSize)))
		return 1
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to create directory: %v", err)))
		return 1
	}

	// Write file
	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// delete removes a file from storage
func (s *Storage) delete(L *lua.LState) int {
	filename := L.CheckString(1)

	fullPath, err := s.validatePath(filename)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// list lists files in storage directory
func (s *Storage) list(L *lua.LState) int {
	pattern := L.OptString(1, "*")

	// List files in storage directory
	files, err := filepath.Glob(filepath.Join(s.config.BaseDir, pattern))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create table of relative paths
	table := L.NewTable()
	for i, file := range files {
		relPath, err := filepath.Rel(s.config.BaseDir, file)
		if err == nil {
			// Skip directories
			info, err := os.Stat(file)
			if err == nil && !info.IsDir() {
				table.RawSetInt(i+1, lua.LString(relPath))
			}
		}
	}

	L.Push(table)
	return 1
}
