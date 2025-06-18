package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec keyboard/ctrl_c_quit
 * @scenario Ctrl+Cでのエディタ終了
 * @description Ctrl+Cキーコンビネーションでエディタを終了する機能の検証
 * @given エディタが実行中の状態
 * @when Ctrl+Cキーイベントを送信する
 * @then エディタが終了状態になる
 * @implementation domain/editor.go, events/key_event.go
 */
func TestCtrlCQuit(t *testing.T) {
	editor := domain.NewEditor()
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running initially")
	}
	
	event := events.KeyEventData{
		Key:  "c",
		Ctrl: true,
	}
	editor.HandleEvent(event)
	
	if editor.IsRunning() {
		t.Error("Editor should have quit after Ctrl+C")
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