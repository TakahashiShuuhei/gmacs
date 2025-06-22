# gmacs Plugin System Implementation Plan

## 概要

HashiCorp go-plugin を使用したプラグインシステムの実装計画。
8週間で段階的に機能を実装し、最終的にソースからビルド可能なプラグインエコシステムを構築する。

## 全体スケジュール

| Phase | 期間 | 内容 | 優先度 |
|-------|------|------|--------|
| Phase 1 | Week 1-2 | 基盤構築 | High |
| Phase 2 | Week 3-4 | コア機能実装 | High |
| Phase 3 | Week 5 | 設定システム | Medium |
| Phase 4 | Week 6-7 | ビルドシステム | Medium |
| Phase 5 | Week 8 | CLI とテスト | Medium |

## Phase 1: 基盤構築（Week 1-2）

### 目標
依存関係の追加とプラグインシステムの基本インターフェースを定義する。

### Week 1: 依存関係とプロジェクト構造

#### タスク
- [ ] **依存関係追加**
  ```bash
  # go.mod に追加
  github.com/hashicorp/go-plugin v1.4.0
  github.com/go-git/go-git/v5 v5.4.2
  google.golang.org/grpc v1.50.0
  google.golang.org/protobuf v1.28.0
  ```

- [ ] **ディレクトリ構造作成**
  ```
  plugin/
  ├── interface.go
  ├── protocol.proto
  ├── rpc.go
  ├── manager.go
  ├── host_api.go
  ├── discovery.go
  └── builder.go
  ```

- [ ] **基本インターフェース定義** (`plugin/interface.go`)
  - `Plugin` interface
  - `HostInterface` interface
  - `PluginManifest` struct
  - `CommandSpec`, `MajorModeSpec`, `MinorModeSpec` structs

### Week 2: gRPC プロトコルと基本実装

#### タスク
- [ ] **gRPCプロトコル定義** (`plugin/protocol.proto`)
  - `PluginService` service
  - `HostService` service
  - メッセージ型定義

- [ ] **プロトコルバッファ生成**
  ```bash
  protoc --go_out=. --go-grpc_out=. plugin/protocol.proto
  ```

- [ ] **gRPC実装** (`plugin/rpc.go`)
  - プラグイン側gRPCサーバー実装
  - ホスト側gRPCクライアント実装

- [ ] **XDGパス管理ユーティリティ**
  - `getXDGDataHome()` 関数
  - `getXDGConfigHome()` 関数
  - パス検索機能

#### 成果物
- 基本的なプラグインインターフェース
- gRPCプロトコル定義
- XDG準拠のパス管理

## Phase 2: コア機能実装（Week 3-4）

### 目標
PluginManagerとHostAPIを実装し、プラグインの基本的なライフサイクル管理を可能にする。

### Week 3: PluginManager基本実装

#### タスク
- [ ] **PluginManager構造体** (`plugin/manager.go`)
  - プラグインプロセス管理
  - ロード・アンロード機能
  - プラグイン状態管理

- [ ] **プラグインライフサイクル管理**
  - プラグインプロセス起動
  - ヘルスチェック
  - クラッシュ検出・回復

- [ ] **プラグインレジストリ**
  - インストール済みプラグイン管理
  - プラグイン情報キャッシュ

### Week 4: HostAPI実装とEditor統合

#### タスク
- [ ] **HostAPI実装** (`plugin/host_api.go`)
  - バッファ操作API
  - ウィンドウ操作API
  - コマンド実行API
  - フック登録API

- [ ] **Editor統合** (`domain/editor.go` 拡張)
  - プラグインマネージャー統合
  - プラグインコマンド登録
  - プラグインモード登録
  - プラグインキーバインド統合

- [ ] **エラーハンドリング**
  - プラグインエラー分離
  - 適切なエラーメッセージ
  - フォールバック処理

#### 成果物
- 動作するプラグインマネージャー
- プラグインとホストの通信機能
- 基本的なプラグイン機能統合

## Phase 3: 設定システム（Week 5）

### 目標
Lua設定システムとの統合、およびプラグイン自動ディスカバリー機能を実装する。

#### タスク
- [ ] **プラグインディスカバリー** (`plugin/discovery.go`)
  - XDG準拠のプラグイン検索
  - manifest.json読み込み
  - プラグイン依存関係チェック

- [ ] **Lua API拡張** (`lua-config/plugin_api.go`)
  - `plugin.setup()` 関数
  - `plugin.enable()` / `plugin.disable()` 関数
  - `plugin.config()` 関数
  - `plugin.list()` 関数

- [ ] **設定ファイル管理**
  - XDG準拠の設定パス
  - レガシー設定ファイルとの互換性
  - プラグイン固有設定管理

- [ ] **既存システムとの統合**
  - 設定読み込み時のプラグイン自動ロード
  - プラグイン設定の適用
  - 設定エラーハンドリング

#### 成果物
- プラグイン自動発見機能
- Lua設定統合
- 設定ファイル管理システム

## Phase 4: ビルドシステム（Week 6-7）

### 目標
ソースコードからプラグインをビルドする機能と、CLI プラグインマネージャーを実装する。

