# gmacs Plugin System Design

## 概要

gmacs に HashiCorp go-plugin を使用したプラグインシステムを導入する設計文書。Emacsライクな拡張性を提供しつつ、既存のアーキテクチャとの互換性を保つ。

## 現在のアーキテクチャ分析

### 既存システムの特徴
- Emacsライクなテキストエディタ
- ドメインロジックとCLI部分の明確な分離
- Luaベースの設定システム (`lua-config/`)
- コマンド登録システム (`CommandRegistry`)
- キーバインディングシステム (`KeyBindingMap`)
- モードシステム（Major/Minor Mode）
- イベント駆動アーキテクチャ

### プラグイン要件
- **Major Mode拡張**: 新しい言語サポート（Go、Rust、Python等）
- **Minor Mode拡張**: 追加機能（Git統合、LSP、Auto-complete等）
- **Command拡張**: 新しいコマンド追加
- **Hook拡張**: イベントハンドリング拡張

## HashiCorp go-plugin アーキテクチャ

### 基本構成

```
gmacs (Host Process)
├── plugin/
│   ├── interface.go     # プラグインインターフェース定義
│   ├── manager.go       # プラグインマネージャー
│   ├── registry.go      # プラグイン登録・発見
│   ├── host_api.go      # ホストAPI実装
│   ├── rpc.go          # gRPC通信層
│   └── protocol.proto   # gRPCプロトコル定義

# ユーザー環境でのプラグイン配置例（XDG準拠）
$XDG_DATA_HOME/gmacs/plugins/    # ~/.local/share/gmacs/plugins/
├── go-mode/
│   ├── manifest.json
│   └── go-mode          # 独立したバイナリ
├── rust-mode/
│   ├── manifest.json
│   └── rust-mode
└── lsp-client/
    ├── manifest.json
    └── lsp-client

/usr/share/gmacs/plugins/        # システムワイドプラグイン
├── markdown-mode/
│   ├── manifest.json
│   └── markdown-mode
└── git-integration/
    ├── manifest.json
    └── git-integration

# 設定ファイル配置
$XDG_CONFIG_HOME/gmacs/          # ~/.config/gmacs/
├── init.lua                     # メイン設定
├── plugins.lua                  # プラグイン設定
└── keybindings.lua             # キーバインド設定
```

### 通信プロトコル
- **gRPCベース**: 型安全な通信、自動シリアライゼーション
- **別プロセス実行**: クラッシュ分離、セキュリティ向上
- **自動プロセス管理**: ヘルスチェック、クラッシュ検出・回復
- **言語非依存**: gRPCサポート言語でプラグイン開発可能

## プラグインインターフェース設計

### Core Plugin Interface

```go
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
```

### Host Interface

```go
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
```

### 専用インターフェース

```go
// MajorModePlugin は新しいメジャーモードを提供
type MajorModePlugin interface {
    Plugin
    
    // モード固有の処理
    OnActivate(buffer BufferInterface) error
    OnDeactivate(buffer BufferInterface) error
    OnFileOpen(buffer BufferInterface, filename string) error
    OnFileSave(buffer BufferInterface, filename string) error
    
    // シンタックスハイライト（将来拡張）
    GetSyntaxHighlighting() SyntaxSpec
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
```

## プラグインマネージャー設計

### PluginManager 構造

```go
type PluginManager struct {
    plugins     map[string]*LoadedPlugin
    registry    *PluginRegistry
    config      *PluginConfig
    hostAPI     *HostAPI
    searchPaths []string
    
    // go-plugin client管理
    clients map[string]*plugin.Client
}

type LoadedPlugin struct {
    Name      string
    Version   string
    Path      string
    Client    *plugin.Client
    Plugin    Plugin
    Config    map[string]interface{}
    State     PluginState
    Manifest  *PluginManifest
}

type PluginState int

const (
    PluginStateUnloaded PluginState = iota
    PluginStateLoading
    PluginStateLoaded
    PluginStateError
)
```

### プラグインライフサイクル管理

```go
// プラグインライフサイクル管理
func (pm *PluginManager) DiscoverPlugins() ([]PluginManifest, error)
func (pm *PluginManager) LoadPlugin(name string) error
func (pm *PluginManager) UnloadPlugin(name string) error
func (pm *PluginManager) ReloadPlugin(name string) error
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool)
func (pm *PluginManager) ListPlugins() []PluginInfo
func (pm *PluginManager) EnablePlugin(name string) error
func (pm *PluginManager) DisablePlugin(name string) error

// 初期化と終了処理
func (pm *PluginManager) Initialize() error
func (pm *PluginManager) Shutdown() error
```

