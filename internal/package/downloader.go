package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Downloader manages package downloads using go get
type Downloader struct {
	workDir     string // Working directory for downloads
	goModPath   string // Path to go.mod file for dependencies
	timeout     time.Duration
}

// NewDownloader creates a new package downloader
func NewDownloader(workDir string) *Downloader {
	return &Downloader{
		workDir:   workDir,
		goModPath: filepath.Join(workDir, "go.mod"),
		timeout:   5 * time.Minute, // 5 minute timeout for downloads
	}
}

// SetTimeout sets the download timeout
func (d *Downloader) SetTimeout(timeout time.Duration) {
	d.timeout = timeout
}

// InitializeWorkspace initializes the workspace for package downloads
func (d *Downloader) InitializeWorkspace() error {
	// Create work directory if it doesn't exist
	err := os.MkdirAll(d.workDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create work directory: %v", err)
	}
	
	// Initialize go.mod if it doesn't exist
	if _, err := os.Stat(d.goModPath); os.IsNotExist(err) {
		err = d.initGoMod()
		if err != nil {
			return fmt.Errorf("failed to initialize go.mod: %v", err)
		}
	}
	
	return nil
}

// initGoMod creates a go.mod file for managing dependencies
func (d *Downloader) initGoMod() error {
	goModContent := `module gmacs-packages

go 1.21

// This module is used to manage gmacs package dependencies
// Do not modify this file manually
`

	err := os.WriteFile(d.goModPath, []byte(goModContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create go.mod: %v", err)
	}
	
	return nil
}

// DownloadPackage downloads a package using go get
func (d *Downloader) DownloadPackage(url, version string) error {
	err := d.InitializeWorkspace()
	if err != nil {
		return err
	}
	
	// Validate package URL
	if err := d.validatePackageURL(url); err != nil {
		return fmt.Errorf("invalid package URL %s: %v", url, err)
	}
	
	// Construct package reference
	packageRef := url
	if version != "" && version != "latest" {
		packageRef = fmt.Sprintf("%s@%s", url, version)
	}
	
	fmt.Printf("Downloading package: %s\n", packageRef)
	
	// Execute go get command
	cmd := exec.Command("go", "get", packageRef)
	cmd.Dir = d.workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	// Set timeout
	done := make(chan error, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("go get failed: %v\nOutput: %s", err, string(output))
		} else {
			done <- nil
		}
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(d.timeout):
		cmd.Process.Kill()
		return fmt.Errorf("download timeout for package %s", packageRef)
	}
	
	fmt.Printf("Successfully downloaded: %s\n", packageRef)
	return nil
}

// validatePackageURL validates if the package URL is valid
func (d *Downloader) validatePackageURL(url string) error {
	if url == "" {
		return fmt.Errorf("package URL cannot be empty")
	}
	
	// Basic validation for common Git hosting services
	validPrefixes := []string{
		"github.com/",
		"gitlab.com/",
		"bitbucket.org/",
		"git.sr.ht/",
		"codeberg.org/",
	}
	
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			hasValidPrefix = true
			break
		}
	}
	
	if !hasValidPrefix {
		return fmt.Errorf("unsupported package host, supported: %v", validPrefixes)
	}
	
	// Check URL format (should have at least user/repo)
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid package URL format, expected: host.com/user/repo")
	}
	
	return nil
}

// GetDownloadedPackages returns a list of downloaded packages
func (d *Downloader) GetDownloadedPackages() ([]PackageInfo, error) {
	if _, err := os.Stat(d.goModPath); os.IsNotExist(err) {
		return []PackageInfo{}, nil
	}
	
	// Parse go.mod to get dependencies
	cmd := exec.Command("go", "list", "-m", "all")
	cmd.Dir = d.workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %v", err)
	}
	
	var packages []PackageInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "gmacs-packages") {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			packages = append(packages, PackageInfo{
				URL:     parts[0],
				Version: parts[1],
			})
		}
	}
	
	return packages, nil
}

// RemovePackage removes a downloaded package
func (d *Downloader) RemovePackage(url string) error {
	// Remove from go.mod
	cmd := exec.Command("go", "mod", "edit", "-droprequire", url)
	cmd.Dir = d.workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %v\nOutput: %s", url, err, string(output))
	}
	
	// Clean up unused modules
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = d.workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to tidy modules: %v\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("Successfully removed package: %s\n", url)
	return nil
}

// CleanWorkspace cleans up the download workspace
func (d *Downloader) CleanWorkspace() error {
	return os.RemoveAll(d.workDir)
}

// GetPackagePath returns the local filesystem path for a downloaded package
func (d *Downloader) GetPackagePath(url string) (string, error) {
	// Use go list to find the package path
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", url)
	cmd.Dir = d.workDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("package %s not found: %v", url, err)
	}
	
	packagePath := strings.TrimSpace(string(output))
	if packagePath == "" {
		return "", fmt.Errorf("package %s has no local path", url)
	}
	
	return packagePath, nil
}