# gmacs プラグイン実装例

このドキュメントでは、gmacs プラグインの実装例を示します。実際に動作する example-plugin を参考に、様々なユースケースでの実装方法を説明します。

## 基本的なプラグイン構造

### ディレクトリ構成

```
my-plugin/
├── manifest.json
├── main.go
├── plugin.go
├── go.mod
└── go.sum
```

### manifest.json

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom gmacs plugin",
  "binary": "my-plugin",
  "dependencies": []
}
```

### go.mod

```go
module github.com/user/my-gmacs-plugin

go 1.22

require (
    github.com/TakahashiShuuhei/gmacs-plugin-sdk v0.1.0
    github.com/hashicorp/go-plugin v1.4.10
)
```

## 例1: シンプルな挨拶プラグイン

### plugin.go

```go
package main

import (
    "context"
    "encoding/gob"
    "fmt"
    "time"
    
    pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type StringError struct {
    Message string
}

func (se StringError) Error() string {
    return se.Message
}

func NewStringError(message string) error {
    return StringError{Message: message}
}

func init() {
    gob.Register(StringError{})
}

type GreetPlugin struct {
    host pluginsdk.HostInterface
}

func (p *GreetPlugin) Name() string {
    return "greet-plugin"
}

func (p *GreetPlugin) Version() string {
    return "1.0.0"
}

func (p *GreetPlugin) Description() string {
    return "Simple greeting plugin"
}

func (p *GreetPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
    p.host = host
    return nil
}

func (p *GreetPlugin) Cleanup() error {
    return nil
}

func (p *GreetPlugin) GetCommands() []pluginsdk.CommandSpec {
    return []pluginsdk.CommandSpec{
        {
            Name:        "greet",
            Description: "Display a greeting message",
            Interactive: true,
            Handler:     "HandleGreet",
        },
        {
            Name:        "greet-time",
            Description: "Display greeting with current time",
            Interactive: true,
            Handler:     "HandleGreetTime",
        },
    }
}

func (p *GreetPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
    return []pluginsdk.MajorModeSpec{}
}

func (p *GreetPlugin) GetMinorModes() []pluginsdk.MinorModeSpec {
    return []pluginsdk.MinorModeSpec{}
}

func (p *GreetPlugin) GetKeyBindings() []pluginsdk.KeyBindingSpec {
    return []pluginsdk.KeyBindingSpec{
        {
            Sequence: "C-c g",
            Command:  "greet",
            Mode:     "",
        },
    }
}

// コマンドハンドラ
func (p *GreetPlugin) HandleGreet() error {
    message := "Hello from greet plugin!"
    p.host.ShowMessage(message)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *GreetPlugin) HandleGreetTime() error {
    now := time.Now().Format("2006-01-02 15:04:05")
    message := fmt.Sprintf("Hello! Current time: %s", now)
    p.host.ShowMessage(message)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

// CommandPlugin インターフェース実装
func (p *GreetPlugin) ExecuteCommand(name string, args ...interface{}) error {
    switch name {
    case "greet":
        return p.HandleGreet()
    case "greet-time":
        return p.HandleGreetTime()
    default:
        return fmt.Errorf("unknown command: %s", name)
    }
}

func (p *GreetPlugin) GetCompletions(command string, prefix string) []string {
    return []string{}
}

var pluginInstance = &GreetPlugin{}
```

## 例2: テキスト処理プラグイン

```go
package main

import (
    "context"
    "strings"
    "unicode"
    
    pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type TextPlugin struct {
    host pluginsdk.HostInterface
}

func (p *TextPlugin) Name() string {
    return "text-processor"
}

func (p *TextPlugin) Version() string {
    return "1.0.0"
}

func (p *TextPlugin) Description() string {
    return "Text processing utilities"
}

func (p *TextPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
    p.host = host
    return nil
}

func (p *TextPlugin) GetCommands() []pluginsdk.CommandSpec {
    return []pluginsdk.CommandSpec{
        {
            Name:        "count-words",
            Description: "Count words in current buffer",
            Interactive: true,
            Handler:     "HandleCountWords",
        },
        {
            Name:        "uppercase-region",
            Description: "Convert text to uppercase",
            Interactive: true,
            Handler:     "HandleUppercase",
        },
        {
            Name:        "insert-separator",
            Description: "Insert a separator line",
            Interactive: true,
            Handler:     "HandleInsertSeparator",
        },
    }
}

func (p *TextPlugin) HandleCountWords() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    content := buffer.Content()
    
    // 単語数をカウント
    words := strings.Fields(content)
    wordCount := len(words)
    
    // 行数をカウント
    lines := strings.Split(content, "\n")
    lineCount := len(lines)
    
    // 文字数をカウント
    charCount := len(content)
    
    // 結果を表示
    result := fmt.Sprintf("[TEXT] %s: %d lines, %d words, %d characters", 
        buffer.Name(), lineCount, wordCount, charCount)
    
    p.host.ShowMessage(result)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", result))
}

func (p *TextPlugin) HandleUppercase() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    // 現在の内容を取得
    content := buffer.Content()
    
    // 大文字に変換
    upperContent := strings.ToUpper(content)
    
    // バッファの内容を更新
    buffer.SetContent(upperContent)
    buffer.MarkDirty()
    
    message := "Text converted to uppercase"
    p.host.ShowMessage(message)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *TextPlugin) HandleInsertSeparator() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    // カーソル位置を取得
    pos := buffer.CursorPosition()
    
    // セパレータ行を挿入
    separator := "\n" + strings.Repeat("-", 50) + "\n"
    buffer.InsertAt(pos, separator)
    
    // カーソルをセパレータの後に移動
    buffer.SetCursorPosition(pos + len(separator))
    buffer.MarkDirty()
    
    message := "Separator line inserted"
    p.host.ShowMessage(message)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *TextPlugin) ExecuteCommand(name string, args ...interface{}) error {
    switch name {
    case "count-words":
        return p.HandleCountWords()
    case "uppercase-region":
        return p.HandleUppercase()
    case "insert-separator":
        return p.HandleInsertSeparator()
    default:
        return fmt.Errorf("unknown command: %s", name)
    }
}

