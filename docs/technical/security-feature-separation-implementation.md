# Security Level & Feature Set Separation Implementation Plan

**Date:** 2025-06-23  
**Author:** Architecture Analysis  
**Status:** DRAFT - Ready for Gold Space Review

## Executive Summary

Complete architectural refactor to separate **Security Levels** (permissions/resources) from **Feature Sets** (go-llmspell functionality). This eliminates the confusing current profiles that mix security and functionality concerns.

## Current Architecture Problems

### 1. Conflated Concerns
- `sandbox` profile: High security + Full features (confusing)
- `development` profile: Moderate security + Missing agent features (broken)
- `production` profile: Same as sandbox (redundant)

### 2. Missing CLI Options
- CLI enum only has 3 profiles but bridge system supports 5
- No way to specify minimal features with trusted security

### 3. Semantic Confusion
- "Sandbox" suggests minimal features but gives full bridges
- "Production" and "sandbox" are identical

## New Architecture: Option A

### Security Levels (Control Engine Permissions & Resources)

**File: `/pkg/security/levels.go`** (rename from profiles.go)

```go
type SecurityLevel string

const (
    SecurityLevelUntrusted   SecurityLevel = "untrusted"   // Maximum restrictions
    SecurityLevelTrusted     SecurityLevel = "trusted"     // Moderate restrictions  
    SecurityLevelPrivileged  SecurityLevel = "privileged"  // Minimal restrictions
)
```

**Untrusted Level:**
- No network, filesystem, environment, exec access
- 64MB memory, 5% CPU, 30s timeout
- Blocks: require(), loadstring, debug, io, os
- For running unknown/user-submitted scripts

**Trusted Level:**
- Network + filesystem access allowed
- 256MB memory, 50% CPU, 300s timeout  
- Allows: require(), io, os (limited), debug
- For development and known scripts

**Privileged Level:**
- All permissions granted
- 1GB memory, 100% CPU, unlimited timeout
- All functions allowed
- For system administration scripts

### Feature Sets (Control Available go-llmspell Functionality)

**File: `/pkg/bridge/registry/feature_sets.go`** (new file)

```go
type FeatureSet string

const (
    FeatureSetMinimal    FeatureSet = "minimal"    // Core + Utility only
    FeatureSetLLM        FeatureSet = "llm"        // + LLM + Structured  
    FeatureSetAgent      FeatureSet = "agent"      // + Agent workflows
    FeatureSetObservable FeatureSet = "observable" // + Observability/debugging
    FeatureSetFull       FeatureSet = "full"       // All bridges
)
```

**Feature Set Bridge Mappings:**
```go
var FeatureSetBridges = map[FeatureSet][]BridgeSet{
    FeatureSetMinimal:    {BridgeSetCore, BridgeSetUtility},
    FeatureSetLLM:        {BridgeSetCore, BridgeSetUtility, BridgeSetLLM, BridgeSetStructured},
    FeatureSetAgent:      {BridgeSetCore, BridgeSetUtility, BridgeSetLLM, BridgeSetStructured, BridgeSetAgent, BridgeSetState},
    FeatureSetObservable: {BridgeSetCore, BridgeSetUtility, BridgeSetLLM, BridgeSetObservability},
    FeatureSetFull:       {BridgeSetCore, BridgeSetUtility, BridgeSetLLM, BridgeSetStructured, BridgeSetAgent, BridgeSetState, BridgeSetObservability},
}
```

### CLI Interface

**File: `/cmd/llmspell/main.go`**

```go
import (
    "github.com/lexlapax/go-llmspell/pkg/security"
    "github.com/lexlapax/go-llmspell/pkg/bridge/registry"
)

type CLI struct {
    // References centralized enum definitions
    SecurityLevel string `help:"Security level" default:"trusted" enum:"untrusted,trusted,privileged"`
    FeatureSet    string `help:"Feature set" default:"full" enum:"minimal,llm,agent,observable,full"`
}

// Validation uses centralized constants
func (c *CLI) Validate() error {
    if !security.IsValidLevel(c.SecurityLevel) { return fmt.Errorf("invalid security level") }
    if !registry.IsValidFeatureSet(c.FeatureSet) { return fmt.Errorf("invalid feature set") }
    return nil
}
```

**REPL Defaults:**
- **Security Level:** `trusted` (REPL needs filesystem access for history)
- **Feature Set:** `full` (complete functionality in interactive mode)

