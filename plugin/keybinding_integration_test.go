package plugin

import (
	"context"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
)

// MockPluginWithKeyBindings はキーバインドを持つテスト用のモックプラグイン
type MockPluginWithKeyBindings struct {
	name        string
	version     string
	description string
	keyBindings []KeyBindingSpec
	commands    []CommandSpec
}

func (m *MockPluginWithKeyBindings) Name() string { return m.name }
func (m *MockPluginWithKeyBindings) Version() string { return m.version }
func (m *MockPluginWithKeyBindings) Description() string { return m.description }
func (m *MockPluginWithKeyBindings) Initialize(ctx context.Context, host HostInterface) error { return nil }
func (m *MockPluginWithKeyBindings) Cleanup() error { return nil }
func (m *MockPluginWithKeyBindings) GetCommands() []CommandSpec { return m.commands }
func (m *MockPluginWithKeyBindings) GetMajorModes() []MajorModeSpec { return nil }
func (m *MockPluginWithKeyBindings) GetMinorModes() []MinorModeSpec { return nil }
func (m *MockPluginWithKeyBindings) GetKeyBindings() []KeyBindingSpec { return m.keyBindings }

func TestPluginGlobalKeyBindingRegistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// 初期状態でのキーバインド数を確認
	sequences, rawSequences := editor.KeyBindings().ListBindings()
	initialSequenceCount := len(sequences)
	initialRawCount := len(rawSequences)
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithKeyBindings{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with key bindings",
		keyBindings: []KeyBindingSpec{
			{
				Sequence: "C-t",
				Command:  "test-global-command",
				Mode:     "global", // グローバルキーバインド
			},
			{
				Sequence: "M-p",
				Command:  "test-plugin-command",
				Mode:     "", // 空文字列はグローバルとして扱う
			},
			{
				Sequence: "C-x p",
				Command:  "test-multi-key-command",
				Mode:     "global",
			},
		},
		commands: []CommandSpec{
			{
				Name:        "test-global-command",
				Description: "Test global command",
				Interactive: true,
				Handler:     "test_global_handler",
			},
		},
	}
	
	// プラグインアダプターを作成
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// キーバインドを登録
	err := editor.RegisterPluginKeyBindings(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginKeyBindings() error = %v", err)
	}
	
	// キーバインドが登録されたことを確認
	sequences, rawSequences = editor.KeyBindings().ListBindings()
	expectedSequenceCount := initialSequenceCount + 3 // 3つのキーバインドを追加
	if len(sequences) != expectedSequenceCount {
		t.Errorf("Expected %d key sequences, got %d", expectedSequenceCount, len(sequences))
	}
	
	if len(rawSequences) != initialRawCount {
		t.Errorf("Raw sequences should not change, expected %d, got %d", initialRawCount, len(rawSequences))
	}
	
	// 個別のキーバインドが登録されているかチェック
	_, found1 := editor.KeyBindings().HasKeySequenceBinding("C-t")
	if !found1 {
		t.Error("C-t key binding not found")
	}
	
	_, found2 := editor.KeyBindings().HasKeySequenceBinding("M-p")
	if !found2 {
		t.Error("M-p key binding not found")
	}
	
	_, found3 := editor.KeyBindings().HasKeySequenceBinding("C-x p")
	if !found3 {
		t.Error("C-x p key binding not found")
	}
}

func TestPluginKeyBindingUnregistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithKeyBindings{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with key bindings",
		keyBindings: []KeyBindingSpec{
			{
				Sequence: "C-t",
				Command:  "test-command",
				Mode:     "global",
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// キーバインドを登録
	err := editor.RegisterPluginKeyBindings(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginKeyBindings() error = %v", err)
	}
	
	// キーバインドが登録されたことを確認
	_, found := editor.KeyBindings().HasKeySequenceBinding("C-t")
	if !found {
		t.Fatal("C-t key binding not found after registration")
	}
	
	// キーバインドの登録を解除
	err = editor.UnregisterPluginKeyBindings(pluginAdapter)
	if err != nil {
		t.Fatalf("UnregisterPluginKeyBindings() error = %v", err)
	}
	
	// キーバインドが削除されたことを確認
	_, found = editor.KeyBindings().HasKeySequenceBinding("C-t")
	if found {
		t.Error("C-t key binding should not exist after unregistration")
	}
}

func TestPluginModeSpecificKeyBindings(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// モード固有のキーバインドを持つプラグインを作成
	mockPlugin := &MockPluginWithKeyBindings{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with mode-specific key bindings",
		keyBindings: []KeyBindingSpec{
			{
				Sequence: "C-t",
				Command:  "test-global-command",
				Mode:     "global", // グローバル
			},
			{
				Sequence: "C-m",
				Command:  "test-mode-command",
				Mode:     "test-mode", // モード固有（グローバルには登録されない）
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// 初期状態でのキーバインド数を記録
	sequences, _ := editor.KeyBindings().ListBindings()
	initialCount := len(sequences)
	
	// キーバインドを登録
	err := editor.RegisterPluginKeyBindings(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginKeyBindings() error = %v", err)
	}
	
	// グローバルキーバインドのみが登録されることを確認
	sequences, _ = editor.KeyBindings().ListBindings()
	expectedCount := initialCount + 1 // C-t のみが追加される
	if len(sequences) != expectedCount {
		t.Errorf("Expected %d key sequences, got %d", expectedCount, len(sequences))
	}
	
	// グローバルキーバインドが登録されている
	_, found1 := editor.KeyBindings().HasKeySequenceBinding("C-t")
	if !found1 {
		t.Error("C-t (global) key binding not found")
	}
	
	// モード固有キーバインドはグローバルには登録されない
	_, found2 := editor.KeyBindings().HasKeySequenceBinding("C-m")
	if found2 {
		t.Error("C-m (mode-specific) key binding should not be in global bindings")
	}
}

func TestPluginKeyBindingExecution(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithKeyBindings{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with executable key bindings",
		keyBindings: []KeyBindingSpec{
			{
				Sequence: "C-t",
				Command:  "test-command",
				Mode:     "global",
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// キーバインドを登録
	err := editor.RegisterPluginKeyBindings(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginKeyBindings() error = %v", err)
	}
	
	// キーバインドを取得して実行
	cmd, found := editor.KeyBindings().HasKeySequenceBinding("C-t")
	if !found {
		t.Fatal("C-t key binding not found after registration")
	}
	
	// コマンドを実行（エラーが発生しないことを確認）
	err = cmd(editor)
	if err != nil {
		t.Errorf("Key binding execution failed: %v", err)
	}
	
	// ミニバッファにメッセージが設定されることを確認（実装に依存）
	// 現在の実装では、ミニバッファのメッセージを取得するAPIが限定的
}

func TestPluginKeyBindingIntegration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// プラグインマネージャーアダプターを作成
	pm := NewPluginManager()
	adapter := NewPluginManagerAdapterWithRegistry(pm, editor)
	
	// エディタにプラグインマネージャーを設定
	editor.SetPluginManager(adapter)
	
	// キーバインド統合が正しく設定されたことを確認
	if editor.PluginManager() != adapter {
		t.Error("Plugin manager not set correctly")
	}
}

func TestKeyBindingMapRemoveSequence(t *testing.T) {
	// 空のキーバインドマップを作成
	kbm := domain.NewEmptyKeyBindingMap()
	
	// テスト用コマンド
	testCmd := func(editor *domain.Editor) error {
		return nil
	}
	
	// キーバインドを追加
	kbm.BindKeySequence("C-t", testCmd)
	kbm.BindKeySequence("C-x C-f", testCmd)
	
	// 追加されたことを確認
	_, found1 := kbm.HasKeySequenceBinding("C-t")
	if !found1 {
		t.Error("C-t binding should exist")
	}
	
	_, found2 := kbm.HasKeySequenceBinding("C-x C-f")
	if !found2 {
		t.Error("C-x C-f binding should exist")
	}
	
	// キーバインドを削除
	removed1 := kbm.RemoveSequence("C-t")
	if !removed1 {
		t.Error("C-t binding should be removed")
	}
	
	// 削除されたことを確認
	_, found1 = kbm.HasKeySequenceBinding("C-t")
	if found1 {
		t.Error("C-t binding should not exist after removal")
	}
	
	// 他のキーバインドは残っていることを確認
	_, found2 = kbm.HasKeySequenceBinding("C-x C-f")
	if !found2 {
		t.Error("C-x C-f binding should still exist")
	}
	
	// 存在しないキーバインドの削除
	removed3 := kbm.RemoveSequence("C-nonexistent")
	if removed3 {
		t.Error("Non-existent binding removal should return false")
	}
}