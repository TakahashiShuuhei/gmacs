package test

import (
	"testing"
	"unicode/utf8"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec cursor/japanese_position
 * @scenario 日本語文字でのカーソル位置計算
 * @description 日本語文字入力時のバイト位置とターミナル表示位置の正確な計算
 * @given エディタを新規作成する
 * @when "あいう"（日本語ひらがな3文字）を入力
 * @then バイト位置が9（3文字 × 3バイト）、ターミナル表示位置が6（3文字 × 2幅）になる
 * @implementation domain/cursor.go, UTF-8処理
 */
func TestCursorPositionWithJapanese(t *testing.T) {
	editor := domain.NewEditor()
	
	// Test "あいう"
	testText := "あいう"
	for _, ch := range []rune(testText) {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	buffer := editor.CurrentBuffer()
	cursor := buffer.Cursor()
	
	// Check buffer content and cursor position
	content := buffer.Content()
	line := content[0]
	
	t.Logf("Line content: %q", line)
	t.Logf("Line bytes: %d", len(line))
	t.Logf("Line runes: %d", utf8.RuneCountInString(line))
	t.Logf("Cursor byte position: %d", cursor.Col)
	
	// Calculate expected positions
	expectedBytes := 0
	expectedRunes := 0
	for _, r := range []rune(testText) {
		expectedBytes += utf8.RuneLen(r)
		expectedRunes++
	}
	
	t.Logf("Expected byte position: %d", expectedBytes)
	t.Logf("Expected rune position: %d", expectedRunes)
	
	if cursor.Col != expectedBytes {
		t.Errorf("Expected cursor at byte %d, got %d", expectedBytes, cursor.Col)
	}
	
	// Test cursor display position
	window := editor.CurrentWindow()
	screenRow, screenCol := window.CursorPosition()
	
	t.Logf("Screen cursor position: (%d, %d)", screenRow, screenCol)
	
	// Now expects terminal width, not rune count
	expectedTerminalWidth := 6 // "あいう" = 3 chars × 2 width = 6
	if screenCol != expectedTerminalWidth {
		t.Errorf("Expected screen cursor at width %d, got %d", expectedTerminalWidth, screenCol)
	}
}

/**
 * @spec cursor/japanese_progression
 * @scenario 日本語文字連続入力時のカーソル進行
 * @description 日本語文字を連続して入力した際のカーソル位置の步進的進行
 * @given エディタを新規作成する
 * @when 日本語文字（あ、い、う、え、お）を1文字ずつ順次入力
 * @then 各文字の入力後にバイト位置とターミナル表示位置が正確に進行する
 * @implementation domain/cursor.go, 文字幅計算
 */
func TestCursorPositionProgression(t *testing.T) {
	editor := domain.NewEditor()
	
	testChars := []rune{'あ', 'い', 'う', 'え', 'お'}
	
	for i, ch := range testChars {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		window := editor.CurrentWindow()
		screenRow, screenCol := window.CursorPosition()
		
		expectedTerminalWidth := (i + 1) * 2 // Each Japanese character is 2 terminal columns
		expectedBytePos := (i + 1) * 3        // Each Japanese character is 3 bytes
		
		t.Logf("After inserting %c: byte pos=%d, screen pos=(%d,%d)", 
			ch, cursor.Col, screenRow, screenCol)
		
		if cursor.Col != expectedBytePos {
			t.Errorf("After %c: expected byte pos %d, got %d", ch, expectedBytePos, cursor.Col)
		}
		
		if screenCol != expectedTerminalWidth {
			t.Errorf("After %c: expected screen col %d, got %d", ch, expectedTerminalWidth, screenCol)
		}
	}
}

/**
 * @spec cursor/mixed_ascii_japanese
 * @scenario ASCIIと日本語混在カーソル位置
 * @description ASCII文字と日本語文字が混在するテキストでのカーソル位置計算
 * @given エディタを新規作成する
 * @when "aあiい"（ASCIIと日本語の混在）を順次入力
 * @then 各文字タイプのバイト数と表示幅の違いを正確に処理してカーソル位置が計算される
 * @implementation domain/cursor.go, 混合文字列処理
 */
func TestMixedASCIIJapaneseCursor(t *testing.T) {
	editor := domain.NewEditor()
	
	// Test "aあiい"
	chars := []rune{'a', 'あ', 'i', 'い'}
	expectedBytes := []int{1, 4, 5, 8}         // a(1) + あ(3) + i(1) + い(3)
	expectedTerminalWidths := []int{1, 3, 4, 6} // a(1) + あ(2) + i(1) + い(2)
	
	for i, ch := range chars {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
		
		buffer := editor.CurrentBuffer()
		cursor := buffer.Cursor()
		window := editor.CurrentWindow()
		screenRow, screenCol := window.CursorPosition()
		
		t.Logf("After inserting %c: byte pos=%d, screen pos=(%d,%d)", 
			ch, cursor.Col, screenRow, screenCol)
		
		if cursor.Col != expectedBytes[i] {
			t.Errorf("After %c: expected byte pos %d, got %d", ch, expectedBytes[i], cursor.Col)
		}
		
		if screenCol != expectedTerminalWidths[i] {
			t.Errorf("After %c: expected screen col %d, got %d", ch, expectedTerminalWidths[i], screenCol)
		}
	}
}