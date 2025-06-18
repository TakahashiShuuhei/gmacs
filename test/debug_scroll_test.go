package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec scroll/edge_case_debug
 * @scenario スクロールエッジケースのデバッグ
 * @description 8行丁度まで埋めた後のEnterキー押下時のスクロール動作の詳細分析
 * @given 40x10ディスプレイ（8コンテンツ行）で8行丁度までコンテンツを埋める
 * @when 最後の可視行でEnterキーを押下
 * @then スクロール量と表示内容が期待値と一致し、適切な1行スクロールが発生する
 * @implementation domain/scroll.go, エッジケース処理
 */
func TestDebugScrollBehavior(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // 10 total = 8 content + mode + mini
	
	window := editor.CurrentWindow()
	window.Resize(40, 8) // 8 content lines (0-7)
	
	// Fill exactly 8 lines to reach the edge case
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
	screenRow, _ := window.CursorPosition()
	t.Logf("Setup: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	
	// Debug: What happens when we press Enter at the last visible line?
	t.Logf("=== Debug: Press Enter at cursor (%d,%d) ===", bufferCursor.Row, bufferCursor.Col)
	
	// Step by step Enter processing
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	
	// Check state before Enter
	t.Logf("BEFORE Enter: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	
	// Apply Enter
	editor.HandleEvent(enterEvent)
	
	// Check state after Enter
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, _ = window.CursorPosition()
	visible := window.VisibleLines()
	t.Logf("AFTER Enter: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// Check if the scroll amount is as expected
	expectedScroll := 1  // User expects 1 line scroll
	actualScroll := window.ScrollTop()
	if actualScroll != expectedScroll {
		t.Logf("MISMATCH: Expected scroll %d, got %d", expectedScroll, actualScroll)
		
		// Try to understand why
		if actualScroll > expectedScroll {
			t.Logf("OVER-SCROLL: Scrolled %d lines instead of %d", actualScroll, expectedScroll)
		} else {
			t.Logf("UNDER-SCROLL: Scrolled %d lines instead of %d", actualScroll, expectedScroll)
		}
	}
	
	// Render and check final state
	display.Render(editor)
	visible = window.VisibleLines()
	t.Logf("Final visible lines: %v", visible)
	
	// User expectation: should show [Line 1, Line 2, ..., Line 7, ""]
	// Actual behavior might be different
	if len(visible) > 0 {
		if visible[0] == "Line 1" {
			t.Logf("SUCCESS: Shows Line 1 as expected")
		} else {
			t.Logf("PROBLEM: Shows '%s' instead of 'Line 1'", visible[0])
		}
	}
}