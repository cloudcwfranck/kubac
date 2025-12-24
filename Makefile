.PHONY: help build test lint clean install e2e-kind e2e-k3d

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build kubac CLI
	go build -o kubac ./cmd/kubac

test: ## Run unit tests
	go test -v -race -coverprofile=coverage.txt ./...

lint: ## Run golangci-lint
	golangci-lint run --timeout=5m

clean: ## Clean build artifacts
	rm -f kubac coverage.txt kubac.yaml kubac-verify-report.json

install: build ## Build and install kubac to /usr/local/bin
	sudo mv kubac /usr/local/bin/

e2e-kind: build ## Run e2e tests on kind
	./test/kind/test-kind.sh

e2e-k3d: build ## Run e2e tests on k3d
	./test/k3d/test-k3d.sh

fmt: ## Format code
	go fmt ./...
	goimports -w -local github.com/cloudcwfranck/kubac .

vet: ## Run go vet
	go vet ./...

mod: ## Tidy and verify go modules
	go mod tidy
	go mod verify

all: fmt vet lint test build ## Run all checks and build

.DEFAULT_GOAL := help
