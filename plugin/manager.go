package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
)

// PluginManager はプラグインのライフサイクルを管理する
type PluginManager struct {
	plugins     map[string]*LoadedPlugin
	clients     map[string]*plugin.Client
	searchPaths []string
	mutex       sync.RWMutex
}

// NewPluginManager は新しいPluginManagerを作成する
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		clients:     make(map[string]*plugin.Client),
		searchPaths: GetDefaultPluginPaths(),
	}
}

// NewPluginManagerWithPaths は指定されたパスでPluginManagerを作成する
func NewPluginManagerWithPaths(searchPaths []string) *PluginManager {
	pm := &PluginManager{
		plugins:     make(map[string]*LoadedPlugin),
		clients:     make(map[string]*plugin.Client),
		searchPaths: searchPaths,
	}
	
	// 起動時にインストール済みプラグインを自動発見（同期実行に変更）
	pm.autoDiscoverPlugins()
	
	return pm
}

// LoadPlugin はプラグインをロードする
func (pm *PluginManager) LoadPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 既にロード済みかチェック
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s is already loaded", name)
	}

	// プラグインバイナリを検索
	pluginPath, err := pm.findPlugin(name)
	if err != nil {
		return fmt.Errorf("plugin %s not found: %v", name, err)
	}

	// マニフェストを読み込み
	manifest, err := pm.loadManifest(filepath.Dir(pluginPath))
	if err != nil {
		return fmt.Errorf("failed to load manifest for %s: %v", name, err)
	}

	// プラグインクライアントを作成
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
	})

	// プラグインに接続
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect to plugin %s: %v", name, err)
	}

	// プラグインインスタンスを取得
	raw, err := rpcClient.Dispense("gmacs-plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin %s: %v", name, err)
	}

	pluginInstance, ok := raw.(Plugin)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin %s does not implement Plugin interface", name)
	}

	// LoadedPluginを作成
	loadedPlugin := &LoadedPlugin{
		Name:     name,
		Version:  manifest.Version,
		Path:     pluginPath,
		Plugin:   pluginInstance,
		Config:   make(map[string]interface{}),
		State:    PluginStateLoaded,
		Manifest: manifest,
		LoadTime: time.Now(),
	}

	// 登録
	pm.plugins[name] = loadedPlugin
	pm.clients[name] = client

	return nil
}

// UnloadPlugin はプラグインをアンロードする
func (pm *PluginManager) UnloadPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	client, exists := pm.clients[name]
	if !exists {
		return fmt.Errorf("plugin client %s not found", name)
	}

	// プラグインのクリーンアップ
	if err := loadedPlugin.Plugin.Cleanup(); err != nil {
		// ログ出力などを行うが、エラーは無視
		fmt.Printf("Warning: plugin %s cleanup failed: %v\n", name, err)
	}

	// クライアントを終了
	client.Kill()

	// 登録解除
	delete(pm.plugins, name)
	delete(pm.clients, name)

	return nil
}

// SetPluginConfig sets the configuration for a plugin
func (pm *PluginManager) SetPluginConfig(name string, config map[string]interface{}) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	// Update plugin configuration
	for key, value := range config {
		loadedPlugin.Config[key] = value
	}

	return nil
}

// GetPluginConfig gets the configuration for a plugin
func (pm *PluginManager) GetPluginConfig(name string) (map[string]interface{}, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s is not loaded", name)
	}

	// Return a copy of the configuration to prevent external modification
	config := make(map[string]interface{})
	for key, value := range loadedPlugin.Config {
		config[key] = value
	}

	return config, nil
}

// SetPluginConfigValue sets a single configuration value for a plugin
func (pm *PluginManager) SetPluginConfigValue(name string, key string, value interface{}) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	loadedPlugin.Config[key] = value
	return nil
}

// GetPlugin はロード済みプラグインを取得する
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return nil, false
	}

	return loadedPlugin.Plugin, true
}

// ListPlugins はロード済みプラグインの一覧を返す
func (pm *PluginManager) ListPlugins() []PluginInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var plugins []PluginInfo
	for _, loadedPlugin := range pm.plugins {
		plugins = append(plugins, PluginInfo{
			Name:        loadedPlugin.Name,
			Version:     loadedPlugin.Version,
			Description: loadedPlugin.Manifest.Description,
			State:       loadedPlugin.State,
			Enabled:     loadedPlugin.State == PluginStateLoaded,
		})
	}

	return plugins
}

// ListInstalledPlugins はディスクにインストールされているプラグインの一覧を返す（ロード状態に関係なく）
func (pm *PluginManager) ListInstalledPlugins() []PluginInfo {
	var plugins []PluginInfo
	seen := make(map[string]bool) // 重複を避けるためのマップ
	
	// 各検索パスをスキャン
	for _, searchPath := range pm.searchPaths {
		if entries, err := os.ReadDir(searchPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					pluginDir := filepath.Join(searchPath, entry.Name())
					if IsPluginDir(pluginDir) {
						// マニフェストを読み込んでプラグイン情報を取得
						if manifest, err := pm.loadSimpleManifest(pluginDir); err == nil {
							// プラグイン名で重複チェック
							if seen[manifest.Name] {
								continue // 既に追加済みの場合はスキップ
							}
							seen[manifest.Name] = true
							
							// ロード状態をチェック
							pm.mutex.RLock()
							loadedPlugin, isLoaded := pm.plugins[entry.Name()]
							pm.mutex.RUnlock()
							
							state := PluginStateUnloaded
							enabled := false
							if isLoaded {
								state = loadedPlugin.State
								enabled = state == PluginStateLoaded
							}
							
							plugins = append(plugins, PluginInfo{
								Name:        manifest.Name,
								Version:     manifest.Version,
								Description: manifest.Description,
								State:       state,
								Enabled:     enabled,
							})
						}
					}
				}
			}
		}
	}
	
	return plugins
}

