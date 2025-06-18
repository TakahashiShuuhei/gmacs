package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestMockDisplayBasic(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// Input some text
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	if len(content) != 3 { // height-2 for mode line and minibuffer
		t.Errorf("Expected 3 content lines, got %d", len(content))
	}
	
	if content[0] != "hello" {
		t.Errorf("Expected 'hello', got %q", content[0])
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 0 || cursorCol != 5 {
		t.Errorf("Expected cursor at (0,5), got (%d,%d)", cursorRow, cursorCol)
	}
}

func TestMockDisplayJapanese(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// Input Japanese text
	text := "あいう"
	for _, ch := range []rune(text) {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	if content[0] != "あいう" {
		t.Errorf("Expected 'あいう', got %q", content[0])
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	t.Logf("Japanese text cursor position: (%d, %d)", cursorRow, cursorCol)
	
	// Show screen with cursor
	screenWithCursor := display.GetScreenWithCursor()
	t.Logf("Screen with cursor:\n%s", screenWithCursor)
	
	// Show detailed info
	t.Logf("Screen info:\n%s", display.GetScreenInfo())
}

func TestMockDisplayCursorProgression(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 5)
	
	testChars := []rune{'a', 'あ', 'b', 'い', 'c'}
	expectedPositions := []struct{ row, col int }{
		{0, 1}, // a(1)
		{0, 3}, // あ(2) -> total 3
		{0, 4}, // b(1) -> total 4
		{0, 6}, // い(2) -> total 6
		{0, 7}, // c(1) -> total 7
	}
	
	for i, ch := range testChars {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
		
		display.Render(editor)
		
		cursorRow, cursorCol := display.GetCursorPosition()
		
		t.Logf("After '%c': cursor at (%d, %d), expected (%d, %d)", 
			ch, cursorRow, cursorCol, expectedPositions[i].row, expectedPositions[i].col)
		t.Logf("Content: %q", display.GetContent()[0])
		t.Logf("Screen with cursor:\n%s\n", display.GetScreenWithCursor())
		
		if cursorRow != expectedPositions[i].row || cursorCol != expectedPositions[i].col {
			t.Errorf("After '%c': expected cursor (%d, %d), got (%d, %d)",
				ch, expectedPositions[i].row, expectedPositions[i].col, cursorRow, cursorCol)
		}
	}
}

func TestMockDisplayMultiline(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// First line
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Second line
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	if content[0] != "hello" {
		t.Errorf("Expected line 0 'hello', got %q", content[0])
	}
	if content[1] != "world" {
		t.Errorf("Expected line 1 'world', got %q", content[1])
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 1 || cursorCol != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursorRow, cursorCol)
	}
	
	t.Logf("Multi-line screen:\n%s", display.GetScreenWithCursor())
}

func TestMockDisplayWidthProblem(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 3)
	
	testCases := []struct {
		input    string
		expected string
	}{
		{"abc", "abc|"},
		{"あいう", "あいう|"},
		{"aあb", "aあb|"},
	}
	
	for _, tc := range testCases {
		// Reset editor
		editor = domain.NewEditor()
		
		// Input text
		for _, ch := range []rune(tc.input) {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		screenWithCursor := display.GetScreenWithCursor()
		lines := strings.Split(screenWithCursor, "\n")
		actualFirstLine := strings.TrimSpace(lines[0])
		
		t.Logf("Input: %q", tc.input)
		t.Logf("Expected: %q", tc.expected)
		t.Logf("Actual: %q", actualFirstLine)
		
		// Note: This test currently shows the cursor position problem
		// The cursor should account for terminal display width, not just rune count
	}
}