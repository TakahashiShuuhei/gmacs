package plugin

import (
	"context"
	"time"
)

// Plugin は全プラグインが実装すべき基本インターフェース
type Plugin interface {
	// プラグイン情報
	Name() string
	Version() string
	Description() string
	
	// ライフサイクル
	Initialize(ctx context.Context, host HostInterface) error
	Cleanup() error
	
	// 機能提供
	GetCommands() []CommandSpec
	GetMajorModes() []MajorModeSpec
	GetMinorModes() []MinorModeSpec
	GetKeyBindings() []KeyBindingSpec
}

// HostInterface はホスト（gmacs）がプラグインに提供するAPI
type HostInterface interface {
	// エディタ操作
	GetCurrentBuffer() BufferInterface
	GetCurrentWindow() WindowInterface
	SetStatus(message string)
	ShowMessage(message string)
	
	// コマンド実行
	ExecuteCommand(name string, args ...interface{}) error
	
	// モード管理
	SetMajorMode(bufferName, modeName string) error
	ToggleMinorMode(bufferName, modeName string) error
	
	// イベント・フック
	AddHook(event string, handler func(...interface{}) error)
	TriggerHook(event string, args ...interface{})
	
	// バッファ操作
	CreateBuffer(name string) BufferInterface
	FindBuffer(name string) BufferInterface
	SwitchToBuffer(name string) error
	
	// ファイル操作
	OpenFile(path string) error
	SaveBuffer(bufferName string) error
	
	// 設定
	GetOption(name string) (interface{}, error)
	SetOption(name string, value interface{}) error
}

// BufferInterface はプラグインからアクセス可能なバッファAPI
type BufferInterface interface {
	Name() string
	Content() string
	SetContent(content string)
	InsertAt(pos int, text string)
	DeleteRange(start, end int)
	CursorPosition() int
	SetCursorPosition(pos int)
	MarkDirty()
	IsDirty() bool
	Filename() string
}

// WindowInterface はプラグインからアクセス可能なウィンドウAPI
type WindowInterface interface {
	Buffer() BufferInterface
	SetBuffer(buffer BufferInterface)
	Width() int
	Height() int
	ScrollOffset() int
	SetScrollOffset(offset int)
}

// CommandSpec はプラグインが提供するコマンド仕様
type CommandSpec struct {
	Name        string
	Description string
	Interactive bool
	Handler     string // プラグイン内のハンドラー名
}

// MajorModeSpec はメジャーモード仕様
type MajorModeSpec struct {
	Name         string
	Extensions   []string // 対象ファイル拡張子
	Description  string
	KeyBindings  []KeyBindingSpec
}

// MinorModeSpec はマイナーモード仕様
type MinorModeSpec struct {
	Name        string
	Description string
	Global      bool // グローバルモードかバッファローカルか
	KeyBindings []KeyBindingSpec
}

// KeyBindingSpec はキーバインディング仕様
type KeyBindingSpec struct {
	Sequence string // "C-c C-g", "M-x" など
	Command  string
	Mode     string // 対象モード（空の場合はグローバル）
}

// MajorModePlugin は新しいメジャーモードを提供
type MajorModePlugin interface {
	Plugin
	
	// モード固有の処理
	OnActivate(buffer BufferInterface) error
	OnDeactivate(buffer BufferInterface) error
	OnFileOpen(buffer BufferInterface, filename string) error
	OnFileSave(buffer BufferInterface, filename string) error
}

// MinorModePlugin は新しいマイナーモードを提供
type MinorModePlugin interface {
	Plugin
	
	// マイナーモード制御
	Enable(buffer BufferInterface) error
	Disable(buffer BufferInterface) error
	IsEnabled(buffer BufferInterface) bool
	
	// モード固有処理
	OnBufferChange(buffer BufferInterface, change ChangeSpec) error
	OnCursorMove(buffer BufferInterface, oldPos, newPos int) error
}

// CommandPlugin は新しいコマンドを提供
type CommandPlugin interface {
	Plugin
	
	// コマンド実行
	ExecuteCommand(name string, args ...interface{}) error
	
	// インタラクティブコマンド用
	GetCompletions(command string, prefix string) []string
}

// ChangeSpec はバッファ変更内容
type ChangeSpec struct {
	Type   ChangeType
	Pos    int
	Length int
	Text   string
}

type ChangeType int

const (
	ChangeTypeInsert ChangeType = iota
	ChangeTypeDelete
	ChangeTypeReplace
)

// PluginManifest はプラグインのメタデータ
type PluginManifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Binary       string                 `json:"binary"`
	Dependencies []string               `json:"dependencies"`
	MinGmacs     string                 `json:"min_gmacs_version"`
	Config       map[string]interface{} `json:"default_config"`
}

// PluginState はプラグインの状態
type PluginState int

const (
	PluginStateUnloaded PluginState = iota
	PluginStateLoading
	PluginStateLoaded
	PluginStateError
)

// LoadedPlugin は読み込み済みプラグイン情報
type LoadedPlugin struct {
	Name      string
	Version   string
	Path      string
	Plugin    Plugin
	Config    map[string]interface{}
	State     PluginState
	Manifest  *PluginManifest
	LoadTime  time.Time
	LastError error
}

// PluginInfo はプラグイン基本情報（リスト表示用）
type PluginInfo struct {
	Name        string
	Version     string
	Description string
	State       PluginState
	Enabled     bool
}

// BuildSpec はプラグインビルド仕様
type BuildSpec struct {
	Repository string // Git repository URL
	Ref        string // branch/tag/commit
	LocalPath  string // ローカルパス（開発用）
}

// BuildCache はビルドキャッシュ情報
type BuildCache struct {
	PluginName   string    // プラグイン名
	Version      string    // バージョン
	Hash         string    // ソースコードハッシュ
	BuildTime    time.Time // ビルド時刻
	BinaryPath   string    // ビルド済みバイナリパス
	ManifestPath string    // manifest.jsonパス
}