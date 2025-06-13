# Go-LLMSpell Documentation

Welcome to the comprehensive go-llmspell documentation. This directory contains all documentation resources organized to serve different audiences and use cases.

## 📚 Documentation Structure

### 👥 [User Guide](user-guide/) 
**For spell writers and end users**
- Complete guides for writing spells in Lua, JavaScript, and Tengo
- Tutorials, examples, and best practices
- API reference for scripting interfaces

### ⚙️ [Technical Documentation](technical/)
**For developers and contributors**
- System architecture and design principles
- Implementation guides and API specifications
- Testing, performance, and security guides

### 🖼️ [Visual Resources](images/)
**Diagrams and visual aids**
- Architecture diagrams (SVG format)
- System flow charts and component relationships
- Visual guides for complex concepts

## 🚀 Quick Start Paths

### 📝 **New to Go-LLMSpell?**
1. Start with [Project README](../README.md) - Project overview and installation
2. Read [User Guide](user-guide/) - Learn to write your first spells
3. Explore [Examples](../examples/) - See working spell examples

### 🔧 **Want to Contribute?**
1. Read [Technical Architecture](technical/architecture.md) - Understand the system design
2. Check [Current Tasks](../TODO.md) - See what needs to be done
3. Review [Completed Work](../TODO-DONE.md) - Understand progress made

### 🏗️ **Building Extensions?**
1. Study [Bridge Development](technical/README.md) - Learn to create bridges
2. Understand [Type System](technical/README.md) - Handle script ↔ Go conversions
3. Follow [Security Guidelines](technical/README.md) - Maintain sandboxing

## 📖 Core Documentation

### Essential Reading
- [**Architecture Overview**](technical/architecture.md) - Complete system design and philosophy
- [**User Guide Index**](user-guide/README.md) - Everything for spell writers
- [**Technical Guide Index**](technical/README.md) - Everything for developers

### Quick References
- [**Main Project README**](../README.md) - Installation, quick start, and overview
- [**Current TODO**](../TODO.md) - Active development tasks
- [**Completed Work**](../TODO-DONE.md) - Development history and achievements

## 🎯 Documentation by Audience

### 🪄 **Spell Writers (Script Users)**
- **Start Here**: [User Guide](user-guide/README.md)
- **Learn Spells**: Writing spells in your preferred language
- **Use APIs**: Built-in functions and capabilities
- **See Examples**: Real-world spell implementations
- **Get Help**: Troubleshooting and community support

### 🔨 **Developers (Go Contributors)**
- **Start Here**: [Technical Documentation](technical/README.md)
- **Understand Architecture**: Bridge-first design principles
- **Implement Features**: Bridge development and engine creation
- **Maintain Quality**: Testing practices and code standards
- **Contribute Upstream**: Guidelines for go-llms contributions

### 🏛️ **System Architects**
- **Start Here**: [Architecture Overview](technical/architecture.md)
- **Study Design**: Bridge-first approach and component relationships
- **Review Decisions**: Architectural choices and trade-offs
- **Plan Extensions**: Integration points and extension mechanisms
- **Assess Security**: Sandboxing model and threat mitigation

## 🔗 Navigation Links

### Project Navigation
- **🏠 [Project Home](../README.md)** - Main project page
- **📋 [Current Tasks](../TODO.md)** - Active development
- **✅ [Completed Work](../TODO-DONE.md)** - Development history
- **🖼️ [Visual Diagrams](images/)** - Architecture diagrams

### Documentation Sections
- **👥 [User Guide](user-guide/README.md)** - For spell writers
- **⚙️ [Technical Docs](technical/README.md)** - For developers
- **🎯 [Examples](../examples/)** - Working code examples

### External Resources
- **📚 [go-llms Documentation](https://github.com/lexlapax/go-llms)** - Underlying LLM library
- **🌐 [GitHub Repository](https://github.com/lexlapax/go-llmspell)** - Source code and issues
- **💬 [Discussions](https://github.com/lexlapax/go-llmspell/discussions)** - Community discussions

## 📊 Current Status (June 2025)

### ✅ **Completed**
- **Architecture Design** - Bridge-first approach documented
- **Phase 1.1** - Script Engine Interface complete
- **Phase 1.2** - Core Bridge Foundation complete
  - State management bridges (Manager, Context, Persistence)
  - Utility bridges (Auth, JSON, LLM, General)
  - Bridge type system with go-llms aliases
- **Documentation** - Comprehensive architecture and guide structure

### 🚧 **In Progress**
- **Phase 1.3** - Core Bridge System (agents, workflows, events, tools)
- **Lua Engine** - First scripting language implementation
- **User Guides** - Spell writing tutorials and examples
- **API Documentation** - Complete scripting interface reference

### 🔮 **Planned**
- **JavaScript Engine** - ES6+ support with async/await
- **Tengo Engine** - High-performance compiled scripts
- **Advanced Examples** - Complex workflow and agent orchestration
- **Community Features** - Spell marketplace and sharing

## 🎯 Key Concepts

### 🌉 **Bridge-First Architecture**
Go-LLMSpell bridges to go-llms functionality rather than reimplementing it. This ensures compatibility, reduces maintenance, and allows us to focus on our core value: scriptability.

### 🪄 **Spells**
Scripts written in Lua, JavaScript, or Tengo that control LLMs and AI agents. Called "spells" because they bring AI capabilities to life through expressive code.

### 🔧 **Engine-Agnostic Design**
All features work identically across scripting languages. Write once in your preferred language, or easily port between languages.

### 🔒 **Secure Execution**
All spells run in sandboxed environments with resource limits, permission controls, and security restrictions.

## 📝 Documentation Standards

Our documentation follows these principles:

- **Clear Structure** - Consistent organization and navigation
- **Practical Examples** - Working code for every concept
- **Multi-Audience** - Appropriate depth for different users
- **Visual Aids** - Diagrams and charts for complex topics
- **Cross-References** - Links between related concepts
- **Version Awareness** - Clear versioning and change tracking

## 🤝 Contributing to Documentation

Documentation improvements are welcome! Please:

1. **Follow Standards** - Match existing style and organization
2. **Include Examples** - Add working code for new concepts
3. **Update Navigation** - Maintain links and indexes
4. **Test Links** - Verify all references work correctly
5. **Submit PRs** - Use the same process as code contributions

For questions about documentation, please open an issue or discussion in the main repository.

---

**Happy Spell Casting!** 🧙‍♂️✨