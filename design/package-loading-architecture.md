# パッケージ読み込みアーキテクチャ設計

## 背景

現在のgmacsパッケージ管理システムでは、Lua設定でパッケージ宣言はできるが、実際のGoパッケージのダウンロード・読み込みが未実装。

## 現状の問題

### 1. パッケージ宣言と実際の読み込みの乖離

**ユーザー設定（期待動作）:**
```lua
-- ~/.config/gmacs/init.lua
gmacs.packages = {
    {"github.com/someone/awesome-package", "v1.0.0"}
}

gmacs.after_packages_loaded(function()
    gmacs.global_set_key("C-i", "awesome-function")
    local doc = gmacs.awesome.show_doc("some_function")
end)
```

**現在の実装状況:**
- ✅ `gmacs.packages` テーブルの定義はできる
- ✅ `after_packages_loaded()` コールバックの登録はできる
- ❌ 実際のGoパッケージダウンロードは未実装
- ❌ Goパッケージからの動的API拡張は未実装
- ❌ `gmacs.awesome` API は存在しない

### 2. LoadConfig()の現在の処理フロー

```go
func (lc *LuaConfig) LoadConfig() error {
    lc.vm = lua.NewState()
    lc.exposeGmacsAPI()           // 基本APIのみ
    
    err := lc.loadDefaultConfig() // default.lua読み込み
    
    // ユーザー設定読み込み
    return lc.vm.DoFile(configPath) // ★ここでpackages宣言だけ実行される
    
    // ★パッケージの実際の読み込み処理は未実装★
}
```

## 設計案

### A案: Lua実行 → パッケージ読み込み → 再実行

```go
func (lc *LuaConfig) LoadConfig() error {
    // 1. 第1段階: パッケージ宣言のみ取得
    lc.vm = lua.NewState()
    lc.exposeGmacsAPI()
    err := lc.vm.DoFile(configPath)
    
    // 2. パッケージ読み込み
    packages := lc.extractDeclaredPackages() // gmacs.packagesから取得
    for _, pkg := range packages {
        err := lc.downloadAndLoadPackage(pkg)
    }
    
    // 3. 第2段階: API拡張後に再実行
    lc.vm.Close()
    lc.vm = lua.NewState()
    lc.exposeGmacsAPI() // 拡張API込み
    err = lc.vm.DoFile(configPath)
    
    // 4. コールバック実行
    return lc.executePackageLoadedCallbacks()
}
```

**メリット:**
- 既存の設定ファイル構造を維持
- Lua側でパッケージ宣言と使用が同じファイルに書ける

**デメリット:**
- 設定ファイルが2回実行される（副作用の懸念）
- 複雑な処理フロー

### B案: 2段階設定ファイル

```lua
-- ~/.config/gmacs/packages.lua (最初に読み込み)
return {
    {"github.com/someone/awesome-package", "v1.0.0"},
    {"github.com/another/great-mode", "v2.1.0"}
}
```

```lua
-- ~/.config/gmacs/init.lua (パッケージ読み込み後)
gmacs.global_set_key("C-i", "awesome-function")
local doc = gmacs.awesome.show_doc("some_function")
```

**メリット:**
- 明確な処理順序
- 副作用の心配なし
- パッケージ宣言がシンプル

**デメリット:**
- 設定ファイルが分離される
- ユーザビリティの低下

### C案: パッケージ宣言専用API

```lua
-- ~/.config/gmacs/init.lua
-- パッケージ宣言フェーズ
gmacs.use_package("github.com/someone/awesome-package", "v1.0.0")
gmacs.use_package("github.com/another/great-mode", "v2.1.0")

-- パッケージ読み込み完了後フェーズ
gmacs.after_packages_loaded(function()
    gmacs.global_set_key("C-i", "awesome-function")
    local doc = gmacs.awesome.show_doc("some_function")
end)
```

**メリット:**
- 設定ファイル1回実行
- 既存のafter_packages_loaded活用
- 明確な処理フェーズ分離

**デメリット:**
- ほぼ全ての設定がafter_packages_loaded内に包まれる
- インデント地獄
- 若干の学習コスト