## File Changes Required

### Single-Source Definition Files (3 files maximum)

#### 1. **SECURITY LEVELS**: `/pkg/security/levels.go` (RENAME from profiles.go)
**SINGLE SOURCE for all security level definitions**
```go
type SecurityLevel string

const (
    SecurityLevelUntrusted   SecurityLevel = "untrusted"
    SecurityLevelTrusted     SecurityLevel = "trusted"
    SecurityLevelPrivileged  SecurityLevel = "privileged"
)

func IsValidLevel(level string) bool { /* validation */ }
func GetLevelConfig(level SecurityLevel) *SecurityConfig { /* config */ }
```

#### 2. **FEATURE SETS**: `/pkg/bridge/registry/feature_sets.go` (NEW FILE)
**SINGLE SOURCE for all feature set definitions**
```go
type FeatureSet string

const (
    FeatureSetMinimal    FeatureSet = "minimal"
    FeatureSetLLM        FeatureSet = "llm"
    FeatureSetAgent      FeatureSet = "agent"
    FeatureSetObservable FeatureSet = "observable"
    FeatureSetFull       FeatureSet = "full"
)

func IsValidFeatureSet(fs string) bool { /* validation */ }
func GetBridgeSetsForFeature(fs FeatureSet) []BridgeSet { /* mapping */ }
```

