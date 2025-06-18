# gmacs テストドキュメント

このドキュメントはテストコードから自動抽出されたBDD仕様書です。

**生成日時:** 2025年06月18日 23:42:15

## 画面表示機能 (display/terminal_width_handling)

### TestMockDisplayWidthProblem

**ファイル:** `test/display_test.go`

**シナリオ:** ターミナル幅と文字幅の問題検証

**説明:** 異なる文字タイプのターミナル幅処理の検証

**前提:** 10x3サイズのMockDisplayを作成

**操作:** ASCII、日本語、混在テキストを各々入力

**結果:** ターミナル表示幅とルーン数の違いを適切に処理する

**実装ファイル:** `test/mock_display.go`, `文字幅計算`

---

## events/quit_handling

### TestQuitEvent

**ファイル:** `test/event_system_test.go`

**シナリオ:** 終了イベントの処理

**説明:** 終了イベントの処理とエディタ状態の変更

**前提:** エディタが実行中の状態

**操作:** QuitEventDataを送信する

**結果:** エディタが終了状態に変更される

**実装ファイル:** `events/quit_event.go`, `domain/editor.go`

---

## resize/multiple_resizes

### TestMultipleResizes

**ファイル:** `test/resize_test.go`

**シナリオ:** 連続的なリサイズ操作

**説明:** 複数回のリサイズ操作でのサイズ更新とコンテンツ保持

**前提:** 80x24サイズで"test content"を入力済み

**操作:** 異なるサイズで複数回連続してリサイズする

**結果:** 各リサイズ後にサイズが正確に更新され、コンテンツが保持される

**実装ファイル:** `domain/window.go`, `events/resize_event.go`

---

## スクロール機能 (scroll/line_wrapping)

### TestLineWrapping

**ファイル:** `test/scrolling_test.go`

**シナリオ:** 行ラップ機能

**説明:** 長い行のラップ機能の有効/無効切り替え検証

**前提:** 10x5の小さいウィンドウと長い行のコンテンツ

**操作:** 行ラップの有効/無効を切り替える

**結果:** ラップ有効時は複数行、無効時は単一行で表示される

**実装ファイル:** `domain/window.go`, `行ラップ処理`

---

## キーボード入力機能 (input/newline)

### TestEnterKeyNewline

**ファイル:** `test/text_input_test.go`

**シナリオ:** Enter キーによる改行

**説明:** Enter キーで行を分割し複数行テキストを作成

**前提:** エディタに "Hi" を入力済み

**操作:** Enter キーを押して "Wo" を入力する

**結果:** 2行に分かれてテキストが表示される

**実装ファイル:** `domain/buffer.go`, `events/key_event.go`

---

## cursor/japanese_support

### TestCursorMovementWithJapanese

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 日本語文字を含むカーソル移動

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル移動

**前提:** "aあbいc"（ASCII+日本語混在）を入力済み

**操作:** C-fで1文字ずつ前進する

**結果:** マルチバイト文字を適切に処理してカーソルが移動する

**実装ファイル:** `domain/cursor.go`, `UTF-8処理`

---

## keyboard/ctrl_x_ctrl_c_quit

### TestCtrlXCtrlCQuit

**ファイル:** `test/keyboard_shortcuts_test.go`

**シナリオ:** C-x C-cでのエディタ終了

**説明:** C-x C-cキーシーケンスでエディタを終了する機能の検証

**前提:** エディタが実行中の状態

**操作:** C-x（prefix key）とC-cキーイベントを順次送信する

**結果:** エディタが終了状態になる

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## commands/mx_unknown

### TestMxUnknownCommand

**ファイル:** `test/mx_command_test.go`

**シナリオ:** 未知のM-xコマンドのエラー処理

**説明:** 存在しないコマンドを実行した際のエラーハンドリング

**前提:** M-xコマンドモードを有効化

**操作:** 存在しないコマンド"nonexistent"を入力してEnterを押下

**結果:** エラーメッセージがミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `エラー処理`

---

## 画面表示機能 (display/newline_rendering)

### TestNewlineDisplay

**ファイル:** `test/newline_display_test.go`

**シナリオ:** 改行表示のレンダリング

**説明:** 改行を含む複数行コンテンツの正確な表示検証

**前提:** 20x5サイズのMockDisplayを作成

**操作:** "hello" + Enter + "world"を入力

**結果:** 2行のコンテンツが正確に表示され、カーソル位置が適切に設定される

**実装ファイル:** `test/mock_display.go`, `改行処理`

---

## 画面表示機能 (display/multiline_newline)

### TestNewlineAtEndOfLine

**ファイル:** `test/newline_display_test.go`

**シナリオ:** 複数改行での行末処理

**説明:** 連続した改行操作での行末処理とコンテンツ構築

**前提:** エディタを新規作成する

**操作:** "abc" + Enter + "def" + Enter + "ghi"を順次入力

