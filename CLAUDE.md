# gmacs - Emacsライクなテキストエディタ

## 概要

gmacs は Go で実装された Emacs ライクなテキストエディタです。ターミナル上で動作し、将来的な GUI 対応を考慮してドメインロジックと CLI 部分を明確に分離したアーキテクチャを採用しています。

## アーキテクチャ

### ディレクトリ構成
```
core/
├── domain/     # ドメインロジック（Buffer、Window、Editor）
├── events/     # イベントシステム
├── cli/        # CLI インターフェース（Display、Terminal）
├── e2e-test/   # E2E テスト（BDD形式、日本語アノテーション）
├── util/       # ユーティリティ関数
├── log/        # ログ機能
├── specs/      # BDD仕様書管理
│   ├── features/    # 日本語Gherkin仕様書
│   ├── tools/       # ドキュメント生成ツール
│   └── test-docs.md # 自動生成テストドキュメント
├── Makefile    # ビルド・テスト自動化
├── main.go     # メインエントリポイント
└── CLAUDE.md   # このファイル
```

### 設計原則

1. **ドメインと CLI の分離**
   - ドメインロジックは CLI に依存しない
   - イベントキューを通じた疎結合な設計
   - テスト可能性を重視

2. **Emacs ライクな構造**
   - Buffer: テキストデータとカーソル位置の管理
   - Window: Buffer の表示とスクロール状態の管理
   - Editor: 全体の状態管理とイベント処理

3. **イベント駆動アーキテクチャ**
   - キーボード入力、ウィンドウリサイズ、終了などのイベント
   - 非同期イベント処理による高いレスポンス性

## キーボード入力対応

### 対応予定の入力タイプ
- **Raw キー入力**: M-x、C-c などの Emacs ライクなキーバインド
- **IME 入力**: 日本語変換対応
- **ASCII 入力**: 通常の文字入力

### 現在の実装状況
- 基本的な ASCII 文字入力
- C-x C-c でのエディタ終了（Emacs準拠）
- 汎用的なキーシーケンスバインディングシステム（BindKeySequence API）
- Enter キーでの改行
- 修飾キー（Ctrl、Meta）の認識
- 基本的なカーソル移動（C-f, C-b, C-p, C-n, C-a, C-e）
- M-x コマンドシステム

## テスト戦略

### テスト分類
#### 1. E2E テスト (`e2e-test/`)
BDD形式の日本語アノテーション付きエンドツーエンドテスト。
`test-docs.md`の自動生成対象。

**主要テストファイル:**
- `editor_startup_test.go`: エディタの起動と基本動作
- `text_input_test.go`: テキスト入力機能（ASCII、日本語、改行）
- `keyboard_shortcuts_test.go`: キーボードショートカット
- `buffer_interactive_test.go`: バッファ管理機能
- `event_system_test.go`: イベントシステムとパフォーマンス

#### 2. 単体テスト (各パッケージ内)
通常のGoテスト形式で、各パッケージのディレクトリ内に配置。
`test-docs.md`には含まれない。

**配置ルール:**
```
domain/
├── buffer.go
├── buffer_test.go          # bufferの単体テスト
├── window.go
├── window_test.go          # windowの単体テスト
└── ...
```

### テスト実行
```bash
make test                       # 全E2Eテスト実行
make test-pattern PATTERN=名前  # 特定E2Eテストのみ実行
go test ./e2e-test/ -v          # E2Eテスト詳細モード
go test ./domain/ -v            # 単体テスト実行例
go test ./... -v                # 全テスト（E2E + 単体）実行
```

## ビルドと実行

### ビルド
```bash
make build    # Makefileを使用（推奨）
go build -o gmacs  # 直接実行
```

### 実行
```bash
./gmacs
```

### 開発サイクル
```bash
make dev      # ビルド + テスト + ドキュメント生成
make verify   # テスト + ドキュメント生成のみ
```

### 終了
- `C-x C-c`: エディタを終了（Emacsと同じ）

## キーバインディングシステム

### キーシーケンスバインディング
gmacs では Emacs スタイルのマルチキーシーケンスをサポートしています。

#### API使用例
```go
// 統一されたキーシーケンスAPI
kbm.BindKeySequence("C-f", ForwardChar)     // 単一キー: C-f
kbm.BindKeySequence("C-x C-c", Quit)        // マルチキー: C-x C-c
kbm.BindKeySequence("C-x C-f", FindFile)    // マルチキー: C-x C-f（将来実装）
kbm.BindRawSequence("\x1b[C", ForwardChar)  // 生エスケープ: 右矢印キー
```

