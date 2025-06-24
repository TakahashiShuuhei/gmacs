# gmacs プラグイン API リファレンス

## HostInterface API

プラグインから gmacs ホストにアクセスするための完全な API リファレンスです。

## インターフェース定義

```go
type HostInterface interface {
    // メッセージ表示
    ShowMessage(message string)
    SetStatus(message string)
    
    // バッファ管理
    GetCurrentBuffer() BufferInterface
    CreateBuffer(name string) BufferInterface
    FindBuffer(name string) BufferInterface
    SwitchToBuffer(name string) error
    
    // ウィンドウ管理
    GetCurrentWindow() WindowInterface
    
    // ファイル操作
    OpenFile(path string) error
    SaveBuffer(bufferName string) error
    
    // コマンド実行
    ExecuteCommand(name string, args ...interface{}) error
    
    // オプション管理
    GetOption(name string) (interface{}, error)
    SetOption(name string, value interface{}) error
    
    // モード管理
    SetMajorMode(bufferName, modeName string) error
    ToggleMinorMode(bufferName, modeName string) error
    
    // フックシステム
    AddHook(event string, handler func(...interface{}) error)
    TriggerHook(event string, args ...interface{})
}
```

## BufferInterface API

```go
type BufferInterface interface {
    // バッファ情報
    Name() string
    Filename() string
    Content() string
    IsDirty() bool
    
    // カーソル操作
    CursorPosition() int
    SetCursorPosition(pos int)
    
    // 編集操作
    InsertAt(pos int, text string)
    DeleteRange(start, end int)
    SetContent(content string)
    
    // フラグ管理
    MarkDirty()
    MarkClean()
}
```

## WindowInterface API

```go
type WindowInterface interface {
    // ウィンドウ情報
    Width() int
    Height() int
    Buffer() BufferInterface
    
    // スクロール操作
    ScrollOffset() int
    SetScrollOffset(offset int)
}
```

---

## メッセージ表示 API

### ShowMessage

ミニバッファまたはメッセージエリアにメッセージを表示します。

```go
host.ShowMessage(message string)
```

**パラメータ:**
- `message`: 表示するメッセージ文字列

**使用例:**
```go
host.ShowMessage("ファイルを保存しました")
host.ShowMessage("[MY-PLUGIN] 処理が完了しました")
```

### SetStatus

ステータスラインにメッセージを設定します。

```go
host.SetStatus(message string)
```

**パラメータ:**
- `message`: ステータスメッセージ

**使用例:**
```go
host.SetStatus("検索中...")
host.SetStatus("準備完了")
```

---

## バッファ管理 API

### GetCurrentBuffer

現在アクティブなバッファを取得します。

```go
buffer := host.GetCurrentBuffer()
```

**戻り値:**
- `BufferInterface`: 現在のバッファ（nil の場合あり）

**使用例:**
```go
buffer := host.GetCurrentBuffer()
if buffer != nil {
    name := buffer.Name()
    host.ShowMessage(fmt.Sprintf("現在のバッファ: %s", name))
}
```

### CreateBuffer

新しいバッファを作成します。

```go
buffer := host.CreateBuffer(name string)
```

**パラメータ:**
- `name`: バッファ名

**戻り値:**
- `BufferInterface`: 作成されたバッファ

**使用例:**
```go
buffer := host.CreateBuffer("*my-plugin-output*")
buffer.SetContent("プラグインの出力結果")
```

### FindBuffer

指定した名前のバッファを検索します。

```go
buffer := host.FindBuffer(name string)
```

**パラメータ:**
- `name`: 検索するバッファ名

**戻り値:**
- `BufferInterface`: 見つかったバッファ（nil の場合あり）

**使用例:**
```go
buffer := host.FindBuffer("*scratch*")
if buffer != nil {
    host.SwitchToBuffer("*scratch*")
}
```

### SwitchToBuffer

指定したバッファに切り替えます。