**結果:** 3行のコンテンツが正確に作成され、カーソルが最終行の末尾に配置される

**実装ファイル:** `domain/buffer.go`, `複数行改行処理`

---

## スクロール機能 (scroll/vertical_scrolling)

### TestVerticalScrolling

**ファイル:** `test/scrolling_test.go`

**シナリオ:** 垂直スクロール動作

**説明:** 大量のコンテンツがある場合の垂直スクロール動作の検証

**前提:** 40x10サイズのウィンドウに20行のコンテンツを作成

**操作:** カーソルが最後の行にある状態でスクロール位置を設定

**結果:** カーソルが可視範囲に保たれるように自動スクロールされる

**実装ファイル:** `domain/window.go`, `domain/scroll.go`

---

## 画面表示機能 (display/terminal_layout)

### TestActualDisplayIssue

**ファイル:** `test/actual_display_issue_test.go`

**シナリオ:** 余分な空行の回避

**説明:** 余分な空行が表示されたユーザー報告シナリオの正確なテスト

**前提:** 12行のターミナル（ユーザーの報告環境）

**操作:** 文字a〜dをそれぞれEnterで区切って入力する

**結果:** 余分な空白なしで実際のコンテンツ行のみがレンダリングされる

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

**バグ修正:** height-2 vs height-3の不整合と無条件改行出力を修正

---

## cursor/mx_commands

### TestInteractiveCommands

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** M-xコマンドによるカーソル移動

**説明:** M-x beginning-of-lineコマンドの実行検証

**前提:** "hello"を入力済みでカーソルが行末にある

**操作:** M-x beginning-of-lineコマンドを実行

**結果:** カーソルが行頭に移動する

**実装ファイル:** `domain/commands.go`, `events/key_event.go`

---

## スクロール機能 (scroll/edge_case_debug)

### TestDebugScrollBehavior

**ファイル:** `test/debug_scroll_test.go`

**シナリオ:** スクロールエッジケースのデバッグ

**説明:** 8行丁度まで埋めた後のEnterキー押下時のスクロール動作の詳細分析

**前提:** 40x10ディスプレイ（8コンテンツ行）で8行丁度までコンテンツを埋める

**操作:** 最後の可視行でEnterキーを押下

**結果:** スクロール量と表示内容が期待値と一致し、適切な1行スクロールが発生する

**実装ファイル:** `domain/scroll.go`, `エッジケース処理`

---

## 画面表示機能 (display/mock_vs_real)

### TestRealVsMockDisplay

**ファイル:** `test/display_layout_test.go`

**シナリオ:** MockDisplayと実際のDisplay比較

**説明:** ユーザー報告シナリオでのMockDisplayと実際のCLI Displayの動作比較

**前提:** 40x10ターミナルでa〜hまで8行のコンテンツを作成

**操作:** 最後にEnterキーを押下

**結果:** MockDisplayの動作がユーザー期待（bから始まる表示）と一致する

**実装ファイル:** `test/mock_display.go`, `cli/display.go`

---

## 画面表示機能 (display/japanese_rendering)

### TestMockDisplayJapanese

**ファイル:** `test/display_test.go`

**シナリオ:** 日本語テキスト表示

**説明:** 日本語文字の表示と表示幅計算の検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "あいう"（ひらがな）を入力する

**結果:** 日本語テキストが正確に表示され、カーソル位置が適切に計算される

**実装ファイル:** `test/mock_display.go`, `UTF-8処理`

---

## 画面表示機能 (display/multiline_rendering)

### TestMockDisplayMultiline

**ファイル:** `test/display_test.go`

**シナリオ:** 複数行テキスト表示

**説明:** 複数行のテキストとカーソル位置の表示検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "hello" + Enter + "world"を入力

**結果:** 2行のテキストが正確に表示され、2行目にカーソルが配置される

**実装ファイル:** `test/mock_display.go`, `複数行処理`

---

## キーボード入力機能 (input/newline_multiple)

### TestMultipleNewlines

**ファイル:** `test/newline_test.go`

**シナリオ:** 連続した改行挿入

**説明:** 複数のEnterキーを連続して押した際の動作

**前提:** 空のバッファから開始

**操作:** "a" + Enter + "b" + Enter + "c"を順次入力

**結果:** 3行のコンテンツが正確に作成され、カーソル位置が適切に設定される

**実装ファイル:** `domain/buffer.go`, `複数行処理`

---

## text/string_width_calculation

### TestStringWidth

**ファイル:** `test/runewidth_test.go`

**シナリオ:** 文字列幅計算機能

**説明:** ASCII、日本語、混合文字列の総表示幅計算

**前提:** 空文字列、ASCII文字列、日本語文字列、混合文字列のテストケース

**操作:** StringWidth関数で各文字列の総表示幅を計算

