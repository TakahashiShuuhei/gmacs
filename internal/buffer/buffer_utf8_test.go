package buffer

import (
	"testing"
)

func TestInsertChar(t *testing.T) {
	buf := New("test")
	
	testCases := []struct {
		name         string
		char         rune
		line         int
		col          int
		expectedText string
	}{
		{"ASCII character", 'a', 0, 0, "a"},
		{"Japanese hiragana", 'あ', 0, 1, "aあ"},
		{"Japanese katakana", 'ア', 0, 2, "aあア"},
		{"Japanese kanji", '文', 0, 3, "aあア文"},
		{"English after Japanese", 'b', 0, 4, "aあア文b"},
		{"Emoji", '😀', 0, 5, "aあア文b😀"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := buf.InsertChar(tc.line, tc.col, tc.char)
			if err != nil {
				t.Errorf("InsertChar failed: %v", err)
				return
			}
			
			if buf.GetLine(tc.line) != tc.expectedText {
				t.Errorf("Expected line content '%s', got '%s'", tc.expectedText, buf.GetLine(tc.line))
			}
		})
	}
}

func TestInsertCharUTF8Properties(t *testing.T) {
	buf := New("utf8-test")
	
	// Test that rune positions work correctly
	text := "Hello世界"
	for i, char := range text {
		err := buf.InsertChar(0, i, char)
		if err != nil {
			t.Errorf("Failed to insert character %c at position %d: %v", char, i, err)
		}
	}
	
	result := buf.GetLine(0)
	if result != text {
		t.Errorf("Expected '%s', got '%s'", text, result)
	}
	
	// Verify the length in runes vs bytes
	runes := []rune(result)
	if len(runes) != 7 { // Hello=5 + 世=1 + 界=1
		t.Errorf("Expected 7 runes, got %d", len(runes))
	}
}

func TestInsertCharInMiddle(t *testing.T) {
	buf := New("middle-test")
	buf.SetLine(0, "こんばんは")
	
	// Insert in the middle: こん[X]ばんは
	err := buf.InsertChar(0, 2, 'X')
	if err != nil {
		t.Errorf("Failed to insert in middle: %v", err)
	}
	
	expected := "こんXばんは"
	if buf.GetLine(0) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.GetLine(0))
	}
}

func TestInsertString(t *testing.T) {
	buf := New("string-test")
	
	testCases := []struct {
		name         string
		text         string
		line         int
		col          int
		expectedText string
	}{
		{"Empty buffer", "Hello", 0, 0, "Hello"},
		{"Japanese string", "あいうえお", 0, 5, "Helloあいうえお"},
		{"Mixed string", "123漢字", 0, 10, "Helloあいうえお123漢字"},
		{"Insert at beginning", "Start", 0, 0, "StartHelloあいうえお123漢字"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := buf.InsertString(tc.line, tc.col, tc.text)
			if err != nil {
				t.Errorf("InsertString failed: %v", err)
				return
			}
			
			if buf.GetLine(tc.line) != tc.expectedText {
				t.Errorf("Expected line content '%s', got '%s'", tc.expectedText, buf.GetLine(tc.line))
			}
		})
	}
}

func TestInsertStringIMESimulation(t *testing.T) {
	buf := New("ime-test")
	
	// Simulate IME input: user types "あいうえお" in one go
	err := buf.InsertString(0, 0, "あいうえお")
	if err != nil {
		t.Errorf("Failed to insert IME string: %v", err)
	}
	
	expected := "あいうえお"
	if buf.GetLine(0) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, buf.GetLine(0))
	}
	
	// Verify the string is 5 characters (runes) long
	runes := []rune(buf.GetLine(0))
	if len(runes) != 5 {
		t.Errorf("Expected 5 runes, got %d", len(runes))
	}
}