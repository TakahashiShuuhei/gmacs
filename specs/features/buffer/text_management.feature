# @spec: buffer/text_management
# @test_files: test/text_input_test.go, test/cursor_movement_test.go
# @implementation: domain/buffer.go, domain/cursor.go

Feature: バッファテキスト管理
  As a ユーザー
  I want to バッファ内のテキストコンテンツを管理したい
  So that 効率的にドキュメントを編集できる

  Background:
    Given gmacsが起動している
    And バッファが利用可能である

  Scenario: カーソル位置での文字挿入
    Given カーソルがバッファ内の特定の位置にある
    When 文字を入力する
    Then 文字がカーソル位置に挿入される
    And カーソルが挿入された文字の後に進む
    And バッファが変更済みとしてマークされる

  Scenario: Enterキーでの行作成
    Given カーソルが行の末尾にある
    When Enterキーを押す
    Then 現在の行の後に新しい行が作成される
    And カーソルが新しい行の先頭に移動する
    And バッファの行数が1増加する

  Scenario: UTF-8文字の処理
    Given カーソルがバッファの先頭にある
    When マルチバイトUTF-8文字を入力する
    Then 文字がバッファに正しく保存される
    And カーソル位置がバイト長を考慮する
    And 表示が正しい文字幅を示す

  Scenario: 行内でのカーソル移動
    Given バッファに "hello world" というテキストが含まれている
    And カーソルが行の先頭にある
    When カーソルを5文字分前進させる
    Then カーソルが "hello" の後に位置する
    And カーソル下の文字がスペースである

  # メモ: バッファ管理の重要な点
  # - cursor.Col はバイト位置で管理（UTF-8対応）
  # - 表示幅計算は util.StringWidth で実行
  # - バッファ変更時の modified フラグ管理