**結果:** 各文字の幅の合計値が正確に計算される（混合文字列は範囲チェック）

**実装ファイル:** `util/runewidth.go`, `文字列幅計算`

---

## スクロール機能 (scroll/toggle_line_wrap)

### TestToggleLineWrap

**ファイル:** `test/scrolling_test.go`

**シナリオ:** 行ラップトグルコマンド

**説明:** ToggleLineWrapコマンドによる行ラップ状態の切り替え

**前提:** エディタを新規作成（デフォルトでラップ有効）

**操作:** ToggleLineWrapコマンドを実行

**結果:** 行ラップの有効/無効が切り替わる

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## 画面表示機能 (display/mock_consistency)

### TestDisplayConsistency

**ファイル:** `test/actual_display_issue_test.go`

**シナリオ:** MockDisplayと実際のDisplay一貫性確認

**説明:** MockDisplayとWindow.VisibleLines()の表示内容が一致することを確認

**前提:** 80x10ターミナル環境

**操作:** 3行のテキスト（a、b、c）を入力する

**結果:** MockDisplayの内容とWindow.VisibleLines()が完全に一致する

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

---

## cursor/vertical_movement

### TestNextPreviousLine

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 垂直方向のカーソル移動（C-p/C-n）

**説明:** 前の行・次の行へのカーソル移動機能の検証

**前提:** 2行のテキスト（"hello"、"world"）を入力済み

**操作:** C-p（前の行）、C-n（次の行）を順次実行

**結果:** カーソルが適切に上下の行を移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## スクロール機能 (scroll/enter_timing_issue)

### TestEnterKeyTimingIssue

**ファイル:** `test/enter_timing_test.go`

**シナリオ:** Enterキータイミング問題の検証

**説明:** 最後の可視行でEnterキーを押した際のスクロールタイミング問題の検証

**前提:** 40x10ディスプレイ（8コンテンツ行）でまず7行を作成

**操作:** 最後の可視行（行7）でEnterキーを押下

**結果:** カーソルが行8に移動し、即座に1行スクロールが発生する

**実装ファイル:** `domain/scroll.go`, `スクロールタイミング修正`

---

## events/queue_operations

### TestEventQueue

**ファイル:** `test/event_system_test.go`

**シナリオ:** イベントキューの基本操作

**説明:** イベントキューのPush/Pop操作の基本動作検証

**前提:** エディタを新規作成してイベントキューを取得

**操作:** KeyEventData('A')をキューにプッシュし、ポップする

**結果:** イベントが正しく取り出され、データが保持される

**実装ファイル:** `events/event_queue.go`

---

## keyboard/ctrl_x_prefix_reset

### TestCtrlXPrefixReset

**ファイル:** `test/keyboard_shortcuts_test.go`

**シナリオ:** C-x prefix key状態のリセット

**説明:** C-x後に無効なキーを押すとprefix状態がリセットされることの検証

**前提:** エディタが実行中の状態

**操作:** C-x後に通常の文字キーを送信する

**結果:** prefix状態がリセットされ、通常のテキスト入力として処理される

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## スクロール機能 (scroll/timing_verification)

### TestScrollStartsAtRightTime

**ファイル:** `test/realistic_scroll_test.go`

**シナリオ:** 異なるウィンドウサイズでのスクロールタイミング検証

**説明:** 複数のウィンドウサイズでスクロール開始タイミングの正確性を検証

**前提:** 異なるターミナル高（6、6、10、24）でテストケースを実行

**操作:** 各サイズでウィンドウ高まで行を追加し、さらに1行追加

**結果:** ウィンドウ高まではスクロールせず、超えた時点でスクロールが発生する

**実装ファイル:** `domain/scroll.go`, `サイズ別タイミング検証`

---

## cursor/japanese_position

### TestCursorPositionWithJapanese

**ファイル:** `test/cursor_position_test.go`

**シナリオ:** 日本語文字でのカーソル位置計算

**説明:** 日本語文字入力時のバイト位置とターミナル表示位置の正確な計算

**前提:** エディタを新規作成する

**操作:** "あいう"（日本語ひらがな3文字）を入力

**結果:** バイト位置が9（3文字 × 3バイト）、ターミナル表示位置が6（3文字 × 2幅）になる

**実装ファイル:** `domain/cursor.go`, `UTF-8処理`

---

## 画面表示機能 (display/basic_rendering)

### TestMockDisplayBasic

**ファイル:** `test/display_test.go`

**シナリオ:** 基本的なテキスト表示

**説明:** MockDisplayでの基本的なテキスト表示とカーソル位置の検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "hello"を入力する

**結果:** テキストが正確に表示され、カーソル位置が適切に設定される

**実装ファイル:** `test/mock_display.go`, `cli/display.go`

---

## エディタ基本機能 (editor/startup)

### TestEditorStartup

**ファイル:** `test/editor_startup_test.go`

