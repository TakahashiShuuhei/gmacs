package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

/**
 * @spec keyboard/key_sequence_binding
 * @scenario キーシーケンスバインディングシステム
 * @description BindKeySequence APIでキーシーケンスを設定し実行する機能の検証
 * @given 新しいキーバインディングマップを作成
 * @when "C-x C-f"のようなキーシーケンスをバインドし、該当するキー入力を送信
 * @then バインドされたコマンドが実行される
 * @implementation domain/keybinding.go, キーシーケンス処理システム
 */
func TestKeySequenceBinding(t *testing.T) {
	kbm := domain.NewEmptyKeyBindingMap()
	
	executed := false
	testCommand := func(editor *domain.Editor) error {
		executed = true
		return nil
	}
	
	// C-x C-f をバインド
	kbm.BindKeySequence("C-x C-f", testCommand)
	
	// C-x を送信（まだ実行されない）
	cmd, matched, continuing := kbm.ProcessKeyPress("x", true, false)
	if matched {
		t.Error("Command should not execute after C-x alone")
	}
	if !continuing {
		t.Error("Should be continuing after C-x prefix")
	}
	if cmd != nil {
		t.Error("Command should be nil for partial sequence")
	}
	
	// C-f を送信（実行される）
	cmd, matched, continuing = kbm.ProcessKeyPress("f", true, false)
	if !matched {
		t.Error("Command should match after C-x C-f sequence")
	}
	if continuing {
		t.Error("Should not be continuing after complete sequence")
	}
	if cmd == nil {
		t.Error("Command should not be nil for complete sequence")
	}
	
	// コマンドを実行
	editor := NewEditorWithDefaults()
	err := cmd(editor)
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}
	
	if !executed {
		t.Error("Test command should have been executed")
	}
}

/**
 * @spec keyboard/sequence_reset
 * @scenario キーシーケンス状態のリセット
 * @description 無効なキーが入力された場合のシーケンス状態リセットの検証
 * @given キーバインディングマップに"C-x C-c"をバインド
 * @when C-x後に無効なキー（'z'）を送信
 * @then シーケンス状態がリセットされ、その後のC-cでは実行されない
 * @implementation domain/keybinding.go, シーケンス状態管理
 */
func TestKeySequenceReset(t *testing.T) {
	kbm := domain.NewEmptyKeyBindingMap()
	
	executed := false
	testCommand := func(editor *domain.Editor) error {
		executed = true
		return nil
	}
	
	// C-x C-c をバインド
	kbm.BindKeySequence("C-x C-c", testCommand)
	
	// C-x を送信
	_, matched, continuing := kbm.ProcessKeyPress("x", true, false)
	if matched || !continuing {
		t.Error("C-x should be continuing but not matched")
	}
	
	// 無効なキー 'z' を送信（シーケンスがリセットされる）
	_, matched, continuing = kbm.ProcessKeyPress("z", false, false)
	if matched || continuing {
		t.Error("Invalid key should reset sequence")
	}
	
	// その後 C-c を送信してもコマンドは実行されない
	_, matched, continuing = kbm.ProcessKeyPress("c", true, false)
	if matched || continuing {
		t.Error("C-c alone should not match or continue after reset")
	}
	
	if executed {
		t.Error("Command should not have been executed after sequence reset")
	}
}

/**
 * @spec keyboard/multiple_sequences
 * @scenario 複数キーシーケンスの同時サポート
 * @description 複数の異なるキーシーケンスを同時にサポートする機能の検証
 * @given キーバインディングマップに"C-x C-c"と"C-x C-f"を両方バインド
 * @when 各シーケンスを順次実行
 * @then それぞれ対応するコマンドが実行される
 * @implementation domain/keybinding.go, 複数シーケンス管理
 */
func TestMultipleKeySequences(t *testing.T) {
	// エディタを作成してそのキーバインディングマップを使用
	editor := NewEditorWithDefaults()
	
	quitExecuted := false
	quitCommand := func(e *domain.Editor) error {
		quitExecuted = true
		return nil
	}
	
	fileExecuted := false
	fileCommand := func(e *domain.Editor) error {
		fileExecuted = true
		return nil
	}
	
	// テスト用に空のキーバインディングマップを作成
	kbm := domain.NewEmptyKeyBindingMap()
	
	// 複数のシーケンスをバインド  
	kbm.BindKeySequence("C-x C-c", quitCommand)
	kbm.BindKeySequence("C-x C-f", fileCommand)
	
	// C-x C-c を実行
	_, _, _ = kbm.ProcessKeyPress("x", true, false)
	cmd, matched, _ := kbm.ProcessKeyPress("c", true, false)
	if !matched || cmd == nil {
		t.Error("C-x C-c should match")
	}
	err := cmd(editor)
	if err != nil {
		t.Errorf("C-x C-c command failed: %v", err)
	}
	
	// C-x C-f を実行
	_, _, _ = kbm.ProcessKeyPress("x", true, false)
	cmd, matched, _ = kbm.ProcessKeyPress("f", true, false)
	if !matched || cmd == nil {
		t.Error("C-x C-f should match")
	}
	err = cmd(editor)
	if err != nil {
		t.Errorf("C-x C-f command failed: %v", err)
	}
	
	if !quitExecuted {
		t.Error("Quit command should have been executed")
	}
	if !fileExecuted {
		t.Error("File command should have been executed")
	}
}