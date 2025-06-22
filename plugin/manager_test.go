package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockPlugin はテスト用のモックプラグイン
type MockPlugin struct {
	name        string
	version     string
	description string
	commands    []CommandSpec
	majorModes  []MajorModeSpec
	minorModes  []MinorModeSpec
	keyBindings []KeyBindingSpec
	initError   error
	cleanupError error
}

func (m *MockPlugin) Name() string { return m.name }
func (m *MockPlugin) Version() string { return m.version }
func (m *MockPlugin) Description() string { return m.description }
func (m *MockPlugin) Initialize(ctx context.Context, host HostInterface) error { return m.initError }
func (m *MockPlugin) Cleanup() error { return m.cleanupError }
func (m *MockPlugin) GetCommands() []CommandSpec { return m.commands }
func (m *MockPlugin) GetMajorModes() []MajorModeSpec { return m.majorModes }
func (m *MockPlugin) GetMinorModes() []MinorModeSpec { return m.minorModes }
func (m *MockPlugin) GetKeyBindings() []KeyBindingSpec { return m.keyBindings }

func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()

	if pm == nil {
		t.Fatal("NewPluginManager() returned nil")
	}

	if pm.plugins == nil {
		t.Error("plugins map not initialized")
	}

	if pm.clients == nil {
		t.Error("clients map not initialized")
	}

	if len(pm.searchPaths) == 0 {
		t.Error("searchPaths not initialized")
	}

	// デフォルトの検索パスが設定されているかチェック
	expectedPaths := GetDefaultPluginPaths()
	if len(pm.searchPaths) != len(expectedPaths) {
		t.Errorf("Expected %d search paths, got %d", len(expectedPaths), len(pm.searchPaths))
	}
}

func TestPluginManager_ListPlugins_Empty(t *testing.T) {
	pm := NewPluginManager()
	plugins := pm.ListPlugins()

	if len(plugins) != 0 {
		t.Errorf("Expected empty plugin list, got %d plugins", len(plugins))
	}
}

func TestPluginManager_GetPlugin_NotFound(t *testing.T) {
	pm := NewPluginManager()
	plugin, found := pm.GetPlugin("nonexistent")

	if found {
		t.Error("Expected plugin not found, but found = true")
	}

	if plugin != nil {
		t.Error("Expected nil plugin for nonexistent plugin")
	}
}

func TestPluginManager_UnloadPlugin_NotLoaded(t *testing.T) {
	pm := NewPluginManager()
	err := pm.UnloadPlugin("nonexistent")

	if err == nil {
		t.Error("Expected error when unloading nonexistent plugin")
	}

	expectedError := "plugin nonexistent is not loaded"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPluginManager_findPlugin(t *testing.T) {
	// 一時ディレクトリでテスト
	tempDir := t.TempDir()

	// テスト用のプラグインディレクトリとバイナリを作成
	pluginDir := filepath.Join(tempDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	binaryPath := filepath.Join(pluginDir, "test-plugin")
	err = os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin binary: %v", err)
	}

	// PluginManagerのsearchPathsを設定
	pm := NewPluginManager()
	pm.searchPaths = []string{tempDir}

	// プラグインを検索
	foundPath, err := pm.findPlugin("test-plugin")
	if err != nil {
		t.Fatalf("findPlugin() error = %v", err)
	}

	if foundPath != binaryPath {
		t.Errorf("findPlugin() = %v, want %v", foundPath, binaryPath)
	}
}

func TestPluginManager_findPlugin_NotFound(t *testing.T) {
	pm := NewPluginManager()
	pm.searchPaths = []string{"/nonexistent"}

	_, err := pm.findPlugin("nonexistent")
	if err == nil {
		t.Error("Expected error when plugin not found")
	}

	expectedError := "plugin binary not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPluginManager_loadManifest(t *testing.T) {
	// 一時ディレクトリでテスト
	tempDir := t.TempDir()
	pluginDir := filepath.Join(tempDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	pm := NewPluginManager()
	manifest, err := pm.loadManifest(pluginDir)
	if err != nil {
		t.Fatalf("loadManifest() error = %v", err)
	}

	if manifest == nil {
		t.Fatal("loadManifest() returned nil manifest")
	}

	// 最小限のマニフェストが返されることを確認
	if manifest.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", manifest.Name)
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", manifest.Version)
	}
}

func TestPluginManager_Shutdown(t *testing.T) {
	pm := NewPluginManager()

	// 空の状態でShutdownを呼び出し
	err := pm.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// プラグインマップが空であることを確認
	if len(pm.plugins) != 0 {
		t.Errorf("Expected empty plugins map after shutdown, got %d", len(pm.plugins))
	}

	if len(pm.clients) != 0 {
		t.Errorf("Expected empty clients map after shutdown, got %d", len(pm.clients))
	}
}

// 統合テスト用のセットアップヘルパー
func setupTestPluginEnvironment(t *testing.T) (string, *PluginManager) {
	tempDir := t.TempDir()
	
	// テスト用プラグインディレクトリ作成
	pluginDir := filepath.Join(tempDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// manifest.json作成
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	manifestContent := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "Test plugin",
		"author": "Test Author",
		"binary": "test-plugin"
	}`
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// ダミーバイナリ作成
	binaryPath := filepath.Join(pluginDir, "test-plugin")
	err = os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create plugin binary: %v", err)
	}

	pm := NewPluginManager()
	pm.searchPaths = []string{tempDir}

	return tempDir, pm
}

func TestLoadedPlugin_Structure(t *testing.T) {
	now := time.Now()
	
	loadedPlugin := &LoadedPlugin{
		Name:     "test",
		Version:  "1.0.0",
		Path:     "/path/to/plugin",
		Plugin:   &MockPlugin{name: "test"},
		Config:   make(map[string]interface{}),
		State:    PluginStateLoaded,
		Manifest: &PluginManifest{Name: "test"},
		LoadTime: now,
	}

	if loadedPlugin.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", loadedPlugin.Name)
	}

	if loadedPlugin.State != PluginStateLoaded {
		t.Errorf("Expected state PluginStateLoaded, got %v", loadedPlugin.State)
	}

	if loadedPlugin.LoadTime != now {
		t.Errorf("Expected LoadTime %v, got %v", now, loadedPlugin.LoadTime)
	}
}

func TestPluginInfo_Structure(t *testing.T) {
	info := PluginInfo{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test description",
		State:       PluginStateLoaded,
		Enabled:     true,
	}

	if info.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", info.Name)
	}

	if !info.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if info.State != PluginStateLoaded {
		t.Errorf("Expected state PluginStateLoaded, got %v", info.State)
	}
}