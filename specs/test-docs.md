# gmacs テストドキュメント

このドキュメントはテストコードから自動抽出されたBDD仕様書です。

**生成日時:** 2025年06月20日 21:25:06

## application/clean_exit

### TestCleanExit

**ファイル:** `e2e-test/clean_exit_test.go`

**シナリオ:** C-x C-c による正常終了

**説明:** C-x C-c コマンドでエディタが正常に終了する機能

**前提:** エディタが実行中の状態

**操作:** C-x C-c キーシーケンスを実行

**結果:** エディタが終了状態になる

**実装ファイル:** `domain/command.go`, `Quit関数`

---

## application/exit_during_input

### TestExitDuringInput

**ファイル:** `e2e-test/clean_exit_test.go`

**シナリオ:** 入力中の終了

**説明:** ミニバッファ入力中にC-x C-cで終了する場合

**前提:** M-xコマンド入力中の状態

**操作:** C-x C-c キーシーケンスを実行

**結果:** ミニバッファがクリアされずに終了する（通常の終了が優先される）

**実装ファイル:** `domain/editor.go`, `キー処理優先順位`

---

## application/signal_exit

### TestSignalExit

**ファイル:** `e2e-test/clean_exit_test.go`

**シナリオ:** シグナルによる終了

**説明:** SIGINTやSIGTERMシグナルでの終了処理

**前提:** エディタが実行中の状態

**操作:** QuitEventDataを受信

**結果:** エディタが終了状態になる

**実装ファイル:** `events/quit_event.go`, `domain/editor.go`

---

## バッファ管理機能 (buffer/kill_buffer_basic)

### TestKillBufferBasic

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** C-x kによる基本的なバッファ削除

**説明:** C-x kキーシーケンスで現在のバッファを削除する機能

**前提:** エディタに複数のバッファが存在し、任意のバッファを選択中

**操作:** C-xを押し、続いてkキーを押下する

**結果:** 現在のバッファが削除され、他のバッファに切り替わる

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/kill_buffer_last)

### TestKillBufferLast

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** 最後のバッファ削除の防止

**説明:** 最後の1つのバッファを削除しようとした場合のエラー処理

**前提:** エディタに1つのバッファのみ存在している状態

**操作:** C-x kキーシーケンスでバッファ削除を試行

**結果:** 削除が拒否され、エラーメッセージが表示される

**実装ファイル:** `domain/buffer_interactive.go`, `エラー処理`

---

## バッファ管理機能 (buffer/list_buffers_basic)

### TestListBuffersBasic

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** C-x C-bによるバッファ一覧表示

**説明:** C-x C-bキーシーケンスでバッファ一覧を表示する機能

**前提:** エディタに複数のバッファが存在している状態

**操作:** C-xを押し、続いてC-bキーを押下する

**結果:** ミニバッファにバッファ一覧と現在のバッファが表示される

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/minibuffer_editing)

### TestBufferMinibufferEditing

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** バッファ選択モードでのミニバッファ編集

**説明:** バッファ選択モードでのカーソル移動と編集機能

**前提:** C-x bでバッファ選択モードを開始し、バッファ名を部分入力済み

**操作:** C-f, C-b, C-a, C-e, C-h, C-dキーで編集操作を実行

**結果:** ミニバッファ内でカーソル移動と文字削除が正常に動作する

**実装ファイル:** `domain/buffer_interactive.go`, `ミニバッファ編集`

---

## バッファ管理機能 (buffer/mx_commands)

### TestBufferMxCommands

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** M-xコマンドによるバッファ操作

**説明:** M-xコマンドでバッファ関連の操作を実行する機能

**前提:** エディタに複数のバッファが存在している状態

**操作:** M-x switch-to-buffer, M-x list-buffers, M-x kill-bufferを実行

**結果:** キーバインドと同等の動作が実行される

**実装ファイル:** `domain/buffer_interactive.go`, `M-xコマンドシステム`

---

## バッファ管理機能 (buffer/switch_to_buffer_basic)

### TestSwitchToBufferBasic

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** C-x bによる基本的なバッファ切り替え

**説明:** C-x bキーシーケンスでバッファ切り替えモードを開始する機能

**前提:** エディタに複数のバッファが存在している状態

**操作:** C-xを押し、続いてbキーを押下する

**結果:** ミニバッファがアクティブになり、"Switch to buffer: "プロンプトが表示される

**実装ファイル:** `domain/buffer_interactive.go`, `domain/editor.go`

---

## バッファ管理機能 (buffer/switch_to_buffer_cancel)

### TestSwitchToBufferCancel

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** C-x bのキャンセル機能

**説明:** Escapeキーでバッファ切り替えをキャンセルする機能

**前提:** C-x bでバッファ切り替えモードを開始し、部分的に名前を入力済み

**操作:** Escapeキーを押下

**結果:** バッファ切り替えがキャンセルされ、ミニバッファがクリアされる

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/switch_to_buffer_empty)

### TestSwitchToBufferEmpty

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** 空文字入力でのバッファ切り替えキャンセル

**説明:** バッファ名を入力せずにEnterを押した場合の動作

**前提:** C-x bでバッファ切り替えモードを開始済み

**操作:** 何も入力せずにEnterキーを押下

**結果:** 現在のバッファのまま変更されず、ミニバッファがクリアされる

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/switch_to_buffer_existing)

### TestSwitchToBufferExisting

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** 既存バッファへの切り替え

**説明:** 存在するバッファ名を入力してバッファを切り替える機能

**前提:** C-x bでバッファ切り替えモードを開始済み

**操作:** 既存のバッファ名"test-buffer"を入力してEnterキーを押下

**結果:** 指定したバッファに切り替わり、成功メッセージが表示される

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/switch_to_buffer_new)

### TestSwitchToBufferNew

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** 新規バッファの作成と切り替え

**説明:** 存在しないバッファ名を入力して新しいバッファを作成する機能

**前提:** C-x bでバッファ切り替えモードを開始済み

**操作:** 存在しないバッファ名"new-buffer"を入力してEnterキーを押下

**結果:** 新しいバッファが作成され、そのバッファに切り替わる

**実装ファイル:** `domain/buffer_interactive.go`

---

## バッファ管理機能 (buffer/tab_completion_multiple)

### TestBufferTabCompletionMultiple

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** バッファ名の複数マッチ補完

**説明:** Tabキーによるバッファ名の自動補完機能（複数マッチ）

**前提:** C-x bでバッファ切り替えモード開始し、複数にマッチする部分文字列を入力済み

**操作:** Tabキーを押下

**結果:** 共通部分まで補完され、マッチした候補一覧が表示される

**実装ファイル:** `domain/buffer_interactive.go`, `補完機能`

---

## バッファ管理機能 (buffer/tab_completion_single)

### TestBufferTabCompletionSingle

**ファイル:** `e2e-test/buffer_interactive_test.go`

**シナリオ:** バッファ名の単一マッチ補完

**説明:** Tabキーによるバッファ名の自動補完機能（単一マッチ）

**前提:** C-x bでバッファ切り替えモード開始し、一意に決まる部分文字列を入力済み

**操作:** Tabキーを押下

**結果:** バッファ名が自動的に完全な名前まで補完される

**実装ファイル:** `domain/buffer_interactive.go`, `補完機能`

---

## commands/mx_basic

### TestMxCommandBasic

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** M-xコマンドの基本動作

**説明:** M-xコマンドモードの有効化とミニバッファ状態の確認

**前提:** エディタを新規作成し、通常モードで起動

**操作:** ESCキーを押し、続いてxキーを押下（M-x）

**結果:** ミニバッファがアクティブになり、"M-x "プロンプトが表示される

**実装ファイル:** `domain/commands.go`, `domain/minibuffer.go`

---

## commands/mx_cancel

### TestMxCancel

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** M-xコマンドのキャンセル

