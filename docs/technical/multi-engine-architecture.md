# Multi-Engine Architecture Design for JavaScript and Tengo

## Overview

This document outlines the strategy for adding JavaScript and Tengo engine support to go-llmspell while maintaining the existing lazy bridge loading architecture and ensuring compatibility with Lua.

## Current Architecture Analysis

### Engine Registry System
- **Factory Pattern**: Engines are registered as lightweight factories in `SetupEngineRegistry()`
- **Lazy Loading**: Bridges are only loaded when `GetEngine()` is called for the first time
- **Caching**: Bridge registration is cached per engine instance + security profile combination
- **Security Profiles**: Map to bridge profiles (sandbox→standard, development→development, etc.)

### Bridge Profile System
- **BridgeProfile**: Defines which bridge sets to register (core, llm, utility, agent, observability, state, structured)
- **StandardProfile**: All bridges (current default for sandbox/production)
- **MinimalProfile**: Core + utility only
- **LLMProfile**: Core + LLM + utility + structured
- **DevelopmentProfile**: Core + LLM + utility + observability

## Multi-Engine Strategy

### 1. Engine Factory Registration Strategy

#### Current State (Lua Only)
```go
// In SetupEngineRegistry()
luaFactory := gopherlua.NewLuaEngineFactory()
if err := registry.Register(luaFactory); err != nil {
    return nil, fmt.Errorf("failed to register Lua engine factory: %w", err)
}
```

#### Proposed Multi-Engine Registration
```go
// In SetupEngineRegistry()
engines := []engine.EngineFactory{
    gopherlua.NewLuaEngineFactory(),
    // TODO: Add when implemented
    // javascript.NewJavaScriptEngineFactory(),
    // tengo.NewTengoEngineFactory(),
}

for _, factory := range engines {
    if err := registry.Register(factory); err != nil {
        return nil, fmt.Errorf("failed to register %s engine factory: %w", factory.Name(), err)
    }
}
```

