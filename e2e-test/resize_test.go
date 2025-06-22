package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec resize/terminal_resize
 * @scenario ターミナルリサイズ処理
 * @description ターミナルサイズ変更時のウィンドウサイズ更新とコンテンツ保持
 * @given 80x24サイズのターミナルで"hello world"を入力済み
 * @when ターミナルを120x30にリサイズする
 * @then ウィンドウサイズが更新され、コンテンツが保持される
 * @implementation domain/window.go, events/resize_event.go
 */
func TestTerminalResize(t *testing.T) {
	editor := NewEditorWithDefaults()
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
	
	// Check window was resized (height should be terminal height - 2)
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	if windowWidth != 120 || windowHeight != 28 { // 30-2 for mode line and minibuffer
		t.Errorf("Expected window size (120, 28), got (%d, %d)", windowWidth, windowHeight)
	}
	
	// Check display was resized
	displayWidth, displayHeight := display.Size()
	if displayWidth != 120 || displayHeight != 30 {
		t.Errorf("Expected display size (120, 30), got (%d, %d)", displayWidth, displayHeight)
	}
	
	// Re-render and check content is still there
	display.Render(editor)
	content := display.GetContent()
	
	// Check actual display height matches resize
	actualDisplayHeight := len(content)
	if actualDisplayHeight != 30 { // Full display height after resize
		t.Errorf("Expected 30 content lines after resize, got %d", actualDisplayHeight)
	}
	
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "hello world" {
		t.Errorf("Expected content preserved after resize, got %q", actualContent)
	}
}

/**
 * @spec resize/smaller_size_resize
 * @scenario 小さいサイズへのリサイズ
 * @description ターミナルを小さいサイズにリサイズした際のコンテンツ保持
 * @given 80x24サイズで複数行のコンテンツを入力済み
 * @when ターミナルのサイズを40x10に縮小する
 * @then ウィンドウサイズが更新され、バッファの全コンテンツが保持される
 * @implementation domain/window.go, domain/buffer.go
 */
func TestResizeToSmallerSize(t *testing.T) {
	editor := NewEditorWithDefaults()
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
	
	// Check sizes updated (height should be terminal height - 2)
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	if windowWidth != 40 || windowHeight != 8 { // 10-2 for mode line and minibuffer
		t.Errorf("Expected window size (40, 8), got (%d, %d)", windowWidth, windowHeight)
	}
	
	// Re-render and check content fits
	display.Render(editor)
	content := display.GetContent()
	if len(content) != 10 { // Full display height
		t.Errorf("Expected 10 content lines after resize, got %d", len(content))
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

/**
 * @spec resize/multiple_resizes
 * @scenario 連続的なリサイズ操作
 * @description 複数回のリサイズ操作でのサイズ更新とコンテンツ保持
 * @given 80x24サイズで"test content"を入力済み
 * @when 異なるサイズで複数回連続してリサイズする
 * @then 各リサイズ後にサイズが正確に更新され、コンテンツが保持される
 * @implementation domain/window.go, events/resize_event.go
 */
func TestMultipleResizes(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 24)
	
	// Add some content
	for _, ch := range "test content" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Perform multiple resizes (expected window heights will be terminal height - 2)
	sizes := []struct{ width, height, expectedWindowHeight int }{
		{100, 30, 28},
		{60, 20, 18},
		{120, 40, 38},
		{80, 24, 22},
	}
	
	for _, size := range sizes {
		resizeEvent := events.ResizeEventData{Width: size.width, Height: size.height}
		editor.HandleEvent(resizeEvent)
		display.Resize(size.width, size.height)
		
		// Check size was applied
		window := editor.CurrentWindow()
		windowWidth, windowHeight := window.Size()
		if windowWidth != size.width || windowHeight != size.expectedWindowHeight {
			t.Errorf("Expected window size (%d, %d), got (%d, %d)", 
				size.width, size.expectedWindowHeight, windowWidth, windowHeight)
		}
		
		// Render and check content is preserved
		display.Render(editor)
		content := display.GetContent()
		expectedLines := size.height // Full display height
		if len(content) != expectedLines {
			t.Errorf("Expected %d content lines for size %dx%d, got %d", 
				expectedLines, size.width, size.height, len(content))
		}
		
		if len(content) > 0 {
			// Trim trailing spaces for comparison
			actualContent := strings.TrimRight(content[0], " ")
			if actualContent != "test content" {
				t.Errorf("Content not preserved after resize to %dx%d: got %q", 
					size.width, size.height, actualContent)
			}
		}
	}
}

/**
 * @spec resize/cursor_position_preservation
 * @scenario リサイズ後のカーソル位置保持
 * @description ターミナルリサイズ後のカーソル位置保持の検証
 * @given "hello"を入力しカーソルを中央（位置2）に設定
 * @when ターミナルを120x30にリサイズする
 * @then カーソル位置がリサイズ後も(0,2)で保持される
 * @implementation domain/window.go, domain/cursor.go
 */
func TestCursorPositionAfterResize(t *testing.T) {
	editor := NewEditorWithDefaults()
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
	if len(content) > 0 {
		// Trim trailing spaces for comparison
		actualContent := strings.TrimRight(content[0], " ")
		if actualContent != "hello" {
			t.Errorf("Expected content 'hello' after resize, got %q", actualContent)
		}
	}
}