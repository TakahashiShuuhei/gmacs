package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
)

// Mock implementations for testing

type mockMinibuffer struct {
	messages []string
}

func (m *mockMinibuffer) ShowMessage(message string) {
	m.messages = append(m.messages, message)
}

type mockEditor struct {
	registry   *command.Registry
	minibuffer *mockMinibuffer
	keyBindings map[string]string
}

func (e *mockEditor) GetMinibuffer() MinibufferInterface {
	return e.minibuffer
}

func (e *mockEditor) GetCommandRegistry() *command.Registry {
	return e.registry
}

func (e *mockEditor) BindKey(keySeq string, command string) error {
	if e.keyBindings == nil {
		e.keyBindings = make(map[string]string)
	}
	e.keyBindings[keySeq] = command
	return nil
}

func TestLuaConfig_LoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "gmacs_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "init.lua")
	configContent := `
-- Test configuration
gmacs.set_variable("test-var", "test-value")
gmacs.register_command("test-command", function()
    gmacs.message("Test command executed")
end, "A test command")
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create mock editor
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	// Create config
	luaConfig := NewLuaConfig(mockEditor)
	luaConfig.SetConfigPath(configFile)

	// Load the config
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Verify that the command was registered
	_, exists := mockEditor.registry.Get("test-command")
	if !exists {
		t.Error("Expected test-command to be registered, but it wasn't found")
	}
}

func TestLuaConfig_NoConfigFile(t *testing.T) {
	// Test with non-existent config file
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)

	// Set the config path to a non-existent file
	luaConfig.SetConfigPath("/nonexistent/path/init.lua")

	// Should not error when config file doesn't exist
	err := luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig should not error when config file doesn't exist, got: %v", err)
	}
}

func TestLuaConfig_CommandRegistration(t *testing.T) {
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)

	// Test script that registers a command
	testScript := `
function test_func()
    return "success"
end

gmacs.register_command("lua-test", test_func, "Test command from Lua")
`

	// Create temporary file
	tempDir, err := os.MkdirTemp("", "gmacs_lua_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scriptFile := filepath.Join(tempDir, "test.lua")
	err = os.WriteFile(scriptFile, []byte(testScript), 0644)
	if err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	// Set config path
	luaConfig.SetConfigPath(scriptFile)

	// Load and test
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Check if command was registered
	cmd, exists := mockEditor.registry.Get("lua-test")
	if !exists {
		t.Error("Expected lua-test command to be registered")
		return
	}

	if cmd.Description != "Test command from Lua" {
		t.Errorf("Expected description 'Test command from Lua', got '%s'", cmd.Description)
	}

	// Try to execute the command
	err = mockEditor.registry.Execute("lua-test")
	if err != nil {
		t.Errorf("Failed to execute lua-test command: %v", err)
	}
}