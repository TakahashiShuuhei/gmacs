package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageManagerDownloaderIntegration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_pkg_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create package manager
	manager := NewManager(tempDir)

	// Test that downloader is initialized
	if manager.downloader == nil {
		t.Error("Expected downloader to be initialized")
	}

	// Test that downloader workspace path is correct
	expectedPath := filepath.Join(tempDir, "cache")
	if manager.downloader.workDir != expectedPath {
		t.Errorf("Expected downloader workDir to be %s, got %s", expectedPath, manager.downloader.workDir)
	}

	// Test downloader initialization
	err = manager.downloader.InitializeWorkspace()
	if err != nil {
		t.Errorf("Failed to initialize downloader workspace: %v", err)
	}

	// Check that workspace directory was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("Expected workspace directory to be created")
	}

	// Check that go.mod was created
	goModPath := filepath.Join(expectedPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Error("Expected go.mod to be created")
	}

	// Test getting downloaded packages (should be empty initially)
	packages, err := manager.GetDownloadedPackages()
	if err != nil {
		t.Errorf("Failed to get downloaded packages: %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("Expected 0 downloaded packages, got %d", len(packages))
	}
}

func TestPackageManagerDeclareAndLoad(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_pkg_declare_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create package manager
	manager := NewManager(tempDir)

	// Declare a package
	testURL := "github.com/test/mock-package"
	testVersion := "v1.0.0"
	manager.DeclarePackage(testURL, testVersion, map[string]any{
		"enabled": true,
	})

	// Check declared packages
	declared := manager.GetDeclaredPackages()
	if len(declared) != 1 {
		t.Errorf("Expected 1 declared package, got %d", len(declared))
	}

	if declared[0].URL != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, declared[0].URL)
	}

	if declared[0].Version != testVersion {
		t.Errorf("Expected version %s, got %s", testVersion, declared[0].Version)
	}

	// Note: We don't test LoadDeclaredPackages() here because it requires
	// actual Go packages to download, which would require internet access
	// and make the test unreliable. The download functionality is tested
	// separately in downloader_test.go
}

func TestPackageManagerValidation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_pkg_validation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create package manager
	manager := NewManager(tempDir)

	// Test package URL validation (through downloader)
	testCases := []struct {
		url     string
		version string
		valid   bool
	}{
		{"github.com/user/repo", "v1.0.0", true},
		{"gitlab.com/user/repo", "latest", true},
		{"invalid-url", "v1.0.0", false},
		{"", "v1.0.0", false},
	}

	for _, tc := range testCases {
		err := manager.downloader.validatePackageURL(tc.url)
		if tc.valid && err != nil {
			t.Errorf("Expected URL %s to be valid, got error: %v", tc.url, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("Expected URL %s to be invalid, but got no error", tc.url)
		}
	}
}