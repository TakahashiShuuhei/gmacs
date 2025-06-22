package plugin

import (
	"context"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
)

// MockPluginWithCommands はコマンドを持つテスト用のモックプラグイン
type MockPluginWithCommands struct {
	name        string
	version     string
	description string
	commands    []CommandSpec
}

func (m *MockPluginWithCommands) Name() string { return m.name }
func (m *MockPluginWithCommands) Version() string { return m.version }
func (m *MockPluginWithCommands) Description() string { return m.description }
func (m *MockPluginWithCommands) Initialize(ctx context.Context, host HostInterface) error { return nil }
func (m *MockPluginWithCommands) Cleanup() error { return nil }
func (m *MockPluginWithCommands) GetCommands() []CommandSpec { return m.commands }
func (m *MockPluginWithCommands) GetMajorModes() []MajorModeSpec { return nil }
func (m *MockPluginWithCommands) GetMinorModes() []MinorModeSpec { return nil }
func (m *MockPluginWithCommands) GetKeyBindings() []KeyBindingSpec { return nil }

func TestPluginCommandRegistration(t *testing.T) {
	// エディタとプラグインマネージャーを作成
	editor := domain.NewEditor()
	
	// 初期状態でのコマンド数を記録
	initialCommandCount := len(editor.CommandRegistry().List())
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithCommands{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with commands",
		commands: []CommandSpec{
			{
				Name:        "test-command-1",
				Description: "First test command",
				Interactive: true,
				Handler:     "test_command_1_handler",
			},
			{
				Name:        "test-command-2",
				Description: "Second test command",
				Interactive: false,
				Handler:     "test_command_2_handler",
			},
		},
	}
	
	// プラグインアダプターを作成
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// コマンドを登録
	err := editor.RegisterPluginCommands(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginCommands() error = %v", err)
	}
	
	// コマンドが登録されたことを確認
	commands := editor.CommandRegistry().List()
	expectedCommandCount := initialCommandCount + 2
	if len(commands) != expectedCommandCount {
		t.Errorf("Expected %d commands, got %d", expectedCommandCount, len(commands))
	}
	
	// 個別のコマンドが登録されているかチェック
	_, exists1 := editor.CommandRegistry().Get("test-command-1")
	if !exists1 {
		t.Error("test-command-1 not found in command registry")
	}
	
	_, exists2 := editor.CommandRegistry().Get("test-command-2")
	if !exists2 {
		t.Error("test-command-2 not found in command registry")
	}
}

func TestPluginCommandUnregistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithCommands{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with commands",
		commands: []CommandSpec{
			{
				Name:        "test-command-1",
				Description: "First test command",
				Interactive: true,
				Handler:     "test_command_1_handler",
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// コマンドを登録
	err := editor.RegisterPluginCommands(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginCommands() error = %v", err)
	}
	
	// コマンドが登録されたことを確認
	_, exists := editor.CommandRegistry().Get("test-command-1")
	if !exists {
		t.Fatal("test-command-1 not found after registration")
	}
	
	// コマンドの登録を解除
	err = editor.UnregisterPluginCommands(pluginAdapter)
	if err != nil {
		t.Fatalf("UnregisterPluginCommands() error = %v", err)
	}
	
	// コマンドが削除されたことを確認
	_, exists = editor.CommandRegistry().Get("test-command-1")
	if exists {
		t.Error("test-command-1 should not exist after unregistration")
	}
}

func TestPluginManagerCommandIntegration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// プラグインマネージャーアダプターを作成
	pm := NewPluginManager()
	adapter := NewPluginManagerAdapterWithRegistry(pm, editor)
	
	// エディタにプラグインマネージャーを設定
	editor.SetPluginManager(adapter)
	
	// プラグインマネージャーが正しく設定されたことを確認
	if editor.PluginManager() != adapter {
		t.Error("Plugin manager not set correctly")
	}
	
	// プラグインマネージャーのリストが空であることを確認
	plugins := adapter.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected empty plugin list, got %d plugins", len(plugins))
	}
}

func TestPluginCommandExecution(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithCommands{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with commands",
		commands: []CommandSpec{
			{
				Name:        "test-command",
				Description: "Test command",
				Interactive: true,
				Handler:     "test_command_handler",
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// コマンドを登録
	err := editor.RegisterPluginCommands(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginCommands() error = %v", err)
	}
	
	// コマンドを取得して実行
	cmd, exists := editor.CommandRegistry().Get("test-command")
	if !exists {
		t.Fatal("test-command not found after registration")
	}
	
	// コマンドを実行（エラーが発生しないことを確認）
	err = cmd.Execute(editor)
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}
	
	// ミニバッファにメッセージが設定されることを確認（現在の実装では確認困難）
	// 実際の実装では、ミニバッファのメッセージを取得するAPIが必要
}