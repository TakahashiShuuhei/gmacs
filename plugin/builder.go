package plugin

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PluginBuilder はプラグインのソースからビルドを行う
type PluginBuilder struct {
	workDir     string              // ビルド作業ディレクトリ
	cacheDir    string              // ビルドキャッシュディレクトリ
	targetDir   string              // プラグイン配置ディレクトリ
	buildCache  map[string]*BuildCache // ビルドキャッシュ
}

// BuildRequest はビルド要求を表す
type BuildRequest struct {
	Repository string // Git repository URL (e.g., "github.com/user/plugin")
	Ref        string // branch/tag/commit (default: "main")
	LocalPath  string // ローカルパス（開発用、Repositoryより優先）
	Force      bool   // キャッシュを無視してビルド
}

// BuildResult はビルド結果を表す
type BuildResult struct {
	PluginName   string    // プラグイン名
	Version      string    // バージョン
	BinaryPath   string    // 生成されたバイナリパス
	ManifestPath string    // manifest.jsonパス
	BuildTime    time.Time // ビルド時刻
	FromCache    bool      // キャッシュから取得したか
}

// NewPluginBuilder は新しいPluginBuilderを作成する
func NewPluginBuilder() (*PluginBuilder, error) {
	// XDG準拠のディレクトリを取得
	workDir := filepath.Join(os.TempDir(), "gmacs-plugin-build")
	cacheDir := filepath.Join(GetXDGCacheHome(), "gmacs", "plugin-build")
	targetDir := filepath.Join(GetXDGDataHome(), "gmacs", "plugins")

	// ディレクトリを作成
	for _, dir := range []string{workDir, cacheDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return &PluginBuilder{
		workDir:    workDir,
		cacheDir:   cacheDir,
		targetDir:  targetDir,
		buildCache: make(map[string]*BuildCache),
	}, nil
}

// BuildFromRepository はGitリポジトリからプラグインをビルドする
func (pb *PluginBuilder) BuildFromRepository(req BuildRequest) (*BuildResult, error) {
	if req.Repository == "" && req.LocalPath == "" {
		return nil, fmt.Errorf("either Repository or LocalPath must be specified")
	}

	// ローカルパスが指定されている場合はそちらを優先
	var sourceDir string
	var err error

	if req.LocalPath != "" {
		sourceDir, err = pb.prepareLocalSource(req.LocalPath)
	} else {
		sourceDir, err = pb.cloneRepository(req)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to prepare source: %v", err)
	}
	defer os.RemoveAll(sourceDir) // クリーンアップ

	// ソースコードハッシュを計算
	sourceHash, err := pb.calculateSourceHash(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate source hash: %v", err)
	}

	// キャッシュチェック
	if !req.Force {
		if cached := pb.getCachedBuild(sourceHash); cached != nil {
			return &BuildResult{
				PluginName:   cached.PluginName,
				Version:      cached.Version,
				BinaryPath:   cached.BinaryPath,
				ManifestPath: cached.ManifestPath,
				BuildTime:    cached.BuildTime,
				FromCache:    true,
			}, nil
		}
	}

	// manifest.jsonを読み込んでプラグイン情報を取得
	manifest, err := pb.loadManifest(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %v", err)
	}
	
	// デバッグ: マニフェスト内容を出力
	fmt.Printf("DEBUG: Loaded manifest - Name: %s, Version: %s, Binary: %s\n", 
		manifest.Name, manifest.Version, manifest.Binary)

	// ビルド実行
	result, err := pb.buildPlugin(sourceDir, manifest, sourceHash)
	if err != nil {
		return nil, fmt.Errorf("build failed: %v", err)
	}

	return result, nil
}

// prepareLocalSource はローカルソースを準備する
func (pb *PluginBuilder) prepareLocalSource(localPath string) (string, error) {
	// 絶対パスに変換
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// ディレクトリが存在するかチェック
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("local path does not exist: %s", absPath)
	}

	// 作業ディレクトリにコピー
	workSourceDir := filepath.Join(pb.workDir, fmt.Sprintf("local-source-%d", time.Now().UnixNano()))
	if err := pb.copyDir(absPath, workSourceDir); err != nil {
		return "", fmt.Errorf("failed to copy source: %v", err)
	}

	return workSourceDir, nil
}

// cloneRepository はGitリポジトリをクローンする
func (pb *PluginBuilder) cloneRepository(req BuildRequest) (string, error) {
	ref := req.Ref
	if ref == "" {
		ref = "main"
	}

	// クローン先ディレクトリ
	cloneDir := filepath.Join(pb.workDir, fmt.Sprintf("clone-%d", time.Now().UnixNano()))

	// git clone実行
	var repoURL string
	if strings.HasPrefix(req.Repository, "http://") || strings.HasPrefix(req.Repository, "https://") {
		repoURL = req.Repository
	} else {
		// github.com/user/repo形式の場合はHTTPS URLに変換
		repoURL = "https://" + req.Repository + ".git"
	}

	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", ref, repoURL, cloneDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %v\nOutput: %s", err, output)
	}

	return cloneDir, nil
}

