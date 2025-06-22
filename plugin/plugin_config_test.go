package plugin

import (
	"os"
	"path/filepath"
	"testing"

	luaconfig "github.com/TakahashiShuuhei/gmacs/lua-config"
)

func TestPluginConfigFileDiscovery(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()
	
	// XDG_CONFIG_HOME環境変数を設定
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
	}()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// テスト用のプラグイン設定ディレクトリを作成
	gmacsConfigDir := filepath.Join(tempDir, "gmacs")
	err := os.MkdirAll(gmacsConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// テスト用のplugins.luaファイルを作成
	pluginConfigPath := filepath.Join(gmacsConfigDir, "plugins.lua")
	pluginConfigContent := `
-- Test plugin configuration
gmacs.defun("test-plugin-command", function()
    gmacs.message("Test plugin command executed")
end)

gmacs.bind_key("C-c t", "test-plugin-command")
`
	err = os.WriteFile(pluginConfigPath, []byte(pluginConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create plugin config file: %v", err)
	}
	
	// ConfigLoaderでプラグイン設定ファイルを検索
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	foundPath, err := configLoader.FindPluginConfigFile()
	if err != nil {
		t.Fatalf("FindPluginConfigFile() error = %v", err)
	}
	
	if foundPath != pluginConfigPath {
		t.Errorf("Expected plugin config path %s, got %s", pluginConfigPath, foundPath)
	}
}

func TestPluginConfigLoading(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()
	
	// テスト用のplugins.luaファイルを作成
	pluginConfigPath := filepath.Join(tempDir, "plugins.lua")
	pluginConfigContent := `
-- Test plugin configuration
gmacs.defun("test-config-command", function()
    gmacs.message("Configuration loaded successfully")
end)
`
	err := os.WriteFile(pluginConfigPath, []byte(pluginConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create plugin config file: %v", err)
	}
	
	// エディタを作成（プラグインマネージャー付き）
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	editor := CreateEditorWithPlugins(configLoader, nil)
	
	// APIバインディングを設定
	vm := configLoader.GetVM()
	apiBindings := luaconfig.NewAPIBindings(editor, vm)
	err = apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register Lua API: %v", err)
	}
	
	// プラグイン設定ファイルをロード
	err = configLoader.LoadConfig(pluginConfigPath)
	if err != nil {
		t.Fatalf("Failed to load plugin config: %v", err)
	}
	
	// 設定で定義されたコマンドが登録されているかチェック
	_, exists := editor.CommandRegistry().Get("test-config-command")
	if !exists {
		t.Error("test-config-command should be registered")
	}
}

func TestLoadPluginConfigIfExists(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir := t.TempDir()
	
	// XDG_CONFIG_HOME環境変数を設定
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
	}()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// テスト用のプラグイン設定ディレクトリとファイルを作成
	gmacsConfigDir := filepath.Join(tempDir, "gmacs")
	err := os.MkdirAll(gmacsConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	pluginConfigPath := filepath.Join(gmacsConfigDir, "plugins.lua")
	pluginConfigContent := `
-- Test auto-load configuration
gmacs.message("Auto-loaded plugin configuration")
`
	err = os.WriteFile(pluginConfigPath, []byte(pluginConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create plugin config file: %v", err)
	}
	
	// ConfigLoaderを作成
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	// エディタとAPIバインディングを設定
	editor := CreateEditorWithPlugins(configLoader, nil)
	vm := configLoader.GetVM()
	apiBindings := luaconfig.NewAPIBindings(editor, vm)
	err = apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register Lua API: %v", err)
	}
	
	// LoadPluginConfigIfExistsを呼び出し
	err = LoadPluginConfigIfExists(configLoader)
	if err != nil {
		t.Fatalf("LoadPluginConfigIfExists() error = %v", err)
	}
}

func TestPluginConfigNotFound(t *testing.T) {
	// 存在しないディレクトリを設定
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
	}()
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent")
	
	// HOME環境変数も存在しないパスに設定
	originalHOME := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", originalHOME)
	}()
	os.Setenv("HOME", "/nonexistent")
	
	// ConfigLoaderを作成
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	// プラグイン設定ファイルが見つからない場合
	foundPath, err := configLoader.FindPluginConfigFile()
	if err != nil {
		t.Fatalf("FindPluginConfigFile() should not error when no file found: %v", err)
	}
	
	if foundPath != "" {
		t.Errorf("Expected empty path when no plugin config found, got %s", foundPath)
	}
	
	// LoadPluginConfigは空の場合でもエラーにならない
	err = configLoader.LoadPluginConfig()
	if err != nil {
		t.Fatalf("LoadPluginConfig() should not error when no file found: %v", err)
	}
}

func TestPluginLuaAPIIntegration(t *testing.T) {
	// エディタを作成（プラグインマネージャー付き）
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	editor := CreateEditorWithPlugins(configLoader, nil)
	
	// APIバインディングを設定
	vm := configLoader.GetVM()
	apiBindings := luaconfig.NewAPIBindings(editor, vm)
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register Lua API: %v", err)
	}
	
	// 一時ファイルでプラグインLua APIをテスト
	tempDir := t.TempDir()
	testConfigPath := filepath.Join(tempDir, "test.lua")
	testConfig := `
-- Test plugin Lua API
local plugins = gmacs.list_plugins()
local loaded = gmacs.plugin_loaded("nonexistent-plugin")

-- These should not error
if not loaded then
    gmacs.message("Plugin not loaded (expected)")
end
`
	err = os.WriteFile(testConfigPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// テスト設定をロード
	err = configLoader.LoadConfig(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
}