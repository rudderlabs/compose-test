.PHONY: default
default: help

.PHONY: test
test:
	go test -v ./...

.PHONY: help
help: ## Show the available commands
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: install-tools
install-tools:
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	bash ./scripts/install-golangci-lint.sh v1.55.2

.PHONY: lint
lint: fmt
	golangci-lint run -v --timeout 5m

.PHONY: fmt
fmt: install-tools
	govulncheck ./...
	gofumpt -l -w -extra  .

