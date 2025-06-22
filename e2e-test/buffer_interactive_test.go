package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec buffer/switch_to_buffer_basic
 * @scenario C-x bによる基本的なバッファ切り替え
 * @description C-x bキーシーケンスでバッファ切り替えモードを開始する機能
 * @given エディタに複数のバッファが存在している状態
 * @when C-xを押し、続いてbキーを押下する
 * @then ミニバッファがアクティブになり、"Switch to buffer: "プロンプトが表示される
 * @implementation domain/buffer_interactive.go, domain/editor.go
 */
func TestSwitchToBufferBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Create additional buffers
	buffer2 := domain.NewBuffer("test-buffer")
	editor.AddBuffer(buffer2)
	buffer3 := domain.NewBuffer("another-buffer")
	editor.AddBuffer(buffer3)
	
	// Initially should be in *scratch* buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "*scratch*" {
		t.Fatalf("Expected initial buffer '*scratch*', got %q", currentBuffer.Name())
	}
	
	// When: Press C-x
	ctrlXEvent := events.KeyEventData{
		Key:  "x",
		Ctrl: true,
	}
	editor.HandleEvent(ctrlXEvent)
	
	// When: Press b
	bEvent := events.KeyEventData{
		Key:  "b",
		Rune: 'b',
	}
	editor.HandleEvent(bEvent)
	
	// Then: Minibuffer should be active with buffer selection prompt
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should be active after C-x b")
	}
	
	if minibuffer.Mode() != domain.MinibufferBufferSelection {
		t.Errorf("Expected MinibufferBufferSelection mode, got %v", minibuffer.Mode())
	}
	
	// Check prompt
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedPrompt := "Switch to buffer: " + strings.Repeat(" ", 62) // 80 - 18 = 62
	if minibufferContent != expectedPrompt {
		t.Errorf("Expected switch to buffer prompt, got %q", minibufferContent)
	}
	
	// The cursor position in minibuffer should be after the prompt
	// (Testing cursor position is display-specific and can be skipped for core functionality)
}

/**
 * @spec buffer/switch_to_buffer_existing
 * @scenario 既存バッファへの切り替え
 * @description 存在するバッファ名を入力してバッファを切り替える機能
 * @given C-x bでバッファ切り替えモードを開始済み
 * @when 既存のバッファ名"test-buffer"を入力してEnterキーを押下
 * @then 指定したバッファに切り替わり、成功メッセージが表示される
 * @implementation domain/buffer_interactive.go
 */
func TestSwitchToBufferExisting(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Create test buffer
	testBuffer := domain.NewBuffer("test-buffer")
	testBuffer.InsertChar('t')
	testBuffer.InsertChar('e')
	testBuffer.InsertChar('s')
	testBuffer.InsertChar('t')
	editor.AddBuffer(testBuffer)
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// When: Type buffer name "test-buffer"
	bufferName := "test-buffer"
	for _, ch := range bufferName {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer shows typed name
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedContent := "Switch to buffer: test-buffer" + strings.Repeat(" ", 51)
	if minibufferContent != expectedContent {
		t.Errorf("Expected buffer name in minibuffer, got %q", minibufferContent)
	}
	
	// When: Press Enter to switch
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Then: Should switch to test-buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "test-buffer" {
		t.Errorf("Expected current buffer 'test-buffer', got %q", currentBuffer.Name())
	}
	
	// Then: Should show success message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() || minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Should show success message in minibuffer")
	}
	
	minibufferContent = display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Switched to buffer: test-buffer") {
		t.Errorf("Expected success message, got %q", minibufferContent)
	}
	
	// Then: Buffer content should be preserved
	display.Render(editor)
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "test" {
		t.Errorf("Expected buffer content 'test', got %q", actualContent)
	}
}

/**
 * @spec buffer/switch_to_buffer_new
 * @scenario 新規バッファの作成と切り替え
 * @description 存在しないバッファ名を入力して新しいバッファを作成する機能
 * @given C-x bでバッファ切り替えモードを開始済み
 * @when 存在しないバッファ名"new-buffer"を入力してEnterキーを押下
 * @then 新しいバッファが作成され、そのバッファに切り替わる
 * @implementation domain/buffer_interactive.go
 */