### 既存システムとの統合

```go
// Editor 拡張
type Editor struct {
    // ... 既存フィールド
    pluginManager *PluginManager
}

// エディタ初期化時にプラグインシステムを統合
func NewEditorWithPlugins(configLoader ConfigLoader, hookManager HookManager) *Editor {
    editor := newEditorWithConfig(EditorConfig{
        ConfigLoader: configLoader,
        HookManager:  hookManager,
    })
    
    // プラグインマネージャー初期化
    editor.pluginManager = NewPluginManager(editor)
    
    // プラグイン自動ロード
    editor.loadConfiguredPlugins()
    
    return editor
}

// プラグインコマンド統合
func (e *Editor) registerPluginCommands() {
    for _, pluginInfo := range e.pluginManager.ListPlugins() {
        plugin, _ := e.pluginManager.GetPlugin(pluginInfo.Name)
        
        for _, cmd := range plugin.GetCommands() {
            // コマンド登録
            e.commandRegistry.RegisterFunc(cmd.Name, func(editor *Editor) error {
                commandPlugin, ok := plugin.(CommandPlugin)
                if !ok {
                    return fmt.Errorf("plugin %s does not support commands", pluginInfo.Name)
                }
                return commandPlugin.ExecuteCommand(cmd.Name, editor)
            })
        }
        
        // キーバインディング登録
        for _, binding := range plugin.GetKeyBindings() {
            if binding.Mode == "" {
                // グローバルキーバインディング
                e.keyBindings.BindKeySequence(binding.Sequence, func(editor *Editor) error {
                    return e.commandRegistry.Execute(binding.Command, editor)
                })
            } else {
                // モード固有キーバインディング
                e.LocalBindKey(binding.Mode, binding.Sequence, binding.Command)
            }
        }
    }
}

// プラグインモード統合
func (e *Editor) registerPluginModes() {
    for _, pluginInfo := range e.pluginManager.ListPlugins() {
        plugin, _ := e.pluginManager.GetPlugin(pluginInfo.Name)
        
        // メジャーモード登録
        for _, modeSpec := range plugin.GetMajorModes() {
            if majorModePlugin, ok := plugin.(MajorModePlugin); ok {
                mode := &PluginMajorMode{
                    name:   modeSpec.Name,
                    plugin: majorModePlugin,
                    spec:   modeSpec,
                }
                e.modeManager.RegisterMajorMode(mode)
            }
        }
        
        // マイナーモード登録
        for _, modeSpec := range plugin.GetMinorModes() {
            if minorModePlugin, ok := plugin.(MinorModePlugin); ok {
                mode := &PluginMinorMode{
                    name:   modeSpec.Name,
                    plugin: minorModePlugin,
                    spec:   modeSpec,
                }
                e.modeManager.RegisterMinorMode(mode)
            }
        }
    }
}
```

## 設定とディスカバリー機能

### プラグイン設定

```go
// プラグイン設定構造
type PluginConfig struct {
    PluginDir     string                 `json:"plugin_dir"`
    SearchPaths   []string              `json:"search_paths"`
    Enabled       []string              `json:"enabled"`
    Disabled      []string              `json:"disabled"`
    AutoLoad      bool                  `json:"auto_load"`
    PluginConfigs map[string]interface{} `json:"plugin_configs"`
}

// デフォルト設定（XDG準拠）
func getDefaultPluginConfig() PluginConfig {
    userDataDir := getXDGDataHome()   // $XDG_DATA_HOME or ~/.local/share
    
    return PluginConfig{
        PluginDir: filepath.Join(userDataDir, "gmacs", "plugins"),
        SearchPaths: []string{
            filepath.Join(userDataDir, "gmacs", "plugins"),  // ユーザーローカル
            "/usr/share/gmacs/plugins/",                     // システムワイド
            "/usr/local/share/gmacs/plugins/",               // ローカルシステム
        },
        AutoLoad: true,
        PluginConfigs: make(map[string]interface{}),
    }
}

// XDG Base Directory Specification準拠のパス取得
func getXDGDataHome() string {
    if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
        return dataHome
    }
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".local", "share")
}

func getXDGConfigHome() string {
    if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
        return configHome
    }
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".config")
}
```

