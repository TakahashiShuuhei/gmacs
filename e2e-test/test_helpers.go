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

// NewEditorWithDefaults creates an editor with the EXACT same initialization as main.go
// This ensures tests use the same configuration as the main application
func NewEditorWithDefaults() *domain.Editor {
	// Step 1: Create editor with Lua configuration support (same as main.go)
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Step 2: Register Lua API (same as main.go - this also registers built-in commands)
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	if err := apiBindings.RegisterGmacsAPI(); err != nil {
		panic("Failed to register Lua API in test: " + err.Error())
	}
	
	// Step 3: Load default configuration (same as main.go)
	err := configLoader.GetVM().ExecuteString(getDefaultConfig())
	if err != nil {
		panic("Failed to load default config in test: " + err.Error())
	}
	
	return editor
}