func TestSwitchToBufferNew(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	initialBufferCount := len(editor.GetBufferNames())
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// When: Type new buffer name
	newBufferName := "new-buffer"
	for _, ch := range newBufferName {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// When: Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Then: Should create and switch to new buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "new-buffer" {
		t.Errorf("Expected current buffer 'new-buffer', got %q", currentBuffer.Name())
	}
	
	// Then: Buffer count should increase
	newBufferCount := len(editor.GetBufferNames())
	if newBufferCount != initialBufferCount+1 {
		t.Errorf("Expected %d buffers, got %d", initialBufferCount+1, newBufferCount)
	}
	
	// Then: New buffer should be empty
	display.Render(editor)
	content := display.GetContent()
	// Trim trailing spaces for comparison
	actualContent := strings.TrimRight(content[0], " ")
	if actualContent != "" {
		t.Errorf("Expected empty new buffer, got %q", actualContent)
	}
}

/**
 * @spec buffer/switch_to_buffer_empty
 * @scenario 空文字入力でのバッファ切り替えキャンセル
 * @description バッファ名を入力せずにEnterを押した場合の動作
 * @given C-x bでバッファ切り替えモードを開始済み
 * @when 何も入力せずにEnterキーを押下
 * @then 現在のバッファのまま変更されず、ミニバッファがクリアされる
 * @implementation domain/buffer_interactive.go
 */
func TestSwitchToBufferEmpty(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	originalBuffer := editor.CurrentBuffer()
	originalBufferName := originalBuffer.Name()
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// When: Press Enter without typing anything
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Then: Should stay in same buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != originalBufferName {
		t.Errorf("Expected to stay in buffer %q, got %q", originalBufferName, currentBuffer.Name())
	}
	
	// Then: Minibuffer should be cleared
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be cleared after empty Enter")
	}
}

/**
 * @spec buffer/switch_to_buffer_cancel
 * @scenario C-x bのキャンセル機能
 * @description Escapeキーでバッファ切り替えをキャンセルする機能
 * @given C-x bでバッファ切り替えモードを開始し、部分的に名前を入力済み
 * @when Escapeキーを押下
 * @then バッファ切り替えがキャンセルされ、ミニバッファがクリアされる
 * @implementation domain/buffer_interactive.go
 */
func TestSwitchToBufferCancel(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	originalBuffer := editor.CurrentBuffer()
	originalBufferName := originalBuffer.Name()
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// Type partial buffer name
	for _, ch := range "test" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Check minibuffer has content
	minibuffer := editor.Minibuffer()
	if minibuffer.Content() != "test" {
		t.Errorf("Expected minibuffer content 'test', got %q", minibuffer.Content())
	}
	
	// When: Press Escape to cancel
	escapeEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escapeEvent)
	
	// Then: Should stay in original buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != originalBufferName {
		t.Errorf("Expected to stay in buffer %q, got %q", originalBufferName, currentBuffer.Name())
	}
	
	// Then: Minibuffer should be cleared
	if minibuffer.IsActive() {
		t.Error("Minibuffer should be cleared after cancel")
	}
}

/**
 * @spec buffer/list_buffers_basic
 * @scenario C-x C-bによるバッファ一覧表示
 * @description C-x C-bキーシーケンスでバッファ一覧を表示する機能
 * @given エディタに複数のバッファが存在している状態
 * @when C-xを押し、続いてC-bキーを押下する
 * @then ミニバッファにバッファ一覧と現在のバッファが表示される
 * @implementation domain/buffer_interactive.go
 */
func TestListBuffersBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(200, 5) // Wide display for buffer list
	
	// Create additional buffers
	buffer2 := domain.NewBuffer("buffer-1")
	editor.AddBuffer(buffer2)
	buffer3 := domain.NewBuffer("buffer-2")
	editor.AddBuffer(buffer3)
	
	// Switch to buffer-1 to make it current
	editor.SwitchToBuffer(buffer2)
	
	// When: Press C-x
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	
	// When: Press C-b
	ctrlBEvent := events.KeyEventData{Key: "b", Ctrl: true}
	editor.HandleEvent(ctrlBEvent)
	
	// Then: Should switch to *Buffer List* buffer and display buffer list
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "*Buffer List*" {
		t.Errorf("Expected to switch to '*Buffer List*' buffer, got %q", currentBuffer.Name())
	}
	
	// Render and check buffer list content
	display.Render(editor)
	content := display.GetContent()
	
	// Should have header line
	if len(content) == 0 || !strings.Contains(content[0], "CRM Buffer") {
		t.Errorf("Expected header line with 'CRM Buffer', got content: %v", content)
	}
	
	// Should contain all buffer names in the content
	allContent := strings.Join(content, "\n")
	expectedBuffers := []string{"*scratch*", "buffer-1", "buffer-2", "*Buffer List*"}
	for _, bufName := range expectedBuffers {
		if !strings.Contains(allContent, bufName) {
			t.Errorf("Expected buffer name %q in list, got content: %s", bufName, allContent)
		}
	}
	
	// Should mark current buffer with "."
	// Note: buffer-1 was current before switching to *Buffer List*
	if !strings.Contains(allContent, ".   buffer-1") {
		t.Errorf("Expected current buffer to be marked with '.', got content: %s", allContent)
	}
}

/**
 * @spec buffer/kill_buffer_basic
 * @scenario C-x kによる基本的なバッファ削除
 * @description C-x kキーシーケンスで現在のバッファを削除する機能
 * @given エディタに複数のバッファが存在し、任意のバッファを選択中
 * @when C-xを押し、続いてkキーを押下する
 * @then 現在のバッファが削除され、他のバッファに切り替わる
 * @implementation domain/buffer_interactive.go
 */
func TestKillBufferBasic(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Create additional buffer and switch to it
	testBuffer := domain.NewBuffer("test-buffer")
	editor.AddBuffer(testBuffer)
	editor.SwitchToBuffer(testBuffer)
	
	// Verify we're in test-buffer
	if editor.CurrentBuffer().Name() != "test-buffer" {
		t.Fatalf("Expected current buffer 'test-buffer', got %q", editor.CurrentBuffer().Name())
	}
	
	initialBufferCount := len(editor.GetBufferNames())
	
	// When: Press C-x k
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	kEvent := events.KeyEventData{Key: "k", Rune: 'k'}
	editor.HandleEvent(kEvent)
	
	// Then: Buffer should be killed
	newBufferCount := len(editor.GetBufferNames())
	if newBufferCount != initialBufferCount-1 {
		t.Errorf("Expected %d buffers after kill, got %d", initialBufferCount-1, newBufferCount)
	}
	
	// Then: Should switch to remaining buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() == "test-buffer" {
		t.Error("Should not still be in killed buffer")
	}
	
	// Then: Should show kill message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() || minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Should show kill message in minibuffer")
	}
	
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Killed buffer: test-buffer") {
		t.Errorf("Expected kill message, got %q", minibufferContent)
	}
}

/**
 * @spec buffer/kill_buffer_last
 * @scenario 最後のバッファ削除の防止
 * @description 最後の1つのバッファを削除しようとした場合のエラー処理
 * @given エディタに1つのバッファのみ存在している状態
 * @when C-x kキーシーケンスでバッファ削除を試行
 * @then 削除が拒否され、エラーメッセージが表示される
 * @implementation domain/buffer_interactive.go, エラー処理
 */
func TestKillBufferLast(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Should only have *scratch* buffer initially
	if len(editor.GetBufferNames()) != 1 {
		t.Fatalf("Expected 1 buffer initially, got %d", len(editor.GetBufferNames()))
	}
	
	// When: Try to kill the last buffer
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	kEvent := events.KeyEventData{Key: "k", Rune: 'k'}
	editor.HandleEvent(kEvent)
	
	// Then: Buffer should not be killed
	if len(editor.GetBufferNames()) != 1 {
		t.Errorf("Expected 1 buffer to remain, got %d", len(editor.GetBufferNames()))
	}
	
	// Then: Should still be in same buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "*scratch*" {
		t.Errorf("Expected to stay in '*scratch*', got %q", currentBuffer.Name())
	}
	
	// Then: Should show error message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() || minibuffer.Mode() != domain.MinibufferMessage {
		t.Error("Should show error message in minibuffer")
	}
	
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Cannot kill the last buffer") {
		t.Errorf("Expected error message, got %q", minibufferContent)
	}
}