**シナリオ:** エディタ初期化と基本状態の確認

**説明:** エディタ起動時の初期状態（バッファ、ウィンドウ、レンダリング）を検証

**前提:** エディタを新規作成する

**操作:** エディタの初期状態を確認する

**結果:** 実行中状態、*scratch*バッファ、ウィンドウが正しく設定される

**実装ファイル:** `domain/editor.go`, `domain/buffer.go`, `domain/window.go`

---

## キーボード入力機能 (input/basic_text)

### TestBasicTextInput

**ファイル:** `test/text_input_test.go`

**シナリオ:** 基本的なテキスト入力

**説明:** ASCII文字の連続入力と表示の検証

**前提:** エディタを新規作成する

**操作:** "Hello, World!"を1文字ずつ入力する

**結果:** 入力したテキストが正確に表示される

**実装ファイル:** `domain/buffer.go`, `domain/editor.go`

---

## スクロール機能 (scroll/cursor_movement_display)

### TestCursorMovementTriggersDisplay

**ファイル:** `test/auto_scroll_test.go`

**シナリオ:** 手動カーソル移動時の表示更新

**説明:** 手動でカーソルを移動した際の適切な表示更新

**前提:** 30x8ウィンドウに20行のコンテンツを作成

**操作:** カーソルを手動でバッファの先頭に移動

**結果:** ウィンドウがスクロールしてカーソルが表示される

**実装ファイル:** `domain/scroll.go`, `domain/cursor.go`

---

## cursor/japanese_progression

### TestCursorPositionProgression

**ファイル:** `test/cursor_position_test.go`

**シナリオ:** 日本語文字連続入力時のカーソル進行

**説明:** 日本語文字を連続して入力した際のカーソル位置の步進的進行

**前提:** エディタを新規作成する

**操作:** 日本語文字（あ、い、う、え、お）を1文字ずつ順次入力

**結果:** 各文字の入力後にバイト位置とターミナル表示位置が正確に進行する

**実装ファイル:** `domain/cursor.go`, `文字幅計算`

---

## events/resize_handling

### TestResizeEvent

**ファイル:** `test/event_system_test.go`

**シナリオ:** リサイズイベントの処理

**説明:** ターミナルリサイズイベントの処理とウィンドウサイズ更新

**前提:** エディタを新規作成する

**操作:** 100x30サイズのリサイズイベントを送信

**結果:** ウィンドウサイズが100x28（モードラインとミニバッファを除いたサイズ）に更新される

**実装ファイル:** `events/resize_event.go`, `domain/window.go`

---

## keyboard/ctrl_modifier_no_insert

### TestCtrlModifierDoesNotInsertText

**ファイル:** `test/keyboard_shortcuts_test.go`

**シナリオ:** Ctrl修飾キーのテキスト非挿入

**説明:** Ctrl+文字キーの組み合わせでテキストが挿入されないことの検証

**前提:** エディタを新規作成する

**操作:** Ctrl+aキーイベントを送信する

**結果:** テキストが挿入されず、空の行が維持される

**実装ファイル:** `domain/editor.go`, `events/key_event.go`

---

## commands/mx_version

### TestMxVersionCommand

**ファイル:** `test/mx_command_test.go`

**シナリオ:** M-x versionコマンドの実行

**説明:** M-x versionコマンドでバージョン情報を表示

**前提:** M-xコマンドモードを有効化

**操作:** "version"を入力してEnterキーを押下

**結果:** バージョンメッセージがミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `domain/minibuffer.go`

---

## キーボード入力機能 (input/newline_split)

### TestNewlineInMiddle

**ファイル:** `test/newline_test.go`

**シナリオ:** 行の中間での改行挿入

**説明:** 行の中間でEnterキーを押した際の行分割動作

**前提:** "hello world"を入力済みでカーソルを"hello"の後（位置5）に移動

**操作:** カーソル位置でEnterキーを押下

**結果:** 行が"hello"と" world"に分割され、カーソルが2行目の先頭に移動する

**実装ファイル:** `domain/buffer.go`, `行分割処理`

---

## スクロール機能 (scroll/horizontal_scrolling)

### TestHorizontalScrolling

**ファイル:** `test/scrolling_test.go`

**シナリオ:** 水平スクロール動作

**説明:** 長い行のコンテンツでの水平スクロール動作の検証

**前提:** 10x5の狭いウィンドウと長い行のコンテンツ

**操作:** 行ラップを無効化して水平スクロールを設定

**結果:** 指定した位置からコンテンツが表示される

**実装ファイル:** `domain/window.go`, `水平スクロール`

---

## cursor/line_wrap_position

### TestCursorPositionWithLineWrapping

**ファイル:** `test/line_wrapping_cursor_test.go`

**シナリオ:** 行ラップ有効時のカーソル位置

**説明:** 行ラップ有効時の長い行でのカーソル位置計算と表示

