-- Test state_context bridge
print("Testing state_context bridge...")

if bridges and bridges.state_context then
    print("\nstate_context bridge available")
    print("Bridge methods:")
    for k, v in pairs(bridges.state_context) do
        print("  ", k, type(v))
    end
else
    print("No state_context bridge found")
end

return "done"