### Week 6: PluginBuilder実装

#### タスク
- [ ] **PluginBuilder構造体** (`plugin/builder.go`)
  - Git リポジトリクローン
  - ソースコード取得機能
  - ビルド作業ディレクトリ管理

- [ ] **ビルドプロセス実装**
  - `go mod download` による依存関係解決
  - `go build` によるバイナリ生成
  - ビルド成果物の配置

- [ ] **ビルドキャッシュシステム**
  - ソースコードハッシュ管理
  - ビルド済みバイナリキャッシュ
  - 差分ビルド対応

### Week 7: CLI インターフェース

#### タスク
- [ ] **CLI サブコマンド実装** (`main.go` 拡張)
  - `gmacs plugin install <repo>`
  - `gmacs plugin update <name>`
  - `gmacs plugin remove <name>`
  - `gmacs plugin list`
  - `gmacs plugin dev <local-path>`

- [ ] **コマンドライン引数処理**
  - サブコマンドルーティング
  - オプション解析
  - ヘルプメッセージ

- [ ] **プラグイン依存関係管理**
  - 依存関係解決
  - バージョン競合検出
  - 依存関係インストール

#### 成果物
- 完全なソースビルド機能
- CLI プラグインマネージャー
- 依存関係管理システム

## Phase 5: CLI とテスト（Week 8）

### 目標
E2Eテストの作成、ドキュメント整備、パフォーマンス最適化を行い、リリース準備を完了する。

#### タスク
- [ ] **E2Eテスト作成** (`e2e-test/plugin_*_test.go`)
  - プラグインインストールテスト
  - プラグインロード・アンロードテスト
  - プラグインコマンド実行テスト
  - エラーハンドリングテスト
  - 設定システムテスト

- [ ] **ドキュメント整備**
  - プラグイン開発ガイド
  - プラグインAPI リファレンス
  - インストール・使用方法ドキュメント
  - トラブルシューティングガイド

- [ ] **パフォーマンス最適化**
  - プラグインロード時間最適化
  - メモリ使用量最適化
  - gRPC通信最適化

- [ ] **リリース準備**
  - バージョニング
  - 変更履歴作成
  - リリースノート
  - プラグインSDK準備

#### 成果物
- 包括的なテストスイート
- 完全なドキュメント
- リリース可能なプラグインシステム

## サンプルプラグイン開発

### プラグインSDK
```go
// github.com/TakahashiShuuhei/gmacs-plugin-sdk
package sdk

type PluginBase struct {
    name        string
    version     string
    description string
}

func (p *PluginBase) Name() string { return p.name }
func (p *PluginBase) Version() string { return p.version }
func (p *PluginBase) Description() string { return p.description }
```

### サンプルプラグイン例
- **gmacs-go-mode**: Go言語サポート
- **gmacs-markdown-mode**: Markdownサポート
- **gmacs-lsp-client**: Language Server Protocol クライアント

## マイルストーン

### Milestone 1 (Week 2)
- [ ] 基本インターフェース完成
- [ ] gRPCプロトコル定義完成
- [ ] 最初のプラグインロード成功

### Milestone 2 (Week 4)
- [ ] PluginManager動作確認
- [ ] Editor統合完成
- [ ] 基本的なプラグイン機能動作

### Milestone 3 (Week 5)
- [ ] Lua設定統合完成
- [ ] プラグイン自動ディスカバリー動作
- [ ] 設定ファイル管理完成

### Milestone 4 (Week 7)
- [ ] ソースからビルド機能完成
- [ ] CLI プラグインマネージャー動作
- [ ] `gmacs plugin install` 成功

### Milestone 5 (Week 8)
- [ ] 全E2Eテスト通過
- [ ] ドキュメント完成
- [ ] リリース準備完了

## リスク管理

### 技術的リスク
- **gRPC通信の複雑さ**: プロトタイプで早期検証
- **プロセス間通信の安定性**: 十分なエラーハンドリング
- **ビルドシステムの複雑性**: 段階的実装

### スケジュールリスク
- **依存関係学習コスト**: Phase 1で十分な調査
- **統合の複雑さ**: 既存システムへの影響最小化
- **テスト網羅性**: 継続的テスト実装

### 対策
- 各フェーズ終了時にマイルストーン確認
- 週次進捗レビュー
- 必要に応じてスコープ調整

## 成功指標

### 技術指標
- [ ] プラグインロード時間 < 100ms
- [ ] メモリオーバーヘッド < 10MB per plugin
- [ ] 全E2Eテスト通過率 100%

### ユーザビリティ指標
- [ ] プラグインインストール 3コマンド以内
- [ ] 設定ファイル記述量 < 5行 per plugin
- [ ] エラーメッセージの明確性

### エコシステム指標
- [ ] サンプルプラグイン 3個以上作成
- [ ] プラグイン開発ドキュメント完備
- [ ] コミュニティ向けリソース準備

---

この計画に従って、段階的にプラグインシステムを実装していきます。各フェーズでマイルストーンを確認し、必要に応じてスケジュールを調整します。