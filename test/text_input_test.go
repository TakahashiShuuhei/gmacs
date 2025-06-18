package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec input/basic_text
 * @scenario 基本的なテキスト入力
 * @description ASCII文字の連続入力と表示の検証
 * @given エディタを新規作成する
 * @when "Hello, World!"を1文字ずつ入力する
 * @then 入力したテキストが正確に表示される
 * @implementation domain/buffer.go, domain/editor.go
 */
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

/**
 * @spec input/newline
 * @scenario Enter キーによる改行
 * @description Enter キーで行を分割し複数行テキストを作成
 * @given エディタに "Hi" を入力済み
 * @when Enter キーを押して "Wo" を入力する
 * @then 2行に分かれてテキストが表示される
 * @implementation domain/buffer.go, events/key_event.go
 */
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

/**
 * @spec input/multiline
 * @scenario 複数行テキスト入力
 * @description 3行のテキストを順次入力し、行分離を検証
 * @given エディタを新規作成する
 * @when "First line", "Second line", "Third line"を Enter で区切って入力する
 * @then 3行が正確に分かれて表示される
 * @implementation domain/buffer.go, domain/editor.go
 */
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

/**
 * @spec input/japanese
 * @scenario 日本語テキスト入力
 * @description ひらがな文字の入力と表示の検証
 * @given エディタを新規作成する
 * @when "あいう"を文字ごとに入力する
 * @then 日本語テキストが正確に表示される
 * @implementation domain/buffer.go, UTF-8処理
 */
func TestJapaneseTextInput(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	testText := "あいう"
	
	// Process each character from the IME input
	for _, ch := range []rune(testText) {
		charEvent := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(charEvent)
	}
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line after Japanese text input")
	}
	
	t.Logf("Input: '%s', Output: '%s'", testText, lines[0])
	
	if lines[0] != testText {
		t.Errorf("Expected '%s', got '%s'", testText, lines[0])
	}
}