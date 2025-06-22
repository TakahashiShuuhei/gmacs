package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec file/find_file_basic
 * @scenario C-x C-f による基本的なファイル開く機能
 * @description ファイルパスを入力してファイルを開く基本機能
 * @given 存在するテストファイルを用意
 * @when C-x C-f コマンドでファイルパスを入力
 * @then ファイルの内容がバッファに読み込まれ、適切に表示される
 * @implementation domain/buffer.go, NewBufferFromFile関数
 */
func TestFindFileBasic(t *testing.T) {
	// テストファイルを作成
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := "Hello, World!\nSecond line\nThird line"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	editor := NewEditorWithDefaults()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// ミニバッファがファイル入力モードになっているかチェック
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferFile {
		t.Errorf("Expected MinibufferFile mode, got %v", minibuffer.Mode())
	}
	if minibuffer.Prompt() != "Find file: " {
		t.Errorf("Expected 'Find file: ' prompt, got %q", minibuffer.Prompt())
	}
	
	// ファイルパスを入力
	for _, ch := range testFile {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enterキーでファイルを開く
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	// ファイルが正しく読み込まれたかチェック
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		t.Fatal("No current buffer")
	}
	
	if buffer.Name() != "test.txt" {
		t.Errorf("Expected buffer name 'test.txt', got %q", buffer.Name())
	}
	
	if buffer.Filepath() != testFile {
		t.Errorf("Expected filepath %q, got %q", testFile, buffer.Filepath())
	}
	
	bufferContent := buffer.Content()
	expectedLines := []string{"Hello, World!", "Second line", "Third line"}
	
	if len(bufferContent) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(bufferContent))
	}
	
	for i, expected := range expectedLines {
		if i < len(bufferContent) && bufferContent[i] != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, bufferContent[i])
		}
	}
}

/**
 * @spec file/find_file_nonexistent
 * @scenario 存在しないファイルを開こうとした場合
 * @description 存在しないファイルパスでC-x C-fを実行した際のエラーハンドリング
 * @given 存在しないファイルパス
 * @when C-x C-f コマンドで存在しないファイルパスを入力
 * @then エラーメッセージが表示され、現在のバッファは変更されない
 * @implementation domain/editor.go, エラーハンドリング
 */
func TestFindFileNonexistent(t *testing.T) {
	editor := NewEditorWithDefaults()
	originalBuffer := editor.CurrentBuffer()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// 存在しないファイルパスを入力
	nonexistentFile := "/path/that/does/not/exist.txt"
	for _, ch := range nonexistentFile {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enterキーでファイルを開こうとする
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	// 元のバッファが維持されているかチェック
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer != originalBuffer {
		t.Error("Buffer should not have changed for nonexistent file")
	}
	
	// エラーメッセージが表示されているかチェック
	minibuffer := editor.Minibuffer()
	message := minibuffer.Message()
	if message == "" {
		t.Error("Expected error message for nonexistent file")
	}
	if !contains(message, "Cannot open file") {
		t.Errorf("Expected error message about 'Cannot open file', got %q", message)
	}
}

/**
 * @spec file/find_file_empty
 * @scenario 空ファイルを開く場合
 * @description 空のファイルを開いた際の適切な処理
 * @given 空のファイル
 * @when C-x C-f コマンドで空ファイルを開く
 * @then 空行が1行あるバッファが作成される
 * @implementation domain/buffer.go, 空ファイル処理
 */
func TestFindFileEmpty(t *testing.T) {
	// 空のテストファイルを作成
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.txt")
	
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	
	editor := NewEditorWithDefaults()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// ファイルパスを入力
	for _, ch := range emptyFile {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enterキーでファイルを開く
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	// 空ファイルが正しく処理されたかチェック
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		t.Fatal("No current buffer")
	}
	
	if buffer.Name() != "empty.txt" {
		t.Errorf("Expected buffer name 'empty.txt', got %q", buffer.Name())
	}
	
	content := buffer.Content()
	if len(content) != 1 {
		t.Errorf("Expected 1 line for empty file, got %d", len(content))
	}
	
	if content[0] != "" {
		t.Errorf("Expected empty line, got %q", content[0])
	}
	
	// カーソルが正しい位置にあるかチェック
	cursor := buffer.Cursor()
	if cursor.Row != 0 || cursor.Col != 0 {
		t.Errorf("Expected cursor at (0,0), got (%d,%d)", cursor.Row, cursor.Col)
	}
}

/**
 * @spec file/find_file_cancel
 * @scenario C-x C-f のキャンセル
 * @description Escapeキーでファイル入力をキャンセルする機能
 * @given C-x C-f を実行してファイル入力モードに入る
 * @when Escapeキーを押す
 * @then ミニバッファがクリアされ、元の状態に戻る
 * @implementation domain/editor.go, キャンセル処理
 */
func TestFindFileCancel(t *testing.T) {
	editor := NewEditorWithDefaults()
	originalBuffer := editor.CurrentBuffer()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// ファイル入力モードになっているかチェック
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferFile {
		t.Errorf("Expected MinibufferFile mode, got %v", minibuffer.Mode())
	}
	
	// 何か入力
	for _, ch := range "some/path" {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Escapeでキャンセル
	event := events.KeyEventData{Key: "Escape"}
	editor.HandleEvent(event)
	
	// ミニバッファがクリアされたかチェック
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be inactive after cancel")
	}
	
	// 元のバッファが維持されているかチェック
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer != originalBuffer {
		t.Error("Buffer should not have changed after cancel")
	}
}

/**
 * @spec file/find_file_japanese
 * @scenario 日本語ファイル名での動作
 * @description 日本語を含むファイルパスでの正常動作
 * @given 日本語ファイル名のテストファイル
 * @when C-x C-f で日本語ファイル名を入力
 * @then ファイルが正常に開かれる
 * @implementation domain/buffer.go, UTF-8ファイル名対応
 */
func TestFindFileJapanese(t *testing.T) {
	// 日本語ファイル名のテストファイルを作成
	tempDir := t.TempDir()
	japaneseFile := filepath.Join(tempDir, "テスト.txt")
	content := "こんにちは、世界！\n日本語のテキストです。"
	
	err := os.WriteFile(japaneseFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create Japanese test file: %v", err)
	}
	
	editor := NewEditorWithDefaults()
	
	// C-x C-f を実行
	event1 := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(event1)
	event2 := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(event2)
	
	// 日本語ファイルパスを入力
	for _, ch := range japaneseFile {
		event := events.KeyEventData{
			Rune: ch,
			Key:  string(ch),
		}
		editor.HandleEvent(event)
	}
	
	// Enterキーでファイルを開く
	event := events.KeyEventData{Key: "Return", Rune: '\n'}
	editor.HandleEvent(event)
	
	// ファイルが正しく読み込まれたかチェック
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		t.Fatal("No current buffer")
	}
	
	if buffer.Name() != "テスト.txt" {
		t.Errorf("Expected buffer name 'テスト.txt', got %q", buffer.Name())
	}
	
	bufferContent := buffer.Content()
	expectedLines := []string{"こんにちは、世界！", "日本語のテキストです。"}
	
	if len(bufferContent) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(bufferContent))
	}
	
	for i, expected := range expectedLines {
		if i < len(bufferContent) && bufferContent[i] != expected {
			t.Errorf("Line %d: expected %q, got %q", i, expected, bufferContent[i])
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr || 
		  containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}