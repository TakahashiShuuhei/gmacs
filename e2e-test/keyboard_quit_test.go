package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec keyboard/quit_mx_command
 * @scenario M-xコマンド入力時のC-gキャンセル
 * @description M-xコマンド入力中にC-gでキャンセルする機能
 * @given M-xコマンド入力モードで部分的にコマンドを入力済み
 * @when C-gキーを押下
 * @then ミニバッファがクリアされ、通常モードに戻る
 * @implementation domain/command.go, KeyboardQuit関数
 */
func TestKeyboardQuitMxCommand(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	// ミニバッファがコマンド入力モードになっているかチェック
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferCommand {
		t.Errorf("Expected MinibufferCommand mode, got %v", minibuffer.Mode())
	}
	
	// 何かコマンドを入力
	for _, ch := range "forward" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// 入力されていることを確認
	if minibuffer.Content() != "forward" {
		t.Errorf("Expected 'forward' in minibuffer, got %q", minibuffer.Content())
	}
	
	// C-g でキャンセル
	event := events.KeyEventData{Key: "g", Ctrl: true}
	editor.HandleEvent(event)
	
	// ミニバッファがクリアされたかチェック
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be inactive after C-g")
	}
	if minibuffer.Content() != "" {
		t.Errorf("Expected empty minibuffer content after C-g, got %q", minibuffer.Content())
	}
}

/**
 * @spec keyboard/quit_find_file
 * @scenario C-x C-f ファイル入力時のC-gキャンセル
 * @description C-x C-f ファイル入力中にC-gでキャンセルする機能
 * @given C-x C-f ファイル入力モードでパスを部分的に入力済み
 * @when C-gキーを押下
 * @then ミニバッファがクリアされ、元の状態に戻る
 * @implementation domain/command.go, KeyboardQuit関数
 */
func TestKeyboardQuitFindFile(t *testing.T) {
	editor := NewEditorWithDefaults()
	originalBuffer := editor.CurrentBuffer()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// ミニバッファがファイル入力モードになっているかチェック
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferFile {
		t.Errorf("Expected MinibufferFile mode, got %v", minibuffer.Mode())
	}
	
	// ファイルパスを部分的に入力
	for _, ch := range "/some/path/to" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// 入力されていることを確認
	if minibuffer.Content() != "/some/path/to" {
		t.Errorf("Expected '/some/path/to' in minibuffer, got %q", minibuffer.Content())
	}
	
	// C-g でキャンセル
	event := events.KeyEventData{Key: "g", Ctrl: true}
	editor.HandleEvent(event)
	
	// ミニバッファがクリアされたかチェック
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be inactive after C-g")
	}
	if minibuffer.Content() != "" {
		t.Errorf("Expected empty minibuffer content after C-g, got %q", minibuffer.Content())
	}
	
	// 元のバッファが維持されているかチェック
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer != originalBuffer {
		t.Error("Buffer should not have changed after C-g cancel")
	}
}

/**
 * @spec keyboard/quit_key_sequence
 * @scenario 進行中のキーシーケンスのC-gキャンセル
 * @description C-x 入力後にC-gでキーシーケンスをキャンセルする機能
 * @given C-x入力済みでキーシーケンスが進行中
 * @when C-gキーを押下
 * @then キーシーケンス状態がリセットされる
 * @implementation domain/command.go, KeyboardQuit関数
 */
func TestKeyboardQuitKeySequence(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// C-x を押下（キーシーケンス開始）
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	
	// この時点ではキーシーケンス進行中でまだ何も実行されていない
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		t.Error("Minibuffer should not be active after C-x only")
	}
	
	// C-g でキーシーケンスをキャンセル
	event2 := events.KeyEventData{Key: "g", Ctrl: true}
	editor.HandleEvent(event2)
	
	// ミニバッファがアクティブでないことを確認（状態がリセットされた）
	if minibuffer.IsActive() {
		t.Error("Minibuffer should not be active after C-g reset")
	}
	
	// この後に C-c を押しても何も起こらないことを確認（キーシーケンスがリセットされたため）
	event3 := events.KeyEventData{Key: "c", Ctrl: true}
	editor.HandleEvent(event3)
	
	// エディタが終了していないことを確認（C-x C-c が実行されていない）
	if !editor.IsRunning() {
		t.Error("Editor should still be running after cancelled C-x sequence")
	}
}

/**
 * @spec keyboard/quit_normal_mode
 * @scenario 通常モードでのC-g動作
 * @description ミニバッファがアクティブでない時のC-g動作
 * @given 通常編集モード
 * @when C-gキーを押下
 * @then 特に何も起こらず、キーシーケンス状態がリセットされる
 * @implementation domain/command.go, KeyboardQuit関数
 */
func TestKeyboardQuitNormalMode(t *testing.T) {
	editor := NewEditorWithDefaults()
	originalBuffer := editor.CurrentBuffer()
	
	// 通常状態でC-gを押す
	event := events.KeyEventData{Key: "g", Ctrl: true}
	editor.HandleEvent(event)
	
	// 何も変わっていないことを確認
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		t.Error("Minibuffer should not be active after C-g in normal mode")
	}
	
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer != originalBuffer {
		t.Error("Buffer should not have changed after C-g in normal mode")
	}
	
	// エディタが正常に動作していることを確認
	if !editor.IsRunning() {
		t.Error("Editor should still be running after C-g in normal mode")
	}
}

/**
 * @spec keyboard/quit_message_clear
 * @scenario メッセージ表示中のC-gクリア
 * @description ミニバッファにメッセージが表示されている時のC-g動作
 * @given ミニバッファにメッセージが表示されている状態
 * @when C-gキーを押下
 * @then メッセージがクリアされる
 * @implementation domain/command.go, KeyboardQuit関数
 */
func TestKeyboardQuitMessageClear(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// メッセージを表示
	editor.SetMinibufferMessage("This is a test message")
	
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Errorf("Expected MinibufferMessage mode, got %v", minibuffer.Mode())
	}
	if minibuffer.Message() != "This is a test message" {
		t.Errorf("Expected test message, got %q", minibuffer.Message())
	}
	
	// C-g でメッセージをクリア
	event := events.KeyEventData{Key: "g", Ctrl: true}
	editor.HandleEvent(event)
	
	// メッセージがクリアされたかチェック
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be inactive after C-g")
	}
	if minibuffer.Message() != "" {
		t.Errorf("Expected empty message after C-g, got %q", minibuffer.Message())
	}
}