package plugin

import (
	"fmt"
	"github.com/TakahashiShuuhei/gmacs/domain"
	gmacslog "github.com/TakahashiShuuhei/gmacs/log"
)

// EditorCommandRegistry defines the interface for registering/unregistering plugin commands, modes, and key bindings
type EditorCommandRegistry interface {
	RegisterPluginCommands(plugin domain.PluginInterface) error
	UnregisterPluginCommands(plugin domain.PluginInterface) error
	RegisterPluginModes(plugin domain.PluginInterface) error
	UnregisterPluginModes(plugin domain.PluginInterface) error
	RegisterPluginKeyBindings(plugin domain.PluginInterface) error
	UnregisterPluginKeyBindings(plugin domain.PluginInterface) error
}

// PluginManagerAdapter はplugin.PluginManagerをdomain.PluginManagerInterfaceに適合させる
type PluginManagerAdapter struct {
	pm       *PluginManager
	registry EditorCommandRegistry
}

// NewPluginManagerAdapter は新しいアダプターを作成する
func NewPluginManagerAdapter(pm *PluginManager) *PluginManagerAdapter {
	return &PluginManagerAdapter{pm: pm}
}

// NewPluginManagerAdapterWithRegistry は新しいアダプターをコマンドレジストリ付きで作成する
func NewPluginManagerAdapterWithRegistry(pm *PluginManager, registry EditorCommandRegistry) *PluginManagerAdapter {
	return &PluginManagerAdapter{
		pm:       pm,
		registry: registry,
	}
}

// LoadPlugin implements domain.PluginManagerInterface
func (a *PluginManagerAdapter) LoadPlugin(name string) error {
	err := a.pm.LoadPlugin(name)
	if err != nil {
		return err
	}
	
	// プラグインが正常にロードされた場合、コマンドとモードを登録
	if a.registry != nil {
		plugin, found := a.pm.GetPlugin(name)
		if found {
			pluginAdapter := &PluginAdapter{plugin: plugin}
			// コマンドを登録
			if err := a.registry.RegisterPluginCommands(pluginAdapter); err != nil {
				return err
			}
			// モードを登録
			if err := a.registry.RegisterPluginModes(pluginAdapter); err != nil {
				return err
			}
			// キーバインドを登録
			return a.registry.RegisterPluginKeyBindings(pluginAdapter)
		}
	}
	
	return nil
}

// UnloadPlugin implements domain.PluginManagerInterface
func (a *PluginManagerAdapter) UnloadPlugin(name string) error {
	// プラグインのアンロード前にコマンドとモードを削除
	if a.registry != nil {
		plugin, found := a.pm.GetPlugin(name)
		if found {
			pluginAdapter := &PluginAdapter{plugin: plugin}
			// キーバインドを削除
			if err := a.registry.UnregisterPluginKeyBindings(pluginAdapter); err != nil {
				// キーバインド削除に失敗してもプラグインのアンロードは続行
				// エラーログは後で実装
			}
			// モードを削除
			if err := a.registry.UnregisterPluginModes(pluginAdapter); err != nil {
				// モード削除に失敗してもプラグインのアンロードは続行
				// エラーログは後で実装
			}
			// コマンドを削除
			if err := a.registry.UnregisterPluginCommands(pluginAdapter); err != nil {
				// コマンド削除に失敗してもプラグインのアンロードは続行
				// エラーログは後で実装
			}
		}
	}
	
	return a.pm.UnloadPlugin(name)
}

// GetPlugin implements domain.PluginManagerInterface
func (a *PluginManagerAdapter) GetPlugin(name string) (domain.PluginInterface, bool) {
	plugin, found := a.pm.GetPlugin(name)
	if !found {
		return nil, false
	}
	return &PluginAdapter{plugin: plugin}, true
}

// ListPlugins implements domain.PluginManagerInterface
func (a *PluginManagerAdapter) ListPlugins() []domain.PluginInfo {
	pluginInfos := a.pm.ListPlugins()
	result := make([]domain.PluginInfo, len(pluginInfos))
	
	for i, info := range pluginInfos {
		result[i] = domain.PluginInfo{
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			State:       int(info.State),
			Enabled:     info.Enabled,
		}
	}
	
	return result
}

// Shutdown implements domain.PluginManagerInterface
func (a *PluginManagerAdapter) Shutdown() error {
	return a.pm.Shutdown()
}

// SetPluginConfig sets the configuration for a plugin
func (a *PluginManagerAdapter) SetPluginConfig(name string, config map[string]interface{}) error {
	return a.pm.SetPluginConfig(name, config)
}

// GetPluginConfig gets the configuration for a plugin
func (a *PluginManagerAdapter) GetPluginConfig(name string) (map[string]interface{}, error) {
	return a.pm.GetPluginConfig(name)
}

// SetPluginConfigValue sets a single configuration value for a plugin
func (a *PluginManagerAdapter) SetPluginConfigValue(name string, key string, value interface{}) error {
	return a.pm.SetPluginConfigValue(name, key, value)
}

