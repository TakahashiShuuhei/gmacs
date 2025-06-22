package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec commands/mx_basic
 * @scenario M-xコマンドの基本動作
 * @description M-xコマンドモードの有効化とミニバッファ状態の確認
 * @given エディタを新規作成し、通常モードで起動
 * @when ESCキーを押し、続いてxキーを押下（M-x）
 * @then ミニバッファがアクティブになり、"M-x "プロンプトが表示される
 * @implementation domain/commands.go, domain/minibuffer.go
 */
func TestMxCommandBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Start with normal mode
	display.Render(editor)
	modeLine := display.GetModeLine()
	expectedModeLine := " *scratch* " + strings.Repeat("-", 69) // 80 - 11 = 69
	if modeLine != expectedModeLine {
		t.Errorf("Expected normal mode line, got %q", modeLine)
	}
	
	// Press ESC (Meta key)
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	
	// Press x for M-x
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Check minibuffer is active
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active after M-x")
	}
	
	if minibuffer.Mode() != domain.MinibufferCommand {
		t.Error("Minibuffer should be in command mode")
	}
	
	// Render and check prompt
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedPrompt := "M-x " + strings.Repeat(" ", 76)
	if minibufferContent != expectedPrompt {
		t.Errorf("Expected M-x prompt, got %q", minibufferContent)
	}
	
	// Check cursor position (should be after "M-x ")
	cursorRow, cursorCol := display.GetCursorPosition()
	if cursorRow != 4 || cursorCol != 4 { // height-1 = 4, after "M-x "
		t.Errorf("Expected cursor at (4, 4), got (%d, %d)", cursorRow, cursorCol)
	}
}

/**
 * @spec commands/mx_version
 * @scenario M-x versionコマンドの実行
 * @description M-x versionコマンドでバージョン情報を表示
 * @given M-xコマンドモードを有効化
 * @when "version"を入力してEnterキーを押下
 * @then バージョンメッセージがミニバッファに表示される
 * @implementation domain/commands.go, domain/minibuffer.go
 */
func TestMxVersionCommand(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type "version"
	versionText := "version"
	for _, ch := range versionText {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer content
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedContent := "M-x version" + strings.Repeat(" ", 69)
	if minibufferContent != expectedContent {
		t.Errorf("Expected 'M-x version', got %q", minibufferContent)
	}
	
	// Press Enter to execute
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Check that command was executed and minibuffer shows version message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should still be active showing version message")
	}
	
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Minibuffer should be in message mode")
	}
	
	minibufferContent = display.GetMinibuffer()
	expectedVersion := "gmacs 0.1.0 - Emacs-like text editor in Go"
	expectedPadding := strings.Repeat(" ", 80-len(expectedVersion))
	expectedLine := expectedVersion + expectedPadding
	if minibufferContent != expectedLine {
		t.Errorf("Expected version message, got %q", minibufferContent)
	}
	
	// Any key should clear the message and insert into buffer
	anyKeyEvent := events.KeyEventData{Key: "a", Rune: 'a'}
	editor.HandleEvent(anyKeyEvent)
	
	// Should return to normal mode and insert the character
	display.Render(editor)
	if editor.Minibuffer().IsActive() {
		t.Error("Minibuffer should be cleared after any key")
	}
	
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent := ""
	if len(content) > 0 {
		actualContent = strings.TrimRight(content[0], " ")
	}
	if actualContent != "a" {
		t.Errorf("Expected 'a' to be inserted, got content: %q", actualContent)
	}
}

/**
 * @spec commands/mx_unknown
 * @scenario 未知のM-xコマンドのエラー処理
 * @description 存在しないコマンドを実行した際のエラーハンドリング
 * @given M-xコマンドモードを有効化
 * @when 存在しないコマンド"nonexistent"を入力してEnterを押下
 * @then エラーメッセージがミニバッファに表示される
 * @implementation domain/commands.go, エラー処理
 */
func TestMxUnknownCommand(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Start M-x and type unknown command
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type "nonexistent"
	for _, ch := range "nonexistent" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Should show error message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should show error message")
	}
	
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Minibuffer should be in message mode")
	}
	
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Unknown command: nonexistent") {
		t.Errorf("Expected unknown command error, got %q", minibufferContent)
	}
}

/**
 * @spec commands/mx_cancel
 * @scenario M-xコマンドのキャンセル
 * @description ESCキーでM-xコマンドをキャンセルする機能
 * @given M-xコマンドモードで部分的にコマンドを入力済み
 * @when ESCキーを押してキャンセルする
 * @then ミニバッファがクリアされ、通常モードに戻る
 * @implementation domain/commands.go, キャンセル処理
 */
func TestMxCancel(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type some text
	for _, ch := range "ver" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer is active with content
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active")
	}
	if minibuffer.Content() != "ver" {
		t.Errorf("Expected content 'ver', got %q", minibuffer.Content())
	}
	
	// Press Escape to cancel
	cancelEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(cancelEvent)
	
	// Minibuffer should be cleared
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be cleared after cancel")
	}
}

/**
 * @spec commands/mx_list_commands
 * @scenario M-x list-commandsコマンドの実行
 * @description 利用可能なコマンド一覧を表示する機能
 * @given M-xコマンドモードを有効化
 * @when "list-commands"を入力してEnterキーを押下
 * @then 利用可能なコマンド一覧がミニバッファに表示される
 * @implementation domain/commands.go, コマンド一覧機能
 */
func TestMxListCommands(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(200, 5) // Wider display to show all commands
	
	// Execute list-commands
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	for _, ch := range "list-commands" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Check message contains available commands
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Available commands:") {
		t.Errorf("Expected command list message, got %q", minibufferContent)
	}
	
	// Check that some key commands are present in the visible portion
	// Note: Output may be truncated due to display width limitations
	// Commands are now sorted alphabetically, so check for commands at the beginning
	expectedCommands := []string{"auto-a-mode"}  // Should be one of the first alphabetically
	for _, cmd := range expectedCommands {
		if !strings.Contains(minibufferContent, cmd) {
			t.Errorf("Expected %s command in list, got %q", cmd, minibufferContent)
		}
	}
	
	// Check that at least some commands are listed (not empty)
	if len(minibufferContent) < 50 {
		t.Errorf("Command list seems too short, got %q", minibufferContent)
	}
	
	// Verify the message starts correctly
	if !strings.HasPrefix(minibufferContent, "Available commands:") {
		t.Errorf("Expected message to start with 'Available commands:', got %q", minibufferContent)
	}
}

/**
 * @spec commands/mx_clear_buffer
 * @scenario M-x clear-bufferコマンドの実行
 * @description バッファの内容を全てクリアする機能
 * @given バッファに"hello world"を入力済み
 * @when M-x clear-bufferコマンドを実行
 * @then バッファが空になり、クリアメッセージが表示される
 * @implementation domain/commands.go, domain/buffer.go
 */
func TestMxClearBuffer(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Add some content to buffer
	for _, ch := range "hello world" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "hello world" {
		t.Errorf("Expected 'hello world', got %q", actualContent)
	}
	
	// Execute clear-buffer
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	for _, ch := range "clear-buffer" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Buffer should be cleared
	display.Render(editor)
	content = display.GetContent()
	// Trim trailing spaces for comparison
	actualContent = strings.TrimRight(content[0], " ")
	if actualContent != "" {
		t.Errorf("Expected empty buffer, got %q", actualContent)
	}
	
	// Should show clear message
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Buffer cleared") {
		t.Errorf("Expected buffer cleared message, got %q", minibufferContent)
	}
}