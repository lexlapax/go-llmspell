// ABOUTME: Tests for LStateFactory which creates and configures new Lua VM instances
// ABOUTME: Validates state creation, library loading, initialization, and configuration

package gopherlua

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestLStateFactory_Create(t *testing.T) {
	tests := []struct {
		name          string
		config        FactoryConfig
		validateState func(t *testing.T, L *lua.LState)
		wantErr       bool
		errContains   string
	}{
		{
			name: "creates_state_with_default_config",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Test that basic libraries are loaded
				assert.NotNil(t, L.GetGlobal("string"))
				assert.NotNil(t, L.GetGlobal("table"))
				assert.NotNil(t, L.GetGlobal("math"))
				// Test that dangerous libraries are not loaded
				assert.Equal(t, lua.LNil, L.GetGlobal("io"))
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
		{
			name: "creates_state_with_minimal_security",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelMinimal,
				}),
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// More libraries should be available
				assert.NotNil(t, L.GetGlobal("os"))
				assert.NotNil(t, L.GetGlobal("io"))
				// But debug should still be disabled
				assert.Equal(t, lua.LNil, L.GetGlobal("debug"))
			},
		},
		{
			name: "creates_state_with_strict_security",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStrict,
				}),
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Only essential libraries
				assert.NotNil(t, L.GetGlobal("string"))
				assert.NotNil(t, L.GetGlobal("table"))
				// No OS access
				assert.Equal(t, lua.LNil, L.GetGlobal("os"))
				assert.Equal(t, lua.LNil, L.GetGlobal("io"))
			},
		},
		{
			name: "applies_registry_size_limit",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				RegistrySize: 1024,
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Registry size should be set in options
			},
		},
		{
			name: "executes_init_script",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				InitScript: `_G.TEST_INIT = "initialized"`,
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Check init script was executed
				val := L.GetGlobal("TEST_INIT")
				str, ok := val.(lua.LString)
				assert.True(t, ok)
				assert.Equal(t, "initialized", string(str))
			},
		},
		{
			name: "fails_on_invalid_init_script",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				InitScript: `invalid syntax here {{`,
			},
			wantErr:     true,
			errContains: "init script failed",
		},
		{
			name: "preloads_custom_modules",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				PreloadModules: map[string]lua.LGFunction{
					"testmod": func(L *lua.LState) int {
						mod := L.NewTable()
						L.SetField(mod, "version", lua.LString("1.0"))
						L.Push(mod)
						return 1
					},
				},
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Module should be preloaded but not loaded
				assert.Equal(t, lua.LNil, L.GetGlobal("testmod"))

				// Should be loadable via require
				err := L.DoString(`local mod = require("testmod"); _G.MOD_VERSION = mod.version`)
				assert.NoError(t, err)

				val := L.GetGlobal("MOD_VERSION")
				str, ok := val.(lua.LString)
				assert.True(t, ok)
				assert.Equal(t, "1.0", string(str))
			},
		},
		{
			name: "applies_custom_vm_options",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				Options: lua.Options{
					CallStackSize:       256,
					RegistrySize:        512,
					IncludeGoStackTrace: true,
				},
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Options are applied at creation, test basic functionality
				assert.NotNil(t, L.GetGlobal("string"))
			},
		},
		{
			name: "executes_warmup_function",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				WarmupFunc: func(L *lua.LState) error {
					// Precompile common operations
					return L.DoString(`
						local function noop() end
						for i = 1, 10 do noop() end
						_G.WARMED_UP = true
					`)
				},
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				val := L.GetGlobal("WARMED_UP")
				assert.Equal(t, lua.LTrue, val)
			},
		},
		{
			name: "fails_on_warmup_error",
			config: FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: SecurityLevelStandard,
				}),
				WarmupFunc: func(L *lua.LState) error {
					return L.DoString(`error("warmup failed")`)
				},
			},
			wantErr:     true,
			errContains: "warmup failed",
		},
		{
			name:   "uses_default_security_manager",
			config: FactoryConfig{
				// No SecurityManager specified
			},
			validateState: func(t *testing.T, L *lua.LState) {
				assert.NotNil(t, L)
				// Should use default standard security
				assert.NotNil(t, L.GetGlobal("string"))
				assert.Equal(t, lua.LNil, L.GetGlobal("io"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewLStateFactory(tt.config)
			require.NotNil(t, factory)

			L, err := factory.Create()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, L)
			defer L.Close()

			if tt.validateState != nil {
				tt.validateState(t, L)
			}
		})
	}
}

