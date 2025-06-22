package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec application/clean_exit
 * @scenario C-x C-c による正常終了
 * @description C-x C-c コマンドでエディタが正常に終了する機能
 * @given エディタが実行中の状態
 * @when C-x C-c キーシーケンスを実行
 * @then エディタが終了状態になる
 * @implementation domain/command.go, Quit関数
 */
func TestCleanExit(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// エディタが実行中であることを確認
	if !editor.IsRunning() {
		t.Error("Editor should be running initially")
	}
	
	// C-x C-c を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	
	event2 := events.KeyEventData{Key: "c", Ctrl: true}
	editor.HandleEvent(event2)
	
	// エディタが終了状態になったことを確認
	if editor.IsRunning() {
		t.Error("Editor should not be running after C-x C-c")
	}
}

/**
 * @spec application/signal_exit
 * @scenario シグナルによる終了
 * @description SIGINTやSIGTERMシグナルでの終了処理
 * @given エディタが実行中の状態
 * @when QuitEventDataを受信
 * @then エディタが終了状態になる
 * @implementation events/quit_event.go, domain/editor.go
 */
func TestSignalExit(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// エディタが実行中であることを確認
	if !editor.IsRunning() {
		t.Error("Editor should be running initially")
	}
	
	// QuitEventData を送信（シグナルハンドラーからの終了イベント）
	quitEvent := events.QuitEventData{}
	editor.HandleEvent(quitEvent)
	
	// エディタが終了状態になったことを確認
	if editor.IsRunning() {
		t.Error("Editor should not be running after quit event")
	}
}

/**
 * @spec application/exit_during_input
 * @scenario 入力中の終了
 * @description ミニバッファ入力中にC-x C-cで終了する場合
 * @given M-xコマンド入力中の状態
 * @when C-x C-c キーシーケンスを実行
 * @then ミニバッファがクリアされずに終了する（通常の終了が優先される）
 * @implementation domain/editor.go, キー処理優先順位
 */
func TestExitDuringInput(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// M-x を実行してミニバッファをアクティブにする
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active after M-x")
	}
	
	// C-x C-c で終了を試行
	event3 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event3)
	event4 := events.KeyEventData{Key: "c", Ctrl: true}
	editor.HandleEvent(event4)
	
	// エディタが終了状態になったことを確認
	if editor.IsRunning() {
		t.Error("Editor should not be running after C-x C-c, even during minibuffer input")
	}
}