// loadSimpleManifest は軽量なマニフェスト読み込み（JSON解析版）
func (pm *PluginManager) loadSimpleManifest(pluginDir string) (*PluginManifest, error) {
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	
	// manifest.jsonの存在チェック
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// フォールバック: ディレクトリ名ベースの最小限マニフェスト
		pluginName := filepath.Base(pluginDir)
		if strings.HasSuffix(pluginName, ".git") {
			pluginName = strings.TrimSuffix(pluginName, ".git")
		}
		
		return &PluginManifest{
			Name:        pluginName,
			Version:     "1.0.0",
			Description: "Plugin installed from source",
			Author:      "Unknown",
			Binary:      pluginName,
		}, nil
	}
	
	// manifest.jsonファイルを読み込み
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %v", err)
	}
	
	// JSONをパース
	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %v", err)
	}
	
	// 必須フィールドのデフォルト値設定
	if manifest.Name == "" {
		manifest.Name = filepath.Base(pluginDir)
	}
	if manifest.Version == "" {
		manifest.Version = "1.0.0"
	}
	if manifest.Description == "" {
		manifest.Description = "Plugin installed from source"
	}
	if manifest.Author == "" {
		manifest.Author = "Unknown"
	}
	if manifest.Binary == "" {
		manifest.Binary = manifest.Name
	}
	
	return &manifest, nil
}

// autoDiscoverPlugins はインストール済みプラグインを自動発見してロードする
func (pm *PluginManager) autoDiscoverPlugins() {
	// 少し待ってからプラグインを発見（初期化完了を待つ）
	time.Sleep(100 * time.Millisecond)
	
	// 重複を避けるため、既に発見したプラグイン名を記録
	discovered := make(map[string]bool)
	
	// 各検索パスをスキャン
	for _, searchPath := range pm.searchPaths {
		if entries, err := os.ReadDir(searchPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					pluginDir := filepath.Join(searchPath, entry.Name())
					if IsPluginDir(pluginDir) {
						// マニフェストを読み込んでプラグイン名を取得
						if manifest, err := pm.loadSimpleManifest(pluginDir); err == nil {
							dirName := entry.Name()  // 実際のディレクトリ名
							pluginName := manifest.Name  // manifest.json内の名前
							
							// 既に発見済みの場合はスキップ
							if discovered[pluginName] {
								continue
							}
							discovered[pluginName] = true
							
							// ディレクトリ名でプラグインをロード（実際のバイナリがある場所）
							pm.LoadPlugin(dirName)
						}
					}
				}
			}
		}
	}
}

// Shutdown は全プラグインをアンロードする
func (pm *PluginManager) Shutdown() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var errors []error

	for name := range pm.plugins {
		if err := pm.unloadPluginUnsafe(name); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// findPlugin はプラグインバイナリを検索する
func (pm *PluginManager) findPlugin(name string) (string, error) {
	for _, searchPath := range pm.searchPaths {
		pluginDir := filepath.Join(searchPath, name)
		
		// ディレクトリが存在するかチェック
		if _, err := os.Stat(pluginDir); err != nil {
			continue
		}

		// バイナリファイルを検索
		binaryPath := filepath.Join(pluginDir, name)
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}

		// 拡張子付きでも検索
		binaryPath = filepath.Join(pluginDir, name+".exe")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("plugin binary not found")
}

// loadManifest はマニフェストファイルを読み込む
func (pm *PluginManager) loadManifest(pluginDir string) (*PluginManifest, error) {
	_ = filepath.Join(pluginDir, "manifest.json") // manifestPath - TODO: JSON読み込み実装
	
	// TODO: JSON読み込み実装
	// 現在は最小限のマニフェストを返す
	return &PluginManifest{
		Name:        filepath.Base(pluginDir),
		Version:     "1.0.0",
		Description: "Plugin",
		Author:      "Unknown",
		Binary:      filepath.Base(pluginDir),
	}, nil
}

// unloadPluginUnsafe は内部用のアンロード関数（ロック不要）
func (pm *PluginManager) unloadPluginUnsafe(name string) error {
	loadedPlugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	client, exists := pm.clients[name]
	if !exists {
		return fmt.Errorf("plugin client %s not found", name)
	}

	// プラグインのクリーンアップ
	if err := loadedPlugin.Plugin.Cleanup(); err != nil {
		fmt.Printf("Warning: plugin %s cleanup failed: %v\n", name, err)
	}

	// クライアントを終了
	client.Kill()

	// 登録解除
	delete(pm.plugins, name)
	delete(pm.clients, name)

	return nil
}