package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestExactUserScenario(t *testing.T) {
	// Setup: Height 10 terminal (8 content + 1 mode + 1 mini)
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // Height 10 exactly as user specified
	
	// Simulate actual resize event like in real gmacs
	resizeEvent := events.ResizeEventData{Width: 40, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("After resize event: window size %dx%d", windowWidth, windowHeight)
	
	// Step 1: Input a, enter, b, enter, ..., h
	// This creates 8 lines: a, b, c, d, e, f, g, h
	chars := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	
	for i, ch := range chars {
		// Input the character
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		// Press Enter (except for the last one initially)
		if i < len(chars)-1 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
	}
	
	display.Render(editor)
	
	// Check state after inputting a through h with enters
	bufferCursor := editor.CurrentBuffer().Cursor()
	visible := window.VisibleLines()
	bufferContent := editor.CurrentBuffer().Content()
	
	t.Logf("After a-h input: cursor (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	t.Logf("Buffer content: %v", bufferContent)
	t.Logf("Visible lines: %v", visible)
	
	// User expectation: should show lines a ~ h, mode line, minibuffer
	// Cursor should be at (7,1) after 'h', scroll should be 0
	if bufferCursor.Row != 7 || bufferCursor.Col != 1 {
		t.Errorf("After a-h input: expected cursor (7,1), got (%d,%d)", 
			bufferCursor.Row, bufferCursor.Col)
	}
	
	if window.ScrollTop() != 0 {
		t.Errorf("After a-h input: expected scroll 0, got %d", window.ScrollTop())
	}
	
	// Should show exactly: [a, b, c, d, e, f, g, h]
	expectedVisible := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	if len(visible) != len(expectedVisible) {
		t.Errorf("After a-h input: expected %d visible lines, got %d", 
			len(expectedVisible), len(visible))
	} else {
		for i, expected := range expectedVisible {
			if i < len(visible) && visible[i] != expected {
				t.Errorf("After a-h input: line %d expected '%s', got '%s'", 
					i, expected, visible[i])
			}
		}
	}
	
	// Step 2: Press Enter while cursor is at end of line h
	// User expectation: should scroll to show b ~ h, empty line, mode, mini
	t.Logf("=== Pressing Enter at end of line h ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	// Check state after Enter
	bufferCursor = editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	bufferContent = editor.CurrentBuffer().Content()
	
	t.Logf("After Enter: cursor (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	t.Logf("Buffer content: %v", bufferContent)
	t.Logf("Visible lines: %v", visible)
	
	// User expectation: cursor should be at (8,0), scroll should be 1
	// Should show: [b, c, d, e, f, g, h, ""]
	if bufferCursor.Row != 8 || bufferCursor.Col != 0 {
		t.Errorf("After Enter: expected cursor (8,0), got (%d,%d)", 
			bufferCursor.Row, bufferCursor.Col)
	}
	
	expectedScroll := 1
	if window.ScrollTop() != expectedScroll {
		t.Errorf("After Enter: expected scroll %d, got %d", 
			expectedScroll, window.ScrollTop())
	}
	
	// Should show exactly: [b, c, d, e, f, g, h, ""]
	expectedVisibleAfterEnter := []string{"b", "c", "d", "e", "f", "g", "h", ""}
	if len(visible) != len(expectedVisibleAfterEnter) {
		t.Errorf("After Enter: expected %d visible lines, got %d", 
			len(expectedVisibleAfterEnter), len(visible))
	} else {
		for i, expected := range expectedVisibleAfterEnter {
			if i < len(visible) && visible[i] != expected {
				t.Errorf("After Enter: line %d expected '%s', got '%s'", 
					i, expected, visible[i])
			}
		}
	}
	
	// The critical test: first visible line should be 'b'
	if len(visible) > 0 && visible[0] != "b" {
		t.Errorf("CRITICAL FAILURE: After Enter, expected first visible line 'b', got '%s'", 
			visible[0])
		t.Errorf("This is the exact issue user reported!")
	} else if len(visible) > 0 && visible[0] == "b" {
		t.Logf("SUCCESS: Shows 'b' as first line as expected by user")
	}
}

// Test the exact scenario step by step for debugging
func TestUserScenarioStepByStep(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10)
	
	window := editor.CurrentWindow()
	window.Resize(40, 8)
	
	// Add content step by step and verify each step
	chars := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	
	for i, ch := range chars {
		// Input character
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		bufferCursor := editor.CurrentBuffer().Cursor()
		t.Logf("After '%c': cursor (%d,%d), scroll %d", 
			ch, bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
		
		// Press Enter (except for last)
		if i < len(chars)-1 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
			
			bufferCursor = editor.CurrentBuffer().Cursor()
			t.Logf("After Enter: cursor (%d,%d), scroll %d", 
				bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
		}
	}
	
	display.Render(editor)
	visible := window.VisibleLines()
	t.Logf("Final state before critical Enter: visible %v", visible)
	
	// Now the critical Enter
	t.Logf("=== CRITICAL ENTER ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor := editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	screenRow, _ := window.CursorPosition()
	
	t.Logf("CRITICAL RESULT: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("CRITICAL RESULT: visible %v", visible)
	
	// Check if this matches user expectation
	if len(visible) > 0 {
		if visible[0] == "b" {
			t.Logf("✅ CORRECT: First visible line is 'b'")
		} else {
			t.Logf("❌ WRONG: First visible line is '%s', expected 'b'", visible[0])
		}
	}
}