### D案: パース→読み込み→実行（推奨）

```lua
-- ~/.config/gmacs/init.lua
-- パッケージ宣言（どこに書いても良い）
gmacs.use_package("github.com/awesome/ruby-mode", {
    ruby_path = "/custom/ruby"
})
gmacs.use_package("github.com/awesome/git-mode")

-- 普通に設定を書く（パッケージ読み込み済み前提）
gmacs.set_variable("theme", "dark")
gmacs.global_set_key("C-x C-f", "find-file")

-- パッケージのAPIも自然に使える
gmacs.global_set_key("C-c C-d", "ruby-show-doc")  -- ruby-mode提供
gmacs.global_set_key("C-x g s", "git-status")     -- git-mode提供

function my_function()
    local doc = gmacs.ruby.show_doc("String#gsub") -- ruby-mode API
    gmacs.message(doc)
end
gmacs.register_command("my-cmd", my_function)
```

```go
func (lc *LuaConfig) LoadConfig() error {
    // 1. 設定ファイルをパース（実行はしない）
    packages, err := lc.parsePackageDeclarations(configPath)
    
    // 2. パッケージダウンロード・読み込み
    for _, pkg := range packages {
        err := lc.downloadAndLoadPackage(pkg)
    }
    
    // 3. 全API準備完了後に設定ファイル実行
    lc.vm = lua.NewState()
    lc.exposeGmacsAPI() // 拡張API込み
    return lc.vm.DoFile(configPath)
}
```

**メリット:**
- ユーザーは普通に設定を書ける
- パッケージAPIも自然に使える
- 複雑なコールバック不要
- インデント地獄回避

**デメリット:**
- Luaファイルのパース処理が必要
- 技術的に若干複雑

## 技術的課題

### 1. Goパッケージの動的読み込み

**選択肢:**
- **Go plugin システム**: `plugin.Open()` で動的ライブラリ読み込み
- **事前コンパイル**: 依存パッケージを含めて再ビルド
- **Go module integration**: `go/build` パッケージで動的ビルド

### 2. パッケージ検索・検証

```go
type PackageRegistry interface {
    Search(query string) ([]PackageInfo, error)
    Validate(url, version string) error
    GetDependencies(url, version string) ([]PackageDeclaration, error)
}
```

### 3. セキュリティ・サンドボックス

- 任意のGoコードの実行を許可するリスク
- パッケージの署名・検証機構
- API権限制御

## 推奨案

**D案（パース→読み込み→実行）**を推奨する理由:

1. **ユーザビリティ**: 設定ファイルが自然に書ける
2. **直感性**: パッケージAPIが普通の関数のように使える
3. **保守性**: インデント地獄やコールバック地獄を回避
4. **柔軟性**: パッケージ宣言の位置を自由に選べる

## 次のステップ

1. **D案の詳細設計**
   - Luaファイルパース機構（AST解析）
   - `gmacs.use_package()` の詳細仕様
   - パッケージダウンロード・検証機構
   - エラーハンドリング

2. **プロトタイプ実装**
   - Luaパーサーの基本実装
   - Go plugin システムでの基本実装
   - サンプルパッケージでの動作確認

3. **セキュリティ設計**
   - パッケージ検証機構
   - API権限制御

## 関連ファイル

- `internal/config/lua_config.go` - Lua設定システム
- `internal/package/manager.go` - パッケージ管理
- `examples/init.lua.example` - 設定例

## 決定事項

- [x] **アーキテクチャ方針決定**: D案（パース→読み込み→実行）を採用
- [ ] パッケージ読み込み方式決定（plugin/rebuild/dynamic）
- [ ] セキュリティ要件定義

## 議論ログ

**2024-XX-XX**: 初期問題提起
- lc.vm.DoFile()でのパッケージ宣言と実際の読み込みの乖離を確認
- A/B/C案を提示

**2024-XX-XX**: D案追加・採用決定
- C案の問題点指摘: ほぼ全てがafter_packages_loaded内になる
- D案提案: パース→読み込み→実行で自然な設定記述を実現
- ユーザビリティ重視でD案採用決定