**説明:** ESCキーでM-xコマンドをキャンセルする機能

**前提:** M-xコマンドモードで部分的にコマンドを入力済み

**操作:** ESCキーを押してキャンセルする

**結果:** ミニバッファがクリアされ、通常モードに戻る

**実装ファイル:** `domain/commands.go`, `キャンセル処理`

---

## commands/mx_clear_buffer

### TestMxClearBuffer

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** M-x clear-bufferコマンドの実行

**説明:** バッファの内容を全てクリアする機能

**前提:** バッファに"hello world"を入力済み

**操作:** M-x clear-bufferコマンドを実行

**結果:** バッファが空になり、クリアメッセージが表示される

**実装ファイル:** `domain/commands.go`, `domain/buffer.go`

---

## commands/mx_list_commands

### TestMxListCommands

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** M-x list-commandsコマンドの実行

**説明:** 利用可能なコマンド一覧を表示する機能

**前提:** M-xコマンドモードを有効化

**操作:** "list-commands"を入力してEnterキーを押下

**結果:** 利用可能なコマンド一覧がミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `コマンド一覧機能`

---

## commands/mx_unknown

### TestMxUnknownCommand

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** 未知のM-xコマンドのエラー処理

**説明:** 存在しないコマンドを実行した際のエラーハンドリング

**前提:** M-xコマンドモードを有効化

**操作:** 存在しないコマンド"nonexistent"を入力してEnterを押下

**結果:** エラーメッセージがミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `エラー処理`

---

## commands/mx_version

### TestMxVersionCommand

**ファイル:** `e2e-test/mx_command_test.go`

**シナリオ:** M-x versionコマンドの実行

**説明:** M-x versionコマンドでバージョン情報を表示

**前提:** M-xコマンドモードを有効化

**操作:** "version"を入力してEnterキーを押下

**結果:** バージョンメッセージがミニバッファに表示される

**実装ファイル:** `domain/commands.go`, `domain/minibuffer.go`

---

## commands/toggle_line_wrap

### TestWrappingToggleCommand

**ファイル:** `e2e-test/line_wrapping_cursor_test.go`

**シナリオ:** 行ラップトグルコマンドの実行

**説明:** M-x toggle-truncate-linesコマンドでの行ラップ状態切り替え

**前提:** エディタを新規作成（デフォルトでラップ有効）

**操作:** ToggleLineWrap関数とM-x toggle-truncate-linesコマンドを実行

**結果:** 行ラップ状態が適切に切り替わり、コマンドが正しく動作する

**実装ファイル:** `domain/commands.go`, `コマンド処理`

---

## cursor/arrow_keys

### TestArrowKeys

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 矢印キーによるカーソル移動

**説明:** 左右矢印キーでのカーソル移動機能の検証

**前提:** "hello"を入力済みでカーソルを行頭に設定

**操作:** 右矢印キー、左矢印キーを順次押下

**結果:** カーソルが適切に左右に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/backward_char

### TestBackwardCharBasic

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 後方向文字移動（C-b）

**説明:** カーソルを1文字左に移動する機能の検証

**前提:** "hello"を入力済みでカーソルが行末にある

**操作:** C-b（backward-char）コマンドを実行

**結果:** カーソルが1文字左に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/forward_char

### TestForwardCharBasic

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 前方向文字移動（C-f）

**説明:** カーソルを1文字右に移動する機能の検証

**前提:** "hello"を入力済みでカーソルを行頭に設定

**操作:** C-f（forward-char）コマンドを実行

**結果:** カーソルが1文字右に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/japanese_position

### TestCursorPositionWithJapanese

**ファイル:** `e2e-test/cursor_position_test.go`

**シナリオ:** 日本語文字でのカーソル位置計算

**説明:** 日本語文字入力時のバイト位置とターミナル表示位置の正確な計算

**前提:** エディタを新規作成する

**操作:** "あいう"（日本語ひらがな3文字）を入力

**結果:** バイト位置が9（3文字 × 3バイト）、ターミナル表示位置が6（3文字 × 2幅）になる

**実装ファイル:** `domain/cursor.go`, `UTF-8処理`

---

## cursor/japanese_progression

### TestCursorPositionProgression

**ファイル:** `e2e-test/cursor_position_test.go`

**シナリオ:** 日本語文字連続入力時のカーソル進行

**説明:** 日本語文字を連続して入力した際のカーソル位置の步進的進行

**前提:** エディタを新規作成する

**操作:** 日本語文字（あ、い、う、え、お）を1文字ずつ順次入力

**結果:** 各文字の入力後にバイト位置とターミナル表示位置が正確に進行する

**実装ファイル:** `domain/cursor.go`, `文字幅計算`

---

## cursor/japanese_support

### TestCursorMovementWithJapanese

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 日本語文字を含むカーソル移動

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル移動

**前提:** "aあbいc"（ASCII+日本語混在）を入力済み

**操作:** C-fで1文字ずつ前進する

**結果:** マルチバイト文字を適切に処理してカーソルが移動する

**実装ファイル:** `domain/cursor.go`, `UTF-8処理`

---

## cursor/line_boundaries

### TestBeginningEndOfLine

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 行頭・行末移動（C-a/C-e）

**説明:** 行の先頭と末尾への移動機能の検証

**前提:** "hello world"を入力済みでカーソルが行末にある

**操作:** C-a（行頭）、C-e（行末）を順次実行

**結果:** カーソルが行頭と行末に適切に移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/line_wrap_position

### TestCursorPositionWithLineWrapping

**ファイル:** `e2e-test/line_wrapping_cursor_test.go`

**シナリオ:** 行ラップ有効時のカーソル位置

**説明:** 行ラップ有効時の長い行でのカーソル位置計算と表示

**前提:** 10x8の小さいウィンドウで行ラップ有効

**操作:** ウィンドウ幅を超える長い行を入力し、カーソルを移動

**結果:** ラップされた行の境界でカーソル位置が正確に計算される

**実装ファイル:** `domain/cursor.go`, `行ラップ処理`

---

## cursor/mixed_ascii_japanese

### TestMixedASCIIJapaneseCursor

**ファイル:** `e2e-test/cursor_position_test.go`

**シナリオ:** ASCIIと日本語混在カーソル位置

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル位置計算

**前提:** エディタを新規作成する

**操作:** "aあiい"（ASCIIと日本語の混在）を順次入力

**結果:** 各文字タイプのバイト数と表示幅の違いを正確に処理してカーソル位置が計算される

**実装ファイル:** `domain/cursor.go`, `混合文字列処理`

---

## cursor/mx_commands

### TestInteractiveCommands

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** M-xコマンドによるカーソル移動

**説明:** M-x beginning-of-lineコマンドの実行検証

**前提:** "hello"を入力済みでカーソルが行末にある

**操作:** M-x beginning-of-lineコマンドを実行

**結果:** カーソルが行頭に移動する

**実装ファイル:** `domain/commands.go`, `events/key_event.go`

---

## cursor/vertical_movement

### TestNextPreviousLine

**ファイル:** `e2e-test/cursor_movement_test.go`

**シナリオ:** 垂直方向のカーソル移動（C-p/C-n）

**説明:** 前の行・次の行へのカーソル移動機能の検証

**前提:** 2行のテキスト（"hello"、"world"）を入力済み

**操作:** C-p（前の行）、C-n（次の行）を順次実行

**結果:** カーソルが適切に上下の行を移動する

**実装ファイル:** `domain/cursor.go`, `events/key_event.go`

---

## cursor/wrapped_line_movement

### TestCursorMovementAcrossWrappedLines

**ファイル:** `e2e-test/line_wrapping_cursor_test.go`

**シナリオ:** ラップされた行をまたいだカーソル移動

**説明:** ラップされた行の境界を跨いだカーソル移動の検証

**前提:** 10x8ウィンドウでラップするコンテンツを作成

