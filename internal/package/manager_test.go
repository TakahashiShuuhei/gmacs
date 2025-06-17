package pkg

import (
	"testing"
)

func TestPackageManager_DeclarePackage(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Declare a package
	pm.DeclarePackage("github.com/user/test-package", "v1.0.0", nil)
	
	declared := pm.GetDeclaredPackages()
	if len(declared) != 1 {
		t.Errorf("Expected 1 declared package, got %d", len(declared))
	}
	
	if declared[0].URL != "github.com/user/test-package" {
		t.Errorf("Expected URL 'github.com/user/test-package', got '%s'", declared[0].URL)
	}
	
	if declared[0].Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", declared[0].Version)
	}
}

func TestPackageManager_DeclareMultiplePackages(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Declare multiple packages
	pm.DeclarePackage("github.com/user/package1", "v1.0.0", nil)
	pm.DeclarePackage("github.com/user/package2", "v2.0.0", map[string]interface{}{
		"enabled": true,
	})
	
	declared := pm.GetDeclaredPackages()
	if len(declared) != 2 {
		t.Errorf("Expected 2 declared packages, got %d", len(declared))
	}
	
	// Check first package
	if declared[0].URL != "github.com/user/package1" {
		t.Errorf("Expected first package URL 'github.com/user/package1', got '%s'", declared[0].URL)
	}
	
	// Check second package
	if declared[1].URL != "github.com/user/package2" {
		t.Errorf("Expected second package URL 'github.com/user/package2', got '%s'", declared[1].URL)
	}
	
	if declared[1].Config == nil {
		t.Error("Expected second package to have config")
	}
}

func TestPackageManager_LoadDeclaredPackages(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Declare a package
	pm.DeclarePackage("github.com/user/test-package", "v1.0.0", nil)
	
	// Load declared packages
	err := pm.LoadDeclaredPackages()
	if err != nil {
		t.Errorf("Expected no error loading packages, got: %v", err)
	}
	
	// Check that package was loaded
	loadedPkg, exists := pm.GetLoadedPackage("github.com/user/test-package")
	if !exists {
		t.Error("Expected package to be loaded")
	}
	
	if loadedPkg.Status != PackageStatusEnabled {
		t.Errorf("Expected package status to be enabled, got %s", loadedPkg.Status)
	}
	
	if loadedPkg.Package == nil {
		t.Error("Expected loaded package to have a Package object")
	}
	
	// Check package info
	info := loadedPkg.Package.GetInfo()
	if info.URL != "github.com/user/test-package" {
		t.Errorf("Expected package URL 'github.com/user/test-package', got '%s'", info.URL)
	}
}

func TestPackageManager_GetAllLoadedPackages(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Declare multiple packages
	pm.DeclarePackage("github.com/user/package1", "v1.0.0", nil)
	pm.DeclarePackage("github.com/user/package2", "v2.0.0", nil)
	
	// Load declared packages
	err := pm.LoadDeclaredPackages()
	if err != nil {
		t.Errorf("Expected no error loading packages, got: %v", err)
	}
	
	// Get all loaded packages
	allLoaded := pm.GetAllLoadedPackages()
	if len(allLoaded) != 2 {
		t.Errorf("Expected 2 loaded packages, got %d", len(allLoaded))
	}
	
	// Check that both packages exist
	_, exists1 := allLoaded["github.com/user/package1"]
	_, exists2 := allLoaded["github.com/user/package2"]
	
	if !exists1 {
		t.Error("Expected package1 to be loaded")
	}
	
	if !exists2 {
		t.Error("Expected package2 to be loaded")
	}
}

func TestPackageManager_EnableDisablePackage(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Declare and load a package
	pm.DeclarePackage("github.com/user/test-package", "v1.0.0", nil)
	err := pm.LoadDeclaredPackages()
	if err != nil {
		t.Errorf("Expected no error loading packages, got: %v", err)
	}
	
	packageURL := "github.com/user/test-package"
	
	// Package should be enabled by default
	loadedPkg, _ := pm.GetLoadedPackage(packageURL)
	if loadedPkg.Status != PackageStatusEnabled {
		t.Errorf("Expected package to be enabled by default, got %s", loadedPkg.Status)
	}
	
	// Disable package
	err = pm.DisablePackage(packageURL)
	if err != nil {
		t.Errorf("Expected no error disabling package, got: %v", err)
	}
	
	loadedPkg, _ = pm.GetLoadedPackage(packageURL)
	if loadedPkg.Status != PackageStatusDisabled {
		t.Errorf("Expected package to be disabled, got %s", loadedPkg.Status)
	}
	
	// Re-enable package
	err = pm.EnablePackage(packageURL)
	if err != nil {
		t.Errorf("Expected no error enabling package, got: %v", err)
	}
	
	loadedPkg, _ = pm.GetLoadedPackage(packageURL)
	if loadedPkg.Status != PackageStatusEnabled {
		t.Errorf("Expected package to be enabled, got %s", loadedPkg.Status)
	}
}

func TestPackageManager_PackageNotFound(t *testing.T) {
	pm := NewManager("/tmp/gmacs-test")
	
	// Try to enable a package that doesn't exist
	err := pm.EnablePackage("github.com/user/nonexistent-package")
	if err == nil {
		t.Error("Expected error when enabling nonexistent package")
	}
	
	// Try to disable a package that doesn't exist
	err = pm.DisablePackage("github.com/user/nonexistent-package")
	if err == nil {
		t.Error("Expected error when disabling nonexistent package")
	}
}

func TestPackageStatus_String(t *testing.T) {
	tests := []struct {
		status   PackageStatus
		expected string
	}{
		{PackageStatusNotLoaded, "not_loaded"},
		{PackageStatusLoading, "loading"},
		{PackageStatusLoaded, "loaded"},
		{PackageStatusEnabled, "enabled"},
		{PackageStatusDisabled, "disabled"},
		{PackageStatusError, "error"},
		{PackageStatus(999), "unknown"},
	}
	
	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected status %d to return '%s', got '%s'", 
				test.status, test.expected, result)
		}
	}
}

func TestMockPackage(t *testing.T) {
	pkg := &mockPackage{
		info: PackageInfo{
			Name:        "test-package",
			Version:     "1.0.0",
			Description: "Test package",
			URL:         "github.com/user/test-package",
		},
	}
	
	// Test basic functionality
	info := pkg.GetInfo()
	if info.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", info.Name)
	}
	
	// Test enable/disable
	if pkg.IsEnabled() {
		t.Error("Expected package to be disabled by default")
	}
	
	err := pkg.Enable()
	if err != nil {
		t.Errorf("Expected no error enabling package, got: %v", err)
	}
	
	if !pkg.IsEnabled() {
		t.Error("Expected package to be enabled after Enable()")
	}
	
	err = pkg.Disable()
	if err != nil {
		t.Errorf("Expected no error disabling package, got: %v", err)
	}
	
	if pkg.IsEnabled() {
		t.Error("Expected package to be disabled after Disable()")
	}
	
	// Test initialization and cleanup
	err = pkg.Initialize()
	if err != nil {
		t.Errorf("Expected no error initializing package, got: %v", err)
	}
	
	err = pkg.Cleanup()
	if err != nil {
		t.Errorf("Expected no error cleaning up package, got: %v", err)
	}
}