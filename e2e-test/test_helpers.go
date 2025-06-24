package test

import (
	"io/ioutil"
	"path/filepath"
	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/lua-config"
	"github.com/TakahashiShuuhei/gmacs/plugin"
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
// For tests, plugin paths are empty to avoid interference from installed plugins
func NewEditorWithDefaults() *domain.Editor {
	// Step 1: Create editor with Lua configuration and plugin system (same as main.go)
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	
	// Use empty plugin paths for tests to ensure isolation
	editor := plugin.CreateEditorWithPluginsAndPaths(configLoader, hookManager, []string{})
	
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

// NewEditorWithTestPlugins creates an editor with test plugins loaded
func NewEditorWithTestPlugins() *domain.Editor {
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	
	// Use test plugin directory
	testPluginPaths := []string{"/tmp/gmacs-test-plugins"}
	editor := plugin.CreateEditorWithPluginsAndPaths(configLoader, hookManager, testPluginPaths)
	
	// Register Lua API
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	if err := apiBindings.RegisterGmacsAPI(); err != nil {
		panic("Failed to register Lua API in test: " + err.Error())
	}
	
	// Load default configuration
	err := configLoader.GetVM().ExecuteString(getDefaultConfig())
	if err != nil {
		panic("Failed to load default config in test: " + err.Error())
	}
	
	return editor
}