**操作:** 行頭に移動し、forward-charで一文字ずつ進む

**結果:** ラップ境界でスクリーンカーソル位置が正しく更新される

**実装ファイル:** `domain/cursor.go`, `ラップ境界処理`

---

## delete/backward_char_basic

### TestDeleteBackwardCharBasic

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** C-h による基本的な文字削除

**説明:** カーソル前の文字を削除する基本的な backspace 機能

**前提:** "hello"を入力済みでカーソルが行末にある

**操作:** C-h（DeleteBackwardChar）コマンドを実行

**結果:** 最後の文字が削除され"hell"になる

**実装ファイル:** `domain/buffer.go`, `DeleteBackward関数`

---

## delete/backward_char_japanese

### TestDeleteBackwardCharJapanese

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** 日本語文字のbackspace削除

**説明:** 日本語文字（マルチバイト）のbackspace削除機能

**前提:** "aあiい"を入力済みでカーソルが行末にある

**操作:** C-h（DeleteBackwardChar）コマンドを実行

**結果:** 最後の日本語文字が削除され"aあi"になる

**実装ファイル:** `domain/buffer.go`, `UTF-8対応削除処理`

---

## delete/backward_line_join

### TestDeleteBackwardLineJoin

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** 行頭でのbackspaceによる行結合

**説明:** 行頭でbackspaceを実行して前の行と結合する機能

**前提:** 2行のテキスト（"hello"、"world"）でカーソルが2行目の行頭

**操作:** C-h（DeleteBackwardChar）コマンドを実行

**結果:** 2行が結合され"helloworld"の1行になる

**実装ファイル:** `domain/buffer.go`, `行結合処理`

---

## delete/edge_cases

### TestDeleteEdgeCases

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** 削除のエッジケース

**説明:** バッファの境界での削除動作

**前提:** 空のバッファまたは境界位置

**操作:** 削除コマンドを実行

**結果:** エラーなく適切に処理される

**実装ファイル:** `domain/buffer.go`, `境界チェック`

---

## delete/forward_char_basic

### TestDeleteCharBasic

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** C-d による基本的な文字削除

**説明:** カーソル位置の文字を削除する delete-char 機能

**前提:** "hello"を入力済みでカーソルを行頭に移動

**操作:** C-d（DeleteChar）コマンドを実行

**結果:** 最初の文字が削除され"ello"になる

**実装ファイル:** `domain/buffer.go`, `DeleteForward関数`

---

## delete/forward_char_japanese

### TestDeleteCharJapanese

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** 日本語文字のdelete-char削除

**説明:** 日本語文字（マルチバイト）のdelete-char削除機能

**前提:** "aあiい"を入力済みでカーソルを"あ"の位置に移動

**操作:** C-d（DeleteChar）コマンドを実行

**結果:** "あ"が削除され"aiい"になる

**実装ファイル:** `domain/buffer.go`, `UTF-8対応削除処理`

---

## delete/forward_line_join

### TestDeleteForwardLineJoin

**ファイル:** `e2e-test/delete_test.go`

**シナリオ:** 行末でのdelete-charによる行結合

**説明:** 行末でdelete-charを実行して次の行と結合する機能

**前提:** 2行のテキスト（"hello"、"world"）でカーソルが1行目の行末

**操作:** C-d（DeleteChar）コマンドを実行

**結果:** 2行が結合され"helloworld"の1行になる

**実装ファイル:** `domain/buffer.go`, `行結合処理`

---

## 画面表示機能 (display/basic_rendering)

### TestMockDisplayBasic

**ファイル:** `e2e-test/display_test.go`

**シナリオ:** 基本的なテキスト表示

**説明:** MockDisplayでの基本的なテキスト表示とカーソル位置の検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "hello"を入力する

**結果:** テキストが正確に表示され、カーソル位置が適切に設定される

**実装ファイル:** `test/mock_display.go`, `cli/display.go`

---

## 画面表示機能 (display/japanese_rendering)

### TestMockDisplayJapanese

**ファイル:** `e2e-test/display_test.go`

**シナリオ:** 日本語テキスト表示

**説明:** 日本語文字の表示と表示幅計算の検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "あいう"（ひらがな）を入力する

**結果:** 日本語テキストが正確に表示され、カーソル位置が適切に計算される

**実装ファイル:** `test/mock_display.go`, `UTF-8処理`

---

## 画面表示機能 (display/key_sequence_cancel)

### TestKeySequenceCancelDisplay

**ファイル:** `e2e-test/prefix_key_display_test.go`

**シナリオ:** キーシーケンスキャンセル後の表示

**説明:** Escapeキーでキーシーケンスをキャンセルした後の表示クリア

**前提:** C-x入力でキーシーケンス進行中

**操作:** Escapeキーを押下

**結果:** キーシーケンス表示がクリアされる

**実装ファイル:** `domain/editor.go`, `Escapeキー処理`

---

## 画面表示機能 (display/key_sequence_format)

### TestKeySequenceFormat

**ファイル:** `e2e-test/prefix_key_display_test.go`

**シナリオ:** キーシーケンス表示フォーマット

**説明:** 各種修飾キーの組み合わせの正しい表示

**前提:** キーバインディングマップを作成

**操作:** 各種キープレス組み合わせをフォーマット

**結果:** 適切な文字列表記が生成される

**実装ファイル:** `domain/keybinding.go`, `FormatSequence関数`

---

## 画面表示機能 (display/layout_analysis)

### TestDisplayLayoutAnalysis

**ファイル:** `e2e-test/display_layout_test.go`

**シナリオ:** 表示レイアウト解析

**説明:** 実際の表示レイアウトと期待されるレイアウトの比較分析

**前提:** 40x10ターミナルでリサイズイベントを送信

**操作:** ウィンドウ高と同じ数の行を追加し、さらに1行追加

**結果:** MockDisplayと実際のCLI Displayの動作が一致し、適切なスクロールタイミングが確認される

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

---

## 画面表示機能 (display/mixed_character_cursor)

### TestMockDisplayCursorProgression

**ファイル:** `e2e-test/display_test.go`

**シナリオ:** ASCII+日本語混在カーソル進行

**説明:** ASCII文字と日本語文字が混在するテキストでのカーソル位置進行

**前提:** 20x5サイズのMockDisplayを作成

**操作:** 'a'、'あ'、'b'、'い'、'c'を順次入力

**結果:** マルチバイト文字の表示幅を考慮してカーソルが適切に進行する

**実装ファイル:** `test/mock_display.go`, `文字幅計算`

---

## 画面表示機能 (display/mock_consistency)

### TestDisplayConsistency

**ファイル:** `e2e-test/actual_display_issue_test.go`

**シナリオ:** MockDisplayと実際のDisplay一貫性確認

**説明:** MockDisplayとWindow.VisibleLines()の表示内容が一致することを確認

**前提:** 80x10ターミナル環境

**操作:** 3行のテキスト（a、b、c）を入力する

**結果:** MockDisplayの内容とWindow.VisibleLines()が完全に一致する

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

---

## 画面表示機能 (display/mock_vs_real)

### TestRealVsMockDisplay

**ファイル:** `e2e-test/display_layout_test.go`

**シナリオ:** MockDisplayと実際のDisplay比較

**説明:** ユーザー報告シナリオでのMockDisplayと実際のCLI Displayの動作比較

**前提:** 40x10ターミナルでa〜hまで8行のコンテンツを作成

**操作:** 最後にEnterキーを押下

**結果:** MockDisplayの動作がユーザー期待（bから始まる表示）と一致する

**実装ファイル:** `test/mock_display.go`, `cli/display.go`

---

## 画面表示機能 (display/multiline_newline)

### TestNewlineAtEndOfLine

**ファイル:** `e2e-test/newline_display_test.go`

**シナリオ:** 複数改行での行末処理

