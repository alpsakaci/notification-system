# Stage 1: Build the application
FROM golang:1.23.2-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git make

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/app main.go

# Stage 2: Create a minimal image
FROM alpine:latest

# Add ca-certificates and tzdata for HTTPS and timezone support
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/bin/app .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./app"]
