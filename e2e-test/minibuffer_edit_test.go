package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec minibuffer/edit_delete_forward
 * @scenario ミニバッファでのC-d文字削除
 * @description M-xコマンド入力中にC-dで前方の文字を削除する機能
 * @given M-xコマンド入力モードで"forward"を入力済み、カーソルが"f"の位置
 * @when C-dキーを押下
 * @then "f"が削除され"orward"になる
 * @implementation domain/minibuffer.go, DeleteForward関数
 */
func TestMinibufferDeleteForward(t *testing.T) {
	editor := domain.NewEditor()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	// "forward"を入力
	for _, ch := range "forward" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルを先頭に移動（"f"の位置）
	minibuffer := editor.Minibuffer()
	minibuffer.MoveCursorToBeginning()
	
	if minibuffer.Content() != "forward" {
		t.Errorf("Expected 'forward', got %q", minibuffer.Content())
	}
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Expected cursor at 0, got %d", minibuffer.CursorPosition())
	}
	
	// C-d で前方削除
	event := events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	// "f"が削除されて"orward"になることを確認
	if minibuffer.Content() != "orward" {
		t.Errorf("Expected 'orward' after C-d, got %q", minibuffer.Content())
	}
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Expected cursor still at 0, got %d", minibuffer.CursorPosition())
	}
}

/**
 * @spec minibuffer/edit_cursor_movement
 * @scenario ミニバッファでのカーソル移動
 * @description M-xコマンド入力中にC-f/C-bでカーソルを移動する機能
 * @given M-xコマンド入力モードで"hello"を入力済み
 * @when C-a（行頭）、C-f（前進）、C-b（後退）、C-e（行末）を順次実行
 * @then カーソルが適切な位置に移動する
 * @implementation domain/minibuffer.go, カーソル移動関数
 */
