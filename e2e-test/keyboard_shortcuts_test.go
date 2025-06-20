package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec keyboard/ctrl_x_ctrl_c_quit
 * @scenario C-x C-cでのエディタ終了
 * @description C-x C-cキーシーケンスでエディタを終了する機能の検証
 * @given エディタが実行中の状態
 * @when C-x（prefix key）とC-cキーイベントを順次送信する
 * @then エディタが終了状態になる
 * @implementation domain/editor.go, prefix key システム
 */
func TestCtrlXCtrlCQuit(t *testing.T) {
	editor := domain.NewEditor()
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running initially")
	}
	
	// Send C-x (prefix key)
	event1 := events.KeyEventData{
		Key:  "x",
		Ctrl: true,
	}
	editor.HandleEvent(event1)
	
	// Editor should still be running after C-x
	if !editor.IsRunning() {
		t.Error("Editor should still be running after C-x")
	}
	
	// Send C-c (quit command)
	event2 := events.KeyEventData{
		Key:  "c",
		Ctrl: true,
	}
	editor.HandleEvent(event2)
	
	// Now editor should have quit
	if editor.IsRunning() {
		t.Error("Editor should have quit after C-x C-c")
	}
}

/**
 * @spec keyboard/ctrl_c_no_quit
 * @scenario C-c単独ではエディタ終了しない
 * @description C-x prefix key なしのC-cではエディタが終了しないことの検証
 * @given エディタが実行中の状態
 * @when C-cキーイベントのみを送信する
 * @then エディタが実行中のまま維持される
 * @implementation domain/editor.go, prefix key システム
 */
func TestCtrlCAloneDoesNotQuit(t *testing.T) {
	editor := domain.NewEditor()
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running initially")
	}
	
	// Send C-c alone (should not quit)
	event := events.KeyEventData{
		Key:  "c",
		Ctrl: true,
	}
	editor.HandleEvent(event)
	
	// Editor should still be running
	if !editor.IsRunning() {
		t.Error("Editor should still be running after C-c alone")
	}
}

/**
 * @spec keyboard/ctrl_x_prefix_reset
 * @scenario C-x prefix key状態のリセット
 * @description C-x後に無効なキーを押すとprefix状態がリセットされることの検証
 * @given エディタが実行中の状態
 * @when C-x後に通常の文字キーを送信する
 * @then prefix状態がリセットされ、通常のテキスト入力として処理される
 * @implementation domain/editor.go, prefix key システム
 */
func TestCtrlXPrefixReset(t *testing.T) {
	editor := domain.NewEditor()
	
	// Send C-x (prefix key)
	event1 := events.KeyEventData{
		Key:  "x",
		Ctrl: true,
	}
	editor.HandleEvent(event1)
	
	// Send a regular character (should reset prefix state)
	event2 := events.KeyEventData{
		Key:  "a",
		Rune: 'a',
	}
	editor.HandleEvent(event2)
	
	// Editor should still be running (prefix state reset)
	if !editor.IsRunning() {
		t.Error("Editor should still be running after C-x followed by regular key")
	}
	
	// Now C-c alone should not quit (prefix state was reset)
	event3 := events.KeyEventData{
		Key:  "c",
		Ctrl: true,
	}
	editor.HandleEvent(event3)
	
	if !editor.IsRunning() {
		t.Error("Editor should still be running - C-x prefix should have been reset")
	}
}

/**
 * @spec keyboard/ctrl_modifier_no_insert
 * @scenario Ctrl修飾キーのテキスト非挿入
 * @description Ctrl+文字キーの組み合わせでテキストが挿入されないことの検証
 * @given エディタを新規作成する
 * @when Ctrl+aキーイベントを送信する
 * @then テキストが挿入されず、空の行が維持される
 * @implementation domain/editor.go, events/key_event.go
 */
func TestCtrlModifierDoesNotInsertText(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	event := events.KeyEventData{
		Key:  "a",
		Rune: 'a',
		Ctrl: true,
	}
	editor.HandleEvent(event)
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}
	
	if lines[0] != "" {
		t.Errorf("Expected empty line after Ctrl+a, got '%s'", lines[0])
	}
}

/**
 * @spec keyboard/meta_modifier_no_insert
 * @scenario Meta修飾キーのテキスト非挿入
 * @description Meta+文字キーの組み合わせでテキストが挿入されないことの検証
 * @given エディタを新規作成する
 * @when Meta+xキーイベントを送信する
 * @then テキストが挿入されず、空の行が維持される
 * @implementation domain/editor.go, events/key_event.go
 */
func TestMetaModifierDoesNotInsertText(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	event := events.KeyEventData{
		Key:  "x",
		Rune: 'x',
		Meta: true,
	}
	editor.HandleEvent(event)
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}
	
	if lines[0] != "" {
		t.Errorf("Expected empty line after Meta+x, got '%s'", lines[0])
	}
}