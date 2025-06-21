# Lua設定システム設計書

## 概要

gmacsに.emacsのようなLua設定システムを追加する設計文書です。ユーザーがLuaコードで自由にエディタの動作をカスタマイズできるようにします。

## アーキテクチャ概要

### ディレクトリ構成
```
core/
├── lua-config/
│   ├── lua_vm.go          # Lua仮想マシン管理
│   ├── config_loader.go   # 設定ファイル読み込み
│   ├── api_bindings.go    # Lua→Go APIブリッジ  
│   ├── mode_registry.go   # Luaモード登録
│   └── event_hooks.go     # イベントフック
└── ...
```

### 設定ファイル構造
```lua
-- ~/.gmacs/init.lua (メイン設定)

-- キーバインド設定
gmacs.bind_key("C-x C-s", "save-buffer")
gmacs.bind_key("C-x C-f", "find-file")

-- モード設定
gmacs.major_mode("lua-mode", {
    file_patterns = {"%.lua$"},
    syntax_highlight = true,
    indent_function = lua_indent
})

-- フック設定
gmacs.add_hook("after-save", function(buffer)
    print("Saved: " .. buffer.filename)
end)
```

## 現在のgmacsアーキテクチャ分析

### 統合ポイント

#### 1. KeyBinding System (`domain/keybinding.go`)
- **現在の仕組み**: `KeyBindingMap`で多段階キーシーケンス処理
- **統合方法**: `BindKeySequence()`メソッドでLua設定からキーバインド登録
- **API**: `gmacs.bind_key(sequence, command)`

#### 2. Mode System (`domain/mode.go`)
- **現在の仕組み**: `MajorMode`/`MinorMode`インターフェースベース
- **統合方法**: `RegisterMajorMode()`/`RegisterMinorMode()`でLuaモード登録
- **API**: `gmacs.major_mode()` / `gmacs.minor_mode()`

#### 3. Command System (`domain/command.go`)
- **現在の仕組み**: `CommandRegistry`で関数ベースコマンド管理
- **統合方法**: `RegisterFunc()`でLuaコマンド登録
- **API**: `gmacs.defun(name, function)`

#### 4. Editor State Management (`domain/editor.go`)
- **現在の仕組み**: 中央コーディネーターパターン
- **統合方法**: Editor起動時に設定読み込み、全サブシステムへアクセス提供

## Lua API設計

### 基本設定API
```lua
-- キーバインディング
gmacs.bind_key(sequence, command)           -- グローバルキーバインド設定
gmacs.unbind_key(sequence)                  -- キーバインド削除
gmacs.local_bind_key(mode_name, sequence, command) -- モード固有バインド

-- mode_nameは文字列で指定：
-- メジャーモード: "text-mode", "lua-mode", "fundamental-mode" など
-- マイナーモード: "auto-a-mode", "line-number-mode" など
-- 例：
-- gmacs.local_bind_key("text-mode", "C-c C-t", "text-specific-command")
-- gmacs.local_bind_key("auto-a-mode", "C-c a", "auto-a-toggle")

-- エディタ設定
gmacs.set_option(name, value)               -- オプション設定
gmacs.get_option(name)                      -- オプション取得

-- コマンド定義
gmacs.defun(name, function(editor)          -- カスタムコマンド定義
    -- Luaでカスタムコマンド実装
end)
```

### モード定義API
```lua
-- メジャーモード
gmacs.major_mode(name, {
    file_patterns = {"%.ext$"},             -- ファイルパターン
    syntax_table = syntax_def,              -- シンタックステーブル
    keymap = {                              -- モード固有キーマップ
        ["C-c C-c"] = "compile",
        ["TAB"] = custom_indent
    },
    hooks = {                               -- モードフック
        on_activate = function(buffer) end,
        on_deactivate = function(buffer) end
    }
})

-- マイナーモード  
gmacs.minor_mode(name, {
    priority = 100,                         -- 優先度
    keymap = {...},                         -- キーマップ
    predicate = function(buffer)            -- 有効条件
        return true 
    end
})
```

### バッファ・ウィンドウ操作API
```lua
-- バッファ操作
local buf = gmacs.current_buffer()
buf:insert_text("Hello")                    -- テキスト挿入
buf:goto_line(10)                          -- 行移動
buf:save()                                 -- 保存

-- ウィンドウ操作
local win = gmacs.current_window()
win:split_vertical()                       -- 縦分割
win:switch_to_buffer("filename")           -- バッファ切り替え
```

