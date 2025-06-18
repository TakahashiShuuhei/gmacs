# @spec: editor/startup
# @test_files: test/editor_startup_test.go, test/event_system_test.go
# @implementation: domain/editor.go, main.go

Feature: エディタ起動と初期化
  As a ユーザー
  I want to gmacsが正しく起動したい
  So that すぐに編集を開始できる

  Scenario: 基本的なエディタ初期化
    When gmacsが起動する
    Then デフォルトバッファ "*scratch*" が作成される
    And バッファ用のウィンドウが作成される
    And ターミナルが初期化される
    And イベントシステムが準備される
    And キーバインディングが読み込まれる

  Scenario: ターミナルサイズ検出
    Given ターミナルのサイズが80x24である
    When gmacsが起動する
    Then ウィンドウコンテンツエリアが80x22になる
    And ディスプレイがターミナルサイズに適応する
    And 初期レンダリングが発生する

  Scenario: イベントシステム初期化
    When gmacsが起動する
    Then イベントキューが準備される
    And キーボード入力ハンドラーがアクティブになる
    And リサイズイベントハンドラーがアクティブになる
    And シグナルハンドラーがインストールされる

  Scenario: デフォルトキーバインディング設定
    When gmacsが起動する
    Then 基本的なEmacsキーバインディングがアクティブになる
    And 矢印キーがカーソル移動に機能する
    And Ctrl+Cでエディタが終了する
    And M-xでミニバッファがアクティブになる

  # メモ: 起動シーケンスの重要な点
  # - ターミナルサイズの検出と初期設定
  # - バッファとウィンドウの初期化順序
  # - イベントシステムの起動
  # - シグナルハンドラの設定