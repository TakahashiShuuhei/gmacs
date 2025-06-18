package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestScrollTimingIssue(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // Height 10 = 8 content + 1 mode line + 1 minibuffer
	
	// Set window content area size (height - 2 for mode line and minibuffer)
	window := editor.CurrentWindow()
	window.Resize(40, 8) // Content area is 8 lines (0-7)
	
	// Step 1: Fill 8 lines exactly (lines 0-7, cursor at end of line 7)
	for i := 0; i < 8; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		// Add line content
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	display.Render(editor)
	
	bufferCursor := editor.CurrentBuffer().Cursor()
	t.Logf("After filling 8 lines: buffer cursor (%d,%d), scroll top: %d", bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	
	visible := window.VisibleLines()
	t.Logf("Visible lines: %v", visible)
	
	// Should show lines 0-7, cursor at (7,6), scroll = 0
	if window.ScrollTop() != 0 {
		t.Errorf("Step 1: Expected scroll top 0, got %d", window.ScrollTop())
	}
	if bufferCursor.Row != 7 {
		t.Errorf("Step 1: Expected cursor row 7, got %d", bufferCursor.Row)
	}
	
	// Step 2: Press Enter on line 7 (8th visible line) - this creates line 8
	// CRITICAL: Cursor goes to (8,0) which is beyond visible area [0-7], should scroll immediately
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	t.Logf("After Enter on line 7: buffer cursor (%d,%d), scroll top: %d", bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	
	visible = window.VisibleLines()
	t.Logf("Visible lines: %v", visible)
	
	// Critical test: cursor at (8,0) should trigger scroll to 1, showing lines 1-8
	if bufferCursor.Row != 8 {
		t.Errorf("Step 2: Expected cursor row 8, got %d", bufferCursor.Row)
	}
	
	// This is the key assertion - user says this fails (scroll should be 1, not 0)
	if window.ScrollTop() != 1 {
		t.Errorf("Step 2: Expected scroll top 1 (to show lines 1-8), got %d", window.ScrollTop())
	}
	
	// Cursor should be at screen row 7 (bottom of visible area)
	screenRow, _ := window.CursorPosition()
	if screenRow != 7 {
		t.Errorf("Step 2: Expected cursor screen row 7, got %d", screenRow)
	}
	
	// Visible lines should be 1-8 (8 lines total)
	if len(visible) != 8 {
		t.Errorf("Step 2: Expected 8 visible lines, got %d", len(visible))
	}
	if len(visible) > 0 && visible[0] != "Line 1" {
		t.Errorf("Step 2: Expected first visible line to be 'Line 1', got '%s'", visible[0])
	}
	
	// Step 3: Add some content and press Enter again
	text := "Content"
	for _, ch := range text {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	t.Logf("After adding content and Enter: buffer cursor (%d,%d), scroll top: %d", bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	
	visible = window.VisibleLines()
	t.Logf("Visible lines: %v", visible)
	
	// Now cursor should be at (9,0) and scroll should be 2 to show lines 2-9
	if bufferCursor.Row != 9 {
		t.Errorf("Step 3: Expected cursor row 9, got %d", bufferCursor.Row)
	}
	
	if window.ScrollTop() != 2 {
		t.Errorf("Step 3: Expected scroll top 2 (to show lines 2-9), got %d", window.ScrollTop())
	}
	
	screenRow, _ = window.CursorPosition()
	if screenRow != 7 {
		t.Errorf("Step 3: Expected cursor screen row 7, got %d", screenRow)
	}
}