package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec display/newline_rendering
 * @scenario 改行表示のレンダリング
 * @description 改行を含む複数行コンテンツの正確な表示検証
 * @given 20x5サイズのMockDisplayを作成
 * @when "hello" + Enter + "world"を入力
 * @then 2行のコンテンツが正確に表示され、カーソル位置が適切に設定される
 * @implementation test/mock_display.go, 改行処理
 */
func TestNewlineDisplay(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 5)
	
	// Type "hello"
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content := display.GetContent()
	t.Logf("After 'hello': display content = %v", content)
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	display.Render(editor)
	content = display.GetContent()
	t.Logf("After Enter: display content = %v", content)
	
	// Type "world"
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content = display.GetContent()
	t.Logf("After 'world': display content = %v", content)
	
	// Check that it displays correctly (trim trailing spaces)
	actualContent0 := strings.TrimRight(content[0], " ")
	if actualContent0 != "hello" {
		t.Errorf("Expected line 0 'hello', got %q", actualContent0)
	}
	actualContent1 := strings.TrimRight(content[1], " ")
	if actualContent1 != "world" {
		t.Errorf("Expected line 1 'world', got %q", actualContent1)
	}
	
	// Check cursor position
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 1 || cursorCol != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursorRow, cursorCol)
	}
	
	// Show actual buffer content vs display content
	buffer := editor.CurrentBuffer()
	bufferContent := buffer.Content()
	t.Logf("Buffer content: %v (length: %d)", bufferContent, len(bufferContent))
	t.Logf("Display content: %v", content)
}

/**
 * @spec display/multiline_newline
 * @scenario 複数改行での行末処理
 * @description 連続した改行操作での行末処理とコンテンツ構築
 * @given エディタを新規作成する
 * @when "abc" + Enter + "def" + Enter + "ghi"を順次入力
 * @then 3行のコンテンツが正確に作成され、カーソルが最終行の末尾に配置される
 * @implementation domain/buffer.go, 複数行改行処理
 */
func TestNewlineAtEndOfLine(t *testing.T) {
	editor := domain.NewEditor()
	
	// Type "abc" + Enter + "def" + Enter + "ghi"
	inputs := []struct {
		text   string
		isEnter bool
	}{
		{"abc", false},
		{"", true},
		{"def", false},
		{"", true},
		{"ghi", false},
	}
	
	for _, input := range inputs {
		if input.isEnter {
			event := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(event)
		} else {
			for _, ch := range input.text {
				event := events.KeyEventData{Rune: ch, Key: string(ch)}
				editor.HandleEvent(event)
			}
		}
	}
	
	buffer := editor.CurrentBuffer()
	content := buffer.Content()
	cursor := buffer.Cursor()
	
	t.Logf("Final content: %v (lines: %d)", content, len(content))
	t.Logf("Final cursor: (%d,%d)", cursor.Row, cursor.Col)
	
	// Should have 3 lines
	if len(content) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(content))
	}
	if content[0] != "abc" {
		t.Errorf("Expected line 0 'abc', got %q", content[0])
	}
	if content[1] != "def" {
		t.Errorf("Expected line 1 'def', got %q", content[1])
	}
	if content[2] != "ghi" {
		t.Errorf("Expected line 2 'ghi', got %q", content[2])
	}
	
	// Cursor should be at end of line 2
	if cursor.Row != 2 || cursor.Col != 3 {
		t.Errorf("Expected cursor at (2,3), got (%d,%d)", cursor.Row, cursor.Col)
	}
}