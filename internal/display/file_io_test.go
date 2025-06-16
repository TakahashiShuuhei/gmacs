package display

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFileAndSave(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gmacs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	editor := NewEditor()
	
	// Test creating a new file
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Simulate loading a non-existent file (should create new buffer)
	buf := editor.currentWin.Buffer()
	buf.SetFilename(testFile)
	buf.SetText("Hello World\nこんにちは\nTest file")
	
	// Test saving
	err = buf.SaveToFile(testFile)
	if err != nil {
		t.Errorf("Failed to save file: %v", err)
	}
	
	// Verify file was created and has correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}
	
	expectedContent := "Hello World\nこんにちは\nTest file"
	if string(content) != expectedContent {
		t.Errorf("File content mismatch.\nExpected: %q\nGot: %q", expectedContent, string(content))
	}
}

func TestLoadExistingFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gmacs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	testFile := filepath.Join(tempDir, "existing.txt")
	testContent := "Existing file content\n日本語テキスト\nLine 3"
	
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	editor := NewEditor()
	buf := editor.currentWin.Buffer()
	
	// Test loading the file
	err = buf.LoadFromFile(testFile)
	if err != nil {
		t.Errorf("Failed to load file: %v", err)
	}
	
	// Verify content was loaded correctly
	loadedContent := buf.GetText()
	if loadedContent != testContent {
		t.Errorf("Loaded content mismatch.\nExpected: %q\nGot: %q", testContent, loadedContent)
	}
	
	// Verify filename is set correctly
	if buf.Filename() != testFile {
		t.Errorf("Filename not set correctly. Expected: %s, Got: %s", testFile, buf.Filename())
	}
	
	// Verify buffer is not marked as modified after loading
	if buf.IsModified() {
		t.Errorf("Buffer should not be marked as modified after loading")
	}
}

func TestSaveModifiedBuffer(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gmacs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testFile := filepath.Join(tempDir, "modified.txt")
	initialContent := "Initial content"
	
	// Create initial file
	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	editor := NewEditor()
	buf := editor.currentWin.Buffer()
	
	// Load the file
	err = buf.LoadFromFile(testFile)
	if err != nil {
		t.Errorf("Failed to load file: %v", err)
	}
	
	// Modify the content
	buf.SetText("Modified content\n追加されたテキスト")
	
	// Verify buffer is marked as modified
	if !buf.IsModified() {
		t.Errorf("Buffer should be marked as modified after changes")
	}
	
	// Save the buffer
	err = buf.Save()
	if err != nil {
		t.Errorf("Failed to save buffer: %v", err)
	}
	
	// Verify buffer is no longer marked as modified
	if buf.IsModified() {
		t.Errorf("Buffer should not be marked as modified after saving")
	}
	
	// Verify file content was updated
	savedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}
	
	expectedContent := "Modified content\n追加されたテキスト"
	if string(savedContent) != expectedContent {
		t.Errorf("Saved content mismatch.\nExpected: %q\nGot: %q", expectedContent, string(savedContent))
	}
}

func TestFileIOWithUTF8(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "gmacs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testFile := filepath.Join(tempDir, "utf8.txt")
	
	// Test content with various UTF-8 characters
	testContent := "English text\n" +
		"日本語のテキスト\n" +
		"한글 텍스트\n" +
		"Русский текст\n" +
		"🎌🗾📝 Emoji test\n" +
		"Math: α β γ δ ε\n"
	
	editor := NewEditor()
	buf := editor.currentWin.Buffer()
	
	// Set content and save
	buf.SetFilename(testFile)
	buf.SetText(testContent)
	
	err = buf.SaveToFile(testFile)
	if err != nil {
		t.Errorf("Failed to save UTF-8 file: %v", err)
	}
	
	// Load in a new buffer to verify
	newBuf := editor.currentWin.Buffer()
	err = newBuf.LoadFromFile(testFile)
	if err != nil {
		t.Errorf("Failed to load UTF-8 file: %v", err)
	}
	
	loadedContent := newBuf.GetText()
	if loadedContent != testContent {
		t.Errorf("UTF-8 content mismatch.\nExpected: %q\nGot: %q", testContent, loadedContent)
	}
}