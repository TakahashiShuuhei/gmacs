package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestScrollTimingProblem(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // 10 total height -> 8 content lines (10-2)
	
	// Set window size to match display content area
	window := editor.CurrentWindow()
	window.Resize(40, 8) // Content area height
	
	t.Logf("Display total height: 10, content area height: 8")
	windowWidth, windowHeight := window.Size()
	t.Logf("Window size: %dx%d", windowWidth, windowHeight)
	
	// Add lines one by one and track when scrolling starts
	for i := 0; i < 12; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		scrollTop := window.ScrollTop()
		screenRow, _ := window.CursorPosition()
		
		// Get visible content
		content := display.GetContent()
		
		t.Logf("Line %d: buffer cursor (%d,%d), scroll top: %d, screen row: %d", 
			i, cursor.Row, cursor.Col, scrollTop, screenRow)
		t.Logf("  Visible content lines: %d, content: %v", len(content), content)
		
		// Check if scrolling should have started
		expectedScrollStart := 8 // Should start scrolling when we reach line 8 (9th line)
		if i >= expectedScrollStart && scrollTop == 0 {
			t.Errorf("Line %d: Expected scrolling to start, but scroll top is still 0", i)
		}
		
		// Check if cursor is actually visible in the rendered content
		if screenRow >= len(content) {
			t.Errorf("Line %d: Screen cursor row %d is beyond visible content (length %d)", 
				i, screenRow, len(content))
		}
	}
}

func TestWindowHeightCalculation(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10)
	
	// Check default window size
	window := editor.CurrentWindow()
	width, height := window.Size()
	t.Logf("Default window size: %dx%d", width, height)
	
	// Simulate resize event (this is what happens in real usage)
	resizeEvent := events.ResizeEventData{Width: 40, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	width, height = window.Size()
	t.Logf("After resize event: window size: %dx%d", width, height)
	
	// Check display content area
	display.Render(editor)
	content := display.GetContent()
	t.Logf("Display content area lines: %d", len(content))
	
	// The window height should match the display content area
	if height != len(content) {
		t.Errorf("Window height (%d) doesn't match display content area (%d)", height, len(content))
	}
}

func TestScrollCondition(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 6) // Small display: 6 total -> 4 content lines
	
	window := editor.CurrentWindow()
	window.Resize(40, 4) // 4 content lines
	
	// Add exactly 4 lines (should fill screen without scrolling)
	for i := 0; i < 4; i++ {
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
	
	display.Render(editor)
	
	buffer := editor.CurrentBuffer()
	cursor := buffer.Cursor()
	scrollTop := window.ScrollTop()
	screenRow, _ := window.CursorPosition()
	content := display.GetContent()
	
	t.Logf("After 4 lines: buffer cursor (%d,%d), scroll top: %d, screen row: %d", 
		cursor.Row, cursor.Col, scrollTop, screenRow)
	t.Logf("Content: %v", content)
	
	// Should not have scrolled yet
	if scrollTop != 0 {
		t.Errorf("Should not scroll with 4 lines in 4-line window, but scroll top = %d", scrollTop)
	}
	
	// Add 5th line - this should trigger scrolling
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	line := "Line 4"
	for _, ch := range line {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	cursor = buffer.Cursor()
	scrollTop = window.ScrollTop()
	screenRow, _ = window.CursorPosition()
	content = display.GetContent()
	
	t.Logf("After 5th line: buffer cursor (%d,%d), scroll top: %d, screen row: %d", 
		cursor.Row, cursor.Col, scrollTop, screenRow)
	t.Logf("Content: %v", content)
	
	// Now it should have scrolled
	if scrollTop == 0 {
		t.Errorf("Should have scrolled after 5th line, but scroll top is still 0")
	}
	
	// Cursor should be at bottom of visible area
	if screenRow != 3 { // Bottom line of 4-line window (0-indexed)
		t.Errorf("Expected cursor at screen row 3, got %d", screenRow)
	}
}