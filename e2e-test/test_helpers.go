package test

import (
	"io/ioutil"
	"path/filepath"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/lua-config"
)

// getDefaultConfig reads the default.lua file
func getDefaultConfig() string {
	content, err := ioutil.ReadFile(filepath.Join("..", "lua-config", "default.lua"))
	if err != nil {
		panic("Failed to read default.lua: " + err.Error())
	}
	return string(content)
}

// NewEditorWithDefaults creates an editor with default configuration loaded
// This ensures tests use the same configuration as the main application
func NewEditorWithDefaults() *domain.Editor {
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Register built-in commands (same as NewEditor does for backward compatibility)
	editor.RegisterBuiltinCommands()
	
	// Register Lua API
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	apiBindings.RegisterGmacsAPI()
	
	// Load default configuration
	err := configLoader.GetVM().ExecuteString(getDefaultConfig())
	if err != nil {
		panic("Failed to load default config in test: " + err.Error())
	}
	
	return editor
}