.PHONY: help build run test lint docker-build docker-up docker-down clean

help:
	@echo "Available targets:"
	@echo "  make build         - Build the application binary"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make lint          - Run linter (golangci-lint)"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start services with docker-compose"
	@echo "  make docker-down   - Stop services"
	@echo "  make clean         - Remove build artifacts"

build:
	@echo "Building application..."
	go build -o bin/api ./cmd/api
	@echo "Build complete: bin/api"

run:
	@echo "Running application..."
	go run ./cmd/api

test:
	@echo "Running tests..."
	go test -v -race -cover ./...

lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
	fi

docker-build:
	@echo "Building Docker image..."
	docker build -t pr-reviewer-service:latest .

docker-up:
	@echo "Starting services..."
	docker-compose up --build

docker-up-detached:
	@echo "Starting services in background..."
	docker-compose up --build -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-down-volumes:
	@echo "Stopping services and removing volumes..."
	docker-compose down -v

docker-logs:
	docker-compose logs -f api

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "Clean complete"

fmt:
	@echo "Formatting code..."
	go fmt ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

mocks:
	@echo "Generating mocks..."
	@if command -v mockery > /dev/null; then \
		mockery --all --keeptree --output=internal/mocks; \
	else \
		echo "mockery not installed"; \
	fi

test-integration:
	@echo "Running integration tests..."
	@echo "Ensure docker-compose is running: docker-compose up -d"
	@go test -v -count=1 ./test/integration/...