#### システム特徴
- **統一API**: `BindKeySequence()` で単一キーもマルチキーも統一
- **汎用的**: 任意の長さのキーシーケンスをサポート
- **設定と実装の分離**: コードとキーバインディング設定が独立
- **テスト可能**: `NewEmptyKeyBindingMap()` でテスト用の空マップを作成可能
- **状態管理**: prefix key の状態を自動管理
- **リセット機能**: Escape キーでシーケンス状態をリセット

#### 実装詳細
- `KeyPress` 構造体で個別のキー押下を表現
- `KeySequenceBinding` で複数キーのシーケンスを管理
- `RawSequenceBinding` でエスケープシーケンス（矢印キー等）を管理
- `ProcessKeyPress()` で段階的なマッチングを実行
- 部分マッチ、完全マッチ、マッチ失敗を区別して処理

## 開発ルール

### コーディング規約
- Go の標準的なコーディングスタイルに従う
- テスト駆動開発を推奨
- 公開関数・メソッドには適切な名前を付ける

### git 管理
- `.gitignore` でビルド成果物や IDE ファイルを除外
- コミット前に `go test ./test/` でテストを実行

### 追加機能の実装方針（BDD方式）
1. **仕様書作成**: `specs/features/` で日本語Gherkin仕様を書く
2. **テスト実装**: 日本語アノテーション付きでテストを実装
3. **ドメインロジック実装**: TDD/BDDに従って実装
4. **CLI 部分の実装**: インターフェース層を実装
5. **ドキュメント更新**: `make docs` で自動ドキュメント生成

## ログ機能

### ログファイル
- プロセス起動ごとに `logs/gmacs_YYYYMMDD_HHMMSS.log` 形式で新しいファイルを作成
- タイムスタンプ付きでマイクロ秒精度のログ出力
- 自動的に `logs/` ディレクトリを作成

### ログレベル
- `Debug`: 詳細なデバッグ情報（キー入力、イベント処理等）
- `Info`: 一般的な情報（起動、終了、ウィンドウリサイズ等）
- `Warn`: 警告（未知のイベント、状態不整合等）
- `Error`: エラー（ターミナル初期化失敗等）

### 使用例
```go
log.Debug("Key event: key=%s, rune=%c", event.Key, event.Rune)
log.Info("gmacs starting up")
log.Warn("No current buffer for key event")
log.Error("Failed to initialize terminal: %v", err)
```

## 既知の問題

### UTF-8/日本語処理
- 現在日本語入力で文字化けが発生
- `TestJapaneseTextInput` が失敗中
- UTF-8 エンコーディング処理の改善が必要

### 今後の改善点
- より詳細なキーバインド対応
- バッファ管理機能の拡張
- ファイル読み書き機能
- 検索・置換機能
- 複数ウィンドウ対応

## 依存関係

### 外部ライブラリ
- `golang.org/x/term`: ターミナル制御
- `golang.org/x/sys`: システム依存処理

### Go バージョン
- Go 1.22.2 以上

## BDD仕様書管理システム

### 概要
gmacs は日本語対応の振る舞い駆動開発(BDD)システムを採用しています。

### ワークフロー
```bash
# 1. 仕様書作成（手動）
vim specs/features/新機能/機能名.feature

# 2. テスト実装（日本語アノテーション付き）
vim test/新機能_test.go

# 3. 自動ドキュメント生成
make docs

# 4. 開発サイクル実行
make dev  # build + test + docs
```

### テストアノテーション形式
```go
/**
 * @spec 機能カテゴリ/機能名
 * @scenario シナリオ名
 * @description 説明
 * @given 前提条件
 * @when 操作
 * @then 期待結果
 * @implementation 実装ファイル
 * @bug_fix バグ修正内容（オプション）
 */
func TestFunctionName(t *testing.T) {
    // テスト実装
}
```

### Gherkin仕様書形式
```gherkin
Feature: 機能名
  As a ユーザー
  I want to やりたいこと
  So that 得られる価値

  Scenario: シナリオ名
    Given 前提条件
    When 操作
    Then 期待結果
    And 追加条件
```

### 自動化コマンド
```bash
make docs        # テストドキュメント生成
make test-docs   # 同上（エイリアス）
make verify      # テスト + ドキュメント生成
make dev         # ビルド + テスト + ドキュメント
```

### 生成されるドキュメント
- `specs/features/*.feature` - 日本語Gherkin仕様書（手動編集）
- `specs/test-docs.md` - テストから抽出した日本語ドキュメント（自動生成）

## 参考情報

このプロジェクトは GNU Emacs の動作を参考にしていますが、完全な互換性は目指していません。基本的な編集機能と Emacs らしい操作感の提供を目標としています。BDD仕様書により、機能の意図と実装の対応関係を明確に管理しています。