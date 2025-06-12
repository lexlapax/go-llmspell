# Technical Documentation

This directory contains technical documentation for go-llmspell developers and contributors.

## 📋 Table of Contents

### Core Architecture
- [**Architecture Overview**](architecture.md) - Complete system architecture, bridge-first design, and technical specifications

### Implementation Guides
- [**API Reference**](api-reference.md) *(Coming Soon)* - Complete API documentation for all interfaces
- [**Bridge Development Guide**](bridge-development.md) *(Coming Soon)* - How to create new bridges to go-llms functionality
- [**Engine Development Guide**](engine-development.md) *(Coming Soon)* - How to implement new script engines
- [**Type System Guide**](type-system.md) *(Coming Soon)* - Deep dive into type conversion system

### Testing & Quality
- [**Testing Guide**](testing-guide.md) *(Coming Soon)* - TDD practices, test patterns, and quality standards
- [**Performance Guide**](performance.md) *(Coming Soon)* - Performance optimization techniques and benchmarking
- [**Security Guide**](security.md) *(Coming Soon)* - Security model, sandboxing, and threat mitigation

### Development Workflows
- [**Contribution Guidelines**](contribution-guidelines.md) *(Coming Soon)* - How to contribute to go-llmspell
- [**Release Process**](release-process.md) *(Coming Soon)* - Version management and release procedures
- [**Upstream Contributions**](upstream-contributions.md) *(Coming Soon)* - Guidelines for contributing to go-llms

## 🏗️ Architecture Quick Reference

```
go-llmspell/
├── pkg/
│   ├── engine/          # Script engine interfaces and implementations
│   │   ├── interface.go # Core interfaces (ScriptEngine, Bridge, TypeConverter)
│   │   ├── registry.go  # Engine registry and discovery
│   │   ├── types.go     # Type system and conversions
│   │   ├── lua/        # Lua engine implementation
│   │   ├── javascript/ # JavaScript engine implementation
│   │   └── tengo/      # Tengo engine implementation
│   │
│   └── bridge/         # Bridges to go-llms functionality
│       ├── manager.go       # Bridge lifecycle management
│       ├── llm_agent.go     # LLM agent orchestration bridge
│       ├── state.go         # State management bridge
│       ├── workflow.go      # Workflow engine bridge
│       ├── tools.go         # Tool system bridge
│       └── ... (other bridges)
```

## 🎯 Key Principles

1. **Bridge, Don't Build** - Leverage go-llms functionality rather than reimplementing
2. **Engine-Agnostic** - All features work identically across scripting languages
3. **Type Safety** - Maintain conversions at bridge boundaries
4. **Security First** - Sandboxed execution with resource limits
5. **Upstream First** - Contribute core improvements to go-llms

## 🔗 Related Documentation

- [**User Guide**](../user-guide/) - Documentation for spell writers and end users
- [**Main README**](../../README.md) - Project overview and quick start
- [**TODO**](../../TODO.md) - Current implementation tasks
- [**TODO-DONE**](../../TODO-DONE.md) - Completed work tracking

## 📊 Visual Resources

- [**Architecture Diagrams**](../images/) - SVG diagrams showing system design
  - `architecture-overview.svg` - High-level system architecture
  - `engine-architecture.svg` - Script engine component details
  - `bridge-architecture.svg` - Bridge layer visualization

## 🚀 Getting Started (Developers)

1. **Read the Architecture** - Start with [architecture.md](architecture.md)
2. **Set up Development Environment** - Follow main [README](../../README.md)
3. **Check Current Tasks** - Review [TODO.md](../../TODO.md)
4. **Run Tests** - `make test` to verify setup
5. **Make Changes** - Follow TDD approach
6. **Quality Checks** - `make all` before committing

## 📝 Documentation Standards

When contributing technical documentation:

- **Clear Structure** - Use consistent headings and organization
- **Code Examples** - Include practical examples for all concepts
- **Diagrams** - Add visual aids when helpful (SVG format)
- **Cross-References** - Link to related documentation
- **Version Awareness** - Note any version-specific information
- **Testing Focus** - Emphasize TDD and quality practices

## 🤝 Contributing

This technical documentation follows the same contribution process as code:

1. **Issue First** - Create issue for significant documentation changes
2. **Branch** - Create feature branch for your changes
3. **Write** - Follow documentation standards
4. **Review** - Submit PR for review
5. **Iterate** - Address feedback and merge

For questions about technical documentation, please open an issue or discussion in the main repository.

---

**Back to:** [Project Root](../../README.md) | [All Documentation](../README.md) | [User Guide](../user-guide/README.md)