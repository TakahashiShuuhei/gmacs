package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec scroll/horizontal_cursor_follow
 * @scenario 水平スクロール時のカーソル追従
 * @description 行ラップ無効時の水平スクロールとカーソル移動の同期検証
 * @given 狭いウィンドウで行ラップを無効にし、長い行を作成
 * @when カーソルを右端まで移動し、その後左に戻る
 * @then カーソル位置に応じて水平スクロールが正しく調整される
 * @implementation domain/scroll.go, 水平スクロール制御
 */
func TestHorizontalScrollCursorFollow(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 8) // 狭いウィンドウ（幅20）
	
	// ウィンドウサイズを狭く設定
	resizeEvent := events.ResizeEventData{Width: 20, Height: 8}
	editor.HandleEvent(resizeEvent)
	
	// 行ラップを無効にする
	window := editor.CurrentWindow()
	if window != nil {
		window.SetLineWrap(false)
	}
	
	// ウィンドウ幅より長い行を作成（40文字）
	longText := "This is a very long line that exceeds window width and should scroll horizontally"
	for _, ch := range longText {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	buffer := editor.CurrentBuffer()
	
	content := buffer.Content()
	if len(content) > 0 {
		t.Logf("Created long line: %q (length: %d)", content[0], len(content[0]))
	}
	width, height := window.Size()
	t.Logf("Window size: %dx%d", width, height)
	t.Logf("Initial cursor position: (%d,%d)", buffer.Cursor().Row, buffer.Cursor().Col)
	
	// 右端での水平スクロール確認
	scrollX := window.ScrollLeft()
	t.Logf("Scroll position after input: scrollX=%d", scrollX)
	
	// カーソルを行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true} // C-a (beginning-of-line)
	editor.HandleEvent(event)
	
	display.Render(editor)
	scrollXAfterHome := window.ScrollLeft()
	cursorAfterHome := buffer.Cursor()
	
	t.Logf("After C-a: cursor=(%d,%d), scrollX=%d", cursorAfterHome.Row, cursorAfterHome.Col, scrollXAfterHome)
	
	// 検証: カーソルが行頭に戻ったとき、水平スクロールも0に戻るべき
	if cursorAfterHome.Col != 0 {
		t.Errorf("Expected cursor column 0 after C-a, got %d", cursorAfterHome.Col)
	}
	
	if scrollXAfterHome != 0 {
		t.Errorf("Expected horizontal scroll to reset to 0 when cursor moves to beginning of line, got scrollX=%d", scrollXAfterHome)
	}
	
	// 段階的にカーソルを右に移動して水平スクロールを確認
	positions := []int{10, 20, 30, 40, 50}
	for _, targetPos := range positions {
		if targetPos > len(longText) {
			continue
		}
		
		// カーソルを目標位置に移動
		for buffer.Cursor().Col < targetPos {
			event := events.KeyEventData{Key: "f", Ctrl: true} // C-f (forward-char)
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		scrollX := window.ScrollLeft()
		cursor := buffer.Cursor()
		
		t.Logf("At position %d: cursor=(%d,%d), scrollX=%d", targetPos, cursor.Row, cursor.Col, scrollX)
		
		// カーソルが可視範囲内にあることを確認
		visibleStartX := scrollX
		width, _ := window.Size()
		visibleEndX := scrollX + width
		
		if cursor.Col < visibleStartX || cursor.Col >= visibleEndX {
			t.Errorf("Cursor at column %d is not visible in range [%d,%d)", cursor.Col, visibleStartX, visibleEndX)
		}
	}
	
	// 今度は左に戻っていく
	t.Logf("\n=== Moving cursor back to the left ===")
	reversePositions := []int{30, 20, 10, 0}
	for _, targetPos := range reversePositions {
		// カーソルを目標位置に移動
		for buffer.Cursor().Col > targetPos {
			event := events.KeyEventData{Key: "b", Ctrl: true} // C-b (backward-char)
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		scrollX := window.ScrollLeft()
		cursor := buffer.Cursor()
		
		t.Logf("Back to position %d: cursor=(%d,%d), scrollX=%d", targetPos, cursor.Row, cursor.Col, scrollX)
		
		// カーソルが可視範囲内にあることを確認
		visibleStartX := scrollX
		width, _ := window.Size()
		visibleEndX := scrollX + width
		
		if cursor.Col < visibleStartX || cursor.Col >= visibleEndX {
			t.Errorf("Cursor at column %d is not visible in range [%d,%d) when moving left", cursor.Col, visibleStartX, visibleEndX)
		}
	}
}

/**
 * @spec scroll/horizontal_boundary_scroll
 * @scenario 水平スクロール境界でのスクロール動作
 * @description カーソルが可視範囲の左右境界を超えた時のスクロール動作
 * @given 行ラップ無効の狭いウィンドウと長い行
 * @when カーソルを左右の境界を超えて移動
 * @then 適切なタイミングでスクロールが発生する
 * @implementation domain/scroll.go, 境界スクロール処理
 */
func TestHorizontalBoundaryScroll(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5) // 非常に狭いウィンドウ（幅10）
	
	// ウィンドウサイズを非常に狭く設定
	resizeEvent := events.ResizeEventData{Width: 10, Height: 5}
	editor.HandleEvent(resizeEvent)
	
	// 行ラップを無効にする
	window := editor.CurrentWindow()
	if window != nil {
		window.SetLineWrap(false)
	}
	
	// 30文字の行を作成
	text := "0123456789ABCDEFGHIJKLMNOPQRST"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルを行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true} // C-a
	editor.HandleEvent(event)
	
	display.Render(editor)
	buffer := editor.CurrentBuffer()
	
	t.Logf("Test text: %q", text)
	width, _ := window.Size()
	t.Logf("Window width: %d", width)
	
	// 1文字ずつ右に移動してスクロールの挙動を確認
	for i := 0; i < len(text); i++ {
		if i > 0 {
			event := events.KeyEventData{Key: "f", Ctrl: true} // C-f
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		scrollX := window.ScrollLeft()
		cursor := buffer.Cursor()
		
		// カーソルが可視範囲内にあることを確認
		visibleStartX := scrollX
		width, _ := window.Size()
		visibleEndX := scrollX + width - 1 // -1 for cursor display
		
		t.Logf("Step %d: cursor.Col=%d, scrollX=%d, visible range=[%d,%d]", 
			i, cursor.Col, scrollX, visibleStartX, visibleEndX)
		
		if cursor.Col < visibleStartX || cursor.Col > visibleEndX {
			t.Errorf("Step %d: Cursor at column %d is not visible in range [%d,%d]", 
				i, cursor.Col, visibleStartX, visibleEndX)
		}
	}
	
	t.Logf("\n=== Moving back left ===")
	
	// 今度は1文字ずつ左に移動してスクロールの挙動を確認
	for i := len(text) - 1; i >= 0; i-- {
		if i < len(text) - 1 {
			event := events.KeyEventData{Key: "b", Ctrl: true} // C-b
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		scrollX := window.ScrollLeft()
		cursor := buffer.Cursor()
		
		// カーソルが可視範囲内にあることを確認
		visibleStartX := scrollX
		width, _ := window.Size()
		visibleEndX := scrollX + width - 1
		
		t.Logf("Back step %d: cursor.Col=%d, scrollX=%d, visible range=[%d,%d]", 
			i, cursor.Col, scrollX, visibleStartX, visibleEndX)
		
		if cursor.Col < visibleStartX || cursor.Col > visibleEndX {
			t.Errorf("Back step %d: Cursor at column %d is not visible in range [%d,%d]", 
				i, cursor.Col, visibleStartX, visibleEndX)
		}
	}
}

/**
 * @spec scroll/horizontal_toggle_wrap_state
 * @scenario 行ラップ切り替え時の水平スクロール状態
 * @description 行ラップの有効/無効切り替え時の水平スクロール状態の保持
 * @given 長い行とカーソルが右端にある状態
 * @when 行ラップの有効/無効を切り替える
 * @then 適切にスクロール状態が管理される
 * @implementation domain/scroll.go, ラップ切り替え処理
 */
func TestHorizontalToggleWrapState(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(15, 6)
	
	// ウィンドウサイズを設定
	resizeEvent := events.ResizeEventData{Width: 15, Height: 6}
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	
	// 長い行を作成
	text := "This is a very long line for testing horizontal scroll behavior"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	t.Logf("Created text: %q (length: %d)", text, len(text))
	width, _ := window.Size()
	t.Logf("Window width: %d", width)
	
	// 初期状態（行ラップ有効）での表示確認
	display.Render(editor)
	buffer := editor.CurrentBuffer()
	scrollX1 := window.ScrollLeft()
	cursor1 := buffer.Cursor()
	isWrapped1 := window.LineWrap()
	
	t.Logf("Initial state: lineWrap=%v, cursor=(%d,%d), scrollX=%d", 
		isWrapped1, cursor1.Row, cursor1.Col, scrollX1)
	
	// 行ラップを無効にする
	if window != nil {
		window.SetLineWrap(false)
	}
	
	display.Render(editor)
	scrollX2 := window.ScrollLeft()
	cursor2 := buffer.Cursor()
	isWrapped2 := window.LineWrap()
	
	t.Logf("After disabling wrap: lineWrap=%v, cursor=(%d,%d), scrollX=%d", 
		isWrapped2, cursor2.Row, cursor2.Col, scrollX2)
	
	// 行ラップ無効時はカーソルが可視範囲内にあることを確認
	if !isWrapped2 {
		visibleStartX := scrollX2
		width, _ := window.Size()
		visibleEndX := scrollX2 + width - 1
		
		if cursor2.Col < visibleStartX || cursor2.Col > visibleEndX {
			t.Errorf("After disabling wrap: Cursor at column %d is not visible in range [%d,%d]", 
				cursor2.Col, visibleStartX, visibleEndX)
		}
	}
	
	// 再度行ラップを有効にする
	if window != nil {
		window.SetLineWrap(true)
	}
	
	display.Render(editor)
	scrollX3 := window.ScrollLeft()
	cursor3 := buffer.Cursor()
	isWrapped3 := window.LineWrap()
	
	t.Logf("After re-enabling wrap: lineWrap=%v, cursor=(%d,%d), scrollX=%d", 
		isWrapped3, cursor3.Row, cursor3.Col, scrollX3)
	
	// カーソル位置は変わらないはず
	if cursor3.Row != cursor1.Row || cursor3.Col != cursor1.Col {
		t.Errorf("Cursor position changed after wrap toggle: expected (%d,%d), got (%d,%d)",
			cursor1.Row, cursor1.Col, cursor3.Row, cursor3.Col)
	}
}