// PluginAdapter はplugin.Pluginをdomain.PluginInterfaceに適合させる
type PluginAdapter struct {
	plugin Plugin
}

// Name implements domain.PluginInterface
func (a *PluginAdapter) Name() string {
	return a.plugin.Name()
}

// Version implements domain.PluginInterface
func (a *PluginAdapter) Version() string {
	return a.plugin.Version()
}

// Description implements domain.PluginInterface
func (a *PluginAdapter) Description() string {
	return a.plugin.Description()
}

// GetCommands implements domain.PluginInterface
func (a *PluginAdapter) GetCommands() []domain.CommandSpec {
	commands := a.plugin.GetCommands()
	result := make([]domain.CommandSpec, len(commands))
	
	for i, cmd := range commands {
		result[i] = domain.CommandSpec{
			Name:        cmd.Name,
			Description: cmd.Description,
			Interactive: cmd.Interactive,
			Handler:     cmd.Handler,
		}
	}
	
	return result
}

// GetMajorModes implements domain.PluginInterface
func (a *PluginAdapter) GetMajorModes() []domain.MajorModeSpec {
	modes := a.plugin.GetMajorModes()
	result := make([]domain.MajorModeSpec, len(modes))
	
	for i, mode := range modes {
		keyBindings := make([]domain.KeyBindingSpec, len(mode.KeyBindings))
		for j, kb := range mode.KeyBindings {
			keyBindings[j] = domain.KeyBindingSpec{
				Sequence: kb.Sequence,
				Command:  kb.Command,
				Mode:     kb.Mode,
			}
		}
		
		result[i] = domain.MajorModeSpec{
			Name:         mode.Name,
			Extensions:   mode.Extensions,
			Description:  mode.Description,
			KeyBindings:  keyBindings,
		}
	}
	
	return result
}

// GetMinorModes implements domain.PluginInterface
func (a *PluginAdapter) GetMinorModes() []domain.MinorModeSpec {
	modes := a.plugin.GetMinorModes()
	result := make([]domain.MinorModeSpec, len(modes))
	
	for i, mode := range modes {
		keyBindings := make([]domain.KeyBindingSpec, len(mode.KeyBindings))
		for j, kb := range mode.KeyBindings {
			keyBindings[j] = domain.KeyBindingSpec{
				Sequence: kb.Sequence,
				Command:  kb.Command,
				Mode:     kb.Mode,
			}
		}
		
		result[i] = domain.MinorModeSpec{
			Name:        mode.Name,
			Description: mode.Description,
			Global:      mode.Global,
			KeyBindings: keyBindings,
		}
	}
	
	return result
}

// GetKeyBindings implements domain.PluginInterface
func (a *PluginAdapter) GetKeyBindings() []domain.KeyBindingSpec {
	keyBindings := a.plugin.GetKeyBindings()
	result := make([]domain.KeyBindingSpec, len(keyBindings))
	
	for i, kb := range keyBindings {
		result[i] = domain.KeyBindingSpec{
			Sequence: kb.Sequence,
			Command:  kb.Command,
			Mode:     kb.Mode,
		}
	}
	
	return result
}

// ExecuteCommand implements CommandPlugin interface for command execution
func (a *PluginAdapter) ExecuteCommand(name string, args ...interface{}) error {
	gmacslog.Debug("Executing command '%s' on plugin '%s'", name, a.plugin.Name())
	
	// Try to cast to CommandPlugin interface
	if cmdPlugin, ok := a.plugin.(CommandPlugin); ok {
		err := cmdPlugin.ExecuteCommand(name, args...)
		if err != nil {
			gmacslog.Error("Plugin command execution failed: %v", err)
		}
		return err
	}
	
	// If the underlying plugin doesn't implement CommandPlugin, return error
	gmacslog.Error("Plugin %s does not implement CommandPlugin interface", a.plugin.Name())
	return fmt.Errorf("plugin %s does not implement CommandPlugin interface", a.plugin.Name())
}

// GetCompletions implements CommandPlugin interface for command completion
func (a *PluginAdapter) GetCompletions(command string, prefix string) []string {
	// Try to cast to CommandPlugin interface
	if cmdPlugin, ok := a.plugin.(CommandPlugin); ok {
		return cmdPlugin.GetCompletions(command, prefix)
	}
	
	// If the underlying plugin doesn't implement CommandPlugin, return empty
	return []string{}
}

// HostInterfaceImpl はプラグインに提供するホストインターフェースの実装
type HostInterfaceImpl struct {
	editor *domain.Editor
}

// NewHostInterface creates a new HostInterface implementation
func NewHostInterface(editor *domain.Editor) *HostInterfaceImpl {
	return &HostInterfaceImpl{editor: editor}
}

// エディタ操作
func (h *HostInterfaceImpl) GetCurrentBuffer() interface{} {
	if h.editor != nil {
		return h.editor.CurrentBuffer()
	}
	return nil
}

func (h *HostInterfaceImpl) GetCurrentWindow() interface{} {
	if h.editor != nil {
		return h.editor.CurrentWindow()
	}
	return nil
}

