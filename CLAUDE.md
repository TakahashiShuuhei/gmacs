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
├── test/       # E2E テスト
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
- Ctrl+C での終了
- Enter キーでの改行
- 修飾キー（Ctrl、Meta）の認識

## テスト戦略

### E2E テストの分類
- `editor_startup_test.go`: エディタの起動と基本動作
- `text_input_test.go`: テキスト入力機能（ASCII、日本語、改行）
- `keyboard_shortcuts_test.go`: キーボードショートカット
- `event_system_test.go`: イベントシステムとパフォーマンス

### テスト実行
```bash
go test ./test/          # 全テスト実行
go test ./test/ -v       # 詳細モード
go test ./test/ -bench=. # ベンチマーク実行
```

## ビルドと実行

### ビルド
```bash
go build -o gmacs
```

### 実行
```bash
./gmacs
```

### 終了
- `Ctrl+C`: エディタを終了

## 開発ルール

### コーディング規約
- Go の標準的なコーディングスタイルに従う
- テスト駆動開発を推奨
- 公開関数・メソッドには適切な名前を付ける

### git 管理
- `.gitignore` でビルド成果物や IDE ファイルを除外
- コミット前に `go test ./test/` でテストを実行

### 追加機能の実装方針
1. まずドメインロジックを実装
2. 対応する E2E テストを作成
3. CLI 部分の実装
4. 統合テスト

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

## 参考情報

このプロジェクトは GNU Emacs の動作を参考にしていますが、完全な互換性は目指していません。基本的な編集機能と Emacs らしい操作感の提供を目標としています。