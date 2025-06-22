package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec window/vertical_split_display
 * @scenario 垂直分割時の表示検証
 * @description C-x 3による垂直分割時のコンテンツ表示とモードライン確認
 * @given 80x10のターミナル環境
 * @when C-x 3で垂直分割し、左ウィンドウに"abc"を入力
 * @then 両ウィンドウにモードラインが表示され、コンテンツが正常に表示される
 * @implementation domain/window_layout.go, cli/display.go
 */
func TestVerticalSplitDisplay(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 10)
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 80, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	t.Logf("=== Initial state ===")
	display.Render(editor)
	initialContent := display.GetContent()
	t.Logf("Initial content lines: %d", len(initialContent))
	
	// Split window vertically (C-x 3)
	t.Logf("=== Performing C-x 3 split ===")
	// Press C-x
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	
	// Press 3
	splitEvent := events.KeyEventData{Key: "3", Rune: '3'}
	editor.HandleEvent(splitEvent)
	
	display.Render(editor)
	
	// Check layout after split
	layout := editor.Layout()
	windowNodes := layout.GetAllWindowNodes()
	t.Logf("After split: found %d window nodes", len(windowNodes))
	
	for i, node := range windowNodes {
		if node.Window != nil {
			t.Logf("Window %d: position (%d,%d), size %dx%d", 
				i, node.X, node.Y, node.Width, node.Height)
		}
	}
	
	// Input "abc" to the left window
	t.Logf("=== Inputting 'abc' ===")
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
	
	// Verify both windows have mode lines
	modeLineCount := 0
	for _, line := range finalContent {
		// Mode line pattern: contains "*scratch*" and dashes
		if strings.Contains(line, "*scratch*") && strings.Contains(line, "---") {
			modeLineCount++
			t.Logf("Found mode line: %q", line)
		}
	}
	
	if modeLineCount != 1 {
		t.Errorf("Expected 1 mode line (containing both windows), found %d", modeLineCount)
	}
	
	// Verify the mode line contains both window sections
	modeLineFound := false
	for _, line := range finalContent {
		if strings.Contains(line, "*scratch*") && strings.Contains(line, "---") {
			// Count "*scratch*" occurrences to verify both windows are represented
			scratchCount := strings.Count(line, "*scratch*")
			if scratchCount >= 2 {
				t.Logf("✅ Mode line shows both windows: %q", line)
				modeLineFound = true
			} else {
				t.Logf("Mode line found but only shows one window: %q", line)
			}
			break
		}
	}
	
	if !modeLineFound {
		t.Errorf("❌ Mode line doesn't properly show both windows")
	}
	
	// Check if content is visible in left window
	contentFound := false
	for _, line := range finalContent {
		if len(line) > 0 && line[0] == 'a' {
			contentFound = true
			t.Logf("Found content line: %q", line)
			
			// Check if 'a' is visible (not hidden by border)
			if len(line) >= 3 && line[0] == 'a' && line[1] == 'b' && line[2] == 'c' {
				t.Logf("✅ Content 'abc' is fully visible")
			} else {
				t.Errorf("❌ Content appears corrupted: expected 'abc', got start of line: %q", line[:min(10, len(line))])
			}
			break
		}
	}
	
	if !contentFound {
		t.Errorf("❌ Content 'abc' not found in display")
	}
}

/**
 * @spec window/border_positioning 
 * @scenario ウィンドウ境界線の位置確認
 * @description 垂直分割時の境界線が正しい位置に描画されることを確認
 * @given 80x10のターミナルでC-x 3による垂直分割
 * @when 左右ウィンドウのサイズを確認
 * @then 境界線がウィンドウ間の正しい位置に表示される
 * @implementation cli/display.go, renderWindowBorders
 */