```go
err := host.SwitchToBuffer(name string)
```

**パラメータ:**
- `name`: 切り替え先のバッファ名

**戻り値:**
- `error`: エラー（バッファが存在しない場合など）

**使用例:**
```go
err := host.SwitchToBuffer("config.txt")
if err != nil {
    host.ShowMessage(fmt.Sprintf("バッファの切り替えに失敗: %v", err))
}
```

---

## バッファ操作 API

### バッファ情報取得

```go
// バッファ名
name := buffer.Name()

// ファイル名（ファイルバッファの場合）
filename := buffer.Filename()

// バッファの全内容
content := buffer.Content()

// 変更フラグ
isDirty := buffer.IsDirty()
```

### カーソル操作

```go
// 現在のカーソル位置（文字単位）
pos := buffer.CursorPosition()

// カーソル位置設定
buffer.SetCursorPosition(100)
```

### テキスト編集

```go
// 指定位置にテキスト挿入
buffer.InsertAt(pos, "挿入するテキスト")

// 範囲削除
buffer.DeleteRange(start, end)

// バッファ内容全体を設定
buffer.SetContent("新しい内容")
```

### フラグ管理

```go
// ダーティフラグ設定（変更ありマーク）
buffer.MarkDirty()

// クリーンフラグ設定（変更なしマーク）
buffer.MarkClean()
```

---

## ウィンドウ管理 API

### GetCurrentWindow

現在アクティブなウィンドウを取得します。

```go
window := host.GetCurrentWindow()
```

**戻り値:**
- `WindowInterface`: 現在のウィンドウ（nil の場合あり）

### ウィンドウ情報取得

```go
// ウィンドウサイズ
width := window.Width()
height := window.Height()

// ウィンドウのバッファ
buffer := window.Buffer()
```

### スクロール操作

```go
// 現在のスクロールオフセット
offset := window.ScrollOffset()

// スクロール位置設定
window.SetScrollOffset(offset + 5)
```

---

## ファイル操作 API

### OpenFile

ファイルを開いてバッファに読み込みます。

```go
err := host.OpenFile(path string)
```

**パラメータ:**
- `path`: ファイルパス

**戻り値:**
- `error`: エラー（ファイルが存在しない場合など）

**使用例:**
```go
err := host.OpenFile("/etc/hosts")
if err != nil {
    host.ShowMessage(fmt.Sprintf("ファイルオープンエラー: %v", err))
} else {
    host.ShowMessage("ファイルを開きました")
}
```

### SaveBuffer

指定したバッファを保存します。

```go
err := host.SaveBuffer(bufferName string)
```

**パラメータ:**
- `bufferName`: 保存するバッファ名

**戻り値:**
- `error`: エラー（書き込み権限がない場合など）

**使用例:**
```go
err := host.SaveBuffer("config.txt")
if err != nil {
    host.ShowMessage(fmt.Sprintf("保存エラー: %v", err))
} else {
    host.ShowMessage("ファイルを保存しました")
}
```

---

## コマンド実行 API

### ExecuteCommand

gmacs の組み込みコマンドを実行します。

```go
err := host.ExecuteCommand(name string, args ...interface{})
```

**パラメータ:**
- `name`: コマンド名
- `args`: コマンド引数（可変長）

**戻り値:**
- `error`: 実行エラー

**使用例:**
```go
// カーソル移動
err := host.ExecuteCommand("forward-char")

// 行移動
err := host.ExecuteCommand("next-line")

// ファイル検索（引数付き）
err := host.ExecuteCommand("find-file", "/path/to/file")
```

**利用可能なコマンド例:**
- `forward-char`, `backward-char`
- `next-line`, `previous-line`
- `beginning-of-line`, `end-of-line`
- `find-file`, `save-buffer`
- `switch-to-buffer`

---

## オプション管理 API

### GetOption

設定オプションの値を取得します。

```go
value, err := host.GetOption(name string)
```

