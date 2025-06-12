# Contributing to Go-LLMSpell

Thank you for your interest in contributing to go-llmspell! This document provides guidelines and information for contributors.

## 🎯 Ways to Contribute

### 🐛 **Bug Reports**
- Search existing issues first
- Use the bug report template
- Include clear reproduction steps
- Provide system information and go-llmspell version

### 💡 **Feature Requests**
- Check if the feature already exists or is planned
- Explain the use case and benefits
- Consider if it belongs in go-llmspell vs go-llms upstream

### 📝 **Documentation**
- Fix typos, improve clarity, add examples
- Write user guides for new features
- Update API documentation
- Translate documentation (future)

### 🔧 **Code Contributions**
- Bug fixes and improvements
- New bridge implementations
- New script engine support
- Performance optimizations

### 🪄 **Spell Examples**
- Real-world spell examples
- Tutorial content
- Best practice demonstrations

## 🏗️ Development Setup

### Prerequisites

- **Go 1.21+** - Latest stable version recommended
- **Git** - For version control
- **Make** - For build automation

### Initial Setup

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/go-llmspell.git
cd go-llmspell

# Add upstream remote
git remote add upstream https://github.com/lexlapax/go-llmspell.git

# Initialize submodules
git submodule update --init --recursive

# Install dependencies and build
make deps
make build

# Run tests to verify setup
make test
```

### Development Workflow

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Make your changes following TDD approach
# 1. Write tests first
# 2. Implement feature
# 3. Ensure tests pass

# Run quality checks
make all  # Runs fmt, vet, lint, test, build

# Commit with clear messages
git commit -m "feat: add new bridge for XYZ functionality"

# Push to your fork and create a pull request
git push origin feature/your-feature-name
```

## 📋 Contribution Guidelines

### 🧪 **Test-Driven Development (TDD)**

We follow strict TDD practices:

1. **Write tests first** - Before implementing any feature
2. **Minimal implementation** - Just enough code to make tests pass
3. **Refactor safely** - Improve code while keeping tests green
4. **Comprehensive coverage** - All code paths should be tested

### 📐 **Code Standards**

#### Go Code Standards
- **Formatting**: Use `gofmt` - enforced by `make fmt`
- **Linting**: Must pass `golangci-lint` - checked by `make lint`
- **Documentation**: All exported functions need godoc comments
- **Error Handling**: Always handle errors explicitly, never panic in library code
- **Naming**: Follow Go naming conventions (camelCase, descriptive names)

#### Code Quality Checklist
```bash
make fmt      # Format code
make vet      # Static analysis
make lint     # Lint checks
make test     # Run tests with race detection
make build    # Verify compilation
```

### 🏛️ **Architecture Principles**

#### 1. Bridge-First Design
- **Bridge, don't build** - Use go-llms functionality when possible
- **Thin wrappers** - Minimal code between scripts and go-llms
- **Type safety** - Handle conversions at bridge boundaries
- **Upstream contributions** - Improve go-llms rather than fork functionality

#### 2. Engine-Agnostic Implementation
- **Common interfaces** - All engines implement the same ScriptEngine interface
- **Portable APIs** - Scripts should work across languages with minimal changes
- **Consistent behavior** - Same functionality across all engines

#### 3. Security First
- **Sandboxed execution** - All scripts run with security restrictions
- **Resource limits** - Memory, CPU, and time constraints
- **Permission model** - Granular control over script capabilities

### 📝 **Documentation Requirements**

#### Code Documentation
```go
// ABOUTME: This file implements the LLM bridge for script engines
// ABOUTME: It provides access to go-llms agent functionality through script interfaces

// NewLLMBridge creates a new bridge for LLM functionality.
// It initializes the bridge with the provided configuration and
// sets up connections to the underlying go-llms provider system.
func NewLLMBridge(config LLMBridgeConfig) (*LLMBridge, error) {
    // Implementation...
}
```

#### File Headers
All Go files should start with a 2-line "ABOUTME" comment explaining the file's purpose.

#### Documentation Updates
- Update relevant documentation when adding features
- Include code examples in documentation
- Update API reference for new functions
- Add user guide sections for new capabilities

### 🔄 **Pull Request Process**

#### Before Submitting
1. **Rebase on latest main** - `git rebase upstream/main`
2. **Run all checks** - `make all` must pass
3. **Update documentation** - Include relevant doc updates
4. **Test edge cases** - Verify error handling and boundary conditions

#### PR Requirements
- **Clear title** - Summarize the change in one line
- **Detailed description** - Explain what, why, and how
- **Link issues** - Reference related issues with "Fixes #123"
- **Screenshots/demos** - For UI or behavior changes
- **Breaking changes** - Clearly document any breaking changes

