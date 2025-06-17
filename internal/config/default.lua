-- gmacs Default Key Bindings
-- This file contains the default key bindings for gmacs

-- Basic file operations
gmacs.global_set_key("C-x C-f", "find-file")
gmacs.global_set_key("C-x C-s", "save-buffer")
gmacs.global_set_key("C-x C-w", "write-file")
gmacs.global_set_key("C-x C-c", "quit")

-- Basic cursor movement
gmacs.global_set_key("C-f", "forward-char")
gmacs.global_set_key("C-b", "backward-char")
gmacs.global_set_key("C-n", "next-line")
gmacs.global_set_key("C-p", "previous-line")

-- Arrow keys (same as Ctrl keys)
gmacs.global_set_key("right", "forward-char")
gmacs.global_set_key("left", "backward-char")
gmacs.global_set_key("down", "next-line")
gmacs.global_set_key("up", "previous-line")

-- Text deletion
gmacs.global_set_key("C-d", "delete-char")
gmacs.global_set_key("backspace", "backward-delete-char")

-- Custom commands defined in Lua
function version_command()
    gmacs.message("gmacs version 0.0.1 - Go Emacs-like Editor")
end

gmacs.register_command("version", version_command, "Show gmacs version")