var pluginInstance = &TextPlugin{}
```

## 例3: ファイル管理プラグイン

```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type FileManagerPlugin struct {
    host pluginsdk.HostInterface
}

func (p *FileManagerPlugin) Name() string {
    return "file-manager"
}

func (p *FileManagerPlugin) Version() string {
    return "1.0.0"
}

func (p *FileManagerPlugin) Description() string {
    return "File management utilities"
}

func (p *FileManagerPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
    p.host = host
    return nil
}

func (p *FileManagerPlugin) GetCommands() []pluginsdk.CommandSpec {
    return []pluginsdk.CommandSpec{
        {
            Name:        "list-directory",
            Description: "List files in current directory",
            Interactive: true,
            Handler:     "HandleListDirectory",
        },
        {
            Name:        "create-backup",
            Description: "Create backup of current file",
            Interactive: true,
            Handler:     "HandleCreateBackup",
        },
        {
            Name:        "open-config-dir",
            Description: "Open configuration directory",
            Interactive: true,
            Handler:     "HandleOpenConfigDir",
        },
    }
}

func (p *FileManagerPlugin) HandleListDirectory() error {
    // 現在のバッファのディレクトリを取得
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    filename := buffer.Filename()
    var dir string
    
    if filename != "" {
        dir = filepath.Dir(filename)
    } else {
        // ファイルバッファでない場合は現在のディレクトリ
        var err error
        dir, err = os.Getwd()
        if err != nil {
            return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
        }
    }
    
    // ディレクトリ内容を取得
    entries, err := os.ReadDir(dir)
    if err != nil {
        return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
    }
    
    // 結果バッファを作成
    listBuffer := p.host.CreateBuffer("*Directory Listing*")
    
    var content strings.Builder
    content.WriteString(fmt.Sprintf("Directory: %s\n", dir))
    content.WriteString(strings.Repeat("=", 50) + "\n\n")
    
    for _, entry := range entries {
        if entry.IsDir() {
            content.WriteString(fmt.Sprintf("📁 %s/\n", entry.Name()))
        } else {
            content.WriteString(fmt.Sprintf("📄 %s\n", entry.Name()))
        }
    }
    
    listBuffer.SetContent(content.String())
    
    // バッファに切り替え
    p.host.SwitchToBuffer("*Directory Listing*")
    
    message := fmt.Sprintf("Listed %d items in %s", len(entries), dir)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *FileManagerPlugin) HandleCreateBackup() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    filename := buffer.Filename()
    if filename == "" {
        return NewStringError("PLUGIN_MESSAGE:ERROR: Buffer is not associated with a file")
    }
    
    // バックアップファイル名を生成
    backupName := filename + ".backup"
    
    // 現在のバッファを保存
    err := p.host.SaveBuffer(buffer.Name())
    if err != nil {
        return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: Failed to save: %v", err))
    }
    
    // ファイルをコピー
    content := buffer.Content()
    err = os.WriteFile(backupName, []byte(content), 0644)
    if err != nil {
        return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: Failed to create backup: %v", err))
    }
    
    message := fmt.Sprintf("Backup created: %s", backupName)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *FileManagerPlugin) HandleOpenConfigDir() error {
    // ホームディレクトリの .config/gmacs を開く
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
    }
    
    configDir := filepath.Join(homeDir, ".config", "gmacs")
    
    // ディレクトリが存在しない場合は作成
    if _, err := os.Stat(configDir); os.IsNotExist(err) {
        err = os.MkdirAll(configDir, 0755)
        if err != nil {
            return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
        }
    }
    
    // 設定ファイルのパスを開く
    configFile := filepath.Join(configDir, "config.lua")
    err = p.host.OpenFile(configFile)
    if err != nil {
        // ファイルが存在しない場合は新規作成
        err = os.WriteFile(configFile, []byte("-- gmacs configuration\n"), 0644)
        if err != nil {
            return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
        }
        err = p.host.OpenFile(configFile)
        if err != nil {
            return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:ERROR: %v", err))
        }
    }
    
    message := fmt.Sprintf("Opened config file: %s", configFile)
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *FileManagerPlugin) ExecuteCommand(name string, args ...interface{}) error {
    switch name {
    case "list-directory":
        return p.HandleListDirectory()
    case "create-backup":
        return p.HandleCreateBackup()
    case "open-config-dir":
        return p.HandleOpenConfigDir()
    default:
        return fmt.Errorf("unknown command: %s", name)
    }
}

