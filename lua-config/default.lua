-- Default gmacs configuration
-- This file contains all the built-in keybindings and commands

-- Note: Built-in Go commands are automatically registered by the API bindings system
-- This file only needs to set up key bindings to existing commands

-- Cursor movement commands
gmacs.bind_key("C-f", "forward-char")
gmacs.bind_key("C-b", "backward-char") 
gmacs.bind_key("C-n", "next-line")
gmacs.bind_key("C-p", "previous-line")
gmacs.bind_key("C-a", "beginning-of-line")
gmacs.bind_key("C-e", "end-of-line")

-- Scrolling commands
gmacs.bind_key("C-v", "page-down")
gmacs.bind_key("M-v", "page-up")
gmacs.bind_key("C-u", "scroll-up")
gmacs.bind_key("C-d", "scroll-down")

-- Buffer management
gmacs.bind_key("C-x b", "switch-to-buffer")
gmacs.bind_key("C-x C-b", "list-buffers") 
gmacs.bind_key("C-x k", "kill-buffer")

-- Window management
gmacs.bind_key("C-x 2", "split-window-below")
gmacs.bind_key("C-x 3", "split-window-right")
gmacs.bind_key("C-x o", "other-window")
gmacs.bind_key("C-x 0", "delete-window")
gmacs.bind_key("C-x 1", "delete-other-windows")

-- Line wrapping toggle
gmacs.bind_key("C-x t", "toggle-truncate-lines")

-- Debug info
gmacs.bind_key("C-x d", "debug-info")

-- Quit and cancel commands
gmacs.bind_key("C-g", "keyboard-quit")
gmacs.bind_key("C-x C-c", "quit")

-- File operations  
gmacs.bind_key("C-x C-f", "find-file")

-- Minor mode commands  
gmacs.bind_key("M-a", "auto-a-mode")

print("Default gmacs configuration loaded")