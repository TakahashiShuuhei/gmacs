-- Test Lua configuration file for gmacs

-- Test key binding
gmacs.bind_key("C-x t", "hello-world")

-- Test option setting
gmacs.set_option("test-option", "test-value")
gmacs.set_option("tab-width", 4)

-- Test custom command definition
gmacs.defun("hello-world", function()
    -- This is a test command
end)

-- Test hook (simple test)
gmacs.add_hook("test-event", function()
    -- This is a test hook
end)