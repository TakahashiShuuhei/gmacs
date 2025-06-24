# gmacs プラグインシステム ドキュメント

gmacs のプラグインシステムに関する包括的なドキュメントです。

## ドキュメント一覧

### [プラグイン開発ガイド](plugin-development.md)
プラグイン開発の基本的な流れとアーキテクチャについて説明します。
- プラグインの基本構造
- インターフェースの実装方法
- メッセージ表示システム
- モードとキーバインドの実装
- ビルドとインストール方法
- デバッグとトラブルシューティング

### [プラグイン API リファレンス](plugin-api-reference.md)
HostInterface API の完全なリファレンスドキュメントです。
- HostInterface の全メソッド詳細
- BufferInterface と WindowInterface API
- パラメータと戻り値の仕様
- 使用例とサンプルコード
- エラーハンドリング

### [プラグイン実装例](plugin-examples.md)
実際に動作するプラグインの実装例を示します。
- シンプルな挨拶プラグイン
- テキスト処理プラグイン
- ファイル管理プラグイン
- 開発者向けユーティリティプラグイン
- 完全なソースコード付き

## 概要

gmacs は HashiCorp go-plugin を使用した強力なプラグインシステムを提供しています。

### 主な特徴

- **独立プロセス**: プラグインは独立したプロセスとして実行され、クラッシュ時もホストに影響しません
- **双方向通信**: プラグインからホストへの API 呼び出しが可能
- **包括的 API**: バッファ、ウィンドウ、ファイル操作などの完全な API
- **拡張可能**: コマンド、モード、キーバインドをプラグインで追加可能
- **型安全**: Go の型システムによる安全な API

### アーキテクチャ

```
gmacs ホスト                     プラグイン
┌─────────────┐                ┌─────────────┐
│   Editor    │                │   Plugin    │
│             │◄──── RPC ────►│             │
│ HostInterface │                │ CommandPlugin │
└─────────────┘                └─────────────┘
```

## クイックスタート

### 1. プラグインプロジェクトの作成

```bash
mkdir my-gmacs-plugin
cd my-gmacs-plugin
go mod init github.com/user/my-gmacs-plugin
```

### 2. 必要な依存関係の追加

```go
// go.mod
require (
    github.com/TakahashiShuuhei/gmacs-plugin-sdk v0.1.0
    github.com/hashicorp/go-plugin v1.4.10
)
```

### 3. manifest.json の作成

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom gmacs plugin",
  "binary": "my-plugin",
  "dependencies": []
}
```

### 4. プラグインの実装

詳細は [プラグイン開発ガイド](plugin-development.md) を参照してください。

### 5. ビルドとインストール

```bash
go build -o my-plugin
gmacs plugin install .
```

### 6. 使用

```
M-x my-command
```

## プラグイン開発フロー

1. **設計**: どのような機能を提供するかを決定
2. **実装**: Plugin インターフェースと CommandPlugin インターフェースを実装
3. **テスト**: ローカルでビルドしてテスト
4. **配布**: GitHub などのリポジトリで公開

## 利用可能な API

プラグインから以下の API を使用できます：

### メッセージ表示
- `ShowMessage()` - ミニバッファにメッセージ表示
- `SetStatus()` - ステータスライン設定

### バッファ操作
- `GetCurrentBuffer()` - 現在のバッファ取得
- `CreateBuffer()` - 新規バッファ作成
- `FindBuffer()` - バッファ検索
- バッファ内容編集、カーソル操作

### ウィンドウ操作
- `GetCurrentWindow()` - 現在のウィンドウ取得
- ウィンドウサイズ、スクロール操作

### ファイル操作
- `OpenFile()` - ファイルを開く
- `SaveBuffer()` - バッファを保存

### その他
- コマンド実行、オプション管理
- モード管理、フックシステム

詳細は [API リファレンス](plugin-api-reference.md) を参照してください。

## 参考リソース

### 公式ドキュメント
- [gmacs メインドキュメント](../CLAUDE.md)
- [gmacs プラグイン SDK](https://github.com/TakahashiShuuhei/gmacs-plugin-sdk)

### サンプルプラグイン
- [gmacs-example-plugin](https://github.com/TakahashiShuuhei/gmacs-example-plugin) - 包括的なサンプル実装

### 外部リソース
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - 使用しているプラグインライブラリ
- [Go 公式ドキュメント](https://golang.org/doc/) - Go 言語の基本

## サポート

プラグイン開発に関する質問やバグ報告は、以下のリポジトリで受け付けています：

- [gmacs メインリポジトリ](https://github.com/TakahashiShuuhei/gmacs)
- [プラグイン SDK リポジトリ](https://github.com/TakahashiShuuhei/gmacs-plugin-sdk)

---

より詳細な情報は、各ドキュメントファイルを参照してください。