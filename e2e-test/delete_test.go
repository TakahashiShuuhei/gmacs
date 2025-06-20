package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec delete/backward_char_basic
 * @scenario C-h による基本的な文字削除
 * @description カーソル前の文字を削除する基本的な backspace 機能
 * @given "hello"を入力済みでカーソルが行末にある
 * @when C-h（DeleteBackwardChar）コマンドを実行
 * @then 最後の文字が削除され"hell"になる
 * @implementation domain/buffer.go, DeleteBackward関数
 */
func TestDeleteBackwardCharBasic(t *testing.T) {
	editor := domain.NewEditor()
	
	// "hello"を入力
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	buffer := editor.CurrentBuffer()
	if len(buffer.Content()) == 0 || buffer.Content()[0] != "hello" {
		t.Errorf("Expected 'hello', got %v", buffer.Content())
	}
	
	// C-h (backspace) を実行
	event := events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) == 0 || content[0] != "hell" {
		t.Errorf("Expected 'hell' after backspace, got %v", content)
	}
	
	// カーソル位置を確認
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 4 {
		t.Errorf("Expected cursor at (0,4), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec delete/forward_char_basic
 * @scenario C-d による基本的な文字削除
 * @description カーソル位置の文字を削除する delete-char 機能
 * @given "hello"を入力済みでカーソルを行頭に移動
 * @when C-d（DeleteChar）コマンドを実行
 * @then 最初の文字が削除され"ello"になる
 * @implementation domain/buffer.go, DeleteForward関数
 */
func TestDeleteCharBasic(t *testing.T) {
	editor := domain.NewEditor()
	
	// "hello"を入力
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルを行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true} // C-a
	editor.HandleEvent(event)
	
	buffer := editor.CurrentBuffer()
	
	// C-d (delete-char) を実行
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) == 0 || content[0] != "ello" {
		t.Errorf("Expected 'ello' after delete-char, got %v", content)
	}
	
	// カーソル位置を確認（行頭のまま）
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec delete/backward_char_japanese
 * @scenario 日本語文字のbackspace削除
 * @description 日本語文字（マルチバイト）のbackspace削除機能
 * @given "aあiい"を入力済みでカーソルが行末にある
 * @when C-h（DeleteBackwardChar）コマンドを実行
 * @then 最後の日本語文字が削除され"aあi"になる
 * @implementation domain/buffer.go, UTF-8対応削除処理
 */
func TestDeleteBackwardCharJapanese(t *testing.T) {
	editor := domain.NewEditor()
	
	// "aあiい"を入力
	text := "aあiい"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	buffer := editor.CurrentBuffer()
	if len(buffer.Content()) == 0 || buffer.Content()[0] != "aあiい" {
		t.Errorf("Expected 'aあiい', got %v", buffer.Content())
	}
	
	// C-h (backspace) を実行
	event := events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) == 0 || content[0] != "aあi" {
		t.Errorf("Expected 'aあi' after backspace, got %v", content)
	}
}

/**
 * @spec delete/forward_char_japanese
 * @scenario 日本語文字のdelete-char削除
 * @description 日本語文字（マルチバイト）のdelete-char削除機能
 * @given "aあiい"を入力済みでカーソルを"あ"の位置に移動
 * @when C-d（DeleteChar）コマンドを実行
 * @then "あ"が削除され"aiい"になる
 * @implementation domain/buffer.go, UTF-8対応削除処理
 */
func TestDeleteCharJapanese(t *testing.T) {
	editor := domain.NewEditor()
	
	// "aあiい"を入力
	text := "aあiい"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルを行頭に移動
	event := events.KeyEventData{Key: "a", Ctrl: true} // C-a
	editor.HandleEvent(event)
	
	// 1文字進む（"a"を通過して"あ"の位置へ）
	event = events.KeyEventData{Key: "f", Ctrl: true} // C-f
	editor.HandleEvent(event)
	
	buffer := editor.CurrentBuffer()
	
	// C-d (delete-char) を実行
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) == 0 || content[0] != "aiい" {
		t.Errorf("Expected 'aiい' after delete-char, got %v", content)
	}
}

