package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestTerminalResize(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 24)
	
	// Check initial size
	width, height := display.Size()
	if width != 80 || height != 24 {
		t.Errorf("Expected initial size (80, 24), got (%d, %d)", width, height)
	}
	
	// Initial content
	for _, ch := range "hello world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	// Resize to larger size
	resizeEvent := events.ResizeEventData{Width: 120, Height: 30}
	editor.HandleEvent(resizeEvent)
	display.Resize(120, 30)
	
	// Check window was resized
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	if windowWidth != 120 || windowHeight != 30 {
		t.Errorf("Expected window size (120, 30), got (%d, %d)", windowWidth, windowHeight)
	}
	
	// Check display was resized
	displayWidth, displayHeight := display.Size()
	if displayWidth != 120 || displayHeight != 30 {
		t.Errorf("Expected display size (120, 30), got (%d, %d)", displayWidth, displayHeight)
	}
	
	// Re-render and check content is still there
	display.Render(editor)
	content := display.GetContent()
	if len(content) != 28 { // height-2 = 28
		t.Errorf("Expected 28 content lines after resize, got %d", len(content))
	}
	
	if content[0] != "hello world" {
		t.Errorf("Expected content preserved after resize, got %q", content[0])
	}
}

func TestResizeToSmallerSize(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 24)
	
	// Add multiple lines of content
	lines := []string{"line1", "line2", "line3", "line4", "line5"}
	for i, line := range lines {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	display.Render(editor)
	
	// Resize to much smaller size
	resizeEvent := events.ResizeEventData{Width: 40, Height: 10}
	editor.HandleEvent(resizeEvent)
	display.Resize(40, 10)
	
	// Check sizes updated
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	if windowWidth != 40 || windowHeight != 10 {
		t.Errorf("Expected window size (40, 10), got (%d, %d)", windowWidth, windowHeight)
	}
	
	// Re-render and check content fits
	display.Render(editor)
	content := display.GetContent()
	if len(content) != 8 { // height-2 = 8
		t.Errorf("Expected 8 content lines after resize, got %d", len(content))
	}
	
	// Content should be preserved but may be scrolled
	buffer := editor.CurrentBuffer()
	bufferContent := buffer.Content()
	if len(bufferContent) != 5 {
		t.Errorf("Expected 5 lines in buffer, got %d", len(bufferContent))
	}
	
	// Check that all original lines are still in buffer
	for i, expectedLine := range lines {
		if i < len(bufferContent) && bufferContent[i] != expectedLine {
			t.Errorf("Expected line %d to be %q, got %q", i, expectedLine, bufferContent[i])
		}
	}
}

func TestMultipleResizes(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 24)
	
	// Add some content
	for _, ch := range "test content" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Perform multiple resizes
	sizes := []struct{ width, height int }{
		{100, 30},
		{60, 20},
		{120, 40},
		{80, 24},
	}
	
	for _, size := range sizes {
		resizeEvent := events.ResizeEventData{Width: size.width, Height: size.height}
		editor.HandleEvent(resizeEvent)
		display.Resize(size.width, size.height)
		
		// Check size was applied
		window := editor.CurrentWindow()
		windowWidth, windowHeight := window.Size()
		if windowWidth != size.width || windowHeight != size.height {
			t.Errorf("Expected window size (%d, %d), got (%d, %d)", 
				size.width, size.height, windowWidth, windowHeight)
		}
		
		// Render and check content is preserved
		display.Render(editor)
		content := display.GetContent()
		expectedLines := size.height - 2
		if len(content) != expectedLines {
			t.Errorf("Expected %d content lines for size %dx%d, got %d", 
				expectedLines, size.width, size.height, len(content))
		}
		
		if len(content) > 0 && content[0] != "test content" {
			t.Errorf("Content not preserved after resize to %dx%d: got %q", 
				size.width, size.height, content[0])
		}
	}
}

func TestCursorPositionAfterResize(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 24)
	
	// Add content and position cursor
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to beginning
	buffer := editor.CurrentBuffer()
	buffer.SetCursor(domain.Position{Row: 0, Col: 2}) // Middle of "hello"
	
	display.Render(editor)
	
	// Check initial cursor position
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 0 || cursorCol != 2 {
		t.Errorf("Expected initial cursor at (0, 2), got (%d, %d)", cursorRow, cursorCol)
	}
	
	// Resize window
	resizeEvent := events.ResizeEventData{Width: 120, Height: 30}
	editor.HandleEvent(resizeEvent)
	display.Resize(120, 30)
	
	// Re-render and check cursor position is preserved
	display.Render(editor)
	cursorRow, cursorCol = display.GetCursorPosition()
	if cursorRow != 0 || cursorCol != 2 {
		t.Errorf("Expected cursor preserved at (0, 2) after resize, got (%d, %d)", cursorRow, cursorCol)
	}
	
	// Content should still be correct
	content := display.GetContent()
	if len(content) > 0 && content[0] != "hello" {
		t.Errorf("Expected content 'hello' after resize, got %q", content[0])
	}
}