### イベントフックAPI
```lua
-- フック登録
gmacs.add_hook("before-save", function(buffer)
    -- 保存前処理
end)

gmacs.add_hook("after-change", function(buffer, start, end, text)
    -- テキスト変更後処理  
end)

-- キーボード処理フック
gmacs.add_hook("key-press", function(key)
    if key == "special-trigger" then
        -- カスタム処理
        return true -- イベント消費
    end
end)
```

## 主要コンポーネント設計

### 1. Lua VM管理 (`lua_vm.go`)
```go
type LuaVM struct {
    state *lua.LState
    sandbox *Sandbox
}

func NewLuaVM() *LuaVM
func (vm *LuaVM) LoadConfig(configPath string) error
func (vm *LuaVM) ReloadConfig() error
func (vm *LuaVM) ExecuteFunction(name string, args ...interface{}) error
```

### エディタ統合と設定読み込み

#### NewEditor の修正
```go
// 修正前（現在）
func NewEditor() *Editor

// 修正後（設定ファイル対応）
func NewEditor() *Editor                    // 設定なし（デフォルト動作）
func NewEditorWithConfig(configPath string) *Editor  // 指定された設定ファイルを読み込み

// main.go での使用例
func main() {
    // ... 既存の初期化処理 ...
    
    // 設定ファイル検索
    configPath := findConfigFile() // ~/.gmacs/init.lua を検索
    
    // エディタ作成
    var editor *domain.Editor
    if configPath != "" {
        gmacslog.Info("Loading config: %s", configPath)
        editor = domain.NewEditorWithConfig(configPath)
    } else {
        gmacslog.Info("No config file found, starting with defaults")
        editor = domain.NewEditor() // 設定なし
    }
    
    // ... 既存のメインループ ...
}

func findConfigFile() string {
    // ~/.gmacs/init.lua を検索
    // 存在すればパスを返す、なければ空文字列
}
```

// 内部実装
type EditorConfig struct {
    ConfigPath string    // 設定ファイルパス（空文字列なら読み込まない）
}

func newEditorWithOptions(config EditorConfig) *Editor
```

### 2. 設定ローダー (`config_loader.go`)
```go
type ConfigLoader struct {
    vm *LuaVM
    configPaths []string
}

func (cl *ConfigLoader) LoadUserConfig() error
func (cl *ConfigLoader) FindConfigFile() (string, error)
func (cl *ConfigLoader) WatchConfigChanges() error
```

### 3. API ブリッジ (`api_bindings.go`)
```go
func RegisterGmacsAPI(L *lua.LState, editor *Editor)
func luaBindKey(L *lua.LState) int          -- gmacs.bind_key()
func luaLocalBindKey(L *lua.LState) int     -- gmacs.local_bind_key()
func luaDefun(L *lua.LState) int            -- gmacs.defun()
func luaMajorMode(L *lua.LState) int        -- gmacs.major_mode()
func luaMinorMode(L *lua.LState) int        -- gmacs.minor_mode()
func luaAddHook(L *lua.LState) int          -- gmacs.add_hook()

// モード名からModeManager経由でモードを取得
func findModeByName(editor *Editor, modeName string) (MajorMode, MinorMode, error)
```

### 4. イベントフック (`event_hooks.go`)
```go
type HookManager struct {
    hooks map[string][]lua.LValue
}

