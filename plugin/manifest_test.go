package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest_ValidJSON(t *testing.T) {
	builder, err := NewPluginBuilder()
	if err != nil {
		t.Fatalf("NewPluginBuilder() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 有効なmanifest.jsonを作成
	manifestContent := `{
		"name": "test-plugin",
		"version": "2.1.0",
		"description": "Test plugin for JSON parsing",
		"author": "Test Author",
		"binary": "test-plugin-bin",
		"dependencies": ["dep1", "dep2"],
		"min_gmacs_version": "0.1.0",
		"default_config": {
			"setting1": "value1",
			"setting2": true
		}
	}`

	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// マニフェストを読み込み
	manifest, err := builder.loadManifest(tempDir)
	if err != nil {
		t.Fatalf("loadManifest() error = %v", err)
	}

	// 値の検証
	if manifest.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", manifest.Name)
	}
	if manifest.Version != "2.1.0" {
		t.Errorf("Expected version '2.1.0', got '%s'", manifest.Version)
	}
	if manifest.Description != "Test plugin for JSON parsing" {
		t.Errorf("Expected description 'Test plugin for JSON parsing', got '%s'", manifest.Description)
	}
	if manifest.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", manifest.Author)
	}
	if manifest.Binary != "test-plugin-bin" {
		t.Errorf("Expected binary 'test-plugin-bin', got '%s'", manifest.Binary)
	}
	if manifest.MinGmacs != "0.1.0" {
		t.Errorf("Expected min_gmacs_version '0.1.0', got '%s'", manifest.MinGmacs)
	}

	t.Logf("Successfully parsed manifest: %+v", manifest)
}

func TestLoadManifest_MinimalJSON(t *testing.T) {
	builder, err := NewPluginBuilder()
	if err != nil {
		t.Fatalf("NewPluginBuilder() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 最小限のmanifest.json（nameのみ）
	manifestContent := `{
		"name": "minimal-plugin"
	}`

	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// マニフェストを読み込み
	manifest, err := builder.loadManifest(tempDir)
	if err != nil {
		t.Fatalf("loadManifest() error = %v", err)
	}

	// デフォルト値の検証
	if manifest.Name != "minimal-plugin" {
		t.Errorf("Expected name 'minimal-plugin', got '%s'", manifest.Name)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", manifest.Version)
	}
	if manifest.Binary != "minimal-plugin" {
		t.Errorf("Expected default binary 'minimal-plugin', got '%s'", manifest.Binary)
	}
	if manifest.Description != "Plugin built from source" {
		t.Errorf("Expected default description, got '%s'", manifest.Description)
	}
	if manifest.Author != "Unknown" {
		t.Errorf("Expected default author 'Unknown', got '%s'", manifest.Author)
	}

	t.Logf("Successfully parsed minimal manifest with defaults: %+v", manifest)
}

func TestLoadManifest_InvalidJSON(t *testing.T) {
	builder, err := NewPluginBuilder()
	if err != nil {
		t.Fatalf("NewPluginBuilder() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// 無効なJSON
	manifestContent := `{
		"name": "invalid-plugin",
		"version": "1.0.0"
		// missing comma and invalid syntax
	}`

	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// マニフェスト読み込みがエラーになることを確認
	_, err = builder.loadManifest(tempDir)
	if err == nil {
		t.Error("Expected error for invalid JSON, but got nil")
	}
	if err != nil {
		t.Logf("✓ Correctly returned error for invalid JSON: %v", err)
	}
}

func TestLoadManifest_MissingNameField(t *testing.T) {
	builder, err := NewPluginBuilder()
	if err != nil {
		t.Fatalf("NewPluginBuilder() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// name フィールドがないJSON
	manifestContent := `{
		"version": "1.0.0",
		"description": "Plugin without name"
	}`

	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// nameフィールドが必須なのでエラーになることを確認
	_, err = builder.loadManifest(tempDir)
	if err == nil {
		t.Error("Expected error for missing name field, but got nil")
	}
	if err != nil {
		t.Logf("✓ Correctly returned error for missing name: %v", err)
	}
}

func TestLoadManifest_MissingFile(t *testing.T) {
	builder, err := NewPluginBuilder()
	if err != nil {
		t.Fatalf("NewPluginBuilder() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成（manifest.jsonなし）
	tempDir := t.TempDir()

	// manifest.jsonが存在しない場合のエラー確認
	_, err = builder.loadManifest(tempDir)
	if err == nil {
		t.Error("Expected error for missing manifest.json, but got nil")
	}
	if err != nil {
		t.Logf("✓ Correctly returned error for missing manifest.json: %v", err)
	}
}

func TestLoadSimpleManifest_ManagerVersion(t *testing.T) {
	manager := NewPluginManager()

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// example-plugin形式のmanifest.json
	manifestContent := `{
		"name": "example-plugin",
		"version": "1.0.0",
		"description": "A simple example plugin for gmacs demonstrating basic functionality",
		"author": "gmacs team",
		"binary": "example-plugin",
		"dependencies": [],
		"min_gmacs_version": "0.1.0",
		"default_config": {
			"greeting_message": "Hello from example plugin!",
			"auto_greet": true,
			"prefix": "[EXAMPLE]"
		}
	}`

	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// マニフェストを読み込み
	manifest, err := manager.loadSimpleManifest(tempDir)
	if err != nil {
		t.Fatalf("loadSimpleManifest() error = %v", err)
	}

	// example-pluginの値が正しく読み込まれることを確認
	if manifest.Name != "example-plugin" {
		t.Errorf("Expected name 'example-plugin', got '%s'", manifest.Name)
	}
	if manifest.Description != "A simple example plugin for gmacs demonstrating basic functionality" {
		t.Errorf("Unexpected description: %s", manifest.Description)
	}
	if manifest.Author != "gmacs team" {
		t.Errorf("Expected author 'gmacs team', got '%s'", manifest.Author)
	}

	t.Logf("Successfully parsed example-plugin manifest: %+v", manifest)
}

func TestLoadSimpleManifest_FallbackMode(t *testing.T) {
	manager := NewPluginManager()

	// テスト用の一時ディレクトリを作成（manifest.jsonなし）
	tempDir := t.TempDir()

	// manifest.jsonが存在しない場合のフォールバック動作
	manifest, err := manager.loadSimpleManifest(tempDir)
	if err != nil {
		t.Fatalf("loadSimpleManifest() error = %v", err)
	}

	// ディレクトリ名ベースのデフォルト値が設定されることを確認
	expectedName := filepath.Base(tempDir)
	if manifest.Name != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, manifest.Name)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", manifest.Version)
	}
	if manifest.Description != "Plugin installed from source" {
		t.Errorf("Expected default description, got '%s'", manifest.Description)
	}

	t.Logf("Successfully used fallback manifest: %+v", manifest)
}