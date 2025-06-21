package luaconfig

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

func TestLocalBindKey(t *testing.T) {
	// Create editor with config
	editor := domain.NewEditor()
	
	// Create a simple Lua config that binds a key to a command in fundamental-mode
	configContent := `
		gmacs.local_bind_key("fundamental-mode", "C-t", "version")
	`
	
	// Create config loader and register API
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Execute the config
	err = configLoader.GetVM().DoString(configContent)
	if err != nil {
		t.Fatalf("Failed to execute config: %v", err)
	}
	
	// Get the fundamental mode and check if the key binding was added
	modeManager := editor.ModeManager()
	fundamentalMode, exists := modeManager.GetMajorModeByName("fundamental-mode")
	if !exists {
		t.Fatal("fundamental-mode should exist")
	}
	
	keyBindings := fundamentalMode.KeyBindings()
	if keyBindings == nil {
		t.Fatal("fundamental-mode should have key bindings")
	}
	
	// Check if the key binding exists
	cmd, found := keyBindings.LookupSequence("C-t")
	if !found {
		t.Error("Key binding C-t should be registered in fundamental-mode")
	}
	
	// Verify the command works by executing it
	if cmd != nil {
		err := cmd(editor)
		if err != nil {
			t.Errorf("Command should execute without error: %v", err)
		}
		
		// Check if version message is displayed
		minibuffer := editor.Minibuffer()
		if !minibuffer.IsActive() {
			t.Error("Minibuffer should be active after version command")
		}
		
		if !strings.Contains(minibuffer.Message(), "gmacs") {
			t.Errorf("Expected version message, got: %s", minibuffer.Message())
		}
	}
}

func TestLocalBindKeyUnknownMode(t *testing.T) {
	editor := domain.NewEditor()
	
	// Try to bind key to non-existent mode
	configContent := `
		gmacs.local_bind_key("non-existent-mode", "C-t", "version")
	`
	
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// This should produce an error
	err = configLoader.GetVM().DoString(configContent)
	if err == nil {
		t.Error("Expected error when binding to unknown mode")
	}
	
	if !strings.Contains(err.Error(), "Unknown mode") {
		t.Errorf("Expected 'Unknown mode' error, got: %v", err)
	}
}

func TestLocalBindKeyUnknownCommand(t *testing.T) {
	editor := domain.NewEditor()
	
	// Try to bind key to non-existent command
	configContent := `
		gmacs.local_bind_key("fundamental-mode", "C-t", "non-existent-command")
	`
	
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// This should produce an error
	err = configLoader.GetVM().DoString(configContent)
	if err == nil {
		t.Error("Expected error when binding to unknown command")
	}
	
	if !strings.Contains(err.Error(), "Unknown command") {
		t.Errorf("Expected 'Unknown command' error, got: %v", err)
	}
}

func TestLocalBindKeyMinorMode(t *testing.T) {
	editor := domain.NewEditor()
	
	// Bind key to minor mode
	configContent := `
		gmacs.local_bind_key("auto-a-mode", "C-a", "version")
	`
	
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Execute the config
	err = configLoader.GetVM().DoString(configContent)
	if err != nil {
		t.Fatalf("Failed to execute config: %v", err)
	}
	
	// Get the auto-a-mode and check if the key binding was added
	modeManager := editor.ModeManager()
	autoAMode, exists := modeManager.GetMinorModeByName("auto-a-mode")
	if !exists {
		t.Fatal("auto-a-mode should exist")
	}
	
	keyBindings := autoAMode.KeyBindings()
	if keyBindings == nil {
		t.Fatal("auto-a-mode should have key bindings")
	}
	
	// Check if the key binding exists
	_, found := keyBindings.LookupSequence("C-a")
	if !found {
		t.Error("Key binding C-a should be registered in auto-a-mode")
	}
}