**前提:** 10x8の小さいウィンドウで行ラップ有効

**操作:** ウィンドウ幅を超える長い行を入力し、カーソルを移動

**結果:** ラップされた行の境界でカーソル位置が正確に計算される

**実装ファイル:** `domain/cursor.go`, `行ラップ処理`

---

## resize/terminal_resize

### TestTerminalResize

**ファイル:** `test/resize_test.go`

**シナリオ:** ターミナルリサイズ処理

**説明:** ターミナルサイズ変更時のウィンドウサイズ更新とコンテンツ保持

**前提:** 80x24サイズのターミナルで"hello world"を入力済み

**操作:** ターミナルを120x30にリサイズする

**結果:** ウィンドウサイズが更新され、コンテンツが保持される

**実装ファイル:** `domain/window.go`, `events/resize_event.go`

---

## キーボード入力機能 (input/japanese)

### TestJapaneseTextInput

**ファイル:** `test/text_input_test.go`

**シナリオ:** 日本語テキスト入力

**説明:** ひらがな文字の入力と表示の検証

**前提:** エディタを新規作成する

**操作:** "あいう"を文字ごとに入力する

**結果:** 日本語テキストが正確に表示される

**実装ファイル:** `domain/buffer.go`, `UTF-8処理`

---

## cursor/forward_char

### TestForwardCharBasic

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 前方向文字移動（C-f）

**説明:** カーソルを1文字右に移動する機能の検証

**前提:** "hello"を入力済みでカーソルを行頭に設定

**操作:** C-f（forward-char）コマンドを実行

**結果:** カーソルが1文字右に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## キーボード入力機能 (input/newline_beginning)

### TestNewlineAtBeginning

**ファイル:** `test/newline_test.go`

**シナリオ:** 行頭での改行挿入

**説明:** 行頭でEnterキーを押した際の新しい行挿入動作

**前提:** "hello"を入力済みでカーソルを行頭に移動

**操作:** 行頭でEnterキーを押下

**結果:** 空の新しい行が挿入され、既存のコンテンツが2行目に移動する

**実装ファイル:** `domain/buffer.go`, `行挿入処理`

---

## resize/smaller_size_resize

### TestResizeToSmallerSize

**ファイル:** `test/resize_test.go`

**シナリオ:** 小さいサイズへのリサイズ

**説明:** ターミナルを小さいサイズにリサイズした際のコンテンツ保持

**前提:** 80x24サイズで複数行のコンテンツを入力済み

**操作:** ターミナルのサイズを40x10に縮小する

**結果:** ウィンドウサイズが更新され、バッファの全コンテンツが保持される

**実装ファイル:** `domain/window.go`, `domain/buffer.go`

---

## キーボード入力機能 (input/newline_basic)

### TestNewlineBasic

**ファイル:** `test/newline_test.go`

**シナリオ:** 基本的な改行挿入

**説明:** 行末でのEnterキーによる基本的な改行動作

**前提:** 空のバッファに"hello"を入力済み

**操作:** 行末でEnterキーを押し、"world"を入力

**結果:** 2行に分かれてテキストが配置され、カーソルが適切な位置に移動する

**実装ファイル:** `domain/buffer.go`, `events/key_event.go`

---

## cursor/wrapped_line_movement

### TestCursorMovementAcrossWrappedLines

**ファイル:** `test/line_wrapping_cursor_test.go`

**シナリオ:** ラップされた行をまたいだカーソル移動

**説明:** ラップされた行の境界を跨いだカーソル移動の検証

**前提:** 10x8ウィンドウでラップするコンテンツを作成

**操作:** 行頭に移動し、forward-charで一文字ずつ進む

**結果:** ラップ境界でスクリーンカーソル位置が正しく更新される

**実装ファイル:** `domain/cursor.go`, `ラップ境界処理`

---

## commands/toggle_line_wrap

### TestWrappingToggleCommand

**ファイル:** `test/line_wrapping_cursor_test.go`

**シナリオ:** 行ラップトグルコマンドの実行

**説明:** M-x toggle-truncate-linesコマンドでの行ラップ状態切り替え

**前提:** エディタを新規作成（デフォルトでラップ有効）

**操作:** ToggleLineWrap関数とM-x toggle-truncate-linesコマンドを実行

**結果:** 行ラップ状態が適切に切り替わり、コマンドが正しく動作する

**実装ファイル:** `domain/commands.go`, `コマンド処理`

---

## commands/mx_basic

### TestMxCommandBasic

**ファイル:** `test/mx_command_test.go`

**シナリオ:** M-xコマンドの基本動作

**説明:** M-xコマンドモードの有効化とミニバッファ状態の確認

**前提:** エディタを新規作成し、通常モードで起動

**操作:** ESCキーを押し、続いてxキーを押下（M-x）

**結果:** ミニバッファがアクティブになり、"M-x "プロンプトが表示される

