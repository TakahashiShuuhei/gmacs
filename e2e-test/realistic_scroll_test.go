package test

import (
	"fmt"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec scroll/realistic_terminal
 * @scenario リアルなターミナルサイズでのスクロール
 * @description 80x24のリアルなターミナルサイズでのスクロール動作検証
 * @given 80x24ターミナル（22コンテンツ行）でリサイズイベントを送信
 * @when 30行のコンテンツを順次追加し、各ステップでスクロール状態を監視
 * @then ウィンドウ高を超えたタイミングでスクロールが開始され、カーソルが常に可視範囲内に保たれる
 * @implementation domain/scroll.go, リアルターミナル環境
 */
func TestRealisticTerminalScroll(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// Simulate a realistic terminal size (80x24)
	display := NewMockDisplay(80, 24) // 24 total -> 22 content lines
	
	// Simulate the resize event that happens at startup
	resizeEvent := events.ResizeEventData{Width: 80, Height: 24}
	editor.HandleEvent(resizeEvent) // This will set window to 80x22
	
	window := editor.CurrentWindow()
	windowWidth, windowHeight := window.Size()
	t.Logf("Terminal size: 80x24, Window size: %dx%d", windowWidth, windowHeight)
	
	// Add lines one by one, tracking when scrolling should start
	expectedScrollStart := windowHeight // Should scroll when we exceed window height
	
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
		
		display.Render(editor)
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		scrollTop := window.ScrollTop()
		screenRow, _ := window.CursorPosition()
		
		t.Logf("Line %d: buffer cursor (%d,%d), scroll top: %d, screen row: %d", 
			i, cursor.Row, cursor.Col, scrollTop, screenRow)
		
		// Check when scrolling starts
		if i == expectedScrollStart-1 {
			t.Logf("  -> This should be the last line before scrolling starts")
			if scrollTop != 0 {
				t.Errorf("Line %d: Premature scrolling, scroll top = %d", i, scrollTop)
			}
		}
		
		if i == expectedScrollStart {
			t.Logf("  -> This should trigger scrolling")
			if scrollTop == 0 {
				t.Errorf("Line %d: Expected scrolling to start, but scroll top is still 0", i)
			}
		}
		
		// Cursor should always be visible
		if screenRow < 0 || screenRow >= windowHeight {
			t.Errorf("Line %d: Screen cursor row %d is outside window [0, %d)", 
				i, screenRow, windowHeight)
		}
	}
}

/**
 * @spec scroll/timing_verification
 * @scenario 異なるウィンドウサイズでのスクロールタイミング検証
 * @description 複数のウィンドウサイズでスクロール開始タイミングの正確性を検証
 * @given 異なるターミナル高（6、6、10、24）でテストケースを実行
 * @when 各サイズでウィンドウ高まで行を追加し、さらに1行追加
 * @then ウィンドウ高まではスクロールせず、超えた時点でスクロールが発生する
 * @implementation domain/scroll.go, サイズ別タイミング検証
 */
func TestScrollStartsAtRightTime(t *testing.T) {
	// Test with different window sizes to verify scroll timing
	testCases := []struct {
		terminalHeight int
		expectedWindowHeight int
	}{
		{6, 4},   // 6 terminal -> 4 content
		{10, 8},  // 10 terminal -> 8 content  
		{24, 22}, // 24 terminal -> 22 content
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Terminal%d", tc.terminalHeight), func(t *testing.T) {
			editor := NewEditorWithDefaults()
			_ = NewMockDisplay(40, tc.terminalHeight)
			
			// Simulate resize event
			resizeEvent := events.ResizeEventData{Width: 40, Height: tc.terminalHeight}
			editor.HandleEvent(resizeEvent)
			
			window := editor.CurrentWindow()
			_, windowHeight := window.Size()
			
			if windowHeight != tc.expectedWindowHeight {
				t.Fatalf("Expected window height %d, got %d", tc.expectedWindowHeight, windowHeight)
			}
			
			// Add lines up to window height
			for i := 0; i < windowHeight; i++ {
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
			
			// Should not have scrolled yet
			scrollTop := window.ScrollTop()
			if scrollTop != 0 {
				t.Errorf("After %d lines: Expected no scrolling, but scroll top = %d", 
					windowHeight, scrollTop)
			}
			
			// Add one more line - this should trigger scrolling
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
			
			line := "Overflow"
			for _, ch := range line {
				event := events.KeyEventData{Rune: ch, Key: string(ch)}
				editor.HandleEvent(event)
			}
			
			// Now it should have scrolled
			scrollTop = window.ScrollTop()
			if scrollTop == 0 {
				t.Errorf("After overflow line: Expected scrolling, but scroll top is still 0")
			}
			
			// Test the debug command
			err := domain.ShowDebugInfo(editor)
			if err != nil {
				t.Errorf("ShowDebugInfo failed: %v", err)
			}
		})
	}
}