/**
 * @spec buffer/tab_completion_single
 * @scenario バッファ名の単一マッチ補完
 * @description Tabキーによるバッファ名の自動補完機能（単一マッチ）
 * @given C-x bでバッファ切り替えモード開始し、一意に決まる部分文字列を入力済み
 * @when Tabキーを押下
 * @then バッファ名が自動的に完全な名前まで補完される
 * @implementation domain/buffer_interactive.go, 補完機能
 */
func TestBufferTabCompletionSingle(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 5)
	
	// Create buffers with distinct prefixes
	testBuffer := domain.NewBuffer("test-buffer")
	editor.AddBuffer(testBuffer)
	anotherBuffer := domain.NewBuffer("another-buffer")
	editor.AddBuffer(anotherBuffer)
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// Type partial name that matches only one buffer
	for _, ch := range "tes" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// When: Press Tab for completion
	tabEvent := events.KeyEventData{Key: "Tab", Rune: '\t'}
	editor.HandleEvent(tabEvent)
	
	// Then: Should complete to full buffer name
	minibuffer := editor.Minibuffer()
	if minibuffer.Content() != "test-buffer" {
		t.Errorf("Expected completion to 'test-buffer', got %q", minibuffer.Content())
	}
	
	// Check display shows completed name
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedContent := "Switch to buffer: test-buffer" + strings.Repeat(" ", 51)
	if minibufferContent != expectedContent {
		t.Errorf("Expected completed name in display, got %q", minibufferContent)
	}
}

/**
 * @spec buffer/tab_completion_multiple
 * @scenario バッファ名の複数マッチ補完
 * @description Tabキーによるバッファ名の自動補完機能（複数マッチ）
 * @given C-x bでバッファ切り替えモード開始し、複数にマッチする部分文字列を入力済み
 * @when Tabキーを押下
 * @then 共通部分まで補完され、マッチした候補一覧が表示される
 * @implementation domain/buffer_interactive.go, 補完機能
 */
func TestBufferTabCompletionMultiple(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(200, 5) // Wide display for matches
	
	// Create buffers with common prefix
	testBuffer1 := domain.NewBuffer("test-buffer-1")
	editor.AddBuffer(testBuffer1)
	testBuffer2 := domain.NewBuffer("test-buffer-2")
	editor.AddBuffer(testBuffer2)
	testFile := domain.NewBuffer("test-file")
	editor.AddBuffer(testFile)
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// Type partial name that matches multiple buffers
	for _, ch := range "test-" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// When: Press Tab for completion
	tabEvent := events.KeyEventData{Key: "Tab", Rune: '\t'}
	editor.HandleEvent(tabEvent)
	
	// Then: Should complete to common prefix and show matches message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	
	// When there are multiple matches with no further common prefix, 
	// it should show the matches message instead of completing
	if minibuffer.Mode() != domain.MinibufferMessage {
		t.Errorf("Expected MinibufferMessage mode after multiple matches, got %v", minibuffer.Mode())
	}
	
	// Check that matches message is shown
	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Matches:") {
		t.Errorf("Expected matches message, got %q", minibufferContent)
	}
	
	// Should show the matching buffers
	expectedMatches := []string{"test-buffer-1", "test-buffer-2", "test-file"}
	for _, match := range expectedMatches {
		if !strings.Contains(minibufferContent, match) {
			t.Errorf("Expected match %q in message, got %q", match, minibufferContent)
		}
	}
}

/**
 * @spec buffer/mx_commands
 * @scenario M-xコマンドによるバッファ操作
 * @description M-xコマンドでバッファ関連の操作を実行する機能
 * @given エディタに複数のバッファが存在している状態
 * @when M-x switch-to-buffer, M-x list-buffers, M-x kill-bufferを実行
 * @then キーバインドと同等の動作が実行される
 * @implementation domain/buffer_interactive.go, M-xコマンドシステム
 */
