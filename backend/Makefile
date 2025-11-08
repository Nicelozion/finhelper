.PHONY: help run build test clean deps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	go mod download
	go mod tidy

run: ## Run the server
	go run .

build: ## Build the binary
	go build -o finhelper-backend .

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	rm -f finhelper-backend
	rm -f coverage.out coverage.html

dev: deps ## Setup and run in development mode
	@echo "Starting development server..."
	@go run .

fmt: ## Format code
	go fmt ./...
	go vet ./...

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Please install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

docker-build: ## Build Docker image
	docker build -t finhelper-backend:latest .

docker-run: ## Run in Docker
	docker run -p 8080:8080 --env-file .env finhelper-backend:latest