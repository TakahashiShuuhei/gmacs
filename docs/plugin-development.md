# gmacs プラグイン開発ガイド

## 概要

gmacs は HashiCorp go-plugin を使用した包括的なプラグインシステムを提供しています。プラグインは独立したプロセスとして実行され、RPC 通信を通じてホストと双方向でやり取りできます。

## プラグインの基本構造

### 必要なファイル

プラグインには以下のファイルが必要です：

1. **`manifest.json`** - プラグインのメタデータ
2. **`main.go`** - プラグインのエントリーポイント
3. **`plugin.go`** - プラグインの実装
4. **`go.mod`** - Go モジュール定義

### manifest.json

```json
{
  "name": "example-plugin",
  "version": "1.0.0",
  "description": "Example plugin for gmacs",
  "binary": "example-plugin",
  "dependencies": []
}
```

## プラグインの実装

### 基本インターフェース

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(ctx context.Context, host HostInterface) error
    Cleanup() error
    
    // コマンド関連
    GetCommands() []CommandSpec
    
    // モード関連
    GetMajorModes() []MajorModeSpec
    GetMinorModes() []MinorModeSpec
    
    // キーバインド関連
    GetKeyBindings() []KeyBindingSpec
}
```

### コマンドプラグインインターフェース

```go
type CommandPlugin interface {
    ExecuteCommand(name string, args ...interface{}) error
    GetCompletions(command string, prefix string) []string
}
```

## HostInterface API

プラグインから gmacs ホストにアクセスするための API です。

### メッセージ表示

```go
// ミニバッファにメッセージを表示
host.ShowMessage("Hello from plugin!")

// ステータスラインにメッセージを設定
host.SetStatus("Plugin status")
```

### バッファ操作

```go
// 現在のバッファを取得
buffer := host.GetCurrentBuffer()
if buffer != nil {
    name := buffer.Name()
    content := buffer.Content()
    pos := buffer.CursorPosition()
    
    // テキスト挿入
    buffer.InsertAt(pos, "inserted text")
    
    // カーソル移動
    buffer.SetCursorPosition(pos + 10)
    
    // ダーティフラグ設定
    buffer.MarkDirty()
}

// バッファ管理
newBuffer := host.CreateBuffer("new-buffer")
existingBuffer := host.FindBuffer("buffer-name")
host.SwitchToBuffer("buffer-name")
```

### ウィンドウ操作

```go
// 現在のウィンドウを取得
window := host.GetCurrentWindow()
if window != nil {
    width := window.Width()
    height := window.Height()
    
    // スクロール操作
    offset := window.ScrollOffset()
    window.SetScrollOffset(offset + 5)
    
    // ウィンドウのバッファを取得
    buffer := window.Buffer()
}
```

### ファイル操作

```go
// ファイルを開く
err := host.OpenFile("/path/to/file.txt")

// バッファを保存
err := host.SaveBuffer("buffer-name")
```

### コマンド実行

```go
// gmacs コマンドを実行
err := host.ExecuteCommand("forward-char")
```

### オプション管理

```go
// オプション取得
value, err := host.GetOption("tab-width")

// オプション設定
err := host.SetOption("my-plugin-option", "value")
```

### モード管理

```go
// メジャーモード設定
err := host.SetMajorMode("buffer-name", "text-mode")

// マイナーモード切り替え
err := host.ToggleMinorMode("buffer-name", "line-number-mode")
```

### フック システム

```go
// フック登録
hookHandler := func(args ...interface{}) error {
    // フック処理
    return nil
}
host.AddHook("before-save-hook", hookHandler)

// フック実行
host.TriggerHook("my-custom-hook", "arg1", "arg2")
```

## コマンドの実装

### CommandSpec 定義

```go
func (p *MyPlugin) GetCommands() []CommandSpec {
    return []CommandSpec{
        {
            Name:        "my-command",
            Description: "My custom command",
            Interactive: true,
            Handler:     "HandleMyCommand",
        },
    }
}
```

### コマンドハンドラ実装

```go
func (p *MyPlugin) HandleMyCommand() error {
    // コマンドの処理
    p.host.ShowMessage("Command executed!")
    
    // ミニバッファにメッセージを表示したい場合は PLUGIN_MESSAGE: プレフィックスを返す
    return NewStringError("PLUGIN_MESSAGE:Command completed successfully")
}

