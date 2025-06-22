package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestSimpleAutoScroll(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	window := editor.CurrentWindow()
	window.Resize(20, 3) // Very small window for testing
	
	// Add 6 lines (more than window height)
	for i := 0; i < 6; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		scrollTop := window.ScrollTop()
		
		t.Logf("After line %d: cursor at buffer (%d,%d), scroll top: %d", 
			i, cursor.Row, cursor.Col, scrollTop)
		
		// Cursor should always be visible
		_, windowHeight := window.Size()
		if cursor.Row >= scrollTop+windowHeight || cursor.Row < scrollTop {
			t.Errorf("Line %d: cursor at buffer row %d is outside scroll window [%d, %d)", 
				i, cursor.Row, scrollTop, scrollTop+windowHeight)
		}
	}
}

func TestManualCursorMovement(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	window := editor.CurrentWindow()
	window.Resize(20, 3)
	
	// Add 10 lines
	for i := 0; i < 10; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	buffer := editor.CurrentBuffer()
	
	// Move cursor to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Check cursor position before EnsureCursorVisible
	screenRow, screenCol := window.CursorPosition()
	t.Logf("Before EnsureCursorVisible: screen cursor (%d, %d), buffer cursor (0, 0)", screenRow, screenCol)
	
	domain.EnsureCursorVisible(editor)
	
	// Check cursor position after EnsureCursorVisible
	screenRow, screenCol = window.CursorPosition()
	scrollTop := window.ScrollTop()
	t.Logf("After EnsureCursorVisible: screen cursor (%d, %d), scroll top = %d", screenRow, screenCol, scrollTop)
	
	if scrollTop != 0 {
		t.Errorf("Expected scroll top to be 0 when cursor at start, got %d", scrollTop)
	}
	
	// Move cursor to middle
	buffer.SetCursor(domain.Position{Row: 5, Col: 0})
	domain.EnsureCursorVisible(editor)
	
	scrollTop = window.ScrollTop()
	t.Logf("After moving to middle (row 5): scroll top = %d", scrollTop)
	
	// Should scroll to show row 5
	_, windowHeight := window.Size()
	if scrollTop > 5 || scrollTop+windowHeight <= 5 {
		t.Errorf("Row 5 should be visible with scroll top %d and window height %d", 
			scrollTop, windowHeight)
	}
	
	// Move cursor to end
	buffer.SetCursor(domain.Position{Row: 9, Col: 0})
	domain.EnsureCursorVisible(editor)
	
	scrollTop = window.ScrollTop()
	t.Logf("After moving to end (row 9): scroll top = %d", scrollTop)
	
	// Should scroll to show row 9
	_, windowHeight = window.Size()
	if scrollTop > 9 || scrollTop+windowHeight <= 9 {
		t.Errorf("Row 9 should be visible with scroll top %d and window height %d", 
			scrollTop, windowHeight)
	}
}