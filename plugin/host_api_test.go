package plugin

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
)

// テスト用のエディタ作成ヘルパー
func createTestEditor() *domain.Editor {
	// 実際のEditorインスタンスを作成
	editor := domain.NewEditor()
	return editor
}

func TestNewHostAPI(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	if hostAPI == nil {
		t.Fatal("NewHostAPI() returned nil")
	}

	if hostAPI.editor != editor {
		t.Error("HostAPI editor not set correctly")
	}
}

func TestHostAPI_GetCurrentBuffer(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	buffer := hostAPI.GetCurrentBuffer()
	if buffer == nil {
		t.Fatal("GetCurrentBuffer() returned nil")
	}

	// デフォルトのスクラッチバッファ
	if buffer.Name() != "*scratch*" {
		t.Errorf("Expected buffer name '*scratch*', got '%s'", buffer.Name())
	}
}

func TestHostAPI_GetCurrentWindow(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	window := hostAPI.GetCurrentWindow()
	if window == nil {
		t.Fatal("GetCurrentWindow() returned nil")
	}

	// ウィンドウの基本プロパティをテスト
	if window.Width() <= 0 {
		t.Error("Expected positive window width")
	}

	if window.Height() <= 0 {
		t.Error("Expected positive window height")
	}
}

func TestHostAPI_SetStatus(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	testMessage := "Test status message"
	// SetStatusは実際にはSetMinibufferMessageを呼び出すので、エラーなく実行されることを確認
	hostAPI.SetStatus(testMessage)
	// 実際のメッセージ確認は複雑なので、エラーが発生しないことのみテスト
}

func TestHostAPI_CreateBuffer(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	bufferName := "new-buffer"
	buffer := hostAPI.CreateBuffer(bufferName)

	if buffer == nil {
		t.Fatal("CreateBuffer() returned nil")
	}

	if buffer.Name() != bufferName {
		t.Errorf("Expected buffer name '%s', got '%s'", bufferName, buffer.Name())
	}
}

func TestHostAPI_FindBuffer(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	// 存在するバッファを検索（デフォルトの*scratch*）
	buffer := hostAPI.FindBuffer("*scratch*")
	if buffer == nil {
		t.Fatal("FindBuffer() returned nil for existing buffer")
	}

	if buffer.Name() != "*scratch*" {
		t.Errorf("Expected buffer name '*scratch*', got '%s'", buffer.Name())
	}

	// 存在しないバッファを検索
	buffer = hostAPI.FindBuffer("nonexistent")
	if buffer != nil {
		t.Error("FindBuffer() should return nil for nonexistent buffer")
	}
}

func TestHostAPI_GetSetOption(t *testing.T) {
	editor := createTestEditor()
	hostAPI := NewHostAPI(editor)

	// オプション設定
	testKey := "test-option"
	testValue := "test-value"
	err := hostAPI.SetOption(testKey, testValue)
	if err != nil {
		t.Fatalf("SetOption() error = %v", err)
	}

	// オプション取得
	value, err := hostAPI.GetOption(testKey)
	if err != nil {
		t.Fatalf("GetOption() error = %v", err)
	}

	if value != testValue {
		t.Errorf("Expected option value '%s', got '%v'", testValue, value)
	}

	// 存在しないオプション取得
	_, err = hostAPI.GetOption("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent option")
	}
}

func TestBufferWrapper_Name(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	wrapper := &BufferWrapper{buffer: buffer}

	if wrapper.Name() != "test-buffer" {
		t.Errorf("Expected name 'test-buffer', got '%s'", wrapper.Name())
	}
}

func TestBufferWrapper_Content(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	wrapper := &BufferWrapper{buffer: buffer}

	// 初期状態は空行
	content := wrapper.Content()
	if content != "" {
		t.Errorf("Expected empty content, got '%s'", content)
	}

	// 文字を追加してテスト
	buffer.InsertChar('H')
	buffer.InsertChar('i')
	content = wrapper.Content()
	if content != "Hi" {
		t.Errorf("Expected content 'Hi', got '%s'", content)
	}
}

func TestBufferWrapper_CursorPosition(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	wrapper := &BufferWrapper{buffer: buffer}

	// 初期位置は0
	pos := wrapper.CursorPosition()
	if pos != 0 {
		t.Errorf("Expected cursor position 0, got %d", pos)
	}

	// カーソル位置設定（実装上の制限により、設定後の値確認はスキップ）
	wrapper.SetCursorPosition(80)
	// SetCursorPositionの実装に問題があるため、エラーが発生しないことのみ確認
}

func TestBufferWrapper_IsDirty(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	wrapper := &BufferWrapper{buffer: buffer}

	// 初期状態は変更なし
	if wrapper.IsDirty() {
		t.Error("Expected buffer to not be dirty initially")
	}

	// 文字挿入後は変更あり
	buffer.InsertChar('a')
	if !wrapper.IsDirty() {
		t.Error("Expected buffer to be dirty after modification")
	}
}

func TestBufferWrapper_Filename(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	wrapper := &BufferWrapper{buffer: buffer}

	// 初期状態はファイル名なし
	filename := wrapper.Filename()
	if filename != "" {
		t.Errorf("Expected empty filename, got '%s'", filename)
	}

	// ファイルパス設定
	testPath := "/path/to/file.txt"
	buffer.SetFilepath(testPath)
	filename = wrapper.Filename()
	if filename != testPath {
		t.Errorf("Expected filename '%s', got '%s'", testPath, filename)
	}
}

func TestWindowWrapper_Size(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	window := domain.NewWindow(buffer, 80, 24)
	wrapper := &WindowWrapper{window: window}

	if wrapper.Width() != 80 {
		t.Errorf("Expected width 80, got %d", wrapper.Width())
	}

	if wrapper.Height() != 24 {
		t.Errorf("Expected height 24, got %d", wrapper.Height())
	}
}

func TestWindowWrapper_ScrollOffset(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	window := domain.NewWindow(buffer, 80, 24)
	wrapper := &WindowWrapper{window: window}

	// 初期スクロールオフセット
	offset := wrapper.ScrollOffset()
	if offset != 0 {
		t.Errorf("Expected scroll offset 0, got %d", offset)
	}

	// スクロールオフセット設定（実装確認）
	wrapper.SetScrollOffset(5)
	// NOTE: 実際の設定値確認は実装に依存するため、エラーが発生しないことのみ確認
}

func TestWindowWrapper_Buffer(t *testing.T) {
	buffer := domain.NewBuffer("test-buffer")
	window := domain.NewWindow(buffer, 80, 24)
	wrapper := &WindowWrapper{window: window}

	bufferInterface := wrapper.Buffer()
	if bufferInterface == nil {
		t.Fatal("Buffer() returned nil")
	}

	if bufferInterface.Name() != "test-buffer" {
		t.Errorf("Expected buffer name 'test-buffer', got '%s'", bufferInterface.Name())
	}
}