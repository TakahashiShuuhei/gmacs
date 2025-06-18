package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec terminal/width_calculation
 * @scenario ターミナル幅計算問題の検証
 * @description ASCII文字と日本語文字の混合テキストでのターミナル表示位置計算
 * @given 20x3のMockDisplayと様々な文字組み合わせのテストケース
 * @when 各テストケースで文字を入力し、カーソル位置を取得
 * @then ASCII文字は1列、日本語文字は2列、混合テキストは合計列数で正確に表示される
 * @implementation test/mock_display.go, ターミナル幅計算処理
 */
func TestTerminalWidthIssue(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 3)
	
	testCases := []struct {
		input              string
		expectedRunePos    int
		expectedTerminalPos int // What it should be in a real terminal
	}{
		{"a", 1, 1},           // ASCII: same
		{"あ", 1, 2},           // Japanese: 1 rune = 2 terminal columns
		{"ab", 2, 2},          // ASCII: same  
		{"あい", 2, 4},         // Japanese: 2 runes = 4 terminal columns
		{"abc", 3, 3},         // ASCII: same
		{"あいう", 3, 6},        // Japanese: 3 runes = 6 terminal columns
		{"aあb", 3, 4},         // Mixed: 1+2+1 = 4 terminal columns
	}
	
	for _, tc := range testCases {
		// Reset editor
		editor = domain.NewEditor()
		
		// Input text
		for _, ch := range []rune(tc.input) {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		_, cursorCol := display.GetCursorPosition()
		
		t.Logf("Input: %-8s | Rune pos: %d | Terminal pos should be: %d | Actual: %d", 
			tc.input, tc.expectedRunePos, tc.expectedTerminalPos, cursorCol)
		
		// Now it should use terminal width position
		if cursorCol != tc.expectedTerminalPos {
			t.Errorf("Terminal position mismatch for %q: expected %d, got %d", 
				tc.input, tc.expectedTerminalPos, cursorCol)
		} else {
			t.Logf("  ✅ Correct terminal width position")
		}
	}
}