**パラメータ:**
- `name`: オプション名

**戻り値:**
- `interface{}`: オプション値
- `error`: エラー（オプションが存在しない場合）

**使用例:**
```go
value, err := host.GetOption("tab-width")
if err == nil {
    host.ShowMessage(fmt.Sprintf("タブ幅: %v", value))
}
```

### SetOption

設定オプションの値を設定します。

```go
err := host.SetOption(name string, value interface{})
```

**パラメータ:**
- `name`: オプション名
- `value`: 設定する値

**戻り値:**
- `error`: 設定エラー

**使用例:**
```go
// 数値設定
err := host.SetOption("tab-width", 4)

// 文字列設定
err := host.SetOption("my-plugin-config", "enabled")

// ブール設定
err := host.SetOption("auto-save", true)
```

---

## モード管理 API

### SetMajorMode

バッファのメジャーモードを設定します。

```go
err := host.SetMajorMode(bufferName, modeName string)
```

**パラメータ:**
- `bufferName`: 対象バッファ名
- `modeName`: モード名

**戻り値:**
- `error`: 設定エラー

**使用例:**
```go
err := host.SetMajorMode("config.txt", "text-mode")
```

### ToggleMinorMode

バッファのマイナーモードを切り替えます。

```go
err := host.ToggleMinorMode(bufferName, modeName string)
```

**パラメータ:**
- `bufferName`: 対象バッファ名
- `modeName`: モード名

**戻り値:**
- `error`: 切り替えエラー

**使用例:**
```go
err := host.ToggleMinorMode("source.go", "line-number-mode")
```

---

## フックシステム API

### AddHook

イベントフックを登録します。

```go
host.AddHook(event string, handler func(...interface{}) error)
```

**パラメータ:**
- `event`: イベント名
- `handler`: フック関数

**使用例:**
```go
hookHandler := func(args ...interface{}) error {
    host.ShowMessage("フックが実行されました")
    return nil
}

host.AddHook("before-save-hook", hookHandler)
```

### TriggerHook

フックイベントを発火します。

```go
host.TriggerHook(event string, args ...interface{})
```

**パラメータ:**
- `event`: イベント名
- `args`: フック関数に渡す引数（可変長）

**使用例:**
```go
host.TriggerHook("my-custom-event", "arg1", 42, true)
```

---

## エラーハンドリング

### StringError 型

プラグインで使用する標準エラー型です。

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

// 初期化関数で gob 登録
func init() {
    gob.Register(StringError{})
}
```

### メッセージ表示用エラー

ミニバッファにメッセージを表示したい場合は `PLUGIN_MESSAGE:` プレフィックスを使用します。

```go
// 成功メッセージ
return NewStringError("PLUGIN_MESSAGE:処理が完了しました")

// エラーメッセージ
return NewStringError("PLUGIN_MESSAGE:ERROR: 処理に失敗しました")

// カテゴリ付きメッセージ  
return NewStringError("PLUGIN_MESSAGE:[MY-PLUGIN] 操作が完了しました")
```

---

## 使用例：完全なコマンド実装

```go
func (p *MyPlugin) HandleFileStats() error {
    // 現在のバッファを取得
    buffer := p.host.GetCurrentBuffer()
    if buffer == nil {
        return NewStringError("PLUGIN_MESSAGE:アクティブなバッファがありません")
    }
    
    // バッファ情報を収集
    name := buffer.Name()
    content := buffer.Content()
    lines := strings.Count(content, "\n") + 1
    chars := len(content)
    words := len(strings.Fields(content))
    
    // 統計情報を表示
    stats := fmt.Sprintf("[STATS] %s: %d行, %d文字, %d単語", 
        name, lines, chars, words)
    
    // ホストにメッセージ表示を依頼
    p.host.ShowMessage(stats)
    
    // ミニバッファにも表示
    return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:%s", stats))
}
```

このAPIリファレンスを参考に、gmacs プラグインの開発を進めてください。