**実装ファイル:** `domain/commands.go`, `domain/minibuffer.go`

---

## スクロール機能 (scroll/auto_scroll_insertion)

### TestAutoScrollOnTextInsertion

**ファイル:** `test/auto_scroll_test.go`

**シナリオ:** テキスト挿入時の自動スクロール

**説明:** 可視範囲を超えるテキスト挿入時のスクロール動作

**前提:** 30x6の小さいウィンドウ（4コンテンツ行）に3行の初期コンテンツ

**操作:** さらに5行の新しいコンテンツを追加

**結果:** スクロールが発生し、カーソルが可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## cursor/mixed_ascii_japanese

### TestMixedASCIIJapaneseCursor

**ファイル:** `test/cursor_position_test.go`

**シナリオ:** ASCIIと日本語混在カーソル位置

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル位置計算

**前提:** エディタを新規作成する

**操作:** "aあiい"（ASCIIと日本語の混在）を順次入力

**結果:** 各文字タイプのバイト数と表示幅の違いを正確に処理してカーソル位置が計算される

**実装ファイル:** `domain/cursor.go`, `混合文字列処理`

---

## 画面表示機能 (display/mixed_character_cursor)

### TestMockDisplayCursorProgression

**ファイル:** `test/display_test.go`

**シナリオ:** ASCII+日本語混在カーソル進行

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル位置進行

**前提:** 20x5サイズのMockDisplayを作成

**操作:** 'a'、'あ'、'b'、'い'、'c'を順次入力

**結果:** マルチバイト文字の表示幅を考慮してカーソルが適切に進行する

**実装ファイル:** `test/mock_display.go`, `文字幅計算`

---

## keyboard/ctrl_c_no_quit

### TestCtrlCAloneDoesNotQuit

**ファイル:** `test/keyboard_shortcuts_test.go`

**シナリオ:** C-c単独ではエディタ終了しない

**説明:** C-x prefix key なしのC-cではエディタが終了しないことの検証

**前提:** エディタが実行中の状態

**操作:** C-cキーイベントのみを送信する

**結果:** エディタが実行中のまま維持される

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## commands/mx_clear_buffer

### TestMxClearBuffer

**ファイル:** `test/mx_command_test.go`

**シナリオ:** M-x clear-bufferコマンドの実行

**説明:** バッファの内容を全てクリアする機能

**前提:** バッファに"hello world"を入力済み

**操作:** M-x clear-bufferコマンドを実行

**結果:** バッファが空になり、クリアメッセージが表示される

**実装ファイル:** `domain/commands.go`, `domain/buffer.go`

---

## text/rune_width_calculation

### TestRuneWidth

**ファイル:** `test/runewidth_test.go`

**シナリオ:** 文字幅計算機能

**説明:** Unicode文字（ASCII、日本語、制御文字）の表示幅計算

**前提:** ASCII文字、日本語文字、制御文字のテストケース

**操作:** RuneWidth関数で各文字の表示幅を計算

**結果:** ASCII文字は幅1、日本語文字は幅2、制御文字は幅0で計算される

**実装ファイル:** `util/runewidth.go`, `文字幅計算`

---

## スクロール機能 (scroll/auto_scroll_lines)

### TestAutoScrollWhenAddingLines

**ファイル:** `test/auto_scroll_test.go`

**シナリオ:** 行追加時の自動スクロール

**説明:** ウィンドウ高を超える行を追加した際の自動スクロール動作

**前提:** 40x10サイズのディスプレイ（8コンテンツ行）

**操作:** 15行のコンテンツを順次追加する

**結果:** カーソルが常に可視範囲内に保たれ、現在の行が表示される

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## keyboard/meta_modifier_no_insert

### TestMetaModifierDoesNotInsertText

**ファイル:** `test/keyboard_shortcuts_test.go`

**シナリオ:** Meta修飾キーのテキスト非挿入

**説明:** Meta+文字キーの組み合わせでテキストが挿入されないことの検証

**前提:** エディタを新規作成する

**操作:** Meta+xキーイベントを送信する

**結果:** テキストが挿入されず、空の行が維持される

**実装ファイル:** `domain/editor.go`, `events/key_event.go`

---

## スクロール機能 (scroll/page_navigation)

### TestPageUpDown

**ファイル:** `test/scrolling_test.go`

**シナリオ:** ページアップ/ダウンナビゲーション

**説明:** PageUp/PageDownコマンドによるページ単位のスクロール

**前提:** 50行の大量コンテンツを持つエディタ

**操作:** PageDown、PageUpコマンドを順次実行

**結果:** スクロール位置がページ単位で適切に変更される

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## スクロール機能 (scroll/scroll_timing)

### TestTerminal12LinesScenario

**ファイル:** `test/terminal_12_lines_test.go`

**シナリオ:** 早すぎるスクロールの回避

