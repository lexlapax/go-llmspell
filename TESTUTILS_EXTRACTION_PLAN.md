# Test Utilities Extraction Plan

## Overview

This document outlines a comprehensive plan to extract common test helper functions and patterns from across the `pkg` directory into a centralized `pkg/testutils` package. This will reduce code duplication, improve maintainability, and provide consistent testing utilities across the project.

## Current State Analysis

### Test File Statistics
- Total test files: 56
- Files with mock implementations: 12+
- Files with repeated patterns: ~40

### Common Patterns Identified

1. **Mock Engine Implementations** - Found in 12+ files
2. **Bridge Setup/Teardown Helpers** - Found in ~20 files
3. **ScriptValue Creation Patterns** - Found in ~30 files
4. **Type Assertion Helpers** - Found in ~25 files
5. **Error Checking Patterns** - Found in ~35 files
6. **Table Test Structures** - Found in ~40 files
7. **Context Creation** - Found in almost all test files

## Proposed Structure

```
pkg/testutils/
├── scriptvalue_helpers.go      # Already exists - ScriptValue extraction/assertion
├── mock_engine.go              # Mock ScriptEngine implementation
├── mock_bridges.go             # Common mock bridge implementations
├── bridge_helpers.go           # Bridge setup/teardown utilities
├── assertions.go               # Common assertion helpers
├── builders.go                 # Test data builders (ScriptValue, configs, etc.)
├── context.go                  # Context creation helpers
├── table_test_helpers.go       # Utilities for table-driven tests
└── numeric.go                  # Numeric conversion helpers
```

## Detailed Extraction Plan

### Phase 1: Core Mock Implementations

#### 1.1 Create `mock_engine.go`
Extract and consolidate mock engine implementations from:
- `/pkg/engine/registry_test.go` (mockRegistryScriptEngine)
- `/pkg/engine/integration_test.go` (mockEngineForIntegration)
- `/pkg/engine/interface_test.go` (mockEngine)
- `/pkg/bridge/manager_test.go` (mockScriptEngine)

**Unified Mock Engine Interface:**
```go
type MockScriptEngine struct {
    mu              sync.RWMutex
    bridges         map[string]engine.Bridge
    initialized     bool
    executeCalls    []ExecuteCall
    executeFunc     func(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error)
    metrics         engine.EngineMetrics
}

type ExecuteCall struct {
    Script string
    Params map[string]interface{}
    Result engine.ScriptValue
    Error  error
}

func NewMockScriptEngine() *MockScriptEngine
func (m *MockScriptEngine) WithExecuteFunc(f func(...) (...)) *MockScriptEngine
func (m *MockScriptEngine) GetExecuteCalls() []ExecuteCall
```

#### 1.2 Create `mock_bridges.go`
Extract common mock bridge patterns from:
- `/pkg/engine/gopherlua/bridge_adapter_test.go` (mockBridge)
- `/pkg/engine/gopherlua/async_bridges_test.go` (mockAsyncBridge)
- `/pkg/bridge/state/manager_test.go` (mockBridge)

**Common Mock Bridge:**
```go
type MockBridge struct {
    ID          string
    Methods     map[string]MethodHandler
    Initialized bool
    CleanedUp   bool
}

type MethodHandler func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error)

func NewMockBridge(id string) *MockBridge
func (b *MockBridge) WithMethod(name string, handler MethodHandler) *MockBridge
```

### Phase 2: Bridge Test Helpers

#### 2.1 Create `bridge_helpers.go`
Extract bridge setup/teardown patterns:

```go
// SetupTestBridge initializes a bridge and returns a cleanup function
func SetupTestBridge(t *testing.T, bridge engine.Bridge) func()

// SetupTestBridgeWithEngine initializes a bridge with a mock engine
func SetupTestBridgeWithEngine(t *testing.T, bridge engine.Bridge) (*MockScriptEngine, func())

// AssertBridgeInitialized verifies bridge initialization
func AssertBridgeInitialized(t *testing.T, bridge engine.Bridge)

// AssertBridgeMethod verifies a bridge method exists and has correct info
func AssertBridgeMethod(t *testing.T, bridge engine.Bridge, methodName string, expectedParams int)
```

### Phase 3: ScriptValue Builders

#### 3.1 Enhance `builders.go`
Create fluent builders for common ScriptValue patterns:

```go
// ScriptValueBuilder provides fluent API for building test ScriptValues
type ScriptValueBuilder struct {
    values []engine.ScriptValue
}

func NewScriptValueBuilder() *ScriptValueBuilder
func (b *ScriptValueBuilder) String(s string) *ScriptValueBuilder
func (b *ScriptValueBuilder) Number(n float64) *ScriptValueBuilder
func (b *ScriptValueBuilder) Bool(v bool) *ScriptValueBuilder
func (b *ScriptValueBuilder) Nil() *ScriptValueBuilder
func (b *ScriptValueBuilder) Object(fields map[string]interface{}) *ScriptValueBuilder
func (b *ScriptValueBuilder) Array(elements ...interface{}) *ScriptValueBuilder
func (b *ScriptValueBuilder) Build() []engine.ScriptValue

// Quick creators for common patterns
func StringValue(s string) engine.ScriptValue
func NumberValue(n float64) engine.ScriptValue
func ObjectFromMap(m map[string]interface{}) engine.ScriptValue
func ArrayFromSlice(s []interface{}) engine.ScriptValue
```

