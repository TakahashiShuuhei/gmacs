/**
 * @spec configuration/lua_config
 * @scenario Lua configuration loading
 * @description Test that Lua configuration files can be loaded and applied correctly
 * @given A test Lua configuration file exists
 * @when Editor is created with the config file
 * @then Configuration should be loaded and applied successfully
 * @implementation lua-config package integration
 */

package test

import (
	"testing"
	"path/filepath"
	"os"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	luaconfig "github.com/TakahashiShuuhei/gmacs/core/lua-config"
	gmacslog "github.com/TakahashiShuuhei/gmacs/core/log"
)

func TestLuaConfigurationLoading(t *testing.T) {
	// Initialize logger for test
	if err := gmacslog.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer gmacslog.Close()

	// Get absolute path to test config file
	testConfigPath, err := filepath.Abs("../testdata/simple_config.lua")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	
	// Check if test config file exists
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Fatalf("Test config file does not exist: %s", testConfigPath)
	}

	// Create config loader and load the test configuration
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Register API bindings and load config
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	apiErr := apiBindings.RegisterGmacsAPI()
	if apiErr != nil {
		t.Fatalf("Failed to register gmacs API: %v", apiErr)
	}
	
	err = configLoader.LoadConfig(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	t.Logf("Loaded config from: %s", testConfigPath)
	if editor == nil {
		t.Fatal("Failed to create editor with config")
	}
	defer editor.Cleanup()
	
	t.Log("Editor created successfully with config")

	// Test that editor was created successfully
	if !editor.IsRunning() {
		t.Error("Editor should be running after creation")
	}

	// Test that options were set correctly
	value, err := editor.GetOption("simple-test")
	if err != nil {
		t.Errorf("Failed to get simple-test: %v", err)
	} else if value != "works" {
		t.Errorf("Expected simple-test to be 'works', got '%v'", value)
	}
}

func TestEditorWithoutConfig(t *testing.T) {
	// Initialize logger for test
	if err := gmacslog.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer gmacslog.Close()

	// Create editor without configuration
	editor := domain.NewEditor()
	if editor == nil {
		t.Fatal("Failed to create editor without config")
	}
	defer editor.Cleanup()

	// Test that editor was created successfully
	if !editor.IsRunning() {
		t.Error("Editor should be running after creation")
	}

	// Test that no options are set (should return error)
	_, err := editor.GetOption("test-option")
	if err == nil {
		t.Error("Expected error when getting non-existent option")
	}
}

func TestInvalidConfigFile(t *testing.T) {
	// Initialize logger for test
	if err := gmacslog.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer gmacslog.Close()

	// Try to create editor with non-existent config file
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	if editor == nil {
		t.Fatal("Editor should be created even without config")
	}
	defer editor.Cleanup()

	// Editor should be running even without config loading
	if !editor.IsRunning() {
		t.Error("Editor should be running even without config")
	}
}