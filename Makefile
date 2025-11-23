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


test:
	@echo "Running unit tests..."
	go test -v -race -cover ./...

test-integration-local:
	@echo "Running integration tests against localhost:8080..."
	@echo "Ensure docker-compose is running: docker-compose up -d"
	@go test -v -count=1 ./test/integration/...

test-e2e:
	@echo "Starting e2e environment..."
	@docker-compose -f docker-compose.e2e.yml up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 8
	@echo "Running tests against http://localhost:8081..."
	@E2E_BASE_URL=http://localhost:8081 go test -v -count=1 ./test/integration/... || \
		(echo "\nTests failed. Showing logs:" && docker-compose -f docker-compose.e2e.yml logs api_e2e && \
		docker-compose -f docker-compose.e2e.yml down -v && exit 1)
	@echo "Tests passed. Cleaning up..."
	@docker-compose -f docker-compose.e2e.yml down -v
	@echo "E2E tests completed successfully!"

test-e2e-up:
	@echo "Starting e2e environment..."
	@docker-compose -f docker-compose.e2e.yml up -d --build
	@echo "Waiting for services..."
	@sleep 5
	@echo ""
	@echo "E2E environment is ready!"
	@echo "API available at: http://localhost:8081"
	@echo "Database available at: localhost:5433"
	@echo ""
	@echo "Run tests with:"
	@echo "  E2E_BASE_URL=http://localhost:8081 go test -v ./test/integration/..."
	@echo ""
	@echo "Stop with: make test-e2e-down"

test-e2e-down:
	@echo "Stopping e2e environment..."
	@docker-compose -f docker-compose.e2e.yml down -v
	@echo "E2E environment stopped and cleaned up."

test-e2e-logs:
	@docker-compose -f docker-compose.e2e.yml logs -f

test-integration: test-integration-local