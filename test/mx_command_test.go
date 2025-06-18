package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestMxCommandBasic(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 5)
	
	// Start with normal mode
	display.Render(editor)
	modeLine := display.GetModeLine()
	expectedModeLine := " *scratch* " + strings.Repeat("-", 69) // 80 - 11 = 69
	if modeLine != expectedModeLine {
		t.Errorf("Expected normal mode line, got %q", modeLine)
	}
	
	// Press ESC (Meta key)
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	
	// Press x for M-x
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Check minibuffer is active
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active after M-x")
	}
	
	if minibuffer.Mode() != domain.MinibufferCommand {
		t.Error("Minibuffer should be in command mode")
	}
	
	// Render and check prompt
	display.Render(editor)
	modeLine = display.GetModeLine()
	expectedPrompt := "M-x " + strings.Repeat(" ", 76)
	if modeLine != expectedPrompt {
		t.Errorf("Expected M-x prompt, got %q", modeLine)
	}
	
	// Check cursor position (should be after "M-x ")
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 4 || cursorCol != 4 { // height-1 = 4, after "M-x "
		t.Errorf("Expected cursor at (4, 4), got (%d, %d)", cursorRow, cursorCol)
	}
}

func TestMxVersionCommand(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 5)
	
	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type "version"
	versionText := "version"
	for _, ch := range versionText {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer content
	display.Render(editor)
	modeLine := display.GetModeLine()
	expectedContent := "M-x version" + strings.Repeat(" ", 69)
	if modeLine != expectedContent {
		t.Errorf("Expected 'M-x version', got %q", modeLine)
	}
	
	// Press Enter to execute
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Check that command was executed and minibuffer shows version message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should still be active showing version message")
	}
	
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Minibuffer should be in message mode")
	}
	
	modeLine = display.GetModeLine()
	expectedVersion := "gmacs 0.1.0 - Emacs-like text editor in Go"
	expectedPadding := strings.Repeat(" ", 80-len(expectedVersion))
	expectedLine := expectedVersion + expectedPadding
	if modeLine != expectedLine {
		t.Errorf("Expected version message, got %q", modeLine)
	}
	
	// Any key should clear the message and insert into buffer
	anyKeyEvent := events.KeyEventData{Key: "a", Rune: 'a'}
	editor.HandleEvent(anyKeyEvent)
	
	// Should return to normal mode and insert the character
	display.Render(editor)
	if editor.Minibuffer().IsActive() {
		t.Error("Minibuffer should be cleared after any key")
	}
	
	content := display.GetContent()
	if len(content) == 0 || content[0] != "a" {
		t.Errorf("Expected 'a' to be inserted, got content: %v", content)
	}
}

func TestMxUnknownCommand(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 5)
	
	// Start M-x and type unknown command
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type "nonexistent"
	for _, ch := range "nonexistent" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Should show error message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should show error message")
	}
	
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Minibuffer should be in message mode")
	}
	
	modeLine := display.GetModeLine()
	if !strings.Contains(modeLine, "Unknown command: nonexistent") {
		t.Errorf("Expected unknown command error, got %q", modeLine)
	}
}

func TestMxCancel(t *testing.T) {
	editor := domain.NewEditor()
	
	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type some text
	for _, ch := range "ver" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer is active with content
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active")
	}
	if minibuffer.Content() != "ver" {
		t.Errorf("Expected content 'ver', got %q", minibuffer.Content())
	}
	
	// Press Escape to cancel
	cancelEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(cancelEvent)
	
	// Minibuffer should be cleared
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be cleared after cancel")
	}
}

func TestMxListCommands(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 5)
	
	// Execute list-commands
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	for _, ch := range "list-commands" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Check message contains available commands
	display.Render(editor)
	modeLine := display.GetModeLine()
	if !strings.Contains(modeLine, "Available commands:") {
		t.Errorf("Expected command list message, got %q", modeLine)
	}
	if !strings.Contains(modeLine, "version") {
		t.Errorf("Expected version command in list, got %q", modeLine)
	}
	if !strings.Contains(modeLine, "list-commands") {
		t.Errorf("Expected list-commands in list, got %q", modeLine)
	}
}

func TestMxClearBuffer(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 5)
	
	// Add some content to buffer
	for _, ch := range "hello world" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content := display.GetContent()
	if content[0] != "hello world" {
		t.Errorf("Expected 'hello world', got %q", content[0])
	}
	
	// Execute clear-buffer
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	for _, ch := range "clear-buffer" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Buffer should be cleared
	display.Render(editor)
	content = display.GetContent()
	if content[0] != "" {
		t.Errorf("Expected empty buffer, got %q", content[0])
	}
	
	// Should show clear message
	modeLine := display.GetModeLine()
	if !strings.Contains(modeLine, "Buffer cleared") {
		t.Errorf("Expected buffer cleared message, got %q", modeLine)
	}
}