### Phase 4: Assertion Helpers

#### 4.1 Create `assertions.go`
Consolidate type assertion and error checking patterns:

```go
// AssertScriptValueType checks if result is of expected type
func AssertScriptValueType(t *testing.T, result engine.ScriptValue, expectedType engine.ValueType)

// AssertErrorValue checks if result is an ErrorValue with expected message
func AssertErrorValue(t *testing.T, result engine.ScriptValue, expectedMessage string)

// AssertObjectHasFields verifies an ObjectValue has expected fields
func AssertObjectHasFields(t *testing.T, result engine.ScriptValue, expectedFields ...string)

// AssertArrayLength verifies an ArrayValue has expected length
func AssertArrayLength(t *testing.T, result engine.ScriptValue, expectedLength int)

// RequireNoGoError asserts no Go error (for methods that should return ErrorValue)
func RequireNoGoError(t *testing.T, err error, msg string)
```

### Phase 5: Table Test Helpers

#### 5.1 Create `table_test_helpers.go`
Provide utilities for common table-driven test patterns:

```go
// MethodTestCase represents a common bridge method test case
type MethodTestCase struct {
    Name        string
    Method      string
    Args        []engine.ScriptValue
    Setup       func(t *testing.T)
    ExpectError bool
    ExpectType  engine.ValueType
    Validate    func(t *testing.T, result engine.ScriptValue)
}

// RunMethodTests executes a table of method test cases
func RunMethodTests(t *testing.T, bridge engine.Bridge, tests []MethodTestCase)

// ValidationTestCase for ValidateMethod tests
type ValidationTestCase struct {
    Name        string
    Method      string
    Args        []engine.ScriptValue
    ExpectError bool
}

// RunValidationTests executes validation test cases
func RunValidationTests(t *testing.T, bridge engine.Bridge, tests []ValidationTestCase)
```

### Phase 6: Context and Numeric Helpers

#### 6.1 Create `context.go`
Simple context creation helpers:

```go
// TestContext creates a context with test-specific values
func TestContext(t *testing.T) context.Context

// TestContextWithTimeout creates a context with timeout
func TestContextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc)

// TestContextWithCancel creates a cancellable context
func TestContextWithCancel(t *testing.T) (context.Context, context.CancelFunc)
```

#### 6.2 Create `numeric.go`
Extract the common toFloat64 helper:

```go
// ToFloat64 converts various numeric types to float64
func ToFloat64(v interface{}) (float64, error)

// MustFloat64 converts to float64 or panics
func MustFloat64(v interface{}) float64
```

## Migration Strategy

### Phase 1: Foundation (Week 1)
1. Create `pkg/testutils` directory structure
2. Implement core mock implementations (mock_engine.go, mock_bridges.go)
3. Move existing scriptvalue_helpers.go content
4. Add comprehensive tests for testutils

### Phase 2: Core Helpers (Week 2)
1. Implement bridge_helpers.go
2. Implement builders.go
3. Implement assertions.go
4. Create migration guide documentation

### Phase 3: Progressive Migration (Weeks 3-4)
1. Migrate test files package by package:
   - Start with `/pkg/engine` tests
   - Then `/pkg/bridge` tests
   - Finally `/pkg/engine/gopherlua` tests
2. Update each test file to use new helpers
3. Remove duplicated code

### Phase 4: Advanced Helpers (Week 5)
1. Implement table_test_helpers.go
2. Implement context.go and numeric.go
3. Identify additional patterns during migration
4. Add any missing utilities

### Phase 5: Cleanup and Documentation (Week 6)
1. Remove all duplicated test code
2. Update test documentation
3. Create testutils usage guide
4. Run full test suite to verify

## Expected Benefits

1. **Code Reduction**: Estimated 30-40% reduction in test code
2. **Consistency**: Uniform testing patterns across all packages
3. **Maintainability**: Single source of truth for test utilities
4. **Discoverability**: Developers can easily find and use test helpers
5. **Type Safety**: Strongly typed test builders and assertions
6. **Performance**: Reusable mock implementations reduce setup overhead

## Success Metrics

1. All tests pass after migration
2. Test line count reduced by >30%
3. No duplicated mock implementations
4. All packages use centralized test utilities
5. New test development time reduced

## Risks and Mitigation

1. **Risk**: Breaking existing tests during migration
   - **Mitigation**: Migrate incrementally, run tests after each change

2. **Risk**: Over-abstraction making tests harder to understand
   - **Mitigation**: Keep helpers simple and well-documented

3. **Risk**: Performance regression from centralized utilities
   - **Mitigation**: Benchmark critical test paths before/after

## Next Steps

1. Review and approve this plan
2. Create initial testutils structure
3. Begin Phase 1 implementation
4. Set up tracking for migration progress