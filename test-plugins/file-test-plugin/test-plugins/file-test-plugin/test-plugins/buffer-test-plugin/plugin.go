package main

import (
	"context"
	"encoding/gob"
	"fmt"

	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type StringError struct {
	Message string
}

func (se StringError) Error() string {
	return se.Message
}

func NewStringError(message string) error {
	return StringError{Message: message}
}

func init() {
	gob.Register(StringError{})
}

type BufferTestPlugin struct {
	host pluginsdk.HostInterface
}

func (p *BufferTestPlugin) Name() string {
	return "buffer-test-plugin"
}

func (p *BufferTestPlugin) Version() string {
	return "1.0.0"
}

func (p *BufferTestPlugin) Description() string {
	return "Plugin for testing buffer operations"
}

func (p *BufferTestPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
	p.host = host
	return nil
}

func (p *BufferTestPlugin) Cleanup() error {
	return nil
}

func (p *BufferTestPlugin) GetCommands() []pluginsdk.CommandSpec {
	return []pluginsdk.CommandSpec{
		{
			Name:        "buffer-test-create",
			Description: "Test buffer creation",
			Interactive: true,
			Handler:     "HandleBufferCreate",
		},
		{
			Name:        "buffer-test-content",
			Description: "Test buffer content operations",
			Interactive: true,
			Handler:     "HandleBufferContent",
		},
		{
			Name:        "buffer-test-cursor",
			Description: "Test buffer cursor operations",
			Interactive: true,
			Handler:     "HandleBufferCursor",
		},
		{
			Name:        "buffer-test-switch",
			Description: "Test buffer switching",
			Interactive: true,
			Handler:     "HandleBufferSwitch",
		},
	}
}

func (p *BufferTestPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
	return []pluginsdk.MajorModeSpec{}
}

func (p *BufferTestPlugin) GetMinorModes() []pluginsdk.MinorModeSpec {
	return []pluginsdk.MinorModeSpec{}
}

func (p *BufferTestPlugin) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	return []pluginsdk.KeyBindingSpec{}
}

// BufferInterface テスト
func (p *BufferTestPlugin) HandleBufferCreate() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	// 新しいバッファを作成
	buffer := p.host.CreateBuffer("*test-buffer*")
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: Failed to create buffer")
	}

	name := buffer.Name()
	return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Buffer created successfully: %s", name))
}

func (p *BufferTestPlugin) HandleBufferContent() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// 内容を設定
	testContent := "Test content from plugin\nLine 2\nLine 3"
	buffer.SetContent(testContent)
	buffer.MarkDirty()

	// 内容を取得して確認
	retrievedContent := buffer.Content()
	if retrievedContent == testContent {
		return NewStringError("PLUGIN_MESSAGE:Buffer content test PASSED")
	} else {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Buffer content test FAILED: expected '%s', got '%s'", testContent, retrievedContent))
	}
}

func (p *BufferTestPlugin) HandleBufferCursor() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// カーソル位置をテスト
	originalPos := buffer.CursorPosition()
	
	// カーソルを移動
	newPos := originalPos + 5
	buffer.SetCursorPosition(newPos)
	
	// 確認
	currentPos := buffer.CursorPosition()
	if currentPos == newPos {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Cursor test PASSED: moved from %d to %d", originalPos, currentPos))
	} else {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Cursor test FAILED: expected %d, got %d", newPos, currentPos))
	}
}

func (p *BufferTestPlugin) HandleBufferSwitch() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	// テストバッファを作成
	testBuffer := p.host.CreateBuffer("*buffer-switch-test*")
	if testBuffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: Failed to create test buffer")
	}

	// バッファに切り替え
	err := p.host.SwitchToBuffer("*buffer-switch-test*")
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Buffer switch test FAILED: %v", err))
	}

	// 現在のバッファを確認
	currentBuffer := p.host.GetCurrentBuffer()
	if currentBuffer == nil {
		return NewStringError("PLUGIN_MESSAGE:Buffer switch test FAILED: no current buffer after switch")
	}

	if currentBuffer.Name() == "*buffer-switch-test*" {
		return NewStringError("PLUGIN_MESSAGE:Buffer switch test PASSED")
	} else {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:Buffer switch test FAILED: expected '*buffer-switch-test*', got '%s'", currentBuffer.Name()))
	}
}

// CommandPlugin インターフェース実装
func (p *BufferTestPlugin) ExecuteCommand(name string, args ...interface{}) error {
	// 初期化確認
	if p.host == nil {
		hostImpl := &SimpleHostInterface{}
		p.Initialize(context.Background(), hostImpl)
	}

	switch name {
	case "buffer-test-create":
		return p.HandleBufferCreate()
	case "buffer-test-content":
		return p.HandleBufferContent()
	case "buffer-test-cursor":
		return p.HandleBufferCursor()
	case "buffer-test-switch":
		return p.HandleBufferSwitch()
	default:
		return fmt.Errorf("unknown command: %s", name)
	}
}

func (p *BufferTestPlugin) GetCompletions(command string, prefix string) []string {
	return []string{}
}

// SimpleHostInterface for testing
type SimpleHostInterface struct{}

func (h *SimpleHostInterface) GetCurrentBuffer() pluginsdk.BufferInterface { return nil }
func (h *SimpleHostInterface) GetCurrentWindow() pluginsdk.WindowInterface { return nil }
func (h *SimpleHostInterface) SetStatus(message string)                     {}
func (h *SimpleHostInterface) ShowMessage(message string)                   {}
func (h *SimpleHostInterface) ExecuteCommand(name string, args ...interface{}) error {
	return nil
}
func (h *SimpleHostInterface) SetMajorMode(bufferName, modeName string) error { return nil }
func (h *SimpleHostInterface) ToggleMinorMode(bufferName, modeName string) error {
	return nil
}
func (h *SimpleHostInterface) AddHook(event string, handler func(...interface{}) error) {}
func (h *SimpleHostInterface) TriggerHook(event string, args ...interface{})             {}
func (h *SimpleHostInterface) CreateBuffer(name string) pluginsdk.BufferInterface       { return nil }
func (h *SimpleHostInterface) FindBuffer(name string) pluginsdk.BufferInterface         { return nil }
func (h *SimpleHostInterface) SwitchToBuffer(name string) error                         { return nil }
func (h *SimpleHostInterface) OpenFile(path string) error                               { return nil }
func (h *SimpleHostInterface) SaveBuffer(bufferName string) error                       { return nil }
func (h *SimpleHostInterface) GetOption(name string) (interface{}, error)               { return nil, nil }
func (h *SimpleHostInterface) SetOption(name string, value interface{}) error           { return nil }

var pluginInstance = &BufferTestPlugin{}