#### 3. **CLI INTEGRATION**: `/cmd/llmspell/commands/common.go`
**SINGLE SOURCE for CLI context helpers (imports #1 and #2)**
```go
import (
    "github.com/lexlapax/go-llmspell/pkg/security"
    "github.com/lexlapax/go-llmspell/pkg/bridge/registry"
)

func GetSecurityLevel(ctx context.Context) security.SecurityLevel { /* helper */ }
func GetFeatureSet(ctx context.Context) registry.FeatureSet { /* helper */ }
```

### Consumer Files (Import Only - No Redefinition)

#### 4. `/pkg/bridge/registry/registry.go`
- **IMPORT** feature_sets.go
- **REMOVE** existing BridgeProfile variables
- **USE** FeatureSet constants (no redefinition)

#### 5. `/pkg/runner/engine_registry.go` 
- **IMPORT** security/levels.go and registry/feature_sets.go
- **REPLACE** `getBridgeProfileForSecurityProfile()` with `getBridgesForFeatureSet()`
- **REMOVE** all hardcoded profile strings

#### 6. `/cmd/llmspell/main.go`
- **IMPORT** security/levels.go and registry/feature_sets.go
- **REPLACE** `Profile string` with dual flags using imported constants
- **USE** validation functions from centralized sources
- **SET** defaults: `SecurityLevel="trusted"`, `FeatureSet="full"`

#### 7. `/cmd/llmspell/commands/run.go`
- **IMPORT** common.go helpers
- **USE** `GetSecurityLevel()` and `GetFeatureSet()` from context
- **REMOVE** all hardcoded profile strings

#### 8. `/cmd/llmspell/commands/repl.go`  
- **IMPORT** common.go helpers
- **USE** same defaults as CLI: `trusted` + `full`
- **SUPPORT** --security-level and --feature-set flags in REPL

#### 9. `/cmd/llmspell/commands/security.go`
- **IMPORT** security/levels.go
- **USE** centralized SecurityLevel constants
- **ADD** feature set management commands

### Engine Integration

#### 10. `/pkg/engine/gopherlua/engine.go`
- **IMPORT** security/levels.go
- **USE** SecurityLevel constants (no redefinition)
- **REMOVE** all hardcoded profile strings

#### 11. `/pkg/engine/gopherlua/security.go`
- **IMPORT** security/levels.go
- **USE** centralized SecurityLevel enum
- **REMOVE** all profile string parsing

#### 12. `/pkg/repl/lua_repl.go`
- **IMPORT** commands/common.go helpers
- **USE** same flags as CLI: `--security-level` and `--feature-set`
- **DEFAULT** to `trusted` + `full` when no flags specified

### Configuration

#### 13. `/pkg/config/config.go`
- **IMPORT** both security/levels.go and registry/feature_sets.go
- **USE** centralized enums in configuration structs
- **REMOVE** old profile settings entirely

### Test Files (34 files need updates)

#### Integration Tests
- `/tests/integration/bridge_system_integration_test.go`
- `/tests/integration/security_profile_test.go`
- All command tests in `/tests/integration/commands/`

#### Unit Tests  
- `/pkg/security/profiles_test.go` → `/pkg/security/levels_test.go`
- `/pkg/bridge/registry/registry_test.go`
- `/pkg/runner/engine_bridge_profiles_test.go`
- All engine tests in `/pkg/engine/gopherlua/`

#### Test Updates Required:
1. Replace profile strings with SecurityLevel + FeatureSet
2. Update test expectations for new bridge loading logic
3. Test dual-flag CLI parsing
4. Test REPL default behavior (`trusted` + `full`)

### Documentation Files

#### 14. `/docs/technical/security-profile-bridge-mapping-analysis.md`
- **ARCHIVE** (keep for historical reference)
- **CREATE** new architecture documentation

#### 15. `/pkg/docs/manpage_llmspell.go`
- **UPDATE** CLI documentation for new flags

## Implementation Sequence

### Phase 1: Core Architecture 
1. Create `/pkg/bridge/registry/feature_sets.go`
2. Create `/pkg/security/levels.go` alongside existing profiles.go
3. Add dual-flag parsing logic to CLI
4. Update engine registry to support SecurityLevel + FeatureSet

### Phase 2: Update Consumers  
1. Update all command files to use new context keys
2. Update REPL to use default `trusted` + `full`
3. Update engine integration files
4. Update configuration handling

### Phase 3: Remove Old System
1. Delete `/pkg/security/profiles.go`
2. Remove old profile support from engine registry
3. Update all tests to use SecurityLevel + FeatureSet

### Phase 4: Documentation
1. Update all documentation
2. Update examples and tutorials
3. Create usage documentation

## REPL Behavior

**REPL uses same dual-flag system as CLI:**

**Default Flags:**
- **--security-level:** `trusted` (REPL needs filesystem access for command history, config files)
- **--feature-set:** `full` (Interactive mode benefits from complete functionality)
- **Bridge Access:** All bridges available (Core + Utility + LLM + Structured + Agent + State + Observability)

**REPL Flag Override Examples:**
```bash
# Default REPL
llmspell repl  # Uses trusted + full

# Restricted REPL
llmspell repl --security-level=untrusted --feature-set=minimal

# Agent-focused REPL
llmspell repl --security-level=privileged --feature-set=agent

# LLM-only REPL
llmspell repl --security-level=trusted --feature-set=llm
```

## Migration Strategy

1. **Breaking Change:** Remove --profile flag entirely
2. **Replace With:** --security-level and --feature-set flags
3. **Update Defaults:** trusted + full for most permissive experience

## CLI Usage Examples

**Basic Usage:**
```bash
# Use defaults (trusted + full)
llmspell run script.lua

# Restrict security for untrusted scripts
llmspell run untrusted.lua --security-level=untrusted

# Use minimal features for lightweight execution
llmspell run simple.lua --feature-set=minimal

# Maximum restrictions
llmspell run unknown.lua --security-level=untrusted --feature-set=minimal

# Maximum permissions for system scripts
llmspell run admin.lua --security-level=privileged --feature-set=full
```

**REPL Usage:**
```bash
# Default REPL (trusted + full)
llmspell repl

# Restricted REPL 
llmspell repl --security-level=untrusted --feature-set=llm
```

## Implementation Effort

- **Files to Change:** ~50 files
- **New Files:** ~5 files  
- **Test Updates:** ~30 test files
- **Estimated Effort:** 2-3 days for core changes, 1-2 days for tests
- **Breaking Changes:** Yes - complete replacement of --profile system

## Single-Source Architecture Principle

**CRITICAL:** Enum definitions exist in exactly 3 files:

1. **`/pkg/security/levels.go`** - All SecurityLevel constants
2. **`/pkg/bridge/registry/feature_sets.go`** - All FeatureSet constants  
3. **`/cmd/llmspell/commands/common.go`** - CLI integration helpers

**ALL OTHER FILES** must import these 3 files and use their constants. **NO REDEFINITION** of enums anywhere else.

**Enforcement:**
- CLI validation functions reference centralized enums
- Engine registry imports and uses centralized constants
- Tests import and use centralized constants
- Configuration imports and uses centralized constants

This prevents enum sprawl and ensures single-source-of-truth for all definitions.

This architecture cleanly separates security concerns from functionality concerns, making the system much more intuitive and flexible.