**説明:** コンテンツがウィンドウコンテンツエリアを真に超えるまでスクロールが発生しないことをテスト

**前提:** 12行のターミナル（10コンテンツ + モード + ミニ）

**操作:** 文字a〜jをそれぞれEnterで区切って入力する

**結果:** すべての10行がスクロールなしで表示される

**実装ファイル:** `domain/scroll.go`, `cli/display.go`

---

### TestTerminal12LinesDebugSteps

**ファイル:** `test/terminal_12_lines_test.go`

**シナリオ:** コンテンツがウィンドウを超えた時のスクロール

**説明:** スクロール動作をステップごとに検証するデバッグテスト

**前提:** 12行のコンテンツエリアを持つターミナル

**操作:** コンテンツエリア限界を超えて一行ずつ追加する

**結果:** 適切なタイミングでスクロールが発生する

**実装ファイル:** `domain/scroll.go`, `cli/display.go`

---

## スクロール機能 (scroll/exact_user_scenario)

### TestExactUserScenario

**ファイル:** `test/exact_user_scenario_test.go`

**シナリオ:** ユーザー報告の正確なシナリオ再現

**説明:** 高さ10ターミナルでa〜hまで入力後のEnter時のスクロール動作

**前提:** 高さ10ターミナル（コンテンツエリア8行）でリサイズイベントを発生

**操作:** a + Enter + b + ... + h を入力し、最後にEnterを押下

**結果:** a〜hが表示され、Enter後はb〜h+空行が表示される

**実装ファイル:** `domain/scroll.go`, `ユーザーシナリオ修正`

---

## commands/mx_cancel

### TestMxCancel

**ファイル:** `test/mx_command_test.go`

**シナリオ:** M-xコマンドのキャンセル

**説明:** ESCキーでM-xコマンドをキャンセルする機能

**前提:** M-xコマンドモードで部分的にコマンドを入力済み

**操作:** ESCキーを押してキャンセルする

**結果:** ミニバッファがクリアされ、通常モードに戻る

**実装ファイル:** `domain/commands.go`, `キャンセル処理`

---

## 画面表示機能 (display/layout_analysis)

### TestDisplayLayoutAnalysis

**ファイル:** `test/display_layout_test.go`

**シナリオ:** 表示レイアウト解析

**説明:** 実際の表示レイアウトと期待されるレイアウトの比較分析

**前提:** 40x10ターミナルでリサイズイベントを送信

**操作:** ウィンドウ高と同じ数の行を追加し、さらに1行追加

**結果:** MockDisplayと実際のCLI Displayの動作が一致し、適切なスクロールタイミングが確認される

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

---

## スクロール機能 (scroll/step_by_step_debug)

### TestUserScenarioStepByStep

**ファイル:** `test/exact_user_scenario_test.go`

**シナリオ:** ユーザーシナリオのステップバイステップデバッグ

**説明:** ユーザー報告シナリオをステップごとに詳細に検証するデバッグテスト

**前提:** 40x10ディスプレイでウィンドウサイズを設定

**操作:** a〜hをステップごとに入力し、各ステップで状態をログ出力

**結果:** 各ステップでカーソル位置とスクロール状態が正しく、最終的に期待結果を得る

**実装ファイル:** `domain/scroll.go`, `デバッグ情報出力`

---

## commands/mx_list_commands

### TestMxListCommands

**ファイル:** `test/mx_command_test.go`

**シナリオ:** M-x list-commandsコマンドの実行

**説明:** 利用可能なコマンド一覧を表示する機能

**前提:** M-xコマンドモードを有効化

**操作:** "list-commands"を入力してEnterキーを押下

**結果:** 利用可能なコマンド一覧がミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `コマンド一覧機能`

---

## resize/cursor_position_preservation

### TestCursorPositionAfterResize

**ファイル:** `test/resize_test.go`

**シナリオ:** リサイズ後のカーソル位置保持

**説明:** ターミナルリサイズ後のカーソル位置保持の検証

**前提:** "hello"を入力しカーソルを中央（位置2）に設定

**操作:** ターミナルを120x30にリサイズする

**結果:** カーソル位置がリサイズ後も(0,2)で保持される

**実装ファイル:** `domain/window.go`, `domain/cursor.go`

---

## スクロール機能 (scroll/individual_scroll_commands)

### TestScrollCommands

**ファイル:** `test/scrolling_test.go`

**シナリオ:** 個別スクロールコマンド

**説明:** ScrollUp/ScrollDownコマンドによる1行単位のスクロール

**前提:** 30行のコンテンツを持つエディタ

**操作:** ScrollDown、ScrollUpコマンドを順次実行

**結果:** スクロール位置が1行単位で正確に変更される

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## キーボード入力機能 (input/multiline)

### TestMultilineTextInput

**ファイル:** `test/text_input_test.go`

