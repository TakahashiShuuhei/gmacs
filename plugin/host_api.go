package plugin

import (
	"fmt"
	"strings"
	"github.com/TakahashiShuuhei/gmacs/domain"
)

// HostAPI はプラグインがホスト（gmacs）にアクセスするためのAPI実装
type HostAPI struct {
	editor *domain.Editor
}

// NewHostAPI は新しいHostAPIを作成する
func NewHostAPI(editor *domain.Editor) *HostAPI {
	return &HostAPI{
		editor: editor,
	}
}

// エディタ操作

func (h *HostAPI) GetCurrentBuffer() BufferInterface {
	buffer := h.editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	return &BufferWrapper{buffer: buffer}
}

func (h *HostAPI) GetCurrentWindow() WindowInterface {
	window := h.editor.CurrentWindow()
	if window == nil {
		return nil
	}
	return &WindowWrapper{window: window}
}

func (h *HostAPI) SetStatus(message string) {
	h.editor.SetMinibufferMessage(message)
}

func (h *HostAPI) ShowMessage(message string) {
	h.editor.SetMinibufferMessage(message)
}

// コマンド実行

func (h *HostAPI) ExecuteCommand(name string, args ...interface{}) error {
	// TODO: コマンド実行の実装
	return fmt.Errorf("command execution not implemented yet")
}

// モード管理

func (h *HostAPI) SetMajorMode(bufferName, modeName string) error {
	buffer := h.editor.FindBuffer(bufferName)
	if buffer == nil {
		return fmt.Errorf("buffer %s not found", bufferName)
	}
	
	return h.editor.ModeManager().SetMajorMode(buffer, modeName)
}

func (h *HostAPI) ToggleMinorMode(bufferName, modeName string) error {
	buffer := h.editor.FindBuffer(bufferName)
	if buffer == nil {
		return fmt.Errorf("buffer %s not found", bufferName)
	}
	
	return h.editor.ModeManager().ToggleMinorMode(buffer, modeName)
}

// イベント・フック

func (h *HostAPI) AddHook(event string, handler func(...interface{}) error) {
	h.editor.AddHook(event, handler)
}

func (h *HostAPI) TriggerHook(event string, args ...interface{}) {
	h.editor.TriggerHook(event, args...)
}

// バッファ操作

func (h *HostAPI) CreateBuffer(name string) BufferInterface {
	buffer := domain.NewBuffer(name)
	h.editor.AddBuffer(buffer)
	return &BufferWrapper{buffer: buffer}
}

func (h *HostAPI) FindBuffer(name string) BufferInterface {
	buffer := h.editor.FindBuffer(name)
	if buffer == nil {
		return nil
	}
	return &BufferWrapper{buffer: buffer}
}

func (h *HostAPI) SwitchToBuffer(name string) error {
	buffer := h.editor.FindBuffer(name)
	if buffer == nil {
		return fmt.Errorf("buffer %s not found", name)
	}
	
	h.editor.SwitchToBuffer(buffer)
	return nil
}

// ファイル操作

func (h *HostAPI) OpenFile(path string) error {
	// TODO: ファイルオープンの実装
	return fmt.Errorf("file operations not implemented yet")
}

func (h *HostAPI) SaveBuffer(bufferName string) error {
	// TODO: バッファ保存の実装
	return fmt.Errorf("buffer save not implemented yet")
}

// 設定

func (h *HostAPI) GetOption(name string) (interface{}, error) {
	return h.editor.GetOption(name)
}

func (h *HostAPI) SetOption(name string, value interface{}) error {
	return h.editor.SetOption(name, value)
}

// BufferWrapper はdomain.BufferをBufferInterfaceでラップする
type BufferWrapper struct {
	buffer *domain.Buffer
}

func (b *BufferWrapper) Name() string {
	return b.buffer.Name()
}

func (b *BufferWrapper) Content() string {
	lines := b.buffer.Content()
	return strings.Join(lines, "\n")
}

func (b *BufferWrapper) SetContent(content string) {
	// TODO: 適切なSetContentの実装
	// 現在domain.Bufferには直接setするメソッドがないため、後で実装
}

func (b *BufferWrapper) InsertAt(pos int, text string) {
	// TODO: 適切な位置挿入の実装
	b.buffer.InsertChar(rune(text[0])) // 簡易実装
}

func (b *BufferWrapper) DeleteRange(start, end int) {
	// TODO: 範囲削除の実装
}

func (b *BufferWrapper) CursorPosition() int {
	cursor := b.buffer.Cursor()
	// 簡易的に行*80+列で位置を計算
	return cursor.Row*80 + cursor.Col
}

func (b *BufferWrapper) SetCursorPosition(pos int) {
	// 簡易的に位置から行・列を計算
	row := pos / 80
	col := pos % 80
	b.buffer.SetCursor(domain.Position{Row: row, Col: col})
}

func (b *BufferWrapper) MarkDirty() {
	// TODO: MarkDirtyの実装（domain.Bufferには対応メソッドがない）
}

func (b *BufferWrapper) IsDirty() bool {
	return b.buffer.IsModified()
}

func (b *BufferWrapper) Filename() string {
	return b.buffer.Filepath()
}

// WindowWrapper はdomain.WindowをWindowInterfaceでラップする
type WindowWrapper struct {
	window *domain.Window
}

func (w *WindowWrapper) Buffer() BufferInterface {
	return &BufferWrapper{buffer: w.window.Buffer()}
}

func (w *WindowWrapper) SetBuffer(buffer BufferInterface) {
	if bufferWrapper, ok := buffer.(*BufferWrapper); ok {
		w.window.SetBuffer(bufferWrapper.buffer)
	}
}

func (w *WindowWrapper) Width() int {
	width, _ := w.window.Size()
	return width
}

func (w *WindowWrapper) Height() int {
	_, height := w.window.Size()
	return height
}

func (w *WindowWrapper) ScrollOffset() int {
	return w.window.ScrollTop()
}

func (w *WindowWrapper) SetScrollOffset(offset int) {
	w.window.SetScrollTop(offset)
}