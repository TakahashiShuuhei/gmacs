package luaconfig

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

func TestLocalBindKey(t *testing.T) {
	// Create editor with Lua config support
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	hookManager := NewHookManager()
	
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Create a simple Lua config that binds a key to a command in fundamental-mode
	configContent := `
		gmacs.local_bind_key("fundamental-mode", "C-t", "forward-char")
	`
	
	// Register API with the same configLoader
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Execute the config
	t.Logf("Executing Lua config: %s", configContent)
	err = configLoader.GetVM().ExecuteString(configContent)
	if err != nil {
		t.Fatalf("Failed to execute config: %v", err)
	}
	t.Log("Lua config executed successfully")
	
	// Debug: Try calling LocalBindKey directly
	err = editor.LocalBindKey("fundamental-mode", "C-x", "forward-char")
	if err != nil {
		t.Fatalf("Direct LocalBindKey call failed: %v", err)
	}
	t.Log("Direct LocalBindKey call succeeded")
	
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
	
	// Check if the direct key binding exists
	cmd, found := keyBindings.HasKeySequenceBinding("C-x")
	if !found {
		t.Error("Key binding C-x should be registered in fundamental-mode (direct call)")
	} else {
		t.Log("Direct key binding C-x found successfully")
	}
	
	// Check if the Lua key binding exists
	cmd, found = keyBindings.HasKeySequenceBinding("C-t")
	if !found {
		t.Error("Key binding C-t should be registered in fundamental-mode (Lua call)")
		t.Logf("fundamental-mode exists: %v", exists)
		t.Logf("keyBindings is nil: %v", keyBindings == nil)
	}
	
	// Verify the command works by executing it
	if cmd != nil {
		// Add some text to the buffer first so forward-char has something to move over
		buffer := editor.CurrentBuffer()
		if buffer != nil {
			buffer.InsertString("hello")
			buffer.SetCursor(domain.Position{Row: 0, Col: 0}) // Move cursor to beginning
			
			// Store initial cursor position
			initialPos := buffer.Cursor()
			
			// Execute the forward-char command
			err := cmd(editor)
			if err != nil {
				t.Errorf("Command should execute without error: %v", err)
			}
			
			// Check if cursor moved forward
			newPos := buffer.Cursor()
			if newPos.Col != initialPos.Col + 1 {
				t.Errorf("Expected cursor to move from col %d to %d, but got %d", 
					initialPos.Col, initialPos.Col + 1, newPos.Col)
			}
		}
	}
}

func TestLocalBindKeyUnknownMode(t *testing.T) {
	editor := domain.NewEditor()
	
	// Try to bind key to non-existent mode
	configContent := `
		gmacs.local_bind_key("non-existent-mode", "C-t", "forward-char")
	`
	
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// This should produce an error
	err = configLoader.GetVM().ExecuteString(configContent)
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
	err = configLoader.GetVM().ExecuteString(configContent)
	if err == nil {
		t.Error("Expected error when binding to unknown command")
	}
	
	if !strings.Contains(err.Error(), "Unknown command") {
		t.Errorf("Expected 'Unknown command' error, got: %v", err)
	}
}

func TestLocalBindKeyMinorMode(t *testing.T) {
	// Create editor with Lua config support
	configLoader := NewConfigLoader()
	defer configLoader.Close()
	hookManager := NewHookManager()
	
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Bind key to minor mode
	configContent := `
		gmacs.local_bind_key("auto-a-mode", "C-a", "auto-a-mode")
	`
	
	// Register API with the same configLoader
	apiBindings := NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Execute the config
	err = configLoader.GetVM().ExecuteString(configContent)
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
	_, found := keyBindings.HasKeySequenceBinding("C-a")
	if !found {
		t.Error("Key binding C-a should be registered in auto-a-mode")
	}
}