func (hm *HookManager) AddHook(event string, function lua.LValue)
func (hm *HookManager) TriggerHook(event string, args ...interface{})
func (hm *HookManager) RemoveHook(event string, function lua.LValue)
```

## 実装計画

### フェーズ1: 基盤構築（テスト対応重視）
1. **Editor API修正**: 設定分離のためのコンストラクタ拡張
   - `NewEditor()`: 設定なし（デフォルト動作、テスト安全）
   - `NewEditorWithConfig(configPath string)`: 指定設定ファイル読み込み
   - `main.go`で設定ファイル検索と適切なコンストラクタ選択

2. **依存関係追加**: gopher-luaライブラリの統合
   - `go.mod`に`github.com/yuin/gopher-lua`追加
   - 基本的なLua実行環境構築

3. **基本構造**: `lua-config`パッケージ作成
   - ディレクトリ構造の作成
   - 基本的なインターフェース定義

4. **VM管理**: Lua仮想マシンとサンドボックス実装
   - 安全なLua実行環境
   - メモリ・実行時間制限

5. **設定読み込み**: 柔軟な設定ローダー実装
   - 設定ファイルパス指定可能
   - テスト用の設定なしモード

### フェーズ2: 基本API実装  
1. **キーバインド**: `gmacs.bind_key()`実装
   - 既存KeyBindingMapとの統合
   - キーシーケンス解析

2. **コマンド**: `gmacs.defun()`でカスタムコマンド
   - CommandRegistryとの統合
   - Lua関数→Go関数ブリッジ

3. **設定**: `gmacs.set_option()`でエディタオプション
   - 設定値の型安全性
   - デフォルト値管理

4. **統合**: main.goでの設定読み込み制御
   - 設定ファイル検索ロジック実装
   - エラー時のフォールバック（設定なしで起動）
   - **重要**: 既存のE2Eテストは変更不要（`NewEditor()`が設定なし）

### フェーズ3: 高度機能
1. **モードシステム**: Luaモード定義API
   - ModeManagerとの統合
   - 動的モード登録

2. **イベントフック**: バッファ・キー操作フック
   - イベントシステムとの統合
   - フック優先度管理

3. **バッファAPI**: Lua→Goブリッジ完全実装
   - バッファ操作の完全なAPI
   - ウィンドウ操作API

4. **エラーハンドリング**: 設定エラーの適切な処理
   - Luaエラーの表示
   - 部分的設定失敗の処理

### フェーズ4: 拡張・最適化
1. **設定再読み込み**: `M-x reload-config`コマンド
   - 動的設定リロード
   - 状態クリーンアップ

2. **デバッグ支援**: Lua設定のデバッグ機能
   - デバッグ情報出力
   - 設定値の確認コマンド

3. **パフォーマンス**: 頻繁に呼ばれるAPIの最適化
   - Lua↔Go呼び出しコスト削減
   - キャッシュ機能

4. **ドキュメント**: Lua API仕様書とサンプル
   - 完全なAPI仕様
   - 設定例とチュートリアル

## 技術的考慮事項

### テスト分離（最重要）
- **E2Eテスト**: `NewEditor()`がデフォルトで設定なし、既存テスト変更不要
- **設定テスト**: `NewEditorWithConfig("testdata/config.lua")`で専用設定使用
- **統合テスト**: 設定読み込み機能の専用テスト
- **本番動作**: `main.go`で設定ファイル検索と読み込み制御

### セキュリティ
- Luaサンドボックス環境での実行
- ファイルシステムアクセス制限
- 無限ループ・メモリリーク防止

### パフォーマンス
- Lua VM初期化コスト
- 頻繁なLua↔Go呼び出しの最適化
- 設定キャッシュ機能

### 互換性
- 既存のgmacsアーキテクチャとの統合
- 設定なしでの正常動作
- 段階的な設定移行サポート

### エラーハンドリング
- Lua文法エラーの適切な表示
- 部分的設定失敗時の継続動作
- デバッグ情報の提供

## 使用例

### 基本的な設定例
```lua
-- ~/.gmacs/init.lua

-- グローバルキーバインド
gmacs.bind_key("C-x C-s", "save-buffer")
gmacs.bind_key("C-x C-f", "find-file")
gmacs.bind_key("C-x C-c", "quit")

-- モード固有キーバインド
gmacs.local_bind_key("text-mode", "C-c C-t", "text-format")
gmacs.local_bind_key("auto-a-mode", "C-c a", "auto-a-toggle")

-- エディタ設定
gmacs.set_option("tab-width", 4)
gmacs.set_option("auto-save", true)

-- カスタムコマンド
gmacs.defun("hello-world", function(editor)
    local buf = editor:current_buffer()
    buf:insert_text("Hello, World!")
end)

gmacs.bind_key("C-x h", "hello-world")
```

### 高度な設定例
```lua
-- Lua言語モード定義
gmacs.major_mode("lua-mode", {
    file_patterns = {"%.lua$"},
    keymap = {
        ["C-c C-c"] = "lua-execute-buffer",
        ["C-c C-l"] = "lua-load-file"
    },
    hooks = {
        on_activate = function(buffer)
            gmacs.message("Lua mode activated")
        end
    }
})

-- 自動保存機能
gmacs.add_hook("after-change", function(buffer)
    if gmacs.get_option("auto-save") then
        -- 5秒後に自動保存をスケジュール
        gmacs.run_after(5000, function()
            buffer:save()
        end)
    end
end)

-- プロジェクト固有設定
if gmacs.file_exists(".gmacs-project") then
    gmacs.load_file(".gmacs-project")
end
```

## まとめ

この設計により、gmacsはEmacsの`.emacs`に匹敵する柔軟で強力な設定システムを持つことができます。既存のアーキテクチャを活かしながら、段階的な実装により安全で使いやすいLua設定環境を提供します。

ユーザーはLuaの豊富な機能を活用して、キーバインド、コマンド、モード、イベントフックなど、エディタのあらゆる側面をカスタマイズできるようになります。