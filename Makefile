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

.PHONY: lint
lint: fmt
	docker run --rm -v $(shell pwd):/app:ro -w /app golangci/golangci-lint:v1.52.2 bash -e -c \
		'golangci-lint run -v --timeout 5m'

.PHONY: fmt
fmt: install-tools
	govulncheck ./...
	gofumpt -l -w -extra  .

