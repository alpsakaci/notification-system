.PHONY: all build run test fmt swag clean

# Default target
all: fmt swag build

# Build the applications
build:
	@echo "Building API and Consumer applications..."
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/consumer cmd/consumer/main.go

# Run the application
run: fmt swag
	@echo "Running the application..."
	@go run main.go

# Run tests and generate HTML coverage report (excluding main, docs, and infrastructure)
test:
	@echo "Running tests..."
	@go test -v ./... -coverprofile=coverage.raw.out
	@cat coverage.raw.out | grep -v "/cmd/" | grep -v "/docs" | grep -v "/internal/infrastructure" > coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@rm coverage.raw.out
	@echo "Coverage report generated at coverage.html"

# Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Generate swagger documentation
swag:
	@echo "Generating Swagger documentation..."
	@swag init -d cmd/api,internal -g main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
