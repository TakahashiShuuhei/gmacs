# gmacs BDD Test Management System

## 概要

このシステムは、gmacsプロジェクトの振る舞い駆動開発(BDD)テスト仕様を管理し、テストコードとドキュメントの同期を自動化します。

## システム構成

### 1. 仕様書管理 (`specs/features/`)

Gherkin記法を使用したBDD仕様書を機能別に分類して管理：

```
specs/features/
├── display/        # 画面表示機能
├── input/          # キーボード入力機能  
├── scroll/         # スクロール機能
├── buffer/         # バッファ管理
└── editor/         # エディタ基本機能
```

### 2. テストコードアノテーション

JavaDoc風のコメントでテストとBDD仕様を連携：

```go
/**
 * @spec scroll/scroll_timing
 * @scenario No premature scrolling
 * @description Tests that scrolling doesn't occur until content truly exceeds the window content area
 * @given A terminal with 12 lines (10 content + mode + mini)
 * @when Input characters a through j with Enter between each  
 * @then All 10 lines should be visible without scrolling
 * @implementation domain/scroll.go, cli/display.go
 */
func TestTerminal12LinesScenario(t *testing.T) {
    // テスト実装
}
```

### 3. ドキュメント生成システム

#### a) BDD仕様書生成 (`specs/tools/generate-docs.go`)

- `.feature`ファイルから HTML/Markdown ドキュメントを生成
- 機能別索引ページの作成
- 特定機能のみの生成も可能

#### b) テスト仕様抽出 (`specs/tools/extract-test-docs.go`)

- テストコードのアノテーションから仕様書を抽出
- テストと仕様の対応関係を明示
- バグ修正履歴の記録

### 4. 自動化システム (`Makefile`)

開発ワークフローの自動化：

```bash
make docs         # BDD仕様書生成
make test-docs    # テスト仕様抽出  
make docs-all     # 両方実行
make verify       # テスト + ドキュメント生成
make dev          # 開発サイクル全体
```

## 使用方法

### 新しい機能の追加

1. **BDD仕様書作成**
   ```bash
   # specs/features/新機能/機能名.feature を作成
   vim specs/features/search/text_search.feature
   ```

2. **テスト実装**
   ```bash
   # test/テスト名_test.go を作成し、アノテーションを追加
   vim test/text_search_test.go
   ```

3. **ドキュメント生成**
   ```bash
   make docs-all
   ```

### 既存機能の修正

1. **仕様書更新**
   - 対応する `.feature` ファイルを修正

2. **テスト更新**
   - テストコードとアノテーションを更新

3. **同期確認**
   ```bash
   make verify  # テスト実行 + ドキュメント再生成
   ```

### ドキュメント閲覧

- **HTML版**: `specs/docs/index.html` をブラウザで開く
- **テスト仕様**: `specs/test-docs.md` を確認

## 利点

### 1. 仕様とテストの同期維持
- アノテーションによる明確な関連付け
- 自動ドキュメント生成による同期確保

### 2. 可読性の向上
- Gherkin記法による自然言語での仕様記述
- HTML出力による見やすいドキュメント

### 3. トレーサビリティ
- バグ修正履歴の記録
- 実装ファイルとの対応関係明示

### 4. 自動化による効率化
- Makefileによるワンコマンド実行
- CI/CDパイプラインへの組み込み可能

## 拡張性

### カスタムアノテーション
- `@priority`, `@tags` などの追加アノテーション
- プロジェクト固有の情報管理

### 出力フォーマット
- PDF生成機能の追加
- Confluence連携
- Slack通知システム

### 自動化拡張
- Git pre-commit hook での自動生成
- CI での仕様書更新チェック

## ベストプラクティス

1. **仕様書は最初に書く** - TDD/BDDの原則に従う
2. **テストアノテーションは必須** - 仕様との関連を明確にする
3. **定期的なドキュメント更新** - `make docs-all` を習慣化
4. **バグ修正時は@bug_fixを記録** - 同じ問題の再発防止

## 今後の改善案

- [ ] Visual Studio Code拡張の開発
- [ ] リアルタイムプレビュー機能
- [ ] テストカバレッジとの連携
- [ ] 仕様変更履歴の追跡
- [ ] 国際化対応（英語版ドキュメント）