### プラグインマニフェスト

```go
// プラグインメタデータ（manifest.json）
type PluginManifest struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    Description  string            `json:"description"`
    Author       string            `json:"author"`
    Binary       string            `json:"binary"`
    Dependencies []string          `json:"dependencies"`
    MinGmacs     string            `json:"min_gmacs_version"`
    Config       map[string]interface{} `json:"default_config"`
}

// manifest.json例
{
    "name": "go-mode",
    "version": "1.0.0",
    "description": "Go language support for gmacs",
    "author": "example@example.com",
    "binary": "go-mode",
    "dependencies": [],
    "min_gmacs_version": "0.1.0",
    "default_config": {
        "auto_format": true,
        "gofmt_on_save": true,
        "use_goimports": false
    }
}
```

### 自動ディスカバリー

```go
type PluginDiscovery struct {
    searchPaths []string
    cache       map[string]*PluginManifest
}

func (pd *PluginDiscovery) ScanPlugins() ([]PluginManifest, error) {
    var manifests []PluginManifest
    
    for _, path := range pd.searchPaths {
        entries, err := os.ReadDir(path)
        if err != nil {
            continue
        }
        
        for _, entry := range entries {
            if entry.IsDir() {
                manifestPath := filepath.Join(path, entry.Name(), "manifest.json")
                if manifest, err := pd.loadManifest(manifestPath); err == nil {
                    manifests = append(manifests, *manifest)
                }
            }
        }
    }
    
    return manifests, nil
}

func (pd *PluginDiscovery) loadManifest(path string) (*PluginManifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var manifest PluginManifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return nil, err
    }
    
    return &manifest, nil
}
```

### Lua設定統合

```go
// 既存のLua設定システムに統合
func (api *APIBindings) RegisterPluginAPI() error {
    // plugin テーブル作成
    pluginTable := api.vm.L.NewTable()
    api.vm.L.SetGlobal("plugin", pluginTable)
    
    // plugin.setup(config)
    api.vm.L.SetField(pluginTable, "setup", api.vm.L.NewFunction(api.pluginSetup))
    
    // plugin.load(name)
    api.vm.L.SetField(pluginTable, "load", api.vm.L.NewFunction(api.pluginLoad))
    
    // plugin.unload(name)  
    api.vm.L.SetField(pluginTable, "unload", api.vm.L.NewFunction(api.pluginUnload))
    
    // plugin.enable(name)
    api.vm.L.SetField(pluginTable, "enable", api.vm.L.NewFunction(api.pluginEnable))
    
    // plugin.disable(name)
    api.vm.L.SetField(pluginTable, "disable", api.vm.L.NewFunction(api.pluginDisable))
    
    // plugin.list()
    api.vm.L.SetField(pluginTable, "list", api.vm.L.NewFunction(api.pluginList))
    
    // plugin.config(name, config)
    api.vm.L.SetField(pluginTable, "config", api.vm.L.NewFunction(api.pluginConfig))
    
    return nil
}
```

### Lua設定例

```lua
-- ~/.config/gmacs/init.lua にプラグイン設定を追加（XDG_CONFIG_HOME準拠）

-- プラグインシステム設定
plugin.setup({
    plugin_dir = "~/.local/share/gmacs/plugins",  -- XDG_DATA_HOME準拠
    auto_load = true,
    search_paths = {
        "~/.local/share/gmacs/plugins/",      -- ユーザーローカル（XDG_DATA_HOME）
        "/usr/share/gmacs/plugins/",          -- システムワイド
        "/usr/local/share/gmacs/plugins/"     -- ローカルシステム
    }
})

-- 個別プラグイン有効化
plugin.enable("go-mode")
plugin.enable("markdown-mode")
plugin.enable("git-integration")

-- プラグイン固有設定
plugin.config("go-mode", {
    auto_format = true,
    gofmt_on_save = true,
    use_goimports = true
})

plugin.config("markdown-mode", {
    live_preview = false,
    wrap_lines = true
})

-- プラグインインストール管理（将来的な拡張）
-- plugin.install("github.com/example/gmacs-python-mode")
-- plugin.update("go-mode")
-- plugin.remove("old-plugin")
```

## gRPCプロトコル定義

