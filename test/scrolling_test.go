package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestVerticalScrolling(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // Small window for testing
	
	// Set window size to match display content area (height-2)
	window := editor.CurrentWindow()
	window.Resize(40, 8)

	// Add many lines of content
	for i := 0; i < 20; i++ {
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

	// Check initial scroll position - should be auto-scrolled to keep cursor visible
	// With 20 lines (0-19) and window height 8, cursor at line 19 should result in scroll position 12
	expectedScrollTop := 12 // 19 (cursor row) - 8 (window height) + 1 = 12
	if window.ScrollTop() != expectedScrollTop {
		t.Errorf("Expected initial scroll top to be %d, got %d", expectedScrollTop, window.ScrollTop())
	}

	// Scroll down and check
	window.SetScrollTop(5)
	display.Render(editor)

	visible := window.VisibleLines()
	if len(visible) == 0 {
		t.Fatal("No visible lines after scroll")
	}

	// Should show lines starting from line 5
	if visible[0] != "Line 5" {
		t.Errorf("Expected first visible line to be 'Line 5', got %q", visible[0])
	}
}

func TestHorizontalScrolling(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5) // Very narrow window
	
	// Set window size to match display content area (height-2)
	window := editor.CurrentWindow()
	window.Resize(10, 3)

	// Add a very long line
	longLine := "This is a very long line that should exceed the window width and require horizontal scrolling"
	for _, ch := range longLine {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}

	// Disable line wrapping to enable horizontal scrolling
	window.SetLineWrap(false)

	display.Render(editor)

	// Test horizontal scrolling
	window.SetScrollLeft(5)
	display.Render(editor)

	visible := window.VisibleLines()
	if len(visible) == 0 {
		t.Fatal("No visible lines after horizontal scroll")
	}

	// The visible line should be a substring starting from position 5
	expectedSubstring := longLine[5:] // Skip first 5 characters
	if len(expectedSubstring) > 10 {
		expectedSubstring = expectedSubstring[:10] // Truncate to window width
	}

	if visible[0] != expectedSubstring {
		t.Errorf("Expected visible line after horizontal scroll to be %q, got %q", expectedSubstring, visible[0])
	}
}

func TestLineWrapping(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5) // Small window
	
	// Set window size to match display content area (height-2)
	window := editor.CurrentWindow()
	window.Resize(10, 3)

	// Add a long line
	longLine := "This is a very long line"
	for _, ch := range longLine {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}

	// Test with line wrapping enabled (default)
	window.SetLineWrap(true)
	display.Render(editor)

	visible := window.VisibleLines()
	if len(visible) < 2 {
		t.Errorf("Expected line to be wrapped into multiple visible lines, got %d lines", len(visible))
	}

	// Test with line wrapping disabled
	window.SetLineWrap(false)
	display.Render(editor)

	visible = window.VisibleLines()
	if len(visible) != 1 {
		t.Errorf("Expected single visible line when wrapping disabled, got %d lines", len(visible))
	}
}

func TestToggleLineWrap(t *testing.T) {
	editor := domain.NewEditor()

	window := editor.CurrentWindow()

	// Check initial state (should be enabled by default)
	if !window.LineWrap() {
		t.Error("Expected line wrap to be enabled by default")
	}

	// Toggle line wrap using the command
	err := domain.ToggleLineWrap(editor)
	if err != nil {
		t.Fatalf("ToggleLineWrap failed: %v", err)
	}

	if window.LineWrap() {
		t.Error("Expected line wrap to be disabled after toggle")
	}

	// Toggle again
	err = domain.ToggleLineWrap(editor)
	if err != nil {
		t.Fatalf("ToggleLineWrap failed: %v", err)
	}

	if !window.LineWrap() {
		t.Error("Expected line wrap to be enabled after second toggle")
	}
}

func TestPageUpDown(t *testing.T) {
	editor := domain.NewEditor()

	// Add many lines
	for i := 0; i < 50; i++ {
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

	window := editor.CurrentWindow()
	initialScroll := window.ScrollTop()

	// Test page down
	err := domain.PageDown(editor)
	if err != nil {
		t.Fatalf("PageDown failed: %v", err)
	}

	if window.ScrollTop() <= initialScroll {
		t.Errorf("Expected scroll position to increase after PageDown, was %d, now %d", initialScroll, window.ScrollTop())
	}

	// Test page up
	scrollAfterPageDown := window.ScrollTop()
	err = domain.PageUp(editor)
	if err != nil {
		t.Fatalf("PageUp failed: %v", err)
	}

	if window.ScrollTop() >= scrollAfterPageDown {
		t.Errorf("Expected scroll position to decrease after PageUp, was %d, now %d", scrollAfterPageDown, window.ScrollTop())
	}
}

func TestScrollCommands(t *testing.T) {
	editor := domain.NewEditor()

	// Add some content
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
	}

	window := editor.CurrentWindow()
	initialScroll := window.ScrollTop()

	// Test scroll down
	err := domain.ScrollDown(editor)
	if err != nil {
		t.Fatalf("ScrollDown failed: %v", err)
	}

	if window.ScrollTop() != initialScroll+1 {
		t.Errorf("Expected scroll top to be %d after ScrollDown, got %d", initialScroll+1, window.ScrollTop())
	}

	// Test scroll up
	err = domain.ScrollUp(editor)
	if err != nil {
		t.Fatalf("ScrollUp failed: %v", err)
	}

	if window.ScrollTop() != initialScroll {
		t.Errorf("Expected scroll top to be %d after ScrollUp, got %d", initialScroll, window.ScrollTop())
	}
}