**Benefits**:
- ✅ No breaking changes to existing Lua functionality
- ✅ Engines are registered only when available (missing engines don't break startup)
- ✅ Extensible for future engines
- ✅ Factory-only registration maintains lazy loading

### 2. Engine-Specific Bridge Profile Mapping

#### Current Implementation
Security profiles map directly to bridge profiles:
- `sandbox` → StandardProfile
- `development` → DevelopmentProfile  
- `production` → StandardProfile
- `minimal` → MinimalProfile
- `llm` → LLMProfile

#### Proposed Engine-Aware Bridge Profile Mapping

**Strategy**: Extend `getBridgeProfileForSecurityProfile()` to be engine-aware:

```go
func (m *EngineRegistryManager) getBridgeProfileForSecurityProfile(securityProfile, engineName string) (registry.BridgeProfile, error) {
    // Engine-specific bridge profile mappings
    switch engineName {
    case "lua":
        return m.getLuaBridgeProfile(securityProfile)
    case "javascript":
        return m.getJavaScriptBridgeProfile(securityProfile)
    case "tengo":
        return m.getTengoBridgeProfile(securityProfile)
    default:
        // Fallback to Lua behavior for unknown engines
        return m.getLuaBridgeProfile(securityProfile)
    }
}

func (m *EngineRegistryManager) getLuaBridgeProfile(securityProfile string) (registry.BridgeProfile, error) {
    // Current implementation (unchanged)
    switch securityProfile {
    case "sandbox", "production":
        return registry.StandardProfile, nil
    case "development":
        return registry.DevelopmentProfile, nil
    case "minimal":
        return registry.MinimalProfile, nil
    case "llm":
        return registry.LLMProfile, nil
    default:
        return registry.StandardProfile, nil
    }
}

func (m *EngineRegistryManager) getJavaScriptBridgeProfile(securityProfile string) (registry.BridgeProfile, error) {
    // JavaScript-optimized bridge profiles
    switch securityProfile {
    case "sandbox":
        return registry.LLMProfile, nil // Lighter profile for JS
    case "development":
        return registry.DevelopmentProfile, nil
    case "production":
        return registry.LLMProfile, nil // Production JS focuses on LLM
    case "minimal":
        return registry.MinimalProfile, nil
    case "llm":
        return registry.LLMProfile, nil
    default:
        return registry.LLMProfile, nil // JS default to LLM-focused
    }
}

func (m *EngineRegistryManager) getTengoBridgeProfile(securityProfile string) (registry.BridgeProfile, error) {
    // Tengo-optimized bridge profiles (minimal by default)
    switch securityProfile {
    case "sandbox", "production":
        return registry.MinimalProfile, nil // Tengo focuses on minimal footprint
    case "development":
        return registry.DevelopmentProfile, nil
    case "minimal":
        return registry.MinimalProfile, nil
    case "llm":
        return registry.LLMProfile, nil
    default:
        return registry.MinimalProfile, nil // Tengo default to minimal
    }
}
```

**Rationale**:
- **Lua**: Maintains current behavior (no breaking changes)
- **JavaScript**: Defaults to LLM-focused profile (common use case for JS in AI)
- **Tengo**: Defaults to minimal profile (Tengo's strength is lightweight execution)

### 3. Configuration Structure for Engine-Specific Bridge Profiles

#### Proposed Configuration Extension

Add to `RunnerConfig`:

```go
type RunnerConfig struct {
    // ... existing fields ...
    
    // Engine-specific bridge profile mappings
    EngineBridgeProfiles map[string]map[string]string `json:"engine_bridge_profiles,omitempty"`
}

// Example configuration:
{
    "engine_bridge_profiles": {
        "lua": {
            "sandbox": "standard",
            "development": "development", 
            "production": "standard",
            "minimal": "minimal",
            "llm": "llm"
        },
        "javascript": {
            "sandbox": "llm",
            "development": "development",
            "production": "llm", 
            "minimal": "minimal",
            "llm": "llm"
        },
        "tengo": {
            "sandbox": "minimal",
            "development": "development",
            "production": "minimal",
            "minimal": "minimal", 
            "llm": "llm"
        }
    }
}
```

**Implementation**:
```go
func (m *EngineRegistryManager) getBridgeProfileForSecurityProfile(securityProfile, engineName string) (registry.BridgeProfile, error) {
    // Check if custom mapping exists in config
    if m.config != nil && m.config.EngineBridgeProfiles != nil {
        if engineProfiles, exists := m.config.EngineBridgeProfiles[engineName]; exists {
            if profileName, exists := engineProfiles[securityProfile]; exists {
                return registry.GetProfileByName(profileName)
            }
        }
    }
    
    // Fallback to default engine-specific mappings
    switch engineName {
    case "lua":
        return m.getLuaBridgeProfile(securityProfile)
    // ... etc
    }
}
```

### 4. Backward Compatibility Strategy

#### No Breaking Changes
- ✅ Existing Lua functionality unchanged
- ✅ Current security profile mappings preserved for Lua
- ✅ Configuration is optional (defaults maintain current behavior)
- ✅ Engine registration is additive (existing engines continue to work)

#### Migration Path
1. **Phase 1**: Deploy multi-engine architecture with Lua only (no behavior change)
2. **Phase 2**: Add JavaScript engine factory when ready
3. **Phase 3**: Add Tengo engine factory when ready  
4. **Phase 4**: Users opt into new engines via file extensions or explicit engine selection

## Implementation Plan

### Task 2.4.4.5.2.1: Design engine factory registration strategy ✅
- Documented above strategy
- No breaking changes to Lua
- Extensible registration pattern

### Task 2.4.4.5.2.2: Plan factory registration without breaking existing functionality
- **Strategy**: Additive registration only
- **Testing**: Ensure existing Lua tests pass unchanged
- **Rollback**: New engines can be disabled without affecting Lua

### Task 2.4.4.5.2.3: Design bridge profile mapping per engine type
- **Strategy**: Engine-aware bridge profile selection
- **Default Behavior**: 
  - Lua: Current mapping (no change)
  - JavaScript: LLM-focused (llm profile default)
  - Tengo: Minimal footprint (minimal profile default)

### Task 2.4.4.5.2.4: Create configuration structure
- **Strategy**: Optional engine-specific bridge profile mapping
- **Backward Compatibility**: Defaults preserve current behavior
- **Flexibility**: Per-engine, per-security-profile customization

### Task 2.4.4.5.2.5: Add tests for lazy loading behavior
- Test engine-specific bridge profile selection
- Test that only requested engines load bridges
- Test configuration override behavior
- Test backward compatibility with Lua-only setup

## Success Criteria

1. **No Breaking Changes**: All existing Lua functionality works unchanged
2. **Extensible Architecture**: New engines can be added without modifying core registry logic
3. **Configurable Profiles**: Bridge profiles can be customized per engine and security profile
4. **Lazy Loading Preserved**: Only requested engines load bridges on-demand
5. **Clean Fallbacks**: Missing engines don't break startup, unknown engines use safe defaults

## Future Considerations

### Engine-Specific Features
- Some bridges may be engine-specific (e.g., Lua-only coroutine bridges)
- Bridge factories could be enhanced to support engine compatibility checks
- Engine capabilities could influence bridge selection

### Performance Optimization
- Engine-specific bridge profiles reduce memory footprint
- Lazy loading ensures unused engines consume no resources
- Bridge caching works per engine instance + profile combination

### Configuration Management
- Engine profiles could be stored in external configuration files
- Dynamic profile switching could be supported for development
- Profile validation could catch configuration errors early