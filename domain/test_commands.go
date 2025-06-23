package domain

import "github.com/TakahashiShuuhei/gmacs/log"

// TestCommand is a simple test command to verify the refactored registration system
func TestCommand(editor *Editor) error {
	message := "Test command executed successfully via new registration system!"
	editor.SetMinibufferMessage(message)
	log.Info("Test command executed")
	return nil
}

// RegisterTestCommands registers test commands to verify the new system
func RegisterTestCommands(registry *CommandRegistry) {
	registry.RegisterFunc("test-refactor", TestCommand)
}