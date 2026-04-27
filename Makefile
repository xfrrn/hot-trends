.PHONY: build run test clean docker-build docker-run

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -f deployments/Dockerfile -t hot-trends:latest .

# Run with docker-compose
docker-run:
	docker-compose -f deployments/docker-compose.yml up -d

# Stop docker-compose
docker-stop:
	docker-compose -f deployments/docker-compose.yml down

# View logs
docker-logs:
	docker-compose -f deployments/docker-compose.yml logs -f

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Development mode with hot reload (requires air)
dev:
	air
