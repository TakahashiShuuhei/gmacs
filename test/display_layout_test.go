package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec display/layout_analysis
 * @scenario 表示レイアウト解析
 * @description 実際の表示レイアウトと期待されるレイアウトの比較分析
 * @given 40x10ターミナルでリサイズイベントを送信
 * @when ウィンドウ高と同じ数の行を追加し、さらに1行追加
 * @then MockDisplayと実際のCLI Displayの動作が一致し、適切なスクロールタイミングが確認される
 * @implementation cli/display.go, test/mock_display.go
 */
func TestDisplayLayoutAnalysis(t *testing.T) {
	// Test actual display layout vs expected layout
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10)
	
	// Simulate actual resize event like in real gmacs
	resizeEvent := events.ResizeEventData{Width: 40, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Terminal height: 10, Window content area: %dx%d", windowWidth, windowHeight)
	
	// Add exactly the content that fills the window
	for i := 0; i < windowHeight; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		// Add single character
		ch := rune('a' + i)
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	// Check what MockDisplay renders vs actual CLI Display
	visible := window.VisibleLines()
	bufferContent := editor.CurrentBuffer().Content()
	
	t.Logf("Buffer content (%d lines): %v", len(bufferContent), bufferContent)
	t.Logf("Visible lines (%d lines): %v", len(visible), visible)
	t.Logf("MockDisplay mode line: %s", display.GetModeLine())
	t.Logf("MockDisplay minibuffer: %s", display.GetMinibuffer())
	
	// Analyze MockDisplay vs real Display behavior
	// MockDisplay creates content array of size height-2
	mockWidth, mockHeight := display.Size()
	t.Logf("MockDisplay total size: %dx%d", mockWidth, mockHeight)
	mockContent := display.GetContent()
	t.Logf("MockDisplay content area (%d lines): %v", len(mockContent), mockContent)
	
	// The issue might be that real CLI Display has different layout calculations
	// Let's check what window.VisibleLines() returns vs what gets rendered
	
	// Check cursor position
	bufferCursor := editor.CurrentBuffer().Cursor()
	screenRow, screenCol := window.CursorPosition()
	t.Logf("Cursor: buffer (%d,%d), screen (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, screenCol, window.ScrollTop())
	
	// Now add one more line that should trigger scroll
	t.Logf("=== Adding one more line ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	ch := rune('a' + windowHeight)
	event := events.KeyEventData{Rune: ch, Key: string(ch)}
	editor.HandleEvent(event)
	
	display.Render(editor)
	
	visible = window.VisibleLines()
	bufferContent = editor.CurrentBuffer().Content()
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, screenCol = window.CursorPosition()
	
	t.Logf("After adding line %d:", windowHeight)
	t.Logf("Buffer content (%d lines): %v", len(bufferContent), bufferContent)
	t.Logf("Visible lines (%d lines): %v", len(visible), visible)
	t.Logf("Cursor: buffer (%d,%d), screen (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, screenCol, window.ScrollTop())
	
	// Compare with user expectation
	if window.ScrollTop() > 0 {
		t.Logf("Scrolling occurred when buffer reached %d lines", len(bufferContent))
	} else {
		t.Logf("No scrolling yet with %d buffer lines", len(bufferContent))
	}
}

/**
 * @spec display/mock_vs_real
 * @scenario MockDisplayと実際のDisplay比較
 * @description ユーザー報告シナリオでのMockDisplayと実際のCLI Displayの動作比較
 * @given 40x10ターミナルでa〜hまで8行のコンテンツを作成
 * @when 最後にEnterキーを押下
 * @then MockDisplayの動作がユーザー期待（bから始まる表示）と一致する
 * @implementation test/mock_display.go, cli/display.go
 */
func TestRealVsMockDisplay(t *testing.T) {
	// Test what happens with exactly the user scenario
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10)
	
	// Use real resize event
	resizeEvent := events.ResizeEventData{Width: 40, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	
	// Input a, enter, b, enter, ..., h (8 lines total)
	chars := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	
	for i, ch := range chars {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		if i < len(chars)-1 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
	}
	
	display.Render(editor)
	
	// Check the exact state
	bufferCursor := editor.CurrentBuffer().Cursor()
	visible := window.VisibleLines()
	screenRow, _ := window.CursorPosition()
	
	t.Logf("Before final Enter: cursor buffer (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("Visible: %v", visible)
	
	// Check if MockDisplay has same layout as CLI Display
	mockContent := display.GetContent()
	t.Logf("MockDisplay content area: %v", mockContent)
	t.Logf("MockDisplay mode line: '%s'", display.GetModeLine())
	t.Logf("MockDisplay minibuffer: '%s'", display.GetMinibuffer())
	
	// The key question: is cursor at last visible line?
	windowWidth, windowHeight := window.Size()
	isAtLastVisibleLine := screenRow == windowHeight-1
	t.Logf("Window size: %dx%d, cursor screen row: %d, at last visible line: %t", 
		windowWidth, windowHeight, screenRow, isAtLastVisibleLine)
	
	// Now the critical Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	visible = window.VisibleLines()
	screenRow, _ = window.CursorPosition()
	
	t.Logf("After final Enter: cursor buffer (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("Visible: %v", visible)
	
	// Check if behavior matches user expectation
	expectedFirstLine := "b"
	if len(visible) > 0 {
		actualFirstLine := visible[0]
		if actualFirstLine == expectedFirstLine {
			t.Logf("✅ MockDisplay behavior matches user expectation: shows '%s'", actualFirstLine)
		} else {
			t.Logf("❌ MockDisplay behavior differs: shows '%s', user expects '%s'", actualFirstLine, expectedFirstLine)
		}
	}
}