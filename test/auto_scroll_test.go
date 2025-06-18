package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec scroll/auto_scroll_lines
 * @scenario 行追加時の自動スクロール
 * @description ウィンドウ高を超える行を追加した際の自動スクロール動作
 * @given 40x10サイズのディスプレイ（8コンテンツ行）
 * @when 15行のコンテンツを順次追加する
 * @then カーソルが常に可視範囲内に保たれ、現在の行が表示される
 * @implementation domain/scroll.go, domain/window.go
 */
func TestAutoScrollWhenAddingLines(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // 8 content lines (10-2)
	
	// Set window size to match display content area
	window := editor.CurrentWindow()
	window.Resize(40, 8)

	// Add lines that exceed the window height
	for i := 0; i < 15; i++ {
		if i > 0 {
			// Add newline
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		// Add line content
		line := "Line " + string(rune('0'+(i%10)))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		// After each line addition, render and check cursor visibility
		display.Render(editor)
		
		cursorRow, cursorCol := display.GetCursorPosition()
		buffer := editor.CurrentBuffer()
		bufferCursor := buffer.Cursor()
		
		t.Logf("Line %d: buffer cursor (%d,%d), screen cursor (%d,%d), scroll top: %d", 
			i, bufferCursor.Row, bufferCursor.Col, cursorRow, cursorCol, window.ScrollTop())
		
		// Check that cursor is visible on screen
		if cursorRow < 0 || cursorRow >= 8 { // 8 is content area height
			t.Errorf("Line %d: Cursor row %d is outside visible area (0-7)", i, cursorRow)
		}
		
		// Check that the current line is visible
		visible := window.VisibleLines()
		if len(visible) == 0 {
			t.Errorf("Line %d: No visible lines", i)
			continue
		}
		
		// The last line should contain the content we just added
		expectedContent := "Line " + string(rune('0'+(i%10)))
		found := false
		for _, visibleLine := range visible {
			if visibleLine == expectedContent {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Line %d: Expected content '%s' not found in visible lines: %v", 
				i, expectedContent, visible)
		}
	}
}

/**
 * @spec scroll/auto_scroll_wrapping
 * @scenario 長い行での自動スクロールと行ラップ
 * @description 行ラップ有効時の長い行での自動スクロール動作
 * @given 20x8の小さいウィンドウで行ラップ有効
 * @when 短い行と長い行（ラップする）を混在して追加
 * @then カーソルが常に可視範囲内に保たれる
 * @implementation domain/scroll.go, domain/window.go
 */
func TestAutoScrollWithLongLines(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 8) // Small window
	
	window := editor.CurrentWindow()
	window.Resize(20, 6)
	window.SetLineWrap(true) // Enable line wrapping

	// Add several lines, some short, some long
	lines := []string{
		"Short",
		"This is a very long line that will wrap multiple times",
		"Medium line here",
		"Another very long line that should definitely wrap around multiple times and create several display lines",
		"End",
	}

	for i, line := range lines {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		cursorRow, cursorCol := display.GetCursorPosition()
		buffer := editor.CurrentBuffer()
		bufferCursor := buffer.Cursor()
		
		t.Logf("After line %d ('%s'): buffer cursor (%d,%d), screen cursor (%d,%d), scroll top: %d", 
			i, line, bufferCursor.Row, bufferCursor.Col, cursorRow, cursorCol, window.ScrollTop())
		
		visible := window.VisibleLines()
		t.Logf("Visible lines: %v", visible)
		
		// Cursor should always be visible
		if cursorRow < 0 || cursorRow >= 6 {
			t.Errorf("After line %d: Cursor row %d is outside visible area (0-5)", i, cursorRow)
		}
	}
}

/**
 * @spec scroll/auto_scroll_insertion
 * @scenario テキスト挿入時の自動スクロール
 * @description 可視範囲を超えるテキスト挿入時のスクロール動作
 * @given 30x6の小さいウィンドウ（4コンテンツ行）に3行の初期コンテンツ
 * @when さらに5行の新しいコンテンツを追加
 * @then スクロールが発生し、カーソルが可視範囲内に保たれる
 * @implementation domain/scroll.go, domain/window.go
 */
func TestAutoScrollOnTextInsertion(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(30, 6) // Very small window - 4 content lines
	
	window := editor.CurrentWindow()
	window.Resize(30, 4)

	// Fill up the screen with lines
	for i := 0; i < 3; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		line := "Initial line " + string(rune('A'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	display.Render(editor)
	initialScrollTop := window.ScrollTop()
	t.Logf("Initial state: scroll top = %d", initialScrollTop)
	
	// Now add more lines - this should trigger auto-scroll
	for i := 0; i < 5; i++ {
		enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
		editor.HandleEvent(enterEvent)
		
		line := "New line " + string(rune('1'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		currentScrollTop := window.ScrollTop()
		cursorRow, cursorCol := display.GetCursorPosition()
		buffer := editor.CurrentBuffer()
		bufferCursor := buffer.Cursor()
		
		t.Logf("After new line %d: buffer cursor (%d,%d), screen cursor (%d,%d), scroll top: %d", 
			i, bufferCursor.Row, bufferCursor.Col, cursorRow, cursorCol, currentScrollTop)
		
		// The scroll should have moved if cursor goes beyond visible area
		if bufferCursor.Row >= 4 { // If buffer has more than 4 lines
			if currentScrollTop == initialScrollTop {
				t.Errorf("After line %d: Expected scroll to increase, but it remained at %d", 
					i, initialScrollTop)
			}
		}
		
		// Most importantly, cursor should be visible
		if cursorRow < 0 || cursorRow >= 4 {
			t.Errorf("After line %d: Cursor row %d is outside visible area (0-3)", i, cursorRow)
		}
	}
}

/**
 * @spec scroll/cursor_movement_display
 * @scenario 手動カーソル移動時の表示更新
 * @description 手動でカーソルを移動した際の適切な表示更新
 * @given 30x8ウィンドウに20行のコンテンツを作成
 * @when カーソルを手動でバッファの先頭に移動
 * @then ウィンドウがスクロールしてカーソルが表示される
 * @implementation domain/scroll.go, domain/cursor.go
 */
func TestCursorMovementTriggersDisplay(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(30, 8)
	
	window := editor.CurrentWindow()
	window.Resize(30, 6)
	
	// Add many lines
	for i := 0; i < 20; i++ {
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
	
	// Move cursor to beginning of buffer
	buffer := editor.CurrentBuffer()
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Ensure cursor is visible after manual cursor movement
	domain.EnsureCursorVisible(editor)
	
	display.Render(editor)
	
	// Window should have scrolled to show the cursor
	scrollTop := window.ScrollTop()
	cursorRow, _ := display.GetCursorPosition()
	
	t.Logf("After moving to start: scroll top = %d, cursor row = %d", scrollTop, cursorRow)
	
	if scrollTop != 0 {
		t.Errorf("Expected scroll top to be 0 when cursor at start, got %d", scrollTop)
	}
	
	if cursorRow != 0 {
		t.Errorf("Expected cursor row to be 0, got %d", cursorRow)
	}
}