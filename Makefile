.PHONY: build run test clean docker-build docker-up docker-down help

# Variables
BINARY_NAME=email-finder
DOCKER_IMAGE=email-finder:latest

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_NAME)"

run: ## Run the application locally
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	@docker-compose up -d
	@echo "Services started. Email Finder API: http://localhost:8080"

docker-down: ## Stop services with Docker Compose
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped"

docker-logs: ## View Docker logs
	@docker-compose logs -f

install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
