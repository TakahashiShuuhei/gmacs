package display

import (
	"strings"
	"testing"
)

func TestSimpleEditor(t *testing.T) {
	editor := NewSimpleEditor()
	
	if editor.currentWin == nil {
		t.Error("SimpleEditor should have a current window")
	}
	
	if editor.registry == nil {
		t.Error("SimpleEditor should have a command registry")
	}
	
	if !editor.running {
		t.Error("SimpleEditor should be running initially")
	}
}

func TestSimpleEditorCommands(t *testing.T) {
	editor := NewSimpleEditor()
	
	// Test quit command
	err := editor.handleCommand("quit")
	if err != nil {
		t.Errorf("Quit command should not return error: %v", err)
	}
	
	if editor.running {
		t.Error("Editor should not be running after quit")
	}
	
	// Reset for more tests
	editor.running = true
	
	// Test help command
	err = editor.handleCommand("help")
	if err != nil {
		t.Errorf("Help command should not return error: %v", err)
	}
	
	// Test direct command execution
	err = editor.handleCommand("version")
	if err != nil {
		t.Errorf("Version command should not return error: %v", err)
	}
}

func TestSimpleEditorRegisteredCommands(t *testing.T) {
	editor := NewSimpleEditor()
	
	// Test that basic commands are registered
	expectedCommands := []string{"version", "hello", "list-commands", "buffer-info", "echo"}
	
	registeredCommands := editor.registry.List()
	
	for _, expected := range expectedCommands {
		found := false
		for _, registered := range registeredCommands {
			if registered == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command '%s' should be registered", expected)
		}
	}
	
	// Test command execution
	err := editor.registry.Execute("hello", "Test")
	if err != nil {
		t.Errorf("Hello command should execute without error: %v", err)
	}
	
	err = editor.registry.Execute("echo", "arg1", "arg2")
	if err != nil {
		t.Errorf("Echo command should execute without error: %v", err)
	}
}

func TestBufferOperations(t *testing.T) {
	editor := NewSimpleEditor()
	
	// Test buffer access
	buf := editor.currentWin.Buffer()
	if buf.Name() != "*scratch*" {
		t.Errorf("Expected buffer name '*scratch*', got '%s'", buf.Name())
	}
	
	// Test that buffer has initial content
	if buf.LineCount() == 0 {
		t.Error("Buffer should have initial content")
	}
	
	content := buf.GetText()
	if !strings.Contains(content, "Welcome") {
		t.Error("Buffer should contain welcome message")
	}
}