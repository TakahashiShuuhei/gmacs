package test

import (
	"fmt"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestRealisticTerminalScroll(t *testing.T) {
	editor := domain.NewEditor()
	
	// Simulate a realistic terminal size (80x24)
	display := NewMockDisplay(80, 24) // 24 total -> 22 content lines
	
	// Simulate the resize event that happens at startup
	resizeEvent := events.ResizeEventData{Width: 80, Height: 24}
	editor.HandleEvent(resizeEvent) // This will set window to 80x22
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Terminal size: 80x24, Window size: %dx%d", windowWidth, windowHeight)
	
	// Add lines one by one, tracking when scrolling should start
	expectedScrollStart := windowHeight // Should scroll when we exceed window height
	
	for i := 0; i < 30; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		line := "Line " + string(rune('0'+(i%10)))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		scrollTop := window.ScrollTop()
		screenRow, _ := window.CursorPosition()
		
		t.Logf("Line %d: buffer cursor (%d,%d), scroll top: %d, screen row: %d", 
			i, cursor.Row, cursor.Col, scrollTop, screenRow)
		
		// Check when scrolling starts
		if i == expectedScrollStart-1 {
			t.Logf("  -> This should be the last line before scrolling starts")
			if scrollTop != 0 {
				t.Errorf("Line %d: Premature scrolling, scroll top = %d", i, scrollTop)
			}
		}
		
		if i == expectedScrollStart {
			t.Logf("  -> This should trigger scrolling")
			if scrollTop == 0 {
				t.Errorf("Line %d: Expected scrolling to start, but scroll top is still 0", i)
			}
		}
		
		// Cursor should always be visible
		if screenRow < 0 || screenRow >= windowHeight {
			t.Errorf("Line %d: Screen cursor row %d is outside window [0, %d)", 
				i, screenRow, windowHeight)
		}
	}
}

func TestScrollStartsAtRightTime(t *testing.T) {
	// Test with different window sizes to verify scroll timing
	testCases := []struct {
		terminalHeight int
		expectedWindowHeight int
	}{
		{6, 4},   // 6 terminal -> 4 content
		{10, 8},  // 10 terminal -> 8 content  
		{24, 22}, // 24 terminal -> 22 content
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Terminal%d", tc.terminalHeight), func(t *testing.T) {
			editor := domain.NewEditor()
			_ = NewMockDisplay(40, tc.terminalHeight)
			
			// Simulate resize event
			resizeEvent := events.ResizeEventData{Width: 40, Height: tc.terminalHeight}
			editor.HandleEvent(resizeEvent)
			
			window := editor.CurrentWindow()
			_, windowHeight := window.Size()
			
			if windowHeight != tc.expectedWindowHeight {
				t.Fatalf("Expected window height %d, got %d", tc.expectedWindowHeight, windowHeight)
			}
			
			// Add lines up to window height
			for i := 0; i < windowHeight; i++ {
				if i > 0 {
					enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
					editor.HandleEvent(enterEvent)
				}
				
				line := "Line " + string(rune('0'+(i%10)))
				for _, ch := range line {
					event := events.KeyEventData{Rune: ch, Key: string(ch)}
					editor.HandleEvent(event)
				}
			}
			
			// Should not have scrolled yet
			scrollTop := window.ScrollTop()
			if scrollTop != 0 {
				t.Errorf("After %d lines: Expected no scrolling, but scroll top = %d", 
					windowHeight, scrollTop)
			}
			
			// Add one more line - this should trigger scrolling
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
			
			line := "Overflow"
			for _, ch := range line {
				event := events.KeyEventData{Rune: ch, Key: string(ch)}
				editor.HandleEvent(event)
			}
			
			// Now it should have scrolled
			scrollTop = window.ScrollTop()
			if scrollTop == 0 {
				t.Errorf("After overflow line: Expected scrolling, but scroll top is still 0")
			}
			
			// Test the debug command
			err := domain.ShowDebugInfo(editor)
			if err != nil {
				t.Errorf("ShowDebugInfo failed: %v", err)
			}
		})
	}
}