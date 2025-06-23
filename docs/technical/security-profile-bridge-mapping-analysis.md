# Security Profile to Bridge Profile Mapping Analysis

**Date:** 2025-06-23  
**Status:** ✅ ALREADY IMPLEMENTED AND WORKING  
**Phase:** 5 - Security Profile Mapping Analysis

## Executive Summary

The security profile mapping to bridge profiles is **already fully implemented** in the codebase. This feature provides intelligent bridge loading based on security context and engine type, with comprehensive configuration support and testing.

## Current Implementation Status

### ✅ Fully Implemented Features

1. **Security Profile → Bridge Profile Mapping**
   - ✅ Engine-specific mappings (Lua, JavaScript, Tengo)
   - ✅ Configurable custom mappings via `EngineBridgeProfiles`
   - ✅ Fallback to sensible defaults
   - ✅ Comprehensive test coverage

2. **Bridge Profiles Available**
   - ✅ `StandardProfile`: Full feature set (Core, LLM, Utility, Agent, Observability, State, Structured)
   - ✅ `MinimalProfile`: Lightweight (Core, Utility only)
   - ✅ `LLMProfile`: LLM-focused (Core, LLM, Utility, Structured)
   - ✅ `DevelopmentProfile`: Development/debugging (Core, LLM, Utility, Observability)

3. **Integration Points**
   - ✅ CLI security profiles propagate to executor
   - ✅ Executor maps security profiles to bridge profiles
   - ✅ Bridge registry loads appropriate bridge sets
   - ✅ Engine receives only necessary bridges

## Default Security Profile → Bridge Profile Mappings

### Lua Engine
```
sandbox     → StandardProfile   (Full functionality with strict security)
development → DevelopmentProfile (Full functionality + debugging tools)
production  → StandardProfile   (Full functionality, production ready)
minimal     → MinimalProfile    (Essential bridges only)
llm         → LLMProfile        (LLM-optimized bridge set)
```

### JavaScript Engine  
```
sandbox     → LLMProfile        (Lighter profile for JS performance)
development → DevelopmentProfile (Full debugging support)
production  → LLMProfile        (Production-ready LLM focus)
minimal     → MinimalProfile    (Essential bridges only)
llm         → LLMProfile        (Native fit for JS+LLM)
```

### Tengo Engine
```
sandbox     → MinimalProfile    (Lightweight for Tengo's simplicity)
development → DevelopmentProfile (Full debugging support)
production  → MinimalProfile    (Production-ready minimal footprint)
minimal     → MinimalProfile    (Natural fit for Tengo)
llm         → LLMProfile        (LLM operations when needed)
```

## Configuration Support

### Custom Engine Bridge Profiles

The system supports custom mappings via `RunnerConfig.EngineBridgeProfiles`:

```go
type RunnerConfig struct {
    // Maps engine name -> security profile -> bridge profile name
    EngineBridgeProfiles map[string]map[string]string
}
```

**Example Configuration:**
```go
config := &RunnerConfig{
    EngineBridgeProfiles: map[string]map[string]string{
        "lua": {
            "sandbox":     "minimal",     // Override: use minimal for sandbox
            "development": "llm",         // Override: use LLM-focused for dev
        },
        "javascript": {
            "sandbox":    "standard",     // Override: use full standard for JS sandbox
            "production": "minimal",      // Override: use minimal for JS production
        },
        "custom-engine": {
            "sandbox": "development",     // Custom engine mapping
        },
    },
}
```

## Implementation Details

### 1. Security Profile Propagation Flow

```
CLI --profile=sandbox
    ↓
main.go: createCommandContext(profile)
    ↓
run.go: GetProfile(ctx) → passes to ExecuteWithOptions
    ↓
executor.go: options.SecurityProfile → getBridgeProfileForSecurityProfile
    ↓
engine_registry.go: maps securityProfile + engineName → BridgeProfile
    ↓
bridge registry: loads appropriate bridge sets
```

### 2. Bridge Profile Selection Logic

```go
func (m *EngineRegistryManager) getBridgeProfileForSecurityProfile(securityProfile, engineName string) (registry.BridgeProfile, error) {
    // 1. Check custom configuration first
    if m.config != nil && m.config.EngineBridgeProfiles != nil {
        if engineProfiles, exists := m.config.EngineBridgeProfiles[engineName]; exists {
            if profileName, exists := engineProfiles[securityProfile]; exists {
                return m.getProfileByName(profileName)
            }
        }
    }
    
    // 2. Fallback to engine-specific defaults
    switch engineName {
    case "lua":
        return m.getLuaBridgeProfile(securityProfile)
    case "javascript":
        return m.getJavaScriptBridgeProfile(securityProfile)
    case "tengo":
        return m.getTengoBridgeProfile(securityProfile)
    default:
        return m.getLuaBridgeProfile(securityProfile) // Fallback to Lua behavior
    }
}
```

