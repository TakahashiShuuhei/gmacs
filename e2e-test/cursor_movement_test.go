package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec cursor/forward_char
 * @scenario 前方向文字移動（C-f）
 * @description カーソルを1文字右に移動する機能の検証
 * @given "hello"を入力済みでカーソルを行頭に設定
 * @when C-f（forward-char）コマンドを実行
 * @then カーソルが1文字右に移動する
 * @implementation domain/cursor.go, events/key_event.go
 */
func TestForwardCharBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Test forward-char (C-f)
	event := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 1 {
		t.Errorf("Expected cursor at (0,1), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec cursor/backward_char
 * @scenario 後方向文字移動（C-b）
 * @description カーソルを1文字左に移動する機能の検証
 * @given "hello"を入力済みでカーソルが行末にある
 * @when C-b（backward-char）コマンドを実行
 * @then カーソルが1文字左に移動する
 * @implementation domain/cursor.go, events/key_event.go
 */
func TestBackwardCharBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Cursor should be at end (0,5)
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (0,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test backward-char (C-b)
	event := events.KeyEventData{Key: "b", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor = buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 4 {
		t.Errorf("Expected cursor at (0,4), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec cursor/arrow_keys
 * @scenario 矢印キーによるカーソル移動
 * @description 左右矢印キーでのカーソル移動機能の検証
 * @given "hello"を入力済みでカーソルを行頭に設定
 * @when 右矢印キー、左矢印キーを順次押下
 * @then カーソルが適切に左右に移動する
 * @implementation domain/cursor.go, events/key_event.go
 */
func TestArrowKeys(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move cursor to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Test right arrow
	rightEvent := events.KeyEventData{Key: "\x1b[C"}
	editor.HandleEvent(rightEvent)
	
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 1 {
		t.Errorf("Expected cursor at (0,1) after right arrow, got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test left arrow
	leftEvent := events.KeyEventData{Key: "\x1b[D"}
	editor.HandleEvent(leftEvent)
	
	cursor = buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (0,0) after left arrow, got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec cursor/vertical_movement
 * @scenario 垂直方向のカーソル移動（C-p/C-n）
 * @description 前の行・次の行へのカーソル移動機能の検証
 * @given 2行のテキスト（"hello"、"world"）を入力済み
 * @when C-p（前の行）、C-n（次の行）を順次実行
 * @then カーソルが適切に上下の行を移動する
 * @implementation domain/cursor.go, events/key_event.go
 */
func TestNextPreviousLine(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Create multi-line content: "hello" + Enter + "world"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Cursor should be at (1,5)
	cursor := buffer.Cursor()
	if cursor.Row != 1 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test previous-line (C-p)
	event := events.KeyEventData{Key: "p", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor = buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (0,5) after C-p, got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test next-line (C-n)
	event = events.KeyEventData{Key: "n", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor = buffer.Cursor()
	if cursor.Row != 1 || cursor.Col != 5 {
		t.Errorf("Expected cursor at (1,5) after C-n, got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec cursor/line_boundaries
 * @scenario 行頭・行末移動（C-a/C-e）
 * @description 行の先頭と末尾への移動機能の検証
 * @given "hello world"を入力済みでカーソルが行末にある
 * @when C-a（行頭）、C-e（行末）を順次実行
 * @then カーソルが行頭と行末に適切に移動する
 * @implementation domain/cursor.go, events/key_event.go
 */
func TestBeginningEndOfLine(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello world"
	for _, ch := range "hello world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Cursor should be at end (0,11)
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 11 {
		t.Errorf("Expected cursor at (0,11), got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test beginning-of-line (C-a)
	event := events.KeyEventData{Key: "a", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor = buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (0,0) after C-a, got (%d,%d)", cursor.Row, cursor.Col)
	}
	
	// Test end-of-line (C-e)
	event = events.KeyEventData{Key: "e", Ctrl: true}
	editor.HandleEvent(event)
	
	cursor = buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 11 {
		t.Errorf("Expected cursor at (0,11) after C-e, got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec cursor/japanese_support
 * @scenario 日本語文字を含むカーソル移動
 * @description ASCII文字と日本語文字が混在するテキストでのカーソル移動
 * @given "aあbいc"（ASCII+日本語混在）を入力済み
 * @when C-fで1文字ずつ前進する
 * @then マルチバイト文字を適切に処理してカーソルが移動する
 * @implementation domain/cursor.go, UTF-8処理
 */
func TestCursorMovementWithJapanese(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "aあbいc" (mixed ASCII and Japanese)
	chars := []rune{'a', 'あ', 'b', 'い', 'c'}
	for _, ch := range chars {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Move to beginning
	buffer.SetCursor(domain.Position{Row: 0, Col: 0})
	
	// Test forward movement through mixed characters
	positions := []struct{ row, col int }{
		{0, 1}, // after 'a'
		{0, 4}, // after 'あ' (3 bytes)
		{0, 5}, // after 'b'
		{0, 8}, // after 'い' (3 bytes)
		{0, 9}, // after 'c'
	}
	
	for i, expected := range positions {
		event := events.KeyEventData{Key: "f", Ctrl: true}
		editor.HandleEvent(event)
		
		cursor := buffer.Cursor()
		if cursor.Row != expected.row || cursor.Col != expected.col {
			t.Errorf("Step %d: expected cursor at (%d,%d), got (%d,%d)", 
				i+1, expected.row, expected.col, cursor.Row, cursor.Col)
		}
	}
}

/**
 * @spec cursor/mx_commands
 * @scenario M-xコマンドによるカーソル移動
 * @description M-x beginning-of-lineコマンドの実行検証
 * @given "hello"を入力済みでカーソルが行末にある
 * @when M-x beginning-of-lineコマンドを実行
 * @then カーソルが行頭に移動する
 * @implementation domain/commands.go, events/key_event.go
 */
func TestInteractiveCommands(t *testing.T) {
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Test M-x forward-char
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	for _, ch := range "beginning-of-line" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Should move cursor to beginning
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (0,0) after M-x beginning-of-line, got (%d,%d)", cursor.Row, cursor.Col)
	}
}