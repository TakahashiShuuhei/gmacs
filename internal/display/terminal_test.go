package display

import (
	"strings"
	"testing"
)

func TestTerminal(t *testing.T) {
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	terminal := NewTerminal(input, output)
	
	// Test basic properties
	width, height := terminal.Size()
	if width <= 0 || height <= 0 {
		t.Errorf("Expected positive dimensions, got %dx%d", width, height)
	}
	
	// Test cursor movement
	terminal.MoveCursor(5, 10)
	line, col := terminal.GetCursorPos()
	if line != 5 || col != 10 {
		t.Errorf("Expected cursor at (5,10), got (%d,%d)", line, col)
	}
	
	// Test printing
	terminal.Print("Hello")
	if !strings.Contains(output.String(), "Hello") {
		t.Error("Output should contain 'Hello'")
	}
}

func TestTerminalColors(t *testing.T) {
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	terminal := NewTerminal(input, output)
	
	// Test color setting
	terminal.SetColor(ColorRed, ColorBlue)
	terminal.Print("Colored text")
	terminal.ResetColor()
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "\033[3") { // Should contain color escape sequence
		t.Error("Output should contain color escape sequences")
	}
}

func TestTerminalBoxDrawing(t *testing.T) {
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	terminal := NewTerminal(input, output)
	
	terminal.DrawBox(1, 1, 10, 5, "Test")
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "┌") || !strings.Contains(outputStr, "┐") {
		t.Error("Box drawing should contain box characters")
	}
}

func TestTerminalClear(t *testing.T) {
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	terminal := NewTerminal(input, output)
	
	terminal.Clear()
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "\033[2J") { // Clear screen escape sequence
		t.Error("Clear should contain clear screen escape sequence")
	}
}