var pluginInstance = &FileManagerPlugin{}
```

## 例4: 開発者向けユーティリティプラグイン

```go
package main

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"
    
    pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type DevUtilsPlugin struct {
    host pluginsdk.HostInterface
}

func (p *DevUtilsPlugin) Name() string {
    return "dev-utils"
}

func (p *DevUtilsPlugin) Version() string {
    return "1.0.0"
}

func (p *DevUtilsPlugin) Description() string {
    return "Developer utilities for code editing"
}

func (p *DevUtilsPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
    p.host = host
    return nil
}

func (p *DevUtilsPlugin) GetCommands() []pluginsdk.CommandSpec {
    return []pluginsdk.CommandSpec{
        {
            Name:        "insert-timestamp",
            Description: "Insert current timestamp",
            Interactive: true,
            Handler:     "HandleInsertTimestamp",
        },
        {
            Name:        "comment-toggle",
            Description: "Toggle line comments",
            Interactive: true,
            Handler:     "HandleCommentToggle",
        },
        {
            Name:        "find-todos",
            Description: "Find TODO comments in buffer",
            Interactive: true,
            Handler:     "HandleFindTodos",
        },
        {
            Name:        "generate-uuid",
            Description: "Generate and insert UUID",
            Interactive: true,
            Handler:     "HandleGenerateUUID",
        },
    }
}