// ExecuteCommand での実装
func (p *MyPlugin) ExecuteCommand(name string, args ...interface{}) error {
    switch name {
    case "my-command":
        return p.HandleMyCommand()
    default:
        return fmt.Errorf("unknown command: %s", name)
    }
}
```

## メッセージ表示システム

プラグインからユーザーにメッセージを表示する方法は2つあります：

### 1. HostInterface.ShowMessage()

```go
// ホストのメッセージシステムを直接使用
p.host.ShowMessage("[MY-PLUGIN] Operation completed")
```

### 2. PLUGIN_MESSAGE: プレフィックス

```go
// コマンドの戻り値としてメッセージを返す
return NewStringError("PLUGIN_MESSAGE:[MY-PLUGIN] Operation completed")
```

2番目の方法はミニバッファに表示され、1番目の方法はログ出力とホストシステムでの処理に使用されます。

## モードの実装

### メジャーモード

```go
func (p *MyPlugin) GetMajorModes() []MajorModeSpec {
    return []MajorModeSpec{
        {
            Name:        "my-mode",
            Extensions:  []string{".myext"},
            Description: "My custom major mode",
            KeyBindings: []KeyBindingSpec{
                {
                    Sequence: "C-c C-e",
                    Command:  "my-command",
                    Mode:     "my-mode",
                },
            },
        },
    }
}
```

### マイナーモード

```go
func (p *MyPlugin) GetMinorModes() []MinorModeSpec {
    return []MinorModeSpec{
        {
            Name:        "my-minor-mode",
            Description: "My custom minor mode",
            Global:      false,
            KeyBindings: []KeyBindingSpec{
                {
                    Sequence: "C-c m",
                    Command:  "my-minor-command",
                    Mode:     "my-minor-mode",
                },
            },
        },
    }
}
```

## キーバインドの実装

### グローバルキーバインド

```go
func (p *MyPlugin) GetKeyBindings() []KeyBindingSpec {
    return []KeyBindingSpec{
        {
            Sequence: "C-c C-x p",
            Command:  "my-plugin-command",
            Mode:     "", // 空文字列はグローバル
        },
    }
}
```

## プラグインのビルドとインストール

### ビルド

```bash
go build -o my-plugin
```

### インストール

```bash
# ローカルディレクトリから
gmacs plugin install /path/to/plugin/directory/

# GitHubリポジトリから
gmacs plugin install github.com/user/my-gmacs-plugin
```

### アンインストール

```bash
gmacs plugin uninstall my-plugin
```

## デバッグとログ

### プラグインでのログ出力

```go
import (
    "log"
    "os"
)

func init() {
    log.SetOutput(os.Stderr)
    log.SetPrefix("[MY-PLUGIN] ")
}

func (p *MyPlugin) HandleMyCommand() error {
    log.Printf("Executing my command...")
    // 処理
    return nil
}
```

### gmacs ログの確認

```bash
# ログファイルの場所
ls logs/gmacs_*.log

# 最新ログの確認
tail -f logs/gmacs_$(date +%Y%m%d)_*.log
```

## 完全なプラグイン例

参考実装として `gmacs-example-plugin` を確認してください：

- **リポジトリ**: https://github.com/TakahashiShuuhei/gmacs-example-plugin
- **機能**: HostInterface の全 API を網羅したテスト用プラグイン
- **コマンド**: 
  - `example-greet` - 基本的な挨拶メッセージ
  - `example-test-host-api` - HostInterface API のテスト
  - `example-buffer-ops` - バッファ操作のテスト
  - `example-file-ops` - ファイル操作のテスト
  - など

## エラーハンドリング

### StringError の実装

```go
type StringError struct {
    Message string
}

func (se StringError) Error() string {
    return se.Message
}

func NewStringError(message string) error {
    return StringError{Message: message}
}

// gob 登録（RPC シリアライゼーション用）
func init() {
    gob.Register(StringError{})
}
```

## トラブルシューティング

### よくある問題

1. **RPC EOF エラー**
   - プラグインプロセスが予期せず終了
   - gob シリアライゼーションエラー
   
2. **コマンドが見つからない**
   - GetCommands() で正しく登録されているか確認
   - ExecuteCommand() で適切にルーティングされているか確認

3. **メッセージが表示されない**
   - `PLUGIN_MESSAGE:` プレフィックスを使用しているか確認
   - ShowMessage() の呼び出しが正しいか確認

### デバッグ手順

1. ログ出力の確認
2. プラグインの手動実行テスト
3. RPC 通信の確認
4. E2E テストでの動作確認

## 今後の拡張予定

- BufferInterface の完全な RPC 実装
- WindowInterface の完全な RPC 実装  
- ファイルシステム操作の拡張
- 設定管理システムの統合
- プラグイン間通信システム

## 参考リソース

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [gmacs プラグイン SDK](https://github.com/TakahashiShuuhei/gmacs-plugin-sdk)
- [gmacs example plugin](https://github.com/TakahashiShuuhei/gmacs-example-plugin)