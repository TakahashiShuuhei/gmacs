package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/events"
)

// TestScrollDelayIssue reproduces the reported scroll delay issue
// User reports: "In 10-line window (8 content + mode + mini), when at line 8 and press Enter,
// should show lines 2-9 but shows 1-8, then next Enter should show 3-10 but shows 1-8,
// then finally next Enter shows 2-9"
func TestScrollDelayIssue(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(40, 10) // 10 total = 8 content + 1 mode + 1 mini
	
	window := editor.CurrentWindow()
	window.Resize(40, 8) // 8 content lines
	
	// Fill exactly 8 lines (0-7) - this should fill screen without scrolling
	for i := 0; i < 8; i++ {
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
	
	bufferCursor := editor.CurrentBuffer().Cursor()
	visible := window.VisibleLines()
	t.Logf("After 8 lines: cursor (%d,%d), scroll %d, visible %v", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop(), visible)
	
	// Should be: cursor (7,6), scroll 0, showing lines 0-7
	if bufferCursor.Row != 7 || window.ScrollTop() != 0 {
		t.Errorf("Initial state wrong: cursor (%d,%d), scroll %d", 
			bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	}
	
	// Step 1: Press Enter at line 7 (end of 8th line)
	// This creates line 8, cursor goes to (8,0)
	// User expects: should scroll to show lines 2-9, but actually shows 1-8
	t.Logf("=== STEP 1: Press Enter at end of line 7 ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	screenRow, _ := window.CursorPosition()
	t.Logf("After 1st Enter: cursor (%d,%d), scroll %d, screen row %d, visible %v", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop(), screenRow, visible)
	
	// According to user report, this should show lines 2-9 but actually shows 1-8
	// Let's see what actually happens
	expectedScroll := 1 // Should scroll to show lines 1-8
	if window.ScrollTop() != expectedScroll {
		t.Errorf("Step 1: Expected scroll %d, got %d", expectedScroll, window.ScrollTop())
	}
	
	// Step 2: Add content and press Enter again  
	// User expects: should show lines 3-10, but actually shows 1-8
	t.Logf("=== STEP 2: Add content and press Enter ===")
	text := "Content"
	for _, ch := range text {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	screenRow, _ = window.CursorPosition()
	t.Logf("After 2nd Enter: cursor (%d,%d), scroll %d, screen row %d, visible %v", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop(), screenRow, visible)
	
	// According to user report, this should show lines 3-10 but shows 1-8
	expectedScroll = 2 // Should scroll to show lines 2-9
	if window.ScrollTop() != expectedScroll {
		t.Errorf("Step 2: Expected scroll %d, got %d", expectedScroll, window.ScrollTop())
	}
	
	// Step 3: Add content and press Enter one more time
	// User says: "Now finally shows 2-9"
	t.Logf("=== STEP 3: Add content and press Enter again ===")
	text = "More"
	for _, ch := range text {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	screenRow, _ = window.CursorPosition()
	t.Logf("After 3rd Enter: cursor (%d,%d), scroll %d, screen row %d, visible %v", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop(), screenRow, visible)
	
	// Check if there's any delay in scrolling
	expectedScroll = 3 // Should scroll to show lines 3-10
	if window.ScrollTop() != expectedScroll {
		t.Errorf("Step 3: Expected scroll %d, got %d", expectedScroll, window.ScrollTop())
	}
}

// TestScrollDelayWithDetailedSteps - more granular testing
func TestScrollDelayWithDetailedSteps(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(40, 10)
	
	window := editor.CurrentWindow()
	window.Resize(40, 8)
	
	// Step-by-step fill and test when scrolling starts
	for i := 0; i < 12; i++ {
		if i > 0 {
			t.Logf("=== Adding line %d ===", i)
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			
			// Check state BEFORE Enter
			bufferCursor := editor.CurrentBuffer().Cursor()
			screenRow, _ := window.CursorPosition()
			t.Logf("BEFORE Enter: cursor (%d,%d), screen row %d, scroll %d", 
				bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
			
			// Apply Enter
			editor.HandleEvent(enterEvent)
			
			// Check state AFTER Enter but before text input
			bufferCursor = editor.CurrentBuffer().Cursor()
			screenRow, _ = window.CursorPosition()
			t.Logf("AFTER Enter: cursor (%d,%d), screen row %d, scroll %d", 
				bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
		}
		
		// Add line content character by character
		line := "Line " + string(rune('0'+(i%10)))
		for j, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
			
			// Only log final state
			if j == len(line)-1 {
				display.Render(editor)
				bufferCursor := editor.CurrentBuffer().Cursor()
				visible := window.VisibleLines()
				screenRow, _ := window.CursorPosition()
				t.Logf("FINAL line %d: cursor (%d,%d), screen row %d, scroll %d, visible count %d", 
					i, bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop(), len(visible))
				
				// Check for unexpected scroll behavior
				if i >= 8 && window.ScrollTop() == 0 {
					t.Errorf("Line %d: Expected scrolling to have started, but scroll is still 0", i)
				}
				
				// Check if cursor is beyond visible area
				if screenRow >= 8 {
					t.Errorf("Line %d: Cursor screen row %d is beyond visible area (0-7)", i, screenRow)
				} else if screenRow < 0 {
					t.Errorf("Line %d: Cursor screen row %d is negative", i, screenRow)
				}
			}
		}
	}
}