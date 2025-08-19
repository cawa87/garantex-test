# Variables
APP_NAME=app
DOCKER_IMAGE=garantex-test
DOCKER_TAG=latest

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty)"

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building application..."
	go build $(LDFLAGS) -o $(APP_NAME) ./cmd/app

# Run unit tests
.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -race -cover ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run the application locally
.PHONY: run
run: build
	@echo "Running application..."
	./$(APP_NAME)

# Run with Docker Compose
.PHONY: run-docker
run-docker:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
.PHONY: stop-docker
stop-docker:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Install linter
.PHONY: install-lint
install-lint:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate protobuf files
.PHONY: proto
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/rate_service.v1/rate_service.proto

# Install protobuf tools
.PHONY: install-proto
install-proto:
	@echo "Installing protobuf tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Run database migrations
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/garantex_test?sslmode=disable" up

# Rollback database migrations
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/garantex_test?sslmode=disable" down

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	go clean -cache

# Clean Docker artifacts
.PHONY: clean-docker
clean-docker:
	@echo "Cleaning Docker artifacts..."
	docker-compose down -v --rmi all
	docker system prune -f

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run unit tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  docker-build  - Build Docker image"
	@echo "  run           - Run the application locally"
	@echo "  run-docker    - Run with Docker Compose"
	@echo "  stop-docker   - Stop Docker Compose services"
	@echo "  lint          - Run linter"
	@echo "  install-lint  - Install linter"
	@echo "  proto         - Generate protobuf files"
	@echo "  install-proto - Install protobuf tools"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  clean         - Clean build artifacts"
	@echo "  clean-docker  - Clean Docker artifacts"
	@echo "  help          - Show this help"
