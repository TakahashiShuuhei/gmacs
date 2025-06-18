package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestNewlineDisplay(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 5)
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content := display.GetContent()
	t.Logf("After 'hello': display content = %v", content)
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	display.Render(editor)
	content = display.GetContent()
	t.Logf("After Enter: display content = %v", content)
	
	// Type "world"
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content = display.GetContent()
	t.Logf("After 'world': display content = %v", content)
	
	// Check that it displays correctly
	if content[0] != "hello" {
		t.Errorf("Expected line 0 'hello', got %q", content[0])
	}
	if content[1] != "world" {
		t.Errorf("Expected line 1 'world', got %q", content[1])
	}
	
	// Check cursor position
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 1 || cursorCol != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursorRow, cursorCol)
	}
	
	// Show actual buffer content vs display content
	buffer := editor.CurrentBuffer()
	bufferContent := buffer.Content()
	t.Logf("Buffer content: %v (length: %d)", bufferContent, len(bufferContent))
	t.Logf("Display content: %v", content)
}

func TestNewlineAtEndOfLine(t *testing.T) {
	editor := domain.NewEditor()
	
	// Type "abc" + Enter + "def" + Enter + "ghi"
	inputs := []struct {
		text   string
		isEnter bool
	}{
		{"abc", false},
		{"", true},
		{"def", false},
		{"", true},
		{"ghi", false},
	}
	
	for _, input := range inputs {
		if input.isEnter {
			event := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(event)
		} else {
			for _, ch := range input.text {
				event := events.KeyEventData{Rune: ch, Key: string(ch)}
				editor.HandleEvent(event)
			}
		}
	}
	
	buffer := editor.CurrentBuffer()
	content := buffer.Content()
	cursor := buffer.Cursor()
	
	t.Logf("Final content: %v (lines: %d)", content, len(content))
	t.Logf("Final cursor: (%d,%d)", cursor.Row, cursor.Col)
	
	// Should have 3 lines
	if len(content) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(content))
	}
	if content[0] != "abc" {
		t.Errorf("Expected line 0 'abc', got %q", content[0])
	}
	if content[1] != "def" {
		t.Errorf("Expected line 1 'def', got %q", content[1])
	}
	if content[2] != "ghi" {
		t.Errorf("Expected line 2 'ghi', got %q", content[2])
	}
	
	// Cursor should be at end of line 2
	if cursor.Row != 2 || cursor.Col != 3 {
		t.Errorf("Expected cursor at (2,3), got (%d,%d)", cursor.Row, cursor.Col)
	}
}