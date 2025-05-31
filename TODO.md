# go-llmspell Implementation TODO

## Overview

This document tracks the implementation progress of go-llmspell. Tasks are organized by priority and dependency order.

## Project Status

- [x] Initial project structure
- [x] Architecture documentation
- [x] go-llms dependency integration
- [x] Core implementation (Phase 1 complete)
- [x] LLM Bridge Enhancement (Phase 2 complete)
- [x] Lua Script Engine Integration (Phase 3 complete)
- [x] Tool System (Phase 4 complete)
- [ ] Agent System (Phase 5)
- [ ] Workflow System (Phase 6)
- [ ] JavaScript/Tengo engines
- [ ] Testing and examples
- [ ] Release preparation

## Phase 1: Core Infrastructure (Priority: Critical) [COMPLETED - See TODO-DONE.md]

## Phase 2: LLM Bridge Enhancement (Priority: Critical) [COMPLETED - See TODO-DONE.md]

## Phase 3: Lua Engine (Priority: High) [COMPLETED - See TODO-DONE.md]

## Phase 4: Tool System (Priority: High) [COMPLETED - See TODO-DONE.md]

## Pending Items from Completed Phases (Revisit)

### From Phase 3 Implementation
- [ ] Add llm.stream_chat_with_history() for message-based streaming
  - Required for full chat-assistant example
  - Accept array of message objects with role and content
- [ ] Implement safe alternatives to io.read/write for interactive spells
  - Required for interactive chat functionality
  - Options: stdin/stdout bridge, event system, or web interface
- [ ] Fix failing promise async tests
  - promise_async_test.go has 4 failing tests
  - Tests are expecting different behavior than current implementation

### From Phase 4 Implementation
- [ ] Investigate and integrate more built-in tools from go-llms
  - Currently only using web_fetch, calculator, string tools
  - Check what other tools are available in go-llms

## Phase 5: Agent System (Priority: High)

### 5.1 Agent Interface
- [ ] Create `pkg/agents/interface.go`
- [ ] Define agent configuration structure
- [ ] Add agent lifecycle management
- [ ] Create agent context handling

### 5.2 Agent Implementation
- [ ] Implement `pkg/agents/agent.go` using go-llms agents
- [ ] Add tool integration for agents - use existing go-llms tools
- [ ] Create conversation memory management
- [ ] Add agent state persistence

### 5.3 Agent Bridge
- [ ] Create `pkg/bridge/agents.go`
- [ ] Implement agent creation from scripts
- [ ] Add agent execution with streaming
- [ ] Create agent composition patterns

### 5.4 Built-in Agents from go-llms
- [ ] Investigate available agents in go-llms
- [ ] Create agent adapter similar to tool adapter
- [ ] Register built-in agents with configuration options
- [ ] Test agent integration with examples

### 5.5 Pre-built Scripted Agents (ask first)
- [ ] Create research assistant agent
- [ ] Implement code review agent
- [ ] Add writing assistant agent
- [ ] Create data analysis agent

## Phase 6: Workflow System (Priority: High)

### 6.1 Workflow Engine
- [ ] Create `pkg/workflows/engine.go`
- [ ] Implement step execution logic
- [ ] Add conditional branching
- [ ] Create parallel execution support

### 6.2 Workflow Bridge
- [ ] Create `pkg/bridge/workflows.go`
- [ ] Implement workflow creation from scripts
- [ ] Add workflow composition (chain, parallel)
- [ ] Create workflow debugging support

### 6.3 Workflow Patterns
- [ ] Implement sequential workflow
- [ ] Add parallel workflow with result aggregation
- [ ] Create conditional workflow with branching
- [ ] Add loop/iteration support


## Phase 7: Spell System (Priority: Medium)

### 7.1 Spell Loader
- [ ] Create `pkg/spells/loader.go`
- [ ] Implement spell discovery from directories
- [ ] Add spell metadata parsing
- [ ] Create spell dependency resolution

### 7.2 Spell Runner
- [ ] Implement `pkg/spells/runner.go`
- [ ] Add parameter validation
- [ ] Create execution isolation
- [ ] Add result formatting

### 7.3 Spell Management
- [ ] Create spell packaging format
- [ ] Add spell versioning support
- [ ] Implement spell dependency management
- [ ] Create spell testing framework

## Phase 8: JavaScript Engine (Priority: Medium)

### 8.1 Goja Integration
- [ ] Create `pkg/engine/javascript/engine.go`
- [ ] Add goja dependency
- [ ] Implement ES6 module support
- [ ] Add promise handling

### 8.2 JavaScript Bridges
- [ ] Create JavaScript bridge adapters
- [ ] Implement async/await support
- [ ] Add event loop integration
- [ ] Create JavaScript-specific utilities

### 8.3 JavaScript Standard Library
- [ ] Port stdlib bridges to JavaScript
- [ ] Add fetch API support
- [ ] Create console object implementation
- [ ] Add timer functions

## Phase 9: Tengo Engine (Priority: Low)

### 9.1 Tengo Integration
- [ ] Create `pkg/engine/tengo/engine.go`
- [ ] Add tengo dependency
- [ ] Implement script compilation
- [ ] Add built-in function registration

