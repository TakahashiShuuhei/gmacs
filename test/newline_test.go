package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestNewlineBasic(t *testing.T) {
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// Initial state
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("Initial: content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After 'hello': content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After Enter: content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Type "world"
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After 'world': content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Check final state
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "hello" {
		t.Errorf("Expected first line 'hello', got %q", content[0])
	}
	if content[1] != "world" {
		t.Errorf("Expected second line 'world', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

func TestNewlineInMiddle(t *testing.T) {
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// Type "hello world"
	for _, ch := range "hello world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to after "hello" (position 5)
	buffer.SetCursor(domain.Position{Row: 0, Col: 5})
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("Before Enter in middle: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Press Enter in the middle
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After Enter in middle: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the split
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "hello" {
		t.Errorf("Expected first line 'hello', got %q", content[0])
	}
	if content[1] != " world" {
		t.Errorf("Expected second line ' world', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (1,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

func TestNewlineAtBeginning(t *testing.T) {
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Press Enter at beginning
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("After Enter at beginning: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the result
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "" {
		t.Errorf("Expected first line empty, got %q", content[0])
	}
	if content[1] != "hello" {
		t.Errorf("Expected second line 'hello', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (1,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

func TestMultipleNewlines(t *testing.T) {
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// Type "a", Enter, "b", Enter, "c"
	chars := []rune{'a', '\n', 'b', '\n', 'c'}
	for _, ch := range chars {
		if ch == '\n' {
			event := events.KeyEventData{Key: "Enter", Rune: ch}
			editor.HandleEvent(event)
		} else {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("After multiple newlines: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the result
	if len(content) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(content))
	}
	if content[0] != "a" {
		t.Errorf("Expected line 0 'a', got %q", content[0])
	}
	if content[1] != "b" {
		t.Errorf("Expected line 1 'b', got %q", content[1])
	}
	if content[2] != "c" {
		t.Errorf("Expected line 2 'c', got %q", content[2])
	}
	if cursor.Row != 2 || cursor.Col != 1 {
		t.Errorf("Expected cursor at (2,1), got (%d,%d)", cursor.Row, cursor.Col)
	}
}