#### Review Process
1. **Automated checks** - CI must pass (tests, linting, building)
2. **Code review** - At least one maintainer approval required
3. **Documentation review** - Ensure docs are updated and accurate
4. **Final testing** - Manual testing of new functionality

### 🎯 **Specific Contribution Areas**

#### 🌉 **Bridge Development**

When adding new bridges to go-llms functionality:

```go
// Example bridge structure
type MyBridge struct {
    // Bridge to existing go-llms component
    component *gollms.MyComponent
}

func (b *MyBridge) Register(engine ScriptEngine) error {
    // Expose go-llms methods to scripts
    engine.RegisterFunction("my.function", b.wrapFunction)
    return nil
}

func (b *MyBridge) wrapFunction(params ScriptParams) (interface{}, error) {
    // Convert script types to Go types
    goParams := b.convertParams(params)
    
    // Call go-llms functionality
    result, err := b.component.DoSomething(goParams)
    if err != nil {
        return nil, b.mapError(err)
    }
    
    // Convert back to script-friendly types
    return b.convertResult(result), nil
}
```

#### 🔧 **Script Engine Development**

When adding support for new scripting languages:

1. **Implement ScriptEngine interface** - Core execution and bridge support
2. **Add type converter** - Handle script ↔ Go type conversions
3. **Security sandbox** - Implement security restrictions
4. **Standard library** - Create language-specific stdlib modules
5. **Documentation** - Add language-specific user guides

#### 📚 **Documentation Contributions**

- **User guides** - Help spell writers learn the system
- **Technical docs** - Help developers understand the architecture
- **Examples** - Working spells demonstrating features
- **API reference** - Complete function and interface documentation

## 🚫 **What We Don't Accept**

### Code Contributions We Decline
- **Duplicate go-llms functionality** - Contribute to go-llms instead
- **Breaking changes without migration** - Must provide upgrade path
- **Unsafe code** - No direct memory manipulation or unsafe operations
- **Code without tests** - All code must have test coverage
- **Large refactors without discussion** - Discuss architecture changes first

### Security Considerations
- **No credential exposure** - Never commit API keys or secrets
- **No unsafe operations** - Don't bypass security sandboxing
- **No privilege escalation** - Scripts should run with minimal permissions

## 🌍 **Community Guidelines**

### 🤝 **Code of Conduct**
- **Be respectful** - Treat all contributors with respect
- **Be inclusive** - Welcome people of all backgrounds and experience levels
- **Be constructive** - Provide helpful feedback and suggestions
- **Be patient** - Remember that everyone is learning

### 💬 **Communication**
- **GitHub Issues** - For bug reports and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Pull Request Reviews** - For code-specific feedback
- **Documentation** - For user guides and technical information

### 🎓 **Learning Resources**
- **[Architecture Overview](docs/technical/architecture.md)** - Understand the system design
- **[User Guide](docs/user-guide/README.md)** - Learn to write spells
- **[TODO.md](TODO.md)** - See current development priorities
- **[Examples](examples/)** - Study working spell implementations

## 🚀 **Getting Help**

### Before Asking for Help
1. **Read the documentation** - Check relevant guides and API docs
2. **Search existing issues** - Your question might already be answered
3. **Check examples** - Look for similar use cases in the examples directory

### Where to Get Help
- **GitHub Discussions** - For general questions and brainstorming
- **GitHub Issues** - For specific bugs or feature requests
- **Documentation** - For guides, tutorials, and API reference
- **Code Comments** - For understanding specific implementation details

### When Reporting Issues
- **Use issue templates** - Helps us understand your problem quickly
- **Provide context** - Include relevant code, error messages, and system info
- **Minimal reproduction** - Create the smallest possible example that shows the issue
- **Search first** - Avoid duplicating existing issues

## 📈 **Recognition**

Contributors are recognized through:

- **Git commit history** - All contributions are permanently recorded
- **Release notes** - Major contributions mentioned in changelog
- **Documentation credits** - Contributors acknowledged in relevant docs
- **Community showcase** - Exemplary contributions highlighted in discussions

## 🔄 **Release Process**

We follow semantic versioning (semver):

- **Major (x.0.0)** - Breaking changes, major new features
- **Minor (0.x.0)** - New features, backwards compatible
- **Patch (0.0.x)** - Bug fixes, small improvements

Contributors can help with:
- **Testing release candidates** - Help verify new versions work correctly
- **Documentation updates** - Ensure release notes and guides are accurate
- **Migration guides** - Help users upgrade between versions

---

## 🎉 **Thank You!**

Thank you for contributing to go-llmspell! Every contribution, whether it's code, documentation, bug reports, or community support, helps make the project better for everyone.

**Happy spell casting!** 🧙‍♂️✨

---

**Quick Links:**
- [Architecture Overview](docs/technical/architecture.md)
- [User Guide](docs/user-guide/README.md)
- [Current TODO](TODO.md)
- [Project README](README.md)