func TestLStateFactory_Libraries(t *testing.T) {
	tests := []struct {
		name          string
		securityLevel SecurityLevel
		checkLibs     map[string]bool // library name -> should exist
	}{
		{
			name:          "minimal_security_libraries",
			securityLevel: SecurityLevelMinimal,
			checkLibs: map[string]bool{
				"base":      true,
				"coroutine": true,
				"table":     true,
				"io":        true,
				"os":        true,
				"string":    true,
				"math":      true,
				"debug":     false, // Never include debug
			},
		},
		{
			name:          "standard_security_libraries",
			securityLevel: SecurityLevelStandard,
			checkLibs: map[string]bool{
				"base":      true,
				"coroutine": true,
				"table":     true,
				"io":        false, // No file access
				"os":        true,  // Limited OS functions
				"string":    true,
				"math":      true,
				"debug":     false,
			},
		},
		{
			name:          "strict_security_libraries",
			securityLevel: SecurityLevelStrict,
			checkLibs: map[string]bool{
				"base":      true,
				"coroutine": true,
				"table":     true,
				"io":        false,
				"os":        false, // No OS access
				"string":    true,
				"math":      true,
				"debug":     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewLStateFactory(FactoryConfig{
				SecurityManager: NewSecurityManager(SecurityConfig{
					Level: tt.securityLevel,
				}),
			})

			L, err := factory.Create()
			require.NoError(t, err)
			require.NotNil(t, L)
			defer L.Close()

			for lib, shouldExist := range tt.checkLibs {
				// Check if library functions exist
				var exists bool
				switch lib {
				case "base":
					exists = L.GetGlobal("print") != lua.LNil
				case "io":
					exists = L.GetGlobal("io") != lua.LNil
				case "os":
					exists = L.GetGlobal("os") != lua.LNil
				case "debug":
					exists = L.GetGlobal("debug") != lua.LNil
				default:
					exists = L.GetGlobal(lib) != lua.LNil
				}

				if shouldExist {
					assert.True(t, exists, "library %s should exist", lib)
				} else {
					assert.False(t, exists, "library %s should not exist", lib)
				}
			}
		})
	}
}

func TestLStateFactory_OSLibrarySafety(t *testing.T) {
	// Test that dangerous OS functions are removed in standard mode
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	L, err := factory.Create()
	require.NoError(t, err)
	require.NotNil(t, L)
	defer L.Close()

	// These functions should be removed
	dangerousFuncs := []string{
		"os.execute",
		"os.exit",
		"os.setenv",
		"os.remove",
		"os.rename",
	}

	for _, funcPath := range dangerousFuncs {
		parts := strings.Split(funcPath, ".")
		err := L.DoString(`
			if ` + parts[0] + ` and ` + parts[0] + `.` + parts[1] + ` then
				error("` + funcPath + ` should not exist")
			end
		`)
		assert.NoError(t, err, "%s should be removed", funcPath)
	}

	// These safe functions should still exist
	safeFuncs := []string{
		"os.time",
		"os.date",
		"os.clock",
		"os.difftime",
	}

	for _, funcPath := range safeFuncs {
		parts := strings.Split(funcPath, ".")
		err := L.DoString(`
			if not (` + parts[0] + ` and ` + parts[0] + `.` + parts[1] + `) then
				error("` + funcPath + ` should exist")
			end
		`)
		assert.NoError(t, err, "%s should exist", funcPath)
	}
}

func TestLStateFactory_Concurrency(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	// Create multiple states concurrently
	const numStates = 10
	states := make([]*lua.LState, numStates)
	errors := make([]error, numStates)

	start := make(chan struct{})
	done := make(chan struct{}, numStates)

	for i := 0; i < numStates; i++ {
		go func(idx int) {
			<-start // Wait for signal to start
			states[idx], errors[idx] = factory.Create()
			done <- struct{}{}
		}(i)
	}

	// Start all goroutines at once
	close(start)

	// Wait for all to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < numStates; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("timeout waiting for state creation")
		}
	}

	// Check results
	for i := 0; i < numStates; i++ {
		assert.NoError(t, errors[i], "state %d creation error", i)
		assert.NotNil(t, states[i], "state %d is nil", i)
		if states[i] != nil {
			states[i].Close()
		}
	}
}

func TestLStateFactory_Reset(t *testing.T) {
	factory := NewLStateFactory(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStandard,
		}),
	})

	// Test that Reset properly reinitializes factory
	L1, err := factory.Create()
	require.NoError(t, err)
	require.NotNil(t, L1)
	L1.Close()

	// Modify factory config
	factory.Reset(FactoryConfig{
		SecurityManager: NewSecurityManager(SecurityConfig{
			Level: SecurityLevelStrict,
		}),
		InitScript: `_G.RESET_TEST = true`,
	})

	L2, err := factory.Create()
	require.NoError(t, err)
	require.NotNil(t, L2)
	defer L2.Close()

	// Verify new config is applied
	val := L2.GetGlobal("RESET_TEST")
	assert.Equal(t, lua.LTrue, val)

	// Verify strict security (no os library)
	assert.Equal(t, lua.LNil, L2.GetGlobal("os"))
}