**説明:** 連続した改行操作での行末処理とコンテンツ構築

**前提:** エディタを新規作成する

**操作:** "abc" + Enter + "def" + Enter + "ghi"を順次入力

**結果:** 3行のコンテンツが正確に作成され、カーソルが最終行の末尾に配置される

**実装ファイル:** `domain/buffer.go`, `複数行改行処理`

---

## 画面表示機能 (display/multiline_rendering)

### TestMockDisplayMultiline

**ファイル:** `e2e-test/display_test.go`

**シナリオ:** 複数行テキスト表示

**説明:** 複数行のテキストとカーソル位置の表示検証

**前提:** 10x5サイズのMockDisplayを作成

**操作:** "hello" + Enter + "world"を入力

**結果:** 2行のテキストが正確に表示され、2行目にカーソルが配置される

**実装ファイル:** `test/mock_display.go`, `複数行処理`

---

## 画面表示機能 (display/newline_rendering)

### TestNewlineDisplay

**ファイル:** `e2e-test/newline_display_test.go`

**シナリオ:** 改行表示のレンダリング

**説明:** 改行を含む複数行コンテンツの正確な表示検証

**前提:** 20x5サイズのMockDisplayを作成

**操作:** "hello" + Enter + "world"を入力

**結果:** 2行のコンテンツが正確に表示され、カーソル位置が適切に設定される

**実装ファイル:** `test/mock_display.go`, `改行処理`

---

## 画面表示機能 (display/prefix_key_display)

### TestPrefixKeyDisplay

**ファイル:** `e2e-test/prefix_key_display_test.go`

**シナリオ:** プレフィックスキーの表示

**説明:** C-x入力後にミニバッファに"C-x -"のような表示が出る機能

**前提:** エディタを新規作成

**操作:** C-xキーを押下

**結果:** キーシーケンス進行中の表示が"C-x -"になる

**実装ファイル:** `domain/keybinding.go`, `cli/display.go`

---

## 画面表示機能 (display/terminal_layout)

### TestActualDisplayIssue

**ファイル:** `e2e-test/actual_display_issue_test.go`

**シナリオ:** 余分な空行の回避

**説明:** 余分な空行が表示されたユーザー報告シナリオの正確なテスト

**前提:** 12行のターミナル（ユーザーの報告環境）

**操作:** 文字a〜dをそれぞれEnterで区切って入力する

**結果:** 余分な空白なしで実際のコンテンツ行のみがレンダリングされる

**実装ファイル:** `cli/display.go`, `test/mock_display.go`

**バグ修正:** height-2 vs height-3の不整合と無条件改行出力を修正

---

## 画面表示機能 (display/terminal_width_handling)

### TestMockDisplayWidthProblem

**ファイル:** `e2e-test/display_test.go`

**シナリオ:** ターミナル幅と文字幅の問題検証

**説明:** 異なる文字タイプのターミナル幅処理の検証

**前提:** 10x3サイズのMockDisplayを作成

**操作:** ASCII、日本語、混在テキストを各々入力

**結果:** ターミナル表示幅とルーン数の違いを適切に処理する

**実装ファイル:** `test/mock_display.go`, `文字幅計算`

---

## エディタ基本機能 (editor/startup)

### TestEditorStartup

**ファイル:** `e2e-test/editor_startup_test.go`

**シナリオ:** エディタ初期化と基本状態の確認

**説明:** エディタ起動時の初期状態（バッファ、ウィンドウ、レンダリング）を検証

**前提:** エディタを新規作成する

**操作:** エディタの初期状態を確認する

**結果:** 実行中状態、*scratch*バッファ、ウィンドウが正しく設定される

**実装ファイル:** `domain/editor.go`, `domain/buffer.go`, `domain/window.go`

---

## events/queue_capacity

### TestEventQueueCapacity

**ファイル:** `e2e-test/event_system_test.go`

**シナリオ:** イベントキューの容量制限

**説明:** イベントキューの容量制限とオーバーフロー処理

**前提:** 容量2のイベントキューを作成

**操作:** 3つのイベント（A、B、C）を順次プッシュ

**結果:** 最初の2つのイベント（A、B）のみが保持され、3番目（C）は破棄される

**実装ファイル:** `events/event_queue.go`, `容量制限処理`

---

## events/queue_operations

### TestEventQueue

**ファイル:** `e2e-test/event_system_test.go`

**シナリオ:** イベントキューの基本操作

**説明:** イベントキューのPush/Pop操作の基本動作検証

**前提:** エディタを新規作成してイベントキューを取得

**操作:** KeyEventData('A')をキューにプッシュし、ポップする

**結果:** イベントが正しく取り出され、データが保持される

**実装ファイル:** `events/event_queue.go`

---

## events/quit_handling

### TestQuitEvent

**ファイル:** `e2e-test/event_system_test.go`

**シナリオ:** 終了イベントの処理

**説明:** 終了イベントの処理とエディタ状態の変更

**前提:** エディタが実行中の状態

**操作:** QuitEventDataを送信する

**結果:** エディタが終了状態に変更される

**実装ファイル:** `events/quit_event.go`, `domain/editor.go`

---

## events/resize_handling

### TestResizeEvent

**ファイル:** `e2e-test/event_system_test.go`

**シナリオ:** リサイズイベントの処理

**説明:** ターミナルリサイズイベントの処理とウィンドウサイズ更新

**前提:** エディタを新規作成する

**操作:** 100x30サイズのリサイズイベントを送信

**結果:** ウィンドウサイズが100x28（モードラインとミニバッファを除いたサイズ）に更新される

**実装ファイル:** `events/resize_event.go`, `domain/window.go`

---

## file/find_file_basic

### TestFindFileBasic

**ファイル:** `e2e-test/file_test.go`

**シナリオ:** C-x C-f による基本的なファイル開く機能

**説明:** ファイルパスを入力してファイルを開く基本機能

**前提:** 存在するテストファイルを用意

**操作:** C-x C-f コマンドでファイルパスを入力

**結果:** ファイルの内容がバッファに読み込まれ、適切に表示される

**実装ファイル:** `domain/buffer.go`, `NewBufferFromFile関数`

---

## file/find_file_cancel

### TestFindFileCancel

**ファイル:** `e2e-test/file_test.go`

**シナリオ:** C-x C-f のキャンセル

**説明:** Escapeキーでファイル入力をキャンセルする機能

**前提:** C-x C-f を実行してファイル入力モードに入る

**操作:** Escapeキーを押す

**結果:** ミニバッファがクリアされ、元の状態に戻る

**実装ファイル:** `domain/editor.go`, `キャンセル処理`

---

## file/find_file_empty

### TestFindFileEmpty

**ファイル:** `e2e-test/file_test.go`

**シナリオ:** 空ファイルを開く場合

**説明:** 空のファイルを開いた際の適切な処理

**前提:** 空のファイル

**操作:** C-x C-f コマンドで空ファイルを開く

**結果:** 空行が1行あるバッファが作成される

**実装ファイル:** `domain/buffer.go`, `空ファイル処理`

---

## file/find_file_japanese

### TestFindFileJapanese

**ファイル:** `e2e-test/file_test.go`

**シナリオ:** 日本語ファイル名での動作

**説明:** 日本語を含むファイルパスでの正常動作

**前提:** 日本語ファイル名のテストファイル

**操作:** C-x C-f で日本語ファイル名を入力

**結果:** ファイルが正常に開かれる

**実装ファイル:** `domain/buffer.go`, `UTF-8ファイル名対応`

---

## file/find_file_nonexistent

### TestFindFileNonexistent

**ファイル:** `e2e-test/file_test.go`

**シナリオ:** 存在しないファイルを開こうとした場合

**説明:** 存在しないファイルパスでC-x C-fを実行した際のエラーハンドリング

**前提:** 存在しないファイルパス