// calculateSourceHash はソースコードのハッシュを計算する
func (pb *PluginBuilder) calculateSourceHash(sourceDir string) (string, error) {
	hasher := sha256.New()

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .gitディレクトリは無視
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// 通常ファイルのみハッシュ対象
		if !info.Mode().IsRegular() {
			return nil
		}

		// ファイル内容をハッシュに追加
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hasher, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// getCachedBuild はキャッシュされたビルドを取得する
func (pb *PluginBuilder) getCachedBuild(sourceHash string) *BuildCache {
	// TODO: ディスクからキャッシュを読み込む実装
	return pb.buildCache[sourceHash]
}

// buildPlugin は実際のビルドを実行する
func (pb *PluginBuilder) buildPlugin(sourceDir string, manifest *PluginManifest, sourceHash string) (*BuildResult, error) {
	// デバッグ: マニフェスト情報を出力
	fmt.Printf("DEBUG: buildPlugin - manifest.Name: %s, sourceDir: %s\n", manifest.Name, sourceDir)
	
	// プラグイン専用ディレクトリを作成
	pluginDir := filepath.Join(pb.targetDir, manifest.Name)
	fmt.Printf("DEBUG: Creating plugin directory: %s\n", pluginDir)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// go mod download
	if err := pb.runGoModDownload(sourceDir); err != nil {
		return nil, fmt.Errorf("go mod download failed: %v", err)
	}

	// go build
	binaryName := manifest.Binary
	if binaryName == "" {
		binaryName = manifest.Name
	}
	binaryPath := filepath.Join(pluginDir, binaryName)

	if err := pb.runGoBuild(sourceDir, binaryPath); err != nil {
		return nil, fmt.Errorf("go build failed: %v", err)
	}

	// manifest.jsonをコピー
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := pb.copyFile(filepath.Join(sourceDir, "manifest.json"), manifestPath); err != nil {
		return nil, fmt.Errorf("failed to copy manifest: %v", err)
	}

	buildTime := time.Now()
	result := &BuildResult{
		PluginName:   manifest.Name,
		Version:      manifest.Version,
		BinaryPath:   binaryPath,
		ManifestPath: manifestPath,
		BuildTime:    buildTime,
		FromCache:    false,
	}

	// キャッシュに保存
	pb.buildCache[sourceHash] = &BuildCache{
		PluginName:   manifest.Name,
		Version:      manifest.Version,
		Hash:         sourceHash,
		BuildTime:    buildTime,
		BinaryPath:   binaryPath,
		ManifestPath: manifestPath,
	}

	return result, nil
}

// runGoModDownload はgo mod downloadを実行する
func (pb *PluginBuilder) runGoModDownload(sourceDir string) error {
	// go.modファイルの相対参照を確認・修正
	if err := pb.fixGoModReplacements(sourceDir); err != nil {
		return fmt.Errorf("failed to fix go.mod replacements: %v", err)
	}
	
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = sourceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod download failed: %v\nOutput: %s", err, output)
	}
	return nil
}

// fixGoModReplacements はgo.modの相対参照を修正する
func (pb *PluginBuilder) fixGoModReplacements(sourceDir string) error {
	goModPath := filepath.Join(sourceDir, "go.mod")
	
	// go.modファイルが存在しない場合はスキップ
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil
	}
	
	// go.modファイルを読み込み
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %v", err)
	}
	
	modContent := string(content)
	
	// ローカル参照（../）を含むreplaceディレクティブを削除または修正
	lines := strings.Split(modContent, "\n")
	var modifiedLines []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 相対パス参照のreplaceディレクティブをスキップ
		if strings.HasPrefix(trimmed, "replace ") && strings.Contains(trimmed, "../") {
			// gmacs-plugin-sdkの場合は公開版に置き換え
			if strings.Contains(trimmed, "gmacs-plugin-sdk") {
				// replace行をコメントアウト（公開リポジトリを使用）
				modifiedLines = append(modifiedLines, "// "+line+" // Commented out for build")
				continue
			}
			// その他の相対参照もコメントアウト
			modifiedLines = append(modifiedLines, "// "+line+" // Commented out for build")
			continue
		}
		
		modifiedLines = append(modifiedLines, line)
	}
	
	// 修正されたgo.modを書き戻し
	modifiedContent := strings.Join(modifiedLines, "\n")
	if err := os.WriteFile(goModPath, []byte(modifiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write modified go.mod: %v", err)
	}
	
	return nil
}

// runGoBuild はgo buildを実行する
func (pb *PluginBuilder) runGoBuild(sourceDir, outputPath string) error {
	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Dir = sourceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %v\nOutput: %s", err, output)
	}
	return nil
}

// copyDir はディレクトリを再帰的にコピーする
func (pb *PluginBuilder) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 相対パスを計算
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return pb.copyFile(path, dstPath)
	})
}

// copyFile はファイルをコピーする
func (pb *PluginBuilder) copyFile(src, dst string) error {
	// ディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// loadManifest はmanifest.jsonを読み込む
func (pb *PluginBuilder) loadManifest(sourceDir string) (*PluginManifest, error) {
	manifestPath := filepath.Join(sourceDir, "manifest.json")
	
	// manifest.jsonの存在チェック
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest.json not found in %s", sourceDir)
	}
	
	// manifest.jsonファイルを読み込み
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %v", err)
	}
	
	// JSONをパース
	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %v", err)
	}
	
	// デバッグ: パース後のmanifest.Name確認
	fmt.Printf("DEBUG: loadManifest parsed name: %s from file: %s\n", manifest.Name, manifestPath)
	
	// 必須フィールドの検証
	if manifest.Name == "" {
		return nil, fmt.Errorf("manifest.json: name field is required")
	}
	if manifest.Version == "" {
		manifest.Version = "1.0.0" // デフォルト値
	}
	if manifest.Binary == "" {
		manifest.Binary = manifest.Name // デフォルトはプラグイン名
	}
	if manifest.Description == "" {
		manifest.Description = "Plugin built from source"
	}
	if manifest.Author == "" {
		manifest.Author = "Unknown"
	}
	
	return &manifest, nil
}