func (h *HostInterfaceImpl) SetStatus(message string) {
	if h.editor != nil {
		h.editor.SetMinibufferMessage(message)
	}
}

func (h *HostInterfaceImpl) ShowMessage(message string) {
	if h.editor != nil {
		h.editor.SetMinibufferMessage(message)
	}
}

// コマンド実行
func (h *HostInterfaceImpl) ExecuteCommand(name string, args ...interface{}) error {
	if h.editor != nil {
		cmd, exists := h.editor.CommandRegistry().Get(name)
		if !exists {
			return fmt.Errorf("command not found: %s", name)
		}
		return cmd.Execute(h.editor)
	}
	return fmt.Errorf("editor not available")
}

// モード管理
func (h *HostInterfaceImpl) SetMajorMode(bufferName, modeName string) error {
	// TODO: Implement major mode setting
	return fmt.Errorf("SetMajorMode not implemented")
}

func (h *HostInterfaceImpl) ToggleMinorMode(bufferName, modeName string) error {
	// TODO: Implement minor mode toggling
	return fmt.Errorf("ToggleMinorMode not implemented")
}

// イベント・フック
func (h *HostInterfaceImpl) AddHook(event string, handler func(...interface{}) error) {
	if h.editor != nil {
		h.editor.AddHook(event, handler)
	}
}

func (h *HostInterfaceImpl) TriggerHook(event string, args ...interface{}) {
	if h.editor != nil {
		h.editor.TriggerHook(event, args...)
	}
}

// バッファ操作
func (h *HostInterfaceImpl) CreateBuffer(name string) interface{} {
	// TODO: Implement buffer creation
	return nil
}

func (h *HostInterfaceImpl) FindBuffer(name string) interface{} {
	if h.editor != nil {
		return h.editor.FindBuffer(name)
	}
	return nil
}

func (h *HostInterfaceImpl) SwitchToBuffer(name string) error {
	if h.editor != nil {
		buffer := h.editor.FindBuffer(name)
		if buffer != nil {
			h.editor.SwitchToBuffer(buffer)
			return nil
		}
		return fmt.Errorf("buffer not found: %s", name)
	}
	return fmt.Errorf("editor not available")
}

// ファイル操作
func (h *HostInterfaceImpl) OpenFile(path string) error {
	// TODO: Implement file opening
	return fmt.Errorf("OpenFile not implemented")
}

func (h *HostInterfaceImpl) SaveBuffer(bufferName string) error {
	// TODO: Implement buffer saving
	return fmt.Errorf("SaveBuffer not implemented")
}

// 設定
func (h *HostInterfaceImpl) GetOption(name string) (interface{}, error) {
	if h.editor != nil {
		return h.editor.GetOption(name)
	}
	return nil, fmt.Errorf("editor not available")
}

func (h *HostInterfaceImpl) SetOption(name string, value interface{}) error {
	if h.editor != nil {
		return h.editor.SetOption(name, value)
	}
	return fmt.Errorf("editor not available")
}

// CreateEditorWithPlugins はプラグインマネージャー付きのエディタを作成する
func CreateEditorWithPlugins(configLoader domain.ConfigLoader, hookManager domain.HookManager) *domain.Editor {
	return CreateEditorWithPluginsAndPaths(configLoader, hookManager, GetDefaultPluginPaths())
}

// CreateEditorWithPluginsAndPaths は指定されたプラグインパスでエディタを作成する
func CreateEditorWithPluginsAndPaths(configLoader domain.ConfigLoader, hookManager domain.HookManager, pluginPaths []string) *domain.Editor {
	// エディタを作成
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// プラグインマネージャーを作成
	pluginManager := NewPluginManagerWithPaths(pluginPaths)
	pluginManagerAdapter := NewPluginManagerAdapterWithRegistry(pluginManager, editor)
	
	// エディタにプラグインマネージャーを設定
	editor.SetPluginManager(pluginManagerAdapter)
	
	// auto-discoveryでロードされたプラグインのコマンドを登録
	loadedPlugins := pluginManager.ListPlugins()
	
	for _, pluginInfo := range loadedPlugins {
		if plugin, found := pluginManager.GetPlugin(pluginInfo.Name); found {
			pluginAdapter := &PluginAdapter{plugin: plugin}
			
			// コマンドを登録
			if err := editor.RegisterPluginCommands(pluginAdapter); err != nil {
				gmacslog.Error("Failed to register commands for plugin %s: %v", pluginInfo.Name, err)
			}
		}
	}
	
	return editor
}

// LoadPluginConfigIfExists loads plugin configuration if available
func LoadPluginConfigIfExists(configLoader domain.ConfigLoader) error {
	// ConfigLoaderがluaconfigパッケージのものかチェックし、
	// plugin configをロードする
	// これは型アサーションで実装
	
	// interface{} として受け取った configLoader から実際の型にキャスト
	if cl, ok := configLoader.(interface{ LoadPluginConfig() error }); ok {
		return cl.LoadPluginConfig()
	}
	
	return nil // plugin config loading not supported
}