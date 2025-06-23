# Coverage Improvement Plan

## Current Coverage Analysis

Based on test results, here are packages that need coverage improvement:

### Priority 1 - Very Low Coverage
- `cmd/llmspell`: 6.9% (needs major improvement)
- `cmd/llmspell/commands`: 45.1% (failing tests, needs fixes)

### Priority 2 - Below Target Coverage
- `pkg/bridge/agent`: 38.2%
- `pkg/engine/gopherlua/stdlib`: 43.8%
- `pkg/bridge/state`: 49.4%
- `pkg/engine/gopherlua/adapters`: 54.1%
- `pkg/bridge/llm`: 60.9%
- `pkg/bridge/util`: 63.7%
- `pkg/engine/gopherlua`: 69.6% (has failing test)

### Priority 3 - Near Target Coverage (80-89%)
- `pkg/bridge/structured`: 82.8%
- `pkg/runner`: 84.0%
- `pkg/template`: 89.5%

### Already Good Coverage (90%+)
- `pkg/errors`: 91.3%
- `pkg/security`: 94.6%

## Strategy

1. **Fix failing tests first**
2. **Create comprehensive test suites for low coverage packages**
3. **Add edge case and error condition tests**
4. **Create integration tests**
5. **Add stress and chaos tests**

## Target: 90%+ overall coverage