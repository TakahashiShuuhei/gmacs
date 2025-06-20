# gmacs Makefile

.PHONY: test build clean docs test-docs all help

# Default target
all: build test docs

# Build the gmacs binary
build:
	go build -o gmacs

# Run all E2E tests
test:
	go test ./e2e-test/ -v

# Run specific E2E test pattern
test-pattern:
	go test ./e2e-test/ -v -run $(PATTERN)

# Generate test documentation from test code annotations
docs:
	@echo "テストドキュメントを抽出しています..."
	go run specs/tools/extract-test-docs.go
	@echo "テストドキュメントが specs/test-docs.md に生成されました"

# Alias for docs (backward compatibility)
test-docs: docs

# Clean build artifacts and generated docs
clean:
	rm -f gmacs
	rm -f specs/test-docs.md

# Run tests and generate documentation
verify: test docs
	@echo "全テストが成功し、ドキュメントが生成されました"

# Development workflow: build, test, and generate docs
dev: build test docs
	@echo "開発サイクルが完了しました"

# Show help
help:
	@echo "gmacs Makefile ターゲット一覧:"
	@echo "  build         - gmacsバイナリをビルド"
	@echo "  test          - 全E2Eテストを実行"
	@echo "  test-pattern  - パターンに一致するE2Eテストを実行 (例: make test-pattern PATTERN=TestName)"
	@echo "  docs          - テストコードからドキュメントを抽出"
	@echo "  test-docs     - docsのエイリアス"
	@echo "  clean         - ビルド成果物とドキュメントを削除"
	@echo "  verify        - テスト実行 + ドキュメント生成"
	@echo "  dev           - 完全な開発サイクル (build + test + docs)"
	@echo "  help          - このヘルプメッセージを表示"
	@echo ""
	@echo "BDD仕様書:"
	@echo "  specs/features/  - 日本語Gherkin仕様書（手動編集）"
	@echo "  specs/test-docs.md - テストから抽出した日本語ドキュメント（自動生成）"