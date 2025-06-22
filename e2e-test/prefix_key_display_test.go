package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec display/prefix_key_display
 * @scenario プレフィックスキーの表示
 * @description C-x入力後にミニバッファに"C-x -"のような表示が出る機能
 * @given エディタを新規作成
 * @when C-xキーを押下
 * @then キーシーケンス進行中の表示が"C-x -"になる
 * @implementation domain/keybinding.go, cli/display.go
 */
func TestPrefixKeyDisplay(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// 初期状態ではキーシーケンス進行中でない
	keySequence := editor.GetKeySequenceInProgress()
	if keySequence != "" {
		t.Errorf("Expected no key sequence initially, got %q", keySequence)
	}
	
	// C-x を送信
	event := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event)
	
	// キーシーケンス進行中の表示を確認
	keySequence = editor.GetKeySequenceInProgress()
	expected := "C-x -"
	if keySequence != expected {
		t.Errorf("Expected %q after C-x, got %q", expected, keySequence)
	}
	
	// C-c を送信してシーケンス完了
	event = events.KeyEventData{Key: "c", Ctrl: true}
	editor.HandleEvent(event)
	
	// シーケンス完了後は表示がクリアされる
	keySequence = editor.GetKeySequenceInProgress()
	if keySequence != "" {
		t.Errorf("Expected no key sequence after completion, got %q", keySequence)
	}
}

/**
 * @spec display/key_sequence_format
 * @scenario キーシーケンス表示フォーマット
 * @description 各種修飾キーの組み合わせの正しい表示
 * @given キーバインディングマップを作成
 * @when 各種キープレス組み合わせをフォーマット
 * @then 適切な文字列表記が生成される
 * @implementation domain/keybinding.go, FormatSequence関数
 */
func TestKeySequenceFormat(t *testing.T) {
	testCases := []struct {
		name     string
		sequence []domain.KeyPress
		expected string
	}{
		{
			name:     "Empty sequence",
			sequence: []domain.KeyPress{},
			expected: "",
		},
		{
			name: "Single Ctrl key",
			sequence: []domain.KeyPress{
				{Key: "x", Ctrl: true, Meta: false},
			},
			expected: "C-x -",
		},
		{
			name: "Single Meta key",
			sequence: []domain.KeyPress{
				{Key: "x", Ctrl: false, Meta: true},
			},
			expected: "M-x -",
		},
		{
			name: "Ctrl+Meta combination",
			sequence: []domain.KeyPress{
				{Key: "x", Ctrl: true, Meta: true},
			},
			expected: "C-M-x -",
		},
		{
			name: "Multiple key sequence",
			sequence: []domain.KeyPress{
				{Key: "x", Ctrl: true, Meta: false},
				{Key: "c", Ctrl: true, Meta: false},
			},
			expected: "C-x C-c -",
		},
		{
			name: "Plain key",
			sequence: []domain.KeyPress{
				{Key: "a", Ctrl: false, Meta: false},
			},
			expected: "a -",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.FormatSequence(tc.sequence)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

/**
 * @spec display/key_sequence_cancel
 * @scenario キーシーケンスキャンセル後の表示
 * @description Escapeキーでキーシーケンスをキャンセルした後の表示クリア
 * @given C-x入力でキーシーケンス進行中
 * @when Escapeキーを押下
 * @then キーシーケンス表示がクリアされる
 * @implementation domain/editor.go, Escapeキー処理
 */
func TestKeySequenceCancelDisplay(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// C-x を送信してキーシーケンス開始
	event := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event)
	
	// キーシーケンス進行中であることを確認
	keySequence := editor.GetKeySequenceInProgress()
	if keySequence != "C-x -" {
		t.Errorf("Expected 'C-x -' after C-x, got %q", keySequence)
	}
	
	// Escape でキャンセル
	event = events.KeyEventData{Key: "Escape"}
	editor.HandleEvent(event)
	
	// キーシーケンス表示がクリアされることを確認
	keySequence = editor.GetKeySequenceInProgress()
	if keySequence != "" {
		t.Errorf("Expected no key sequence after Escape, got %q", keySequence)
	}
}

/**
 * @spec minibuffer/cursor_position_accuracy
 * @scenario ミニバッファでのカーソル位置精度
 * @description ミニバッファでの日本語文字を含むテキストでのカーソル位置計算精度
 * @given M-xコマンド入力モードでマルチバイト文字を入力
 * @when カーソル移動を行う
 * @then 正確なバイト位置とルーン位置が計算される
 * @implementation domain/minibuffer.go, カーソル位置計算
 */
func TestMinibufferCursorPositionAccuracy(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	minibuffer := editor.Minibuffer()
	
	// "aあb"を入力（ASCII + Japanese + ASCII）
	testText := "aあb"
	for _, ch := range testText {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルが最後にあることを確認（ルーン位置3）
	expectedRunePos := 3
	if minibuffer.CursorPosition() != expectedRunePos {
		t.Errorf("Expected cursor at rune position %d, got %d", expectedRunePos, minibuffer.CursorPosition())
	}
	
	// カーソルを"あ"の前に移動（ルーン位置1）
	minibuffer.MoveCursorToBeginning()
	minibuffer.MoveCursorForward() // 'a'の後
	
	expectedRunePos = 1
	if minibuffer.CursorPosition() != expectedRunePos {
		t.Errorf("Expected cursor at rune position %d after moving, got %d", expectedRunePos, minibuffer.CursorPosition())
	}
	
	// 日本語文字を削除
	event := events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	// コンテンツが"ab"になり、カーソル位置が維持されることを確認
	expected := "ab"
	if minibuffer.Content() != expected {
		t.Errorf("Expected content %q after deleting あ, got %q", expected, minibuffer.Content())
	}
	
	// カーソル位置がルーン位置1に維持されることを確認
	if minibuffer.CursorPosition() != 1 {
		t.Errorf("Expected cursor still at rune position 1, got %d", minibuffer.CursorPosition())
	}
}