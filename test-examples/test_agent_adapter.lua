print("Testing agent adapter vs bridge...")

-- Check what's available globally
print("\nGlobal agent_core:", type(agent_core))
print("Global bridges:", type(bridges))

if bridges and bridges.agent_core then
    local bridge = bridges.agent_core
    print("\nChecking bridge methods:")
    print("  run:", type(bridge.run))
    print("  runAgent:", type(bridge.runAgent))
    print("  lifecycleCreate:", type(bridge.lifecycleCreate))
    print("  createAgent:", type(bridge.createAgent))
end

-- Check if there's an agent_core global (adapter)
if agent_core then
    print("\nChecking agent_core adapter methods:")
    print("  run:", type(agent_core.run))
    print("  lifecycleCreate:", type(agent_core.lifecycleCreate))
end