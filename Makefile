.PHONY: default
default: help

GO=go
GOLANGCI=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3
gofumpt=mvdan.cc/gofumpt@latest
govulncheck=golang.org/x/vuln/cmd/govulncheck@latest
actionlint=github.com/rhysd/actionlint/cmd/actionlint@latest

.PHONY: test
test:
	go test -v ./...

.PHONY: help
help: ## Show the available commands
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: fmt  ## Run linters on all go files
	$(GO) run $(GOLANGCI) run -v
	$(GO) run $(actionlint)
	@$(MAKE) sec

.PHONY: fmt
fmt: ## Formats all go files
	$(GO) run $(gofumpt) -l -w -extra  .

.PHONY: sec
sec: ## Run security checks
	$(GO) run $(govulncheck) ./...

