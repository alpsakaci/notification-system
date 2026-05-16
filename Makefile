.PHONY: all build run test fmt swag clean

# Default target
all: fmt swag build

# Build the application
build:
	@echo "Building the application..."
	@go build -o bin/api main.go

# Run the application
run: fmt swag
	@echo "Running the application..."
	@go run main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Generate swagger documentation
swag:
	@echo "Generating Swagger documentation..."
	@swag init

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
