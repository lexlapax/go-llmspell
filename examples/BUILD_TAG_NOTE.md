# Build Tag Notice

All Go files in the examples directory are marked with the build tag `to_be_migrated`.

This means they will be excluded from normal builds until they are updated for the new multi-engine architecture.

## Why?

The examples were written for the old single-engine (Lua-only) implementation. They need to be migrated to work with the new multi-engine architecture that supports Lua, JavaScript, and Tengo.

## How to build examples (when needed)

If you need to build an example file for reference, you can use:

```bash
go build -tags=to_be_migrated examples/integration/lua_integration.go
```

## Migration Status

- [ ] examples/integration/lua_integration.go - Needs update for new engine interfaces
- [ ] All spell examples (*.lua files) - Need to be tested with new bridges

The spell examples themselves (Lua scripts) may work with minimal changes once the Lua engine implementation is complete.