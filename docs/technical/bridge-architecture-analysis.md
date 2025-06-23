# Bridge Architecture Analysis

## Overview

This document analyzes the three-layer bridge architecture in go-llmspell and identifies naming inconsistencies that need to be addressed.

## Current Architecture

### Layer 1: Bridge Implementation (`pkg/bridge/*`)

The bridge layer contains the actual implementations. Current bridge IDs:

**LLM Package (`pkg/bridge/llm/`):**
- `"llm"` - Core LLM operations
- `"providers"` - Provider management
- `"pool"` - LLM pool management

**Util Package (`pkg/bridge/util/`):**
- `"slog"` - Structured logging
- `"script_logger"` - Script-specific logging
- `"util_auth"` - Authentication utilities
- `"util_llm"` - LLM utilities
- `"util"` - Core utilities
- `"debug"` - Debug utilities
- `"util_json"` - JSON utilities
- `"util_errors"` - Error utilities

**Agent Package (`pkg/bridge/agent/`):**
- `"hooks"` - Agent hooks
- `"tools_registry"` - Tool registry
- `"tools"` - Tool operations
- `"events"` - Event system
- `"agent"` - Core agent operations
- `"workflow"` - Workflow management

**Observability Package (`pkg/bridge/observability/`):**
- `"metrics"` - Metrics collection
- `"tracing"` - Distributed tracing
- `"guardrails"` - Safety guardrails

**State Package (`pkg/bridge/state/`):**
- `"state_context"` - State context management
- `"state_manager"` - State lifecycle management

**Structured Package (`pkg/bridge/structured/`):**
- `"schema"` - Schema validation

**Root Bridge Package (`pkg/bridge/`):**
- `"modelinfo"` - Model information

### Layer 2: Bridge Adapters (`pkg/engine/gopherlua/adapters/*`)

The adapter layer wraps bridges for use in the Lua engine:

- `agent/` - Agent-related adapters
- `events/` - Event system adapters
- `hooks/` - Hook system adapters
- `llm/` - LLM adapters
- `modelinfo/` - Model info adapters
- `observability/` - Observability adapters
- `state/` - State management adapters
- `structured/` - Schema/structured data adapters
- `tools/` - Tool adapters
- `utils/` - Utility adapters
- `workflow/` - Workflow adapters

### Layer 3: Stdlib Modules (`pkg/engine/gopherlua/stdlib/*.lua`)

The stdlib layer provides Lua-friendly APIs. Current modules and their bridge expectations:

**Working Correctly:**
- `state.lua` → expects `bridges.state_context`, `bridges.state_manager` ✅
- `tools.lua` → expects `bridges.tools` ✅
- `agent.lua` → expects `bridges.agent`, `bridges.workflow` ✅
- `events.lua` → expects `bridges.events` ✅
- `errors.lua` → expects `bridges.util_errors` ✅
- `data.lua` → expects `bridges.util` ✅

**Naming Mismatches:**
- `llm.lua` → expects `bridges.llm_bridge` ❌ (actual: `"llm"`)
- `llm.lua` → expects `bridges.llm_util_bridge` ❌ (actual: `"util_llm"`)
- `logging.lua` → expects `bridges.util_debug` ❌ (actual: `"debug"`)
- `logging.lua` → expects `bridges.util_script_logger` ❌ (actual: `"script_logger"`)
- `logging.lua` → expects `bridges.util_slog` ❌ (actual: `"slog"`)
- `observability.lua` → expects `bridges.slog` ❌ (should be `"util_slog"`)

**Missing Components:**
- `auth.lua` → expects `bridges.security` ❌ (bridge doesn't exist)
- No `structured.lua` module (despite having `"schema"` bridge)

## Inconsistencies Summary

### 1. Naming Convention Inconsistency
- Some util bridges use prefix: `util_auth`, `util_llm`, `util_json`, `util_errors`
- Others don't: `debug`, `slog`, `script_logger`
- No clear pattern for when to use category prefix

### 2. Bridge ID vs Expected Name Mismatches
- Stdlib modules expect different names than bridges provide
- Examples: `llm_bridge` vs `llm`, `util_debug` vs `debug`

### 3. Missing Implementations
- `security` bridge expected by `auth.lua` doesn't exist
- `structured.lua` stdlib module missing despite having schema bridge

### 4. Category Confusion
- `events` bridge is in agent package but used by multiple modules
- `slog` is used by both logging and observability modules

## Proposed Naming Standard

### Convention: `<category>_<function>`

**Categories:**
- `llm_` - Language model operations
- `agent_` - Agent and workflow operations
- `util_` - Utility functions
- `observability_` - Metrics, tracing, monitoring
- `state_` - State management
- `structured_` - Schema and data validation

### Proposed Renaming:

**LLM:**
- `"llm"` → `"llm_core"`
- `"providers"` → `"llm_providers"`
- `"pool"` → `"llm_pool"`
- `"modelinfo"` → `"llm_modelinfo"`

**Util:**
- `"slog"` → `"util_slog"`
- `"script_logger"` → `"util_script_logger"`
- `"debug"` → `"util_debug"`
- `"util"` → `"util_core"`
- Keep: `"util_auth"`, `"util_llm"`, `"util_json"`, `"util_errors"`

**Agent:**
- `"agent"` → `"agent_core"`
- `"tools"` → `"agent_tools"`
- `"tools_registry"` → `"agent_tools_registry"`
- `"events"` → `"agent_events"`
- `"workflow"` → `"agent_workflow"`
- `"hooks"` → `"agent_hooks"`

**Observability:**
- `"metrics"` → `"observability_metrics"`
- `"tracing"` → `"observability_tracing"`
- `"guardrails"` → `"observability_guardrails"`

**State:**
- Keep: `"state_context"`, `"state_manager"` (already consistent)

**Structured:**
- `"schema"` → `"structured_schema"`

## Implementation Strategy

1. **Phase 1**: Update all GetID() methods in bridges
2. **Phase 2**: Update bridge adapter registrations and references
3. **Phase 3**: Update stdlib module bridge references
4. **Phase 4**: Add missing structured.lua module
5. **Phase 5**: Fix security profile propagation
6. **Phase 6**: Comprehensive testing

This standardization will ensure consistency across all layers and make the architecture more maintainable.