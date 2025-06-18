package test

import (
	"testing"
	"unicode/utf8"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

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