**操作:** C-x C-f コマンドで存在しないファイルパスを入力

**結果:** エラーメッセージが表示され、現在のバッファは変更されない

**実装ファイル:** `domain/editor.go`, `エラーハンドリング`

---

## キーボード入力機能 (input/basic_text)

### TestBasicTextInput

**ファイル:** `e2e-test/text_input_test.go`

**シナリオ:** 基本的なテキスト入力

**説明:** ASCII文字の連続入力と表示の検証

**前提:** エディタを新規作成する

**操作:** "Hello, World!"を1文字ずつ入力する

**結果:** 入力したテキストが正確に表示される

**実装ファイル:** `domain/buffer.go`, `domain/editor.go`

---

## キーボード入力機能 (input/japanese)

### TestJapaneseTextInput

**ファイル:** `e2e-test/text_input_test.go`

**シナリオ:** 日本語テキスト入力

**説明:** ひらがな文字の入力と表示の検証

**前提:** エディタを新規作成する

**操作:** "あいう"を文字ごとに入力する

**結果:** 日本語テキストが正確に表示される

**実装ファイル:** `domain/buffer.go`, `UTF-8処理`

---

## キーボード入力機能 (input/multiline)

### TestMultilineTextInput

**ファイル:** `e2e-test/text_input_test.go`

**シナリオ:** 複数行テキスト入力

**説明:** 3行のテキストを順次入力し、行分離を検証

**前提:** エディタを新規作成する

**操作:** "First line", "Second line", "Third line"を Enter で区切って入力する

**結果:** 3行が正確に分かれて表示される

**実装ファイル:** `domain/buffer.go`, `domain/editor.go`

---

## キーボード入力機能 (input/newline)

### TestEnterKeyNewline

**ファイル:** `e2e-test/text_input_test.go`

**シナリオ:** Enter キーによる改行

**説明:** Enter キーで行を分割し複数行テキストを作成

**前提:** エディタに "Hi" を入力済み

**操作:** Enter キーを押して "Wo" を入力する

**結果:** 2行に分かれてテキストが表示される

**実装ファイル:** `domain/buffer.go`, `events/key_event.go`

---

## キーボード入力機能 (input/newline_basic)

### TestNewlineBasic

**ファイル:** `e2e-test/newline_test.go`

**シナリオ:** 基本的な改行挿入

**説明:** 行末でのEnterキーによる基本的な改行動作

**前提:** 空のバッファに"hello"を入力済み

**操作:** 行末でEnterキーを押し、"world"を入力

**結果:** 2行に分かれてテキストが配置され、カーソルが適切な位置に移動する

**実装ファイル:** `domain/buffer.go`, `events/key_event.go`

---

## キーボード入力機能 (input/newline_beginning)

### TestNewlineAtBeginning

**ファイル:** `e2e-test/newline_test.go`

**シナリオ:** 行頭での改行挿入

**説明:** 行頭でEnterキーを押した際の新しい行挿入動作

**前提:** "hello"を入力済みでカーソルを行頭に移動

**操作:** 行頭でEnterキーを押下

**結果:** 空の新しい行が挿入され、既存のコンテンツが2行目に移動する

**実装ファイル:** `domain/buffer.go`, `行挿入処理`

---

## キーボード入力機能 (input/newline_multiple)

### TestMultipleNewlines

**ファイル:** `e2e-test/newline_test.go`

**シナリオ:** 連続した改行挿入

**説明:** 複数のEnterキーを連続して押した際の動作

**前提:** 空のバッファから開始

**操作:** "a" + Enter + "b" + Enter + "c"を順次入力

**結果:** 3行のコンテンツが正確に作成され、カーソル位置が適切に設定される

**実装ファイル:** `domain/buffer.go`, `複数行処理`

---

## キーボード入力機能 (input/newline_split)

### TestNewlineInMiddle

**ファイル:** `e2e-test/newline_test.go`

**シナリオ:** 行の中間での改行挿入

**説明:** 行の中間でEnterキーを押した際の行分割動作

**前提:** "hello world"を入力済みでカーソルを"hello"の後（位置5）に移動

**操作:** カーソル位置でEnterキーを押下

**結果:** 行が"hello"と" world"に分割され、カーソルが2行目の先頭に移動する

**実装ファイル:** `domain/buffer.go`, `行分割処理`

---

## keyboard/ctrl_c_no_quit

### TestCtrlCAloneDoesNotQuit

**ファイル:** `e2e-test/keyboard_shortcuts_test.go`

**シナリオ:** C-c単独ではエディタ終了しない

**説明:** C-x prefix key なしのC-cではエディタが終了しないことの検証

**前提:** エディタが実行中の状態

**操作:** C-cキーイベントのみを送信する

**結果:** エディタが実行中のまま維持される

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## keyboard/ctrl_modifier_no_insert

### TestCtrlModifierDoesNotInsertText

**ファイル:** `e2e-test/keyboard_shortcuts_test.go`

**シナリオ:** Ctrl修飾キーのテキスト非挿入

**説明:** Ctrl+文字キーの組み合わせでテキストが挿入されないことの検証

**前提:** エディタを新規作成する

**操作:** Ctrl+aキーイベントを送信する

**結果:** テキストが挿入されず、空の行が維持される

**実装ファイル:** `domain/editor.go`, `events/key_event.go`

---

## keyboard/ctrl_x_ctrl_c_quit

### TestCtrlXCtrlCQuit

**ファイル:** `e2e-test/keyboard_shortcuts_test.go`

**シナリオ:** C-x C-cでのエディタ終了

**説明:** C-x C-cキーシーケンスでエディタを終了する機能の検証

**前提:** エディタが実行中の状態

**操作:** C-x（prefix key）とC-cキーイベントを順次送信する

**結果:** エディタが終了状態になる

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## keyboard/ctrl_x_prefix_reset

### TestCtrlXPrefixReset

**ファイル:** `e2e-test/keyboard_shortcuts_test.go`

**シナリオ:** C-x prefix key状態のリセット

**説明:** C-x後に無効なキーを押すとprefix状態がリセットされることの検証

**前提:** エディタが実行中の状態

**操作:** C-x後に通常の文字キーを送信する

**結果:** prefix状態がリセットされ、通常のテキスト入力として処理される

**実装ファイル:** `domain/editor.go`, `prefix key システム`

---

## keyboard/key_sequence_binding

### TestKeySequenceBinding

**ファイル:** `e2e-test/key_sequence_test.go`

**シナリオ:** キーシーケンスバインディングシステム

**説明:** BindKeySequence APIでキーシーケンスを設定し実行する機能の検証

**前提:** 新しいキーバインディングマップを作成

**操作:** "C-x C-f"のようなキーシーケンスをバインドし、該当するキー入力を送信

**結果:** バインドされたコマンドが実行される

**実装ファイル:** `domain/keybinding.go`, `キーシーケンス処理システム`

---

## keyboard/meta_modifier_no_insert

### TestMetaModifierDoesNotInsertText

**ファイル:** `e2e-test/keyboard_shortcuts_test.go`

**シナリオ:** Meta修飾キーのテキスト非挿入

**説明:** Meta+文字キーの組み合わせでテキストが挿入されないことの検証

**前提:** エディタを新規作成する

**操作:** Meta+xキーイベントを送信する

**結果:** テキストが挿入されず、空の行が維持される

**実装ファイル:** `domain/editor.go`, `events/key_event.go`

---

## keyboard/multiple_sequences

### TestMultipleKeySequences

**ファイル:** `e2e-test/key_sequence_test.go`

**シナリオ:** 複数キーシーケンスの同時サポート

**説明:** 複数の異なるキーシーケンスを同時にサポートする機能の検証

**前提:** キーバインディングマップに"C-x C-c"と"C-x C-f"を両方バインド

**操作:** 各シーケンスを順次実行

**結果:** それぞれ対応するコマンドが実行される