```protobuf
// plugin/protocol.proto
syntax = "proto3";

package gmacs.plugin;

option go_package = "github.com/TakahashiShuuhei/gmacs/plugin";

// プラグインサービス
service PluginService {
    rpc Initialize(InitRequest) returns (InitResponse);
    rpc GetInfo(InfoRequest) returns (InfoResponse);
    rpc GetCommands(CommandsRequest) returns (CommandsResponse);
    rpc GetModes(ModesRequest) returns (ModesResponse);
    rpc ExecuteCommand(ExecuteRequest) returns (ExecuteResponse);
    rpc Cleanup(CleanupRequest) returns (CleanupResponse);
}

// ホストサービス（プラグインからホストへの呼び出し）
service HostService {
    rpc GetCurrentBuffer(BufferRequest) returns (BufferResponse);
    rpc ExecuteCommand(HostExecuteRequest) returns (HostExecuteResponse);
    rpc SetStatus(StatusRequest) returns (StatusResponse);
    rpc AddHook(HookRequest) returns (HookResponse);
}

message InitRequest {
    map<string, string> config = 1;
}

message InitResponse {
    bool success = 1;
    string error = 2;
}

message InfoRequest {}

message InfoResponse {
    string name = 1;
    string version = 2;
    string description = 3;
}

message CommandsRequest {}

message CommandsResponse {
    repeated CommandSpec commands = 1;
}

message CommandSpec {
    string name = 1;
    string description = 2;
    bool interactive = 3;
}

message ExecuteRequest {
    string command = 1;
    repeated string args = 2;
}

message ExecuteResponse {
    bool success = 1;
    string error = 2;
    string result = 3;
}
```

## プラグインビルドシステム設計

### PluginBuilder 構造

```go
type PluginBuilder struct {
    workspace   string                    // ビルド作業ディレクトリ
    goPath      string                    // Go実行パス
    targetDir   string                    // プラグイン配置先
    cache       map[string]*BuildCache    // ビルドキャッシュ
}

type BuildSpec struct {
    Repository string                     // Git repository URL
    Ref        string                     // branch/tag/commit
    LocalPath  string                     // ローカルパス（開発用）
}

type BuildCache struct {
    Hash       string                     // ソースコードハッシュ
    BuildTime  time.Time                  // ビルド時刻
    BinaryPath string                     // ビルド済みバイナリパス
}

// ビルドプロセス
func (pb *PluginBuilder) BuildPlugin(spec BuildSpec) (*PluginManifest, error) {
    // 1. ソースコード取得
    srcDir, err := pb.fetchSource(spec)
    if err != nil {
        return nil, err
    }
    
    // 2. manifest.json読み込み
    manifest, err := pb.loadManifest(srcDir)
    if err != nil {
        return nil, err
    }
    
    // 3. 依存関係チェック
    if err := pb.checkDependencies(srcDir); err != nil {
        return nil, err
    }
    
    // 4. ビルド実行
    binaryPath, err := pb.buildBinary(srcDir, manifest.Name)
    if err != nil {
        return nil, err
    }
    
    // 5. プラグインディレクトリに配置
    if err := pb.installPlugin(manifest, binaryPath); err != nil {
        return nil, err
    }
    
    return manifest, nil
}
```

### CLI インターフェース

```go
// gmacs plugin サブコマンド実装
func main() {
    if len(os.Args) > 1 && os.Args[1] == "plugin" {
        pluginManager := NewPluginManager()
        
        switch os.Args[2] {
        case "install":
            repo := os.Args[3]
            err := pluginManager.InstallFromSource(repo)
            
        case "update":
            name := os.Args[3]
            err := pluginManager.UpdatePlugin(name)
            
        case "remove":
            name := os.Args[3]
            err := pluginManager.RemovePlugin(name)
            
        case "list":
            pluginManager.ListPlugins()
            
        case "dev":
            // 開発モード: ローカルパスからビルド
            localPath := os.Args[3]
            err := pluginManager.InstallFromLocal(localPath)
        }
    }
}
```

## 実装プラン

### Phase 1: 基盤構築（Week 1-2）

1. **依存関係追加**
   ```bash
   go mod tidy
   # go.modに追加:
   # github.com/hashicorp/go-plugin v1.4.0
   # github.com/go-git/go-git/v5 v5.4.2  # Git operations
   ```

2. **基本インターフェース実装**
   - `plugin/interface.go`: コアインターフェース定義
   - `plugin/protocol.proto`: gRPCプロトコル定義
   - `plugin/rpc.go`: gRPC実装

3. **プロトコルバッファ生成**
   ```bash
   protoc --go_out=. --go-grpc_out=. plugin/protocol.proto
   ```

