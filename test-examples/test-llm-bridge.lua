-- Test LLM bridge methods
print("Testing LLM bridge methods...")

if bridges and bridges.llm_core then
    print("\nllm_core bridge methods:")
    for k, v in pairs(bridges.llm_core) do
        print("  ", k, type(v))
    end
else
    print("No llm_core bridge found")
end

return "done"