**実装ファイル:** `domain/keybinding.go`, `複数シーケンス管理`

---

## keyboard/quit_find_file

### TestKeyboardQuitFindFile

**ファイル:** `e2e-test/keyboard_quit_test.go`

**シナリオ:** C-x C-f ファイル入力時のC-gキャンセル

**説明:** C-x C-f ファイル入力中にC-gでキャンセルする機能

**前提:** C-x C-f ファイル入力モードでパスを部分的に入力済み

**操作:** C-gキーを押下

**結果:** ミニバッファがクリアされ、元の状態に戻る

**実装ファイル:** `domain/command.go`, `KeyboardQuit関数`

---

## keyboard/quit_key_sequence

### TestKeyboardQuitKeySequence

**ファイル:** `e2e-test/keyboard_quit_test.go`

**シナリオ:** 進行中のキーシーケンスのC-gキャンセル

**説明:** C-x 入力後にC-gでキーシーケンスをキャンセルする機能

**前提:** C-x入力済みでキーシーケンスが進行中

**操作:** C-gキーを押下

**結果:** キーシーケンス状態がリセットされる

**実装ファイル:** `domain/command.go`, `KeyboardQuit関数`

---

## keyboard/quit_message_clear

### TestKeyboardQuitMessageClear

**ファイル:** `e2e-test/keyboard_quit_test.go`

**シナリオ:** メッセージ表示中のC-gクリア

**説明:** ミニバッファにメッセージが表示されている時のC-g動作

**前提:** ミニバッファにメッセージが表示されている状態

**操作:** C-gキーを押下

**結果:** メッセージがクリアされる

**実装ファイル:** `domain/command.go`, `KeyboardQuit関数`

---

## keyboard/quit_mx_command

### TestKeyboardQuitMxCommand

**ファイル:** `e2e-test/keyboard_quit_test.go`

**シナリオ:** M-xコマンド入力時のC-gキャンセル

**説明:** M-xコマンド入力中にC-gでキャンセルする機能

**前提:** M-xコマンド入力モードで部分的にコマンドを入力済み

**操作:** C-gキーを押下

**結果:** ミニバッファがクリアされ、通常モードに戻る

**実装ファイル:** `domain/command.go`, `KeyboardQuit関数`

---

## keyboard/quit_normal_mode

### TestKeyboardQuitNormalMode

**ファイル:** `e2e-test/keyboard_quit_test.go`

**シナリオ:** 通常モードでのC-g動作

**説明:** ミニバッファがアクティブでない時のC-g動作

**前提:** 通常編集モード

**操作:** C-gキーを押下

**結果:** 特に何も起こらず、キーシーケンス状態がリセットされる

**実装ファイル:** `domain/command.go`, `KeyboardQuit関数`

---

## keyboard/sequence_reset

### TestKeySequenceReset

**ファイル:** `e2e-test/key_sequence_test.go`

**シナリオ:** キーシーケンス状態のリセット

**説明:** 無効なキーが入力された場合のシーケンス状態リセットの検証

**前提:** キーバインディングマップに"C-x C-c"をバインド

**操作:** C-x後に無効なキー（'z'）を送信

**結果:** シーケンス状態がリセットされ、その後のC-cでは実行されない

**実装ファイル:** `domain/keybinding.go`, `シーケンス状態管理`

---

## minibuffer/cursor_position_accuracy

### TestMinibufferCursorPositionAccuracy

**ファイル:** `e2e-test/prefix_key_display_test.go`

**シナリオ:** ミニバッファでのカーソル位置精度

**説明:** ミニバッファでの日本語文字を含むテキストでのカーソル位置計算精度

**前提:** M-xコマンド入力モードでマルチバイト文字を入力

**操作:** カーソル移動を行う

**結果:** 正確なバイト位置とルーン位置が計算される

**実装ファイル:** `domain/minibuffer.go`, `カーソル位置計算`

---

## minibuffer/edit_boundary_conditions

### TestMinibufferEditBoundaryConditions

**ファイル:** `e2e-test/minibuffer_edit_test.go`

**シナリオ:** ミニバッファ編集の境界条件

**説明:** カーソルが境界位置にある時の編集動作

**前提:** M-xコマンド入力モードで"test"を入力済み

**操作:** 境界位置での削除とカーソル移動を試行

**結果:** エラーなく適切に処理される

**実装ファイル:** `domain/minibuffer.go`, `境界チェック`

---

## minibuffer/edit_cursor_movement

### TestMinibufferCursorMovement

**ファイル:** `e2e-test/minibuffer_edit_test.go`

**シナリオ:** ミニバッファでのカーソル移動

**説明:** M-xコマンド入力中にC-f/C-bでカーソルを移動する機能

**前提:** M-xコマンド入力モードで"hello"を入力済み

**操作:** C-a（行頭）、C-f（前進）、C-b（後退）、C-e（行末）を順次実行

**結果:** カーソルが適切な位置に移動する

**実装ファイル:** `domain/minibuffer.go`, `カーソル移動関数`

---

## minibuffer/edit_delete_forward

### TestMinibufferDeleteForward

**ファイル:** `e2e-test/minibuffer_edit_test.go`

**シナリオ:** ミニバッファでのC-d文字削除

**説明:** M-xコマンド入力中にC-dで前方の文字を削除する機能

**前提:** M-xコマンド入力モードで"forward"を入力済み、カーソルが"f"の位置

**操作:** C-dキーを押下

**結果:** "f"が削除され"orward"になる

**実装ファイル:** `domain/minibuffer.go`, `DeleteForward関数`

---

## minibuffer/edit_file_input

### TestMinibufferFileInputEdit

**ファイル:** `e2e-test/minibuffer_edit_test.go`

**シナリオ:** ファイル入力モードでの編集機能

**説明:** C-x C-fファイル入力中にC-h/C-dで編集する機能

**前提:** C-x C-fファイル入力モードで"/path/to/file.txt"を入力済み

**操作:** カーソル移動と削除コマンドを実行

**結果:** ファイルパスが適切に編集される

**実装ファイル:** `domain/minibuffer.go`, `ファイル入力モード編集`

---

## minibuffer/edit_japanese_characters

### TestMinibufferJapaneseEdit

**ファイル:** `e2e-test/minibuffer_edit_test.go`

**シナリオ:** ミニバッファでの日本語文字編集

**説明:** M-xコマンド入力中に日本語文字を含むテキストを編集する機能

**前提:** M-xコマンド入力モードで"aあbいc"を入力済み

**操作:** カーソル移動と削除を行う

**結果:** 日本語文字が適切に処理される

**実装ファイル:** `domain/minibuffer.go`, `UTF-8対応編集`

---

## resize/cursor_position_preservation

### TestCursorPositionAfterResize

**ファイル:** `e2e-test/resize_test.go`

**シナリオ:** リサイズ後のカーソル位置保持

**説明:** ターミナルリサイズ後のカーソル位置保持の検証

**前提:** "hello"を入力しカーソルを中央（位置2）に設定

**操作:** ターミナルを120x30にリサイズする

**結果:** カーソル位置がリサイズ後も(0,2)で保持される

**実装ファイル:** `domain/window.go`, `domain/cursor.go`

---

## resize/multiple_resizes

### TestMultipleResizes

**ファイル:** `e2e-test/resize_test.go`

**シナリオ:** 連続的なリサイズ操作

**説明:** 複数回のリサイズ操作でのサイズ更新とコンテンツ保持

**前提:** 80x24サイズで"test content"を入力済み

**操作:** 異なるサイズで複数回連続してリサイズする

**結果:** 各リサイズ後にサイズが正確に更新され、コンテンツが保持される

**実装ファイル:** `domain/window.go`, `events/resize_event.go`

---

## resize/smaller_size_resize

### TestResizeToSmallerSize

**ファイル:** `e2e-test/resize_test.go`

