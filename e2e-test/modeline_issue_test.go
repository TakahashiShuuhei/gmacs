package test

import (
	"fmt"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
)

/**
 * @spec window/modeline_visibility
 * @scenario モードライン消失問題の調査
 * @description 実際の画面表示でモードラインが消える原因を特定
 * @given ウィンドウ分割後の状態
 * @when レンダリング処理を詳細に追跡
 * @then モードラインが正しく表示されることを確認
 * @implementation cli/display.go, MockDisplay比較
 */
func TestModeLineVisibilityIssue(t *testing.T) {
	editor := NewEditorWithDefaults()
	display := NewMockDisplay(80, 10)
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 80, Height: 10}
	editor.HandleEvent(resizeEvent)
	
	t.Logf("=== Before split ===")
	display.Render(editor)
	beforeContent := display.GetContent()
	
	// Count lines with content before split
	beforeLines := 0
	for i, line := range beforeContent {
		if len(line) > 0 {
			beforeLines++
			t.Logf("Before line %d: %q", i, line)
		}
	}
	
	// Check layout before split
	layout := editor.Layout()
	beforeNodes := layout.GetAllWindowNodes()
	t.Logf("Before split: %d window nodes", len(beforeNodes))
	for i, node := range beforeNodes {
		t.Logf("Before node %d: pos(%d,%d), size %dx%d", 
			i, node.X, node.Y, node.Width, node.Height)
	}
	
	// Split window vertically (C-x 3)
	t.Logf("=== Performing C-x 3 split ===")
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "3", Rune: '3'}
	editor.HandleEvent(splitEvent)
	
	t.Logf("=== After split ===")
	display.Render(editor)
	afterContent := display.GetContent()
	
	// Count lines with content after split
	afterLines := 0
	for i, line := range afterContent {
		if len(line) > 0 {
			afterLines++
			t.Logf("After line %d: %q", i, line)
		}
	}
	
	// Check layout after split
	afterNodes := layout.GetAllWindowNodes()
	t.Logf("After split: %d window nodes", len(afterNodes))
	for i, node := range afterNodes {
		t.Logf("After node %d: pos(%d,%d), size %dx%d", 
			i, node.X, node.Y, node.Width, node.Height)
	}
	
	// Verify we have the expected number of windows
	if len(afterNodes) != 2 {
		t.Fatalf("Expected 2 windows after split, got %d", len(afterNodes))
	}
	
	// Analyze each window's expected mode line position
	for i, node := range afterNodes {
		if node.Window != nil {
			_, contentHeight := node.Window.Size()
			expectedModeLineRow := node.Y + contentHeight
			t.Logf("Window %d: content ends at row %d, mode line expected at row %d", 
				i, node.Y+contentHeight-1, expectedModeLineRow)
			
			// Check if there's content at the expected mode line position
			if expectedModeLineRow < len(afterContent) {
				modeLineContent := afterContent[expectedModeLineRow]
				if len(modeLineContent) > 0 {
					t.Logf("Window %d mode line content: %q", i, modeLineContent)
				} else {
					t.Errorf("❌ Window %d: Empty mode line at row %d", i, expectedModeLineRow)
				}
			} else {
				t.Errorf("❌ Window %d: Mode line row %d exceeds display height %d", 
					i, expectedModeLineRow, len(afterContent))
			}
		}
	}
	
	// Look for mode line patterns in the content
	modeLinePatterns := 0
	for i, line := range afterContent {
		if len(line) > 0 && (line[0] == ' ' || (len(line) > 1 && line[1] == '*')) {
			// Potential mode line
			hasBuffer := false
			hasDashes := false
			for j := 0; j < len(line); j++ {
				if j < len(line)-8 && line[j:j+9] == "*scratch*" {
					hasBuffer = true
				}
				if line[j] == '-' {
					hasDashes = true
				}
			}
			if hasBuffer && hasDashes {
				modeLinePatterns++
				t.Logf("✅ Found mode line pattern at row %d: %q", i, line)
			}
		}
	}
	
	t.Logf("=== Summary ===")
	t.Logf("Before split: %d content lines", beforeLines)
	t.Logf("After split: %d content lines", afterLines)
	t.Logf("Mode line patterns found: %d", modeLinePatterns)
	
	if modeLinePatterns == 0 {
		t.Errorf("❌ No mode line patterns found after split")
	}
}

/**
 * @spec window/render_order_analysis
 * @scenario レンダリング順序の詳細分析
 * @description ウィンドウ、モードライン、境界線の描画順序を確認
 * @given 垂直分割された状態
 * @when 各レンダリングステップを個別に実行
 * @then 正しい順序で描画されることを確認
 * @implementation cli/display.go, レンダリング順序
 */
func TestRenderOrderAnalysis(t *testing.T) {
	editor := NewEditorWithDefaults()
	
	// Setup terminal size
	resizeEvent := events.ResizeEventData{Width: 60, Height: 8}
	editor.HandleEvent(resizeEvent)
	
	// Split window vertically
	ctrlXEvent := events.KeyEventData{Key: "x", Ctrl: true}
	editor.HandleEvent(ctrlXEvent)
	splitEvent := events.KeyEventData{Key: "3", Rune: '3'}
	editor.HandleEvent(splitEvent)
	
	// Add some content to make it visible
	for _, ch := range "LEFT" {
		event := events.KeyEventData{Rune: ch, Key: string(ch)}
		editor.HandleEvent(event)
	}
	
	// Test with a custom mock that tracks render operations
	trackingDisplay := NewTrackingMockDisplay(60, 8)
	trackingDisplay.Render(editor)
	
	t.Logf("=== Render operation sequence ===")
	for i, op := range trackingDisplay.operations {
		t.Logf("Op %d: %s", i, op)
	}
	
	t.Logf("=== Final content ===")
	content := trackingDisplay.GetContent()
	for i, line := range content {
		t.Logf("Line %d: %q", i, line)
	}
}

// TrackingMockDisplay tracks the sequence of render operations
type TrackingMockDisplay struct {
	*MockDisplay
	operations []string
}

func NewTrackingMockDisplay(width, height int) *TrackingMockDisplay {
	return &TrackingMockDisplay{
		MockDisplay: NewMockDisplay(width, height),
		operations:  make([]string, 0),
	}
}

func (d *TrackingMockDisplay) renderWindow(node *domain.WindowLayoutNode) {
	d.operations = append(d.operations, 
		fmt.Sprintf("renderWindow: pos(%d,%d), size %dx%d", 
			node.X, node.Y, node.Width, node.Height))
	d.MockDisplay.renderWindow(node)
}

func (d *TrackingMockDisplay) renderWindowModeLine(node *domain.WindowLayoutNode) {
	_, contentHeight := node.Window.Size()
	modeLineRow := node.Y + contentHeight
	d.operations = append(d.operations, 
		fmt.Sprintf("renderWindowModeLine: window pos(%d,%d), mode line at row %d", 
			node.X, node.Y, modeLineRow))
	d.MockDisplay.renderWindowModeLine(node)
}

func (d *TrackingMockDisplay) renderWindowBorders(layout *domain.WindowLayout) {
	d.operations = append(d.operations, "renderWindowBorders: start")
	d.MockDisplay.renderWindowBorders(layout)
	d.operations = append(d.operations, "renderWindowBorders: end")
}