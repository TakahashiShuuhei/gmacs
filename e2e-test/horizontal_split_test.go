package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec window/horizontal_split_display
 * @scenario 水平分割時の表示検証
 * @description C-x 2による水平分割時のコンテンツ表示確認
 * @given 80x10のターミナル環境
 * @when C-x 2で水平分割し、上ウィンドウに"abc"を入力
 * @then 下ウィンドウの1行目がコンテンツで隠されていないことを確認
 * @implementation domain/window_layout.go, cli/display.go
 */
func TestHorizontalSplitDisplay(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(80, 10)
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 80, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	t.Logf("=== Initial state ===")
	display.Render(editor)
	
	// Split window horizontally (C-x 2)
	t.Logf("=== Performing C-x 2 split ===")
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "2", Rune: '2'}
	editor.HandleEvent(splitEvent)
	
	display.Render(editor)
	
	// Check layout after split
	layout := editor.Layout()
	windowNodes := layout.GetAllWindowNodes()
	t.Logf("After split: found %d window nodes", len(windowNodes))
	
	if len(windowNodes) != 2 {
		t.Fatalf("Expected 2 windows after split, got %d", len(windowNodes))
	}
	
	topWindow := windowNodes[0]
	bottomWindow := windowNodes[1]
	
	t.Logf("Top window: pos(%d,%d), size %dx%d", 
		topWindow.X, topWindow.Y, topWindow.Width, topWindow.Height)
	t.Logf("Bottom window: pos(%d,%d), size %dx%d", 
		bottomWindow.X, bottomWindow.Y, bottomWindow.Width, bottomWindow.Height)
	
	// Input "abc" to the top window
	t.Logf("=== Inputting 'abc' to top window ===")
	for _, ch := range "abc" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	finalContent := display.GetContent()
	
	t.Logf("=== Final display content ===")
	for i, line := range finalContent {
		t.Logf("Line %d: %q", i, line)
	}
	
	// Verify top window shows content
	topContentFound := false
	for i := topWindow.Y; i < topWindow.Y+topWindow.Height; i++ {
		if i < len(finalContent) && strings.Contains(finalContent[i], "abc") {
			topContentFound = true
			t.Logf("✅ Top window content found at line %d: %q", i, finalContent[i])
			break
		}
	}
	
	if !topContentFound {
		t.Errorf("❌ Top window content 'abc' not found")
	}
	
	// Verify bottom window content area is not obscured
	_, topContentHeight := topWindow.Window.Size()
	topModeLineRow := topWindow.Y + topContentHeight
	bottomContentStart := bottomWindow.Y
	
	t.Logf("Top mode line at row %d, bottom content starts at row %d", 
		topModeLineRow, bottomContentStart)
	
	// Check if there's a horizontal line blocking bottom window content
	if bottomContentStart < len(finalContent) {
		bottomFirstLine := finalContent[bottomContentStart]
		hasHorizontalLine := strings.Contains(bottomFirstLine, "─")
		
		if hasHorizontalLine {
			t.Errorf("❌ Bottom window first line is blocked by horizontal line: %q", bottomFirstLine)
		} else {
			t.Logf("✅ Bottom window first line is clear: %q", bottomFirstLine)
		}
	}
	
	// Switch to bottom window and add content
	t.Logf("=== Switching to bottom window ===")
	ctrlXEvent = events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	oEvent := events.KeyEventData{Key: "o", Rune: 'o'}
	editor.HandleEvent(oEvent)
	
	// Add content to bottom window
	for _, ch := range "xyz" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	finalContentWithBoth := display.GetContent()
	
	t.Logf("=== Content with both windows filled ===")
	for i, line := range finalContentWithBoth {
		t.Logf("Line %d: %q", i, line)
	}
	
	// Verify bottom window shows its content
	bottomContentFound := false
	for i := bottomWindow.Y; i < bottomWindow.Y+bottomWindow.Height; i++ {
		if i < len(finalContentWithBoth) && strings.Contains(finalContentWithBoth[i], "xyz") {
			bottomContentFound = true
			t.Logf("✅ Bottom window content found at line %d: %q", i, finalContentWithBoth[i])
			break
		}
	}
	
	if !bottomContentFound {
		t.Errorf("❌ Bottom window content 'xyz' not found")
	}
}

/**
 * @spec window/horizontal_border_analysis
 * @scenario 水平分割時の境界線分析
 * @description モードラインが境界として機能することを確認
 * @given 水平分割された状態
 * @when 各ウィンドウの境界を分析
 * @then モードラインが適切に境界として機能することを確認
 * @implementation cli/display.go, ボーダー描画ロジック
 */
func TestHorizontalBorderAnalysis(t *testing.T) {
	editor := domain.NewEditor()
	display := NewMockDisplay(60, 8) // Smaller for easier analysis
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 60, Height: 8}
	editor.HandleEvent(resizeEvent)
	
	// Split window horizontally
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "2", Rune: '2'}
	editor.HandleEvent(splitEvent)
	
	display.Render(editor)
	content := display.GetContent()
	
	t.Logf("=== Horizontal split layout analysis ===")
	for i, line := range content {
		t.Logf("Line %d: %q", i, line)
	}
	
	// Count horizontal line characters
	horizontalLineCount := 0
	modeLineCount := 0
	
	for i, line := range content {
		hasHorizontalChars := strings.Contains(line, "─")
		hasModeLinePattern := strings.Contains(line, "*scratch*") && strings.Contains(line, "---")
		
		if hasHorizontalChars {
			horizontalLineCount++
			t.Logf("Found horizontal line chars at row %d: %q", i, line)
		}
		
		if hasModeLinePattern {
			modeLineCount++
			t.Logf("Found mode line at row %d: %q", i, line)
		}
	}
	
	t.Logf("=== Analysis results ===")
	t.Logf("Horizontal line chars found: %d", horizontalLineCount)
	t.Logf("Mode lines found: %d", modeLineCount)
	
	if horizontalLineCount > 0 {
		t.Logf("⚠️  Horizontal lines detected - may be unnecessary for horizontal splits")
	}
	
	if modeLineCount < 2 {
		t.Errorf("❌ Expected 2 mode lines for horizontal split, found %d", modeLineCount)
	} else {
		t.Logf("✅ Correct number of mode lines found")
	}
}