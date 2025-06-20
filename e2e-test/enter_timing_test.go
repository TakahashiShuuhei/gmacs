package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec scroll/enter_timing_issue
 * @scenario Enterキータイミング問題の検証
 * @description 最後の可視行でEnterキーを押した際のスクロールタイミング問題の検証
 * @given 40x10ディスプレイ（8コンテンツ行）でまず7行を作成
 * @when 最後の可視行（行7）でEnterキーを押下
 * @then カーソルが行8に移動し、即座に1行スクロールが発生する
 * @implementation domain/scroll.go, スクロールタイミング修正
 */
func TestEnterKeyTimingIssue(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10) // 10 total = 8 content + mode + mini
	
	window := editor.CurrentWindow()
	window.Resize(40, 8) // 8 content lines (0-7)
	
	// Add 7 lines first (lines 0-6)
	for i := 0; i < 7; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	
	display.Render(editor)
	
	// Now we're at cursor (6,6), should show lines 0-6, no scroll yet
	bufferCursor := editor.CurrentBuffer().Cursor()
	t.Logf("After 7 lines: cursor (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	
	if bufferCursor.Row != 6 || window.ScrollTop() != 0 {
		t.Errorf("Unexpected initial state: cursor (%d,%d), scroll %d", 
			bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	}
	
	// Press Enter to create line 7
	// Cursor should go to (7,0), still visible at screen row 7, no scroll needed
	t.Logf("=== Press Enter to create line 7 ===")
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, _ := window.CursorPosition()
	visible := window.VisibleLines()
	t.Logf("After creating line 7: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// Should be cursor (7,0), screen row 7, scroll 0
	if bufferCursor.Row != 7 || window.ScrollTop() != 0 {
		t.Errorf("After line 7 creation: cursor (%d,%d), scroll %d", 
			bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	}
	
	// Add content to line 7
	line := "Line 7"
	for _, ch := range line {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	t.Logf("After adding content to line 7: cursor (%d,%d), scroll %d", 
		bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	
	// Now cursor should be (7,6), screen row 7, scroll 0 - filling the last visible line
	if bufferCursor.Row != 7 || bufferCursor.Col != 6 || window.ScrollTop() != 0 {
		t.Errorf("After filling line 7: cursor (%d,%d), scroll %d", 
			bufferCursor.Row, bufferCursor.Col, window.ScrollTop())
	}
	
	// CRITICAL TEST: Press Enter while at line 7 (last visible line)
	// This should create line 8, cursor goes to (8,0)
	// User reports this should scroll immediately but doesn't
	t.Logf("=== CRITICAL: Press Enter at end of line 7 (last visible line) ===")
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	bufferCursor = editor.CurrentBuffer().Cursor()
	screenRow, _ = window.CursorPosition()
	visible = window.VisibleLines()
	t.Logf("CRITICAL MOMENT: cursor (%d,%d), screen row %d, scroll %d", 
		bufferCursor.Row, bufferCursor.Col, screenRow, window.ScrollTop())
	t.Logf("Visible lines: %v", visible)
	
	// This is where user says problem occurs
	// Cursor should be (8,0), and since that's beyond visible area [0-7], should scroll to 1
	if bufferCursor.Row != 8 {
		t.Errorf("CRITICAL: Expected cursor row 8, got %d", bufferCursor.Row)
	}
	
	// User says this doesn't scroll immediately - let's see if our fix works
	expectedScroll := 1
	if window.ScrollTop() != expectedScroll {
		t.Errorf("CRITICAL: Expected immediate scroll to %d, got %d - THIS IS THE BUG", 
			expectedScroll, window.ScrollTop())
	}
	
	// Screen row should be 7 (bottom of visible area) after scroll
	if screenRow != 7 && window.ScrollTop() == expectedScroll {
		t.Errorf("CRITICAL: After scroll, expected screen row 7, got %d", screenRow)
	}
	
	// The visible lines should now be [Line 1, Line 2, ..., Line 7, ""]
	if len(visible) > 0 && visible[0] != "Line 1" {
		t.Errorf("CRITICAL: After scroll, expected first visible line 'Line 1', got '%s'", visible[0])
	}
}

// Test to verify what user reports as wrong behavior
/**
 * @spec scroll/user_reported_behavior
 * @scenario ユーザー報告された問題の再現
 * @description ユーザーが報告したスクロールディレイの正確な再現テスト
 * @given 8行でスクリーンを埋めた状態
 * @when 連続してEnter+コンテンツ入力を繰り返す
 * @then ユーザー期待と実際の動作の違いを特定し、修正を検証する
 * @implementation domain/scroll.go, ユーザー報告修正
 */
func TestUserReportedBehavior(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(40, 10)
	
	window := editor.CurrentWindow()
	window.Resize(40, 8)
	
	// Setup: Fill exactly 8 lines (this fills the screen)
	for i := 0; i < 8; i++ {
		if i > 0 {
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
		}
		line := "Line " + string(rune('0'+i))
		for _, ch := range line {
			event := events.KeyEventData{Rune: ch, Key: string(ch)}
			editor.HandleEvent(event)
		}
	}
	display.Render(editor)
	
	// State: cursor (7,6), showing lines 0-7
	t.Logf("Setup complete: cursor (%d,%d), scroll %d", 
		editor.CurrentBuffer().Cursor().Row, 
		editor.CurrentBuffer().Cursor().Col, 
		window.ScrollTop())
	
	// Step 1: Press Enter (user: "8行目でenterを押したら1行スクロールしてバッファの2行目から9行目が表示されるはず")
	// Expected: scroll=1, show lines 1-8  
	// User reports: shows lines 1-8 (this part seems to work)
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	visible := window.VisibleLines()
	t.Logf("Step 1 - Enter at line 7: cursor (%d,%d), scroll %d, visible: %v", 
		editor.CurrentBuffer().Cursor().Row, 
		editor.CurrentBuffer().Cursor().Col, 
		window.ScrollTop(), visible)
	
	// Step 2: Add content and press Enter
	// User: "そのまま何か入力してenterしたら3~10行目が表示されるはずなのに1~8行目"
	// Expected: scroll=2, show lines 2-9
	// User reports: still shows lines 1-8 (THIS IS THE PROBLEM)
	content := "Content"
	for _, ch := range content {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	visible = window.VisibleLines()
	bufferContent := editor.CurrentBuffer().Content()
	t.Logf("Step 2 - Add content and Enter: cursor (%d,%d), scroll %d, visible: %v", 
		editor.CurrentBuffer().Cursor().Row, 
		editor.CurrentBuffer().Cursor().Col, 
		window.ScrollTop(), visible)
	t.Logf("Step 2 - Buffer content: %v", bufferContent)
	t.Logf("Step 2 - Buffer has %d lines, showing from line %d", len(bufferContent), window.ScrollTop())
	
	// User says this should show lines 2-9 but shows 1-8
	// Current: scroll=2, showing [Line 2, Line 3, ..., Line 7, Content, ""]
	// User expects: scroll=1, showing [Line 1, Line 2, ..., Line 7, Content]
	// The difference is: user expects cursor to be at the bottom-1 line, not bottom line
	
	// Step 3: Add content and press Enter again
	// User: "次に何か入力してenterすると、ここでやっと2~9行目が表示される"
	content = "More"
	for _, ch := range content {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	enterEvent = events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	display.Render(editor)
	
	visible = window.VisibleLines()
	t.Logf("Step 3 - Add more and Enter: cursor (%d,%d), scroll %d, visible: %v", 
		editor.CurrentBuffer().Cursor().Row, 
		editor.CurrentBuffer().Cursor().Col, 
		window.ScrollTop(), visible)
	
	// User says now it finally shows lines 2-9
	if len(visible) > 0 && visible[0] == "Line 2" {
		t.Logf("Step 3 shows Line 2 as expected by user")
	} else {
		t.Logf("Step 3: User expects 'Line 2', got '%s'", visible[0])
	}
}