// メジャーモード定義（Go ファイル用の拡張）
func (p *DevUtilsPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
    return []pluginsdk.MajorModeSpec{
        {
            Name:        "enhanced-go-mode",
            Extensions:  []string{".go"},
            Description: "Enhanced Go mode with dev utilities",
            KeyBindings: []pluginsdk.KeyBindingSpec{
                {
                    Sequence: "C-c C-t",
                    Command:  "insert-timestamp",
                    Mode:     "enhanced-go-mode",
                },
                {
                    Sequence: "C-c /",
                    Command:  "comment-toggle",
                    Mode:     "enhanced-go-mode",
                },
            },
        },
    }
}

func (p *DevUtilsPlugin) HandleInsertTimestamp() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    // 現在時刻を生成
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    
    // カーソル位置に挿入
    pos := buffer.CursorPosition()
    buffer.InsertAt(pos, timestamp)
    buffer.SetCursorPosition(pos + len(timestamp))
    buffer.MarkDirty()
    
    message := "Timestamp inserted"
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *DevUtilsPlugin) HandleCommentToggle() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    content := buffer.Content()
    lines := strings.Split(content, "\n")
    pos := buffer.CursorPosition()
    
    // カーソル位置から行番号を計算
    currentLine := 0
    charCount := 0
    for i, line := range lines {
        if charCount + len(line) >= pos {
            currentLine = i
            break
        }
        charCount += len(line) + 1 // +1 for newline
    }
    
    if currentLine >= len(lines) {
        return NewStringError("PLUGIN_MESSAGE:ERROR: Invalid cursor position")
    }
    
    // ファイル拡張子に基づいてコメント文字を決定
    filename := buffer.Filename()
    commentPrefix := "//"  // デフォルト
    
    if strings.HasSuffix(filename, ".py") {
        commentPrefix = "#"
    } else if strings.HasSuffix(filename, ".sh") {
        commentPrefix = "#"
    } else if strings.HasSuffix(filename, ".lua") {
        commentPrefix = "--"
    }
    
    // 行のコメント状態を切り替え
    line := lines[currentLine]
    trimmed := strings.TrimSpace(line)
    
    if strings.HasPrefix(trimmed, commentPrefix) {
        // コメントを削除
        uncommented := strings.TrimPrefix(trimmed, commentPrefix)
        uncommented = strings.TrimSpace(uncommented)
        lines[currentLine] = strings.Repeat(" ", len(line)-len(trimmed)) + uncommented
    } else {
        // コメントを追加
        indent := len(line) - len(strings.TrimLeft(line, " \t"))
        lines[currentLine] = line[:indent] + commentPrefix + " " + line[indent:]
    }
    
    // バッファを更新
    newContent := strings.Join(lines, "\n")
    buffer.SetContent(newContent)
    buffer.MarkDirty()
    
    message := "Comment toggled"
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *DevUtilsPlugin) HandleFindTodos() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    content := buffer.Content()
    lines := strings.Split(content, "\n")
    
    // TODO パターンを検索
    todoPattern := regexp.MustCompile(`(?i)(TODO|FIXME|XXX|HACK|NOTE):?\s*(.*)`)
    
    var todos []string
    for i, line := range lines {
        matches := todoPattern.FindStringSubmatch(line)
        if len(matches) > 0 {
            todoItem := fmt.Sprintf("Line %d: %s %s", i+1, matches[1], matches[2])
            todos = append(todos, todoItem)
        }
    }
    
    if len(todos) == 0 {
        return NewStringError("PLUGIN_MESSAGE:No TODO items found")
    }
    
    // 結果バッファを作成
    todoBuffer := p.host.CreateBuffer("*TODO List*")
    
    var content_builder strings.Builder
    content_builder.WriteString(fmt.Sprintf("TODO items in %s\n", buffer.Name()))
    content_builder.WriteString(strings.Repeat("=", 40) + "\n\n")
    
    for _, todo := range todos {
        content_builder.WriteString(todo + "\n")
    }
    
    todoBuffer.SetContent(content_builder.String())
    p.host.SwitchToBuffer("*TODO List*")
    
    message := fmt.Sprintf("Found %d TODO items", len(todos))
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *DevUtilsPlugin) HandleGenerateUUID() error {
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:ERROR: No active buffer")
    }
    
    // 簡易UUID生成（実際の実装では uuid ライブラリを使用推奨）
    uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
        time.Now().UnixNano()&0xffffffff,
        time.Now().UnixNano()>>32&0xffff,
        time.Now().UnixNano()>>48&0xffff,
        time.Now().UnixNano()>>16&0xffff,
        time.Now().UnixNano()&0xffffffffffff)
    
    // カーソル位置に挿入
    pos := buffer.CursorPosition()
    buffer.InsertAt(pos, uuid)
    buffer.SetCursorPosition(pos + len(uuid))
    buffer.MarkDirty()
    
    message := "UUID generated and inserted"
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", message))
}

