package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestBasicTextInput(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	testText := "Hello, World!"
	
	for _, ch := range testText {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line after text input")
	}
	
	if lines[0] != testText {
		t.Errorf("Expected '%s', got '%s'", testText, lines[0])
	}
}

func TestEnterKeyNewline(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	editor.HandleEvent(events.KeyEventData{Rune: 'H', Key: "H"})
	editor.HandleEvent(events.KeyEventData{Rune: 'i', Key: "i"})
	editor.HandleEvent(events.KeyEventData{Key: "Enter", Rune: '\n'})
	editor.HandleEvent(events.KeyEventData{Rune: 'W', Key: "W"})
	editor.HandleEvent(events.KeyEventData{Rune: 'o', Key: "o"})
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines after Enter, got %d", len(lines))
	}
	
	if lines[0] != "Hi" {
		t.Errorf("Expected first line 'Hi', got '%s'", lines[0])
	}
	
	if lines[1] != "Wo" {
		t.Errorf("Expected second line 'Wo', got '%s'", lines[1])
	}
}

func TestMultilineTextInput(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	testLines := []string{"First line", "Second line", "Third line"}
	
	for i, line := range testLines {
		for _, ch := range line {
			editor.HandleEvent(events.KeyEventData{Rune: ch, Key: string(ch)})
		}
		if i < len(testLines)-1 {
			editor.HandleEvent(events.KeyEventData{Key: "Enter", Rune: '\n'})
		}
	}
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) != len(testLines) {
		t.Fatalf("Expected %d lines, got %d", len(testLines), len(lines))
	}
	
	for i, expected := range testLines {
		if lines[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}
}

func TestJapaneseTextInput(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	testText := "こんにちは世界"
	
	for _, ch := range testText {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line after Japanese text input")
	}
	
	// Unicode handling might need improvement
	t.Logf("Input: '%s', Output: '%s'", testText, lines[0])
	
	if len([]rune(lines[0])) != len([]rune(testText)) {
		t.Errorf("Expected %d characters, got %d", len([]rune(testText)), len([]rune(lines[0])))
	}
}