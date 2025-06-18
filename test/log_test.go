package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TakahashiShuuhei/gmacs/core/log"
)

func TestLoggerCreation(t *testing.T) {
	logger, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
}

func TestLogLevels(t *testing.T) {
	logger, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.SetLevel(log.InfoLevel)

	// Debug should not be logged
	logger.Debug("Debug message")
	
	// Info, Warn, Error should be logged
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
}

func TestLogFileCreation(t *testing.T) {
	logger, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Test log message")

	// Check if logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		t.Fatal("Logs directory should be created")
	}

	// Check if log file exists
	files, err := filepath.Glob("logs/gmacs_*.log")
	if err != nil {
		t.Fatalf("Failed to check log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("At least one log file should exist")
	}
}

func TestLogFileNaming(t *testing.T) {
	logger, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Wait a bit and create another logger
	time.Sleep(1100 * time.Millisecond)
	
	logger2, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create second logger: %v", err)
	}
	defer logger2.Close()

	files, err := filepath.Glob("logs/gmacs_*.log")
	if err != nil {
		t.Fatalf("Failed to check log files: %v", err)
	}

	// Should have at least 2 different log files
	if len(files) < 2 {
		t.Fatalf("Expected at least 2 log files, got %d", len(files))
	}

	// Check filename format
	for _, file := range files {
		filename := filepath.Base(file)
		if !strings.HasPrefix(filename, "gmacs_") || !strings.HasSuffix(filename, ".log") {
			t.Errorf("Invalid log filename format: %s", filename)
		}
	}
}

func TestGlobalLogger(t *testing.T) {
	// Clean up any existing global logger
	log.Close()

	err := log.Init()
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}
	defer log.Close()

	// Test global logging functions
	log.Debug("Global debug")
	log.Info("Global info")
	log.Warn("Global warn")
	log.Error("Global error")

	// Should not panic
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    log.Level
		expected string
	}{
		{log.DebugLevel, "DEBUG"},
		{log.InfoLevel, "INFO"},
		{log.WarnLevel, "WARN"},
		{log.ErrorLevel, "ERROR"},
	}

	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestLoggerClose(t *testing.T) {
	logger, err := log.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}

	// Should be safe to close again
	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close logger twice: %v", err)
	}
}

func cleanup() {
	os.RemoveAll("logs")
}

func TestMain(m *testing.M) {
	cleanup()
	code := m.Run()
	cleanup()
	os.Exit(code)
}