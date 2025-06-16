package buffer

import (
	"testing"
)

func TestNew(t *testing.T) {
	buf := New("test-buffer")
	
	if buf.Name() != "test-buffer" {
		t.Errorf("Expected name 'test-buffer', got '%s'", buf.Name())
	}
	
	if buf.LineCount() != 1 {
		t.Errorf("Expected line count 1, got %d", buf.LineCount())
	}
	
	if buf.GetLine(0) != "" {
		t.Errorf("Expected empty first line, got '%s'", buf.GetLine(0))
	}
	
	if buf.IsModified() {
		t.Error("New buffer should not be modified")
	}
}

func TestSetText(t *testing.T) {
	buf := New("test")
	text := "Hello, World!\nThis is a test buffer.\nLine 3"
	buf.SetText(text)
	
	if buf.LineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buf.LineCount())
	}
	
	if buf.GetLine(0) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buf.GetLine(0))
	}
	
	if buf.GetLine(1) != "This is a test buffer." {
		t.Errorf("Expected 'This is a test buffer.', got '%s'", buf.GetLine(1))
	}
	
	if buf.GetLine(2) != "Line 3" {
		t.Errorf("Expected 'Line 3', got '%s'", buf.GetLine(2))
	}
	
	if !buf.IsModified() {
		t.Error("Buffer should be modified after SetText")
	}
}

func TestInsertLine(t *testing.T) {
	buf := New("test")
	buf.SetText("Line 1\nLine 2\nLine 3")
	
	buf.InsertLine(1, "Inserted line")
	
	if buf.LineCount() != 4 {
		t.Errorf("Expected 4 lines after insertion, got %d", buf.LineCount())
	}
	
	if buf.GetLine(1) != "Inserted line" {
		t.Errorf("Expected 'Inserted line' at line 1, got '%s'", buf.GetLine(1))
	}
	
	if buf.GetLine(2) != "Line 2" {
		t.Errorf("Expected 'Line 2' at line 2, got '%s'", buf.GetLine(2))
	}
}

func TestDeleteLine(t *testing.T) {
	buf := New("test")
	buf.SetText("Line 1\nLine 2\nLine 3")
	
	buf.DeleteLine(1)
	
	if buf.LineCount() != 2 {
		t.Errorf("Expected 2 lines after deletion, got %d", buf.LineCount())
	}
	
	if buf.GetLine(1) != "Line 3" {
		t.Errorf("Expected 'Line 3' at line 1, got '%s'", buf.GetLine(1))
	}
}

func TestGetText(t *testing.T) {
	buf := New("test")
	text := "Hello\nWorld\nTest"
	buf.SetText(text)
	
	result := buf.GetText()
	if result != text {
		t.Errorf("Expected '%s', got '%s'", text, result)
	}
}

func TestReadOnly(t *testing.T) {
	buf := New("test")
	buf.SetReadOnly(true)
	
	if !buf.IsReadOnly() {
		t.Error("Buffer should be read-only")
	}
	
	// Test that modifications don't work on read-only buffer
	originalModified := buf.IsModified()
	buf.SetText("Should not change")
	
	if buf.IsModified() != originalModified {
		t.Error("Read-only buffer should not be marked as modified")
	}
}