/**
 * @spec delete/backward_line_join
 * @scenario 行頭でのbackspaceによる行結合
 * @description 行頭でbackspaceを実行して前の行と結合する機能
 * @given 2行のテキスト（"hello"、"world"）でカーソルが2行目の行頭
 * @when C-h（DeleteBackwardChar）コマンドを実行
 * @then 2行が結合され"helloworld"の1行になる
 * @implementation domain/buffer.go, 行結合処理
 */
func TestDeleteBackwardLineJoin(t *testing.T) {
	editor := domain.NewEditor()
	
	// "hello" + Enter + "world"を入力
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enter
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	text = "world"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// カーソルを2行目の行頭に移動
	event = events.KeyEventData{Key: "a", Ctrl: true} // C-a
	editor.HandleEvent(event)
	
	buffer := editor.CurrentBuffer()
	
	// 2行あることを確認
	if len(buffer.Content()) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(buffer.Content()))
	}
	
	// C-h (backspace) を実行
	event = events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) != 1 {
		t.Errorf("Expected 1 line after join, got %d", len(content))
	}
	if content[0] != "helloworld" {
		t.Errorf("Expected 'helloworld' after join, got %v", content[0])
	}
	
	// カーソル位置を確認（"hello"の末尾）
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (0,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec delete/forward_line_join
 * @scenario 行末でのdelete-charによる行結合
 * @description 行末でdelete-charを実行して次の行と結合する機能
 * @given 2行のテキスト（"hello"、"world"）でカーソルが1行目の行末
 * @when C-d（DeleteChar）コマンドを実行
 * @then 2行が結合され"helloworld"の1行になる
 * @implementation domain/buffer.go, 行結合処理
 */
func TestDeleteForwardLineJoin(t *testing.T) {
	editor := domain.NewEditor()
	
	// "hello" + Enter + "world"を入力
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enter
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	text = "world"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// 1行目の行末に移動
	event = events.KeyEventData{Key: "p", Ctrl: true} // C-p (上の行)
	editor.HandleEvent(event)
	event = events.KeyEventData{Key: "e", Ctrl: true} // C-e (行末)
	editor.HandleEvent(event)
	
	buffer := editor.CurrentBuffer()
	
	// 2行あることを確認
	if len(buffer.Content()) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(buffer.Content()))
	}
	
	// C-d (delete-char) を実行
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	// 結果を確認
	content := buffer.Content()
	if len(content) != 1 {
		t.Errorf("Expected 1 line after join, got %d", len(content))
	}
	if content[0] != "helloworld" {
		t.Errorf("Expected 'helloworld' after join, got %v", content[0])
	}
	
	// カーソル位置を確認（"hello"の末尾）
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (0,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec delete/edge_cases
 * @scenario 削除のエッジケース
 * @description バッファの境界での削除動作
 * @given 空のバッファまたは境界位置
 * @when 削除コマンドを実行
 * @then エラーなく適切に処理される
 * @implementation domain/buffer.go, 境界チェック
 */
func TestDeleteEdgeCases(t *testing.T) {
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// 空のバッファでC-h（何も起こらない）
	event := events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(event)
	
	content := buffer.Content()
	if len(content) != 1 || content[0] != "" {
		t.Errorf("Expected empty content after backspace on empty buffer, got %v", content)
	}
	
	// 空のバッファでC-d（何も起こらない）
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	content = buffer.Content()
	if len(content) != 1 || content[0] != "" {
		t.Errorf("Expected empty content after delete-char on empty buffer, got %v", content)
	}
	
	// 1文字入力
	event = events.KeyEventData{Rune: 'a', Key: "a"}
	editor.HandleEvent(event)
	
	// 行末でC-d（何も起こらない）
	event = events.KeyEventData{Key: "d", Ctrl: true}
	editor.HandleEvent(event)
	
	content = buffer.Content()
	if len(content) != 1 || content[0] != "a" {
		t.Errorf("Expected 'a' after delete-char at end of line, got %v", content)
	}
}