### 9.2 Tengo Bridges
- [ ] Create Tengo bridge adapters
- [ ] Handle Tengo's type system
- [ ] Add error propagation
- [ ] Create Tengo-specific helpers

## Phase 10: Security Implementation (Priority: High)

### 10.1 Security Policies
- [ ] Create `pkg/security/policy.go`
- [ ] Implement filesystem sandboxing
- [ ] Add network access control
- [ ] Create resource limit enforcement

### 10.2 Sandbox Implementation
- [ ] Implement `pkg/security/sandbox.go`
- [ ] Add path validation and jailing
- [ ] Create domain allowlisting
- [ ] Add rate limiting

### 10.3 Resource Monitoring
- [ ] Add memory usage tracking
- [ ] Implement CPU time limits
- [ ] Create goroutine limits
- [ ] Add metrics collection

## Phase 11: CLI and User Interface (Priority: Medium)

### 11.1 CLI Commands
- [ ] Enhance `cmd/llmspell/main.go`
- [ ] Add `spell run` command
- [ ] Create `spell list` command
- [ ] Implement `spell create` wizard
- [ ] Add `spell test` command

### 11.2 Configuration
- [ ] Implement config file loading
- [ ] Add environment variable support
- [ ] Create config validation
- [ ] Add config generation command

### 11.3 Output Formatting
- [ ] Add JSON output support
- [ ] Create pretty printing for results
- [ ] Add progress indicators
- [ ] Implement verbose/debug modes

## Phase 12: Testing (Priority: Critical)

### 12.1 Unit Tests
- [ ] Engine interface tests
- [ ] Registry tests
- [ ] Bridge tests
- [ ] Security policy tests

### 12.2 Integration Tests
- [ ] Lua engine integration tests
- [ ] Tool system tests
- [ ] Agent system tests
- [ ] End-to-end spell execution tests

### 12.3 Example Spells
- [ ] Create basic example spells
- [ ] Add advanced spell examples
- [ ] Create tutorial spells
- [ ] Add benchmark spells

### 12.4 Performance Tests
- [ ] Add execution benchmarks
- [ ] Create memory usage tests
- [ ] Add concurrency stress tests
- [ ] Create resource limit tests

## Phase 13: Documentation (Priority: High)

### 13.1 API Documentation
- [ ] Generate godoc documentation
- [ ] Create bridge API reference
- [ ] Document spell API for each language
- [ ] Add troubleshooting guide

### 13.2 Tutorials
- [ ] Create "Writing Your First Spell" tutorial
- [ ] Add "Creating Custom Tools" guide
- [ ] Write "Building Agents" tutorial
- [ ] Create "Security Best Practices" guide

### 13.3 Examples Documentation
- [ ] Document all example spells
- [ ] Create example index
- [ ] Add usage scenarios
- [ ] Create cookbook recipes

## Phase 14: Release Preparation (Priority: Medium)

### 14.1 Build and Packaging
- [ ] Update Makefile for all targets
- [ ] Create release scripts
- [ ] Add cross-compilation support
- [ ] Create distribution packages

### 14.2 Quality Assurance
- [ ] Run full test suite
- [ ] Perform security audit
- [ ] Check resource usage
- [ ] Validate all examples

### 14.3 Release Artifacts
- [ ] Create GitHub releases
- [ ] Generate changelog
- [ ] Update version numbers
- [ ] Create release notes

## Dependencies

### External Dependencies to Add
- [x] github.com/yuin/gopher-lua (Lua engine)
- [x] github.com/joho/godotenv (Environment file loading)
- [ ] github.com/dop251/goja (JavaScript engine)
- [ ] github.com/d5/tengo (Tengo engine)
- [ ] github.com/google/uuid (UUID generation)
- [x] github.com/stretchr/testify (Testing)

### Internal Dependencies
- [x] github.com/lexlapax/go-llms v0.2.6

## Milestones

### Milestone 1: Core Infrastructure (Week 1-2) [COMPLETED]
- Engine interface and registry
- Basic bridge system
- Security context

### Milestone 2: Lua Support (Week 3-4) [COMPLETED]
- Complete Lua engine with full Engine interface implementation
- LLM bridge fully working in Lua
- Standard library modules (JSON, HTTP, Storage, Log, Promise)
- Security sandbox implemented
- Example spells created and tested

### Milestone 3: Tools and Agents (Week 5-6) [IN PROGRESS]
- Tool system complete (Phase 4) ✅
- Agent system working (Phase 5) - Current focus
- Integration with go-llms tools and agents

### Milestone 4: Multi-language Support (Week 7-8)
- JavaScript engine complete
- Tengo engine complete
- Cross-language testing

### Milestone 5: Production Ready (Week 9-12)
- Security hardening
- Performance optimization
- Documentation complete
- Release preparation

## Notes

- Prioritize Lua engine first as it's the most mature scripting option
- Ensure security is built-in from the start, not added later
- Keep bridges language-agnostic where possible
- Focus on developer experience for spell creators
- Maintain compatibility with go-llms updates

## Contributing

When working on tasks:
1. Create a feature branch
2. Mark task as in-progress
3. Write tests alongside implementation
4. Update documentation
5. Mark task as complete after PR merge