## Test Coverage

### ✅ Comprehensive Test Suite

Located in `pkg/runner/engine_bridge_profiles_test.go`:

1. **`TestEngineSpecificBridgeProfiles`**
   - Tests default mappings for all engines
   - Verifies correct profile selection per security level
   - Tests unknown security profiles fallback

2. **`TestCustomEngineBridgeProfiles`**
   - Tests custom configuration override
   - Tests invalid profile name handling
   - Tests nil/empty configuration behavior

3. **`TestLazyBridgeLoadingWithProfiles`**
   - Tests integration with lazy loading
   - Verifies bridge caching works with profiles

4. **`TestGetProfileByName`**
   - Tests profile name resolution
   - Tests invalid profile name error handling

## Bridge Sets in Each Profile

### StandardProfile (Lua: sandbox, production)
- ✅ BridgeSetCore (modelinfo)
- ✅ BridgeSetLLM (llm_core, llm_providers, llm_pool)  
- ✅ BridgeSetUtility (all util_* bridges)
- ✅ BridgeSetAgent (all agent_* bridges)
- ✅ BridgeSetObservability (all observability_* bridges)
- ✅ BridgeSetState (state_manager, state_context)
- ✅ BridgeSetStructured (structured_schema)

### DevelopmentProfile (All engines: development)
- ✅ BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetObservability
- ❌ Excludes: Agent workflows, State management (focuses on development tools)

### LLMProfile (JS: sandbox/production, Tengo: llm)
- ✅ BridgeSetCore, BridgeSetLLM, BridgeSetUtility, BridgeSetStructured
- ❌ Excludes: Agent workflows, Observability, State (LLM-focused)

### MinimalProfile (Tengo: sandbox/production)
- ✅ BridgeSetCore, BridgeSetUtility
- ❌ Excludes: Everything else (minimal footprint)

## Security Implications

### 1. Defense in Depth
- **Engine Security**: Each engine has own sandbox mechanisms
- **Bridge Security**: Bridge profiles limit available functionality
- **Profile Security**: Security profiles control both engine settings AND bridge availability

### 2. Principle of Least Privilege
- **Minimal Profile**: Only essential bridges for basic operations
- **LLM Profile**: Only LLM-related functionality
- **Development Profile**: Adds observability but excludes production concerns

### 3. Engine-Appropriate Defaults
- **Lua**: Mature engine → StandardProfile by default
- **JavaScript**: Performance-focused → LLMProfile by default  
- **Tengo**: Simplicity-focused → MinimalProfile by default

## Verification of Current Behavior

### CLI Test Results
```bash
# Sandbox profile - strict security, full bridges (Lua)
./llmspell run script.lua --profile=sandbox --debug
[DEBUG] Executing script: script.lua with profile: sandbox
# → Uses StandardProfile → All bridges available
# → require() blocked by engine security level

# Development profile - standard security, debugging bridges
./llmspell run script.lua --profile=development --debug  
[DEBUG] Executing script: script.lua with profile: development
# → Uses DevelopmentProfile → Core + LLM + Utility + Observability
# → require() allowed by engine security level
```

## Conclusion

### ✅ The Question is ANSWERED

**"Should security profile mapping propagate to bridge registry and bridge profiles?"**

**Answer: IT ALREADY DOES, and it's working perfectly!**

1. **✅ Security profiles are mapped to bridge profiles**
2. **✅ Each engine has appropriate mappings**
3. **✅ Configuration allows customization**
4. **✅ Tests verify the behavior**
5. **✅ CLI integration is working**

### What was the "issue"?

The issue wasn't that bridge profile mapping was missing - it was that CLI security profile propagation to the executor had a bug (which we fixed earlier in this phase). The bridge profile mapping was always working correctly.

### Recommendations

1. **✅ No implementation needed** - system is already complete
2. **✅ No architecture changes needed** - design is sound
3. **✅ Current mappings are sensible** - good defaults for each engine
4. **✅ Configuration is flexible** - allows overrides when needed

### Phase 5 Status: COMPLETE

Both parts of Phase 5 are now complete:
1. ✅ Fix CLI profile propagation to executor (completed)
2. ✅ Analyze bridge profile mapping (analysis shows it's already implemented and working)

The security profile system is robust, well-tested, and provides appropriate bridge loading based on security context and engine capabilities.