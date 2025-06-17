package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
)


func TestLuaVersionCommand(t *testing.T) {
	// Create a temporary config file with custom command (not version to avoid conflict)
	tempDir, err := os.MkdirTemp("", "gmacs_lua_command_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "test.lua")
	configContent := `
function custom_command()
    gmacs.message("Custom command executed")
end

gmacs.register_command("custom-test", custom_command, "Test custom command")
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create mock editor and config
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)
	luaConfig.SetConfigPath(configFile)

	// Load config
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
		return
	}

	// Check if custom command was registered
	_, exists := mockEditor.registry.Get("custom-test")
	if !exists {
		t.Error("Expected custom-test command to be registered")
		return
	}

	// Execute the custom command
	err = mockEditor.registry.Execute("custom-test")
	if err != nil {
		t.Errorf("Failed to execute custom-test command: %v", err)
		return
	}

	// Check if message was displayed
	if len(mockEditor.minibuffer.messages) == 0 {
		t.Error("Expected at least 1 message")
		return
	}

	// Find our custom message
	found := false
	for _, msg := range mockEditor.minibuffer.messages {
		if msg == "Custom command executed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find 'Custom command executed' in messages: %v", mockEditor.minibuffer.messages)
	}
}

func TestLuaDefaultCommands(t *testing.T) {
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)

	// Initialize Lua VM and expose API (like LoadConfig does)
	luaConfig.vm = lua.NewState()
	defer luaConfig.vm.Close()
	luaConfig.exposeGmacsAPI()

	// Load default config (which includes version command)
	err := luaConfig.loadDefaultConfig()
	if err != nil {
		t.Errorf("loadDefaultConfig failed: %v", err)
		return
	}

	// Check if version command from default.lua was registered
	cmd, exists := mockEditor.registry.Get("version")
	if !exists {
		t.Error("Expected version command to be registered from default.lua")
		return
	}

	if !strings.Contains(cmd.Description, "version") {
		t.Errorf("Expected version command description to contain 'version', got '%s'", cmd.Description)
	}

	// Execute the version command
	err = mockEditor.registry.Execute("version")
	if err != nil {
		t.Errorf("Failed to execute version command: %v", err)
		return
	}

	// Check if message was displayed
	if len(mockEditor.minibuffer.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mockEditor.minibuffer.messages))
		return
	}

	expectedMessage := "gmacs version 0.0.1 - Go Emacs-like Editor"
	if mockEditor.minibuffer.messages[0] != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, mockEditor.minibuffer.messages[0])
	}
}