### Phase 2: コア機能実装（Week 3-4）

1. **PluginManager実装**
   - `plugin/manager.go`: プラグインライフサイクル管理
   - `plugin/registry.go`: プラグイン登録・発見
   - プロセス管理、エラーハンドリング

2. **HostAPI実装**
   - `plugin/host_api.go`: プラグインからgmacsへのAPI
   - バッファ・ウィンドウ操作API
   - コマンド実行API

3. **Editor統合**
   - `domain/editor.go`拡張: プラグインマネージャー統合
   - コマンド・モード・キーバインド統合
   - 既存システムとの互換性保持

### Phase 3: 設定システム（Week 5）

1. **Lua統合**
   - `lua-config/plugin_api.go`: プラグイン設定API
   - Luaからのプラグイン制御
   - XDG準拠の設定ファイルサポート

2. **ディスカバリー機能**
   - `plugin/discovery.go`: 自動プラグイン発見
   - マニフェストファイル処理
   - XDG準拠のパス管理
   - レガシー設定パスとの互換性

### Phase 4: プラグインSDKとビルドシステム（Week 6-7）

1. **プラグインSDK作成**
   - `gmacs-plugin-sdk` パッケージ
   - プラグイン開発用ヘルパー
   - プラグインテンプレート生成ツール

2. **ビルドシステム実装**
   - `plugin/builder.go`: ソースからビルド
   - Git リポジトリクローン機能
   - Go ビルドチェーン統合
   - 依存関係管理

3. **サンプルプラグイン（別リポジトリ）**
   - `gmacs-go-mode`: Goファイル用メジャーモード
   - `gmacs-markdown-mode`: Markdownファイル用メジャーモード
   - ビルド可能な形式で配布

### Phase 5: テストとドキュメント（Week 8）

1. **E2Eテスト追加**
   - プラグインロード・アンロードテスト
   - プラグインコマンド実行テスト
   - エラーハンドリングテスト

2. **ドキュメント更新**
   - `PLUGIN_DESIGN.md`（このファイル）
   - プラグイン開発ガイド
   - プラグインAPI リファレンス
   - プラグイン配布ガイドライン
   - XDG準拠のディレクトリ構造ガイド

## 利点と考慮事項

### 利点

1. **安全性**
   - 別プロセス実行によるクラッシュ分離
   - プラグインがホストをクラッシュさせない
   - セキュリティサンドボックス

2. **拡張性**
   - 言語非依存（gRPCサポート言語）
   - Emacsライクな豊富なプラグインエコシステム可能
   - 動的ロード・アンロード

3. **保守性**
   - 既存アーキテクチャとの互換性保持
   - プラグインとホストの独立開発
   - 明確なAPI境界

4. **パフォーマンス**
   - プラグインプロセス分離によるパフォーマンス分離
   - 必要に応じたプラグインロード
   - プラグイン障害がエディタに影響しない

5. **エコシステム**
   - サードパーティ開発者による自由なプラグイン開発
   - 独立したリリースサイクル
   - コミュニティ主導の機能拡張

### 考慮事項

1. **プロセス間通信オーバーヘッド**
   - gRPC通信コスト
   - 大量データ転送時の性能
   - 対策: バッチ処理、キャッシュ活用

2. **複雑性増加**
   - プロセス管理の複雑さ
   - デバッグの困難さ
   - 対策: 充実したログ、デバッグツール

3. **プラグイン開発コスト**
   - gRPCプロトコルの学習コスト
   - ビルド・デプロイ手順
   - 対策: SDKとテンプレート提供

4. **配布とインストール**
   - プラグインバイナリの配布方法
   - バージョン管理と互換性
   - 対策: パッケージマネージャー統合（将来）

## プラグインインストール手順

### 方法1: 手動インストール（初期段階）

1. **プラグインバイナリをダウンロード**
   ```bash
   # GitHubリリースからダウンロード例（初期のみ）
   wget https://github.com/example/gmacs-go-mode/releases/latest/download/go-mode-linux-amd64.tar.gz
   tar -xzf go-mode-linux-amd64.tar.gz
   ```

2. **プラグインディレクトリに配置**
   ```bash
   # ユーザーローカルにインストール
   mkdir -p ~/.local/share/gmacs/plugins/go-mode
   cp go-mode ~/.local/share/gmacs/plugins/go-mode/
   cp manifest.json ~/.local/share/gmacs/plugins/go-mode/
   ```

