package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec scroll/vertical_scrolling
 * @scenario 垂直スクロール動作
 * @description 大量のコンテンツがある場合の垂直スクロール動作の検証
 * @given 40x10サイズのウィンドウに20行のコンテンツを作成
 * @when カーソルが最後の行にある状態でスクロール位置を設定
 * @then カーソルが可視範囲に保たれるように自動スクロールされる
 * @implementation domain/window.go, domain/scroll.go
 */
func TestVerticalScrolling(t *testing.T) {
	editor := NewEditorWithDefaults()
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

/**
 * @spec scroll/horizontal_scrolling
 * @scenario 水平スクロール動作
 * @description 長い行のコンテンツでの水平スクロール動作の検証
 * @given 10x5の狭いウィンドウと長い行のコンテンツ
 * @when 行ラップを無効化して水平スクロールを設定
 * @then 指定した位置からコンテンツが表示される
 * @implementation domain/window.go, 水平スクロール
 */
func TestHorizontalScrolling(t *testing.T) {
	editor := NewEditorWithDefaults()
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
	// With continuation indicators: \ at start (for left scroll) and \ at end (if content continues)
	expectedSubstring := longLine[5:] // Skip first 5 characters
	
	// Account for continuation indicators in the expected result
	availableWidth := 10 - 2 // Subtract 2 for left (\) and right (\) indicators
	if len(expectedSubstring) > availableWidth {
		expectedSubstring = expectedSubstring[:availableWidth]
	}
	expectedWithIndicators := "\\" + expectedSubstring + "\\"

	if visible[0] != expectedWithIndicators {
		t.Errorf("Expected visible line after horizontal scroll to be %q, got %q", expectedWithIndicators, visible[0])
	}
}

/**
 * @spec scroll/line_wrapping
 * @scenario 行ラップ機能
 * @description 長い行のラップ機能の有効/無効切り替え検証
 * @given 10x5の小さいウィンドウと長い行のコンテンツ
 * @when 行ラップの有効/無効を切り替える
 * @then ラップ有効時は複数行、無効時は単一行で表示される
 * @implementation domain/window.go, 行ラップ処理
 */
func TestLineWrapping(t *testing.T) {
	editor := NewEditorWithDefaults()
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

/**
 * @spec scroll/toggle_line_wrap
 * @scenario 行ラップトグルコマンド
 * @description ToggleLineWrapコマンドによる行ラップ状態の切り替え
 * @given エディタを新規作成（デフォルトでラップ有効）
 * @when ToggleLineWrapコマンドを実行
 * @then 行ラップの有効/無効が切り替わる
 * @implementation domain/commands.go, domain/window.go
 */
func TestToggleLineWrap(t *testing.T) {
	editor := NewEditorWithDefaults()

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

/**
 * @spec scroll/page_navigation
 * @scenario ページアップ/ダウンナビゲーション
 * @description PageUp/PageDownコマンドによるページ単位のスクロール
 * @given 50行の大量コンテンツを持つエディタ
 * @when PageDown、PageUpコマンドを順次実行
 * @then スクロール位置がページ単位で適切に変更される
 * @implementation domain/commands.go, domain/window.go
 */
func TestPageUpDown(t *testing.T) {
	editor := NewEditorWithDefaults()

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

/**
 * @spec scroll/individual_scroll_commands
 * @scenario 個別スクロールコマンド
 * @description ScrollUp/ScrollDownコマンドによる1行単位のスクロール
 * @given 30行のコンテンツを持つエディタ
 * @when ScrollDown、ScrollUpコマンドを順次実行
 * @then スクロール位置が1行単位で正確に変更される
 * @implementation domain/commands.go, domain/window.go
 */
func TestScrollCommands(t *testing.T) {
	editor := NewEditorWithDefaults()

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