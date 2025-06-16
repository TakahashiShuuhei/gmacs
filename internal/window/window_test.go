package window

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
)

func TestNew(t *testing.T) {
	buf := buffer.New("test")
	win := New(buf, 10, 80)
	
	if win.Height() != 10 {
		t.Errorf("Expected height 10, got %d", win.Height())
	}
	
	if win.Width() != 80 {
		t.Errorf("Expected width 80, got %d", win.Width())
	}
	
	if win.Buffer().Name() != "test" {
		t.Errorf("Expected buffer name 'test', got '%s'", win.Buffer().Name())
	}
	
	if win.TopLine() != 0 {
		t.Errorf("Expected top line 0, got %d", win.TopLine())
	}
}

func TestSetSize(t *testing.T) {
	buf := buffer.New("test")
	win := New(buf, 10, 80)
	
	win.SetSize(20, 120)
	
	if win.Height() != 20 {
		t.Errorf("Expected height 20, got %d", win.Height())
	}
	
	if win.Width() != 120 {
		t.Errorf("Expected width 120, got %d", win.Width())
	}
	
	// Test minimum size constraints
	win.SetSize(0, 0)
	if win.Height() != 1 || win.Width() != 1 {
		t.Errorf("Expected minimum size (1,1), got (%d,%d)", win.Height(), win.Width())
	}
}

func TestVisibleLines(t *testing.T) {
	buf := buffer.New("test")
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	win := New(buf, 5, 80) // 5 lines visible
	
	start, end := win.VisibleLines()
	if start != 0 || end != 4 {
		t.Errorf("Expected visible lines (0,4), got (%d,%d)", start, end)
	}
	
	// Test scrolling
	win.SetTopLine(3)
	start, end = win.VisibleLines()
	if start != 3 || end != 7 {
		t.Errorf("Expected visible lines (3,7) after scroll, got (%d,%d)", start, end)
	}
}

func TestIsLineVisible(t *testing.T) {
	buf := buffer.New("test")
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	win := New(buf, 5, 80)
	
	// Lines 0-4 should be visible initially
	for i := 0; i < 5; i++ {
		if !win.IsLineVisible(i) {
			t.Errorf("Line %d should be visible", i)
		}
	}
	
	// Lines 5+ should not be visible
	for i := 5; i < 10; i++ {
		if win.IsLineVisible(i) {
			t.Errorf("Line %d should not be visible", i)
		}
	}
}

func TestScrolling(t *testing.T) {
	buf := buffer.New("test")
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	win := New(buf, 5, 80)
	
	// Test scroll down
	win.ScrollDown(2)
	if win.TopLine() != 2 {
		t.Errorf("Expected top line 2 after ScrollDown(2), got %d", win.TopLine())
	}
	
	// Test scroll up
	win.ScrollUp(1)
	if win.TopLine() != 1 {
		t.Errorf("Expected top line 1 after ScrollUp(1), got %d", win.TopLine())
	}
	
	// Test scroll beyond limits
	win.ScrollUp(10)
	if win.TopLine() != 0 {
		t.Errorf("Expected top line 0 after scrolling up beyond limit, got %d", win.TopLine())
	}
}

func TestEnsureCursorVisible(t *testing.T) {
	buf := buffer.New("test")
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	win := New(buf, 5, 80)
	
	// Move cursor below visible area
	win.Cursor().SetPoint(7, 0)
	win.EnsureCursorVisible()
	
	// Window should scroll to make line 7 visible
	if !win.IsLineVisible(7) {
		t.Error("Cursor line should be visible after EnsureCursorVisible")
	}
	
	// Move cursor above visible area
	win.Cursor().SetPoint(0, 0)
	win.EnsureCursorVisible()
	
	if win.TopLine() != 0 {
		t.Errorf("Expected top line 0 when cursor at top, got %d", win.TopLine())
	}
}

func TestGetVisibleText(t *testing.T) {
	buf := buffer.New("test")
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	win := New(buf, 3, 80)
	
	visibleText := win.GetVisibleText()
	expected := []string{"Line 1", "Line 2", "Line 3"}
	
	if len(visibleText) != len(expected) {
		t.Errorf("Expected %d visible lines, got %d", len(expected), len(visibleText))
	}
	
	for i, line := range expected {
		if i < len(visibleText) && visibleText[i] != line {
			t.Errorf("Expected line %d to be '%s', got '%s'", i, line, visibleText[i])
		}
	}
}

func TestCursorScreenPosition(t *testing.T) {
	buf := buffer.New("test")
	// Add enough lines to allow scrolling
	buf.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12")
	win := New(buf, 5, 80) // Smaller window
	win.SetLeftMargin(4)
	
	win.Cursor().SetPoint(2, 5)
	win.SetTopLine(1)
	
	screenLine, screenCol := win.CursorScreenPosition()
	
	// Screen line should be cursor line - top line
	expectedScreenLine := 2 - 1 // 1
	if screenLine != expectedScreenLine {
		t.Errorf("Expected screen line %d, got %d", expectedScreenLine, screenLine)
	}
	
	// Screen col should be cursor col + left margin
	expectedScreenCol := 5 + 4 // 9
	if screenCol != expectedScreenCol {
		t.Errorf("Expected screen col %d, got %d", expectedScreenCol, screenCol)
	}
}