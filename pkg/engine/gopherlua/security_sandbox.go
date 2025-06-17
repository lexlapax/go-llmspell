// ABOUTME: SandboxEnforcer implements comprehensive sandbox enforcement for Lua execution
// ABOUTME: Handles ApplySandbox, environment filtering, metatable protection, and escape prevention

package gopherlua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// SandboxEnforcer manages sandbox enforcement for Lua states
type SandboxEnforcer struct {
	level            SecurityLevel
	libraryLoader    *SafeLibraryLoader
	allowedGlobals   map[string]bool
	blockedFunctions map[string]bool
}

// NewSandboxEnforcer creates a new sandbox enforcer
func NewSandboxEnforcer(level SecurityLevel) *SandboxEnforcer {
	enforcer := &SandboxEnforcer{
		level:            level,
		libraryLoader:    NewSafeLibraryLoader(level),
		allowedGlobals:   make(map[string]bool),
		blockedFunctions: make(map[string]bool),
	}

	// Configure allowed globals based on security level
	enforcer.configureAllowedGlobals()

	// Configure blocked functions
	enforcer.configureBlockedFunctions()

	return enforcer
}

// ApplySandbox applies comprehensive sandbox restrictions to a Lua state
func (se *SandboxEnforcer) ApplySandbox(L *lua.LState) error {
	// Step 1: Load libraries with restrictions
	libraries := se.getAllowedLibraries()
	if err := se.libraryLoader.LoadLibraries(L, libraries); err != nil {
		return fmt.Errorf("failed to load libraries: %w", err)
	}

	// Step 2: Remove dangerous functions
	if err := se.libraryLoader.RemoveDangerousFunctions(L); err != nil {
		return fmt.Errorf("failed to remove dangerous functions: %w", err)
	}

	// Step 3: Apply custom safe replacements
	if err := se.libraryLoader.ApplyCustomReplacements(L); err != nil {
		return fmt.Errorf("failed to apply safe replacements: %w", err)
	}

	// Step 4: Filter global environment
	if err := se.filterGlobalEnvironment(L); err != nil {
		return fmt.Errorf("failed to filter globals: %w", err)
	}

	// Step 5: Protect metatables
	if err := se.protectMetatables(L); err != nil {
		return fmt.Errorf("failed to protect metatables: %w", err)
	}

	// Step 6: Install require restrictions
	if err := se.installRequireRestrictions(L); err != nil {
		return fmt.Errorf("failed to install require restrictions: %w", err)
	}

	return nil
}

// filterGlobalEnvironment removes or replaces dangerous global functions
func (se *SandboxEnforcer) filterGlobalEnvironment(L *lua.LState) error {
	// Block dangerous global functions based on security level
	dangerousGlobals := se.getDangerousGlobals()

	for funcName := range dangerousGlobals {
		if se.isGlobalBlocked(funcName) {
			L.SetGlobal(funcName, lua.LNil)
		}
	}

	// Install safe replacements for critical functions
	return se.installSafeGlobalReplacements(L)
}

// protectMetatables protects built-in metatables from modification
func (se *SandboxEnforcer) protectMetatables(L *lua.LState) error {
	// Protect string metatable
	if err := se.protectStringMetatable(L); err != nil {
		return err
	}

	// Block access to debug functions that could manipulate metatables
	se.blockDebugAccess(L)

	return nil
}

// installRequireRestrictions installs restrictions on require function
func (se *SandboxEnforcer) installRequireRestrictions(L *lua.LState) error {
	// Preserve original require function
	originalRequire := L.GetGlobal("require")
	L.SetGlobal("_original_require", originalRequire)

	switch se.level {
	case SecurityLevelStrict:
		// Block require completely
		L.SetGlobal("require", L.NewFunction(se.blockedRequire))

	case SecurityLevelStandard:
		// Allow only whitelisted modules
		L.SetGlobal("require", L.NewFunction(se.restrictedRequire))

	case SecurityLevelMinimal:
		// Allow most modules but block dangerous ones
		L.SetGlobal("require", L.NewFunction(se.filteredRequire))
	}

	return nil
}

// configureAllowedGlobals sets up allowed global functions based on security level
func (se *SandboxEnforcer) configureAllowedGlobals() {
	// Always allowed globals
	basicGlobals := []string{
		"type", "tostring", "tonumber", "print", "pairs", "ipairs",
		"next", "select", "unpack", "pcall", "xpcall", "error",
		"assert", "rawget", "rawset", "rawequal", "rawlen",
		"getmetatable", "setmetatable",
	}

	for _, global := range basicGlobals {
		se.allowedGlobals[global] = true
	}

	// Additional globals based on security level
	switch se.level {
	case SecurityLevelMinimal:
		se.allowedGlobals["dofile"] = false // Will be replaced with safe version
		se.allowedGlobals["loadfile"] = false
		se.allowedGlobals["load"] = true
		se.allowedGlobals["loadstring"] = true
		se.allowedGlobals["require"] = true
		se.allowedGlobals["getfenv"] = true
		se.allowedGlobals["setfenv"] = true

	case SecurityLevelStandard:
		se.allowedGlobals["dofile"] = false
		se.allowedGlobals["loadfile"] = false
		se.allowedGlobals["load"] = false
		se.allowedGlobals["loadstring"] = false
		se.allowedGlobals["require"] = true // Restricted
		se.allowedGlobals["getfenv"] = false
		se.allowedGlobals["setfenv"] = false

	case SecurityLevelStrict:
		se.allowedGlobals["dofile"] = false
		se.allowedGlobals["loadfile"] = false
		se.allowedGlobals["load"] = false
		se.allowedGlobals["loadstring"] = false
		se.allowedGlobals["require"] = false
		se.allowedGlobals["getfenv"] = false
		se.allowedGlobals["setfenv"] = false
	}
}

