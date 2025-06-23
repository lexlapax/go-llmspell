// ABOUTME: Embeds all stdlib Lua modules for deployment without filesystem dependencies
// ABOUTME: Provides access to embedded .lua files and module listing functionality

package stdlib

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

// StdlibFS embeds all .lua files in the stdlib directory.
// This allows the stdlib modules to be included in the binary
// and accessed without filesystem dependencies.
//
//go:embed *.lua
var StdlibFS embed.FS

// GetEmbeddedModules returns a list of all embedded module names.
// Module names are derived from filenames by removing the .lua extension.
func GetEmbeddedModules() ([]string, error) {
	var modules []string

	err := fs.WalkDir(StdlibFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .lua files
		if filepath.Ext(path) == ".lua" {
			// Remove .lua extension to get module name
			moduleName := strings.TrimSuffix(filepath.Base(path), ".lua")
			modules = append(modules, moduleName)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil
}

// ReadEmbeddedModule reads the content of an embedded module by name.
// The name should not include the .lua extension.
func ReadEmbeddedModule(moduleName string) ([]byte, error) {
	filename := moduleName + ".lua"
	return StdlibFS.ReadFile(filename)
}

// ModuleExists checks if a module exists in the embedded filesystem.
func ModuleExists(moduleName string) bool {
	filename := moduleName + ".lua"
	_, err := StdlibFS.ReadFile(filename)
	return err == nil
}