func TestMinibufferCursorMovement(t *testing.T) {
	editor := domain.NewEditor()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	// "hello"を入力
	for _, ch := range "hello" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	minibuffer := editor.Minibuffer()
	
	// 初期状態（行末）
	if minibuffer.CursorPosition() != 5 {
		t.Errorf("Expected cursor at 5, got %d", minibuffer.CursorPosition())
	}
	
	// C-a で行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Expected cursor at 0 after C-a, got %d", minibuffer.CursorPosition())
	}
	
	// C-f で前進
	event = events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 1 {
		t.Errorf("Expected cursor at 1 after C-f, got %d", minibuffer.CursorPosition())
	}
	
	// C-f で再度前進
	event = events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 2 {
		t.Errorf("Expected cursor at 2 after second C-f, got %d", minibuffer.CursorPosition())
	}
	
	// C-b で後退
	event = events.KeyEventData{Key: "b", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 1 {
		t.Errorf("Expected cursor at 1 after C-b, got %d", minibuffer.CursorPosition())
	}
	
	// C-e で行末に移動
	event = events.KeyEventData{Key: "e", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 5 {
		t.Errorf("Expected cursor at 5 after C-e, got %d", minibuffer.CursorPosition())
	}
}

/**
 * @spec minibuffer/edit_file_input
 * @scenario ファイル入力モードでの編集機能
 * @description C-x C-fファイル入力中にC-h/C-dで編集する機能
 * @given C-x C-fファイル入力モードで"/path/to/file.txt"を入力済み
 * @when カーソル移動と削除コマンドを実行
 * @then ファイルパスが適切に編集される
 * @implementation domain/minibuffer.go, ファイル入力モード編集
 */
func TestMinibufferFileInputEdit(t *testing.T) {
	editor := domain.NewEditor()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferFile {
		t.Errorf("Expected MinibufferFile mode, got %v", minibuffer.Mode())
	}
	
	// "/path/to/file.txt"を入力
	testPath := "/path/to/file.txt"
	for _, ch := range testPath {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	if minibuffer.Content() != testPath {
		t.Errorf("Expected '%s', got %q", testPath, minibuffer.Content())
	}
	
	// C-a で行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Expected cursor at 0, got %d", minibuffer.CursorPosition())
	}
	
	// C-d で"/"を削除
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	expected := "path/to/file.txt"
	if minibuffer.Content() != expected {
		t.Errorf("Expected '%s' after C-d, got %q", expected, minibuffer.Content())
	}
	
	// C-e で行末に移動
	event = events.KeyEventData{Key: "e", Ctrl: true}
	editor.HandleEvent(event)
	expectedPos := len([]rune(expected))
	if minibuffer.CursorPosition() != expectedPos {
		t.Errorf("Expected cursor at %d, got %d", expectedPos, minibuffer.CursorPosition())
	}
	
	// C-h で".txt"の"t"を削除
	event = events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	expected = "path/to/file.tx"
	if minibuffer.Content() != expected {
		t.Errorf("Expected '%s' after C-h, got %q", expected, minibuffer.Content())
	}
}

/**
 * @spec minibuffer/edit_japanese_characters
 * @scenario ミニバッファでの日本語文字編集
 * @description M-xコマンド入力中に日本語文字を含むテキストを編集する機能
 * @given M-xコマンド入力モードで"aあbいc"を入力済み
 * @when カーソル移動と削除を行う
 * @then 日本語文字が適切に処理される
 * @implementation domain/minibuffer.go, UTF-8対応編集
 */
func TestMinibufferJapaneseEdit(t *testing.T) {
	editor := domain.NewEditor()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	// "aあbいc"を入力
	testText := "aあbいc"
	for _, ch := range testText {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	minibuffer := editor.Minibuffer()
	if minibuffer.Content() != testText {
		t.Errorf("Expected '%s', got %q", testText, minibuffer.Content())
	}
	
	// カーソルを"あ"の前に移動（位置1）
	minibuffer.MoveCursorToBeginning()
	minibuffer.MoveCursorForward() // 'a'の後
	
	if minibuffer.CursorPosition() != 1 {
		t.Errorf("Expected cursor at 1, got %d", minibuffer.CursorPosition())
	}
	
	// C-d で"あ"を削除
	event := events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	expected := "abいc"
	if minibuffer.Content() != expected {
		t.Errorf("Expected '%s' after deleting あ, got %q", expected, minibuffer.Content())
	}
	
	// カーソルを"い"の位置に移動（位置2）
	minibuffer.MoveCursorForward() // 'b'の後
	
	// C-h で"b"を削除
	event = events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	expected = "aいc"
	if minibuffer.Content() != expected {
		t.Errorf("Expected '%s' after deleting b, got %q", expected, minibuffer.Content())
	}
}

/**
 * @spec minibuffer/edit_boundary_conditions
 * @scenario ミニバッファ編集の境界条件
 * @description カーソルが境界位置にある時の編集動作
 * @given M-xコマンド入力モードで"test"を入力済み
 * @when 境界位置での削除とカーソル移動を試行
 * @then エラーなく適切に処理される
 * @implementation domain/minibuffer.go, 境界チェック
 */
func TestMinibufferEditBoundaryConditions(t *testing.T) {
	editor := domain.NewEditor()
	
	// M-x を実行
	event1 := events.KeyEventData{Key: "\x1b"} // Escape for Meta
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "x"}
	editor.HandleEvent(event2)
	
	// "test"を入力
	for _, ch := range "test" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	minibuffer := editor.Minibuffer()
	
	// 行末でC-fを試行（カーソルが範囲外に移動しないことを確認）
	initialPos := minibuffer.CursorPosition()
	event := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != initialPos {
		t.Errorf("Cursor moved beyond end: expected %d, got %d", initialPos, minibuffer.CursorPosition())
	}
	
	// 行末でC-dを試行（何も削除されないことを確認）
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.Content() != "test" {
		t.Errorf("Content changed unexpectedly: got %q", minibuffer.Content())
	}
	
	// 行頭に移動
	minibuffer.MoveCursorToBeginning()
	
	// 行頭でC-bを試行（カーソルが負の値にならないことを確認）
	event = events.KeyEventData{Key: "b", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Cursor moved before beginning: got %d", minibuffer.CursorPosition())
	}
	
	// 行頭でC-hを試行（何も削除されないことを確認）
	event = events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	if minibuffer.Content() != "test" {
		t.Errorf("Content changed unexpectedly: got %q", minibuffer.Content())
	}
}