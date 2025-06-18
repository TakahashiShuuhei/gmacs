package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestActualDisplayIssue(t *testing.T) {
	// Test the exact scenario user reported: a, enter, b, enter, c, enter, d, enter
	editor := domain.NewEditor()
	display := NewMockDisplay(120, 12) // Match user's 12-line terminal
	
	// Simulate actual resize event 
	resizeEvent := events.ResizeEventData{Width: 120, Height: 12}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Setup: Terminal 12 lines, Window content area: %dx%d", windowWidth, windowHeight)
	
	// Input: a, enter, b, enter, c, enter, d, enter
	chars := []rune{'a', 'b', 'c', 'd'}
	
	for _, ch := range chars {
		// Input character
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		// Press Enter
		enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
		editor.HandleEvent(enterEvent)
		
		// Debug after each step
		bufferCursor := editor.CurrentBuffer().Cursor()
		screenRow, _ := window.CursorPosition()
		visible := window.VisibleLines()
		
		t.Logf("After '%c' + Enter: cursor buffer (%d,%d), screen row %d, scroll %d", 
			ch, bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
		t.Logf("  Visible lines (%d): %v", len(visible), visible)
	}
	
	display.Render(editor)
	
	// Check final state
	bufferContent := editor.CurrentBuffer().Content()
	visible := window.VisibleLines()
	mockContent := display.GetContent()
	
	t.Logf("Final state:")
	t.Logf("  Buffer content (%d lines): %v", len(bufferContent), bufferContent)
	t.Logf("  Window visible lines (%d): %v", len(visible), visible)
	t.Logf("  MockDisplay content (%d): %v", len(mockContent), mockContent)
	
	// Check for the problem user reported
	// User sees: b, c, d, then lots of empty lines
	// Expected: a, b, c, d, then empty lines only for unused content area
	
	expectedContent := []string{"a", "b", "c", "d", ""}
	for i, expected := range expectedContent {
		if i < len(visible) {
			if visible[i] != expected {
				t.Logf("❌ Line %d: expected '%s', got '%s'", i, expected, visible[i])
			} else {
				t.Logf("✅ Line %d: '%s' correct", i, expected)
			}
		}
	}
	
	// Check if we have excessive empty lines
	nonEmptyLines := 0
	for _, line := range visible {
		if line != "" {
			nonEmptyLines++
		}
	}
	
	t.Logf("Non-empty visible lines: %d out of %d", nonEmptyLines, len(visible))
	
	// The content should fill the beginning of the visible area, not have gaps
	if len(visible) > 0 && visible[0] != "a" {
		t.Errorf("❌ CRITICAL: First visible line should be 'a', got '%s'", visible[0])
		t.Errorf("This matches user's reported issue!")
	}
}

func TestDisplayConsistency(t *testing.T) {
	// Test if MockDisplay and actual Display logic are consistent
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 10)
	
	resizeEvent := events.ResizeEventData{Width: 80, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	
	t.Logf("Window reports content area: %dx%d", windowWidth, windowHeight)
	
	// Add some content
	for i := 0; i < 3; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		ch := rune('a' + i)
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	visible := window.VisibleLines()
	mockContent := display.GetContent()
	
	t.Logf("Window.VisibleLines(): %v", visible)
	t.Logf("MockDisplay.GetContent(): %v", mockContent)
	
	// These should be consistent
	if len(visible) != len(mockContent) {
		t.Errorf("Inconsistency: VisibleLines has %d items, MockDisplay has %d", 
			len(visible), len(mockContent))
	}
	
	for i := 0; i < len(visible) && i < len(mockContent); i++ {
		if visible[i] != mockContent[i] {
			t.Errorf("Line %d inconsistency: VisibleLines='%s', MockDisplay='%s'", 
				i, visible[i], mockContent[i])
		}
	}
}