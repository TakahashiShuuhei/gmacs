package plugin

import (
	"context"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
)

// MockPluginWithModes はモードを持つテスト用のモックプラグイン
type MockPluginWithModes struct {
	name        string
	version     string
	description string
	majorModes  []MajorModeSpec
	minorModes  []MinorModeSpec
	commands    []CommandSpec
}

func (m *MockPluginWithModes) Name() string { return m.name }
func (m *MockPluginWithModes) Version() string { return m.version }
func (m *MockPluginWithModes) Description() string { return m.description }
func (m *MockPluginWithModes) Initialize(ctx context.Context, host HostInterface) error { return nil }
func (m *MockPluginWithModes) Cleanup() error { return nil }
func (m *MockPluginWithModes) GetCommands() []CommandSpec { return m.commands }
func (m *MockPluginWithModes) GetMajorModes() []MajorModeSpec { return m.majorModes }
func (m *MockPluginWithModes) GetMinorModes() []MinorModeSpec { return m.minorModes }
func (m *MockPluginWithModes) GetKeyBindings() []KeyBindingSpec { return nil }

func TestPluginMajorModeRegistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// 初期状態でのモード数を記録
	modeManager := editor.ModeManager()
	_, fundamentalExists := modeManager.GetMajorModeByName("fundamental-mode")
	if !fundamentalExists {
		t.Fatal("fundamental-mode should exist initially")
	}
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithModes{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with modes",
		majorModes: []MajorModeSpec{
			{
				Name:        "test-major-mode",
				Extensions:  []string{"test", "tst"},
				Description: "Test major mode",
				KeyBindings: []KeyBindingSpec{
					{
						Sequence: "C-t",
						Command:  "test-major-command",
						Mode:     "test-major-mode",
					},
				},
			},
		},
	}
	
	// プラグインアダプターを作成
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// モードを登録
	err := editor.RegisterPluginModes(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginModes() error = %v", err)
	}
	
	// モードが登録されたことを確認
	testMode, exists := modeManager.GetMajorModeByName("test-major-mode")
	if !exists {
		t.Error("test-major-mode not found in mode manager")
	}
	
	if testMode == nil {
		t.Fatal("test-major-mode is nil")
	}
	
	// モードの詳細を確認
	if testMode.Name() != "test-major-mode" {
		t.Errorf("Expected mode name 'test-major-mode', got '%s'", testMode.Name())
	}
	
	// ファイルパターンの確認
	pattern := testMode.FilePattern()
	if pattern == nil {
		t.Fatal("FilePattern should not be nil")
	}
	
	if !pattern.MatchString("test.test") {
		t.Error("Pattern should match 'test.test'")
	}
	
	if !pattern.MatchString("example.tst") {
		t.Error("Pattern should match 'example.tst'")
	}
}

func TestPluginMinorModeRegistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithModes{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with minor modes",
		minorModes: []MinorModeSpec{
			{
				Name:        "test-minor-mode",
				Description: "Test minor mode",
				Global:      false,
				KeyBindings: []KeyBindingSpec{
					{
						Sequence: "C-m",
						Command:  "test-minor-command",
						Mode:     "test-minor-mode",
					},
				},
			},
		},
	}
	
	// プラグインアダプターを作成
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// モードを登録
	err := editor.RegisterPluginModes(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginModes() error = %v", err)
	}
	
	// モードが登録されたことを確認
	modeManager := editor.ModeManager()
	testMode, exists := modeManager.GetMinorModeByName("test-minor-mode")
	if !exists {
		t.Error("test-minor-mode not found in mode manager")
	}
	
	if testMode == nil {
		t.Fatal("test-minor-mode is nil")
	}
	
	// モードの詳細を確認
	if testMode.Name() != "test-minor-mode" {
		t.Errorf("Expected mode name 'test-minor-mode', got '%s'", testMode.Name())
	}
}

func TestPluginModeUnregistration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithModes{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with modes",
		majorModes: []MajorModeSpec{
			{
				Name:        "test-major-mode",
				Extensions:  []string{"test"},
				Description: "Test major mode",
			},
		},
		minorModes: []MinorModeSpec{
			{
				Name:        "test-minor-mode",
				Description: "Test minor mode",
				Global:      false,
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// モードを登録
	err := editor.RegisterPluginModes(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginModes() error = %v", err)
	}
	
	// モードが登録されたことを確認
	modeManager := editor.ModeManager()
	_, majorExists := modeManager.GetMajorModeByName("test-major-mode")
	if !majorExists {
		t.Fatal("test-major-mode not found after registration")
	}
	
	_, minorExists := modeManager.GetMinorModeByName("test-minor-mode")
	if !minorExists {
		t.Fatal("test-minor-mode not found after registration")
	}
	
	// モードの登録を解除
	err = editor.UnregisterPluginModes(pluginAdapter)
	if err != nil {
		t.Fatalf("UnregisterPluginModes() error = %v", err)
	}
	
	// モードが削除されたことを確認
	_, majorExists = modeManager.GetMajorModeByName("test-major-mode")
	if majorExists {
		t.Error("test-major-mode should not exist after unregistration")
	}
	
	_, minorExists = modeManager.GetMinorModeByName("test-minor-mode")
	if minorExists {
		t.Error("test-minor-mode should not exist after unregistration")
	}
}

func TestPluginModeAutoDetection(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// テスト用プラグインを作成
	mockPlugin := &MockPluginWithModes{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin with auto-detection",
		majorModes: []MajorModeSpec{
			{
				Name:        "test-mode",
				Extensions:  []string{"tst"},
				Description: "Test mode for .tst files",
			},
		},
	}
	
	pluginAdapter := &PluginAdapter{plugin: mockPlugin}
	
	// モードを登録
	err := editor.RegisterPluginModes(pluginAdapter)
	if err != nil {
		t.Fatalf("RegisterPluginModes() error = %v", err)
	}
	
	// テストバッファを作成
	buffer := domain.NewBuffer("test.tst")
	buffer.SetFilepath("/path/to/test.tst")
	
	// 自動検出をテスト
	modeManager := editor.ModeManager()
	detectedMode, err := modeManager.AutoDetectMajorMode(buffer)
	if err != nil {
		t.Fatalf("AutoDetectMajorMode() error = %v", err)
	}
	
	if detectedMode.Name() != "test-mode" {
		t.Errorf("Expected detected mode 'test-mode', got '%s'", detectedMode.Name())
	}
}

func TestPluginModeIntegration(t *testing.T) {
	// エディタを作成
	editor := domain.NewEditor()
	
	// プラグインマネージャーアダプターを作成
	pm := NewPluginManager()
	adapter := NewPluginManagerAdapterWithRegistry(pm, editor)
	
	// エディタにプラグインマネージャーを設定
	editor.SetPluginManager(adapter)
	
	// 統合が正しく設定されたことを確認
	if editor.PluginManager() != adapter {
		t.Error("Plugin manager not set correctly")
	}
}