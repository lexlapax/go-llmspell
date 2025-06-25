-- Test hooks module
local success, hooks = pcall(require, "hooks")
if success then
    print("Hooks module loaded:", hooks ~= nil)
    if hooks then
        print("Hooks functions:")
        for k, v in pairs(hooks) do
            print("  ", k, type(v))
        end
    end
else
    print("Failed to load hooks module:", hooks)
end

-- Check bridges for hooks
if bridges and bridges.agent_hooks then
    print("\nHooks bridge available")
    print("Bridge methods:")
    for k, v in pairs(bridges.agent_hooks) do
        print("  ", k, type(v))
    end
end

return "done"