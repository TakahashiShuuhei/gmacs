package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// TestDocumentation represents extracted documentation from test files
type TestDocumentation struct {
	Function     string
	FilePath     string
	SpecRef      string
	Scenario     string
	Description  string
	Given        string
	When         string
	Then         string
	Implementation []string
	BugFix       string
}

var (
	testDirFlag = flag.String("testdir", "e2e-test", "E2Eテストファイルディレクトリ")
	outputFlag  = flag.String("output", "specs/test-docs.md", "テストドキュメント出力ファイル")
)

func main() {
	flag.Parse()

	docs, err := extractTestDocs(*testDirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "テストドキュメント抽出エラー: %v\n", err)
		os.Exit(1)
	}

	err = generateTestDocumentation(docs, *outputFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ドキュメント生成エラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%d個のテスト関数のドキュメントを %s に生成しました\n", len(docs), *outputFlag)
}

func extractTestDocs(testDir string) ([]TestDocumentation, error) {
	var docs []TestDocumentation

	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			fileDocs, err := parseTestFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Error parsing %s: %v\n", path, err)
				return nil
			}
			docs = append(docs, fileDocs...)
		}

		return nil
	})

	return docs, err
}

func parseTestFile(filePath string) ([]TestDocumentation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var docs []TestDocumentation
	scanner := bufio.NewScanner(file)
	
	var currentDoc *TestDocumentation
	inComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Start of javadoc-style comment
		if strings.HasPrefix(line, "/**") {
			currentDoc = &TestDocumentation{FilePath: filePath}
			inComment = true
			continue
		}

		// End of javadoc-style comment
		if strings.HasPrefix(line, "*/") {
			inComment = false
			continue
		}

		// Parse javadoc annotations
		if inComment && strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
			
			if strings.HasPrefix(line, "@spec ") {
				currentDoc.SpecRef = strings.TrimSpace(strings.TrimPrefix(line, "@spec "))
			} else if strings.HasPrefix(line, "@scenario ") {
				currentDoc.Scenario = strings.TrimSpace(strings.TrimPrefix(line, "@scenario "))
			} else if strings.HasPrefix(line, "@description ") {
				currentDoc.Description = strings.TrimSpace(strings.TrimPrefix(line, "@description "))
			} else if strings.HasPrefix(line, "@given ") {
				currentDoc.Given = strings.TrimSpace(strings.TrimPrefix(line, "@given "))
			} else if strings.HasPrefix(line, "@when ") {
				currentDoc.When = strings.TrimSpace(strings.TrimPrefix(line, "@when "))
			} else if strings.HasPrefix(line, "@then ") {
				currentDoc.Then = strings.TrimSpace(strings.TrimPrefix(line, "@then "))
			} else if strings.HasPrefix(line, "@implementation ") {
				impl := strings.TrimSpace(strings.TrimPrefix(line, "@implementation "))
				currentDoc.Implementation = strings.Split(impl, ",")
				for i := range currentDoc.Implementation {
					currentDoc.Implementation[i] = strings.TrimSpace(currentDoc.Implementation[i])
				}
			} else if strings.HasPrefix(line, "@bug_fix ") {
				currentDoc.BugFix = strings.TrimSpace(strings.TrimPrefix(line, "@bug_fix "))
			}
			continue
		}

		// Parse function declaration
		if !inComment && currentDoc != nil {
			funcRegex := regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)
			if matches := funcRegex.FindStringSubmatch(line); matches != nil {
				currentDoc.Function = matches[1]
				docs = append(docs, *currentDoc)
				currentDoc = nil
			}
		}
	}

	return docs, scanner.Err()
}

func generateTestDocumentation(docs []TestDocumentation, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# gmacs テストドキュメント\n\n")
	fmt.Fprintf(file, "このドキュメントはテストコードから自動抽出されたBDD仕様書です。\n\n")
	fmt.Fprintf(file, "**生成日時:** %s\n\n", time.Now().Format("2006年01月02日 15:04:05"))

	// 仕様別にグループ化
	specGroups := make(map[string][]TestDocumentation)
	for _, doc := range docs {
		if doc.SpecRef != "" {
			specGroups[doc.SpecRef] = append(specGroups[doc.SpecRef], doc)
		}
	}

	// 機能名の日本語マッピング
	categoryNames := map[string]string{
		"display": "画面表示機能",
		"input":   "キーボード入力機能", 
		"scroll":  "スクロール機能",
		"buffer":  "バッファ管理機能",
		"editor":  "エディタ基本機能",
	}

	// specRefをソートして確定的な順序にする
	var sortedSpecs []string
	for specRef := range specGroups {
		sortedSpecs = append(sortedSpecs, specRef)
	}
	sort.Strings(sortedSpecs)

	for _, specRef := range sortedSpecs {
		specDocs := specGroups[specRef]
		
		// 機能カテゴリを日本語で表示
		parts := strings.Split(specRef, "/")
		categoryJP := specRef
		if len(parts) > 0 {
			if jp, exists := categoryNames[parts[0]]; exists {
				categoryJP = jp + " (" + specRef + ")"
			}
		}
		
		fmt.Fprintf(file, "## %s\n\n", categoryJP)
		
		// 各spec内でも関数名でソート
		sort.Slice(specDocs, func(i, j int) bool {
			return specDocs[i].Function < specDocs[j].Function
		})
		
		for _, doc := range specDocs {
			fmt.Fprintf(file, "### %s\n\n", doc.Function)
			fmt.Fprintf(file, "**ファイル:** `%s`\n\n", doc.FilePath)
			
			if doc.Scenario != "" {
				fmt.Fprintf(file, "**シナリオ:** %s\n\n", doc.Scenario)
			}
			
			if doc.Description != "" {
				fmt.Fprintf(file, "**説明:** %s\n\n", doc.Description)
			}
			
			if doc.Given != "" {
				fmt.Fprintf(file, "**前提:** %s\n\n", doc.Given)
			}
			
			if doc.When != "" {
				fmt.Fprintf(file, "**操作:** %s\n\n", doc.When)
			}
			
			if doc.Then != "" {
				fmt.Fprintf(file, "**結果:** %s\n\n", doc.Then)
			}
			
			if len(doc.Implementation) > 0 {
				fmt.Fprintf(file, "**実装ファイル:** ")
				for i, impl := range doc.Implementation {
					if i > 0 {
						fmt.Fprintf(file, ", ")
					}
					fmt.Fprintf(file, "`%s`", impl)
				}
				fmt.Fprintf(file, "\n\n")
			}
			
			if doc.BugFix != "" {
				fmt.Fprintf(file, "**バグ修正:** %s\n\n", doc.BugFix)
			}
			
			fmt.Fprintf(file, "---\n\n")
		}
	}

	fmt.Fprintf(file, "*このドキュメントは自動生成されています。修正はテストファイルのアノテーションを編集してください。*\n")

	return nil
}