func TestBufferMxCommands(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(200, 5)
	
	// Create test buffer
	testBuffer := domain.NewBuffer("mx-test-buffer")
	editor.AddBuffer(testBuffer)
	
	// Test M-x switch-to-buffer
	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type command
	commandName := "switch-to-buffer"
	for _, ch := range commandName {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Execute command
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Should be in buffer selection mode
	minibuffer := editor.Minibuffer()
	if minibuffer.Mode() != domain.MinibufferBufferSelection {
		t.Error("M-x switch-to-buffer should activate buffer selection mode")
	}
	
	// Cancel and test list-buffers
	escEvent = events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	
	// Test M-x list-buffers
	escEvent = events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent = events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	commandName = "list-buffers"
	for _, ch := range commandName {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	// Check that we switched to *Buffer List* buffer
	currentBuffer := editor.CurrentBuffer()
	if currentBuffer.Name() != "*Buffer List*" {
		t.Errorf("M-x list-buffers should switch to *Buffer List* buffer, got %q", currentBuffer.Name())
	}
	
	// Check buffer list content
	display.Render(editor)
	content := display.GetContent()
	allContent := strings.Join(content, "\n")
	if !strings.Contains(allContent, "CRM Buffer") {
		t.Errorf("M-x list-buffers should show buffer list header, got %q", allContent)
	}
}

/**
 * @spec buffer/minibuffer_editing
 * @scenario バッファ選択モードでのミニバッファ編集
 * @description バッファ選択モードでのカーソル移動と編集機能
 * @given C-x bでバッファ選択モードを開始し、バッファ名を部分入力済み
 * @when C-f, C-b, C-a, C-e, C-h, C-dキーで編集操作を実行
 * @then ミニバッファ内でカーソル移動と文字削除が正常に動作する
 * @implementation domain/buffer_interactive.go, ミニバッファ編集
 */
func TestBufferMinibufferEditing(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// Start C-x b
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	bEvent := events.KeyEventData{Key: "b", Rune: 'b'}
	editor.HandleEvent(bEvent)
	
	// Type some text
	for _, ch := range "test-buffer" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	minibuffer := editor.Minibuffer()
	if minibuffer.Content() != "test-buffer" {
		t.Fatalf("Expected 'test-buffer', got %q", minibuffer.Content())
	}
	
	// Test C-a (beginning of line)
	ctrlAEvent := events.KeyEventData{Key: "a", Ctrl: true}
	editor.HandleEvent(ctrlAEvent)
	if minibuffer.CursorPosition() != 0 {
		t.Errorf("Expected cursor at position 0 after C-a, got %d", minibuffer.CursorPosition())
	}
	
	// Test C-e (end of line)
	ctrlEEvent := events.KeyEventData{Key: "e", Ctrl: true}
	editor.HandleEvent(ctrlEEvent)
	if minibuffer.CursorPosition() != len("test-buffer") {
		t.Errorf("Expected cursor at end after C-e, got %d", minibuffer.CursorPosition())
	}
	
	// Test C-b (backward char)
	ctrlBEvent := events.KeyEventData{Key: "b", Ctrl: true}
	editor.HandleEvent(ctrlBEvent)
	if minibuffer.CursorPosition() != len("test-buffer")-1 {
		t.Errorf("Expected cursor moved back, got %d", minibuffer.CursorPosition())
	}
	
	// Test C-f (forward char)
	ctrlFEvent := events.KeyEventData{Key: "f", Ctrl: true}
	editor.HandleEvent(ctrlFEvent)
	if minibuffer.CursorPosition() != len("test-buffer") {
		t.Errorf("Expected cursor moved forward, got %d", minibuffer.CursorPosition())
	}
	
	// Test C-h (delete backward)
	ctrlHEvent := events.KeyEventData{Key: "h", Ctrl: true}
	editor.HandleEvent(ctrlHEvent)
	if minibuffer.Content() != "test-buffe" {
		t.Errorf("Expected 'test-buffe' after C-h, got %q", minibuffer.Content())
	}
}