**シナリオ:** 小さいサイズへのリサイズ

**説明:** ターミナルを小さいサイズにリサイズした際のコンテンツ保持

**前提:** 80x24サイズで複数行のコンテンツを入力済み

**操作:** ターミナルのサイズを40x10に縮小する

**結果:** ウィンドウサイズが更新され、バッファの全コンテンツが保持される

**実装ファイル:** `domain/window.go`, `domain/buffer.go`

---

## resize/terminal_resize

### TestTerminalResize

**ファイル:** `e2e-test/resize_test.go`

**シナリオ:** ターミナルリサイズ処理

**説明:** ターミナルサイズ変更時のウィンドウサイズ更新とコンテンツ保持

**前提:** 80x24サイズのターミナルで"hello world"を入力済み

**操作:** ターミナルを120x30にリサイズする

**結果:** ウィンドウサイズが更新され、コンテンツが保持される

**実装ファイル:** `domain/window.go`, `events/resize_event.go`

---

## スクロール機能 (scroll/auto_scroll_insertion)

### TestAutoScrollOnTextInsertion

**ファイル:** `e2e-test/auto_scroll_test.go`

**シナリオ:** テキスト挿入時の自動スクロール

**説明:** 可視範囲を超えるテキスト挿入時のスクロール動作

**前提:** 30x6の小さいウィンドウ（4コンテンツ行）に3行の初期コンテンツ

**操作:** さらに5行の新しいコンテンツを追加

**結果:** スクロールが発生し、カーソルが可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## スクロール機能 (scroll/auto_scroll_lines)

### TestAutoScrollWhenAddingLines

**ファイル:** `e2e-test/auto_scroll_test.go`

**シナリオ:** 行追加時の自動スクロール

**説明:** ウィンドウ高を超える行を追加した際の自動スクロール動作

**前提:** 40x10サイズのディスプレイ（8コンテンツ行）

**操作:** 15行のコンテンツを順次追加する

**結果:** カーソルが常に可視範囲内に保たれ、現在の行が表示される

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## スクロール機能 (scroll/auto_scroll_wrapping)

### TestAutoScrollWithLongLines

**ファイル:** `e2e-test/auto_scroll_test.go`

**シナリオ:** 長い行での自動スクロールと行ラップ

**説明:** 行ラップ有効時の長い行での自動スクロール動作

**前提:** 20x8の小さいウィンドウで行ラップ有効

**操作:** 短い行と長い行（ラップする）を混在して追加

**結果:** カーソルが常に可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `domain/window.go`

---

## スクロール機能 (scroll/cursor_movement_display)

### TestCursorMovementTriggersDisplay

**ファイル:** `e2e-test/auto_scroll_test.go`

**シナリオ:** 手動カーソル移動時の表示更新

**説明:** 手動でカーソルを移動した際の適切な表示更新

**前提:** 30x8ウィンドウに20行のコンテンツを作成

**操作:** カーソルを手動でバッファの先頭に移動

**結果:** ウィンドウがスクロールしてカーソルが表示される

**実装ファイル:** `domain/scroll.go`, `domain/cursor.go`

---

## スクロール機能 (scroll/edge_case_debug)

### TestDebugScrollBehavior

**ファイル:** `e2e-test/debug_scroll_test.go`

**シナリオ:** スクロールエッジケースのデバッグ

**説明:** 8行丁度まで埋めた後のEnterキー押下時のスクロール動作の詳細分析

**前提:** 40x10ディスプレイ（8コンテンツ行）で8行丁度までコンテンツを埋める

**操作:** 最後の可視行でEnterキーを押下

**結果:** スクロール量と表示内容が期待値と一致し、適切な1行スクロールが発生する

**実装ファイル:** `domain/scroll.go`, `エッジケース処理`

---

## スクロール機能 (scroll/enter_timing_issue)

### TestEnterKeyTimingIssue

**ファイル:** `e2e-test/enter_timing_test.go`

**シナリオ:** Enterキータイミング問題の検証

**説明:** 最後の可視行でEnterキーを押した際のスクロールタイミング問題の検証

**前提:** 40x10ディスプレイ（8コンテンツ行）でまず7行を作成

**操作:** 最後の可視行（行7）でEnterキーを押下

**結果:** カーソルが行8に移動し、即座に1行スクロールが発生する

**実装ファイル:** `domain/scroll.go`, `スクロールタイミング修正`

---

## スクロール機能 (scroll/exact_user_scenario)

### TestExactUserScenario

**ファイル:** `e2e-test/exact_user_scenario_test.go`

**シナリオ:** ユーザー報告の正確なシナリオ再現

**説明:** 高さ10ターミナルでa〜hまで入力後のEnter時のスクロール動作

**前提:** 高さ10ターミナル（コンテンツエリア8行）でリサイズイベントを発生

**操作:** a + Enter + b + ... + h を入力し、最後にEnterを押下

**結果:** a〜hが表示され、Enter後はb〜h+空行が表示される

**実装ファイル:** `domain/scroll.go`, `ユーザーシナリオ修正`

---

## スクロール機能 (scroll/horizontal_boundary_scroll)

### TestHorizontalBoundaryScroll

**ファイル:** `e2e-test/horizontal_scroll_test.go`

**シナリオ:** 水平スクロール境界でのスクロール動作

**説明:** カーソルが可視範囲の左右境界を超えた時のスクロール動作

**前提:** 行ラップ無効の狭いウィンドウと長い行

**操作:** カーソルを左右の境界を超えて移動

**結果:** 適切なタイミングでスクロールが発生する

**実装ファイル:** `domain/scroll.go`, `境界スクロール処理`

---

## スクロール機能 (scroll/horizontal_cursor_follow)

### TestHorizontalScrollCursorFollow

**ファイル:** `e2e-test/horizontal_scroll_test.go`

**シナリオ:** 水平スクロール時のカーソル追従

**説明:** 行ラップ無効時の水平スクロールとカーソル移動の同期検証

**前提:** 狭いウィンドウで行ラップを無効にし、長い行を作成

**操作:** カーソルを右端まで移動し、その後左に戻る

**結果:** カーソル位置に応じて水平スクロールが正しく調整される

**実装ファイル:** `domain/scroll.go`, `水平スクロール制御`

---

## スクロール機能 (scroll/horizontal_scrolling)

### TestHorizontalScrolling

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** 水平スクロール動作

**説明:** 長い行のコンテンツでの水平スクロール動作の検証

**前提:** 10x5の狭いウィンドウと長い行のコンテンツ

**操作:** 行ラップを無効化して水平スクロールを設定

**結果:** 指定した位置からコンテンツが表示される

**実装ファイル:** `domain/window.go`, `水平スクロール`

---

## スクロール機能 (scroll/horizontal_toggle_wrap_state)

### TestHorizontalToggleWrapState

**ファイル:** `e2e-test/horizontal_scroll_test.go`

**シナリオ:** 行ラップ切り替え時の水平スクロール状態

**説明:** 行ラップの有効/無効切り替え時の水平スクロール状態の保持

**前提:** 長い行とカーソルが右端にある状態

**操作:** 行ラップの有効/無効を切り替える

**結果:** 適切にスクロール状態が管理される

**実装ファイル:** `domain/scroll.go`, `ラップ切り替え処理`

---

## スクロール機能 (scroll/individual_scroll_commands)

### TestScrollCommands

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** 個別スクロールコマンド

**説明:** ScrollUp/ScrollDownコマンドによる1行単位のスクロール

**前提:** 30行のコンテンツを持つエディタ

**操作:** ScrollDown、ScrollUpコマンドを順次実行

**結果:** スクロール位置が1行単位で正確に変更される

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## スクロール機能 (scroll/line_wrapping)

### TestLineWrapping

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** 行ラップ機能

**説明:** 長い行のラップ機能の有効/無効切り替え検証

