# Test Fixes Documentation

This document tracks all the test fixes applied after the ScriptValue refactoring.

## Helper Function Added

Added `extractScriptValue` helper function to handle ScriptValue to primitive conversion in tests.

## Files Modified

### 1. `/pkg/engine/gopherlua/bridge_adapter_test.go`
- Fixed `TestBridgeAdapter_MethodWrapping` - multiReturn handling
- Fixed `TestBridgeAdapter_TypeConversion` - complex types handling
- Fixed `TestBridgeAdapter_MethodValidation` - validation error handling
- Added `convertMapToScriptValue` helper function
- Modified `mockBridge.ExecuteMethod` to handle slices and maps properly
- Modified `mockBridge.ValidateMethod` to use validateFunc when available

### 2. `/pkg/engine/gopherlua/converter_scriptvalue_test.go`
- Fixed `TestCircularReferenceDetection` - made error message check more flexible

### 3. `/pkg/engine/gopherlua/engine_execute_test.go`
- Fixed `TestExecutionPipeline_ParameterInjection` - all subtests
- Fixed `TestExecutionPipeline_ResultExtraction` - all subtests (using helper functions)
- Fixed `TestExecutionPipeline_ChunkCacheIntegration` - replaced assert.Equal with assertScriptValueEquals
- Fixed `TestExecutionPipeline_MemoryManagement` - replaced assert.Equal with assertScriptValueEquals

### 4. `/pkg/engine/gopherlua/engine_test.go`
- Fixed `TestLuaEngine_BasicExecution` - all subtests

### 5. `/pkg/engine/gopherlua/engine_integration_test.go`
- Fixed `TestLuaEngine_FullIntegration/concurrent_execution_stability` - using extractScriptValueMap
- Fixed `TestLuaEngine_FullIntegration/error_handling_and_recovery` - using assertScriptValueEquals
- Fixed `TestLuaEngine_FullIntegration/resource_management` - using assertScriptValueEquals

### 6. `/pkg/engine/gopherlua/test_helpers.go` (NEW FILE)
- Created helper functions for ScriptValue extraction and assertions:
  - `extractScriptValue` - extracts Go value from ScriptValue
  - `assertScriptValueEquals` - asserts ScriptValue equals expected value
  - `assertScriptValueNil` - asserts ScriptValue is nil
  - `extractScriptValueMap` - extracts map from ScriptValue
  - `extractScriptValueSlice` - extracts slice from ScriptValue (handles Lua 1-based indexing)

### 7. `/pkg/bridge/util/auth.go`
- Fixed deadlock in `ExecuteMethod` - removed RLock during method execution to prevent RWMutex upgrade deadlock

### 8. `/pkg/engine/conversion.go`
- Fixed `ConvertToScriptValue` to properly handle channels and functions
- Added specific cases for `chan interface{}` and function types  
- Added reflection-based helpers `isFunction` and `isChannel` for generic detection
- Fixed error handling to use `ErrorValue` instead of `StringValue`

### 9. `/pkg/engine/scriptvalue_test.go`
- Updated test expectations for error case to use `ErrorValue` instead of `StringValue`

### 10. `/pkg/bridge/util/json.go`
- Fixed `marshalIndent` method to handle custom prefixes using standard library
- json-iterator doesn't support prefixes, so fallback to `encoding/json` when prefix is non-empty

### 11. **Test Helper Refactoring**
- Moved test helpers from `/pkg/engine/gopherlua/test_helpers.go` to `/pkg/testutils/scriptvalue_helpers.go`
- Centralized location makes helpers available to all packages
- Exported functions with proper capitalization:
  - `extractScriptValue` → `ExtractScriptValue`
  - `assertScriptValueEquals` → `AssertScriptValueEquals`  
  - `assertScriptValueNil` → `AssertScriptValueNil`
  - `extractScriptValueMap` → `ExtractScriptValueMap`
  - `extractScriptValueSlice` → `ExtractScriptValueSlice`
- Updated imports in:
  - `/pkg/engine/gopherlua/engine_integration_test.go`
  - `/pkg/engine/gopherlua/engine_execute_test.go`

## Pattern of Fixes

Most fixes follow this pattern:

**Before:**
```go
assert.Equal(t, expectedValue, result)
```

**After:**
```go
sv, ok := result.(engine.ScriptValue)
require.True(t, ok)
assert.Equal(t, expectedValue, sv.ToGo())
```

## Special Cases

1. **Nil handling**: ScriptValue returns `engine.NilValue{}` instead of Go `nil`
2. **Array/Table handling**: Lua tables can be converted to either arrays or maps
3. **Error messages**: Some tests check exact error messages that changed with refactoring

## Tests Still Failing

### `/pkg/bridge/agent/workflow_test.go`
1. `TestWorkflowBridge_GetMetadata` - expects different metadata values
2. `TestWorkflowBridge_Methods` - many workflow methods are missing
3. Multiple `TestWorkflowBridge_ExecuteMethod_*` tests - methods not implemented
4. `TestWorkflowBridge_TypeMappings` - missing type mappings

### `/pkg/bridge/util/auth_test.go`
1. `TestUtilAuthBridgeMultiSchemeAuth` - timeout after 30s (deadlock in registerAuthScheme)

## Notes

The workflow bridge tests are failing because many methods were removed during cleanup of unused code. These may need to be re-implemented if the workflow functionality is needed.

The auth bridge test has a deadlock issue that needs investigation.