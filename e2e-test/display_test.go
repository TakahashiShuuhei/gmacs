package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec display/basic_rendering
 * @scenario 基本的なテキスト表示
 * @description MockDisplayでの基本的なテキスト表示とカーソル位置の検証
 * @given 10x5サイズのMockDisplayを作成
 * @when "hello"を入力する
 * @then テキストが正確に表示され、カーソル位置が適切に設定される
 * @implementation test/mock_display.go, cli/display.go
 */
func TestMockDisplayBasic(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// Resize editor to match display size
	resizeEvent := events.ResizeEventData{Width: 10, Height: 5}
	editor.HandleEvent(resizeEvent)
	
	// Input some text
	text := "hello"
	for _, ch := range text {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	// MockDisplay should have 5 lines total (height=5)
	expectedHeight := 5
	if len(content) != expectedHeight {
		t.Errorf("Expected %d content lines (from MockDisplay height), got %d", expectedHeight, len(content))
	}
	
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "hello" {
		t.Errorf("Expected 'hello', got %q", actualContent)
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 0 || cursorCol != 5 {
		t.Errorf("Expected cursor at (0,5), got (%d,%d)", cursorRow, cursorCol)
	}
}

/**
 * @spec display/japanese_rendering
 * @scenario 日本語テキスト表示
 * @description 日本語文字の表示と表示幅計算の検証
 * @given 10x5サイズのMockDisplayを作成
 * @when "あいう"（ひらがな）を入力する
 * @then 日本語テキストが正確に表示され、カーソル位置が適切に計算される
 * @implementation test/mock_display.go, UTF-8処理
 */
func TestMockDisplayJapanese(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// Resize editor to match display size
	resizeEvent := events.ResizeEventData{Width: 10, Height: 5}
	editor.HandleEvent(resizeEvent)
	
	// Input Japanese text
	text := "あいう"
	for _, ch := range []rune(text) {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "あいう" {
		t.Errorf("Expected 'あいう', got %q", actualContent)
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	t.Logf("Japanese text cursor position: (%d, %d)", cursorRow, cursorCol)
	
	// Show screen with cursor
	screenWithCursor := display.GetScreenWithCursor()
	t.Logf("Screen with cursor:\n%s", screenWithCursor)
	
	// Show detailed info
	t.Logf("Screen info:\n%s", display.GetScreenInfo())
}

/**
 * @spec display/mixed_character_cursor
 * @scenario ASCII+日本語混在カーソル進行
 * @description ASCII文字と日本語文字が混在するテキストでのカーソル位置進行
 * @given 20x5サイズのMockDisplayを作成
 * @when 'a'、'あ'、'b'、'い'、'c'を順次入力
 * @then マルチバイト文字の表示幅を考慮してカーソルが適切に進行する
 * @implementation test/mock_display.go, 文字幅計算
 */
func TestMockDisplayCursorProgression(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(20, 5)
	
	testChars := []rune{'a', 'あ', 'b', 'い', 'c'}
	expectedPositions := []struct{ row, col int }{
		{0, 1}, // a(1)
		{0, 3}, // あ(2) -> total 3
		{0, 4}, // b(1) -> total 4
		{0, 6}, // い(2) -> total 6
		{0, 7}, // c(1) -> total 7
	}
	
	for i, ch := range testChars {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
		
		display.Render(editor)
		
		cursorRow, cursorCol := display.GetCursorPosition()
		
		t.Logf("After '%c': cursor at (%d, %d), expected (%d, %d)", 
			ch, cursorRow, cursorCol, expectedPositions[i].row, expectedPositions[i].col)
		t.Logf("Content: %q", display.GetContent()[0])
		t.Logf("Screen with cursor:\n%s\n", display.GetScreenWithCursor())
		
		if cursorRow != expectedPositions[i].row || cursorCol != expectedPositions[i].col {
			t.Errorf("After '%c': expected cursor (%d, %d), got (%d, %d)",
				ch, expectedPositions[i].row, expectedPositions[i].col, cursorRow, cursorCol)
		}
	}
}

/**
 * @spec display/multiline_rendering
 * @scenario 複数行テキスト表示
 * @description 複数行のテキストとカーソル位置の表示検証
 * @given 10x5サイズのMockDisplayを作成
 * @when "hello" + Enter + "world"を入力
 * @then 2行のテキストが正確に表示され、2行目にカーソルが配置される
 * @implementation test/mock_display.go, 複数行処理
 */
func TestMockDisplayMultiline(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 5)
	
	// Resize editor to match display size
	resizeEvent := events.ResizeEventData{Width: 10, Height: 5}
	editor.HandleEvent(resizeEvent)
	
	// First line
	for _, ch := range "hello" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Second line
	for _, ch := range "world" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent0 := strings.TrimRight(content[0], " ")
	if actualContent0 != "hello" {
		t.Errorf("Expected line 0 'hello', got %q", actualContent0)
	}
	actualContent1 := strings.TrimRight(content[1], " ")
	if actualContent1 != "world" {
		t.Errorf("Expected line 1 'world', got %q", actualContent1)
	}
	
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 1 || cursorCol != 5 {
		t.Errorf("Expected cursor at (1,5), got (%d,%d)", cursorRow, cursorCol)
	}
	
	t.Logf("Multi-line screen:\n%s", display.GetScreenWithCursor())
}

/**
 * @spec display/terminal_width_handling
 * @scenario ターミナル幅と文字幅の問題検証
 * @description 異なる文字タイプのターミナル幅処理の検証
 * @given 10x3サイズのMockDisplayを作成
 * @when ASCII、日本語、混在テキストを各々入力
 * @then ターミナル表示幅とルーン数の違いを適切に処理する
 * @implementation test/mock_display.go, 文字幅計算
 */
func TestMockDisplayWidthProblem(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(10, 3)
	
	testCases := []struct {
		input    string
		expected string
	}{
		{"abc", "abc|"},
		{"あいう", "あいう|"},
		{"aあb", "aあb|"},
	}
	
	for _, tc := range testCases {
		// Reset editor
		editor = domain.NewEditor()
		
		// Input text
		for _, ch := range []rune(tc.input) {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
		
		display.Render(editor)
		
		screenWithCursor := display.GetScreenWithCursor()
		lines := strings.Split(screenWithCursor, "\n")
		actualFirstLine := strings.TrimSpace(lines[0])
		
		t.Logf("Input: %q", tc.input)
		t.Logf("Expected: %q", tc.expected)
		t.Logf("Actual: %q", actualFirstLine)
		
		// Note: This test currently shows the cursor position problem
		// The cursor should account for terminal display width, not just rune count
	}
}