package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

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

type FileTestPlugin struct {
	host pluginsdk.HostInterface
}

func (p *FileTestPlugin) Name() string {
	return "file-test-plugin"
}

func (p *FileTestPlugin) Version() string {
	return "1.0.0"
}

func (p *FileTestPlugin) Description() string {
	return "Plugin for testing file operations"
}

func (p *FileTestPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
	p.host = host
	return nil
}

func (p *FileTestPlugin) Cleanup() error {
	return nil
}

func (p *FileTestPlugin) GetCommands() []pluginsdk.CommandSpec {
	return []pluginsdk.CommandSpec{
		{
			Name:        "file-test-create",
			Description: "Test file creation and opening",
			Interactive: true,
			Handler:     "HandleFileCreate",
		},
		{
			Name:        "file-test-save",
			Description: "Test file saving",
			Interactive: true,
			Handler:     "HandleFileSave",
		},
		{
			Name:        "file-test-content",
			Description: "Test file content operations",
			Interactive: true,
			Handler:     "HandleFileContent",
		},
	}
}

func (p *FileTestPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
	return []pluginsdk.MajorModeSpec{}
}

func (p *FileTestPlugin) GetMinorModes() []pluginsdk.MinorModeSpec {
	return []pluginsdk.MinorModeSpec{}
}

func (p *FileTestPlugin) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	return []pluginsdk.KeyBindingSpec{}
}

func (p *FileTestPlugin) HandleFileCreate() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	// テスト用ファイルを作成
	testFile := filepath.Join(os.TempDir(), "gmacs-plugin-test.txt")
	testContent := "Test file created by plugin\nLine 2\nLine 3"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File creation FAILED: %v", err))
	}

	// ファイルを開く
	err = p.host.OpenFile(testFile)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File open FAILED: %v", err))
	}

	return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File create/open test PASSED: %s", testFile))
}

func (p *FileTestPlugin) HandleFileSave() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// バッファの内容を変更
	originalContent := buffer.Content()
	modifiedContent := originalContent + "\nModified by plugin"
	buffer.SetContent(modifiedContent)
	buffer.MarkDirty()

	// バッファを保存
	bufferName := buffer.Name()
	err := p.host.SaveBuffer(bufferName)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File save FAILED: %v", err))
	}

	// ダーティフラグをチェック
	if buffer.IsDirty() {
		return NewStringError("PLUGIN_MESSAGE:File save test PARTIAL: file saved but dirty flag still set")
	}

	return NewStringError("PLUGIN_MESSAGE:File save test PASSED")
}

func (p *FileTestPlugin) HandleFileContent() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// ファイル情報をテスト
	name := buffer.Name()
	filename := buffer.Filename()
	content := buffer.Content()
	isDirty := buffer.IsDirty()

	if filename == "" {
		return NewStringError("PLUGIN_MESSAGE:File content test PARTIAL: buffer has no associated file")
	}

	// ファイルが実際に存在するかチェック
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File content test FAILED: file does not exist: %s", filename))
	}

	result := fmt.Sprintf("File: %s, Buffer: %s, Content length: %d, Dirty: %t", 
		filename, name, len(content), isDirty)
	return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File content test PASSED - %s", result))
}

// CommandPlugin インターフェース実装
func (p *FileTestPlugin) ExecuteCommand(name string, args ...interface{}) error {
	// 初期化確認
	if p.host == nil {
		hostImpl := &SimpleHostInterface{}
		p.Initialize(context.Background(), hostImpl)
	}

	switch name {
	case "file-test-create":
		return p.HandleFileCreate()
	case "file-test-save":
		return p.HandleFileSave()
	case "file-test-content":
		return p.HandleFileContent()
	default:
		return fmt.Errorf("unknown command: %s", name)
	}
}

func (p *FileTestPlugin) GetCompletions(command string, prefix string) []string {
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

var pluginInstance = &FileTestPlugin{}