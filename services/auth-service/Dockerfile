# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for go mod if needed
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/main.go

# Release stage
FROM alpine:3.19

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/auth-service .

# Copy .env file if present
COPY .env .env

# Expose gRPC port
EXPOSE 50051

# Run the service
ENTRYPOINT ["./auth-service"]