3. **設定ファイルで有効化**
   ```lua
   -- ~/.config/gmacs/init.lua
   plugin.enable("go-mode")
   ```

4. **gmacs再起動**
   ```bash
   gmacs  # プラグインが自動ロードされる
   ```

### 方法2: スクリプト経由インストール（中期）

1. **インストールスクリプト実行**
   ```bash
   # プラグイン配布者が提供するインストールスクリプト（ソースビルド版）
   curl -fsSL https://raw.githubusercontent.com/example/gmacs-go-mode/main/install.sh | bash
   
   # または
   wget -qO- https://github.com/example/gmacs-go-mode/main/install.sh | bash
   ```

2. **自動ビルドプロセス**
   ```bash
   # スクリプトが自動で以下を実行:
   # - git clone でソースコード取得
   # - go mod download で依存関係解決
   # - go build でバイナリ作成
   # - プラグインディレクトリに配置
   # - manifest.json作成
   ```

### 方法3: ソースからビルド（推奨）

1. **GitHubリポジトリから直接インストール**
   ```bash
   # gmacs自体にプラグインマネージャー機能を追加
   gmacs plugin install github.com/example/gmacs-go-mode
   gmacs plugin install gitlab.com/user/gmacs-rust-mode
   gmacs plugin install https://git.example.com/plugins/lsp-client
   
   # ローカルパスからもインストール可能
   gmacs plugin install ./my-custom-plugin/
   ```

2. **自動ビルドプロセス**
   ```bash
   # gmacs が自動で以下を実行:
   # 1. git clone でソースコード取得
   # 2. go mod download で依存関係解決
   # 3. go build でバイナリ作成
   # 4. 適切なディレクトリに配置
   # 5. manifest.json 生成
   ```

3. **システムパッケージマネージャー使用（補完的）**
   ```bash
   # Arch Linux (AUR) - ビルドスクリプト付き
   yay -S gmacs-go-mode-plugin
   
   # Ubuntu/Debian (将来的にPPAで)
   sudo apt install gmacs-plugin-go-mode
   
   # Homebrew (macOS)
   brew install gmacs-go-mode
   ```

### プラグイン発見方法

1. **公式プラグインリスト**
   - GitHub: `gmacs-plugins` organization
   - ウェブサイト: プラグインカタログページ

2. **GitHub検索**
   ```bash
   # GitHubでプラグインを検索
   topic:gmacs-plugin
   "gmacs plugin" language:go
   ```

3. **コミュニティリソース**
   - Reddit: r/gmacs
   - Discord/Slack コミュニティ
   - awesome-gmacs リポジトリ

### プラグイン設定例

```lua
-- ~/.config/gmacs/init.lua

-- プラグインシステム設定
plugin.setup({
    auto_load = true,
    search_paths = {
        "~/.local/share/gmacs/plugins/",
        "/usr/share/gmacs/plugins/"
    }
})

-- インストール済みプラグインを有効化
plugin.enable("go-mode")
plugin.enable("rust-mode")
plugin.enable("lsp-client")

-- プラグイン固有設定
plugin.config("go-mode", {
    auto_format = true,
    gofmt_on_save = true,
    use_goimports = true
})

plugin.config("lsp-client", {
    servers = {
        gopls = { cmd = {"gopls"} },
        rust_analyzer = { cmd = {"rust-analyzer"} }
    }
})
```

### トラブルシューティング

1. **プラグインが見つからない**
   ```bash
   # デバッグモードで起動
   gmacs --debug
   
   # プラグインパスを確認
   ls -la ~/.local/share/gmacs/plugins/
   ```

2. **プラグインが動作しない**
   ```lua
   -- Luaコンソールで確認
   :lua print(plugin.list())
   :lua print(plugin.status("go-mode"))
   ```

3. **設定の確認**
   ```bash
   # 設定ファイルの文法チェック
   lua -c "dofile('~/.config/gmacs/init.lua')"
   ```

## まとめ

この設計により、gmacsは以下を実現できる：

- **安全で強力なプラグインシステム**
- **Emacsライクな豊富な拡張性**
- **既存アーキテクチャとの互換性**
- **段階的なプラグインエコシステム構築**
- **ユーザーフレンドリーなインストール体験**

HashiCorp go-pluginの採用により、業界実績のある安定したプラグインアーキテクチャを基盤として、gmacs独自の拡張可能なエディタを構築できる。