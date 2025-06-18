# @spec: input/text_input
# @test_files: test/text_input_test.go, test/enter_timing_test.go
# @implementation: domain/buffer.go, domain/editor.go

Feature: テキスト入力と文字処理
  As a ユーザー
  I want to 自然にテキストを入力したい
  So that 効率的にコンテンツを書いて編集できる

  Background:
    Given gmacsが実行されている
    And バッファが開いている

  Scenario: 基本的なASCII文字入力
    Given カーソルが空のバッファの先頭にある
    When 文字 'a' を入力する
    Then その文字がバッファに表示される
    And カーソルが1文字分進む
    And 必要に応じて自動スクロールが発動する

  Scenario: Enterキーが新しい行を作成
    Given カーソルが "hello" という内容の行の末尾にある
    When Enterキーを押す
    Then 新しい行が作成される
    And カーソルが新しい行の先頭に移動する
    And 必要に応じて自動スクロールが発動する

  Scenario: 文字とEnterの連続入力
    Given 空のバッファがある
    When 'a' を入力し、Enterを押し、'b' を入力し、Enterを押す
    Then バッファに "a" と "b" の2行が含まれる
    And カーソルが3行目の先頭にある
    And 自動スクロールがカーソルの可視性を維持する

  Scenario: 日本語テキスト入力（UTF-8）
    Given カーソルが空のバッファの先頭にある
    When 日本語文字 "こんにちは" を入力する
    Then 日本語テキストが正しく表示される
    And カーソル位置が文字幅を考慮する
    And 表示幅が正しく計算される

  # メモ: UTF-8処理の注意点
  # - cursor.Col はバイト位置で管理
  # - 表示幅計算には util.StringWidth を使用
  # - 日本語文字は表示幅2、バイト長3の場合が多い