func (p *DevUtilsPlugin) ExecuteCommand(name string, args ...interface{}) error {
    switch name {
    case "insert-timestamp":
        return p.HandleInsertTimestamp()
    case "comment-toggle":
        return p.HandleCommentToggle()
    case "find-todos":
        return p.HandleFindTodos()
    case "generate-uuid":
        return p.HandleGenerateUUID()
    default:
        return fmt.Errorf("unknown command: %s", name)
    }
}

var pluginInstance = &DevUtilsPlugin{}
```

## main.go の実装

すべての例で共通して使用する main.go：

```go
package main

import (
    "context"
    "log"
    "net/rpc"
    "os"
    
    "github.com/hashicorp/go-plugin"
    pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

// RPC サーバー実装
type RPCServer struct {
    Impl pluginsdk.Plugin
}

func (s *RPCServer) ExecuteCommand(args map[string]interface{}, resp *error) error {
    name, _ := args["name"].(string)
    argsSlice, _ := args["args"].([]interface{})
    
    // プラグインの初期化確認
    if s.Impl == nil {
        *resp = NewStringError("Plugin not initialized")
        return nil
    }
    
    // 一時的にホストインターフェースを設定
    hostImpl := &SimpleHostInterface{}
    s.Impl.Initialize(context.Background(), hostImpl)
    
    // コマンド実行
    if cmdPlugin, ok := s.Impl.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
        *resp = cmdPlugin.ExecuteCommand(name, argsSlice...)
    } else {
        *resp = NewStringError("Plugin does not support command execution")
    }
    
    return nil
}

// その他の RPC メソッドは省略（基本的にすべて同じパターン）

// SimpleHostInterface の実装
type SimpleHostInterface struct{}

func (h *SimpleHostInterface) ShowMessage(message string) {
    log.Printf("[GMACS_HOST_MESSAGE]%s", message)
}

// その他のメソッドは nil 実装

// プラグインサーバー設定
type CustomRPCPlugin struct{}

func (p *CustomRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
    return &RPCServer{Impl: pluginInstance}, nil
}

func (p *CustomRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
    return nil, nil
}

func main() {
    log.SetOutput(os.Stderr)
    log.SetPrefix("[PLUGIN] ")
    
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.HandshakeConfig{
            ProtocolVersion:  1,
            MagicCookieKey:   "GMACS_PLUGIN",
            MagicCookieValue: "gmacs-plugin-magic-cookie",
        },
        Plugins: map[string]plugin.Plugin{
            "gmacs-plugin": &CustomRPCPlugin{},
        },
    })
}
```

## ビルドとテスト

```bash
# ビルド
go build -o my-plugin

# インストール
gmacs plugin install /path/to/plugin/

# テスト
M-x my-command
```

これらの例を参考に、独自の gmacs プラグインを開発してください。各例は特定のユースケースに特化していますが、組み合わせることでより複雑で強力なプラグインを作成できます。