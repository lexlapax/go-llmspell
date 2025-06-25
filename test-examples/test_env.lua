local env_var = os.getenv("TEST_VAR"); return "TEST_VAR=" .. (env_var or "not set")
