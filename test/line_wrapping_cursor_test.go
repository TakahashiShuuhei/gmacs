package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestCursorPositionWithLineWrapping(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 8) // Small window for testing wrapping
	
	// Set window size to match display content area (height-2)
	window := editor.CurrentWindow()
	window.Resize(10, 6)
	
	// Enable line wrapping (should be default)
	window.SetLineWrap(true)

	// Add a long line that will wrap
	longLine := "This is a very long line that should wrap"
	for _, ch := range longLine {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}

	display.Render(editor)

	// Check cursor position after typing
	cursorRow, cursorCol := display.GetCursorPosition()
	t.Logf("Cursor position after typing long line: (%d, %d)", cursorRow, cursorCol)

	// The line should be wrapped, so cursor might be on a different screen row
	visible := window.VisibleLines()
	t.Logf("Visible lines: %v", visible)

	// Test cursor movement at line wrap boundaries
	buffer := editor.CurrentBuffer()
	cursor := buffer.Cursor()
	t.Logf("Buffer cursor position: (%d, %d)", cursor.Row, cursor.Col)

	// Move cursor to middle of wrapped line and test
	buffer.SetCursor(domain.Position{Row: 0, Col: 15}) // Middle of the line
	display.Render(editor)

	cursorRow, cursorCol = display.GetCursorPosition()
	t.Logf("Cursor position at Col 15: (%d, %d)", cursorRow, cursorCol)

	// Test with line wrapping disabled
	window.SetLineWrap(false)
	display.Render(editor)

	cursorRow, cursorCol = display.GetCursorPosition()
	t.Logf("Cursor position with wrapping disabled: (%d, %d)", cursorRow, cursorCol)

	visible = window.VisibleLines()
	t.Logf("Visible lines (no wrap): %v", visible)
}

func TestCursorMovementAcrossWrappedLines(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 8)
	
	window := editor.CurrentWindow()
	window.Resize(10, 6)
	window.SetLineWrap(true)

	// Add content that will create multiple wrapped lines
	content := "Line1 Line2 Line3 Line4"
	for _, ch := range content {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}

	// Move cursor to beginning of line
	err := domain.BeginningOfLine(editor)
	if err != nil {
		t.Fatalf("BeginningOfLine failed: %v", err)
	}

	display.Render(editor)
	cursorRow, cursorCol := display.GetCursorPosition()
	t.Logf("After beginning-of-line: cursor at (%d, %d)", cursorRow, cursorCol)

	// Move forward char by char and check positions
	for i := 0; i < 10; i++ {
		err := domain.ForwardChar(editor)
		if err != nil {
			t.Fatalf("ForwardChar failed at step %d: %v", i, err)
		}
		
		display.Render(editor)
		cursorRow, cursorCol = display.GetCursorPosition()
		buffer := editor.CurrentBuffer()
		bufferCursor := buffer.Cursor()
		
		t.Logf("Step %d: buffer cursor (%d,%d), screen cursor (%d,%d)", 
			i+1, bufferCursor.Row, bufferCursor.Col, cursorRow, cursorCol)
	}
}

func TestWrappingToggleCommand(t *testing.T) {
	editor := domain.NewEditor()
	
	window := editor.CurrentWindow()
	
	// Check initial state (should be wrapping enabled)
	if !window.LineWrap() {
		t.Error("Expected line wrap to be enabled by default")
	}

	// Execute toggle command via M-x
	err := domain.ToggleLineWrap(editor)
	if err != nil {
		t.Fatalf("ToggleLineWrap failed: %v", err)
	}

	if window.LineWrap() {
		t.Error("Expected line wrap to be disabled after toggle")
	}

	// Test via minibuffer command execution
	
	// Simulate M-x toggle-truncate-lines
	editor.Minibuffer().StartCommandInput()
	
	// Type the command name
	for _, ch := range "toggle-truncate-lines" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Execute the command
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Should be back to wrapping enabled
	if !window.LineWrap() {
		t.Error("Expected line wrap to be enabled after M-x toggle-truncate-lines")
	}
}