**前提:** 10x5の小さいウィンドウと長い行のコンテンツ

**操作:** 行ラップの有効/無効を切り替える

**結果:** ラップ有効時は複数行、無効時は単一行で表示される

**実装ファイル:** `domain/window.go`, `行ラップ処理`

---

## スクロール機能 (scroll/page_navigation)

### TestPageUpDown

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** ページアップ/ダウンナビゲーション

**説明:** PageUp/PageDownコマンドによるページ単位のスクロール

**前提:** 50行の大量コンテンツを持つエディタ

**操作:** PageDown、PageUpコマンドを順次実行

**結果:** スクロール位置がページ単位で適切に変更される

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## スクロール機能 (scroll/realistic_terminal)

### TestRealisticTerminalScroll

**ファイル:** `e2e-test/realistic_scroll_test.go`

**シナリオ:** リアルなターミナルサイズでのスクロール

**説明:** 80x24のリアルなターミナルサイズでのスクロール動作検証

**前提:** 80x24ターミナル（22コンテンツ行）でリサイズイベントを送信

**操作:** 30行のコンテンツを順次追加し、各ステップでスクロール状態を監視

**結果:** ウィンドウ高を超えたタイミングでスクロールが開始され、カーソルが常に可視範囲内に保たれる

**実装ファイル:** `domain/scroll.go`, `リアルターミナル環境`

---

## スクロール機能 (scroll/scroll_timing)

### TestTerminal12LinesDebugSteps

**ファイル:** `e2e-test/terminal_12_lines_test.go`

**シナリオ:** コンテンツがウィンドウを超えた時のスクロール

**説明:** スクロール動作をステップごとに検証するデバッグテスト

**前提:** 12行のコンテンツエリアを持つターミナル

**操作:** コンテンツエリア限界を超えて一行ずつ追加する

**結果:** 適切なタイミングでスクロールが発生する

**実装ファイル:** `domain/scroll.go`, `cli/display.go`

---

### TestTerminal12LinesScenario

**ファイル:** `e2e-test/terminal_12_lines_test.go`

**シナリオ:** 早すぎるスクロールの回避

**説明:** コンテンツがウィンドウコンテンツエリアを真に超えるまでスクロールが発生しないことをテスト

**前提:** 12行のターミナル（10コンテンツ + モード + ミニ）

**操作:** 文字a〜jをそれぞれEnterで区切って入力する

**結果:** すべての10行がスクロールなしで表示される

**実装ファイル:** `domain/scroll.go`, `cli/display.go`

---

## スクロール機能 (scroll/step_by_step_debug)

### TestUserScenarioStepByStep

**ファイル:** `e2e-test/exact_user_scenario_test.go`

**シナリオ:** ユーザーシナリオのステップバイステップデバッグ

**説明:** ユーザー報告シナリオをステップごとに詳細に検証するデバッグテスト

**前提:** 40x10ディスプレイでウィンドウサイズを設定

**操作:** a〜hをステップごとに入力し、各ステップで状態をログ出力

**結果:** 各ステップでカーソル位置とスクロール状態が正しく、最終的に期待結果を得る

**実装ファイル:** `domain/scroll.go`, `デバッグ情報出力`

---

## スクロール機能 (scroll/timing_verification)

### TestScrollStartsAtRightTime

**ファイル:** `e2e-test/realistic_scroll_test.go`

**シナリオ:** 異なるウィンドウサイズでのスクロールタイミング検証

**説明:** 複数のウィンドウサイズでスクロール開始タイミングの正確性を検証

**前提:** 異なるターミナル高（6、6、10、24）でテストケースを実行

**操作:** 各サイズでウィンドウ高まで行を追加し、さらに1行追加

**結果:** ウィンドウ高まではスクロールせず、超えた時点でスクロールが発生する

**実装ファイル:** `domain/scroll.go`, `サイズ別タイミング検証`

---

## スクロール機能 (scroll/toggle_line_wrap)

### TestToggleLineWrap

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** 行ラップトグルコマンド

**説明:** ToggleLineWrapコマンドによる行ラップ状態の切り替え

**前提:** エディタを新規作成（デフォルトでラップ有効）

**操作:** ToggleLineWrapコマンドを実行

**結果:** 行ラップの有効/無効が切り替わる

**実装ファイル:** `domain/commands.go`, `domain/window.go`

---

## スクロール機能 (scroll/user_reported_behavior)

### TestUserReportedBehavior

**ファイル:** `e2e-test/enter_timing_test.go`

**シナリオ:** ユーザー報告された問題の再現

**説明:** ユーザーが報告したスクロールディレイの正確な再現テスト

**前提:** 8行でスクリーンを埋めた状態

**操作:** 連続してEnter+コンテンツ入力を繰り返す

**結果:** ユーザー期待と実際の動作の違いを特定し、修正を検証する

**実装ファイル:** `domain/scroll.go`, `ユーザー報告修正`

---

## スクロール機能 (scroll/vertical_scrolling)

### TestVerticalScrolling

**ファイル:** `e2e-test/scrolling_test.go`

**シナリオ:** 垂直スクロール動作

**説明:** 大量のコンテンツがある場合の垂直スクロール動作の検証

**前提:** 40x10サイズのウィンドウに20行のコンテンツを作成

**操作:** カーソルが最後の行にある状態でスクロール位置を設定

**結果:** カーソルが可視範囲に保たれるように自動スクロールされる

**実装ファイル:** `domain/window.go`, `domain/scroll.go`

---

## terminal/width_calculation

### TestTerminalWidthIssue

**ファイル:** `e2e-test/terminal_width_test.go`

**シナリオ:** ターミナル幅計算問題の検証

**説明:** ASCII文字と日本語文字の混合テキストでのターミナル表示位置計算

**前提:** 20x3のMockDisplayと様々な文字組み合わせのテストケース

**操作:** 各テストケースで文字を入力し、カーソル位置を取得

**結果:** ASCII文字は1列、日本語文字は2列、混合テキストは合計列数で正確に表示される

**実装ファイル:** `test/mock_display.go`, `ターミナル幅計算処理`

---

## text/partial_string_width

### TestStringWidthUpTo

**ファイル:** `e2e-test/runewidth_test.go`

**シナリオ:** 部分文字列幅計算機能

**説明:** 指定バイト位置までの文字列の表示幅計算

**前提:** ASCII文字列、日本語文字列、混合文字列と様々なバイト位置

**操作:** StringWidthUpTo関数で指定位置までの表示幅を計算

**結果:** マルチバイト文字の境界を考慮した正確な部分幅が計算される

**実装ファイル:** `util/runewidth.go`, `部分文字列幅計算`

---

## text/rune_width_calculation

### TestRuneWidth

**ファイル:** `e2e-test/runewidth_test.go`

**シナリオ:** 文字幅計算機能

**説明:** Unicode文字（ASCII、日本語、制御文字）の表示幅計算

**前提:** ASCII文字、日本語文字、制御文字のテストケース

**操作:** RuneWidth関数で各文字の表示幅を計算

**結果:** ASCII文字は幅1、日本語文字は幅2、制御文字は幅0で計算される

**実装ファイル:** `util/runewidth.go`, `文字幅計算`

---

## text/string_width_calculation

### TestStringWidth

**ファイル:** `e2e-test/runewidth_test.go`

**シナリオ:** 文字列幅計算機能

**説明:** ASCII、日本語、混合文字列の総表示幅計算

**前提:** 空文字列、ASCII文字列、日本語文字列、混合文字列のテストケース

**操作:** StringWidth関数で各文字列の総表示幅を計算

**結果:** 各文字の幅の合計値が正確に計算される（混合文字列は範囲チェック）

**実装ファイル:** `util/runewidth.go`, `文字列幅計算`

---

*このドキュメントは自動生成されています。修正はテストファイルのアノテーションを編集してください。*