// configureBlockedFunctions sets up functions to block completely
func (se *SandboxEnforcer) configureBlockedFunctions() {
	// Always blocked
	se.blockedFunctions["debug"] = true
	se.blockedFunctions["collectgarbage"] = true

	// Block based on security level
	switch se.level {
	case SecurityLevelStandard:
		se.blockedFunctions["dofile"] = true
		se.blockedFunctions["loadfile"] = true
		se.blockedFunctions["getfenv"] = true
		se.blockedFunctions["setfenv"] = true

	case SecurityLevelStrict:
		se.blockedFunctions["dofile"] = true
		se.blockedFunctions["loadfile"] = true
		se.blockedFunctions["getfenv"] = true
		se.blockedFunctions["setfenv"] = true
		se.blockedFunctions["load"] = true
		se.blockedFunctions["loadstring"] = true
		se.blockedFunctions["require"] = true
	}
}

// getAllowedLibraries returns libraries allowed for the security level
func (se *SandboxEnforcer) getAllowedLibraries() []string {
	switch se.level {
	case SecurityLevelMinimal:
		return []string{"base", "coroutine", "table", "io", "os", "string", "math", "package"}

	case SecurityLevelStandard:
		return []string{"base", "coroutine", "table", "string", "math", "os", "package"}

	case SecurityLevelStrict:
		return []string{"base", "coroutine", "table", "string", "math", "package"}

	default:
		return []string{"base", "string", "table", "math"}
	}
}

// getDangerousGlobals returns globals that should be filtered
func (se *SandboxEnforcer) getDangerousGlobals() map[string]bool {
	return map[string]bool{
		"dofile":         true,
		"loadfile":       true,
		"load":           true,
		"loadstring":     true,
		"getfenv":        true,
		"setfenv":        true,
		"debug":          true,
		"collectgarbage": true,
	}
}

// isGlobalBlocked checks if a global should be blocked at current security level
func (se *SandboxEnforcer) isGlobalBlocked(name string) bool {
	allowed, exists := se.allowedGlobals[name]
	if !exists {
		// If not explicitly configured, allow basic functions
		return se.blockedFunctions[name]
	}
	return !allowed
}

// installSafeGlobalReplacements installs safe versions of critical functions
func (se *SandboxEnforcer) installSafeGlobalReplacements(L *lua.LState) error {
	// Safe dofile replacement (always blocked)
	L.SetGlobal("dofile", L.NewFunction(func(L *lua.LState) int {
		L.RaiseError("dofile is disabled for security")
		return 0
	}))

	// Safe loadfile replacement (always blocked)
	L.SetGlobal("loadfile", L.NewFunction(func(L *lua.LState) int {
		L.RaiseError("loadfile is disabled for security")
		return 0
	}))

	// Safe collectgarbage replacement
	L.SetGlobal("collectgarbage", L.NewFunction(func(L *lua.LState) int {
		L.RaiseError("collectgarbage is disabled for security")
		return 0
	}))

	return nil
}

// protectStringMetatable protects the string metatable from modification
func (se *SandboxEnforcer) protectStringMetatable(L *lua.LState) error {
	// Get string metatable
	err := L.DoString(`
		local mt = getmetatable("")
		if mt then
			-- Make it read-only by replacing __newindex
			mt.__newindex = function(t, k, v)
				error("string metatable is protected")
			end
		end
	`)
	return err
}

// blockDebugAccess ensures debug library is not accessible
func (se *SandboxEnforcer) blockDebugAccess(L *lua.LState) {
	L.SetGlobal("debug", lua.LNil)
}

// Require restriction functions
func (se *SandboxEnforcer) blockedRequire(L *lua.LState) int {
	L.RaiseError("require is disabled in strict security mode")
	return 0
}

func (se *SandboxEnforcer) restrictedRequire(L *lua.LState) int {
	module := L.CheckString(1)

	// Whitelist of allowed modules in standard mode
	allowedModules := map[string]bool{
		"string":    true,
		"table":     true,
		"math":      true,
		"coroutine": true,
		"package":   true,
	}

	if !allowedModules[module] {
		L.RaiseError("module '%s' is not allowed in standard security mode", module)
		return 0
	}

	// Use original require implementation for allowed modules
	originalRequire := L.GetGlobal("_original_require")
	if originalRequire != lua.LNil {
		L.Push(originalRequire)
		L.Push(lua.LString(module))
		L.Call(1, 1)
		return 1
	}

	L.RaiseError("original require function not available")
	return 0
}

func (se *SandboxEnforcer) filteredRequire(L *lua.LState) int {
	module := L.CheckString(1)

	// Blacklist of dangerous modules in minimal mode
	blockedModules := map[string]bool{
		"socket": true,
		"lfs":    true,
		"ffi":    true,
		"posix":  true,
		"winapi": true,
		"debug":  true,
	}

	if blockedModules[module] {
		L.RaiseError("module '%s' is blocked for security", module)
		return 0
	}

	// Use original require for allowed modules
	originalRequire := L.GetGlobal("_original_require")
	if originalRequire != lua.LNil {
		L.Push(originalRequire)
		L.Push(lua.LString(module))
		L.Call(1, 1)
		return 1
	}

	L.RaiseError("original require function not available")
	return 0
}
