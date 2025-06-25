-- Test LLM methods
local llm = require("llm")

print("Testing LLM module...")
if llm then
    print("LLM methods:")
    for k, v in pairs(llm) do
        print("  ", k, type(v))
    end
end

return "done"