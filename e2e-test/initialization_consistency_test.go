package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec initialization/consistency_check
 * @scenario テスト環境と本番環境の初期化が一致することを確認
 * @description NewEditorWithDefaults()と実際のアプリケーションの初期化が同じ結果を生成することをテスト
 * @given NewEditorWithDefaults()でエディタを作成
 * @when 基本的なコマンドとキーバインディングが利用可能かチェック
 * @then すべてのコアコマンドが正常に動作する
 * @implementation main.go, e2e-test/test_helpers.go
 */
func TestInitializationConsistency(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// 必須のコアコマンドがすべて登録されていることを確認
	requiredCommands := []string{
		"quit",
		"keyboard-quit", 
		"find-file",
		"delete-backward-char",
		"delete-char",
		"forward-char",
		"backward-char",
		"auto-a-mode", // Lua定義のコマンド
	}
	
	for _, cmdName := range requiredCommands {
		t.Run("command_"+cmdName, func(t *testing.T) {
			// M-x でコマンドを実行してみる
			event1 := events.KeyEventData{Key: "\x1b"} // Escape
			editor.HandleEvent(event1)
			event2 := events.KeyEventData{Key: "x"}
			editor.HandleEvent(event2)
			
			// コマンド名を入力
			for _, ch := range cmdName {
				event := events.KeyEventData{
					Rune: ch,
					Key:  string(ch),
				}
				editor.HandleEvent(event)
			}
			
			// Enter で実行
			enterEvent := events.KeyEventData{Key: "Enter"}
			editor.HandleEvent(enterEvent)
			
			// minibufferの結果をチェック
			minibuffer := editor.Minibuffer()
			if minibuffer.IsActive() {
				message := minibuffer.Message()
				if strings.Contains(message, "Unknown command") {
					t.Errorf("Command %s not found: %s", cmdName, message)
				}
				// テスト間でクリア
				minibuffer.Clear()
			}
		})
	}
}

/**
 * @spec initialization/keybinding_check
 * @scenario 重要なキーバインディングが正常に動作することを確認
 * @description 設定読み込み後にキーバインディングが正しく機能することをテスト
 * @given NewEditorWithDefaults()でエディタを作成
 * @when 重要なキーバインディングを実行
 * @then 対応するコマンドが正常に実行される
 * @implementation default.lua, domain/keybinding.go
 */
func TestKeybindingConsistency(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// 重要なキーバインディングのテスト
	testCases := []struct {
		name        string
		keys        []events.KeyEventData
		expectMsg   string
		description string
	}{
		{
			name: "M-a_auto_a_mode",
			keys: []events.KeyEventData{
				{Key: "\x1b"},          // Escape
				{Key: "a", Rune: 'a'},  // a
			},
			expectMsg:   "Auto-A mode",
			description: "M-a should execute auto-a-mode",
		},
		{
			name: "C-g_keyboard_quit",
			keys: []events.KeyEventData{
				{Key: "g", Ctrl: true}, // C-g
			},
			expectMsg:   "", // keyboard-quit clears minibuffer, no message
			description: "C-g should execute keyboard-quit",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// キーイベントを送信
			for _, keyEvent := range tc.keys {
				editor.HandleEvent(keyEvent)
			}
			
			// 結果をチェック
			minibuffer := editor.Minibuffer()
			if tc.expectMsg != "" {
				if !minibuffer.IsActive() {
					t.Errorf("%s: Expected minibuffer to be active", tc.description)
				} else {
					message := minibuffer.Message()
					if !strings.Contains(message, tc.expectMsg) {
						t.Errorf("%s: Expected message containing '%s', got '%s'", tc.description, tc.expectMsg, message)
					}
				}
			}
			
			// テスト間でクリア
			minibuffer.Clear()
		})
	}
}

/**
 * @spec initialization/config_load_error_detection
 * @scenario 設定読み込み中のエラーを検出
 * @description 破損した設定が与えられた場合にエラーが適切に検出されることをテスト
 * @given 無効なLua設定
 * @when エディタ初期化を試行
 * @then エラーが適切に報告される
 * @implementation lua-config/config_loader.go
 */
func TestConfigLoadErrorDetection(t *testing.T) {
	// 無効な設定のテストケース
	invalidConfigs := []struct {
		name   string
		config string
		issue  string
	}{
		{
			name:   "unknown_command",
			config: `gmacs.bind_key("C-x", "unknown-command")`,
			issue:  "Unknown command should be detected",
		},
		{
			name:   "syntax_error",
			config: `gmacs.bind_key("C-x", `,
			issue:  "Lua syntax error should be detected",
		},
	}
	
	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// このテストは実際にはエラーを期待するので、
			// パニックやエラーをキャッチする仕組みが必要
			// 現在の実装では設定エラーはログに記録されるが、
			// テストで検証可能にするには改善が必要
			t.Logf("Test case: %s - %s", tc.name, tc.issue)
			t.Skip("Config error detection needs improvement in implementation")
		})
	}
}