# @spec: scroll/scroll_timing
# @test_files: test/scroll_timing_test.go, test/terminal_12_lines_test.go, test/exact_user_scenario_test.go
# @implementation: domain/scroll.go, cli/display.go

Feature: スクロールタイミングとカーソル可視性
  As a ユーザー
  I want to 適切なタイミングでスクロールが発生したい
  So that カーソルが見え続け、コンテンツが自然に流れる

  Background:
    Given gmacsが実行されている
    And 行折り返しが有効である

  Scenario: 早すぎるスクロールの回避
    Given 12行のターミナル（10コンテンツ + モード + ミニ）
    When 文字a〜jをそれぞれEnterで区切って入力する
    Then すべての10行がスクロールなしで表示される
    And 最初の行に 'a' が表示されている
    And 最後の行に 'j' が表示されている
    And カーソルがコンテンツエリアの底部にある

  Scenario: コンテンツがウィンドウを超えた時のスクロール
    Given 12行のターミナル（10コンテンツエリア）
    And バッファにa〜jの10行が含まれている
    When 新しい行で 'k' を入力する
    Then スクロールが発生する
    And 最初の表示行が 'b' になる
    And 最後の表示行が 'k' を示す
    And カーソルが画面下部に残る

  Scenario: 底部でのEnter時の即座スクロール
    Given カーソルがコンテンツエリアの最後の表示行にある
    And コンテンツがコンテンツエリア全体を満たしている
    When Enterキーを押して新しい行を作成する
    Then スクロールが即座に発生する
    And カーソルが底部で見え続ける
    And スクロール応答に遅延が発生しない

  Scenario: カーソル可視性の維持
    Given コンテンツが複数ページにまたがっている
    When カーソルが表示エリアを超えて移動する
    Then 自動スクロールが発動する
    And カーソルが見えるようになる
    And 最小限のスクロールが使用される

  # メモ: ユーザー報告のバグ
  # 「高さ10のターミナルで a enter b enter ... h まで入力して次にenterを押したら
  #  b ~ h の行が表示される」← aが消えるのが早すぎる問題
  # 
  # 原因: display.goでコンテンツエリアの計算が間違っていた
  # - for i := 0; i < height-2 だが if i < height-3 で改行という不整合
  # - 実際には height-2 = 10行使えるのに 7行しか使っていなかった