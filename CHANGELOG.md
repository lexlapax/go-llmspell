# Changelog

All notable changes to the go-llmspell project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Multi-engine architecture design (Lua, JavaScript, Tengo)
- Phase 1.2: Core Agent System (in progress)

## [0.1.0-alpha] - 2025-06-12

### Overview
Initial implementation of the multi-engine architecture foundation for go-llmspell, supporting go-llms v0.3.3.

### Added

#### Phase 1.1: Script Engine Interface (Complete)
- Core interfaces: ScriptEngine, Bridge, TypeConverter with comprehensive test coverage
- Engine Registry with thread-safe registration and discovery system
- Type System with common type representations and conversion utilities
- Bridge Manager with lifecycle management and dependency resolution
- Core LLM Bridge supporting multiple providers with streaming capabilities
- Essential Utilities Bridge providing JSON, environment, and auth utilities
- Model Info Bridge with caching, filtering, and provider management
- Comprehensive test-driven development approach for all components

#### Development Infrastructure
- Complete Makefile with all standard Go development commands
- Linting configuration with golangci-lint
- Test coverage reporting
- Build system for multi-platform support

### Changed
- Migrated from single Lua-only implementation to multi-engine architecture
- Updated to support go-llms v0.3.3 interfaces and capabilities
- Redesigned bridge system for engine-agnostic operation

### Technical Details
- All components follow TDD principles
- 100% test coverage for implemented features
- Thread-safe implementations throughout
- Clean separation between engine-specific and engine-agnostic code