**シナリオ:** 複数行テキスト入力

**説明:** 3行のテキストを順次入力し、行分離を検証

**前提:** エディタを新規作成する

**操作:** "First line", "Second line", "Third line"を Enter で区切って入力する

**結果:** 3行が正確に分かれて表示される

**実装ファイル:** `domain/buffer.go`, `domain/editor.go`

---

## スクロール機能 (scroll/auto_scroll_wrapping)

### TestAutoScrollWithLongLines

**ファイル:** `test/auto_scroll_test.go`

**シナリオ:** 長い行での自動スクロールと行ラップ

**説明:** 行ラップ有効時の長い行での自動スクロール動作

**前提:** 20x8の小さいウィンドウで行ラップ有効

**操作:** 短い行と長い行（ラップする）を混在して追加

**結果:** カーソルが常に可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## cursor/backward_char

### TestBackwardCharBasic

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 後方向文字移動（C-b）

**説明:** カーソルを1文字左に移動する機能の検証

**前提:** "hello"を入力済みでカーソルが行末にある

**操作:** C-b（backward-char）コマンドを実行

**結果:** カーソルが1文字左に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## スクロール機能 (scroll/user_reported_behavior)

### TestUserReportedBehavior

**ファイル:** `test/enter_timing_test.go`

**シナリオ:** ユーザー報告された問題の再現

**説明:** ユーザーが報告したスクロールディレイの正確な再現テスト

**前提:** 8行でスクリーンを埋めた状態

**操作:** 連続してEnter+コンテンツ入力を繰り返す

**結果:** ユーザー期待と実際の動作の違いを特定し、修正を検証する

**実装ファイル:** `domain/scroll.go`, `ユーザー報告修正`

---

## text/partial_string_width

### TestStringWidthUpTo

**ファイル:** `test/runewidth_test.go`

**シナリオ:** 部分文字列幅計算機能

**説明:** 指定バイト位置までの文字列の表示幅計算

**前提:** ASCII文字列、日本語文字列、混合文字列と様々なバイト位置

**操作:** StringWidthUpTo関数で指定位置までの表示幅を計算

**結果:** マルチバイト文字の境界を考慮した正確な部分幅が計算される

**実装ファイル:** `util/runewidth.go`, `部分文字列幅計算`

---

## cursor/arrow_keys

### TestArrowKeys

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 矢印キーによるカーソル移動

**説明:** 左右矢印キーでのカーソル移動機能の検証

**前提:** "hello"を入力済みでカーソルを行頭に設定

**操作:** 右矢印キー、左矢印キーを順次押下

**結果:** カーソルが適切に左右に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/line_boundaries

### TestBeginningEndOfLine

**ファイル:** `test/cursor_movement_test.go`

**シナリオ:** 行頭・行末移動（C-a/C-e）

**説明:** 行の先頭と末尾への移動機能の検証

**前提:** "hello world"を入力済みでカーソルが行末にある

**操作:** C-a（行頭）、C-e（行末）を順次実行

**結果:** カーソルが行頭と行末に適切に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## events/queue_capacity

### TestEventQueueCapacity

**ファイル:** `test/event_system_test.go`

**シナリオ:** イベントキューの容量制限

**説明:** イベントキューの容量制限とオーバーフロー処理

**前提:** 容量2のイベントキューを作成

**操作:** 3つのイベント（A、B、C）を順次プッシュ

**結果:** 最初の2つのイベント（A、B）のみが保持され、3番目（C）は破棄される

**実装ファイル:** `events/event_queue.go`, `容量制限処理`

---

## スクロール機能 (scroll/realistic_terminal)

### TestRealisticTerminalScroll

**ファイル:** `test/realistic_scroll_test.go`

**シナリオ:** リアルなターミナルサイズでのスクロール

**説明:** 80x24のリアルなターミナルサイズでのスクロール動作検証

**前提:** 80x24ターミナル（22コンテンツ行）でリサイズイベントを送信

**操作:** 30行のコンテンツを順次追加し、各ステップでスクロール状態を監視

**結果:** ウィンドウ高を超えたタイミングでスクロールが開始され、カーソルが常に可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `リアルターミナル環境`

---

## terminal/width_calculation

### TestTerminalWidthIssue

**ファイル:** `test/terminal_width_test.go`

**シナリオ:** ターミナル幅計算問題の検証

**説明:** ASCII文字と日本語文字の混合テキストでのターミナル表示位置計算

**前提:** 20x3のMockDisplayと様々な文字組み合わせのテストケース

**操作:** 各テストケースで文字を入力し、カーソル位置を取得

**結果:** ASCII文字は1列、日本語文字は2列、混合テキストは合計列数で正確に表示される

**実装ファイル:** `test/mock_display.go`, `ターミナル幅計算処理`

---

*このドキュメントは自動生成されています。修正はテストファイルのアノテーションを編集してください。*
