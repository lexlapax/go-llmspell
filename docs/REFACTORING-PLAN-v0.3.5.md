# Go-LLMSpell Refactoring Plan for go-llms v0.3.5

## Overview

This document outlines comprehensive refactoring opportunities to leverage go-llms v0.3.5's new capabilities. The plan focuses on maximizing the use of new features while simplifying our codebase.

## Key v0.3.5 Features to Leverage

1. **Schema System** - Repositories, generators, validation, bridge integration
2. **Structured Output Support** - Parsers with recovery, format conversion
3. **Event System Enhancements** - Bridge events, serialization, storage, replay
4. **Workflow Serialization** - JSON export/import, script handlers
5. **Testing Infrastructure** - Centralized mocks, fixtures, scenarios
6. **Documentation Generation** - OpenAPI, Markdown, JSON formats
7. **Bridge-Friendly Type System** - Serializable errors, conversion registry
8. **Runtime Tool Registration** - Enhanced discovery API

## Component-by-Component Refactoring Plan

### 1. pkg/bridge/interfaces.go

**Current State**: Type aliases for go-llms types

**Refactoring Opportunities**:
- ✅ Already updated with v0.3.5 types
- ADD: New event types for bridge communication
- ADD: Schema repository and generator types
- ADD: Structured output parser types
- ADD: Documentation generator interfaces
- ADD: Error serialization types

**Action Items**:
```go
// Add new v0.3.5 types
type (
    // Schema system
    SchemaRepository = schemaDomain.SchemaRepository
    SchemaGenerator  = schemaDomain.SchemaGenerator
    SchemaVersion    = repository.SchemaVersion
    
    // Structured output
    OutputParser     = structured.OutputParser
    JSONParser       = structured.JSONParser
    XMLParser        = structured.XMLParser
    YAMLParser       = structured.YAMLParser
    
    // Event system
    EventStore       = events.EventStore
    EventFilter      = events.Filter
    EventReplayer    = events.Replayer
    EventSerializer  = events.Serializer
    
    // Documentation
    DocGenerator     = docs.Generator
    OpenAPIGenerator = docs.OpenAPIGenerator
    
    // Errors
    SerializableError = errors.SerializableError
    ErrorRecovery     = errors.RecoveryStrategy
)
```

### 2. pkg/bridge/manager.go

**Current State**: Basic lifecycle management

**Refactoring Opportunities**:
- ADD: Event emission for bridge lifecycle (using v0.3.5 event system)
- ADD: Bridge metadata export for documentation generation
- ADD: Serialization support for bridge state
- ADD: Error recovery strategies for initialization failures
- ADD: Performance monitoring using event metrics

**Action Items**:
```go
// Add event emitter
type BridgeManager struct {
    // ... existing fields ...
    eventEmitter *events.EventEmitter
    eventStore   events.EventStore
    metrics      *metrics.BridgeMetrics
}

// Emit lifecycle events
func (m *BridgeManager) InitializeBridge(ctx context.Context, bridgeID string) error {
    m.eventEmitter.Emit(events.BridgeEvent{
        Type: events.BridgeInitializing,
        BridgeID: bridgeID,
        Timestamp: time.Now(),
    })
    // ... existing code ...
}

// Add documentation export
func (m *BridgeManager) GenerateDocumentation(format string) ([]byte, error) {
    generator := docs.NewGenerator(format)
    return generator.GenerateBridgeAPI(m.bridges)
}

// Add serialization
func (m *BridgeManager) ExportState() (*BridgeManagerState, error) {
    // Leverage v0.3.5 serialization
}
```

### 3. pkg/bridge/state/* (context.go, manager.go)

**Current State**: Basic state management wrappers

**Refactoring Opportunities**:
- ADD: State persistence using v0.3.5 schema repositories
- ADD: State validation using schema system
- ADD: State event emission with replay capability
- ADD: State transformation pipeline integration
- ADD: Versioned state snapshots

**Action Items**:
```go
// StateContextBridge enhancements
type StateContextBridge struct {
    // ... existing fields ...
    schemaRepo   schemaDomain.SchemaRepository
    stateSchema  *schemaDomain.Schema
    eventEmitter *events.EventEmitter
    validator    schemaDomain.Validator
}

// Add schema validation
func (b *StateContextBridge) ValidateState(state map[string]interface{}) error {
    return b.validator.ValidateStruct(b.stateSchema, state)
}

// Add event emission
func (b *StateContextBridge) contextSet(ctx, key, value interface{}) error {
    // ... existing code ...
    b.eventEmitter.Emit(events.StateChangeEvent{
        Key: key,
        OldValue: oldValue,
        NewValue: value,
        Context: ctx.(*ScriptSharedContext).id,
    })
}

// Add persistence
func (b *StateContextBridge) persistState(ctx interface{}) error {
    state := b.contextToState(ctx)
    return b.schemaRepo.Save(ctx.ID, state)
}
```

### 4. pkg/bridge/util/* (auth.go, json.go, llm.go, util.go)

**Current State**: Basic utility wrappers

**Refactoring Opportunities**:

**auth.go**:
- ADD: OAuth2 discovery integration
- ADD: Token validation with schema system
- ADD: Auth event logging
- ADD: Credential serialization

**json.go**:
- REPLACE: Use v0.3.5 structured output parsers
- ADD: JSON schema validation
- ADD: Recovery mechanisms for malformed JSON
- ADD: Format conversion (JSON ↔ YAML ↔ XML)

**llm.go**:
- ADD: Provider capability metadata exposure
- ADD: Model discovery API integration
- ADD: Response parsing with recovery
- ADD: Streaming event emission

**util.go**:
- ADD: Error serialization utilities
- ADD: Type conversion registry integration
- ADD: Performance monitoring helpers

**Action Items**:
```go
// JSONBridge replacement
type JSONBridge struct {
    parser    structured.JSONParser
    validator schemaDomain.Validator
    converter structured.FormatConverter
}

func (b *JSONBridge) parse(data string) (interface{}, error) {
    // Use v0.3.5 parser with recovery
    result, err := b.parser.Parse(data)
    if err != nil {
        // Try recovery strategies
        result, err = b.parser.ParseWithRecovery(data)
    }
    return result, err
}

// LLMBridge enhancements
func (b *LLMBridge) getProviderCapabilities(provider string) (map[string]interface{}, error) {
    metadata := b.registry.GetProviderMetadata(provider)
    return b.serializeCapabilities(metadata)
}
```

### 5. pkg/bridge/llm/llm.go

**Current State**: Basic LLM provider wrapper

**Refactoring Opportunities**:
- ADD: Provider metadata and capability discovery
- ADD: Response validation using schemas
- ADD: Streaming with event emission
- ADD: Error recovery strategies
- ADD: Model-specific configuration schemas

**Action Items**:
```go
type LLMBridge struct {
    // ... existing fields ...
    schemaValidator  schemaDomain.Validator
    responseSchemas  map[string]*schemaDomain.Schema
    eventEmitter     *events.EventEmitter
    errorRecovery    errors.RecoveryStrategy
}

// Add schema-validated generation
func (b *LLMBridge) generateWithSchema(provider, prompt, schemaName string) (interface{}, error) {
    schema := b.responseSchemas[schemaName]
    response, err := b.provider.Generate(prompt, 
        llm.WithResponseSchema(schema),
        llm.WithRecovery(b.errorRecovery),
    )
    
    // Emit generation event
    b.eventEmitter.Emit(events.LLMGenerationEvent{
        Provider: provider,
        Model: b.model,
        Schema: schemaName,
        Success: err == nil,
    })
    
    return response, err
}
```

### 6. pkg/bridge/structured/schema.go

**Current State**: Basic schema validation wrapper

**Refactoring Opportunities**:
- USE: Full v0.3.5 schema repository features
- ADD: Schema versioning and migration
- ADD: Schema generation from examples
- ADD: Tag-based generation support
- ADD: Schema export/import
- ADD: Custom validators

**Action Items**:
```go
type SchemaBridge struct {
    // ... existing fields ...
    fileRepo     *repository.FileSchemaRepository
    tagGenerator *generator.TagSchemaGenerator
    migrations   map[string]SchemaMigration
}

// Add versioned schema support
func (b *SchemaBridge) saveSchemaVersion(id string, schema interface{}, version int) error {
    return b.fileRepo.SaveVersion(id, schema, version)
}

// Add schema generation from struct tags
func (b *SchemaBridge) generateFromTags(structType interface{}) (*schemaDomain.Schema, error) {
    return b.tagGenerator.GenerateSchema(structType)
}

// Add schema migration
func (b *SchemaBridge) migrateSchema(id string, fromVersion, toVersion int) error {
    migration := b.migrations[fmt.Sprintf("%d-%d", fromVersion, toVersion)]
    return migration.Apply(b.repository, id)
}
```

### 7. pkg/bridge/modelinfo.go

**Current State**: Basic model info wrapper

**Refactoring Opportunities**:
- ADD: Real-time model capability discovery
- ADD: Model performance metrics from events
- ADD: Cost calculation with pricing data
- ADD: Model recommendation based on task
- ADD: Export model catalog as OpenAPI

**Action Items**:
```go
type ModelInfoBridge struct {
    // ... existing fields ...
    capabilityIndex map[string][]ModelCapability
    metricsStore    events.EventStore
    pricingData     map[string]PricingInfo
    docGenerator    docs.Generator
}

// Add capability-based discovery
func (b *ModelInfoBridge) findModelsWithCapabilities(capabilities []string) ([]ModelInfo, error) {
    // Use enhanced discovery API
}

// Add performance analytics
func (b *ModelInfoBridge) getModelPerformance(model string, timeRange time.Duration) (*PerformanceReport, error) {
    events := b.metricsStore.Query(
        storage.WithType(EventLLMGeneration),
        storage.WithModel(model),
        storage.WithTimeRange(time.Now().Add(-timeRange), time.Now()),
    )
    return b.analyzePerformance(events)
}
```

### 8. pkg/bridge/agent/* (agent.go, events.go, tools.go, workflow.go)

**Current State**: Basic agent system wrappers

**Refactoring Opportunities**:

**agent.go**:
- ADD: Agent state serialization for persistence
- ADD: Agent replay from event history
- ADD: Agent performance profiling
- ADD: Multi-agent coordination patterns

**events.go**:
- REPLACE: with v0.3.5 enhanced event system
- ADD: Event filtering and aggregation
- ADD: Event replay capabilities
- ADD: Bridge-specific event types

**tools.go**:
- USE: Enhanced tool discovery API
- ADD: Tool schema validation
- ADD: Tool execution event tracking
- ADD: Dynamic tool registration
- ADD: Tool documentation generation

**workflow.go**:
- ADD: Workflow serialization/deserialization
- ADD: Script step handlers
- ADD: Workflow templates
- ADD: Visual workflow export

**Action Items**:
```go
// Enhanced ToolsBridge
type ToolsBridge struct {
    // ... existing fields ...
    schemaValidator schemaDomain.Validator
    eventReplayer   events.Replayer
    docGenerator    docs.ToolDocGenerator
}

// Add tool execution with validation
func (b *ToolsBridge) executeToolValidated(name string, params interface{}) (interface{}, error) {
    tool := b.discovery.GetTool(name)
    schema := tool.GetSchema()
    
    // Validate parameters
    if err := b.schemaValidator.ValidateStruct(schema.Input, params); err != nil {
        return nil, fmt.Errorf("invalid parameters: %w", err)
    }
    
    // Execute with event tracking
    start := time.Now()
    result, err := tool.Execute(params)
    
    b.eventEmitter.Emit(ToolExecutionEvent{
        Tool:     name,
        Duration: time.Since(start),
        Success:  err == nil,
        Error:    err,
    })
    
    return result, err
}

// Enhanced WorkflowBridge  
func (b *WorkflowBridge) exportWorkflow(id string) (string, error) {
    workflow := b.workflows[id]
    return workflow.Serialize() // New v0.3.5 feature
}

func (b *WorkflowBridge) importWorkflow(data string) (string, error) {
    workflow, err := workflow.Deserialize(data)
    if err != nil {
        return "", err
    }
    return b.registerWorkflow(workflow)
}
```

### 9. pkg/engine/* (interface.go, types.go, registry.go)

**Current State**: Core engine interfaces

**Refactoring Opportunities**:
- ADD: Engine-level event bus
- ADD: Type conversion registry integration
- ADD: Performance profiling hooks
- ADD: Serialization interfaces
- ADD: Documentation metadata

**Action Items**:
```go
// Enhanced ScriptEngine interface
type ScriptEngine interface {
    // ... existing methods ...
    
    // New v0.3.5 integration methods
    GetEventBus() events.EventEmitter
    RegisterTypeConverter(from, to reflect.Type, converter TypeConverter) error
    ExportAPI() (*EngineAPISpec, error)
    EnableProfiling(profiler Profiler) error
}

// Enhanced Bridge interface
type Bridge interface {
    // ... existing methods ...
    
    // New serialization support
    Serialize() ([]byte, error)
    Deserialize(data []byte) error
    
    // New documentation support
    GetAPISchema() (*APISchema, error)
    GetExamples() []Example
}
```

## Implementation Priority

### Phase 1: Foundation (Week 1-2)
1. Update interfaces.go with all new v0.3.5 types
2. Enhance BridgeManager with event system
3. Integrate testing infrastructure (mocks, fixtures)

### Phase 2: Core Enhancements (Week 3-4)
1. Upgrade schema.go with full repository/generator support
2. Replace JSON utilities with structured output parsers
3. Enhance state bridges with persistence and validation

### Phase 3: Advanced Features (Week 5-6)
1. Implement workflow serialization
2. Add event replay capabilities
3. Integrate documentation generation
4. Add performance monitoring

### Phase 4: Polish (Week 7-8)
1. Complete error recovery strategies
2. Add comprehensive examples
3. Generate API documentation
4. Performance optimization

## Testing Strategy

1. **Use v0.3.5 Testing Infrastructure**:
   - Replace custom mocks with centralized mocks
   - Use fixture library for test data
   - Implement scenario-based testing

2. **Event-Driven Testing**:
   - Verify event emission
   - Test event replay
   - Validate event sequences

3. **Schema Validation Testing**:
   - Test all bridge methods against schemas
   - Verify type conversions
   - Test error recovery

## Migration Notes

1. **No Backward Compatibility Required** - We can make breaking changes
2. **Incremental Migration** - Can be done component by component
3. **Parallel Development** - Multiple components can be updated simultaneously
4. **Documentation First** - Generate docs as we refactor

## Success Metrics

1. **Code Reduction**: Expect 20-30% less code by using go-llms features
2. **Test Coverage**: Achieve 90%+ coverage using v0.3.5 test infrastructure
3. **Performance**: Event-based monitoring should show <5ms overhead
4. **Documentation**: 100% API coverage with generated docs
5. **Type Safety**: All script-Go interactions validated by schemas

## Next Steps

1. Review and approve this plan
2. Create detailed tickets for each component
3. Set up v0.3.5 test infrastructure
4. Begin with Phase 1 implementation
5. Weekly progress reviews