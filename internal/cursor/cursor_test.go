package cursor

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	
	if c.Line() != 0 {
		t.Errorf("Expected line 0, got %d", c.Line())
	}
	
	if c.Col() != 0 {
		t.Errorf("Expected col 0, got %d", c.Col())
	}
	
	if c.HasMark() {
		t.Error("New cursor should not have mark")
	}
}

func TestMovement(t *testing.T) {
	c := New()
	
	// Test right movement
	c.MoveRight()
	if c.Col() != 1 {
		t.Errorf("Expected col 1 after MoveRight, got %d", c.Col())
	}
	
	// Test down movement
	c.MoveDown()
	if c.Line() != 1 {
		t.Errorf("Expected line 1 after MoveDown, got %d", c.Line())
	}
	
	// Test left movement
	c.MoveLeft()
	if c.Col() != 0 {
		t.Errorf("Expected col 0 after MoveLeft, got %d", c.Col())
	}
	
	// Test up movement
	c.MoveUp()
	if c.Line() != 0 {
		t.Errorf("Expected line 0 after MoveUp, got %d", c.Line())
	}
}

func TestSetPoint(t *testing.T) {
	c := New()
	c.SetPoint(5, 10)
	
	if c.Line() != 5 {
		t.Errorf("Expected line 5, got %d", c.Line())
	}
	
	if c.Col() != 10 {
		t.Errorf("Expected col 10, got %d", c.Col())
	}
	
	// Test negative values are handled
	c.SetPoint(-1, -1)
	if c.Line() != 0 || c.Col() != 0 {
		t.Errorf("Expected (0,0) for negative values, got (%d,%d)", c.Line(), c.Col())
	}
}

func TestBeginningEndOfLine(t *testing.T) {
	c := New()
	c.SetPoint(2, 5)
	
	c.BeginningOfLine()
	if c.Col() != 0 {
		t.Errorf("Expected col 0 after BeginningOfLine, got %d", c.Col())
	}
	if c.Line() != 2 {
		t.Errorf("Expected line unchanged at 2, got %d", c.Line())
	}
	
	c.EndOfLine(20)
	if c.Col() != 20 {
		t.Errorf("Expected col 20 after EndOfLine(20), got %d", c.Col())
	}
}

func TestMark(t *testing.T) {
	c := New()
	c.SetPoint(3, 7)
	
	if c.HasMark() {
		t.Error("Should not have mark initially")
	}
	
	c.SetMark()
	if !c.HasMark() {
		t.Error("Should have mark after SetMark")
	}
	
	mark := c.Mark()
	if mark.Line != 3 || mark.Col != 7 {
		t.Errorf("Expected mark at (3,7), got (%d,%d)", mark.Line, mark.Col)
	}
	
	c.ClearMark()
	if c.HasMark() {
		t.Error("Should not have mark after ClearMark")
	}
}

func TestGetRegion(t *testing.T) {
	c := New()
	c.SetPoint(2, 5)
	c.SetMark()
	c.SetPoint(4, 3)
	
	start, end, hasRegion := c.GetRegion()
	if !hasRegion {
		t.Error("Should have region")
	}
	
	// Region should be ordered (start before end)
	if start.Line != 2 || start.Col != 5 {
		t.Errorf("Expected start at (2,5), got (%d,%d)", start.Line, start.Col)
	}
	
	if end.Line != 4 || end.Col != 3 {
		t.Errorf("Expected end at (4,3), got (%d,%d)", end.Line, end.Col)
	}
	
	// Test reverse case (cursor before mark)
	c.SetPoint(1, 2)
	start, end, hasRegion = c.GetRegion()
	if !hasRegion {
		t.Error("Should have region")
	}
	
	if start.Line != 1 || start.Col != 2 {
		t.Errorf("Expected start at (1,2), got (%d,%d)", start.Line, start.Col)
	}
	
	if end.Line != 2 || end.Col != 5 {
		t.Errorf("Expected end at (2,5), got (%d,%d)", end.Line, end.Col)
	}
}