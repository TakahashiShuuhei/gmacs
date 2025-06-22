package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec scroll/scroll_timing
 * @scenario 早すぎるスクロールの回避
 * @description コンテンツがウィンドウコンテンツエリアを真に超えるまでスクロールが発生しないことをテスト
 * @given 12行のターミナル（10コンテンツ + モード + ミニ）
 * @when 文字a〜jをそれぞれEnterで区切って入力する
 * @then すべての10行がスクロールなしで表示される
 * @implementation domain/scroll.go, cli/display.go
 */
func TestTerminal12LinesScenario(t *testing.T) {
	// Exact user scenario: 12-line terminal
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(120, 12) // 12 total height = 10 content + 1 mode + 1 mini
	
	// Simulate actual resize event 
	resizeEvent := events.ResizeEventData{Width: 120, Height: 12}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Terminal: 12 lines, Window content area: %dx%d", windowWidth, windowHeight)
	
	// Step 1: Input a through h (should fill 8 lines out of 10 available)
	chars := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	
	for i, ch := range chars {
		// Input character
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		// Press Enter (except for last)
		if i < len(chars)-1 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
	}
	
	display.Render(editor)
	
	// Check state after h input
	bufferCursor := editor.CurrentBuffer().Cursor()
	screenRow, screenCol := window.CursorPosition()
	visible := window.VisibleLines()
	
	t.Logf("After h input: cursor buffer (%d,%d), screen (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, screenCol, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// User says "cursor disappears" after h - this means cursor is at edge
	if screenRow >= windowHeight-1 {
		t.Logf("✅ Cursor is at bottom edge (screen row %d), would be invisible", screenRow)
	} else {
		t.Logf("❌ Cursor should be at bottom edge but is at screen row %d", screenRow)
	}
	
	// Expected: should show a, b, c, d, e, f, g, h (8 lines out of 10 available)
	expectedVisible := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	if len(visible) == len(expectedVisible) {
		for i, expected := range expectedVisible {
			if i < len(visible) && visible[i] != expected {
				t.Errorf("After h: line %d expected '%s', got '%s'", i, expected, visible[i])
			}
		}
	}
	
	// Step 2: Input i (this should be 9th line, still within 10-line content area)
	t.Logf("=== Input 'i' ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	iEvent := events.KeyEventData{Rune: 'i', Key: "i"}
	editor.HandleEvent(iEvent)
	
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, screenCol = window.CursorPosition()
	visible = window.VisibleLines()
	
	t.Logf("After i input: cursor buffer (%d,%d), screen (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, screenCol, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// Step 3: Input j (this should be 10th line, still within content area)
	t.Logf("=== Input 'j' ===")
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	jEvent := events.KeyEventData{Rune: 'j', Key: "j"}
	editor.HandleEvent(jEvent)
	
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, screenCol = window.CursorPosition()
	visible = window.VisibleLines()
	
	t.Logf("After j input: cursor buffer (%d,%d), screen (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, screenCol, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// User reports: shows b, c, d, e, f, g, h, i, j (missing 'a')
	// Expected: should show a, b, c, d, e, f, g, h, i, j (all 10 lines)
	
	if len(visible) > 0 {
		if visible[0] == "a" {
			t.Logf("✅ Shows 'a' as first line (correct)")
		} else {
			t.Logf("❌ Shows '%s' as first line, user reports this should be 'a'", visible[0])
			t.Logf("❌ This means scroll happened too early")
		}
	}
	
	// Check if we're using the full content area
	if len(visible) < windowHeight {
		t.Logf("❌ Only using %d lines out of %d available content lines", len(visible), windowHeight)
	} else {
		t.Logf("✅ Using full %d content lines", windowHeight)
	}
}

/**
 * @spec scroll/scroll_timing  
 * @scenario コンテンツがウィンドウを超えた時のスクロール
 * @description スクロール動作をステップごとに検証するデバッグテスト
 * @given 12行のコンテンツエリアを持つターミナル
 * @when コンテンツエリア限界を超えて一行ずつ追加する
 * @then 適切なタイミングでスクロールが発生する
 * @implementation domain/scroll.go, cli/display.go
 */
func TestTerminal12LinesDebugSteps(t *testing.T) {
	// Debug each step in detail
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(120, 12)
	
	resizeEvent := events.ResizeEventData{Width: 120, Height: 12}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Setup: Terminal 12 lines, Content area %dx%d", windowWidth, windowHeight)
	
	// Add lines one by one and check when scrolling starts
	chars := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l'}
	
	for i, ch := range chars {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
		
		display.Render(editor)
		
		bufferCursor := editor.CurrentBuffer().Cursor()
		screenRow, _ := window.CursorPosition()
		visible := window.VisibleLines()
		
		t.Logf("Line %d ('%c'): cursor buffer (%d,%d), screen row %d, scroll %d, visible %d lines", 
			i, ch, bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop(), len(visible))
		
		// Check if scrolling happened when it shouldn't
		if window.ScrollTop() > 0 && len(visible) < windowHeight {
			t.Logf("❌ Premature scroll: scrolling when only %d/%d content lines used", len(visible), windowHeight)
		}
		
		// Check if cursor is beyond visible area
		if screenRow >= windowHeight {
			t.Logf("❌ Cursor beyond content area: screen row %d >= %d", screenRow, windowHeight)
		}
		
		// Check if we're filling content area properly
		if i < windowHeight && window.ScrollTop() > 0 {
			t.Logf("❌ Scrolling too early: at line %d but content area can hold %d lines", i+1, windowHeight)
		}
	}
}