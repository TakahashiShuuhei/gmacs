package domain

import (
	"regexp"
)

// PluginMajorMode はプラグインから提供されるメジャーモードのラッパー
type PluginMajorMode struct {
	spec   MajorModeSpec
	plugin PluginInterface
	keyBindings *KeyBindingMap
}

// NewPluginMajorMode は新しいプラグインメジャーモードを作成する
func NewPluginMajorMode(spec MajorModeSpec, plugin PluginInterface) *PluginMajorMode {
	mode := &PluginMajorMode{
		spec:   spec,
		plugin: plugin,
		keyBindings: NewKeyBindingMap(),
	}
	
	// プラグインのキーバインディングを設定
	for _, kb := range spec.KeyBindings {
		// TODO: 実際のプラグインコマンドハンドラーを呼び出す実装
		// 現在はプレースホルダー
		cmdFunc := func(editor *Editor) error {
			message := "Plugin major mode command: " + kb.Command + " from " + plugin.Name()
			editor.SetMinibufferMessage(message)
			return nil
		}
		mode.keyBindings.BindKeySequence(kb.Sequence, cmdFunc)
	}
	
	return mode
}

// Name implements MajorMode interface
func (pm *PluginMajorMode) Name() string {
	return pm.spec.Name
}

// FilePattern implements MajorMode interface
func (pm *PluginMajorMode) FilePattern() *regexp.Regexp {
	if len(pm.spec.Extensions) == 0 {
		return nil
	}
	
	// 拡張子リストから正規表現を構築
	pattern := "\\."
	if len(pm.spec.Extensions) == 1 {
		pattern += regexp.QuoteMeta(pm.spec.Extensions[0]) + "$"
	} else {
		pattern += "(" + regexp.QuoteMeta(pm.spec.Extensions[0])
		for _, ext := range pm.spec.Extensions[1:] {
			pattern += "|" + regexp.QuoteMeta(ext)
		}
		pattern += ")$"
	}
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	
	return regex
}

// KeyBindings implements MajorMode interface
func (pm *PluginMajorMode) KeyBindings() *KeyBindingMap {
	return pm.keyBindings
}

// Commands implements MajorMode interface
func (pm *PluginMajorMode) Commands() map[string]*Command {
	// プラグインコマンドは別途コマンドレジストリに登録されるため、ここでは空
	return make(map[string]*Command)
}

// IndentFunction implements MajorMode interface
func (pm *PluginMajorMode) IndentFunction() IndentFunc {
	// デフォルトのインデント機能
	return func(buffer *Buffer, line int) int {
		return 0 // 基本的なインデント
	}
}

// SyntaxHighlighting implements MajorMode interface
func (pm *PluginMajorMode) SyntaxHighlighting() SyntaxHighlighter {
	// TODO: プラグインからシンタックスハイライト情報を取得する実装
	return nil
}

// Initialize implements MajorMode interface
func (pm *PluginMajorMode) Initialize(buffer *Buffer) error {
	// プラグインの初期化処理
	return nil
}

// OnActivate implements MajorMode interface
func (pm *PluginMajorMode) OnActivate(buffer *Buffer) error {
	// プラグインのアクティベーション処理
	return nil
}

// OnDeactivate implements MajorMode interface
func (pm *PluginMajorMode) OnDeactivate(buffer *Buffer) error {
	// プラグインの非アクティベーション処理
	return nil
}

// PluginMinorMode はプラグインから提供されるマイナーモードのラッパー
type PluginMinorMode struct {
	spec        MinorModeSpec
	plugin      PluginInterface
	keyBindings *KeyBindingMap
	enabled     map[*Buffer]bool
}

// NewPluginMinorMode は新しいプラグインマイナーモードを作成する
func NewPluginMinorMode(spec MinorModeSpec, plugin PluginInterface) *PluginMinorMode {
	mode := &PluginMinorMode{
		spec:        spec,
		plugin:      plugin,
		keyBindings: NewKeyBindingMap(),
		enabled:     make(map[*Buffer]bool),
	}
	
	// プラグインのキーバインディングを設定
	for _, kb := range spec.KeyBindings {
		// TODO: 実際のプラグインコマンドハンドラーを呼び出す実装
		// 現在はプレースホルダー
		cmdFunc := func(editor *Editor) error {
			message := "Plugin minor mode command: " + kb.Command + " from " + plugin.Name()
			editor.SetMinibufferMessage(message)
			return nil
		}
		mode.keyBindings.BindKeySequence(kb.Sequence, cmdFunc)
	}
	
	return mode
}

// Name implements MinorMode interface
func (pm *PluginMinorMode) Name() string {
	return pm.spec.Name
}

// KeyBindings implements MinorMode interface
func (pm *PluginMinorMode) KeyBindings() *KeyBindingMap {
	return pm.keyBindings
}

// Commands implements MinorMode interface
func (pm *PluginMinorMode) Commands() map[string]*Command {
	// プラグインコマンドは別途コマンドレジストリに登録されるため、ここでは空
	return make(map[string]*Command)
}

// Enable implements MinorMode interface
func (pm *PluginMinorMode) Enable(buffer *Buffer) error {
	pm.enabled[buffer] = true
	return nil
}

// Disable implements MinorMode interface
func (pm *PluginMinorMode) Disable(buffer *Buffer) error {
	pm.enabled[buffer] = false
	return nil
}

// IsEnabled implements MinorMode interface
func (pm *PluginMinorMode) IsEnabled(buffer *Buffer) bool {
	enabled, exists := pm.enabled[buffer]
	return exists && enabled
}

// Priority implements MinorMode interface
func (pm *PluginMinorMode) Priority() int {
	// プラグインマイナーモードのデフォルト優先度
	return 100
}