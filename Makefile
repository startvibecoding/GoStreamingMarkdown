SHELL := /bin/bash

.DEFAULT_GOAL := help

BINARY  := GoStreamingMarkdown
MODULE  := .
GOFLAGS :=

.PHONY: help
help: ## Show available targets.
	@color=0; \
	if [ -t 1 ] && [ -z "$${NO_COLOR:-}" ] && [ "$${TERM:-}" != "dumb" ]; then color=1; fi; \
	COLOR="$$color" awk 'BEGIN { \
		FS = ":.*## "; \
		if (ENVIRON["COLOR"] == "1") { bold = "\033[1m"; cyan = "\033[36m"; reset = "\033[0m" } \
		printf "%sAvailable targets:%s\n", bold, reset \
	} /^[a-zA-Z0-9_.-]+:.*## / {printf "  %s%-24s%s %s\n", cyan, $$1, reset, $$2}' $(MAKEFILE_LIST)

# ── Build ────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build the mdrender binary.
	go build $(GOFLAGS) -o $(BINARY) $(MODULE)

.PHONY: install
install: ## Install the binary to $GOPATH/bin.
	go install $(GOFLAGS) $(MODULE)

.PHONY: clean
clean: ## Remove build artifacts.
	rm -f $(BINARY)

# ── Test ─────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests.
	go test ./... -count=1

.PHONY: test-v
test-v: ## Run all tests with verbose output.
	go test ./... -v -count=1

.PHONY: test-parser
test-parser: ## Run parser tests only.
	go test ./parser/... -v -count=1

.PHONY: test-renderer
test-renderer: ## Run renderer tests only.
	go test ./renderer/... -v -count=1

.PHONY: test-race
test-race: ## Run tests with race detector.
	go test ./... -race -count=1

.PHONY: test-cover
test-cover: ## Run tests with coverage report.
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out
	@rm -f coverage.out

.PHONY: bench
bench: ## Run benchmarks.
	go test ./... -bench=. -benchmem -count=1

# ── Code Quality ─────────────────────────────────────────────────────────────

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: fmt
fmt: ## Format all Go source files.
	go fmt ./...

.PHONY: lint
lint: vet ## Run all lint checks (vet + staticcheck if available).
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo 'staticcheck not installed (go install honnef.co/go/tools/cmd/staticcheck@latest)'

.PHONY: tidy
tidy: ## Run go mod tidy.
	go mod tidy

# ── CI ───────────────────────────────────────────────────────────────────────

.PHONY: ci
ci: fmt vet test build ## Run the same checks as CI.

# ── Utilities ────────────────────────────────────────────────────────────────

.PHONY: cloc
cloc: ## Count lines of code.
	@command -v cloc >/dev/null 2>&1 || { echo 'cloc not found. Install: apt install cloc / brew install cloc'; exit 1; }
	@cloc --vcs=git --exclude-dir=vendor 2>/dev/null || cloc .

.PHONY: run
run: build ## Build and run with example markdown.
	@echo '# Hello **world**' | ./$(BINARY)

.PHONY: demo
demo: build ## Run demo with a rich markdown document.
	@printf '%s\n' \
		'# GoStreamingMarkdown Demo' '' \
		'This is a **bold** and *italic* paragraph.' '' \
		'```go' 'func main() {' '    fmt.Println("Hello!")' '}' '```' '' \
		'| Name | Age |' '|------|-----|' '| Alice | 30 |' '| Bob | 25 |' '' \
		'> A blockquote.' '' \
		'- Item one' '- [x] Done' '- [ ] Todo' '' \
		'1. First' '2. Second' '' \
		'---' '' \
		'See [Go](https://go.dev).' \
		| ./$(BINARY) -w $${COLUMNS:-80}

.PHONY: demo-stream
demo-stream: build ## Run streaming demo.
	@echo 'Streaming in 3 seconds...'; sleep 3
	@( \
		echo '---BEGIN---'; sleep 0.3; \
		echo ''; sleep 0.2; \
		echo 'This text appears **word** by word.'; sleep 0.3; \
		echo ''; sleep 0.2; \
		echo '```go'; sleep 0.2; \
		echo 'func main() {'; sleep 0.2; \
		echo '    fmt.Println("Hello!")'; sleep 0.2; \
		echo '}'; sleep 0.2; \
		echo '```'; sleep 0.3; \
		echo ''; \
		echo '> Streaming blockquote with *italic* text.'; \
	) | ./$(BINARY) --stream --delay 50ms
