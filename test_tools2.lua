print("Testing tools"); local success, result = pcall(require, "tools"); print("Tools require result:", success, result); return {success=success, result=type(result)}
