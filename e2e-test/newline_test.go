package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec input/newline_basic
 * @scenario 基本的な改行挿入
 * @description 行末でのEnterキーによる基本的な改行動作
 * @given 空のバッファに"hello"を入力済み
 * @when 行末でEnterキーを押し、"world"を入力
 * @then 2行に分かれてテキストが配置され、カーソルが適切な位置に移動する
 * @implementation domain/buffer.go, events/key_event.go
 */
func TestNewlineBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Initial state
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("Initial: content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After 'hello': content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After Enter: content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Type "world"
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After 'world': content=%v (len=%d), cursor=(%d,%d)", content, len(content), cursor.Row, cursor.Col)
	
	// Check final state
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "hello" {
		t.Errorf("Expected first line 'hello', got %q", content[0])
	}
	if content[1] != "world" {
		t.Errorf("Expected second line 'world', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec input/newline_split
 * @scenario 行の中間での改行挿入
 * @description 行の中間でEnterキーを押した際の行分割動作
 * @given "hello world"を入力済みでカーソルを"hello"の後（位置5）に移動
 * @when カーソル位置でEnterキーを押下
 * @then 行が"hello"と" world"に分割され、カーソルが2行目の先頭に移動する
 * @implementation domain/buffer.go, 行分割処理
 */
func TestNewlineInMiddle(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello world"
	for _, ch := range "hello world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to after "hello" (position 5)
	buffer.SetCursor(domain.Position{Row: 0, Col: 5})
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("Before Enter in middle: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Press Enter in the middle
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content = buffer.Content()
	cursor = buffer.Cursor()
	t.Logf("After Enter in middle: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the split
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "hello" {
		t.Errorf("Expected first line 'hello', got %q", content[0])
	}
	if content[1] != " world" {
		t.Errorf("Expected second line ' world', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (1,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec input/newline_beginning
 * @scenario 行頭での改行挿入
 * @description 行頭でEnterキーを押した際の新しい行挿入動作
 * @given "hello"を入力済みでカーソルを行頭に移動
 * @when 行頭でEnterキーを押下
 * @then 空の新しい行が挿入され、既存のコンテンツが2行目に移動する
 * @implementation domain/buffer.go, 行挿入処理
 */
func TestNewlineAtBeginning(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Press Enter at beginning
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("After Enter at beginning: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the result
	if len(content) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(content))
	}
	if content[0] != "" {
		t.Errorf("Expected first line empty, got %q", content[0])
	}
	if content[1] != "hello" {
		t.Errorf("Expected second line 'hello', got %q", content[1])
	}
	if cursor.Row != 1 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (1,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec input/newline_multiple
 * @scenario 連続した改行挿入
 * @description 複数のEnterキーを連続して押した際の動作
 * @given 空のバッファから開始
 * @when "a" + Enter + "b" + Enter + "c"を順次入力
 * @then 3行のコンテンツが正確に作成され、カーソル位置が適切に設定される
 * @implementation domain/buffer.go, 複数行処理
 */
func TestMultipleNewlines(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "a", Enter, "b", Enter, "c"
	chars := []rune{'a', '\n', 'b', '\n', 'c'}
	for _, ch := range chars {
		if ch == '\n' {
			event := events.KeyEventData{Key: "Enter", Rune: ch}
			editor.HandleEvent(event)
		} else {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	content := buffer.Content()
	cursor := buffer.Cursor()
	t.Logf("After multiple newlines: content=%v, cursor=(%d,%d)", content, cursor.Row, cursor.Col)
	
	// Check the result
	if len(content) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(content))
	}
	if content[0] != "a" {
		t.Errorf("Expected line 0 'a', got %q", content[0])
	}
	if content[1] != "b" {
		t.Errorf("Expected line 1 'b', got %q", content[1])
	}
	if content[2] != "c" {
		t.Errorf("Expected line 2 'c', got %q", content[2])
	}
	if cursor.Row != 2 || cursor.Col != 1 {
		t.Errorf("Expected cursor at (2,1), got (%d,%d)", cursor.Row, cursor.Col)
	}
}