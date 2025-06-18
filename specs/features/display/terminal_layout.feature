# @spec: display/terminal_layout
# @test_files: test/display_layout_test.go, test/resize_test.go
# @implementation: cli/display.go

Feature: ターミナルレイアウト管理
  As a ユーザー
  I want to エディタがターミナル画面レイアウトを適切に管理したい
  So that コンテンツが無駄なスペースなく正しく表示される

  Background:
    Given gmacsがターミナルで実行されている
    And ターミナルが固定サイズである

  Scenario: 適切なコンテンツエリア計算
    Given 高さ12行のターミナル
    When gmacsがディスプレイを初期化する
    Then コンテンツエリアは10行である
    And モードライン用に1行が確保される
    And ミニバッファ用に1行が確保される

  Scenario: 余分な空行の回避
    Given 高さ10行のターミナル
    And 8行のコンテンツエリア
    When 5行のテキストを表示する
    Then 5行のコンテンツのみがレンダリングされる
    And 余分な空行が表示されない
    And モードラインがコンテンツの直後に表示される
    And ミニバッファがモードラインの後に表示される

  Scenario: コンテンツエリアの完全利用
    Given 高さ12行のターミナル
    And 10行のコンテンツエリア
    When バッファに10行のテキストが含まれている
    Then すべての10行が表示される
    And スクロールが発生しない
    And カーソルが画面上で見える

  Scenario: ターミナルリサイズ処理
    Given gmacsがコンテンツを表示している
    When ターミナルが24x80から12x40にリサイズされる
    Then コンテンツエリアが新しい高さ-2に調整される
    And ディスプレイが正しく再レンダリングされる
    And カーソル位置が有効のままである

  # メモ: 以前のバグ
  # - cli/display.goで height-2 と height-3 の不整合によりレンダリングエリアがずれていた
  # - 無条件の改行出力により大量の空行が発生していた
  # - MockDisplayと実際のDisplayで異なる動作をしていた