func TestWindowBorderPositioning(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 10)
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 80, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	// Split window vertically (C-x 3)
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "3", Rune: '3'}
	editor.HandleEvent(splitEvent)
	
	display.Render(editor)
	
	// Check layout
	layout := editor.Layout()
	windowNodes := layout.GetAllWindowNodes()
	
	if len(windowNodes) != 2 {
		t.Fatalf("Expected 2 windows after split, got %d", len(windowNodes))
	}
	
	leftWindow := windowNodes[0]
	rightWindow := windowNodes[1]
	
	t.Logf("Left window: pos(%d,%d), size %dx%d", 
		leftWindow.X, leftWindow.Y, leftWindow.Width, leftWindow.Height)
	t.Logf("Right window: pos(%d,%d), size %dx%d", 
		rightWindow.X, rightWindow.Y, rightWindow.Width, rightWindow.Height)
	
	// Expected border position should be at leftWindow.X + leftWindow.Width
	expectedBorderX := leftWindow.X + leftWindow.Width
	t.Logf("Expected border at column %d", expectedBorderX)
	
	// Check if border exists and doesn't overlap with content
	content := display.GetContent()
	borderFound := false
	
	for i, line := range content {
		if len(line) > expectedBorderX && string([]rune(line)[expectedBorderX]) == "│" {
			borderFound = true
			t.Logf("Found border at line %d, column %d", i, expectedBorderX)
		}
		
		// Check if content in left window is not overlapped by border
		if len(line) > 0 {
			leftContent := ""
			rightContent := ""
			
			if len(line) > leftWindow.Width {
				leftContent = line[:leftWindow.Width]
				if len(line) > rightWindow.X {
					rightContent = line[rightWindow.X:]
				}
			} else {
				leftContent = line
			}
			
			if leftContent != "" || rightContent != "" {
				t.Logf("Line %d: left='%s' | right='%s'", i, leftContent, rightContent)
			}
		}
	}
	
	if !borderFound {
		t.Errorf("❌ Border not found at expected position %d", expectedBorderX)
	}
	
	// Verify left window width accounts for border
	totalExpectedWidth := leftWindow.Width + 1 + rightWindow.Width // +1 for border
	if totalExpectedWidth > 80 {
		t.Errorf("❌ Total width %d exceeds terminal width 80", totalExpectedWidth)
	}
}

/**
 * @spec window/content_overlap
 * @scenario コンテンツと境界線の重なり検証
 * @description 境界線がコンテンツと重ならないことを確認
 * @given 垂直分割されたウィンドウ
 * @when 両ウィンドウにコンテンツを入力
 * @then コンテンツが境界線と重ならず正しく表示される
 * @implementation domain/window.go, cli/display.go
 */
func TestContentBorderOverlap(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(60, 8) // Smaller terminal to make issues more apparent
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 60, Height: 8}
	editor.HandleEvent(resizeEvent)
	
	// Split window vertically (C-x 3)
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "3", Rune: '3'}
	editor.HandleEvent(splitEvent)
	
	// Add content to fill the left window width
	longText := "This is a long line that should fill the window width"
	for _, ch := range longText {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	
	// Switch to right window and add content
	ctrlXEvent = events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	oEvent := events.KeyEventData{Key: "o", Rune: 'o'}
	editor.HandleEvent(oEvent)
	
	// Add content to right window
	rightText := "RIGHT"
	for _, ch := range rightText {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	display.Render(editor)
	content := display.GetContent()
	
	t.Logf("=== Content with both windows filled ===")
	for i, line := range content {
		t.Logf("Line %d: %q", i, line)
	}
	
	// Analyze layout
	layout := editor.Layout()
	windowNodes := layout.GetAllWindowNodes()
	
	if len(windowNodes) != 2 {
		t.Fatalf("Expected 2 windows, got %d", len(windowNodes))
	}
	
	leftWindow := windowNodes[0]
	rightWindow := windowNodes[1]
	borderX := leftWindow.X + leftWindow.Width
	
	t.Logf("Left window width: %d, Right window start: %d, Border at: %d", 
		leftWindow.Width, rightWindow.X, borderX)
	
	// Check for content overlap with border
	for i, line := range content {
		runes := []rune(line)
		if len(runes) > borderX {
			if borderX > 0 && borderX < len(runes) {
				charAtBorder := runes[borderX]
				if string(charAtBorder) != "│" && charAtBorder != ' ' {
					t.Errorf("❌ Line %d: Content character '%c' at border position %d", 
						i, charAtBorder, borderX)
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}