-- Very simple test config
print("Lua config is being executed!")

-- Test a simple option
print("About to call set_option...")
gmacs.set_option("simple-test", "works")
print("set_option called successfully!")